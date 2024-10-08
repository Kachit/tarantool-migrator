package tarantool_migrator

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type MigrationsCollectionTestSuite struct {
	suite.Suite
	testable MigrationsCollection
}

func (suite *MigrationsCollectionTestSuite) SetupTest() {
	suite.testable = make(MigrationsCollection)
}

func (suite *MigrationsCollectionTestSuite) TestIsEmpty() {
	migration := &Migration{ID: "test"}
	assert.True(suite.T(), suite.testable.IsEmpty())
	suite.testable.Add(migration)
	assert.False(suite.T(), suite.testable.IsEmpty())
}

func (suite *MigrationsCollectionTestSuite) TestFindFound() {
	id := "test"
	migration := &Migration{ID: id}
	suite.testable.Add(migration)
	result, err := suite.testable.Find(id)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), result)
	assert.Equal(suite.T(), id, result.ID)
}

func (suite *MigrationsCollectionTestSuite) TestFindNotFound() {
	id := "test"
	migration := &Migration{ID: id}
	suite.testable.Add(migration)
	result, err := suite.testable.Find("test-2")
	assert.Error(suite.T(), err)
	assert.IsType(suite.T(), ErrMigrationIDDoesNotExist, err)
	assert.Equal(suite.T(), "tried to migrate to an ID that doesn't exist", err.Error())
	assert.Empty(suite.T(), result)
}

func TestMigrationsCollectionTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationsCollectionTestSuite))
}
