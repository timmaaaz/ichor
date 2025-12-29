# Status Domains Color & Icon Enhancement Plan

## Overview

Add `primary_color`, `secondary_color`, and `icon` fields to all status domains across the codebase. The database migrations have already been applied. This plan implements the changes across all layers (store, business, app, API) with comprehensive testing for each domain.

## Database Schema (Already Applied)

All five status tables now have these additional columns:
```sql
primary_color VARCHAR(50) NULL,
secondary_color VARCHAR(50) NULL,
icon VARCHAR(100) NULL,
```

**Affected Tables:**
1. `hr.user_approval_status` (lines 129-137)
2. `assets.approval_status` (lines 239-247)
3. `assets.fulfillment_status` (lines 251-259)
4. `sales.order_fulfillment_statuses` (lines 764-773)
5. `sales.line_item_fulfillment_statuses` (lines 775-784)

## Architecture Pattern

Each domain follows the Ardan Labs layered architecture:
- **Store Layer** (`*db`): Database models and SQL operations
- **Business Layer** (`*bus`): Domain models and business logic
- **App Layer** (`*app`): API models, validation, and conversion
- **API Layer** (`*api`): HTTP handlers and routing

## Implementation Phases

Each phase represents one domain and can be implemented independently.

---

## Phase 1: HR User Approval Status

### Domain: `hr.user_approval_status`

**Current State:**
- Has `IconID` field (UUID reference to icons table)
- Need to add direct color/icon string fields

### Files to Modify:

#### 1.1 Business Layer Models
**File:** `business/domain/hr/approvalbus/model.go`

Add fields to all three structs:
```go
type UserApprovalStatus struct {
    ID             uuid.UUID
    Name           string
    IconID         uuid.UUID
    PrimaryColor   string    // NEW
    SecondaryColor string    // NEW
    Icon           string    // NEW
}

type NewUserApprovalStatus struct {
    Name           string
    IconID         uuid.UUID
    PrimaryColor   string    // NEW
    SecondaryColor string    // NEW
    Icon           string    // NEW
}

type UpdateUserApprovalStatus struct {
    Name           *string
    IconID         *uuid.UUID
    PrimaryColor   *string   // NEW
    SecondaryColor *string   // NEW
    Icon           *string   // NEW
}
```

#### 1.2 Store Layer Models
**File:** `business/domain/hr/approvalbus/stores/approvaldb/model.go`

Add database fields and update conversion functions:
```go
type userApprovalStatus struct {
    ID             uuid.UUID      `db:"id"`
    Name           string         `db:"name"`
    IconID         sql.NullString `db:"icon_id"`
    PrimaryColor   sql.NullString `db:"primary_color"`    // NEW
    SecondaryColor sql.NullString `db:"secondary_color"`  // NEW
    Icon           sql.NullString `db:"icon"`             // NEW
}
```

Update `toDBUserApprovalStatus()` and `toBusUserApprovalStatus()` conversion functions.

#### 1.3 Store Layer SQL
**File:** `business/domain/hr/approvalbus/stores/approvaldb/statusdb.go`

Update SQL queries:
- `Create()`: Add new columns to INSERT statement
- `Update()`: Add new columns to UPDATE statement
- `Query()`: Add new columns to SELECT statement
- `QueryByID()`: Add new columns to SELECT statement

#### 1.4 App Layer Models
**File:** `app/domain/hr/approvalapp/model.go`

Add fields to API models and update conversions:
```go
type UserApprovalStatus struct {
    ID             string `json:"id"`
    IconID         string `json:"icon_id"`
    Name           string `json:"name"`
    PrimaryColor   string `json:"primary_color"`    // NEW
    SecondaryColor string `json:"secondary_color"`  // NEW
    Icon           string `json:"icon"`             // NEW
}

type NewUserApprovalStatus struct {
    IconID         string `json:"icon_id" validate:"omitempty,uuid"`
    Name           string `json:"name" validate:"required,min=3,max=100"`
    PrimaryColor   string `json:"primary_color" validate:"omitempty,max=50"`    // NEW
    SecondaryColor string `json:"secondary_color" validate:"omitempty,max=50"`  // NEW
    Icon           string `json:"icon" validate:"omitempty,max=100"`            // NEW
}

type UpdateUserApprovalStatus struct {
    IconID         *string `json:"icon_id" validate:"omitempty,uuid"`
    Name           *string `json:"name" validate:"omitempty,min=3,max=100"`
    PrimaryColor   *string `json:"primary_color" validate:"omitempty,max=50"`    // NEW
    SecondaryColor *string `json:"secondary_color" validate:"omitempty,max=50"`  // NEW
    Icon           *string `json:"icon" validate:"omitempty,max=100"`            // NEW
}
```

