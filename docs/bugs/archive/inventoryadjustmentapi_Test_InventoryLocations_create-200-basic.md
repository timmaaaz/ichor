# Test Failure: Test_InventoryLocations/create-200-basic

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/inventoryadjustmentapi`
- **Duration**: 0.02s

## Investigation

### iter-1
target: `api/cmd/services/ichor/tests/inventory/inventoryadjustmentapi/create_test.go:39`
classification: test bug
confidence: high
gap_notes:
- none

## Fix

- **File**: `api/cmd/services/ichor/tests/inventory/inventoryadjustmentapi/create_test.go:44`
- **Classification**: test bug
- **Change**: Added `ApprovalStatus: "pending"` to `ExpResp` struct in `create200()`; business layer sets this default on create but test was written before that feature existed
- **Verified**: `go test -v -run Test_InventoryLocations/create-200-basic github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/inventoryadjustmentapi` ✓

## Failure Output

```
    apitest.go:73: DIFF
    apitest.go:74:   &inventoryadjustmentapp.InventoryAdjustment{
          	... // 3 identical fields
          	AdjustedBy:     "4f6a1706-cbf7-493a-8162-558ca1eae2ec",
          	ApprovedBy:     "",
        - 	ApprovalStatus: "pending",
        + 	ApprovalStatus: "",
          	ApprovalReason: "",
          	RejectedBy:     "",
          	... // 7 identical fields
          }
    apitest.go:75: GOT
    apitest.go:76: &inventoryadjustmentapp.InventoryAdjustment{InventoryAdjustmentID:"e7c3f85e-2834-4d5a-a8d3-8121e70b24a6", ProductID:"43fdf36c-b838-4050-b9ed-66b75383c55c", LocationID:"001280e9-43e3-4121-9d0d-482ad05cf5b1", AdjustedBy:"4f6a1706-cbf7-493a-8162-558ca1eae2ec", ApprovedBy:"", ApprovalStatus:"pending", ApprovalReason:"", RejectedBy:"", RejectionReason:"", QuantityChange:"10", ReasonCode:"Purchase", Notes:"New purchase", AdjustmentDate:"2026-03-13 11:39:40 +0000 UTC", CreatedDate:"2026-03-13 11:39:40 +0000 UTC", UpdatedDate:"2026-03-13 11:39:40 +0000 UTC"}
    apitest.go:77: EXP
    apitest.go:78: &inventoryadjustmentapp.InventoryAdjustment{InventoryAdjustmentID:"e7c3f85e-2834-4d5a-a8d3-8121e70b24a6", ProductID:"43fdf36c-b838-4050-b9ed-66b75383c55c", LocationID:"001280e9-43e3-4121-9d0d-482ad05cf5b1", AdjustedBy:"4f6a1706-cbf7-493a-8162-558ca1eae2ec", ApprovedBy:"", ApprovalStatus:"", ApprovalReason:"", RejectedBy:"", RejectionReason:"", QuantityChange:"10", ReasonCode:"Purchase", Notes:"New purchase", AdjustmentDate:"2026-03-13 11:39:40 +0000 UTC", CreatedDate:"2026-03-13 11:39:40 +0000 UTC", UpdatedDate:"2026-03-13 11:39:40 +0000 UTC"}
    apitest.go:79: Should get the expected response
--- FAIL: Test_InventoryLocations/create-200-basic (0.02s)
```
