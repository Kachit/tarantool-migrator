package tarantool_migrator

import (
	"context"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/pool"
)

// MigrateFunc is the func signature for migrating.
type MigrateFunc func(context.Context, pool.Pooler, *Options) error

// RollbackFunc is the func signature for rollback.
type RollbackFunc func(context.Context, pool.Pooler, *Options) error

// Migration represents a database migration (a modification to be made on the database).
type Migration struct {
	// ID is the migration identifier. Usually a timestamp like "201601021504".
	ID string
	// Migrate is a function that will br executed while running this migration.
	Migrate MigrateFunc
	// Rollback will be executed on rollback. Can be nil.
	Rollback RollbackFunc
}

func (mg *Migration) isValidForMigrate() error {
	if len(mg.ID) == 0 {
		return ErrMissingID
	}
	if mg.Migrate == nil {
		return ErrMissingMigrateFunc
	}

	return nil
}

func (mg *Migration) isValidForRollback() error {
	if len(mg.ID) == 0 {
		return ErrMissingID
	}
	if mg.Rollback == nil {
		return ErrMissingRollbackFunc
	}

	return nil
}

func NewGenericMigrateFunction(req string) func(context.Context, pool.Pooler, *Options) error {
	return func(ctx context.Context, tt pool.Pooler, opts *Options) error {
		_, err := tt.Do(tarantool.NewEvalRequest(req).Context(ctx), opts.WriteMode).Get()
		if err != nil {
			return err
		}

		return nil
	}
}
