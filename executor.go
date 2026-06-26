package tarantool_migrator

import (
	"context"
	"fmt"
	"strings"

	"github.com/tarantool/go-tarantool/v3"
	"github.com/tarantool/go-tarantool/v3/pool"
)

type executor interface {
	createMigrationsSpaceIfNotExists(ctx context.Context, path string) error
	hasAppliedMigration(ctx context.Context, migrationID string) (bool, error)
	applyMigration(ctx context.Context, migration *Migration) error
	rollbackMigration(ctx context.Context, migration *Migration) error
	findLastAppliedMigration(ctx context.Context) (*migrationTuple, error)
}

type executorBase struct {
	tt   pool.Pooler
	opts *Options
}

func newExecutor(tt pool.Pooler, opts *Options) executor {
	base := executorBase{tt: tt, opts: opts}

	return &noTxExecutor{executorBase: base}
}

func (e *executorBase) createMigrationsSpaceIfNotExists(ctx context.Context, path string) error {
	data, err := LuaFs.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read lua script: %w", err)
	}

	migrationSpaceRequest := string(data)
	migrationSpaceRequest = strings.ReplaceAll(migrationSpaceRequest, "_migrations_space_", e.opts.MigrationsSpace)

	_, err = e.tt.Do(tarantool.NewEvalRequest(migrationSpaceRequest).Context(ctx), e.opts.WriteMode).Get()
	if err != nil {
		return fmt.Errorf("exec create migrations space: %w", err)
	}

	return nil
}

func (e *executorBase) hasAppliedMigration(ctx context.Context, migrationID string) (bool, error) {
	var tuples []migrationTuple

	req := tarantool.NewSelectRequest(e.opts.MigrationsSpace).Context(ctx).Key([]any{migrationID})

	err := e.tt.Do(req, e.opts.ReadMode).GetTyped(&tuples)
	if err != nil {
		return false, fmt.Errorf("check applied migration: %w", err)
	}

	return len(tuples) > 0, nil
}

func (e *executorBase) insertMigration(ctx context.Context, tt pool.Pooler, migrationID string) error {
	tuple := newMigrationTuple(migrationID)

	_, err := tt.Do(tarantool.NewInsertRequest(e.opts.MigrationsSpace).Context(ctx).Tuple(tuple.ToSlice()),
		e.opts.WriteMode,
	).Get()
	if err != nil {
		return fmt.Errorf("insert migration record: %w", err)
	}

	return nil
}

func (e *executorBase) deleteMigration(ctx context.Context, tt pool.Pooler, migrationID string) error {
	req := tarantool.NewDeleteRequest(e.opts.MigrationsSpace).Context(ctx).Key([]any{migrationID})

	_, err := tt.Do(req, e.opts.WriteMode).Get()
	if err != nil {
		return fmt.Errorf("delete migration record: %w", err)
	}

	return nil
}

func (e *executorBase) findLastAppliedMigration(ctx context.Context) (*migrationTuple, error) {
	var tuples []migrationTuple

	cmd := "return box.space._migrations_space_.index.id:max()"
	expr := strings.ReplaceAll(cmd, "_migrations_space_", e.opts.MigrationsSpace)

	err := e.tt.Do(tarantool.NewEvalRequest(expr).Context(ctx), e.opts.ReadMode).GetTyped(&tuples)
	if err != nil {
		return nil, fmt.Errorf("find last applied migration: %w", err)
	}

	if len(tuples) == 0 {
		return nil, ErrNoAppliedMigrations
	}

	return &tuples[0], nil
}
