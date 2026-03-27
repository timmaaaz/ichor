# create_put_away_task Workflow Action — Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a `create_put_away_task` workflow action handler that creates a put-away task when triggered by a PO line item receive event, plus a seeded default rule so receiving auto-generates put-away work.

**Architecture:** New `CreatePutAwayTaskHandler` in `workflowactions/inventory/` implementing `workflow.ActionHandler`. Handler is sync (not async), registered under the nil-guard pattern in `RegisterGranularInventoryActions`. Seeded default rule fires on any `on_update` of `purchase_order_line_items` with no trigger conditions — the handler's own delta logic (FieldChanges.OldValue vs NewValue) filters out non-receive updates.

**Tech Stack:** Go 1.23, `putawaytaskbus`, `supplierproductbus`, `purchaseorderbus`, `business/sdk/unitest` + `dbtest.NewDatabase()` for real-DB tests, `business/sdk/workflow.ActionHandler` interface.

**Spec:** `docs/plans/create-put-away-task-action.md`

**Arch docs:** `docs/arch/inventory-ops.md`, `docs/arch/workflow-engine.md`

---

## File Map

| File | Change |
|---|---|
| `business/sdk/workflow/workflowactions/inventory/createputawaytask.go` | **Create** — handler + config + helpers |
| `business/sdk/workflow/workflowactions/inventory/createputawaytask_test.go` | **Create** — validate + execute tests with real DB |
| `business/sdk/workflow/workflowactions/register.go` | **Modify** — add `PutAwayTask` to `BusDependencies`, register in `RegisterGranularInventoryActions` |
| `api/cmd/services/ichor/build/all/all.go` | **Modify** — add `PutAwayTask: putAwayTaskBus` to `inventoryAndProcurementConfig.Buses` |
| `business/sdk/dbtest/seed_workflow.go` | **Modify** — add action template + seeded default rule |
| `docs/workflow/README.md` | **Modify** — add `create_put_away_task` to handler catalog |

---

## Chunk 1: Wiring — register.go and all.go

### Task 1: Add PutAwayTask to BusDependencies

**Files:**
- Modify: `business/sdk/workflow/workflowactions/register.go`

> **Note on chunk scope:** This task only adds `PutAwayTask` to the `BusDependencies` struct. The registration call inside `RegisterGranularInventoryActions` is in **Task 4 Step 4** (Chunk 2), after the handler file exists.

- [ ] **Step 1: Add import for putawaytaskbus**

In `register.go`, add to imports:
```go
"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
```

- [ ] **Step 2: Add PutAwayTask field to BusDependencies struct**

In the `BusDependencies` struct, add after `TransferOrder`:
```go
PutAwayTask *putawaytaskbus.Business
```

The struct block should look like:
```go
type BusDependencies struct {
    // Inventory domain
    InventoryItem        *inventoryitembus.Business
    InventoryLocation    *inventorylocationbus.Business
    InventoryTransaction *inventorytransactionbus.Business
    InventoryAdjustment  *inventoryadjustmentbus.Business
    TransferOrder        *transferorderbus.Business
    PutAwayTask          *putawaytaskbus.Business   // ← add this line
    Product              *productbus.Business
    // ... rest unchanged
```

- [ ] **Step 3: Build to verify no compile errors**

