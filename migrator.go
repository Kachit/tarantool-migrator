package tarantool_migrator

import (
	"context"
	"fmt"
	"github.com/tarantool/go-tarantool/v2/pool"
	"time"
)

func NewMigrator(tt pool.Pooler, migrations MigrationsCollection, options ...func(*Migrator)) *Migrator {
	m := &Migrator{
		logger:     DefaultLogger,
		migrations: migrations,
		opts:       DefaultOptions,
	}
	for _, opt := range options {
		opt(m)
	}
	m.ex = newExecutor(tt, m.opts)
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
		return ErrNoDefinedMigrations
	}
	err := m.ex.createMigrationsSpaceIfNotExists(ctx, createMigrationsSpacePath)
	if err != nil {
		return err
	}
	for _, migration := range m.migrations {
		m.logger.Info(ctx, fmt.Sprintf(`migration "%s" process started`, migration.ID))
		err = migration.isValidForMigrate()
		if err != nil {
			return err
		}
		exists, err := m.ex.hasAppliedMigration(ctx, migration.ID)
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
		return ErrNoDefinedMigrations
	}
	mgr, err := m.ex.findLastAppliedMigration(ctx)
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
