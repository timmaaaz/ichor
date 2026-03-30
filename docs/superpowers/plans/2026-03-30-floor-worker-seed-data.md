# Floor Worker Seed Data Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `make seed-frontend` produce realistic warehouse data so floor_worker1 can run every manual test suite (T-03 through T-14) without creating data manually.

**Architecture:** Modify existing `TestSeed*` functions in `{entity}bus/testutil.go` to produce realistic field values. Wire 4 missing seed functions into `seedFrontend.go`. Integration tests self-heal via `cmp.Diff` against seed structs — only count assertions need updating if `n` changes.

**Tech Stack:** Go 1.23, PostgreSQL 16.4, Ardan Labs service architecture

**Spec:** `docs/superpowers/specs/2026-03-30-floor-worker-seed-data-design.md`

---

## File Map

### PR 1a: Products + Inspections
- Modify: `business/domain/products/productbus/testutil.go`
- Modify: `business/domain/inventory/inspectionbus/testutil.go`

### PR 1b: Locations + Inventory Items
- Modify: `business/domain/inventory/inventorylocationbus/testutil.go`
- Modify: `business/domain/inventory/inventoryitembus/testutil.go`

### PR 1c: Put-Away + Pick Tasks
- Modify: `business/domain/inventory/putawaytaskbus/testutil.go`
- Modify: `business/sdk/dbtest/seedFrontend.go`
- Modify: `business/sdk/dbtest/seed_sales.go` (return SalesSeed struct)
- Create: `business/sdk/dbtest/seed_tasks.go`

### PR 2a: Lots + Serials
- Modify: `business/domain/inventory/lottrackingsbus/testutil.go`
- Modify: `business/domain/inventory/serialnumberbus/testutil.go`

### PR 2b: POs + Transfers
- Modify: `business/domain/procurement/purchaseorderbus/testutil.go`
- Modify: `business/domain/inventory/transferorderbus/testutil.go`

### PR 3a: Adjustments + Transactions
- Modify: `business/domain/inventory/inventoryadjustmentbus/testutil.go`
- Modify: `business/domain/inventory/inventorytransactionbus/testutil.go`
- Modify: `business/sdk/dbtest/seed_inventory.go` (approve some adjustments post-create)

### PR 3b: Cycle Counts + Approvals + Config
- Modify: `business/sdk/dbtest/seedFrontend.go`
- Modify: `business/sdk/dbtest/seed_tasks.go` (add cycle count + approval seeding)
- Create: `business/domain/workflow/approvalrequestbus/testutil.go`
- Modify: `business/sdk/dbtest/dbtest.go` (add ApprovalRequest to BusDomain)

---

## PR 1a: Products + Inspections

### Task 1: Make Products Realistic

**Files:**
- Modify: `business/domain/products/productbus/testutil.go`

- [ ] **Step 1: Update TestNewProducts to use realistic names, UPCs, tracking types**

Replace the body of `TestNewProducts` in `business/domain/products/productbus/testutil.go`:

```go
func TestNewProducts(n int, brandIDs, productCategoryIDs uuid.UUIDs) []NewProduct {
	newProducts := make([]NewProduct, n)

	productNames := []string{
		"Industrial Bearing 6205",
		"Nitrile Gloves Box/100",
		"Hydraulic Filter HF-302",
		"LED Panel Light 60W",
		"Stainless Steel Bolt M10",
		"Thermal Paste Tube 5g",
		"Safety Goggles Clear",
		"Rubber Gasket Set",
		"Wire Spool CAT6 100m",
		"Epoxy Adhesive 2-Part",
	}

	trackingTypes := []string{"none", "lot", "serial"}

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		trackingType := trackingTypes[i%len(trackingTypes)]

		np := NewProduct{
			Name:                 productNames[i%len(productNames)],
			BrandID:              brandIDs[rand.Intn(len(brandIDs))],
			ProductCategoryID:    productCategoryIDs[rand.Intn(len(productCategoryIDs))],
			Description:          fmt.Sprintf("High-quality %s for warehouse operations", productNames[i%len(productNames)]),
			SKU:                  fmt.Sprintf("SKU-%04d", i+1),
			ModelNumber:          fmt.Sprintf("MDL-%04d", i+1),
			UpcCode:              fmt.Sprintf("0123456789%02d", i%100),
			Status:               "active",
			IsActive:             idx%2 == 0,
			IsPerishable:         trackingType == "lot",
			HandlingInstructions: fmt.Sprintf("Handling instructions %d", idx),
			UnitsPerCase:         idx * 5,
			TrackingType:         trackingType,
		}
		newProducts[i] = np
	}

	return newProducts
}
```

- [ ] **Step 2: Update TestNewProductsHistorical with the same changes**

Apply the same name/UPC/tracking changes to `TestNewProductsHistorical`. The only difference is the `CreatedDate` field:

```go
func TestNewProductsHistorical(n int, daysBack int, brandIDs, productCategoryIDs uuid.UUIDs) []NewProduct {
	newProducts := make([]NewProduct, n)
	now := time.Now()

	productNames := []string{
		"Industrial Bearing 6205",
		"Nitrile Gloves Box/100",
		"Hydraulic Filter HF-302",
		"LED Panel Light 60W",
		"Stainless Steel Bolt M10",
		"Thermal Paste Tube 5g",
		"Safety Goggles Clear",
		"Rubber Gasket Set",
		"Wire Spool CAT6 100m",
		"Epoxy Adhesive 2-Part",
	}

	trackingTypes := []string{"none", "lot", "serial"}

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		daysAgo := (i * daysBack) / n
		createdDate := now.AddDate(0, 0, -daysAgo)

		trackingType := trackingTypes[i%len(trackingTypes)]

		np := NewProduct{
			Name:                 productNames[i%len(productNames)],
			BrandID:              brandIDs[rand.Intn(len(brandIDs))],
			ProductCategoryID:    productCategoryIDs[rand.Intn(len(productCategoryIDs))],
			Description:          fmt.Sprintf("High-quality %s for warehouse operations", productNames[i%len(productNames)]),
			SKU:                  fmt.Sprintf("SKU-%04d", i+1),
			ModelNumber:          fmt.Sprintf("MDL-%04d", i+1),
			UpcCode:              fmt.Sprintf("0123456789%02d", i%100),
			Status:               "active",
			IsActive:             idx%2 == 0,
			IsPerishable:         trackingType == "lot",
			HandlingInstructions: fmt.Sprintf("Handling instructions %d", idx),
			UnitsPerCase:         idx * 5,
			TrackingType:         trackingType,
			CreatedDate:          &createdDate,
		}
		newProducts[i] = np
	}

	return newProducts
}
```

