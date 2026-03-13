# Test Failure: Test_AlertWS_E2E

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/workflow/alertws`
- **Duration**: 14.61s

## Failure Output

```
    alertws_test.go:51: RabbitMQ Started: servicetest-rabbit at localhost:56410 (URL: amqp://guest:guest@localhost:56410/)
--- FAIL: Test_AlertWS_E2E (14.61s)
```

## Fix

- **File 1**: `api/cmd/services/ichor/tests/workflow/alertws/e2e_test.go:371,381`
- **File 2**: `api/cmd/services/ichor/tests/workflow/alertws/seed_test.go`
- **Classification**: test bug
- **Change**: The `testE2ETestAlertEndpoint` subtest used `sd.UserToken(0)` for both the WebSocket connection and the POST. The POST endpoint requires `RuleAdminOnly` + Create permission on `workflow.alerts`. Fixed by: (1) using `sd.AdminToken()` for both WS connection and POST (alert targets caller's userID — both must match), (2) adding admin user → role assignment + `TestSeedTableAccess` call in seed to grant table access.
- **Verified**: `go test -v -run Test_AlertWS_E2E/e2e-test-alert-endpoint ./api/cmd/services/ichor/tests/workflow/alertws/...` ✓
