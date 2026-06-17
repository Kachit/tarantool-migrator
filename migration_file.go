package tarantool_migrator

import (
	"io/fs"
	"strings"
)

const MigrationFileSuffixUp = "up"
const MigrationFileSuffixDown = "down"
const MigrationFilePrefixExcluded = "--"

type MigrationFile struct {
	path string
	name string
	cmd  string
}

func (mf *MigrationFile) GetPath() string {
	return mf.path
}

func (mf *MigrationFile) GetName() string {
	return mf.name
}

func (mf *MigrationFile) GetCmd() string {
	return mf.cmd
}

func NewMigrationFile(path string, file fs.DirEntry) (*MigrationFile, error) {
	fileName := file.Name()
	if !strings.HasSuffix(fileName, ".lua") {
		return nil, ErrWrongMigrationFileFormat
	}

	baseName := strings.TrimSuffix(fileName, ".lua")

	lastDot := strings.LastIndex(baseName, ".")
	if lastDot < 0 {
		return nil, ErrWrongMigrationFileFormat
	}

	name := baseName[:lastDot]
	cmd := baseName[lastDot+1:]

	if cmd != MigrationFileSuffixUp && cmd != MigrationFileSuffixDown {
		return nil, ErrWrongMigrationCmdFormat
	}

	return &MigrationFile{
		path: path + "/" + fileName,
		name: name,
		cmd:  cmd,
	}, nil
}
