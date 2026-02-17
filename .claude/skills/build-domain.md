# Build Domain Skill

Automatically generates a complete, **self-validating** domain implementation from a SQL CREATE TABLE statement. Writes all files across all layers, then iteratively builds, tests, and fixes until the scaffold compiles and passes tests — or reports what couldn't be auto-resolved.

## Usage

```
/build-domain <paste your SQL CREATE TABLE statement>
```

## Overview

Unlike `/add-domain` (which is interactive and guided), `/build-domain` parses your SQL schema and **writes all files immediately** — business layer, database store, application layer, API layer, tests, wiring, and registration. It then enters an automated validation pipeline:

1. **Build loop** (max 5 iterations) — compile, catch errors, fix, rebuild
2. **Existing test validation** (max 3 iterations) — run existing tests to catch wiring breakage, fix
3. **New test validation** (max 5 iterations) — run the generated CRUD tests, fix until passing
4. **Lint check** — catch style issues

The user receives a working, tested scaffold — not a pile of files to debug.

This skill is portable to any project using the **Ardan Labs Service Starter Kit** architecture (Domain-Driven, Data-Oriented Design).

## Your Task

When the user invokes this skill with SQL, execute all steps below without stopping for confirmation. Write every file, then run the validation pipeline. The user can review and adjust after.

---

## Step 1: Parse the SQL Schema

Extract from the `CREATE TABLE` statement:

| Element | Example |
|---------|---------|
| **Schema** | `assets`, `core`, `hr`, `inventory`, `products`, `procurement`, `sales`, `config`, `workflow` |
| **Table name** | `asset_conditions`, `warehouses`, `user_assets` |
| **Columns** | name, type, constraints |
| **Primary key** | Usually `id UUID` |
| **Foreign keys** | `REFERENCES other_table(id)` |
| **Unique constraints** | `UNIQUE (name)` |
| **NOT NULL columns** | Required fields |
| **Nullable columns** | Optional fields |
| **Default values** | `DEFAULT gen_random_uuid()`, `DEFAULT TRUE` |

### Derive Names

From the parsed schema, compute:

| Derived | Rule | Example |
|---------|------|---------|
| **Entity** (singular, PascalCase) | Singularize table name | `AssetCondition`, `Warehouse`, `UserAsset` |
| **entity** (singular, camelCase) | For variables | `assetCondition`, `warehouse`, `userAsset` |
| **Bus package** | `{entity}bus` | `assetconditionbus`, `warehousebus`, `userassetbus` |
| **DB package** | `{entity}db` | `assetconditiondb`, `warehousedb`, `userassetdb` |
| **App package** | `{entity}app` | `assetconditionapp`, `warehouseapp`, `userassetapp` |
| **API package** | `{entity}api` | `assetconditionapi`, `warehouseapi`, `userassetapi` |
| **DB struct** | `db{Entity}` (private) | `dbAssetCondition` — avoids conflicts with SDK packages |
| **URL path** | kebab-case of table name | `/assets/asset-conditions`, `/inventory/warehouses` |
| **URL param** | snake_case + `_id` | `asset_condition_id`, `warehouse_id` |
| **RouteTable** | `{schema}.{table}` | `assets.asset_conditions`, `inventory.warehouses` |
| **DomainName** | lowercase no separators | `assetcondition`, `warehouse`, `userasset` |
| **EntityName** | table name (for workflow) | `asset_conditions`, `warehouses`, `user_assets` |
| **Directory area** | schema name | `assets`, `inventory`, `core` |

### Type Mapping

Map SQL types to Go types across layers:

| SQL Type | Business Model | DB Model | App Model (JSON) | Filter Type |
|----------|---------------|----------|-------------------|-------------|
| `UUID NOT NULL` | `uuid.UUID` | `uuid.UUID` | `string` | `*uuid.UUID` |
| `UUID NULL` | `uuid.UUID` | `uuid.UUID` | `string` | `*uuid.UUID` |
| `TEXT NOT NULL` | `string` | `string` | `string` | `*string` |
| `TEXT NULL` | `string` | `sql.NullString` | `string` | `*string` |
| `INTEGER/INT NOT NULL` | `int` | `int` | `int` | `*int` |
| `INTEGER NULL` | `int` | `sql.NullInt64` | `int` | `*int` |
| `BOOLEAN NOT NULL` | `bool` | `bool` | `bool` | `*bool` |
| `BOOLEAN DEFAULT x` | `bool` | `bool` | `bool` | `*bool` |
| `TIMESTAMP/TIMESTAMPTZ` | `time.Time` | `time.Time` | `string` (RFC3339) | N/A usually |
| `NUMERIC/DECIMAL` | `decimal.Decimal` | `decimal.Decimal` | `string` | N/A usually |
| `JSONB` | `json.RawMessage` | `json.RawMessage` | `json.RawMessage` | N/A |

### Classify Fields

For each column (excluding `id`):

- **Required fields** (NOT NULL, no DEFAULT): go into `NewEntity` as value types, app model gets `validate:"required"`
- **Optional fields** (NULL or has DEFAULT): go into `NewEntity` as value types but app model gets `validate:"omitempty"`
- **All non-PK fields**: go into `UpdateEntity` as pointer types
- **Filterable fields**: text and UUID fields typically get filters; booleans sometimes
- **Audit fields** (`created_by`, `updated_by`, `created_date`, `updated_date`): handled specially — set from context, not user input
- **Foreign key fields**: UUID fields referencing other tables — need FK violation error handling in store

### Detect Domain Complexity

- **Simple domain**: No foreign keys, no audit fields, no timestamps (e.g., `asset_conditions`, `tags`)
  - Conversion functions are **exported** (PascalCase): `ToBusNewEntity(app) bus.NewEntity`
  - No error return from conversions (no UUID/time parsing needed)
- **Complex domain**: Has foreign keys, audit fields, or timestamp fields (e.g., `user_assets`, `valid_assets`)
  - Conversion functions are **private** (camelCase): `toBusNewEntity(app) (bus.NewEntity, error)`
  - Return errors because UUID parsing and time parsing can fail
  - App layer Create/Update methods must handle conversion errors before calling business layer
  - TestSeed functions accept individual named `uuid.UUIDs` params per FK (not a generic `parentIDs`)

---

## Step 2: Find the Go Module Path

Read `go.mod` in the project root to get the module path (e.g., `github.com/timmaaaz/ichor`). All import paths derive from this.

---

## Step 3: Find the Migration Version

Read `business/sdk/migrate/sql/migrate.sql`, find the last `-- Version: X.YY` line, and increment for the new migration.

---

## Step 4: Write All Files

### File Order

Write files bottom-up (dependencies first):

1. Migration SQL (append to `migrate.sql`)
2. Business layer: `model.go`, `filter.go`, `order.go`, `event.go`, `{entity}bus.go`, `testutil.go`
3. Database store: `model.go`, `filter.go`, `order.go`, `{entity}db.go`
4. Application layer: `model.go`, `filter.go`, `order.go`, `{entity}app.go`
5. API layer: `{entity}api.go`, `route.go`, `filter.go`
6. Wiring edits: `all.go` (imports, bus instantiation, delegate registration, route registration)
7. Test infrastructure edits: `dbtest.go` (BusDomain field + instantiation), `apitest/model.go` (SeedData field), `tableaccessbus/testutil.go` (table access entry)
8. Test files: `{entity}_test.go`, `seed_test.go`, `create_test.go`, `update_test.go`, `delete_test.go`, `query_test.go`

---

### 4.1: Migration SQL

**Append** to `business/sdk/migrate/sql/migrate.sql`:

```sql
-- Version: {next_version}
-- Description: Create table {table_name}
{original CREATE TABLE statement}
```

---

### 4.2: Business Layer

**Directory**: `business/domain/{area}/{entity}bus/`

#### `model.go`

```go
package {entity}bus

import (
    "github.com/google/uuid"
    // Add "time" if timestamps exist
    // Add "github.com/shopspring/decimal" if money fields exist
)

// {Entity} represents a {entity} in the system.
type {Entity} struct {
    ID          uuid.UUID `json:"id"`
    // All columns as Go types with json tags (snake_case in JSON)
    // Example:
    // Name        string    `json:"name"`
    // Description string    `json:"description"`
}

// New{Entity} contains information needed to create a new {entity}.
type New{Entity} struct {
    // All required/optional fields EXCEPT id, created_date, updated_date
    // Audit fields (created_by) come from context, NOT from this struct
    // Foreign keys are uuid.UUID types
}

// Update{Entity} defines what information may be provided to modify an existing {entity}.
// All fields are pointers so we can distinguish not provided from zero value.
type Update{Entity} struct {
    // Same fields as New{Entity} but ALL as pointers
    // Example: Name *string `json:"name,omitempty"`
}
```

