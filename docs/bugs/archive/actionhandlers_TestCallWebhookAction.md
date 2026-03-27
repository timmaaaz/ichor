# Test Failure: TestCallWebhookAction

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/workflow/actionhandlers`
- **Duration**: 4.68s

## Failure Output

```
    comms_test.go:76: Temporal Started: servicetest-temporal at localhost:56405
2026/03/13 11:42:47 INFO  No logger configured for temporal client. Created default one.
2026/03/13 11:42:47 INFO  No logger configured for temporal client. Created default one.
2026/03/13 11:42:47 INFO  No logger configured for temporal client. Created default one.
2026/03/13 11:42:47 INFO  Started Worker Namespace default TaskQueue test-workflow-notif-TestSendNotificationAction WorkerID 42587@Jakes-MacBook-Pro.local@
2026/03/13 11:42:47 INFO  Started Worker Namespace default TaskQueue test-workflow-email-TestSendEmailAction WorkerID 42587@Jakes-MacBook-Pro.local@
2026/03/13 11:42:47 INFO  Started Worker Namespace default TaskQueue test-workflow-webhook-TestCallWebhookAction WorkerID 42587@Jakes-MacBook-Pro.local@
    comms_test.go:146: creating rule: create: namedexeccontext: foreign key violation
--- FAIL: TestCallWebhookAction (4.68s)
```

## Fix

- **File**: `api/cmd/services/ichor/tests/workflow/actionhandlers/comms_test.go:119` and `:444`
- **Classification**: test bug
- **Change**: Replaced `adminID := uuid.New()` with `userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, ...)` in both `TestCallWebhookAction` and `seedCommsActionRule` helper; added `userBus *userbus.Business` parameter to helper; updated two callers to pass `db.BusDomain.User`
- **Verified**: `go test -v -run TestCallWebhookAction ./api/cmd/services/ichor/tests/workflow/actionhandlers/...` ✓

## Investigation

### iter-1
target: `comms_test.go:119` and `comms_test.go:444`
classification: test bug
confidence: high
gap_notes:
- none
detail: `adminID := uuid.New()` creates a random UUID never inserted into `core.users`. `workflow.automation_rules.created_by` has REFERENCES core.users(id). Fix: use `userbus.TestSeedUsersWithNoFKs` (same pattern as inventory_test.go:68). Same bug in `seedCommsActionRule()` helper (line 444) used by TestSendEmailAction and TestSendNotificationAction.
