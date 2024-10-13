package tarantool_migrator

import (
	"context"
	"fmt"
	"github.com/kachit/tarantool-migrator/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tarantool/go-iproto"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/pool"
	"github.com/tarantool/go-tarantool/v2/test_helpers"
	"reflect"
	"strings"
	"testing"
)

type ExecutorTestSuite struct {
	suite.Suite
	ctx           context.Context
	mock          *mocks.PoolerMock
	tupleResponse *test_helpers.MockResponse
	testable      *Executor
}

func (suite *ExecutorTestSuite) SetupTest() {
	suite.mock = &mocks.PoolerMock{}
	suite.ctx = context.Background()
	suite.tupleResponse = test_helpers.NewMockResponse(suite.T(), newMigrationTupleStubResponseBody())
	suite.testable = newExecutor(suite.mock, &Options{
		MigrationsSpace: "migrations",
		ReadMode:        pool.ANY,
		WriteMode:       pool.RW,
	})
}

func (suite *ExecutorTestSuite) TestApplyMigrationInDryRunMode() {
	suite.testable.opts.DryRun = true

	err := suite.testable.applyMigration(suite.ctx, &Migration{
		ID: "migration-apply-dry-run",
	})
	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), calls, 0)
}

func (suite *ExecutorTestSuite) TestApplyMigrationWithInsertError() {
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		test_helpers.NewMockResponse(suite.T(), [][]interface{}{}),
		fmt.Errorf("tarantool error"),
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.applyMigration(suite.ctx, &Migration{
		ID:      "migration-with-insert-error",
		Migrate: NewGenericMigrateFunction("box.info"),
	})
	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "tarantool error", err.Error())
	assert.Len(suite.T(), calls, 2)
	assert.Equal(suite.T(), pool.RW, calls[0].Mode)
	assert.Equal(suite.T(), pool.RW, calls[1].Mode)

	migrateReqRef := reflect.ValueOf(calls[0].Req).Elem()
	migrateReq := migrateReqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, migrateReq)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, migrateReq.Type())
	exprField := migrateReqRef.FieldByName("expr")
	assert.Equal(suite.T(), "box.info", exprField.String())
	argsField := migrateReqRef.FieldByName("args")
	assert.Equal(suite.T(), "[]", fmt.Sprintf("%v", argsField))

	insertReqRef := reflect.ValueOf(calls[1].Req).Elem()
	insertReq := insertReqRef.Interface().(tarantool.InsertRequest)
	assert.IsType(suite.T(), tarantool.InsertRequest{}, insertReq)
	assert.Equal(suite.T(), iproto.IPROTO_INSERT, insertReq.Type())
	spaceField := insertReqRef.FieldByName("space")
	assert.Equal(suite.T(), "migrations", fmt.Sprintf("%v", spaceField))
}

func (suite *ExecutorTestSuite) TestApplyMigrationSuccessful() {
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		test_helpers.NewMockResponse(suite.T(), [][]interface{}{}),
		suite.tupleResponse,
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.applyMigration(suite.ctx, &Migration{
		ID:      "apply-migration-successful",
		Migrate: NewGenericMigrateFunction("box.info"),
	})
	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), calls, 2)
	assert.Equal(suite.T(), pool.RW, calls[0].Mode)
	assert.Equal(suite.T(), pool.RW, calls[1].Mode)

	migrateReqRef := reflect.ValueOf(calls[0].Req).Elem()
	migrateReq := migrateReqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, migrateReq)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, migrateReq.Type())
	exprField := migrateReqRef.FieldByName("expr")
	assert.Equal(suite.T(), "box.info", exprField.String())
	argsField := migrateReqRef.FieldByName("args")
	assert.Equal(suite.T(), "[]", fmt.Sprintf("%v", argsField))

	insertReqRef := reflect.ValueOf(calls[1].Req).Elem()
	insertReq := insertReqRef.Interface().(tarantool.InsertRequest)
	assert.IsType(suite.T(), tarantool.InsertRequest{}, insertReq)
	assert.Equal(suite.T(), iproto.IPROTO_INSERT, insertReq.Type())
	spaceField := insertReqRef.FieldByName("space")
	assert.Equal(suite.T(), "migrations", fmt.Sprintf("%v", spaceField))
}

func (suite *ExecutorTestSuite) TestRollbackMigrationInDryRunMode() {
	suite.testable.opts.DryRun = true

	err := suite.testable.rollbackMigration(suite.ctx, &Migration{
		ID: "migration-rollback-dry-run",
	})
	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), calls, 0)
}