Update conversion functions: `ToAppUserApprovalStatus()`, `toBusNewUserApprovalStatus()`, `toBusUpdateUserApprovalStatus()`

#### 1.5 App Layer Filter (Optional)
**File:** `app/domain/hr/approvalapp/filter.go`

Consider adding filter support for new fields if needed:
```go
type QueryParams struct {
    Page           string
    Rows           string
    OrderBy        string
    ID             string
    IconID         string
    Name           string
    PrimaryColor   string  // NEW (if filtering needed)
    SecondaryColor string  // NEW (if filtering needed)
    Icon           string  // NEW (if filtering needed)
}
```

#### 1.6 Tests - Create
**File:** `api/cmd/services/ichor/tests/hr/userapprovalstatusapi/create_test.go`

Add test cases that include new fields:
```go
{
    Name:       "with colors and icon",
    URL:        "/v1/hr/user-approval-status",
    Token:      sd.Users[0].Token,
    Method:     http.MethodPost,
    StatusCode: http.StatusOK,
    Input: &approvalapp.NewUserApprovalStatus{
        Name:           "ColoredStatus",
        PrimaryColor:   "#FF5733",
        SecondaryColor: "#33FF57",
        Icon:           "check-circle",
    },
    GotResp: &approvalapp.UserApprovalStatus{},
    ExpResp: &approvalapp.UserApprovalStatus{
        Name:           "ColoredStatus",
        IconID:         "00000000-0000-0000-0000-000000000000",
        PrimaryColor:   "#FF5733",
        SecondaryColor: "#33FF57",
        Icon:           "check-circle",
    },
    // ... CmpFunc
}
```

#### 1.7 Tests - Update
**File:** `api/cmd/services/ichor/tests/hr/userapprovalstatusapi/update_test.go`

Add test cases for updating color/icon fields:
```go
{
    Name:       "update colors",
    URL:        "/v1/hr/user-approval-status/" + sd.Approvals[0].ID,
    Token:      sd.Users[0].Token,
    Method:     http.MethodPut,
    StatusCode: http.StatusOK,
    Input: &approvalapp.UpdateUserApprovalStatus{
        PrimaryColor:   convert.StringPointer("#AABBCC"),
        SecondaryColor: convert.StringPointer("#DDEEFF"),
        Icon:           convert.StringPointer("star"),
    },
    // ... GotResp, ExpResp, CmpFunc
}
```

#### 1.8 Tests - Query
**File:** `api/cmd/services/ichor/tests/hr/userapprovalstatusapi/query_test.go`

Verify new fields are returned in query results.

#### 1.9 Tests - Seed
**File:** `api/cmd/services/ichor/tests/hr/userapprovalstatusapi/seed_test.go`

Add color/icon data to seed records if needed for tests.

---

## Phase 2: Assets Approval Status

### Domain: `assets.approval_status`

**Current State:**
- Has `IconID` field (UUID reference)
- Need to add color/icon string fields

### Files to Modify:

#### 2.1 Business Layer Models
**File:** `business/domain/assets/approvalstatusbus/model.go`

Add `PrimaryColor`, `SecondaryColor`, `Icon` to:
- `ApprovalStatus`
- `NewApprovalStatus`
- `UpdateApprovalStatus` (as pointers)

#### 2.2 Store Layer Models
**File:** `business/domain/assets/approvalstatusbus/stores/approvalstatusdb/model.go`

Add database fields (using `sql.NullString`) and update conversions:
- `toDBApprovalStatus()`
- `toBusApprovalStatus()`
- `toBusApprovalStatuses()`

