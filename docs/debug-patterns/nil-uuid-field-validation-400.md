# nil-uuid-field-validation-400

**Signal**: `update-200-basic` (or any update test) returns 400 instead of 200; test input uses `&sd.Entity[n].NullableField`; field is a UUID string in the app model; no obvious validation error in test code
**Root cause**: Testutil seeds nullable `*uuid.UUID` as `nil` → `ToApp*` converts nil to `""` → test sends `&sd.Entity[n].Field` which is a non-nil `*string` pointing to `""` → go-playground/validator v10 `hasValue()` returns `!IsNil()` for pointers, so `omitempty` does NOT skip a non-nil pointer to empty string → `min=36` fails → 400.
**Fix**:
1. Grep testutil for the nullable field: `grep -n "Field: nil" business/domain/{area}/{entity}bus/testutil.go`
2. Change `FieldName: nil` → `FieldName: &someValidID` using a real ID from the input slice
3. If the test `ExpResp` references `sd.Entity[n].FieldName`, also verify the expected field is populated

**Key mechanic**: validator v10 `omitempty` only skips when the pointer itself is nil. A non-nil pointer to `""` is NOT treated as empty — all subsequent tags (`min=36,max=36`) still run.

**See also**: `docs/arch/seeding.md`
**Examples**:
- `inventoryadjustmentapi_Test_InventoryLocations_update-200-basic.md` — `ApprovedBy: nil` in `inventoryadjustmentbus/testutil.go` → `""` in app model → test sent `&""` → `min=36` failed; fixed by `ApprovedBy: &approvedByID`
