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
	file fs.DirEntry
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
	mgrParts := strings.Split(file.Name(), ".")
	if len(mgrParts) < 3 {
		return nil, ErrWrongMigrationFileFormat
	}
	if mgrParts[1] != MigrationFileSuffixUp && mgrParts[1] != MigrationFileSuffixDown {
		return nil, ErrWrongMigrationCmdFormat
	}

	return &MigrationFile{
		path: path + "/" + file.Name(),
		file: file,
		name: mgrParts[0],
		cmd:  mgrParts[1],
	}, nil
}
