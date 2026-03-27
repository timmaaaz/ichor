# Progress Summary: seeding.md

## Overview
Go-based test seeding pipeline. Generates randomized, meaningful test data for unit/integration testing. Contrast with customer seeding: provides random data for testing, not stable data for production.

## TestDataReference [test]

All values are randomized (names, IDs, dates). Only counts and relationships are fixed.

### Counts

| Entity | Count | Notes |
|--------|-------|-------|
| Admin users | 1 | foundation.Admins[0] — creator/approver throughout |
| Regular users (reporters) | 20 | role=User |
| Regular users (bosses) | 10 | role=User |
| Titles | 10 | |
| Currencies | 5 + USD | USD from seed.sql; 5 additional test |
| Cities | 5 | |
| Streets | 5 | |
| Contact infos | 5 | |
| Offices | 10 | |
| Brands | 5 | |
| Product categories | 10 | |
| Products | 20 | historical, 180-day window |
| Product costs | 20 | all USD |
| Physical attributes | 20 | |
| Metrics | 40 | |
| Cost history | 40 | historical, 180-day window |
| Warehouses | 5 | historical, 365-day window |
| Zones | 12 | |
| Inventory locations | 25 | |
| Inventory items | 30 | |
| Transfer orders | 20 | |
| Inventory transactions | 40 | |
| Inventory adjustments | 20 | |
| Customers | 5 | historical, 180-day window |
| Orders | 200 | weighted random, 90-day window |
| Order line items | 5 per order | |
| Order fulfillment statuses | from seedmodels.OrderFulfillmentStatusData | named |
| Line item fulfillment statuses | 9 | ALLOCATED, CANCELLED, PACKED, PENDING, PICKED, PARTIALLY_PICKED, BACKORDERED, PENDING_REVIEW, SHIPPED |
| Suppliers | 25 | |
| Supplier products | 10 | |
| Purchase orders | 10 | historical, 120-day window |
| PO line items | 25 | |
| Lot trackings | 15 | |
| Lot locations | 15 | |
| Inspections | 10 | |
| Serial numbers | 50 | |
| Action templates | 15 | workflow |
| Default automation rules | 8 | workflow |

### Fixed UUIDs (Safe to Hardcode)

**Source:** `business/sdk/migrate/sql/seed.sql` — identical on every environment.

#### Users
```
admin_gopher  id=5cf37266-3473-4006-984f-9325122678b7  roles={ADMIN}  email=admin@example.com
user_gopher   id=45b5fbd3-755f-4379-8f07-a58d4a30fa2f  roles={USER}   email=user@example.com
```

#### Role
```
ZZZADMIN  id=54bb2165-71e1-41a6-af3e-7da4a0e1e2c1
```
(admin_gopher linked via user_roles)

#### HR User Approval Statuses
```
PENDING      id=89173300-3f4e-4606-872c-f34914bbee19
APPROVED     id=0394acac-ace4-4e8f-b64e-68625b0af14a
DENIED       id=7b901e2e-3f33-40c1-9201-b4e8b1718b4b
UNDER REVIEW id=132a2572-b7a0-4b56-a165-55e1c244c3e2
```

#### Payment Terms (Patterned IDs)
Both name and ID are stable:
```
a0000000-0000-4000-8000-000000000001  Net 30
a0000000-0000-4000-8000-000000000002  Net 60
a0000000-0000-4000-8000-000000000003  Due on Receipt
...
a0000000-0000-4000-8000-000000000018  Letter of Credit
```

### Stable by Name/Code (Query These — IDs Random per Environment)

#### Currencies (10 total — query by `code`)
USD, EUR, GBP, CAD, AUD, JPY, CHF, CNY, INR, MXN

#### Geography (All use `uuid_generate_v4()` — query by code/name)
- `geography.countries` ~250 rows — query by alpha_2 (e.g. 'US') or alpha_3
- `geography.regions` 50 US states — query by code (e.g. 'CA', 'NY', 'TX')
- `geography.timezones` ~37 rows — query by name

#### Asset Conditions (Query by Name)
PERFECT, GOOD, USED, POOR, END_OF_LIFE

#### Asset Approval/Fulfillment Statuses (Query by Name)
SUCCESS, ERROR, WAITING, REJECTED, IN_PROGRESS

#### Named Users (18 Additional — Query by `username`)
manager1, manager2, finance_admin, hr_admin, employee1–4, readonly, temp_admin, sales_east, sales_west, it_systems, it_dev, accounting, payroll, recruitment, benefits

All use password hash for `admin123` except admin_gopher.

### Stable from seedmodels (Query by Name)

#### Order Fulfillment Statuses
PENDING, PROCESSING, PICKING, PACKING, READY_TO_SHIP, SHIPPED, DELIVERED, CANCELLED

#### Line Item Fulfillment Statuses
ALLOCATED, CANCELLED, PACKED, PENDING, PICKED, PARTIALLY_PICKED, BACKORDERED, PENDING_REVIEW, SHIPPED

