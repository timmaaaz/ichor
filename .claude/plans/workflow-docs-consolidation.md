# Workflow Documentation Consolidation & Audit

## Overview
Consolidate all workflow engine documentation into a single comprehensive source of truth, then audit the documentation against the actual implementation for accuracy.

## Current Documentation Locations (5 files, ~2,500 lines total)

| File | Lines | Content |
|------|-------|---------|
| `business/sdk/workflow/docs/workflow-config-validator-spec.md` | ~1000 | Field reference, validation rules, action schemas, template variables |
| `business/sdk/workflow/workflowactions/communication/ALERT_SYSTEM.md` | ~330 | Alert system architecture, API, frontend guide |
| `business/sdk/workflow/workflowactions/data/updatefield_overview.md` | ~95 | UpdateFieldHandler overview |
| `business/sdk/workflow/workflowactions/inventory/allocate_overview.md` | ~380 | Inventory allocation system guide |
| `.claude/plans/completed-plans/WORKFLOW_EVENT_FIRING_INFRASTRUCTURE.md` | ~1650 | Event infrastructure, delegate pattern, domain checklist |

## Target Location
`docs/workflow/` - New consolidated documentation directory

## Archive Location
`.archive/workflow-docs/` - Original documentation files will be moved here after consolidation

---

## Phase 1: Consolidation
**Status**: ✅ COMPLETE

### Created Files (18 total)

```
docs/workflow/
├── README.md                    ✅ Created
├── architecture.md              ✅ Created
├── configuration/
│   ├── triggers.md              ✅ Created
│   ├── rules.md                 ✅ Created
│   └── templates.md             ✅ Created
├── actions/
│   ├── overview.md              ✅ Created
│   ├── create-alert.md          ✅ Created
│   ├── update-field.md          ✅ Created
│   ├── send-email.md            ✅ Created
│   ├── send-notification.md     ✅ Created
│   ├── seek-approval.md         ✅ Created
│   └── allocate-inventory.md    ✅ Created
├── database-schema.md           ✅ Created
├── api-reference.md             ✅ Created
├── event-infrastructure.md      ✅ Created
├── testing.md                   ✅ Created
└── adding-domains.md            ✅ Created
```

### Tasks
- [x] Create `docs/workflow/` directory structure
- [x] Create `docs/workflow/README.md` (main entry point with TOC)
- [x] Create `docs/workflow/architecture.md` (consolidate event flow from all docs)
- [x] Create `docs/workflow/configuration/triggers.md` (from validator-spec §1-6)
- [x] Create `docs/workflow/configuration/rules.md` (from validator-spec §4,7-9)
- [x] Create `docs/workflow/configuration/templates.md` (from validator-spec §11)
- [x] Create `docs/workflow/actions/overview.md` (action registry, interfaces)
- [x] Create `docs/workflow/actions/create-alert.md` (from ALERT_SYSTEM.md + validator-spec §10.1)
- [x] Create `docs/workflow/actions/update-field.md` (from updatefield_overview.md + validator-spec §10.2)
- [x] Create `docs/workflow/actions/send-email.md` (from validator-spec §10.3)
- [x] Create `docs/workflow/actions/send-notification.md` (from validator-spec §10.4)
- [x] Create `docs/workflow/actions/seek-approval.md` (from validator-spec §10.5)
- [x] Create `docs/workflow/actions/allocate-inventory.md` (from allocate_overview.md + validator-spec §10.6)
- [x] Create `docs/workflow/database-schema.md` (from validator-spec §13-14, ALERT_SYSTEM.md)
- [x] Create `docs/workflow/api-reference.md` (from ALERT_SYSTEM.md API section)
- [x] Create `docs/workflow/event-infrastructure.md` (from WORKFLOW_EVENT_FIRING_INFRASTRUCTURE.md)
- [x] Create `docs/workflow/testing.md` (from event infrastructure testing section)
- [x] Create `docs/workflow/adding-domains.md` (from delegate pattern section)