func (suite *ExecutorTestSuite) TestRollbackMigrationWithDeleteError() {
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		test_helpers.NewMockResponse(suite.T(), [][]interface{}{}),
		fmt.Errorf("tarantool error"),
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.rollbackMigration(suite.ctx, &Migration{
		ID:       "migration-with-delete-error",
		Rollback: NewGenericMigrateFunction("box.info"),
	})
	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "tarantool error", err.Error())
	assert.Len(suite.T(), calls, 2)
	assert.Equal(suite.T(), pool.RW, calls[0].Mode)
	assert.Equal(suite.T(), pool.RW, calls[1].Mode)

	migrateReqRef := reflect.ValueOf(calls[0].Req).Elem()
	migrateReq := migrateReqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, migrateReq)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, migrateReq.Type())
	exprField := migrateReqRef.FieldByName("expr")
	assert.Equal(suite.T(), "box.info", exprField.String())
	argsField := migrateReqRef.FieldByName("args")
	assert.Equal(suite.T(), "[]", fmt.Sprintf("%v", argsField))

	reqRef := reflect.ValueOf(calls[1].Req).Elem()
	req := reqRef.Interface().(tarantool.DeleteRequest)
	assert.IsType(suite.T(), tarantool.DeleteRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_DELETE, req.Type())
	spaceField := reqRef.FieldByName("space")
	assert.Equal(suite.T(), "migrations", fmt.Sprintf("%v", spaceField))
	keyField := reqRef.FieldByName("key")
	assert.Equal(suite.T(), "[migration-with-delete-error]", fmt.Sprintf("%v", keyField))
}

func (suite *ExecutorTestSuite) TestRollbackMigrationSuccessful() {
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		test_helpers.NewMockResponse(suite.T(), [][]interface{}{}),
		test_helpers.NewMockResponse(suite.T(), [][]interface{}{}),
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.rollbackMigration(suite.ctx, &Migration{
		ID:       "migration-rollback-success",
		Rollback: NewGenericMigrateFunction("box.info"),
	})
	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), calls, 2)
	assert.Equal(suite.T(), pool.RW, calls[0].Mode)
	assert.Equal(suite.T(), pool.RW, calls[1].Mode)

	migrateReqRef := reflect.ValueOf(calls[0].Req).Elem()
	migrateReq := migrateReqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, migrateReq)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, migrateReq.Type())
	exprField := migrateReqRef.FieldByName("expr")
	assert.Equal(suite.T(), "box.info", exprField.String())
	argsField := migrateReqRef.FieldByName("args")
	assert.Equal(suite.T(), "[]", fmt.Sprintf("%v", argsField))

	reqRef := reflect.ValueOf(calls[1].Req).Elem()
	req := reqRef.Interface().(tarantool.DeleteRequest)
	assert.IsType(suite.T(), tarantool.DeleteRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_DELETE, req.Type())
	spaceField := reqRef.FieldByName("space")
	assert.Equal(suite.T(), "migrations", fmt.Sprintf("%v", spaceField))
	keyField := reqRef.FieldByName("key")
	assert.Equal(suite.T(), "[migration-rollback-success]", fmt.Sprintf("%v", keyField))
}

func (suite *ExecutorTestSuite) TestHasAppliedMigrationFound() {
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		suite.tupleResponse,
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}
	found, err := suite.testable.hasAppliedMigration(suite.ctx, "qwerty")

	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), found)
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.ANY, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req).Elem()
	req := reqRef.Interface().(tarantool.SelectRequest)
	assert.IsType(suite.T(), tarantool.SelectRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_SELECT, req.Type())
	spaceField := reqRef.FieldByName("space")
	assert.Equal(suite.T(), "migrations", fmt.Sprintf("%v", spaceField))
	keyField := reqRef.FieldByName("key")
	assert.Equal(suite.T(), "[qwerty]", fmt.Sprintf("%v", keyField))
	indexField := reqRef.FieldByName("index")
	assert.True(suite.T(), indexField.IsNil())
}

func (suite *ExecutorTestSuite) TestHasAppliedMigrationNotFound() {
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		test_helpers.NewMockResponse(suite.T(), [][]interface{}{}),
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}
	found, err := suite.testable.hasAppliedMigration(suite.ctx, "qwerty")

	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), found)
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.ANY, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req).Elem()
	req := reqRef.Interface().(tarantool.SelectRequest)
	assert.IsType(suite.T(), tarantool.SelectRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_SELECT, req.Type())
	spaceField := reqRef.FieldByName("space")
	assert.Equal(suite.T(), "migrations", fmt.Sprintf("%v", spaceField))
	keyField := reqRef.FieldByName("key")
	assert.Equal(suite.T(), "[qwerty]", fmt.Sprintf("%v", keyField))
	indexField := reqRef.FieldByName("index")
	assert.True(suite.T(), indexField.IsNil())
}