- [ ] **Step 3: Build to verify compilation**

Run: `go build ./business/domain/products/productbus/...`
Expected: Clean compilation, no errors.

- [ ] **Step 4: Run product integration tests**

Run: `go test ./business/domain/products/productbus/... -count=1 -v`
Expected: All tests pass. The `cmp.Diff` assertions compare against `sd.Products[i]` which now contains the new realistic values — self-healing.

- [ ] **Step 5: Run downstream API tests that consume product seeds**

Run: `go test ./api/cmd/services/ichor/tests/products/productapi/... -count=1 -v`
Expected: All tests pass. If any fail on count assertions, update the hardcoded `Total:` value to match.

- [ ] **Step 6: Commit**

```bash
git add business/domain/products/productbus/testutil.go
git commit -m "feat(seed): make products realistic — names, UPCs, tracking types"
```

### Task 2: Update Inspections for Floor Worker

**Files:**
- Modify: `business/domain/inventory/inspectionbus/testutil.go`

- [ ] **Step 1: Update TestNewInspections to assign floor_worker1 and set recent dates**

Replace the body of `TestNewInspections`:

```go
func TestNewInspections(n int, productIDs, inspectorIDs, lotIDs uuid.UUIDs) []NewInspection {
	newInspections := make([]NewInspection, n)

	// floor_worker1 UUID — stable across all environments (from seed.sql)
	floorWorker1 := uuid.MustParse("c0000000-0000-4000-8000-000000000001")

	idx := rand.Intn(10000)

	for i := 0; i < n; i++ {
		idx++

		// First 5 inspections assigned to floor_worker1
		inspectorID := inspectorIDs[idx%len(inspectorIDs)]
		if i < 5 {
			inspectorID = floorWorker1
		}

		newInspections[i] = NewInspection{
			ProductID:          productIDs[idx%len(productIDs)],
			InspectorID:        inspectorID,
			LotID:              lotIDs[idx%len(lotIDs)],
			InspectionDate:     time.Now().AddDate(0, 0, -(i % 7)),
			Status:             "pending",
			NextInspectionDate: time.Now().AddDate(0, 0, i+7),
		}
	}

	return newInspections
}
```

- [ ] **Step 2: Build to verify compilation**

Run: `go build ./business/domain/inventory/inspectionbus/...`
Expected: Clean compilation.

- [ ] **Step 3: Run inspection integration tests**

Run: `go test ./business/domain/inventory/inspectionbus/... -count=1 -v`
Expected: All tests pass.

Run: `go test ./api/cmd/services/ichor/tests/inventory/inspectionapi/... -count=1 -v`
Expected: All tests pass.

- [ ] **Step 4: Commit**

```bash
git add business/domain/inventory/inspectionbus/testutil.go
git commit -m "feat(seed): assign inspections to floor_worker1, use recent dates"
```

### Task 3: PR 1a — Build, Full Test, Push

- [ ] **Step 1: Full build check**

Run: `go build ./...`
Expected: Clean compilation.

- [ ] **Step 2: Run all product and inspection domain tests**

Run: `go test ./business/domain/products/... ./business/domain/inventory/inspectionbus/... -count=1`
Expected: All pass.

- [ ] **Step 3: Run all API tests that consume product/inspection seeds**

Run: `go test ./api/cmd/services/ichor/tests/products/... ./api/cmd/services/ichor/tests/inventory/inspectionapi/... -count=1`
Expected: All pass. Fix any count assertion mismatches.

- [ ] **Step 4: Create PR**

```bash
git push origin HEAD
gh pr create --title "feat(seed): PR 1a — realistic products + floor_worker1 inspections" --body "..."
```

---

## PR 1b: Locations + Inventory Items

### Task 4: Make Locations Warehouse-Realistic

**Files:**
- Modify: `business/domain/inventory/inventorylocationbus/testutil.go`

- [ ] **Step 1: Update TestNewInventoryLocation with warehouse codes**

Replace the body of `TestNewInventoryLocation`:

```go
func TestNewInventoryLocation(n int, warehouseIDs, zoneIDs []uuid.UUID) []NewInventoryLocation {
	newInventoryLocations := make([]NewInventoryLocation, n)

	aisles := []string{"A", "B", "C"}

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		aisle := aisles[i%len(aisles)]
		rack := fmt.Sprintf("%02d", (i/12)%5+1)
		shelf := fmt.Sprintf("%02d", (i/3)%4+1)
		bin := fmt.Sprintf("%02d", i%6+1)
		locationCode := fmt.Sprintf("%s-%s-%s-%s", aisle, rack, shelf, bin)

		// Alternate: even indices are pick locations, odd are reserve
		isPickLocation := i%2 == 0

		newInventoryLocations[i] = NewInventoryLocation{
			WarehouseID:        warehouseIDs[idx%len(warehouseIDs)],
			ZoneID:             zoneIDs[idx%len(zoneIDs)],
			Aisle:              aisle,
			Rack:               rack,
			Shelf:              shelf,
			Bin:                bin,
			LocationCode:       &locationCode,
			IsPickLocation:     isPickLocation,
			IsReserveLocation:  !isPickLocation,
			MaxCapacity:        idx%100 + 10,
			CurrentUtilization: types.RoundedFloat{Value: float64(idx % 100)},
		}
	}

	return newInventoryLocations
}
```