**Rules**:
- Business models use Go-native types: `uuid.UUID`, `time.Time`, `string`, `int`, `bool`
- JSON tags use lowercase/snake_case (e.g., `json:"id"`, `json:"name"`, `json:"created_by"`) — matches workflow event serialization
- `New*` struct: value types, excludes `id` and audit fields
- `Update*` struct: all pointer types

#### `filter.go`

```go
package {entity}bus

import "github.com/google/uuid"

// QueryFilter holds the available fields to filter by.
type QueryFilter struct {
    ID   *uuid.UUID
    // Pointer fields for each filterable column
    // Text fields: *string (for ILIKE search)
    // UUID fields: *uuid.UUID (for exact match)
    // Boolean fields: *bool (for exact match)
}
```

#### `order.go`

```go
package {entity}bus

import "github.com/{module}/business/sdk/order"

// DefaultOrderBy represents the default way results are ordered.
var DefaultOrderBy = order.NewBy(OrderByName, order.ASC)

const (
    OrderByID   = "id"
    OrderByName = "name"
    // One constant per sortable column
)
```

#### `event.go`

```go
package {entity}bus

import (
    "encoding/json"

    "github.com/google/uuid"
    "github.com/{module}/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "{domainname}"

// EntityName is the workflow entity name used for event matching.
const EntityName = "{table_name}"

const (
    ActionCreated = "created"
    ActionUpdated = "updated"
    ActionDeleted = "deleted"
)

// =============================================================================
// Created Event
// =============================================================================

type ActionCreatedParms struct {
    EntityID uuid.UUID `json:"entityID"`
    UserID   uuid.UUID `json:"userID"`
    Entity   {Entity}  `json:"entity"`
}

func (p *ActionCreatedParms) Marshal() ([]byte, error) {
    return json.Marshal(p)
}

func ActionCreatedData({entityVar} {Entity}) delegate.Data {
    params := ActionCreatedParms{
        EntityID: {entityVar}.ID,
        UserID:   uuid.Nil, // Set to actual user ID if audit fields exist
        Entity:   {entityVar},
    }

    rawParams, err := params.Marshal()
    if err != nil {
        panic(err)
    }

    return delegate.Data{
        Domain:    DomainName,
        Action:    ActionCreated,
        RawParams: rawParams,
    }
}

// =============================================================================
// Updated Event — same pattern as Created
// =============================================================================

type ActionUpdatedParms struct {
    EntityID uuid.UUID `json:"entityID"`
    UserID   uuid.UUID `json:"userID"`
    Entity   {Entity}  `json:"entity"`
}

func (p *ActionUpdatedParms) Marshal() ([]byte, error) {
    return json.Marshal(p)
}

func ActionUpdatedData({entityVar} {Entity}) delegate.Data {
    params := ActionUpdatedParms{
        EntityID: {entityVar}.ID,
        UserID:   uuid.Nil,
        Entity:   {entityVar},
    }

    rawParams, err := params.Marshal()
    if err != nil {
        panic(err)
    }

    return delegate.Data{
        Domain:    DomainName,
        Action:    ActionUpdated,
        RawParams: rawParams,
    }
}

// =============================================================================
// Deleted Event — same pattern
// =============================================================================

type ActionDeletedParms struct {
    EntityID uuid.UUID `json:"entityID"`
    UserID   uuid.UUID `json:"userID"`
    Entity   {Entity}  `json:"entity"`
}

func (p *ActionDeletedParms) Marshal() ([]byte, error) {
    return json.Marshal(p)
}

func ActionDeletedData({entityVar} {Entity}) delegate.Data {
    params := ActionDeletedParms{
        EntityID: {entityVar}.ID,
        UserID:   uuid.Nil,
        Entity:   {entityVar},
    }

    rawParams, err := params.Marshal()
    if err != nil {
        panic(err)
    }

    return delegate.Data{
        Domain:    DomainName,
        Action:    ActionDeleted,
        RawParams: rawParams,
    }
}
```

#### `{entity}bus.go`

```go
package {entity}bus

import (
    "context"
    "errors"
    "fmt"

    "github.com/google/uuid"
    "github.com/{module}/business/sdk/delegate"
    "github.com/{module}/business/sdk/order"
    "github.com/{module}/business/sdk/page"
    "github.com/{module}/business/sdk/sqldb"
    "github.com/{module}/foundation/logger"
    "github.com/{module}/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
    ErrNotFound    = errors.New("{entity} not found")
    ErrUniqueEntry = errors.New("{entity} entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
    NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
    Create(ctx context.Context, {entityVar} {Entity}) error
    Update(ctx context.Context, {entityVar} {Entity}) error
    Delete(ctx context.Context, {entityVar} {Entity}) error
    Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]{Entity}, error)
    Count(ctx context.Context, filter QueryFilter) (int, error)
    QueryByID(ctx context.Context, {entityVar}ID uuid.UUID) ({Entity}, error)
}

// Business manages the set of APIs for {entity} access.
type Business struct {
    log      *logger.Logger
    storer   Storer
    delegate *delegate.Delegate
}

// NewBusiness constructs a {entity} business API for use.
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, storer Storer) *Business {
    return &Business{
        log:      log,
        delegate: delegate,
        storer:   storer,
    }
}

// NewWithTx constructs a new business value that will use the
// specified transaction in any store related calls.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
    storer, err := b.storer.NewWithTx(tx)
    if err != nil {
        return nil, err
    }

    bus := Business{
        log:      b.log,
        delegate: b.delegate,
        storer:   storer,
    }

    return &bus, nil
}

// Create adds a new {entity} to the system.
func (b *Business) Create(ctx context.Context, n{entityAbbrev} New{Entity}) ({Entity}, error) {
    ctx, span := otel.AddSpan(ctx, "business.{entity}bus.Create")
    defer span.End()

    {entityVar} := {Entity}{
        ID: uuid.New(),
        // Map all fields from New{Entity} to {Entity}
    }

    if err := b.storer.Create(ctx, {entityVar}); err != nil {
        if errors.Is(err, ErrUniqueEntry) {
            return {Entity}{}, fmt.Errorf("create: %w", ErrUniqueEntry)
        }
        return {Entity}{}, err
    }

    if err := b.delegate.Call(ctx, ActionCreatedData({entityVar})); err != nil {
        b.log.Error(ctx, "{entity}bus: delegate call failed", "action", ActionCreated, "err", err)
    }

    return {entityVar}, nil
}

// Update updates an existing {entity}.
func (b *Business) Update(ctx context.Context, {entityVar} {Entity}, u{entityAbbrev} Update{Entity}) ({Entity}, error) {
    ctx, span := otel.AddSpan(ctx, "business.{entity}bus.Update")
    defer span.End()

    // Apply pointer-based partial updates
    // if u{entityAbbrev}.Name != nil { {entityVar}.Name = *u{entityAbbrev}.Name }
    // Repeat for each field...

    if err := b.storer.Update(ctx, {entityVar}); err != nil {
        if errors.Is(err, ErrUniqueEntry) {
            return {Entity}{}, fmt.Errorf("update: %w", ErrUniqueEntry)
        }
        return {Entity}{}, fmt.Errorf("update: %w", err)
    }

    if err := b.delegate.Call(ctx, ActionUpdatedData({entityVar})); err != nil {
        b.log.Error(ctx, "{entity}bus: delegate call failed", "action", ActionUpdated, "err", err)
    }

    return {entityVar}, nil
}

// Delete removes a {entity} from the system.
func (b *Business) Delete(ctx context.Context, {entityVar} {Entity}) error {
    ctx, span := otel.AddSpan(ctx, "business.{entity}bus.Delete")
    defer span.End()

    if err := b.storer.Delete(ctx, {entityVar}); err != nil {
        return fmt.Errorf("delete: %w", err)
    }

    if err := b.delegate.Call(ctx, ActionDeletedData({entityVar})); err != nil {
        b.log.Error(ctx, "{entity}bus: delegate call failed", "action", ActionDeleted, "err", err)
    }

    return nil
}

// Query retrieves a list of existing {entities} from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]{Entity}, error) {
    ctx, span := otel.AddSpan(ctx, "business.{entity}bus.Query")
    defer span.End()

    {entities}, err := b.storer.Query(ctx, filter, orderBy, page)
    if err != nil {
        return nil, fmt.Errorf("query: %w", err)
    }

    return {entities}, nil
}

// Count returns the total number of {entities}.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
    ctx, span := otel.AddSpan(ctx, "business.{entity}bus.Count")
    defer span.End()

    return b.storer.Count(ctx, filter)
}

// QueryByID finds the {entity} by the specified ID.
func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) ({Entity}, error) {
    ctx, span := otel.AddSpan(ctx, "business.{entity}bus.QueryByID")
    defer span.End()

    {entityVar}, err := b.storer.QueryByID(ctx, id)
    if err != nil {
        return {Entity}{}, fmt.Errorf("query: {entityVar}ID[%s]: %w", id, err)
    }

    return {entityVar}, nil
}
```

