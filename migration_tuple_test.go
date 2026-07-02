package tarantool_migrator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tarantool/go-tarantool/v3/datetime"
)

type MigrationTupleTestSuite struct {
	suite.Suite
	testable *migrationTuple
}

func (suite *MigrationTupleTestSuite) SetupTest() {
	suite.testable = newMigrationTuple("1234_qwerty")
}

func (suite *MigrationTupleTestSuite) TestToSlice() {
	result := suite.testable.ToSlice()

	tm := result[1].(datetime.Datetime)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), suite.testable.ID, result[0])
	assert.False(suite.T(), tm.ToTime().IsZero())
}

func TestMigrationTupleTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationTupleTestSuite))
}