- [ ] **Step 2: Build and test**

Run: `go build ./business/domain/inventory/inventorylocationbus/...`
Expected: Clean compilation.

Run: `go test ./business/domain/inventory/inventorylocationbus/... -count=1 -v`
Expected: All tests pass.

Run: `go test ./api/cmd/services/ichor/tests/inventory/inventorylocationapi/... -count=1 -v`
Expected: All tests pass. The `"TESTLOC-001"` hardcode is set via a separate `.Update()` — unaffected.

- [ ] **Step 3: Commit**

```bash
git add business/domain/inventory/inventorylocationbus/testutil.go
git commit -m "feat(seed): warehouse-realistic location codes (A-01-02-03 format)"
```

### Task 5: Ensure Inventory Items Have Adequate Quantities

**Files:**
- Modify: `business/domain/inventory/inventoryitembus/testutil.go`

- [ ] **Step 1: Verify current TestNewInventoryProducts already has good quantities**

Read `business/domain/inventory/inventoryitembus/testutil.go`. The current code already sets:
```go
Quantity: 100 + i,
```
This gives quantities of 100-129 for 30 items. This meets the >=50 requirement and the 100+ floor.

The grid pairing `locationIDs[i%nL]` with `productIDs[(i/nL)%nP]` already guarantees unique product-location pairs and will naturally create multi-product locations when `n > len(locationIDs)`.

If the quantities or distribution already meet requirements, no code changes are needed — just verify and document.

- [ ] **Step 2: Build and test**

Run: `go test ./business/domain/inventory/inventoryitembus/... -count=1 -v`
Expected: All tests pass.

Run: `go test ./api/cmd/services/ichor/tests/inventory/inventoryitemapi/... -count=1 -v`
Expected: All tests pass.

- [ ] **Step 3: Commit (if changes were made)**

```bash
git add business/domain/inventory/inventoryitembus/testutil.go
git commit -m "feat(seed): ensure inventory items have adequate quantities for testing"
```

### Task 6: PR 1b — Build, Full Test, Push

- [ ] **Step 1: Full build and test**

Run: `go build ./...`
Run: `go test ./business/domain/inventory/inventorylocationbus/... ./business/domain/inventory/inventoryitembus/... -count=1`
Run: `go test ./api/cmd/services/ichor/tests/inventory/inventorylocationapi/... ./api/cmd/services/ichor/tests/inventory/inventoryitemapi/... -count=1`
Expected: All pass.

- [ ] **Step 2: Create PR**

```bash
git push origin HEAD
gh pr create --title "feat(seed): PR 1b — warehouse-realistic locations + inventory quantities" --body "..."
```

---

## PR 1c: Put-Away + Pick Tasks

### Task 7: Modify seedSales to Return SalesSeed

**Files:**
- Modify: `business/sdk/dbtest/seed_sales.go`

Pick tasks need sales order IDs and line item IDs. Currently `seedSales` returns only `error`. We need it to return a struct.

- [ ] **Step 1: Add SalesSeed struct and update seedSales return type**

In `business/sdk/dbtest/seed_sales.go`, add the struct and change the function signature:

```go
// SalesSeed holds the results of seeding sales data.
type SalesSeed struct {
	OrderIDs         uuid.UUIDs
	OrderLineItemIDs uuid.UUIDs
}
```

Change function signature from:
```go
func seedSales(ctx context.Context, busDomain BusDomain, foundation FoundationSeed, geoHR GeographyHRSeed, products ProductsSeed) error {
```
to:
```go
func seedSales(ctx context.Context, busDomain BusDomain, foundation FoundationSeed, geoHR GeographyHRSeed, products ProductsSeed) (SalesSeed, error) {
```

Update all `return fmt.Errorf(...)` to `return SalesSeed{}, fmt.Errorf(...)`.

Capture line item IDs from `TestSeedOrderLineItemsHistorical`:
```go
	lineItems, err := orderlineitemsbus.TestSeedOrderLineItemsHistorical(ctx, count, orderDates, orderIDs, productIDs, olStatusIDs, userIDs, busDomain.OrderLineItem)
	if err != nil {
		return SalesSeed{}, fmt.Errorf("seeding Order Line Items: %w", err)
	}

	lineItemIDs := make(uuid.UUIDs, len(lineItems))
	for i, li := range lineItems {
		lineItemIDs[i] = li.ID
	}

	return SalesSeed{
		OrderIDs:         orderIDs,
		OrderLineItemIDs: lineItemIDs,
	}, nil
```

Note: Check the return type of `TestSeedOrderLineItemsHistorical` — if it returns `([]OrderLineItem, error)`, capture the first return. If it returns `(int, error)` or `error`, you'll need to query for line item IDs after seeding, or adjust.

- [ ] **Step 2: Update seedFrontend.go to capture SalesSeed**

In `business/sdk/dbtest/seedFrontend.go`, change:
```go
	if err := seedSales(ctx, busDomain, foundation, geoHR, products); err != nil {
		return fmt.Errorf("seeding sales: %w", err)
	}
```
to:
```go
	sales, err := seedSales(ctx, busDomain, foundation, geoHR, products)
	if err != nil {
		return fmt.Errorf("seeding sales: %w", err)
	}
```

- [ ] **Step 3: Build to verify**

Run: `go build ./business/sdk/dbtest/...`
Expected: Clean compilation.

- [ ] **Step 4: Commit**

```bash
git add business/sdk/dbtest/seed_sales.go business/sdk/dbtest/seedFrontend.go
git commit -m "refactor(seed): make seedSales return SalesSeed with order/line item IDs"
```

### Task 8: Create seed_tasks.go and Wire Put-Away + Pick Tasks

