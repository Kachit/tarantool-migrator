package tarantool_migrator

import (
	"io/fs"
)

const MigrationFileSuffixUp = "up"
const MigrationFileSuffixDown = "down"

type MigrationFile struct {
	path string
	file fs.DirEntry
}
