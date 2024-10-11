package tarantool_migrator

import (
	"fmt"
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

func (suite *EmbedFsLoaderTestSuite) TestLoadMigrationsValidEmptyDir() {
	result, err := suite.testable.LoadMigrations("lua/stubs")
	assert.Empty(suite.T(), result)
	assert.NoError(suite.T(), err)
	fmt.Println(result)
}

func (suite *EmbedFsLoaderTestSuite) TestLoadMigrationsValid() {
	result, err := suite.testable.LoadMigrations("lua/stubs/valid")
	assert.NotEmpty(suite.T(), result)
	assert.NoError(suite.T(), err)
}

func TestEmbedFsLoaderTestSuite(t *testing.T) {
	suite.Run(t, new(EmbedFsLoaderTestSuite))
}
