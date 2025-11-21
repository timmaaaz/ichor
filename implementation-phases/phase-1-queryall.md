# Phase 1: QueryAll Endpoints Implementation

**Status**: Optional but Recommended
**Priority**: Medium
**Estimated Time**: 3-4 hours
**Unblocks**: Frontend Phase 2 (Admin List Views)

---

## Overview

Add "list all" endpoints to three config domains (forms, page-configs, table-configs) to enable the frontend admin UI to display all configurations without requiring direct database queries.

### What You'll Build

Three new endpoints:
- `GET /v1/config/forms/all` - List all forms
- `GET /v1/config/page-configs/all` - List all page configurations
- `GET /v1/data/configs/all` - List all table configurations

### Pattern to Follow

Use the **purchaseorderstatusapi** pattern from `/api/domain/http/procurement/purchaseorderstatusapi/` as your reference implementation.

---

## Why This Pattern?

**QueryAll vs Query**:
- `Query`: Paginated, filtered, complex
- `QueryAll`: Simple, returns everything, no filters

**Use Cases**:
- Admin dropdowns (select a form, select a page config)
- Admin list views (show all configs)
- Reference data (forms, statuses, types)

**Key Differences from Paginated Query**:
- No `QueryParams` parsing
- No pagination (`query.Result[T]`)
- Returns wrapper type (`Entities []Entity`)
- Simple SQL: `SELECT * FROM table ORDER BY sort_column`
- Endpoint convention: `/entities/all` (not `/entities`)

---

## Implementation Steps

For each entity (Forms, Page Configs, Table Configs), follow these steps in order:

### Step 1: Business Layer (Add QueryAll to Storer)

**Files to Modify**:
1. `business/domain/config/formbus/formbus.go`
2. `business/domain/config/pageconfigbus/pageconfigbus.go`
3. `business/domain/databus/...` (table configs - may vary)

**Pattern**:

#### 1a. Add to Storer Interface

```go
type Storer interface {
    // ... existing methods
    QueryAll(ctx context.Context) ([]Form, error)  // Add this
}
```

#### 1b. Add Business Method

```go
func (b *Business) QueryAll(ctx context.Context) ([]Form, error) {
    ctx, span := otel.AddSpan(ctx, "business.formbus.queryall")
    defer span.End()

    forms, err := b.storer.QueryAll(ctx)
    if err != nil {
        return nil, fmt.Errorf("queryall: %w", err)
    }

    return forms, nil
}
```

**Key Points**:
- Add OpenTelemetry span for tracing
- Simple passthrough to storer
- Consistent error wrapping

---

### Step 2: Database Layer (Implement QueryAll)

**Files to Modify**:
1. `business/domain/config/formbus/stores/formdb/formdb.go`
2. `business/domain/config/pageconfigbus/stores/pageconfigdb/pageconfigdb.go`
3. `business/domain/databus/stores/...db/` (table configs)

**Pattern**:

```go
func (s *Store) QueryAll(ctx context.Context) ([]formbus.Form, error) {
    const q = `
    SELECT
        id,
        name,
        is_reference_data,
        allow_inline_create
    FROM
        config.forms
    ORDER BY
        name`

    var dbForms []dbForm
    if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, nil, &dbForms); err != nil {
        return nil, fmt.Errorf("namedqueryslice: %w", err)
    }

    return toBusForms(dbForms), nil
}
```

**SQL Guidelines**:
- Select all columns needed for the model
- Use simple `ORDER BY` (no complex joins)
- No WHERE clause (return everything)
- No LIMIT/OFFSET (not paginated)

**For Page Configs**:
```sql
SELECT
    id, name, user_id, is_default
FROM
    config.page_configs
ORDER BY
    name
```

**For Table Configs**:
```sql
SELECT
    id, name, description, config, created_by, updated_by, created_date, updated_date
FROM
    config.table_configs
ORDER BY
    name
```

---

### Step 3: Application Layer (Add QueryAll with Wrapper Type)

**Files to Modify**:
1. `app/domain/config/formapp/formapp.go`
2. `app/domain/config/formapp/model.go`
3. `app/domain/config/pageconfigapp/pageconfigapp.go`
4. `app/domain/config/pageconfigapp/model.go`
5. `app/domain/dataapp/dataapp.go`
6. `app/domain/dataapp/model.go`

**Pattern**:

#### 3a. Add Wrapper Type in `model.go`

```go
// Forms is a collection wrapper that implements the Encoder interface.
type Forms []Form

func (app Forms) Encode() ([]byte, string, error) {
    data, err := json.Marshal(app)
    return data, "application/json", err
}
```

**Why?** The wrapper type implements `web.Encoder` interface, required by API layer.

#### 3b. Add App Method in `{entity}app.go`

