package tarantool_migrator

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/kachit/tarantool-migrator/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tarantool/go-iproto"
	"github.com/tarantool/go-tarantool/v3"
	"github.com/tarantool/go-tarantool/v3/pool"
	"github.com/tarantool/go-tarantool/v3/test_helpers"
)

type NoTxExecutorTestSuite struct {
	suite.Suite
	ctx      context.Context
	mock     *mocks.PoolerMock
	testable *noTxExecutor
}

func (suite *NoTxExecutorTestSuite) SetupTest() {
	suite.mock = &mocks.PoolerMock{}
	suite.ctx = context.Background()
	suite.testable = &noTxExecutor{
		executorBase: executorBase{
			tt: suite.mock,
			opts: &Options{
				MigrationsSpace: "migrations",
				ReadMode:        pool.ModeAny,
				WriteMode:       pool.ModeRW,
			},
		},
	}
}

func (suite *NoTxExecutorTestSuite) TestApplyMigrationInDryRunMode() {
	suite.testable.opts.DryRun = true

	err := suite.testable.applyMigration(suite.ctx, &Migration{
		ID: "migration-apply-dry-run",
	})
	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), calls, 0)
}

func (suite *NoTxExecutorTestSuite) TestApplyMigrationWithMigrateError() {
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
	assert.Equal(suite.T(), "user migrate: tarantool error", err.Error())
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

func (suite *NoTxExecutorTestSuite) TestApplyMigrationWithInsertError() {
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
	assert.Equal(suite.T(), "insert migration record: tarantool error", err.Error())
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

func (suite *NoTxExecutorTestSuite) TestApplyMigrationSuccessful() {
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

func (suite *NoTxExecutorTestSuite) TestRollbackMigrationInDryRunMode() {
	suite.testable.opts.DryRun = true

	err := suite.testable.rollbackMigration(suite.ctx, &Migration{
		ID: "migration-rollback-dry-run",
	})
	calls := suite.mock.DoCalls()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), calls, 0)
}

func (suite *NoTxExecutorTestSuite) TestRollbackMigrationWithRollbackError() {
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
	assert.Equal(suite.T(), "user rollback: tarantool error", err.Error())
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

func (suite *NoTxExecutorTestSuite) TestRollbackMigrationWithDeleteError() {
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
	assert.Equal(suite.T(), "delete migration record: tarantool error", err.Error())
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

func (suite *NoTxExecutorTestSuite) TestRollbackMigrationSuccessful() {
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

func (suite *NoTxExecutorTestSuite) TestNewGenericMigrateFunctionSuccessExecute() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseRaw([][]interface{}{})
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	mgrFunc := NewGenericMigrateFunction("box.info")
	err := mgrFunc(suite.ctx, suite.mock, *suite.testable.opts)
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

func (suite *NoTxExecutorTestSuite) TestNewGenericMigrateFunctionErrorExecute() {
	mockDoer := test_helpers.NewMockDoer(suite.T())
	mockDoer.AddResponseError(fmt.Errorf("tarantool error"))
	suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
		return mockDoer.Do(req)
	}
	mgrFunc := NewGenericMigrateFunction("box.info")
	err := mgrFunc(suite.ctx, suite.mock, *suite.testable.opts)
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

func TestNoTxExecutorTestSuite(t *testing.T) {
	suite.Run(t, new(NoTxExecutorTestSuite))
}
