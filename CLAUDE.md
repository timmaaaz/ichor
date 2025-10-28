# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Ichor is a production-grade ERP system built using the **Ardan Labs Service Starter Kit** architecture. It implements Domain Driven, Data Oriented Design patterns with full Kubernetes deployment support. The project is a fork/adaptation of the Ardan Labs service architecture specifically for ERP use cases covering HR, Assets, Inventory, Products, Procurement, Sales, and Workflow automation.

**Module**: `github.com/timmaaaz/ichor`
**Go Version**: 1.23
**Architecture**: Ardan Labs Domain-Driven, Data-Oriented Design
**Deployment**: Kubernetes (KIND for local development)
**Database**: PostgreSQL 16.4 with multi-schema design

## Essential Commands

### Development Setup
```bash
# Install Go tooling
make dev-gotooling

# Install Homebrew dependencies (kind, kubectl, kustomize, pgcli, watch)
make dev-brew

# Pull Docker images
make dev-docker
```

### Testing
```bash
# Run all tests with linting and vulnerability checks
make test

# Run tests with race detector
make test-race

# Run only tests (no linting)
make test-only

# Lint code
make lint

# Check for vulnerabilities
make vuln-check

# Shutdown test containers
make test-down
```

### Local Kubernetes Development
```bash
# Start KIND cluster with all services
make dev-up

# Build containers and deploy to KIND
make dev-update-apply

# View logs (formatted)
make dev-logs

# View auth service logs
make dev-logs-auth

# View init container logs
make dev-logs-init

# Restart deployments (after code changes)
make dev-update

# Check pod status
make dev-status

# Shutdown cluster
make dev-down
```

### Database Operations
```bash
# Run migrations
make migrate

# Seed database with test data
make seed

# Seed frontend configuration
make seed-frontend

# Access PostgreSQL CLI
make pgcli

# Recreate database (deletes all data!)
make dev-database-recreate
```

### Docker Compose (Alternative to Kubernetes)
```bash
# Start with existing images
make compose-up

# Build and start
make compose-build-up

# View logs
make compose-logs

# Shutdown
make compose-down
```

### Running Locally (Without Containers)
```bash
# Run main service locally
make run

# Run with help output
make run-help

# Run admin tooling
make admin
```

### Authentication & API Testing
```bash
# Get authentication token
make token

# Export token for subsequent requests
export TOKEN=<COPY_TOKEN_STRING>

# Test users endpoint
make users

# Create new user
make curl-create

# Test liveness probe
make live

# Test readiness probe
make ready
```

### Load Testing
```bash
# Run load test (100 concurrent, 1000 requests)
make load
```

## Architecture

### Layer Structure (Ardan Labs Pattern)

The codebase follows strict layering from top to bottom:

```
┌─────────────────────────────────────────────────────┐
│  api/             HTTP handlers, routes, tests      │
│  ├── domain/http/  Domain-specific HTTP APIs        │
│  ├── cmd/services/  Service entry points            │
│  └── sdk/http/      HTTP framework utilities        │
└─────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────┐
│  app/             Application layer                 │
│  ├── domain/       Domain apps (validation, conv)   │
│  └── sdk/          App-level utilities              │
└─────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────┐
│  business/        Business logic layer              │
│  ├── domain/       Domain business packages (*bus)  │
│  └── sdk/          Business utilities, migration    │
└─────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────┐
│  foundation/      Framework-agnostic utilities      │
│                   (logger, keystore, otel, web)     │
└─────────────────────────────────────────────────────┘
```

**Key Rules:**
- Higher layers can import lower layers, NEVER the reverse
- Business layer contains ALL domain logic
- App layer validates and transforms between API ↔ Business models
- API layer handles HTTP concerns only (routing, middleware, serialization)

### Domain Organization

Domains are organized by business area using PostgreSQL schemas:

- **core/** - Users, roles, permissions, contact info (`core.*` tables)
- **hr/** - Offices, titles, reports-to, homes, approval (`hr.*` tables)
- **geography/** - Countries, regions, cities, streets (`geography.*` tables)
- **assets/** - Asset types, conditions, valid assets, user assets (`assets.*` tables)
- **inventory/** - Warehouses, zones, locations, items, tracking (`inventory.*` tables)
- **products/** - Products, brands, categories, costs, attributes (`products.*` tables)
- **procurement/** - Suppliers, supplier products (`procurement.*` tables)
- **sales/** - Customers, orders, line items, fulfillment (`sales.*` tables)
- **config/** - Table configs, page configs, forms (`config.*` tables)
- **workflow/** - Automation rules, actions, entities (`workflow.*` tables)

### Naming Conventions

**Business Layer Packages** end in `bus`:
- `userbus`, `assetbus`, `productbus`, etc.
- Located in `business/domain/{area}/{entity}bus/`
- Data stores in `business/domain/{area}/{entity}bus/stores/{entity}db/`

**Application Layer Packages** end in `app`:
- `userapp`, `assetapp`, `productapp`, etc.
- Located in `app/domain/{area}/{entity}app/`

**API Layer Packages** end in `api`:
- `userapi`, `assetapi`, `productapi`, etc.
- Located in `api/domain/http/{area}/{entity}api/`

**Model Naming**:
- Creation: `New{Entity}` (e.g., `NewUser`, `NewAsset`)
- Update: `Update{Entity}` (e.g., `UpdateUser`, `UpdateAsset`)
- Response: `{Entity}` (e.g., `User`, `Asset`)

### Service Architecture

**Main Service**: `api/cmd/services/ichor/main.go`
- Entry point for the Ichor API service
- Configures: Database, Auth, OAuth, Tracing, CORS
- Routes can be built in different configurations:
  - `all` - All routes (default)
  - `crud` - Transactional endpoints only
  - `reporting` - Reporting endpoints only

**Auth Service**: `api/cmd/services/auth/` (separate microservice)

**Metrics Service**: `api/cmd/services/metrics/` (observability)

**Route Binding**: `api/cmd/services/ichor/build/`
- `all/all.go` - Binds all domain routes
- `crud/crud.go` - CRUD-only routes
- `reporting/reporting.go` - Reporting-only routes

### Key Patterns

#### Business Layer (`*bus` packages)

```go
// Core structure
type Business struct {
    log      *logger.Logger
    delegate *delegate.Delegate  // Handles UUID generation, time
    storer   Storer              // Interface to database
}

// Always expose interface for storage
type Storer interface {
    Create(ctx context.Context, entity Entity) error
    QueryByID(ctx context.Context, id uuid.UUID) (Entity, error)
    // ... other methods
}
```

#### Application Layer (`*app` packages)

```go
// Converts between API and Business models
// Validates business rules at API boundary

type App struct {
    business *entitybus.Business
}

func (a *App) Create(ctx context.Context, app NewEntity) (Entity, error) {
    // 1. Validate app model
    if err := app.Validate(); err != nil {
        return Entity{}, err
    }

    // 2. Convert app → bus
    bus := toBusNewEntity(app)

    // 3. Call business layer
    busEntity, err := a.business.Create(ctx, bus)
    if err != nil {
        return Entity{}, err
    }

    // 4. Convert bus → app
    return toAppEntity(busEntity), nil
}
```

#### API Layer (`*api` packages)

```go
// HTTP handlers ONLY
func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
    var app appModel.NewEntity
    if err := web.Decode(r, &app); err != nil {
        return errs.NewError(errs.InvalidArgument, err)
    }

    entity, err := api.entityApp.Create(ctx, app)
    if err != nil {
        return errs.NewError(errs.Internal, err)
    }

    return entity
}
```

### Testing

**Integration Tests** are located at:
```
api/cmd/services/ichor/tests/{domain}/{entityapi}/
```

Example test structure:
```go
func Test_{Entity}API(t *testing.T) {
    test := apitest.StartTest(t, "{entityapi_test}")

    // Seed test data
    sd := test.SeedData()

    // Run test tables
    test.Run(t, query200(sd), "query-200")
    test.Run(t, create200(sd), "create-200")
    test.Run(t, update200(sd), "update-200")
}
```

**Test Helpers**:
- `business/sdk/unitest` - Unit test helpers for business layer
- `business/sdk/dbtest` - Database test setup
- Use `apitest.Table` pattern for HTTP integration tests

### Database Migrations

**Location**: `business/sdk/migrate/sql/migrate.sql`

**Format**:
```sql
-- Version: X.YY
-- Description: What this migration does
CREATE TABLE schema.table_name (
    id UUID PRIMARY KEY,
    ...
);
```

**Schemas** are versioned separately:
- Version 0.xx - Schema creation
- Version 1.xx - Core tables
- Version 2.xx - Configuration tables

**Apply migrations**: Use `make migrate` or admin tooling

### FormData Dynamic System

A powerful feature for multi-entity transactional operations. See `FORMDATA_IMPLEMENTATION.md` for complete details.

**Key Points**:
- Registry-based entity registration in `api/cmd/services/ichor/build/all/formdata_registry.go`
- Supports CREATE and UPDATE operations with template variables
- Automatic form validation via reflection on `validate:"required"` tags
- All operations run in a single database transaction

**Adding New Entity** (2 lines per entity):
```go
registry.Register(formdataregistry.EntityRegistration{
    Name: "products",
    CreateModel: productapp.NewProduct{},  // For validation
    UpdateModel: productapp.UpdateProduct{}, // For validation
    DecodeNew: func(data json.RawMessage) (interface{}, error) { /*...*/ },
    CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) { /*...*/ },
    DecodeUpdate: func(data json.RawMessage) (interface{}, error) { /*...*/ },
    UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) { /*...*/ },
})
```

### Authentication & Authorization

**JWT-based** with RSA keys:
- Keys stored in `zarf/keys/` or environment variable `ICHOR_KEYS`
- OAuth providers: Google, Development (for testing)
- Basic auth also supported for testing

**Permission System**:
- Role-based with table-level permissions
- Cached in `permissionsbus`, `rolecache`, `tableaccesscache`
- Check permissions via `PermissionsBus.CheckAccess()`

**Middleware**:
- `mid.Authenticate()` - Validates JWT
- `mid.Authorize()` - Checks permissions

### Caching Strategy

**Sturdyc** used for business layer caching:
- User cache: 1 minute TTL
- Role cache: 60 minutes TTL
- Table access cache: 60 minutes TTL
- Permissions cache: 60 minutes TTL

**Pattern**:
```go
cache := entitycache.NewStore(log,
    entitydb.NewStore(log, db),
    60*time.Minute)
