# Synopsis: camelCase Usage in Export/Import APIs

## Part 1: Thorough Research Findings

---

## Executive Summary

The export/import/validate endpoints for page-configs use **camelCase** in their JSON payloads, which differs from the **snake_case** convention used by regular CRUD endpoints. This document provides comprehensive analysis of why this exists and options for addressing it.

---

## The Three Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/v1/config/page-configs/validate` | POST | Validate a page config blob before import |
| `/v1/config/page-configs/import-blob` | POST | Import a page config blob |
| `/v1/config/page-configs/{id}/export` | GET | Export a page config by ID |

---

## Current JSON Casing Analysis

### 1. Export Response (camelCase)

The export endpoint returns the **business layer model directly**, bypassing app layer conversion:

```go
// pageconfigapi.go:221-240
func (api *api) exportBlob(ctx context.Context, r *http.Request) web.Encoder {
    pkg, err := api.pageConfigApp.ExportBlob(ctx, configID)
    data, err := json.Marshal(pkg)  // Marshals business layer type directly
    return rawJSON(data)
}
```

This results in **camelCase** output from the business layer models:

```json
{
    "pageConfig": {
        "id": "...",
        "name": "...",
        "userId": "...",        // camelCase
        "isDefault": true       // camelCase
    },
    "contents": [{
        "id": "...",
        "pageConfigId": "...",   // camelCase
        "contentType": "...",    // camelCase
        "tableConfigId": "...",  // camelCase
        "formId": "...",         // camelCase
        "chartConfigId": "...",  // camelCase
        "orderIndex": 1,         // camelCase
        "parentId": "...",       // camelCase
        "isVisible": true,       // camelCase
        "isDefault": false       // camelCase
    }],
    "actions": {
        "buttons": [{
            "pageConfigId": "...",         // camelCase
            "actionType": "...",           // camelCase
            "actionOrder": 1,              // camelCase
            "isActive": true,              // camelCase
            "button": {
                "targetPath": "...",       // camelCase
                "confirmationPrompt": ""   // camelCase
            }
        }]
    }
}
```

### 2. Import/Validate Input (camelCase)

The import and validate endpoints accept the **same business layer format**:

```go
// Test file shows expected input format (export_import_test.go:139-151)
json.RawMessage(`{
    "PageConfig": {
        "Name": "...",
        "UserID": "...",
        "IsDefault": true
    },
    "Contents": [],
    "Actions": {...}
}`)
```

### 3. Validate/Import Responses (snake_case)

The response types use snake_case because they go through the app layer:

```go
// pageconfigapp/model.go:507-511
type ImportStats struct {
    ImportedCount int `json:"imported_count"`   // snake_case
    UpdatedCount  int `json:"updated_count"`    // snake_case
    SkippedCount  int `json:"skipped_count"`    // snake_case
}
```

---

## Why camelCase Exists Here

### Root Cause: Business Layer Direct Serialization

The business layer models in [pageconfigbus/model.go](business/domain/config/pageconfigbus/model.go) use camelCase JSON tags:

```go
// pageconfigbus/model.go:10-15
type PageConfig struct {
    ID        uuid.UUID `json:"id"`
    Name      string    `json:"name"`
    UserID    uuid.UUID `json:"userId"`      // camelCase
    IsDefault bool      `json:"isDefault"`   // camelCase
}
```

**Contrast with the standard CRUD models** in the same file:

```go
// pageconfigbus/model.go:18-22
type NewPageConfig struct {
    Name      string    `json:"name"`
    UserID    uuid.UUID `json:"user_id"`     // snake_case
    IsDefault bool      `json:"is_default"`  // snake_case
}
```

### Historical Reasoning (Likely)

1. **Round-trip consistency**: Export output = Import input. If export uses camelCase, import must accept camelCase.
2. **Business layer proximity**: These models were designed to mirror internal Go structs more closely.
3. **Different use case**: Export/import is for config portability (admin tooling), not regular API consumers.

---

## The Inconsistency Problem

### Within pageconfigbus Itself

The same file has **mixed conventions**:

| Model | Field | JSON Tag |
|-------|-------|----------|
| `PageConfig` (main entity) | `UserID` | `"userId"` (camelCase) |
| `NewPageConfig` (create) | `UserID` | `"user_id"` (snake_case) |
| `UpdatePageConfig` (update) | `UserID` | `"user_id"` (snake_case) |
| `PageContentExport` | `PageConfigID` | `"pageConfigId"` (camelCase) |
| `ImportStats` | `ImportedCount` | `"imported_count"` (snake_case) |

### Compared to Other Domains

All other domains consistently use **snake_case** in app layer responses:
- `userapp.User` → `"user_id"`, `"is_admin"`
- `assetapp.Asset` → `"asset_type_id"`, `"is_active"`
- `orderapp.Order` → `"order_date"`, `"customer_id"`

---

## Files Affected

### Business Layer (camelCase currently)
- [business/domain/config/pageconfigbus/model.go](business/domain/config/pageconfigbus/model.go) - Lines 10-95
  - `PageConfig` struct
  - `PageConfigWithRelations` struct
  - `PageContentExport` struct
  - `PageActionsExport` struct
  - `PageActionExport` struct
  - `ButtonActionExport` struct
  - `DropdownActionExport` struct
  - `DropdownItemExport` struct

### App Layer (mixed - has both conventions)
- [app/domain/config/pageconfigapp/model.go](app/domain/config/pageconfigapp/model.go) - Lines 169-232
  - `PageConfigPackage` struct (camelCase)
  - `PageContentApp` struct (camelCase)
  - `PageActionsApp` struct (camelCase)
  - `PageActionApp` struct (camelCase)
  - `ButtonActionApp` struct (camelCase)
  - `DropdownActionApp` struct (camelCase)
  - `DropdownItemApp` struct (camelCase)

### API Layer
- [api/domain/http/config/pageconfigapi/pageconfigapi.go](api/domain/http/config/pageconfigapi/pageconfigapi.go) - Lines 220-240
  - `exportBlob` handler directly marshals business type

### Tests
- [api/cmd/services/ichor/tests/config/pageconfigapi/export_import_test.go](api/cmd/services/ichor/tests/config/pageconfigapi/export_import_test.go)
  - Lines 139-192: Test blobs use `"PageConfig"`, `"UserID"`, `"IsDefault"` (PascalCase in test JSON)

---

## Recommendation: Standardize to snake_case

### Why Change to snake_case

1. **API consistency**: All other endpoints use snake_case
2. **Frontend alignment**: Your frontend likely expects snake_case
3. **Industry convention**: REST APIs typically use snake_case for JSON
4. **Maintainability**: One convention is easier to document and enforce

### Migration Strategy

1. **Update business layer models** in `pageconfigbus/model.go` to use snake_case JSON tags
2. **Update app layer export models** in `pageconfigapp/model.go` to use snake_case
3. **Ensure export handler** routes through proper app layer conversion (not direct business marshal)
4. **Update tests** in `export_import_test.go` to use snake_case input
5. **Coordinate with frontend** - this is a breaking change for any existing exports

### Breaking Change Warning

If there are existing exported configs in the wild using camelCase, they will become incompatible. Consider:
- Adding a migration path (accept both formats temporarily)
- Versioning the export format
- Documenting the change for users

---

## Part 2: Options for Resolution

---

## Option A: Keep camelCase (Do Nothing)

### Rationale
- Export/import is a specialized admin feature, not a general API consumer feature
- Round-trip consistency is preserved (export → import works without transformation)
- No breaking changes for any existing exported configs
- These endpoints are intentionally different - they're "file format" not "API response"

### Drawbacks
- Inconsistent with rest of API
- Frontend must handle two conventions
- Confusing for developers onboarding

### Effort: None

---

## Option B: Standardize to snake_case (Recommended)

### Rationale
- Consistent with 100% of other API endpoints
- Frontend only needs one JSON casing strategy
- Follows REST API conventions
- Matches recent snake_case standardization effort across the codebase

### Implementation Steps

1. **Update business layer export models** (`pageconfigbus/model.go`)
   - Change JSON tags on `PageConfig`, `PageConfigWithRelations`, `PageContentExport`, etc.
   - ~30 JSON tag changes

2. **Update app layer export models** (`pageconfigapp/model.go`)
   - Change JSON tags on `PageConfigPackage`, `PageContentApp`, `PageActionsApp`, etc.
   - ~25 JSON tag changes

