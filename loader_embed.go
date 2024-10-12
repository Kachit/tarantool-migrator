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

	migrations := make(MigrationsCollection, 0)
	coll := make(map[string]*Migration)
	for _, file := range files {
		if file.IsDir() || strings.HasPrefix(file.Name(), MigrationFilePrefixExcluded) {
			continue
		}
		mgrFile, err := NewMigrationFile(path, file)
		if err != nil {
			return nil, err
		}
		var migration *Migration
		migration, ok := coll[mgrFile.GetName()]
		if !ok {
			migration = &Migration{
				ID: mgrFile.GetName(),
			}
			coll[mgrFile.GetName()] = migration
		}

		fileData, err := fl.fs.ReadFile(mgrFile.GetPath())
		if err != nil {
			return nil, err
		}
		if mgrFile.GetCmd() == MigrationFileSuffixUp {
			migration.Migrate = NewGenericMigrateFunction(string(fileData))
		} else {
			migration.Rollback = NewGenericMigrateFunction(string(fileData))
		}
	}

	for _, migration := range coll {
		migrations = append(migrations, migration)
	}
	migrations.sort()
	return migrations, nil
}

func NewEmbedFsLoader(fs embed.FS) *EmbedFsLoader {
	return &EmbedFsLoader{fs}
}
