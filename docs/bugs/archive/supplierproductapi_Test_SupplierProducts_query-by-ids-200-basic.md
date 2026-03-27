# Test Failure: Test_SupplierProducts/query-by-ids-200-basic

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/procurement/supplierproductapi`
- **Duration**: 0.02s

## Failure Output

```
    apitest.go:73: DIFF
    apitest.go:74:   &supplierproductapp.SupplierProducts{
        - 	{
        - 		SupplierProductID:  "254a86c2-48ab-4745-8a47-94f329293f8d",
        - 		SupplierID:         "ade19cfb-0ad2-451c-9ee8-8b0757ebb6a0",
        - 		ProductID:          "2f198b15-2040-4868-93a5-7a77a3592348",
        - 		SupplierPartNumber: "SupplierPartNumber2395",
        - 		MinOrderQuantity:   "2385",
        - 		MaxOrderQuantity:   "2405",
        - 		LeadTimeDays:       "2395",
        - 		UnitCost:           "6.69",
        - 		IsPrimarySupplier:  "false",
        - 		CreatedDate:        "2026-03-13 11:41:14 +0000 UTC",
        - 		UpdatedDate:        "2026-03-13 11:41:14 +0000 UTC",
        - 	},
        + 	{
        + 		SupplierProductID:  "254a86c2-48ab-4745-8a47-94f329293f8d",
        + 		SupplierID:         "ade19cfb-0ad2-451c-9ee8-8b0757ebb6a0",
        + 		ProductID:          "2f198b15-2040-4868-93a5-7a77a3592348",
        + 		SupplierPartNumber: "SupplierPartNumber2395",
        + 		MinOrderQuantity:   "2385",
        + 		MaxOrderQuantity:   "2405",
        + 		LeadTimeDays:       "2395",
        + 		UnitCost:           "6.69",
        + 		IsPrimarySupplier:  "false",
        + 		CreatedDate:        "2026-03-13 11:41:14 +0000 UTC",
        + 		UpdatedDate:        "2026-03-13 11:41:14 +0000 UTC",
        + 	},
          }
    apitest.go:75: GOT
    apitest.go:76: &supplierproductapp.SupplierProducts{supplierproductapp.SupplierProduct{SupplierProductID:"254a86c2-48ab-4745-8a47-94f329293f8d", SupplierID:"ade19cfb-0ad2-451c-9ee8-8b0757ebb6a0", ProductID:"2f198b15-2040-4868-93a5-7a77a3592348", SupplierPartNumber:"SupplierPartNumber2395", MinOrderQuantity:"2385", MaxOrderQuantity:"2405", LeadTimeDays:"2395", UnitCost:"6.69", IsPrimarySupplier:"false", CreatedDate:"2026-03-13 11:41:14 +0000 UTC", UpdatedDate:"2026-03-13 11:41:14 +0000 UTC"}, supplierproductapp.SupplierProduct{SupplierProductID:"d3a20621-28e2-4a7f-8e2e-c89e7ced5f4a", SupplierID:"a2a037f6-6d2b-4c25-9455-101a0972e0fe", ProductID:"0b4b6e96-d1b8-41b9-a99f-1d8e7b36ff1a", SupplierPartNumber:"SupplierPartNumber2409", MinOrderQuantity:"2399", MaxOrderQuantity:"2419", LeadTimeDays:"2409", UnitCost:"4.33", IsPrimarySupplier:"false", CreatedDate:"2026-03-13 11:41:14 +0000 UTC", UpdatedDate:"2026-03-13 11:41:14 +0000 UTC"}}
    apitest.go:77: EXP
    apitest.go:78: &supplierproductapp.SupplierProducts{supplierproductapp.SupplierProduct{SupplierProductID:"d3a20621-28e2-4a7f-8e2e-c89e7ced5f4a", SupplierID:"a2a037f6-6d2b-4c25-9455-101a0972e0fe", ProductID:"0b4b6e96-d1b8-41b9-a99f-1d8e7b36ff1a", SupplierPartNumber:"SupplierPartNumber2409", MinOrderQuantity:"2399", MaxOrderQuantity:"2419", LeadTimeDays:"2409", UnitCost:"4.33", IsPrimarySupplier:"false", CreatedDate:"2026-03-13 11:41:14 +0000 UTC", UpdatedDate:"2026-03-13 11:41:14 +0000 UTC"}, supplierproductapp.SupplierProduct{SupplierProductID:"254a86c2-48ab-4745-8a47-94f329293f8d", SupplierID:"ade19cfb-0ad2-451c-9ee8-8b0757ebb6a0", ProductID:"2f198b15-2040-4868-93a5-7a77a3592348", SupplierPartNumber:"SupplierPartNumber2395", MinOrderQuantity:"2385", MaxOrderQuantity:"2405", LeadTimeDays:"2395", UnitCost:"6.69", IsPrimarySupplier:"false", CreatedDate:"2026-03-13 11:41:14 +0000 UTC", UpdatedDate:"2026-03-13 11:41:14 +0000 UTC"}}
    apitest.go:79: Should get the expected response
--- FAIL: Test_SupplierProducts/query-by-ids-200-basic (0.02s)
```

## Fix
- **File**: `business/domain/procurement/supplierproductbus/stores/supplierproductdb/supplierproductdb.go:186`
- **Classification**: code bug
- **Change**: Added `ORDER BY product_id ASC` to `QueryByIDs` SQL — it had no ORDER BY while `Query()` did, causing non-deterministic result order
- **Verified**: `go test -v -run Test_SupplierProducts/query-by-ids-200-basic ./api/cmd/services/ichor/tests/procurement/supplierproductapi/` ✓