**Files:**
- Create: `business/sdk/dbtest/seed_tasks.go`
- Modify: `business/sdk/dbtest/seedFrontend.go`
- Modify: `business/domain/inventory/putawaytaskbus/testutil.go` (assign floor_worker1)

- [ ] **Step 1: Update putawaytaskbus TestNewPutAwayTasks to assign floor_worker1**

In `business/domain/inventory/putawaytaskbus/testutil.go`, update to assign first task to floor_worker1:

```go
func TestNewPutAwayTasks(n int, productIDs, locationIDs, createdByIDs []uuid.UUID) []NewPutAwayTask {
	tasks := make([]NewPutAwayTask, n)

	// floor_worker1 UUID — stable across all environments (from seed.sql)
	floorWorker1 := uuid.MustParse("c0000000-0000-4000-8000-000000000001")

	for i := range n {
		createdBy := createdByIDs[i%len(createdByIDs)]
		if i == 0 {
			createdBy = floorWorker1
		}

		tasks[i] = NewPutAwayTask{
			ProductID:       productIDs[i%len(productIDs)],
			LocationID:      locationIDs[i%len(locationIDs)],
			Quantity:        (i + 1) * 10,
			ReferenceNumber: fmt.Sprintf("PO-HIST-%d", i+1),
			CreatedBy:       createdBy,
		}
	}

	return tasks
}
```

Add the `fmt` import if not already present.

- [ ] **Step 2: Create seed_tasks.go**

Create `business/sdk/dbtest/seed_tasks.go`:

```go
package dbtest

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
)

func seedTasks(ctx context.Context, busDomain BusDomain, foundation FoundationSeed, products ProductsSeed, inventory InventorySeed, sales SalesSeed) error {
	productIDs := make(uuid.UUIDs, len(products.Products))
	for i, p := range products.Products {
		productIDs[i] = p.ProductID
	}

	locationIDs := make(uuid.UUIDs, len(inventory.InventoryLocations))
	for i, loc := range inventory.InventoryLocations {
		locationIDs[i] = loc.LocationID
	}

	adminIDs := make(uuid.UUIDs, len(foundation.Admins))
	for i, a := range foundation.Admins {
		adminIDs[i] = a.ID
	}

	// Seed 15 put-away tasks for frontend
	_, err := putawaytaskbus.TestSeedPutAwayTasks(ctx, 15, productIDs, locationIDs, adminIDs, busDomain.PutAwayTask)
	if err != nil {
		return fmt.Errorf("seeding put-away tasks: %w", err)
	}

	// Seed 15 pick tasks for frontend
	_, err = picktaskbus.TestSeedPickTasks(ctx, 15, sales.OrderIDs, sales.OrderLineItemIDs, productIDs, locationIDs, adminIDs, busDomain.PickTask)
	if err != nil {
		return fmt.Errorf("seeding pick tasks: %w", err)
	}

	return nil
}
```

- [ ] **Step 3: Wire seedTasks into seedFrontend.go**

In `business/sdk/dbtest/seedFrontend.go`, add the call after `seedProcurement`:

```go
	if err := seedProcurement(ctx, busDomain, foundation, geoHR, products, inventory); err != nil {
		return fmt.Errorf("seeding procurement: %w", err)
	}

	if err := seedTasks(ctx, busDomain, foundation, products, inventory, sales); err != nil {
		return fmt.Errorf("seeding tasks: %w", err)
	}
```

Make sure the `sales` variable from the earlier `seedSales` call is available at this point.

- [ ] **Step 4: Build and test**

Run: `go build ./business/sdk/dbtest/...`
Expected: Clean compilation.

Run: `go test ./api/cmd/services/ichor/tests/inventory/putawaytaskapi/... -count=1 -v`
Run: `go test ./api/cmd/services/ichor/tests/inventory/picktaskapi/... -count=1 -v`
Expected: All pass.

- [ ] **Step 5: Commit**

```bash
git add business/domain/inventory/putawaytaskbus/testutil.go business/sdk/dbtest/seed_tasks.go business/sdk/dbtest/seedFrontend.go
git commit -m "feat(seed): wire put-away and pick tasks into seedFrontend"
```

### Task 9: PR 1c — Build, Full Test, Push

- [ ] **Step 1: Full build and targeted tests**

Run: `go build ./...`
Run: `go test ./business/sdk/dbtest/... -count=1`
Run: `go test ./api/cmd/services/ichor/tests/inventory/putawaytaskapi/... ./api/cmd/services/ichor/tests/inventory/picktaskapi/... -count=1`
Expected: All pass.

- [ ] **Step 2: Create PR**

```bash
git push origin HEAD
gh pr create --title "feat(seed): PR 1c — wire put-away and pick tasks into seedFrontend" --body "..."
```

---

## PR 2a: Lots + Serials

### Task 10: Make Lot Trackings Realistic

**Files:**
- Modify: `business/domain/inventory/lottrackingsbus/testutil.go`

- [ ] **Step 1: Update TestNewLotTrackings with human-readable lot numbers and relative dates**

Replace the body of `TestNewLotTrackings`:

```go
func TestNewLotTrackings(n int, supplierProductIDs uuid.UUIDs) []NewLotTrackings {
	newLotTrackingss := make([]NewLotTrackings, n)
	now := time.Now()

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		// Expiration dates distributed for dashboard buckets:
		// idx 0: urgent (within 7 days), idx 1: warning (within 30 days),
		// idx 2: monitor (within 90 days), rest: 30-180 days out
		var expirationDate time.Time
		switch i {
		case 0:
			expirationDate = now.AddDate(0, 0, 5) // urgent
		case 1:
			expirationDate = now.AddDate(0, 0, 20) // warning
		case 2:
			expirationDate = now.AddDate(0, 0, 60) // monitor
		default:
			expirationDate = now.AddDate(0, 0, 30+rand.Intn(150))
		}

		// First 3 are "good" (for dashboard visibility), then cycle to ensure quarantined
		qualityStatuses := []string{"good", "good", "good", "quarantined", "on_hold", "released", "expired"}
		qualityStatus := qualityStatuses[i%len(qualityStatuses)]

		newLotTrackingss[i] = NewLotTrackings{
			SupplierProductID: supplierProductIDs[i%len(supplierProductIDs)],
			LotNumber:         fmt.Sprintf("LOT-2026-%03d", i+1),
			ManufactureDate:   now.AddDate(0, -3, 0),
			ExpirationDate:    expirationDate,
			ReceivedDate:      now.AddDate(0, 0, -rand.Intn(30)),
			QualityStatus:     qualityStatus,
			Quantity:           rand.Intn(1000),
		}
	}

	return newLotTrackingss
}
```