#### `testutil.go`

```go
package {entity}bus

import (
    "context"
    "fmt"
    "math/rand"
    "sort"
)

// TestNew{Entities} is a helper method for testing.
func TestNew{Entities}(n int) []New{Entity} {
    new{Entities} := make([]New{Entity}, n)

    idx := rand.Intn(10000)
    for i := 0; i < n; i++ {
        idx++

        n{entityAbbrev} := New{Entity}{
            // Fill with test data using fmt.Sprintf("{Entity}%d", idx)
            // For FK UUIDs: accept IDs as function params
        }

        new{Entities}[i] = n{entityAbbrev}
    }

    return new{Entities}
}

// TestSeed{Entities} is a helper method for testing.
func TestSeed{Entities}(ctx context.Context, n int, api *Business) ([]{Entity}, error) {
    new{Entities} := TestNew{Entities}(n)

    {entities} := make([]{Entity}, len(new{Entities}))
    for i, n{entityAbbrev} := range new{Entities} {
        {entityVar}, err := api.Create(ctx, n{entityAbbrev})
        if err != nil {
            return nil, fmt.Errorf("seeding {entity}: idx: %d : %w", i, err)
        }

        {entities}[i] = {entityVar}
    }

    sort.Slice({entities}, func(i, j int) bool {
        return {entities}[i].Name <= {entities}[j].Name
    })

    return {entities}, nil
}
```

**Note for complex domains**: If the entity has foreign keys, `TestNew{Entities}` and `TestSeed{Entities}` must accept parent entity IDs as parameters. Example:

```go
func TestSeed{Entities}(ctx context.Context, n int, parentIDs uuid.UUIDs, api *Business) ([]{Entity}, error) {
```

---

### 4.3: Database Store

**Directory**: `business/domain/{area}/{entity}bus/stores/{entity}db/`

#### `model.go`

```go
package {entity}db

import (
    "database/sql"

    "github.com/google/uuid"
    "github.com/{module}/business/domain/{area}/{entity}bus"
)

type {dbEntity} struct {
    ID          uuid.UUID      `db:"id"`
    // Map each column:
    // NOT NULL text → string
    // NULL text → sql.NullString
    // NOT NULL int → int
    // NULL int → sql.NullInt64
    // UUID → uuid.UUID
    // BOOLEAN → bool
    // TIMESTAMP → time.Time
}

func toDB{Entity}(bus {entity}bus.{Entity}) {dbEntity} {
    db := {dbEntity}{
        ID:   bus.ID,
        // Map all fields from bus → db
        // For nullable strings: check if empty → sql.NullString{String: val, Valid: val != ""}
    }
    return db
}

func toBus{Entity}(db {dbEntity}) {entity}bus.{Entity} {
    return {entity}bus.{Entity}{
        ID:   db.ID,
        // Map all fields from db → bus
        // For sql.NullString: use .String (extracts value, empty if NULL)
    }
}

func toBus{Entities}(dbs []{dbEntity}) []{entity}bus.{Entity} {
    items := make([]{entity}bus.{Entity}, len(dbs))
    for i, db := range dbs {
        items[i] = toBus{Entity}(db)
    }
    return items
}
```

**Important**: If the table has `TIMESTAMP` or `time.Time` fields, add an `init()` function:

```go
func init() {
    time.Local = time.UTC
}
```

#### `filter.go`

```go
package {entity}db

import (
    "bytes"
    "strings"

    "github.com/{module}/business/domain/{area}/{entity}bus"
)

func applyFilter(filter {entity}bus.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
    var wc []string

    if filter.ID != nil {
        data["id"] = *filter.ID
        wc = append(wc, "id = :id")
    }

    // For text fields: ILIKE with wildcards
    // if filter.Name != nil {
    //     data["name"] = "%" + *filter.Name + "%"
    //     wc = append(wc, "name ILIKE :name")
    // }

    // For UUID FK fields: exact match
    // if filter.UserID != nil {
    //     data["user_id"] = *filter.UserID
    //     wc = append(wc, "user_id = :user_id")
    // }

    // For boolean fields: exact match
    // if filter.IsActive != nil {
    //     data["is_active"] = *filter.IsActive
    //     wc = append(wc, "is_active = :is_active")
    // }

    if len(wc) > 0 {
        buf.WriteString(" WHERE ")
        buf.WriteString(strings.Join(wc, " AND "))
    }
}
```

#### `order.go`

```go
package {entity}db

import (
    "fmt"

    "github.com/{module}/business/domain/{area}/{entity}bus"
    "github.com/{module}/business/sdk/order"
)

var orderByFields = map[string]string{
    {entity}bus.OrderByID:   "id",
    {entity}bus.OrderByName: "name",
    // Map each OrderBy constant to its DB column name
}

func orderByClause(orderBy order.By) (string, error) {
    by, exists := orderByFields[orderBy.Field]
    if !exists {
        return "", fmt.Errorf("field %q does not exist", orderBy.Field)
    }

    return " ORDER BY " + by + " " + orderBy.Direction, nil
}
```

#### `{entity}db.go`

```go
package {entity}db

import (
    "bytes"
    "context"
    "errors"
    "fmt"

    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"
    "github.com/{module}/business/domain/{area}/{entity}bus"
    "github.com/{module}/business/sdk/order"
    "github.com/{module}/business/sdk/page"
    "github.com/{module}/business/sdk/sqldb"
    "github.com/{module}/foundation/logger"
)

// Store manages the set of APIs for {entity} database access.
type Store struct {
    log *logger.Logger
    db  sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
    return &Store{
        log: log,
        db:  db,
    }
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) ({entity}bus.Storer, error) {
    ec, err := sqldb.GetExtContext(tx)
    if err != nil {
        return nil, err
    }

    store := Store{
        log: s.log,
        db:  ec,
    }

    return &store, nil
}

// Create inserts a new {entity} into the database.
func (s *Store) Create(ctx context.Context, {entityVar} {entity}bus.{Entity}) error {
    const q = `
    INSERT INTO {schema}.{table} (
        {comma_separated_columns}
    ) VALUES (
        {comma_separated_named_params}
    )
    `
    if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDB{Entity}({entityVar})); err != nil {
        if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
            return fmt.Errorf("namedexeccontext: %w", {entity}bus.ErrUniqueEntry)
        }
        // Add FK violation handling for complex domains:
        // if errors.Is(err, sqldb.ErrForeignKeyViolation) {
        //     return fmt.Errorf("namedexeccontext: %w", {entity}bus.ErrFKViolation)
        // }
        return fmt.Errorf("namedexeccontext: %w", err)
    }

    return nil
}

// Update modifies data about a {entity} in the database.
func (s *Store) Update(ctx context.Context, {entityVar} {entity}bus.{Entity}) error {
    const q = `
    UPDATE
        {schema}.{table}
    SET
        {set_clauses}
    WHERE
        id = :id
    `
    if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDB{Entity}({entityVar})); err != nil {
        if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
            return fmt.Errorf("namedexeccontext: %w", {entity}bus.ErrUniqueEntry)
        }
        return fmt.Errorf("namedexeccontext: %w", err)
    }

    return nil
}

// Delete removes a {entity} from the database.
func (s *Store) Delete(ctx context.Context, {entityVar} {entity}bus.{Entity}) error {
    const q = `
    DELETE FROM
        {schema}.{table}
    WHERE
        id = :id
    `
    if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDB{Entity}({entityVar})); err != nil {
        return fmt.Errorf("namedexeccontext: %w", err)
    }

    return nil
}

// Query retrieves a list of existing {entities} from the database.
func (s *Store) Query(ctx context.Context, filter {entity}bus.QueryFilter, orderBy order.By, page page.Page) ([]{entity}bus.{Entity}, error) {
    data := map[string]any{
        "offset":        (page.Number() - 1) * page.RowsPerPage(),
        "rows_per_page": page.RowsPerPage(),
    }

    const q = `
    SELECT
        {comma_separated_columns}
    FROM
        {schema}.{table}`

    buf := bytes.NewBufferString(q)
    applyFilter(filter, data, buf)

    orderByClause, err := orderByClause(orderBy)
    if err != nil {
        return nil, err
    }

    buf.WriteString(orderByClause)
    buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

    var dbItems []{dbEntity}
    if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbItems); err != nil {
        return nil, fmt.Errorf("namedqueryslice: %w", err)
    }

    return toBus{Entities}(dbItems), nil
}

// Count returns the total number of {entities} in the DB.
func (s *Store) Count(ctx context.Context, filter {entity}bus.QueryFilter) (int, error) {
    data := map[string]any{}

    const q = `
    SELECT
        COUNT(1) AS count
    FROM
        {schema}.{table}`

    buf := bytes.NewBufferString(q)
    applyFilter(filter, data, buf)

    var count struct {
        Count int `db:"count"`
    }
    if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
        return 0, fmt.Errorf("namedquerysingle: %w", err)
    }

    return count.Count, nil
}

// QueryByID retrieves a single {entity} by its id.
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) ({entity}bus.{Entity}, error) {
    data := struct {
        ID string `db:"id"`
    }{
        ID: id.String(),
    }

    const q = `
    SELECT
        {comma_separated_columns}
    FROM
        {schema}.{table}
    WHERE
        id = :id
    `

    var db{entityAbbrev} {dbEntity}
    if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &db{entityAbbrev}); err != nil {
        if errors.Is(err, sqldb.ErrDBNotFound) {
            return {entity}bus.{Entity}{}, fmt.Errorf("db: %w", {entity}bus.ErrNotFound)
        }
        return {entity}bus.{Entity}{}, fmt.Errorf("querystruct: %w", err)
    }

    return toBus{Entity}(db{entityAbbrev}), nil
}
```

