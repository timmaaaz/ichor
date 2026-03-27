# Picking FEFO SQL Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.
>
> **Worktree:** Create a worktree before executing: `create a worktree for blocker-002-fefo and execute this plan`

**Goal:** Implement FEFO allocation strategy so lot-tracked inventory picks honor expiration dates instead of silently falling back to FIFO.

**Architecture:** Add a `case "fefo":` branch to the strategy switch in `QueryAvailableForAllocation` with a LEFT JOIN through `serial_numbers → lot_trackings` for expiration date ordering. No migration needed.

**Tech Stack:** Go 1.23, PostgreSQL, Ardan Labs service architecture

---

## Context

All three call sites in `pickingapp.go` (lines 149, 527, 606) pass `"fefo"` as the strategy string. Today this falls through to the `default` case and silently orders by `created_date ASC` (FIFO). The schema has no direct FK from `inventory_items` to `lot_trackings`, but the path exists through `serial_numbers` which shares `product_id` + `location_id` with `inventory_items` and has a `lot_id FK → lot_trackings`.

**Key files:**
- `business/domain/inventory/inventoryitembus/stores/inventoryitemdb/inventoryitemdb.go` — the store query (lines 236-314)
- `business/domain/inventory/inventoryitembus/inventoryitembus.go` — business method (pass-through, line 231)
- `app/domain/sales/pickingapp/pickingapp.go` — consumer (lines 149, 527, 606)

---

## Steps

### Step 1: Add FEFO case to strategy switch

- [ ] **Edit** `business/domain/inventory/inventoryitembus/stores/inventoryitemdb/inventoryitemdb.go`

Replace the strategy switch block (lines 268-276) with a version that handles `"fefo"`. The FEFO case must:
1. LEFT JOIN `inventory.serial_numbers sn` on matching `product_id` and `location_id`
2. LEFT JOIN `inventory.lot_trackings lt` on `sn.lot_id = lt.id`
3. Add `GROUP BY ii.id` + all selected columns (required because LEFT JOIN can produce duplicates when multiple serial numbers exist per inventory item)
4. Order by `MIN(lt.expiration_date) ASC NULLS LAST` then `ii.created_date ASC` as tiebreaker

Current code (lines 268-276):
```go
	// Apply ordering based on strategy
	switch strategy {
	case "fifo":
		q += " ORDER BY ii.created_date ASC"
	case "lifo":
		q += " ORDER BY ii.created_date DESC"
	default:
		q += " ORDER BY ii.created_date ASC"
	}
```

Replace with:
```go
	// Apply ordering based on strategy
	switch strategy {
	case "fefo":
		q += `
        LEFT JOIN inventory.serial_numbers sn
            ON sn.product_id = ii.product_id AND sn.location_id = ii.location_id
        LEFT JOIN inventory.lot_trackings lt
            ON lt.id = sn.lot_id
        GROUP BY ii.id, ii.product_id, ii.location_id, ii.quantity,
            ii.reserved_quantity, ii.allocated_quantity, ii.minimum_stock,
            ii.maximum_stock, ii.reorder_point, ii.economic_order_quantity,
            ii.safety_stock, ii.avg_daily_usage, ii.created_date, ii.updated_date
        ORDER BY MIN(lt.expiration_date) ASC NULLS LAST, ii.created_date ASC`
	case "lifo":
		q += " ORDER BY ii.created_date DESC"
	default:
		q += " ORDER BY ii.created_date ASC"
	}
```

Note: the `"fifo"` case is removed and merged into `default` since they are identical. The `"fefo"` case is listed first for readability since it is the primary use case in picking.

**Why GROUP BY instead of DISTINCT?** An inventory item can have multiple serial numbers, each potentially linking to a different lot with a different expiration date. We need `MIN(lt.expiration_date)` to pick the earliest expiry across all lots at that location, which requires aggregation.

**Why NULLS LAST?** Items without lot tracking (no serial numbers, or serial numbers without lots) get `NULL` for `expiration_date`. These should sort after items with known expiration dates, falling back to FIFO via the `created_date` tiebreaker.

