# Test Failure: TestSendNotificationAction

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/workflow/actionhandlers`
- **Duration**: 4.68s

## Failure Output

```
    comms_test.go:376: creating rule: create: namedexeccontext: foreign key violation
2026/03/13 11:42:47 INFO  Stopped Worker Namespace default TaskQueue test-workflow-webhook-TestCallWebhookAction WorkerID 42587@Jakes-MacBook-Pro.local@
--- FAIL: TestSendNotificationAction (4.68s)
```

## Investigation

### iter-1
target: `comms_test.go:381`
classification: test bug
confidence: high
gap_notes:
- none

The original FK violation was already fixed (seedCommsActionRule now seeds a real user).
The actual failure was: `recipients` field in the config JSON was an object `{"users":[...],"roles":[]}`
instead of the handler's expected `[]string`. Also `priority` was missing.

## Fix

- **File**: `api/cmd/services/ichor/tests/workflow/actionhandlers/comms_test.go:381`
- **Classification**: test bug
- **Change**: Changed `send_notification` config JSON `recipients` from object form `{"users":[...],"roles":[]}` to flat `[]string`, and added required `priority` field.
- **Verified**: `go test -v -run TestSendNotificationAction ./api/cmd/services/ichor/tests/workflow/actionhandlers/...` ✓