- [ ] **Step 2: Remove the RandomDate helper function**

Delete the `RandomDate()` function from the file — it's no longer used.

- [ ] **Step 3: Build and test**

Run: `go build ./business/domain/inventory/lottrackingsbus/...`
Expected: Clean compilation. If `RandomDate` is used elsewhere, keep it — check with grep first.

Run: `go test ./business/domain/inventory/lottrackingsbus/... -count=1 -v`
Run: `go test ./api/cmd/services/ichor/tests/inventory/lottrackingsapi/... -count=1 -v`
Expected: All pass.

- [ ] **Step 4: Commit**

```bash
git add business/domain/inventory/lottrackingsbus/testutil.go
git commit -m "feat(seed): human-readable lot numbers, relative expiry dates for dashboard"
```

### Task 11: Make Serial Numbers Realistic

**Files:**
- Modify: `business/domain/inventory/serialnumberbus/testutil.go`

- [ ] **Step 1: Update TestNewSerialNumbers with real statuses**

Replace the body of `TestNewSerialNumbers`:

```go
func TestNewSerialNumbers(n int, lotIDs, productIDs, locationIDs []uuid.UUID) []NewSerialNumber {
	newSerialNumbers := make([]NewSerialNumber, n)

	statuses := []string{"available", "reserved", "quarantined", "shipped"}

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++
		newSerialNumbers[i] = NewSerialNumber{
			LotID:        lotIDs[idx%len(lotIDs)],
			ProductID:    productIDs[idx%len(productIDs)],
			LocationID:   locationIDs[idx%len(locationIDs)],
			SerialNumber: fmt.Sprintf("SN-2026-%04d", i+1),
			Status:       statuses[i%len(statuses)],
		}
	}

	return newSerialNumbers
}
```

- [ ] **Step 2: Build and test**

Run: `go build ./business/domain/inventory/serialnumberbus/...`
Run: `go test ./business/domain/inventory/serialnumberbus/... -count=1 -v`
Run: `go test ./api/cmd/services/ichor/tests/inventory/serialnumberapi/... -count=1 -v`
Expected: All pass.

- [ ] **Step 3: Commit**

```bash
git add business/domain/inventory/serialnumberbus/testutil.go
git commit -m "feat(seed): human-readable serial numbers with real statuses"
```

### Task 12: PR 2a — Build, Full Test, Push

- [ ] **Step 1: Full build and test**

Run: `go build ./...`
Run: `go test ./business/domain/inventory/lottrackingsbus/... ./business/domain/inventory/serialnumberbus/... -count=1`
Run: `go test ./api/cmd/services/ichor/tests/inventory/lottrackingsapi/... ./api/cmd/services/ichor/tests/inventory/serialnumberapi/... ./api/cmd/services/ichor/tests/inventory/scanapi/... -count=1`
Expected: All pass.

- [ ] **Step 2: Create PR**

```bash
git push origin HEAD
gh pr create --title "feat(seed): PR 2a — realistic lots + serial numbers" --body "..."
```

---

## PR 2b: POs + Transfers

### Task 13: Make Purchase Orders Have Date Spread

**Files:**
- Modify: `business/domain/procurement/purchaseorderbus/testutil.go`

- [ ] **Step 1: Update TestNewPurchaseOrdersHistorical with varied delivery dates**

Replace the date logic in the for loop of `TestNewPurchaseOrdersHistorical`:

```go
		daysAgo := (i * daysBack) / n
		orderDate := now.AddDate(0, 0, -daysAgo)

		// Varied delivery dates for dashboard windows:
		// i=0: delivery today, i=1: delivery in 3 days (7-day window),
		// i=2: overdue (past, no actual delivery), rest: orderDate + 14 days
		var expectedDelivery time.Time
		switch i {
		case 0:
			expectedDelivery = now // Today window
		case 1:
			expectedDelivery = now.AddDate(0, 0, 3) // 7 Days window
		case 2:
			expectedDelivery = now.AddDate(0, 0, -7) // Overdue (past)
		default:
			expectedDelivery = orderDate.AddDate(0, 0, 14) // Standard 2 weeks
		}
```

Also ensure status cycling includes active statuses. The `statusIDs` parameter cycles through PO status IDs — verify the order in `seed_procurement.go` to ensure at least 1 lands on SENT or APPROVED.

- [ ] **Step 2: Build and test**

Run: `go build ./business/domain/procurement/purchaseorderbus/...`
Run: `go test ./business/domain/procurement/purchaseorderbus/... -count=1 -v`
Run: `go test ./api/cmd/services/ichor/tests/procurement/purchaseorderapi/... -count=1 -v`
Expected: All pass. Audit `purchaseorderapi/query_test.go` carefully — it has 8+ Total assertions. If any filter by date range, counts may shift.

- [ ] **Step 3: Commit**

```bash
git add business/domain/procurement/purchaseorderbus/testutil.go
git commit -m "feat(seed): PO date spread — today, overdue, 7-day window"
```

### Task 14: Make Transfer Orders Have Mixed Statuses

**Files:**
- Modify: `business/domain/inventory/transferorderbus/testutil.go`

- [ ] **Step 1: Update TestNewTransferOrders with status distribution**

In `TestNewTransferOrders`, replace the status line. The `Create` method accepts status directly from `nto.Status`:

