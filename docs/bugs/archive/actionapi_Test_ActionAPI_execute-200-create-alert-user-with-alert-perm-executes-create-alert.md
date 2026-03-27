# Test Failure: Test_ActionAPI/execute-200-create-alert-user-with-alert-perm-executes-create-alert

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/workflow/actionapi`
- **Duration**: 0.01s

## Failure Output

```
    apitest.go:57: user-with-alert-perm-executes-create-alert: Should receive a status code of 200 for the response : 404
--- FAIL: Test_ActionAPI/execute-200-create-alert-user-with-alert-perm-executes-create-alert (0.01s)
```

## Fix

- **File**: `business/sdk/workflow/workflowactions/communication/alert.go:173` + `business/sdk/workflow/workflowactions/register.go:184` + `api/cmd/services/ichor/build/all/all.go:501`
- **Classification**: code bug
- **Change**: Added nil guard for `alertBus` in `Execute()` (graceful degradation), registered `create_alert` handler with nil buses in `RegisterCoreActions`, and upgraded with real `alertBus` in `all.go` — matching the existing pattern for `seek_approval`
- **Verified**: `go test -v -run Test_ActionAPI/execute-200-create-alert github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/workflow/actionapi` ✓
