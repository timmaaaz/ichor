# Phase 3: Import/Export Endpoints

**Status**: CRITICAL - Hard Blocker
**Priority**: High
**Estimated Time**: 6-8 hours
**Unblocks**: Frontend Phase 8 (JSON Import/Export & Templates)

---

## Overview

Add bulk export/import capabilities to three config domains (forms, page-configs, table-configs) to enable the frontend to:
- Export configurations as JSON packages
- Import configurations from JSON (with conflict resolution)
- Create template libraries
- Backup/restore configurations
- Share configurations between environments

### What You'll Build

**6 Endpoints**:
1. `POST /v1/config/forms/export` - Export forms (with fields)
2. `POST /v1/config/forms/import` - Import forms
3. `POST /v1/config/page-configs/export` - Export page configs (with content & actions)
4. `POST /v1/config/page-configs/import` - Import page configs
5. `POST /v1/data/configs/export` - Export table configs
6. `POST /v1/data/configs/import` - Import table configs

**Key Features**:
- Export includes related records (forms + fields, page-configs + content/actions)
- Import handles conflict resolution (merge/skip/replace modes)
- Transactional imports (all-or-nothing)
- ID remapping for foreign key relationships
- Validation and error handling

---

## Why Import/Export?

**Frontend Use Cases**:
- **Templates**: Share common configurations (e.g., "Standard Sales Dashboard")
- **Backup**: Export before making changes
- **Deployment**: Move configs from dev → staging → production
- **Collaboration**: Share configs between team members
- **Bulk Operations**: Create multiple configs at once

**Backend Requirements**:
- Serialize complete config packages (including dependencies)
- Handle ID remapping (UUIDs change on import)
- Validate before importing
- Rollback on errors

---

## Architecture

### JSON Package Format

Each export/import uses a standard JSON package structure:

```json
{
  "version": "1.0",
  "type": "forms",
  "exported_at": "2025-11-19T10:00:00Z",
  "count": 2,
  "data": [
    {
      "form": {
        "name": "contact_form",
        "isReferenceData": false,
        "allowInlineCreate": true
      },
      "fields": [
        {
          "name": "first_name",
          "fieldType": "text",
          "isRequired": true,
          "sortOrder": 1
        },
        {
          "name": "email",
          "fieldType": "email",
          "isRequired": true,
          "sortOrder": 2
        }
      ]
    }
  ]
}
```

**Key Points**:
- `version`: Format version (for future compatibility)
- `type`: Entity type (forms, page-configs, table-configs)
- `data`: Array of complete entity packages (including related records)
- Nested structure: Parent + children in single object

---

## Implementation Strategy

For each config type (forms, page-configs, table-configs):

### Export Flow
1. **Receive IDs** - Frontend sends list of entity IDs to export
2. **Fetch Entities** - Query business layer for entities + related records
3. **Serialize** - Convert to JSON package format
4. **Return** - Send JSON to frontend

### Import Flow
1. **Receive JSON** - Frontend sends JSON package + conflict mode
2. **Validate** - Check JSON structure, required fields, business rules
3. **Resolve Conflicts** - Check for name collisions, apply mode (merge/skip/replace)
4. **Transaction Start** - Begin database transaction
5. **Create/Update** - Insert/update entities + related records
6. **ID Remapping** - Track old ID → new ID mappings for relationships
7. **Transaction Commit** - Commit all changes or rollback on error
8. **Return** - Send created/updated IDs to frontend

---

## Step-by-Step Implementation

### Part 1: Forms Export/Import

Forms have related records (`form_fields`), so export/import must handle parent + children.

---

#### Step 1a: Define Export/Import Models (App Layer)

**File**: `app/domain/config/formapp/model.go`