**Why LEFT JOIN (not INNER JOIN)?** Not all inventory items are lot-tracked. An INNER JOIN would exclude non-lot-tracked items entirely, causing them to never be allocated when using FEFO strategy.

### Step 2: Remove the TODO block

- [ ] **Delete** the TODO comment block (lines 278-304) in the same file. The `nearest_expiry` strategy described there is effectively what FEFO implements. The other strategies (`lowest_cost`, `nearest_location`, `load_balancing`, `priority_zone`) are speculative and not on the roadmap. If needed later they can be planned from scratch.

### Step 3: Verify it compiles

- [ ] **Run:**
```bash
go build ./business/domain/inventory/inventoryitembus/...
```

Expected: clean build, zero errors.

### Step 4: Verify picking app compiles

- [ ] **Run:**
```bash
go build ./app/domain/sales/pickingapp/...
```

Expected: clean build. No changes needed in pickingapp since the interface and return type are unchanged.

### Step 5: Run existing inventory item tests

- [ ] **Run:**
```bash
go test ./business/domain/inventory/inventoryitembus/...
```

Expected: all existing tests pass. The change only affects `QueryAvailableForAllocation` which has no direct tests today, so no regressions expected.

### Step 6: Commit

- [ ] **Commit** with message:
```
feat(inventory): implement FEFO allocation strategy in QueryAvailableForAllocation

Add LEFT JOIN through serial_numbers → lot_trackings to order inventory
items by earliest expiration date. Items without lot tracking fall back
to FIFO ordering via NULLS LAST.
```

---

## Verification

After implementation, the FEFO query should produce SQL equivalent to:

```sql
SELECT
    ii.id, ii.product_id, ii.location_id, ii.quantity, ii.reserved_quantity,
    ii.allocated_quantity, ii.minimum_stock, ii.maximum_stock, ii.reorder_point,
    ii.economic_order_quantity, ii.safety_stock, ii.avg_daily_usage,
    ii.created_date, ii.updated_date
FROM
    inventory.inventory_items ii
LEFT JOIN inventory.serial_numbers sn
    ON sn.product_id = ii.product_id AND sn.location_id = ii.location_id
LEFT JOIN inventory.lot_trackings lt
    ON lt.id = sn.lot_id
WHERE
    ii.product_id = :product_id
    AND (ii.quantity - ii.reserved_quantity - ii.allocated_quantity) > 0
GROUP BY ii.id, ii.product_id, ii.location_id, ii.quantity,
    ii.reserved_quantity, ii.allocated_quantity, ii.minimum_stock,
    ii.maximum_stock, ii.reorder_point, ii.economic_order_quantity,
    ii.safety_stock, ii.avg_daily_usage, ii.created_date, ii.updated_date
ORDER BY MIN(lt.expiration_date) ASC NULLS LAST, ii.created_date ASC
LIMIT :limit FOR UPDATE
```

## Risks & Edge Cases

| Scenario | Behavior |
|----------|----------|
| Inventory item has no serial numbers | LEFT JOIN produces NULL expiration → sorts last (FIFO fallback) |
| Inventory item has serial numbers but lot has no expiration | `expiration_date` is `NOT NULL` in schema, so this cannot happen |
| Multiple serial numbers per item with different lots | `MIN(lt.expiration_date)` picks the earliest, which is correct for FEFO |
| `FOR UPDATE` with GROUP BY | PostgreSQL allows this — locks the rows in `ii` that contributed to the grouped result |
| Performance with large inventory | The JOIN adds cost. If this becomes a bottleneck, consider adding a materialized `earliest_expiry` column to `inventory_items` or indexing `serial_numbers(product_id, location_id)` |

## Out of Scope

- Adding a direct `lot_id` FK to `inventory_items` (schema change, broader impact)
- Other speculative strategies (lowest_cost, nearest_location, load_balancing, priority_zone)
- Integration tests for FEFO ordering (no test infrastructure exists for `QueryAvailableForAllocation` today)
