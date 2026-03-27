# invalid-enum-check-constraint

**Signal**: `create-200-basic` or `update-200-basic` returns 500; no panic in logs; error maps to a DB constraint violation; test uses a string literal for a status/quality/type field (e.g., `QualityStatus: "poor"`, `Status: "active"`)
**Root cause**: Test uses a value that is NOT in the DB CHECK constraint for that column. The app and bus layers accept any string (no Go-side enum validation), so the error only surfaces at the DB layer as a constraint violation → 500.
**Fix**:
1. Grep the migration for the failing field: `grep -n "CHECK" business/sdk/migrate/sql/migrate.sql | grep field_name`
2. Identify the valid enum values from the CHECK constraint
3. Replace the invalid string literal in the test with a valid value

**Valid values by domain** (common ones — always verify in migrate.sql):
- `lot_trackings.quality_status`: `good`, `on_hold`, `quarantined`, `released`, `expired`
- Any field with `oneof=` in app model validation: check `validate` tag for the full list

**See also**: `docs/arch/sqldb.md`
**Examples**:
- `lottrackingsapi_Test_ProductCost_create-200-basic.md` — `QualityStatus: "poor"` and `"perfect"` used in create/update tests; valid values are `good/on_hold/quarantined/released/expired`; fixed by changing to `"good"`
- `transferorderapi_Test_InventoryLocations_update-200-basic.md` — `Status: "Adjustment"` used in update test; DB CHECK constraint requires `"pending"` (and other valid states); fixed by changing to `"pending"`
- `inspectionbus_Test_Inspections_create-Create.md` — `Status: "Pending"` (capitalized) used in create; CHECK constraint requires lowercase `"pending"`
- `inspectionbus_Test_Inspections_update-Update.md` — `Status: "In Progress"` used in update; not a valid CHECK value; fixed by changing to `"passed"`
