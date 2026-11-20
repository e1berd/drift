package gitrepo

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type worktreeEntry struct {
	path   string
	branch string
}

func (r repo) worktreeEntries(ctx context.Context) ([]worktreeEntry, error) {
	out, err := r.git(ctx, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		return nil, err
	}
	var entries []worktreeEntry
	for field := range strings.SplitSeq(string(out), "\x00") {
		if value, ok := strings.CutPrefix(field, "worktree "); ok {
			entries = append(entries, worktreeEntry{path: canonical(value)})
		} else if value, ok := strings.CutPrefix(field, "branch "); ok && len(entries) > 0 {
			entries[len(entries)-1].branch = strings.TrimPrefix(value, "refs/heads/")
		}
	}
	return entries, nil
}

func canonical(p string) string {
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		return resolved
	}
	return filepath.Clean(p)
}

func (r repo) matchWorktree(entries []worktreeEntry, ref CreatedRef) (worktreeEntry, bool) {
	if !strings.ContainsRune(ref.Worktree, '/') && !strings.HasPrefix(ref.Worktree, ".") {
		for _, entry := range entries {
			if filepath.Base(entry.path) == ref.Worktree {
				return entry, true
			}
		}
	}
	want := canonical(r.absolute(ref.Dir, ref.Worktree))
	for _, entry := range entries {
		if entry.path == want {
			return entry, true
		}
	}
	return worktreeEntry{}, false
}

func (r repo) absolute(dir, target string) string {
	if home, err := os.UserHomeDir(); err == nil {
		dir = expandHome(dir, home)
		target = expandHome(target, home)
	}
	if filepath.IsAbs(target) {
		return target
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(r.dir, dir)
	}
	return filepath.Join(dir, target)
}

func expandHome(p, home string) string {
	if p == "~" {
		return home
	}
	if rest, ok := strings.CutPrefix(p, "~/"); ok {
		return filepath.Join(home, rest)
	}
	return p
}

func (r repo) displayPath(p string) string {
	if rel, err := filepath.Rel(filepath.Dir(r.dir), p); err == nil && !strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(rel)
	}
	if home, err := os.UserHomeDir(); err == nil {
		if rel, err := filepath.Rel(home, p); err == nil && !strings.HasPrefix(rel, "..") {
			return "~/" + filepath.ToSlash(rel)
		}
	}
	return p
}

func worktreeChanges(ctx context.Context, dir string) int {
	tree, err := open(dir)
	if err != nil {
		return 0
	}
	out, err := tree.git(ctx, "status", "--porcelain", "-z", "--untracked-files=all")
	if err != nil {
		return 0
	}
	count := 0
	for record := range strings.SplitSeq(string(out), "\x00") {
		if len(record) > 3 && record[2] == ' ' {
			count++
		}
	}
	return count
}

func (r repo) countCommits(ctx context.Context, from, to string) int {
	if from == "" {
		return 0
	}
	out, err := r.git(ctx, "rev-list", "--count", from+".."+to, "--")
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(strings.TrimSpace(string(out)))
	return n
}
