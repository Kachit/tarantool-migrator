package tarantool_migrator

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type EmbedFsLoaderTestSuite struct {
	suite.Suite
	testable *EmbedFsLoader
}

func (suite *EmbedFsLoaderTestSuite) SetupTest() {
	suite.testable = NewEmbedFsLoader(LuaFs)
}

func (suite *EmbedFsLoaderTestSuite) TestLoadMigrationsInvalidWrongMigrationsPath() {
	result, err := suite.testable.LoadMigrations("stubs")
	assert.Nil(suite.T(), result)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "open stubs: file does not exist", err.Error())
}

func (suite *EmbedFsLoaderTestSuite) TestLoadMigrationsInvalidWrongMigrationFilename() {
	result, err := suite.testable.LoadMigrations("lua/stubs/invalid-wrong-migration-filename")
	assert.Nil(suite.T(), result)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "wrong migration file format", err.Error())
}

func (suite *EmbedFsLoaderTestSuite) TestLoadMigrationsInvalidWrongMigrationCmd() {
	result, err := suite.testable.LoadMigrations("lua/stubs/invalid-wrong-migration-cmd")
	assert.Nil(suite.T(), result)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "wrong migration cmd format", err.Error())
}

func (suite *EmbedFsLoaderTestSuite) TestLoadMigrationsValidEmptyDir() {
	result, err := suite.testable.LoadMigrations("lua/stubs")
	assert.Empty(suite.T(), result)
	assert.NoError(suite.T(), err)
}

func (suite *EmbedFsLoaderTestSuite) TestLoadMigrationsValid() {
	result, err := suite.testable.LoadMigrations("lua/stubs/valid")
	assert.NotEmpty(suite.T(), result)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Contains(suite.T(), result, "202410082345_test_migration_1")
	assert.Contains(suite.T(), result, "202410091201_test_migration_2")
}

func TestEmbedFsLoaderTestSuite(t *testing.T) {
	suite.Run(t, new(EmbedFsLoaderTestSuite))
}
