package tarantool_migrator

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/fs"
	"testing"
)

type MigrationFileTestSuite struct {
	suite.Suite
}

func (suite *MigrationFileTestSuite) TestNewMigrationFileDownValid() {
	files, _ := LuaFs.ReadDir("lua/stubs/valid")
	var file fs.DirEntry
	for _, f := range files {
		if f.Name() == "202410082345_test_migration_1.down.lua" {
			file = f
		}
	}
	result, err := NewMigrationFile("lua/stubs/valid", file)
	assert.NotEmpty(suite.T(), result)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "202410082345_test_migration_1", result.GetName())
	assert.Equal(suite.T(), "down", result.GetCmd())
	assert.Equal(suite.T(), "lua/stubs/valid/202410082345_test_migration_1.down.lua", result.GetPath())
}

func (suite *MigrationFileTestSuite) TestNewMigrationFileUpValid() {
	files, _ := LuaFs.ReadDir("lua/stubs/valid")
	var file fs.DirEntry
	for _, f := range files {
		if f.Name() == "202410082345_test_migration_1.up.lua" {
			file = f
		}
	}
	result, err := NewMigrationFile("lua/stubs/valid", file)
	assert.NotEmpty(suite.T(), result)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "202410082345_test_migration_1", result.GetName())
	assert.Equal(suite.T(), "up", result.GetCmd())
	assert.Equal(suite.T(), "lua/stubs/valid/202410082345_test_migration_1.up.lua", result.GetPath())
}

func (suite *MigrationFileTestSuite) TestNewMigrationFileInvalidMigrationFileFormat() {
	files, _ := LuaFs.ReadDir("lua/stubs/invalid-wrong-migration-filename")
	result, err := NewMigrationFile("lua/stubs/invalid-wrong-migration-filename", files[0])
	assert.Empty(suite.T(), result)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "wrong migration file format", err.Error())
}

func (suite *MigrationFileTestSuite) TestNewMigrationFileInvalidMigrationFileCmd() {
	files, _ := LuaFs.ReadDir("lua/stubs/invalid-wrong-migration-cmd")
	result, err := NewMigrationFile("lua/stubs/invalid-wrong-migration-cmd", files[0])
	assert.Empty(suite.T(), result)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "wrong migration cmd format", err.Error())
}

func TestMigrationFileTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationFileTestSuite))
}
