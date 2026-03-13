# Test Failure: Test_ActionAPI/list-actions-200-user-with-permissions-user-sees-permitted-actions-only

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/workflow/actionapi`
- **Duration**: 0.01s

## Failure Output

```
    apitest.go:73: DIFF
    apitest.go:74:   &actionapp.AvailableActions{
        - 	{
        - 		Type:        "create_alert",
        - 		Description: "Create an alert notification for users or roles",
        - 	},
          }
    apitest.go:75: GOT
    apitest.go:76: &actionapp.AvailableActions{actionapp.AvailableAction{Type:"send_notification", Description:"Send an in-app notification", IsAsync:false}, actionapp.AvailableAction{Type:"create_alert", Description:"Create an alert notification for users or roles", IsAsync:false}}
    apitest.go:77: EXP
    apitest.go:78: &actionapp.AvailableActions{actionapp.AvailableAction{Type:"send_notification", Description:"Send an in-app notification", IsAsync:false}}
    apitest.go:79: Should get the expected response
--- FAIL: Test_ActionAPI/list-actions-200-user-with-permissions-user-sees-permitted-actions-only (0.01s)
```

## Fix
- **File**: `api/cmd/services/ichor/tests/workflow/actionapi/list_test.go:42-44`
- **Classification**: test bug
- **Change**: Added `create_alert` to expected actions — `RegisterCoreActions` registers both `send_notification` and `create_alert` (with nil buses), both return `SupportsManualExecution()=true`, and the user has permissions for both. The test comment was wrong.
- **Verified**: `go test -v -run Test_ActionAPI/list-actions-200-user-with-permissions-user-sees-permitted-actions-only ./api/cmd/services/ichor/tests/workflow/actionapi/...` ✓
