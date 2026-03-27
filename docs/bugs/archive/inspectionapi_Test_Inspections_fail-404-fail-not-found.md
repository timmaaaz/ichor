# Test Failure: Test_Inspections/fail-404-fail-not-found

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/inspectionapi`
- **Duration**: 0.01s

## Failure Output

```
    apitest.go:68: Should be able to unmarshal the response : json: Unmarshal(nil)
--- FAIL: Test_Inspections/fail-404-fail-not-found (0.01s)
```

## Fix
- **File**: `api/cmd/services/ichor/tests/inventory/inspectionapi/fail_test.go:82-100`
- **Classification**: test bug
- **Change**: Added missing `GotResp`, `ExpResp`, and `CmpFunc` fields to `fail404()` test table entry — the apitest harness requires `GotResp` as the unmarshal target
- **Verified**: `go test -v -run Test_Inspections/fail-404-fail-not-found ./api/cmd/services/ichor/tests/inventory/inspectionapi/...` ✓