#### Purchase Order Statuses
DRAFT, PENDING_APPROVAL, APPROVED, SENT, PARTIALLY_RECEIVED, RECEIVED, CANCELLED, CLOSED

#### PO Line Item Statuses
PENDING, ORDERED, PARTIALLY_RECEIVED, RECEIVED, BACKORDERED, CANCELLED

### What is Random (Never Hardcode)
- All UUIDs not listed above — always query by stable name/code/status
- All names, descriptions, addresses, emails from TestSeed* functions
- All dates — historical relative to seed time; assert on ranges, not exact values
- All monetary amounts

## NewDatabase [dbtest] — `business/sdk/dbtest/dbtest.go`

```go
func NewDatabase(t *testing.T, testName string) *Database

type Database struct {
    DB        *sqlx.DB
    Log       *logger.Logger
    BusDomain BusDomain
}
```

### Key Facts
- **DB name:** random 4 lowercase letters (`abcdefghijklmnopqrstuvwxyz`)
- **TimeZone:** `SET TIME ZONE 'America/New_York'`
- **Runs migrations + seeds** before returning
- **Each test gets its own isolated database instance**
- **Two seeding entry points:**
  - `InsertSeedData` — live K8s/Docker DB for `make seed-frontend`
  - `NewDatabase` — isolated test DB
- **Both use same BusDomain struct and seed_*.go chain**

## BusDomain [dbtest]

**Central struct holding every instantiated business package.**

Passed into all seed functions. Contains 60+ bus instances:

**Geography:** Country, Region, City, Street, Timezone, Home

**HR / Users:** User, Title, Office, ReportsTo, UserApprovalStatus, UserApprovalComment

**Core:** Role, UserRole, TableAccess, Permissions, Page, RolePage, Introspection, ActionPermissions

**Assets:** ApprovalStatus, FulfillmentStatus, Tag, AssetTag, ValidAsset, AssetType, AssetCondition, UserAsset, Asset

**Products:** Brand, ProductCategory, Product, ProductUOM, PhysicalAttribute, ProductCost, Metrics, CostHistory

**Procurement:** Supplier, SupplierProduct, PaymentTerm, PurchaseOrderStatus, PurchaseOrderLineItemStatus, PurchaseOrder, PurchaseOrderLineItem, LotTrackings, LotLocation, Inspection

**Inventory:** Warehouse, Zones, InventoryLocation, InventoryItem, SerialNumber, InventoryTransaction, InventoryAdjustment, TransferOrder

**Sales:** ContactInfos, Customers, Currency, Order, OrderLineItem, OrderFulfillmentStatus, LineItemFulfillmentStatus

**Config:** ConfigStore, TableStore, Form, FormField, PageAction, PageConfig, PageContent

**Workflow:** Workflow, Alert, Notification

## SeedOrchestrator [dbtest] — `business/sdk/dbtest/seedFrontend.go`

**Execution order is authoritative — do not reorder.**

```
seedFoundation(ctx, busDomain)
  ↓
seedGeographyHR(ctx, busDomain)
  ↓
seedAssets(ctx, busDomain, foundation)
  ↓
seedProducts(ctx, busDomain, geoHR, foundation)
  ↓
seedInventory(ctx, busDomain, foundation, geoHR, products)
  ↓
seedSales(ctx, busDomain, foundation, geoHR, products)
  ↓
seedProcurement(ctx, busDomain, foundation, geoHR, products, inventory)
  ↓
seedTableBuilder(ctx, busDomain, adminID)
  ↓
seedPages(ctx, log, busDomain)
  ↓
seedForms(ctx, log, busDomain)
  ↓
seedWorkflow(ctx, log, busDomain, adminID)
```

`adminID` extracted from `foundation.Admins[0].ID` after `seedFoundation`.

## SeedFunctions [dbtest]

### seed_foundation.go
**Returns:** FoundationSeed (Admins, Reporters, Bosses, USDCurrencyID, Currencies)

Seeds: users, roles, user roles, table access, currencies.

### seed_geography_hr.go
**Returns:** GeographyHRSeed (Cities, Streets, ContactInfos, Offices)

Seeds: cities, streets, contact infos, offices, titles, reports-to, approval comments.

### seed_assets.go
Seeds: asset types, conditions, valid assets, assets, approval/fulfillment statuses, tags, user assets.

### seed_products.go
**Returns:** ProductsSeed

Seeds: brands, categories, products, costs, physical attributes, metrics, cost histories.

### seed_inventory.go
**Returns:** InventorySeed

Seeds: warehouses, zones, locations, inventory items, lot trackings, lot locations, serial numbers, inspections, transfer orders, transactions, adjustments.

### seed_sales.go
Seeds: customers, order fulfillment statuses, orders, line item fulfillment statuses, order line items.

### seed_procurement.go
Seeds: suppliers, supplier products, PO statuses, POs, PO line items.