---

### 4.4: Application Layer

**Directory**: `app/domain/{area}/{entity}app/`

#### `model.go`

```go
package {entity}app

import (
    "encoding/json"

    "github.com/{module}/app/sdk/errs"
    "github.com/{module}/business/domain/{area}/{entity}bus"
)

// QueryParams represents the query parameters that can be used.
type QueryParams struct {
    Page    string
    Rows    string
    OrderBy string
    ID      string
    // One string field per filterable column
}

// {Entity} represents a {entity} in the API layer.
type {Entity} struct {
    ID   string `json:"id"`
    // All fields as strings for UUIDs, native types for int/bool
    // Time fields as string (RFC3339)
    // JSON tags use camelCase for frontend consumption (e.g., sortOrder, isActive)
}

// Encode implements the encoder interface.
func (app {Entity}) Encode() ([]byte, string, error) {
    data, err := json.Marshal(app)
    return data, "application/json", err
}

// ToApp{Entity} converts a business {entity} to an app {entity}.
func ToApp{Entity}(bus {entity}bus.{Entity}) {Entity} {
    return {Entity}{
        ID:   bus.ID.String(),
        // UUID fields: .String()
        // Time fields: .Format(time.RFC3339)
        // Other fields: direct assignment
    }
}

// ToApp{Entities} converts a slice of business {entities} to app {entities}.
func ToApp{Entities}(bus []{entity}bus.{Entity}) []{Entity} {
    app := make([]{Entity}, len(bus))
    for i, v := range bus {
        app[i] = ToApp{Entity}(v)
    }
    return app
}

// =============================================================================

// New{Entity} contains information needed to create a new {entity}.
type New{Entity} struct {
    // Required fields: validate:"required"
    // Optional fields: validate:"omitempty"
    // UUID FK fields: string with validate:"required" if NOT NULL
    // Example:
    // Name string `json:"name" validate:"required,min=3,max=50"`
}

// Decode implements the decoder interface.
func (app *New{Entity}) Decode(data []byte) error {
    return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app New{Entity}) Validate() error {
    if err := errs.Check(app); err != nil {
        return errs.Newf(errs.InvalidArgument, "validate: %s", err)
    }

    return nil
}

// ToBusNew{Entity} converts an app new {entity} to a bus new {entity}.
//
// SIMPLE DOMAINS (no FK UUIDs or time parsing needed):
//   Exported function, no error return:
//   func ToBusNew{Entity}(app New{Entity}) {entity}bus.New{Entity} { ... }
//
// COMPLEX DOMAINS (FK UUIDs or time fields need parsing):
//   Private function, returns error:
//   func toBusNew{Entity}(app New{Entity}) ({entity}bus.New{Entity}, error) {
//       userID, err := uuid.Parse(app.UserID)
//       if err != nil {
//           return {entity}bus.New{Entity}{}, errs.NewFieldsError("userID", err)
//       }
//       dateReceived, err := time.Parse(time.RFC3339, app.DateReceived)
//       if err != nil {
//           return {entity}bus.New{Entity}{}, errs.NewFieldsError("dateReceived", err)
//       }
//       return {entity}bus.New{Entity}{UserID: userID, DateReceived: dateReceived, ...}, nil
//   }
func ToBusNew{Entity}(app New{Entity}) {entity}bus.New{Entity} {
    return {entity}bus.New{Entity}{
        // Map fields directly for simple domains
    }
}

// =============================================================================

// Update{Entity} defines what information may be provided to modify an existing {entity}.
type Update{Entity} struct {
    // All pointer fields
    // Example: Name *string `json:"name" validate:"omitempty,min=3,max=50"`
}

// Decode implements the decoder interface.
func (app *Update{Entity}) Decode(data []byte) error {
    return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app Update{Entity}) Validate() error {
    if err := errs.Check(app); err != nil {
        return errs.Newf(errs.InvalidArgument, "validate: %s", err)
    }

    return nil
}

// ToBusUpdate{Entity} converts an app update to bus update.
func ToBusUpdate{Entity}(app Update{Entity}) {entity}bus.Update{Entity} {
    return {entity}bus.Update{Entity}{
        // Direct pointer field mapping for simple types
        // For UUID pointer fields, parse in the app layer method instead
    }
}
```

**Important for complex domains**: If `ToBusNew{Entity}` or `ToBusUpdate{Entity}` need to parse UUIDs or timestamps, they should return errors:

```go
func toBusNew{Entity}(app New{Entity}) ({entity}bus.New{Entity}, error) {
    userID, err := uuid.Parse(app.UserID)
    if err != nil {
        return {entity}bus.New{Entity}{}, errs.NewFieldsError("userID", err)
    }
    // ... parse other FK UUIDs
    return {entity}bus.New{Entity}{
        UserID: userID,
        // ...
    }, nil
}
```

#### `filter.go`

```go
package {entity}app

import (
    "github.com/google/uuid"
    "github.com/{module}/app/sdk/errs"
    "github.com/{module}/business/domain/{area}/{entity}bus"
)

func parseFilter(qp QueryParams) ({entity}bus.QueryFilter, error) {
    var filter {entity}bus.QueryFilter

    if qp.ID != "" {
        id, err := uuid.Parse(qp.ID)
        if err != nil {
            return {entity}bus.QueryFilter{}, errs.NewFieldsError("id", err)
        }
        filter.ID = &id
    }

    // For string fields:
    // if qp.Name != "" {
    //     filter.Name = &qp.Name
    // }

    // For UUID FK fields:
    // if qp.UserID != "" {
    //     id, err := uuid.Parse(qp.UserID)
    //     if err != nil {
    //         return {entity}bus.QueryFilter{}, errs.NewFieldsError("userID", err)
    //     }
    //     filter.UserID = &id
    // }

    return filter, nil
}
```

#### `order.go`

```go
package {entity}app

import (
    "github.com/{module}/business/domain/{area}/{entity}bus"
    "github.com/{module}/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
    "id":   {entity}bus.OrderByID,
    "name": {entity}bus.OrderByName,
    // Map camelCase API field names to business layer OrderBy constants
}
```

#### `{entity}app.go`