func (suite *ExecutorTestSuite) TestHasAppliedMigrationError() {
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		fmt.Errorf("tarantool error"),
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}
	found, err := suite.testable.hasAppliedMigration(suite.ctx, "qwerty")

	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "tarantool error", err.Error())
	assert.False(suite.T(), found)
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.ANY, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req).Elem()
	req := reqRef.Interface().(tarantool.SelectRequest)
	assert.IsType(suite.T(), tarantool.SelectRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_SELECT, req.Type())
	spaceField := reqRef.FieldByName("space")
	assert.Equal(suite.T(), "migrations", fmt.Sprintf("%v", spaceField))
	keyField := reqRef.FieldByName("key")
	assert.Equal(suite.T(), "[qwerty]", fmt.Sprintf("%v", keyField))
	indexField := reqRef.FieldByName("index")
	assert.True(suite.T(), indexField.IsNil())
}

func (suite *ExecutorTestSuite) TestFindLastAppliedMigrationFound() {
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		suite.tupleResponse,
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}
	result, err := suite.testable.findLastAppliedMigration(suite.ctx)

	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), result)
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.ANY, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req).Elem()
	req := reqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, req.Type())
	exprField := reqRef.FieldByName("expr")
	assert.Equal(suite.T(), "return box.space.migrations.index.id:max()", exprField.String())
	argsField := reqRef.FieldByName("args")
	assert.Equal(suite.T(), "[]", fmt.Sprintf("%v", argsField))
}

func (suite *ExecutorTestSuite) TestFindLastAppliedMigrationNotFound() {
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		test_helpers.NewMockResponse(suite.T(), [][]interface{}{}),
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}
	result, err := suite.testable.findLastAppliedMigration(suite.ctx)

	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "no applied migrations", err.Error())
	assert.Empty(suite.T(), result)
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.ANY, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req).Elem()
	req := reqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, req.Type())
	exprField := reqRef.FieldByName("expr")
	assert.Equal(suite.T(), "return box.space.migrations.index.id:max()", exprField.String())
	argsField := reqRef.FieldByName("args")
	assert.Equal(suite.T(), "[]", fmt.Sprintf("%v", argsField))
}

func (suite *ExecutorTestSuite) TestFindLastAppliedMigrationError() {
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		fmt.Errorf("tarantool error"),
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}
	result, err := suite.testable.findLastAppliedMigration(suite.ctx)

	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "tarantool error", err.Error())
	assert.Empty(suite.T(), result)
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.ANY, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req).Elem()
	req := reqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, req.Type())
	exprField := reqRef.FieldByName("expr")
	assert.Equal(suite.T(), "return box.space.migrations.index.id:max()", exprField.String())
	argsField := reqRef.FieldByName("args")
	assert.Equal(suite.T(), "[]", fmt.Sprintf("%v", argsField))
}

func (suite *ExecutorTestSuite) TestCreateMigrationsSpaceIfNotExistsSuccess() {
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		test_helpers.NewMockResponse(suite.T(), [][]interface{}{}),
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.createMigrationsSpaceIfNotExists(suite.ctx, createMigrationsSpacePath)

	data, _ := LuaFs.ReadFile("lua/migrations/create_migrations_space.up.lua")
	migrationSpaceRequest := string(data)
	migrationSpaceRequest = strings.ReplaceAll(migrationSpaceRequest, "_migrations_space_", "migrations")

	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.RW, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req).Elem()
	req := reqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, req.Type())
	exprField := reqRef.FieldByName("expr")
	assert.Equal(suite.T(), migrationSpaceRequest, exprField.String())
	argsField := reqRef.FieldByName("args")
	assert.Equal(suite.T(), "[]", fmt.Sprintf("%v", argsField))
}

func (suite *ExecutorTestSuite) TestCreateMigrationsSpaceIfNotExistsWrongMigrationsPathError() {
	err := suite.testable.createMigrationsSpaceIfNotExists(suite.ctx, "lua/migrations/create_migrations_space.up.dua")
	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "open lua/migrations/create_migrations_space.up.dua: file does not exist", err.Error())
	assert.Len(suite.T(), calls, 0)
}