```

### Configuration

**Environment Variables** (prefixed with `ICHOR_`):
- `ICHOR_DB_HOST` - Database host
- `ICHOR_DB_USER` - Database user
- `ICHOR_DB_PASSWORD` - Database password
- `ICHOR_KEYS` - RSA keys for JWT (multiline)
- `ICHOR_WEB_API_HOST` - API listen address

**Config Parsing**: Uses `github.com/ardanlabs/conf/v3`

See `api/cmd/services/ichor/main.go` for complete configuration structure.

### Observability

**Tracing**: OpenTelemetry → Tempo (Grafana stack)
- Configured in `foundation/otel/`
- 5% sampling by default

**Metrics**: Exposed via `/metrics` endpoint
- View with: `make metrics-view`
- Prometheus format

**Logging**: Structured JSON logs
- `foundation/logger/`
- Trace ID injection
- Format tool: `api/cmd/tooling/logfmt/`

**Visualization**:
- Grafana: `make grafana` (http://localhost:3100)
- Statsviz: `make statsviz` (http://localhost:3010/debug/statsviz)

## Development Workflow

### Adding a New Domain Entity

1. **Create Business Layer** (`business/domain/{area}/{entity}bus/`)
   - Define `{entity}bus.go` with Business struct and methods
   - Define `stores/{entity}db/{entity}db.go` for database operations
   - Define models: `Entity`, `NewEntity`, `UpdateEntity`
   - Write unit tests if needed

2. **Create Application Layer** (`app/domain/{area}/{entity}app/`)
   - Define `{entity}app.go` with App struct
   - Define `model.go` with API models and validation
   - Implement conversion functions: `toBus*()`, `toApp*()`

3. **Create API Layer** (`api/domain/http/{area}/{entity}api/`)
   - Define `{entity}api.go` with HTTP handlers
   - Define `route.go` with route configuration
   - Implement CRUD handlers: create, query, queryByID, update, delete

4. **Wire Dependencies** (`api/cmd/services/ichor/build/all/all.go`)
   - Instantiate business layer: `entityBus := entitybus.NewBusiness(...)`
   - Instantiate app layer if needed
   - Call `entityapi.Routes(app, entityapi.Config{...})`

5. **Add Tests** (`api/cmd/services/ichor/tests/{area}/{entity}api/`)
   - Create `{entity}_test.go`, `query_test.go`, `create_test.go`, etc.
   - Seed test data in `seed_test.go`
   - Use `apitest.Table` pattern

6. **Create Migration** (`business/sdk/migrate/sql/migrate.sql`)
   - Add table creation with appropriate schema
   - Follow version numbering convention

### Adding a Domain from SQL Schema (Step-by-Step)

When you have SQL table definitions and need to implement a complete domain, follow this comprehensive checklist. This example uses a hypothetical `pages` table in the `core` schema.

#### Prerequisites

**Given SQL**:
```sql
CREATE TABLE core.pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    path TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    module TEXT NOT NULL,
    icon TEXT,
    sort_order INTEGER DEFAULT 1000,
    is_active BOOLEAN DEFAULT TRUE
);
```

#### Step 1: Add Database Migration

**File**: `business/sdk/migrate/sql/migrate.sql`

1. Find the last version number in the file
2. Add your tables with incremented version numbers
3. Include descriptive comments

```sql
-- Version: 1.28
-- Description: Create table pages
CREATE TABLE core.pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    path TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    module TEXT NOT NULL,
    icon TEXT,
    sort_order INTEGER DEFAULT 1000,
    is_active BOOLEAN DEFAULT TRUE
);
```

**Important**: If inserting between existing versions, renumber all subsequent migrations.

#### Step 2: Business Layer (Core Logic)

**Directory**: `business/domain/core/pagebus/`

##### 2a. Create `model.go`
```go
package pagebus

