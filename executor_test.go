package tarantool_migrator

import (
	"context"
	"fmt"
	"github.com/kachit/tarantool-migrator/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tarantool/go-iproto"
	"github.com/tarantool/go-tarantool/v3"
	"github.com/tarantool/go-tarantool/v3/pool"
	"github.com/tarantool/go-tarantool/v3/test_helpers"
	"reflect"
	"strings"
	"testing"
)

type ExecutorTestSuite struct {
	suite.Suite
	ctx      context.Context
	mock     *mocks.PoolerMock
	testable *Executor
}

func (suite *ExecutorTestSuite) SetupTest() {
	suite.mock = &mocks.PoolerMock{}
	suite.ctx = context.Background()
	suite.testable = newExecutor(suite.mock, &Options{
		MigrationsSpace: "migrations",
		ReadMode:        pool.ModeAny,
		WriteMode:       pool.ModeRW,
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

func (suite *ExecutorTestSuite) TestApplyMigrationWithMigrateError() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseError(fmt.Errorf("tarantool error"))
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.applyMigration(suite.ctx, &Migration{
		ID:      "migration-with-migrate-error",
		Migrate: NewGenericMigrateFunction("box.info"),
	})
	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "tarantool error", err.Error())
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.ModeRW, calls[0].Mode)

	migrateReqRef := reflect.ValueOf(calls[0].Req)
	migrateReq := migrateReqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, migrateReq)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, migrateReq.Type())
	exprField := migrateReqRef.FieldByName("expr")
	assert.Equal(suite.T(), "box.info", exprField.String())
	argsField := migrateReqRef.FieldByName("args")
	assert.Equal(suite.T(), "[]", fmt.Sprintf("%v", argsField))
}

func (suite *ExecutorTestSuite) TestApplyMigrationWithInsertError() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseError(fmt.Errorf("tarantool error"))
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
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
	assert.Equal(suite.T(), pool.ModeRW, calls[0].Mode)
	assert.Equal(suite.T(), pool.ModeRW, calls[1].Mode)

	migrateReqRef := reflect.ValueOf(calls[0].Req)
	migrateReq := migrateReqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, migrateReq)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, migrateReq.Type())
	exprField := migrateReqRef.FieldByName("expr")
	assert.Equal(suite.T(), "box.info", exprField.String())
	argsField := migrateReqRef.FieldByName("args")
	assert.Equal(suite.T(), "[]", fmt.Sprintf("%v", argsField))

	insertReqRef := reflect.ValueOf(calls[1].Req)
	insertReq := insertReqRef.Interface().(tarantool.InsertRequest)
	assert.IsType(suite.T(), tarantool.InsertRequest{}, insertReq)
	assert.Equal(suite.T(), iproto.IPROTO_INSERT, insertReq.Type())
	spaceField := insertReqRef.FieldByName("space")
	assert.Equal(suite.T(), "migrations", fmt.Sprintf("%v", spaceField))
}

func (suite *ExecutorTestSuite) TestApplyMigrationSuccessful() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseRaw(newMigrationTupleStubResponseBody())
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.applyMigration(suite.ctx, &Migration{
		ID:      "apply-migration-successful",
		Migrate: NewGenericMigrateFunction("box.info"),
	})
	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), calls, 2)
	assert.Equal(suite.T(), pool.ModeRW, calls[0].Mode)
	assert.Equal(suite.T(), pool.ModeRW, calls[1].Mode)

	migrateReqRef := reflect.ValueOf(calls[0].Req)
	migrateReq := migrateReqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, migrateReq)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, migrateReq.Type())
	exprField := migrateReqRef.FieldByName("expr")
	assert.Equal(suite.T(), "box.info", exprField.String())
	argsField := migrateReqRef.FieldByName("args")
	assert.Equal(suite.T(), "[]", fmt.Sprintf("%v", argsField))

	insertReqRef := reflect.ValueOf(calls[1].Req)
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