#### 2.3 Store Layer SQL
**File:** `business/domain/assets/approvalstatusbus/stores/approvalstatusdb/approvalstatusdb.go`

Update SQL statements in:
- `Create()`
- `Update()`
- `Query()`
- `QueryByID()`

#### 2.4 App Layer Models
**File:** `app/domain/assets/approvalstatusapp/model.go`

Add fields to:
- `ApprovalStatus` (with JSON tags: `primary_color`, `secondary_color`, `icon`)
- `NewApprovalStatus` (with validation: `validate:"omitempty,max=50"` for colors, `max=100` for icon)
- `UpdateApprovalStatus` (as pointers with same validation)

Update conversions:
- `ToAppApprovalStatus()`
- `ToAppApprovalStatuses()`
- `toBusNewApprovalStatus()`
- `toBusUpdateApprovalStatus()`

#### 2.5 App Layer Filter
**File:** `app/domain/assets/approvalstatusapp/filter.go`

Optionally add filter parameters for new fields.

#### 2.6 Tests - Create
**File:** `api/cmd/services/ichor/tests/assets/approvalstatusapi/create_test.go`

Add test cases with color/icon fields.

#### 2.7 Tests - Update
**File:** `api/cmd/services/ichor/tests/assets/approvalstatusapi/update_test.go`

Add test cases for updating color/icon fields.

#### 2.8 Tests - Query
**File:** `api/cmd/services/ichor/tests/assets/approvalstatusapi/query_test.go`

Verify fields in results.

#### 2.9 Tests - Seed
**File:** `api/cmd/services/ichor/tests/assets/approvalstatusapi/seed_test.go`

Update seed data if needed.

---

## Phase 3: Assets Fulfillment Status

### Domain: `assets.fulfillment_status`

**Current State:**
- Has `IconID` field (UUID reference)
- Need to add color/icon string fields

### Files to Modify:

#### 3.1 Business Layer Models
**File:** `business/domain/assets/fulfillmentstatusbus/model.go`

Add `PrimaryColor`, `SecondaryColor`, `Icon` to:
- `FulfillmentStatus`
- `NewFulfillmentStatus`
- `UpdateFulfillmentStatus` (as pointers)

#### 3.2 Store Layer Models
**File:** `business/domain/assets/fulfillmentstatusbus/stores/fulfillmentstatusdb/model.go`

Add database fields (using `sql.NullString`) and update conversions.

#### 3.3 Store Layer SQL
**File:** `business/domain/assets/fulfillmentstatusbus/stores/fulfillmentstatusdb/fulfillmentstatusdb.go`

Update SQL statements in all CRUD operations.

#### 3.4 App Layer Models
**File:** `app/domain/assets/fulfillmentstatusapp/model.go`

Add fields with JSON tags and validation to all model structs.
Update conversion functions.

#### 3.5 App Layer Filter
**File:** `app/domain/assets/fulfillmentstatusapp/filter.go`

Optionally add filter parameters.

#### 3.6 Tests - Create
**File:** `api/cmd/services/ichor/tests/assets/fulfillmentstatusapi/create_test.go`

Add test cases with color/icon fields.

#### 3.7 Tests - Update
**File:** `api/cmd/services/ichor/tests/assets/fulfillmentstatusapi/update_test.go`

Add test cases for updates.

#### 3.8 Tests - Query
**File:** `api/cmd/services/ichor/tests/assets/fulfillmentstatusapi/query_test.go`

Verify fields in results.

#### 3.9 Tests - Seed
**File:** `api/cmd/services/ichor/tests/assets/fulfillmentstatusapi/seed_test.go`

Update seed data if needed.

---

## Phase 4: Sales Order Fulfillment Status

### Domain: `sales.order_fulfillment_statuses`

**Current State:**
- Has `Description` field (no IconID)
- Need to add color/icon string fields

### Files to Modify:

#### 4.1 Business Layer Models
**File:** `business/domain/sales/orderfulfillmentstatusbus/model.go`

Add `PrimaryColor`, `SecondaryColor`, `Icon` to:
- `OrderFulfillmentStatus`
- `NewOrderFulfillmentStatus`
- `UpdateOrderFulfillmentStatus` (as pointers)