import "github.com/google/uuid"

type Page struct {
    ID        uuid.UUID
    Path      string
    Name      string
    Module    string
    Icon      string
    SortOrder int
    IsActive  bool
}

type NewPage struct {
    Path      string
    Name      string
    Module    string
    Icon      string
    SortOrder int
    IsActive  bool
}

type UpdatePage struct {
    Path      *string
    Name      *string
    Module    *string
    Icon      *string
    SortOrder *int
    IsActive  *bool
}
```

##### 2b. Create `filter.go`
```go
package pagebus

import "github.com/google/uuid"

type QueryFilter struct {
    ID       *uuid.UUID
    Path     *string
    Name     *string
    Module   *string
    IsActive *bool
}
```

##### 2c. Create `order.go`
```go
package pagebus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderBySortOrder, order.ASC)

const (
    OrderByID        = "id"
    OrderByPath      = "path"
    OrderByName      = "name"
    OrderByModule    = "module"
    OrderBySortOrder = "sort_order"
    OrderByIsActive  = "is_active"
)
```

##### 2d. Create `event.go`
```go
package pagebus

import (
    "encoding/json"
    "fmt"
    "github.com/google/uuid"
    "github.com/timmaaaz/ichor/business/sdk/delegate"
)

const DomainName = "page"

const (
    ActionCreated   = "created"
    ActionUpdated   = "updated"
    ActionDeleted   = "deleted"
)

// Create similar event structures for all actions
// See existing domains for full implementation
```

##### 2e. Create `pagebus.go` (Main business file)
```go
package pagebus

