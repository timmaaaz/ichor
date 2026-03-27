# reference-category-count-mismatch

**Signal**: `expected N types, got N+M`, `Category "X": expected N types`, category consistency test fails; test hardcodes expected count of action types per category
**Root cause**: Reference API tests assert a fixed number of action types per category (e.g., "inventory" should have 11 types). When a new action type is registered in a category, the hardcoded count goes stale. Similar to `table-access-count` but for the action type registry instead of table_access.
**Fix**:
1. Check which new action types were added (compare GOT vs EXP in failure output)
2. Add the missing action type(s) to the expected list in the test file
3. Update the expected count
4. Run: `go test -v -run Test_ActionTypeSchemas ./api/cmd/services/ichor/tests/workflow/referenceapi/...`

**See also**: `docs/arch/workflow-engine.md`
**Examples**:
- `referenceapi_Test_ActionTypeSchemas_CategoryConsistency.md` — `create_put_away_task` action type added to inventory category; expected count went from 11 to 12; fixed by adding to expected types list
