// SPDX-License-Identifier: EUPL-1.2

// Package git provides utilities for git operations across multiple repositories.
package git

import (
	"bytes"
	"context"
	fmt "dappco.re/go/core/scm/internal/ax/fmtx"
	os "dappco.re/go/core/scm/internal/ax/osx"
	strings "dappco.re/go/core/scm/internal/ax/stringsx"
	"errors"
	exec "golang.org/x/sys/execabs"
	"io"
	"iter"
	stdos "os"
	"slices"
	"strconv"
	"sync"
)

// RepoStatus represents the git status of a single repository.
type RepoStatus struct {
	Name      string
	Path      string
	Modified  int
	Untracked int
	Staged    int
	Ahead     int
	Behind    int
	Branch    string
	Error     error
}

// IsDirty returns true if there are uncommitted changes.
// Usage: IsDirty(...)
func (s *RepoStatus) IsDirty() bool {
	return s.Modified > 0 || s.Untracked > 0 || s.Staged > 0
}

// HasUnpushed returns true if there are commits to push.
// Usage: HasUnpushed(...)
func (s *RepoStatus) HasUnpushed() bool {
	return s.Ahead > 0
}

// HasUnpulled returns true if there are commits to pull.
// Usage: HasUnpulled(...)
func (s *RepoStatus) HasUnpulled() bool {
	return s.Behind > 0
}

// StatusOptions configures the status check.
type StatusOptions struct {
	// Paths is a list of repo paths to check
	Paths []string
	// Names maps paths to display names
	Names map[string]string
}

// Status checks git status for multiple repositories in parallel.
// Usage: Status(...)
func Status(ctx context.Context, opts StatusOptions) []RepoStatus {
	var wg sync.WaitGroup
	results := make([]RepoStatus, len(opts.Paths))

	for i, path := range opts.Paths {
		wg.Add(1)
		go func(idx int, repoPath string) {
			defer wg.Done()
			name := opts.Names[repoPath]
			if name == "" {
				name = repoPath
			}
			results[idx] = getStatus(ctx, repoPath, name)
		}(i, path)
	}

	wg.Wait()
	return results
}

// StatusIter returns an iterator over git status for multiple repositories.
// Usage: StatusIter(...)
func StatusIter(ctx context.Context, opts StatusOptions) iter.Seq[RepoStatus] {
	return func(yield func(RepoStatus) bool) {
		results := Status(ctx, opts)
		for _, r := range results {
			if !yield(r) {
				return
			}
		}
	}
}

// getStatus gets the git status for a single repository.
func getStatus(ctx context.Context, path, name string) RepoStatus {
	status := RepoStatus{
		Name: name,
		Path: path,
	}

	// Get current branch
	branch, err := gitCommand(ctx, path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		status.Error = err
		return status
	}
	status.Branch = strings.TrimSpace(branch)

	// Get porcelain status
	porcelain, err := gitCommand(ctx, path, "status", "--porcelain")
	if err != nil {
		status.Error = err
		return status
	}

	// Parse status output
	for line := range strings.SplitSeq(porcelain, "\n") {
		if len(line) < 2 {
			continue
		}
		x, y := line[0], line[1]

		// Untracked
		if x == '?' && y == '?' {
			status.Untracked++
			continue
		}

		// Staged (index has changes)
		if x == 'A' || x == 'D' || x == 'R' || x == 'M' {
			status.Staged++
		}

		// Modified in working tree
		if y == 'M' || y == 'D' {
			status.Modified++
		}
	}

	// Get ahead/behind counts
	ahead, behind := getAheadBehind(ctx, path)
	status.Ahead = ahead
	status.Behind = behind

	return status
}

// getAheadBehind returns the number of commits ahead and behind upstream.
func getAheadBehind(ctx context.Context, path string) (ahead, behind int) {
	// Try to get ahead count
	aheadStr, err := gitCommand(ctx, path, "rev-list", "--count", "@{u}..HEAD")
	if err == nil {
		ahead, _ = strconv.Atoi(strings.TrimSpace(aheadStr))
	}

	// Try to get behind count
	behindStr, err := gitCommand(ctx, path, "rev-list", "--count", "HEAD..@{u}")
	if err == nil {
		behind, _ = strconv.Atoi(strings.TrimSpace(behindStr))
	}

	return ahead, behind
}

