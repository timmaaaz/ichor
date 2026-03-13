# Test Failure: Test_InventoryProduct/update-200-basic

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/products/productapi`
- **Duration**: 0.02s

## Failure Output

```
    apitest.go:73: DIFF
    apitest.go:74:   &productapp.Product{
          	... // 11 identical fields
          	HandlingInstructions: "test handling instructions",
          	UnitsPerCase:         "20",
        - 	TrackingType:         "none",
        + 	TrackingType:         "",
          	InventoryType:        "",
          	CreatedDate:          "2026-03-13 11:41:42 +0000 UTC",
          	UpdatedDate:          "2026-03-13 11:41:44 +0000 UTC",
          }
    apitest.go:75: GOT
    apitest.go:76: &productapp.Product{ProductID:"e25c16be-a454-41bf-aea0-06613157e5cf", SKU:"sku123", BrandID:"5002ccc6-0c6d-45cd-a25c-f1ca34207823", ProductCategoryID:"ca445934-8a5e-46d7-8633-291798d7857a", Name:"Test Product", Description:"test description", ModelNumber:"test model number", UpcCode:"test upc code", Status:"test status", IsActive:"true", IsPerishable:"false", HandlingInstructions:"test handling instructions", UnitsPerCase:"20", TrackingType:"none", InventoryType:"", CreatedDate:"2026-03-13 11:41:42 +0000 UTC", UpdatedDate:"2026-03-13 11:41:44 +0000 UTC"}
    apitest.go:77: EXP
    apitest.go:78: &productapp.Product{ProductID:"e25c16be-a454-41bf-aea0-06613157e5cf", SKU:"sku123", BrandID:"5002ccc6-0c6d-45cd-a25c-f1ca34207823", ProductCategoryID:"ca445934-8a5e-46d7-8633-291798d7857a", Name:"Test Product", Description:"test description", ModelNumber:"test model number", UpcCode:"test upc code", Status:"test status", IsActive:"true", IsPerishable:"false", HandlingInstructions:"test handling instructions", UnitsPerCase:"20", TrackingType:"", InventoryType:"", CreatedDate:"2026-03-13 11:41:42 +0000 UTC", UpdatedDate:"2026-03-13 11:41:44 +0000 UTC"}
    apitest.go:79: Should get the expected response
--- FAIL: Test_InventoryProduct/update-200-basic (0.02s)
```

## Fix
- **File**: `api/cmd/services/ichor/tests/products/productapi/update_test.go:52`
- **Classification**: test bug
- **Change**: Added `TrackingType: "none"` to ExpResp — product creation defaults empty TrackingType to "none" (productbus.go:80-83), but the test expected the zero-value ""
- **Verified**: `go test -v -run "Test_InventoryProduct/update-200-basic" github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/products/productapi` ✓
