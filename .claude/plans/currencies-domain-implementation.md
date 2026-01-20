# Currencies Domain Implementation Plan

## Overview

This plan covers the implementation of a new `currencies` domain in the `core` schema, along with updates to related domains that currently use string-based currency codes. The implementation follows the Ardan Labs Service architecture patterns used throughout the Ichor codebase.

**Key Decisions (from user input):**
- Caching: Yes, with 60+ minute TTL
- Audit fields: Allow NULL for system/seed operations
- Permissions: Read for any authenticated user, Write operations admin-only
- Scope: Update all currency-related tables (orders, customers, product costs, purchase orders, cost history)

---

## Phase 1: Core Currency Domain Implementation

**Goal:** Create the complete `currencybus` domain following established patterns.

### 1.1 Database Migration

**File:** `business/sdk/migrate/sql/migrate.sql`

Add new migration (find current max version and increment):

```sql
-- Version: X.XX
-- Description: Create currencies reference table
CREATE TABLE core.currencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(3) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    locale VARCHAR(10) NOT NULL,
    decimal_places INT NOT NULL DEFAULT 2,
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort_order INT NOT NULL DEFAULT 0,
    created_by UUID REFERENCES core.users(id),
    created_date TIMESTAMPTZ DEFAULT NOW(),
    updated_by UUID REFERENCES core.users(id),
    updated_date TIMESTAMPTZ DEFAULT NOW()
);
```

### 1.2 Business Layer

**Directory:** `business/domain/core/currencybus/`

Files to create:
- `model.go` - Currency, NewCurrency, UpdateCurrency structs
- `filter.go` - QueryFilter with ID, Code, Name, IsActive filters
- `order.go` - OrderBy constants (ID, Code, Name, SortOrder, IsActive)
- `event.go` - Domain events for workflow integration
- `currencybus.go` - Business logic with Storer interface
- `testutil.go` - Test helper functions

**Directory:** `business/domain/core/currencybus/stores/currencydb/`

Files to create:
- `model.go` - DB struct and conversion functions
- `filter.go` - SQL WHERE clause building
- `order.go` - SQL ORDER BY mapping
- `currencydb.go` - Storer implementation

**Directory:** `business/domain/core/currencybus/stores/currencycache/`

Files to create:
- `currencycache.go` - Cache wrapper using sturdyc (60 min TTL)

### 1.3 Application Layer

**Directory:** `app/domain/core/currencyapp/`

Files to create:
- `model.go` - App models with JSON tags, Encode/Decode, Validate methods
- `filter.go` - Query parameter parsing
- `order.go` - Order field mapping
- `currencyapp.go` - App logic calling business layer

### 1.4 API Layer

**Directory:** `api/domain/http/core/currencyapi/`

Files to create:
- `currencyapi.go` - HTTP handlers (create, update, delete, query, queryByID, queryAll)
- `filter.go` - HTTP query parameter extraction
- `route.go` - Route registration with middleware

**Route Table:** `core.currencies`

**Endpoints:**
- `GET /v1/core/currencies` - Query (any authenticated)
- `GET /v1/core/currencies/:id` - QueryByID (any authenticated)
- `GET /v1/core/currencies/all` - QueryAll (any authenticated)
- `POST /v1/core/currencies` - Create (admin only)
- `PUT /v1/core/currencies/:id` - Update (admin only)
- `DELETE /v1/core/currencies/:id` - Delete (admin only)

### 1.5 Wiring

**File:** `api/cmd/services/ichor/build/all/all.go`

- Import new packages
- Instantiate currencyBus with cache wrapper
- Register routes with Config

### 1.6 Seed Data

**File:** `business/sdk/migrate/sql/seed.sql`

