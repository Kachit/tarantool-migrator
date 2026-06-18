package tarantool_migrator

import (
	"context"
	"fmt"
	"testing"

	"github.com/kachit/tarantool-migrator/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tarantool/go-tarantool/v3"
	"github.com/tarantool/go-tarantool/v3/pool"
	"github.com/tarantool/go-tarantool/v3/test_helpers"
)

type MigratorTestSuite struct {
	suite.Suite
	ctx      context.Context
	mock     *mocks.PoolerMock
	testable *Migrator
}

func (suite *MigratorTestSuite) SetupTest() {
	suite.mock = &mocks.PoolerMock{}
	suite.ctx = context.Background()
	suite.testable = NewMigrator(suite.mock, nil, WithLogger(SilentLogger), WithOptions(&Options{
		MigrationsSpace: "migrations",
		ReadMode:        pool.ModeAny,
		WriteMode:       pool.ModeRW,
	}))
}

func (suite *MigratorTestSuite) TestChangeLogger() {
	assert.Equal(suite.T(), SilentLogger, suite.testable.logger)
	fn := WithLogger(DebugLogger)
	fn(suite.testable)
	assert.Equal(suite.T(), DebugLogger, suite.testable.logger)
}

func (suite *MigratorTestSuite) TestChangeOptions() {
	assert.False(suite.T(), suite.testable.opts.DryRun)
	opts := &Options{}
	opts.DryRun = true
	fn := WithOptions(opts)
	fn(suite.testable)
	assert.True(suite.T(), suite.testable.opts.DryRun)
}

func (suite *MigratorTestSuite) TestMigrateWithoutMigrations() {
	err := suite.testable.Migrate(suite.ctx)
	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "no defined migrations", err.Error())
	assert.Len(suite.T(), calls, 0)
}

func (suite *MigratorTestSuite) TestMigrateCreateMigrationSpaceError() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseError(fmt.Errorf("tarantool error"))
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}

	migrations := make(MigrationsCollection, 0, 1)
	migrations = append(migrations, &Migration{
		ID: "init-migrations-space-error",
	})
	suite.testable.migrations = migrations
	err := suite.testable.Migrate(suite.ctx)
	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), `init migrations space error: exec create migrations space: tarantool error`, err.Error())
	assert.Len(suite.T(), calls, 1)
}

func (suite *MigratorTestSuite) TestMigrateMigrationWithoutFunction() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}

	migrations := make(MigrationsCollection, 0, 1)
	migrations = append(migrations, &Migration{
		ID: "missing-migrate-function",
	})
	suite.testable.migrations = migrations
	err := suite.testable.Migrate(suite.ctx)
	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), `migration "missing-migrate-function" error: missing migrate function in migration`,
		err.Error())
	assert.Len(suite.T(), calls, 1)
}

func (suite *MigratorTestSuite) TestMigrateMigrationIsAlreadyApplied() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseRaw(newMigrationTupleStubResponseBody())
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}

	migrations := make(MigrationsCollection, 0, 1)
	migrations = append(migrations, &Migration{
		ID:      "migration-already-applied",
		Migrate: NewGenericMigrateFunction("box.info"),
	})
	suite.testable.migrations = migrations
	err := suite.testable.Migrate(suite.ctx)
	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), calls, 2)
}

func (suite *MigratorTestSuite) TestMigrateMigrationHasAppliedError() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseError(fmt.Errorf("tarantool error"))
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}

	migrations := make(MigrationsCollection, 0, 1)
	migrations = append(migrations, &Migration{
		ID:      "migration-has-applied-error",
		Migrate: NewGenericMigrateFunction("box.info"),
	})
	suite.testable.migrations = migrations
	err := suite.testable.Migrate(suite.ctx)
	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), `migration "migration-has-applied-error" error: check applied migration: tarantool error`, err.Error())
	assert.Len(suite.T(), calls, 2)
}

func (suite *MigratorTestSuite) TestMigrateMigrationMigrateError() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseError(fmt.Errorf("tarantool error"))
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}

	migrations := make(MigrationsCollection, 0, 1)
	migrations = append(migrations, &Migration{
		ID:      "migrate-error",
		Migrate: NewGenericMigrateFunction("box.info"),
	})
	suite.testable.migrations = migrations
	err := suite.testable.Migrate(suite.ctx)
	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), `migration "migrate-error" error: user migrate: tarantool error`, err.Error())
	assert.Len(suite.T(), calls, 3)
}

func (suite *MigratorTestSuite) TestMigrateSuccess() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseRaw(newMigrationTupleStubResponseBody())
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}

	migrations := make(MigrationsCollection, 0, 1)
	migrations = append(migrations, &Migration{
		ID:      "migration-success",
		Migrate: NewGenericMigrateFunction("box.info"),
	})
	suite.testable.migrations = migrations
	err := suite.testable.Migrate(suite.ctx)
	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), calls, 4)
}

