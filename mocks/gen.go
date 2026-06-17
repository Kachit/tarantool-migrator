package mocks

import (
	"context"
	"github.com/tarantool/go-tarantool/v3"
	"github.com/tarantool/go-tarantool/v3/pool"
	"time"
)

//go:generate moq -out pool.go . Pooler

// TopologyEditor is the interface that must be implemented by a connection pool.
// It describes edit topology methods.
type TopologyEditor interface {
	Add(ctx context.Context, instance pool.Instance) error
	Remove(name string) error
}

// Pooler is the interface that must be implemented by a connection pool.
type Pooler interface {
	TopologyEditor

	ConnectedNow(mode pool.Mode) (bool, error)
	Close() error
	CloseGraceful() error
	ConfiguredTimeout(mode pool.Mode) (time.Duration, error)
	NewPrepared(expr string, mode pool.Mode) (*tarantool.Prepared, error)
	NewStream(mode pool.Mode) (*tarantool.Stream, error)
	NewWatcher(key string, callback tarantool.WatchCallback,
		mode pool.Mode) (tarantool.Watcher, error)
	Do(req tarantool.Request, mode pool.Mode) tarantool.Future
}
