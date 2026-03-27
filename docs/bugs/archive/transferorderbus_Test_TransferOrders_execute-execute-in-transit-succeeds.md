# Test Failure: Test_TransferOrders/execute-execute-in-transit-succeeds

- **Package**: `github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus`
- **Duration**: 0s

## Failure Output

```
    unittest.go:17: DIFF
    unittest.go:18:   any(
        - 	e"executing transfer order: execute: transfer order is not in a state that allows approval/rejection: must be in_transit, got status4",
        + 	string("completed"),
          )
    unittest.go:19: GOT
    unittest.go:20: &fmt.wrapError{msg:"executing transfer order: execute: transfer order is not in a state that allows approval/rejection: must be in_transit, got status4", err:(*fmt.wrapError)(0x14000331f60)}
    unittest.go:21: EXP
    unittest.go:22: "completed"
    unittest.go:23: Should get the expected response
--- FAIL: Test_TransferOrders/execute-execute-in-transit-succeeds (0.00s)
```

## Fix
- **File**: `business/domain/inventory/transferorderbus/testutil.go:26`
- **Classification**: test bug
- **Change**: Replaced `fmt.Sprintf("status%d", idx%5)` with `StatusPending` — seed must use valid status constants; transfer orders always start as "pending"
- **Verified**: `go test -v -run Test_TransferOrders ./business/domain/inventory/transferorderbus/...` ✓
- **pattern-match**: `invalid-enum-check-constraint`
