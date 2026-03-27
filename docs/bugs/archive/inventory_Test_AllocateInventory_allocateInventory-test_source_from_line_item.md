# Test Failure: Test_AllocateInventory/allocateInventory-test_source_from_line_item

- **Package**: `github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory`
- **Duration**: 0.01s

## Failure Output

```
    unittest.go:17: DIFF
    unittest.go:18: got expected QueuedAllocationResponse, got map[string]interface {}, want success
    unittest.go:19: GOT
    unittest.go:20: &errors.errorString{s:"expected QueuedAllocationResponse, got map[string]interface {}"}
    unittest.go:21: EXP
    unittest.go:22: "success"
    unittest.go:23: Should get the expected response
--- FAIL: Test_AllocateInventory/allocateInventory-test_source_from_line_item (0.01s)
```

## Fix

- **File**: `business/sdk/workflow/workflowactions/inventory/allocate.go` (return type) + `allocate_test.go` (assertions)
- **Classification**: stale bug — already resolved
- **Change**: Commit `ead95fc3` changed `Execute()` to return `map[string]any` (hybrid output ports) and updated test assertions to match. Bug was recorded before that commit landed.
- **Verified**: `go test -v -run Test_AllocateInventory/allocateInventory-test_source_from_line_item ./business/sdk/workflow/workflowactions/inventory/...` ✓ PASS
