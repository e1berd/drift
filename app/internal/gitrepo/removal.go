package gitrepo

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"github.com/labstack/echo/v5"
)

func (r repo) presentEntry(ctx context.Context, h History, kind, name string) (ActivityEntry, error) {
	ok, err := r.isRepository(ctx)
	if err != nil {
		return ActivityEntry{}, err
	}
	base := ""
	if ok {
		base = r.validBase(ctx, h.Base)
	}
	entries, err := r.activity(ctx, base, h, ok)
	if err != nil {
		return ActivityEntry{}, err
	}
	for _, entry := range entries {
		if entry.Kind == kind && entry.Name == name {
			if entry.Removed {
				return entry, echo.NewHTTPError(http.StatusConflict, kind+" is already removed")
			}
			return entry, nil
		}
	}
	return ActivityEntry{}, echo.NewHTTPError(http.StatusNotFound, "the session did not create this "+kind)
}

func (r repo) removeFile(ctx context.Context, h History, name string) (Removal, error) {
	entry, err := r.presentEntry(ctx, h, KindFile, name)
	if err != nil {
		return Removal{}, err
	}
	if err := os.Remove(entry.path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return Removal{}, fmt.Errorf("remove file: %w", err)
	}
	return Removal{SourceID: entry.MessageID, Target: KindFile, Name: entry.Name}, nil
}

func (r repo) removeWorktree(ctx context.Context, h History, name string, force, deleteBranch bool) (Removal, error) {
	entry, err := r.presentEntry(ctx, h, KindWorktree, name)
	if err != nil {
		return Removal{}, err
	}
	if entry.Changes > 0 && !force {
		return Removal{}, echo.NewHTTPError(http.StatusConflict, "the worktree has uncommitted changes")
	}
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force", "--force")
	}
	if _, err := r.git(ctx, append(args, entry.path)...); err != nil {
		return Removal{}, err
	}
	removal := Removal{SourceID: entry.MessageID, Target: KindWorktree, Name: entry.Name}
	if deleteBranch && entry.Branch != "" {
		if _, err := r.git(ctx, "branch", "-D", entry.Branch); err != nil {
			return removal, err
		}
		removal.Branch = entry.Branch
	}
	return removal, nil
}
