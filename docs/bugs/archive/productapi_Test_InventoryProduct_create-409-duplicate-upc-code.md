# Test Failure: Test_InventoryProduct/create-409-duplicate-upc-code

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/products/productapi`
- **Duration**: 0.01s

## Fix
- **File**: `business/sdk/migrate/sql/migrate.sql` (Version 2.14)
- **Classification**: code bug
- **Change**: Added `ALTER TABLE products.products ADD CONSTRAINT products_upc_code_unique UNIQUE (upc_code)` as migration version 2.14
- **Verified**: `go test -v -run Test_InventoryProduct/create-409-duplicate-upc-code ./api/cmd/services/ichor/tests/products/productapi/...` ✓

## Failure Output

```
    apitest.go:57: duplicate-upc-code: Should receive a status code of 409 for the response : 200
--- FAIL: Test_InventoryProduct/create-409-duplicate-upc-code (0.01s)
```
