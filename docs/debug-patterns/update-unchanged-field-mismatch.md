# update-unchanged-field-mismatch

**Signal**: update-200 DIFF on a field the update did NOT modify (e.g. `TrackingType`); the EXP value is a guessed constant or comes from a different/template record than the one actually updated; the fields the test *does* mutate match fine
**Root cause**: The `Update` method overwrites only fields present (non-nil) in the `Update*` struct and preserves everything else from the existing record. If the test's `ExpResp` hardcodes an unchanged field, or is built from the wrong base record, preserved fields diverge from the actual seeded row.
**Fix**:
1. Build `ExpResp` from the actual record being updated (`sd.X[n]`), overwriting only the fields the test mutates.
2. Replace any hardcoded unchanged-field literal with a reference to that seed record's field.
3. Confirm against the `Update` implementation which fields are conditionally applied (non-nil only).

**See also**: `docs/arch/domain-template.md`, `docs/arch/seeding.md`
**Examples**:
- `productapi…update-200` — hardcoded `TrackingType "none"` → `sd.Products[1].TrackingType`.
- `productbus…update-Update` — `ExpResp` built from a template product; rebuilt from the original record being updated to preserve `TrackingType` (`lot` vs `serial`).
