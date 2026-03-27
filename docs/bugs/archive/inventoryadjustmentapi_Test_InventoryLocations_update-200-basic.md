# Test Failure: Test_InventoryLocations/update-200-basic

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/inventoryadjustmentapi`
- **Duration**: 0.01s

## Failure Output

```
    apitest.go:57: basic: Should receive a status code of 200 for the response : 400
--- FAIL: Test_InventoryLocations/update-200-basic (0.01s)
```

## Investigation

### iter-1
target: `business/domain/inventory/inventoryadjustmentbus/testutil.go:20`
classification: test bug
confidence: high
gap_notes:
- none

## Fix

- **File**: `business/domain/inventory/inventoryadjustmentbus/testutil.go:20`
- **Classification**: test bug
- **Change**: `ApprovedBy: nil` → `ApprovedBy: &approvedByID`. Nil bus pointer converts to `""` in `ToAppInventoryAdjustment`; test sends `&""` (non-nil `*string`); validator v10 `hasValue` returns `!IsNil()` for pointers so `omitempty` does NOT skip; `min=36` fails → 400.
- **Additional**: `update_test.go` ExpResp updated to include `ApprovalStatus` field.
- **Verified**: `go test -v -run Test_InventoryLocations/update-200-basic ./api/cmd/services/ichor/tests/inventory/inventoryadjustmentapi/...` ✓
