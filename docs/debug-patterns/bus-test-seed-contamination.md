# bus-test-seed-contamination

**Signal**: Bus-layer query test DIFF where GOT contains rows from a different product/entity group than EXP; EXP uses `sd.Items[0:N]` (first N seeded rows) but GOT page-1 returns items that don't belong to the same entity subset; items have different ProductID/entity FK than expected
**Root cause**: Bus-layer tests that query without an entity-specific filter (e.g., no `ProductID` filter) fetch ALL globally seeded rows for that domain. Since multiple product groups or related-entity groups are seeded, page-1 results depend on row insertion order across the whole table — not just the test's own seed set. The hardcoded `sd.Items[0:N]` slice assumes the test's items land first, which breaks when sibling seed rows interleave.
**Fix**:
1. Identify which FK/entity field uniquely scopes the rows this test owns (e.g., `ProductID`, `WarehouseID`)
2. Set that field's value from the test seed data (e.g., `p1ID := sd.Products[1].ProductID`)
3. Add the filter to the `QueryFilter` in the test: `filter.ProductID = &p1ID`
4. Replace the hardcoded `sd.Items[0:N]` expected slice with a dynamically computed slice: filter `sd.Items` where `item.ProductID == p1ID`, then take `[:pageSize]`
5. Verify the test's QueryFilter `Page`/`RowsPerPage` still makes sense after narrowing

**See also**: `docs/arch/seeding.md`, `docs/arch/testing.md`
**Examples**:
- `inventoryitembus_Test_InventoryItem_query-Query.md` — Bus query test fetched all inventory items globally; items from `Products[0]` (seeded by `TestSeedInventoryItems`) landed in page-1 alongside items for `Products[1]`; fixed by adding `ProductID: &p1ID` filter and computing `expItems` dynamically from the filtered seed slice
