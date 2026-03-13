# Test Failure: Test_ProductUOM/create-200-basic

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/products/productuomapi`
- **Duration**: 0.02s

## Failure Output

```
    apitest.go:57: basic: Should receive a status code of 200 for the response : 409
--- FAIL: Test_ProductUOM/create-200-basic (0.02s)
```

## Fix
- **File**: `api/cmd/services/ichor/tests/products/productuomapi/create_test.go:26,36`
- **Classification**: test bug
- **Change**: Changed `IsBase: true` → `IsBase: false` in both `Input` and `ExpResp` of `create-200-basic`. The `products.product_uoms` table has a partial unique index `(product_id) WHERE is_base = TRUE`, and the seed randomly assigns the base UOM to one of 5 products (~20% chance of hitting `Products[0]`), causing a flaky 409.
- **Verified**: `go test -v -run "Test_ProductUOM/create-200-basic" github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/products/productuomapi` ✓
