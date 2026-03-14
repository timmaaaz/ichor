# create-missing-querybyid-join-fields

**Signal**: `create-200-basic` returns 200 but DIFF shows UUID fields as `"00000000-0000-0000-0000-000000000000"` in GOT; or JOIN-enriched fields (e.g., ProductName, ProductSKU, LocationCode) are empty/nil in response but populated in DB; entity has fields sourced from a JOIN in the query view
**Root cause**: `App.Create` calls `bus.Create` and returns the result directly. The bus `Create` only sets fields present in the INSERT — JOIN-enriched fields (populated via a view or JOIN in `Query`/`QueryByID`) are absent. `uuid.Nil.String()` → `"00000000-0000-0000-0000-000000000000"` for unpopulated UUID fields.
**Fix**:
1. In `app/domain/{area}/{entity}app/{entity}app.go`, find the `Create` method
2. After `bus.Create(ctx, newEntity)`, add: `entity, err = a.{entity}Bus.QueryByID(ctx, entity.ID)` (re-fetch with full JOIN)
3. Return the re-fetched entity instead of the insert result
4. In the test `CmpFunc`, add `cmpopts.IgnoreFields` for any JOIN fields whose values are random seed data (ProductID, ProductName, etc.)

**See also**: `docs/arch/domain-template.md`, `docs/arch/sqldb.md`
**Examples**:
- `lottrackingsapi_Test_ProductCost_create-200-basic.md` — `App.Create` returned partial struct; `ProductID`/`ProductName`/`ProductSKU` were `uuid.Nil.String()` because they come from a JOIN; fixed by adding `QueryByID` after create in `lottrackingsapp.go:41`
- `lottrackingsapi_Test_ProductCost_update-200-basic.md` — same JOIN fields (`ProductID`, `ProductName`, `ProductSKU`) were not copied in the update test `CmpFunc`; fix: add those fields to the CmpFunc copy-from-got block so the test doesn't assert on seed-derived JOIN data
