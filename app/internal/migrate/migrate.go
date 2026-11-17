// Package migrate applies embedded up-migrations on startup. It keeps golang-migrate's
// schema_migrations table (one row: version, dirty), so `task app:migration:*` keeps working.
package migrate

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const lockKey = 7_301_942_118

type migration struct {
	version uint64
	name    string
}

// Up applies every *.up.sql in dir newer than the recorded version, each in its own transaction.
func Up(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS, dir string, log *slog.Logger) error {
	list, err := collect(fsys, dir)
	if err != nil {
		return err
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", lockKey); err != nil {
		return fmt.Errorf("migrate: lock: %w", err)
	}
	defer conn.Exec(context.WithoutCancel(ctx), "SELECT pg_advisory_unlock($1)", lockKey)

	if _, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version bigint NOT NULL PRIMARY KEY, dirty boolean NOT NULL)`); err != nil {
		return fmt.Errorf("migrate: create table: %w", err)
	}

	var current int64
	var dirty bool
	err = conn.QueryRow(ctx, "SELECT version, dirty FROM schema_migrations LIMIT 1").Scan(&current, &dirty)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("migrate: read version: %w", err)
	}
	if dirty {
		return fmt.Errorf("migrate: schema is dirty at version %d, fix it and run `task app:migration:force`", current)
	}

	for _, m := range list {
		if m.version <= uint64(current) {
			continue
		}
		sql, err := fs.ReadFile(fsys, path.Join(dir, m.name))
		if err != nil {
			return err
		}
		err = pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, string(sql)); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, "DELETE FROM schema_migrations"); err != nil {
				return err
			}
			_, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version, dirty) VALUES ($1, false)", int64(m.version))
			return err
		})
		if err != nil {
			return fmt.Errorf("migrate: %s: %w", m.name, err)
		}
		log.Info("migration applied", "file", m.name)
	}
	return nil
}

func collect(fsys fs.FS, dir string) ([]migration, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}
	var list []migration
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		prefix, _, ok := strings.Cut(name, "_")
		if !ok {
			return nil, fmt.Errorf("migrate: bad file name %q", name)
		}
		v, err := strconv.ParseUint(prefix, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("migrate: bad version in %q: %w", name, err)
		}
		list = append(list, migration{version: v, name: name})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].version < list[j].version })
	return list, nil
}
