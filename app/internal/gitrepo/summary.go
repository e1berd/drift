package gitrepo

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	topChangedFiles    = 5
	maxUntrackedFiles  = 2000
	maxCountedFileSize = 1 << 20
)

// History returns the commit the session started from and the agent's tool calls, oldest first.
type History func(ctx context.Context, session string) (base string, calls []ToolCall, err error)

// Summary is the singleton git overview of a session (AIP-156): what the session changed since it
// started and the branches and worktrees its agent created.
type Summary struct {
	Repository bool            `json:"repository"`
	Base       string          `json:"base,omitempty"`
	Changes    ChangeSummary   `json:"changes"`
	Branches   []SessionBranch `json:"branches"`
	Worktrees  []SessionTree   `json:"worktrees"`
}

// ChangeSummary counts changed lines from the session's base commit to the working tree.
type ChangeSummary struct {
	Additions int        `json:"additions"`
	Deletions int        `json:"deletions"`
	FileCount int        `json:"fileCount"`
	Files     []FileStat `json:"files"`
}

// FileStat is the changed line count of one file.
type FileStat struct {
	Path      string `json:"path"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

// SessionBranch is a branch the session created that still exists.
type SessionBranch struct {
	Name      string `json:"name"`
	MessageID string `json:"messageId"`
	Current   bool   `json:"current"`
	Upstream  string `json:"upstream,omitempty"`
	Commits   int    `json:"commits"`
	Unpushed  int    `json:"unpushed"`
}

// SessionTree is a worktree the session created that still exists.
type SessionTree struct {
	Path      string `json:"path"`
	Branch    string `json:"branch,omitempty"`
	MessageID string `json:"messageId"`
	Changes   int    `json:"changes"`
}

type worktreeEntry struct {
	path   string
	branch string
}

func (r repo) summary(ctx context.Context, base string, calls []ToolCall) (Summary, error) {
	s := Summary{Branches: []SessionBranch{}, Worktrees: []SessionTree{}, Changes: ChangeSummary{Files: []FileStat{}}}
	ok, err := r.isRepository(ctx)
	if err != nil || !ok {
		return s, err
	}
	s.Repository = true
	s.Base = r.validBase(ctx, base)
	if s.Changes, err = r.changesSince(ctx, s.Base); err != nil {
		return s, err
	}
	refs := CreatedRefs(calls)
	if s.Worktrees, err = r.sessionWorktrees(ctx, refs); err != nil {
		return s, err
	}
	s.Branches, err = r.sessionBranches(ctx, s.Base, refs, s.Worktrees)
	if len(s.Base) > 12 {
		s.Base = s.Base[:12]
	}
	return s, err
}

func (r repo) validBase(ctx context.Context, base string) string {
	if base != "" {
		if _, err := r.git(ctx, "cat-file", "-e", base+"^{commit}"); err == nil {
			return base
		}
	}
	if !r.hasHead(ctx) {
		return ""
	}
	out, err := r.git(ctx, "rev-parse", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (r repo) changesSince(ctx context.Context, base string) (ChangeSummary, error) {
	files := map[string]*FileStat{}
	if base != "" {
		out, err := r.git(ctx, "diff", "--numstat", "-z", "--no-renames", base, "--")
		if err != nil {
			return ChangeSummary{Files: []FileStat{}}, err
		}
		for _, stat := range parseNumstat(out) {
			files[stat.Path] = &stat
		}
	}
	untracked, err := r.git(ctx, "ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return ChangeSummary{Files: []FileStat{}}, err
	}
	for i, name := range strings.Split(strings.TrimRight(string(untracked), "\x00"), "\x00") {
		if name == "" || i >= maxUntrackedFiles {
			continue
		}
		files[name] = &FileStat{Path: name, Additions: r.lineCount(name)}
	}
	return summarize(files), nil
}

func summarize(files map[string]*FileStat) ChangeSummary {
	list := make([]FileStat, 0, len(files))
	c := ChangeSummary{FileCount: len(files)}
	for _, stat := range files {
		c.Additions += stat.Additions
		c.Deletions += stat.Deletions
		list = append(list, *stat)
	}
	sort.Slice(list, func(i, j int) bool {
		a, b := list[i].Additions+list[i].Deletions, list[j].Additions+list[j].Deletions
		if a != b {
			return a > b
		}
		return list[i].Path < list[j].Path
	})
	c.Files = list[:min(len(list), topChangedFiles)]
	return c
}

func parseNumstat(out []byte) []FileStat {
	var stats []FileStat
	for record := range bytes.SplitSeq(out, []byte{0}) {
		fields := strings.SplitN(string(record), "\t", 3)
		if len(fields) != 3 || fields[2] == "" {
			continue
		}
		added, _ := strconv.Atoi(fields[0])
		deleted, _ := strconv.Atoi(fields[1])
		stats = append(stats, FileStat{Path: fields[2], Additions: added, Deletions: deleted})
	}
	return stats
}

func (r repo) lineCount(name string) int {
	full := filepath.Join(r.dir, filepath.FromSlash(name))
	info, err := os.Stat(full)
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxCountedFileSize {
		return 0
	}
	data, err := os.ReadFile(full)
	if err != nil || bytes.IndexByte(data, 0) >= 0 || len(data) == 0 {
		return 0
	}
	lines := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		lines++
	}
	return lines
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

func (r repo) sessionWorktrees(ctx context.Context, refs []CreatedRef) ([]SessionTree, error) {
	trees := []SessionTree{}
	entries, err := r.worktreeEntries(ctx)
	if err != nil || len(entries) < 2 {
		return trees, err
	}
	seen := map[string]bool{}
	for _, ref := range refs {
		if ref.Worktree == "" {
			continue
		}
		entry, ok := r.matchWorktree(entries[1:], ref)
		if !ok || seen[entry.path] {
			continue
		}
		seen[entry.path] = true
		trees = append(trees, SessionTree{
			Path: r.displayPath(entry.path), Branch: entry.branch, MessageID: ref.MessageID,
			Changes: worktreeChanges(ctx, entry.path),
		})
	}
	return trees, nil
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

func (r repo) sessionBranches(ctx context.Context, base string, refs []CreatedRef, trees []SessionTree) ([]SessionBranch, error) {
	list := []SessionBranch{}
	existing, err := r.branches(ctx)
	if err != nil {
		return list, err
	}
	byName := make(map[string]Branch, len(existing))
	for _, b := range existing {
		byName[b.Name] = b
	}
	seen := map[string]bool{}
	add := func(name, messageID string) {
		b, ok := byName[name]
		if !ok || seen[name] {
			return
		}
		seen[name] = true
		list = append(list, SessionBranch{
			Name: name, MessageID: messageID, Current: b.Current, Upstream: b.Upstream,
			Commits: r.countCommits(ctx, base, name), Unpushed: r.countCommits(ctx, b.Upstream, name),
		})
	}
	for _, ref := range refs {
		if ref.Branch != "" {
			add(ref.Branch, ref.MessageID)
		}
	}
	for _, tree := range trees {
		add(tree.Branch, tree.MessageID)
	}
	return list, nil
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

// HeadCommit is the full hash HEAD points to in folder, or "" when it is not a repository with commits.
func HeadCommit(ctx context.Context, folder string) string {
	r, err := open(folder)
	if err != nil {
		return ""
	}
	if ok, err := r.isRepository(ctx); err != nil || !ok {
		return ""
	}
	return r.validBase(ctx, "")
}
