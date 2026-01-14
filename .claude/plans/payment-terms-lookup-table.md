# Payment Terms Lookup Table Implementation Plan

## Summary

Convert `payment_terms` from a free-form VARCHAR/TEXT field to a FK reference to a new `core.payment_terms` lookup table. This enables dropdown selection in dynamic forms and ensures data consistency.

## User Requirements
- **Schema**: `core` (shared across sales + procurement)
- **Structure**: Simple - name + description only (no colors/icons)
- **Seed data**: Comprehensive set of common payment terms with hardcoded UUIDs

---

## Implementation Phases

### Phase 1: Database Migration

**File**: [migrate.sql](business/sdk/migrate/sql/migrate.sql)

#### 1a. Add new payment_terms table

Insert as **Version 1.42** (before suppliers at 1.43). This requires renumbering all migrations from current 1.42 through 1.78 by incrementing each version by 1.

**Version renumbering required**:
- Current 1.42 → 1.43
- Current 1.43 (suppliers) → 1.44
- ... continue incrementing through 1.78 → 1.79

**New Version 1.42**:
```sql
-- Version: 1.42
-- Description: Create payment_terms lookup table
CREATE TABLE core.payment_terms (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   name VARCHAR(100) UNIQUE NOT NULL,
   description TEXT
);
```

#### 1b. Modify sales.orders table definition (Version 1.57, formerly 1.56)

Change:
```sql
   payment_terms VARCHAR(50) NULL,
```
To:
```sql
   payment_terms_id UUID NULL,
```

Add FK constraint at bottom of CREATE TABLE:
```sql
   FOREIGN KEY (payment_terms_id) REFERENCES core.payment_terms(id),
```

#### 1c. Modify procurement.suppliers table definition (Version 1.44, formerly 1.43)

Change:
```sql
   payment_terms TEXT NOT NULL,
```
To:
```sql
   payment_terms_id UUID NULL,
```

Add FK constraint at bottom of CREATE TABLE:
```sql
   FOREIGN KEY (payment_terms_id) REFERENCES core.payment_terms(id),
```

**File**: [seed.sql](business/sdk/migrate/sql/seed.sql)

#### 1d. Add payment terms seed data with hardcoded UUIDs

Use hardcoded UUIDs for predictable test seeding:

```sql
-- Payment Terms lookup data (hardcoded UUIDs for test predictability)
INSERT INTO core.payment_terms (id, name, description) VALUES
    -- Standard Net Terms (first 5 are commonly used in tests)
    ('a0000000-0000-0000-0000-000000000001', 'Net 30', 'Payment due within 30 days of invoice date'),
    ('a0000000-0000-0000-0000-000000000002', 'Net 60', 'Payment due within 60 days of invoice date'),
    ('a0000000-0000-0000-0000-000000000003', 'Due on Receipt', 'Payment due immediately upon receipt of invoice'),
    ('a0000000-0000-0000-0000-000000000004', 'Net 15', 'Payment due within 15 days of invoice date'),
    ('a0000000-0000-0000-0000-000000000005', 'Net 45', 'Payment due within 45 days of invoice date'),
    -- Additional Net Terms
    ('a0000000-0000-0000-0000-000000000006', 'Net 7', 'Payment due within 7 days of invoice date'),
    ('a0000000-0000-0000-0000-000000000007', 'Net 10', 'Payment due within 10 days of invoice date'),
    ('a0000000-0000-0000-0000-000000000008', 'Net 21', 'Payment due within 21 days of invoice date'),
    ('a0000000-0000-0000-0000-000000000009', 'Net 90', 'Payment due within 90 days of invoice date'),
    -- Prepayment Terms
    ('a0000000-0000-0000-0000-000000000010', 'Prepaid', 'Full payment required before order fulfillment'),
    ('a0000000-0000-0000-0000-000000000011', 'COD', 'Cash on Delivery - payment due at time of delivery'),
    ('a0000000-0000-0000-0000-000000000012', 'CIA', 'Cash in Advance - full payment before shipping'),
    ('a0000000-0000-0000-0000-000000000013', '50% Deposit', '50% payment due upfront, balance on delivery'),
    -- End of Month Terms
    ('a0000000-0000-0000-0000-000000000014', 'EOM', 'Payment due at end of month'),
    ('a0000000-0000-0000-0000-000000000015', 'MFI', 'Month Following Invoice - due end of next month'),
    ('a0000000-0000-0000-0000-000000000016', '15 MFI', '15 days after month following invoice'),
    -- Credit Terms
    ('a0000000-0000-0000-0000-000000000017', 'Open Account', 'Standard credit account with flexible terms'),
    ('a0000000-0000-0000-0000-000000000018', 'Letter of Credit', 'Payment secured by letter of credit');
```