func (suite *ExecutorTestSuite) TestRollbackMigrationWithRollbackError() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseError(fmt.Errorf("tarantool error"))
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.rollbackMigration(suite.ctx, &Migration{
		ID:       "migration-with-rollback-error",
		Rollback: NewGenericMigrateFunction("box.info"),
	})
	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "tarantool error", err.Error())
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.ModeRW, calls[0].Mode)

	migrateReqRef := reflect.ValueOf(calls[0].Req)
	migrateReq := migrateReqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, migrateReq)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, migrateReq.Type())
	exprField := migrateReqRef.FieldByName("expr")
	assert.Equal(suite.T(), "box.info", exprField.String())
	argsField := migrateReqRef.FieldByName("args")
	assert.Equal(suite.T(), "[]", fmt.Sprintf("%v", argsField))
}

func (suite *ExecutorTestSuite) TestRollbackMigrationWithDeleteError() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseError(fmt.Errorf("tarantool error"))
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
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
	assert.Equal(suite.T(), pool.ModeRW, calls[0].Mode)
	assert.Equal(suite.T(), pool.ModeRW, calls[1].Mode)

	migrateReqRef := reflect.ValueOf(calls[0].Req)
	migrateReq := migrateReqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, migrateReq)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, migrateReq.Type())
	exprField := migrateReqRef.FieldByName("expr")
	assert.Equal(suite.T(), "box.info", exprField.String())
	argsField := migrateReqRef.FieldByName("args")
	assert.Equal(suite.T(), "[]", fmt.Sprintf("%v", argsField))

	reqRef := reflect.ValueOf(calls[1].Req)
	req := reqRef.Interface().(tarantool.DeleteRequest)
	assert.IsType(suite.T(), tarantool.DeleteRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_DELETE, req.Type())
	spaceField := reqRef.FieldByName("space")
	assert.Equal(suite.T(), "migrations", fmt.Sprintf("%v", spaceField))
	keyField := reqRef.FieldByName("key")
	assert.Equal(suite.T(), "[migration-with-delete-error]", fmt.Sprintf("%v", keyField))
}

func (suite *ExecutorTestSuite) TestRollbackMigrationSuccessful() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseRaw([][]interface{}{})
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.rollbackMigration(suite.ctx, &Migration{
		ID:       "migration-rollback-success",
		Rollback: NewGenericMigrateFunction("box.info"),
	})
	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), calls, 2)
	assert.Equal(suite.T(), pool.ModeRW, calls[0].Mode)
	assert.Equal(suite.T(), pool.ModeRW, calls[1].Mode)

	migrateReqRef := reflect.ValueOf(calls[0].Req)
	migrateReq := migrateReqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, migrateReq)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, migrateReq.Type())
	exprField := migrateReqRef.FieldByName("expr")
	assert.Equal(suite.T(), "box.info", exprField.String())
	argsField := migrateReqRef.FieldByName("args")
	assert.Equal(suite.T(), "[]", fmt.Sprintf("%v", argsField))

	reqRef := reflect.ValueOf(calls[1].Req)
	req := reqRef.Interface().(tarantool.DeleteRequest)
	assert.IsType(suite.T(), tarantool.DeleteRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_DELETE, req.Type())
	spaceField := reqRef.FieldByName("space")
	assert.Equal(suite.T(), "migrations", fmt.Sprintf("%v", spaceField))
	keyField := reqRef.FieldByName("key")
	assert.Equal(suite.T(), "[migration-rollback-success]", fmt.Sprintf("%v", keyField))
}

