# Test Failure: Test_Inspections/create-Create

- **Package**: `github.com/timmaaaz/ichor/business/domain/inventory/inspectionbus`
- **Duration**: 0s

## Failure Output

```
    unittest.go:17: DIFF
    unittest.go:18: got is not a inspection: create: namedexeccontext: ERROR: new row for relation "quality_inspections" violates check constraint "quality_inspections_status_check" (SQLSTATE 23514)
    unittest.go:19: GOT
    unittest.go:20: &fmt.wrapError{msg:"create: namedexeccontext: ERROR: new row for relation \"quality_inspections\" violates check constraint \"quality_inspections_status_check\" (SQLSTATE 23514)", err:(*fmt.wrapError)(0x140002b0960)}
    unittest.go:21: EXP
    unittest.go:22: inspectionbus.Inspection{InspectionID:uuid.UUID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, ProductID:uuid.UUID{0x4c, 0xcc, 0x27, 0x8c, 0xd0, 0xcc, 0x44, 0x2, 0xbd, 0xc1, 0x75, 0xde, 0x3e, 0x61, 0x3f, 0xcc}, InspectorID:uuid.UUID{0x85, 0xab, 0x79, 0x3c, 0xb1, 0xf5, 0x4a, 0x84, 0xb7, 0xfd, 0xe2, 0x9a, 0x95, 0x81, 0x53, 0x11}, LotID:uuid.UUID{0xad, 0x36, 0x59, 0xe7, 0x85, 0xd0, 0x40, 0x1a, 0xb4, 0xbc, 0x6e, 0xc8, 0xf7, 0xe8, 0x55, 0x1e}, Status:"Pending", Notes:"Initial inspection", InspectionDate:time.Date(2026, time.March, 27, 15, 27, 0, 129849000, time.UTC), NextInspectionDate:time.Date(2026, time.March, 28, 15, 27, 0, 129849000, time.UTC), UpdatedDate:time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), CreatedDate:time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}
    unittest.go:23: Should get the expected response
--- FAIL: Test_Inspections/create-Create (0.00s)
```

## Fix
- **File**: `business/domain/inventory/inspectionbus/inspectionbus_test.go:232,243`
- **Classification**: test bug
- **Change**: Changed `"Pending"` to `"pending"` — DB CHECK constraint requires lowercase values (`pending`, `passed`, `failed`)
- **Verified**: `go test -v -run Test_Inspections/create-Create ./business/domain/inventory/inspectionbus/...` ✓
- **pattern-match**: `invalid-enum-check-constraint`