```go
func (a *App) QueryAll(ctx context.Context) (Forms, error) {
    forms, err := a.business.QueryAll(ctx)
    if err != nil {
        return nil, errs.Newf(errs.Internal, "queryall: %s", err)
    }

    return Forms(ToAppForms(forms)), nil
}
```

**Key Points**:
- Call business layer
- Convert bus models to app models
- Wrap in collection type
- Return typed wrapper (not `[]Form`)

---

### Step 4: API Layer (Add HTTP Handler)

**Files to Modify**:
1. `api/domain/http/config/formapi/formapi.go`
2. `api/domain/http/config/pageconfigapi/pageconfigapi.go`
3. `api/domain/http/dataapi/dataapi.go`

**Pattern**:

```go
func (api *api) queryAll(ctx context.Context, r *http.Request) web.Encoder {
    forms, err := api.formApp.QueryAll(ctx)
    if err != nil {
        return errs.NewError(err)
    }

    return forms  // Already implements web.Encoder
}
```

**Key Points**:
- No request decoding (GET endpoint, no body)
- Simple call to app layer
- Return wrapper type directly
- Error handling via `errs.NewError()`

---

### Step 5: Register Route

**Files to Modify**:
1. `api/domain/http/config/formapi/routes.go`
2. `api/domain/http/config/pageconfigapi/routes.go`
3. `api/domain/http/dataapi/routes.go`

**Pattern**:

```go
app.HandlerFunc(http.MethodGet, version, "/config/forms/all", api.queryAll, authen,
    mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))
```

**Authorization**:
- Use `auth.RuleAny` for any authenticated user
- Use `auth.RuleAdminOnly` if admin-only access required
- Action: `permissionsbus.Actions.Read`

**Endpoint Naming**:
- Pattern: `/{domain}/{entity}/all`
- Examples:
  - `/config/forms/all`
  - `/config/page-configs/all`
  - `/data/configs/all`

---

## Complete Implementation Checklist

### Forms Domain

- [ ] **Business Layer** (`formbus.go`)
  - [ ] Add `QueryAll(ctx context.Context) ([]Form, error)` to Storer interface
  - [ ] Add `QueryAll()` method to Business struct

- [ ] **Database Layer** (`formdb.go`)
  - [ ] Implement `QueryAll()` with SQL query
  - [ ] Order by name

- [ ] **Application Layer** (`formapp.go`, `model.go`)
  - [ ] Create `Forms` wrapper type in `model.go`
  - [ ] Implement `Encode()` for `Forms` type
  - [ ] Add `QueryAll()` method to App struct

- [ ] **API Layer** (`formapi.go`, `routes.go`)
  - [ ] Add `queryAll()` handler
  - [ ] Register `GET /v1/config/forms/all` route

---

### Page Configs Domain

- [ ] **Business Layer** (`pageconfigbus.go`)
  - [ ] Add `QueryAll(ctx context.Context) ([]PageConfig, error)` to Storer interface
  - [ ] Add `QueryAll()` method to Business struct

- [ ] **Database Layer** (`pageconfigdb.go`)
  - [ ] Implement `QueryAll()` with SQL query
  - [ ] Order by name

- [ ] **Application Layer** (`pageconfigapp.go`, `model.go`)
  - [ ] Create `PageConfigs` wrapper type in `model.go`
  - [ ] Implement `Encode()` for `PageConfigs` type
  - [ ] Add `QueryAll()` method to App struct

- [ ] **API Layer** (`pageconfigapi.go`, `routes.go`)
  - [ ] Add `queryAll()` handler
  - [ ] Register `GET /v1/config/page-configs/all` route

---

### Table Configs Domain

**Note**: Table configs may be in `dataapi` rather than `config` domain. Adjust paths accordingly.

- [ ] **Business Layer**
  - [ ] Add `QueryAll(ctx context.Context) ([]TableConfig, error)` to Storer interface
  - [ ] Add `QueryAll()` method to Business struct

- [ ] **Database Layer**
  - [ ] Implement `QueryAll()` with SQL query
  - [ ] Order by name

- [ ] **Application Layer**
  - [ ] Create `TableConfigs` wrapper type in `model.go`
  - [ ] Implement `Encode()` for `TableConfigs` type
  - [ ] Add `QueryAll()` method to App struct

- [ ] **API Layer**
  - [ ] Add `queryAll()` handler
  - [ ] Register `GET /v1/data/configs/all` route

---

## Testing

### Manual Testing

```bash
# Start the service
make dev-up
make dev-update-apply

# Get authentication token
export TOKEN=$(make token | grep -o '"token":"[^"]*' | cut -d'"' -f4)

# Test forms endpoint
curl -H "Authorization: Bearer $TOKEN" http://localhost:3000/v1/config/forms/all | jq

# Test page-configs endpoint
curl -H "Authorization: Bearer $TOKEN" http://localhost:3000/v1/config/page-configs/all | jq

# Test table-configs endpoint
curl -H "Authorization: Bearer $TOKEN" http://localhost:3000/v1/data/configs/all | jq
```

