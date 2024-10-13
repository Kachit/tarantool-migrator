package tarantool_migrator

import (
	"context"
	"github.com/kachit/tarantool-migrator/mocks"
	"github.com/stretchr/testify/suite"
	"github.com/tarantool/go-tarantool/v2/test_helpers"
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
	suite.testable = NewMigrator(suite.mock, nil, WithLogger(SilentLogger))
}

//func (suite *MigratorTestSuite) TestChangeLogger() {
//	ref := reflect.ValueOf(suite.testable).Elem()
//	loggerField := ref.FieldByName("logger")
//	lg := loggerField.Interface().(logger)
//	fmt.Println(lg)
//}

func TestMigratorTestSuite(t *testing.T) {
	suite.Run(t, new(MigratorTestSuite))
}
