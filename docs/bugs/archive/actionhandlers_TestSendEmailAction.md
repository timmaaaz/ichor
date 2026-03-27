# Test Failure: TestSendEmailAction

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/workflow/actionhandlers`
- **Duration**: 4.67s

## Failure Output

```
    comms_test.go:270: creating rule: create: namedexeccontext: foreign key violation
2026/03/13 11:42:47 INFO  Stopped Worker Namespace default TaskQueue test-workflow-email-TestSendEmailAction WorkerID 42587@Jakes-MacBook-Pro.local@
2026/03/13 11:42:47 INFO  Stopped Worker Namespace default TaskQueue test-workflow-notif-TestSendNotificationAction WorkerID 42587@Jakes-MacBook-Pro.local@
--- FAIL: TestSendEmailAction (4.67s)
```

## Fix

- **File**: `api/cmd/services/ichor/tests/workflow/actionhandlers/comms_test.go`
- **Classification**: test bug
- **Change**: Replaced `uuid.New()` (random UUID not in DB) with `userbus.TestSeedUsersWithNoFKs()` so `CreatedBy` references a real `core.users` row, satisfying the FK constraint on `workflow.automation_rules.created_by`
- **Verified**: `go test -v -run TestSendEmailAction github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/workflow/actionhandlers` ✓
