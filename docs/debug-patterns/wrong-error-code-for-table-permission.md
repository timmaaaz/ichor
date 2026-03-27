# wrong-error-code-for-table-permission

**Signal**: API test expects 403 but gets 401; endpoint uses table-level permission check; `authorize.go` middleware; `errs.Unauthenticated` vs `errs.PermissionDenied`
**Root cause**: The `authorize` middleware in `app/sdk/mid/authorize.go` returns `errs.Unauthenticated` (401) when a user lacks table-level access, but the correct HTTP semantics for "authenticated but not authorized" is `errs.PermissionDenied` (403). This is a code bug, not a test bug.
**Fix**:
1. Open `app/sdk/mid/authorize.go`
2. Find the table permission denial branch (where table access check fails)
3. Change `errs.Unauthenticated` to `errs.PermissionDenied`
4. Run affected API fail tests: `go test -v -run Test_.*_fail-403 ./api/cmd/services/ichor/tests/...`

**See also**: `docs/arch/auth.md`, `docs/arch/errs.md`
**Examples**:
- `inspectionapi_Test_Inspections_fail-403-fail-forbidden.md` — table permission denial returned 401 instead of 403; fixed by changing errs.Unauthenticated to errs.PermissionDenied in authorize.go
