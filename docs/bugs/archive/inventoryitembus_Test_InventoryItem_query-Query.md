# Test Failure: Test_InventoryItem/query-Query

- **Package**: `github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus`
- **Duration**: 0s

## Failure Output

```
    unittest.go:17: DIFF
    unittest.go:18:   []inventoryitembus.InventoryItem{
        - 	{
        - 		ID:                    s"04afb281-5dc8-4348-83b0-583bf3a947a5",
        - 		ProductID:             s"c8abe78f-61bd-4003-be13-9b326c997be4",
        - 		LocationID:            s"9bd4758a-cb26-4eac-87f9-0f3ffdff7c4a",
        - 		Quantity:              114,
        - 		MinimumStock:          10,
        - 		MaximumStock:          228,
        - 		ReorderPoint:          20,
        - 		EconomicOrderQuantity: 50,
        - 		SafetyStock:           15,
        - 		AvgDailyUsage:         5,
        - 		CreatedDate:           s"2026-03-13 17:34:28.033302 +0000 UTC",
        - 		UpdatedDate:           s"2026-03-13 17:34:28.033302 +0000 UTC",
        - 	},
        - 	{
        - 		ID:                    s"14eef219-ad6f-4d23-b416-9650e847cbaa",
        - 		ProductID:             s"c8abe78f-61bd-4003-be13-9b326c997be4",
        - 		LocationID:            s"d377afc9-f37d-42bf-8b14-c8a0ab2f57e9",
        - 		Quantity:              121,
        - 		MinimumStock:          10,
        - 		MaximumStock:          242,
        - 		ReorderPoint:          20,
        - 		EconomicOrderQuantity: 50,
        - 		SafetyStock:           15,
        - 		AvgDailyUsage:         5,
        - 		CreatedDate:           s"2026-03-13 17:34:28.053357 +0000 UTC",
        - 		UpdatedDate:           s"2026-03-13 17:34:28.053357 +0000 UTC",
        - 	},
        - 	{
        - 		ID:                    s"1c594af1-c2fc-4b57-9d1e-e5ec37b60096",
        - 		ProductID:             s"c8abe78f-61bd-4003-be13-9b326c997be4",
        - 		LocationID:            s"39ba5392-9f0c-4a3a-8581-9d99c7408170",
        - 		Quantity:              105,
        - 		MinimumStock:          10,
        - 		MaximumStock:          210,
        - 		ReorderPoint:          20,
        - 		EconomicOrderQuantity: 50,
        - 		SafetyStock:           15,
        - 		AvgDailyUsage:         5,
        - 		CreatedDate:           s"2026-03-13 17:34:28.002691 +0000 UTC",
        - 		UpdatedDate:           s"2026-03-13 17:34:28.002691 +0000 UTC",
        - 	},
        - 	{
        - 		ID:                    s"21729ad2-0133-4eb4-a38a-b6b8d8a7535e",
        - 		ProductID:             s"c8abe78f-61bd-4003-be13-9b326c997be4",
        - 		LocationID:            s"e27ecfa8-a513-4d45-869f-f5fbaacc0e1a",
        - 		Quantity:              122,
        - 		MinimumStock:          10,
        - 		MaximumStock:          244,
        - 		ReorderPoint:          20,
        - 		EconomicOrderQuantity: 50,
        - 		SafetyStock:           15,
        - 		AvgDailyUsage:         5,
        - 		CreatedDate:           s"2026-03-13 17:34:28.056309 +0000 UTC",
        - 		UpdatedDate:           s"2026-03-13 17:34:28.056309 +0000 UTC",
        - 	},
        + 	{
        + 		ID:                    s"56bdf962-7b04-462e-a71f-fbef43b521d8",
        + 		ProductID:             s"bb85f9ee-8a55-4580-86c7-5dad61b6da3b",
        + 		LocationID:            s"14fe1b12-5caf-41f7-bdb7-b5b9a773313e",
        + 		Quantity:              125,
        + 		MinimumStock:          10,
        + 		MaximumStock:          250,
        + 		ReorderPoint:          20,
        + 		EconomicOrderQuantity: 50,
        + 		SafetyStock:           15,
        + 		AvgDailyUsage:         5,
        + 		CreatedDate:           s"2026-03-13 17:34:28.064613 +0000 UTC m=+2.020990168",
        + 		UpdatedDate:           s"2026-03-13 17:34:28.064613 +0000 UTC m=+2.020990168",
        + 	},
        + 	{
        + 		ID:                    s"7affda19-f456-4a92-a258-b597bd4d46e3",
        + 		ProductID:             s"bb85f9ee-8a55-4580-86c7-5dad61b6da3b",
        + 		LocationID:            s"2a5a6e05-a2dc-4773-a3e2-577fcf15bb99",
        + 		Quantity:              127,
        + 		MinimumStock:          10,
        + 		MaximumStock:          254,
        + 		ReorderPoint:          20,
        + 		EconomicOrderQuantity: 50,
        + 		SafetyStock:           15,
        + 		AvgDailyUsage:         5,
        + 		CreatedDate:           s"2026-03-13 17:34:28.070109 +0000 UTC m=+2.026486126",
        + 		UpdatedDate:           s"2026-03-13 17:34:28.070109 +0000 UTC m=+2.026486126",
        + 	},
        + 	{
        + 		ID:                    s"f4ae781e-1dbf-4ab9-86bc-c4fccc7c3a79",
        + 		ProductID:             s"bb85f9ee-8a55-4580-86c7-5dad61b6da3b",
        + 		LocationID:            s"2b0c6024-cdd3-4cc8-8525-55a7266ca79e",
        + 		Quantity:              128,
        + 		MinimumStock:          10,
        + 		MaximumStock:          256,
        + 		ReorderPoint:          20,
        + 		EconomicOrderQuantity: 50,
        + 		SafetyStock:           15,
        + 		AvgDailyUsage:         5,
        + 		CreatedDate:           s"2026-03-13 17:34:28.072952 +0000 UTC m=+2.029329126",
        + 		UpdatedDate:           s"2026-03-13 17:34:28.072952 +0000 UTC m=+2.029329126",
        + 	},
        + 	{
        + 		ID:                    s"37175c1b-4652-4e5b-aff2-357144031741",
        + 		ProductID:             s"bb85f9ee-8a55-4580-86c7-5dad61b6da3b",
        + 		LocationID:            s"398f2bb6-01f5-493a-aeff-654b4b8dc3ba",
        + 		Quantity:              129,
        + 		MinimumStock:          10,
        + 		MaximumStock:          258,
        + 		ReorderPoint:          20,
        + 		EconomicOrderQuantity: 50,
        + 		SafetyStock:           15,
        + 		AvgDailyUsage:         5,
        + 		CreatedDate:           s"2026-03-13 17:34:28.076445 +0000 UTC m=+2.032822126",
        + 		UpdatedDate:           s"2026-03-13 17:34:28.076445 +0000 UTC m=+2.032822126",
        + 	},
          }
    unittest.go:19: GOT
    unittest.go:20: []inventoryitembus.InventoryItem{inventoryitembus.InventoryItem{ID:uuid.UUID{0x4, 0xaf, 0xb2, 0x81, 0x5d, 0xc8, 0x43, 0x48, 0x83, 0xb0, 0x58, 0x3b, 0xf3, 0xa9, 0x47, 0xa5}, ProductID:uuid.UUID{0xc8, 0xab, 0xe7, 0x8f, 0x61, 0xbd, 0x40, 0x3, 0xbe, 0x13, 0x9b, 0x32, 0x6c, 0x99, 0x7b, 0xe4}, LocationID:uuid.UUID{0x9b, 0xd4, 0x75, 0x8a, 0xcb, 0x26, 0x4e, 0xac, 0x87, 0xf9, 0xf, 0x3f, 0xfd, 0xff, 0x7c, 0x4a}, Quantity:114, ReservedQuantity:0, AllocatedQuantity:0, MinimumStock:10, MaximumStock:228, ReorderPoint:20, EconomicOrderQuantity:50, SafetyStock:15, AvgDailyUsage:5, CreatedDate:time.Date(2026, time.March, 13, 17, 34, 28, 33302000, time.UTC), UpdatedDate:time.Date(2026, time.March, 13, 17, 34, 28, 33302000, time.UTC)}, inventoryitembus.InventoryItem{ID:uuid.UUID{0x14, 0xee, 0xf2, 0x19, 0xad, 0x6f, 0x4d, 0x23, 0xb4, 0x16, 0x96, 0x50, 0xe8, 0x47, 0xcb, 0xaa}, ProductID:uuid.UUID{0xc8, 0xab, 0xe7, 0x8f, 0x61, 0xbd, 0x40, 0x3, 0xbe, 0x13, 0x9b, 0x32, 0x6c, 0x99, 0x7b, 0xe4}, LocationID:uuid.UUID
ime.UTC), UpdatedDate:time.Date(2026, time.March, 13, 17, 34, 28, 2691000, time.UTC)}, inventoryitembus.InventoryItem{ID:uuid.UUID{0x21, 0x72, 0x9a, 0xd2, 0x1, 0x33, 0x4e, 0xb4, 0xa3, 0x8a, 0xb6, 0xb8, 0xd8, 0xa7, 0x53, 0x5e}, ProductID:uuid.UUID{0xc8, 0xab, 0xe7, 0x8f, 0x61, 0xbd, 0x40, 0x3, 0xbe, 0x13, 0x9b, 0x32, 0x6c, 0x99, 0x7b, 0xe4}, LocationID:uuid.UUID{0xe2, 0x7e, 0xcf, 0xa8, 0xa5, 0x13, 0x4d, 0x45, 0x86, 0x9f, 0xf5, 0xfb, 0xaa, 0xcc, 0xe, 0x1a}, Quantity:122, ReservedQuantity:0, AllocatedQuantity:0, MinimumStock:10, MaximumStock:244, ReorderPoint:20, EconomicOrderQuantity:50, SafetyStock:15, AvgDailyUsage:5, CreatedDate:time.Date(2026, time.March, 13, 17, 34, 28, 56309000, time.UTC), UpdatedDate:time.Date(2026, time.March, 13, 17, 34, 28, 56309000, time.UTC)}, inventoryitembus.InventoryItem{ID:uuid.UUID{0x2f, 0x26, 0x86, 0xf5, 0x9f, 0xa5, 0x44, 0x64, 0xa3, 0x1b, 0xec, 0xd6, 0x6f, 0x79, 0xfe, 0x9e}, ProductID:uuid.UUID{0xbb, 0x85, 0xf9, 0xee, 0x8a, 0x55, 0x45, 0x80, 0x86, 0xc7, 0x5d, 0xad, 0x61, 0xb6
, 0xda, 0x3b}, LocationID:uuid.UUID{0x24, 0x12, 0xf3, 0x87, 0x80, 0x94, 0x49, 0xba, 0xb1, 0x44, 0x91, 0x58, 0xf3, 0xbf, 0x14, 0x4}, Quantity:126, ReservedQuantity:0, AllocatedQuantity:0, MinimumStock:10, MaximumStock:252, ReorderPoint:20, EconomicOrderQuantity:50, SafetyStock:15, AvgDailyUsage:5, CreatedDate:time.Date(2026, time.March, 13, 17, 34, 28, 67363000, time.UTC), UpdatedDate:time.Date(2026, time.March, 13, 17, 34, 28, 67363000, time.UTC)}}
    unittest.go:21: EXP
    unittest.go:22: []inventoryitembus.InventoryItem{inventoryitembus.InventoryItem{ID:uuid.UUID{0x56, 0xbd, 0xf9, 0x62, 0x7b, 0x4, 0x46, 0x2e, 0xa7, 0x1f, 0xfb, 0xef, 0x43, 0xb5, 0x21, 0xd8}, ProductID:uuid.UUID{0xbb, 0x85, 0xf9, 0xee, 0x8a, 0x55, 0x45, 0x80, 0x86, 0xc7, 0x5d, 0xad, 0x61, 0xb6, 0xda, 0x3b}, LocationID:uuid.UUID{0x14, 0xfe, 0x1b, 0x12, 0x5c, 0xaf, 0x41, 0xf7, 0xbd, 0xb7, 0xb5, 0xb9, 0xa7, 0x73, 0x31, 0x3e}, Quantity:125, ReservedQuantity:0, AllocatedQuantity:0, MinimumStock:10, MaximumStock:250, ReorderPoint:20, EconomicOrderQuantity:50, SafetyStock:15, AvgDailyUsage:5, CreatedDate:time.Date(2026, time.March, 13, 17, 34, 28, 64613000, time.UTC), UpdatedDate:time.Date(2026, time.March, 13, 17, 34, 28, 64613000, time.UTC)}, inventoryitembus.InventoryItem{ID:uuid.UUID{0x2f, 0x26, 0x86, 0xf5, 0x9f, 0xa5, 0x44, 0x64, 0xa3, 0x1b, 0xec, 0xd6, 0x6f, 0x79, 0xfe, 0x9e}, ProductID:uuid.UUID{0xbb, 0x85, 0xf9, 0xee, 0x8a, 0x55, 0x45, 0x80, 0x86, 0xc7, 0x5d, 0xad, 0x61, 0xb6, 0xda, 0x3b}, LocationID:uuid.U
UID{0x24, 0x12, 0xf3, 0x87, 0x80, 0x94, 0x49, 0xba, 0xb1, 0x44, 0x91, 0x58, 0xf3, 0xbf, 0x14, 0x4}, Quantity:126, ReservedQuantity:0, AllocatedQuantity:0, MinimumStock:10, MaximumStock:252, ReorderPoint:20, EconomicOrderQuantity:50, SafetyStock:15, AvgDailyUsage:5, CreatedDate:time.Date(2026, time.March, 13, 17, 34, 28, 67363000, time.UTC), UpdatedDate:time.Date(2026, time.March, 13, 17, 34, 28, 67363000, time.UTC)}, inventoryitembus.InventoryItem{ID:uuid.UUID{0x7a, 0xff, 0xda, 0x19, 0xf4, 0x56, 0x4a, 0x92, 0xa2, 0x58, 0xb5, 0x97, 0xbd, 0x4d, 0x46, 0xe3}, ProductID:uuid.UUID{0xbb, 0x85, 0xf9, 0xee, 0x8a, 0x55, 0x45, 0x80, 0x86, 0xc7, 0x5d, 0xad, 0x61, 0xb6, 0xda, 0x3b}, LocationID:uuid.UUID{0x2a, 0x5a, 0x6e, 0x5, 0xa2, 0xdc, 0x47, 0x73, 0xa3, 0xe2, 0x57, 0x7f, 0xcf, 0x15, 0xbb, 0x99}, Quantity:127, ReservedQuantity:0, AllocatedQuantity:0, MinimumStock:10, MaximumStock:254, ReorderPoint:20, EconomicOrderQuantity:50, SafetyStock:15, AvgDailyUsage:5, CreatedDate:time.Date(2026, time.March, 13, 17, 34, 28, 701090
00, time.UTC), UpdatedDate:time.Date(2026, time.March, 13, 17, 34, 28, 70109000, time.UTC)}, inventoryitembus.InventoryItem{ID:uuid.UUID{0xf4, 0xae, 0x78, 0x1e, 0x1d, 0xbf, 0x4a, 0xb9, 0x86, 0xbc, 0xc4, 0xfc, 0xcc, 0x7c, 0x3a, 0x79}, ProductID:uuid.UUID{0xbb, 0x85, 0xf9, 0xee, 0x8a, 0x55, 0x45, 0x80, 0x86, 0xc7, 0x5d, 0xad, 0x61, 0xb6, 0xda, 0x3b}, LocationID:uuid.UUID{0x2b, 0xc, 0x60, 0x24, 0xcd, 0xd3, 0x4c, 0xc8, 0x85, 0x25, 0x55, 0xa7, 0x26, 0x6c, 0xa7, 0x9e}, Quantity:128, ReservedQuantity:0, AllocatedQuantity:0, MinimumStock:10, MaximumStock:256, ReorderPoint:20, EconomicOrderQuantity:50, SafetyStock:15, AvgDailyUsage:5, CreatedDate:time.Date(2026, time.March, 13, 17, 34, 28, 72952000, time.UTC), UpdatedDate:time.Date(2026, time.March, 13, 17, 34, 28, 72952000, time.UTC)}, inventoryitembus.InventoryItem{ID:uuid.UUID{0x37, 0x17, 0x5c, 0x1b, 0x46, 0x52, 0x4e, 0x5b, 0xaf, 0xf2, 0x35, 0x71, 0x44, 0x3, 0x17, 0x41}, ProductID:uuid.UUID{0xbb, 0x85, 0xf9, 0xee, 0x8a, 0x55, 0x45, 0x80, 0x86, 0xc7, 0x5d, 0xad, 0x6
1, 0xb6, 0xda, 0x3b}, LocationID:uuid.UUID{0x39, 0x8f, 0x2b, 0xb6, 0x1, 0xf5, 0x49, 0x3a, 0xae, 0xff, 0x65, 0x4b, 0x4b, 0x8d, 0xc3, 0xba}, Quantity:129, ReservedQuantity:0, AllocatedQuantity:0, MinimumStock:10, MaximumStock:258, ReorderPoint:20, EconomicOrderQuantity:50, SafetyStock:15, AvgDailyUsage:5, CreatedDate:time.Date(2026, time.March, 13, 17, 34, 28, 76445000, time.UTC), UpdatedDate:time.Date(2026, time.March, 13, 17, 34, 28, 76445000, time.UTC)}}
    unittest.go:23: Should get the expected response
--- FAIL: Test_InventoryItem/query-Query (0.00s)
```

## Fix
- **File**: `business/domain/inventory/inventoryitembus/inventoryitembus_test.go:196`
- **Classification**: test bug
- **Change**: Added `ProductID: &p1ID` filter to scope the query to test-specific rows (Products[1] = exactly 5 items). Replaced hardcoded `sd.InventoryItems[0:5]` with dynamically-computed `expItems` filtered from the already-sorted slice. Prevents global-seed inventory items from contaminating the unfiltered page-1 result.
- **Verified**: `go test -v -run Test_InventoryItem github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus` ✓