// Push pushes commits for a single repository.
// Uses interactive mode to support SSH passphrase prompts.
// Usage: Push(...)
func Push(ctx context.Context, path string) error {
	return gitInteractive(ctx, path, "push")
}

// Pull pulls changes for a single repository.
// Uses interactive mode to support SSH passphrase prompts.
// Usage: Pull(...)
func Pull(ctx context.Context, path string) error {
	return gitInteractive(ctx, path, "pull", "--rebase")
}

// PushWithSSH pushes commits using an explicit SSH configuration.
// Usage: PushWithSSH(...)
func PushWithSSH(ctx context.Context, path string, opts SSHOptions) error {
	return gitInteractiveEnv(ctx, path, gitEnvWithSSH(opts), "push")
}

// PullWithSSH pulls changes using an explicit SSH configuration.
// Usage: PullWithSSH(...)
func PullWithSSH(ctx context.Context, path string, opts SSHOptions) error {
	return gitInteractiveEnv(ctx, path, gitEnvWithSSH(opts), "pull", "--rebase")
}

// CreateBranch creates and checks out a new branch from the current HEAD or an
// optional start point.
// Usage: CreateBranch(...)
func CreateBranch(ctx context.Context, path, branch, startPoint string) error {
	args := []string{"checkout", "-b", branch}
	if strings.TrimSpace(startPoint) != "" {
		args = append(args, startPoint)
	}
	return gitInteractive(ctx, path, args...)
}

// Clone clones a repository into dest. When ref is non-empty it is checked out
// during clone, which supports both branches and tags.
// Usage: Clone(...)
func Clone(ctx context.Context, repo, dest, ref string) error {
	args := []string{"clone", "--depth=1"}
	if ref != "" {
		args = append(args, "--branch", ref)
	}
	args = append(args, repo, dest)
	return gitInteractive(ctx, "", args...)
}

// CloneWithSSH clones a repository using an explicit SSH configuration.
// Usage: CloneWithSSH(...)
func CloneWithSSH(ctx context.Context, repo, dest, ref string, opts SSHOptions) error {
	args := []string{"clone", "--depth=1"}
	if ref != "" {
		args = append(args, "--branch", ref)
	}
	args = append(args, repo, dest)
	return gitInteractiveEnv(ctx, "", gitEnvWithSSH(opts), args...)
}

// Fetch fetches refs from the given remote.
// When branch is non-empty, it is fetched explicitly from origin.
// Usage: Fetch(...)
func Fetch(ctx context.Context, path, branch string) error {
	args := []string{"fetch", "origin"}
	if branch != "" {
		args = append(args, branch)
	}
	return gitInteractive(ctx, path, args...)
}

// FetchWithSSH fetches refs using an explicit SSH configuration.
// Usage: FetchWithSSH(...)
func FetchWithSSH(ctx context.Context, path, branch string, opts SSHOptions) error {
	args := []string{"fetch", "origin"}
	if branch != "" {
		args = append(args, branch)
	}
	return gitInteractiveEnv(ctx, path, gitEnvWithSSH(opts), args...)
}

// ResetHard resets the working tree to the given ref.
// Usage: ResetHard(...)
func ResetHard(ctx context.Context, path, ref string) error {
	_, err := gitCommand(ctx, path, "reset", "--hard", ref)
	return err
}

// FetchTags fetches all tags from origin.
// Usage: FetchTags(...)
func FetchTags(ctx context.Context, path string) error {
	return gitInteractive(ctx, path, "fetch", "--tags", "origin")
}

// Checkout switches the repository to the given ref.
// Usage: Checkout(...)
func Checkout(ctx context.Context, path, ref string) error {
	return gitInteractive(ctx, path, "checkout", ref)
}

// SwitchBranch switches the repository to the named branch.
// Usage: SwitchBranch(...)
func SwitchBranch(ctx context.Context, path, branch string) error {
	return Checkout(ctx, path, branch)
}

