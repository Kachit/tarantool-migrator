package tarantool_migrator

import (
	"embed"
	"strings"
)

type EmbedFsLoader struct {
	fs embed.FS
}

func (fl *EmbedFsLoader) LoadMigrations(path string) (MigrationsCollection, error) {
	files, err := fl.fs.ReadDir(path)
	if err != nil {
		return nil, err
	}

	coll := make(MigrationsCollection)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		mgrParts := strings.Split(file.Name(), ".")
		if len(mgrParts) < 3 {
			return nil, ErrWrongMigrationFileFormat
		}

		if mgrParts[1] != MigrationFileSuffixUp && mgrParts[1] != MigrationFileSuffixDown {
			return nil, ErrWrongMigrationCmdFormat
		}
		var migration *Migration
		migration, ok := coll[mgrParts[0]]
		if !ok {
			migration = &Migration{
				ID: mgrParts[0],
			}
			coll[mgrParts[0]] = migration
		}

		fileData, err := fl.fs.ReadFile(path + "/" + file.Name())
		if err != nil {
			return nil, err
		}
		if mgrParts[1] == MigrationFileSuffixUp {
			migration.Migrate = NewGenericMigrateFunction(string(fileData))
		} else {
			migration.Rollback = NewGenericMigrateFunction(string(fileData))
		}
	}
	return coll, nil
}

func NewEmbedFsLoader(fs embed.FS) *EmbedFsLoader {
	return &EmbedFsLoader{fs}
}
