package runner

import (
	"context"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/gitrepo"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *Runner) rememberGitBase(ctx context.Context, session pgtype.UUID, dir string) {
	base := gitrepo.HeadCommit(ctx, dir)
	if base == "" {
		return
	}
	if err := r.q.SetSessionGitBase(r.ctx, db.SetSessionGitBaseParams{ID: session, GitBase: base}); err != nil {
		r.log.Warn("save session git base", "err", err)
	}
}
