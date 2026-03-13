# Test Failure: Test_InventoryAdjustment/update-update

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
    unittest.go:20: inventoryadjustmentbus.InventoryAdjustment{InventoryAdjustmentID:uuid.UUID{0x13, 0x85, 0x8d, 0x78, 0x94, 0xa7, 0x47, 0x64, 0xa6, 0xca, 0xc6, 0xcf, 0x21, 0x32, 0x29, 0xf4}, ProductID:uuid.UUID{0x31, 0xd1, 0x39, 0xe5, 0x6f, 0xa7, 0x4d, 0xfc, 0x9a, 0x0, 0x88, 0x42, 0x3c, 0x80, 0xb2, 0xb6}, LocationID:uuid.UUID{0x4, 0x7f, 0x11, 0x8a, 0xa1, 0x36, 0x4d, 0x57, 0xbe, 0xcf, 0x69, 0x57, 0x52, 0xf3, 0xff, 0x2}, AdjustedBy:uuid.UUID{0x68, 0x4c, 0x74, 0x8c, 0xfe, 0xee, 0x4a, 0x17, 0xbb, 0xa5, 0x7b, 0xa9, 0x12, 0x9f, 0xe5, 0xa3}, ApprovedBy:(*uuid.UUID)(nil), ApprovalStatus:"pending", ApprovalReason:"", RejectedBy:(*uuid.UUID)(nil), RejectionReason:"", QuantityChange:20, ReasonCode:"Adjustment", Notes:"Updated adjustment", AdjustmentDate:time.Date(2026, time.March, 13, 11, 47, 27, 206943000, time.UTC), CreatedDate:time.Date(2026, time.March, 13, 11, 47, 27, 142923000, time.UTC), UpdatedDate:time.Date(2026, time.March, 13, 11, 47, 27, 206951000, time.UTC)}
    unittest.go:21: EXP
    unittest.go:22: inventoryadjustmentbus.InventoryAdjustment{InventoryAdjustmentID:uuid.UUID{0x13, 0x85, 0x8d, 0x78, 0x94, 0xa7, 0x47, 0x64, 0xa6, 0xca, 0xc6, 0xcf, 0x21, 0x32, 0x29, 0xf4}, ProductID:uuid.UUID{0x31, 0xd1, 0x39, 0xe5, 0x6f, 0xa7, 0x4d, 0xfc, 0x9a, 0x0, 0x88, 0x42, 0x3c, 0x80, 0xb2, 0xb6}, LocationID:uuid.UUID{0x4, 0x7f, 0x11, 0x8a, 0xa1, 0x36, 0x4d, 0x57, 0xbe, 0xcf, 0x69, 0x57, 0x52, 0xf3, 0xff, 0x2}, AdjustedBy:uuid.UUID{0x68, 0x4c, 0x74, 0x8c, 0xfe, 0xee, 0x4a, 0x17, 0xbb, 0xa5, 0x7b, 0xa9, 0x12, 0x9f, 0xe5, 0xa3}, ApprovedBy:(*uuid.UUID)(nil), ApprovalStatus:"", ApprovalReason:"", RejectedBy:(*uuid.UUID)(nil), RejectionReason:"", QuantityChange:20, ReasonCode:"Adjustment", Notes:"Updated adjustment", AdjustmentDate:time.Date(2026, time.March, 13, 11, 47, 27, 206943000, time.UTC), CreatedDate:time.Date(2026, time.March, 13, 11, 47, 27, 142923000, time.UTC), UpdatedDate:time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}
    unittest.go:23: Should get the expected response
--- FAIL: Test_InventoryAdjustment/update-update (0.00s)
```

## Fix

- **File**: `business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus.go:90`
- **Classification**: test bug (stale bug file — fix already applied)
- **Change**: `Create` method was updated to set `ApprovalStatus: ApprovalStatusPending`; seed calls `Create`, so `sd.InventoryAdjustments[1].ApprovalStatus` = `"pending"` which matches GOT
- **pattern-match**: `business-default-in-test`
- **Verified**: `go test -v -run Test_InventoryAdjustment/update-update ./business/domain/inventory/inventoryadjustmentbus/...` ✓
