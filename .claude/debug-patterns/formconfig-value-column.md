# formconfig-value-column

**Signal**: `DROPDOWN_COLUMN_NOT_FOUND`, `value column "X" not found in table Y` in `TestFormConfigsAgainstSchema`; field name is a FK column (ends in `_id`) but error says it isn't found in the *target* table
**Root cause**: In a smart-combobox form field config, `value_column` must be a column in the **target/lookup table** (usually `id`), not the FK column name in the source table. E.g., for `zone_id` in `inventory_locations`, the value column is `id` in `inventory.zones`, not `zone_id`.
**Fix**:
1. Find the failing form config in `business/sdk/dbtest/seedmodels/tableforms.go`
2. Search for the field with the wrong `value_column`
3. Change `"value_column": "<fk_column_name>"` → `"value_column": "id"` (the target table's PK)

Also check that field names in the config match actual column names in the source table (e.g., `product_category_id` vs `category_id`).

**See also**: `docs/arch/form-data.md`
**Examples**:
- `dbtest_TestFormConfigsAgainstSchema_GetInventoryLocationFormFields.md` — `value_column: "zone_id"` in zone dropdown → should be `"id"` (the PK of `inventory.zones`)
- `dbtest_TestFormConfigsAgainstSchema.md` — `product_category_id` field name wrong (actual column: `category_id`) + same `value_column` issue
