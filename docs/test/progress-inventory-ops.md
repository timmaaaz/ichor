# Progress Summary: inventory-ops.md

## Overview
Architecture for two inventory operation domains: put-away tasks (receiving/stocking) and barcode scanning with fan-out pattern.

## PutAwayTask State Machine

```
Pending → InProgress → Completed
                    ↘
                     Cancelled
```

**Terminal guard:** No transitions out of Completed or Cancelled

### Transition Side Effects
- **→ InProgress** — auto-assigns: `assigned_to = authenticated user`, `assigned_at = now()`
- **→ Completed** — triggers 3-way atomic write (see below)
- **→ Cancelled** — plain update, no side effects

## PutAwayTaskApp [app] — `app/domain/inventory/putawaytaskapp/putawaytaskapp.go`

**Responsibility:** Manage put-away task lifecycle and inventory updates.

### Struct
```go
type App struct {
    putAwayTaskBus       *putawaytaskbus.Business
    invTransactionBus    *inventorytransactionbus.Business
    invItemBus           *inventoryitembus.Business
    db                   *sqlx.DB
}
```

### Methods
- `NewApp(putAwayTaskBus, invTransactionBus, invItemBus, db) *App`
- `Create(ctx, app NewPutAwayTask) (PutAwayTask, error)`
- `Update(ctx, taskID uuid.UUID, app UpdatePutAwayTask) (PutAwayTask, error)`
- `Delete(ctx, taskID uuid.UUID) error`
- `Query(ctx, qp QueryParams) (query.Result[PutAwayTask], error)`
- `QueryByID(ctx, taskID uuid.UUID) (PutAwayTask, error)`

### Status Constants
- `Pending` → "pending"
- `InProgress` → "in_progress"
- `Completed` → "completed"
- `Cancelled` → "cancelled"

### 3-Way Atomic Write (Completed Transition)

**TX isolation: sql.LevelReadCommitted**

When task status → Completed, execute atomically:
1. **UPDATE inventory.put_away_tasks** — set status=completed, completed_by=userID, completed_at=now()
2. **INSERT inventory.inventory_transactions** — new PUT_AWAY ledger record: ProductID, LocationID, UserID, Quantity, TransactionDate
3. **UPSERT inventory.inventory_items** — add Quantity at (ProductID + LocationID) destination; creates row if absent

**Key:** All three writes succeed or all fail (atomicity).

## ScanApp [app] — `app/domain/inventory/scanapp/scanapp.go`

**Responsibility:** Barcode scanning with multi-concurrent enrichment.

### Struct
```go
type App struct {
    productBus       *productbus.Business
    inventoryItemBus *inventoryitembus.Business
    locationBus      *inventorylocationbus.Business
    lotTrackingsBus  *lottrackingsbus.Business
    serialNumberBus  *serialnumberbus.Business
}
```

### Methods
- `NewApp(productBus, inventoryItemBus, locationBus, lotTrackingsBus, serialNumberBus) *App`
- `Scan(ctx, barcode string) (ScanResult, error)`

### Fan-Out Pattern

Uses `sync.WaitGroup` with 4 concurrent goroutines (wg.Add(4)):
1. **goroutine 1** → productBus.QueryByUPC(barcode)
2. **goroutine 2** → locationBus.QueryByCode(barcode)
3. **goroutine 3** → lotTrackingsBus.QueryByLotNumber(barcode)
4. **goroutine 4** → serialNumberBus.QueryBySerialNumber(barcode)

### Key Facts
- Each goroutine uses `sync.Mutex` to guard writes to shared result slice
- **Fail-open:** Individual query errors return nil and are skipped; never block other goroutines
- `wg.Wait()` after all 4 launch; then priority selection on results

### Result Priority (Highest First)
1. **serial** — serial number match (most specific)
2. **lot** — lot number match
3. **product** — UPC code match
4. **location** — location code match
5. **unknown** — no match (ScanResult{Type: "unknown", Data: nil})

First matching type in priority order wins; lower-priority results discarded.

### Per-Type Enrichment

After priority selection, enrich with location/quantity details:

**Serial:**
```
serialNumberBus.QueryLocationBySerialID(sn.SerialID)
→ LocationID, LocationCode, Aisle, Rack, Shelf, Bin, WarehouseName, ZoneName
fail-open: partial result without location on error
```

**Lot:**
```
lotTrackingsBus.QueryLocationsByLotID(lot.LotID)
→ []Location{LocationID, LocationCode, Aisle, Rack, Shelf, Bin, Quantity}
fail-open: empty locations list on error
```

**Product:**
```
inventoryItemBus.QueryWithLocationDetails(filter{ProductID}, page "1-100")
→ []InventoryItem{LocationID, LocationCode, Quantity, ReservedQuantity}
fail-open: nil items on error
```

**Location:**
```
inventoryItemBus.QueryItemsWithProductAtLocation(loc.LocationID)
→ []Product{ProductID, ProductName, ProductSKU, TrackingType, Quantity}
fail-open: nil items on error
```

## Change Patterns

### ⚠ Adding a New Put-Away Status Transition
Affects 4 areas:
1. `business/domain/inventory/putawaytaskbus/putawaytaskbus.go` — status constants
2. `app/domain/inventory/putawaytaskapp/putawaytaskapp.go` — transition switch + side effect
3. `business/sdk/migrate/sql/migrate.sql` — if new status value added
4. `api/cmd/services/ichor/tests/inventory/putawaytaskapi/` — update transition test

### ⚠ Adding a New Scan Result Type (New Barcode Domain)
Affects 3 areas:
1. `app/domain/inventory/scanapp/scanapp.go` — add goroutine in wg.Add(N), add to priority check
2. `app/domain/inventory/scanapp/model.go` — new *ScanResult struct + type constant
3. `app/domain/sales/pickingapp/` or other consumers — handle new ScanResult.Type

### ⚠ Changing the 3-Way Write (e.g., Adding a 4th Write)
Affects 3 areas:
1. `app/domain/inventory/putawaytaskapp/putawaytaskapp.go` — complete() method (all writes in one BeginTx block)
2. `business/sdk/migrate/sql/migrate.sql` — new table/column if write targets new schema
3. `api/cmd/services/ichor/tests/inventory/putawaytaskapi/` — verify atomic behavior in test

## Critical Points
- **3-way write must be atomic** — all succeed or all fail (no partial updates)
- **Scan fan-out is fail-open** — individual query errors never block other goroutines
- **Priority selection is deterministic** — highest priority always wins
- **Per-type enrichment is optional** — fail-open if location queries fail

## Notes for Future Development
Put-away tasks implement careful atomic writes for stocking operations. Scan handles barcode ambiguity via concurrency and priority. Most changes will be:
- Adding new status transitions (moderate)
- Adding new scan result types (moderate)
- Changing atomic write operations (risky, requires testing atomicity)