3. **Refactor export handler** (`pageconfigapi.go`)
   - Currently: Marshals business type directly → camelCase
   - Change to: Use app layer conversion function → snake_case
   - The conversion function `toAppPageConfigWithRelations()` already exists but is unused

4. **Update tests** (`export_import_test.go`)
   - Update all test JSON blobs to use snake_case
   - ~8 test blobs need updating

### Breaking Change Mitigation
- This is a breaking change for any saved exports
- Option: Accept both formats during import for one release cycle

### Effort: Medium (~2-3 hours)

---

## Option C: Hybrid - Accept Both, Output snake_case

### Rationale
- Provides backwards compatibility for existing exports
- New exports use correct convention
- Gradual migration path

### Implementation
- Add custom JSON unmarshaler that accepts both `userId` and `user_id`
- Export only outputs snake_case
- Deprecate camelCase input in documentation

### Drawbacks
- More complex code
- Perpetuates inconsistency in import format

### Effort: High (~4-5 hours)

---

## Recommendation

**Option B (Standardize to snake_case)** is recommended because:

1. The codebase recently underwent snake_case standardization (commit `a4e9d18`)
2. Export/import is an admin feature with limited usage
3. Any saved exports are likely recent and can be re-exported
4. Clean, consistent API is worth the breaking change

---

## Detailed Implementation Plan (If Option B Selected)

### Phase 1: Business Layer (`pageconfigbus/model.go`)

```go
// Before
type PageConfig struct {
    UserID    uuid.UUID `json:"userId"`
    IsDefault bool      `json:"isDefault"`
}

// After
type PageConfig struct {
    UserID    uuid.UUID `json:"user_id"`
    IsDefault bool      `json:"is_default"`
}
```

Full list of changes:
| Struct | Field | Current | New |
|--------|-------|---------|-----|
| `PageConfig` | `UserID` | `userId` | `user_id` |
| `PageConfig` | `IsDefault` | `isDefault` | `is_default` |
| `PageConfigWithRelations` | `PageConfig` | `pageConfig` | `page_config` |
| `PageContentExport` | `PageConfigID` | `pageConfigId` | `page_config_id` |
| `PageContentExport` | `ContentType` | `contentType` | `content_type` |
| `PageContentExport` | `TableConfigID` | `tableConfigId` | `table_config_id` |
| `PageContentExport` | `FormID` | `formId` | `form_id` |
| `PageContentExport` | `ChartConfigID` | `chartConfigId` | `chart_config_id` |
| `PageContentExport` | `OrderIndex` | `orderIndex` | `order_index` |
| `PageContentExport` | `ParentID` | `parentId` | `parent_id` |
| `PageContentExport` | `IsVisible` | `isVisible` | `is_visible` |
| `PageContentExport` | `IsDefault` | `isDefault` | `is_default` |
| `PageActionExport` | `PageConfigID` | `pageConfigId` | `page_config_id` |
| `PageActionExport` | `ActionType` | `actionType` | `action_type` |
| `PageActionExport` | `ActionOrder` | `actionOrder` | `action_order` |
| `PageActionExport` | `IsActive` | `isActive` | `is_active` |
| `ButtonActionExport` | `TargetPath` | `targetPath` | `target_path` |
| `ButtonActionExport` | `ConfirmationPrompt` | `confirmationPrompt` | `confirmation_prompt` |
| `DropdownActionExport` | (none - all single word) | - | - |
| `DropdownItemExport` | `TargetPath` | `targetPath` | `target_path` |
| `DropdownItemExport` | `ItemOrder` | `itemOrder` | `item_order` |

### Phase 2: App Layer (`pageconfigapp/model.go`)

Same pattern - update all export-related structs to use snake_case JSON tags.

### Phase 3: API Handler (`pageconfigapi.go`)

```go
// Before (line 220-240)
func (api *api) exportBlob(ctx context.Context, r *http.Request) web.Encoder {
    pkg, err := api.pageConfigApp.ExportBlob(ctx, configID)
    data, err := json.Marshal(pkg)  // Business type → camelCase
    return rawJSON(data)
}

// After
func (api *api) exportBlob(ctx context.Context, r *http.Request) web.Encoder {
    pkg, err := api.pageConfigApp.ExportBlob(ctx, configID)
    appPkg := toAppPageConfigWithRelations(pkg)  // Convert to app type
    return appPkg  // App type implements Encoder → snake_case
}
```

