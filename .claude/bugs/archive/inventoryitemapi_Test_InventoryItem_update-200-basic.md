# Test Failure: Test_InventoryItem/update-200-basic

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/inventoryitemapi`
- **Duration**: 0.01s

## Failure Output

```
    apitest.go:57: basic: Should receive a status code of 200 for the response : 409
--- FAIL: Test_InventoryItem/update-200-basic (0.01s)
```

## Fix

- **File 1**: `business/domain/inventory/inventoryitembus/testutil.go` — changed sort from UUID to `(product_id, location_id)` for deterministic grid-order output
- **File 2**: `api/cmd/services/ichor/tests/inventory/inventoryitemapi/update_test.go:24,39` — changed `Products[1]` → `Products[2]` (products[0] and [1] fill all 50 seeded slots; products[2] is guaranteed free)
- **Classification**: test bug — update target collided with an already-seeded `(product_id, location_id)` pair, triggering the unique constraint → 409
- **Verified**: `go test -v -run Test_InventoryItem/update-200-basic github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/inventoryitemapi` ✓
