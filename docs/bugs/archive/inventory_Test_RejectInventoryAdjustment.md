# Test Failure: Test_RejectInventoryAdjustment

- **Package**: `github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory`
- **Duration**: 1.37s

## Failure Output

```
    reject_adjustment_test.go:23: seeding: creating pre-approve adjustment: create: invalid reason code
--- FAIL: Test_RejectInventoryAdjustment (1.37s)
```

## Fix
- **File**: `business/sdk/workflow/workflowactions/inventory/approve_adjustment_test.go:209,227`
- **Classification**: test bug
- **Change**: Same file as Approve bug — shared seed function used `"correction"` and `"damage"` instead of valid reason codes (`"data_entry_error"`, `"damaged"`)
- **Verified**: `go test -v -run Test_RejectInventoryAdjustment ./business/sdk/workflow/workflowactions/inventory/...` ✓
- **pattern-match**: `invalid-enum-check-constraint`