import (
    "context"
    "errors"
    "fmt"
    "github.com/google/uuid"
    "github.com/timmaaaz/ichor/business/sdk/convert"
    "github.com/timmaaaz/ichor/business/sdk/delegate"
    "github.com/timmaaaz/ichor/business/sdk/order"
    "github.com/timmaaaz/ichor/business/sdk/page"
    "github.com/timmaaaz/ichor/business/sdk/sqldb"
    "github.com/timmaaaz/ichor/foundation/logger"
    "github.com/timmaaaz/ichor/foundation/otel"
)

var (
    ErrNotFound = errors.New("page not found")
    ErrUnique   = errors.New("not unique")
)

type Storer interface {
    NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
    Create(ctx context.Context, page Page) error
    Update(ctx context.Context, page Page) error
    Delete(ctx context.Context, page Page) error
    Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Page, error)
    Count(ctx context.Context, filter QueryFilter) (int, error)
    QueryByID(ctx context.Context, pageID uuid.UUID) (Page, error)
}

type Business struct {
    log    *logger.Logger
    storer Storer
    del    *delegate.Delegate
}

func NewBusiness(log *logger.Logger, del *delegate.Delegate, storer Storer) *Business {
    return &Business{log: log, del: del, storer: storer}
}

// Implement CRUD methods: Create, Update, Delete, Query, Count, QueryByID
// Follow pattern from rolebus or other existing domains
```

##### 2f. Create Database Store Files

**Directory**: `business/domain/core/pagebus/stores/pagedb/`

**IMPORTANT**: Avoid naming conflicts with `business/sdk/page`. Use `dbPage` instead of `page` for structs.

**`model.go`**:
```go
package pagedb

import (
    "github.com/google/uuid"
    "github.com/timmaaaz/ichor/business/domain/core/pagebus"
)

type dbPage struct {  // Use dbPage to avoid conflict with page package
    ID        uuid.UUID `db:"id"`
    Path      string    `db:"path"`
    Name      string    `db:"name"`
    Module    string    `db:"module"`
    Icon      string    `db:"icon"`
    SortOrder int       `db:"sort_order"`
    IsActive  bool      `db:"is_active"`
}

func toDBPage(bus pagebus.Page) dbPage {
    return dbPage{/* map fields */}
}

func toBusPage(db dbPage) pagebus.Page {
    return pagebus.Page{/* map fields */}
}

func toBusPages(dbs []dbPage) []pagebus.Page {
    pages := make([]pagebus.Page, len(dbs))
    for i, db := range dbs {
        pages[i] = toBusPage(db)
    }
    return pages
}
```

**`filter.go`**, **`order.go`**, **`pagedb.go`**: Follow patterns from `roledb`

#### Step 3: Application Layer (Validation & Conversion)

**Directory**: `app/domain/core/pageapp/`

##### 3a. Create `model.go`
```go
package pageapp

import (
    "encoding/json"
    "github.com/timmaaaz/ichor/app/sdk/errs"
    "github.com/timmaaaz/ichor/business/domain/core/pagebus"
    "github.com/timmaaaz/ichor/business/sdk/convert"
)

type QueryParams struct {
    Page     string
    Rows     string
    OrderBy  string
    ID       string
    Path     string
    Name     string
    Module   string
    IsActive string
}

type Page struct {
    ID        string `json:"id"`
    Path      string `json:"path"`
    Name      string `json:"name"`
    Module    string `json:"module"`
    Icon      string `json:"icon"`
    SortOrder int    `json:"sortOrder"`
    IsActive  bool   `json:"isActive"`
}

func (app Page) Encode() ([]byte, string, error) {
    data, err := json.Marshal(app)
    return data, "application/json", err
}

type NewPage struct {
    Path      string `json:"path" validate:"required"`
    Name      string `json:"name" validate:"required"`
    Module    string `json:"module" validate:"required"`
    Icon      string `json:"icon"`
    SortOrder int    `json:"sortOrder"`
    IsActive  bool   `json:"isActive"`
}

func (app *NewPage) Decode(data []byte) error {
    return json.Unmarshal(data, &app)
}

