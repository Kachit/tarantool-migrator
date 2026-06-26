package tarantool_migrator

import (
	"time"

	"github.com/tarantool/go-tarantool/v3/datetime"
)

type migrationTuple struct {
	ID         string
	ExecutedAt datetime.Datetime
}

func (m *migrationTuple) ToSlice() []any {
	return []any{
		m.ID,
		m.ExecutedAt,
	}
}

func newMigrationTuple(migrationID string) *migrationTuple {
	dt, _ := datetime.NewDatetime(time.Now().UTC())

	return &migrationTuple{
		ID:         migrationID,
		ExecutedAt: dt,
	}
}
