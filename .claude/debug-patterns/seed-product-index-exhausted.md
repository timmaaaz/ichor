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
- `inventoryitemapi_Test_InventoryItem_create-200-basic.md` -- Products[0] exhausted across 25 locations; fix: use Products[2]
- `inventoryitemapi_Test_InventoryItem_update-200-basic.md` -- Products[1] also exhausted; same fix
