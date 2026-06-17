package tarantool_migrator

import (
	"context"
	"time"

	"github.com/tarantool/go-tarantool/v3"
	"github.com/tarantool/go-tarantool/v3/pool"
)

type streamer interface {
	Do(req tarantool.Request) tarantool.Future
}

type streamPooler struct {
	s streamer
}

func (sp *streamPooler) Do(req tarantool.Request, _ pool.Mode) tarantool.Future {
	return sp.s.Do(req)
}

func (sp *streamPooler) Add(_ context.Context, _ pool.Instance) error { return nil }

func (sp *streamPooler) Remove(_ string) error { return nil }

func (sp *streamPooler) ConnectedNow(_ pool.Mode) (bool, error) { return true, nil }

func (sp *streamPooler) Close() error { return nil }

func (sp *streamPooler) CloseGraceful() error { return nil }

func (sp *streamPooler) ConfiguredTimeout(_ pool.Mode) (time.Duration, error) {
	return 0, nil
}

func (sp *streamPooler) NewPrepared(_ string, _ pool.Mode) (*tarantool.Prepared, error) {
	return nil, nil
}

func (sp *streamPooler) NewStream(_ pool.Mode) (*tarantool.Stream, error) {
	return nil, nil
}

func (sp *streamPooler) NewWatcher(_ string, _ tarantool.WatchCallback, _ pool.Mode) (tarantool.Watcher, error) {
	return nil, nil
}