#### 1e. Add table_access permission for payment_terms (in the core schema section):

```sql
    (gen_random_uuid(), '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'core.payment_terms', true, true, true, true),
```

Insert alphabetically in the core schema section (after `core.pages`, before `core.role_pages`).

---

### Phase 2: New paymenttermbus Package

**Directory**: `business/domain/core/paymenttermbus/`

Following [assetconditionbus](business/domain/assets/assetconditionbus/) pattern with **singular naming**:

| File | Contents |
|------|----------|
| `model.go` | PaymentTerm, NewPaymentTerm, UpdatePaymentTerm structs with JSON tags |
| `filter.go` | QueryFilter with ID, Name, Description |
| `order.go` | OrderBy constants: OrderByID, OrderByName, OrderByDescription |
| `event.go` | Domain events (DomainName="paymentterm", EntityName="payment_terms") |
| `paymenttermbus.go` | Business struct with CRUD methods + QueryAll |
| `testutil.go` | Test helpers for seeding |

#### 2a. model.go

```go
package paymenttermbus

import "github.com/google/uuid"

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type PaymentTerm struct {
    ID          uuid.UUID `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
}

type NewPaymentTerm struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

type UpdatePaymentTerm struct {
    Name        *string `json:"name,omitempty"`
    Description *string `json:"description,omitempty"`
}
```

#### 2b. event.go

```go
package paymenttermbus

const DomainName = "paymentterm"   // Singular, no underscore
const EntityName = "payment_terms" // Matches table name

const (
    ActionCreated = "created"
    ActionUpdated = "updated"
    ActionDeleted = "deleted"
)

// ActionCreatedParms, ActionUpdatedParms, ActionDeletedParms
// Follow pattern from assetconditionbus/event.go
```

#### 2c. paymenttermbus.go - Include QueryAll method

```go
// QueryAll retrieves all payment terms for dropdown population (no pagination).
func (b *Business) QueryAll(ctx context.Context) ([]PaymentTerm, error) {
    ctx, span := otel.AddSpan(ctx, "business.paymenttermbus.QueryAll")
    defer span.End()

    paymentTerms, err := b.storer.QueryAll(ctx)
    if err != nil {
        return nil, fmt.Errorf("queryall: %w", err)
    }

    return paymentTerms, nil
}
```

#### 2d. testutil.go

```go
package paymenttermbus

import (
    "context"
    "fmt"
    "math/rand"
    "sort"
)

// Hardcoded UUIDs for common test payment terms (matches seed.sql)
var (
    Net30ID        = uuid.MustParse("a0000000-0000-0000-0000-000000000001")
    Net60ID        = uuid.MustParse("a0000000-0000-0000-0000-000000000002")
    DueOnReceiptID = uuid.MustParse("a0000000-0000-0000-0000-000000000003")
    Net15ID        = uuid.MustParse("a0000000-0000-0000-0000-000000000004")
    Net45ID        = uuid.MustParse("a0000000-0000-0000-0000-000000000005")
)

