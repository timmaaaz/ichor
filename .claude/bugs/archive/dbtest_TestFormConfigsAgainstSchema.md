# Test Failure: TestFormConfigsAgainstSchema

- **Package**: `github.com/timmaaaz/ichor/business/sdk/dbtest`
- **Duration**: 3s

## Failure Output

```
    formvalidation_test.go:30: 
        === Validating 55 Registered Forms Against Schema ===
    formvalidation_test.go:87: 
        === Form Validation Summary ===
    formvalidation_test.go:88: Forms validated: 55
    formvalidation_test.go:89: Total errors: 2
    formvalidation_test.go:90: Total warnings: 0
    formvalidation_test.go:93: 
        Failed forms:
    formvalidation_test.go:95:   - GetProductFormFields
    formvalidation_test.go:95:   - GetInventoryLocationFormFields
--- FAIL: TestFormConfigsAgainstSchema (3.00s)
```

## Investigation

### iter-1
target: `business/sdk/dbtest/seedmodels/tableforms.go:575`
classification: test bug (seed data wrong)
confidence: high
gap_notes:
- none

**GetProductFormFields**: `product_category_id` — column does not exist in `products.products`. Actual column is `category_id`. Wrong "fix" introduced by commit `3a77ada4`.

**GetInventoryLocationFormFields**: `value_column: "zone_id"` in zone_id smart-combobox config — `zone_id` does not exist in `inventory.zones` (has `id` instead). Already fixed by uncommitted change in working tree (→ `"id"`).

## Fix

- **File**: `business/sdk/dbtest/seedmodels/tableforms.go:575`
- **Classification**: test bug
- **Change**: Reverted `product_category_id` → `category_id` in `GetProductFormFields` (undoes bad commit `3a77ada4`). The uncommitted fix for `GetInventoryLocationFormFields` (`zone_id` → `id` in `value_column`) was already present.
- **Verified**: `go build ./business/sdk/dbtest/...` ✓ (DB required for full test run)
