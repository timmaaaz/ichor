# inventory-ops

[bus]=business [app]=application [api]=HTTP [db]=store [sdk]=shared
→=depends on ⊕=writes ⊗=reads ⚡=external [tx]=transaction [cache]=cached

---

## PutAwayTask — StateMachine

```
Pending → InProgress → Completed
                    ↘
                     Cancelled
```

terminal guard: no transitions out of Completed or Cancelled

transition side effects:
  → InProgress   auto-assigns: assigned_to = authenticated user, assigned_at = now()
  → Completed    triggers 3-way atomic write (see below)
  → Cancelled    plain update, no side effects

---

## PutAwayTaskApp [app]

file: app/domain/inventory/putawaytaskapp/putawaytaskapp.go

```go
type App struct {
    putAwayTaskBus       *putawaytaskbus.Business
    invTransactionBus    *inventorytransactionbus.Business
    invItemBus           *inventoryitembus.Business
    db                   *sqlx.DB
}
```

  NewApp(putAwayTaskBus, invTransactionBus, invItemBus, db) *App
  Create(ctx, app NewPutAwayTask) (PutAwayTask, error)
  Update(ctx, taskID uuid.UUID, app UpdatePutAwayTask) (PutAwayTask, error)
  Delete(ctx, taskID uuid.UUID) error
  Query(ctx, qp QueryParams) (query.Result[PutAwayTask], error)
  QueryByID(ctx, taskID uuid.UUID) (PutAwayTask, error)

Status Constants:
  Pending    → "pending"
  InProgress → "in_progress"
  Completed  → "completed"
  Cancelled  → "cancelled"

3-Way Atomic Write (Completed transition) — TX isolation: sql.LevelReadCommitted:
  1. ⊕ UPDATE inventory.put_away_tasks
       set status=completed, completed_by=userID, completed_at=now()
  2. ⊕ INSERT inventory.inventory_transactions
       new PUT_AWAY ledger record: ProductID, LocationID, UserID, Quantity, TransactionDate
  3. ⊕ UPSERT inventory.inventory_items
       add Quantity at (ProductID + LocationID) destination — creates row if absent

---

## ScanApp [app]

file: app/domain/inventory/scanapp/scanapp.go

```go
type App struct {
    productBus       *productbus.Business
    inventoryItemBus *inventoryitembus.Business
    locationBus      *inventorylocationbus.Business
    lotTrackingsBus  *lottrackingsbus.Business
    serialNumberBus  *serialnumberbus.Business
}
```

  NewApp(productBus, inventoryItemBus, locationBus, lotTrackingsBus, serialNumberBus) *App
  Scan(ctx, barcode string) (ScanResult, error)

Fan-Out Pattern — sync.WaitGroup with 4 concurrent goroutines (wg.Add(4)):
  goroutine 1 → productBus.QueryByUPC(barcode)
  goroutine 2 → locationBus.QueryByCode(barcode)
  goroutine 3 → lotTrackingsBus.QueryByLotNumber(barcode)
  goroutine 4 → serialNumberBus.QueryBySerialNumber(barcode)

key facts:
  - each goroutine: mu sync.Mutex guards write to shared result slice
  - fail-open: individual query errors return nil and are skipped — never block other goroutines
  - wg.Wait() after all 4 launch; then priority selection on results

Result Priority (highest first):
  1. serial   — serial number match
  2. lot      — lot number match
  3. product  — UPC code match
  4. location — location code match
  5. unknown  — no match (ScanResult{Type: "unknown", Data: nil})

first matching type in priority order wins; lower-priority results discarded

Per-Type Enrichment:
```
serial →  serialNumberBus.QueryLocationBySerialID(sn.SerialID)
            → LocationID, LocationCode, Aisle, Rack, Shelf, Bin, WarehouseName, ZoneName
            fail-open: partial result without location on error

lot    →  lotTrackingsBus.QueryLocationsByLotID(lot.LotID)
            → []Location{LocationID, LocationCode, Aisle, Rack, Shelf, Bin, Quantity}
            fail-open: empty locations list on error

product → inventoryItemBus.QueryWithLocationDetails(filter{ProductID}, page "1-100")
            → []InventoryItem{LocationID, LocationCode, Quantity, ReservedQuantity}
            fail-open: nil items on error

location → inventoryItemBus.QueryItemsWithProductAtLocation(loc.LocationID)
            → []Product{ProductID, ProductName, ProductSKU, TrackingType, Quantity}
            fail-open: nil items on error
```

---

## ⚠ Adding a new put-away status transition

  business/domain/inventory/putawaytaskbus/putawaytaskbus.go    (status constants)
  app/domain/inventory/putawaytaskapp/putawaytaskapp.go         (transition switch + side effect)
  business/sdk/migrate/sql/migrate.sql                          (if new status value added)
  api/cmd/services/ichor/tests/inventory/putawaytaskapi/        (update transition test)

## ⚠ Adding a new scan result type (new barcode domain)

  app/domain/inventory/scanapp/scanapp.go                       (add goroutine in wg.Add(N), add to priority check)
  app/domain/inventory/scanapp/model.go                         (new *ScanResult struct + type constant)
  app/domain/sales/pickingapp/ or other consumers               (handle new ScanResult.Type)

## ⚠ Changing the 3-way write (e.g. adding a 4th write)

  app/domain/inventory/putawaytaskapp/putawaytaskapp.go         (complete() method — all writes in one BeginTx block)
  business/sdk/migrate/sql/migrate.sql                          (new table/column if write targets new schema)
  api/cmd/services/ichor/tests/inventory/putawaytaskapi/        (verify atomic behavior in test)
