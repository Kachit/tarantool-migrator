package tarantool_migrator

import (
	"context"
	"github.com/kachit/tarantool-migrator/mocks"
	"github.com/stretchr/testify/suite"
	"github.com/tarantool/go-tarantool/v2/test_helpers"
	"testing"
)

type ExecutorTestSuite struct {
	suite.Suite
	ctx          context.Context
	mock         *mocks.PoolerMock
	mockResponse *test_helpers.MockResponse
	testable     *Executor
}

func (suite *ExecutorTestSuite) SetupTest() {
	suite.mock = &mocks.PoolerMock{}
	suite.ctx = context.Background()
	suite.testable = newExecutor(suite.mock, DefaultOptions)
}

func TestExecutorTestSuite(t *testing.T) {
	suite.Run(t, new(ExecutorTestSuite))
}
