# Blocker 003: Adjustments — Approve Does Not Apply qty_change to Inventory

> **STATUS: CONFIRMED** (verified 2026-03-16)
> `Approve()` at `inventoryadjustmentbus.go:189-213` only sets status fields.
> Bus struct has no `inventoryItemBus` or `inventoryTransactionBus`.
> App struct (`inventoryadjustmentapp.go:18-21`) also has no inventory deps.
> `all.go:1047-1052` wires only the adjustment bus — no inventory buses passed.
> Zero `inventory_transaction` rows with type `"adjustment"` exist in codebase.

**Severity:** CRITICAL — approving an adjustment creates a paper trail but doesn't change actual inventory
**Domain:** `inventory/inventoryadjustment`
**Backend repo:** `../../../ichor/` relative to Vue frontend

---

## Problem

`inventoryadjustmentbus.Approve()` transitions `ApprovalStatus: pending → approved` and saves
the record, but **never touches `inventory_items` or `inventory_transactions`**. Approving an
adjustment is a no-op on actual inventory.

This also cascades to Cycle Count: variance submission creates adjustments, which go through the
same broken approval path, so reconciled counts never change inventory.

**Exact location:**
```
business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus.go  Approve()
```

Current `Approve` only does:
```go
ia.ApprovalStatus = ApprovalStatusApproved
b.storer.Update(ctx, ia)   // saves status change — NO inventory mutation
```

---

## What Should Happen

When an adjustment is approved, a 3-way atomic write should occur (same pattern as
`putawaytaskapp` — see `app/domain/inventory/putawaytaskapp/putawaytaskapp.go`):

1. **UPDATE** `inventory.inventory_adjustments` — set `approval_status = approved`
2. **INSERT** `inventory.inventory_transactions` — new record:
   - `type = "adjustment"`
   - `quantity = ia.QtyChange` (signed — positive = add, negative = subtract)
   - `product_id`, `location_id`, `user_id` from the adjustment record
3. **UPSERT** `inventory.inventory_items` — apply `qty_change`:
   - Find row by `(product_id, location_id)`
   - `SET quantity = quantity + ia.QtyChange`
   - If no row exists and `qty_change > 0`, INSERT new inventory_items row

TX isolation: `sql.LevelReadCommitted` (matches `putawaytaskapp` pattern)

---

## Changes Required

### Option A: Elevate to an App layer (recommended — matches codebase patterns)

Create `app/domain/inventory/inventoryadjustmentapp/inventoryadjustmentapp.go` as an
**orchestrating app** (like `putawaytaskapp`) that holds `inventoryItemBus` and
`inventoryTransactionBus` in addition to `inventoryadjustmentbus`.

```go
type App struct {
    inventoryadjustmentbus  *inventoryadjustmentbus.Business
    inventoryTransactionBus *inventorytransactionbus.Business
    inventoryItemBus        *inventoryitembus.Business
    db                      *sqlx.DB
    auth                    *auth.Auth
}
```

The `Approve` method in this app layer runs the 3-way write inside a transaction.

**Files to create/modify:**
- `app/domain/inventory/inventoryadjustmentapp/inventoryadjustmentapp.go` — add `db`, `invTransactionBus`, `invItemBus` to `App` struct; update `Approve` with TX block
- `api/cmd/services/ichor/build/all/all.go` — wire the new dependencies when constructing `inventoryadjustmentApp`

### Option B: Add side effects inside business layer

Extend `inventoryadjustmentbus.Business` struct to hold `inventoryItemBus` and
`inventoryTransactionBus` (acceptable if the team prefers keeping logic in bus). Less
idiomatic for this codebase — check with arch docs first.

---

## Inventory Model Fields to Verify

Check `business/domain/inventory/inventoryadjustmentbus/model.go` for field names:
- `QtyChange` (or `QuantityChange`) — the signed delta to apply
- `ProductID` — FK to products
- `LocationID` — FK to inventory_locations
- `RequestedBy` — the user submitting the adjustment
- `ApprovedBy` — set during Approve

---

## Key Files to Read Before Implementing

- `docs/arch/inventory-ops.md` — 3-way atomic write pattern (PutAwayTask Completed transition)
- `app/domain/inventory/putawaytaskapp/putawaytaskapp.go` — exact reference implementation
- `business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus.go` — current Approve
- `business/domain/inventory/inventoryadjustmentbus/model.go` — field names
- `business/domain/inventory/inventorytransactionbus/model.go` — transaction type constants

---

## Acceptance Criteria

- [ ] `POST /v1/inventory/inventory-adjustments/{id}/approve` applies `qty_change` to `inventory_items`
- [ ] An `inventory_transactions` row with `type="adjustment"` is created on approval
- [ ] Rejection (`/reject`) makes no inventory change (existing behavior correct)
- [ ] Negative `qty_change` decrements inventory (manual shrinkage correction)
- [ ] Approving a zero-qty adjustment is rejected or is a no-op (decide and document)
- [ ] TX rolls back entirely if any of the 3 writes fail
- [ ] Integration test: approve adjustment, query inventory item, verify quantity changed
- [ ] `go test ./app/domain/inventory/inventoryadjustmentapp/...` passes
- [ ] `go build ./...` passes
