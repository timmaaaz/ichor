# Test Failure: Test_OrderLineItem/update-200-basic

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/sales/orderlineitemsapi`
- **Duration**: 0.01s

## Failure Output

```
    apitest.go:73: DIFF
    apitest.go:74:   &orderlineitemsapp.OrderLineItem{
          	... // 8 identical fields
          	LineTotal:                     "171.37",
          	LineItemFulfillmentStatusesID: "2b933b6c-a26d-4153-8abd-f9a8ffba16e8",
        - 	PickedQuantity:                "",
        + 	PickedQuantity:                "0",
        - 	BackorderedQuantity:           "",
        + 	BackorderedQuantity:           "0",
          	ShortPickReason:               nil,
          	CreatedBy:                     "e89b4157-aceb-414b-90a7-1c2644b7f410",
          	... // 3 identical fields
          }
    apitest.go:75: GOT
    apitest.go:76: &orderlineitemsapp.OrderLineItem{ID:"45ebfd1b-55eb-49e3-945b-59eeb88436d3", OrderID:"ab3d6916-ac40-4c5e-8e74-aeccaf320e6f", ProductID:"af1fb1b6-2146-4c99-966f-5161ad9f93c3", Description:"Standard item", Quantity:"2", UnitPrice:"175.37", Discount:"4.00", DiscountType:"flat", LineTotal:"171.37", LineItemFulfillmentStatusesID:"2b933b6c-a26d-4153-8abd-f9a8ffba16e8", PickedQuantity:"0", BackorderedQuantity:"0", ShortPickReason:(*string)(nil), CreatedBy:"e89b4157-aceb-414b-90a7-1c2644b7f410", CreatedDate:"2026-03-13T11:42:20Z", UpdatedBy:"e89b4157-aceb-414b-90a7-1c2644b7f410", UpdatedDate:""}
    apitest.go:77: EXP
    apitest.go:78: &orderlineitemsapp.OrderLineItem{ID:"45ebfd1b-55eb-49e3-945b-59eeb88436d3", OrderID:"ab3d6916-ac40-4c5e-8e74-aeccaf320e6f", ProductID:"af1fb1b6-2146-4c99-966f-5161ad9f93c3", Description:"Standard item", Quantity:"2", UnitPrice:"175.37", Discount:"4.00", DiscountType:"flat", LineTotal:"171.37", LineItemFulfillmentStatusesID:"2b933b6c-a26d-4153-8abd-f9a8ffba16e8", PickedQuantity:"", BackorderedQuantity:"", ShortPickReason:(*string)(nil), CreatedBy:"e89b4157-aceb-414b-90a7-1c2644b7f410", CreatedDate:"2026-03-13T11:42:20Z", UpdatedBy:"e89b4157-aceb-414b-90a7-1c2644b7f410", UpdatedDate:""}
    apitest.go:79: Should get the expected response
--- FAIL: Test_OrderLineItem/update-200-basic (0.01s)
```

## Fix
- **File**: `api/cmd/services/ichor/tests/sales/orderlineitemsapi/update_test.go:36`
- **Classification**: test bug
- **Change**: Added `PickedQuantity` and `BackorderedQuantity` to `ExpResp` — fields default to `"0"` via `strconv.Itoa(0)` in app layer but were omitted from expected struct.
- **Verified**: `go test -v -run Test_OrderLineItem/update-200-basic github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/sales/orderlineitemsapi` ✓