// TestNewPaymentTerms generates n new payment terms for testing.
func TestNewPaymentTerms(n int) []NewPaymentTerm {
    newPaymentTerms := make([]NewPaymentTerm, n)
    idx := rand.Intn(10000)
    for i := 0; i < n; i++ {
        idx++
        newPaymentTerms[i] = NewPaymentTerm{
            Name:        fmt.Sprintf("PaymentTerm%d", idx),
            Description: fmt.Sprintf("PaymentTerm%d Description", idx),
        }
    }
    return newPaymentTerms
}

// TestSeedPaymentTerms creates n payment terms in the database for testing.
func TestSeedPaymentTerms(ctx context.Context, n int, api *Business) ([]PaymentTerm, error) {
    newPaymentTerms := TestNewPaymentTerms(n)
    paymentTerms := make([]PaymentTerm, len(newPaymentTerms))
    for i, np := range newPaymentTerms {
        paymentTerm, err := api.Create(ctx, np)
        if err != nil {
            return nil, fmt.Errorf("seeding payment term: idx: %d : %w", i, err)
        }
        paymentTerms[i] = paymentTerm
    }

    sort.Slice(paymentTerms, func(i, j int) bool {
        return paymentTerms[i].Name <= paymentTerms[j].Name
    })

    return paymentTerms, nil
}
```

**Store Directory**: `business/domain/core/paymenttermbus/stores/paymenttermdb/`

| File | Contents |
|------|----------|
| `model.go` | `paymentTerm` struct (lowercase, unexported) with db tags |
| `filter.go` | Filter to SQL WHERE clauses |
| `order.go` | Order field mapping |
| `paymenttermdb.go` | Store with SQL operations including QueryAll |

#### Store model.go - Use lowercase struct name

```go
package paymenttermdb

import (
    "database/sql"
    "github.com/google/uuid"
    "github.com/timmaaaz/ichor/business/domain/core/paymenttermbus"
)

// paymentTerm - lowercase to avoid conflict with business/sdk/page package
type paymentTerm struct {
    ID          uuid.UUID      `db:"id"`
    Name        string         `db:"name"`
    Description sql.NullString `db:"description"`
}

func toDBPaymentTerm(bus paymenttermbus.PaymentTerm) paymentTerm {
    return paymentTerm{
        ID:   bus.ID,
        Name: bus.Name,
        Description: sql.NullString{
            String: bus.Description,
            Valid:  bus.Description != "",
        },
    }
}

func toBusPaymentTerm(db paymentTerm) paymenttermbus.PaymentTerm {
    return paymenttermbus.PaymentTerm{
        ID:          db.ID,
        Name:        db.Name,
        Description: db.Description.String,
    }
}
```

---

### Phase 3: New paymenttermapp Package

**Directory**: `app/domain/core/paymenttermapp/`

| File | Contents |
|------|----------|
| `model.go` | App models with JSON tags, Encode/Decode, validation, **Entities wrapper** |
| `filter.go` | Parse QueryParams to bus filter |
| `order.go` | Map API order to bus order |
| `paymenttermapp.go` | App struct with business calls including QueryAll |

#### 3a. model.go - Include Entities wrapper for QueryAll

```go
package paymenttermapp

import (
    "encoding/json"
    "github.com/timmaaaz/ichor/app/sdk/errs"
    "github.com/timmaaaz/ichor/business/domain/core/paymenttermbus"
)

type QueryParams struct {
    Page        string
    Rows        string
    OrderBy     string
    ID          string
    Name        string
    Description string
}

type PaymentTerm struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
}

func (app PaymentTerm) Encode() ([]byte, string, error) {
    data, err := json.Marshal(app)
    return data, "application/json", err
}

// Entities is a collection wrapper that implements the Encoder interface.
type Entities []PaymentTerm

func (app Entities) Encode() ([]byte, string, error) {
    data, err := json.Marshal(app)
    return data, "application/json", err
}

type NewPaymentTerm struct {
    Name        string `json:"name" validate:"required"`
    Description string `json:"description"`
}

func (app *NewPaymentTerm) Decode(data []byte) error {
    return json.Unmarshal(data, &app)
}

