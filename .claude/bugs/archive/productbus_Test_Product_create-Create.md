# Test Failure: Test_Product/create-Create

- **Package**: `github.com/timmaaaz/ichor/business/domain/products/productbus`
- **Duration**: 0s

## Failure Output

```
    unittest.go:17: DIFF
    unittest.go:18: error occurred
    unittest.go:19: GOT
    unittest.go:20: &fmt.wrapError{msg:"create: namedexeccontext: product entry is not unique", err:(*fmt.wrapError)(0x140002b01e0)}
    unittest.go:21: EXP
    unittest.go:22: productbus.Product{ProductID:uuid.UUID{0x3d, 0xec, 0xbe, 0xc8, 0xab, 0xc1, 0x40, 0xd0, 0xab, 0x3b, 0xc, 0x55, 0xa6, 0x9c, 0xc7, 0x7e}, SKU:"SKU8108", BrandID:uuid.UUID{0xf2, 0xa5, 0xc3, 0xae, 0xe0, 0xe6, 0x42, 0xee, 0x87, 0xfa, 0x74, 0xb, 0x7d, 0x28, 0x39, 0x29}, ProductCategoryID:uuid.UUID{0x4a, 0xaa, 0x23, 0x62, 0x3a, 0x7f, 0x42, 0xc8, 0xbf, 0x99, 0x83, 0xc9, 0xe5, 0xe8, 0x0, 0x90}, Name:"Product8108", Description:"Description8108", ModelNumber:"ModelNumber8108", UpcCode:"UpcCode8108", Status:"Status8108", IsActive:true, IsPerishable:false, HandlingInstructions:"Handling instructions 8108", UnitsPerCase:40540, TrackingType:"none", InventoryType:(*productbus.InventoryType)(nil), CreatedDate:time.Date(2026, time.March, 13, 17, 35, 24, 649342000, time.UTC), UpdatedDate:time.Date(2026, time.March, 13, 17, 35, 24, 649342000, time.UTC)}
    unittest.go:23: Should get the expected response
--- FAIL: Test_Product/create-Create (0.00s)
```

## Fix
- **File**: `business/domain/products/productbus/productbus_test.go:163`
- **Classification**: test bug
- **Change**: `create()` was picking a random seeded product and trying to INSERT it again, hitting the `UpcCode` UNIQUE constraint. Replaced with a fresh `NewProduct` using fixed index `99999` (outside seeded range 1–10020) so no collision occurs. `ExpResp` is constructed to match the new values, with `TrackingType: "none"` to match the business-layer default.
- **pattern-match**: seed-unique-pair-exhausted
- **Verified**: `go test -v -run Test_Product/create-Create ./business/domain/products/productbus/...` ✓
