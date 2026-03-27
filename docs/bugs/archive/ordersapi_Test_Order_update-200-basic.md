# Test Failure: Test_Order/update-200-basic

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/sales/ordersapi`
- **Duration**: 0.02s

## Failure Output

```
    apitest.go:73: DIFF
    apitest.go:74:   &ordersapp.Order{
          	... // 6 identical fields
          	BillingAddressID:  "",
          	ShippingAddressID: "",
        - 	AssignedTo:        "",
        + 	AssignedTo:        "8d2a1ebe-db06-4f19-8ce9-5ac4072dd79b",
          	Subtotal:          "109.00",
          	TaxRate:           "8.00",
          	... // 10 identical fields
          }
    apitest.go:75: GOT
    apitest.go:76: &ordersapp.Order{ID:"2f5d1e03-02ea-41cc-b638-060145c93ee6", Number:"TST-000000001", CustomerID:"90a89bb9-e792-44e3-9fcf-2213f9e6457e", DueDate:"2026-04-22", FulfillmentStatusID:"803e82f2-f795-4a30-a32a-23f2073ea7bb", OrderDate:"2026-03-13", BillingAddressID:"", ShippingAddressID:"", AssignedTo:"8d2a1ebe-db06-4f19-8ce9-5ac4072dd79b", Subtotal:"109.00", TaxRate:"8.00", TaxAmount:"8.72", ShippingCost:"25.00", TotalAmount:"142.72", CurrencyID:"2f0a4e6c-6b20-4de2-b374-6fd2117ed593", PaymentTermID:"", Notes:"Test order 1", CreatedBy:"8d2a1ebe-db06-4f19-8ce9-5ac4072dd79b", UpdatedBy:"8d2a1ebe-db06-4f19-8ce9-5ac4072dd79b", CreatedDate:"2026-03-13T11:42:25Z", UpdatedDate:"2026-03-13T11:42:26Z"}
    apitest.go:77: EXP
    apitest.go:78: &ordersapp.Order{ID:"2f5d1e03-02ea-41cc-b638-060145c93ee6", Number:"TST-000000001", CustomerID:"90a89bb9-e792-44e3-9fcf-2213f9e6457e", DueDate:"2026-04-22", FulfillmentStatusID:"803e82f2-f795-4a30-a32a-23f2073ea7bb", OrderDate:"2026-03-13", BillingAddressID:"", ShippingAddressID:"", AssignedTo:"", Subtotal:"109.00", TaxRate:"8.00", TaxAmount:"8.72", ShippingCost:"25.00", TotalAmount:"142.72", CurrencyID:"2f0a4e6c-6b20-4de2-b374-6fd2117ed593", PaymentTermID:"", Notes:"Test order 1", CreatedBy:"8d2a1ebe-db06-4f19-8ce9-5ac4072dd79b", UpdatedBy:"8d2a1ebe-db06-4f19-8ce9-5ac4072dd79b", CreatedDate:"2026-03-13T11:42:25Z", UpdatedDate:"2026-03-13T11:42:26Z"}
    apitest.go:79: Should get the expected response
--- FAIL: Test_Order/update-200-basic (0.02s)
```

## Fix
- **File**: `api/cmd/services/ichor/tests/sales/ordersapi/update_test.go:32`
- **Classification**: test bug
- **Change**: Added `AssignedTo: sd.Orders[0].AssignedTo,` to the `ExpResp` struct — `assigned_to` was seeded with the admin's ID (for filter testing) but the update test's expected response was never updated to include it.
- **Verified**: `go test -v -run Test_Order/update-200-basic ./api/cmd/services/ichor/tests/sales/ordersapi/...` ✓
