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
	suite.testable = MigrationsCollection{}
}

func (suite *MigrationsCollectionTestSuite) TestIsEmpty() {
	migration := &Migration{ID: "test"}
	assert.True(suite.T(), suite.testable.IsEmpty())
	suite.testable = append(suite.testable, migration)
	assert.False(suite.T(), suite.testable.IsEmpty())
}

func (suite *MigrationsCollectionTestSuite) TestFindFound() {
	id := "test"
	migration := &Migration{ID: id}
	suite.testable = append(suite.testable, migration)
	result, err := suite.testable.Find(id)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), result)
	assert.Equal(suite.T(), id, result.ID)
}

func (suite *MigrationsCollectionTestSuite) TestFindNotFound() {
	id := "test"
	migration := &Migration{ID: id}
	suite.testable = append(suite.testable, migration)
	result, err := suite.testable.Find("test-2")
	assert.Error(suite.T(), err)
	assert.IsType(suite.T(), ErrMigrationIDDoesNotExist, err)
	assert.Equal(suite.T(), "tried to migrate to an ID that doesn't exist", err.Error())
	assert.Empty(suite.T(), result)
}

func (suite *MigrationsCollectionTestSuite) TestSort() {
	migration1 := &Migration{ID: "202410082345_test_migration_1"}
	migration2 := &Migration{ID: "202410091201_test_migration_2"}
	migration3 := &Migration{ID: "202410091545_test_migration_3"}
	suite.testable = append(suite.testable, migration3)
	suite.testable = append(suite.testable, migration2)
	suite.testable = append(suite.testable, migration1)
	suite.testable.sort()
	assert.Equal(suite.T(), suite.testable[0].ID, "202410082345_test_migration_1")
	assert.Equal(suite.T(), suite.testable[1].ID, "202410091201_test_migration_2")
	assert.Equal(suite.T(), suite.testable[2].ID, "202410091545_test_migration_3")
}

func TestMigrationsCollectionTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationsCollectionTestSuite))
}
