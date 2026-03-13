# Test Failure: Test_OrderLineItem/create-200-basic

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/sales/orderlineitemsapi`
- **Duration**: 0.01s

## Failure Output

```
    apitest.go:73: DIFF
    apitest.go:74:   &orderlineitemsapp.OrderLineItem{
          	... // 8 identical fields
          	LineTotal:                     "",
          	LineItemFulfillmentStatusesID: "2b933b6c-a26d-4153-8abd-f9a8ffba16e8",
        - 	PickedQuantity:                "0",
        + 	PickedQuantity:                "",
        - 	BackorderedQuantity:           "0",
        + 	BackorderedQuantity:           "",
          	ShortPickReason:               nil,
          	CreatedBy:                     "e89b4157-aceb-414b-90a7-1c2644b7f410",
          	... // 3 identical fields
          }
    apitest.go:75: GOT
    apitest.go:76: &orderlineitemsapp.OrderLineItem{ID:"5d12cd28-245b-43c0-b4e6-e7010a9ba474", OrderID:"ab3d6916-ac40-4c5e-8e74-aeccaf320e6f", ProductID:"af1fb1b6-2146-4c99-966f-5161ad9f93c3", Description:"", Quantity:"1", UnitPrice:"", Discount:"0.00", DiscountType:"flat", LineTotal:"", LineItemFulfillmentStatusesID:"2b933b6c-a26d-4153-8abd-f9a8ffba16e8", PickedQuantity:"0", BackorderedQuantity:"0", ShortPickReason:(*string)(nil), CreatedBy:"e89b4157-aceb-414b-90a7-1c2644b7f410", CreatedDate:"2026-03-13T11:42:21Z", UpdatedBy:"e89b4157-aceb-414b-90a7-1c2644b7f410", UpdatedDate:"2026-03-13T11:42:21Z"}
    apitest.go:77: EXP
    apitest.go:78: &orderlineitemsapp.OrderLineItem{ID:"5d12cd28-245b-43c0-b4e6-e7010a9ba474", OrderID:"ab3d6916-ac40-4c5e-8e74-aeccaf320e6f", ProductID:"af1fb1b6-2146-4c99-966f-5161ad9f93c3", Description:"", Quantity:"1", UnitPrice:"", Discount:"0.00", DiscountType:"flat", LineTotal:"", LineItemFulfillmentStatusesID:"2b933b6c-a26d-4153-8abd-f9a8ffba16e8", PickedQuantity:"", BackorderedQuantity:"", ShortPickReason:(*string)(nil), CreatedBy:"e89b4157-aceb-414b-90a7-1c2644b7f410", CreatedDate:"2026-03-13T11:42:21Z", UpdatedBy:"e89b4157-aceb-414b-90a7-1c2644b7f410", UpdatedDate:"2026-03-13T11:42:21Z"}
    apitest.go:79: Should get the expected response
--- FAIL: Test_OrderLineItem/create-200-basic (0.01s)
```

## Fix
- **File**: `api/cmd/services/ichor/tests/sales/orderlineitemsapi/create_test.go:29`
- **Classification**: test bug
- **Change**: Added `PickedQuantity: "0"` and `BackorderedQuantity: "0"` to `ExpResp` — these fields are `int` in the business model so they default to 0, and `strconv.Itoa(0)` returns `"0"`.
- **Verified**: `go test -v -run Test_OrderLineItem/create-200-basic github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/sales/orderlineitemsapi` ✓
- **pattern-match**: business-default-in-test
