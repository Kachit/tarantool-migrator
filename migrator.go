package tarantool_migrator

import (
	"context"
	"github.com/tarantool/go-tarantool/v2/pool"
)

func NewMigrator(tt pool.Pooler, migrations Migrations, options *Options) *Migrator {
	if options == nil {
		options = DefaultOptions
	}
	ex := newExecutor(tt, options)
	return &Migrator{ex: ex, migrations: migrations, opts: options}
}

type Migrator struct {
	ex         *Executor
	opts       *Options
	migrations Migrations
}

func (m *Migrator) Migrate(ctx context.Context) error {
	if m.migrations.IsEmpty() {
		return ErrNoMigrationsDefined
	}
	err := m.ex.createMigrationsSpaceIfNotExists(ctx)
	if err != nil {
		return err
	}
	for _, mgr := range m.migrations {
		err = mgr.isValidForMigrate()
		if err != nil {
			return err
		}
		exists, err := m.ex.hasConfirmedMigration(ctx, mgr.ID)
		if err != nil {
			return err
		}
		if !exists {

		}
	}
	return nil
}

func (m *Migrator) RollbackLast(ctx context.Context) error {
	return nil
}
