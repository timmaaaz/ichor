# Progress Summary: errs.md

## Overview
Ichor's error code system mapping gRPC-style status codes to HTTP statuses. Provides standardized error responses across the entire API.

## errs [sdk] — `app/sdk/errs/errs.go`

**Responsibility:** Define error codes, HTTP status mapping, and field validation errors.

### Key Facts
- **1,065 usages** of errs.New across 151 files in app/ and api/ (verified 2026-03-09)
- **616 usages** of errs.NewFieldsError across 138 files (verified 2026-03-09)
- **gRPC-style codes** mapped to HTTP statuses

### ErrCode Type
```go
type ErrCode struct{ value int }
```

### ErrCode Constants → HTTP Status Mapping

| Constant              | Value | HTTP Status |
|-----------------------|-------|-------------|
| OK                    | 0     | 200         |
| NoContent             | 1     | 204         |
| Canceled              | 2     | 504         |
| Unknown               | 3     | 500         |
| InvalidArgument       | 4     | 400         |
| DeadlineExceeded      | 5     | 504         |
| NotFound              | 6     | 404         |
| AlreadyExists         | 7     | 409         |
| PermissionDenied      | 8     | 403         |
| ResourceExhausted     | 9     | 429         |
| FailedPrecondition    | 10    | 400         |
| Aborted               | 11    | 409         |
| OutOfRange            | 12    | 400         |
| Unimplemented         | 13    | 501         |
| Internal              | 14    | 500         |
| Unavailable           | 15    | 503         |
| DataLoss              | 16    | 500         |
| Unauthenticated       | 17    | 401         |
| TooManyRequests       | 18    | 429         |
| InternalOnlyLog       | 19    | 500         |

### Most Common Patterns
- `errs.NewFieldsError("field", err)` — validation failure
- `errs.NotFound` — ErrCode for 404 responses
- `errs.InvalidArgument` — ErrCode for 400 responses
- `errs.Unauthenticated` — ErrCode for 401 responses

## FieldErrors [sdk]

**Field-level validation errors.**

```go
type FieldError struct {
    Field string `json:"field"`
    Err   string `json:"error"`
}

type FieldErrors []FieldError

func NewFieldsError(field string, err error) FieldErrors
```

### JSON Serialization
```json
[{"field":"name","error":"is required"}]
```

### Usage
- Returned from [app] layer validation
- Encoded in HTTP 400 response body
- FieldErrors implements error interface — return directly from app methods

## Change Patterns

### ⚠ Adding a New ErrCode
Affects 2 areas:
1. `app/sdk/errs/errs.go` — add constant with value + HTTP status mapping
2. `app/sdk/errs/errs.go` — add to httpStatus map or equivalent switch
3. **Note:** No [app] files need to change unless they reference the new code by name

### ⚠ Returning a New Validation Error from [app] Layer
Affects 2 areas:
1. `app/domain/{area}/{entity}app/{entity}app.go` — call errs.NewFieldsError in validation
2. `app/domain/{area}/{entity}app/model.go` — Validate() method on New{Entity}/Update{Entity}
3. **Note:** FieldErrors implements error interface — return directly from app methods

### ⚠ Changing FieldError Struct Shape
High impact — affects 138 call sites:
1. `app/sdk/errs/errs.go` — struct definition + JSON tags
2. ALL 138 files calling NewFieldsError — all callers need verification
3. **Frontend API client** — parses `field` + `error` keys in 400 responses; must be updated
4. **Verify first:** findReferences(app/sdk/errs/errs.go:129:6) — exact caller count before mass edit

## Critical Points
- Error codes follow gRPC status code conventions for consistency
- HTTP status mapping is deterministic — same ErrCode always produces same HTTP status
- FieldError JSON format is fixed: `field` and `error` keys (frontend expects this)
- **Most common codes:** NotFound (404), InvalidArgument (400), Unauthenticated (401), Internal (500)

## Notes for Future Development
Error handling is straightforward in practice:
- Adding new error codes is low-risk (add constant + mapping)
- Adding field validation is common (easy — just call errs.NewFieldsError)
- Changing FieldError struct is high-risk (138 call sites + frontend must update)

Most changes will be adding new validation errors (straightforward) rather than changing core error structures (risky).
