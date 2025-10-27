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
