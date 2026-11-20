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

// History is what a session's transcript says about its folder: the commit it started from, the
// agent's tool calls and the removals made from the info bar, all oldest first.
type History struct {
	Base     string
	Calls    []ToolCall
	Removals []Removal
}

// Journal reads the history of a session and records removals made from the info bar.
type Journal interface {
	History(ctx context.Context, session string) (History, error)
	RecordRemoval(ctx context.Context, owner, session string, removal Removal) error
}

// Summary is the singleton git overview of a session (AIP-156): what the session changed since it
// started and, in order, the branches, worktrees and files its agent created.
type Summary struct {
	Repository bool            `json:"repository"`
	Base       string          `json:"base,omitempty"`
	Changes    ChangeSummary   `json:"changes"`
	Activity   []ActivityEntry `json:"activity"`
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

func (r repo) summary(ctx context.Context, h History) (Summary, error) {
	s := Summary{Activity: []ActivityEntry{}, Changes: ChangeSummary{Files: []FileStat{}}}
	ok, err := r.isRepository(ctx)
	if err != nil {
		return s, err
	}
	s.Repository = ok
	if ok {
		s.Base = r.validBase(ctx, h.Base)
		if s.Changes, err = r.changesSince(ctx, s.Base); err != nil {
			return s, err
		}
	}
	s.Activity, err = r.activity(ctx, s.Base, h, ok)
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
