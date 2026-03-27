# Test Failure: Test_InventoryItem/update-200-basic

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/inventoryitemapi`
- **Duration**: 0.01s

## Failure Output

```
    apitest.go:57: basic: Should receive a status code of 200 for the response : 409
--- FAIL: Test_InventoryItem/update-200-basic (0.01s)
```

## Fix
- **File**: `api/cmd/services/ichor/tests/inventory/inventoryitemapi/update_test.go:24,39`
- **Classification**: test bug
- **Change**: Changed `Products[2]` → `Products[3]` in both Input and ExpResp; `create-200` and `update-200` both used `(Products[2], InventoryLocations[0])` causing a unique constraint violation when update ran after create in the same test suite
- **Verified**: `go test -v -run Test_InventoryItem/update-200-basic github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/inventoryitemapi` ✓
- **pattern-match**: seed-product-index-exhausted
