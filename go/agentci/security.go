// SPDX-License-Identifier: EUPL-1.2

package agentci

import (
	// Note: context.Context is retained as the public cancellation contract for SSH command construction.
	"context"
	"os/exec" // Note: AX-6 — process invocation via os/exec is intrinsic for SSH exec.Cmd construction returned by this public API.
	// Note: regexp is retained for path-element allowlist validation; no core equivalent covers compiled regexes.
	"regexp"

	core "dappco.re/go"
)

const (
	sonarSecurityAgentciResolvepathwithinroot = "agentci.ResolvePathWithinRoot"
	sonarSecurityAgentciSanitizepath          = "agentci.SanitizePath"
	sonarSecurityAgentciValidateremotedir     = "agentci.ValidateRemoteDir"
)

var safeNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\-\_\.]+$`)

// SanitizePath ensures a filename or directory name is safe and prevents path traversal.
// Returns the validated input unchanged.
// Usage: SanitizePath(...)
func SanitizePath(input string) (string, error) {
	if input == "" {
		return "", core.E(sonarSecurityAgentciSanitizepath, "path element is required", nil)
	}
	if input == "." || input == ".." {
		return "", core.E(sonarSecurityAgentciSanitizepath, core.Sprintf("invalid path element: %s", input), nil)
	}
	if core.Contains(input, "/") || core.Contains(input, `\`) {
		return "", core.E(sonarSecurityAgentciSanitizepath, core.Sprintf("path separators are not allowed: %s", input), nil)
	}
	if !safeNameRegex.MatchString(input) {
		return "", core.E(sonarSecurityAgentciSanitizepath, core.Sprintf("invalid characters in path element: %s", input), nil)
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
		return "", core.E("agentci.ValidatePathElement", core.Sprintf("path separators are not allowed: %s", input), nil)
	}
	return safeName, nil
}

// ResolvePathWithinRoot resolves a validated path element beneath a root directory.
// Usage: ResolvePathWithinRoot(...)
func ResolvePathWithinRoot(root string, input string) (string, string, error) {
	if core.Trim(root) == "" {
		return "", "", core.E(sonarSecurityAgentciResolvepathwithinroot, "root is required", nil)
	}

	safeName, err := ValidatePathElement(input)
	if err != nil {
		return "", "", core.E(sonarSecurityAgentciResolvepathwithinroot, "invalid path element", err)
	}

	absRoot := absoluteLocalPath(root)
	resolved := cleanLocalPath(core.Join(localPathSeparator(), absRoot, safeName))
	cleanRoot := cleanLocalPath(absRoot)
	if cleanRoot == localPathSeparator() {
		if !core.HasPrefix(resolved, cleanRoot) {
			return "", "", core.E(sonarSecurityAgentciResolvepathwithinroot, "resolved path escaped root", nil)
		}
		return safeName, resolved, nil
	}

	rootPrefix := core.Concat(cleanRoot, localPathSeparator())
	if resolved != cleanRoot && !core.HasPrefix(resolved, rootPrefix) {
		return "", "", core.E(sonarSecurityAgentciResolvepathwithinroot, "resolved path escaped root", nil)
	}

	return safeName, resolved, nil
}

// ValidateRemoteDir validates a remote directory path used over SSH.
// Usage: ValidateRemoteDir(...)
func ValidateRemoteDir(dir string) (string, error) {
	if core.Trim(dir) == "" {
		return "", core.E(sonarSecurityAgentciValidateremotedir, "directory is required", nil)
	}
	if core.Contains(dir, `\`) {
		return "", core.E(sonarSecurityAgentciValidateremotedir, "backslashes are not allowed", nil)
	}

	switch dir {
	case "/", "~":
		return dir, nil
	}

	prefix, rest := splitRemoteDirPrefix(dir)
	if err := validateRemoteDirSegments(rest); err != nil {
		return "", err
	}

	if rest == "" || rest == "." {
		return remoteDirRoot(prefix), nil
	}

	return cleanRemotePath(dir), nil
}

func splitRemoteDirPrefix(dir string) (prefix, rest string) {
	if core.HasPrefix(dir, "~/") {
		return "~/", core.TrimPrefix(dir, "~/")
	}
	if core.HasPrefix(dir, "/") {
		return "/", core.TrimPrefix(dir, "/")
	}
	return "", dir
}

func validateRemoteDirSegments(rest string) error {
	for _, part := range core.Split(rest, "/") {
		if part == "" {
			continue
		}
		if part == "." || part == ".." {
			return core.E(sonarSecurityAgentciValidateremotedir, "directory escaped root", nil)
		}
		if _, err := ValidatePathElement(part); err != nil {
			return core.E(sonarSecurityAgentciValidateremotedir, "invalid directory segment", err)
		}
	}
	return nil
}

func remoteDirRoot(prefix string) string {
	if prefix == "~/" {
		return "~"
	}
	return prefix
}

// JoinRemotePath joins validated remote path elements using forward slashes.
// Usage: JoinRemotePath(...)
func JoinRemotePath(base string, parts ...string) (string, error) {
	safeBase, err := ValidateRemoteDir(base)
	if err != nil {
		return "", core.E("agentci.JoinRemotePath", "invalid base directory", err)
	}
	if len(parts) == 0 {
		return safeBase, nil
	}

	cleanParts := make([]string, 0, len(parts))
	for _, part := range parts {
		safePart, partErr := ValidatePathElement(part)
		if partErr != nil {
			return "", core.E("agentci.JoinRemotePath", "invalid path element", partErr)
		}
		cleanParts = append(cleanParts, safePart)
	}

	if safeBase == "~" {
		return joinRemotePath("~", joinRemotePath(cleanParts...)), nil
	}
	if core.HasPrefix(safeBase, "~/") {
		return core.Concat("~/", joinRemotePath(core.TrimPrefix(safeBase, "~/"), joinRemotePath(cleanParts...))), nil
	}
	return joinRemotePath(append([]string{safeBase}, cleanParts...)...), nil
}

// EscapeShellArg wraps a string in single quotes for safe remote shell insertion.
// Prefer exec.Command arguments over constructing shell strings where possible.
// Usage: EscapeShellArg(...)
func EscapeShellArg(arg string) string {
	return core.Concat("'", core.Replace(arg, "'", "'\\''"), "'")
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
	return core.Concat(token[:4], "****", token[len(token)-4:])
}

func localPathSeparator() string {
	ds := core.Env("DS")
	if ds == "" {
		return "/"
	}
	return ds
}

func cleanLocalPath(p string) string {
	return core.CleanPath(p, localPathSeparator())
}

func absoluteLocalPath(p string) string {
	if core.PathIsAbs(p) {
		return cleanLocalPath(p)
	}

	cwd := core.Env("DIR_CWD")
	if cwd == "" {
		cwd = "."
	}
	return cleanLocalPath(core.Join(localPathSeparator(), cwd, p))
}

func cleanRemotePath(p string) string {
	return core.CleanPath(p, "/")
}

func joinRemotePath(parts ...string) string {
	return cleanRemotePath(core.JoinPath(parts...))
}