```go
package {entity}app

import (
    "context"
    "errors"

    "github.com/google/uuid"
    "github.com/{module}/app/sdk/auth"
    "github.com/{module}/app/sdk/errs"
    "github.com/{module}/app/sdk/query"
    "github.com/{module}/business/domain/{area}/{entity}bus"
    "github.com/{module}/business/sdk/order"
    "github.com/{module}/business/sdk/page"
)

// App manages the set of app layer api functions for the {entity} domain.
type App struct {
    {entityVar}Bus *{entity}bus.Business
    auth           *auth.Auth
}

// NewApp constructs a {entity} app API for use.
func NewApp({entityVar}Bus *{entity}bus.Business) *App {
    return &App{
        {entityVar}Bus: {entityVar}Bus,
    }
}

// NewAppWithAuth constructs a {entity} app API for use with auth support.
func NewAppWithAuth({entityVar}Bus *{entity}bus.Business, ath *auth.Auth) *App {
    return &App{
        auth:           ath,
        {entityVar}Bus: {entityVar}Bus,
    }
}

// Create adds a new {entity} to the system.
//
// SIMPLE DOMAINS (no error from conversion):
func (a *App) Create(ctx context.Context, app New{Entity}) ({Entity}, error) {
    {entityVar}, err := a.{entityVar}Bus.Create(ctx, ToBusNew{Entity}(app))
    if err != nil {
        if errors.Is(err, {entity}bus.ErrUniqueEntry) {
            return {Entity}{}, errs.New(errs.Aborted, {entity}bus.ErrUniqueEntry)
        }
        return {Entity}{}, errs.Newf(errs.Internal, "create: {entity}[%+v]: %s", {entityVar}, err)
    }

    return ToApp{Entity}({entityVar}), nil
}
// COMPLEX DOMAINS (conversion returns error — FK UUID/time parsing):
//   func (a *App) Create(ctx context.Context, app New{Entity}) ({Entity}, error) {
//       na, err := toBusNew{Entity}(app)
//       if err != nil {
//           return {Entity}{}, errs.New(errs.InvalidArgument, err)
//       }
//       {entityVar}, err := a.{entityVar}Bus.Create(ctx, na)
//       if err != nil { ... }
//       return ToApp{Entity}({entityVar}), nil
//   }

// Update updates an existing {entity}.
func (a *App) Update(ctx context.Context, app Update{Entity}, id uuid.UUID) ({Entity}, error) {
    u{entityAbbrev} := ToBusUpdate{Entity}(app)
    // COMPLEX: u{entityAbbrev}, err := toBusUpdate{Entity}(app); if err != nil { return ... }

    {entityVar}, err := a.{entityVar}Bus.QueryByID(ctx, id)
    if err != nil {
        return {Entity}{}, errs.Newf(errs.NotFound, "update: {entity}[%s]: %s", id, err)
    }

    updated, err := a.{entityVar}Bus.Update(ctx, {entityVar}, u{entityAbbrev})
    if err != nil {
        if errors.Is(err, {entity}bus.ErrNotFound) {
            return {Entity}{}, errs.New(errs.NotFound, err)
        }
        return {Entity}{}, errs.Newf(errs.Internal, "update: {entity}[%+v]: %s", updated, err)
    }

    return ToApp{Entity}(updated), nil
}

// Delete removes an existing {entity}.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
    {entityVar}, err := a.{entityVar}Bus.QueryByID(ctx, id)
    if err != nil {
        return errs.Newf(errs.NotFound, "delete: {entity}[%s]: %s", id, err)
    }

    if err := a.{entityVar}Bus.Delete(ctx, {entityVar}); err != nil {
        return errs.Newf(errs.Internal, "delete: {entity}[%+v]: %s", {entityVar}, err)
    }

    return nil
}

// Query returns a list of {entities}.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[{Entity}], error) {
    page, err := page.Parse(qp.Page, qp.Rows)
    if err != nil {
        return query.Result[{Entity}]{}, errs.NewFieldsError("page", err)
    }

    filter, err := parseFilter(qp)
    if err != nil {
        return query.Result[{Entity}]{}, errs.NewFieldsError("filter", err)
    }

    orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
    if err != nil {
        return query.Result[{Entity}]{}, errs.NewFieldsError("orderby", err)
    }

    {entities}, err := a.{entityVar}Bus.Query(ctx, filter, orderBy, page)
    if err != nil {
        return query.Result[{Entity}]{}, errs.Newf(errs.Internal, "query: %s", err)
    }

    total, err := a.{entityVar}Bus.Count(ctx, filter)
    if err != nil {
        return query.Result[{Entity}]{}, errs.Newf(errs.Internal, "count: %s", err)
    }

    return query.NewResult(ToApp{Entities}({entities}), total, page), nil
}

// QueryByID returns a single {entity} based on the id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) ({Entity}, error) {
    {entityVar}, err := a.{entityVar}Bus.QueryByID(ctx, id)
    if err != nil {
        return {Entity}{}, errs.Newf(errs.NotFound, "query: {entity}[%s]: %s", id, err)
    }

    return ToApp{Entity}({entityVar}), nil
}
```

---

### 4.5: API Layer

**Directory**: `api/domain/http/{area}/{entity}api/`

#### `{entity}api.go`

```go
package {entity}api

import (
    "context"
    "net/http"

    "github.com/google/uuid"
    "github.com/{module}/app/domain/{area}/{entity}app"
    "github.com/{module}/app/sdk/errs"
    "github.com/{module}/foundation/web"
)

type api struct {
    {entityVar}app *{entity}app.App
}

func newAPI({entityVar}app *{entity}app.App) *api {
    return &api{
        {entityVar}app: {entityVar}app,
    }
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
    var app {entity}app.New{Entity}
    if err := web.Decode(r, &app); err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    {entityVar}, err := api.{entityVar}app.Create(ctx, app)
    if err != nil {
        return errs.NewError(err)
    }

    return {entityVar}
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
    var app {entity}app.Update{Entity}
    if err := web.Decode(r, &app); err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    {entityVar}ID := web.Param(r, "{url_param_id}")
    parsed, err := uuid.Parse({entityVar}ID)
    if err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    {entityVar}, err := api.{entityVar}app.Update(ctx, app, parsed)
    if err != nil {
        return errs.NewError(err)
    }

    return {entityVar}
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
    {entityVar}ID := web.Param(r, "{url_param_id}")

    parsed, err := uuid.Parse({entityVar}ID)
    if err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    err = api.{entityVar}app.Delete(ctx, parsed)
    if err != nil {
        return errs.NewError(err)
    }

    return nil
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
    qp, err := parseQueryParams(r)
    if err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    result, err := api.{entityVar}app.Query(ctx, qp)
    if err != nil {
        return errs.NewError(err)
    }

    return result
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
    {entityVar}ID := web.Param(r, "{url_param_id}")

    parsed, err := uuid.Parse({entityVar}ID)
    if err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    result, err := api.{entityVar}app.QueryByID(ctx, parsed)
    if err != nil {
        return errs.NewError(err)
    }

    return result
}
```

#### `route.go`

```go
package {entity}api

import (
    "net/http"

    "github.com/{module}/api/sdk/http/mid"
    "github.com/{module}/app/domain/{area}/{entity}app"
    "github.com/{module}/app/sdk/auth"
    "github.com/{module}/app/sdk/authclient"
    "github.com/{module}/business/domain/{area}/{entity}bus"
    "github.com/{module}/business/domain/core/permissionsbus"
    "github.com/{module}/foundation/logger"
    "github.com/{module}/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
    Log            *logger.Logger
    {Entity}Bus    *{entity}bus.Business
    AuthClient     *authclient.Client
    PermissionsBus *permissionsbus.Business
}

const (
    RouteTable = "{schema}.{table}"
)

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
    const version = "v1"

    authen := mid.Authenticate(cfg.AuthClient)

    api := newAPI({entity}app.NewApp(cfg.{Entity}Bus))
    app.HandlerFunc(http.MethodGet, version, "/{area}/{url-path}", api.query, authen,
        mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))
    app.HandlerFunc(http.MethodGet, version, "/{area}/{url-path}/{{url_param_id}}", api.queryByID, authen,
        mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))
    app.HandlerFunc(http.MethodPost, version, "/{area}/{url-path}", api.create, authen,
        mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAny))
    app.HandlerFunc(http.MethodPut, version, "/{area}/{url-path}/{{url_param_id}}", api.update, authen,
        mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAny))
    app.HandlerFunc(http.MethodDelete, version, "/{area}/{url-path}/{{url_param_id}}", api.delete, authen,
        mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete, auth.RuleAny))
}
```

#### `filter.go`

```go
package {entity}api

import (
    "net/http"

    "github.com/{module}/app/domain/{area}/{entity}app"
)

func parseQueryParams(r *http.Request) ({entity}app.QueryParams, error) {
    values := r.URL.Query()

    filter := {entity}app.QueryParams{
        Page:    values.Get("page"),
        Rows:    values.Get("rows"),
        OrderBy: values.Get("orderBy"),
        ID:      values.Get("id"),
        // One values.Get() per filterable field
    }

    return filter, nil
}
```

---

### 4.6: Wiring Edits

**File**: `api/cmd/services/ichor/build/all/all.go`

