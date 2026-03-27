# Progress Summary: form-data.md

## Overview
Architecture for multi-entity form submissions. Handles atomic batch operations on multiple entities in correct FK-safe execution order within a single database transaction.

## FormdataApp [app] — `app/domain/formdata/formdataapp/formdataapp.go`

**Responsibility:** Execute batch multi-entity operations atomically.

### Struct
```go
type App struct {
    log          *logger.Logger
    registry     *formdataregistry.Registry     // entity registration lookup
    db           *sqlx.DB
    formBus      *formbus.Business              // form metadata
    formFieldBus *formfieldbus.Business         // form field definitions
    templateProc *workflow.TemplateProcessor    // {{entity.id}} resolution
}
```

### Main Entry Point
```go
UpsertFormData(ctx context.Context, formID uuid.UUID, req FormDataRequest) (FormDataResponse, error)
```

### Key Facts
- **Single HTTP POST → N entity creates/updates** in FK-safe execution order
- **Sorts operations by OperationMeta.Order** (FK-safe execution)
- **Single DB transaction** — `sql.LevelReadCommitted` isolation level
- **{{entity.id}} template resolution** — happens after each operation (for FK references)
- **FK value resolution** — human-readable names resolved to UUIDs via Registry.Get(name).QueryByNameFunc
- **DoS protection** — maxArrayItems = 1000 per operation (enforced before tx begins)

## Request / Response Models

```go
type FormDataRequest struct {
    Operations map[string]OperationMeta    // "orders": {Operation: "create", Order: 1}
    Data       map[string]json.RawMessage  // "orders": {...} | "line_items": [{...},...]
}

type OperationMeta struct {
    Operation formdataregistry.EntityOperation  // "create" | "update" | "delete"
    Order     int                               // execution sequence (1 = first)
}

type FormDataResponse struct {
    Success bool
    Results map[string]interface{}  // created/updated IDs keyed by operation name
}
```

## Registry [sdk] — `app/sdk/formdataregistry/`

**Responsibility:** Centralized registry of entities that participate in multi-entity operations.

Files: registry.go, types.go, reflection.go

### EntityRegistration
```go
type EntityRegistration struct {
    Name              string
    DecodeNew         func(json.RawMessage) (interface{}, error)       // decode request
    CreateFunc        func(context.Context, interface{}) (interface{}, error)
    CreateModel       interface{}
    DecodeUpdate      func(json.RawMessage) (interface{}, error)
    UpdateFunc        func(context.Context, uuid.UUID, interface{}) (interface{}, error)
    UpdateModel       interface{}
    QueryByNameFunc   func(ctx context.Context, name string) (uuid.UUID, error) // for FK resolution
}

type Registry struct {
    mu         sync.RWMutex
    entities   map[string]*EntityRegistration
    entityByID map[uuid.UUID]*EntityRegistration
}
```

### Methods
- `New() *Registry` — create new registry
- `Register(reg EntityRegistration) error` — register entity by name
- `RegisterWithID(entityID uuid.UUID, reg EntityRegistration) error` — register entity by UUID
- `Get(name string) (*EntityRegistration, error)` — lookup by name
- `GetByID(id uuid.UUID) (*EntityRegistration, error)` — lookup by UUID
- `ListEntities() []string` — list all registered names

### Key Facts
- **Thread-safe:** RWMutex guards all reads and writes
- **Read-only after startup:** registration happens once in all.go
- **Registered at startup** — in api/cmd/services/ichor/build/all/all.go

## FormBus / FormFieldBus [bus]

Files:
- `business/domain/config/formbus/formbus.go`
- `business/domain/config/formfieldbus/formfieldbus.go`

### Key Facts
- **FormField.Config** — json.RawMessage holds:
  - `execution_order` (int) — which operation group this field belongs to
  - `dropdown_config` (object) — `{entity, display_field, inline_create}`
- **LineItemsFieldConfig.ExecutionOrder is UNRELATED to workflow** — different domain (form configuration, not workflow execution)
- **Data sources:**
  - ⊗ config.forms
  - ⊗ config.form_fields

## TemplateProcessor [sdk]

File: `business/sdk/workflow/` (shared with workflow engine)

Type: `workflow.TemplateProcessor`

### Key Facts
- `Process(template string, context ActionExecutionContext) string` — resolve templates
- **Resolves** — `{{entity_id}}`, `{{field_name}}`, dot notation: `{{order.customer.email}}`
- **Filters** — `{{created_date | dateFormat}}`
- **Applied per-item** — when Data value is array

## Change Patterns

### ⚠ Adding a New Entity to Registry
Affects 3-4 areas:
1. `app/sdk/formdataregistry/registry.go` — EntityRegistration struct (no change needed)
2. `api/cmd/services/ichor/build/all/all.go` — registry.Register() or RegisterWithID() call at startup
3. `business/domain/{area}/{entity}bus/{entity}bus.go` — QueryByNameFunc must exist or be added
4. `business/sdk/dbtest/seedmodels/forms.go` — add seed helper if entity needs test form data

### ⚠ Adding a New OperationMeta Field
Affects 2 areas:
1. `app/domain/formdata/formdataapp/model.go` — OperationMeta struct definition
2. `app/domain/formdata/formdataapp/formdataapp.go` — read + apply new field in UpsertFormData
3. **Update design doc:** FORMDATA_IMPLEMENTATION.md

### ⚠ Changing Transaction Isolation Level
Affects 1 file:
1. `app/domain/formdata/formdataapp/formdataapp.go` — sql.TxOptions (currently sql.LevelReadCommitted)
2. **Warning:** Upgrading to Serializable may cause serialization failures under concurrent load — test thoroughly

### ⚠ Changing FormDataRequest Shape (Data Field Type)
Affects multiple areas:
1. `app/domain/formdata/formdataapp/model.go` — Data map[string]json.RawMessage definition
2. **All HTTP callers** — frontend API client + integration tests
3. `api/cmd/services/ichor/tests/.../formdataapi/` — update test request builders

## Critical Points
- **Execution order matters** — operations sorted by OperationMeta.Order before execution
- **FK-safe execution** — dependent entities must have lower Order values
- **Template resolution per-item** — enables derived fields like {{order.id}} in line items
- **Single transaction boundary** — atomic: all succeed or all fail
- **DoS protection** — maxArrayItems = 1000 limits payload size

## Notes for Future Development
FormData is a sophisticated multi-entity orchestrator with careful ordering and template resolution. Most changes will be:
- Adding new entities to registry (straightforward)
- Adding new OperationMeta fields (moderate)
- Changing isolation level (risky, requires testing)

Never change FormDataRequest.Data type lightly (breaks API contracts and frontend).