```go
// ExportPackage represents a JSON export package for forms.
type ExportPackage struct {
    Version    string       `json:"version"`
    Type       string       `json:"type"`
    ExportedAt string       `json:"exportedAt"`
    Count      int          `json:"count"`
    Data       []FormPackage `json:"data"`
}

func (app ExportPackage) Encode() ([]byte, string, error) {
    data, err := json.Marshal(app)
    return data, "application/json", err
}

// FormPackage represents a single form with its fields.
type FormPackage struct {
    Form   Form        `json:"form"`
    Fields []FormField `json:"fields"`
}

// ImportPackage represents a JSON import package for forms.
type ImportPackage struct {
    Mode string        `json:"mode"` // "merge", "skip", "replace"
    Data []FormPackage `json:"data"`
}

func (app *ImportPackage) Decode(data []byte) error {
    return json.Unmarshal(data, &app)
}

func (app ImportPackage) Validate() error {
    if app.Mode != "merge" && app.Mode != "skip" && app.Mode != "replace" {
        return errs.Newf(errs.InvalidArgument, "mode must be 'merge', 'skip', or 'replace'")
    }

    if len(app.Data) == 0 {
        return errs.Newf(errs.InvalidArgument, "data cannot be empty")
    }

    for _, pkg := range app.Data {
        if err := pkg.Form.Validate(); err != nil {
            return err
        }
    }

    return nil
}

// ImportResult represents the result of an import operation.
type ImportResult struct {
    ImportedCount int      `json:"importedCount"`
    SkippedCount  int      `json:"skippedCount"`
    UpdatedCount  int      `json:"updatedCount"`
    Errors        []string `json:"errors,omitempty"`
}

func (app ImportResult) Encode() ([]byte, string, error) {
    data, err := json.Marshal(app)
    return data, "application/json", err
}
```

**Key Points**:
- `ExportPackage`: Standard export format
- `ImportPackage`: Request format with conflict mode
- `ImportResult`: Response format with statistics
- All types implement `Encoder`/`Decoder` interfaces

---

#### Step 1b: Add Export Method (Business Layer)

**File**: `business/domain/config/formbus/formbus.go`

```go
// ExportByIDs exports forms and their fields by IDs.
func (b *Business) ExportByIDs(ctx context.Context, formIDs []uuid.UUID) ([]FormWithFields, error) {
    ctx, span := otel.AddSpan(ctx, "business.formbus.exportbyids")
    defer span.End()

    var results []FormWithFields

    for _, formID := range formIDs {
        form, err := b.storer.QueryByID(ctx, formID)
        if err != nil {
            return nil, fmt.Errorf("query form %s: %w", formID, err)
        }

        fields, err := b.formFieldBus.QueryByFormID(ctx, formID)
        if err != nil {
            return nil, fmt.Errorf("query fields for form %s: %w", formID, err)
        }

        results = append(results, FormWithFields{
            Form:   form,
            Fields: fields,
        })
    }

    return results, nil
}

// FormWithFields represents a form with its fields.
type FormWithFields struct {
    Form   Form
    Fields []FormField
}
```

**Note**: You'll need access to `formFieldBus` from `formbus`. Add it as a field:

```go
type Business struct {
    log          *logger.Logger
    storer       Storer
    del          *delegate.Delegate
    formFieldBus *formfieldbus.Business  // Add this
}

func NewBusiness(log *logger.Logger, del *delegate.Delegate, storer Storer, formFieldBus *formfieldbus.Business) *Business {
    return &Business{
        log:          log,
        storer:       storer,
        del:          del,
        formFieldBus: formFieldBus,
    }
}
```

---

#### Step 1c: Add Export Method (App Layer)

**File**: `app/domain/config/formapp/formapp.go`