Make these edits (do NOT rewrite the file — use targeted edits):

1. **Add imports**: Add imports for `{entity}api`, `{entity}app`, `{entity}bus`, `{entity}db` packages
2. **Business instantiation**: Add near other bus instantiations in the same schema group:
   ```go
   {entityVar}Bus := {entity}bus.NewBusiness(cfg.Log, delegate, {entity}db.NewStore(cfg.Log, cfg.DB))
   ```
3. **Delegate registration**: Add near other delegate registrations:
   ```go
   delegateHandler.RegisterDomain(delegate, {entity}bus.DomainName, {entity}bus.EntityName)
   ```
4. **Route registration**: Add near other route registrations in the same schema group:
   ```go
   {entity}api.Routes(app, {entity}api.Config{
       {Entity}Bus:    {entityVar}Bus,
       AuthClient:     cfg.AuthClient,
       Log:            cfg.Log,
       PermissionsBus: permissionsBus,
   })
   ```

---

### 4.6.1: FormData Registration (Optional)

If the entity should support multi-entity transactional operations via the FormData system:

**File**: `api/cmd/services/ichor/build/all/formdata_registry.go`

1. **Add import** for `{entity}app` package
2. **Add parameter** to `buildFormDataRegistry` (or equivalent function): `{entityVar}App *{entity}app.App`
3. **Add registration**:
```go
if err := registry.Register(formdataregistry.EntityRegistration{
    Name: "{schema}.{table}",
    DecodeNew: func(data json.RawMessage) (interface{}, error) {
        var app {entity}app.New{Entity}
        if err := json.Unmarshal(data, &app); err != nil {
            return nil, err
        }
        if err := app.Validate(); err != nil {
            return nil, err
        }
        return app, nil
    },
    CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
        return {entityVar}App.Create(ctx, model.({entity}app.New{Entity}))
    },
    CreateModel: {entity}app.New{Entity}{},
    DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
        var app {entity}app.Update{Entity}
        if err := json.Unmarshal(data, &app); err != nil {
            return nil, err
        }
        if err := app.Validate(); err != nil {
            return nil, err
        }
        return app, nil
    },
    UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
        return {entityVar}App.Update(ctx, model.({entity}app.Update{Entity}), id)
    },
    UpdateModel: {entity}app.Update{Entity}{},
}); err != nil {
    return nil, fmt.Errorf("register {table}: %w", err)
}
```
4. **Update call site** in `all.go` to pass `{entity}app.NewApp({entityVar}Bus)` to `buildFormDataRegistry`

---

### 4.7: Test Infrastructure Edits

#### `business/sdk/dbtest/dbtest.go`

1. **Add field to BusDomain struct**:
   ```go
   {Entity} *{entity}bus.Business
   ```

2. **Add instantiation** in the `NewDatabase` function (or wherever BusDomain is populated):
   ```go
   {Entity}: {entity}bus.NewBusiness(log, delegate, {entity}db.NewStore(log, db)),
   ```

#### `api/sdk/http/apitest/model.go`

1. **Add field to SeedData struct**:
   ```go
   {Entities} []{entity}app.{Entity}
   ```

#### `business/domain/core/tableaccessbus/testutil.go`

1. **Add table access entry** in the `TestSeedTableAccess` function's `newTAs` slice, in the appropriate schema section:
   ```go
   {RoleID: uuid.Nil, TableName: "{schema}.{table}", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
   ```

---

### 4.8: Test Files

**Directory**: `api/cmd/services/ichor/tests/{area}/{entity}api/`

**Package name**: `{entity}_test` (e.g., `assetcondition_test`)

#### `{entity}_test.go`

```go
package {entity}_test

import (
    "testing"

    "github.com/{module}/api/sdk/http/apitest"
)

func Test_{Entity}(t *testing.T) {
    t.Parallel()

    test := apitest.StartTest(t, "Test_{Entity}")

    sd, err := insertSeedData(test.DB, test.Auth)
    if err != nil {
        t.Fatalf("seeding error %s", err)
    }

    test.Run(t, query200(sd), "query-200")
    test.Run(t, queryByID200(sd), "query-by-id-200")

    test.Run(t, create200(sd), "create-200")
    test.Run(t, create400(sd), "create-400")
    test.Run(t, create401(sd), "create-401")

    test.Run(t, update200(sd), "update-200")
    test.Run(t, update400(sd), "update-400")
    test.Run(t, update401(sd), "update-401")

    test.Run(t, delete200(sd), "delete-200")
    test.Run(t, delete401(sd), "delete-401")
}
```

#### `seed_test.go`

```go
package {entity}_test

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    "github.com/{module}/api/domain/http/{area}/{entity}api"
    "github.com/{module}/api/sdk/http/apitest"
    "github.com/{module}/app/domain/{area}/{entity}app"
    "github.com/{module}/app/sdk/auth"
    "github.com/{module}/business/domain/{area}/{entity}bus"
    "github.com/{module}/business/domain/core/rolebus"
    "github.com/{module}/business/domain/core/tableaccessbus"
    "github.com/{module}/business/domain/core/userbus"
    "github.com/{module}/business/domain/core/userrolebus"
    "github.com/{module}/business/sdk/dbtest"
)

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
    ctx := context.Background()
    busDomain := db.BusDomain

    // Create regular user (restricted permissions)
    usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
    if err != nil {
        return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
    }
    tu1 := apitest.User{
        User:  usrs[0],
        Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
    }

    // Create admin user (full permissions)
    usrs, err = userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
    if err != nil {
        return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
    }
    tu2 := apitest.User{
        User:  usrs[0],
        Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
    }

    // Seed domain entities
    // For simple domains:
    {entities}, err := {entity}bus.TestSeed{Entities}(ctx, 10, busDomain.{Entity})
    if err != nil {
        return apitest.SeedData{}, err
    }
    // For complex domains with FK dependencies:
    // Create parent entities first, then pass their IDs

    // =========================================================================
    // Permissions setup
    // =========================================================================
    roles, err := rolebus.TestSeedRoles(ctx, 2, busDomain.Role)
    if err != nil {
        return apitest.SeedData{}, fmt.Errorf("seeding roles : %w", err)
    }

    roleIDs := make(uuid.UUIDs, len(roles))
    for i, r := range roles {
        roleIDs[i] = r.ID
    }

    userIDs := make(uuid.UUIDs, 2)
    userIDs[0] = tu1.ID
    userIDs[1] = tu2.ID

    _, err = userrolebus.TestSeedUserRoles(ctx, userIDs, roleIDs, busDomain.UserRole)
    if err != nil {
        return apitest.SeedData{}, fmt.Errorf("seeding user roles : %w", err)
    }

    _, err = tableaccessbus.TestSeedTableAccess(ctx, roleIDs, busDomain.TableAccess)
    if err != nil {
        return apitest.SeedData{}, fmt.Errorf("seeding table access : %w", err)
    }

    // Restrict tu1 to read-only for this table
    ur1, err := busDomain.UserRole.QueryByUserID(ctx, tu1.ID)
    if err != nil {
        return apitest.SeedData{}, fmt.Errorf("querying user1 roles : %w", err)
    }

    usrRoleIDs := make(uuid.UUIDs, len(ur1))
    for i, r := range ur1 {
        usrRoleIDs[i] = r.RoleID
    }

    tas, err := busDomain.TableAccess.QueryByRoleIDs(ctx, usrRoleIDs)
    if err != nil {
        return apitest.SeedData{}, fmt.Errorf("querying table access : %w", err)
    }

    for _, ta := range tas {
        if ta.TableName == {entity}api.RouteTable {
            update := tableaccessbus.UpdateTableAccess{
                CanCreate: dbtest.BoolPointer(false),
                CanUpdate: dbtest.BoolPointer(false),
                CanDelete: dbtest.BoolPointer(false),
                CanRead:   dbtest.BoolPointer(true),
            }
            _, err := busDomain.TableAccess.Update(ctx, ta, update)
            if err != nil {
                return apitest.SeedData{}, fmt.Errorf("updating table access : %w", err)
            }
        }
    }

    return apitest.SeedData{
        Users:       []apitest.User{tu1},
        Admins:      []apitest.User{tu2},
        {Entities}:  {entity}app.ToApp{Entities}({entities}),
    }, nil
}
```

#### `create_test.go`