### Phase 4: Tests (`export_import_test.go`)

Update all test blobs:
```go
// Before
json.RawMessage(`{
    "PageConfig": {
        "Name": "...",
        "UserID": "...",
        "IsDefault": true
    }
}`)

// After
json.RawMessage(`{
    "page_config": {
        "name": "...",
        "user_id": "...",
        "is_default": true
    }
}`)
```

### Phase 5: Validation Helpers (`validation_helpers.go`)

Update the `layoutJSON` struct used for validation:
```go
// Line 139-140
type layoutJSON struct {
    ColSpan *ResponsiveValue `json:"col_span,omitempty"`  // Was colSpan
    RowSpan *int             `json:"row_span,omitempty"`  // Was rowSpan
}
```

---

## Verification

After implementation:
1. Run `make test` - all tests should pass
2. Test export endpoint: `GET /v1/config/page-configs/{id}/export`
   - Response should use snake_case
3. Test import endpoint: `POST /v1/config/page-configs/import-blob`
   - Should accept snake_case input
4. Test validate endpoint: `POST /v1/config/page-configs/validate`
   - Should accept snake_case input

---

## Files to Modify (Complete List)

| File | Purpose |
|------|---------|
| `business/domain/config/pageconfigbus/model.go` | Business layer export models |
| `business/domain/config/pageconfigbus/validation_helpers.go` | Layout validation struct |
| `app/domain/config/pageconfigapp/model.go` | App layer export models |
| `api/domain/http/config/pageconfigapi/pageconfigapi.go` | Export handler |
| `api/cmd/services/ichor/tests/config/pageconfigapi/export_import_test.go` | Tests |

---

## Implementation Checklist (Detailed for Fresh Prompt Execution)

### Phase 1: Business Layer Models

**File: `business/domain/config/pageconfigbus/model.go`**

- [ ] **Task 1.1**: Update `PageConfig` struct (lines 10-15)
  ```go
  // Change:
  UserID    uuid.UUID `json:"userId"`
  IsDefault bool      `json:"isDefault"`
  // To:
  UserID    uuid.UUID `json:"user_id"`
  IsDefault bool      `json:"is_default"`
  ```

- [ ] **Task 1.2**: Update `PageConfigWithRelations` struct (lines 31-36)
  ```go
  // Change:
  PageConfig PageConfig          `json:"pageConfig"`
  // To:
  PageConfig PageConfig          `json:"page_config"`
  ```

- [ ] **Task 1.3**: Update `PageContentExport` struct (lines 38-52)
  ```go
  // Change these fields:
  PageConfigID  uuid.UUID       `json:"pageConfigId"`   -> `json:"page_config_id"`
  ContentType   string          `json:"contentType"`    -> `json:"content_type"`
  TableConfigID uuid.UUID       `json:"tableConfigId"`  -> `json:"table_config_id"`
  FormID        uuid.UUID       `json:"formId"`         -> `json:"form_id"`
  ChartConfigID uuid.UUID       `json:"chartConfigId"`  -> `json:"chart_config_id"`
  OrderIndex    int             `json:"orderIndex"`     -> `json:"order_index"`
  ParentID      uuid.UUID       `json:"parentId"`       -> `json:"parent_id"`
  IsVisible     bool            `json:"isVisible"`      -> `json:"is_visible"`
  IsDefault     bool            `json:"isDefault"`      -> `json:"is_default"`
  ```

- [ ] **Task 1.4**: Update `PageActionExport` struct (lines 62-70)
  ```go
  // Change these fields:
  PageConfigID uuid.UUID `json:"pageConfigId"`  -> `json:"page_config_id"`
  ActionType   string    `json:"actionType"`    -> `json:"action_type"`
  ActionOrder  int       `json:"actionOrder"`   -> `json:"action_order"`
  IsActive     bool      `json:"isActive"`      -> `json:"is_active"`
  ```

- [ ] **Task 1.5**: Update `ButtonActionExport` struct (lines 73-80)
  ```go
  // Change these fields:
  TargetPath         string `json:"targetPath"`         -> `json:"target_path"`
  ConfirmationPrompt string `json:"confirmationPrompt"` -> `json:"confirmation_prompt"`
  ```

