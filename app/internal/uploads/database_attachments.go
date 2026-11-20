package uploads

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

// MoveDatabaseAttachments writes images that older versions kept in Postgres to disk, so every
// attachment is a file afterwards.
func (s *Server) MoveDatabaseAttachments(ctx context.Context) error {
	for {
		rows, err := s.q.ListDatabaseAttachments(ctx)
		if err != nil {
			return fmt.Errorf("list database attachments: %w", err)
		}
		if len(rows) == 0 {
			return nil
		}
		for _, row := range rows {
			if err := s.moveToDisk(ctx, row); err != nil {
				return fmt.Errorf("move attachment to disk: %w", err)
			}
		}
	}
}

func (s *Server) moveToDisk(ctx context.Context, row db.ListDatabaseAttachmentsRow) error {
	id, _ := row.ID.Value()
	rel := relativePath(fmt.Sprint(id), row.Name)
	target := s.store.Path(rel)
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(target, row.Data, 0o600); err != nil {
		return err
	}
	return s.q.MoveAttachmentToDisk(ctx, db.MoveAttachmentToDiskParams{ID: row.ID, Path: pgtype.Text{String: rel, Valid: true}})
}