Insert initial currencies:
```sql
INSERT INTO core.currencies (id, code, name, symbol, locale, decimal_places, is_active, sort_order) VALUES
('uuid1', 'USD', 'US Dollar', '$', 'en-US', 2, true, 1),
('uuid2', 'EUR', 'Euro', '€', 'en-EU', 2, true, 2),
('uuid3', 'GBP', 'British Pound', '£', 'en-GB', 2, true, 3),
('uuid4', 'CAD', 'Canadian Dollar', '$', 'en-CA', 2, true, 4),
('uuid5', 'AUD', 'Australian Dollar', '$', 'en-AU', 2, true, 5),
('uuid6', 'JPY', 'Japanese Yen', '¥', 'ja-JP', 0, true, 6),
('uuid7', 'CHF', 'Swiss Franc', 'CHF', 'de-CH', 2, true, 7),
('uuid8', 'CNY', 'Chinese Yuan', '¥', 'zh-CN', 2, true, 8),
('uuid9', 'INR', 'Indian Rupee', '₹', 'en-IN', 2, true, 9),
('uuid10', 'MXN', 'Mexican Peso', '$', 'es-MX', 2, true, 10);
```

### 1.7 Tests

**Business Layer Tests:**
- `business/domain/core/currencybus/currencybus_test.go` - Unit tests

**Integration Tests:**
**Directory:** `api/cmd/services/ichor/tests/core/currencyapi/`

Files to create:
- `currency_test.go` - Main test orchestrator
- `seed_test.go` - Test data seeding
- `query_test.go` - Query endpoint tests (200, 401)
- `create_test.go` - Create endpoint tests (200, 400, 401, 409)
- `update_test.go` - Update endpoint tests (200, 400, 401)
- `delete_test.go` - Delete endpoint tests (200, 401)

**Test Currency:** Use a non-seeded currency code for testing (e.g., "TST", "XYZ", or "ZZZ") to ensure tests create/modify real data rather than relying on seeds.

### 1.8 Table Access

**File:** `business/domain/core/tableaccessbus/testutil.go`

Add entry:
```go
{RoleID: uuid.Nil, TableName: "core.currencies", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
```

---

## Phase 2: Update Test Utilities and Related Domains

**Goal:** Replace hardcoded currency strings with proper FK references in testutil files.

### 2.1 Updates Required

**Files to modify:**

1. `business/domain/products/productcostbus/testutil.go`
   - Remove `var currencyCodes = []string{...}`
   - Add `currencyBus *currencybus.Business` parameter to test functions
   - Query currencies from database
   - Use currency IDs instead of codes

2. `business/domain/sales/ordersbus/testutil.go`
   - Remove `var testCurrencies = []string{...}`
   - Add currency business layer dependency
   - Use currency IDs in test data generation

3. `business/domain/procurement/purchaseorderbus/testutil.go`
   - Remove hardcoded `"USD"`
   - Add currency business layer dependency
   - Use currency IDs

4. `business/domain/products/costhistorybus/testutil.go`
   - Remove generated `Currency%d` strings
   - Add currency business layer dependency
   - Use currency IDs

### 2.2 Model Updates

Each domain that has a `Currency string` field needs to be updated:

**productcostbus:**
- `model.go`: Change `Currency string` to `CurrencyID uuid.UUID`
- `stores/productcostdb/model.go`: Update DB mapping
- `filter.go`: Add CurrencyID filter

**ordersbus:**
- `model.go`: Change `Currency string` to `CurrencyID uuid.UUID`
- `stores/ordersdb/model.go`: Update DB mapping
- `filter.go`: Add CurrencyID filter

**purchaseorderbus:**
- `model.go`: Change `Currency string` to `CurrencyID uuid.UUID`
- `stores/purchaseorderdb/model.go`: Update DB mapping
- `filter.go`: Add CurrencyID filter

**costhistorybus:**
- `model.go`: Change `Currency string` to `CurrencyID uuid.UUID`
- `stores/costhistorydb/model.go`: Update DB mapping
- `filter.go`: Add CurrencyID filter

### 2.3 App Layer Updates

Update corresponding app layer models:
- `productcostapp/model.go`
- `ordersapp/model.go`
- `purchaseorderapp/model.go`
- `costhistoryapp/model.go`

### 2.4 Database Migrations

