# seed-product-index-exhausted

**Signal**: `409`, `unique_violation`, `duplicate key`, `already exists`, inventory item create test, `Products[0]`, `Products[1]`, `InventoryLocations[0]`
**Root cause**: Seed creates N items using `(loc[i%L], prod[i/L])` combinatorics; low-index products fully saturated across all locations. Test picks `Products[0]` for create -> conflict.
**Fix**:
1. Count seed items (N), locations (L), products (P) in `{entity}bus/testutil.go`
2. Compute saturation threshold: `ceil(N/L)` products are fully used
3. In test, use `Products[idx]` where `idx >= ceil(N/L)`
4. If deterministic ordering matters, verify `testutil.go` ORDER BY clause includes all unique-constraint columns (e.g., `product_id, location_id`)

**See also**: `docs/arch/seeding.md`
**Examples**:
- `inventoryitemapi_Test_InventoryItem_create-200-basic.md` — Products[0] exhausted across 25 locations; fixed by using Products[2]
- `inventoryitemapi_Test_InventoryItem_update-200-basic.md` — Products[2] also exhausted when both create-200 and update-200 subtests used `(Products[2], InventoryLocations[0])`; fixed by using Products[3]
- `productbus_Test_Product_create-Create.md` — UPC/SKU UNIQUE constraint hit at bus layer; create test randomly picked a seeded product whose UpcCode already existed; fixed by using a fresh `NewProduct` with index `99999` (well outside seeded range 1–10020)
- `productbus_Test_Product_update-Update.md` — Same UNIQUE constraint (`products_upc_code_unique`) in update test; update values were copied from an existing seeded row still in DB; fixed by generating fresh unique `UpdatedSKU%d`, `UpdatedUpc%d`, `UpdatedModel%d` strings
