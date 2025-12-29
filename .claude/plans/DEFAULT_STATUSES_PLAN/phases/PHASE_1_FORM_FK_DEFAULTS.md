# Phase 1: Form Configuration FK Default Resolution

**Status**: Pending
**Category**: backend
**Dependencies**: None

---

## CRITICAL CONTEXT: DATA FLOW AND SECURITY MODEL

**READ AND UNDERSTAND THIS SECTION BEFORE IMPLEMENTING ANY TASKS.**

### Data Sources - Where Each Value Comes From

```
COMPILE-TIME DATA (Trusted, hardcoded in Go source files):
├── formdata_registry.go
│   └── Registry entries: map of entity names like "sales.orders", "core.users"
│       These are the ONLY entities that can be queried for FK resolution
│
└── tableforms.go / forms.go (seed data)
    └── Form field configs containing:
        ├── entity: "sales.line_item_fulfillment_statuses"
        ├── label_column: "name"
        ├── value_column: "id"
        └── default_value_create: "Pending"

RUNTIME DATA:
└── User HTTP request body
    └── Field values submitted by user (ALWAYS parameterized in queries)
```

### The Validation Chain

When `resolveFKByName(ctx, dropdownConfig, value)` is called:

1. **INPUT**: `dropdownConfig.Entity` = `"sales.line_item_fulfillment_statuses"`
2. **VALIDATION**: Call `registry.Get("sales.line_item_fulfillment_statuses")`
3. **IF NOT IN REGISTRY**: Return error immediately. NO query is executed.
4. **IF IN REGISTRY**: Entity is trusted. Proceed to query.
5. **QUERY**: Uses `dropdownConfig.LabelColumn` and `dropdownConfig.ValueColumn` from config
6. **VALUE**: The lookup value (e.g., `"Pending"`) is ALWAYS parameterized as `:value`

### Why Column Names Are Safe

Column names (`label_column`, `value_column`) come from form field config JSON stored in `config.form_fields` table. This data is:
- Seeded from `tableforms.go` and `forms.go` at application startup
- Hardcoded as Go string literals in those files
- NOT user-modifiable via normal API operations

Even if form field configs WERE modifiable via API:
- The `entity` must still pass registry validation
- Attackers cannot add new entities to the registry (compile-time only)
- They can only reference entities already in the registry
- Those entities exist and are legitimate

### Security Summary

| Query Component | Source | Trusted? | Protection |
|-----------------|--------|----------|------------|
| Table name | `dropdownConfig.Entity` | YES after validation | Registry whitelist check |
| Column names | `dropdownConfig.LabelColumn`, `dropdownConfig.ValueColumn` | YES | From compile-time seed data |
| WHERE value | `value` parameter | NO (may be user input) | Always parameterized as `:value` |

**DO NOT add additional column validation. The registry check on Entity is the security boundary.**

---

## Overview

Enable form fields to specify default values by human-readable names (e.g., `"Pending"`) for FK fields. The `formdataapp` resolves names to UUIDs during form processing.

### Existing Types Used

**DropdownConfig** (defined in `business/domain/config/formfieldbus/model.go:68-74`):
```go
type DropdownConfig struct {
    Entity         string   `json:"entity"`
    LabelColumn    string   `json:"label_column"`
    ValueColumn    string   `json:"value_column"`
    DisplayColumns []string `json:"additional_display_columns,omitempty"`
}
```

**FieldDefaultConfig** (defined in `business/domain/config/formfieldbus/model.go:58-66`):
```go
type FieldDefaultConfig struct {
    DefaultValue       string `json:"default_value,omitempty"`
    DefaultValueCreate string `json:"default_value_create,omitempty"`
    DefaultValueUpdate string `json:"default_value_update,omitempty"`
    Hidden             bool   `json:"hidden,omitempty"`
}
```

---

## Tasks

### Task 1: Add Logger to formdataapp

**File**: `app/domain/formdata/formdataapp/formdataapp.go`

**Action**: Add `log *logger.Logger` as the FIRST field in the `App` struct and FIRST parameter in `NewApp`.

**Before** (current struct):
```go
type App struct {
	registry     *formdataregistry.Registry
	db           *sqlx.DB
	formBus      *formbus.Business
	formFieldBus *formfieldbus.Business
	templateProc *workflow.TemplateProcessor
}
```

**After** (modified struct):
```go
type App struct {
	log          *logger.Logger  // ADD: first field
	registry     *formdataregistry.Registry
	db           *sqlx.DB
	formBus      *formbus.Business
	formFieldBus *formfieldbus.Business
	templateProc *workflow.TemplateProcessor
}
```

**Before** (current constructor):
```go
func NewApp(
	registry *formdataregistry.Registry,
	db *sqlx.DB,
	formBus *formbus.Business,
	formFieldBus *formfieldbus.Business,
) *App
```

**After** (modified constructor):
```go
func NewApp(
	log *logger.Logger,  // ADD: first parameter
	registry *formdataregistry.Registry,
	db *sqlx.DB,
	formBus *formbus.Business,
	formFieldBus *formfieldbus.Business,
) *App
```

