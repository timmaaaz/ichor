# errs

[sdk]=shared SDK [app]=application layer
→=depends on

---

## Overview

App-layer error code system mapping gRPC-style status codes to HTTP statuses.
Used for structured validation errors (FieldErrors) and HTTP response codes.
496+ importers across app/ and api/ layers.

---

## errs [sdk]

file: app/sdk/errs/errs.go

### ErrCode

```go
type ErrCode struct{ value int }
```

### ErrCode Constants → HTTP Status

| Constant          | value | HTTP Status |
|-------------------|-------|-------------|
| OK                | 0     | 200         |
| NoContent         | 1     | 204         |
| Canceled          | 2     | 504         |
| Unknown           | 3     | 500         |
| InvalidArgument   | 4     | 400         |
| DeadlineExceeded  | 5     | 504         |
| NotFound          | 6     | 404         |
| AlreadyExists     | 7     | 409         |
| PermissionDenied  | 8     | 403         |
| ResourceExhausted | 9     | 429         |
| FailedPrecondition| 10    | 400         |
| Aborted           | 11    | 409         |
| OutOfRange        | 12    | 400         |
| Unimplemented     | 13    | 501         |
| Internal          | 14    | 500         |
| Unavailable       | 15    | 503         |
| DataLoss          | 16    | 500         |
| Unauthenticated   | 17    | 401         |
| TooManyRequests   | 18    | 429         |
| InternalOnlyLog   | 19    | 500         |

---

## FieldErrors

```go
type FieldError struct {
    Field string `json:"field"`
    Err   string `json:"error"`
}

type FieldErrors []FieldError

func NewFieldsError(field string, err error) FieldErrors
```

JSON serialization: `[{"field":"name","error":"is required"}]`

Usage: returned from [app] layer validation, encoded in HTTP 400 response body.

---

## Usage Scope

~203 files in app/domain/ import errs. Most common patterns:
  errs.NewFieldsError("field", err)         — validation failure
  errs.NotFound                             — ErrCode for 404 responses
  errs.InvalidArgument                      — ErrCode for 400 responses
  errs.Unauthenticated                      — ErrCode for 401 responses

---

## ⚠ Adding a new ErrCode

  app/sdk/errs/errs.go          (add constant with value + HTTP status mapping)
  app/sdk/errs/errs.go          (add to httpStatus map or equivalent switch)
  Note: no [app] files need to change unless they reference the new code by name

## ⚠ Returning a new validation error from [app] layer

  app/domain/{area}/{entity}app/{entity}app.go   (call errs.NewFieldsError in validation)
  app/domain/{area}/{entity}app/model.go         (Validate() method on New{Entity}/Update{Entity})
  Note: FieldErrors implements error interface — return directly from app methods

## ⚠ Changing FieldError struct shape

  app/sdk/errs/errs.go                 (struct definition + JSON tags)
  ALL ~203 app/domain/ files           (all callers of NewFieldsError — check JSON consumer compatibility)
  Frontend API client                  (parses `field` + `error` keys in 400 responses)
