package tarantool_migrator

import (
	"fmt"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/pool"
)

func mInitMigrationsSpace(space string) *Migration {
	return &Migration{
		ID: "init_migrations_space",
		Migrate: func(tt *pool.ConnectionPool, opts *Options) error {
			createSpaceQuery := fmt.Sprintf("box.schema.space.create('%s');", space)
			_, err := tt.Do(tarantool.NewEvalRequest(createSpaceQuery), pool.ANY).Get()
			if err != nil {
				return err
			}

			formatSpaceQuery := fmt.Sprintf(`
box.space.%s:format({
    {'id',type='string'},
    {'executed_at',type='string'},
})
`, space)
			_, err = tt.Do(tarantool.NewEvalRequest(formatSpaceQuery), pool.ANY).Get()
			if err != nil {
				return err
			}

			addIndexesQuery := fmt.Sprintf(`
box.space.%s:create_index('id', {parts = {'id'}, unique = true})
`, space)
			_, err = tt.Do(tarantool.NewEvalRequest(addIndexesQuery), pool.ANY).Get()
			if err != nil {
				return err
			}
			return nil
		},
		Rollback: func(tt *pool.ConnectionPool, opts *Options) error {
			dropSpaceQuery := fmt.Sprintf("box.space.%s:drop();", space)
			_, err := tt.Do(tarantool.NewEvalRequest(dropSpaceQuery), pool.ANY).Get()
			if err != nil {
				return err
			}
			return nil
		},
	}
}