```go
	// Status distribution: ~40% pending, ~40% approved, ~20% completed
	transferStatuses := []string{StatusPending, StatusPending, StatusApproved, StatusApproved, StatusCompleted}
```

Then in the loop, replace `Status: StatusPending` with:
```go
		Status:         transferStatuses[i%len(transferStatuses)],
```

- [ ] **Step 2: Build and test**

Run: `go build ./business/domain/inventory/transferorderbus/...`
Run: `go test ./business/domain/inventory/transferorderbus/... -count=1 -v`
Run: `go test ./api/cmd/services/ichor/tests/inventory/transferorderapi/... -count=1 -v`
Expected: All pass. The `Total: 10` assertion is unfiltered — safe.

- [ ] **Step 3: Commit**

```bash
git add business/domain/inventory/transferorderbus/testutil.go
git commit -m "feat(seed): mixed transfer order statuses (pending/approved/completed)"
```

### Task 15: PR 2b — Build, Full Test, Push

- [ ] **Step 1: Full build and test**

Run: `go build ./...`
Run: `go test ./business/domain/procurement/purchaseorderbus/... ./business/domain/inventory/transferorderbus/... -count=1`
Run: `go test ./api/cmd/services/ichor/tests/procurement/purchaseorderapi/... ./api/cmd/services/ichor/tests/inventory/transferorderapi/... -count=1`
Expected: All pass.

- [ ] **Step 2: Create PR**

```bash
git push origin HEAD
gh pr create --title "feat(seed): PR 2b — PO date spread + mixed transfer statuses" --body "..."
```

---

## PR 3a: Adjustments + Transactions

### Task 16: Make Adjustments Have Varied Reasons and Approve Some

**Files:**
- Modify: `business/domain/inventory/inventoryadjustmentbus/testutil.go`
- Modify: `business/sdk/dbtest/seed_inventory.go`

- [ ] **Step 1: Update TestNewInventoryAdjustment with varied reason codes and floor_worker1**

Replace the body of `TestNewInventoryAdjustment`:

```go
func TestNewInventoryAdjustment(n int, productIDs, locationIDs, adjustedByIDs uuid.UUIDs) []NewInventoryAdjustment {
	newInventoryAdjustments := make([]NewInventoryAdjustment, n)

	// floor_worker1 UUID — stable across all environments
	floorWorker1 := uuid.MustParse("c0000000-0000-4000-8000-000000000001")

	reasonCodes := []string{
		ReasonCodeDamaged,
		ReasonCodeTheft,
		ReasonCodeDataEntryError,
		ReasonCodeFoundStock,
		ReasonCodePickingError,
		ReasonCodeReceivingError,
		ReasonCodeCycleCount,
		ReasonCodeOther,
	}

	idx := rand.Intn(10000)

	for i := 0; i < n; i++ {
		approvedByID := adjustedByIDs[(idx+1)%len(adjustedByIDs)]

		// First 2 adjustments created by floor_worker1
		adjustedBy := adjustedByIDs[idx%len(adjustedByIDs)]
		if i < 2 {
			adjustedBy = floorWorker1
		}

		newInventoryAdjustments[i] = NewInventoryAdjustment{
			ProductID:      productIDs[idx%len(productIDs)],
			LocationID:     locationIDs[idx%len(locationIDs)],
			AdjustedBy:     adjustedBy,
			ApprovedBy:     &approvedByID,
			QuantityChange: rand.Intn(100) - 50,
			ReasonCode:     reasonCodes[i%len(reasonCodes)],
			Notes:          fmt.Sprintf("Adjustment for %s", reasonCodes[i%len(reasonCodes)]),
			AdjustmentDate: time.Now().AddDate(0, 0, -(i % 30)),
		}
	}

	return newInventoryAdjustments
}
```

Add `fmt` import if not already present.

- [ ] **Step 2: Approve some adjustments in seed_inventory.go**

In `business/sdk/dbtest/seed_inventory.go`, change the existing `_, err =` to capture the return value, then approve the first adjustment. The `Create` forces `ApprovalStatusPending`, so we need a post-create approve.

Change:
```go
	_, err = inventoryadjustmentbus.TestSeedInventoryAdjustments(ctx, 20, productIDs, inventoryLocationsIDs, reporterIDs[:2], busDomain.InventoryAdjustment)
```
to:
```go
	adjustments, err := inventoryadjustmentbus.TestSeedInventoryAdjustments(ctx, 20, productIDs, inventoryLocationsIDs, reporterIDs[:2], busDomain.InventoryAdjustment)
```

Then add after it:
```go
	// Approve the first adjustment for varied status display
	if len(adjustments) > 0 {
		_, err = busDomain.InventoryAdjustment.Approve(ctx, adjustments[0], foundation.Admins[0].ID, "Approved during seeding")
		if err != nil {
			return InventorySeed{}, fmt.Errorf("approving adjustment: %w", err)
		}
	}
```

The `foundation` parameter is available — check the signature: `func seedInventory(ctx context.Context, busDomain BusDomain, foundation FoundationSeed, geoHR GeographyHRSeed, products ProductsSeed) (InventorySeed, error)`.

- [ ] **Step 3: Build and test**

Run: `go build ./business/domain/inventory/inventoryadjustmentbus/... ./business/sdk/dbtest/...`
Run: `go test ./business/domain/inventory/inventoryadjustmentbus/... -count=1 -v`
Run: `go test ./api/cmd/services/ichor/tests/inventory/inventoryadjustmentapi/... -count=1 -v`
Expected: All pass.

- [ ] **Step 4: Commit**

```bash
git add business/domain/inventory/inventoryadjustmentbus/testutil.go business/sdk/dbtest/seed_inventory.go
git commit -m "feat(seed): varied adjustment reasons, floor_worker1 creator, approve one"
```

### Task 17: Make Transactions Have Varied Types

**Files:**
- Modify: `business/domain/inventory/inventorytransactionbus/testutil.go`