// Commit creates a git commit with the supplied message.
// Usage: Commit(...)
func Commit(ctx context.Context, path, message string) error {
	_, err := gitCommand(ctx, path, "commit", "-m", message)
	return err
}

// AddAll stages all tracked and untracked changes.
// Usage: AddAll(...)
func AddAll(ctx context.Context, path string) error {
	_, err := gitCommand(ctx, path, "add", "-A")
	return err
}

// CurrentTag returns the tag pointing at HEAD, or empty if HEAD is not tagged.
// Usage: CurrentTag(...)
func CurrentTag(ctx context.Context, path string) (string, error) {
	out, err := gitCommand(ctx, path, "describe", "--tags", "--exact-match", "HEAD")
	if err != nil {
		if isNoExactTagError(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// ListRemoteTags lists tag names advertised by a remote repository.
// Usage: ListRemoteTags(...)
func ListRemoteTags(ctx context.Context, repo string) ([]string, error) {
	out, err := gitCommand(ctx, "", "ls-remote", "--tags", "--refs", repo)
	if err != nil {
		return nil, err
	}

	var tags []string
	for line := range strings.SplitSeq(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			return nil, &GitError{Err: nil, Stderr: fmt.Sprintf("invalid ls-remote output: %s", line)}
		}

		ref := strings.TrimSpace(parts[1])
		if strings.HasPrefix(ref, "refs/tags/") {
			tags = append(tags, strings.TrimPrefix(ref, "refs/tags/"))
		}
	}

	return tags, nil
}

// VerifyCommitSignature verifies the commit signature at ref.
// Unsigned or invalid commits return (false, nil).
// Usage: VerifyCommitSignature(...)
func VerifyCommitSignature(ctx context.Context, path, ref string) (bool, error) {
	return verifySignature(ctx, path, "verify-commit", ref)
}

// VerifyTagSignature verifies the tag signature for tag.
// Unsigned or invalid tags return (false, nil).
// Usage: VerifyTagSignature(...)
func VerifyTagSignature(ctx context.Context, path, tag string) (bool, error) {
	return verifySignature(ctx, path, "verify-tag", tag)
}

// SSHOptions configures git operations that authenticate over SSH.
type SSHOptions struct {
	KeyPath                      string
	KnownHostsPath               string
	DisableStrictHostKeyChecking bool
}

// SSHCommand renders a safe GIT_SSH_COMMAND string for git child processes.
// Usage: SSHCommand(...)
func SSHCommand(opts SSHOptions) string {
	parts := []string{
		"ssh",
		"-o", "BatchMode=yes",
		"-o", "IdentitiesOnly=yes",
	}

	strict := "yes"
	if opts.DisableStrictHostKeyChecking {
		strict = "no"
	}
	parts = append(parts, "-o", "StrictHostKeyChecking="+strict)

	if keyPath := strings.TrimSpace(opts.KeyPath); keyPath != "" {
		parts = append(parts, "-i", keyPath)
	}

	knownHosts := strings.TrimSpace(opts.KnownHostsPath)
	switch {
	case knownHosts != "":
		parts = append(parts, "-o", "UserKnownHostsFile="+knownHosts)
	case opts.DisableStrictHostKeyChecking:
		parts = append(parts, "-o", "UserKnownHostsFile=/dev/null")
	}

	return shellJoin(parts)
}

// ConfigureSSH sets GIT_SSH_COMMAND on cmd using the supplied SSH options.
// Usage: ConfigureSSH(...)
func ConfigureSSH(cmd *exec.Cmd, opts SSHOptions) {
	if cmd == nil {
		return
	}

	cmd.Env = gitEnvWithSSH(opts)
}

// IsNonFastForward checks if an error is a non-fast-forward rejection.
// Usage: IsNonFastForward(...)
func IsNonFastForward(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "non-fast-forward") ||
		strings.Contains(msg, "fetch first") ||
		strings.Contains(msg, "tip of your current branch is behind")
}

// gitInteractive runs a git command with terminal attached for user interaction.
func gitInteractive(ctx context.Context, dir string, args ...string) error {
	return gitInteractiveEnv(ctx, dir, nil, args...)
}

func gitInteractiveEnv(ctx context.Context, dir string, env []string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	if len(env) > 0 {
		cmd.Env = env
	}

	// Connect to terminal for SSH passphrase prompts
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	// Capture stderr for error reporting while also showing it
	var stderr bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return &GitError{Err: err, Stderr: stderr.String()}
		}
		return err
	}

	return nil
}