func (suite *ExecutorTestSuite) TestCreateMigrationsSpaceIfNotExistsError() {
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		fmt.Errorf("tarantool error"),
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.createMigrationsSpaceIfNotExists(suite.ctx, createMigrationsSpacePath)

	data, _ := LuaFs.ReadFile("lua/migrations/create_migrations_space.up.lua")
	migrationSpaceRequest := string(data)
	migrationSpaceRequest = strings.ReplaceAll(migrationSpaceRequest, "_migrations_space_", "migrations")

	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "tarantool error", err.Error())
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.RW, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req).Elem()
	req := reqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, req.Type())
	exprField := reqRef.FieldByName("expr")
	assert.Equal(suite.T(), migrationSpaceRequest, exprField.String())
	argsField := reqRef.FieldByName("args")
	assert.Equal(suite.T(), "[]", fmt.Sprintf("%v", argsField))
}

func (suite *ExecutorTestSuite) TestInsertMigrationSuccess() {
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		suite.tupleResponse,
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.insertMigration(suite.ctx, "qwerty")

	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.RW, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req).Elem()
	req := reqRef.Interface().(tarantool.InsertRequest)
	assert.IsType(suite.T(), tarantool.InsertRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_INSERT, req.Type())
	spaceField := reqRef.FieldByName("space")
	assert.Equal(suite.T(), "migrations", fmt.Sprintf("%v", spaceField))
}

func (suite *ExecutorTestSuite) TestInsertMigrationError() {
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		fmt.Errorf("tarantool error"),
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.insertMigration(suite.ctx, "qwerty")

	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "tarantool error", err.Error())
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.RW, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req).Elem()
	req := reqRef.Interface().(tarantool.InsertRequest)
	assert.IsType(suite.T(), tarantool.InsertRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_INSERT, req.Type())
	spaceField := reqRef.FieldByName("space")
	assert.Equal(suite.T(), "migrations", fmt.Sprintf("%v", spaceField))
}

func (suite *ExecutorTestSuite) TestDeleteMigrationSuccess() {
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		test_helpers.NewMockResponse(suite.T(), [][]interface{}{}),
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.deleteMigration(suite.ctx, "qwerty")

	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.RW, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req).Elem()
	req := reqRef.Interface().(tarantool.DeleteRequest)
	assert.IsType(suite.T(), tarantool.DeleteRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_DELETE, req.Type())
	spaceField := reqRef.FieldByName("space")
	assert.Equal(suite.T(), "migrations", fmt.Sprintf("%v", spaceField))
	keyField := reqRef.FieldByName("key")
	assert.Equal(suite.T(), "[qwerty]", fmt.Sprintf("%v", keyField))
}

func (suite *ExecutorTestSuite) TestDeleteMigrationError() {
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		fmt.Errorf("tarantool error"),
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.deleteMigration(suite.ctx, "qwerty")

	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "tarantool error", err.Error())
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.RW, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req).Elem()
	req := reqRef.Interface().(tarantool.DeleteRequest)
	assert.IsType(suite.T(), tarantool.DeleteRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_DELETE, req.Type())
	spaceField := reqRef.FieldByName("space")
	assert.Equal(suite.T(), "migrations", fmt.Sprintf("%v", spaceField))
	keyField := reqRef.FieldByName("key")
	assert.Equal(suite.T(), "[qwerty]", fmt.Sprintf("%v", keyField))
}

func (suite *ExecutorTestSuite) TestNewGenericMigrateFunctionSuccessExecute() {
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		test_helpers.NewMockResponse(suite.T(), [][]interface{}{}),
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}
	mgrFunc := NewGenericMigrateFunction("box.info")
	err := mgrFunc(suite.ctx, suite.mock, suite.testable.opts)
	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.RW, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req).Elem()
	req := reqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, req.Type())
	exprField := reqRef.FieldByName("expr")
	assert.Equal(suite.T(), "box.info", exprField.String())
}

func (suite *ExecutorTestSuite) TestNewGenericMigrateFunctionErrorExecute() {
	mockDoer := test_helpers.NewMockDoer(suite.T(),
		fmt.Errorf("tarantool error"),
	)
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) *tarantool.Future {
		return mockDoer.Do(req)
	}
	mgrFunc := NewGenericMigrateFunction("box.info")
	err := mgrFunc(suite.ctx, suite.mock, suite.testable.opts)
	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "tarantool error", err.Error())
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.RW, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req).Elem()
	req := reqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, req.Type())
	exprField := reqRef.FieldByName("expr")
	assert.Equal(suite.T(), "box.info", exprField.String())
}

func TestExecutorTestSuite(t *testing.T) {
	suite.Run(t, new(ExecutorTestSuite))
}
