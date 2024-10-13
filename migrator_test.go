package tarantool_migrator

import (
	"context"
	"fmt"
	"github.com/kachit/tarantool-migrator/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/pool"
	"github.com/tarantool/go-tarantool/v2/test_helpers"
	"reflect"
	"testing"
)

type MigratorTestSuite struct {
	suite.Suite
	ctx           context.Context
	mock          *mocks.PoolerMock
	tupleResponse *test_helpers.MockResponse
	testable      *Migrator
}

func (suite *MigratorTestSuite) SetupTest() {
	suite.mock = &mocks.PoolerMock{}
	suite.ctx = context.Background()
	suite.tupleResponse = test_helpers.NewMockResponse(suite.T(), newMigrationTupleStubResponseBody())
	suite.testable = NewMigrator(suite.mock, nil, WithLogger(SilentLogger), WithOptions(&Options{
		MigrationsSpace: "migrations",
		ReadMode:        pool.ANY,
		WriteMode:       pool.RW,
	}))
}

func (suite *MigratorTestSuite) TestChangeLogger() {
	reqLg := reflect.ValueOf(suite.testable.logger).Elem()
	lg := reqLg.Interface().(logger)
	assert.Equal(suite.T(), LogLevelSilent, lg.LogLevel)
	fn := WithLogger(DebugLogger)
	fn(suite.testable)

	reqLg = reflect.ValueOf(suite.testable.logger).Elem()
	lg = reqLg.Interface().(logger)
	assert.Equal(suite.T(), LogLevelDebug, lg.LogLevel)
}

func (suite *MigratorTestSuite) TestChangeOptions() {
	assert.False(suite.T(), suite.testable.opts.TransactionsEnabled)
	opts := &Options{}
	opts.TransactionsEnabled = true
	fn := WithOptions(opts)
	fn(suite.testable)
	assert.True(suite.T(), suite.testable.opts.TransactionsEnabled)
}

func (suite *MigratorTestSuite) TestMigrateWithoutMigrations() {
	err := suite.testable.Migrate(suite.ctx)
	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "no defined migrations", err.Error())
	assert.Len(suite.T(), calls, 0)
}

func (suite *MigratorTestSuite) TestMigrateCreateMigrationSpaceError() {
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		fmt.Errorf("tarantool error"),
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}

	migrations := make(MigrationsCollection, 0)
	migrations = append(migrations, &Migration{
		ID: "init-migrations-space-error",
	})
	suite.testable.migrations = migrations
	err := suite.testable.Migrate(suite.ctx)
	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), `init migrations space error: tarantool error`, err.Error())
	assert.Len(suite.T(), calls, 1)
}

func (suite *MigratorTestSuite) TestMigrateMigrationWithoutFunction() {
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		test_helpers.NewMockResponse(suite.T(), [][]interface{}{}),
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}

	migrations := make(MigrationsCollection, 0)
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
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		test_helpers.NewMockResponse(suite.T(), [][]interface{}{}),
		suite.tupleResponse,
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}

	migrations := make(MigrationsCollection, 0)
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

func (suite *MigratorTestSuite) TestMigrateMigrationMigrateError() {
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		test_helpers.NewMockResponse(suite.T(), [][]interface{}{}),
		test_helpers.NewMockResponse(suite.T(), [][]interface{}{}),
		fmt.Errorf("tarantool error"),
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}

	migrations := make(MigrationsCollection, 0)
	migrations = append(migrations, &Migration{
		ID:      "migrate-error",
		Migrate: NewGenericMigrateFunction("box.info"),
	})
	suite.testable.migrations = migrations
	err := suite.testable.Migrate(suite.ctx)
	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), `migration "migrate-error" error: tarantool error`, err.Error())
	assert.Len(suite.T(), calls, 3)
}

func (suite *MigratorTestSuite) TestMigrateMigrationInDriveRunMode() {
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		test_helpers.NewMockResponse(suite.T(), [][]interface{}{}),
		test_helpers.NewMockResponse(suite.T(), [][]interface{}{}),
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}

	migrations := make(MigrationsCollection, 0)
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

func TestMigratorTestSuite(t *testing.T) {
	suite.Run(t, new(MigratorTestSuite))
}
