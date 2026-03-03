# form-data

[app]=application layer [bus]=business layer [sdk]=shared
→=depends on ⊕=writes ⊗=reads [tx]=transaction [cache]=sturdyc

---

## Overview

Multi-entity transactional form submission. Single HTTP POST → N entity creates/updates
in FK-safe execution order. Template variables link created IDs across operations.
max array items per operation: 1000 (DoS protection).

---

## FormdataApp [app]

file: app/domain/formdata/formdataapp/formdataapp.go
```go
type App struct {
    log          *logger.Logger
    registry     *formdataregistry.Registry
    db           *sqlx.DB
    formBus      *formbus.Business
    formFieldBus *formfieldbus.Business
    templateProc *workflow.TemplateProcessor
}
```

main entry:
  UpsertFormData(ctx context.Context, formID uuid.UUID, req FormDataRequest) (FormDataResponse, error)

key facts:
  - Sorts operations by OperationMeta.Order (FK-safe execution)
  - Runs inside single DB [tx] at isolation level sql.LevelReadCommitted
  - {{entity.id}} templates resolved via workflow.TemplateProcessor after each operation
  - FK values: human-readable names resolved to UUIDs via Registry.Get(name).QueryByNameFunc
  - maxArrayItems = 1000 per operation (enforced before tx begins)

---

## Request / Response

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

---

## Registry [sdk]

file: app/sdk/formdataregistry/
files: registry.go, types.go, reflection.go

```go
type EntityRegistration struct {
    Name          string
    DecodeNew     func(json.RawMessage) (interface{}, error)
    CreateFunc    func(context.Context, interface{}) (interface{}, error)
    CreateModel   interface{}
    DecodeUpdate  func(json.RawMessage) (interface{}, error)
    UpdateFunc    func(context.Context, uuid.UUID, interface{}) (interface{}, error)
    UpdateModel   interface{}
    QueryByNameFunc func(ctx context.Context, name string) (uuid.UUID, error)
}

type Registry struct {
    mu         sync.RWMutex
    entities   map[string]*EntityRegistration
    entityByID map[uuid.UUID]*EntityRegistration
}
```

api:
  New() *Registry
  Register(reg EntityRegistration) error
  RegisterWithID(entityID uuid.UUID, reg EntityRegistration) error
  Get(name string) (*EntityRegistration, error)
  GetByID(id uuid.UUID) (*EntityRegistration, error)
  ListEntities() []string

Thread-safe: RWMutex guards all reads and writes. Read-only after startup registration.
Registered at startup in api/cmd/services/ichor/build/all/all.go.

---

## FormBus / FormFieldBus [bus]

files:
  business/domain/config/formbus/formbus.go
  business/domain/config/formfieldbus/formfieldbus.go
key facts:
  - FormField.Config json.RawMessage holds:
      execution_order (int) — which operation group this field belongs to
      dropdown_config (object) — {entity, display_field, inline_create}
  - LineItemsFieldConfig.ExecutionOrder is UNRELATED to workflow (different domain)
⊗ config.forms
⊗ config.form_fields

---

## TemplateProcessor [sdk]

file: business/sdk/workflow/ (shared with workflow engine)
type: workflow.TemplateProcessor
key facts:
  - Process(template string, context ActionExecutionContext) string
  - Resolves {{entity_id}}, {{field_name}}, dot notation: {{order.customer.email}}
  - Filters: {{created_date | dateFormat}}
  - Applied per-item when Data value is array

---

## ⚠ Adding a new entity to Registry

  app/sdk/formdataregistry/registry.go                              (EntityRegistration struct — no change needed)
  api/cmd/services/ichor/build/all/all.go                           (registry.Register() or RegisterWithID() call at startup)
  business/domain/{area}/{entity}bus/{entity}bus.go                 (QueryByNameFunc must exist or be added)
  business/sdk/dbtest/seedmodels/forms.go                           (add seed helper if entity needs test form data)

## ⚠ Adding a new OperationMeta field

  app/domain/formdata/formdataapp/model.go                          (OperationMeta struct)
  app/domain/formdata/formdataapp/formdataapp.go                    (read + apply new field in UpsertFormData)
  FORMDATA_IMPLEMENTATION.md                                         (update design doc)

## ⚠ Changing transaction isolation level

  app/domain/formdata/formdataapp/formdataapp.go                    (sql.TxOptions — currently sql.LevelReadCommitted)
  Note: upgrading to Serializable may cause serialization failures under concurrent load — test thoroughly

## ⚠ Changing FormDataRequest shape (Data field type)

  app/domain/formdata/formdataapp/model.go                          (Data map[string]json.RawMessage)
  All HTTP callers sending form data payloads                        (frontend API client + integration tests)
  api/cmd/services/ichor/tests/.../formdataapi/                     (update test request builders)