**Required import**: `"github.com/timmaaaz/ichor/foundation/logger"`

---

### Task 2: Add resolveFKByName Method

**File**: `app/domain/formdata/formdataapp/formdataapp.go`

**Action**: Add this method to the App type.

```go
// resolveFKByName resolves a human-readable FK value to a UUID.
//
// Security model (see CRITICAL CONTEXT section in phase plan):
// - dropdownConfig.Entity is validated against registry whitelist
// - dropdownConfig.LabelColumn and ValueColumn come from compile-time seed data
// - value is always parameterized, never interpolated
//
// Returns the UUID unchanged if value is already a valid UUID (fast path).
func (a *App) resolveFKByName(ctx context.Context, dropdownConfig *formfieldbus.DropdownConfig, value string) (uuid.UUID, error) {
	// Fast path: value is already a UUID
	if id, err := uuid.Parse(value); err == nil {
		return id, nil
	}

	// Security: validate entity against registry whitelist
	if _, err := a.registry.Get(dropdownConfig.Entity); err != nil {
		return uuid.Nil, fmt.Errorf("entity %q not registered for FK resolution", dropdownConfig.Entity)
	}

	// Build query using config values (all from compile-time seed data)
	data := struct {
		Value string `db:"value"`
	}{
		Value: value,
	}

	q := fmt.Sprintf(`SELECT %s FROM %s WHERE %s = :value`,
		dropdownConfig.ValueColumn,
		dropdownConfig.Entity,
		dropdownConfig.LabelColumn,
	)

	var id uuid.UUID
	if err := sqldb.NamedQueryStruct(ctx, a.log, a.db, q, data, &id); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return uuid.Nil, fmt.Errorf("value %q not found in %s.%s", value, dropdownConfig.Entity, dropdownConfig.LabelColumn)
		}
		return uuid.Nil, fmt.Errorf("FK resolution query failed: %w", err)
	}

	return id, nil
}
```

**Required imports**:
- `"errors"`
- `"fmt"`
- `"github.com/timmaaaz/ichor/business/sdk/sqldb"`

---

### Task 3: Update mergeFieldDefaults

**File**: `app/domain/formdata/formdataapp/formdataapp.go`

**Action**:
1. Add `ctx context.Context` as FIRST parameter
2. After determining `defaultVal`, check if field has `entity` in config
3. If so, call `resolveFKByName` to convert name to UUID

**Signature change**:
```go
// Before:
func (a *App) mergeFieldDefaults(data json.RawMessage, fieldConfigs []formfieldbus.FormField, operation formdataregistry.EntityOperation) (json.RawMessage, InjectionResult, error)

// After:
func (a *App) mergeFieldDefaults(ctx context.Context, data json.RawMessage, fieldConfigs []formfieldbus.FormField, operation formdataregistry.EntityOperation) (json.RawMessage, InjectionResult, error)
```

**Logic to add** (after determining `defaultVal` and before setting `dataMap[field.Name]`):
```go
// Check if field has dropdown config for FK resolution
var dropdownCfg struct {
	Entity      string `json:"entity"`
	LabelColumn string `json:"label_column"`
	ValueColumn string `json:"value_column"`
}
if err := json.Unmarshal(field.Config, &dropdownCfg); err == nil && dropdownCfg.Entity != "" {
	ddConfig := &formfieldbus.DropdownConfig{
		Entity:      dropdownCfg.Entity,
		LabelColumn: dropdownCfg.LabelColumn,
		ValueColumn: dropdownCfg.ValueColumn,
	}
	resolvedID, err := a.resolveFKByName(ctx, ddConfig, defaultVal)
	if err != nil {
		a.log.Warn(ctx, "FK default resolution failed",
			"field", field.Name,
			"entity", dropdownCfg.Entity,
			"default", defaultVal,
			"error", err)
		continue
	}
	defaultVal = resolvedID.String()
}
```

**Update all call sites** to pass `ctx` as first argument.

---

### Task 4: Update mergeLineItemFieldDefaults

**File**: `app/domain/formdata/formdataapp/formdataapp.go`

**Action**:
1. Add `ctx context.Context` as FIRST parameter
2. If field has `DropdownConfig` with `Entity`, call `resolveFKByName`

**Signature change**:
```go
// Before:
func (a *App) mergeLineItemFieldDefaults(data json.RawMessage, lineItemFields []formfieldbus.LineItemField, operation formdataregistry.EntityOperation) (json.RawMessage, error)

// After:
func (a *App) mergeLineItemFieldDefaults(ctx context.Context, data json.RawMessage, lineItemFields []formfieldbus.LineItemField, operation formdataregistry.EntityOperation) (json.RawMessage, error)
```

**Logic to add** (after determining `defaultVal` and before setting `dataMap[field.Name]`):
```go
if field.DropdownConfig != nil && field.DropdownConfig.Entity != "" {
	resolvedID, err := a.resolveFKByName(ctx, field.DropdownConfig, defaultVal)
	if err != nil {
		a.log.Warn(ctx, "FK default resolution failed for line item field",
			"field", field.Name,
			"entity", field.DropdownConfig.Entity,
			"default", defaultVal,
			"error", err)
		continue
	}
	defaultVal = resolvedID.String()
}
```