```go
// ExportByIDs exports forms by IDs as a JSON package.
func (a *App) ExportByIDs(ctx context.Context, formIDs []string) (ExportPackage, error) {
    // Convert string IDs to UUIDs
    uuids := make([]uuid.UUID, len(formIDs))
    for i, id := range formIDs {
        uid, err := uuid.Parse(id)
        if err != nil {
            return ExportPackage{}, errs.Newf(errs.InvalidArgument, "invalid form ID %s: %s", id, err)
        }
        uuids[i] = uid
    }

    // Export from business layer
    results, err := a.business.ExportByIDs(ctx, uuids)
    if err != nil {
        return ExportPackage{}, errs.Newf(errs.Internal, "export: %s", err)
    }

    // Convert to app models
    var packages []FormPackage
    for _, result := range results {
        packages = append(packages, FormPackage{
            Form:   ToAppForm(result.Form),
            Fields: ToAppFormFields(result.Fields),
        })
    }

    return ExportPackage{
        Version:    "1.0",
        Type:       "forms",
        ExportedAt: time.Now().Format(time.RFC3339),
        Count:      len(packages),
        Data:       packages,
    }, nil
}
```

---

#### Step 1d: Add Import Method (Business Layer)

**File**: `business/domain/config/formbus/formbus.go`

```go
// ImportForms imports forms with conflict resolution.
func (b *Business) ImportForms(ctx context.Context, packages []FormWithFields, mode string) (ImportStats, error) {
    ctx, span := otel.AddSpan(ctx, "business.formbus.importforms")
    defer span.End()

    stats := ImportStats{}

    // Start transaction
    tx, err := b.storer.NewWithTx(sqldb.NewBeginner(b.db))
    if err != nil {
        return stats, fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback()

    for _, pkg := range packages {
        // Check if form exists by name
        existing, err := b.storer.QueryByName(ctx, pkg.Form.Name)
        if err != nil && !errors.Is(err, ErrNotFound) {
            return stats, fmt.Errorf("query by name %s: %w", pkg.Form.Name, err)
        }

        existsAlready := !errors.Is(err, ErrNotFound)

        switch mode {
        case "skip":
            if existsAlready {
                stats.SkippedCount++
                continue
            }
            // Create new
            if err := b.createFormWithFields(ctx, tx, pkg); err != nil {
                return stats, err
            }
            stats.ImportedCount++

        case "replace":
            if existsAlready {
                // Delete existing and create new
                if err := tx.Delete(ctx, existing); err != nil {
                    return stats, fmt.Errorf("delete existing: %w", err)
                }
                stats.UpdatedCount++
            }
            if err := b.createFormWithFields(ctx, tx, pkg); err != nil {
                return stats, err
            }
            if !existsAlready {
                stats.ImportedCount++
            }

        case "merge":
            if existsAlready {
                // Update existing form, merge fields
                if err := b.updateFormWithFields(ctx, tx, existing.ID, pkg); err != nil {
                    return stats, fmt.Errorf("update form: %w", err)
                }
                stats.UpdatedCount++
            } else {
                // Create new
                if err := b.createFormWithFields(ctx, tx, pkg); err != nil {
                    return stats, err
                }
                stats.ImportedCount++
            }
        }
    }

    // Commit transaction
    if err := tx.Commit(); err != nil {
        return stats, fmt.Errorf("commit: %w", err)
    }

    return stats, nil
}

// ImportStats represents statistics from an import operation.
type ImportStats struct {
    ImportedCount int
    SkippedCount  int
    UpdatedCount  int
}

func (b *Business) createFormWithFields(ctx context.Context, tx Storer, pkg FormWithFields) error {
    // Create form
    newForm := NewForm{
        Name:              pkg.Form.Name,
        IsReferenceData:   pkg.Form.IsReferenceData,
        AllowInlineCreate: pkg.Form.AllowInlineCreate,
    }

    form, err := b.Create(ctx, newForm)
    if err != nil {
        return fmt.Errorf("create form: %w", err)
    }

    // Create fields
    for _, field := range pkg.Fields {
        newField := formfieldbus.NewFormField{
            FormID:      form.ID,
            Name:        field.Name,
            FieldType:   field.FieldType,
            IsRequired:  field.IsRequired,
            SortOrder:   field.SortOrder,
        }
        if _, err := b.formFieldBus.Create(ctx, newField); err != nil {
            return fmt.Errorf("create field %s: %w", field.Name, err)
        }
    }

    return nil
}

func (b *Business) updateFormWithFields(ctx context.Context, tx Storer, formID uuid.UUID, pkg FormWithFields) error {
    // Update form
    updateForm := UpdateForm{
        Name:              &pkg.Form.Name,
        IsReferenceData:   &pkg.Form.IsReferenceData,
        AllowInlineCreate: &pkg.Form.AllowInlineCreate,
    }

    if _, err := b.Update(ctx, updateForm, formID); err != nil {
        return fmt.Errorf("update form: %w", err)
    }

    // Delete existing fields and recreate (simple approach)
    existingFields, err := b.formFieldBus.QueryByFormID(ctx, formID)
    if err != nil {
        return fmt.Errorf("query existing fields: %w", err)
    }

    for _, field := range existingFields {
        if err := b.formFieldBus.Delete(ctx, field.ID); err != nil {
            return fmt.Errorf("delete field %s: %w", field.ID, err)
        }
    }

    // Create new fields
    for _, field := range pkg.Fields {
        newField := formfieldbus.NewFormField{
            FormID:      formID,
            Name:        field.Name,
            FieldType:   field.FieldType,
            IsRequired:  field.IsRequired,
            SortOrder:   field.SortOrder,
        }
        if _, err := b.formFieldBus.Create(ctx, newField); err != nil {
            return fmt.Errorf("create field %s: %w", field.Name, err)
        }
    }

    return nil
}
```

