package tarantool_migrator

import (
	"context"
	"embed"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/pool"
)

//go:embed lua
var LuaFs embed.FS

func NewLuaMigration(id string, upRequest string, downRequest string) *Migration {
	return &Migration{
		ID: id,
		Migrate: func(tt pool.Pooler, ctx context.Context, opts *Options) error {
			_, err := tt.Do(tarantool.NewEvalRequest(upRequest), opts.WriteMode).Get()
			if err != nil {
				return err
			}
			return nil
		},
		Rollback: func(tt pool.Pooler, ctx context.Context, opts *Options) error {
			_, err := tt.Do(tarantool.NewEvalRequest(downRequest), opts.WriteMode).Get()
			if err != nil {
				return err
			}
			return nil
		},
	}
}