**Update all call sites** to pass `ctx` as first argument.

---

### Task 5: Update all.go Call Site

**File**: `api/cmd/services/ichor/build/all/all.go`

**Action**: Pass `cfg.Log` as first argument to `formdataapp.NewApp`.

**Before**:
```go
formdataApp := formdataapp.NewApp(formDataRegistry, cfg.DB, formBus, formFieldBus)
```

**After**:
```go
formdataApp := formdataapp.NewApp(cfg.Log, formDataRegistry, cfg.DB, formBus, formFieldBus)
```

---

### Task 6: Add Max Array Size Validation

**File**: `app/domain/formdata/formdataapp/formdataapp.go`

**Action**: Add constant and validation at start of `executeArrayOperation`.

**Add constant** (at package level):
```go
const maxArrayItems = 1000
```

**Add validation** (at start of `executeArrayOperation`, after checking `len(items) == 0`):
```go
if len(items) > maxArrayItems {
	return nil, errs.Newf(errs.InvalidArgument, "array for %s exceeds maximum size of %d items", step.EntityName, maxArrayItems)
}
```

---

### Task 7: Update Form Seed Data

**Files**:
- `business/sdk/dbtest/seedmodels/tableforms.go`
- `business/sdk/dbtest/seedmodels/forms.go`

**Action**: Add `"default_value_create": "Pending"` to status FK field configs.

**Fields to update in tableforms.go**:

| Function | Field Name | Entity |
|----------|-----------|--------|
| `GetSalesOrderFormFields` | `fulfillment_status_id` | `sales.order_fulfillment_statuses` |
| `GetSalesOrderLineItemFormFields` | `line_item_fulfillment_status_id` | `sales.line_item_fulfillment_statuses` |
| `GetPurchaseOrderFormFields` | `purchase_order_status_id` | `procurement.purchase_order_statuses` |
| `GetPurchaseOrderLineItemFormFields` | `line_item_status_id` | `procurement.purchase_order_line_item_statuses` |
| `GetUserAssetFormFields` | `approval_status_id` | `assets.approval_status` |
| `GetUserAssetFormFields` | `fulfillment_status_id` | `assets.fulfillment_status` |

**Example config update**:
```go
Config: json.RawMessage(`{
	"entity": "sales.order_fulfillment_statuses",
	"label_column": "name",
	"value_column": "id",
	"default_value_create": "Pending"
}`),
```

**Fields to update in forms.go**:

For `LineItemField` structs, add `DefaultValueCreate` field:
```go
{
	Name:               "line_item_fulfillment_statuses_id",
	Label:              "Fulfillment Status",
	Type:               "dropdown",
	Required:           true,
	DefaultValueCreate: "Pending",  // ADD THIS
	DropdownConfig: &formfieldbus.DropdownConfig{
		Entity:      "sales.line_item_fulfillment_statuses",
		LabelColumn: "name",
		ValueColumn: "id",
	},
},
```

---

### Task 8: Add Unit Tests

**File**: `app/domain/formdata/formdataapp/fkresolution_test.go` (create new file)

**Test cases**:
1. `TestResolveFKByName_ValidName` - Resolves "Pending" to correct UUID
2. `TestResolveFKByName_UUIDPassthrough` - Returns UUID unchanged if already valid
3. `TestResolveFKByName_UnknownEntity` - Returns error for unregistered entity
4. `TestResolveFKByName_ValueNotFound` - Returns error when name not in table

---

### Task 9: Add Integration Tests

**Files**:
- `api/cmd/services/ichor/tests/formdata/formdataapi/upsert_test.go`
- `api/cmd/services/ichor/tests/formdata/formdataapi/seed_test.go`

**Test cases**:
1. Create order without status -> gets "Pending" UUID
2. Create order with explicit status -> user value preserved
3. Create line items without status -> each gets "Pending" UUID

---

## Validation Criteria

- [ ] `go build ./...` passes
- [ ] `make test` passes
- [ ] Orders without status get "Pending" UUID
- [ ] Line items without status get "Pending" UUID
- [ ] Invalid names log warning but don't fail request
- [ ] User-provided values preserved (not overwritten)
- [ ] UUID values pass through without database lookup
- [ ] Unregistered entities rejected before query
- [ ] Array size limited to 1000 items

---

## Deliverables

1. `formdataapp.App.log` field added
2. `formdataapp.resolveFKByName` method
3. Updated `mergeFieldDefaults` with ctx and FK resolution
4. Updated `mergeLineItemFieldDefaults` with ctx and FK resolution
5. Updated `all.go` call site
6. `maxArrayItems` constant and validation
7. Form seed data with `default_value_create: "Pending"`
8. Unit tests for FK resolution
9. Integration tests for default status injection

---

**Last Updated**: 2025-12-29
