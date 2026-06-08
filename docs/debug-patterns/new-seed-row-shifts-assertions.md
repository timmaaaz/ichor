# new-seed-row-shifts-assertions

**Signal**: query DIFF in role / userrole / permission (or any small reference-table) tests; a real seeded row appears that the test didn't expect (e.g. `FLOOR_WORKER` sorts before `Role0`); `Total` off by the number of new global rows (`13→14`, `3→4`); slice indices shifted by one
**Root cause**: A row was added to a globally-seeded reference table (e.g. the `FLOOR_WORKER` role + its `user_role` + `table_access` rows). Tests that hardcode `Total`/indices, or assume only their own test-seeded rows exist, drift — real seed rows coexist with test-seeded `Role0..RoleN`.
**Fix** (by test style):
1. **Hardcoded count/indices (apitest)**: update `Total` and slice bounds to include the new global row(s) — e.g. `Total: 13→14`, `items[5:10]→items[4:9]`.
2. **Isolatable by name (bus unittest)**: add a `Name` filter to the `QueryFilter` so only test rows return — e.g. `Name: dbtest.StringPointer("Role")` (ILIKE `%Role%`) excludes `FLOOR_WORKER`.
3. One new role usually adds rows to `user_roles` and `table_access` too — check those tests (see `table-access-count`).

**See also**: `docs/arch/seeding.md`; related `table-access-count`
**Examples**:
- FLOOR_WORKER role added to seed → `roleapi…query-200` (Total 13→14 + indices), `rolebus…query-Query` (added `Name` ILIKE filter), `userroleapi…query-200` (Total 3→4, floor_worker1 user_role).