// PushResult represents the result of a push operation.
type PushResult struct {
	Name    string
	Path    string
	Success bool
	Error   error
}

// PushMultiple pushes multiple repositories sequentially.
// Sequential because SSH passphrase prompts need user interaction.
// Usage: PushMultiple(...)
func PushMultiple(ctx context.Context, paths []string, names map[string]string) []PushResult {
	return slices.Collect(PushMultipleIter(ctx, paths, names))
}

// PushMultipleIter returns an iterator that pushes repositories sequentially and yields results.
// Usage: PushMultipleIter(...)
func PushMultipleIter(ctx context.Context, paths []string, names map[string]string) iter.Seq[PushResult] {
	return func(yield func(PushResult) bool) {
		for _, path := range paths {
			name := names[path]
			if name == "" {
				name = path
			}

			result := PushResult{
				Name: name,
				Path: path,
			}

			err := Push(ctx, path)
			if err != nil {
				result.Error = err
			} else {
				result.Success = true
			}

			if !yield(result) {
				return
			}
		}
	}
}

// gitCommand runs a git command and returns stdout.
func gitCommand(ctx context.Context, dir string, args ...string) (string, error) {
	return gitCommandEnv(ctx, dir, nil, args...)
}

func gitCommandEnv(ctx context.Context, dir string, env []string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	if len(env) > 0 {
		cmd.Env = env
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Include stderr in error message for better diagnostics
		if stderr.Len() > 0 {
			return "", &GitError{Err: err, Stderr: stderr.String()}
		}
		return "", err
	}

	return stdout.String(), nil
}

func verifySignature(ctx context.Context, path string, args ...string) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = path

	output, err := cmd.CombinedOutput()
	if err == nil {
		return true, nil
	}

	msg := strings.TrimSpace(string(output))
	if signatureVerificationFailure(msg) {
		return false, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}
	if msg != "" {
		return false, &GitError{Err: err, Stderr: msg}
	}

	return false, err
}

func signatureVerificationFailure(msg string) bool {
	if msg == "" {
		return false
	}

	msg = strings.ToLower(msg)
	for _, needle := range []string{
		"does not have a gpg signature",
		"does not have a good signature",
		"bad signature",
		"no signature",
		"can't check signature",
		"cannot check signature",
		"missing object referenced by",
		"cannot verify a non-tag object",
	} {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}

func isNoExactTagError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	for _, needle := range []string{
		"no names found, cannot describe anything",
		"no tag exactly matches",
	} {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}

func gitEnvWithSSH(opts SSHOptions) []string {
	return append(stdos.Environ(), "GIT_SSH_COMMAND="+SSHCommand(opts))
}

func shellJoin(parts []string) string {
	quoted := make([]string, 0, len(parts))
	for _, part := range parts {
		quoted = append(quoted, shellQuote(part))
	}
	return strings.Join(quoted, " ")
}

func shellQuote(part string) string {
	if part == "" {
		return "''"
	}

	safe := true
	for _, r := range part {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			continue
		}
		switch r {
		case '/', '.', '-', '_', ':', '=', '@':
			continue
		default:
			safe = false
		}
		if !safe {
			break
		}
	}
	if safe {
		return part
	}

	return "'" + strings.ReplaceAll(part, "'", `'\''`) + "'"
}

// GitError wraps a git command error with stderr output.
type GitError struct {
	Err    error
	Stderr string
}

// Error returns the git error message, preferring stderr output.
// Usage: Error(...)
func (e *GitError) Error() string {
	// Return just the stderr message, trimmed
	msg := strings.TrimSpace(e.Stderr)
	if msg != "" {
		return msg
	}
	return e.Err.Error()
}

// Unwrap returns the underlying error for error chain inspection.
// Usage: Unwrap(...)
func (e *GitError) Unwrap() error {
	return e.Err
}
