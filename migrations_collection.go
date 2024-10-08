package tarantool_migrator

type MigrationsCollection map[string]*Migration

func (m *MigrationsCollection) IsEmpty() bool {
	return len(*m) == 0
}

func (m *MigrationsCollection) Add(migration *Migration) *MigrationsCollection {
	mm := *m
	mm[migration.ID] = migration
	return m
}

func (m *MigrationsCollection) Find(migrationID string) (*Migration, error) {
	for _, mgr := range *m {
		if mgr.ID == migrationID {
			return mgr, nil
		}
	}
	return nil, ErrMigrationIDDoesNotExist
}
