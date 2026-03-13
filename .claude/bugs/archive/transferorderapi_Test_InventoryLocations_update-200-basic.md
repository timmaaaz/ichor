# Test Failure: Test_InventoryLocations/update-200-basic

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/transferorderapi`
- **Duration**: 0.01s

## Failure Output

```
    apitest.go:57: basic: Should receive a status code of 200 for the response : 400
--- FAIL: Test_InventoryLocations/update-200-basic (0.01s)
```

## Fix
- **File**: `api/cmd/services/ichor/tests/inventory/transferorderapi/update_test.go:35,46`
- **Classification**: test bug
- **Change**: Changed `Status: "Adjustment"` to `Status: "pending"` (both in input and ExpResp); "Adjustment" is not in the app-layer `oneof=pending approved in_progress completed cancelled` validator for `UpdateTransferOrder.Status`
- **pattern-match**: nil-uuid-field-validation-400 (adjacent pattern — same 400 symptom but different cause: invalid oneof value vs nil UUID)
- **Verified**: `go test -v -run Test_InventoryLocations/update-200-basic ./api/cmd/services/ichor/tests/inventory/transferorderapi/...` ✓
