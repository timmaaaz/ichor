# Domain Implementation Guide

This guide walks through implementing a complete domain from a SQL table definition. It covers all layers of the Ardan Labs architecture: business, application, and API.

## Overview

When you have SQL table definitions and need to implement a complete domain, follow this comprehensive checklist. This example uses a hypothetical `pages` table in the `core` schema.

## Prerequisites

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

## Step 1: Add Database Migration

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

## Step 2: Business Layer (Core Logic)

**Directory**: `business/domain/core/pagebus/`

### 2a. Create `model.go`
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

### 2b. Create `filter.go`
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

### 2c. Create `order.go`
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

### 2d. Create `event.go`
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

### 2e. Create `pagebus.go` (Main business file)
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

### 2f. Create Database Store Files

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

## Step 3: Application Layer (Validation & Conversion)

**Directory**: `app/domain/core/pageapp/`

### 3a. Create `model.go`
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

### 3b. Create `filter.go`, `order.go`, `pageapp.go`

Follow patterns from `roleapp` for parsing and business layer calls.

## Step 4: API Layer (HTTP Handlers)

**Directory**: `api/domain/http/core/pageapi/`

### 4a. Create `pageapi.go`
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

### 4b. Create `route.go`
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

### 4c. Create `filter.go`

## Step 5: Wire Everything Together

**File**: `api/cmd/services/ichor/build/all/all.go`

### 5a. Add imports
```go
import (
    "github.com/timmaaaz/ichor/api/domain/http/core/pageapi"
    "github.com/timmaaaz/ichor/app/domain/core/pageapp"
    "github.com/timmaaaz/ichor/business/domain/core/pagebus"
    "github.com/timmaaaz/ichor/business/domain/core/pagebus/stores/pagedb"
)
```

### 5b. Instantiate business layer (around line 320)
```go
pageBus := pagebus.NewBusiness(cfg.Log, delegate, pagedb.NewStore(cfg.Log, cfg.DB))
```

### 5c. Register routes (around line 520)
```go
pageapi.Routes(app, pageapi.Config{
    Log:            cfg.Log,
    PageBus:        pageBus,
    AuthClient:     cfg.AuthClient,
    PermissionsBus: permissionsBus,
})
```

## Step 6: Add Permissions for Tests

**File**: `business/domain/core/tableaccessbus/testutil.go`

Add entries for your new tables:
```go
{RoleID: uuid.Nil, TableName: "pages", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
```

## Step 7: Register in FormData System (Optional)

If you want your entity to work with multi-entity transactions:

**File**: `api/cmd/services/ichor/build/all/formdata_registry.go`

### 7a. Add parameter to function signature
```go
func buildFormDataRegistry(
    // ... existing params
    pageApp *pageapp.App,
    // ... remaining params
) (*formdataregistry.Registry, error) {
```

### 7b. Register entity
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

### 7c. Update call site in `all.go` (around line 730)
```go
formDataRegistry, err := buildFormDataRegistry(
    // ... existing params
    pageapp.NewApp(pageBus),
    // ... remaining params
)
```

## Step 8: Run Tests

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

## Common Pitfalls

1. **Naming Conflicts**: Avoid naming database structs the same as SDK packages (e.g., use `dbPage` instead of `page` to avoid conflict with `business/sdk/page`)
2. **Version Numbers**: Always increment and never skip migration version numbers
3. **Import Paths**: Use the full module path: `github.com/timmaaaz/ichor/...`
4. **Auth Rules**: For admin-only endpoints, use `auth.RuleAdminOnly`; for any authenticated user, use `auth.RuleAny`
5. **Pointer Fields**: Use pointers in `Update` structs to distinguish between "not provided" and "provided as zero value"
6. **Validation Tags**: Use `validate:"required"` for required fields in `New` structs
7. **JSON Tags**: Use camelCase in JSON tags to match frontend conventions

## Quick Reference: Files to Create

| Layer | Directory | Files |
|-------|-----------|-------|
| Business | `business/domain/{area}/{entity}bus/` | `model.go`, `filter.go`, `order.go`, `event.go`, `{entity}bus.go` |
| Business Store | `business/domain/{area}/{entity}bus/stores/{entity}db/` | `model.go`, `filter.go`, `order.go`, `{entity}db.go` |
| Application | `app/domain/{area}/{entity}app/` | `model.go`, `filter.go`, `order.go`, `{entity}app.go` |
| API | `api/domain/http/{area}/{entity}api/` | `{entity}api.go`, `route.go`, `filter.go` |
| Tests | `api/cmd/services/ichor/tests/{area}/{entity}api/` | `{entity}_test.go`, `seed_test.go`, `query_test.go`, etc. |

## See Also

- [Layer Patterns](layer-patterns.md) - Encoder/Decoder interfaces, Storer pattern
- [Financial Calculations](financial-calculations.md) - Decimal arithmetic for money
- [Debugging Guide](debugging.md) - Troubleshooting common issues
