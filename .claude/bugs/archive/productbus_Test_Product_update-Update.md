# Test Failure: Test_Product/update-Update

- **Package**: `github.com/timmaaaz/ichor/business/domain/products/productbus`
- **Duration**: 0s

## Failure Output

```
    unittest.go:17: DIFF
    unittest.go:18: error occurred
    unittest.go:19: GOT
    unittest.go:20: &fmt.wrapError{msg:"update: namedexeccontext: product entry is not unique", err:(*fmt.wrapError)(0x140002b08e0)}
    unittest.go:21: EXP
    unittest.go:22: productbus.Product{ProductID:uuid.UUID{0xab, 0xd3, 0x8a, 0x12, 0x82, 0x8b, 0x45, 0x90, 0xb2, 0xe8, 0x87, 0xea, 0xa, 0x51, 0xa3, 0xa5}, SKU:"SKU8098", BrandID:uuid.UUID{0xac, 0x74, 0x35, 0xa7, 0x31, 0xe7, 0x40, 0x54, 0x9e, 0xec, 0xf3, 0x5a, 0x4e, 0x6d, 0x87, 0x44}, ProductCategoryID:uuid.UUID{0x34, 0xa1, 0x98, 0x44, 0x75, 0x3f, 0x44, 0xb5, 0x8e, 0x61, 0xe2, 0x26, 0xf8, 0x7, 0x8f, 0x72}, Name:"Product8098", Description:"Description8098", ModelNumber:"ModelNumber8098", UpcCode:"UpcCode8098", Status:"Status8098", IsActive:true, IsPerishable:false, HandlingInstructions:"Handling instructions 8098", UnitsPerCase:40490, TrackingType:"none", InventoryType:(*productbus.InventoryType)(nil), CreatedDate:time.Date(2026, time.March, 13, 17, 35, 24, 609077000, time.UTC), UpdatedDate:time.Date(2026, time.March, 13, 17, 35, 24, 609077000, time.UTC)}
    unittest.go:23: Should get the expected response
--- FAIL: Test_Product/update-Update (0.00s)
```

## Fix
- **File**: `business/domain/products/productbus/productbus_test.go:230`
- **Classification**: test bug
- **Change**: Generate fresh unique values (`UpdatedSKU%d`, `UpdatedUpc%d`, `UpdatedModel%d`) for unique-constrained fields instead of copying them from `updateProduct` (a row still in the DB), preventing `products_upc_code_unique` constraint violation.
- **Verified**: `go test -v -run Test_Product ./business/domain/products/productbus/...` ✓
