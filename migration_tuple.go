package tarantool_migrator

import (
	"fmt"
	"time"
)

type migrationTuple struct {
	ID         string
	ExecutedAt string
}

func newMigrationTupleStubResponseBody() [][]interface{} {
	ts := time.Now().UTC()
	migration := migrationTuple{
		ID:         fmt.Sprintf("%d", ts.Unix()),
		ExecutedAt: ts.String(),
	}

	return [][]interface{}{
		{
			migration.ID,
			migration.ExecutedAt,
		},
	}
}