```go
package {entity}_test

import (
    "net/http"

    "github.com/google/go-cmp/cmp"
    "github.com/{module}/api/sdk/http/apitest"
    "github.com/{module}/app/domain/{area}/{entity}app"
    "github.com/{module}/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
    table := []apitest.Table{
        {
            Name:       "basic",
            URL:        "/v1/{area}/{url-path}",
            Token:      sd.Admins[0].Token,
            Method:     http.MethodPost,
            StatusCode: http.StatusOK,
            Input: &{entity}app.New{Entity}{
                // Fill with valid test data
            },
            GotResp: &{entity}app.{Entity}{},
            ExpResp: &{entity}app.{Entity}{
                // Expected response (ID will be overwritten)
            },
            CmpFunc: func(got any, exp any) string {
                gotResp, exists := got.(*{entity}app.{Entity})
                if !exists {
                    return "error occurred"
                }

                expResp := exp.(*{entity}app.{Entity})
                expResp.ID = gotResp.ID

                return cmp.Diff(got, exp)
            },
        },
    }

    return table
}

func create400(sd apitest.SeedData) []apitest.Table {
    table := []apitest.Table{
        {
            Name:       "missing required field",
            URL:        "/v1/{area}/{url-path}",
            Token:      sd.Admins[0].Token,
            Method:     http.MethodPost,
            StatusCode: http.StatusBadRequest,
            Input: &{entity}app.New{Entity}{
                // Omit a required field
            },
            GotResp: &errs.Error{},
            ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"{required_field}\",\"error\":\"{required_field} is a required field\"}]"),
            CmpFunc: func(got any, exp any) string {
                gotResp, exists := got.(*errs.Error)
                if !exists {
                    return "error occurred"
                }
                return cmp.Diff(exp, gotResp)
            },
        },
    }
    return table
}

func create401(sd apitest.SeedData) []apitest.Table {
    table := []apitest.Table{
        {
            Name:       "emptytoken",
            URL:        "/v1/{area}/{url-path}",
            Token:      "&nbsp;",
            Method:     http.MethodPost,
            StatusCode: http.StatusUnauthorized,
            GotResp:    &errs.Error{},
            ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
            CmpFunc: func(got any, exp any) string {
                return cmp.Diff(got, exp)
            },
        },
        {
            Name:       "badtoken",
            URL:        "/v1/{area}/{url-path}",
            Token:      sd.Admins[0].Token[:10],
            Method:     http.MethodPost,
            StatusCode: http.StatusUnauthorized,
            GotResp:    &errs.Error{},
            ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
            CmpFunc: func(got any, exp any) string {
                return cmp.Diff(got, exp)
            },
        },
        {
            Name:       "badsig",
            URL:        "/v1/{area}/{url-path}",
            Token:      sd.Admins[0].Token + "A",
            Method:     http.MethodPost,
            StatusCode: http.StatusUnauthorized,
            GotResp:    &errs.Error{},
            ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
            CmpFunc: func(got any, exp any) string {
                return cmp.Diff(got, exp)
            },
        },
        {
            Name:       "wronguser",
            URL:        "/v1/{area}/{url-path}",
            Token:      sd.Users[0].Token,
            Method:     http.MethodPost,
            StatusCode: http.StatusUnauthorized,
            GotResp:    &errs.Error{},
            ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: {schema}.{table}"),
            CmpFunc: func(got any, exp any) string {
                return cmp.Diff(got, exp)
            },
        },
    }
    return table
}
```

#### `query_test.go`

```go
package {entity}_test

import (
    "net/http"

    "github.com/google/go-cmp/cmp"
    "github.com/{module}/api/sdk/http/apitest"
    "github.com/{module}/app/domain/{area}/{entity}app"
    "github.com/{module}/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
    table := []apitest.Table{
        {
            Name:       "basic",
            URL:        "/v1/{area}/{url-path}?page=1&rows=10",
            Token:      sd.Admins[0].Token,
            StatusCode: http.StatusOK,
            Method:     http.MethodGet,
            GotResp:    &query.Result[{entity}app.{Entity}]{},
            ExpResp: &query.Result[{entity}app.{Entity}]{
                Page:        1,
                RowsPerPage: 10,
                Total:       len(sd.{Entities}),
                Items:       sd.{Entities},
            },
            CmpFunc: func(got any, exp any) string {
                return cmp.Diff(got, exp)
            },
        },
    }

    return table
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
    table := []apitest.Table{
        {
            Name:       "basic",
            URL:        "/v1/{area}/{url-path}/" + sd.{Entities}[0].ID,
            Token:      sd.Admins[0].Token,
            StatusCode: http.StatusOK,
            Method:     http.MethodGet,
            GotResp:    &{entity}app.{Entity}{},
            ExpResp:    &sd.{Entities}[0],
            CmpFunc: func(got any, exp any) string {
                return cmp.Diff(got, exp)
            },
        },
    }

    return table
}
```

#### `update_test.go`

```go
package {entity}_test

import (
    "net/http"

    "github.com/google/go-cmp/cmp"
    "github.com/{module}/api/sdk/http/apitest"
    "github.com/{module}/app/domain/{area}/{entity}app"
    "github.com/{module}/app/sdk/errs"
    "github.com/{module}/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
    table := []apitest.Table{
        {
            Name:       "basic",
            URL:        "/v1/{area}/{url-path}/" + sd.{Entities}[0].ID,
            Token:      sd.Admins[0].Token,
            Method:     http.MethodPut,
            StatusCode: http.StatusOK,
            Input: &{entity}app.Update{Entity}{
                // Update a field using dbtest.StringPointer("Updated Value")
            },
            GotResp: &{entity}app.{Entity}{},
            ExpResp: &{entity}app.{Entity}{},
            CmpFunc: func(got any, exp any) string {
                gotResp, exists := got.(*{entity}app.{Entity})
                if !exists {
                    return "error occurred"
                }

                expResp := exp.(*{entity}app.{Entity})
                expResp.ID = gotResp.ID
                // Copy other server-generated fields

                return cmp.Diff(got, exp)
            },
        },
    }

    return table
}

func update400(sd apitest.SeedData) []apitest.Table {
    table := []apitest.Table{
        {
            Name:       "bad id",
            URL:        "/v1/{area}/{url-path}/not-a-uuid",
            Token:      sd.Admins[0].Token,
            Method:     http.MethodPut,
            StatusCode: http.StatusBadRequest,
            Input: &{entity}app.Update{Entity}{},
            GotResp: &errs.Error{},
            ExpResp: errs.Newf(errs.InvalidArgument, "invalid UUID length: 10"),
            CmpFunc: func(got any, exp any) string {
                return cmp.Diff(got, exp)
            },
        },
    }
    return table
}

func update401(sd apitest.SeedData) []apitest.Table {
    table := []apitest.Table{
        {
            Name:       "emptytoken",
            URL:        "/v1/{area}/{url-path}/" + sd.{Entities}[0].ID,
            Token:      "&nbsp;",
            Method:     http.MethodPut,
            StatusCode: http.StatusUnauthorized,
            GotResp:    &errs.Error{},
            ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
            CmpFunc: func(got any, exp any) string {
                return cmp.Diff(got, exp)
            },
        },
        {
            Name:       "wronguser",
            URL:        "/v1/{area}/{url-path}/" + sd.{Entities}[0].ID,
            Token:      sd.Users[0].Token,
            Method:     http.MethodPut,
            StatusCode: http.StatusUnauthorized,
            GotResp:    &errs.Error{},
            ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: {schema}.{table}"),
            CmpFunc: func(got any, exp any) string {
                return cmp.Diff(got, exp)
            },
        },
    }
    return table
}
```

#### `delete_test.go`

```go
package {entity}_test

import (
    "net/http"

    "github.com/google/go-cmp/cmp"
    "github.com/{module}/api/sdk/http/apitest"
    "github.com/{module}/app/sdk/errs"
)

func delete200(sd apitest.SeedData) []apitest.Table {
    table := []apitest.Table{
        {
            Name:       "basic",
            URL:        "/v1/{area}/{url-path}/" + sd.{Entities}[0].ID,
            Token:      sd.Admins[0].Token,
            Method:     http.MethodDelete,
            StatusCode: http.StatusNoContent,
        },
    }

    return table
}

func delete401(sd apitest.SeedData) []apitest.Table {
    table := []apitest.Table{
        {
            Name:       "emptytoken",
            URL:        "/v1/{area}/{url-path}/" + sd.{Entities}[0].ID,
            Token:      "&nbsp;",
            Method:     http.MethodDelete,
            StatusCode: http.StatusUnauthorized,
            GotResp:    &errs.Error{},
            ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
            CmpFunc: func(got any, exp any) string {
                return cmp.Diff(got, exp)
            },
        },
        {
            Name:       "wronguser",
            URL:        "/v1/{area}/{url-path}/" + sd.{Entities}[0].ID,
            Token:      sd.Users[0].Token,
            Method:     http.MethodDelete,
            StatusCode: http.StatusUnauthorized,
            GotResp:    &errs.Error{},
            ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission DELETE for table: {schema}.{table}"),
            CmpFunc: func(got any, exp any) string {
                return cmp.Diff(got, exp)
            },
        },
    }
    return table
}
```

