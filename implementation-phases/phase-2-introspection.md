# Phase 2: Database Introspection Domain

**Status**: CRITICAL - Hard Blocker
**Priority**: Highest
**Estimated Time**: 8-10 hours
**Unblocks**: Frontend Phase 4 (Table Builder UI)

---

## Overview

Create a complete new domain (`introspectionbus`) for PostgreSQL schema introspection. This enables the frontend Table Builder UI to discover database structure (schemas, tables, columns, relationships) dynamically.

### What You'll Build

**New Domain**: `introspectionbus` (read-only, stateless)

**4 Endpoints**:
1. `GET /v1/introspection/schemas` - List all database schemas
2. `GET /v1/introspection/schemas/{schema}/tables` - List tables in a schema
3. `GET /v1/introspection/tables/{schema}/{table}/columns` - Get column metadata
4. `GET /v1/introspection/tables/{schema}/{table}/relationships` - Get foreign key relationships

**Key Characteristics**:
- Read-only (no CRUD operations)
- Queries PostgreSQL system catalogs (`information_schema`, `pg_catalog`)
- Admin-only access (`auth.RuleAdminOnly`)
- Stateless (no database writes, no transactions)

---

## Why Introspection?

**Frontend Need**: Table Builder UI requires:
- Schema dropdown (core, hr, assets, sales, etc.)
- Table dropdown (users, offices, products, orders, etc.)
- Column metadata (name, type, nullable, primary key)
- Relationship suggestions (auto-join based on foreign keys)

**Backend Solution**: Query PostgreSQL metadata tables to return this information dynamically.

---

## Architecture

### Domain Structure

```
business/domain/introspectionbus/
├── introspectionbus.go          # Business logic
├── model.go                     # Business models (Schema, Table, Column, Relationship)
└── stores/introspectiondb/
    ├── introspectiondb.go       # Database queries
    └── model.go                 # Database models (dbSchema, dbTable, etc.)

app/domain/introspectionapp/
├── introspectionapp.go          # Application layer
└── model.go                     # App models (with JSON tags, Encoder interface)

api/domain/http/introspectionapi/
├── introspectionapi.go          # HTTP handlers
└── routes.go                    # Route registration
```

### Differences from Standard Domain

**No CRUD Operations**: This domain is read-only, so you won't need:
- Create/Update/Delete methods
- NewEntity/UpdateEntity models
- Transactions

**No Storer Interface**: Since there are no writes, you can skip the Storer interface pattern and call the database layer directly (or use a simplified read-only interface).

**Simplified Pattern**: Follow `purchaseorderstatusapi` (read-only reference data) rather than full CRUD domains.

---

## Step-by-Step Implementation

### Step 1: Create Business Layer Models

**File**: `business/domain/introspectionbus/model.go`

```go
package introspectionbus

// Schema represents a PostgreSQL schema.
type Schema struct {
    Name string
}

// Table represents a table within a schema.
type Table struct {
    Schema         string
    Name           string
    RowCountEstimate int
}

// Column represents a column within a table.
type Column struct {
    Name          string
    DataType      string
    IsNullable    bool
    IsPrimaryKey  bool
    DefaultValue  string
}

// Relationship represents a foreign key relationship.
type Relationship struct {
    ForeignKeyName      string
    ColumnName          string
    ReferencedSchema    string
    ReferencedTable     string
    ReferencedColumn    string
    RelationshipType    string  // "many-to-one", "one-to-many", etc.
}
```

---

### Step 2: Create Business Layer Logic

**File**: `business/domain/introspectionbus/introspectionbus.go`