---

## Phase 2: Audit - Core Engine
**Status**: ✅ COMPLETE (All 15 discrepancies fixed)

### Files to Audit Against
| File | Purpose |
|------|---------|
| `business/sdk/workflow/models.go` | Core data structures |
| `business/sdk/workflow/engine.go` | Workflow engine |
| `business/sdk/workflow/trigger.go` | Trigger processing |
| `business/sdk/workflow/template.go` | Template variable processing |
| `business/sdk/workflow/queue.go` | RabbitMQ queue management |
| `business/sdk/workflow/interfaces.go` | Action interfaces |
| `business/sdk/workflow/eventpublisher.go` | Event publishing |
| `business/sdk/workflow/delegatehandler.go` | Delegate pattern bridge |

### Audit Checklist
- [x] **models.go**: Verify all struct fields match documentation
- [x] **engine.go**: Verify execution flow matches architecture docs
- [x] **trigger.go**: Verify operators and condition evaluation logic
- [x] **template.go**: Verify all filters exist and work as documented
- [x] **queue.go**: Verify queue management matches documentation
- [x] **interfaces.go**: Verify ActionHandler interface is documented
- [x] **eventpublisher.go**: Verify PublishCreateEvent/Update/Delete methods
- [x] **delegatehandler.go**: Verify domain registration and event mapping

### Discrepancies Found

#### 1. interfaces.go - ActionHandler Interface Mismatch
**Location**: `docs/workflow/architecture.md:218-224`
**Issue**: Documentation shows incorrect interface signature

**Documentation says:**
```go
type ActionHandler interface {
    Type() string
    Validate(config json.RawMessage) error
    Execute(ctx context.Context, config json.RawMessage, execCtx ExecutionContext) (*ActionResult, error)
}
```

**Actual code (interfaces.go:38-42):**
```go
type ActionHandler interface {
    Execute(ctx context.Context, config json.RawMessage, context ActionExecutionContext) (any, error)
    Validate(config json.RawMessage) error
    GetType() string
}
```

**Differences:**
- Method order is different (Execute first in code, Type first in docs)
- Method name is `GetType()` not `Type()`
- Return type is `(any, error)` not `(*ActionResult, error)`
- Parameter name is `ActionExecutionContext` not `ExecutionContext`

**Action Required**: Update `docs/workflow/architecture.md:218-224` to match actual interface

---

#### 2. interfaces.go - Missing AsyncActionHandler Documentation
**Location**: `docs/workflow/architecture.md`
**Issue**: The `AsyncActionHandler` interface (interfaces.go:88-95) is not documented

**Actual code:**
```go
type AsyncActionHandler interface {
    ActionHandler
    ProcessQueued(ctx context.Context, payload json.RawMessage, publisher *EventPublisher) error
}
```

**Action Required**: Add AsyncActionHandler documentation to architecture.md or create separate interface docs

---

#### 3. queue.go - QueueManager Struct Mismatch
**Location**: `docs/workflow/architecture.md:120-127`
**Issue**: Documentation shows simplified struct, missing several fields

**Documentation says:**
```go
type QueueManager struct {
    log           *logger.Logger
    db            *sqlx.DB
    engine        *Engine
    rabbitClient  *rabbitmq.Client
    queue         *rabbitmq.WorkflowQueue
}
```

**Actual code (queue.go:19-45) has additional fields:**
- `config QueueConfig`
- `mu sync.RWMutex`
- `isRunning atomic.Bool`
- `consumers map[string]*rabbitmq.Consumer`
- `stopChan chan struct{}`
- `processingWG sync.WaitGroup`
- `metrics QueueMetrics`
- `metricsLock sync.RWMutex`
- `circuitBreakerManager *CircuitBreakerManager`
- `handlerRegistry *websocket.HandlerRegistry`

