# Test Failure: Test_ActionTypeSchemas_CategoryConsistency

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/workflow/referenceapi`
- **Duration**: 0s

## Failure Output

```
    schema_alignment_test.go:171: Category "inventory": expected 11 types, got 12 (found: [allocate_inventory approve_inventory_adjustment approve_transfer_order check_inventory check_reorder_point commit_allocation create_put_away_task receive_inventory reject_inventory_adjustment reject_transfer_order release_reservation reserve_inventory])
--- FAIL: Test_ActionTypeSchemas_CategoryConsistency (0.00s)
```

## Fix
- **File**: `api/cmd/services/ichor/tests/workflow/referenceapi/schema_alignment_test.go:149-154`
- **Classification**: test bug
- **Change**: Added `"create_put_away_task"` to inventory category expected types (count 11 → 12)
- **Verified**: `go test -v -run Test_ActionTypeSchemas_CategoryConsistency ./api/cmd/services/ichor/tests/workflow/referenceapi/...` ✓
- **pattern-match**: `hardcoded-action-type-list`
