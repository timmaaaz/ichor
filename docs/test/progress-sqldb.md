# Progress Summary: sqldb.md

## Overview
Ichor's database abstraction layer. Thin wrapper over sqlx providing parameterized query helpers, connection config, and sentinel errors.

## sqldb [sdk] — `business/sdk/sqldb/sqldb.go`

**Responsibility:** Database connection management and query execution abstraction.

### Key Facts
- **Most imported SDK package** — NamedQuerySlice alone has 129 call sites across 75 db store files (verified 2026-03-09)
- **Thin wrapper over sqlx** — parameterized query helpers + connection config + sentinel errors
- **Production-grade PostgreSQL support** — TLS, connection pooling, schema support

### Config Struct
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

func Open(cfg Config) (*sqlx.DB, error)
```

## QueryHelpers [sdk]

**Three main query execution patterns.**

```go
// Returns nil (NOT ErrDBNotFound) when result set is empty.
func NamedQuerySlice[T any](ctx context.Context, log *logger.Logger, db sqlx.ExtContext, query string, data any, dest *[]T) error

// Returns ErrDBNotFound when no row matches.
func NamedQueryStruct(ctx context.Context, log *logger.Logger, db sqlx.ExtContext, query string, data any, dest any) error

// Execute with no result scanning (INSERT/UPDATE/DELETE that don't return rows).
func ExecContext(ctx context.Context, log *logger.Logger, db sqlx.ExtContext, query string) error
```

### Key Distinctions
- **NamedQuerySlice** — returns `nil` for empty result sets (NOT ErrDBNotFound)
- **NamedQueryStruct** — returns `ErrDBNotFound` when single-row query finds no match
- **ExecContext** — for mutations without result scanning

## SentinelErrors [sdk]

**Database-specific error constants.**

```go
var ErrDBNotFound        = sql.ErrNoRows          // single-row query found no match
var ErrDBDuplicatedEntry = errors.New("duplicated entry")
var ErrUndefinedTable    = errors.New("undefined table")
var ErrForeignKeyViolation = errors.New("foreign key violation")
```

## Change Patterns

### ⚠ NamedQuerySlice Returns nil — NOT ErrDBNotFound

Critical distinction that affects all db store implementations:

**Pattern to avoid:**
```go
var items []Item
err := sqldb.NamedQuerySlice(ctx, log, db, query, data, &items)
if err == sqldb.ErrDBNotFound { ... }  // ❌ WRONG — NamedQuerySlice never returns this
```

**Correct pattern:**
```go
var items []Item
err := sqldb.NamedQuerySlice(ctx, log, db, query, data, &items)
if err != nil { ... }           // handle errors
if items == nil { ... }         // handle empty result (not an error)
```

Affected files:
- `business/domain/{area}/{entity}bus/stores/{entity}db/{entity}db.go` — Query method
- `business/sdk/workflow/temporal/stores/edgedb/edgedb.go` — LoadActions, LoadEdges

### ⚠ Adding a New DB Store
Affects 3 areas:
1. `business/domain/{area}/{entity}bus/stores/{entity}db/{entity}db.go` — call NamedQuerySlice/NamedQueryStruct/ExecContext
2. `business/sdk/dbtest/dbtest.go` — BusDomain wire-up for test store instantiation
3. `api/cmd/services/ichor/build/all/all.go` — pass *sqlx.DB to store constructor

### ⚠ Changing Config Fields
Affects 4 areas:
1. `business/sdk/sqldb/sqldb.go` — Config struct definition
2. `api/cmd/services/ichor/main.go` — config parsing via conf/v3
3. `zarf/k8s/` — K8s manifests and secrets
4. `Makefile` — dev database targets
5. **Note:** Environment variables must follow `ICHOR_DB_*` pattern

## Critical Points
- **NamedQuerySlice returns nil for empty results** — NOT an error, handle explicitly
- **NamedQueryStruct returns ErrDBNotFound for missing rows** — different behavior from Slice
- **All queries are parameterized** — no string interpolation, injection-safe
- **Connection pooling configured** — MaxIdleConns and MaxOpenConns tunable
- **TLS enabled by default** — DisableTLS=true only for dev environments

## Notes for Future Development
The sqldb package is foundational and used across the entire codebase. Changes should be minimal and carefully tested:
- Adding new query helpers (rare, requires widespread impact analysis)
- Changing Config fields (moderate, requires coordination across dev/test/prod)
- Fixing error handling patterns (low-risk if consistent)

The nil vs. ErrDBNotFound distinction is subtle but critical — it's the source of many bugs if not handled correctly.
