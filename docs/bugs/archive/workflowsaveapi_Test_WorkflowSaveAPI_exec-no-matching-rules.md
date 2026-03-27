# Test Failure: Test_WorkflowSaveAPI/exec-no-matching-rules

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/workflow/workflowsaveapi`
- **Duration**: 0s

## Failure Output

```
    execution_test.go:251: unexpected error for no-match event: process event: invalid trigger event: Timestamp is required
--- FAIL: Test_WorkflowSaveAPI/exec-no-matching-rules (0.00s)
```

## Fix
- **File**: `api/cmd/services/ichor/tests/workflow/workflowsaveapi/execution_test.go:248`
- **Classification**: test bug
- **Change**: Added `Timestamp: time.Now()` to `TriggerEvent` construction — `trigger.go:175-177` validates `Timestamp` is non-zero but test omitted it.
- **Verified**: `go test -v -run "Test_WorkflowSaveAPI/exec-no-matching-rules" github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/workflow/workflowsaveapi` ✓
