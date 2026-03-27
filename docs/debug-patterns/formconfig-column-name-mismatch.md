# formconfig-column-name-mismatch

**Signal**: `COLUMN_NOT_FOUND`, `TestFormConfigsAgainstSchema`, `column "X" does not exist`, `tableforms.go`, form field Name mismatch
**Root cause**: Form config `Name` field set to FK column name from referencing table (e.g., `product_category_id`) instead of actual column name in target table (e.g., `category_id`).
**Fix**:
1. Find failing field in `business/sdk/dbtest/seedmodels/tableforms.go`
2. Check actual column name: `\d schema.table_name` or grep `migrate.sql` for the CREATE TABLE
3. Change `Name: "wrong_column"` to match the real column name in the target table
4. Re-run `TestFormConfigsAgainstSchema` to confirm

**See also**: `docs/arch/form-data.md`, `docs/arch/table-builder.md`
**Examples**:
- `dbtest_TestFormConfigsAgainstSchema_GetProductFormFields.md` -- used `product_category_id` instead of `category_id` in products.products
