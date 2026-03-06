# Put-Away Task Queue + Location Code + UPC Uniqueness
## Implementation Plan — Phase 3 Backend

**Goal**: Three backend capabilities enabling a mobile floor worker put-away UX.
**Approach**: 5 phased increments, each independently testable.

---

## Architecture Decisions

### location_code
- **`VARCHAR(50) NULL` + UNIQUE constraint** — NULLs don't conflict with UNIQUE, so warehouses that don't use codes just leave it NULL. Consistent with codebase nullable pattern (`sql.NullString` in DB model, plain `string` in bus model).
- **Auto-generate on Create** (optional): if `location_code` omitted or empty, backend derives it from non-empty segments of `aisle`/`rack`/`shelf`/`bin` joined by `-`. Example: aisle="A", rack="01", shelf="", bin="" → "A-01". If all segments empty, `location_code` stays NULL.
- **On Update**: if `location_code` explicitly provided, use it; otherwise leave unchanged (don't re-derive).

### put_away_tasks completion side-effects
- **Workflow action handler** — the domain bus is pure CRUD + status transition validation. When status transitions to `completed`, the bus fires a delegate event. A new workflow action handler (`put_away_complete`) catches the `changed_to: "completed"` event via the TriggerEvent FieldChanges system and runs the multi-bus transaction (create inventory_transaction + find/update/create inventory_item).
- This is async/eventually consistent (Temporal), which matches industry-standard WMS patterns for put-away. Temporal's retry mechanism handles failures without human intervention.

### UPC uniqueness (products)
- Add DB constraint (migration).
- Products already has ErrUniqueEntry handling in `Create` — add same handling to `Update` in productapp.

---

## Phase 1: Database Schema Migrations
**Scope**: migrate.sql only. No Go code changes.
**Test**: `make migrate && make seed` succeeds; verify schema with pgcli.

### Checklist
- [ ] Version 1.998: `ALTER TABLE inventory.inventory_locations ADD COLUMN location_code VARCHAR(50) NULL, ADD CONSTRAINT inventory_locations_location_code_key UNIQUE (location_code)`
- [ ] Version 1.999: `ALTER TABLE products.products ADD CONSTRAINT products_upc_code_key UNIQUE (upc_code)`
- [ ] Version 2.000: `CREATE TABLE inventory.put_away_tasks (...)` with all columns + FK constraints + indexes
- [ ] Add `table_access` entries for `inventory.put_away_tasks` in seed.sql (grant admin full access)

### put_away_tasks DDL
```sql
-- Version: 2.000
-- Description: add inventory.put_away_tasks
CREATE TABLE inventory.put_away_tasks (
    id UUID NOT NULL,
    product_id UUID NOT NULL,
    location_id UUID NOT NULL,
    quantity INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    reference_number VARCHAR(100) NOT NULL,
    assigned_to UUID,
    assigned_at TIMESTAMP,
    completed_by UUID,
    completed_at TIMESTAMP,
    notes VARCHAR(500),
    created_date TIMESTAMP NOT NULL,
    updated_date TIMESTAMP NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (product_id) REFERENCES products.products(id),
    FOREIGN KEY (location_id) REFERENCES inventory.inventory_locations(id),
    FOREIGN KEY (assigned_to) REFERENCES core.users(id),
    FOREIGN KEY (completed_by) REFERENCES core.users(id)
);
CREATE INDEX idx_put_away_tasks_status ON inventory.put_away_tasks(status);
CREATE INDEX idx_put_away_tasks_assigned_to ON inventory.put_away_tasks(assigned_to);
```

---

## Phase 2: `location_code` on `inventory.inventory_locations`
**Builds on**: Phase 1 (schema must exist).
**Test**: `go test ./api/cmd/services/ichor/tests/inventory/inventorylocationapi/...`

This is a vertical slice change — all 4 layers plus tests.

### 2a — DB layer (`inventorylocationbus/stores/inventorylocationdb/`)
- [ ] `model.go`: add `LocationCode sql.NullString \`db:"location_code"\`` to `inventoryLocation` struct; update `toDBInvLocation` and `toBusInvLocation` conversions (Valid check for null handling)
- [ ] `inventorylocationdb.go`: add `location_code` to INSERT col list + `:location_code` to VALUES; add `location_code = :location_code` to UPDATE SET; add `il.location_code` to SELECT list
- [ ] `filter.go`: add `if filter.LocationCode != nil` block generating `location_code = :location_code`
- [ ] `order.go`: add `inventorylocationbus.OrderByLocationCode: "location_code"` entry

### 2b — Business layer (`inventorylocationbus/`)
- [ ] `model.go`: add `LocationCode string` to `InventoryLocation`; add `LocationCode string` to `NewInventoryLocation`; add `LocationCode *string` to `UpdateInventoryLocation`
- [ ] `filter.go`: add `LocationCode *string` to `QueryFilter`
- [ ] `order.go`: add `OrderByLocationCode = "location_code"` constant
- [ ] `inventorylocationbus.go`: in `Create()`, call auto-generate helper if `LocationCode == ""`; add `LocationCode` assignment; in `Update()`, add `if u.LocationCode != nil` block

### 2c — Auto-generate logic (in bus or a private helper)
```go
// If LocationCode is empty, derive from non-empty segments
func deriveLocationCode(aisle, rack, shelf, bin string) string {
    parts := []string{}
    for _, s := range []string{aisle, rack, shelf, bin} {
        if s != "" { parts = append(parts, s) }
    }
    return strings.Join(parts, "-")
}
```
Called in `Create()` when `newLoc.LocationCode == ""`.

### 2d — App layer (`inventorylocationapp/`)
- [ ] `model.go`: add `LocationCode string \`json:"location_code"\`` to `InventoryLocation` (response); add `LocationCode string \`json:"location_code"\`` to `NewInventoryLocation` (NOT required — allow empty for auto-gen); add `LocationCode *string \`json:"location_code"\`` to `UpdateInventoryLocation`; update all `ToApp*` and `toBus*` conversion functions; add `LocationCode string` to `QueryParams`
- [ ] `filter.go`: add `if qp.LocationCode != ""` block
- [ ] `order.go`: add `"location_code": inventorylocationbus.OrderByLocationCode` entry

### 2e — API layer (`inventorylocationapi/`)
- [ ] `filter.go`: add `LocationCode: values.Get("location_code")` to `QueryParams`

### 2f — Tests + testutil
- [ ] `testutil.go`: add `LocationCode` generation in `TestNewInventoryLocation` (e.g. `fmt.Sprintf("LOC-%d", idx)`)
- [ ] `create_test.go`: add `LocationCode` to all create inputs and expected responses; add test for auto-generate (omit location_code, verify it derives from aisle/rack/shelf/bin)
- [ ] `update_test.go`: add `LocationCode` to update inputs and expected responses
- [ ] `query_test.go`: add `location_code` filter test: GET with `?location_code=LOC-1` returns exactly 1 result

---

## Phase 3: UPC Uniqueness on `products.products`
**Builds on**: Phase 1.
**Test**: `go test ./api/cmd/services/ichor/tests/products/productapi/...`

Minimal changes — products already has the error handling wired for `Create`; just fill the gap for `Update`.

### Checklist
- [ ] `productapp/productapp.go`: in `Update()`, add `errors.Is(err, productbus.ErrUniqueEntry)` check returning `errs.New(errs.AlreadyExists, productbus.ErrUniqueEntry)` (mirrors the existing `Create()` check)
- [ ] `update_test.go`: add `update409` case for duplicate UPC code (create two products; try to update second with first's UPC code; expect 409)
- [ ] `create_test.go`: add `create409` case for duplicate UPC code (create product; try to create another with same UPC; expect 409)
- [ ] Verify no duplicate UPC codes in test seed data (confirmed: testutil generates unique via `idx` counter — no changes needed)

---

## Phase 4: `put_away_tasks` Domain — Pure CRUD + Status Transitions
**Builds on**: Phase 1.
**Test**: `go test ./api/cmd/services/ichor/tests/inventory/putawaytaskapi/...`

Full new domain following `inventorytransactionbus` structure. Side-effects (inventory update) are NOT in this phase — they come in Phase 5 via workflow action.

### 4a — Business layer (NEW files)
- [ ] `putawaytaskbus/model.go`: define `PutAwayTask`, `NewPutAwayTask`, `UpdatePutAwayTask`
  - Nullable fields: `AssignedTo uuid.UUID`, `AssignedAt time.Time`, `CompletedBy uuid.UUID`, `CompletedAt time.Time`, `Notes string` (zero value = not set)
  - Status constants: `StatusPending`, `StatusInProgress`, `StatusCompleted`, `StatusCancelled`
- [ ] `putawaytaskbus/filter.go`: `QueryFilter` with `Status *string`, `ProductID *uuid.UUID`, `LocationID *uuid.UUID`, `AssignedTo *uuid.UUID`
- [ ] `putawaytaskbus/order.go`: constants for all orderable fields
- [ ] `putawaytaskbus/event.go`: `DomainName`, `EntityName`, `ActionCreated/Updated/Deleted` constants; `ActionCreatedParms`, `ActionUpdatedParms` (include `BeforeEntity` for FieldChanges); `ActionXxxData()` functions
- [ ] `putawaytaskbus/putawaytaskbus.go`: `Business` struct, `NewBusiness()`, CRUD methods; status transition validation in `Update()`:
  - `pending → in_progress`: allowed (sets AssignedTo, AssignedAt)
  - `in_progress → completed`: allowed (sets CompletedBy, CompletedAt)
  - `in_progress → pending`: allowed (clears AssignedTo, AssignedAt — "abandon")
  - `any → cancelled`: allowed if current != completed
  - Invalid transition: return `ErrInvalidStatusTransition` error
  - Fire delegate events on create/update/delete for workflow triggering
- [ ] `putawaytaskbus/testutil.go`: `TestNewPutAwayTask(n int, productIDs, locationIDs []uuid.UUID)` helper

### 4b — DB store layer (NEW files)
- [ ] `stores/putawaytaskdb/model.go`: DB struct with `sql.NullTime` for nullable timestamps, `uuid.NullUUID` for nullable UUID FKs, `sql.NullString` for notes; `toDBPutAwayTask()` and `toBusPutAwayTask()` conversions
- [ ] `stores/putawaytaskdb/putawaytaskdb.go`: `Store` struct implementing `Storer` interface; all CRUD + `NewWithTx()` method
- [ ] `stores/putawaytaskdb/filter.go`: `applyFilter()` generating SQL WHERE clause
- [ ] `stores/putawaytaskdb/order.go`: SQL column mapping

### 4c — App layer (NEW files)
- [ ] `putawaytaskapp/model.go`: `PutAwayTask`, `NewPutAwayTask`, `UpdatePutAwayTask` with JSON tags + validate tags; `QueryParams`; conversion functions
- [ ] `putawaytaskapp/filter.go`: parse query params into `QueryFilter`
- [ ] `putawaytaskapp/putawaytaskapp.go`: thin App struct; error mapping for `ErrInvalidStatusTransition` → `errs.FailedPrecondition` (HTTP 400); FK violations → `errs.AlreadyExists`

### 4d — API layer (NEW files)
- [ ] `putawaytaskapi/putawaytaskapi.go`: handlers for create, query, queryByID, update, delete
- [ ] `putawaytaskapi/routes.go`: route registration; `RouteTable = "inventory.put_away_tasks"`
- [ ] `putawaytaskapi/filter.go`: parse HTTP query params

### 4e — Wiring in `all.go`
- [ ] Instantiate: `putAwayTaskBus := putawaytaskbus.NewBusiness(cfg.Log, delegate, putawaytaskdb.NewStore(cfg.Log, cfg.DB))`
- [ ] Register delegate: `delegateHandler.RegisterDomain(delegate, putawaytaskbus.DomainName, putawaytaskbus.EntityName)`
- [ ] Register routes: `putawaytaskapi.Routes(app, putawaytaskapi.Config{PutAwayTaskBus: putAwayTaskBus, AuthClient: cfg.AuthClient, Log: cfg.Log, PermissionsBus: permissionsBus})`

### 4f — Integration tests
- [ ] `tests/inventory/putawaytaskapi/seed_test.go`: seed 2 users, 1 warehouse, zones, products, locations, 4 put_away_tasks (1 pending, 1 in_progress, 1 completed, 1 cancelled)
- [ ] `tests/inventory/putawaytaskapi/create_test.go`: create200, create400, create401
- [ ] `tests/inventory/putawaytaskapi/query_test.go`: query200, queryByID200, query with status filter
- [ ] `tests/inventory/putawaytaskapi/update_test.go`: update200 (claim pending→in_progress), update200 (abandon in_progress→pending), update400 (invalid transition pending→completed), update401
- [ ] `tests/inventory/putawaytaskapi/delete_test.go`: delete200, delete401

---

## Phase 5: Put-Away Completion Workflow Action (side-effects)
**Builds on**: Phase 4 (put_away_tasks domain + event firing must exist).
**Test**: `go test ./api/cmd/services/ichor/tests/inventory/putawaytaskapi/...` (add workflow test); `go build ./...`

This phase wires the async inventory side-effects via the Temporal workflow system.

### How it works
1. Worker calls `PUT /v1/inventory/put-away-tasks/{id}` with `{"status": "completed", "completed_by": "..."}`
2. `putawaytaskbus.Update()` validates the transition, saves the record, fires an `ActionUpdated` delegate event with `BeforeEntity.Status = "in_progress"`, `Entity.Status = "completed"`
3. `DelegateHandler` computes `FieldChanges["status"] = {OldValue: "in_progress", NewValue: "completed"}` automatically
4. `TriggerProcessor` matches against workflow rules — finds the rule: `trigger=on_update, entity=put_away_tasks, condition: status changed_to "completed"`
5. Temporal dispatches the `put_away_complete` workflow → executes the action handler
6. Action handler atomically: creates `inventory_transaction` (type "PUT_AWAY") + finds/creates/updates `inventory_item`

### Checklist

#### New workflow action handler
- [ ] `business/sdk/workflow/workflowactions/inventory/putawaycomplete.go`:
  - `PutAwayCompleteHandler` struct holding `log`, `db *sqldb.Database`, `inventoryItemBus`, `inventoryTransactionBus`
  - `GetType() string` returns `"put_away_complete"`
  - `IsAsync() bool` returns `false`
  - `Execute()`:
    1. Extract `task` from `execCtx.TriggerEvent.RawData` (JSON decode into `putawaytaskbus.PutAwayTask`)
    2. `tx, err := h.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})`
    3. `defer tx.Rollback()`
    4. `txItemBus, _ := h.inventoryItemBus.NewWithTx(tx)`
    5. `txTxnBus, _ := h.inventoryTransactionBus.NewWithTx(tx)`
    6. Query `txItemBus` with `{ProductID: task.ProductID, LocationID: task.LocationID}`
    7. If found: `txItemBus.Update(ctx, item, {Quantity: &newQty, LocationID: &task.LocationID})`
    8. If not found: `txItemBus.Create(ctx, NewInventoryItem{ProductID: task.ProductID, LocationID: task.LocationID, Quantity: task.Quantity, ...defaults})`
    9. `txTxnBus.Create(ctx, NewInventoryTransaction{ProductID: task.ProductID, LocationID: task.LocationID, UserID: task.CompletedBy, Quantity: task.Quantity, TransactionType: "PUT_AWAY", ReferenceNumber: task.ReferenceNumber, TransactionDate: now})`
    10. `tx.Commit()`

#### Register the action
- [ ] `register.go`: add `NewPutAwayCompleteHandler` to `RegisterGranularInventoryActions()` (or `RegisterAll()`); add `InventoryItemBus` and `InventoryTransactionBus` to `BusDependencies` if not already present
- [ ] `all.go`: pass `InventoryItemBus: inventoryItemBus, InventoryTransactionBus: inventoryTransactionBus` in the inventory action config block

#### Seed workflow rule for tests
- [ ] `tests/inventory/putawaytaskapi/seed_test.go`: after seeding put-away tasks, seed a workflow rule:
  - Trigger: `on_update`, entity: `put_away_tasks`
  - Condition: `{field_name: "status", operator: "changed_to", value: "completed"}`
  - Action: `{action_type: "put_away_complete", config: {}}`
  - Edge: start → action (no branches)

#### Integration test for completion side-effects
- [ ] `tests/inventory/putawaytaskapi/update_test.go`: add `update200_complete` test:
  1. Claim a pending task (PUT status=in_progress)
  2. Complete it (PUT status=completed, completed_by=userID)
  3. Wait for Temporal to process (or use synchronous test infra with `apitest.WorkflowInfra`)
  4. Query `inventory_transactions` — verify PUT_AWAY transaction exists
  5. Query `inventory_items` by product+location — verify quantity updated

---

## Testing Strategy Per Phase

| Phase | Test Command | What's Verified |
|-------|-------------|-----------------|
| 1 | `make migrate && make seed` | Schema correct, no constraint errors |
| 2 | `go test ./api/cmd/services/ichor/tests/inventory/inventorylocationapi/...` | location_code CRUD + filter + auto-gen |
| 3 | `go test ./api/cmd/services/ichor/tests/products/productapi/...` | Duplicate UPC → 409 on create + update |
| 4 | `go test ./api/cmd/services/ichor/tests/inventory/putawaytaskapi/...` | CRUD + status transitions + auth |
| 5 | `go test ./api/cmd/services/ichor/tests/inventory/putawaytaskapi/...` | Completion creates transaction + updates item |
| All | `go build ./...` | Compilation across all layers |

---

## Key Files Reference (by phase)

### Phase 2 touch points
| Layer | File |
|-------|------|
| Migration | `business/sdk/migrate/sql/migrate.sql` |
| DB model | `business/domain/inventory/inventorylocationbus/stores/inventorylocationdb/model.go` |
| DB store | `business/domain/inventory/inventorylocationbus/stores/inventorylocationdb/inventorylocationdb.go` |
| DB filter | `business/domain/inventory/inventorylocationbus/stores/inventorylocationdb/filter.go` |
| DB order | `business/domain/inventory/inventorylocationbus/stores/inventorylocationdb/order.go` |
| Bus model | `business/domain/inventory/inventorylocationbus/model.go` |
| Bus filter | `business/domain/inventory/inventorylocationbus/filter.go` |
| Bus order | `business/domain/inventory/inventorylocationbus/order.go` |
| Bus main | `business/domain/inventory/inventorylocationbus/inventorylocationbus.go` |
| Bus testutil | `business/domain/inventory/inventorylocationbus/testutil.go` |
| App model | `app/domain/inventory/inventorylocationapp/model.go` |
| App filter | `app/domain/inventory/inventorylocationapp/filter.go` |
| App order | `app/domain/inventory/inventorylocationapp/order.go` |
| API filter | `api/domain/http/inventory/inventorylocationapi/filter.go` |
| Tests | `api/cmd/services/ichor/tests/inventory/inventorylocationapi/*.go` |

### Phase 3 touch points
| Layer | File |
|-------|------|
| Migration | `business/sdk/migrate/sql/migrate.sql` |
| App | `app/domain/products/productapp/productapp.go` |
| Tests | `api/cmd/services/ichor/tests/products/productapi/create_test.go` |
| Tests | `api/cmd/services/ichor/tests/products/productapi/update_test.go` |

### Phase 4 new files
All in `business/domain/inventory/putawaytaskbus/`, `app/domain/inventory/putawaytaskapp/`, `api/domain/http/inventory/putawaytaskapi/`, `api/cmd/services/ichor/tests/inventory/putawaytaskapi/`.
Wiring: `api/cmd/services/ichor/build/all/all.go`.

### Phase 5 new files
`business/sdk/workflow/workflowactions/inventory/putawaycomplete.go`
Wiring: `business/sdk/workflow/workflowactions/register.go` + `all.go`.

---

## Reference Implementations

| Need | Look at |
|------|---------|
| Vertical slice structure | `business/domain/inventory/inventorytransactionbus/` → through all layers |
| Nullable DB fields (sql.NullString, uuid.NullUUID, sql.NullTime) | `business/domain/procurement/purchaseorderbus/stores/purchaseorderdb/model.go` |
| Multi-bus DB transaction (atomic side-effects) | `business/sdk/workflow/workflowactions/inventory/receive.go` |
| Workflow action handler interface | `business/sdk/workflow/workflowactions/inventory/receive.go` (implements ActionHandler) |
| Event.go pattern | `business/domain/inventory/inventorylocationbus/event.go` |
| All.go wiring for inventory | `api/cmd/services/ichor/build/all/all.go` lines ~385–420 and ~940–987 |
| Status transition error mapping | `app/domain/inventory/inventorytransactionapp/inventorytransactionapp.go` (FK + not-found handling) |
