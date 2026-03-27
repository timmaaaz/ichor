---
name: missing-order-by
description: SQL query missing ORDER BY yields non-deterministic row order — test DIFF failures are intermittent and seed-count-dependent
type: feedback
---

# missing-order-by

**Signal**: query-200 DIFF where row order differs between GOT and EXP; failure is intermittent or only appears at certain seed counts; the query is a `QueryByIDs`, `QueryByFilter`, or similar multi-row fetch; no `ORDER BY` in the SQL for that method
**Root cause**: PostgreSQL does not guarantee row order without an explicit `ORDER BY`. A query without one returns rows in heap scan order, which changes based on vacuums, updates, and concurrent activity. Tests that sort `expResp` but rely on the DB to return the same order will fail non-deterministically.
**Fix**:
1. Find the failing SQL query in `business/domain/{area}/{entity}bus/stores/{entity}db/{entity}db.go`
2. Add `ORDER BY {stable_column} ASC` at the end of the SELECT — typically `id`, `created_date`, or a natural key
3. If the test `CmpFunc` sorts both sides, the `ORDER BY` may be redundant but is still good practice
4. Verify that `query_test.go` for this entity has a `CmpFunc` that sorts both sides if insertion order is not guaranteed

**Common stable sort columns**:
- Primary key (`id UUID`) — always deterministic
- `created_date ASC` — good for temporal ordering
- Natural key (e.g., `product_id ASC`) — use when querying by a FK

**See also**: `docs/arch/sqldb.md`
**Examples**:
- `supplierproductapi_Test_SupplierProducts_query-by-ids-200-basic.md` — `QueryByIDs` had no `ORDER BY`; result order was non-deterministic; fixed by adding `ORDER BY product_id ASC` to the SQL in `supplierproductdb.go`
- `inventoryitemapi_Test_InventoryItem_query-200-basic.md` — SQL had `ORDER BY id ASC` (DefaultOrderBy) but `TestSeedInventoryItems` in `testutil.go` sorted seed rows by `(product_id, location_id)`; test expected slice was in wrong order; fixed by changing testutil sort to `id ASC` to match DefaultOrderBy
