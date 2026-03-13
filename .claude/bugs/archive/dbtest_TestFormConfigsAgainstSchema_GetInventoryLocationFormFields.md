# Test Failure: TestFormConfigsAgainstSchema/GetInventoryLocationFormFields

- **Package**: `github.com/timmaaaz/ichor/business/sdk/dbtest`
- **Duration**: 0.01s

## Failure Output

```
    formvalidation_test.go:65: [SCHEMA] fields[0].config.value_column: value column "zone_id" not found in table inventory.zones (DROPDOWN_COLUMN_NOT_FOUND)
--- FAIL: TestFormConfigsAgainstSchema/GetInventoryLocationFormFields (0.01s)
```

## Fix

- **File**: `business/sdk/dbtest/seedmodels/tableforms.go:690`
- **Classification**: test bug (wrong static config in seed data)
- **Change**: Changed `"value_column": "zone_id"` to `"value_column": "id"` in the `GetInventoryLocationFormFields` smart-combobox config. `zone_id` is a FK column in `inventory.inventory_locations`, not a column in `inventory.zones`. The correct value column for the dropdown lookup is the zones table's primary key `id`.
- **Verified**: `go test -v -run TestFormConfigsAgainstSchema/GetInventoryLocationFormFields ./business/sdk/dbtest/...` ✓