func (app NewPaymentTerm) Validate() error {
    if err := errs.Check(app); err != nil {
        return errs.Newf(errs.InvalidArgument, "validate: %s", err)
    }
    return nil
}

type UpdatePaymentTerm struct {
    Name        *string `json:"name"`
    Description *string `json:"description"`
}

func (app *UpdatePaymentTerm) Decode(data []byte) error {
    return json.Unmarshal(data, &app)
}

func (app UpdatePaymentTerm) Validate() error {
    if err := errs.Check(app); err != nil {
        return errs.Newf(errs.InvalidArgument, "validate: %s", err)
    }
    return nil
}

// ToAppPaymentTerm converts bus model to app model
func ToAppPaymentTerm(bus paymenttermbus.PaymentTerm) PaymentTerm {
    return PaymentTerm{
        ID:          bus.ID.String(),
        Name:        bus.Name,
        Description: bus.Description,
    }
}

// ToAppPaymentTerms converts slice of bus models to app models
func ToAppPaymentTerms(bus []paymenttermbus.PaymentTerm) []PaymentTerm {
    app := make([]PaymentTerm, len(bus))
    for i, b := range bus {
        app[i] = ToAppPaymentTerm(b)
    }
    return app
}
```

#### 3b. paymenttermapp.go - Include QueryAll

```go
// QueryAll returns all payment terms for dropdown population.
func (a *App) QueryAll(ctx context.Context) (Entities, error) {
    paymentTerms, err := a.paymenttermbus.QueryAll(ctx)
    if err != nil {
        return nil, errs.Newf(errs.Internal, "queryall: %s", err)
    }
    return Entities(ToAppPaymentTerms(paymentTerms)), nil
}
```

---

### Phase 4: New paymenttermapi Package

**Directory**: `api/domain/http/core/paymenttermapi/`

| File | Contents |
|------|----------|
| `paymenttermapi.go` | HTTP handlers: create, update, delete, query, queryByID, **queryAll** |
| `route.go` | Route registration with **auth.RuleAdminOnly for writes** |
| `filter.go` | Parse HTTP query params |

#### 4a. route.go - Auth rules

```go
package paymenttermapi

import (
    "net/http"

    "github.com/timmaaaz/ichor/api/sdk/http/mid"
    "github.com/timmaaaz/ichor/app/domain/core/paymenttermapp"
    "github.com/timmaaaz/ichor/app/sdk/auth"
    "github.com/timmaaaz/ichor/app/sdk/authclient"
    "github.com/timmaaaz/ichor/business/domain/core/paymenttermbus"
    "github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
    "github.com/timmaaaz/ichor/foundation/logger"
    "github.com/timmaaaz/ichor/foundation/web"
)

type Config struct {
    Log             *logger.Logger
    PaymentTermBus  *paymenttermbus.Business
    AuthClient      *authclient.Client
    PermissionsBus  *permissionsbus.Business
}

const RouteTable = "core.payment_terms"

