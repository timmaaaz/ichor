# Test Failure: Test_InventoryAdjustment/create-create

- **Package**: `github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus`
- **Duration**: 0s

## Failure Output

```
    unittest.go:17: DIFF
    unittest.go:18:   inventoryadjustmentbus.InventoryAdjustment{
          	... // 3 identical fields
          	AdjustedBy:     {0x68, 0x4c, 0x74, 0x8c, ...},
          	ApprovedBy:     nil,
        - 	ApprovalStatus: "pending",
        + 	ApprovalStatus: "",
          	ApprovalReason: "",
          	RejectedBy:     nil,
          	... // 7 identical fields
          }
    unittest.go:19: GOT
    unittest.go:20: inventoryadjustmentbus.InventoryAdjustment{InventoryAdjustmentID:uuid.UUID{0xd7, 0x93, 0xcf, 0xf, 0x8f, 0x3c, 0x47, 0x91, 0xb1, 0x2c, 0x9b, 0x84, 0x99, 0xc, 0xc, 0x64}, ProductID:uuid.UUID{0x31, 0xd1, 0x39, 0xe5, 0x6f, 0xa7, 0x4d, 0xfc, 0x9a, 0x0, 0x88, 0x42, 0x3c, 0x80, 0xb2, 0xb6}, LocationID:uuid.UUID{0x4, 0x7f, 0x11, 0x8a, 0xa1, 0x36, 0x4d, 0x57, 0xbe, 0xcf, 0x69, 0x57, 0x52, 0xf3, 0xff, 0x2}, AdjustedBy:uuid.UUID{0x68, 0x4c, 0x74, 0x8c, 0xfe, 0xee, 0x4a, 0x17, 0xbb, 0xa5, 0x7b, 0xa9, 0x12, 0x9f, 0xe5, 0xa3}, ApprovedBy:(*uuid.UUID)(nil), ApprovalStatus:"pending", ApprovalReason:"", RejectedBy:(*uuid.UUID)(nil), RejectionReason:"", QuantityChange:10, ReasonCode:"Purchase", Notes:"New purchase", AdjustmentDate:time.Date(2026, time.March, 13, 11, 47, 27, 203728000, time.UTC), CreatedDate:time.Date(2026, time.March, 13, 11, 47, 27, 203738000, time.UTC), UpdatedDate:time.Date(2026, time.March, 13, 11, 47, 27, 203738000, time.UTC)}
    unittest.go:21: EXP
    unittest.go:22: inventoryadjustmentbus.InventoryAdjustment{InventoryAdjustmentID:uuid.UUID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, ProductID:uuid.UUID{0x31, 0xd1, 0x39, 0xe5, 0x6f, 0xa7, 0x4d, 0xfc, 0x9a, 0x0, 0x88, 0x42, 0x3c, 0x80, 0xb2, 0xb6}, LocationID:uuid.UUID{0x4, 0x7f, 0x11, 0x8a, 0xa1, 0x36, 0x4d, 0x57, 0xbe, 0xcf, 0x69, 0x57, 0x52, 0xf3, 0xff, 0x2}, AdjustedBy:uuid.UUID{0x68, 0x4c, 0x74, 0x8c, 0xfe, 0xee, 0x4a, 0x17, 0xbb, 0xa5, 0x7b, 0xa9, 0x12, 0x9f, 0xe5, 0xa3}, ApprovedBy:(*uuid.UUID)(nil), ApprovalStatus:"", ApprovalReason:"", RejectedBy:(*uuid.UUID)(nil), RejectionReason:"", QuantityChange:10, ReasonCode:"Purchase", Notes:"New purchase", AdjustmentDate:time.Date(2026, time.March, 13, 11, 47, 27, 203728000, time.UTC), CreatedDate:time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), UpdatedDate:time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}
    unittest.go:23: Should get the expected response
--- FAIL: Test_InventoryAdjustment/create-create (0.00s)
```

## Fix

- **File**: `business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus_test.go:230`
- **Classification**: test bug
- **Change**: Added `ApprovalStatus: inventoryadjustmentbus.ApprovalStatusPending` to `create` test `ExpResp`; also added `ApprovalStatus: sd.InventoryAdjustments[1].ApprovalStatus` to `update` test `ExpResp`
- **pattern-match**: business-default-in-test
- **Verified**: `go test -v -run Test_InventoryAdjustment/create ./business/domain/inventory/inventoryadjustmentbus/...` ✓
