package runner

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/e1berd/drift/app/internal/workspace"
	"github.com/jackc/pgx/v5/pgtype"
)

func (t *transcript) createsFile(name, input string) bool {
	if name != "Write" || t.dir == "" || strings.TrimSpace(input) == "" {
		return false
	}
	path := strings.TrimSpace(input)
	if !filepath.IsAbs(path) {
		path = filepath.Join(t.dir, path)
	}
	_, err := os.Lstat(path)
	return errors.Is(err, fs.ErrNotExist)
}

func (r *Runner) withRemovalNotice(session pgtype.UUID, text string) string {
	rows, err := r.q.ListUndeliveredRemovals(r.ctx, session)
	if err != nil {
		r.log.Warn("load removal notes", "err", err)
		return text
	}
	if len(rows) == 0 {
		return text
	}
	lines := make([]string, 0, len(rows))
	ids := make([]pgtype.UUID, 0, len(rows))
	for _, row := range rows {
		var note workspace.RemovalNote
		if json.Unmarshal(row.Body, &note) != nil {
			continue
		}
		lines = append(lines, removalLine(note))
		ids = append(ids, row.ID)
	}
	if err := r.q.MarkRemovalsDelivered(r.ctx, ids); err != nil {
		r.log.Warn("mark removal notes delivered", "err", err)
	}
	return removalNotice(lines) + text
}

func removalLine(note workspace.RemovalNote) string {
	if note.Target == "worktree" && note.Branch != "" {
		return "- worktree " + note.Path + " and its branch " + note.Branch
	}
	return "- " + note.Target + " " + note.Path
}

func removalNotice(lines []string) string {
	return "[Drift] Since your last turn the user deleted these from the session's info bar; they no longer exist, " +
		"so do not use them and do not recreate them unless asked:\n" + strings.Join(lines, "\n") + "\n\n"
}
