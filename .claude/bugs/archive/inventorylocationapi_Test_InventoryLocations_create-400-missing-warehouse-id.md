# Test Failure: Test_InventoryLocations/create-400-missing-warehouse-id

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/inventorylocationapi`
- **Duration**: 0.01s

## Failure Output

```
    apitest.go:57: missing-warehouse-id: Should receive a status code of 400 for the response : 200
--- FAIL: Test_InventoryLocations/create-400-missing-warehouse-id (0.01s)
```

## Fix

- **File**: `app/domain/inventory/inventorylocationapp/model.go:93-103`
- **Classification**: code bug
- **Change**: Changed `validate:"omitempty"` → `validate:"required"` for `WarehouseID`, `IsPickLocation`, `IsReserveLocation`, `MaxCapacity`, `CurrentUtilization` on `NewInventoryLocation`. Applies to all 5 related bugs.
- **Verified**: `go test -v -run "Test_InventoryLocations/create-400" ./api/cmd/services/ichor/tests/inventory/inventorylocationapi/...` — all 10 subtests PASS ✓
