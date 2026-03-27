# Blocker 004: Transfers — No Execute Flow (Missing in_transit/completed Statuses + Atomic Stock Move)

> **STATUS: CONFIRMED** (verified 2026-03-16)
> Only 3 statuses: `pending|approved|rejected` at `transferorderbus.go:28-32`.
> No `Claim`/`Execute` methods at any layer. No `/claim` or `/execute` routes.
> `Approve()` at lines 209-234 only sets status — no stock movement.
> `all.go:1064-1069` wires only the transfer order bus — no inventory deps.

**Severity:** CRITICAL — transfers can be created and approved but never actually move stock
**Domain:** `inventory/transferorder`
**Backend repo:** `../../../ichor/` relative to Vue frontend

---

## Problem

`transferorderbus` has only three statuses: `pending | approved | rejected`. The floor execute
flow requires `in_transit` (worker claims and walks to source) and `completed` (worker delivers
to destination, stock moves). Neither status exists and there is no endpoint that atomically
moves inventory between locations.

**Current status constants:**
```
business/domain/inventory/transferorderbus/transferorderbus.go:29-31
    StatusPending  = "pending"
    StatusApproved = "approved"
    StatusRejected = "rejected"
```

**Missing:** `StatusInTransit = "in_transit"` and `StatusCompleted = "completed"`

---

## What Should Happen

The floor execute flow has two phases:

### Phase 1: Claim (approved → in_transit)
Worker claims an approved transfer. Assigns it to themselves and marks it in-transit.

Endpoint: `POST /v1/inventory/transfer-orders/{id}/claim`
- Validates: `status == approved`
- Writes: `status = in_transit`, `assigned_to = userID`, `assigned_at = now()`
- No inventory change at this point

### Phase 2: Execute (in_transit → completed)
Worker delivers goods to destination. Stock moves atomically.

Endpoint: `POST /v1/inventory/transfer-orders/{id}/execute`
TX isolation: `sql.LevelReadCommitted`

3-way atomic write:
1. **UPDATE** `inventory.transfer_orders` → `status = completed`, `completed_at = now()`
2. **UPDATE** `inventory.inventory_items` at source location: `quantity -= transfer.Quantity`
   - Error if source qty < transfer.Quantity (prevent negative inventory)
3. **UPSERT** `inventory.inventory_items` at destination location: `quantity += transfer.Quantity`
   - Creates new row if product not yet at destination
4. **INSERT** `inventory.inventory_transactions` — two records:
   - `type = "transfer_out"`, source location, negative qty
   - `type = "transfer_in"`, destination location, positive qty

---

## Changes Required

### 1. Add statuses to business layer
File: `business/domain/inventory/transferorderbus/transferorderbus.go`

```go
const (
    StatusPending   = "pending"
    StatusApproved  = "approved"
    StatusRejected  = "rejected"
    StatusInTransit = "in_transit"
    StatusCompleted = "completed"
)
```

### 2. Add status values to migration
File: `business/sdk/migrate/sql/migrate.sql`

If `status` is a VARCHAR (not CHECK constraint), no migration needed.
If it's an enum or has a CHECK constraint, add new values.

Check: `grep -A 10 "transfer_orders" business/sdk/migrate/sql/migrate.sql`

### 3. Create orchestrating app layer
File: `app/domain/inventory/transferorderapp/transferorderapp.go`

Add `Claim` and `Execute` methods to the `App` struct. `Execute` needs `inventoryItemBus`,
`inventoryTransactionBus`, and `db` for transaction management (same pattern as `putawaytaskapp`).

```go
type App struct {
    transferorderbus        *transferorderbus.Business
    inventoryItemBus        *inventoryitembus.Business
    inventoryTransactionBus *inventorytransactionbus.Business
    db                      *sqlx.DB
    auth                    *auth.Auth
}
```

### 4. Add HTTP endpoints to transferorderapi
File: `api/domain/http/inventory/transferorderapi/`

Add routes:
- `POST /v1/inventory/transfer-orders/{id}/claim` → `app.Claim()`
- `POST /v1/inventory/transfer-orders/{id}/execute` → `app.Execute()`

### 5. Wire new dependencies
File: `api/cmd/services/ichor/build/all/all.go`

Add `inventoryItemBus` and `inventoryTransactionBus` when constructing `transferorderApp`.

---

## Key Files to Read Before Implementing

- `docs/arch/inventory-ops.md` — 3-way atomic write pattern (exact reference)
- `app/domain/inventory/putawaytaskapp/putawaytaskapp.go` — reference orchestrating app
- `business/domain/inventory/transferorderbus/transferorderbus.go` — current business layer
- `business/domain/inventory/transferorderbus/model.go` — TransferOrder model fields (SourceLocationID, DestLocationID, ProductID, Quantity)
- `api/domain/http/inventory/transferorderapi/` — existing routes to extend
- `business/sdk/migrate/sql/migrate.sql` — check status column constraints

---

## Acceptance Criteria

- [ ] `POST /v1/inventory/transfer-orders/{id}/claim` transitions `approved → in_transit`, assigns `assigned_to`
- [ ] `POST /v1/inventory/transfer-orders/{id}/execute` atomically: decrements source, increments dest, creates two transaction records, marks completed
- [ ] Source inventory goes negative → 400 error, full rollback
- [ ] `claim` on non-approved transfer → 400 error
- [ ] `execute` on non-in_transit transfer → 400 error
- [ ] Only the assigned worker (or admin) can execute their claimed transfer
- [ ] Integration test: full claim → execute cycle, verify both inventory_items rows updated
- [ ] `go build ./...` passes