**Key Points**:
- Transactional (all-or-nothing)
- Three conflict modes: skip, replace, merge
- Creates parent (form) then children (fields)
- Returns statistics

---

#### Step 1e: Add Import Method (App Layer)

**File**: `app/domain/config/formapp/formapp.go`

```go
// ImportForms imports forms from a JSON package.
func (a *App) ImportForms(ctx context.Context, pkg ImportPackage) (ImportResult, error) {
    // Validate package
    if err := pkg.Validate(); err != nil {
        return ImportResult{}, err
    }

    // Convert app models to business models
    var busPackages []formbus.FormWithFields
    for _, formPkg := range pkg.Data {
        busPackages = append(busPackages, formbus.FormWithFields{
            Form:   toBusForm(formPkg.Form),
            Fields: toBusFormFields(formPkg.Fields),
        })
    }

    // Import via business layer
    stats, err := a.business.ImportForms(ctx, busPackages, pkg.Mode)
    if err != nil {
        return ImportResult{
            Errors: []string{err.Error()},
        }, errs.Newf(errs.Internal, "import: %s", err)
    }

    return ImportResult{
        ImportedCount: stats.ImportedCount,
        SkippedCount:  stats.SkippedCount,
        UpdatedCount:  stats.UpdatedCount,
    }, nil
}
```

---

#### Step 1f: Add API Handlers

**File**: `api/domain/http/config/formapi/formapi.go`

```go
func (api *api) exportForms(ctx context.Context, r *http.Request) web.Encoder {
    type request struct {
        IDs []string `json:"ids" validate:"required,min=1"`
    }

    var req request
    if err := web.Decode(r, &req); err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    pkg, err := api.formApp.ExportByIDs(ctx, req.IDs)
    if err != nil {
        return errs.NewError(err)
    }

    return pkg
}

func (api *api) importForms(ctx context.Context, r *http.Request) web.Encoder {
    var pkg formapp.ImportPackage
    if err := web.Decode(r, &pkg); err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    result, err := api.formApp.ImportForms(ctx, pkg)
    if err != nil {
        return errs.NewError(err)
    }

    return result
}
```

---

#### Step 1g: Register Routes

**File**: `api/domain/http/config/formapi/routes.go`

