// SPDX-License-Identifier: EUPL-1.2

package agentci

import (
	strings "dappco.re/go/core/scm/internal/ax/stringsx"
	exec "golang.org/x/sys/execabs"
	"path"
	"regexp"

	coreerr "dappco.re/go/core/log"
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
)

var safeNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\-\_\.]+$`)

// SanitizePath ensures a filename or directory name is safe and prevents path traversal.
// Returns the validated input unchanged.
//
func SanitizePath(input string) (string, error) {
	if input == "" {
		return "", coreerr.E("agentci.SanitizePath", "path element is required", nil)
	}
	if strings.ContainsAny(input, `/\`) {
		return "", coreerr.E("agentci.SanitizePath", "path separators are not allowed: "+input, nil)
	}
	if input == "." || input == ".." {
		return "", coreerr.E("agentci.SanitizePath", "invalid path element: "+input, nil)
	}
	if !safeNameRegex.MatchString(input) {
		return "", coreerr.E("agentci.SanitizePath", "invalid characters in path element: "+input, nil)
	}
	return input, nil
}

// ValidatePathElement validates a single local path element and returns its safe form.
//
func ValidatePathElement(input string) (string, error) {
	return SanitizePath(input)
}

// ResolvePathWithinRoot resolves a validated path element beneath a root directory.
//
func ResolvePathWithinRoot(root string, input string) (string, string, error) {
	safeName, err := ValidatePathElement(input)
	if err != nil {
		return "", "", coreerr.E("agentci.ResolvePathWithinRoot", "invalid path element", err)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", "", coreerr.E("agentci.ResolvePathWithinRoot", "resolve root", err)
	}

	resolved := filepath.Clean(filepath.Join(absRoot, safeName))
	cleanRoot := filepath.Clean(absRoot)
	rootPrefix := cleanRoot + string(filepath.Separator)
	if resolved != cleanRoot && !strings.HasPrefix(resolved, rootPrefix) {
		return "", "", coreerr.E("agentci.ResolvePathWithinRoot", "resolved path escaped root", nil)
	}

	return safeName, resolved, nil
}

// ValidateRemoteDir validates a remote directory path used over SSH.
//
func ValidateRemoteDir(dir string) (string, error) {
	if strings.TrimSpace(dir) == "" {
		return "", coreerr.E("agentci.ValidateRemoteDir", "directory is required", nil)
	}
	if strings.ContainsAny(dir, `\`) {
		return "", coreerr.E("agentci.ValidateRemoteDir", "backslashes are not allowed", nil)
	}

	switch dir {
	case "/", "~":
		return dir, nil
	}

	cleaned := path.Clean(dir)
	prefix := ""
	rest := cleaned

	if strings.HasPrefix(dir, "~/") {
		prefix = "~/"
		rest = strings.TrimPrefix(cleaned, "~/")
	}
	if strings.HasPrefix(dir, "/") {
		prefix = "/"
		rest = strings.TrimPrefix(cleaned, "/")
	}

	if rest == "." || rest == ".." || strings.HasPrefix(rest, "../") {
		return "", coreerr.E("agentci.ValidateRemoteDir", "directory escaped root", nil)
	}

	for _, part := range strings.Split(rest, "/") {
		if part == "" {
			continue
		}
		if _, err := ValidatePathElement(part); err != nil {
			return "", coreerr.E("agentci.ValidateRemoteDir", "invalid directory segment", err)
		}
	}

	if rest == "" || rest == "." {
		return prefix, nil
	}

	return prefix + rest, nil
}

// JoinRemotePath joins validated remote path elements using forward slashes.
//
func JoinRemotePath(base string, parts ...string) (string, error) {
	safeBase, err := ValidateRemoteDir(base)
	if err != nil {
		return "", coreerr.E("agentci.JoinRemotePath", "invalid base directory", err)
	}

	cleanParts := make([]string, 0, len(parts))
	for _, part := range parts {
		safePart, partErr := ValidatePathElement(part)
		if partErr != nil {
			return "", coreerr.E("agentci.JoinRemotePath", "invalid path element", partErr)
		}
		cleanParts = append(cleanParts, safePart)
	}

	if safeBase == "~" {
		return path.Join("~", path.Join(cleanParts...)), nil
	}
	if strings.HasPrefix(safeBase, "~/") {
		return "~/" + path.Join(strings.TrimPrefix(safeBase, "~/"), path.Join(cleanParts...)), nil
	}
	return path.Join(append([]string{safeBase}, cleanParts...)...), nil
}

// EscapeShellArg wraps a string in single quotes for safe remote shell insertion.
// Prefer exec.Command arguments over constructing shell strings where possible.
//
func EscapeShellArg(arg string) string {
	return "'" + strings.ReplaceAll(arg, "'", "'\\''") + "'"
}

// SecureSSHCommand creates an SSH exec.Cmd with strict host key checking and batch mode.
//
func SecureSSHCommand(host string, remoteCmd string) *exec.Cmd {
	return exec.Command("ssh",
		"-o", "StrictHostKeyChecking=yes",
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=10",
		host,
		remoteCmd,
	)
}

// MaskToken returns a masked version of a token for safe logging.
//
func MaskToken(token string) string {
	if len(token) < 8 {
		return "*****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}
