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

func (e *Executor) initMigrationsSpace(ctx context.Context) error {
	migrationSpaceRequest := `
box.schema.create_space('{migrations_space}', { if_not_exists = true, format={
    {'id',type='string'},
    {'executed_at',type='string'},
}})

box.space.{migrations_space}:create_index('id', {parts = {'id'}, if_not_exists = true, unique = true})
`
	migrationSpaceRequest = strings.ReplaceAll(migrationSpaceRequest, "{migrations_space}", e.opts.SpaceName)
	_, err := e.tt.Do(tarantool.NewEvalRequest(migrationSpaceRequest).Context(ctx), e.opts.WriteMode).Get()
	if err != nil {
		return err
	}
	return nil
}

func (e *Executor) hasConfirmedMigration(ctx context.Context, migrationID string) (bool, error) {
	var tuples []migrationTuple
	err := e.tt.Do(tarantool.NewSelectRequest(e.opts.SpaceName).Context(ctx).Key([]any{migrationID}), e.opts.ReadMode).GetTyped(&tuples)
	if err != nil {
		return false, err
	}
	return len(tuples) > 0, nil
}

func (e *Executor) confirmMigration(ctx context.Context, migrationID string) error {
	_, err := e.tt.Do(tarantool.NewReplaceRequest(e.opts.SpaceName).Context(ctx).Tuple([]interface{}{
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

func (e *Executor) rejectMigration(ctx context.Context, migrationID string) error {
	_, err := e.tt.Do(tarantool.NewDeleteRequest(e.opts.SpaceName).Context(ctx).Key([]any{migrationID}), e.opts.WriteMode).Get()
	if err != nil {
		return err
	}
	return nil
}

func (e *Executor) findLastConfirmedMigration(ctx context.Context) (*migrationTuple, error) {
	var tuples []migrationTuple
	expr := fmt.Sprintf("return box.space.%s.index.id:max()", e.opts.SpaceName)
	err := e.tt.Do(tarantool.NewEvalRequest(expr).Context(ctx), e.opts.ReadMode).GetTyped(&tuples)
	if err != nil {
		return nil, err
	}
	if len(tuples) == 0 {
		return nil, fmt.Errorf("no confirmed migrations")
	}
	return &tuples[0], nil
}