```go
package introspectionbus

import (
    "context"
    "fmt"

    "github.com/timmaaaz/ichor/business/sdk/sqldb"
    "github.com/timmaaaz/ichor/foundation/logger"
    "github.com/timmaaaz/ichor/foundation/otel"
)

type Business struct {
    log *logger.Logger
    db  sqldb.Execer
}

func NewBusiness(log *logger.Logger, db sqldb.Execer) *Business {
    return &Business{
        log: log,
        db:  db,
    }
}

// QuerySchemas returns all database schemas (excluding system schemas).
func (b *Business) QuerySchemas(ctx context.Context) ([]Schema, error) {
    ctx, span := otel.AddSpan(ctx, "business.introspectionbus.queryschemas")
    defer span.End()

    const q = `
    SELECT
        schema_name
    FROM
        information_schema.schemata
    WHERE
        schema_name NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
    ORDER BY
        schema_name`

    var schemas []Schema
    if err := sqldb.NamedQuerySlice(ctx, b.log, b.db, q, nil, &schemas); err != nil {
        return nil, fmt.Errorf("query schemas: %w", err)
    }

    return schemas, nil
}

// QueryTables returns all tables in a given schema.
func (b *Business) QueryTables(ctx context.Context, schema string) ([]Table, error) {
    ctx, span := otel.AddSpan(ctx, "business.introspectionbus.querytables")
    defer span.End()

    const q = `
    SELECT
        t.table_schema AS schema,
        t.table_name AS name,
        COALESCE(c.reltuples::bigint, 0) AS row_count_estimate
    FROM
        information_schema.tables t
    LEFT JOIN
        pg_class c ON c.relname = t.table_name
    LEFT JOIN
        pg_namespace n ON n.oid = c.relnamespace AND n.nspname = t.table_schema
    WHERE
        t.table_schema = :schema
        AND t.table_type = 'BASE TABLE'
    ORDER BY
        t.table_name`

    data := struct {
        Schema string `db:"schema"`
    }{
        Schema: schema,
    }

    var tables []Table
    if err := sqldb.NamedQuerySlice(ctx, b.log, b.db, q, data, &tables); err != nil {
        return nil, fmt.Errorf("query tables: %w", err)
    }

    return tables, nil
}

// QueryColumns returns all columns for a given table.
func (b *Business) QueryColumns(ctx context.Context, schema, table string) ([]Column, error) {
    ctx, span := otel.AddSpan(ctx, "business.introspectionbus.querycolumns")
    defer span.End()

    const q = `
    SELECT
        c.column_name AS name,
        c.data_type AS data_type,
        c.is_nullable = 'YES' AS is_nullable,
        COALESCE(c.column_default, '') AS default_value,
        EXISTS(
            SELECT 1
            FROM information_schema.table_constraints tc
            JOIN information_schema.key_column_usage kcu
                ON tc.constraint_name = kcu.constraint_name
                AND tc.table_schema = kcu.table_schema
            WHERE tc.constraint_type = 'PRIMARY KEY'
                AND tc.table_schema = :schema
                AND tc.table_name = :table
                AND kcu.column_name = c.column_name
        ) AS is_primary_key
    FROM
        information_schema.columns c
    WHERE
        c.table_schema = :schema
        AND c.table_name = :table
    ORDER BY
        c.ordinal_position`

    data := struct {
        Schema string `db:"schema"`
        Table  string `db:"table"`
    }{
        Schema: schema,
        Table:  table,
    }

    var columns []Column
    if err := sqldb.NamedQuerySlice(ctx, b.log, b.db, q, data, &columns); err != nil {
        return nil, fmt.Errorf("query columns: %w", err)
    }

    return columns, nil
}

// QueryRelationships returns all foreign key relationships for a given table.
func (b *Business) QueryRelationships(ctx context.Context, schema, table string) ([]Relationship, error) {
    ctx, span := otel.AddSpan(ctx, "business.introspectionbus.queryrelationships")
    defer span.End()

    const q = `
    SELECT
        tc.constraint_name AS foreign_key_name,
        kcu.column_name AS column_name,
        ccu.table_schema AS referenced_schema,
        ccu.table_name AS referenced_table,
        ccu.column_name AS referenced_column,
        'many-to-one' AS relationship_type
    FROM
        information_schema.table_constraints tc
    JOIN
        information_schema.key_column_usage kcu
        ON tc.constraint_name = kcu.constraint_name
        AND tc.table_schema = kcu.table_schema
    JOIN
        information_schema.constraint_column_usage ccu
        ON tc.constraint_name = ccu.constraint_name
        AND tc.table_schema = ccu.table_schema
    WHERE
        tc.constraint_type = 'FOREIGN KEY'
        AND tc.table_schema = :schema
        AND tc.table_name = :table
    ORDER BY
        kcu.ordinal_position`

    data := struct {
        Schema string `db:"schema"`
        Table  string `db:"table"`
    }{
        Schema: schema,
        Table:  table,
    }

    var relationships []Relationship
    if err := sqldb.NamedQuerySlice(ctx, b.log, b.db, q, data, &relationships); err != nil {
        return nil, fmt.Errorf("query relationships: %w", err)
    }

    return relationships, nil
}
```

