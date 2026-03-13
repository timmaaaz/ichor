# Test Failure: Test_PutAwayTask/create-200-basic

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/putawaytaskapi`
- **Duration**: 0.01s

## Failure Output

```
    apitest.go:57: basic: Should receive a status code of 200 for the response : 401
--- FAIL: Test_PutAwayTask/create-200-basic (0.01s)
```

## Fix

- **Root cause**: `inventory.put_away_tasks` missing from `core.table_access` seed in two places
- **Classification**: mixed (permission seeding gap + test bugs in same package)
- **Changes**:
  1. `business/sdk/migrate/sql/seed.sql` — added `inventory.put_away_tasks` to table_access INSERT block (live/staging DB seed)
  2. `business/domain/core/tableaccessbus/testutil.go` — added `inventory.put_away_tasks` to TestSeedTableAccess (test DB seed)
  3. `business/domain/inventory/putawaytaskbus/putawaytaskbus.go` — QueryByID: return ErrNotFound directly without wrapping (code bug; matched inventorylocationbus pattern)
  4. `api/cmd/services/ichor/tests/inventory/putawaytaskapi/query_test.go` — fix query-200-all to use `query.Result[PutAwayTask]` with Page/RowsPerPage (test bug)
  5. `api/cmd/services/ichor/tests/inventory/putawaytaskapi/create_test.go` — fix create-409 expected error to include full chain "create: namedexeccontext: foreign key violation" (test bug)
- **Verified**: `go test -count=1 -p 1 -run "Test_PutAwayTask|TestUpdate200Complete|TestUpdate400TerminalState" ./api/cmd/services/ichor/tests/inventory/putawaytaskapi/...` — 27/27 PASS ✓
