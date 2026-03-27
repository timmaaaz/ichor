# Test Failure: Test_Inspections/fail-403-fail-forbidden

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/inspectionapi`
- **Duration**: 0.01s

## Failure Output

```
    apitest.go:57: fail-forbidden: Should receive a status code of 403 for the response : 401
--- FAIL: Test_Inspections/fail-403-fail-forbidden (0.01s)
```

## Fix
- **File**: `app/sdk/mid/authorize.go:74`
- **Classification**: code bug — wrong error code in AuthorizeTable middleware
- **Change**: Changed `errs.Unauthenticated` to `errs.PermissionDenied` when `hasTablePermission` returns false. User is authenticated but lacks table permission — that's 403, not 401.
- **Also fixed**: `fail_test.go` — added `GotResp`/`ExpResp`/`CmpFunc` to `fail403` for proper response validation
- **Verified**: `go test -v -run Test_Inspections/fail-403-fail-forbidden ./api/cmd/services/ichor/tests/inventory/inspectionapi/...` PASS

Novel pattern candidate: `authorize-table-401-vs-403` — AuthorizeTable middleware returns Unauthenticated instead of PermissionDenied for table permission failures. Run /distill-bugs