#### 4.2 Store Layer Models
**File:** `business/domain/sales/orderfulfillmentstatusbus/stores/orderfulfillmentstatusdb/model.go`

Add database fields (using `sql.NullString`) and update conversions:
- `toDBOrderFulfillmentStatus()`
- `toBusOrderFulfillmentStatus()`
- `toBusOrderFulfillmentStatuses()`

#### 4.3 Store Layer SQL
**File:** `business/domain/sales/orderfulfillmentstatusbus/stores/orderfulfillmentstatusdb/orderfulfillmentstatusdb.go`

Update SQL statements in:
- `Create()`
- `Update()`
- `Query()`
- `QueryByID()`

#### 4.4 App Layer Models
**File:** `app/domain/sales/orderfulfillmentstatusapp/model.go`

Add fields to:
- `OrderFulfillmentStatus` (with JSON tags)
- `NewOrderFulfillmentStatus` (with validation)
- `UpdateOrderFulfillmentStatus` (as pointers with validation)

Update conversions:
- `ToAppOrderFulfillmentStatus()`
- `ToAppOrderFulfillmentStatuses()`
- `toBusNewOrderFulfillmentStatus()`
- `toBusUpdateOrderFulfillmentStatus()`

#### 4.5 App Layer Filter
**File:** `app/domain/sales/orderfulfillmentstatusapp/filter.go`

Optionally add filter parameters.

#### 4.6 Tests - Create
**File:** `api/cmd/services/ichor/tests/sales/orderfulfillmentstatusapi/create_test.go`

Add test cases with color/icon fields.

#### 4.7 Tests - Update
**File:** `api/cmd/services/ichor/tests/sales/orderfulfillmentstatusapi/update_test.go`

Add test cases for updates.

#### 4.8 Tests - Query
**File:** `api/cmd/services/ichor/tests/sales/orderfulfillmentstatusapi/query_test.go`

Verify fields in results.

#### 4.9 Tests - Seed
**File:** `api/cmd/services/ichor/tests/sales/orderfulfillmentstatusapi/seed_test.go`

Update seed data if needed.

---

## Phase 5: Sales Line Item Fulfillment Status

### Domain: `sales.line_item_fulfillment_statuses`

**Current State:**
- Has `Description` field (no IconID)
- Need to add color/icon string fields

### Files to Modify:

#### 5.1 Business Layer Models
**File:** `business/domain/sales/lineitemfulfillmentstatusbus/model.go`

Add `PrimaryColor`, `SecondaryColor`, `Icon` to:
- `LineItemFulfillmentStatus`
- `NewLineItemFulfillmentStatus`
- `UpdateLineItemFulfillmentStatus` (as pointers)

#### 5.2 Store Layer Models
**File:** `business/domain/sales/lineitemfulfillmentstatusbus/stores/lineitemfulfillmentstatusdb/model.go`

Add database fields (using `sql.NullString`) and update conversions:
- `toDBLineItemFulfillmentStatus()`
- `toBusLineItemFulfillmentStatus()`
- `toBusLineItemFulfillmentStatuses()`

#### 5.3 Store Layer SQL
**File:** `business/domain/sales/lineitemfulfillmentstatusbus/stores/lineitemfulfillmentstatusdb/lineitemfulfillmentstatusdb.go`

Update SQL statements in:
- `Create()`
- `Update()`
- `Query()`
- `QueryByID()`

#### 5.4 App Layer Models
**File:** `app/domain/sales/lineitemfulfillmentstatusapp/model.go`

Add fields to:
- `LineItemFulfillmentStatus` (with JSON tags)
- `NewLineItemFulfillmentStatus` (with validation)
- `UpdateLineItemFulfillmentStatus` (as pointers with validation)

Update conversions:
- `ToAppLineItemFulfillmentStatus()`
- `ToAppLineItemFulfillmentStatuses()`
- `toBusNewLineItemFulfillmentStatus()`
- `toBusUpdateLineItemFulfillmentStatus()`