```go
// POST /v1/config/forms/export
app.HandlerFunc(http.MethodPost, version, "/config/forms/export", api.exportForms, authen,
    mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

// POST /v1/config/forms/import
app.HandlerFunc(http.MethodPost, version, "/config/forms/import", api.importForms, authen,
    mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAny))
```

---

### Part 2: Page Configs Export/Import

Page configs have TWO types of related records:
- `page_content` (page sections/tabs)
- `page_actions` (buttons/dropdowns)

Follow the same pattern as forms, but with more complex relationships.

---

#### Key Differences for Page Configs

**Export**:
```go
// Business Layer
type PageConfigWithRelations struct {
    PageConfig PageConfig
    Contents   []PageContent
    Actions    []PageAction
}

func (b *Business) ExportByIDs(ctx context.Context, ids []uuid.UUID) ([]PageConfigWithRelations, error) {
    var results []PageConfigWithRelations

    for _, id := range ids {
        config, _ := b.storer.QueryByID(ctx, id)
        contents, _ := b.pageContentBus.QueryByPageConfigID(ctx, id)
        actions, _ := b.pageActionBus.QueryByPageConfigID(ctx, id)

        results = append(results, PageConfigWithRelations{
            PageConfig: config,
            Contents:   contents,
            Actions:    actions,
        })
    }

    return results, nil
}
```

**Import**:
- Create page config first (parent)
- Create contents (children)
- Create actions (children)
- Handle nested content (parent/child content relationships)

---

### Part 3: Table Configs Export/Import

Table configs are simpler (no related records), stored as JSONB.

---

#### Key Differences for Table Configs

**Export**:
```go
// Business Layer
func (b *Business) ExportByIDs(ctx context.Context, ids []uuid.UUID) ([]TableConfig, error) {
    var results []TableConfig

    for _, id := range ids {
        config, err := b.storer.QueryByID(ctx, id)
        if err != nil {
            return nil, fmt.Errorf("query table config %s: %w", id, err)
        }
        results = append(results, config)
    }

    return results, nil
}
```

**Import**:
- Simple create/update (no children)
- Conflict resolution by name
- Validate JSONB `config` field

---

## Complete Implementation Checklist

### Forms Export/Import
- [ ] **Business Layer** (`formbus.go`, `model.go`)
  - [ ] Add `FormWithFields` struct
  - [ ] Add `ExportByIDs()` method
  - [ ] Add `ImportForms()` method
  - [ ] Add `createFormWithFields()` helper
  - [ ] Add `updateFormWithFields()` helper
  - [ ] Add `ImportStats` struct
  - [ ] Add `formFieldBus` dependency

- [ ] **Application Layer** (`formapp.go`, `model.go`)
  - [ ] Add `ExportPackage` struct with `Encode()`
  - [ ] Add `ImportPackage` struct with `Decode()` and `Validate()`
  - [ ] Add `ImportResult` struct with `Encode()`
  - [ ] Add `FormPackage` struct
  - [ ] Add `ExportByIDs()` method
  - [ ] Add `ImportForms()` method

- [ ] **API Layer** (`formapi.go`, `routes.go`)
  - [ ] Add `exportForms()` handler
  - [ ] Add `importForms()` handler
  - [ ] Register POST `/v1/config/forms/export` route
  - [ ] Register POST `/v1/config/forms/import` route

---

### Page Configs Export/Import
- [ ] **Business Layer**
  - [ ] Add `PageConfigWithRelations` struct
  - [ ] Add `ExportByIDs()` method
  - [ ] Add `ImportPageConfigs()` method
  - [ ] Add helpers for creating contents and actions
  - [ ] Add `pageContentBus` and `pageActionBus` dependencies

- [ ] **Application Layer**
  - [ ] Add export/import models
  - [ ] Add `ExportByIDs()` method
  - [ ] Add `ImportPageConfigs()` method