### seed_tablebuilder.go
Seeds stored JSON column config blobs for 20+ modules via `busDomain.TableStore`.

References static configs from `seedmodels/tables_*.go` and `seedmodels/charts.go`.

### seed_pages.go
Seeds: page configs, page content blocks (tables + charts), page action buttons, nav pages, role-page access.

References `seedmodels/pageactions.go` and `seedmodels/pages.go`.

### seed_forms.go
**Key:** uses `busDomain.Workflow.QueryEntityByName` (read-only lookup) — must run after `seedFoundation`, NOT after `seedWorkflow`

Seeds: forms and all form fields.

References `seedmodels/tableforms.go`, `seedmodels/forms.go`, `seedmodels/formregistry.go`

### seed_workflow.go
**Must run LAST** — depends on forms and templates already seeded.

Seeds: 15 action templates + 8 default automation rules.

## SeedModels [seedmodels] — `business/sdk/dbtest/seedmodels/`

| File | Contents |
|------|----------|
| tables_admin.go | admin module table configs |
| tables_assets.go | asset table configs |
| tables_hr.go | HR/employee table configs |
| tables_inventory.go | warehouse/items/adjustments |
| tables_procurement.go | PO/supplier table configs |
| tables_products.go | product table configs |
| tables_sales.go | orders/customers/line items |
| charts.go | 14 chart type configs |
| tableforms.go | 48 entity form field generators |
| forms.go | 4 composite form field generators |
| formregistry.go | form registry + ~50 registrations |
| pageactions.go | page button action definitions |
| pages.go | allPages slice — nav page route records |
| enums.go | status name slices |

## Permission Seeding — Two-Layer System

### Layer 1: Page Access
- **File:** `business/sdk/dbtest/seedmodels/pages.go` → AllPages slice
- **seedPages()** auto-grants can_access=true to ZZZADMIN for every AllPages entry
- **Missing:** router redirects to /login; page never renders

### Layer 2: Table Access
- **File:** `business/sdk/migrate/sql/seed.sql` → core.table_access INSERT block
- **Format:** (gen_random_uuid(), '<ZZZADMIN_id>', 'schema.table', true, true, true, true)
- **Missing:** 401 "user does not have permission READ for table: ..."

### Layer 3: Action Permissions
- **File:** `business/sdk/migrate/sql/seed.sql` → workflow.action_permissions INSERT block
- **Missing:** 403 "permission_denied" on POST /v1/workflow/actions/{type}/execute

### Symptom → Fix
| Symptom | Fix |
|---------|-----|
| Redirect to /login | Add to AllPages |
| 401 "no permission for table" | Add to seed.sql table_access |
| 403 "permission_denied" | Add to seed.sql action_permissions |

## Change Patterns

### ⚠ Adding a New Domain to Seeding
Affects 4 areas:
1. `business/sdk/dbtest/dbtest.go` — add bus field(s) to BusDomain struct + instantiate in newBusDomains
2. `business/sdk/dbtest/seed_<domain>.go` — create new seed file with seed<Domain>() function + result struct
3. `business/sdk/dbtest/seedFrontend.go` — add call in dependency chain at correct position
4. `business/sdk/dbtest/seedmodels/tables_<domain>.go` — add static table configs if needed

### ⚠ Adding a New TestSeed* Function
Affects 2 areas:
1. `business/domain/{area}/{entity}bus/testutil.go` — function lives here in bus package, NOT in test file
2. `api/cmd/services/ichor/tests/{area}/{entity}api/seed_test.go` — call from insertSeedData in integration tests

**Key:** TestSeed* functions live in {entity}bus/testutil.go (importable from integration tests without circular imports).

**Signature pattern:**
```go
func TestSeed{Entities}(ctx context.Context, n int, ...depIDs, api *Business) ([]Entity, error)
```

### ⚠ Changing Seed Ordering (Dependency Chain)
Affects 1 area:
1. `business/sdk/dbtest/seedFrontend.go` — orchestrator chain (order is authoritative)

**Dependency rule:** A seed function may only reference IDs from result structs of functions called before it. Never reach into busDomain for IDs that should come from an earlier seed result struct.

## Critical Points
- **Fixed UUIDs are environment-invariant** — safe to hardcode admin_gopher, ZZZADMIN, payment terms
- **Random values never hardcoded** — always query by stable name/code/status
- **Deterministic counts** — same for every test (reproducible)
- **Historical windows** — dates relative to seed time; assert on ranges, not exact values
- **Dependency chain matters** — seed order is load-bearing; don't reorder

## Notes for Future Development
Test seeding is well-designed for reproducibility with randomization. Most changes will be:
- Adding new TestSeed* functions (straightforward, lives in bus package testutil.go)
- Adding new domains (moderate, requires orchestrator entry + models)
- Changing seed ordering (risky, breaks dependency assumptions)

The fixed UUID set (admin_gopher, ZZZADMIN, etc.) is intentional and stable — never regenerate these in seed.sql.
