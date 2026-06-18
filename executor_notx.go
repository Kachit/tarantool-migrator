package tarantool_migrator

import (
	"context"
	"fmt"
)

type noTxExecutor struct {
	executorBase
}

func (e *noTxExecutor) applyMigration(ctx context.Context, migration *Migration) error {
	if e.opts.DryRun {
		return nil
	}

	if err := migration.Migrate(ctx, e.tt, *e.opts); err != nil {
		return fmt.Errorf("user migrate: %w", err)
	}

	return e.insertMigration(ctx, e.tt, migration.ID)
}

func (e *noTxExecutor) rollbackMigration(ctx context.Context, migration *Migration) error {
	if e.opts.DryRun {
		return nil
	}

	if err := migration.Rollback(ctx, e.tt, *e.opts); err != nil {
		return fmt.Errorf("user rollback: %w", err)
	}

	return e.deleteMigration(ctx, e.tt, migration.ID)
}
