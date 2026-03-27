# Blocker 002: Picking — FEFO Allocation Falls Back to FIFO (SQL TODO)

> **STATUS: CONFIRMED** (verified 2026-03-16)
> `inventoryitemdb.go:269-276` — switch has `fifo`/`lifo` cases but no `fefo`. Default = FIFO.
> TODO block at lines 278-304 documents the gap explicitly.

**Severity:** HIGH — lot-tracked inventory picks ignore expiry date ordering
**Domain:** `inventory/inventoryitem`
**Backend repo:** `../../../ichor/` relative to Vue frontend

---

## Problem

`QueryAvailableForAllocation` in `inventoryitemdb.go` has a commented-out TODO for FEFO
(First Expired, First Out). When `strategy = "fefo"` is requested, the query silently falls
back to FIFO because the lot join is not implemented.

**Exact location:**
```
business/domain/inventory/inventoryitembus/stores/inventoryitemdb/inventoryitemdb.go:278
```

The comment reads:
```
/* TODO: Advanced Allocation Strategies
 * nearest_expiry: Requires joining with lot_trackings table
 *   - ORDER BY lt.expiration_date ASC
```

---

## What Exists

`QueryAvailableForAllocation` already:
- Takes `strategy string` parameter
- Applies `FOR UPDATE` pessimistic locking
- Returns `[]inventoryitembus.InventoryItem` ordered for allocation
- Falls back to `ORDER BY ii.created_date ASC` (FIFO) for all strategies

---

## What Is Missing

### Implement FEFO JOIN in the allocation query

The query needs a `LEFT JOIN inventory.lot_trackings lt ON ii.lot_id = lt.lot_id` and
conditional `ORDER BY` based on strategy:

```sql
-- FEFO strategy
SELECT ii.*
FROM inventory.inventory_items ii
LEFT JOIN inventory.lot_trackings lt ON ii.lot_id = lt.lot_id
WHERE ii.product_id = :product_id
  AND (ii.quantity - COALESCE(ii.allocated_quantity, 0)) > 0
  -- optional location/warehouse filters
ORDER BY
  CASE WHEN :strategy = 'fefo' THEN lt.expiration_date END ASC NULLS LAST,
  ii.created_date ASC
FOR UPDATE OF ii
```

Alternatively, two separate query strings branched on `strategy` value.

### Changes required

1. **`business/domain/inventory/inventoryitembus/stores/inventoryitemdb/inventoryitemdb.go`**
   - Replace the `TODO` block with the FEFO JOIN query
   - Ensure `NULL` expiry dates sort last (products without lots still allocate via FIFO)
   - Keep `FOR UPDATE` on `ii` only, not on `lt` (avoid lock contention on shared lot rows)

2. **Confirm `lot_trackings` schema has `expiration_date` column**
   Check: `business/sdk/migrate/sql/migrate.sql` or
   `business/domain/inventory/lottrackingsbus/model.go`

3. **Update integration test**
   `api/cmd/services/ichor/tests/sales/pickingapi/` (see Blocker 001)
   - Add test: seed two lots with different expiry dates, verify FEFO picks the earlier one first

---

## Key Files to Read Before Implementing

- `docs/arch/picking.md` — specifically `## ⚠ Changing the FEFO allocation query`
- `business/domain/inventory/inventoryitembus/stores/inventoryitemdb/inventoryitemdb.go:236-290`
- `business/domain/inventory/lottrackingsbus/model.go` — verify `ExpirationDate` field name
- `business/sdk/migrate/sql/migrate.sql` — verify lot_trackings schema

---

## Acceptance Criteria

- [ ] `QueryAvailableForAllocation(ctx, productID, nil, nil, "fefo", 10)` returns items sorted by lot expiry date ASC
- [ ] Items with no lot (no expiry) sort after lot-tracked items
- [ ] `strategy = "fifo"` still returns items ordered by `created_date ASC`
- [ ] No regression on existing allocation tests
- [ ] `go build ./...` passes
- [ ] `go test ./business/domain/inventory/inventoryitembus/...` passes
