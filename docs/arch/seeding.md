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
| Products | 40 | historical dates, 180-day window; tracking distribution 28 none / 8 lot / 4 serial |
| Product costs | 40 | all USD |
| Physical attributes | 40 | |
| Metrics | 80 | |
| Cost history entries | 80 | historical, 180-day window |
| Label catalog | 79 | 19 location + 20 container + 40 product (one per seeded product) |
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
| Put-away tasks | 15 | first assigned to floor_worker1 |
| Pick tasks | 15 | cycle over 5 line items via modulo |
| Action templates | 15 | workflow |
| Default automation rules | 8 | workflow |
| Cycle count sessions | 3 | |
| Cycle count items | 15 | 5 per session |
| Automation executions | 5 | created for approval request FK |
| Approval requests | 5 | pending, for supervisor inbox |

### Fixed UUIDs (safe to hardcode in tests)

Source: `business/sdk/migrate/sql/seed.sql` — identical on every environment.

Users:
```
admin_gopher   id=5cf37266-3473-4006-984f-9325122678b7  roles={ADMIN}  email=admin@example.com
user_gopher    id=45b5fbd3-755f-4379-8f07-a58d4a30fa2f  roles={USER}   email=user@example.com
floor_worker1  id=c0000000-0000-4000-8000-000000000001  roles={USER}   email=floor_worker1@example.com  password=admin123
```

Roles:
```
ZZZADMIN      id=54bb2165-71e1-41a6-af3e-7da4a0e1e2c1
FLOOR_WORKER  id=b0000000-0000-4000-8000-000000000001
```
(admin_gopher is linked to ZZZADMIN, floor_worker1 is linked to FLOOR_WORKER via user_roles)

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
  - Three seeding entry points: InsertSeedData (live K8s/Docker DB for `make seed-frontend`), NewDatabase (isolated test DB), and InsertPlatformConfig (`make seed-platform` — fresh customer DB bootstrap, platform config only, no demo data)
  - All use the same BusDomain struct and same seed_*.go functions

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

// Labels
Label

// Procurement
Supplier, SupplierProduct, PaymentTerm
PurchaseOrderStatus, PurchaseOrderLineItemStatus, PurchaseOrder, PurchaseOrderLineItem
LotTrackings, LotLocation, Inspection

// Inventory
Warehouse, Zones, InventoryLocation, InventoryItem, SerialNumber
InventoryTransaction, InventoryAdjustment, TransferOrder
PutAwayTask, PickTask

// Sales
ContactInfos, Customers, Currency, Order, OrderLineItem
OrderFulfillmentStatus, LineItemFulfillmentStatus

// Config
ConfigStore, TableStore, Form, FormField, Settings
PageAction, PageConfig, PageContent

// Workflow
Workflow, Alert, Notification, ApprovalRequest

// Inventory (cycle counts)
CycleCountSession, CycleCountItem

// Scenarios
Scenario
```

---

## SeedOrchestrator [dbtest]

file: business/sdk/dbtest/seedFrontend.go (~71 lines)

```go
func InsertSeedData(log *logger.Logger, cfg sqldb.Config) error
func InsertSeedDataWithDB(log *logger.Logger, db *sqlx.DB) error
```

`InsertSeedData` opens a connection from `cfg`, defers Close, and delegates
to `InsertSeedDataWithDB`. Integration tests (e.g. `Test_Seed_Integration`)
call `InsertSeedDataWithDB` directly against the `dbtest.NewDatabase`-managed
`*sqlx.DB` so the seed chain reuses the test container's connection.

Dependency chain (execution order is authoritative — do not reorder):

```
seedFoundation(ctx, busDomain)                                  → FoundationSeed
    ↓
seedGeographyHR(ctx, busDomain)                                 → GeographyHRSeed
    ↓
seedAssets(ctx, busDomain, foundation)
    ↓
seedProducts(ctx, busDomain, geoHR, foundation)                 → ProductsSeed
    ↓
seedLabels(ctx, busDomain.Label, products)                      → 79 deterministic catalog rows
    ↓
seedInventory(ctx, busDomain, foundation, geoHR, products)      → inventory result
    ↓
seedSales(ctx, busDomain, foundation, geoHR, products)          → SalesSeed
    ↓
seedProcurement(ctx, busDomain, foundation, geoHR, products, inventory)
    ↓
seedTasks(ctx, busDomain, foundation, products, inventory, sales) → TasksSeed
    ↓
seedTableBuilder(ctx, busDomain, adminID)
    ↓
seedPages(ctx, log, busDomain)
    ↓