**Key Points**:
- No Storer interface (simplified for read-only)
- Direct SQL queries using `sqldb.NamedQuerySlice`
- OpenTelemetry spans for observability
- Exclude system schemas (pg_catalog, information_schema)

---

### Step 3: Create Application Layer Models

**File**: `app/domain/introspectionapp/model.go`

```go
package introspectionapp

import (
    "encoding/json"

    "github.com/timmaaaz/ichor/business/domain/introspectionbus"
)

// Schema represents a database schema.
type Schema struct {
    Name string `json:"name"`
}

func (app Schema) Encode() ([]byte, string, error) {
    data, err := json.Marshal(app)
    return data, "application/json", err
}

// Schemas is a collection wrapper.
type Schemas []Schema

func (app Schemas) Encode() ([]byte, string, error) {
    data, err := json.Marshal(app)
    return data, "application/json", err
}

// Table represents a database table.
type Table struct {
    Schema           string `json:"schema"`
    Name             string `json:"name"`
    RowCountEstimate int    `json:"rowCountEstimate"`
}

func (app Table) Encode() ([]byte, string, error) {
    data, err := json.Marshal(app)
    return data, "application/json", err
}

// Tables is a collection wrapper.
type Tables []Table

func (app Tables) Encode() ([]byte, string, error) {
    data, err := json.Marshal(app)
    return data, "application/json", err
}

// Column represents a table column.
type Column struct {
    Name         string `json:"name"`
    DataType     string `json:"dataType"`
    IsNullable   bool   `json:"isNullable"`
    IsPrimaryKey bool   `json:"isPrimaryKey"`
    DefaultValue string `json:"defaultValue"`
}

func (app Column) Encode() ([]byte, string, error) {
    data, err := json.Marshal(app)
    return data, "application/json", err
}

// Columns is a collection wrapper.
type Columns []Column

func (app Columns) Encode() ([]byte, string, error) {
    data, err := json.Marshal(app)
    return data, "application/json", err
}

// Relationship represents a foreign key relationship.
type Relationship struct {
    ForeignKeyName   string `json:"foreignKeyName"`
    ColumnName       string `json:"columnName"`
    ReferencedSchema string `json:"referencedSchema"`
    ReferencedTable  string `json:"referencedTable"`
    ReferencedColumn string `json:"referencedColumn"`
    RelationshipType string `json:"relationshipType"`
}

func (app Relationship) Encode() ([]byte, string, error) {
    data, err := json.Marshal(app)
    return data, "application/json", err
}

// Relationships is a collection wrapper.
type Relationships []Relationship

func (app Relationships) Encode() ([]byte, string, error) {
    data, err := json.Marshal(app)
    return data, "application/json", err
}

// Conversion functions

func ToAppSchema(bus introspectionbus.Schema) Schema {
    return Schema{
        Name: bus.Name,
    }
}

func ToAppSchemas(bus []introspectionbus.Schema) []Schema {
    schemas := make([]Schema, len(bus))
    for i, s := range bus {
        schemas[i] = ToAppSchema(s)
    }
    return schemas
}

func ToAppTable(bus introspectionbus.Table) Table {
    return Table{
        Schema:           bus.Schema,
        Name:             bus.Name,
        RowCountEstimate: bus.RowCountEstimate,
    }
}

func ToAppTables(bus []introspectionbus.Table) []Table {
    tables := make([]Table, len(bus))
    for i, t := range bus {
        tables[i] = ToAppTable(t)
    }
    return tables
}

func ToAppColumn(bus introspectionbus.Column) Column {
    return Column{
        Name:         bus.Name,
        DataType:     bus.DataType,
        IsNullable:   bus.IsNullable,
        IsPrimaryKey: bus.IsPrimaryKey,
        DefaultValue: bus.DefaultValue,
    }
}

func ToAppColumns(bus []introspectionbus.Column) []Column {
    columns := make([]Column, len(bus))
    for i, c := range bus {
        columns[i] = ToAppColumn(c)
    }
    return columns
}

func ToAppRelationship(bus introspectionbus.Relationship) Relationship {
    return Relationship{
        ForeignKeyName:   bus.ForeignKeyName,
        ColumnName:       bus.ColumnName,
        ReferencedSchema: bus.ReferencedSchema,
        ReferencedTable:  bus.ReferencedTable,
        ReferencedColumn: bus.ReferencedColumn,
        RelationshipType: bus.RelationshipType,
    }
}

func ToAppRelationships(bus []introspectionbus.Relationship) []Relationship {
    relationships := make([]Relationship, len(bus))
    for i, r := range bus {
        relationships[i] = ToAppRelationship(r)
    }
    return relationships
}
```