Add migration to:
1. Add `currency_id UUID` column to each table
2. Populate with default currency (USD) for existing data
3. Add FK constraints
4. Drop old `currency VARCHAR` columns

Tables affected:
- `sales.orders`
- `sales.customers` (add default_currency_id)
- `products.product_costs`
- `procurement.purchase_orders`
- `products.cost_history`

---

## Phase 3: Template.go Currency Formatting Update

**Goal:** Update workflow template to use currency data from database.

### 3.1 Current Implementation

**File:** `business/sdk/workflow/template.go` (lines 676-696)

Current hardcoded switch:
```go
switch currency {
case "EUR": symbol = "€"
case "GBP": symbol = "£"
case "JPY": symbol = "¥"
default: symbol = "$"
}
```

### 3.2 Proposed Changes

Options to consider:
1. **Simple enhancement**: Add all 10 currency symbols to the switch statement
2. **Dynamic lookup**: Query currencybus for symbol based on code
3. **Hybrid**: Maintain switch for performance but expand coverage

Recommended: Option 1 for Phase 3, with enhancement to use decimal_places for formatting:

```go
"currency": func(value interface{}, args ...string) (interface{}, error) {
    num, err := toFloat64(value)
    if err != nil {
        return value, err
    }
    currency := "USD"
    if len(args) > 0 {
        currency = args[0]
    }
    symbol := "$"
    decimals := 2
    switch currency {
    case "EUR": symbol = "€"
    case "GBP": symbol = "£"
    case "JPY": symbol = "¥"; decimals = 0
    case "CHF": symbol = "CHF"
    case "CNY": symbol = "¥"
    case "INR": symbol = "₹"
    case "MXN": symbol = "$"
    case "CAD": symbol = "$"
    case "AUD": symbol = "$"
    }
    format := fmt.Sprintf("%%s%%.%df", decimals)
    return fmt.Sprintf(format, symbol, num), nil
},
```

---

## Phase 4: Backend Impact Analysis and Additional Updates

**Goal:** Identify and update any other backend areas affected by this change.

### 4.1 Areas to Investigate

1. **Seed Models** (`business/sdk/dbtest/seedmodels/`)
   - `tables.go` - Currency type fields
   - `tableforms.go` - Form configurations with currency defaults

2. **FormData Registry** (`api/cmd/services/ichor/build/all/formdata_registry.go`)
   - Register currency entity for multi-entity transactions

3. **API Tests** (all affected domains)
   - Update test fixtures to use currency IDs
   - Update expected responses

4. **Frontend Seed Data**
   - Page configs referencing currency fields
   - Form configurations

### 4.2 Specific File Updates

**formdata_registry.go:**
- Add currency entity registration
- Update existing registrations that use currency strings

**Affected Integration Tests:**
- `api/cmd/services/ichor/tests/sales/ordersapi/`
- `api/cmd/services/ichor/tests/products/productcostapi/`
- `api/cmd/services/ichor/tests/procurement/purchaseorderapi/`

### 4.3 Backwards Compatibility

Consider:
- API response changes (string → UUID)
- Any external integrations expecting currency codes
- Migration strategy for existing data

---

## Implementation Order

1. **Phase 1** (Core Domain): Complete currency domain with tests - establishes foundation
2. **Phase 2** (Test Utils + Related Domains): Update dependencies - requires Phase 1
3. **Phase 3** (Template): Update workflow formatting - can be done in parallel with Phase 2
4. **Phase 4** (Impact Analysis): Final cleanup - after Phases 2 & 3

---

## Validation Checklist

After each phase:
- [ ] `make lint` passes
- [ ] `make test` passes
- [ ] `make build` succeeds
- [ ] Manual API testing confirms endpoints work

---

## Notes

- Use "TST" or "XYZ" as test currency code to ensure tests create fresh data
- All JSON tags should use `snake_case` per codebase conventions
- Maintain backwards compatibility in API responses where possible
- Consider adding a "queryByCode" endpoint for convenience lookups
