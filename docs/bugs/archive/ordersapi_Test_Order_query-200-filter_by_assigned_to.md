# Test Failure: Test_Order/query-200-filter_by_assigned_to

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/sales/ordersapi`
- **Duration**: 0.01s

## Failure Output

```
    apitest.go:73: DIFF
    apitest.go:74:   &query.Result[github.com/timmaaaz/ichor/app/domain/sales/ordersapp.Order]{
          	Items:       {{ID: "2f5d1e03-02ea-41cc-b638-060145c93ee6", Number: "TST-000000001", CustomerID: "ddd8138f-6c1f-436f-925d-103162ab2f5b", DueDate: "2026-04-22", ...}},
        - 	Total:       1,
        + 	Total:       5,
          	Page:        1,
          	RowsPerPage: 10,
          }
    apitest.go:75: GOT
    apitest.go:76: &query.Result[github.com/timmaaaz/ichor/app/domain/sales/ordersapp.Order]{Items:[]ordersapp.Order{ordersapp.Order{ID:"2f5d1e03-02ea-41cc-b638-060145c93ee6", Number:"TST-000000001", CustomerID:"ddd8138f-6c1f-436f-925d-103162ab2f5b", DueDate:"2026-04-22", FulfillmentStatusID:"803e82f2-f795-4a30-a32a-23f2073ea7bb", OrderDate:"2026-03-13", BillingAddressID:"", ShippingAddressID:"", AssignedTo:"8d2a1ebe-db06-4f19-8ce9-5ac4072dd79b", Subtotal:"109.00", TaxRate:"8.00", TaxAmount:"8.72", ShippingCost:"25.00", TotalAmount:"142.72", CurrencyID:"2f0a4e6c-6b20-4de2-b374-6fd2117ed593", PaymentTermID:"", Notes:"Test order 1", CreatedBy:"8d2a1ebe-db06-4f19-8ce9-5ac4072dd79b", UpdatedBy:"8d2a1ebe-db06-4f19-8ce9-5ac4072dd79b", CreatedDate:"2026-03-13T11:42:25Z", UpdatedDate:"2026-03-13T11:42:25Z"}}, Total:1, Page:1, RowsPerPage:10}
    apitest.go:77: EXP
    apitest.go:78: &query.Result[github.com/timmaaaz/ichor/app/domain/sales/ordersapp.Order]{Items:[]ordersapp.Order{ordersapp.Order{ID:"2f5d1e03-02ea-41cc-b638-060145c93ee6", Number:"TST-000000001", CustomerID:"ddd8138f-6c1f-436f-925d-103162ab2f5b", DueDate:"2026-04-22", FulfillmentStatusID:"803e82f2-f795-4a30-a32a-23f2073ea7bb", OrderDate:"2026-03-13", BillingAddressID:"", ShippingAddressID:"", AssignedTo:"8d2a1ebe-db06-4f19-8ce9-5ac4072dd79b", Subtotal:"109.00", TaxRate:"8.00", TaxAmount:"8.72", ShippingCost:"25.00", TotalAmount:"142.72", CurrencyID:"2f0a4e6c-6b20-4de2-b374-6fd2117ed593", PaymentTermID:"", Notes:"Test order 1", CreatedBy:"8d2a1ebe-db06-4f19-8ce9-5ac4072dd79b", UpdatedBy:"8d2a1ebe-db06-4f19-8ce9-5ac4072dd79b", CreatedDate:"2026-03-13T11:42:25Z", UpdatedDate:"2026-03-13T11:42:25Z"}}, Total:5, Page:1, RowsPerPage:10}
    apitest.go:79: Should get the expected response
--- FAIL: Test_Order/query-200-filter_by_assigned_to (0.01s)
```

## Fix
- **File**: `api/cmd/services/ichor/tests/sales/ordersapi/query_test.go:58`
- **Classification**: test bug
- **Change**: Changed `Total: 5` to `Total: 1` in "filter by assigned_to" test case — seed only assigns 1 order to admin, so filter should return Total=1 not Total=5 (copy-paste from basic query test)
- **Verified**: `go test -v -run Test_Order/query-200-filter_by_assigned_to github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/sales/ordersapi` ✓
