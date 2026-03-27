# Test Failure: Test_LotLocation

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/lotlocationapi`
- **Duration**: 3.07s

## Failure Output

```
    lotlocation_test.go:16: Seeding error: seeding lot locations : seeding error: create: namedexeccontext: lot location entry is not unique
--- FAIL: Test_LotLocation (3.07s)
```

## Fix

- **Files**:
  - `api/cmd/services/ichor/tests/inventory/lotlocationapi/seed_test.go:209` — changed `n=15` to `n=10`
  - `api/cmd/services/ichor/tests/inventory/lotlocationapi/query_test.go:22` — changed `Total: 15` to `Total: 10`
  - `api/cmd/services/ichor/tests/inventory/lotlocationapi/create_test.go:22-29` — changed `InventoryLocations[0]` to `InventoryLocations[1]`
  - `business/sdk/migrate/sql/seed.sql` — added `inventory.lot_locations` to table_access INSERT
  - `business/domain/core/tableaccessbus/testutil.go` — added `inventory.lot_locations` entry
- **Classification**: test bug (3 compounding issues: seed count > unique pairs, missing table_access, create-200 collision with seeded pair)
- **Verified**: `go test -v -run Test_LotLocation github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/lotlocationapi` — all 27 subtests PASS ✓
