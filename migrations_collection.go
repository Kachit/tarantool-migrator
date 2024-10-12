package tarantool_migrator

import "sort"

type MigrationsCollection []*Migration

func (m *MigrationsCollection) IsEmpty() bool {
	return len(*m) == 0
}

func (m *MigrationsCollection) Find(migrationID string) (*Migration, error) {
	for _, mgr := range *m {
		if mgr.ID == migrationID {
			return mgr, nil
		}
	}
	return nil, ErrMigrationIDDoesNotExist
}

func (m *MigrationsCollection) sort() {
	mm := *m
	sort.Slice(mm, func(i, j int) bool {
		return mm[i].ID < mm[j].ID
	})
}