func Routes(app *web.App, cfg Config) {
    const version = "v1"

    api := newAPI(paymenttermapp.NewApp(cfg.PaymentTermBus))
    authen := mid.Authenticate(cfg.AuthClient)

    // READ endpoints - allow any authenticated user
    app.HandlerFunc(http.MethodGet, version, "/core/payment-terms", api.query, authen,
        mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))
    app.HandlerFunc(http.MethodGet, version, "/core/payment-terms/all", api.queryAll, authen,
        mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))
    app.HandlerFunc(http.MethodGet, version, "/core/payment-terms/{payment_term_id}", api.queryByID, authen,
        mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

    // WRITE endpoints - admin only
    app.HandlerFunc(http.MethodPost, version, "/core/payment-terms", api.create, authen,
        mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAdminOnly))
    app.HandlerFunc(http.MethodPut, version, "/core/payment-terms/{payment_term_id}", api.update, authen,
        mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAdminOnly))
    app.HandlerFunc(http.MethodDelete, version, "/core/payment-terms/{payment_term_id}", api.delete, authen,
        mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete, auth.RuleAdminOnly))
}
```

**Routes**:
- `GET /v1/core/payment-terms` - Query with pagination (auth.RuleAny)
- `GET /v1/core/payment-terms/all` - Query all for dropdowns (auth.RuleAny)
- `GET /v1/core/payment-terms/{payment_term_id}` - Get by ID (auth.RuleAny)
- `POST /v1/core/payment-terms` - Create (**auth.RuleAdminOnly**)
- `PUT /v1/core/payment-terms/{payment_term_id}` - Update (**auth.RuleAdminOnly**)
- `DELETE /v1/core/payment-terms/{payment_term_id}` - Delete (**auth.RuleAdminOnly**)

---

### Phase 5: Update Orders Domain (sales.orders)

**Files to modify**:

1. [ordersbus/model.go](business/domain/sales/ordersbus/model.go)
   - `PaymentTerms string` → `PaymentTermsID *uuid.UUID` (nullable FK)

2. [ordersbus/filter.go](business/domain/sales/ordersbus/filter.go)
   - Update filter for PaymentTermsID

3. [ordersbus/stores/ordersdb/model.go](business/domain/sales/ordersbus/stores/ordersdb/model.go)
   - Update dbOrder struct: `PaymentTermsID uuid.NullUUID`

4. [ordersbus/stores/ordersdb/filter.go](business/domain/sales/ordersbus/stores/ordersdb/filter.go)
   - Update filter SQL

5. [ordersapp/model.go](app/domain/sales/ordersapp/model.go)
   - Update app models, JSON tags (`payment_terms_id`), conversion functions

6. [ordersbus/testutil.go](business/domain/sales/ordersbus/testutil.go)
   - Add PaymentTermsIDs parameter to test functions
   - Use hardcoded UUIDs from paymenttermbus for test seeding

---

### Phase 6: Update Supplier Domain (procurement.suppliers)

**Files to modify**:

1. [supplierbus/model.go](business/domain/procurement/supplierbus/model.go)
   - `PaymentTerms string` → `PaymentTermsID *uuid.UUID` (nullable FK)

2. [supplierbus/filter.go](business/domain/procurement/supplierbus/filter.go)
   - Update filter

3. [supplierbus/stores/supplierdb/model.go](business/domain/procurement/supplierbus/stores/supplierdb/model.go)
   - Update dbSupplier struct: `PaymentTermsID uuid.NullUUID`

4. [supplierapp/model.go](app/domain/procurement/supplierapp/model.go)
   - Update app models (`payment_terms_id`)

5. [supplierbus/testutil.go](business/domain/procurement/supplierbus/testutil.go)
   - Add PaymentTermsIDs parameter
   - Use hardcoded UUIDs from paymenttermbus

---

### Phase 7: Wiring in all.go

**File**: [all.go](api/cmd/services/ichor/build/all/all.go)

1. Add imports:
```go
import (
    "github.com/timmaaaz/ichor/api/domain/http/core/paymenttermapi"
    "github.com/timmaaaz/ichor/business/domain/core/paymenttermbus"
    "github.com/timmaaaz/ichor/business/domain/core/paymenttermbus/stores/paymenttermdb"
)
```

2. Instantiate business layer (around line 320):
```go
paymentTermBus := paymenttermbus.NewBusiness(cfg.Log, delegate, paymenttermdb.NewStore(cfg.Log, cfg.DB))
```

3. Register domain for workflow events:
```go
delegateHandler.RegisterDomain(delegate, paymenttermbus.DomainName, paymenttermbus.EntityName)
```

4. Add routes (around line 520):
```go
paymenttermapi.Routes(app, paymenttermapi.Config{
    Log:            cfg.Log,
    PaymentTermBus: paymentTermBus,
    AuthClient:     cfg.AuthClient,
    PermissionsBus: permissionsBus,
})
```

---

### Phase 8: FormData Registry

**File**: [formdata_registry.go](api/cmd/services/ichor/build/all/formdata_registry.go)

1. Add parameter to function signature:
```go
func buildFormDataRegistry(
    // ... existing params
    paymentTermApp *paymenttermapp.App,
    // ... remaining params
) (*formdataregistry.Registry, error) {
```

2. Register entity:
```go
if err := registry.Register(formdataregistry.EntityRegistration{
    Name: "core.payment_terms",
    DecodeNew: func(data json.RawMessage) (interface{}, error) {
        var app paymenttermapp.NewPaymentTerm
        if err := json.Unmarshal(data, &app); err != nil {
            return nil, err
        }
        if err := app.Validate(); err != nil {
            return nil, err
        }
        return app, nil
    },
    CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
        return paymentTermApp.Create(ctx, model.(paymenttermapp.NewPaymentTerm))
    },
    CreateModel: paymenttermapp.NewPaymentTerm{},
    DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
        var app paymenttermapp.UpdatePaymentTerm
        if err := json.Unmarshal(data, &app); err != nil {
            return nil, err
        }
        if err := app.Validate(); err != nil {
            return nil, err
        }
        return app, nil
    },
    UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
        return paymentTermApp.Update(ctx, model.(paymenttermapp.UpdatePaymentTerm), id)
    },
    UpdateModel: paymenttermapp.UpdatePaymentTerm{},
}); err != nil {
    return nil, fmt.Errorf("register core.payment_terms: %w", err)
}
```

3. Update call site in `all.go`:
```go
formDataRegistry, err := buildFormDataRegistry(
    // ... existing params
    paymenttermapp.NewApp(paymentTermBus),
    // ... remaining params
)
```

---

### Phase 9: Form Configuration Updates

#### 9a. Update tableforms.go

**File**: [tableforms.go](business/sdk/dbtest/seedmodels/tableforms.go)

**Sales Orders** - Change from:
```go
{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "orders", Name: "payment_terms", Label: "Payment Terms", FieldType: "text", FieldOrder: 9, Required: false, Config: json.RawMessage(`{}`)}
```
To:
```go
{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "orders", Name: "payment_terms_id", Label: "Payment Terms", FieldType: "smart-combobox", FieldOrder: 9, Required: false, Config: json.RawMessage(`{"entity": "core.payment_terms", "display_field": "name"}`)}
```

**Suppliers** - Change from:
```go
{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "suppliers", Name: "payment_terms", Label: "Payment Terms", FieldType: "textarea", FieldOrder: 3, Required: true, Config: json.RawMessage(`{}`)}
```
To:
```go
{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "suppliers", Name: "payment_terms_id", Label: "Payment Terms", FieldType: "smart-combobox", FieldOrder: 3, Required: false, Config: json.RawMessage(`{"entity": "core.payment_terms", "display_field": "name"}`)}
```

#### 9b. Update forms.go

**File**: [forms.go](business/sdk/dbtest/seedmodels/forms.go)

**GetFullSupplierFormFields** - Change the payment_terms field to:
```go
{
    FormID:       formID,
    EntityID:     supplierEntityID,
    EntitySchema: "procurement",
    EntityTable:  "suppliers",
    Name:         "payment_terms_id",
    Label:        "Payment Terms",
    FieldType:    "smart-combobox",
    FieldOrder:   order,
    Required:     false,
    Config:       json.RawMessage(`{"execution_order": 2, "entity": "core.payment_terms", "display_field": "name"}`),
},
```

---

### Phase 10: Test Updates

**New test directory**: `api/cmd/services/ichor/tests/core/paymenttermapi/`

#### 10a. Test file structure

| File | Contents |
|------|----------|
| `paymentterm_test.go` | Main test orchestration |
| `seed_test.go` | Seed data with Users AND Admins |
| `query_test.go` | query200 tests (uses any authenticated user) |
| `create_test.go` | create200 (admin), create401 (non-admin) |
| `update_test.go` | update200 (admin), update401 (non-admin) |
| `delete_test.go` | delete200 (admin), delete401 (non-admin) |

#### 10b. seed_test.go - Admin authorization pattern

```go
func insertSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
    ctx := context.Background()
    busDomain := db.BusDomain

    // Regular user (for 401 tests)
    usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
    if err != nil {
        return apitest.SeedData{}, fmt.Errorf("seeding user: %w", err)
    }
    tu1 := apitest.User{
        User:  usrs[0],
        Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
    }

    // Admin user (for 200 tests on write operations)
    usrs, err = userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
    if err != nil {
        return apitest.SeedData{}, fmt.Errorf("seeding admin: %w", err)
    }
    tu2 := apitest.User{
        User:  usrs[0],
        Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
    }

    // Seed payment terms for query/update/delete tests
    paymentTerms, err := paymenttermbus.TestSeedPaymentTerms(ctx, 2, busDomain.PaymentTerm)
    if err != nil {
        return apitest.SeedData{}, fmt.Errorf("seeding payment terms: %w", err)
    }

    return apitest.SeedData{
        Users:        []apitest.User{tu1},    // Regular users
        Admins:       []apitest.User{tu2},    // Admin users
        PaymentTerms: paymenttermapp.ToAppPaymentTerms(paymentTerms),
    }, nil
}
```

#### 10c. create_test.go - Admin-only authorization tests

```go
// create200 - successful creation with admin token
func create200(sd apitest.SeedData) []apitest.Table {
    return []apitest.Table{
        {
            Name:       "basic",
            URL:        "/v1/core/payment-terms",
            Token:      sd.Admins[0].Token,  // Admin token - should succeed
            Method:     http.MethodPost,
            StatusCode: http.StatusOK,
            Input: &paymenttermapp.NewPaymentTerm{
                Name:        "Test Payment Term",
                Description: "Test Description",
            },
            GotResp: &paymenttermapp.PaymentTerm{},
            ExpResp: &paymenttermapp.PaymentTerm{
                Name:        "Test Payment Term",
                Description: "Test Description",
            },
            CmpFunc: func(got, exp any) string {
                gotResp := got.(*paymenttermapp.PaymentTerm)
                expResp := exp.(*paymenttermapp.PaymentTerm)
                expResp.ID = gotResp.ID
                return cmp.Diff(gotResp, expResp)
            },
        },
    }
}

