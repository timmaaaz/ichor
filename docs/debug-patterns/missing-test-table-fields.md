# missing-test-table-fields

**Signal**: `json: Unmarshal(nil)`, test table entry returns unexpected nil response, API test framework panics or gets zero-value struct; test table row missing `GotResp`, `ExpResp`, or `CmpFunc` fields
**Root cause**: When adding a new test case to an API test table, one or more required struct fields (GotResp, ExpResp, CmpFunc) were omitted. The test framework attempts to unmarshal the response into a nil pointer, causing a panic or nonsensical comparison.
**Fix**:
1. Find the failing test table entry in the test file
2. Add the missing fields — typically:
   - `GotResp: &model.Entity{}`
   - `ExpResp: &expectedValue`
   - `CmpFunc: func(got, exp any) string { ... }` (or reference a shared cmp function)
3. Run the test to verify

**See also**: `docs/arch/testing.md`
**Examples**:
- `inspectionapi_Test_Inspections_fail-404-fail-not-found.md` — test table entry for fail-not-found case was missing GotResp, ExpResp, and CmpFunc fields; json.Unmarshal received nil target; fixed by adding all three fields
