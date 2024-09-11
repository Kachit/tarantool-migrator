package tarantool_migrator

import (
	"fmt"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/pool"
	"log"
	"time"
)

func NewMigrator(tt *pool.ConnectionPool, migrations []*Migration, options *Options) *Migrator {
	if options == nil {
		options = DefaultOptions
	}
	return &Migrator{tt: tt, migrations: migrations, opts: options}
}

type Migrator struct {
	tt         *pool.ConnectionPool
	opts       *Options
	migrations []*Migration
}

type migrationRow struct {
	ID         string
	ExecutedAt string
}

func (m *Migrator) createMigrationsSpaceIfNotExists() error {
	var exists []bool
	expr := fmt.Sprintf("return box.space.%s ~= nil", m.opts.SpaceName)
	err := m.tt.Do(tarantool.NewEvalRequest(expr), pool.ANY).GetTyped(&exists)
	if err != nil {
		return err
	}
	if len(exists) == 0 {
		return fmt.Errorf("empty result")
	}
	if !exists[0] {
		migration := mInitMigrationsSpace(m.opts.SpaceName)
		err = migration.Migrate(m.tt, m.opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Migrator) hasConfirmedMigration(migrationID string) (bool, error) {
	var rows []migrationRow
	err := m.tt.Do(tarantool.NewSelectRequest(m.opts.SpaceName).Key([]any{migrationID}), pool.ANY).GetTyped(&rows)
	if err != nil {
		return false, err
	}
	return len(rows) > 0, nil
}

func (m *Migrator) confirmMigration(migrationID string) error {
	_, err := m.tt.Do(tarantool.NewReplaceRequest(m.opts.SpaceName).Tuple([]interface{}{
		migrationID,
		time.Now().UTC().String(),
	}),
		pool.ANY,
	).Get()
	if err != nil {
		return err
	}
	return nil
}

func (m *Migrator) findLastConfirmedMigration() (*migrationRow, error) {
	var rows []migrationRow
	expr := fmt.Sprintf("return box.space.%s.index.id:max()", m.opts.SpaceName)
	err := m.tt.Do(tarantool.NewEvalRequest(expr), pool.ANY).GetTyped(&rows)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no confirmed migrations")
	}
	return &rows[0], nil
}

func (m *Migrator) rejectMigration(migrationID string) error {
	_, err := m.tt.Do(tarantool.NewDeleteRequest(m.opts.SpaceName).Key([]any{migrationID}), pool.ANY).Get()
	if err != nil {
		return err
	}
	return nil
}

func (m *Migrator) hasMigrations() bool {
	return len(m.migrations) > 0
}

func (m *Migrator) findAvailableMigration(migrationID string) (*Migration, error) {
	for _, mgr := range m.migrations {
		if mgr.ID == migrationID {
			return mgr, nil
		}
	}
	return nil, ErrMigrationIDDoesNotExist
}

func (m *Migrator) beginTransaction() error {
	if m.opts.UseTransaction {
		_, err := m.tt.Do(tarantool.NewBeginRequest(), pool.ANY).Get()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Migrator) commitTransaction() error {
	if m.opts.UseTransaction {
		_, err := m.tt.Do(tarantool.NewCommitRequest(), pool.ANY).Get()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Migrator) rollbackTransaction() error {
	if m.opts.UseTransaction {
		_, err := m.tt.Do(tarantool.NewRollbackRequest(), pool.ANY).Get()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Migrator) Migrate() error {
	if !m.hasMigrations() {
		return ErrNoMigrationsDefined
	}
	err := m.createMigrationsSpaceIfNotExists()
	if err != nil {
		return err
	}
	for _, mgr := range m.migrations {
		log.Println(fmt.Sprintf(`tarantool-migrator: Migration "%s" process started`, mgr.ID))
		err = mgr.isValidForMigrate()
		if err != nil {
			return err
		}
		exists, err := m.hasConfirmedMigration(mgr.ID)
		if err != nil {
			return err
		}
		if !exists || m.opts.WithoutConfirm {
			startedAt := time.Now().UTC()
			err = m.beginTransaction()
			if err != nil {
				return err
			}
			errTx := mgr.Migrate(m.tt, m.opts)
			if errTx != nil {
				err = m.rollbackTransaction()
				if err != nil {
					return err
				}
				return errTx
			}
			migratedAt := time.Now().UTC().Sub(startedAt)
			if !m.opts.WithoutConfirm {
				errTx = m.confirmMigration(mgr.ID)
				if errTx != nil {
					err = m.rollbackTransaction()
					if err != nil {
						return err
					}
					return errTx
				}
			}
			err = m.commitTransaction()
			if err != nil {
				return err
			}

			log.Println(fmt.Sprintf(`tarantool-migrator: Migration "%s" successfully migrated in %.3fms`, mgr.ID, float64(migratedAt.Nanoseconds())/1e6))
		} else {
			log.Println(fmt.Sprintf(`tarantool-migrator: Migration "%s" is already migrated`, mgr.ID))
		}
	}
	return nil
}

func (m *Migrator) RollbackLast() error {
	if !m.hasMigrations() {
		return ErrNoMigrationsDefined
	}
	mgr, err := m.findLastConfirmedMigration()
	if err != nil {
		return err
	}
	log.Println(fmt.Sprintf(`tarantool-migrator: Migration "%s" found for rollback`, mgr.ID))
	migration, err := m.findAvailableMigration(mgr.ID)
	if err != nil {
		return err
	}
	err = migration.isValidForRollback()
	if err != nil {
		return err
	}
	startedAt := time.Now().UTC()
	err = migration.Rollback(m.tt, m.opts)
	if err != nil {
		return err
	}
	rolledAt := time.Now().UTC().Sub(startedAt)
	err = m.rejectMigration(migration.ID)
	if err != nil {
		return err
	}
	log.Println(fmt.Sprintf(`tarantool-migrator: Migration "%s" successfully rollbacked in %.3fms`, mgr.ID, float64(rolledAt.Nanoseconds())/1e6))
	return nil
}