// create401 - unauthorized creation with non-admin token
func create401(sd apitest.SeedData) []apitest.Table {
    return []apitest.Table{
        {
            Name:       "non-admin",
            URL:        "/v1/core/payment-terms",
            Token:      sd.Users[0].Token,  // Regular user token - should fail
            Method:     http.MethodPost,
            StatusCode: http.StatusUnauthorized,
            Input: &paymenttermapp.NewPaymentTerm{
                Name:        "Test Payment Term",
                Description: "Test Description",
            },
            GotResp: &errs.Error{},
            ExpResp: errs.Newf(errs.Unauthenticated,
                "authorize: you are not authorized for that action, claims[[USER]] rule[rule_admin_only]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
            CmpFunc: func(got, exp any) string {
                return cmp.Diff(got, exp)
            },
        },
    }
}
```

#### 10d. paymentterm_test.go - Test orchestration

```go
func Test_PaymentTerm(t *testing.T) {
    t.Parallel()

    test := apitest.StartTest(t, "Test_PaymentTerm")
    sd, err := insertSeedData(test.DB, test.Auth)
    if err != nil {
        t.Fatalf("seeding error %s", err)
    }

    // Read operations - any authenticated user
    test.Run(t, query200(sd), "query-200")
    test.Run(t, queryByID200(sd), "queryByID-200")

    // Write operations - admin only (test both success and failure)
    test.Run(t, create200(sd), "create-200")    // Admin succeeds
    test.Run(t, create401(sd), "create-401")    // Non-admin fails

    test.Run(t, update200(sd), "update-200")    // Admin succeeds
    test.Run(t, update401(sd), "update-401")    // Non-admin fails

    test.Run(t, delete200(sd), "delete-200")    // Admin succeeds
    test.Run(t, delete401(sd), "delete-401")    // Non-admin fails
}
```

#### 10e. Update apitest.SeedData model

**File**: `api/sdk/http/apitest/model.go`

Add PaymentTerms field if not present:
```go
type SeedData struct {
    Users        []User
    Admins       []User
    PaymentTerms []paymenttermapp.PaymentTerm
    // ... other fields
}
```

#### 10f. Update tableaccessbus/testutil.go

**File**: [tableaccessbus/testutil.go](business/domain/core/tableaccessbus/testutil.go)

Add permission entry in the Core schema section:
```go
// Core schema
{RoleID: uuid.Nil, TableName: "core.users", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
{RoleID: uuid.Nil, TableName: "core.contact_infos", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
{RoleID: uuid.Nil, TableName: "core.roles", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
{RoleID: uuid.Nil, TableName: "core.user_roles", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
{RoleID: uuid.Nil, TableName: "core.pages", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
{RoleID: uuid.Nil, TableName: "core.payment_terms", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},  // ADD THIS
{RoleID: uuid.Nil, TableName: "core.role_pages", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
```

**Existing tests to update**:
- [ordersapi tests](api/cmd/services/ichor/tests/sales/ordersapi/) - Update to seed PaymentTerms before Orders, use hardcoded UUIDs
- [supplierapi tests](api/cmd/services/ichor/tests/procurement/supplierapi/) - Update to seed PaymentTerms before Suppliers, use hardcoded UUIDs

---

## Verification Plan

1. **Build check**: `go build ./api/cmd/services/ichor/...`
2. **Run migrations**: `make migrate`
3. **Run all tests**: `make test`
4. **Manual API tests**:
   - `GET /v1/core/payment-terms` - Verify dropdown data available (any auth user)
   - `GET /v1/core/payment-terms/all` - Verify all terms returned (any auth user)
   - `POST /v1/core/payment-terms` with non-admin - Verify 401 Unauthorized
   - `POST /v1/core/payment-terms` with admin - Verify 200 OK
   - `POST /v1/sales/orders` with `payment_terms_id` - Verify FK works
   - `POST /v1/procurement/suppliers` with `payment_terms_id` - Verify FK works
5. **Form test**: Verify supplier/order forms show dropdown instead of textarea

---

## Critical Files

1. [assetconditionbus.go](business/domain/assets/assetconditionbus/assetconditionbus.go) - Pattern to follow
2. [roleapi/route.go](api/domain/http/core/roleapi/route.go) - Auth rule pattern (RuleAdminOnly for writes)
3. [ordersbus/model.go](business/domain/sales/ordersbus/model.go) - FK change needed
4. [supplierbus/model.go](business/domain/procurement/supplierbus/model.go) - FK change needed
5. [all.go](api/cmd/services/ichor/build/all/all.go) - Central wiring
6. [migrate.sql](business/sdk/migrate/sql/migrate.sql) - Schema changes

---

## Summary of Changes from Original Plan

| Issue | Original | Fixed |
|-------|----------|-------|
| Package naming | `paymenttermsbus` (plural) | `paymenttermbus` (singular) |
| DomainName | Not specified | `"paymentterm"` (singular, no underscore) |
| Migration version | `1.XX` placeholder | `1.42` (before suppliers, requires renumbering) |
| DB struct naming | Not specified | `paymentTerm` (lowercase, unexported) |
| Auth rules | `auth.RuleAny` for all | `auth.RuleAdminOnly` for Create/Update/Delete |
| Seed UUIDs | `gen_random_uuid()` | Hardcoded UUIDs for test predictability |
| QueryAll method | Missing | Added for dropdown population |
| Test patterns | Basic structure | Full admin/non-admin authorization tests |
| Entities wrapper | Missing | Added for QueryAll response |
