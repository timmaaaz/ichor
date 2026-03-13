# Test Failure: TestFormConfigsAgainstSchema/GetProductFormFields

- **Package**: `github.com/timmaaaz/ichor/business/sdk/dbtest`
- **Duration**: 0.01s

## Failure Output

```
    formvalidation_test.go:65: [SCHEMA] fields[4]: column "product_category_id" not found in table products.products (COLUMN_NOT_FOUND)
--- FAIL: TestFormConfigsAgainstSchema/GetProductFormFields (0.01s)
```

## Fix

- **File**: `business/sdk/dbtest/seedmodels/tableforms.go:575`
- **Classification**: test bug (form config used wrong column name)
- **Change**: Changed `Name: "product_category_id"` to `Name: "category_id"` to match the actual DB column in `products.products`
- **Verified**: `go test -v -run TestFormConfigsAgainstSchema/GetProductFormFields ./business/sdk/dbtest/...` ✓