**Expected Response**:
```json
[
  {
    "id": "uuid-here",
    "name": "example_form",
    "isReferenceData": true,
    "allowInlineCreate": false
  },
  ...
]
```

### Integration Tests

**Location**: `api/cmd/services/ichor/tests/{domain}/{entity}api/`

**Test File**: Create `queryall_test.go`

**Pattern**:

```go
func queryAll200(sd apitest.SeedData) apitest.Table {
    table := apitest.Table{
        Name:   "queryall-200",
        URL:    "/v1/config/forms/all",
        Method: http.MethodGet,
        Token:  sd.Admins[0].Token,
        StatusCode: http.StatusOK,
    }

    return table
}
```

**Add to test suite**:

```go
func Test_FormAPI(t *testing.T) {
    test := apitest.StartTest(t, "formapi_test")
    sd := test.SeedData()

    // ... existing tests
    test.Run(t, queryAll200(sd), "queryall-200")
}
```

---

## Common Issues

### Issue 1: Wrapper Type Not Encoding

**Error**: `type Forms does not implement web.Encoder`

**Fix**: Ensure `Forms` type implements `Encode()`:
```go
type Forms []Form

func (app Forms) Encode() ([]byte, string, error) {
    data, err := json.Marshal(app)
    return data, "application/json", err
}
```

### Issue 2: Wrong Endpoint Pattern

**Error**: Route conflicts or not found

**Fix**: Use `/all` suffix:
- ✅ `/config/forms/all`
- ❌ `/config/forms` (conflicts with paginated query)

### Issue 3: Authorization Fails

**Error**: 403 Forbidden

**Fix**: Check authorization rule:
- Use `auth.RuleAny` for any authenticated user
- Verify `RouteTable` constant matches table name in `tableaccessbus`

### Issue 4: Empty Response

**Error**: Returns `[]` but database has records

**Fix**: Check SQL query:
- Verify table name and schema
- Check column names match struct tags
- Add logging: `log.Info(ctx, "queryall returned", "count", len(forms))`

---

## Reference Implementation

**Location**: `/api/domain/http/procurement/purchaseorderstatusapi/`

**Files to Review**:
- `purchaseorderstatusapi.go:138` - `queryAll()` handler
- `routes.go:55` - Route registration
- `../../app/domain/procurement/purchaseorderstatusapp/purchaseorderstatusapp.go:95` - App layer
- `../../business/domain/procurement/purchaseorderstatusbus/purchaseorderstatusbus.go:74` - Business layer
- `stores/purchaseorderstatusdb/purchaseorderstatusdb.go:121` - Database layer

---

## Success Criteria

### Functional
- [ ] All 3 endpoints return HTTP 200
- [ ] Response is valid JSON array
- [ ] Returns all records from database (no filtering)
- [ ] Records are sorted (by name or sort_order)

### Code Quality
- [ ] Follows Ardan Labs layering (bus → app → api)
- [ ] Implements `web.Encoder` interface
- [ ] Error handling consistent with codebase
- [ ] OpenTelemetry spans added

### Testing
- [ ] Manual testing passes for all 3 endpoints
- [ ] Integration tests created and passing
- [ ] Authorization works for admin and regular users

---

## Estimated Time Breakdown

- **Forms**: 1 hour
- **Page Configs**: 1 hour
- **Table Configs**: 1 hour
- **Testing**: 30 minutes
- **Debugging**: 30 minutes

**Total**: 3-4 hours

---

## Next Steps

After completing Phase 1:
1. Test all endpoints manually
2. Run integration tests: `make test`
3. Commit changes: `git commit -m "feat: add QueryAll endpoints for config entities"`
4. Move to [Phase 2: Introspection](phase-2-introspection.md) (critical for frontend Phase 4)

---

## Questions?

**Q: Can I skip this phase?**
A: Yes, frontend can query database directly as a workaround. But this is the cleaner solution.

**Q: Why not use the existing Query() method with empty filters?**
A: `Query()` requires pagination and is more complex. `QueryAll()` is simpler and more explicit for "get everything" use cases.

**Q: Should QueryAll be admin-only?**
A: Depends on sensitivity. For config data, `auth.RuleAny` (any authenticated user) is usually fine. For sensitive data, use `auth.RuleAdminOnly`.

**Q: What if QueryAll returns too many records?**
A: If you expect >1000 records, consider adding pagination. For config data (forms, page-configs), this is rarely an issue.

---

**Ready to implement?** Start with Forms domain, then replicate the pattern for Page Configs and Table Configs. Good luck!
