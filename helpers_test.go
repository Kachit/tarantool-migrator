package tarantool_migrator

import (
	"fmt"
	"time"
)

func newMigrationTupleStubResponseBody() [][]interface{} {
	ts := time.Now().UTC()

	return [][]interface{}{
		{
			fmt.Sprintf("%d", ts.Unix()),
			ts.Format(time.RFC3339),
		},
	}
}