---

## Step 5: Build Validation Loop (max 5 iterations)

After writing all files, enter an iterative build-fix cycle. The goal is a clean compilation before moving on.

### Iteration Protocol

```
For iteration = 1 to 5:
  1. Run: go build ./...
  2. If clean → proceed to Step 6
  3. If errors → analyze, fix, continue to next iteration
```

### Common Build Errors and Fixes

| Error Pattern | Likely Cause | Fix |
|--------------|-------------|-----|
| `cannot find package` | Import path typo (e.g., `timmaaez` vs `timmaaaz`) | Correct the import path |
| `undefined: {Type}` | Missing field in conversion function or wrong struct name | Add missing field or fix name |
| `too many arguments` / `not enough arguments` | Function signature mismatch between layers | Align signatures |
| `ambiguous import` | Package name conflicts with SDK (`page`, `order`) | Use `db{Entity}` prefix for DB structs |
| `missing comma` | Struct literal formatting | Add trailing comma |
| `imported and not used` | Premature import | Remove unused import or add usage |
| `cannot use X as type Y` | Wrong type in conversion (e.g., `string` vs `uuid.UUID`) | Fix type mapping per SQL→Go table |
| `duplicate field` | Copy-paste error in struct or wiring | Remove duplicate |

### If build fails after 5 iterations

Stop and report:
- What was attempted
- Remaining errors
- Likely root cause

Do NOT continue to Step 6 with a broken build.

---

## Step 6: Existing Test Validation

Run existing tests to catch wiring breakage caused by edits to shared files (`all.go`, `dbtest.go`, `apitest/model.go`, `tableaccessbus/testutil.go`).

```
For iteration = 1 to 3:
  1. Run: go test ./api/cmd/services/ichor/tests/... -count=1 -timeout 120s 2>&1 | head -100
  2. If all pass → proceed to Step 7
  3. If failures → analyze and fix wiring issues, continue
```

### Common Wiring Failures

| Failure | Cause | Fix |
|---------|-------|-----|
| `cannot find symbol` in `all.go` | Missing import or instantiation | Add import + bus instantiation |
| `too few values in struct literal` | New field added to `BusDomain` or `SeedData` but not populated | Add field initialization |
| `undefined: busDomain.{Entity}` | Field name mismatch in `dbtest.go` | Match exact field name |
| `undefined: sd.{Entities}` | Field name mismatch in `apitest/model.go` | Match exact field name |
| Test timeout | Unrelated infrastructure issue | Note and skip — not a wiring problem |

### If existing tests fail after 3 iterations

Stop, report remaining failures, and distinguish:
- **Wiring failures** (caused by this scaffold) — must fix
- **Pre-existing failures** (already broken before scaffold) — note and move on

---

## Step 7: New Domain Test Validation

Run the newly generated test suite for the scaffolded domain.

```
For iteration = 1 to 5:
  1. Run: go test ./api/cmd/services/ichor/tests/{area}/{entity}api/ -v -count=1 -timeout 120s
  2. If all pass → proceed to Step 8
  3. If failures → analyze, fix test code or domain code, continue
```

### Common New Test Failures

| Failure | Cause | Fix |
|---------|-------|-----|
| `404 Not Found` | Route not registered or wrong URL path | Check `route.go` URL and `all.go` registration |
| `401 Unauthorized` | Table access not seeded for test role | Check `tableaccessbus/testutil.go` entry |
| `400 Bad Request` on create | Validation tag mismatch or missing required field in test input | Align test input with `validate` tags |
| Wrong JSON field name | camelCase vs snake_case mismatch between app model and test expectation | Fix JSON tags |
| `pq: relation does not exist` | Migration not appended or wrong schema/table name | Fix migration SQL |
| Seed data count mismatch | `TestSeed{Entities}` count doesn't match query expectation | Align seed count with test assertions |
| `duplicate key` on seed | Unique constraint hit by random test data | Make test data more unique (add UUID suffix) |

### If new tests fail after 5 iterations

Stop and report:
- Tests passing vs failing
- Root cause analysis
- Suggested manual fixes

---

## Step 8: Lint Check

Run the linter to catch style issues:

```bash
make lint
```

Fix any linting errors in the generated files. Common issues:
- Missing error checks (`errcheck`)
- Unused parameters (`unparam`)
- Ineffective assignments (`ineffassign`)
- Style violations (`revive`, `gocritic`)

---

## Step 9: Report Summary

After validation completes (or stops due to max iterations), output a summary:

```
Domain generated: {Entity}
Build status: PASS/FAIL (N iterations)
Existing tests: PASS/FAIL
New tests: PASS/FAIL (N iterations)
Lint: PASS/FAIL

Files created:
  Business Layer:
    - business/domain/{area}/{entity}bus/model.go
    - business/domain/{area}/{entity}bus/filter.go
    - business/domain/{area}/{entity}bus/order.go
    - business/domain/{area}/{entity}bus/event.go
    - business/domain/{area}/{entity}bus/{entity}bus.go
    - business/domain/{area}/{entity}bus/testutil.go

  Database Store:
    - business/domain/{area}/{entity}bus/stores/{entity}db/model.go
    - business/domain/{area}/{entity}bus/stores/{entity}db/filter.go
    - business/domain/{area}/{entity}bus/stores/{entity}db/order.go
    - business/domain/{area}/{entity}bus/stores/{entity}db/{entity}db.go

  Application Layer:
    - app/domain/{area}/{entity}app/model.go
    - app/domain/{area}/{entity}app/filter.go
    - app/domain/{area}/{entity}app/order.go
    - app/domain/{area}/{entity}app/{entity}app.go

  API Layer:
    - api/domain/http/{area}/{entity}api/{entity}api.go
    - api/domain/http/{area}/{entity}api/route.go
    - api/domain/http/{area}/{entity}api/filter.go

  Tests:
    - api/cmd/services/ichor/tests/{area}/{entity}api/{entity}_test.go
    - api/cmd/services/ichor/tests/{area}/{entity}api/seed_test.go
    - api/cmd/services/ichor/tests/{area}/{entity}api/create_test.go
    - api/cmd/services/ichor/tests/{area}/{entity}api/update_test.go
    - api/cmd/services/ichor/tests/{area}/{entity}api/delete_test.go
    - api/cmd/services/ichor/tests/{area}/{entity}api/query_test.go

Files modified:
    - business/sdk/migrate/sql/migrate.sql (added migration)
    - api/cmd/services/ichor/build/all/all.go (wiring)
    - business/sdk/dbtest/dbtest.go (BusDomain)
    - api/sdk/http/apitest/model.go (SeedData)
    - business/domain/core/tableaccessbus/testutil.go (table access)

Build iterations: N
Test fix iterations: N
Issues auto-resolved:
    - [list each error that was caught and fixed]

Remaining issues (if any):
    - [list any unresolved problems]
```

---

## Common Pitfalls

1. **Package naming conflicts**: Never name a DB struct `page` or `order` — conflicts with SDK packages. Use `db{Entity}` prefix.
2. **Layer violations**: Business layer NEVER imports app or API. App NEVER imports API.
3. **Conversion completeness**: Every field in the business model must appear in both toDB and toBus conversion functions.
4. **Filter/Order consistency**: The same fields must appear in bus filter, db filter, app filter, app order, db order, and bus order.
5. **SeedData field name**: Must match the `apitest.SeedData` struct field name exactly.
6. **BusDomain field name**: Must match the `dbtest.BusDomain` struct field name exactly.
7. **Table access entry**: Must use the schema-qualified table name (e.g., `assets.asset_conditions`).
8. **URL path**: Use kebab-case (hyphens), not snake_case (underscores).
9. **URL param**: Use snake_case with `_id` suffix (e.g., `asset_condition_id`).
10. **JSON tags**: Business models use lowercase/snake_case (e.g., `json:"sort_order"`). App models use camelCase for frontend (e.g., `json:"sortOrder"`).
11. **Validate tags**: Use `required` for NOT NULL fields without defaults in `New*` struct.
12. **Pointer fields**: ALL fields in `Update*` structs must be pointers.
13. **sql.NullString**: Only use in DB model for NULL-able text columns. Business model uses plain `string`.
