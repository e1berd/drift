package gitrepo

import (
	"context"
	"os"
	"strings"
	"time"
)

// Kinds of activity entries and removal targets.
const (
	KindBranch   = "branch"
	KindWorktree = "worktree"
	KindFile     = "file"
)

// RemovedByUser marks an entry removed from the info bar; anything else missing was removed elsewhere.
const RemovedByUser = "user"

// Removal is a file or worktree deleted from the info bar, stored as a system note of the session.
type Removal struct {
	SourceID string
	Target   string
	Name     string
	Branch   string
	Time     time.Time
}

// ActivityEntry is a branch, worktree or file the session created; a removed one stays in the list.
type ActivityEntry struct {
	Key        string     `json:"key"`
	Kind       string     `json:"kind"`
	MessageID  string     `json:"messageId"`
	Name       string     `json:"name"`
	Removed    bool       `json:"removed"`
	RemovedBy  string     `json:"removedBy,omitempty"`
	RemoveTime *time.Time `json:"removeTime,omitempty"`
	Branch     string     `json:"branch,omitempty"`
	Current    bool       `json:"current,omitempty"`
	Upstream   string     `json:"upstream,omitempty"`
	Commits    int        `json:"commits,omitempty"`
	Unpushed   int        `json:"unpushed,omitempty"`
	Changes    int        `json:"changes,omitempty"`
	path       string
}

type activityBuilder struct {
	r        repo
	ctx      context.Context
	base     string
	trees    []worktreeEntry
	branches map[string]Branch
	removals map[string]Removal
	seen     map[string]bool
	entries  []ActivityEntry
}

func (r repo) activity(ctx context.Context, base string, h History, repository bool) ([]ActivityEntry, error) {
	b := activityBuilder{r: r, ctx: ctx, base: base, removals: removalIndex(h.Removals), seen: map[string]bool{}}
	if repository {
		if err := b.loadRepository(); err != nil {
			return nil, err
		}
	}
	for _, ref := range CreatedRefs(h.Calls) {
		b.addRef(ref)
	}
	if b.entries == nil {
		return []ActivityEntry{}, nil
	}
	return b.entries, nil
}

func removalIndex(removals []Removal) map[string]Removal {
	index := make(map[string]Removal, len(removals))
	for _, removal := range removals {
		if removal.Target == KindWorktree {
			index[worktreeRemovalKey(removal.SourceID)] = removal
		} else {
			index[removal.Target+":"+removal.Name] = removal
		}
		if removal.Branch != "" {
			index[KindBranch+":"+removal.Branch] = removal
		}
	}
	return index
}

func worktreeRemovalKey(sourceID string) string {
	return KindWorktree + "@" + sourceID
}

func (b *activityBuilder) removalOf(entry ActivityEntry) (Removal, bool) {
	if entry.Kind == KindWorktree {
		removal, ok := b.removals[worktreeRemovalKey(entry.MessageID)]
		return removal, ok
	}
	removal, ok := b.removals[entry.Kind+":"+entry.Name]
	return removal, ok
}

func (b *activityBuilder) loadRepository() error {
	entries, err := b.r.worktreeEntries(b.ctx)
	if err != nil {
		return err
	}
	if len(entries) > 1 {
		b.trees = entries[1:]
	}
	list, err := b.r.branches(b.ctx)
	if err != nil {
		return err
	}
	b.branches = make(map[string]Branch, len(list))
	for _, branch := range list {
		b.branches[branch.Name] = branch
	}
	return nil
}

func (b *activityBuilder) repository() bool {
	return b.branches != nil
}

func (b *activityBuilder) addRef(ref CreatedRef) {
	if ref.File != "" {
		b.add(b.file(ref))
	}
	if !b.repository() {
		return
	}
	treeBranch := ""
	if ref.Worktree != "" {
		tree := b.worktree(ref)
		treeBranch = tree.Branch
		b.add(tree)
	}
	if branch := firstNonEmpty(ref.Branch, treeBranch); branch != "" {
		b.add(b.branch(branch, ref.MessageID))
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func (b *activityBuilder) add(entry ActivityEntry) {
	if removal, ok := b.removalOf(entry); ok && entry.Removed {
		markRemovedByUser(&entry, removal)
	}
	entry.Key = entry.Kind + ":" + entry.Name
	if b.seen[entry.Key] {
		return
	}
	b.seen[entry.Key] = true
	b.entries = append(b.entries, entry)
}

func markRemovedByUser(entry *ActivityEntry, removal Removal) {
	entry.RemovedBy = RemovedByUser
	removeTime := removal.Time
	entry.RemoveTime = &removeTime
	if entry.Kind == KindWorktree && removal.Name != "" {
		entry.Name = removal.Name
	}
}

func (b *activityBuilder) file(ref CreatedRef) ActivityEntry {
	full := b.r.absolute("", ref.File)
	info, err := os.Lstat(full)
	present := err == nil && !info.IsDir()
	return ActivityEntry{Kind: KindFile, MessageID: ref.MessageID, Name: ref.File, Removed: !present, path: full}
}

func (b *activityBuilder) worktree(ref CreatedRef) ActivityEntry {
	entry, ok := b.r.matchWorktree(b.trees, ref)
	if !ok {
		return ActivityEntry{Kind: KindWorktree, MessageID: ref.MessageID, Name: b.r.missingWorktreeName(ref), Removed: true}
	}
	return ActivityEntry{
		Kind: KindWorktree, MessageID: ref.MessageID, Name: b.r.displayPath(entry.path), Branch: entry.branch,
		Changes: worktreeChanges(b.ctx, entry.path), path: entry.path,
	}
}

func (r repo) missingWorktreeName(ref CreatedRef) string {
	if !strings.ContainsRune(ref.Worktree, '/') && !strings.HasPrefix(ref.Worktree, ".") {
		return ref.Worktree
	}
	return r.displayPath(canonical(r.absolute(ref.Dir, ref.Worktree)))
}

func (b *activityBuilder) branch(name, messageID string) ActivityEntry {
	found, ok := b.branches[name]
	if !ok {
		return ActivityEntry{Kind: KindBranch, MessageID: messageID, Name: name, Removed: true}
	}
	return ActivityEntry{
		Kind: KindBranch, MessageID: messageID, Name: name, Current: found.Current, Upstream: found.Upstream,
		Commits: b.r.countCommits(b.ctx, b.base, name), Unpushed: b.r.countCommits(b.ctx, found.Upstream, name),
	}
}