func (suite *ExecutorTestSuite) TestHasAppliedMigrationFound() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw(newMigrationTupleStubResponseBody())
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	found, err := suite.testable.hasAppliedMigration(suite.ctx, "qwerty")

	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), found)
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.ModeAny, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req)
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
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	found, err := suite.testable.hasAppliedMigration(suite.ctx, "qwerty")

	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), found)
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.ModeAny, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req)
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
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseError(fmt.Errorf("tarantool error"))
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	found, err := suite.testable.hasAppliedMigration(suite.ctx, "qwerty")

	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "tarantool error", err.Error())
	assert.False(suite.T(), found)
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.ModeAny, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req)
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
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw(newMigrationTupleStubResponseBody())
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	result, err := suite.testable.findLastAppliedMigration(suite.ctx)

	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), result)
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.ModeAny, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req)
	req := reqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, req.Type())
	exprField := reqRef.FieldByName("expr")
	assert.Equal(suite.T(), "return box.space.migrations.index.id:max()", exprField.String())
	argsField := reqRef.FieldByName("args")
	assert.Equal(suite.T(), "[]", fmt.Sprintf("%v", argsField))
}

func (suite *ExecutorTestSuite) TestFindLastAppliedMigrationNotFound() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	result, err := suite.testable.findLastAppliedMigration(suite.ctx)

	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "no applied migrations", err.Error())
	assert.Empty(suite.T(), result)
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.ModeAny, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req)
	req := reqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, req.Type())
	exprField := reqRef.FieldByName("expr")
	assert.Equal(suite.T(), "return box.space.migrations.index.id:max()", exprField.String())
	argsField := reqRef.FieldByName("args")
	assert.Equal(suite.T(), "[]", fmt.Sprintf("%v", argsField))
}

func (suite *ExecutorTestSuite) TestFindLastAppliedMigrationError() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseError(fmt.Errorf("tarantool error"))
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	result, err := suite.testable.findLastAppliedMigration(suite.ctx)

	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "tarantool error", err.Error())
	assert.Empty(suite.T(), result)
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.ModeAny, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req)
	req := reqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, req.Type())
	exprField := reqRef.FieldByName("expr")
	assert.Equal(suite.T(), "return box.space.migrations.index.id:max()", exprField.String())
	argsField := reqRef.FieldByName("args")
	assert.Equal(suite.T(), "[]", fmt.Sprintf("%v", argsField))
}

func (suite *ExecutorTestSuite) TestCreateMigrationsSpaceIfNotExistsSuccess() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.createMigrationsSpaceIfNotExists(suite.ctx, createMigrationsSpacePath)

	data, _ := LuaFs.ReadFile("lua/migrations/create_migrations_space.up.lua")
	migrationSpaceRequest := string(data)
	migrationSpaceRequest = strings.ReplaceAll(migrationSpaceRequest, "_migrations_space_", "migrations")

	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.ModeRW, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req)
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
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseError(fmt.Errorf("tarantool error"))
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
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
	assert.Equal(suite.T(), pool.ModeRW, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req)
	req := reqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, req.Type())
	exprField := reqRef.FieldByName("expr")
	assert.Equal(suite.T(), migrationSpaceRequest, exprField.String())
	argsField := reqRef.FieldByName("args")
	assert.Equal(suite.T(), "[]", fmt.Sprintf("%v", argsField))
}

func (suite *ExecutorTestSuite) TestInsertMigrationSuccess() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw(newMigrationTupleStubResponseBody())
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.insertMigration(suite.ctx, "qwerty")

	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.ModeRW, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req)
	req := reqRef.Interface().(tarantool.InsertRequest)
	assert.IsType(suite.T(), tarantool.InsertRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_INSERT, req.Type())
	spaceField := reqRef.FieldByName("space")
	assert.Equal(suite.T(), "migrations", fmt.Sprintf("%v", spaceField))
}

func (suite *ExecutorTestSuite) TestInsertMigrationError() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseError(fmt.Errorf("tarantool error"))
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.insertMigration(suite.ctx, "qwerty")

	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "tarantool error", err.Error())
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.ModeRW, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req)
	req := reqRef.Interface().(tarantool.InsertRequest)
	assert.IsType(suite.T(), tarantool.InsertRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_INSERT, req.Type())
	spaceField := reqRef.FieldByName("space")
	assert.Equal(suite.T(), "migrations", fmt.Sprintf("%v", spaceField))
}

