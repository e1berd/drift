package gitrepo

import (
	"bytes"
	"context"
	"encoding/base64"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	maxDiffBytes    = 512 << 10
	maxBranches     = 500
	defaultPageSize = 30
	maxPageSize     = 100
)

var commitHash = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)

// Branch is a local branch.
type Branch struct {
	Name     string `json:"name"`
	Current  bool   `json:"current"`
	Upstream string `json:"upstream,omitempty"`
}

// Commit is one entry of the history; AuthorTime follows AIP-142 naming.
type Commit struct {
	Hash        string    `json:"hash"`
	ShortHash   string    `json:"shortHash"`
	Subject     string    `json:"subject"`
	AuthorName  string    `json:"authorName"`
	AuthorEmail string    `json:"authorEmail"`
	AuthorTime  time.Time `json:"authorTime"`
}

// Diff is the unified diff of one path, cut at 512 KiB.
type Diff struct {
	Path      string `json:"path"`
	Staged    bool   `json:"staged"`
	Diff      string `json:"diff"`
	Truncated bool   `json:"truncated,omitempty"`
}

func (r repo) branches(ctx context.Context) ([]Branch, error) {
	out, err := r.git(ctx, "for-each-ref", "--sort=-committerdate", "--count=500",
		"--format=%(refname:short)%00%(HEAD)%00%(upstream:short)", "refs/heads")
	if err != nil {
		return nil, err
	}
	list := make([]Branch, 0, 16)
	for line := range strings.SplitSeq(string(out), "\n") {
		f := strings.Split(line, "\x00")
		if len(f) != 3 || len(list) == maxBranches {
			continue
		}
		list = append(list, Branch{Name: f[0], Current: f[1] == "*", Upstream: f[2]})
	}
	return list, nil
}

type commitPage struct {
	head string
	skip int
}

func (p commitPage) token() string {
	return base64.RawURLEncoding.EncodeToString([]byte(p.head + ":" + strconv.Itoa(p.skip)))
}

func parsePageToken(token string) (commitPage, bool) {
	if token == "" {
		return commitPage{}, true
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return commitPage{}, false
	}
	head, skip, ok := strings.Cut(string(raw), ":")
	n, err := strconv.Atoi(skip)
	if !ok || err != nil || n < 0 || !commitHash.MatchString(head) {
		return commitPage{}, false
	}
	return commitPage{head: head, skip: n}, true
}

func (r repo) commits(ctx context.Context, page commitPage, size int) ([]Commit, string, error) {
	if page.head == "" {
		out, err := r.git(ctx, "rev-parse", "--verify", "--quiet", "HEAD")
		if err != nil {
			return []Commit{}, "", nil
		}
		page.head = strings.TrimSpace(string(out))
	}
	out, err := r.git(ctx, "log", "--no-color", "-z", "--skip="+strconv.Itoa(page.skip),
		"--max-count="+strconv.Itoa(size+1), "--format=%H%x1f%h%x1f%s%x1f%an%x1f%ae%x1f%aI", page.head, "--")
	if err != nil {
		return nil, "", err
	}
	list := make([]Commit, 0, size+1)
	for rec := range bytes.SplitSeq(out, []byte{0}) {
		f := strings.Split(string(rec), "\x1f")
		if len(f) != 6 {
			continue
		}
		at, _ := time.Parse(time.RFC3339, f[5])
		list = append(list, Commit{Hash: f[0], ShortHash: f[1], Subject: f[2], AuthorName: f[3], AuthorEmail: f[4], AuthorTime: at})
	}
	if len(list) > size {
		return list[:size], commitPage{head: page.head, skip: page.skip + size}.token(), nil
	}
	return list, "", nil
}

func (r repo) diff(ctx context.Context, path string, staged bool) (Diff, error) {
	args := []string{"diff", "--no-color", "--no-ext-diff", "--no-textconv"}
	if staged {
		args = append(args, "--cached")
	}
	out := &capped{limit: maxDiffBytes}
	if err := r.run(ctx, command{args: append(args, "--", path), timeout: localTimeout, stdout: out}); err != nil {
		return Diff{}, err
	}
	if out.bytes() == nil && !staged && r.isUntracked(ctx, path) {
		out = &capped{limit: maxDiffBytes}
		err := r.run(ctx, command{
			args:    []string{"diff", "--no-color", "--no-ext-diff", "--no-textconv", "--no-index", "--", "/dev/null", path},
			timeout: localTimeout, stdout: out, okCodes: []int{1},
		})
		if err != nil {
			return Diff{}, err
		}
	}
	return Diff{Path: path, Staged: staged, Diff: string(out.bytes()), Truncated: out.truncated}, nil
}

func (r repo) isUntracked(ctx context.Context, path string) bool {
	out, err := r.git(ctx, "ls-files", "--others", "--exclude-standard", "-z", "--", path)
	return err == nil && string(bytes.TrimSuffix(out, []byte{0})) == path
}

func (r repo) untracked(ctx context.Context, paths []string) (map[string]bool, error) {
	out, err := r.git(ctx, append([]string{"ls-files", "--others", "--exclude-standard", "-z", "--"}, paths...)...)
	if err != nil {
		return nil, err
	}
	set := make(map[string]bool)
	for p := range bytes.SplitSeq(out, []byte{0}) {
		if len(p) > 0 {
			set[string(p)] = true
		}
	}
	return set, nil
}
