# seeding

[dbtest]=test infrastructure [seedmodels]=static seed data
→=depends on ⊗=reads from

---

## Overview

`business/sdk/dbtest/` provides two seeding entry points:
- `InsertSeedData` — populates a live Kubernetes/Docker database for `make seed-frontend`
- `NewDatabase` — spins up an isolated test database (random name, 4-letter suffix) for integration tests

Both use the same `BusDomain` struct and the same seed_*.go chain.
Static data (table configs, form field definitions, page action buttons, chart configs) lives in `seedmodels/`.

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

- DB name: random 4 lowercase letters (`abcdefghijklmnopqrstuvwxyz`)
- TimeZone: `SET TIME ZONE 'America/New_York'`
- Runs migrations + seeds before returning
- Each test gets its own isolated database instance

---

## BusDomain [dbtest]

file: business/sdk/dbtest/dbtest.go

Central struct holding every instantiated business package. Passed into all seed functions.

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
Brand, ProductCategory, Product, PhysicalAttribute, ProductCost, Metrics, CostHistory

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
Workflow, Alert
```

---

## Seed Orchestrator [dbtest]

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

## Seed Functions + Result Structs

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
Cross-domain note: uses `busDomain.Workflow.QueryEntityByName` (read-only lookup) — must run after
`seedFoundation` (which sets up `busDomain.Workflow`), but NOT after `seedWorkflow`.
References `seedmodels/tableforms.go`, `seedmodels/forms.go`, `seedmodels/formregistry.go`.

### seed_workflow.go
```go
func seedWorkflow(ctx context.Context, log *logger.Logger, busDomain BusDomain, adminID uuid.UUID) error
```
Seeds: 15 action templates + 8 default automation rules.
Must run LAST — depends on forms and templates already seeded.

---

## seedmodels/ — Static Data

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

## seed_config_pages.go

file: business/sdk/dbtest/seed_config_pages.go

Holds the `allPages` slice used by `seedPages` for navigation page route records.
Renamed from the original `model.go`.

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