**Key Points**:
- JSON tags use camelCase (frontend convention)
- All collection types implement `Encoder` interface
- Conversion functions map business → app models

---

### Step 4: Create Application Layer Logic

**File**: `app/domain/introspectionapp/introspectionapp.go`

```go
package introspectionapp

import (
    "context"

    "github.com/timmaaaz/ichor/app/sdk/errs"
    "github.com/timmaaaz/ichor/business/domain/introspectionbus"
)

type App struct {
    business *introspectionbus.Business
}

func NewApp(business *introspectionbus.Business) *App {
    return &App{
        business: business,
    }
}

func (a *App) QuerySchemas(ctx context.Context) (Schemas, error) {
    schemas, err := a.business.QuerySchemas(ctx)
    if err != nil {
        return nil, errs.Newf(errs.Internal, "query schemas: %s", err)
    }

    return Schemas(ToAppSchemas(schemas)), nil
}

func (a *App) QueryTables(ctx context.Context, schema string) (Tables, error) {
    tables, err := a.business.QueryTables(ctx, schema)
    if err != nil {
        return nil, errs.Newf(errs.Internal, "query tables: %s", err)
    }

    return Tables(ToAppTables(tables)), nil
}

func (a *App) QueryColumns(ctx context.Context, schema, table string) (Columns, error) {
    columns, err := a.business.QueryColumns(ctx, schema, table)
    if err != nil {
        return nil, errs.Newf(errs.Internal, "query columns: %s", err)
    }

    return Columns(ToAppColumns(columns)), nil
}

func (a *App) QueryRelationships(ctx context.Context, schema, table string) (Relationships, error) {
    relationships, err := a.business.QueryRelationships(ctx, schema, table)
    if err != nil {
        return nil, errs.Newf(errs.Internal, "query relationships: %s", err)
    }

    return Relationships(ToAppRelationships(relationships)), nil
}
```

---

### Step 5: Create API Layer Handlers

**File**: `api/domain/http/introspectionapi/introspectionapi.go`

```go
package introspectionapi

import (
    "context"
    "net/http"

    "github.com/timmaaaz/ichor/app/domain/introspectionapp"
    "github.com/timmaaaz/ichor/app/sdk/errs"
    "github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
    introspectionApp *introspectionapp.App
}

func newAPI(introspectionApp *introspectionapp.App) *api {
    return &api{
        introspectionApp: introspectionApp,
    }
}

func (api *api) querySchemas(ctx context.Context, r *http.Request) web.Encoder {
    schemas, err := api.introspectionApp.QuerySchemas(ctx)
    if err != nil {
        return errs.NewError(err)
    }

    return schemas
}

func (api *api) queryTables(ctx context.Context, r *http.Request) web.Encoder {
    schema := web.Param(r, "schema")

    tables, err := api.introspectionApp.QueryTables(ctx, schema)
    if err != nil {
        return errs.NewError(err)
    }

    return tables
}

func (api *api) queryColumns(ctx context.Context, r *http.Request) web.Encoder {
    schema := web.Param(r, "schema")
    table := web.Param(r, "table")

    columns, err := api.introspectionApp.QueryColumns(ctx, schema, table)
    if err != nil {
        return errs.NewError(err)
    }

    return columns
}

func (api *api) queryRelationships(ctx context.Context, r *http.Request) web.Encoder {
    schema := web.Param(r, "schema")
    table := web.Param(r, "table")

    relationships, err := api.introspectionApp.QueryRelationships(ctx, schema, table)
    if err != nil {
        return errs.NewError(err)
    }

    return relationships
}
```