#### 5.5 App Layer Filter
**File:** `app/domain/sales/lineitemfulfillmentstatusapp/filter.go`

Optionally add filter parameters.

#### 5.6 Tests - Create
**File:** `api/cmd/services/ichor/tests/sales/lineitemfulfillmentstatusapi/create_test.go`

Add test cases with color/icon fields.

#### 5.7 Tests - Update
**File:** `api/cmd/services/ichor/tests/sales/lineitemfulfillmentstatusapi/update_test.go`

Add test cases for updates.

#### 5.8 Tests - Query
**File:** `api/cmd/services/ichor/tests/sales/lineitemfulfillmentstatusapi/query_test.go`

Verify fields in results.

#### 5.9 Tests - Seed
**File:** `api/cmd/services/ichor/tests/sales/lineitemfulfillmentstatusapi/seed_test.go`

Update seed data if needed.

---

## Common Patterns & Guidelines

### Field Handling
- **Nullability**: All three new fields are nullable in the database (`NULL` allowed)
- **Database Type**: Use `sql.NullString` in store layer structs
- **Business Type**: Use plain `string` (empty string for null values)
- **Update Type**: Use `*string` pointers to distinguish "not provided" from "set to empty"

### Validation Rules
- `primary_color`: `validate:"omitempty,max=50"`
- `secondary_color`: `validate:"omitempty,max=50"`
- `icon`: `validate:"omitempty,max=100"`

### JSON Field Naming
Use snake_case for consistency with existing API:
- `primary_color`
- `secondary_color`
- `icon`

### SQL Query Updates
When updating SQL queries, add the new columns in this order (after existing fields):
1. `primary_color`
2. `secondary_color`
3. `icon`

### Conversion Utilities
Use existing nulltype utilities from `business/sdk/sqldb/nulltypes`:
- `ToNullableString(value string) sql.NullString`
- `FromNullableString(ns sql.NullString) string`

### Test Data Examples
```go
PrimaryColor:   "#FF5733"      // Hex color
SecondaryColor: "#33FF57"      // Hex color
Icon:           "check-circle"  // Icon name/identifier
```

---

## Execution Strategy

### Per-Phase Workflow
1. **Modify Store Layer** (database models, SQL queries)
2. **Modify Business Layer** (domain models)
3. **Modify App Layer** (API models, validation, conversions)
4. **Update Tests** (create, update, query, seed)
5. **Run Tests** for the domain
6. **Verify** manually if needed

### Testing Commands
```bash
# Test specific domain
go test -v ./api/cmd/services/ichor/tests/hr/userapprovalstatusapi
go test -v ./api/cmd/services/ichor/tests/assets/approvalstatusapi
go test -v ./api/cmd/services/ichor/tests/assets/fulfillmentstatusapi
go test -v ./api/cmd/services/ichor/tests/sales/orderfulfillmentstatusapi
go test -v ./api/cmd/services/ichor/tests/sales/lineitemfulfillmentstatusapi

# Run all tests
make test
```

### Migration Note
Database migrations have already been applied. No migration step required during implementation.

---

## Success Criteria

For each phase:
- [ ] All model structs include the three new fields
- [ ] SQL queries properly select and insert/update new columns
- [ ] Conversion functions handle all new fields correctly
- [ ] Validation rules are applied to API models
- [ ] Integration tests pass with new fields
- [ ] Create operations accept and store new fields
- [ ] Update operations can modify new fields
- [ ] Query operations return new fields in responses
- [ ] Null values are handled correctly throughout the stack

---

## Risks & Considerations

1. **Breaking Changes**: Adding fields to response models is generally safe (additive change)
2. **Null Handling**: Ensure empty strings are used consistently for null database values
3. **Test Data**: Existing tests may need updates if they use strict comparison
4. **API Contracts**: Frontend may need updates to consume new fields
5. **Caching**: If any of these domains use caching, ensure cache keys/TTLs are appropriate

---

## Estimated Effort

Per Phase:
- Store Layer: 15 minutes
- Business Layer: 5 minutes
- App Layer: 20 minutes
- Tests: 20 minutes
- **Total per phase: ~60 minutes**

**Total for all 5 phases: ~5 hours**
