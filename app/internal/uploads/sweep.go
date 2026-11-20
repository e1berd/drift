package uploads

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

const (
	sweepInterval      = time.Hour
	partialUploadTTL   = 7 * 24 * time.Hour
	orphanGracePeriod  = time.Hour
	orphanLookupsBatch = 500
)

// Sweep deletes unclaimed attachments, abandoned partial uploads and files whose attachment
// is gone (its message or session was deleted), now and then every hour until ctx ends.
func (s *Server) Sweep(ctx context.Context) {
	ticker := time.NewTicker(sweepInterval)
	defer ticker.Stop()
	for {
		s.sweepOnce(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *Server) sweepOnce(ctx context.Context) {
	if _, err := s.q.DeleteStaleAttachments(ctx); err != nil {
		s.log.Warn("delete stale attachments", "err", err)
	}
	s.removeAbandonedPartials(time.Now())
	if err := s.removeOrphanFiles(ctx, time.Now()); err != nil {
		s.log.Warn("remove orphan attachment files", "err", err)
	}
}

func (s *Server) removeAbandonedPartials(now time.Time) {
	entries, err := os.ReadDir(s.store.partialDir())
	if err != nil {
		s.log.Warn("list partial uploads", "err", err)
		return
	}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil || now.Sub(info.ModTime()) < partialUploadTTL {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".info")
		if s.uploadTouchedSince(id, now.Add(-partialUploadTTL)) {
			continue
		}
		_ = os.Remove(filepath.Join(s.store.partialDir(), entry.Name()))
	}
}

func (s *Server) uploadTouchedSince(id string, since time.Time) bool {
	info, err := os.Stat(filepath.Join(s.store.partialDir(), id))
	return err == nil && info.ModTime().After(since)
}

func (s *Server) removeOrphanFiles(ctx context.Context, now time.Time) error {
	entries, err := os.ReadDir(s.store.filesDir())
	if err != nil {
		return err
	}
	candidates := make(map[pgtype.UUID]string)
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil || now.Sub(info.ModTime()) < orphanGracePeriod {
			continue
		}
		var id pgtype.UUID
		if id.Scan(entry.Name()) == nil {
			candidates[id] = entry.Name()
		}
	}
	for batch := range batches(candidates) {
		kept, err := s.q.ListStoredAttachmentIDs(ctx, batch)
		if err != nil {
			return err
		}
		for _, id := range kept {
			delete(candidates, id)
		}
	}
	for _, name := range candidates {
		_ = os.RemoveAll(filepath.Join(s.store.filesDir(), name))
	}
	return nil
}

func batches(ids map[pgtype.UUID]string) func(yield func([]pgtype.UUID) bool) {
	return func(yield func([]pgtype.UUID) bool) {
		batch := make([]pgtype.UUID, 0, orphanLookupsBatch)
		for id := range ids {
			batch = append(batch, id)
			if len(batch) == orphanLookupsBatch {
				if !yield(batch) {
					return
				}
				batch = make([]pgtype.UUID, 0, orphanLookupsBatch)
			}
		}
		if len(batch) > 0 {
			yield(batch)
		}
	}
}