**Key Points**:
- Extract path parameters using `web.Param(r, "name")`
- No request body decoding (GET endpoints)
- Return wrapper types (already implement `web.Encoder`)

---

### Step 6: Register Routes

**File**: `api/domain/http/introspectionapi/routes.go`

```go
package introspectionapi

import (
    "net/http"

    "github.com/timmaaaz/ichor/api/sdk/http/mid"
    "github.com/timmaaaz/ichor/app/domain/introspectionapp"
    "github.com/timmaaaz/ichor/app/sdk/auth"
    "github.com/timmaaaz/ichor/app/sdk/authclient"
    "github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
    "github.com/timmaaaz/ichor/business/domain/introspectionbus"
    "github.com/timmaaaz/ichor/foundation/logger"
    "github.com/timmaaaz/ichor/foundation/web"
)

type Config struct {
    Log             *logger.Logger
    IntrospectionBus *introspectionbus.Business
    AuthClient      *authclient.Client
    PermissionsBus  *permissionsbus.Business
}

const RouteTable = "introspection"

func Routes(app *web.App, cfg Config) {
    const version = "v1"
    api := newAPI(introspectionapp.NewApp(cfg.IntrospectionBus))
    authen := mid.Authenticate(cfg.AuthClient)

    // GET /v1/introspection/schemas
    app.HandlerFunc(http.MethodGet, version, "/introspection/schemas", api.querySchemas, authen,
        mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAdminOnly))

    // GET /v1/introspection/schemas/{schema}/tables
    app.HandlerFunc(http.MethodGet, version, "/introspection/schemas/:schema/tables", api.queryTables, authen,
        mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAdminOnly))

    // GET /v1/introspection/tables/{schema}/{table}/columns
    app.HandlerFunc(http.MethodGet, version, "/introspection/tables/:schema/:table/columns", api.queryColumns, authen,
        mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAdminOnly))

    // GET /v1/introspection/tables/{schema}/{table}/relationships
    app.HandlerFunc(http.MethodGet, version, "/introspection/tables/:schema/:table/relationships", api.queryRelationships, authen,
        mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAdminOnly))
}
```

**Key Points**:
- All routes use `auth.RuleAdminOnly` (introspection is admin-only)
- Path parameters use `:param` syntax
- Action: `permissionsbus.Actions.Read`

---

### Step 7: Wire Domain in Main Service

**File**: `api/cmd/services/ichor/build/all/all.go`

#### 7a. Add Imports

```go
import (
    // ... existing imports
    "github.com/timmaaaz/ichor/api/domain/http/introspectionapi"
    "github.com/timmaaaz/ichor/business/domain/introspectionbus"
)
```

#### 7b. Instantiate Business Layer (around line 320)

```go
// Introspection domain
introspectionBus := introspectionbus.NewBusiness(cfg.Log, cfg.DB)
```

#### 7c. Register Routes (around line 520)

```go
introspectionapi.Routes(app, introspectionapi.Config{
    Log:              cfg.Log,
    IntrospectionBus: introspectionBus,
    AuthClient:       cfg.AuthClient,
    PermissionsBus:   permissionsBus,
})
```

---

### Step 8: Add Permissions for Testing

**File**: `business/domain/core/tableaccessbus/testutil.go`

Add introspection table to test permissions:

```go
{RoleID: uuid.Nil, TableName: "introspection", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
```

---

## Testing

### Manual Testing