func (app NewPage) Validate() error {
    if err := errs.Check(app); err != nil {
        return errs.Newf(errs.InvalidArgument, "validate: %s", err)
    }
    return nil
}

// Conversion functions: ToAppPage(), toBusNewPage(), etc.
```

##### 3b. Create `filter.go`, `order.go`, `pageapp.go`

Follow patterns from `roleapp` for parsing and business layer calls.

#### Step 4: API Layer (HTTP Handlers)

**Directory**: `api/domain/http/core/pageapi/`

##### 4a. Create `pageapi.go`
```go
package pageapi

import (
    "context"
    "net/http"
    "github.com/google/uuid"
    "github.com/timmaaaz/ichor/app/domain/core/pageapp"
    "github.com/timmaaaz/ichor/app/sdk/errs"
    "github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
    pageapp *pageapp.App
}

func newAPI(pageapp *pageapp.App) *api {
    return &api{pageapp: pageapp}
}

// Implement handlers: create, update, delete, query, queryByID
```

##### 4b. Create `route.go`
```go
package pageapi

import (
    "net/http"
    "github.com/timmaaaz/ichor/api/sdk/http/mid"
    "github.com/timmaaaz/ichor/app/domain/core/pageapp"
    "github.com/timmaaaz/ichor/app/sdk/auth"
    "github.com/timmaaaz/ichor/app/sdk/authclient"
    "github.com/timmaaaz/ichor/business/domain/core/pagebus"
    "github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
    "github.com/timmaaaz/ichor/foundation/logger"
    "github.com/timmaaaz/ichor/foundation/web"
)

type Config struct {
    Log            *logger.Logger
    PageBus        *pagebus.Business
    AuthClient     *authclient.Client
    PermissionsBus *permissionsbus.Business
}

const RouteTable = "pages"

func Routes(app *web.App, cfg Config) {
    const version = "v1"
    api := newAPI(pageapp.NewApp(cfg.PageBus))
    authen := mid.Authenticate(cfg.AuthClient)

    // Use auth.RuleAdminOnly for admin-only endpoints
    app.HandlerFunc(http.MethodGet, version, "/core/pages", api.query, authen,
        mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAdminOnly))
    app.HandlerFunc(http.MethodPost, version, "/core/pages", api.create, authen,
        mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAdminOnly))
    // Add other routes...
}
```

##### 4c. Create `filter.go`

#### Step 5: Wire Everything Together

**File**: `api/cmd/services/ichor/build/all/all.go`

##### 5a. Add imports
```go
import (
    "github.com/timmaaaz/ichor/api/domain/http/core/pageapi"
    "github.com/timmaaaz/ichor/app/domain/core/pageapp"
    "github.com/timmaaaz/ichor/business/domain/core/pagebus"
    "github.com/timmaaaz/ichor/business/domain/core/pagebus/stores/pagedb"
)
```

##### 5b. Instantiate business layer (around line 320)
```go
pageBus := pagebus.NewBusiness(cfg.Log, delegate, pagedb.NewStore(cfg.Log, cfg.DB))
```

##### 5c. Register routes (around line 520)
```go
pageapi.Routes(app, pageapi.Config{
    Log:            cfg.Log,
    PageBus:        pageBus,
    AuthClient:     cfg.AuthClient,
    PermissionsBus: permissionsBus,
})
```

#### Step 6: Add Permissions for Tests

**File**: `business/domain/core/tableaccessbus/testutil.go`

Add entries for your new tables:
```go
{RoleID: uuid.Nil, TableName: "pages", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
```

#### Step 7: Register in FormData System (Optional)

If you want your entity to work with multi-entity transactions:

**File**: `api/cmd/services/ichor/build/all/formdata_registry.go`

##### 7a. Add parameter to function signature
```go
func buildFormDataRegistry(
    // ... existing params
    pageApp *pageapp.App,
    // ... remaining params
) (*formdataregistry.Registry, error) {
```

##### 7b. Register entity
```go
if err := registry.Register(formdataregistry.EntityRegistration{
    Name: "pages",
    DecodeNew: func(data json.RawMessage) (interface{}, error) {
        var app pageapp.NewPage
        if err := json.Unmarshal(data, &app); err != nil {
            return nil, err
        }
        if err := app.Validate(); err != nil {
            return nil, err
        }
        return app, nil
    },
    CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
        return pageApp.Create(ctx, model.(pageapp.NewPage))
    },
    CreateModel: pageapp.NewPage{},
    DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
        var app pageapp.UpdatePage
        if err := json.Unmarshal(data, &app); err != nil {
            return nil, err
        }
        if err := app.Validate(); err != nil {
            return nil, err
        }
        return app, nil
    },
    UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
        return pageApp.Update(ctx, model.(pageapp.UpdatePage), id)
    },
    UpdateModel: pageapp.UpdatePage{},
}); err != nil {
    return nil, fmt.Errorf("register pages: %w", err)
}
```

##### 7c. Update call site in `all.go` (around line 730)
```go
formDataRegistry, err := buildFormDataRegistry(
    // ... existing params
    pageapp.NewApp(pageBus),
    // ... remaining params
)
```

#### Step 8: Run Tests

```bash
# Build to check for compilation errors
go build ./api/cmd/services/ichor/...

