package tarantool_migrator

type Migrations map[string]*Migration

func (m *Migrations) IsEmpty() bool {
	return len(*m) == 0
}

func (m *Migrations) Find(migrationID string) (*Migration, error) {
	for _, mgr := range *m {
		if mgr.ID == migrationID {
			return mgr, nil
		}
	}
	return nil, ErrMigrationIDDoesNotExist
}
