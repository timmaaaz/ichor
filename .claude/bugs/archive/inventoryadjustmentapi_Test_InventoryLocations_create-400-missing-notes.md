# Test Failure: Test_InventoryLocations/create-400-missing-notes

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/inventoryadjustmentapi`
- **Duration**: 0.01s

## Failure Output

```
    apitest.go:57: missing-notes: Should receive a status code of 400 for the response : 200
--- FAIL: Test_InventoryLocations/create-400-missing-notes (0.01s)
```

## Investigation

### iter-1
target: `app/domain/inventory/inventoryadjustmentapp/model.go:105`
classification: code bug
confidence: High
gap_notes: none

Root cause: commit `a1c51f0e` changed `Notes` from `validate:"required"` → `validate:"omitempty"` ("matches UI intent"). Test expects 400 when Notes is missing, but omitempty allows empty string → returns 200.

Fix already present in working directory: reverted `Notes` back to `validate:"required"` (also reverted `ApprovedBy` from `omitempty,min=36,max=36` → `required,min=36,max=36`).

## Fix

- **File**: `app/domain/inventory/inventoryadjustmentapp/model.go:105`
- **Classification**: code bug
- **Change**: Reverted `Notes validate:"omitempty"` → `validate:"required"` and `ApprovedBy validate:"omitempty,min=36,max=36"` → `validate:"required,min=36,max=36"` (fix was already in working directory)
- **Verified**: `go build ./app/domain/inventory/inventoryadjustmentapp/...` ✓
