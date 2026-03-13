# Test Failure: Test_InventoryLocations/update-Update

- **Package**: `github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus`
- **Duration**: 0s

## Failure Output

```
    unittest.go:17: DIFF
    unittest.go:18:   inventorylocationbus.InventoryLocation{
          	... // 5 identical fields
          	Shelf:             "UpdatedShelf",
          	Bin:               "UpdatedBin",
        - 	LocationCode:      nil,
        + 	LocationCode:      &"Aisle500-Rack500-Shelf500-Bin500",
          	IsPickLocation:    true,
          	IsReserveLocation: false,
          	... // 4 identical fields
          }
    unittest.go:19: GOT
    unittest.go:20: inventorylocationbus.InventoryLocation{LocationID:uuid.UUID{0x3, 0xaa, 0x8d, 0x8c, 0x2f, 0x6f, 0x47, 0x28, 0xb1, 0x40, 0x69, 0x8, 0x72, 0x29, 0xc2, 0xf9}, WarehouseID:uuid.UUID{0xe9, 0x97, 0xe2, 0x8e, 0xc2, 0xea, 0x47, 0x99, 0x8d, 0xb7, 0x6b, 0x4d, 0x47, 0x2b, 0xe1, 0x7e}, ZoneID:uuid.UUID{0x8, 0xa6, 0x6a, 0x9e, 0x7e, 0x70, 0x4e, 0x6c, 0x81, 0x10, 0xe0, 0x53, 0x38, 0xd, 0x99, 0x9a}, Aisle:"UpdatedAisle", Rack:"UpdatedRack", Shelf:"UpdatedShelf", Bin:"UpdatedBin", LocationCode:(*string)(0x1400010ee30), IsPickLocation:true, IsReserveLocation:false, MaxCapacity:100, CurrentUtilization:types.RoundedFloat{Value:0.57}, CreatedDate:time.Date(2026, time.March, 13, 11, 47, 34, 342226000, time.UTC), UpdatedDate:time.Date(2026, time.March, 13, 11, 47, 34, 361081000, time.UTC)}
    unittest.go:21: EXP
    unittest.go:22: inventorylocationbus.InventoryLocation{LocationID:uuid.UUID{0x3, 0xaa, 0x8d, 0x8c, 0x2f, 0x6f, 0x47, 0x28, 0xb1, 0x40, 0x69, 0x8, 0x72, 0x29, 0xc2, 0xf9}, WarehouseID:uuid.UUID{0xe9, 0x97, 0xe2, 0x8e, 0xc2, 0xea, 0x47, 0x99, 0x8d, 0xb7, 0x6b, 0x4d, 0x47, 0x2b, 0xe1, 0x7e}, ZoneID:uuid.UUID{0x8, 0xa6, 0x6a, 0x9e, 0x7e, 0x70, 0x4e, 0x6c, 0x81, 0x10, 0xe0, 0x53, 0x38, 0xd, 0x99, 0x9a}, Aisle:"UpdatedAisle", Rack:"UpdatedRack", Shelf:"UpdatedShelf", Bin:"UpdatedBin", LocationCode:(*string)(nil), IsPickLocation:true, IsReserveLocation:false, MaxCapacity:100, CurrentUtilization:types.RoundedFloat{Value:0.57}, CreatedDate:time.Date(2026, time.March, 13, 11, 47, 34, 342226000, time.UTC), UpdatedDate:time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}
    unittest.go:23: Should get the expected response
--- FAIL: Test_InventoryLocations/update-Update (0.00s)
```

## Fix

- **File**: `business/domain/inventory/inventorylocationbus/inventorylocationbus.go:133`
- **Classification**: code bug
- **Change**: Added an `else if` branch so that when any of Aisle/Rack/Shelf/Bin are changed without an explicit new LocationCode, the stale LocationCode is cleared to nil
- **Verified**: `go test -v -run Test_InventoryLocations/update-Update ./business/domain/inventory/inventorylocationbus/...` ✓
