# Test Failure: Test_InventoryLocations/create-400-missing-approved-by

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/inventoryadjustmentapi`
- **Duration**: 0.01s

## Failure Output

```
    apitest.go:57: missing-approved-by: Should receive a status code of 400 for the response : 200
--- FAIL: Test_InventoryLocations/create-400-missing-approved-by (0.01s)
```

## Fix

- **File**: `app/domain/inventory/inventoryadjustmentapp/model.go:102`
- **Classification**: code bug
- **Change**: Changed `approved_by` and `notes` from `validate:"omitempty,..."` to `validate:"required,..."` in `NewInventoryAdjustment`; updated `testutil.go` to seed a real `ApprovedBy` UUID; added `ApprovalStatus` to `update_test.go` ExpResp
- **Verified**: `go test -run Test_InventoryLocations github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/inventoryadjustmentapi` ✓
- **pattern-match**: omitempty-to-required
