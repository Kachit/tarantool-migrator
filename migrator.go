package tarantool_migrator

import (
	"context"
	"fmt"
	"github.com/tarantool/go-tarantool/v2/pool"
	"time"
)

func NewMigrator(tt pool.Pooler, migrations Migrations, options *Options) *Migrator {
	if options == nil {
		options = DefaultOptions
	}
	ex := newExecutor(tt, options)
	return &Migrator{
		ex:         ex,
		logger:     DefaultLogger.SetLogLevel(options.LogLevel),
		migrations: migrations,
		opts:       options,
	}
}

func (m *Migrator) SetLogger(lg Logger) *Migrator {
	m.logger = lg
	return m
}

type Migrator struct {
	ex         *Executor
	opts       *Options
	logger     Logger
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
		m.logger.Info(ctx, fmt.Sprintf(`migration "%s" process started`, mgr.ID))
		err = mgr.isValidForMigrate()
		if err != nil {
			return err
		}
		exists, err := m.ex.hasConfirmedMigration(ctx, mgr.ID)
		if err != nil {
			return err
		}
		if !exists {
			startedAt := time.Now().UTC()
			//migration process
			migratedAt := time.Now().UTC().Sub(startedAt)
			m.logger.Info(ctx, fmt.Sprintf(`migration "%s" successfully migrated in %.3fms`, mgr.ID, float64(migratedAt.Nanoseconds())/1e6))
		} else {
			m.logger.Info(ctx, fmt.Sprintf(`migration "%s" is already migrated`, mgr.ID))
		}
	}
	return nil
}

func (m *Migrator) RollbackLast(ctx context.Context) error {
	return nil
}