**Action Required**: The documentation intentionally shows simplified view; add note that it's simplified

---

#### 4. queue.go - Missing Circuit Breaker Documentation
**Location**: `docs/workflow/architecture.md`
**Issue**: Circuit breaker feature is not documented at all

**Actual features (queue.go:86-130):**
- Per-queue-type circuit breakers
- Global circuit breaker fallback
- Configurable thresholds and timeouts
- Half-open recovery state

**Action Required**: Add circuit breaker section to architecture.md or create separate reliability docs

---

#### 5. queue.go - Queue Types Not Documented
**Location**: `docs/workflow/architecture.md:409-413`
**Issue**: Documentation only mentions `workflow_events` queue

**Actual queue types (queue.go:289-296):**
- `QueueTypeWorkflow`
- `QueueTypeApproval`
- `QueueTypeNotification`
- `QueueTypeInventory`
- `QueueTypeEmail`
- `QueueTypeAlert`

**Action Required**: Update queue settings documentation to list all queue types

---

#### 6. trigger.go - Missing `scheduled` Trigger Type in Validation
**Location**: `docs/workflow/configuration/triggers.md:14`
**Issue**: Documentation lists `scheduled` trigger type but code confirms it exists

**Actual code (trigger.go:435):**
```go
supportedTypes := []string{"on_create", "on_update", "on_delete", "scheduled"}
```

**Status**: Documentation is CORRECT - no action needed

---

#### 7. template.go - Missing Filters Documentation
**Location**: `docs/workflow/configuration/templates.md:102-109`
**Issue**: Currency filter documentation incomplete

**Documentation lists these currencies:**
> USD, EUR, GBP, JPY, CAD, AUD, CHF, CNY, INR, and more

**Actual code (template.go:687-709) also includes:**
- MXN (Mexican Peso with $ symbol)

**Action Required**: Update currency list or add note about supported currencies

---

#### 8. template.go - formatDate Filter Output Format Mismatch
**Location**: `docs/workflow/configuration/templates.md:116-118`
**Issue**: Documentation shows AM/PM format, but code uses 24-hour format

**Documentation says:**
| `formatDate:time` | Time only | `2:30 PM` |
| `formatDate:datetime` | Date and time | `Jan 15, 2024 2:30 PM` |

**Actual code (template.go:771-772):**
```go
case "time":
    format = "15:04:05"
case "datetime":
    format = "2006-01-02 15:04:05"
```

These output `14:30:05` and `2024-01-15 14:30:05` respectively (24-hour format, no AM/PM)

**Action Required**: Update documentation table to show actual 24-hour output

---

#### 9. eventpublisher.go - Missing Methods Documentation
**Location**: `docs/workflow/architecture.md:77-81`
**Issue**: Documentation missing several methods

**Documentation shows:**
```go
func (ep *EventPublisher) PublishCreateEvent(ctx, entityName, result, userID)
func (ep *EventPublisher) PublishUpdateEvent(ctx, entityName, result, fieldChanges, userID)
func (ep *EventPublisher) PublishDeleteEvent(ctx, entityName, entityID, userID)
```

**Actual code also has (eventpublisher.go:35-76):**
- `PublishCreateEventsBlocking()` - for FormData batch operations
- `PublishUpdateEventsBlocking()` - for FormData batch operations
- `PublishCustomEvent()` - for async action handlers

**Action Required**: Add documentation for blocking and custom event methods

---

#### 10. delegatehandler.go - Missing domainMappings Field
**Location**: `docs/workflow/architecture.md:96-104`
**Issue**: Documentation shows `domainMappings` field that doesn't exist in code

**Documentation says:**
```go
type DelegateHandler struct {
    log            *logger.Logger
    eventPublisher *EventPublisher
    domainMappings map[string]string  // domain -> entity name
}
```

**Actual code (delegatehandler.go:17-20):**
```go
type DelegateHandler struct {
    log       *logger.Logger
    publisher *EventPublisher
}
```

