---
name: go-unit-testing
description: Write Go unit tests for existing code in this project. Use whenever asked to "add tests", "write tests for X", "test this function", "add coverage", "cover this method", or any request that involves adding test coverage to existing Go code — even if the user doesn't say "unit test" explicitly.
---

# Go Unit Testing (Test-After)

You're writing tests for existing code. Read the code under test first, identify every branch and error path, then produce tests that hit all of them.

## Decide on structure

**Use `testify/suite`** when testing a struct with multiple methods or when you need shared setup:

```go
type FooTestSuite struct {
    suite.Suite
    ctx      context.Context
    mock     *mocks.PoolerMock
    testable *Foo
}

func (suite *FooTestSuite) SetupTest() {
    suite.mock = &mocks.PoolerMock{}
    suite.ctx = context.Background()
    suite.testable = NewFoo(suite.mock, &Options{
        MigrationsSpace: "migrations",
        ReadMode:        pool.ModeAny,
        WriteMode:       pool.ModeRW,
    })
}

func TestFooTestSuite(t *testing.T) {
    suite.Run(t, new(FooTestSuite))
}
```

**Use table-driven tests** for pure/stateless functions (no external dependencies, no mocks):

```go
func TestParseSomething(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid", "foo.up.lua", "up", false},
        {"no extension", "foo", "", true},
        {"wrong cmd", "foo.test.lua", "", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseSomething(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

## Mock setup

Assign `DoFunc` per test method, not in `SetupTest` — each test controls its own responses:

```go
func (suite *FooTestSuite) TestDoSomethingSuccess() {
    mockDoer := test_helpers.NewMockDoer(suite.T())
    mockDoer.AddResponseRaw([][]interface{}{})           // empty success
    mockDoer.AddResponseRaw(newMigrationTupleStubResponseBody()) // tuple result
    suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
        return mockDoer.Do(req)
    }
    // ...
}

func (suite *FooTestSuite) TestDoSomethingError() {
    mockDoer := test_helpers.NewMockDoer(suite.T())
    mockDoer.AddResponseError(fmt.Errorf("tarantool error"))
    suite.mock.DoFunc = func(req tarantool.Request, mode pool.Mode) tarantool.Future {
        return mockDoer.Do(req)
    }
    // ...
}
```

Responses are consumed in order — first `AddResponse*` call matches the first `Do()` call, and so on.

## What to assert

Always verify both the **return value** and the **interactions**:

```go
calls := suite.mock.DoCalls()
assert.NoError(suite.T(), err)
assert.Len(suite.T(), calls, 2)                          // correct number of DB calls
assert.Equal(suite.T(), pool.ModeAny, calls[0].Mode)    // read uses ModeAny
assert.Equal(suite.T(), pool.ModeRW, calls[1].Mode)     // write uses ModeRW
```

To verify the request type and fields, use `reflect.ValueOf`:

```go
reqRef := reflect.ValueOf(calls[0].Req)
req := reqRef.Interface().(tarantool.SelectRequest)
assert.IsType(suite.T(), tarantool.SelectRequest{}, req)
assert.Equal(suite.T(), iproto.IPROTO_SELECT, req.Type())
spaceField := reqRef.FieldByName("space")
assert.Equal(suite.T(), "migrations", fmt.Sprintf("%v", spaceField))
keyField := reqRef.FieldByName("key")
assert.Equal(suite.T(), "[myid]", fmt.Sprintf("%v", keyField))
```

Request types map to iproto constants: `SelectRequest` → `IPROTO_SELECT`, `InsertRequest` → `IPROTO_INSERT`, `DeleteRequest` → `IPROTO_DELETE`, `EvalRequest` → `IPROTO_EVAL`.

## What to cover

For every function/method, write tests for:

- **Happy path** — normal success case
- **Each error path** — every `if err != nil` branch in the implementation
- **No-op / early return** — e.g., `DryRun` flag, empty collection
- **Not found** — queries returning empty results (empty tuples slice)

## Naming

- Suite struct: `TypeNameTestSuite`
- Test methods: `Test<MethodName><Scenario>` — e.g., `TestMigrateSuccess`, `TestMigrateNoMigrations`, `TestHasAppliedMigrationError`
- Runner: `func TestTypeNameTestSuite(t *testing.T) { suite.Run(t, new(TypeNameTestSuite)) }`

## Run and verify

```bash
go test ./... -v -run TestYourSuite
```

Check coverage:
```bash
go test -coverprofile=coverage.out ./... && go tool cover -func coverage.out
```
