# Test Failure: Test_Inspections/fail-200-no-quarantine-fail-without-quarantine

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/inspectionapi`
- **Duration**: 0.01s

## Failure Output

```
    apitest.go:57: fail-without-quarantine: Should receive a status code of 200 for the response : 404
--- FAIL: Test_Inspections/fail-200-no-quarantine-fail-without-quarantine (0.01s)
```

## Fix
- **File**: `api/cmd/services/ichor/tests/inventory/inspectionapi/fail_test.go:44`
- **Classification**: test bug
- **Change**: Changed `sd.Inspections[1]` to `sd.Inspections[3]` — index 1 was deleted by `delete200()` which runs earlier in the sequential test suite
- **Verified**: `go test -v -run Test_Inspections/fail-200-no-quarantine-fail-without-quarantine ./api/cmd/services/ichor/tests/inventory/inspectionapi/...` ✓
