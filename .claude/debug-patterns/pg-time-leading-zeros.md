# pg-time-leading-zeros

**Signal**: DIFF shows time value like `"8:00:00"` in EXP but `"08:00:00"` in GOT; field name often `AvailableHoursStart`, `AvailableHoursEnd`, or any `time` / `timetz` column
**Root cause**: PostgreSQL always returns TIME values in canonical `HH:MM:SS` format with leading zeros (e.g., `08:00:00`). Test expectation was written with a short form like `"8:00:00"` which never matches.
**Fix**:
1. Find the test file asserting the time value (search for the short-form string)
2. Update the expected string to use zero-padded hours: `"8:00:00"` → `"08:00:00"`
3. Apply to both start and end times if both are present

**See also**: `docs/arch/seeding.md`
**Examples**:
- `contactinfsoapi_Test_ContactInfos_create-200-basic.md` — `AvailableHoursStart: "8:00:00"` → `"08:00:00"` in `create_test.go:40`