- [ ] **API Layer**
  - [ ] Add export/import handlers
  - [ ] Register routes

---

### Table Configs Export/Import
- [ ] **Business Layer**
  - [ ] Add `ExportByIDs()` method (simple)
  - [ ] Add `ImportTableConfigs()` method

- [ ] **Application Layer**
  - [ ] Add export/import models
  - [ ] Add `ExportByIDs()` method
  - [ ] Add `ImportTableConfigs()` method

- [ ] **API Layer**
  - [ ] Add export/import handlers
  - [ ] Register routes

---

## Testing

### Manual Testing

**Export Forms**:
```bash
# Export form by ID
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"ids": ["form-id-uuid"]}' \
  http://localhost:3000/v1/config/forms/export | jq

# Expected response:
# {
#   "version": "1.0",
#   "type": "forms",
#   "exportedAt": "2025-11-19T10:00:00Z",
#   "count": 1,
#   "data": [
#     {
#       "form": { "name": "contact_form", ... },
#       "fields": [ ... ]
#     }
#   ]
# }
```

**Import Forms**:
```bash
# Import with "skip" mode (skip existing)
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "mode": "skip",
    "data": [
      {
        "form": {
          "name": "imported_form",
          "isReferenceData": false,
          "allowInlineCreate": true
        },
        "fields": [
          {
            "name": "field1",
            "fieldType": "text",
            "isRequired": true,
            "sortOrder": 1
          }
        ]
      }
    ]
  }' \
  http://localhost:3000/v1/config/forms/import | jq

# Expected response:
# {
#   "importedCount": 1,
#   "skippedCount": 0,
#   "updatedCount": 0,
#   "errors": []
# }
```

**Round-Trip Test**:
```bash
# 1. Export a form
EXPORT=$(curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"ids": ["form-id"]}' \
  http://localhost:3000/v1/config/forms/export)

# 2. Import it back (with different name to avoid conflict)
echo $EXPORT | jq '.data[0].form.name = "imported_form"' | \
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d @- \
  http://localhost:3000/v1/config/forms/import

# 3. Verify it exists
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:3000/v1/config/forms/name/imported_form | jq
```

### Integration Tests

**Location**: `api/cmd/services/ichor/tests/config/formapi/`

**Create**: `export_test.go`, `import_test.go`

```go
func exportForms200(sd apitest.SeedData) apitest.Table {
    body := struct {
        IDs []string `json:"ids"`
    }{
        IDs: []string{sd.Forms[0].ID},
    }

    return apitest.Table{
        Name:       "export-forms-200",
        URL:        "/v1/config/forms/export",
        Method:     http.MethodPost,
        Token:      sd.Admins[0].Token,
        StatusCode: http.StatusOK,
        Input:      body,
    }
}

func importForms200(sd apitest.SeedData) apitest.Table {
    body := formapp.ImportPackage{
        Mode: "skip",
        Data: []formapp.FormPackage{
            {
                Form: formapp.Form{
                    Name:              "test_import_form",
                    IsReferenceData:   false,
                    AllowInlineCreate: true,
                },
                Fields: []formapp.FormField{
                    {
                        Name:       "test_field",
                        FieldType:  "text",
                        IsRequired: true,
                        SortOrder:  1,
                    },
                },
            },
        },
    }

    return apitest.Table{
        Name:       "import-forms-200",
        URL:        "/v1/config/forms/import",
        Method:     http.MethodPost,
        Token:      sd.Admins[0].Token,
        StatusCode: http.StatusOK,
        Input:      body,
    }
}
```

---

## Common Issues

### Issue 1: Transaction Rollback Not Working

**Error**: Partial import on error (some records created, others failed)

**Fix**: Ensure transaction is properly passed:
```go
// Business layer must use tx for all operations
tx, err := b.storer.NewWithTx(...)
defer tx.Rollback()

// Use tx (not b.storer) for all queries
tx.Create(ctx, form)
```

