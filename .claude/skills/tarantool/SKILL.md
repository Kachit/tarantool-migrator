---
name: tarantool
description: Reference for the go-tarantool v3 API patterns used in this project. Use whenever writing, reading, or modifying code that touches pool.Pooler, tarantool requests, Lua eval, mock responses, or the local Tarantool dev setup. Also use when looking up pool modes, request builder chains, Future consumption, or iproto constants for test assertions.
---

# Tarantool Go Client Reference (v3)

Library: `github.com/tarantool/go-tarantool/v3`

## Pool setup

```go
dialer := tarantool.NetDialer{
    Address:  "127.0.0.1:3301",
    User:     "migrator",
    Password: "migrator-pwd",
}
opts := tarantool.Opts{Timeout: time.Second}
instance := pool.Instance{Name: "instance001", Dialer: dialer, Opts: opts}
tt, err := pool.New(ctx, []pool.Instance{instance})
```

## Pool modes

| Operation | Mode |
|---|---|
| Reads (SELECT, EVAL returning data) | `pool.ModeAny` |
| Writes (INSERT, DELETE, EVAL mutating) | `pool.ModeRW` |

## Requests

All requests are built with a fluent chain. Every request gets `.Context(ctx)` before being passed to `Do`.

```go
// Lua eval
req := tarantool.NewEvalRequest("return box.info").Context(ctx)

// Select by key
req := tarantool.NewSelectRequest("migrations").Context(ctx).Key([]any{"my-id"})

// Insert a tuple
req := tarantool.NewInsertRequest("migrations").Context(ctx).Tuple([]interface{}{
    "my-id",
    time.Now().UTC().Format(time.RFC3339),
})

// Delete by key
req := tarantool.NewDeleteRequest("migrations").Context(ctx).Key([]any{"my-id"})
```

## Executing and reading results

```go
// Raw result (ignore data)
_, err := tt.Do(req, pool.ModeRW).Get()

// Typed result (decode into slice of structs)
var tuples []myStruct
err := tt.Do(req, pool.ModeAny).GetTyped(&tuples)
```

`GetTyped` decodes msgpack into the slice. An empty result sets the slice to zero length (not an error). Check `len(tuples) == 0` for "not found".

## Typed struct decoding

Fields are decoded positionally — struct field order must match the Tarantool tuple field order:

```go
type migrationTuple struct {
    ID         string
    ExecutedAt string
}
```

## Lua eval with space name substitution

The `_migrations_space_` token is a placeholder replaced at runtime:

```go
cmd := "return box.space._migrations_space_.index.id:max()"
expr := strings.ReplaceAll(cmd, "_migrations_space_", opts.MigrationsSpace)
```

The Lua file at `lua/migrations/create_migrations_space.up.lua` uses the same convention.

## iproto constants (for test assertions)

```go
import "github.com/tarantool/go-iproto"

iproto.IPROTO_EVAL    // NewEvalRequest
iproto.IPROTO_SELECT  // NewSelectRequest
iproto.IPROTO_INSERT  // NewInsertRequest
iproto.IPROTO_DELETE  // NewDeleteRequest
```

Use `req.Type()` to assert the request type in tests.

## Inspecting request fields in tests

Request structs have unexported fields — use `reflect.ValueOf` to read them:

```go
reqRef := reflect.ValueOf(calls[0].Req)
req := reqRef.Interface().(tarantool.EvalRequest)
assert.Equal(t, iproto.IPROTO_EVAL, req.Type())
exprField := reqRef.FieldByName("expr")
assert.Equal(t, "expected lua", exprField.String())
keyField := reqRef.FieldByName("key")
assert.Equal(t, "[my-id]", fmt.Sprintf("%v", keyField))
spaceField := reqRef.FieldByName("space")
assert.Equal(t, "migrations", fmt.Sprintf("%v", spaceField))
```

## Mock setup in tests

```go
mockDoer := test_helpers.NewMockDoer(suite.T())
mockDoer.AddResponseRaw([][]interface{}{})           // empty success
mockDoer.AddResponseRaw(newMigrationTupleStubResponseBody()) // tuple result
mockDoer.AddResponseError(fmt.Errorf("tarantool error"))     // error response

suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
    return mockDoer.Do(req)
}
```

Responses are consumed in the order they were added — first `AddResponse*` matches the first `Do()` call.

## Local dev

```
Image:    tarantool/tarantool:3.7
Port:     127.0.0.1:3301:3301
User:     migrator
Password: migrator-pwd
Config:   .docker/config.yaml
```

```bash
docker compose up -d    # start
cd tmp && go run .      # run testdrive against live instance
docker compose down     # stop (keeps volume)
docker compose down -v  # stop and wipe data
```