func (suite *MigratorTestSuite) TestMigrateMigrationInDriveRunMode() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseRaw([][]interface{}{})
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}

	migrations := make(MigrationsCollection, 0, 1)
	migrations = append(migrations, &Migration{
		ID:      "migrate-in-drive-run-mode",
		Migrate: NewGenericMigrateFunction("box.info"),
	})
	suite.testable.migrations = migrations
	suite.testable.opts.DryRun = true
	err := suite.testable.Migrate(suite.ctx)
	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), calls, 2)
}

func (suite *MigratorTestSuite) TestRollbackLastWithoutMigrations() {
	err := suite.testable.RollbackLast(suite.ctx)
	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "no defined migrations", err.Error())
	assert.Len(suite.T(), calls, 0)
}

func (suite *MigratorTestSuite) TestRollbackMigrationFindLastError() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseError(fmt.Errorf("tarantool error"))
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}

	migrations := make(MigrationsCollection, 0, 1)
	migrations = append(migrations, &Migration{
		ID:       "find-last-error",
		Rollback: NewGenericMigrateFunction("box.info"),
	})
	suite.testable.migrations = migrations
	err := suite.testable.RollbackLast(suite.ctx)
	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), `find applied migration error: find last applied migration: tarantool error`,
		err.Error())
	assert.Len(suite.T(), calls, 1)
}

func (suite *MigratorTestSuite) TestRollbackMigrationNotExists() {
	body := newMigrationTupleStubResponseBody()
	migrationId := fmt.Sprintf("%v", body[0][0])

	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw(body)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}

	migrations := make(MigrationsCollection, 0, 1)
	migrations = append(migrations, &Migration{
		ID:       "not-exists",
		Rollback: NewGenericMigrateFunction("box.info"),
	})
	suite.testable.migrations = migrations
	err := suite.testable.RollbackLast(suite.ctx)
	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), fmt.Sprintf(`migration "%s" error: tried to migrate to an ID that doesn't exist`,
		migrationId),
		err.Error())
	assert.Len(suite.T(), calls, 1)
}

func (suite *MigratorTestSuite) TestRollbackMigrationWithoutRollbackFunction() {
	body := newMigrationTupleStubResponseBody()
	migrationId := fmt.Sprintf("%v", body[0][0])

	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw(body)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}

	migrations := make(MigrationsCollection, 0, 1)
	migrations = append(migrations, &Migration{
		ID: migrationId,
	})
	suite.testable.migrations = migrations
	err := suite.testable.RollbackLast(suite.ctx)
	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), fmt.Sprintf(`migration "%s" error: missing rollback function in migration`,
		migrationId),
		err.Error())
	assert.Len(suite.T(), calls, 1)
}

func (suite *MigratorTestSuite) TestRollbackMigrationRollbackError() {
	body := newMigrationTupleStubResponseBody()
	migrationId := fmt.Sprintf("%v", body[0][0])

	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw(body)
	mockDoer.AddResponseError(fmt.Errorf("tarantool error"))
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}

	migrations := make(MigrationsCollection, 0, 1)
	migrations = append(migrations, &Migration{
		ID:       migrationId,
		Rollback: NewGenericMigrateFunction("box.info"),
	})
	suite.testable.migrations = migrations
	err := suite.testable.RollbackLast(suite.ctx)
	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), fmt.Sprintf(`migration "%s" error: user rollback: tarantool error`,
		migrationId),
		err.Error())
	assert.Len(suite.T(), calls, 2)
}

func (suite *MigratorTestSuite) TestRollbackMigrationSuccess() {
	body := newMigrationTupleStubResponseBody()
	migrationId := fmt.Sprintf("%v", body[0][0])

	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw(body)
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseRaw([][]interface{}{})
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}

	migrations := make(MigrationsCollection, 0, 1)
	migrations = append(migrations, &Migration{
		ID:       migrationId,
		Rollback: NewGenericMigrateFunction("box.info"),
	})
	suite.testable.migrations = migrations
	err := suite.testable.RollbackLast(suite.ctx)
	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), calls, 3)
}

func (suite *MigratorTestSuite) TestRollbackMigrationInDriveRunMode() {
	body := newMigrationTupleStubResponseBody()
	migrationId := fmt.Sprintf("%v", body[0][0])

	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw(body)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}

	migrations := make(MigrationsCollection, 0, 1)
	migrations = append(migrations, &Migration{
		ID:       migrationId,
		Rollback: NewGenericMigrateFunction("box.info"),
	})
	suite.testable.migrations = migrations
	suite.testable.opts.DryRun = true
	err := suite.testable.RollbackLast(suite.ctx)
	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), calls, 1)
}

func TestMigratorTestSuite(t *testing.T) {
	suite.Run(t, new(MigratorTestSuite))
}
