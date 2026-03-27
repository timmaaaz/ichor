# Test Failure: Test_Inspections/update-Update

- **Package**: `github.com/timmaaaz/ichor/business/domain/inventory/inspectionbus`
- **Duration**: 0s

## Failure Output

```
    unittest.go:17: DIFF
    unittest.go:18: got is not a inspection: update: namedexeccontext: ERROR: new row for relation "quality_inspections" violates check constraint "quality_inspections_status_check" (SQLSTATE 23514)
    unittest.go:19: GOT
    unittest.go:20: &fmt.wrapError{msg:"update: namedexeccontext: ERROR: new row for relation \"quality_inspections\" violates check constraint \"quality_inspections_status_check\" (SQLSTATE 23514)", err:(*fmt.wrapError)(0x140002b1020)}
    unittest.go:21: EXP
    unittest.go:22: inspectionbus.Inspection{InspectionID:uuid.UUID{0x8d, 0x1b, 0x76, 0x90, 0x8b, 0xf9, 0x4d, 0x8e, 0x8c, 0x95, 0x94, 0xb4, 0xaa, 0xa1, 0x4f, 0x0}, ProductID:uuid.UUID{0x42, 0x9d, 0x7b, 0x59, 0xa, 0xce, 0x44, 0xc8, 0x8e, 0xb4, 0xeb, 0xdd, 0x99, 0x5f, 0xe2, 0x16}, InspectorID:uuid.UUID{0x85, 0xab, 0x79, 0x3c, 0xb1, 0xf5, 0x4a, 0x84, 0xb7, 0xfd, 0xe2, 0x9a, 0x95, 0x81, 0x53, 0x11}, LotID:uuid.UUID{0x56, 0x47, 0x88, 0x78, 0x20, 0xc2, 0x44, 0x97, 0xbd, 0x6a, 0x95, 0x76, 0x52, 0xe3, 0xe2, 0x27}, Status:"In Progress", Notes:"Updated inspection", InspectionDate:time.Date(2026, time.March, 27, 15, 27, 0, 95787000, time.UTC), NextInspectionDate:time.Date(2026, time.April, 10, 15, 27, 0, 95787000, time.UTC), UpdatedDate:time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), CreatedDate:time.Date(2026, time.March, 27, 15, 27, 0, 120385000, time.UTC)}
    unittest.go:23: Should get the expected response
--- FAIL: Test_Inspections/update-Update (0.00s)
```

## Fix
- **File**: `business/domain/inventory/inspectionbus/inspectionbus_test.go:281,292`
- **Classification**: test bug
- **Change**: Changed `"In Progress"` to `"passed"` — DB CHECK constraint only allows (`pending`, `passed`, `failed`); "In Progress" is not a valid status
- **Verified**: `go test -v -run Test_Inspections/update-Update ./business/domain/inventory/inspectionbus/...` ✓
- **pattern-match**: `invalid-enum-check-constraint`
