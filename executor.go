package tarantool_migrator

import (
	"context"
	"strings"
	"time"

	"github.com/tarantool/go-tarantool/v3"
	"github.com/tarantool/go-tarantool/v3/pool"
)

type Executor struct {
	tt   pool.Pooler
	opts *Options
}

func newExecutor(tt pool.Pooler, options *Options) *Executor {
	return &Executor{tt: tt, opts: options}
}

func (e *Executor) createMigrationsSpaceIfNotExists(ctx context.Context, path string) error {
	data, err := LuaFs.ReadFile(path)
	if err != nil {
		return err
	}

	migrationSpaceRequest := string(data)
	migrationSpaceRequest = strings.ReplaceAll(migrationSpaceRequest, "_migrations_space_", e.opts.MigrationsSpace)

	_, err = e.tt.Do(tarantool.NewEvalRequest(migrationSpaceRequest).Context(ctx), e.opts.WriteMode).Get()

	return err
}

func (e *Executor) hasAppliedMigration(ctx context.Context, migrationID string) (bool, error) {
	var tuples []migrationTuple

	req := tarantool.NewSelectRequest(e.opts.MigrationsSpace).Context(ctx).Key([]any{migrationID})

	err := e.tt.Do(req, e.opts.ReadMode).GetTyped(&tuples)
	if err != nil {
		return false, err
	}

	return len(tuples) > 0, nil
}

func (e *Executor) insertMigration(ctx context.Context, migrationID string) error {
	_, err := e.tt.Do(tarantool.NewInsertRequest(e.opts.MigrationsSpace).Context(ctx).Tuple([]interface{}{
		migrationID,
		time.Now().UTC().Format(time.RFC3339),
	}),
		e.opts.WriteMode,
	).Get()

	return err
}

func (e *Executor) deleteMigration(ctx context.Context, migrationID string) error {
	req := tarantool.NewDeleteRequest(e.opts.MigrationsSpace).Context(ctx).Key([]any{migrationID})

	_, err := e.tt.Do(req, e.opts.WriteMode).Get()

	return err
}

func (e *Executor) applyMigration(ctx context.Context, migration *Migration) error {
	if e.opts.DryRun {
		return nil
	}

	if e.opts.TransactionsEnabled {
		stream, err := e.tt.NewStream(e.opts.WriteMode)
		if err != nil {
			return err
		}

		return e.applyMigrationInTx(ctx, stream, migration)
	}

	if err := migration.Migrate(ctx, e.tt, e.opts); err != nil {
		return err
	}

	return e.insertMigration(ctx, migration.ID)
}

func (e *Executor) rollbackMigration(ctx context.Context, migration *Migration) error {
	if e.opts.DryRun {
		return nil
	}

	if e.opts.TransactionsEnabled {
		stream, err := e.tt.NewStream(e.opts.WriteMode)
		if err != nil {
			return err
		}

		return e.rollbackMigrationInTx(ctx, stream, migration)
	}

	if err := migration.Rollback(ctx, e.tt, e.opts); err != nil {
		return err
	}

	return e.deleteMigration(ctx, migration.ID)
}

func (e *Executor) runInTx(ctx context.Context, s streamer, fn func(pool.Pooler) error) error {
	if _, err := s.Do(tarantool.NewBeginRequest().Context(ctx)).Get(); err != nil {
		return err
	}

	if err := fn(&streamPooler{s: s}); err != nil {
		_, _ = s.Do(tarantool.NewRollbackRequest().Context(ctx)).Get()

		return err
	}

	if _, err := s.Do(tarantool.NewCommitRequest().Context(ctx)).Get(); err != nil {
		_, _ = s.Do(tarantool.NewRollbackRequest().Context(ctx)).Get()

		return err
	}

	return nil
}

func (e *Executor) applyMigrationInTx(ctx context.Context, s streamer, migration *Migration) error {
	return e.runInTx(ctx, s, func(sp pool.Pooler) error {
		if err := migration.Migrate(ctx, sp, e.opts); err != nil {
			return err
		}

		_, err := s.Do(tarantool.NewInsertRequest(e.opts.MigrationsSpace).Context(ctx).Tuple([]interface{}{
			migration.ID,
			time.Now().UTC().Format(time.RFC3339),
		})).Get()

		return err
	})
}

func (e *Executor) rollbackMigrationInTx(ctx context.Context, s streamer, migration *Migration) error {
	return e.runInTx(ctx, s, func(sp pool.Pooler) error {
		if err := migration.Rollback(ctx, sp, e.opts); err != nil {
			return err
		}

		_, err := s.Do(tarantool.NewDeleteRequest(e.opts.MigrationsSpace).Context(ctx).Key([]any{migration.ID})).Get()

		return err
	})
}

func (e *Executor) findLastAppliedMigration(ctx context.Context) (*migrationTuple, error) {
	var tuples []migrationTuple

	cmd := "return box.space._migrations_space_.index.id:max()"
	expr := strings.ReplaceAll(cmd, "_migrations_space_", e.opts.MigrationsSpace)

	err := e.tt.Do(tarantool.NewEvalRequest(expr).Context(ctx), e.opts.ReadMode).GetTyped(&tuples)
	if err != nil {
		return nil, err
	}

	if len(tuples) == 0 {
		return nil, ErrNoAppliedMigrations
	}

	return &tuples[0], nil
}
