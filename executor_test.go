package tarantool_migrator

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/kachit/tarantool-migrator/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tarantool/go-iproto"
	"github.com/tarantool/go-tarantool/v3"
	"github.com/tarantool/go-tarantool/v3/pool"
	"github.com/tarantool/go-tarantool/v3/test_helpers"
)

type ExecutorBaseTestSuite struct {
	suite.Suite
	ctx      context.Context
	mock     *mocks.PoolerMock
	testable *executorBase
}

func (suite *ExecutorBaseTestSuite) SetupTest() {
	suite.mock = &mocks.PoolerMock{}
	suite.ctx = context.Background()
	suite.testable = &executorBase{
		tt: suite.mock,
		opts: &Options{
			MigrationsSpace: "migrations",
			ReadMode:        pool.ModeAny,
			WriteMode:       pool.ModeRW,
		},
	}
}

func (suite *ExecutorBaseTestSuite) TestCreateMigrationsSpaceIfNotExistsSuccess() {
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

func (suite *ExecutorBaseTestSuite) TestCreateMigrationsSpaceIfNotExistsWrongMigrationsPathError() {
	err := suite.testable.createMigrationsSpaceIfNotExists(suite.ctx, "lua/migrations/create_migrations_space.up.dua")
	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "read lua script: open lua/migrations/create_migrations_space.up.dua: file does not exist", err.Error())
	assert.Len(suite.T(), calls, 0)
}

func (suite *ExecutorBaseTestSuite) TestCreateMigrationsSpaceIfNotExistsError() {
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
	assert.Equal(suite.T(), "exec create migrations space: tarantool error", err.Error())
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

func (suite *ExecutorBaseTestSuite) TestHasAppliedMigrationFound() {
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

func (suite *ExecutorBaseTestSuite) TestHasAppliedMigrationNotFound() {
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

func (suite *ExecutorBaseTestSuite) TestHasAppliedMigrationError() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseError(fmt.Errorf("tarantool error"))
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	found, err := suite.testable.hasAppliedMigration(suite.ctx, "qwerty")

	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "check applied migration: tarantool error", err.Error())
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

func (suite *ExecutorBaseTestSuite) TestFindLastAppliedMigrationFound() {
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

func (suite *ExecutorBaseTestSuite) TestFindLastAppliedMigrationNotFound() {
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

func (suite *ExecutorBaseTestSuite) TestFindLastAppliedMigrationError() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseError(fmt.Errorf("tarantool error"))
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	result, err := suite.testable.findLastAppliedMigration(suite.ctx)

	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "find last applied migration: tarantool error", err.Error())
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

func (suite *ExecutorBaseTestSuite) TestInsertMigrationSuccess() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw(newMigrationTupleStubResponseBody())
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.insertMigration(suite.ctx, suite.mock, "qwerty")

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

func (suite *ExecutorBaseTestSuite) TestInsertMigrationError() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseError(fmt.Errorf("tarantool error"))
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.insertMigration(suite.ctx, suite.mock, "qwerty")

	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "insert migration record: tarantool error", err.Error())
	assert.Len(suite.T(), calls, 1)
	assert.Equal(suite.T(), pool.ModeRW, calls[0].Mode)

	reqRef := reflect.ValueOf(calls[0].Req)
	req := reqRef.Interface().(tarantool.InsertRequest)
	assert.IsType(suite.T(), tarantool.InsertRequest{}, req)
	assert.Equal(suite.T(), iproto.IPROTO_INSERT, req.Type())
	spaceField := reqRef.FieldByName("space")
	assert.Equal(suite.T(), "migrations", fmt.Sprintf("%v", spaceField))
}

func (suite *ExecutorBaseTestSuite) TestDeleteMigrationSuccess() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.deleteMigration(suite.ctx, suite.mock, "qwerty")

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

func (suite *ExecutorBaseTestSuite) TestDeleteMigrationError() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseError(fmt.Errorf("tarantool error"))
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	err := suite.testable.deleteMigration(suite.ctx, suite.mock, "qwerty")

	calls := suite.mock.DoCalls()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "delete migration record: tarantool error", err.Error())
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

func TestExecutorBaseTestSuite(t *testing.T) {
	suite.Run(t, new(ExecutorBaseTestSuite))
}
