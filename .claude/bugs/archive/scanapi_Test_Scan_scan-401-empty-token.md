# Test Failure: Test_Scan/scan-401-empty-token

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/scanapi`
- **Duration**: 0s

## Failure Output

```
    apitest.go:73: DIFF
    apitest.go:74:   (*errs.Error)(
        - 	e"expected authorization header format: Bearer <token>",
        + 	e"error parsing token: token contains an invalid number of segments",
          )
    apitest.go:75: GOT
    apitest.go:76: &errs.Error{Code:errs.ErrCode{value:17}, Message:"expected authorization header format: Bearer <token>", FuncName:"", FileName:""}
    apitest.go:77: EXP
    apitest.go:78: &errs.Error{Code:errs.ErrCode{value:17}, Message:"error parsing token: token contains an invalid number of segments", FuncName:"github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/scanapi_test.scan401", FileName:"/Users/jaketimmer/src/work/superior/ichor/ichor/api/cmd/services/ichor/tests/inventory/scanapi/query_test.go:128"}
    apitest.go:79: Should get the expected response
--- FAIL: Test_Scan/scan-401-empty-token (0.00s)
```

## Fix
- **File**: `api/cmd/services/ichor/tests/inventory/scanapi/query_test.go:128`
- **Classification**: test bug
- **Change**: Updated ExpResp from "error parsing token: token contains an invalid number of segments" to "expected authorization header format: Bearer <token>" — `Token: ""` produces `"Bearer "` which HTTP trims to `"Bearer"`, failing the Bearer prefix check before JWT parsing.
- **Verified**: `go test -v -run Test_Scan/scan-401 ./api/cmd/services/ichor/tests/inventory/scanapi/` ✓
