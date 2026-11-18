package gitrepo

import (
	"bytes"
	"context"
	"strconv"
	"strings"
)

const maxStatusEntries = 2000

// Change kinds of a file, as git status reports them per side (index or worktree).
const (
	Modified    = "modified"
	Added       = "added"
	Deleted     = "deleted"
	Renamed     = "renamed"
	Copied      = "copied"
	TypeChanged = "typeChanged"
	Conflicted  = "conflicted"
	Untracked   = "untracked"
)

var changeKinds = map[byte]string{
	'M': Modified, 'A': Added, 'D': Deleted, 'R': Renamed, 'C': Copied, 'T': TypeChanged,
}

// FileChange is one changed path on one side: staged (index vs HEAD) or not (worktree vs index).
type FileChange struct {
	Path     string `json:"path"`
	OrigPath string `json:"origPath,omitempty"`
	Status   string `json:"status"`
}

// Status is the singleton git state of a session folder (AIP-156).
type Status struct {
	Repository bool         `json:"repository"`
	Branch     string       `json:"branch,omitempty"`
	Detached   bool         `json:"detached"`
	Head       string       `json:"head,omitempty"`
	Upstream   string       `json:"upstream,omitempty"`
	Ahead      int          `json:"ahead"`
	Behind     int          `json:"behind"`
	Staged     []FileChange `json:"staged"`
	Unstaged   []FileChange `json:"unstaged"`
	Truncated  bool         `json:"truncated,omitempty"`
}

func (r repo) status(ctx context.Context) (Status, error) {
	empty := Status{Staged: []FileChange{}, Unstaged: []FileChange{}}
	ok, err := r.isRepository(ctx)
	if err != nil || !ok {
		return empty, err
	}
	out, err := r.git(ctx, "status", "--porcelain=v2", "--branch", "-z", "--untracked-files=all")
	if err != nil {
		return empty, err
	}
	return ParseStatus(out), nil
}

// ParseStatus reads the output of `git status --porcelain=v2 --branch -z`.
func ParseStatus(out []byte) Status {
	s := Status{Repository: true, Staged: []FileChange{}, Unstaged: []FileChange{}}
	records := bytes.Split(out, []byte{0})
	for i := 0; i < len(records); i++ {
		rec := string(records[i])
		if rec == "" {
			continue
		}
		switch rec[0] {
		case '#':
			s.readHeader(rec)
		case '1':
			if f := strings.SplitN(rec, " ", 9); len(f) == 9 {
				s.add(f[1], f[8], "")
			}
		case '2':
			if f := strings.SplitN(rec, " ", 10); len(f) == 10 && i+1 < len(records) {
				i++
				s.add(f[1], f[9], string(records[i]))
			}
		case 'u':
			if f := strings.SplitN(rec, " ", 11); len(f) == 11 {
				s.push(&s.Unstaged, FileChange{Path: f[10], Status: Conflicted})
			}
		case '?':
			s.push(&s.Unstaged, FileChange{Path: strings.TrimPrefix(rec, "? "), Status: Untracked})
		}
	}
	return s
}

func (s *Status) readHeader(rec string) {
	name, value, _ := strings.Cut(strings.TrimPrefix(rec, "# "), " ")
	switch name {
	case "branch.oid":
		if value != "(initial)" {
			s.Head = value[:min(len(value), 12)]
		}
	case "branch.head":
		if value == "(detached)" {
			s.Detached = true
		} else {
			s.Branch = value
		}
	case "branch.upstream":
		s.Upstream = value
	case "branch.ab":
		ahead, behind, _ := strings.Cut(value, " ")
		s.Ahead, _ = strconv.Atoi(strings.TrimPrefix(ahead, "+"))
		s.Behind, _ = strconv.Atoi(strings.TrimPrefix(behind, "-"))
	}
}

func (s *Status) add(xy, path, orig string) {
	if len(xy) != 2 {
		return
	}
	if kind, ok := changeKinds[xy[0]]; ok {
		s.push(&s.Staged, FileChange{Path: path, OrigPath: orig, Status: kind})
	}
	if kind, ok := changeKinds[xy[1]]; ok {
		s.push(&s.Unstaged, FileChange{Path: path, Status: kind})
	}
}

func (s *Status) push(list *[]FileChange, change FileChange) {
	if len(*list) >= maxStatusEntries {
		s.Truncated = true
		return
	}
	*list = append(*list, change)
}