### Issue 2: Foreign Key Constraint Violation

**Error**: Import fails because related IDs don't exist

**Fix**: Create parent records before children:
```go
// 1. Create form first
form, _ := b.Create(ctx, newForm)

// 2. Use new form ID for fields
for _, field := range fields {
    field.FormID = form.ID  // Use newly created ID
    b.formFieldBus.Create(ctx, field)
}
```

### Issue 3: Name Collision Not Detected

**Error**: Import creates duplicate names

**Fix**: Query by name before creating:
```go
existing, err := b.storer.QueryByName(ctx, form.Name)
if err == nil {
    // Name exists, apply conflict mode
}
```

### Issue 4: Nested Content Not Imported

**Error**: Page content imports but loses parent/child relationships

**Fix**: Import in correct order (parents first, then children):
```go
// 1. Import content with nil ParentID first
for _, content := range contents {
    if content.ParentID == nil {
        created, _ := b.pageContentBus.Create(ctx, content)
        idMap[content.OldID] = created.ID  // Track mapping
    }
}

// 2. Import children, remapping ParentID
for _, content := range contents {
    if content.ParentID != nil {
        content.ParentID = idMap[*content.ParentID]  // Remap
        b.pageContentBus.Create(ctx, content)
    }
}
```

---

## Success Criteria

### Functional
- [ ] All 6 endpoints return HTTP 200
- [ ] Export includes all related records (fields, contents, actions)
- [ ] Import creates all records in a single transaction
- [ ] Conflict modes work correctly (skip, replace, merge)
- [ ] Round-trip test: export → import → verify identical

### Data Integrity
- [ ] Foreign key relationships maintained
- [ ] IDs properly remapped on import
- [ ] Transactions rollback on error (no partial imports)
- [ ] Name uniqueness validated

### Code Quality
- [ ] Follows Ardan Labs architecture
- [ ] Error handling consistent
- [ ] OpenTelemetry spans added
- [ ] All models implement Encoder/Decoder

### Testing
- [ ] Integration tests pass
- [ ] Manual testing successful
- [ ] Round-trip testing successful

---

## Estimated Time Breakdown

- **Forms Export/Import**: 2.5 hours
- **Page Configs Export/Import**: 2.5 hours
- **Table Configs Export/Import**: 1 hour
- **Testing**: 1.5 hours
- **Debugging**: 1-1.5 hours

**Total**: 6-8 hours

---

## Next Steps

After completing Phase 3:
1. Test all endpoints manually
2. Run integration tests: `make test`
3. Perform round-trip testing for each config type
4. Commit changes: `git commit -m "feat: add import/export for config entities"`
5. **Notify frontend team**: Phase 8 (Import/Export UI) is unblocked

---

## Questions?

**Q: What if import takes too long (timeout)?**
A: For large imports (100+ entities), consider:
- Batch processing (import N at a time)
- Async processing (job queue)
- Progress reporting (websocket updates)

**Q: Should I validate all fields before importing?**
A: Yes, validate early to avoid partial imports:
```go
// Validate ALL packages first
for _, pkg := range packages {
    if err := pkg.Validate(); err != nil {
        return ImportResult{Errors: []string{err.Error()}}, err
    }
}

// Then import (all validation done)
for _, pkg := range packages {
    b.importPackage(ctx, pkg)
}
```

**Q: How do I handle circular dependencies?**
A: For page content (parent/child relationships):
1. Import all parents first (ParentID = nil)
2. Track old ID → new ID mappings
3. Import children, remapping ParentID using map

**Q: Should imports be admin-only?**
A: Depends on use case:
- Admin-only: Use `auth.RuleAdminOnly`
- User-specific configs: Use `auth.RuleAny` and check ownership

---

**Ready to implement?** Start with forms (simplest), then page-configs (most complex), then table-configs (simplest). Good luck!
