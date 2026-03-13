# Test Failure: Test_AllocateInventory/allocateInventory-test_order_grouped_allocation

- **Package**: `github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory`
- **Duration**: 0.01s

## Failure Output

```
    unittest.go:17: DIFF
    unittest.go:18: order grouping failed: got expected QueuedAllocationResponse for A1, got map[string]interface {}, want true
    unittest.go:19: GOT
    unittest.go:20: &errors.errorString{s:"expected QueuedAllocationResponse for A1, got map[string]interface {}"}
    unittest.go:21: EXP
    unittest.go:22: true
    unittest.go:23: Should get the expected response
--- FAIL: Test_AllocateInventory/allocateInventory-test_order_grouped_allocation (0.01s)
```

## Fix

- **File**: `business/sdk/workflow/workflowactions/inventory/allocate_test.go:699,776,815`
- **Classification**: test bug
- **Change**: PR #74 changed `Execute()` to return `map[string]any` (required for Temporal serialization + MergedContext routing). Updated 3 type assertions from `inventory.QueuedAllocationResponse` to `map[string]any` and field accesses from `.ReferenceID`/`.Status` to `["reference_id"].(string)`/`["status"].(string)`
- **Verified**: `go test -v -run Test_AllocateInventory/allocateInventory-test_order_grouped_allocation ./business/sdk/workflow/workflowactions/inventory/...` ✓