- [ ] **Step 1: Update TestNewInventoryTransaction with varied types and floor_worker1**

Replace the body of `TestNewInventoryTransaction`:

```go
func TestNewInventoryTransaction(n int, locationIDs, productIDs, userIDs uuid.UUIDs) []NewInventoryTransaction {
	newInventoryTransactions := make([]NewInventoryTransaction, n)

	// floor_worker1 UUID — stable across all environments
	floorWorker1 := uuid.MustParse("c0000000-0000-4000-8000-000000000001")

	transactionTypes := []string{"receive", "pick", "putaway", "transfer", "adjustment", "count"}

	idx := rand.Intn(10000)

	for i := 0; i < n; i++ {
		idx++

		userID := userIDs[idx%len(userIDs)]
		if i < 5 {
			userID = floorWorker1
		}

		newInventoryTransactions[i] = NewInventoryTransaction{
			LocationID:      locationIDs[idx%len(locationIDs)],
			ProductID:       productIDs[idx%len(productIDs)],
			UserID:          userID,
			TransactionType: transactionTypes[i%len(transactionTypes)],
			Quantity:        rand.Intn(100) + 1,
			ReferenceNumber: fmt.Sprintf("REF-%04d", i+1),
			TransactionDate: time.Now().AddDate(0, 0, -(i % 7)),
		}
	}

	return newInventoryTransactions
}
```

- [ ] **Step 2: Build and test**

Run: `go build ./business/domain/inventory/inventorytransactionbus/...`
Run: `go test ./business/domain/inventory/inventorytransactionbus/... -count=1 -v`
Run: `go test ./api/cmd/services/ichor/tests/inventory/inventorytransactionapi/... -count=1 -v`
Expected: All pass.

- [ ] **Step 3: Commit**

```bash
git add business/domain/inventory/inventorytransactionbus/testutil.go
git commit -m "feat(seed): varied transaction types, floor_worker1, recent dates"
```

### Task 18: PR 3a — Build, Full Test, Push

- [ ] **Step 1: Full build and test**

Run: `go build ./...`
Run: `go test ./business/domain/inventory/inventoryadjustmentbus/... ./business/domain/inventory/inventorytransactionbus/... -count=1`
Run: `go test ./api/cmd/services/ichor/tests/inventory/inventoryadjustmentapi/... ./api/cmd/services/ichor/tests/inventory/inventorytransactionapi/... -count=1`
Expected: All pass.

- [ ] **Step 2: Create PR**

```bash
git push origin HEAD
gh pr create --title "feat(seed): PR 3a — varied adjustments + transaction types" --body "..."
```

---

## PR 3b: Cycle Counts + Approvals + Config

### Task 19: Wire Cycle Count Sessions into seedFrontend

**Files:**
- Modify: `business/sdk/dbtest/seed_tasks.go`

- [ ] **Step 1: Add cycle count seeding to seed_tasks.go**

Add imports for cycle count packages and add to the `seedTasks` function:

```go
import (
	// ... existing imports ...
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
)
```

At the end of `seedTasks`, add:

```go
	// Seed 3 cycle count sessions
	sessions, err := cyclecountsessionbus.TestSeedCycleCountSessions(ctx, 3, adminIDs, busDomain.CycleCountSession)
	if err != nil {
		return fmt.Errorf("seeding cycle count sessions: %w", err)
	}

	sessionIDs := make(uuid.UUIDs, len(sessions))
	for i, s := range sessions {
		sessionIDs[i] = s.ID
	}

	// Seed cycle count items (5 per session = 15 total)
	_, err = cyclecountitembus.TestSeedCycleCountItems(ctx, 15, sessionIDs, productIDs, locationIDs, busDomain.CycleCountItem)
	if err != nil {
		return fmt.Errorf("seeding cycle count items: %w", err)
	}
```

- [ ] **Step 2: Build and test**

Run: `go build ./business/sdk/dbtest/...`
Expected: Clean compilation.

- [ ] **Step 3: Commit**

```bash
git add business/sdk/dbtest/seed_tasks.go
git commit -m "feat(seed): wire cycle count sessions + items into seedFrontend"
```

### Task 20: Add ApprovalRequest to BusDomain and Create TestSeed

**Files:**
- Modify: `business/sdk/dbtest/dbtest.go`
- Create: `business/domain/workflow/approvalrequestbus/testutil.go`
- Modify: `business/sdk/dbtest/seed_tasks.go`

- [ ] **Step 1: Add ApprovalRequest field to BusDomain**

In `business/sdk/dbtest/dbtest.go`, add to the BusDomain struct (in the Workflow section):

```go
	ApprovalRequest *approvalrequestbus.Business
```

Add the import:
```go
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
```

In `newBusDomains()`, add the instantiation. Check the `NewBusiness` signature — it takes `(log, delegate, storer)`:

```go
	ApprovalRequest: approvalrequestbus.NewBusiness(log, delegate, approvalrequestdb.NewStore(log, db)),
```

Add the store import:
```go
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus/stores/approvalrequestdb"
```

- [ ] **Step 2: Create testutil.go for approvalrequestbus**

Create `business/domain/workflow/approvalrequestbus/testutil.go`:

```go
package approvalrequestbus

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// TestNewApprovalRequests generates n new approval requests for testing.
func TestNewApprovalRequests(n int, approverIDs uuid.UUIDs) []NewApprovalRequest {
	requests := make([]NewApprovalRequest, n)

	taskNames := []string{
		"Transfer Order Approval",
		"Inventory Adjustment Review",
		"Purchase Order Approval",
		"Quality Hold Release",
		"Cycle Count Variance Review",
	}

	for i := range n {
		requests[i] = NewApprovalRequest{
			ExecutionID:     uuid.New(),
			RuleID:          uuid.New(),
			ActionName:      fmt.Sprintf("approve_%s", taskNames[i%len(taskNames)]),
			Approvers:       approverIDs,
			ApprovalType:    ApprovalTypeAny,
			TimeoutHours:    24,
			TaskToken:       fmt.Sprintf("task-token-%04d", i+1),
			ApprovalMessage: fmt.Sprintf("Please review: %s #%d", taskNames[i%len(taskNames)], i+1),
		}
	}

	return requests
}

// TestSeedApprovalRequests creates n approval requests in the database for testing.
func TestSeedApprovalRequests(ctx context.Context, n int, approverIDs uuid.UUIDs, api *Business) ([]ApprovalRequest, error) {
	newRequests := TestNewApprovalRequests(n, approverIDs)

	requests := make([]ApprovalRequest, len(newRequests))
	for i, nr := range newRequests {
		req, err := api.Create(ctx, nr)
		if err != nil {
			return nil, fmt.Errorf("seeding approval request %d: %w", i, err)
		}
		requests[i] = req
	}

	sort.Slice(requests, func(i, j int) bool {
		return requests[i].ID.String() < requests[j].ID.String()
	})

	return requests, nil
}
```

- [ ] **Step 3: Wire approval requests into seed_tasks.go**

Add to `seedTasks` function:

```go
	// Seed 5 workflow approval requests
	_, err = approvalrequestbus.TestSeedApprovalRequests(ctx, 5, adminIDs, busDomain.ApprovalRequest)
	if err != nil {
		return fmt.Errorf("seeding approval requests: %w", err)
	}
```

Add the import:
```go
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
```

- [ ] **Step 4: Build and test**

Run: `go build ./business/sdk/dbtest/... ./business/domain/workflow/approvalrequestbus/...`
Expected: Clean compilation.

- [ ] **Step 5: Commit**

```bash
git add business/sdk/dbtest/dbtest.go business/domain/workflow/approvalrequestbus/testutil.go business/sdk/dbtest/seed_tasks.go
git commit -m "feat(seed): add ApprovalRequest to BusDomain, seed 5 pending approvals"
```

### Task 21: Seed Config Settings (Optional)

**Files:**
- Modify: `business/sdk/dbtest/seed_tasks.go` (or create `seed_config.go`)

- [ ] **Step 1: Check if ConfigStore has a Create/Set method**

Read `business/sdk/dbtest/dbtest.go` and find the `ConfigStore` field type. Then check the business package for a method to set config values. If no suitable method exists, this can be deferred — the frontend falls back to hardcoded defaults.

- [ ] **Step 2: If method exists, add config seeding**

Add to `seedTasks` or a new `seedConfig` function:

```go
	// Seed expiry threshold config settings
	configEntries := map[string]string{
		"inventory.expiry_warning_1_days": "7",
		"inventory.expiry_warning_2_days": "30",
		"inventory.expiry_warning_3_days": "90",
	}
	for key, value := range configEntries {
		// Call busDomain.ConfigStore.Set(ctx, key, value) or equivalent
	}
```

The exact API depends on the ConfigStore interface — check and adapt.

- [ ] **Step 3: Commit**

```bash
git add business/sdk/dbtest/seed_tasks.go
git commit -m "feat(seed): add expiry threshold config settings"
```

### Task 22: PR 3b — Build, Full Test, Push

- [ ] **Step 1: Full build and test**

Run: `go build ./...`
Run: `go test ./business/domain/workflow/approvalrequestbus/... -count=1`
Expected: All pass.

- [ ] **Step 2: Smoke test seedFrontend end-to-end**

If you have a local database: `make seed-frontend`
Then verify with psql:

```sql
-- Products have realistic data
SELECT name, upc_code, tracking_type, is_perishable FROM products.products LIMIT 5;

-- Locations have warehouse codes
SELECT location_code, is_pick_location, is_reserve_location FROM inventory.inventory_locations LIMIT 5;

-- Put-away tasks exist
SELECT COUNT(*) FROM inventory.put_away_tasks;

-- Pick tasks exist
SELECT COUNT(*) FROM inventory.pick_tasks;

-- Lots have human-readable numbers
SELECT lot_number, quality_status, expiration_date FROM inventory.lot_trackings LIMIT 5;

-- Serials have real statuses
SELECT serial_number, status FROM inventory.serial_numbers LIMIT 5;

-- Transfers have mixed statuses
SELECT status, COUNT(*) FROM inventory.transfer_orders GROUP BY status;

-- Adjustments have varied reasons
SELECT reason_code, COUNT(*) FROM inventory.inventory_adjustments GROUP BY reason_code;

-- Transactions have varied types
SELECT transaction_type, COUNT(*) FROM inventory.inventory_transactions GROUP BY transaction_type;

-- Approval requests exist
SELECT COUNT(*), status FROM workflow.approval_requests GROUP BY status;

-- Cycle count sessions exist
SELECT COUNT(*) FROM inventory.cycle_count_sessions;

-- Inspections assigned to floor_worker1
SELECT COUNT(*) FROM inventory.inspections WHERE inspector_id = 'c0000000-0000-4000-8000-000000000001';
```

- [ ] **Step 3: Create PR**

```bash
git push origin HEAD
gh pr create --title "feat(seed): PR 3b — cycle counts, approval requests, config settings" --body "..."
```

---

## Cross-PR Test Verification Checklist

After all 7 PRs are merged, run the full test suite for all affected domains:

```bash
# All product-related
go test ./business/domain/products/... -count=1

# All inventory-related
go test ./business/domain/inventory/... -count=1

# All procurement-related
go test ./business/domain/procurement/... -count=1

# All workflow-related
go test ./business/domain/workflow/approvalrequestbus/... -count=1

# All API tests for affected domains
go test ./api/cmd/services/ichor/tests/products/... -count=1
go test ./api/cmd/services/ichor/tests/inventory/... -count=1
go test ./api/cmd/services/ichor/tests/procurement/... -count=1
```

Then run `make seed-frontend` and do a manual walkthrough as floor_worker1.
