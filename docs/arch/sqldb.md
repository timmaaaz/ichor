# sqldb

[sdk]=shared SDK [db]=store layer
→=depends on

---

## Overview

Thin wrapper over sqlx providing parameterized query helpers, connection config,
and sentinel errors used by every [db] store layer. Most imported SDK package:
185+ importers.

---

## sqldb [sdk]

file: business/sdk/sqldb/sqldb.go

### Config

```go
type Config struct {
    User         string
    Password     string
    Host         string
    Name         string
    Schema       string
    MaxIdleConns int
    MaxOpenConns int
    DisableTLS   bool
}
```

### Open

```go
func Open(cfg Config) (*sqlx.DB, error)
```

---

## Query Helpers

```go
// Returns nil (NOT ErrDBNotFound) when result set is empty.
func NamedQuerySlice[T any](ctx context.Context, log *logger.Logger, db sqlx.ExtContext, query string, data any, dest *[]T) error

// Returns ErrDBNotFound when no row matches.
func NamedQueryStruct(ctx context.Context, log *logger.Logger, db sqlx.ExtContext, query string, data any, dest any) error

// Execute with no result scanning (INSERT/UPDATE/DELETE that don't return rows).
func ExecContext(ctx context.Context, log *logger.Logger, db sqlx.ExtContext, query string) error
```

---

## Sentinel Errors

```go
var ErrDBNotFound        = sql.ErrNoRows          // single-row query found no match
var ErrDBDuplicatedEntry = errors.New("duplicated entry")
var ErrUndefinedTable    = errors.New("undefined table")
var ErrForeignKeyViolation = errors.New("foreign key violation")
```

---

## ⚠ NamedQuerySlice returns nil — NOT ErrDBNotFound

Every [db] store layer that calls NamedQuerySlice must handle nil slice as "no results"
rather than as an error. Do NOT check for ErrDBNotFound after NamedQuerySlice.

Affected pattern (every file matching this path):
  business/domain/{area}/{entity}bus/stores/{entity}db/{entity}db.go   (Query method)
  business/sdk/workflow/temporal/stores/edgedb/edgedb.go               (LoadActions, LoadEdges)

## ⚠ Adding a new DB store

  business/domain/{area}/{entity}bus/stores/{entity}db/{entity}db.go   (call NamedQuerySlice/NamedQueryStruct/ExecContext)
  business/sdk/dbtest/dbtest.go                                         (BusDomain wire-up — new store instantiation)
  api/cmd/services/ichor/build/all/all.go                               (pass *sqlx.DB to store constructor)

## ⚠ Changing Config fields

  business/sdk/sqldb/sqldb.go                     (Config struct)
  api/cmd/services/ichor/main.go                  (config parsing via conf/v3)
  zarf/k8s/ (K8s manifests / secrets)             (env vars must match ICHOR_DB_* pattern)
  Makefile                                         (dev database targets)
