# Test Failure: Test_InventoryItem/create-200-basic

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/inventoryitemapi`
- **Duration**: 0.02s

## Fix

- **File**: `api/cmd/services/ichor/tests/inventory/inventoryitemapi/create_test.go:23,37`
- **Classification**: test bug
- **Change**: Changed `Products[0]` → `Products[2]` in both Input and ExpResp; seed exhausts products[0] and products[1] across all 25 locations (50 items, formula `loc[i%25], prod[i/25]`), so index ≥ 2 is safe
- **Verified**: `go test -v -run Test_InventoryItem/create-200-basic github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/inventoryitemapi` ✓

## Failure Output

```
    apitest.go:57: basic: Should receive a status code of 200 for the response : 409
--- FAIL: Test_InventoryItem/create-200-basic (0.02s)
```