- [ ] **Task 1.6**: Update `DropdownItemExport` struct (lines 90-95)
  ```go
  // Change these fields:
  TargetPath string    `json:"targetPath"`  -> `json:"target_path"`
  ItemOrder  int       `json:"itemOrder"`   -> `json:"item_order"`
  ```

**File: `business/domain/config/pageconfigbus/validation_helpers.go`**

- [ ] **Task 1.7**: Update `layoutJSON` struct (lines 138-142)
  ```go
  // Change:
  ColSpan *ResponsiveValue `json:"colSpan,omitempty"`
  RowSpan *int             `json:"rowSpan,omitempty"`
  ColStart *int            `json:"colStart,omitempty"`
  RowStart *int            `json:"rowStart,omitempty"`
  GridCols *ResponsiveValue `json:"gridCols,omitempty"`
  ClassName string          `json:"className,omitempty"`
  ContainerType string      `json:"containerType,omitempty"`
  // To:
  ColSpan *ResponsiveValue `json:"col_span,omitempty"`
  RowSpan *int             `json:"row_span,omitempty"`
  ColStart *int            `json:"col_start,omitempty"`
  RowStart *int            `json:"row_start,omitempty"`
  GridCols *ResponsiveValue `json:"grid_cols,omitempty"`
  ClassName string          `json:"class_name,omitempty"`
  ContainerType string      `json:"container_type,omitempty"`
  ```

---

### Phase 2: App Layer Models

**File: `app/domain/config/pageconfigapp/model.go`**

- [ ] **Task 2.1**: Update `PageConfigPackage` struct (lines 169-174)
  ```go
  // Change:
  PageConfig PageConfig        `json:"pageConfig"`
  // To:
  PageConfig PageConfig        `json:"page_config"`
  ```

- [ ] **Task 2.2**: Update `PageContentApp` struct (lines 176-189)
  ```go
  // Change these fields:
  PageConfigID  string `json:"pageConfigId"`   -> `json:"page_config_id"`
  ContentType   string `json:"contentType"`    -> `json:"content_type"`
  TableConfigID string `json:"tableConfigId"`  -> `json:"table_config_id"`
  FormID        string `json:"formId"`         -> `json:"form_id"`
  OrderIndex    int    `json:"orderIndex"`     -> `json:"order_index"`
  ParentID      string `json:"parentId"`       -> `json:"parent_id"`
  IsVisible     bool   `json:"isVisible"`      -> `json:"is_visible"`
  IsDefault     bool   `json:"isDefault"`      -> `json:"is_default"`
  ```

- [ ] **Task 2.3**: Update `PageActionApp` struct (lines 198-207)
  ```go
  // Change these fields:
  PageConfigID string `json:"pageConfigId"`  -> `json:"page_config_id"`
  ActionType   string `json:"actionType"`    -> `json:"action_type"`
  ActionOrder  int    `json:"actionOrder"`   -> `json:"action_order"`
  IsActive     bool   `json:"isActive"`      -> `json:"is_active"`
  ```

- [ ] **Task 2.4**: Update `ButtonActionApp` struct (lines 209-217)
  ```go
  // Change these fields:
  TargetPath         string `json:"targetPath"`         -> `json:"target_path"`
  ConfirmationPrompt string `json:"confirmationPrompt"` -> `json:"confirmation_prompt"`
  ```

- [ ] **Task 2.5**: Update `DropdownItemApp` struct (lines 227-232)
  ```go
  // Change these fields:
  TargetPath string `json:"targetPath"`  -> `json:"target_path"`
  ItemOrder  int    `json:"itemOrder"`   -> `json:"item_order"`
  ```

- [ ] **Task 2.6**: Add `Encode()` method to `PageConfigPackage` if not present
  ```go
  func (app PageConfigPackage) Encode() ([]byte, string, error) {
      data, err := json.Marshal(app)
      return data, "application/json", err
  }
  ```

---

### Phase 3: API Handler Refactor

**File: `api/domain/http/config/pageconfigapi/pageconfigapi.go`**

