# seeding

[bus]=business [app]=application [api]=HTTP [db]=store [sdk]=shared [dbtest]=test-infra [seedmodels]=static-seed-data
→=depends on ⊕=writes ⊗=reads ⚡=external [tx]=transaction [cache]=cached

---

## TestDataReference [test]

All values are randomized (names, IDs, dates). Only counts and relationships are fixed.

| Entity | Count | Notes |
|--------|-------|-------|
| Admin users | 1 | `foundation.Admins[0]` — used as creator/approver throughout |
| Regular users (reporters) | 20 | role=User |
| Regular users (bosses) | 10 | role=User |
| Titles | 10 | |
| Currencies | 5 + USD | USD queried from seed.sql; 5 additional test currencies |
| Cities | 5 | |
| Streets | 5 | |
| Contact infos | 5 | |
| Offices | 10 | |
| Brands | 5 | |
| Product categories | 10 | |
| Products | 20 | historical dates, 180-day window |
| Product costs | 20 | all USD |
| Physical attributes | 20 | |
| Metrics | 40 | |
| Cost history entries | 40 | historical, 180-day window |
| Warehouses | 5 | historical, 365-day window |
| Zones | 12 | |
| Inventory locations | 25 | |
| Inventory items | 30 | |
| Transfer orders | 20 | |
| Inventory transactions | 40 | |
| Inventory adjustments | 20 | |
| Customers | 5 | historical, 180-day window |
| Orders | 200 | weighted random distribution, 90-day window |
| Order line items | 5 per order | |
| Order fulfillment statuses | from seedmodels.OrderFulfillmentStatusData | named statuses |
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

### Fixed UUIDs (safe to hardcode in tests)

Source: `business/sdk/migrate/sql/seed.sql` — identical on every environment.

Users:
```
admin_gopher  id=5cf37266-3473-4006-984f-9325122678b7  roles={ADMIN}  email=admin@example.com
user_gopher   id=45b5fbd3-755f-4379-8f07-a58d4a30fa2f  roles={USER}   email=user@example.com
```

Role:
```
ZZZADMIN  id=54bb2165-71e1-41a6-af3e-7da4a0e1e2c1
```
(admin_gopher is linked to this role via user_roles)

HR user approval statuses:
```
PENDING      id=89173300-3f4e-4606-872c-f34914bbee19
APPROVED     id=0394acac-ace4-4e8f-b64e-68625b0af14a
DENIED       id=7b901e2e-3f33-40c1-9201-b4e8b1718b4b
UNDER REVIEW id=132a2572-b7a0-4b56-a165-55e1c244c3e2
```

Payment terms (patterned IDs — both name and ID are stable):
```
a0000000-0000-4000-8000-000000000001  Net 30
a0000000-0000-4000-8000-000000000002  Net 60
a0000000-0000-4000-8000-000000000003  Due on Receipt
a0000000-0000-4000-8000-000000000004  Net 15
a0000000-0000-4000-8000-000000000005  Net 45
a0000000-0000-4000-8000-000000000006  Net 7
a0000000-0000-4000-8000-000000000007  Net 10
a0000000-0000-4000-8000-000000000008  Net 21
a0000000-0000-4000-8000-000000000009  Net 90
a0000000-0000-4000-8000-000000000010  Prepaid
a0000000-0000-4000-8000-000000000011  COD
a0000000-0000-4000-8000-000000000012  CIA
a0000000-0000-4000-8000-000000000013  50% Deposit
a0000000-0000-4000-8000-000000000014  EOM
a0000000-0000-4000-8000-000000000015  MFI
a0000000-0000-4000-8000-000000000016  15 MFI
a0000000-0000-4000-8000-000000000017  Open Account
a0000000-0000-4000-8000-000000000018  Letter of Credit
```

### Stable by name/code (query by these — IDs are random per environment)

Currencies (10 total — query by `code`):
```
USD, EUR, GBP, CAD, AUD, JPY, CHF, CNY, INR, MXN
```

Geography — all use `uuid_generate_v4()` so query by code:
```
geography.countries  ~250 rows  — query by alpha_2 (e.g. 'US') or alpha_3
geography.regions    50 US states — query by code (e.g. 'CA', 'NY', 'TX')
geography.timezones  ~37 rows   — query by name
```