```bash
cd /path/to/ichor && go build ./business/sdk/workflow/workflowactions/...
```
Expected: success (no errors — handler doesn't exist yet but BusDependencies compiles fine)

- [ ] **Step 4: Commit**

```bash
git add business/sdk/workflow/workflowactions/register.go
git commit -m "feat(workflow): add PutAwayTask to BusDependencies"
```

---

### Task 2: Wire PutAwayTask into all.go

**Files:**
- Modify: `api/cmd/services/ichor/build/all/all.go` (around line 481)

- [ ] **Step 1: Add PutAwayTask to inventoryAndProcurementConfig**

> **No new import needed.** `putawaytaskbus` is already imported in `all.go` at line ~220. Do not add a duplicate import.

Find the `inventoryAndProcurementConfig` block (around line 478). Add `PutAwayTask: putAwayTaskBus` to `Buses`:

```go
inventoryAndProcurementConfig := workflowactions.ActionConfig{
    Log: cfg.Log,
    DB:  cfg.DB,
    Buses: workflowactions.BusDependencies{
        InventoryItem:         inventoryItemBus,
        InventoryTransaction:  inventoryTransactionBus,
        InventoryAdjustment:   inventoryAdjustmentBus,
        TransferOrder:         transferOrderBus,
        PutAwayTask:           putAwayTaskBus,     // ← add this line
        SupplierProduct:       supplierProductBus,
        PurchaseOrder:         purchaseOrderBus,
        PurchaseOrderLineItem: purchaseOrderLineItemBus,
        Workflow:              workflowBus,
    },
}
```

Note: `putAwayTaskBus` is already declared earlier in `all.go` (line ~438). No new variable needed.

- [ ] **Step 2: Build to verify**

```bash
go build ./api/cmd/services/ichor/...
```
Expected: success

- [ ] **Step 3: Commit**

```bash
git add api/cmd/services/ichor/build/all/all.go
git commit -m "feat(workflow): wire PutAwayTask bus into action config"
```

---

## Chunk 2: Handler and Tests (TDD)

### Task 3: Write the test file (all tests should fail to compile until handler exists)

**Files:**
- Create: `business/sdk/workflow/workflowactions/inventory/createputawaytask_test.go`

- [ ] **Step 1: Create the test file**

```go
package inventory_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitemstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

func Test_CreatePutAwayTask(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_CreatePutAwayTask")

	sd, err := insertCreatePutAwayTaskSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	sd.Handler = inventory.NewCreatePutAwayTaskHandler(
		log,
		db.BusDomain.PutAwayTask,
		db.BusDomain.SupplierProduct,
		db.BusDomain.PurchaseOrder,
	)

	unitest.Run(t, createPutAwayTaskValidateTests(sd), "validate")
	unitest.Run(t, createPutAwayTaskExecuteTests(db.BusDomain, sd), "execute")
}

// =============================================================================

type createPutAwayTaskSeedData struct {
	unitest.SeedData
	Handler              *inventory.CreatePutAwayTaskHandler
	Products             []productbus.Product
	InventoryLocations   []inventorylocationbus.InventoryLocation
	SupplierProducts     []supplierproductbus.SupplierProduct
	POWithLocation       purchaseorderbus.PurchaseOrder // has DeliveryLocationID set
	POWithoutLocation    purchaseorderbus.PurchaseOrder // DeliveryLocationID = uuid.Nil
	LineItems            []purchaseorderlineitembus.PurchaseOrderLineItem
	ExecutionContext      workflow.ActionExecutionContext
}

func insertCreatePutAwayTaskSeedData(busDomain dbtest.BusDomain) (createPutAwayTaskSeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding user: %w", err)
	}
	adminIDs := make([]uuid.UUID, len(admins))
	for i, a := range admins {
		adminIDs[i] = a.ID
	}

	// Geography chain (required for warehouses and inventory locations)
	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("querying regions: %w", err)
	}
	regionIDs := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		regionIDs = append(regionIDs, r.ID)
	}

	cities, err := citybus.TestSeedCities(ctx, 3, regionIDs, busDomain.City)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding cities: %w", err)
	}
	cityIDs := make([]uuid.UUID, 0, len(cities))
	for _, c := range cities {
		cityIDs = append(cityIDs, c.ID)
	}

	streets, err := streetbus.TestSeedStreets(ctx, 3, cityIDs, busDomain.Street)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding streets: %w", err)
	}
	streetIDs := make([]uuid.UUID, 0, len(streets))
	for _, s := range streets {
		streetIDs = append(streetIDs, s.ID)
	}

	tzs, err := busDomain.Timezone.QueryAll(ctx)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("querying timezones: %w", err)
	}
	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 3, streetIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding contact infos: %w", err)
	}
	contactIDs := make(uuid.UUIDs, len(contactInfos))
	for i, c := range contactInfos {
		contactIDs[i] = c.ID
	}

	// Products (required for put_away_tasks.product_id FK)
	brands, err := brandbus.TestSeedBrands(ctx, 2, contactIDs, busDomain.Brand)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding brands: %w", err)
	}
	brandIDs := make(uuid.UUIDs, len(brands))
	for i, b := range brands {
		brandIDs[i] = b.BrandID
	}

	productCategories, err := productcategorybus.TestSeedProductCategories(ctx, 2, busDomain.ProductCategory)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding product categories: %w", err)
	}
	productCategoryIDs := make(uuid.UUIDs, len(productCategories))
	for i, pc := range productCategories {
		productCategoryIDs[i] = pc.ProductCategoryID
	}

	products, err := productbus.TestSeedProducts(ctx, 5, brandIDs, productCategoryIDs, busDomain.Product)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding products: %w", err)
	}
	productIDs := make(uuid.UUIDs, len(products))
	for i, p := range products {
		productIDs[i] = p.ProductID
	}

	// Inventory locations (required for put_away_tasks.location_id FK)
	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 2, adminIDs[0], streetIDs, busDomain.Warehouse)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding warehouses: %w", err)
	}
	warehouseIDs := make(uuid.UUIDs, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	zones, err := zonebus.TestSeedZone(ctx, 4, warehouseIDs, busDomain.Zones)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding zones: %w", err)
	}
	zoneIDs := make([]uuid.UUID, len(zones))
	for i, z := range zones {
		zoneIDs[i] = z.ZoneID
	}

	inventoryLocations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 5, warehouseIDs, zoneIDs, busDomain.InventoryLocation)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding inventory locations: %w", err)
	}

	// Procurement entities (for PO-based location resolution and supplier_product -> product mapping)
	suppliers, err := supplierbus.TestSeedSuppliers(ctx, 2, contactIDs, busDomain.Supplier)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding suppliers: %w", err)
	}
	supplierIDs := make(uuid.UUIDs, len(suppliers))
	for i, s := range suppliers {
		supplierIDs[i] = s.SupplierID
	}

	supplierProducts, err := supplierproductbus.TestSeedSupplierProducts(ctx, 3, productIDs, supplierIDs, busDomain.SupplierProduct)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding supplier products: %w", err)
	}

	// Purchase order setup — query existing reference data seeded by seed.sql
	poStatuses, err := busDomain.PurchaseOrderStatus.QueryAll(ctx)
	if err != nil || len(poStatuses) == 0 {
		return createPutAwayTaskSeedData{}, fmt.Errorf("querying PO statuses: %w", err)
	}
	poStatusIDs := make(uuid.UUIDs, len(poStatuses))
	for i, s := range poStatuses {
		poStatusIDs[i] = s.ID
	}

	currencies, err := busDomain.Currency.QueryAll(ctx)
	if err != nil || len(currencies) == 0 {
		return createPutAwayTaskSeedData{}, fmt.Errorf("querying currencies: %w", err)
	}
	currencyIDs := make(uuid.UUIDs, len(currencies))
	for i, c := range currencies {
		currencyIDs[i] = c.ID
	}

	// PO with delivery location set — used for happy-path po_delivery test
	poWithLocation, err := busDomain.PurchaseOrder.Create(ctx, purchaseorderbus.NewPurchaseOrder{
		OrderNumber:             "TEST-PO-WITH-LOC-001",
		SupplierID:              supplierIDs[0],
		PurchaseOrderStatusID:   poStatusIDs[0],
		DeliveryWarehouseID:     warehouseIDs[0],
		DeliveryLocationID:      inventoryLocations[0].LocationID, // set — key for po_delivery test
		DeliveryStreetID:        streetIDs[0],
		OrderDate:               time.Now().UTC(),
		ExpectedDeliveryDate:    time.Now().UTC().Add(time.Hour * 24 * 14),
		Subtotal:                1000.00,
		TaxAmount:               80.00,
		ShippingCost:            50.00,
		TotalAmount:             1130.00,
		CurrencyID:              currencyIDs[0],
		RequestedBy:             adminIDs[0],
		Notes:                   "Test PO with delivery location",
		SupplierReferenceNumber: "SUP-TEST-001",
		CreatedBy:               adminIDs[0],
	})
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("creating PO with location: %w", err)
	}

	// PO without delivery location — used for no_location output port test
	poWithoutLocation, err := busDomain.PurchaseOrder.Create(ctx, purchaseorderbus.NewPurchaseOrder{
		OrderNumber:             "TEST-PO-NO-LOC-002",
		SupplierID:              supplierIDs[0],
		PurchaseOrderStatusID:   poStatusIDs[0],
		DeliveryWarehouseID:     warehouseIDs[0],
		DeliveryLocationID:      uuid.Nil, // no location set
		DeliveryStreetID:        streetIDs[0],
		OrderDate:               time.Now().UTC(),
		ExpectedDeliveryDate:    time.Now().UTC().Add(time.Hour * 24 * 14),
		Subtotal:                500.00,
		TaxAmount:               40.00,
		ShippingCost:            25.00,
		TotalAmount:             565.00,
		CurrencyID:              currencyIDs[0],
		RequestedBy:             adminIDs[0],
		Notes:                   "Test PO without delivery location",
		SupplierReferenceNumber: "SUP-TEST-002",
		CreatedBy:               adminIDs[0],
	})
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("creating PO without location: %w", err)
	}

	// PO line items (for completeness — tests read from RawData directly, not from DB)
	liStatuses, err := purchaseorderlineitemstatusbus.TestSeedPurchaseOrderLineItemStatuses(ctx, 1, busDomain.PurchaseOrderLineItemStatus)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding line item statuses: %w", err)
	}
	liStatusIDs := make(uuid.UUIDs, len(liStatuses))
	for i, s := range liStatuses {
		liStatusIDs[i] = s.ID
	}

	spIDs := make(uuid.UUIDs, len(supplierProducts))
	for i, sp := range supplierProducts {
		spIDs[i] = sp.SupplierProductID
	}
	poIDs := uuid.UUIDs{poWithLocation.ID, poWithoutLocation.ID}

	lineItems, err := purchaseorderlineitembus.TestSeedPurchaseOrderLineItems(ctx, 4, poIDs, spIDs, liStatusIDs, adminIDs, busDomain.PurchaseOrderLineItem)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding line items: %w", err)
	}

	ruleID := uuid.New()
	execContext := workflow.ActionExecutionContext{
		EntityID:      uuid.New(),
		EntityName:    "purchase_order_line_items",
		EventType:     "on_update",
		UserID:        adminIDs[0],
		RuleID:        &ruleID,
		RuleName:      "Test Create Put-Away Rule",
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now().UTC(),
		TriggerSource: workflow.TriggerSourceAutomation,
	}

	return createPutAwayTaskSeedData{
		SeedData: unitest.SeedData{
			Admins:             []unitest.User{{User: admins[0]}},
			Products:           products,
			InventoryLocations: inventoryLocations,
		},
		Products:           products,
		InventoryLocations: inventoryLocations,
		SupplierProducts:   supplierProducts,
		POWithLocation:     poWithLocation,
		POWithoutLocation:  poWithoutLocation,
		LineItems:          lineItems,
		ExecutionContext:    execContext,
	}, nil
}

// =============================================================================
// Validate Tests

func createPutAwayTaskValidateTests(sd createPutAwayTaskSeedData) []unitest.Table {
	return []unitest.Table{
		putAwayValidateMissingLocationStrategy(sd),
		putAwayValidateInvalidLocationStrategy(sd),
		putAwayValidateMissingProductIDWhenStatic(sd),
		putAwayValidateInvalidProductID(sd),
		putAwayValidateMissingLocationIDWhenStatic(sd),
		putAwayValidateInvalidLocationID(sd),
		putAwayValidateSourceFromPOPODeliveryValid(sd),
	}
}

func putAwayValidateMissingLocationStrategy(sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "missing_location_strategy",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{"source_from_po":false,"product_id":"` + uuid.New().String() + `"}`)
			err := sd.Handler.Validate(config)
			if err == nil {
				return false
			}
			return strings.Contains(err.Error(), "location_strategy")
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func putAwayValidateInvalidLocationStrategy(sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "invalid_location_strategy",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{"location_strategy":"bad_value","product_id":"` + uuid.New().String() + `"}`)
			err := sd.Handler.Validate(config)
			if err == nil {
				return false
			}
			return strings.Contains(err.Error(), "location_strategy")
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func putAwayValidateMissingProductIDWhenStatic(sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "missing_product_id_when_static",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{"location_strategy":"static","location_id":"` + uuid.New().String() + `"}`)
			err := sd.Handler.Validate(config)
			if err == nil {
				return false
			}
			return strings.Contains(err.Error(), "product_id")
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func putAwayValidateInvalidProductID(sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "invalid_product_id",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{"location_strategy":"static","product_id":"not-a-uuid","location_id":"` + uuid.New().String() + `"}`)
			err := sd.Handler.Validate(config)
			if err == nil {
				return false
			}
			return strings.Contains(err.Error(), "product_id")
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func putAwayValidateMissingLocationIDWhenStatic(sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "missing_location_id_when_static",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{"location_strategy":"static","product_id":"` + uuid.New().String() + `"}`)
			err := sd.Handler.Validate(config)
			if err == nil {
				return false
			}
			return strings.Contains(err.Error(), "location_id")
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func putAwayValidateInvalidLocationID(sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "invalid_location_id",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{"location_strategy":"static","product_id":"` + uuid.New().String() + `","location_id":"not-a-uuid"}`)
			err := sd.Handler.Validate(config)
			if err == nil {
				return false
			}
			return strings.Contains(err.Error(), "location_id")
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func putAwayValidateSourceFromPOPODeliveryValid(sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "source_from_po_po_delivery_valid",
		ExpResp: nil,
		ExcFunc: func(ctx context.Context) any {
			// source_from_po=true with po_delivery requires no product_id or location_id
			config := json.RawMessage(`{"source_from_po":true,"location_strategy":"po_delivery"}`)
			return sd.Handler.Validate(config)
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("expected nil error, got: %v", got)
			}
			return ""
		},
	}
}

// =============================================================================
// Execute Tests

func createPutAwayTaskExecuteTests(busDomain dbtest.BusDomain, sd createPutAwayTaskSeedData) []unitest.Table {
	return []unitest.Table{
		putAwayExecuteHappyPathStatic(busDomain, sd),
		putAwayExecuteSourceFromPODelivery(busDomain, sd),
		putAwayExecuteNoLocationOnPO(sd),
		putAwayExecuteProductNotFound(sd),
		putAwayExecuteZeroDelta(sd),
		putAwayExecuteNegativeDelta(sd),
		putAwayExecuteTemplateReferenceNumber(busDomain, sd),
	}
}

func putAwayExecuteHappyPathStatic(busDomain dbtest.BusDomain, sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "happy_path_static",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			product := sd.Products[1]
			location := sd.InventoryLocations[1]
			delta := 10

			cfg := inventory.CreatePutAwayTaskConfig{
				SourceFromPO:     false,
				ProductID:        product.ProductID.String(),
				LocationStrategy: "static",
				LocationID:       location.LocationID.String(),
				ReferenceNumber:  "STATIC-TEST",
			}
			configJSON, _ := json.Marshal(cfg)

			execCtx := sd.ExecutionContext
			execCtx.ExecutionID = uuid.New()
			execCtx.FieldChanges = map[string]workflow.FieldChange{
				"quantity_received": {
					OldValue: float64(0),
					NewValue: float64(delta),
				},
			}

			result, err := sd.Handler.Execute(ctx, configJSON, execCtx)
			if err != nil {
				return fmt.Errorf("execute failed: %w", err)
			}

			resultMap, ok := result.(map[string]any)
			if !ok {
				return fmt.Errorf("expected map[string]any, got %T", result)
			}

			if resultMap["output"] != "created" {
				return fmt.Errorf("expected output=created, got %v", resultMap["output"])
			}

			taskIDStr, ok := resultMap["task_id"].(string)
			if !ok {
				return fmt.Errorf("task_id missing or wrong type: %v", resultMap["task_id"])
			}
			taskID, err := uuid.Parse(taskIDStr)
			if err != nil {
				return fmt.Errorf("task_id not a valid UUID: %s", taskIDStr)
			}

			// Verify task exists in DB with correct values
			task, err := busDomain.PutAwayTask.QueryByID(ctx, taskID)
			if err != nil {
				return fmt.Errorf("querying created task: %w", err)
			}
			if task.ProductID != product.ProductID {
				return fmt.Errorf("product_id mismatch: expected %s, got %s", product.ProductID, task.ProductID)
			}
			if task.LocationID != location.LocationID {
				return fmt.Errorf("location_id mismatch: expected %s, got %s", location.LocationID, task.LocationID)
			}
			if task.Quantity != delta {
				return fmt.Errorf("quantity mismatch: expected %d, got %d", delta, task.Quantity)
			}

			return true
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func putAwayExecuteSourceFromPODelivery(busDomain dbtest.BusDomain, sd createPutAwayTaskSeedData) unitest.Table {
	if len(sd.SupplierProducts) == 0 {
		return unitest.Table{
			Name: "source_from_po_po_delivery_skip", ExpResp: "skipped",
			ExcFunc: func(ctx context.Context) any { return "skipped" },
			CmpFunc: func(got, exp any) string { return "" },
		}
	}
	return unitest.Table{
		Name:    "source_from_po_po_delivery",
		ExpResp: "created",
		ExcFunc: func(ctx context.Context) any {
			sp := sd.SupplierProducts[0]

			cfg := inventory.CreatePutAwayTaskConfig{
				SourceFromPO:     true,
				LocationStrategy: "po_delivery",
			}
			configJSON, _ := json.Marshal(cfg)

			execCtx := sd.ExecutionContext
			execCtx.ExecutionID = uuid.New()
			execCtx.FieldChanges = map[string]workflow.FieldChange{
				"quantity_received": {OldValue: float64(0), NewValue: float64(15)},
			}
			execCtx.RawData = map[string]any{
				"supplier_product_id": sp.SupplierProductID.String(),
				"purchase_order_id":   sd.POWithLocation.ID.String(),
			}

			result, err := sd.Handler.Execute(ctx, configJSON, execCtx)
			if err != nil {
				return fmt.Errorf("execute failed: %w", err)
			}
			resultMap, ok := result.(map[string]any)
			if !ok {
				return fmt.Errorf("expected map, got %T", result)
			}
			// Verify location matches po.DeliveryLocationID
			if output := resultMap["output"]; output != "created" {
				return fmt.Errorf("expected output=created, got %v", output)
			}
			taskID, _ := uuid.Parse(resultMap["task_id"].(string))
			task, err := busDomain.PutAwayTask.QueryByID(ctx, taskID)
			if err != nil {
				return fmt.Errorf("querying task: %w", err)
			}
			if task.LocationID != sd.POWithLocation.DeliveryLocationID {
				return fmt.Errorf("location_id should match PO.DeliveryLocationID: expected %s, got %s",
					sd.POWithLocation.DeliveryLocationID, task.LocationID)
			}
			if task.ProductID != sp.ProductID {
				return fmt.Errorf("product_id should match supplier product: expected %s, got %s",
					sp.ProductID, task.ProductID)
			}
			return resultMap["output"]
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func putAwayExecuteNoLocationOnPO(sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "no_location_on_po",
		ExpResp: "no_location",
		ExcFunc: func(ctx context.Context) any {
			cfg := inventory.CreatePutAwayTaskConfig{
				SourceFromPO:     true,
				LocationStrategy: "po_delivery",
			}
			configJSON, _ := json.Marshal(cfg)

			execCtx := sd.ExecutionContext
			execCtx.ExecutionID = uuid.New()
			execCtx.FieldChanges = map[string]workflow.FieldChange{
				"quantity_received": {OldValue: float64(0), NewValue: float64(5)},
			}
			execCtx.RawData = map[string]any{
				"supplier_product_id": sd.SupplierProducts[0].SupplierProductID.String(),
				"purchase_order_id":   sd.POWithoutLocation.ID.String(), // no DeliveryLocationID
			}

			result, err := sd.Handler.Execute(ctx, configJSON, execCtx)
			if err != nil {
				return fmt.Errorf("unexpected hard error: %w", err)
			}
			resultMap, ok := result.(map[string]any)
			if !ok {
				return fmt.Errorf("expected map, got %T", result)
			}
			return resultMap["output"]
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func putAwayExecuteProductNotFound(sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "product_not_found",
		ExpResp: "product_not_found",
		ExcFunc: func(ctx context.Context) any {
			cfg := inventory.CreatePutAwayTaskConfig{
				SourceFromPO:     true,
				LocationStrategy: "po_delivery",
			}
			configJSON, _ := json.Marshal(cfg)

			execCtx := sd.ExecutionContext
			execCtx.ExecutionID = uuid.New()
			execCtx.FieldChanges = map[string]workflow.FieldChange{
				"quantity_received": {OldValue: float64(0), NewValue: float64(5)},
			}
			execCtx.RawData = map[string]any{
				"supplier_product_id": uuid.New().String(), // unknown — lookup will fail
				"purchase_order_id":   sd.POWithLocation.ID.String(),
			}

			result, err := sd.Handler.Execute(ctx, configJSON, execCtx)
			if err != nil {
				return fmt.Errorf("unexpected hard error: %w", err)
			}
			resultMap, ok := result.(map[string]any)
			if !ok {
				return fmt.Errorf("expected map, got %T", result)
			}
			return resultMap["output"]
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func putAwayExecuteZeroDelta(sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "zero_delta",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			cfg := inventory.CreatePutAwayTaskConfig{
				SourceFromPO:     false,
				ProductID:        sd.Products[0].ProductID.String(),
				LocationStrategy: "static",
				LocationID:       sd.InventoryLocations[0].LocationID.String(),
			}
			configJSON, _ := json.Marshal(cfg)

			execCtx := sd.ExecutionContext
			execCtx.ExecutionID = uuid.New()
			execCtx.FieldChanges = map[string]workflow.FieldChange{
				"quantity_received": {OldValue: float64(10), NewValue: float64(10)}, // same = delta 0
			}

			result, err := sd.Handler.Execute(ctx, configJSON, execCtx)
			if err != nil {
				return fmt.Errorf("execute failed: %w", err)
			}
			resultMap, ok := result.(map[string]any)
			if !ok {
				return fmt.Errorf("expected map, got %T", result)
			}
			skipped, _ := resultMap["skipped"].(bool)
			return resultMap["output"] == "created" && skipped
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func putAwayExecuteNegativeDelta(sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "negative_delta",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			cfg := inventory.CreatePutAwayTaskConfig{
				SourceFromPO:     false,
				ProductID:        sd.Products[0].ProductID.String(),
				LocationStrategy: "static",
				LocationID:       sd.InventoryLocations[0].LocationID.String(),
			}
			configJSON, _ := json.Marshal(cfg)

			execCtx := sd.ExecutionContext
			execCtx.ExecutionID = uuid.New()
			execCtx.FieldChanges = map[string]workflow.FieldChange{
				"quantity_received": {OldValue: float64(10), NewValue: float64(5)}, // correction: delta = -5
			}

			result, err := sd.Handler.Execute(ctx, configJSON, execCtx)
			if err != nil {
				return fmt.Errorf("execute failed: %w", err)
			}
			resultMap, ok := result.(map[string]any)
			if !ok {
				return fmt.Errorf("expected map, got %T", result)
			}
			skipped, _ := resultMap["skipped"].(bool)
			return resultMap["output"] == "created" && skipped
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func putAwayExecuteTemplateReferenceNumber(busDomain dbtest.BusDomain, sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "template_reference_number",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			cfg := inventory.CreatePutAwayTaskConfig{
				SourceFromPO:     false,
				ProductID:        sd.Products[2].ProductID.String(),
				LocationStrategy: "static",
				LocationID:       sd.InventoryLocations[2].LocationID.String(),
				ReferenceNumber:  "PO-RCV-{{purchase_order_id}}",
			}
			configJSON, _ := json.Marshal(cfg)

			poID := sd.POWithLocation.ID.String()
			execCtx := sd.ExecutionContext
			execCtx.ExecutionID = uuid.New()
			execCtx.FieldChanges = map[string]workflow.FieldChange{
				"quantity_received": {OldValue: float64(0), NewValue: float64(7)},
			}
			execCtx.RawData = map[string]any{
				"purchase_order_id": poID,
			}

			result, err := sd.Handler.Execute(ctx, configJSON, execCtx)
			if err != nil {
				return fmt.Errorf("execute failed: %w", err)
			}
			resultMap, ok := result.(map[string]any)
			if !ok {
				return fmt.Errorf("expected map, got %T", result)
			}
			if resultMap["output"] != "created" {
				return fmt.Errorf("expected output=created, got %v", resultMap["output"])
			}

			taskID, _ := uuid.Parse(resultMap["task_id"].(string))
			task, err := busDomain.PutAwayTask.QueryByID(ctx, taskID)
			if err != nil {
				return fmt.Errorf("querying task: %w", err)
			}

			expected := "PO-RCV-" + poID
			if task.ReferenceNumber != expected {
				return fmt.Errorf("reference_number mismatch: expected %q, got %q", expected, task.ReferenceNumber)
			}
			return true
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}
```

- [ ] **Step 2: Verify test file fails to compile (handler doesn't exist yet)**

```bash
go build ./business/sdk/workflow/workflowactions/inventory/...
```
Expected: compile error referencing `inventory.NewCreatePutAwayTaskHandler` and `inventory.CreatePutAwayTaskConfig`

---

### Task 4: Implement the handler

**Files:**
- Create: `business/sdk/workflow/workflowactions/inventory/createputawaytask.go`

- [ ] **Step 1: Create the handler file**

```go
package inventory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// CreatePutAwayTaskConfig defines the configuration for the create_put_away_task action.
type CreatePutAwayTaskConfig struct {
	// SourceFromPO: when true, resolves product_id from trigger context's supplier_product_id.
	// When false, uses ProductID from config.
	SourceFromPO bool `json:"source_from_po"`

	// ProductID is a static product UUID. Used only when SourceFromPO is false.
	ProductID string `json:"product_id,omitempty"`

	// LocationStrategy determines how the destination location is resolved.
	// "po_delivery" — use the PO's delivery_location_id (requires PO lookup via RawData["purchase_order_id"])
	// "static"      — use LocationID field below
	LocationStrategy string `json:"location_strategy"`

	// LocationID is a static location UUID. Used when LocationStrategy is "static".
	LocationID string `json:"location_id,omitempty"`

	// ReferenceNumber is a template string. Supports {{variable}} substitution from RawData.
	// Example: "PO-RCV-{{purchase_order_id}}"
	// Defaults to "PO-{purchase_order_id}" when empty and purchase_order_id is in RawData.
	ReferenceNumber string `json:"reference_number,omitempty"`
}

// CreatePutAwayTaskHandler creates a put-away task for received inventory.
// It is triggered by on_update events on purchase_order_line_items when quantity_received increases.
//
// Execute returns map[string]any with key "output" (string) and one of:
//   - "created"  — task created; also includes "task_id" string
//   - "created" + "skipped": true — delta <= 0, no task needed
//   - "no_location"      — po_delivery strategy but PO has no delivery_location_id
//   - "product_not_found" — supplier_product_id lookup failed
//   - "failure"           — unexpected error
type CreatePutAwayTaskHandler struct {
	log                *logger.Logger
	putAwayTaskBus     *putawaytaskbus.Business
	supplierProductBus *supplierproductbus.Business
	purchaseOrderBus   *purchaseorderbus.Business
}

// NewCreatePutAwayTaskHandler creates a new CreatePutAwayTaskHandler.
func NewCreatePutAwayTaskHandler(
	log *logger.Logger,
	putAwayTaskBus *putawaytaskbus.Business,
	supplierProductBus *supplierproductbus.Business,
	purchaseOrderBus *purchaseorderbus.Business,
) *CreatePutAwayTaskHandler {
	return &CreatePutAwayTaskHandler{
		log:                log,
		putAwayTaskBus:     putAwayTaskBus,
		supplierProductBus: supplierProductBus,
		purchaseOrderBus:   purchaseOrderBus,
	}
}

// GetType returns the action type identifier.
func (h *CreatePutAwayTaskHandler) GetType() string {
	return "create_put_away_task"
}

// GetDescription returns a human-readable description for discovery APIs.
func (h *CreatePutAwayTaskHandler) GetDescription() string {
	return "Creates a put-away task directing floor workers to shelve received goods"
}

// SupportsManualExecution returns true — this action can be triggered manually.
func (h *CreatePutAwayTaskHandler) SupportsManualExecution() bool {
	return true
}

// IsAsync returns false — this action completes inline.
func (h *CreatePutAwayTaskHandler) IsAsync() bool {
	return false
}

// Validate validates the action configuration.
func (h *CreatePutAwayTaskHandler) Validate(config json.RawMessage) error {
	var cfg CreatePutAwayTaskConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	if cfg.LocationStrategy != "po_delivery" && cfg.LocationStrategy != "static" {
		return fmt.Errorf("location_strategy must be 'po_delivery' or 'static', got %q", cfg.LocationStrategy)
	}

	if !cfg.SourceFromPO {
		if cfg.ProductID == "" {
			return fmt.Errorf("product_id is required when source_from_po is false")
		}
		if _, err := uuid.Parse(cfg.ProductID); err != nil {
			return fmt.Errorf("invalid product_id: %w", err)
		}
	}

	if cfg.LocationStrategy == "static" {
		if cfg.LocationID == "" {
			return fmt.Errorf("location_id is required when location_strategy is 'static'")
		}
		if _, err := uuid.Parse(cfg.LocationID); err != nil {
			return fmt.Errorf("invalid location_id: %w", err)
		}
	}

	return nil
}

// Execute creates a put-away task based on the config and execution context.
// Must return map[string]any with an "output" key for edge routing.
func (h *CreatePutAwayTaskHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
	var cfg CreatePutAwayTaskConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return map[string]any{"output": "failure", "error": err.Error()}, nil
	}

	// Step 1: Compute quantity delta from FieldChanges.
	// JSON deserializes numbers as float64 — use explicit conversion.
	delta := 0
	if fc, ok := execCtx.FieldChanges["quantity_received"]; ok {
		oldF, _ := toFloat64(fc.OldValue)
		newF, _ := toFloat64(fc.NewValue)
		delta = int(newF - oldF)
	} else if raw, ok := execCtx.RawData["quantity_received"]; ok {
		if v, ok2 := toFloat64(raw); ok2 {
			delta = int(v)
			h.log.Warn(ctx, "create_put_away_task: FieldChanges not present; falling back to RawData quantity_received",
				"entity_id", execCtx.EntityID)
		}
	}

	if delta <= 0 {
		h.log.Info(ctx, "create_put_away_task: delta <= 0, skipping task creation",
			"delta", delta, "entity_id", execCtx.EntityID)
		return map[string]any{"output": "created", "skipped": true, "reason": "delta <= 0"}, nil
	}

	// Step 2: Resolve product_id.
	var productID uuid.UUID
	if cfg.SourceFromPO {
		spIDStr, _ := execCtx.RawData["supplier_product_id"].(string)
		spID, err := uuid.Parse(spIDStr)
		if err != nil {
			return map[string]any{"output": "product_not_found", "error": "invalid or missing supplier_product_id in RawData"}, nil
		}
		sp, err := h.supplierProductBus.QueryByID(ctx, spID)
		if err != nil {
			h.log.Info(ctx, "create_put_away_task: supplier product lookup failed",
				"supplier_product_id", spIDStr, "error", err)
			return map[string]any{"output": "product_not_found", "error": err.Error()}, nil
		}
		productID = sp.ProductID
	} else {
		productID, _ = uuid.Parse(cfg.ProductID)
	}

	// Step 3: Resolve location_id.
	var locationID uuid.UUID
	if cfg.LocationStrategy == "po_delivery" {
		poIDStr, _ := execCtx.RawData["purchase_order_id"].(string)
		poID, err := uuid.Parse(poIDStr)
		if err != nil {
			return map[string]any{"output": "no_location", "error": "invalid or missing purchase_order_id in RawData"}, nil
		}
		po, err := h.purchaseOrderBus.QueryByID(ctx, poID)
		if err != nil {
			return map[string]any{"output": "no_location", "error": err.Error()}, nil
		}
		if po.DeliveryLocationID == uuid.Nil {
			h.log.Info(ctx, "create_put_away_task: PO has no delivery_location_id",
				"purchase_order_id", poIDStr)
			return map[string]any{"output": "no_location"}, nil
		}
		locationID = po.DeliveryLocationID
	} else {
		locationID, _ = uuid.Parse(cfg.LocationID)
	}

	// Step 4: Resolve reference number (supports {{variable}} template substitution).
	refNum := cfg.ReferenceNumber
	if refNum == "" {
		if poIDStr, ok := execCtx.RawData["purchase_order_id"].(string); ok {
			refNum = "PO-" + poIDStr
		}
	} else {
		refNum = resolveTemplate(refNum, execCtx.RawData)
	}

	// Step 5: Create the task.
	task, err := h.putAwayTaskBus.Create(ctx, putawaytaskbus.NewPutAwayTask{
		ProductID:       productID,
		LocationID:      locationID,
		Quantity:        delta,
		ReferenceNumber: refNum,
		CreatedBy:       execCtx.UserID,
	})
	if err != nil {
		h.log.Error(ctx, "create_put_away_task: failed to create task", "error", err)
		return map[string]any{"output": "failure", "error": err.Error()}, nil
	}

	h.log.Info(ctx, "create_put_away_task: task created",
		"task_id", task.ID, "product_id", productID, "location_id", locationID, "quantity", delta)

	return map[string]any{
		"output":  "created",
		"task_id": task.ID.String(),
	}, nil
}

// resolveTemplate replaces {{key}} placeholders in tmpl with values from data.
func resolveTemplate(tmpl string, data map[string]any) string {
	result := tmpl
	for k, v := range data {
		result = strings.ReplaceAll(result, "{{"+k+"}}", fmt.Sprintf("%v", v))
	}
	return result
}

// toFloat64 converts any numeric value to float64.
// JSON deserializes numbers as float64, but RawData may contain int types in test code.
func toFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}
```

- [ ] **Step 2: Build the package**

```bash
go build ./business/sdk/workflow/workflowactions/inventory/...
```
Expected: success

- [ ] **Step 3: Run the tests**

```bash
go test ./business/sdk/workflow/workflowactions/inventory/... -run Test_CreatePutAwayTask -v -count=1
```
Expected: all validate and execute subtests PASS

- [ ] **Step 4: Register the handler in register.go**

In `RegisterGranularInventoryActions`, add after the TransferOrder nil-guard block:
```go
if config.Buses.PutAwayTask != nil {
    registry.Register(inventory.NewCreatePutAwayTaskHandler(
        config.Log,
        config.Buses.PutAwayTask,
        config.Buses.SupplierProduct,
        config.Buses.PurchaseOrder,
    ))
}
```

- [ ] **Step 5: Build to verify registration compiles**

```bash
go build ./business/sdk/workflow/workflowactions/...
```
Expected: success

- [ ] **Step 6: Run the full inventory action test suite (regression check)**

```bash
go test ./business/sdk/workflow/workflowactions/inventory/... -v -count=1
```
Expected: all existing tests still pass

- [ ] **Step 7: Commit**

```bash
git add \
  business/sdk/workflow/workflowactions/inventory/createputawaytask.go \
  business/sdk/workflow/workflowactions/inventory/createputawaytask_test.go \
  business/sdk/workflow/workflowactions/register.go
git commit -m "feat(workflow): implement create_put_away_task action handler with tests"
```

---

## Chunk 3: Seed Data and Docs

### Task 5: Add action template to seed_workflow.go

**Files:**
- Modify: `business/sdk/dbtest/seed_workflow.go`

- [ ] **Step 1: Add the action template**

In `seedWorkflow()`, find the block where other action templates are created (around line 59). Add the new template alongside them. Find a good insertion point — after the last procurement-related template (e.g., after the `create_purchase_order` template around line 288):

```go
_, err = busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
    Name:          "Create Put-Away Task",
    Description:   "Creates a put-away task directing floor workers to shelve received goods",
    ActionType:    "create_put_away_task",
    Icon:          "material-symbols:shelves",
    DefaultConfig: json.RawMessage(`{"source_from_po":true,"location_strategy":"po_delivery","reference_number":"PO-RCV-{{purchase_order_id}}"}`),
    CreatedBy:     adminID,
})
if err != nil {
    log.Error(ctx, "Failed to create create_put_away_task template", "error", err)
}
```

Note: no need to capture the return value if the rule seeding is handled separately in a guard block.

- [ ] **Step 2: Add the seeded default rule**

At the end of `seedWorkflow()`, add a new guard block (after the existing default workflow blocks). This block queries the entity and trigger type, then creates the rule:

```go
// --- Default Workflow N: Auto-Create Put-Away on Receive ---
poLineItemsEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "purchase_order_line_items")
if err != nil {
    log.Error(ctx, "Failed to query purchase_order_line_items entity for put-away rule", "error", err)
} else if poLineItemsEntity.ID != uuid.Nil && onUpdateTrigger.ID != uuid.Nil {
    putAwayRule, err := busDomain.Workflow.CreateRule(ctx, workflow.NewAutomationRule{
        Name:          "Auto-Create Put-Away on Receive",
        Description:   "When a PO line item receives quantity, create a put-away task at the PO's delivery location",
        EntityID:      poLineItemsEntity.ID,
        EntityTypeID:  wfEntityType.ID,
        TriggerTypeID: onUpdateTrigger.ID,
        IsActive:      true,
        IsDefault:     true,
        CreatedBy:     adminID,
        // TriggerConditions is nil — no conditions.
        // The handler filters by checking FieldChanges["quantity_received"] delta > 0.
    })
    if err != nil {
        log.Error(ctx, "Failed to create Auto-Create Put-Away rule", "error", err)
    } else {
        putAwayAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
            AutomationRuleID: putAwayRule.ID,
            Name:             "Create Put-Away Task from PO Receive",
            Description:      "Creates a put-away task when a PO line item quantity_received increases",
            ActionConfig:     json.RawMessage(`{"source_from_po":true,"location_strategy":"po_delivery","reference_number":"PO-RCV-{{purchase_order_id}}"}`),
            IsActive:         true,
            // TemplateID is optional — omit or set if the template ID is captured above
        })
        if err != nil {
            log.Error(ctx, "Failed to create put-away action", "error", err)
        } else {
            _, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
                RuleID:         putAwayRule.ID,
                SourceActionID: nil,        // nil = start edge
                TargetActionID: putAwayAction.ID,
                EdgeType:       "start",
                EdgeOrder:      0,
            })
            if err != nil {
                log.Error(ctx, "Failed to create edge for put-away action", "error", err)
            }
            log.Info(ctx, "✅ Created 'Auto-Create Put-Away on Receive' rule")
        }
    }
}
```

**Important:** `onUpdateTrigger` is already queried earlier in the function (around line 858). Make sure your block is placed after that query, not before it.

- [ ] **Step 3: Build to verify**

```bash
go build ./business/sdk/dbtest/...
```
Expected: success

- [ ] **Step 4: Commit**

```bash
git add business/sdk/dbtest/seed_workflow.go
git commit -m "feat(workflow): seed create_put_away_task action template and default rule"
```

---

### Task 6: Update workflow handler catalog

**Files:**
- Modify: `docs/workflow/README.md`

- [ ] **Step 1: Add create_put_away_task to the handler table**

Find the handler catalog table in `docs/workflow/README.md`. Add a row for the new handler alongside the other inventory actions:

```markdown
| `create_put_away_task` | Inventory | Creates a put-away task for received goods. Resolves product from `supplier_product_id` (when `source_from_po: true`) and location from PO's `delivery_location_id` (when `location_strategy: "po_delivery"`). Delta-aware: reads `FieldChanges["quantity_received"]` and skips if delta ≤ 0. Output ports: `created`, `no_location`, `product_not_found`, `failure`. | No |
```

- [ ] **Step 2: Commit**

```bash
git add docs/workflow/README.md
git commit -m "docs(workflow): add create_put_away_task to handler catalog"
```

---

### Task 7: Final build and test verification

- [ ] **Step 1: Build the full service**

```bash
go build ./...
```
Expected: no errors

- [ ] **Step 2: Run the handler tests**

```bash
go test ./business/sdk/workflow/workflowactions/inventory/... -v -count=1
```
Expected: all tests pass including `Test_CreatePutAwayTask`

- [ ] **Step 3: Run the workflowactions package tests**

```bash
go test ./business/sdk/workflow/workflowactions/... -count=1
```
Expected: all pass

- [ ] **Step 4: Run the dbtest seed package (compilation check)**

```bash
go test ./business/sdk/dbtest/... -run TestDoesNotExist -count=1
```
Expected: no compile errors, 0 tests run

---

## Key Implementation Notes

**Delta computation:** `FieldChange.OldValue` and `FieldChange.NewValue` are typed `any`. JSON deserializes all numbers to `float64`, so the `toFloat64()` helper handles this correctly. Test code may pass `int` literals — the helper covers that too.

**No trigger conditions on the seeded rule:** The `TriggerProcessor` does not support a `"changed"` operator — unknown operators return `Matched: false` silently. Using no conditions means the rule fires on any `on_update` of `purchase_order_line_items`, and the handler's delta check filters out non-receive updates.

**PO `DeliveryLocationID` check:** `purchaseorderbus.PurchaseOrder.DeliveryLocationID` is `uuid.UUID` (not a pointer). Check `== uuid.Nil` to detect "no delivery location set".

**FK constraints:** `put_away_tasks` has FKs on `product_id → products.products` and `location_id → inventory.inventory_locations`. Test seeds must include real products and real inventory locations — the test does not need inventory items.

**arch ⚠️ Adding a new ActionHandler:** Per `docs/arch/workflow-engine.md`, also verify:
- `business/sdk/workflow/temporal/activities.go` — `create_put_away_task` is sync, so it uses `Registry` (not `AsyncRegistry`). No changes needed since sync is the default path.
- `api/cmd/services/ichor/build/all/all.go` — covered in Task 2.
