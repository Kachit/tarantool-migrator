package tarantool_migrator

import (
	"fmt"
	"time"

	"github.com/tarantool/go-tarantool/v3/datetime"
)

func newMigrationTupleStubResponseBody() [][]interface{} {
	ts := time.Now().UTC()
	dt, _ := datetime.NewDatetime(ts)

	return [][]interface{}{
		{
			fmt.Sprintf("%d", ts.Unix()),
			dt,
		},
	}
}