**Differences:**
- Field is `publisher` not `eventPublisher`
- No `domainMappings` field exists (entity name passed directly to RegisterDomain)

**Action Required**: Update documentation struct to match actual implementation

---

#### 11. engine.go - Struct Mismatch
**Location**: `docs/workflow/architecture.md:146-153`
**Issue**: Documentation shows simplified struct

**Documentation says:**
```go
type Engine struct {
    log            *logger.Logger
    db             *sqlx.DB
    triggerProc    *TriggerProcessor
    actionRegistry *ActionRegistry
    workflowBus    *Business
}
```

**Actual code (engine.go:16-34) has:**
- `triggerProcessor` not `triggerProc`
- `dependencies *DependencyResolver` (not documented)
- `executor *ActionExecutor` (not documented)
- No direct `actionRegistry` field (accessed via executor)
- State management fields: `mu`, `isInitialized`, `activeExecutions`, `executionHistory`, `stats`, `config`

**Action Required**: Update documentation to reflect actual structure (or note it's simplified)

---

#### 12. trigger.go - TriggerProcessor Struct Mismatch
**Location**: `docs/workflow/architecture.md:174-178`
**Issue**: Documentation shows different struct

**Documentation says:**
```go
type TriggerProcessor struct {
    log   *logger.Logger
    rules []AutomationRule
}
```

**Actual code (trigger.go:63-72):**
```go
type TriggerProcessor struct {
    log         *logger.Logger
    db          *sqlx.DB
    workflowBus *Business
    activeRules  []AutomationRuleView  // Note: AutomationRuleView not AutomationRule
    lastLoadTime time.Time
    cacheTimeout time.Duration
}
```

**Differences:**
- Uses `AutomationRuleView` not `AutomationRule`
- Has additional fields: `db`, `workflowBus`, `lastLoadTime`, `cacheTimeout`

**Action Required**: Update TriggerProcessor documentation

---

#### 13. trigger.go - Missing LoadRules Method
**Location**: `docs/workflow/architecture.md:180-181`
**Issue**: Documentation shows method that doesn't exist

**Documentation says:**
```go
func (tp *TriggerProcessor) LoadRules(rules []AutomationRule)
func (tp *TriggerProcessor) EvaluateEvent(event TriggerEvent) []MatchedRule
```

**Actual methods (trigger.go):**
- `Initialize(ctx)` - loads rules internally
- `ProcessEvent(ctx, event)` - returns `*ProcessingResult` not `[]MatchedRule`
- `RefreshRules(ctx)` - forces reload
- No public `LoadRules()` method

**Action Required**: Update TriggerProcessor method documentation

---

#### 14. models.go - Line Number References Outdated
**Location**: Multiple places in docs
**Issue**: Line number references in documentation may be outdated

Examples found:
- `docs/workflow/configuration/triggers.md:28` references `:152-159`
- `docs/workflow/configuration/rules.md:29` references `:238-252`

**Action Required**: Verify all line number references are still accurate after code changes

---

#### 15. templates.md - Old/New Field Change Variables Not Implemented
**Location**: `docs/workflow/configuration/templates.md:73-85`
**Issue**: Documentation describes `old_{field_name}` and `new_{field_name}` variables

**Documentation says:**
```
| `old_{field_name}` | Previous value |
| `new_{field_name}` | New value |
```

**Actual implementation (template.go):**
- Template processor doesn't automatically create `old_*` / `new_*` prefixed variables
- Field changes are accessed via the `FieldChanges` map, not as separate variables
- The delegatehandler.go:97-98 notes: "FieldChanges is nil in Phase 2"

**Action Required**: Either implement this feature or update documentation to reflect actual behavior

---

## Phase 3: Audit - Action Handlers
**Status**: ✅ COMPLETE (4 discrepancies found and fixed)

### Files to Audit Against
| File | Action Type |
|------|-------------|
| `business/sdk/workflow/workflowactions/register.go` | Registry |
| `business/sdk/workflow/workflowactions/communication/alert.go` | create_alert |
| `business/sdk/workflow/workflowactions/communication/email.go` | send_email |
| `business/sdk/workflow/workflowactions/communication/notification.go` | send_notification |
| `business/sdk/workflow/workflowactions/data/updatefield.go` | update_field |
| `business/sdk/workflow/workflowactions/inventory/allocate.go` | allocate_inventory |
| `business/sdk/workflow/workflowactions/approval/seek.go` | seek_approval |

### Audit Checklist
- [x] **register.go**: Verified all 6 action types are registered ✓
- [x] **alert.go**: Config schema matches, validation rules correct ✓
- [x] **email.go**: Config schema matches, validation rules correct ✓
- [x] **notification.go**: Config schema matches, channel types documented ✓
- [x] **updatefield.go**: Whitelist matches, operators match (except `not_in`) ✓
- [x] **allocate.go**: Modes, strategies, async processing all correct ✓
- [x] **seek.go**: Approval types correct ✓

### Discrepancies Found

#### 1. overview.md - ActionHandler Interface Still Shows Old Signature
**Location**: `docs/workflow/actions/overview.md:20-26`
**Issue**: Documentation shows incorrect interface signature (same issue as Phase 2)

**Documentation says:**
```go
type ActionHandler interface {
    Type() string
    Validate(config json.RawMessage) error
    Execute(ctx context.Context, config json.RawMessage, execCtx ExecutionContext) (*ActionResult, error)
}
```

**Actual code (interfaces.go:38-42):**
```go
type ActionHandler interface {
    Execute(ctx context.Context, config json.RawMessage, context ActionExecutionContext) (any, error)
    Validate(config json.RawMessage) error
    GetType() string
}
```

**Differences:**
- Method name is `GetType()` not `Type()`
- Method order is different
- Return type is `(any, error)` not `(*ActionResult, error)`
- Context type is `ActionExecutionContext` not `ExecutionContext`

**Action Required**: Update `docs/workflow/actions/overview.md:20-26` to match actual interface

---

#### 2. overview.md - ActionResult Struct Not Used in Code
**Location**: `docs/workflow/actions/overview.md:40-51`
**Issue**: Documentation describes ActionResult struct that handlers don't actually return

**Documentation says:**
```go
type ActionResult struct {
    Status       string
    ActionType   string
    Message      string
    Data         map[string]interface{}
    Error        string
    Duration     time.Duration
}
```

**Actual code:**
- Handlers return `map[string]interface{}` or custom structs (e.g., `QueuedAllocationResponse`)
- No standardized `ActionResult` type is used

**Action Required**: Either add note that this is a conceptual model or update to show actual return patterns

---

#### 3. overview.md - ExecutionContext Struct Name Wrong
**Location**: `docs/workflow/actions/overview.md:55-64`
**Issue**: Type name is wrong

**Documentation says:**
```go
type ExecutionContext struct { ... }
```

**Actual type name:**
```go
type ActionExecutionContext struct { ... }
```

**Action Required**: Update struct name in documentation

---

#### 4. update-field.md - Missing `not_in` Operator Implementation Note
**Location**: `docs/workflow/actions/update-field.md:70-71`
**Issue**: Documentation lists `not_in` operator but code doesn't implement it

**Documentation lists:**
```
| `not_in` | Value not in array |
```

**Actual code (updatefield.go:414-425):**
```go
validOperators := []string{
    "equals", "not_equals", "greater_than", "less_than",
    "contains", "is_null", "is_not_null", "in", "not_in",
}
```

However, in `buildWhereClause` (updatefield.go:228-252), `not_in` falls through to the default case which does `= :cond` instead of `NOT IN`.

**Action Required**: Either implement `not_in` properly in code or add note that it's pending implementation

---

### Fixes Applied

1. ✅ **Fixed overview.md:20-26** - Updated ActionHandler interface to show `GetType()`, correct return type `(any, error)`, and `ActionExecutionContext`
2. ✅ **Fixed overview.md:40-51** - Removed incorrect `ActionResult` struct, added note about return type flexibility
3. ✅ **Fixed overview.md:55-64** - Changed `ExecutionContext` to `ActionExecutionContext` with accurate fields
4. ✅ **Fixed update-field.md:70-71** - Added warning note about `not_in` operator limitation

---

## Phase 4: Audit - Database & API
**Status**: Not Started

### Files to Audit Against

**Database Migrations:**
- `business/sdk/migrate/sql/migrate.sql` (workflow tables ~versions 1.70-1.80)

**Business Layer:**
| Package | Purpose |
|---------|---------|
| `business/domain/workflow/alertbus/` | Alert business logic |
| `business/domain/workflow/alertbus/model.go` | Alert models |
| `business/domain/workflow/alertbus/stores/alertdb/` | Alert database store |

**API Layer:**
| Package | Purpose |
|---------|---------|
| `api/domain/http/workflow/alertapi/` | Alert HTTP handlers |
| `api/domain/http/workflow/alertapi/route.go` | Route definitions |
| `api/domain/http/workflow/alertapi/model.go` | API models |

### Audit Checklist
- [ ] **migrate.sql**: Verify all workflow tables match schema docs
- [ ] **alertbus/model.go**: Verify Alert, AlertRecipient, AlertAcknowledgment structs
- [ ] **alertdb/**: Verify CRUD operations match documentation
- [ ] **alertapi/route.go**: Verify all endpoints match API reference
- [ ] **alertapi/model.go**: Verify request/response models

### Discrepancies Found
_(To be filled during audit)_

---

## Phase 5: Final Review & Cleanup
**Status**: Not Started

### Tasks
- [ ] Review consolidated documentation for completeness
- [ ] Add cross-references between sections
- [ ] Ensure all code links use correct line numbers
- [ ] Move original files to `.archive/workflow-docs/`
- [ ] Update CLAUDE.md to reference new documentation location
- [ ] Add workflow section to main README if exists
- [ ] Verify all internal links work

### Archive File Mapping
| Original Location | Archive Location |
|-------------------|------------------|
| `business/sdk/workflow/docs/workflow-config-validator-spec.md` | `.archive/workflow-docs/workflow-config-validator-spec.md` |
| `business/sdk/workflow/workflowactions/communication/ALERT_SYSTEM.md` | `.archive/workflow-docs/ALERT_SYSTEM.md` |
| `business/sdk/workflow/workflowactions/data/updatefield_overview.md` | `.archive/workflow-docs/updatefield_overview.md` |
| `business/sdk/workflow/workflowactions/inventory/allocate_overview.md` | `.archive/workflow-docs/allocate_overview.md` |
| `.claude/plans/completed-plans/WORKFLOW_EVENT_FIRING_INFRASTRUCTURE.md` | `.archive/workflow-docs/WORKFLOW_EVENT_FIRING_INFRASTRUCTURE.md` |

---

## Progress Log

| Date | Phase | Notes |
|------|-------|-------|
| 2026-01-27 | Planning | Read all 5 source documents, created consolidation plan |
| 2026-01-27 | Phase 1 | Created 18 consolidated documentation files in docs/workflow/ |
| 2026-01-27 | Phase 2 | Audited 8 core engine files, found 15 discrepancies |
| 2026-01-27 | Phase 2 | Fixed all 15 discrepancies in architecture.md and templates.md |
| 2026-01-27 | Phase 3 | Audited 7 action handler files, found 4 discrepancies |
| 2026-01-27 | Phase 3 | Fixed all 4 discrepancies in overview.md and update-field.md |
| | | |