```bash
# Build and deploy
make dev-up
make dev-update-apply

# Get admin token
export TOKEN=$(make token | grep -o '"token":"[^"]*' | cut -d'"' -f4)

# Test schemas endpoint
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:3000/v1/introspection/schemas | jq

# Expected response:
# [
#   {"name": "core"},
#   {"name": "hr"},
#   {"name": "assets"},
#   ...
# ]

# Test tables endpoint
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:3000/v1/introspection/schemas/core/tables | jq

# Expected response:
# [
#   {"schema": "core", "name": "users", "rowCountEstimate": 150},
#   {"schema": "core", "name": "roles", "rowCountEstimate": 5},
#   ...
# ]

# Test columns endpoint
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:3000/v1/introspection/tables/core/users/columns | jq

# Expected response:
# [
#   {"name": "id", "dataType": "uuid", "isNullable": false, "isPrimaryKey": true, "defaultValue": "gen_random_uuid()"},
#   {"name": "email", "dataType": "text", "isNullable": false, "isPrimaryKey": false, "defaultValue": ""},
#   ...
# ]

# Test relationships endpoint
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:3000/v1/introspection/tables/hr/offices/relationships | jq

# Expected response:
# [
#   {
#     "foreignKeyName": "fk_city_id",
#     "columnName": "city_id",
#     "referencedSchema": "geography",
#     "referencedTable": "cities",
#     "referencedColumn": "id",
#     "relationshipType": "many-to-one"
#   }
# ]
```

### Integration Tests

**Location**: `api/cmd/services/ichor/tests/introspectionapi/`

**Create**: `introspection_test.go`

```go
package introspectionapi_test

import (
    "net/http"
    "testing"

    "github.com/timmaaaz/ichor/api/sdk/apitest"
)

func Test_IntrospectionAPI(t *testing.T) {
    test := apitest.StartTest(t, "introspectionapi_test")
    sd := test.SeedData()

    test.Run(t, querySchemas200(sd), "query-schemas-200")
    test.Run(t, queryTables200(sd), "query-tables-200")
    test.Run(t, queryColumns200(sd), "query-columns-200")
    test.Run(t, queryRelationships200(sd), "query-relationships-200")
}

func querySchemas200(sd apitest.SeedData) apitest.Table {
    return apitest.Table{
        Name:       "query-schemas-200",
        URL:        "/v1/introspection/schemas",
        Method:     http.MethodGet,
        Token:      sd.Admins[0].Token,
        StatusCode: http.StatusOK,
    }
}

func queryTables200(sd apitest.SeedData) apitest.Table {
    return apitest.Table{
        Name:       "query-tables-200",
        URL:        "/v1/introspection/schemas/core/tables",
        Method:     http.MethodGet,
        Token:      sd.Admins[0].Token,
        StatusCode: http.StatusOK,
    }
}

func queryColumns200(sd apitest.SeedData) apitest.Table {
    return apitest.Table{
        Name:       "query-columns-200",
        URL:        "/v1/introspection/tables/core/users/columns",
        Method:     http.MethodGet,
        Token:      sd.Admins[0].Token,
        StatusCode: http.StatusOK,
    }
}

func queryRelationships200(sd apitest.SeedData) apitest.Table {
    return apitest.Table{
        Name:       "query-relationships-200",
        URL:        "/v1/introspection/tables/hr/offices/relationships",
        Method:     http.MethodGet,
        Token:      sd.Admins[0].Token,
        StatusCode: http.StatusOK,
    }
}
```

**Run Tests**:
```bash
go test -v ./api/cmd/services/ichor/tests/introspectionapi
```

---

## Complete Implementation Checklist

### Business Layer
- [ ] Create `business/domain/introspectionbus/model.go`
  - [ ] Define `Schema`, `Table`, `Column`, `Relationship` structs
- [ ] Create `business/domain/introspectionbus/introspectionbus.go`
  - [ ] Implement `QuerySchemas()`
  - [ ] Implement `QueryTables()`
  - [ ] Implement `QueryColumns()`
  - [ ] Implement `QueryRelationships()`

### Application Layer
- [ ] Create `app/domain/introspectionapp/model.go`
  - [ ] Define app models with JSON tags
  - [ ] Implement `Encode()` for all models
  - [ ] Create conversion functions
- [ ] Create `app/domain/introspectionapp/introspectionapp.go`
  - [ ] Implement `QuerySchemas()`
  - [ ] Implement `QueryTables()`
  - [ ] Implement `QueryColumns()`
  - [ ] Implement `QueryRelationships()`

### API Layer
- [ ] Create `api/domain/http/introspectionapi/introspectionapi.go`
  - [ ] Implement `querySchemas()` handler
  - [ ] Implement `queryTables()` handler
  - [ ] Implement `queryColumns()` handler
  - [ ] Implement `queryRelationships()` handler
