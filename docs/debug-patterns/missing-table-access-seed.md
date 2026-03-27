# missing-table-access-seed

**Signal**: 401 on ALL endpoints for one entity (create, query, query-by-id, update, delete — everything 401); entity is newly added; other entities work fine
**Root cause**: New entity's table name is missing from `core.table_access` seeding in two places: `seed.sql` (for live/staging DB) and `tableaccessbus/testutil.go` (for test DB). Without a table_access row, the auth middleware rejects all requests.
**Fix**:
1. Add `'schema.table_name'` to the table_access INSERT in `business/sdk/migrate/sql/seed.sql`
2. Add `'schema.table_name'` to `TestSeedTableAccess` in `business/domain/core/tableaccessbus/testutil.go`
3. Run the full entity test suite (not just one subtest — all will be 401)

Example seed.sql addition (inside the INSERT INTO core.table_access block):
```sql
-- Add alongside other tables in the same schema
```

Example testutil.go addition (inside the `TestSeedTableAccess` function, in the `tableNames` slice or equivalent):
```go
"schema.table_name",
```

**See also**: `docs/arch/seeding.md`, `docs/arch/auth.md`
**Examples**:
- `putawaytaskapi_Test_PutAwayTask_create-200-basic.md` — `inventory.put_away_tasks` missing from both seed files; caused 19 cascading 401 failures across the entire putawaytaskapi test suite
- `alertws_Test_AlertWS_E2E.md` — alertws seed was missing admin user → role assignment + TestSeedTableAccess call, causing 401 on test alert endpoint
