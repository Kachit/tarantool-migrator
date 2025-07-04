package tarantool_migrator

import (
	"context"
	"fmt"
	"github.com/tarantool/go-tarantool/v2/pool"
	"time"
)

func NewMigrator(tt pool.Pooler, migrations MigrationsCollection, options ...func(*Migrator)) *Migrator {
	opts := DefaultOptions
	m := &Migrator{
		logger:     DefaultLogger,
		opts:       &opts,
		migrations: migrations,
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
	m.logger.Debug(ctx, fmt.Sprintf(`started "migrate" command with "%d" migrations and options:`,
		len(m.migrations)), m.opts)
	if m.migrations.IsEmpty() {
		return ErrNoDefinedMigrations
	}
	err := m.ex.createMigrationsSpaceIfNotExists(ctx, createMigrationsSpacePath)
	if err != nil {
		return fmt.Errorf(`init migrations space error: %w`, err)
	}
	for _, migration := range m.migrations {
		m.logger.Info(ctx, fmt.Sprintf(`migration "%s" process started`, migration.ID))
		err = migration.isValidForMigrate()
		if err != nil {
			return fmt.Errorf(`migration "%s" error: %w`, migration.ID, err)
		}
		exists, err := m.ex.hasAppliedMigration(ctx, migration.ID)
		if err != nil {
			return fmt.Errorf(`migration "%s" error: %w`, migration.ID, err)
		}
		if !exists {
			startedAt := time.Now().UTC()
			err = m.ex.applyMigration(ctx, migration)
			if err != nil {
				return fmt.Errorf(`migration "%s" error: %w`, migration.ID, err)
			}
			migratedAt := time.Now().UTC().Sub(startedAt)
			m.logger.Info(ctx, fmt.Sprintf(`migration "%s" successfully migrated in %.3fms`,
				migration.ID, formatDurationToMs(migratedAt)))
		} else {
			m.logger.Info(ctx, fmt.Sprintf(`migration "%s" is already migrated`, migration.ID))
		}
	}

	return nil
}

func (m *Migrator) RollbackLast(ctx context.Context) error {
	m.logger.Debug(ctx, fmt.Sprintf(`started "rollback-last" command with "%d" migrations and options:`,
		len(m.migrations)), m.opts)
	if m.migrations.IsEmpty() {
		return ErrNoDefinedMigrations
	}
	mgr, err := m.ex.findLastAppliedMigration(ctx)
	if err != nil {
		return fmt.Errorf(`find applied migration error: %w`, err)
	}
	m.logger.Info(ctx, fmt.Sprintf(`migration "%s" found for rollback`, mgr.ID))
	migration, err := m.migrations.Find(mgr.ID)
	if err != nil {
		return fmt.Errorf(`migration "%s" error: %w`, mgr.ID, err)
	}
	err = migration.isValidForRollback()
	if err != nil {
		return fmt.Errorf(`migration "%s" error: %w`, mgr.ID, err)
	}
	startedAt := time.Now().UTC()
	err = m.ex.rollbackMigration(ctx, migration)
	if err != nil {
		return fmt.Errorf(`migration "%s" error: %w`, mgr.ID, err)
	}
	rolledAt := time.Now().UTC().Sub(startedAt)
	m.logger.Info(ctx, fmt.Sprintf(`migration "%s" successfully rollbacked in %.3fms`,
		mgr.ID, formatDurationToMs(rolledAt)))

	return nil
}

func WithLogger(lg Logger) func(migrator *Migrator) {
	return func(m *Migrator) {
		m.logger = lg
	}
}

func WithOptions(op *Options) func(migrator *Migrator) {
	return func(m *Migrator) {
		m.opts = op
	}
}
