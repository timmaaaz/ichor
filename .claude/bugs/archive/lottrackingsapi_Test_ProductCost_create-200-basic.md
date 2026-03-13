# Test Failure: Test_ProductCost/create-200-basic

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/lottrackingsapi`
- **Duration**: 0.02s

## Failure Output

```
    apitest.go:57: basic: Should receive a status code of 200 for the response : 500
--- FAIL: Test_ProductCost/create-200-basic (0.02s)
```

## Investigation

### iter-1
target: `create_test.go:38` — QualityStatus: "poor" violates DB CHECK constraint
classification: test bug (initial) → code bug (revealed after fix)
confidence: high

Root cause (layered):
1. `QualityStatus: "poor"` in create200 test fails DB CHECK constraint (valid: good/on_hold/quarantined/released/expired) → 500
2. After fixing invalid enum, revealed `App.Create` (lottrackingsapp.go:41) returns partial struct without JOIN-enriched ProductID/ProductName/ProductSKU — `uuid.Nil.String()` = "00000000-0000-0000-0000-000000000000" vs expected ""
gap_notes: none (high confidence, 1 iteration for core fix; second fix was expected pattern)

## Fix

- **File 1**: `api/cmd/services/ichor/tests/inventory/lottrackingsapi/create_test.go:38,48,353` — Changed `QualityStatus: "poor"` → `"good"` (valid CHECK constraint value)
- **File 2**: `app/domain/inventory/lottrackingsapp/lottrackingsapp.go:41` — Added `QueryByID` after `Create` to populate JOIN-enriched fields (ProductID, ProductName, ProductSKU)
- **File 3**: `create_test.go:61` — Added `cmpopts.IgnoreFields` for ProductID/ProductName/ProductSKU in CmpFunc (JOIN values are random seed data)
- **File 4**: `update_test.go` — Changed `"perfect"` → `"good"` for QualityStatus (also invalid CHECK value; separate bug `update-200-basic` remains)
- **Classification**: Test bug (invalid enum) + Code bug (missing QueryByID in App.Create)
- **Verified**: `go test -v -run Test_ProductCost/create-200-basic github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/lottrackingsapi` ✓