- [ ] Create `api/domain/http/introspectionapi/routes.go`
  - [ ] Register all 4 routes with admin-only authorization

### Integration
- [ ] Update `api/cmd/services/ichor/build/all/all.go`
  - [ ] Import introspection packages
  - [ ] Instantiate `introspectionBus`
  - [ ] Register routes
- [ ] Update `business/domain/core/tableaccessbus/testutil.go`
  - [ ] Add introspection table permissions

### Testing
- [ ] Create integration tests
- [ ] Manual testing of all 4 endpoints
- [ ] Verify admin-only authorization

---

## Common Issues

### Issue 1: SQL Query Returns Empty Results

**Error**: Endpoints return `[]` but schemas/tables exist

**Fix**: Check query parameters:
```sql
-- Debug: List all schemas
SELECT schema_name FROM information_schema.schemata;

-- Debug: Check if filtering is correct
SELECT schema_name FROM information_schema.schemata
WHERE schema_name NOT IN ('pg_catalog', 'information_schema', 'pg_toast');
```

### Issue 2: Authorization Fails

**Error**: 403 Forbidden

**Fix**: Ensure:
1. Using admin token: `sd.Admins[0].Token`
2. `RouteTable = "introspection"` matches test permissions
3. `auth.RuleAdminOnly` is set correctly

### Issue 3: Path Parameters Not Extracted

**Error**: `schema` or `table` is empty string

**Fix**: Check route definition:
- ✅ `/introspection/schemas/:schema/tables` (correct)
- ❌ `/introspection/schemas/{schema}/tables` (wrong)

### Issue 4: Foreign Keys Not Showing

**Error**: Relationships endpoint returns `[]`

**Fix**: Test with tables that have foreign keys:
```bash
# Test with hr.offices (has city_id FK)
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:3000/v1/introspection/tables/hr/offices/relationships | jq
```

---

## Success Criteria

### Functional
- [ ] All 4 endpoints return HTTP 200
- [ ] Schemas endpoint returns all user schemas (core, hr, assets, etc.)
- [ ] Tables endpoint returns all tables in a schema
- [ ] Columns endpoint returns metadata (name, type, nullable, PK)
- [ ] Relationships endpoint returns foreign key relationships

### Authorization
- [ ] Admin users can access all endpoints
- [ ] Non-admin users get 403 Forbidden
- [ ] Unauthenticated requests get 401 Unauthorized

### Code Quality
- [ ] Follows Ardan Labs architecture (bus → app → api)
- [ ] OpenTelemetry spans added
- [ ] Error handling consistent
- [ ] All models implement `Encoder` interface

### Testing
- [ ] Integration tests pass
- [ ] Manual testing successful
- [ ] Frontend can consume endpoints

---

## Estimated Time Breakdown

- **Business Layer**: 2 hours
- **Application Layer**: 1 hour
- **API Layer**: 1 hour
- **Integration/Wiring**: 1 hour
- **Testing**: 2 hours
- **Debugging**: 1-2 hours

**Total**: 8-10 hours

---

## Next Steps

After completing Phase 2:
1. Test all endpoints manually with curl
2. Run integration tests: `make test`
3. Commit changes: `git commit -m "feat: add database introspection domain"`
4. **Notify frontend team**: Phase 4 (Table Builder) is unblocked
5. Move to [Phase 3: Import/Export](phase-3-import-export.md)

---

## Questions?

**Q: Why admin-only access?**
A: Introspection reveals database structure, which may be considered sensitive. If you want to allow all users, change `auth.RuleAdminOnly` to `auth.RuleAny`.

**Q: Should I cache introspection results?**
A: Schema metadata changes infrequently, so caching could improve performance. Consider adding a 60-minute cache using Sturdyc (see `business/domain/core/userbus/stores/usercache/` for example).

**Q: What about views and materialized views?**
A: Current implementation only returns `BASE TABLE`. To include views, modify the query:
```sql
WHERE t.table_type IN ('BASE TABLE', 'VIEW')
```

**Q: How do I handle composite foreign keys?**
A: Current implementation returns one row per column. For composite keys, group by `constraint_name` in the frontend.

---

**Ready to implement?** This is the highest priority phase. Complete this to unblock frontend Phase 4 immediately. Good luck!