Asset conditions (query by name):
```
PERFECT, GOOD, USED, POOR, END_OF_LIFE
```

Asset approval/fulfillment statuses (query by name):
```
SUCCESS, ERROR, WAITING, REJECTED, IN_PROGRESS
```

Named users (18 additional users with `gen_random_uuid()` IDs — query by `username`):
```
manager1, manager2, finance_admin, hr_admin, employee1–4,
readonly, temp_admin, sales_east, sales_west,
it_systems, it_dev, accounting, payroll, recruitment, benefits
```
All use password hash for `admin123` except admin_gopher.

### Stable from seedmodels (seeded via InsertSeedData / TestSeed chain)

Order fulfillment statuses (query by name):
```
PENDING, PROCESSING, PICKING, PACKING, READY_TO_SHIP, SHIPPED, DELIVERED, CANCELLED
```

Line item fulfillment statuses (query by name):
```
ALLOCATED, CANCELLED, PACKED, PENDING, PICKED, PARTIALLY_PICKED, BACKORDERED, PENDING_REVIEW, SHIPPED
```

Purchase order statuses (query by name):
```
DRAFT, PENDING_APPROVAL, APPROVED, SENT, PARTIALLY_RECEIVED, RECEIVED, CANCELLED, CLOSED
```

PO line item statuses (query by name):
```
PENDING, ORDERED, PARTIALLY_RECEIVED, RECEIVED, BACKORDERED, CANCELLED
```

### What is random (never hardcode)

- All UUIDs not listed above — always query by stable name/code/status
- All names, descriptions, addresses, emails from `TestSeed*` functions
- All dates — historical relative to seed time; assert on ranges, not exact values
- All monetary amounts

---

## NewDatabase [dbtest]

file: business/sdk/dbtest/dbtest.go

```go
func NewDatabase(t *testing.T, testName string) *Database
```

returns:
```go
type Database struct {
    DB        *sqlx.DB
    Log       *logger.Logger
    BusDomain BusDomain
}
```

key facts:
  - DB name: random 4 lowercase letters (`abcdefghijklmnopqrstuvwxyz`)
  - TimeZone: `SET TIME ZONE 'America/New_York'`
  - Runs migrations + seeds before returning
  - Each test gets its own isolated database instance
  - Two seeding entry points: InsertSeedData (live K8s/Docker DB for `make seed-frontend`) + NewDatabase (isolated test DB)
  - Both use the same BusDomain struct and same seed_*.go chain

---

## BusDomain [dbtest]

file: business/sdk/dbtest/dbtest.go
key facts:
  - Central struct holding every instantiated business package
  - Passed into all seed functions

```
Delegate

// Geography
Country, Region, City, Street, Timezone, Home

// HR / Users
User, Title, Office, ReportsTo, UserApprovalStatus, UserApprovalComment

// Core
Role, UserRole, TableAccess, Permissions, Page, RolePage, Introspection
ActionPermissions

// Assets
ApprovalStatus, FulfillmentStatus, Tag, AssetTag, ValidAsset
AssetType, AssetCondition, UserAsset, Asset

// Products
Brand, ProductCategory, Product, ProductUOM, PhysicalAttribute, ProductCost, Metrics, CostHistory

// Procurement
Supplier, SupplierProduct, PaymentTerm
PurchaseOrderStatus, PurchaseOrderLineItemStatus, PurchaseOrder, PurchaseOrderLineItem
LotTrackings, LotLocation, Inspection

// Inventory
Warehouse, Zones, InventoryLocation, InventoryItem, SerialNumber
InventoryTransaction, InventoryAdjustment, TransferOrder

// Sales
ContactInfos, Customers, Currency, Order, OrderLineItem
OrderFulfillmentStatus, LineItemFulfillmentStatus

// Config
ConfigStore, TableStore, Form, FormField
PageAction, PageConfig, PageContent

// Workflow
Workflow, Alert, Notification
```

---

## SeedOrchestrator [dbtest]

file: business/sdk/dbtest/seedFrontend.go (~71 lines)

```go
func InsertSeedData(log *logger.Logger, cfg sqldb.Config) error
```

