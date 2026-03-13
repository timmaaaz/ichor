# omitempty-to-required

**Signal**: create-400 test for "missing-X" returns 200 instead of 400; the field X is optional in the request but the test expects it to be required
**Root cause**: The `NewEntity` struct in the app layer has `validate:"omitempty"` on a field that should be `validate:"required"`. The validator accepts the request even when the field is absent.
**Source of truth**: The migration SQL in `business/sdk/migrate/sql/migrate.sql`. `NOT NULL` column → `validate:"required"`. Nullable column → `validate:"omitempty"`. Check the schema before deciding.

**Fix**:
1. Find the `New{Entity}` struct in `app/domain/{area}/{entity}app/model.go`
2. Check `migrate.sql` for the column nullability
3. Change `validate:"omitempty"` → `validate:"required"` for `NOT NULL` columns
4. Run all create-400 subtests for the entity at once — they often share the same root cause

**Side effect**: If seed data uses `nil` for a field now marked `required`, `sd.{Entity}[n].{Field}` will be `""` and break other tests that pass it as input. Fix by updating `testutil.go` to seed a real value.

**See also**: `docs/arch/domain-template.md`
**Examples**:
- `inventorylocationapi_Test_InventoryLocations_create-400-missing-warehouse-id.md` (and 4 related bugs) — `WarehouseID`, `IsPickLocation`, `IsReserveLocation`, `MaxCapacity`, `CurrentUtilization` all had `omitempty` in `app/domain/inventory/inventorylocationapp/model.go`
- `inventoryadjustmentapi_Test_InventoryLocations_create-400-missing-approved-by.md` + `missing-notes` — both `approved_by` and `notes` are `NOT NULL` in schema but had `omitempty`; seed also needed `ApprovedBy` populated so other tests using `sd.InventoryAdjustments[n].ApprovedBy` didn't submit `""`
