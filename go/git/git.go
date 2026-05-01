// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"context"
	"iter"
	"strconv"

	core "dappco.re/go"
	process "dappco.re/go/process"
)

type GitError struct {
	Err    error
	Stderr string
}

func (e *GitError) Error() string {
	if e == nil {
		return ""
	}
	if core.Trim(e.Stderr) != "" {
		return core.Trim(e.Stderr)
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "git error"
}

func (e *GitError) Unwrap() error  /* v090-result-boundary */ {
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
	msg := core.Lower(err.Error())
	return core.Contains(msg, "non-fast-forward") || core.Contains(msg, "fetch first") || core.Contains(msg, "rejected")
}

func runGit(ctx context.Context, path string, args ...string) ([]byte, []byte, error)  /* v090-result-boundary */ {
	if ctx == nil {
		ctx = context.Background()
	}
	r := process.RunWithOptions(ctx, process.RunOptions{
		Command: "git",
		Args:    append([]string{"-C", path}, args...),
	})
	output, _ := r.Value.(string)
	if !r.OK {
		return []byte(output), []byte(output), &GitError{Err: r.Value.(error), Stderr: output}
	}
	return []byte(output), nil, nil
}

func Pull(ctx context.Context, path string) error  /* v090-result-boundary */ {
	if err := ensurePath(path); err != nil {
		return err
	}
	_, _, err := runGit(ctx, path, "pull", "--ff-only")
	return err
}

func Push(ctx context.Context, path string) error  /* v090-result-boundary */ {
	if err := ensurePath(path); err != nil {
		return err
	}
	_, _, err := runGit(ctx, path, "push")
	return err
}

// Sync fetches the default Forge remote and hard-resets the working tree to
// match the requested branch.
func Sync(ctx context.Context, path string) error  /* v090-result-boundary */ {
	return SyncWithRemote(ctx, path, "origin", "dev")
}

// SyncWithRemote fetches a branch from the given remote and resets the local
// working tree to match it.
func SyncWithRemote(ctx context.Context, path, remote, branch string) error  /* v090-result-boundary */ {
	if err := ensurePath(path); err != nil {
		return err
	}
	if core.Trim(remote) == "" {
		remote = "origin"
	}
	if core.Trim(branch) == "" {
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
			name = core.PathBase(path)
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
			name = core.PathBase(path)
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

func PullMultiple(ctx context.Context, paths []string, names map[string]string) []PushResult {
	var out []PushResult
	for _, path := range paths {
		name := names[path]
		if name == "" {
			name = core.PathBase(path)
		}
		r := PushResult{Name: name, Path: path}
		if err := Pull(ctx, path); err != nil {
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
				name = core.PathBase(path)
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

func PullMultipleIter(ctx context.Context, paths []string, names map[string]string) iter.Seq[PushResult] {
	return func(yield func(PushResult) bool) {
		for _, path := range paths {
			name := names[path]
			if name == "" {
				name = core.PathBase(path)
			}
			r := PushResult{Name: name, Path: path}
			if err := Pull(ctx, path); err != nil {
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
			name = core.PathBase(path)
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
				name = core.PathBase(path)
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
	for _, line := range core.Split(output, "\n") {
		line = core.Trim(line)
		switch {
		case core.HasPrefix(line, "## "):
			parseBranchLine(core.TrimPrefix(line, "## "), st)
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
	parts := core.Split(line, "...")
	if len(parts) > 0 {
		st.Branch = core.Trim(parts[0])
	}
	if len(parts) < 2 {
		return
	}
	rest := parts[1]
	if rangeParts := core.SplitN(rest, "[", 2); len(rangeParts) == 2 {
		ranges := core.TrimSuffix(rangeParts[1], "]")
		for _, part := range core.Split(ranges, ",") {
			part = core.Trim(part)
			switch {
			case core.HasPrefix(part, "ahead "):
				st.Ahead = parseCount(core.TrimPrefix(part, "ahead "))
			case core.HasPrefix(part, "behind "):
				st.Behind = parseCount(core.TrimPrefix(part, "behind "))
			}
		}
	}
}

func ensurePath(path string) error  /* v090-result-boundary */ {
	if core.Trim(path) == "" {
		return core.E("", "git path is required", nil)
	}
	return nil
}

func parseCount(raw string) int {
	n, _ := strconv.Atoi(core.Trim(raw))
	return n
}
