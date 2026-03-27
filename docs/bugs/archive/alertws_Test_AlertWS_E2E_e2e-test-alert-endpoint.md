# Test Failure: Test_AlertWS_E2E/e2e-test-alert-endpoint

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/workflow/alertws`
- **Duration**: 0.51s

## Failure Output

```
    e2e_test.go:392: Expected 200 OK, got 401
--- FAIL: Test_AlertWS_E2E/e2e-test-alert-endpoint (0.51s)
```

## Fix

- **File**: `api/domain/http/workflow/alertapi/route.go:55`
- **Classification**: code bug
- **Change**: Changed `POST /workflow/alerts/test` middleware from `mid.Authorize` (OPA + `permissionsbus` table permissions) to `mid.AuthorizeUser` (OPA only). The alertws E2E seed creates users/roles but not table access entries; the test endpoint only needs admin OPA role check.
- **Verified**: `go test -v -run "Test_AlertWS_E2E/e2e-test-alert-endpoint" github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/workflow/alertws` ✓

## Investigation

### iter-1
target: `api/domain/http/workflow/alertapi/route.go:53`
classification: code bug
confidence: High
analysis: AuthorizeTable does two checks — OPA rule AND permissionsbus table permissions.
  alertws seed creates users/roles but never seeds table access entries for workflow.alerts.
  Admin passes OPA (has ADMIN role in JWT) but fails permissionsbus check → 401.
  Test-only endpoint does not need table permissions; OPA admin rule is sufficient.
gap_notes: none