- [ ] **Task 3.1**: Modify `exportBlob` handler (lines 220-240)

  **Current code:**
  ```go
  func (api *api) exportBlob(ctx context.Context, r *http.Request) web.Encoder {
      configID, err := uuid.Parse(web.Param(r, "config_id"))
      if err != nil {
          return errs.New(errs.InvalidArgument, err)
      }

      pkg, err := api.pageConfigApp.ExportBlob(ctx, configID)
      if err != nil {
          return errs.NewError(err)
      }

      data, err := json.Marshal(pkg)
      if err != nil {
          return errs.New(errs.Internal, err)
      }

      return rawJSON(data)
  }
  ```

  **Change to:**
  ```go
  func (api *api) exportBlob(ctx context.Context, r *http.Request) web.Encoder {
      configID, err := uuid.Parse(web.Param(r, "config_id"))
      if err != nil {
          return errs.New(errs.InvalidArgument, err)
      }

      pkg, err := api.pageConfigApp.ExportBlobAsApp(ctx, configID)
      if err != nil {
          return errs.NewError(err)
      }

      return pkg  // PageConfigPackage now implements Encode()
  }
  ```

- [ ] **Task 3.2**: Add `ExportBlobAsApp` method to `pageconfigapp.App` if needed, or modify `ExportBlob` to return app type

**File: `app/domain/config/pageconfigapp/pageconfigapp.go`**

- [ ] **Task 3.3**: Update or add method that returns `PageConfigPackage` instead of business type
  ```go
  func (a *App) ExportBlobAsApp(ctx context.Context, configID uuid.UUID) (PageConfigPackage, error) {
      busPkg, err := a.pageConfigBus.ExportByIDs(ctx, []uuid.UUID{configID})
      if err != nil {
          return PageConfigPackage{}, err
      }
      if len(busPkg) == 0 {
          return PageConfigPackage{}, errs.Newf(errs.NotFound, "page config not found")
      }
      return toAppPageConfigWithRelations(busPkg[0]), nil
  }
  ```

---

### Phase 4: Update Tests

**File: `api/cmd/services/ichor/tests/config/pageconfigapi/export_import_test.go`**

- [ ] **Task 4.1**: Update `skipBlob` test JSON (around line 139)
  ```go
  // Change from:
  skipBlob := json.RawMessage(`{
      "PageConfig": {
          "Name": "Import Test Page ...",
          "UserID": "00000000-0000-0000-0000-000000000000",
          "IsDefault": true
      },
      "Contents": [],
      "Actions": {
          "Buttons": [],
          "Dropdowns": [],
          "Separators": []
      }
  }`)

  // To:
  skipBlob := json.RawMessage(`{
      "page_config": {
          "name": "Import Test Page ...",
          "user_id": "00000000-0000-0000-0000-000000000000",
          "is_default": true
      },
      "contents": [],
      "actions": {
          "buttons": [],
          "dropdowns": [],
          "separators": []
      }
  }`)
  ```

- [ ] **Task 4.2**: Update `replaceBlob` test JSON (around line 153)
- [ ] **Task 4.3**: Update `mergeBlob` test JSON (around line 167)
- [ ] **Task 4.4**: Update `defaultBlob` test JSON (around line 181)
- [ ] **Task 4.5**: Update `import400` missing-name test JSON (around line 292)
  ```go
  // Change from:
  Input: json.RawMessage(`{"PageConfig":{},"Contents":[],"Actions":{"Buttons":[],"Dropdowns":[],"Separators":[]}}`),

  // To:
  Input: json.RawMessage(`{"page_config":{},"contents":[],"actions":{"buttons":[],"dropdowns":[],"separators":[]}}`),
  ```

- [ ] **Task 4.6**: Update `import401` test blob (around line 312)
  ```go
  // Change from:
  blob := json.RawMessage(`{"PageConfig":{"Name":"Test"},"Contents":[],"Actions":{"Buttons":[],"Dropdowns":[],"Separators":[]}}`)

  // To:
  blob := json.RawMessage(`{"page_config":{"name":"Test"},"contents":[],"actions":{"buttons":[],"dropdowns":[],"separators":[]}}`)
  ```

- [ ] **Task 4.7**: Update export test `GotResp` type if needed (line 46)
  - Verify `PageConfigPackage` fields now use snake_case in assertions

---

### Phase 5: Verification

- [ ] **Task 5.1**: Run `make test` - all tests should pass
- [ ] **Task 5.2**: Run `go build ./...` - verify no compilation errors
- [ ] **Task 5.3**: Start local server and test export endpoint manually
  ```bash
  curl -H "Authorization: Bearer $TOKEN" \
    http://localhost:3000/v1/config/page-configs/id/{config_id}/export
  ```
  - Verify response uses snake_case (`page_config`, `user_id`, etc.)

