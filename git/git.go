// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitError struct {
	Err    error
	Stderr string
}

func (e *GitError) Error() string {
	if e == nil {
		return ""
	}
	if strings.TrimSpace(e.Stderr) != "" {
		return strings.TrimSpace(e.Stderr)
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "git error"
}

func (e *GitError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

type PushResult struct {
	Name    string
	Path    string
	Success bool
	Error   error
}

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

type SyncResult struct {
	Name    string
	Path    string
	Success bool
	Error   error
}

type StatusOptions struct {
	Paths []string
	Names map[string]string
}

func (s *RepoStatus) HasUnpulled() bool { return s != nil && s.Behind > 0 }
func (s *RepoStatus) HasUnpushed() bool { return s != nil && s.Ahead > 0 }
func (s *RepoStatus) IsDirty() bool {
	return s != nil && (s.Modified > 0 || s.Untracked > 0 || s.Staged > 0)
}

func IsNonFastForward(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "non-fast-forward") || strings.Contains(msg, "fetch first") || strings.Contains(msg, "rejected")
}

func runGit(ctx context.Context, path string, args ...string) ([]byte, []byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", path}, args...)...)
	stdout, stderr := &strings.Builder{}, &strings.Builder{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		return []byte(stdout.String()), []byte(stderr.String()), &GitError{Err: err, Stderr: stderr.String()}
	}
	return []byte(stdout.String()), []byte(stderr.String()), nil
}

func Pull(ctx context.Context, path string) error {
	_, _, err := runGit(ctx, path, "pull", "--ff-only")
	return err
}

func Push(ctx context.Context, path string) error {
	_, _, err := runGit(ctx, path, "push")
	return err
}

// Sync fetches the default Forge remote and hard-resets the working tree to
// match the requested branch.
func Sync(ctx context.Context, path string) error {
	return SyncWithRemote(ctx, path, "origin", "dev")
}

// SyncWithRemote fetches a branch from the given remote and resets the local
// working tree to match it.
func SyncWithRemote(ctx context.Context, path, remote, branch string) error {
	if err := ensurePath(path); err != nil {
		return err
	}
	if strings.TrimSpace(remote) == "" {
		remote = "origin"
	}
	if strings.TrimSpace(branch) == "" {
		branch = "dev"
	}
	if _, _, err := runGit(ctx, path, "fetch", remote, branch); err != nil {
		return err
	}
	_, _, err := runGit(ctx, path, "reset", "--hard", remote+"/"+branch)
	return err
}

// SyncMultiple synchronizes a set of local clones and returns a per-repo result.
func SyncMultiple(ctx context.Context, paths []string, names map[string]string, remote, branch string) []SyncResult {
	var out []SyncResult
	for _, path := range paths {
		name := names[path]
		if name == "" {
			name = filepath.Base(path)
		}
		r := SyncResult{Name: name, Path: path}
		if err := SyncWithRemote(ctx, path, remote, branch); err != nil {
			r.Error = err
		} else {
			r.Success = true
		}
		out = append(out, r)
	}
	return out
}

func PushMultiple(ctx context.Context, paths []string, names map[string]string) []PushResult {
	var out []PushResult
	for _, path := range paths {
		name := names[path]
		if name == "" {
			name = filepath.Base(path)
		}
		r := PushResult{Name: name, Path: path}
		if err := Push(ctx, path); err != nil {
			r.Error = err
		} else {
			r.Success = true
		}
		out = append(out, r)
	}
	return out
}

func PushMultipleIter(ctx context.Context, paths []string, names map[string]string) iter.Seq[PushResult] {
	return func(yield func(PushResult) bool) {
		for _, path := range paths {
			name := names[path]
			if name == "" {
				name = filepath.Base(path)
			}
			r := PushResult{Name: name, Path: path}
			if err := Push(ctx, path); err != nil {
				r.Error = err
			} else {
				r.Success = true
			}
			if !yield(r) {
				return
			}
		}
	}
}

func Status(ctx context.Context, opts StatusOptions) []RepoStatus {
	var out []RepoStatus
	for _, path := range opts.Paths {
		name := opts.Names[path]
		if name == "" {
			name = filepath.Base(path)
		}
		st := RepoStatus{Name: name, Path: path}
		outText, _, err := runGit(ctx, path, "status", "--porcelain", "--branch")
		if err != nil {
			st.Error = err
			out = append(out, st)
			continue
		}
		parseStatus(string(outText), &st)
		out = append(out, st)
	}
	return out
}

func StatusIter(ctx context.Context, opts StatusOptions) iter.Seq[RepoStatus] {
	return func(yield func(RepoStatus) bool) {
		for _, path := range opts.Paths {
			name := opts.Names[path]
			if name == "" {
				name = filepath.Base(path)
			}
			st := RepoStatus{Name: name, Path: path}
			out, _, err := runGit(ctx, path, "status", "--porcelain", "--branch")
			if err != nil {
				st.Error = err
				if !yield(st) {
					return
				}
				continue
			}
			parseStatus(string(out), &st)
			if !yield(st) {
				return
			}
		}
	}
}

func parseStatus(output string, st *RepoStatus) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "## "):
			parseBranchLine(strings.TrimPrefix(line, "## "), st)
		case len(line) >= 2:
			switch {
			case line[0] == '?' && line[1] == '?':
				st.Untracked++
			case line[0] != ' ' && line[0] != '?':
				st.Staged++
			case line[1] != ' ':
				st.Modified++
			}
		}
	}
}

func parseBranchLine(line string, st *RepoStatus) {
	parts := strings.Split(line, "...")
	if len(parts) > 0 {
		st.Branch = strings.TrimSpace(parts[0])
	}
	if len(parts) < 2 {
		return
	}
	rest := parts[1]
	if idx := strings.Index(rest, "["); idx >= 0 {
		ranges := strings.TrimSuffix(rest[idx+1:], "]")
		for _, part := range strings.Split(ranges, ",") {
			part = strings.TrimSpace(part)
			switch {
			case strings.HasPrefix(part, "ahead "):
				fmt.Sscanf(strings.TrimPrefix(part, "ahead "), "%d", &st.Ahead)
			case strings.HasPrefix(part, "behind "):
				fmt.Sscanf(strings.TrimPrefix(part, "behind "), "%d", &st.Behind)
			}
		}
	}
}

func ensurePath(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("git path is required")
	}
	return nil
}
