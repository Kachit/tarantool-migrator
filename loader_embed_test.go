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

func (suite *EmbedFsLoaderTestSuite) TestLoadMigrationsWrongMigrationsPath() {
	result, err := suite.testable.LoadMigrations("stubs-foo")
	assert.Nil(suite.T(), result)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "open stubs-foo: file does not exist", err.Error())
}

func TestEmbedFsLoaderTestSuite(t *testing.T) {
	suite.Run(t, new(EmbedFsLoaderTestSuite))
}
