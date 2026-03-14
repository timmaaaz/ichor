# table-access-count

**Signal**: `len(gotResp.TableAccess) != N. Len = M`, count mismatch in `permissionsbus_test.go`
**Root cause**: Hardcoded expected table count in `permissionsbus_test.go` is stale after a new table was added to `seed.sql`'s `core.table_access` INSERT block.
**Fix**:
1. Check what the actual count is (from GOT in failure output)
2. Update the hardcoded count in `business/domain/core/permissionsbus/permissionsbus_test.go` (search for the `len(gotResp.TableAccess) != N` assertion)
3. Run: `go test -v -run Test_Permissions/query-Query ./business/domain/core/permissionsbus/...`

**See also**: `docs/arch/seeding.md`, `docs/arch/auth.md`
**Examples**:
- `permissionsbus_Test_Permissions_query-Query.md` — `workflow.action_templates` added to table_access in seed.sql, count went from 70 to 71
- `permissionsbus_Test_Permissions_query-Query.md` (recurrence) — 2 new tables added to seed.sql, count went from 71 to 73
