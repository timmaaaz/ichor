# Test Failure: Test_Order

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/sales/ordersapi`
- **Duration**: 4.01s

## Failure Output

```
--- FAIL: Test_Order (4.01s)
2026/03/13 11:42:27 INFO  No logger configured for temporal client. Created default one.
```

## Fix
- **File**: `api/cmd/services/ichor/tests/sales/ordersapi/update_test.go:36`
- **Classification**: test bug
- **Change**: Added `AssignedTo: sd.Orders[0].AssignedTo` to ExpResp in `update200` — field was added to `ordersapp.Order` but test expected struct was not updated.
- **Verified**: `go test -v -run Test_Order/update-200-basic github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/sales/ordersapi` ✓
