# Test Failure: Test_InventoryLocations/create-400-missing-current-utilization

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/inventorylocationapi`
- **Duration**: 0.01s

## Failure Output

```
    apitest.go:57: missing-current-utilization: Should receive a status code of 400 for the response : 200
--- FAIL: Test_InventoryLocations/create-400-missing-current-utilization (0.01s)
```

## Fix

- **File**: `app/domain/inventory/inventorylocationapp/model.go:93,100-103`
- **Classification**: code bug
- **Change**: Changed `omitempty` to `required` on `WarehouseID`, `IsPickLocation`, `IsReserveLocation`, `MaxCapacity`, `CurrentUtilization` in `NewInventoryLocation` struct
- **Verified**: `go test -v -run Test_InventoryLocations github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/inventorylocationapi` ✓