seedForms(ctx, log, busDomain, db)
    ↓
seedWorkflow(ctx, log, busDomain, adminID)
    ↓
seedAlerts(ctx, log, busDomain, adminID)
    ↓
seedCycleCounts(ctx, busDomain, foundation, products, inventory)
    ↓
seedApprovals(ctx, busDomain, foundation)                         → queries rules from seedWorkflow
    ↓
seedSettings(ctx, busDomain)                                      → 11 scan-discipline lever rows (pick.productScan locked; 10 overridable by scenarios)
    ↓
seedScenarios(ctx, busDomain)                                     → loads YAML fixtures from deployments/scenarios
```

`adminID` is extracted from `foundation.Admins[0].ID` after `seedFoundation`.

---

## InsertPlatformConfig [dbtest]

file: business/sdk/dbtest/seedFrontend.go (`InsertPlatformConfig`)

```go
func InsertPlatformConfig(log *logger.Logger, cfg sqldb.Config) error
```

Subset of the orchestrator above. Used by `make seed-platform` (sole caller: `api/cmd/tooling/admin/commands/seedPlatform.go`) to bootstrap a fresh customer DB with **platform configuration only** — no demo users, products, orders, or inventory. Requires migrations + `seed.sql` (admin_gopher) to have run first.

Chain (execution order is authoritative — do not reorder):

```
seedTableBuilder(ctx, busDomain, adminID)
    ↓
seedPages(ctx, log, busDomain)
    ↓
seedForms(ctx, log, busDomain, db)
    ↓
seedWorkflow(ctx, log, busDomain, adminID)
    ↓
seedAlerts(ctx, log, busDomain, adminID)
    ↓