func (suite *ExecutorTestSuite) TestDeleteMigrationSuccess() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.deleteMigration(suite.ctx, "qwerty")

	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.ModeRW, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req)
	req := reqRef.Interface().(tarantool.DeleteRequest)
	assert.IsType(suite.T(), tarantool.DeleteRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_DELETE, req.Type())
	spaceField := reqRef.FieldByName("space")
	assert.Equal(suite.T(), "migrations", fmt.Sprintf("%v", spaceField))
	keyField := reqRef.FieldByName("key")
	assert.Equal(suite.T(), "[qwerty]", fmt.Sprintf("%v", keyField))
}

func (suite *ExecutorTestSuite) TestDeleteMigrationError() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseError(fmt.Errorf("tarantool error"))
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.deleteMigration(suite.ctx, "qwerty")

	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "tarantool error", err.Error())
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.ModeRW, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req)
	req := reqRef.Interface().(tarantool.DeleteRequest)
	assert.IsType(suite.T(), tarantool.DeleteRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_DELETE, req.Type())
	spaceField := reqRef.FieldByName("space")
	assert.Equal(suite.T(), "migrations", fmt.Sprintf("%v", spaceField))
	keyField := reqRef.FieldByName("key")
	assert.Equal(suite.T(), "[qwerty]", fmt.Sprintf("%v", keyField))
}

func (suite *ExecutorTestSuite) TestNewGenericMigrateFunctionSuccessExecute() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	mgrFunc := NewGenericMigrateFunction("box.info")
	err := mgrFunc(suite.ctx, suite.mock, suite.testable.opts)
	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.ModeRW, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req)
	req := reqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, req.Type())
	exprField := reqRef.FieldByName("expr")
	assert.Equal(suite.T(), "box.info", exprField.String())
}

func (suite *ExecutorTestSuite) TestNewGenericMigrateFunctionErrorExecute() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseError(fmt.Errorf("tarantool error"))
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	mgrFunc := NewGenericMigrateFunction("box.info")
	err := mgrFunc(suite.ctx, suite.mock, suite.testable.opts)
	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "tarantool error", err.Error())
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.ModeRW, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req)
	req := reqRef.Interface().(tarantool.EvalRequest)
	assert.IsType(suite.T(), tarantool.EvalRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_EVAL, req.Type())
	exprField := reqRef.FieldByName("expr")
	assert.Equal(suite.T(), "box.info", exprField.String())
}

func (suite *ExecutorTestSuite) TestApplyMigrationWithTransactionsEnabledNewStreamError() {
	suite.testable.opts.TransactionsEnabled = true
	suite.mock.NewStreamFunc = func(_ pool.Mode) (*tarantool.Stream, error) {
		return nil, fmt.Errorf("stream error")
	}

	err := suite.testable.applyMigration(suite.ctx, &Migration{
		ID:      "tx-new-stream-error",
		Migrate: NewGenericMigrateFunction("box.info"),
	})

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "stream error", err.Error())
}

func (suite *ExecutorTestSuite) TestRollbackMigrationWithTransactionsEnabledNewStreamError() {
	suite.testable.opts.TransactionsEnabled = true
	suite.mock.NewStreamFunc = func(_ pool.Mode) (*tarantool.Stream, error) {
		return nil, fmt.Errorf("stream error")
	}

	err := suite.testable.rollbackMigration(suite.ctx, &Migration{
		ID:       "tx-rollback-stream-error",
		Rollback: NewGenericMigrateFunction("box.info"),
	})

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "stream error", err.Error())
}

func (suite *ExecutorTestSuite) TestApplyMigrationInTxBeginError() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseError(fmt.Errorf("begin error"))

	err := suite.testable.applyMigrationInTx(suite.ctx, mockDoer, &Migration{
		ID:      "tx-begin-error",
		Migrate: NewGenericMigrateFunction("box.info"),
	})

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "begin error", err.Error())
}

