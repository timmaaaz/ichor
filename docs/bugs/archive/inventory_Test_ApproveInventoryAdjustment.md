# Test Failure: Test_ApproveInventoryAdjustment

- **Package**: `github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory`
- **Duration**: 1.46s

## Failure Output

```
    approve_adjustment_test.go:38: seeding: creating pre-approve adjustment: create: invalid reason code
--- FAIL: Test_ApproveInventoryAdjustment (1.46s)
```

## Fix
- **File**: `business/sdk/workflow/workflowactions/inventory/approve_adjustment_test.go:209,227`
- **Classification**: test bug
- **Change**: Changed `"correction"` to `"data_entry_error"` and `"damage"` to `"damaged"` — must match valid reason codes in `inventoryadjustmentbus.ValidReasonCodes`
- **Verified**: `go test -v -run Test_ApproveInventoryAdjustment ./business/sdk/workflow/workflowactions/inventory/...` ✓
- **pattern-match**: `invalid-enum-check-constraint`