Dependency chain (execution order is authoritative — do not reorder):

```
seedFoundation(ctx, busDomain)                                  → FoundationSeed
    ↓
seedGeographyHR(ctx, busDomain)                                 → GeographyHRSeed
    ↓
seedAssets(ctx, busDomain, foundation)
    ↓
seedProducts(ctx, busDomain, geoHR, foundation)                 → products result
    ↓
seedInventory(ctx, busDomain, foundation, geoHR, products)      → inventory result
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

`adminID` is extracted from `foundation.Admins[0].ID` after `seedFoundation`.

---

## SeedFunctions [dbtest]

### seed_foundation.go
```go
type FoundationSeed struct {
    Admins        []userbus.User
    Reporters     []userbus.User
    Bosses        []userbus.User
    USDCurrencyID uuid.UUID
    Currencies    []currencybus.Currency
}
func seedFoundation(ctx context.Context, busDomain BusDomain) (FoundationSeed, error)
```
Seeds: users, roles, user roles, table access, currencies.

### seed_geography_hr.go
```go
type GeographyHRSeed struct {
    Cities       []citybus.City
    Streets      []streetbus.Street
    ContactInfos []contactinfosbus.ContactInfos
    Offices      []officebus.Office
}
func seedGeographyHR(ctx context.Context, busDomain BusDomain) (GeographyHRSeed, error)
```
Seeds: cities, streets, contact infos, offices, titles, reports-to, approval comments.

### seed_assets.go
```go
func seedAssets(ctx context.Context, busDomain BusDomain, foundation FoundationSeed) error
```
Seeds: asset types, conditions, valid assets, assets, approval/fulfillment statuses, tags, user assets.

### seed_products.go
```go
func seedProducts(ctx context.Context, busDomain BusDomain, geoHR GeographyHRSeed, foundation FoundationSeed) (ProductsSeed, error)
```
Seeds: brands, categories, products, costs, physical attributes, metrics, cost histories.

### seed_inventory.go
```go
func seedInventory(ctx context.Context, busDomain BusDomain, foundation FoundationSeed, geoHR GeographyHRSeed, products ProductsSeed) (InventorySeed, error)
```
Seeds: warehouses, zones, locations, inventory items, lot trackings, lot locations, serial numbers,
inspections, transfer orders, transactions, adjustments.

### seed_sales.go
```go
func seedSales(ctx context.Context, busDomain BusDomain, foundation FoundationSeed, geoHR GeographyHRSeed, products ProductsSeed) error
```
Seeds: customers, order fulfillment statuses, orders, line item fulfillment statuses, order line items.

### seed_procurement.go
```go
func seedProcurement(ctx context.Context, busDomain BusDomain, foundation FoundationSeed, geoHR GeographyHRSeed, products ProductsSeed, inventory InventorySeed) error
```
Seeds: suppliers, supplier products, PO statuses, POs, PO line items.

### seed_tablebuilder.go
```go
func seedTableBuilder(ctx context.Context, busDomain BusDomain, adminID uuid.UUID) error
```
Seeds stored JSON column config blobs for 20+ modules via `busDomain.TableStore`.
References static configs from `seedmodels/tables_*.go` and `seedmodels/charts.go`.

### seed_pages.go
```go
func seedPages(ctx context.Context, log *logger.Logger, busDomain BusDomain) error
```
Seeds: page configs, page content blocks (tables + charts), page action buttons, nav pages, role-page access.
References `seedmodels/pageactions.go` and `seedmodels/pages.go`.

### seed_forms.go
```go
func seedForms(ctx context.Context, log *logger.Logger, busDomain BusDomain) error
```
Seeds: forms and all form fields.
key facts:
  - uses `busDomain.Workflow.QueryEntityByName` (read-only lookup) — must run after `seedFoundation`, NOT after `seedWorkflow`
  - References `seedmodels/tableforms.go`, `seedmodels/forms.go`, `seedmodels/formregistry.go`

### seed_workflow.go
```go
func seedWorkflow(ctx context.Context, log *logger.Logger, busDomain BusDomain, adminID uuid.UUID) error
```
Seeds: 15 action templates + 8 default automation rules.
Must run LAST — depends on forms and templates already seeded.

---

## SeedModels [seedmodels]

file: business/sdk/dbtest/seedmodels/

| File | Contents |
|------|----------|
| `tables_admin.go` | admin module table configs |
| `tables_assets.go` | asset table configs |
| `tables_hr.go` | HR/employee table configs |
| `tables_inventory.go` | warehouse/items/adjustments configs |
| `tables_procurement.go` | PO/supplier table configs |
| `tables_products.go` | product table configs |
| `tables_sales.go` | orders/customers/line items configs |
| `charts.go` | 14 chart type configs |
| `tableforms.go` | 48 entity form field generators (mixed domains) |
| `forms.go` | 4 composite form field generators |
| `formregistry.go` | form registry + ~50 registrations |
| `pageactions.go` | page button action definitions |
| `pages.go` | `allPages` slice — nav page route records |
| `enums.go` | status name slices (approval, fulfillment, etc.) |

---

## ConfigPages [dbtest]

file: business/sdk/dbtest/seed_config_pages.go
key facts:
  - Holds the `allPages` slice used by `seedPages` for navigation page route records
  - Renamed from the original `model.go`

---

## ⚠ Permission seeding — two-layer system

Layer 1: Page access (router → canAccessPage)
  file: business/sdk/dbtest/seedmodels/pages.go → AllPages slice
  seedPages() auto-grants can_access=true to ZZZADMIN for every AllPages entry
  missing → router redirects to /login; page never renders

Layer 2: Table access (API auth middleware → core.table_access)
  file: business/sdk/migrate/sql/seed.sql → core.table_access INSERT block
  ZZZADMIN id: 54bb2165-71e1-41a6-af3e-7da4a0e1e2c1
  format: (gen_random_uuid(), '<ZZZADMIN_id>', 'schema.table', true, true, true, true)
  missing → 401 "user does not have permission READ for table: ..."

Layer 3: Action permissions (workflow.action_permissions)
  file: business/sdk/migrate/sql/seed.sql → workflow.action_permissions INSERT block
  missing → 403 "permission_denied" on POST /v1/workflow/actions/{type}/execute

symptom → fix:
  redirect to /login             → add to AllPages
  401 "no permission for table"  → add to seed.sql table_access
  403 "permission_denied"        → add to seed.sql action_permissions

---

## ⚠ Form field seeding

  fields: seedmodels/tableforms.go → Get<Entity>FormFields(formID, entityID)
  registration: seedmodels/formregistry.go → RegisterForm("Form Name", supportsUpdate, GetFn)
  name in RegisterForm must match loadFullFormConfigByName("Form Name") in frontend
  changes require re-seed (make dev-bounce) — not hot-reloaded

  symptom → fix:
    "X is required" but no input visible  → X missing from Get*FormFields in tableforms.go
    null columns after submit              → column missing from Get*FormFields
    inline_create sub-form fails to load   → "form_name" in Config JSON not registered in FormRegistry

---

## ⚠ Adding a new domain to seeding

  business/sdk/dbtest/dbtest.go          (add bus field(s) to BusDomain struct + instantiate in newBusDomains)
  business/sdk/dbtest/seed_<domain>.go   (create new seed file with seed<Domain>() function + result struct)
  business/sdk/dbtest/seedFrontend.go    (add call in dependency chain at correct position)
  business/sdk/dbtest/seedmodels/tables_<domain>.go  (add static table configs if needed)

## ⚠ Adding a new TestSeed* function

  business/domain/{area}/{entity}bus/testutil.go    (function lives here in the bus package, NOT in a test file)
  api/cmd/services/ichor/tests/{area}/{entity}api/seed_test.go   (call from insertSeedData in integration tests)

  Key: TestSeed* functions live in {entity}bus/testutil.go so they are importable from integration tests
  without circular imports. Signature pattern:
    func TestSeed{Entities}(ctx context.Context, n int, ...depIDs, api *Business) ([]Entity, error)

## ⚠ Changing seed ordering (dependency chain)

  business/sdk/dbtest/seedFrontend.go    (orchestrator chain — order is authoritative)
  Dependency rule: a seed function may only reference IDs from result structs of
  functions called before it. Never reach into busDomain for IDs that should come
  from an earlier seed result struct.