func (suite *ExecutorTestSuite) TestApplyMigrationInTxMigrateError() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseError(fmt.Errorf("migrate error"))
	mockDoer.AddResponseRaw([][]interface{}{})

	err := suite.testable.applyMigrationInTx(suite.ctx, mockDoer, &Migration{
		ID:      "tx-migrate-error",
		Migrate: NewGenericMigrateFunction("box.info"),
	})

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "migrate error", err.Error())
}

func (suite *ExecutorTestSuite) TestApplyMigrationInTxInsertError() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseError(fmt.Errorf("insert error"))
	mockDoer.AddResponseRaw([][]interface{}{})

	err := suite.testable.applyMigrationInTx(suite.ctx, mockDoer, &Migration{
		ID:      "tx-insert-error",
		Migrate: NewGenericMigrateFunction("box.info"),
	})

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "insert error", err.Error())
}

func (suite *ExecutorTestSuite) TestApplyMigrationInTxCommitError() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseRaw(newMigrationTupleStubResponseBody())
	mockDoer.AddResponseError(fmt.Errorf("commit error"))
	mockDoer.AddResponseRaw([][]interface{}{})

	err := suite.testable.applyMigrationInTx(suite.ctx, mockDoer, &Migration{
		ID:      "tx-commit-error",
		Migrate: NewGenericMigrateFunction("box.info"),
	})

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "commit error", err.Error())
}

func (suite *ExecutorTestSuite) TestApplyMigrationInTxSuccess() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseRaw(newMigrationTupleStubResponseBody())
	mockDoer.AddResponseRaw([][]interface{}{})

	err := suite.testable.applyMigrationInTx(suite.ctx, mockDoer, &Migration{
		ID:      "tx-apply-success",
		Migrate: NewGenericMigrateFunction("box.info"),
	})

	assert.NoError(suite.T(), err)
}

func (suite *ExecutorTestSuite) TestRollbackMigrationInTxBeginError() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseError(fmt.Errorf("begin error"))

	err := suite.testable.rollbackMigrationInTx(suite.ctx, mockDoer, &Migration{
		ID:       "tx-rollback-begin-error",
		Rollback: NewGenericMigrateFunction("box.info"),
	})

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "begin error", err.Error())
}

func (suite *ExecutorTestSuite) TestRollbackMigrationInTxRollbackFuncError() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseError(fmt.Errorf("rollback func error"))
	mockDoer.AddResponseRaw([][]interface{}{})

	err := suite.testable.rollbackMigrationInTx(suite.ctx, mockDoer, &Migration{
		ID:       "tx-rollback-func-error",
		Rollback: NewGenericMigrateFunction("box.info"),
	})

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "rollback func error", err.Error())
}

func (suite *ExecutorTestSuite) TestRollbackMigrationInTxDeleteError() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseError(fmt.Errorf("delete error"))
	mockDoer.AddResponseRaw([][]interface{}{})

	err := suite.testable.rollbackMigrationInTx(suite.ctx, mockDoer, &Migration{
		ID:       "tx-rollback-delete-error",
		Rollback: NewGenericMigrateFunction("box.info"),
	})

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "delete error", err.Error())
}

func (suite *ExecutorTestSuite) TestRollbackMigrationInTxSuccess() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseRaw([][]interface{}{})
	mockDoer.AddResponseRaw([][]interface{}{})

	err := suite.testable.rollbackMigrationInTx(suite.ctx, mockDoer, &Migration{
		ID:       "tx-rollback-success",
		Rollback: NewGenericMigrateFunction("box.info"),
	})

	assert.NoError(suite.T(), err)
}

func TestExecutorTestSuite(t *testing.T) {
	suite.Run(t, new(ExecutorTestSuite))
}
