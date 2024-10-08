package tarantool_migrator

import (
	"context"
	"fmt"
	"github.com/tarantool/go-tarantool/v2/pool"
	"time"
)

func NewMigrator(tt pool.Pooler, migrations MigrationsCollection, options *Options) *Migrator {
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
	migrations MigrationsCollection
}

func (m *Migrator) Migrate(ctx context.Context) error {
	if m.migrations.IsEmpty() {
		return ErrNoMigrationsDefined
	}
	err := m.ex.createMigrationsSpaceIfNotExists(ctx)
	if err != nil {
		return err
	}
	for _, migration := range m.migrations {
		m.logger.Info(ctx, fmt.Sprintf(`migration "%s" process started`, migration.ID))
		err = migration.isValidForMigrate()
		if err != nil {
			return err
		}
		exists, err := m.ex.hasConfirmedMigration(ctx, migration.ID)
		if err != nil {
			return err
		}
		if !exists {
			startedAt := time.Now().UTC()
			err = m.ex.applyMigration(ctx, migration)
			if err != nil {
				return err
			}
			migratedAt := time.Now().UTC().Sub(startedAt)
			m.logger.Info(ctx, fmt.Sprintf(`migration "%s" successfully migrated in %.3fms`, migration.ID, float64(migratedAt.Nanoseconds())/1e6))
		} else {
			m.logger.Info(ctx, fmt.Sprintf(`migration "%s" is already migrated`, migration.ID))
		}
	}
	return nil
}

func (m *Migrator) RollbackLast(ctx context.Context) error {
	if m.migrations.IsEmpty() {
		return ErrNoMigrationsDefined
	}
	mgr, err := m.ex.findLastConfirmedMigration(ctx)
	if err != nil {
		return err
	}
	m.logger.Info(ctx, fmt.Sprintf(`migration "%s" found for rollback`, mgr.ID))
	migration, err := m.migrations.Find(mgr.ID)
	if err != nil {
		return err
	}
	err = migration.isValidForRollback()
	if err != nil {
		return err
	}
	startedAt := time.Now().UTC()
	err = m.ex.rollbackMigration(ctx, migration)
	if err != nil {
		return err
	}
	rolledAt := time.Now().UTC().Sub(startedAt)
	m.logger.Info(ctx, fmt.Sprintf(`migration "%s" successfully rollbacked in %.3fms`, mgr.ID, float64(rolledAt.Nanoseconds())/1e6))
	return nil
}
