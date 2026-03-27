# Test Failure: Test_InventoryLocations/create-400-missing-approved-by-id

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/transferorderapi`
- **Duration**: 0.01s

## Failure Output

```
    apitest.go:57: missing-approved-by-id: Should receive a status code of 400 for the response : 200
--- FAIL: Test_InventoryLocations/create-400-missing-approved-by-id (0.01s)
```

## Fix
- **File**: `api/cmd/services/ichor/tests/inventory/transferorderapi/create_test.go:156`
- **Classification**: test bug
- **Change**: Removed `missing-approved-by-id` subtest — `approved_by` was intentionally made nullable by migration ("Make transfer_orders.approved_by nullable to support pending-approval workflow"), so omitting it is valid. Model correctly uses `validate:"omitempty"`.
- **Verified**: `go test -v -run "Test_InventoryLocations/create-400" github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/transferorderapi` ✓
- **pattern-match**: omitempty-to-required (inverted — schema IS nullable, test was wrong to assert required)
