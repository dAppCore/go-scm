// SPDX-License-Identifier: EUPL-1.2

package agentci

import (
	// Note: context.Context is retained as the public cancellation contract for SSH command construction.
	"context"
	// Note: fmt.Errorf is retained for wrapped validation errors across path and SSH helpers.
	"fmt"
	// Note: os/exec is retained because these helpers intentionally construct ssh exec.Cmd values.
	"os/exec"
	// Note: path is retained for remote POSIX path normalization independent of host OS filepath semantics.
	"path"
	// Note: filepath is retained for host OS path containment checks under local roots.
	"path/filepath"
	// Note: regexp is retained for path-element allowlist validation; no core equivalent covers compiled regexes.
	"regexp"
	// Note: strings helpers are retained for path sanitation and shell-argument escaping.
	"strings"
)

var safeNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\-\_\.]+$`)

// SanitizePath ensures a filename or directory name is safe and prevents path traversal.
// Returns the validated input unchanged.
// Usage: SanitizePath(...)
func SanitizePath(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("agentci.SanitizePath: path element is required")
	}
	if input == "." || input == ".." {
		return "", fmt.Errorf("agentci.SanitizePath: invalid path element: %s", input)
	}
	if strings.ContainsAny(input, `/\`) {
		return "", fmt.Errorf("agentci.SanitizePath: path separators are not allowed: %s", input)
	}
	if !safeNameRegex.MatchString(input) {
		return "", fmt.Errorf("agentci.SanitizePath: invalid characters in path element: %s", input)
	}
	return input, nil
}

// ValidatePathElement validates a single local path element and returns its safe form.
// Usage: ValidatePathElement(...)
func ValidatePathElement(input string) (string, error) {
	safeName, err := SanitizePath(input)
	if err != nil {
		return "", err
	}
	if safeName != input {
		return "", fmt.Errorf("agentci.ValidatePathElement: path separators are not allowed: %s", input)
	}
	return safeName, nil
}

// ResolvePathWithinRoot resolves a validated path element beneath a root directory.
// Usage: ResolvePathWithinRoot(...)
func ResolvePathWithinRoot(root string, input string) (string, string, error) {
	if strings.TrimSpace(root) == "" {
		return "", "", fmt.Errorf("agentci.ResolvePathWithinRoot: root is required")
	}

	safeName, err := ValidatePathElement(input)
	if err != nil {
		return "", "", fmt.Errorf("agentci.ResolvePathWithinRoot: invalid path element: %w", err)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", "", fmt.Errorf("agentci.ResolvePathWithinRoot: resolve root: %w", err)
	}

	resolved := filepath.Clean(filepath.Join(absRoot, safeName))
	cleanRoot := filepath.Clean(absRoot)
	if cleanRoot == string(filepath.Separator) {
		if !strings.HasPrefix(resolved, cleanRoot) {
			return "", "", fmt.Errorf("agentci.ResolvePathWithinRoot: resolved path escaped root")
		}
		return safeName, resolved, nil
	}

	rootPrefix := cleanRoot + string(filepath.Separator)
	if resolved != cleanRoot && !strings.HasPrefix(resolved, rootPrefix) {
		return "", "", fmt.Errorf("agentci.ResolvePathWithinRoot: resolved path escaped root")
	}

	return safeName, resolved, nil
}

// ValidateRemoteDir validates a remote directory path used over SSH.
// Usage: ValidateRemoteDir(...)
func ValidateRemoteDir(dir string) (string, error) {
	if strings.TrimSpace(dir) == "" {
		return "", fmt.Errorf("agentci.ValidateRemoteDir: directory is required")
	}
	if strings.ContainsAny(dir, `\`) {
		return "", fmt.Errorf("agentci.ValidateRemoteDir: backslashes are not allowed")
	}

	switch dir {
	case "/", "~":
		return dir, nil
	}

	prefix := ""
	rest := dir

	if strings.HasPrefix(dir, "~/") {
		prefix = "~/"
		rest = strings.TrimPrefix(dir, "~/")
	}
	if strings.HasPrefix(dir, "/") {
		prefix = "/"
		rest = strings.TrimPrefix(dir, "/")
	}

	for _, part := range strings.Split(rest, "/") {
		if part == "" {
			continue
		}
		if part == "." || part == ".." {
			return "", fmt.Errorf("agentci.ValidateRemoteDir: directory escaped root")
		}
		if _, err := ValidatePathElement(part); err != nil {
			return "", fmt.Errorf("agentci.ValidateRemoteDir: invalid directory segment: %w", err)
		}
	}

	if rest == "" || rest == "." {
		if prefix == "~/" {
			return "~", nil
		}
		return prefix, nil
	}

	return path.Clean(dir), nil
}

// JoinRemotePath joins validated remote path elements using forward slashes.
// Usage: JoinRemotePath(...)
func JoinRemotePath(base string, parts ...string) (string, error) {
	safeBase, err := ValidateRemoteDir(base)
	if err != nil {
		return "", fmt.Errorf("agentci.JoinRemotePath: invalid base directory: %w", err)
	}
	if len(parts) == 0 {
		return safeBase, nil
	}

	cleanParts := make([]string, 0, len(parts))
	for _, part := range parts {
		safePart, partErr := ValidatePathElement(part)
		if partErr != nil {
			return "", fmt.Errorf("agentci.JoinRemotePath: invalid path element: %w", partErr)
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
// Usage: EscapeShellArg(...)
func EscapeShellArg(arg string) string {
	return "'" + strings.ReplaceAll(arg, "'", "'\\''") + "'"
}

// SecureSSHCommand creates an SSH exec.Cmd with strict host key checking and batch mode.
// Usage: SecureSSHCommand(...)
func SecureSSHCommand(host string, remoteCmd string) *exec.Cmd {
	return SecureSSHCommandContext(context.Background(), host, remoteCmd)
}

// SecureSSHCommandContext creates an SSH exec.Cmd with strict host key checking and batch mode.
// Usage: SecureSSHCommandContext(...)
func SecureSSHCommandContext(ctx context.Context, host string, remoteCmd string) *exec.Cmd {
	if ctx == nil {
		ctx = context.Background()
	}

	return exec.CommandContext(ctx, "ssh",
		"-o", "StrictHostKeyChecking=yes",
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=10",
		host,
		remoteCmd,
	)
}

// MaskToken returns a masked version of a token for safe logging.
// Usage: MaskToken(...)
func MaskToken(token string) string {
	if len(token) < 8 {
		return "*****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}