seedSettings(ctx, busDomain)                                      → 11 scan-discipline lever rows (pick.productScan locked; 10 overridable by scenarios)
```

`adminID` is hardcoded to `5cf37266-3473-4006-984f-9325122678b7` (admin_gopher from `seed.sql`) — no `seedFoundation` runs in this path.

`seedScenarios` is intentionally omitted — `make seed-platform` is the customer-bootstrap path, not a demo/test seed. Customers load scenarios on demand via the scenarios admin UI / `POST /v1/scenarios/{id}/load`, which writes to the `seedSettings`-seeded base rows via `lever_overrides`.

⚠ **Not idempotent.** `seedSettings` (and other seeders) call `Create` directly with no upsert; running `make seed-platform` twice against the same DB fails on `ErrUniqueEntry`. Operators must wipe the DB before re-running.

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
After initial user seeding, `reporters[0]` is updated with `assigned_zones=["STG-A","STG-B"]`
and `reporters[1]` with `assigned_zones=["STG-C","PCK"]` so picking can fan out to
multiple zone-scoped workers in tests. Other reporters keep the empty default.

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
Seeds: brands, categories, 40 products, costs, physical attributes, metrics, cost histories.
Products use the `productbus.TestSeedProductsHistoricalWithDistribution` helper with an explicit
28×`"none"` + 8×`"lot"` + 4×`"serial"` tracking-type distribution so downstream
inventory/lot-tracking/serial-number seed logic has a stable shape to fan out from.

### seed_labels.go
```go
func seedLabels(ctx context.Context, bus *labelbus.Business, products ProductsSeed) error
```
Seeds 79 rows: 19 location + 20 container + 40 product. Catalog UUIDs are deterministic
via `detUUID("label:" + code)` (UUID v5 over the `deadbeef-…-beefdeadbeef` namespace) so
`make reseed-frontend` produces byte-identical label IDs across runs.

Type breakdown:
- **Location** — RCV-NN, QA-NN, STG-{A|B|C}NN, PCK-NN, PKG-NN, SHP-NN. Empty `payload_json: "{}"`.
- **Container** — TOTE-001..TOTE-020. Empty `payload_json: "{}"`.
- **Product** — `PRD-{SKU}` (one per seeded product). `entity_ref` is the product UUID.
  `payload_json` carries `{"sku":..., "upc":..., "productName":...}` for ZPL rendering.

Note: product `entity_ref` is non-deterministic until `productbus.Business.SeedCreate` lands —
all other label fields are stable across reseeds.

### seed_inventory.go
```go
func seedInventory(ctx context.Context, busDomain BusDomain, foundation FoundationSeed, geoHR GeographyHRSeed, products ProductsSeed) (InventorySeed, error)
```
Seeds: warehouses, zones, locations, inventory items, lot trackings, lot locations, serial numbers,
inspections, transfer orders, transactions, adjustments.

### seed_sales.go
```go
type SalesSeed struct {
    OrderIDs         uuid.UUIDs
    OrderLineItemIDs uuid.UUIDs
}
func seedSales(ctx context.Context, busDomain BusDomain, foundation FoundationSeed, geoHR GeographyHRSeed, products ProductsSeed) (SalesSeed, error)
```
Seeds: customers, order fulfillment statuses, orders, line item fulfillment statuses, order line items.

### seed_tasks.go
```go
type TasksSeed struct {
    PutAwayTasks []putawaytaskbus.PutAwayTask
    PickTasks    []picktaskbus.PickTask
}
func seedTasks(ctx context.Context, busDomain BusDomain, foundation FoundationSeed, products ProductsSeed, inventory InventorySeed, sales SalesSeed) (TasksSeed, error)
```
Seeds: 15 put-away tasks, 15 pick tasks. First put-away task assigned to floor_worker1.

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
func seedForms(ctx context.Context, log *logger.Logger, busDomain BusDomain, db *sqlx.DB) error
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
Must run after forms and templates are seeded.

### seed_alerts.go
```go
func seedAlerts(ctx context.Context, log *logger.Logger, busDomain BusDomain, adminID uuid.UUID) error
```
Seeds: alert records.

### seed_cyclecounts.go
```go
func seedCycleCounts(ctx context.Context, busDomain BusDomain, foundation FoundationSeed, products ProductsSeed, inventory InventorySeed) error
```
Seeds: 3 cycle count sessions, 15 cycle count items (5 per session).

### seed_approvals.go
```go
func seedApprovals(ctx context.Context, busDomain BusDomain, foundation FoundationSeed) error
```
Seeds: 5 automation executions (FK prerequisite) + 5 pending approval requests.
Must run after seedWorkflow — queries rules from DB.

### seed_settings.go
```go
func seedSettings(ctx context.Context, busDomain BusDomain) error
```
Seeds: 11 canonical scan-discipline lever rows from `levers.Defaults`
(`business/domain/config/settingsbus/levers`). Single source of truth
for default lever values; scenarios may override individual keys via
`config.scenario_setting_overrides`. Must run before seedScenarios so
each override has a base row for the settings GET LEFT JOIN to merge
(semantic ordering; no FK enforces it).

⚠ **`pick.productScan` is non-overridable** — listed in `levers.nonOverridableKeys`
per design doc §3.3 invariant 1 (always `"required"`). It is still seeded as one
of the 11 base rows but `levers.IsOverridable` rejects it from any scenario's
`lever_overrides`. Of the 11 seeded keys, 10 are overridable.

### seed_scenarios.go
```go
func seedScenarios(ctx context.Context, busDomain BusDomain) error
func SeedScenariosFromRoot(ctx context.Context, busDomain BusDomain, scenariosDir string) error
```
Seeds: scenarios + scenario_fixtures rows from YAML files under
`deployments/scenarios/`. `seedScenarios` discovers the root via
`findRepoRoot()` (private, walks for go.mod); `SeedScenariosFromRoot`
takes an explicit path so integration tests can point at a temp dir.
Must run last — depends on every preceding seeder for FK references
(products, locations, totes) resolved by `seed_scenarios_refs.go`.

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

## ⚠ Scenario fixture `_ref → _id` resolver

file: business/sdk/dbtest/seed_scenarios_refs.go

state.yaml rows may use stable human-readable codes for three FK types.
The seeder resolves them to UUIDs before writing to
`inventory.scenario_fixtures.payload_json`, so `scenariodb.ApplyFixtures`
stays table-agnostic and `jsonb_populate_record` receives real UUIDs
on every typed column.

| state.yaml key | Resolved via | Output key |
|---|---|---|
| `product_ref` | `productbus.Query{SKU}` | `product_id` |
| `location_ref` | `inventorylocationbus.Query{LocationCodeExact}` | `location_id` |
| `tote_ref` | `labelbus.QueryByCode` | `label_catalog_id` |

Any key ending in `_ref` that is NOT one of the three above is a
fail-hard error — prevents silent mis-seeding when new ref types are
added to YAML before their resolvers land.

`scenario_id` is also auto-injected when absent, so `DeleteScopedRows`
can remove the fixture rows on the next Load.

Further `_ref` types (e.g. `supplier_ref`, `warehouse_ref`, `currency_ref`,
`user_ref`, `purchase_order_status_ref`) are expected to land alongside
the first real scenario fixtures that need them. Add each new resolver
as a field on `refLookups` and a case in `resolveRefs`'s switch.

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