- [ ] **Task 5.4**: Test import endpoint manually
  ```bash
  curl -X POST -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"page_config":{"name":"Test"},"contents":[],"actions":{"buttons":[],"dropdowns":[],"separators":[]}}' \
    http://localhost:3000/v1/config/page-configs/import-blob?mode=skip
  ```
  - Verify it accepts snake_case input

- [ ] **Task 5.5**: Test validate endpoint manually
  ```bash
  curl -X POST -H "Content-Type: application/json" \
    -d '{"page_config":{"name":"Test"},"contents":[],"actions":{"buttons":[],"dropdowns":[],"separators":[]}}' \
    http://localhost:3000/v1/config/page-configs/validate
  ```
  - Verify it accepts snake_case input

---

## Part 3: Layout JSON - Intentionally Excluded from snake_case Conversion

---

### Decision: Layout JSON Remains camelCase

The `Layout` field in `PageContent` uses **camelCase** JSON keys (`colSpan`, `containerType`, etc.) and will **NOT** be converted to snake_case. This is an intentional architectural decision, not technical debt.

---

### Why Layout JSON is Different

#### 1. It's Frontend Configuration, Not an API Contract

Layout JSON is **stored frontend state** that happens to pass through the backend. The data flow is:

```
Frontend creates layout → Backend stores as opaque JSONB → Frontend reads and renders
```

The backend never interprets these keys - it just validates structure and stores them. This is fundamentally different from API request/response bodies where the backend defines the contract.

#### 2. camelCase is Idiomatic for CSS/JS Configuration

The Layout JSON keys map directly to CSS/Tailwind concepts:

| JSON Key | Maps To |
|----------|---------|
| `colSpan` | CSS `grid-column` / Tailwind `col-span-*` |
| `rowSpan` | CSS `grid-row` / Tailwind `row-span-*` |
| `gridCols` | CSS `grid-template-columns` / Tailwind `grid-cols-*` |
| `className` | HTML/React `className` attribute |
| `containerType` | Component variant prop |

These follow JavaScript/CSS naming conventions where camelCase is standard. Using `col_span` would be inconsistent with the ecosystem the frontend operates in.

#### 3. Only the Frontend Consumes This Data

No external API consumers parse Layout JSON. It's written by the frontend, stored by the backend, and read by the same frontend. Converting to snake_case would require the frontend to transform keys on every read/write for no benefit.

---

### What Remains camelCase

**Files with LayoutConfig structs (JSON tags stay camelCase):**
- `business/domain/config/pagecontentbus/model.go` (lines 89-113)
- `business/domain/config/pageconfigbus/validation_helpers.go` (lines 138-142)
- `business/sdk/tablebuilder/model.go` (lines 566-590)

**Seed data and tests with Layout JSON strings (stay camelCase):**
- `business/sdk/dbtest/seedFrontend.go` (~65 occurrences)
- `business/domain/config/pagecontentbus/testutil.go`
- `api/cmd/services/ichor/tests/config/pagecontentapi/*.go`
- `business/domain/config/pagecontentbus/pagecontentbus_test.go`
- `business/domain/config/pageconfigbus/export_import_test.go`
- `business/domain/config/pageconfigbus/validation_helpers_test.go`
- `api/domain/http/config/pageconfigapi/pageconfig_validate_test.go`

**The specific JSON keys that remain camelCase:**
- `colSpan`, `rowSpan`, `colStart`, `rowStart`
- `gridCols`, `className`, `containerType`

---

### Summary: Files to Modify (Final Scope)

| File | Changes |
|------|---------|
| `business/domain/config/pageconfigbus/model.go` | ~30 JSON tag changes |
| `business/domain/config/pageconfigbus/validation_helpers.go` | Layout validation struct JSON tags stay camelCase; only export model tags change |
| `app/domain/config/pageconfigapp/model.go` | ~25 JSON tag changes |
| `api/domain/http/config/pageconfigapi/pageconfigapi.go` | Handler refactor |
| `api/cmd/services/ichor/tests/config/pageconfigapi/export_import_test.go` | ~8 test blob updates |
