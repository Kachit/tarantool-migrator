package tarantool_migrator

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type MigrationTestSuite struct {
	suite.Suite
	testable *Migration
}

func (suite *MigrationTestSuite) SetupTest() {
	suite.testable = &Migration{}
}

func (suite *MigrationTestSuite) TestIsValidForMigrateValid() {
	suite.testable.ID = "test-1"
	suite.testable.Migrate = NewGenericMigrateFunction("foo")
	err := suite.testable.isValidForMigrate()
	assert.NoError(suite.T(), err)
}

func (suite *MigrationTestSuite) TestIsValidForMigrateInvalid() {
	err := suite.testable.isValidForMigrate()
	assert.Error(suite.T(), err)
	assert.IsType(suite.T(), ErrMissingID, err)
	assert.Equal(suite.T(), "missing ID in migration", err.Error())
	suite.testable.ID = "test-2"
	err = suite.testable.isValidForMigrate()
	assert.Error(suite.T(), err)
	assert.IsType(suite.T(), ErrMissingMigrateFunc, err)
	assert.Equal(suite.T(), "missing migrate function in migration", err.Error())
}

func (suite *MigrationTestSuite) TestIsValidForRollbackValid() {
	suite.testable.ID = "test-3"
	suite.testable.Rollback = NewGenericMigrateFunction("foo")
	err := suite.testable.isValidForRollback()
	assert.NoError(suite.T(), err)
}

func (suite *MigrationTestSuite) TestIsValidForRollbackInvalid() {
	err := suite.testable.isValidForRollback()
	assert.Error(suite.T(), err)
	assert.IsType(suite.T(), ErrMissingID, err)
	assert.Equal(suite.T(), "missing ID in migration", err.Error())
	suite.testable.ID = "test-4"
	err = suite.testable.isValidForRollback()
	assert.Error(suite.T(), err)
	assert.IsType(suite.T(), ErrMissingRollbackFunc, err)
	assert.Equal(suite.T(), "missing rollback function in migration", err.Error())
}

func TestMigrationTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationTestSuite))
}