# Run migrations
make migrate

# Run all tests
make test

# Run tests for your specific domain
go test -v ./api/cmd/services/ichor/tests/core/pageapi
```

#### Common Pitfalls

1. **Naming Conflicts**: Avoid naming database structs the same as SDK packages (e.g., use `dbPage` instead of `page` to avoid conflict with `business/sdk/page`)
2. **Version Numbers**: Always increment and never skip migration version numbers
3. **Import Paths**: Use the full module path: `github.com/timmaaaz/ichor/...`
4. **Auth Rules**: For admin-only endpoints, use `auth.RuleAdminOnly`; for any authenticated user, use `auth.RuleAny`
5. **Pointer Fields**: Use pointers in `Update` structs to distinguish between "not provided" and "provided as zero value"
6. **Validation Tags**: Use `validate:"required"` for required fields in `New` structs
7. **JSON Tags**: Use camelCase in JSON tags to match frontend conventions

### Running a Single Test

```bash
# Run specific test function
go test -v ./api/cmd/services/ichor/tests/{area}/{entity}api -run TestFunctionName

# Run all tests in a package
go test -v ./api/cmd/services/ichor/tests/{area}/{entity}api

# Run with race detector
go test -race -v ./api/cmd/services/ichor/tests/{area}/{entity}api
```

### Debugging

**View database directly**:
```bash
make pgcli
# Then: SELECT * FROM schema.table_name;
```

**View logs**:
```bash
# Service logs (formatted)
make dev-logs

# Raw logs
kubectl logs -n ichor-system -l app=ichor --all-containers=true -f
```

**Describe resources**:
```bash
make dev-describe-ichor     # Ichor pods
make dev-describe-database  # Database pod
make dev-describe-node      # Cluster nodes
```

### Common Issues

**"No keys exist"**: Set `ICHOR_KEYS` or add keys to `zarf/keys/`

**Database connection fails**:
- Check `ICHOR_DB_HOST` matches service name
- Verify database pod is running: `make dev-status`

**Tests failing**:
- Ensure test database is clean: `make test-down` then `make test`
- Check migrations are current: `make migrate`

**Build fails**:
- Verify Go version: `go version` (must be 1.23+)
- Clean and rebuild: `go clean -modcache && go mod download && make build`

## Important Notes

- **Never skip migrations** - Always add new version, never edit existing
- **Business layer is source of truth** - All validation and logic goes here
- **Keep layers pure** - No business logic in API, no HTTP in business
- **Use delegate** - For UUID generation, timestamps (testing seams)
- **Cache carefully** - Only cache read-heavy, infrequently changing data
- **Test everything** - Integration tests are primary test strategy

## Additional Resources

- **Ardan Labs Course**: https://github.com/ardanlabs/service/wiki
- **FormData System**: `FORMDATA_IMPLEMENTATION.md` in this repo
- **Makefile Help**: `make help` for all available commands
