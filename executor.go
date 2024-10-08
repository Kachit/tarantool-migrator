package tarantool_migrator

import (
	"context"
	"fmt"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/pool"
	"strings"
	"time"
)

type migrationTuple struct {
	ID         string
	ExecutedAt string
}

type Executor struct {
	tt   pool.Pooler
	opts *Options
}

func newExecutor(tt pool.Pooler, options *Options) *Executor {
	return &Executor{tt: tt, opts: options}
}

func (e *Executor) createMigrationsSpaceIfNotExists(ctx context.Context) error {
	data, err := LuaFs.ReadFile("lua/migrations/create_migrations_space.up.lua")
	if err != nil {
		return err
	}
	migrationSpaceRequest := string(data)
	migrationSpaceRequest = strings.ReplaceAll(migrationSpaceRequest, "_migrations_space_", e.opts.MigrationsSpace)
	_, err = e.tt.Do(tarantool.NewEvalRequest(migrationSpaceRequest).Context(ctx), e.opts.WriteMode).Get()
	if err != nil {
		return err
	}
	return nil
}

func (e *Executor) hasConfirmedMigration(ctx context.Context, migrationID string) (bool, error) {
	var tuples []migrationTuple
	err := e.tt.Do(tarantool.NewSelectRequest(e.opts.MigrationsSpace).Context(ctx).Key([]any{migrationID}), e.opts.ReadMode).GetTyped(&tuples)
	if err != nil {
		return false, err
	}
	return len(tuples) > 0, nil
}

func (e *Executor) insertMigration(ctx context.Context, migrationID string) error {
	_, err := e.tt.Do(tarantool.NewInsertRequest(e.opts.MigrationsSpace).Context(ctx).Tuple([]interface{}{
		migrationID,
		time.Now().UTC().String(),
	}),
		e.opts.WriteMode,
	).Get()
	if err != nil {
		return err
	}
	return nil
}

func (e *Executor) deleteMigration(ctx context.Context, migrationID string) error {
	_, err := e.tt.Do(tarantool.NewDeleteRequest(e.opts.MigrationsSpace).Context(ctx).Key([]any{migrationID}), e.opts.WriteMode).Get()
	if err != nil {
		return err
	}
	return nil
}

func (e *Executor) applyMigration(ctx context.Context, migration *Migration) error {
	err := migration.Migrate(ctx, e.tt, e.opts)
	if err != nil {
		return err
	}
	return nil
}

func (e *Executor) rollbackMigration(ctx context.Context, migration *Migration) error {
	err := migration.Rollback(ctx, e.tt, e.opts)
	if err != nil {
		return err
	}
	return nil
}

func (e *Executor) findLastConfirmedMigration(ctx context.Context) (*migrationTuple, error) {
	var tuples []migrationTuple
	expr := strings.ReplaceAll("return box.space._migrations_space_.index.id:max()", "_migrations_space_", e.opts.MigrationsSpace)
	err := e.tt.Do(tarantool.NewEvalRequest(expr).Context(ctx), e.opts.ReadMode).GetTyped(&tuples)
	if err != nil {
		return nil, err
	}
	if len(tuples) == 0 {
		return nil, fmt.Errorf("no confirmed migrations")
	}
	return &tuples[0], nil
}
