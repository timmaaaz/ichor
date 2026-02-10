# Phase 13: Dead Code Removal & Temporal Rewiring

**Category**: backend
**Status**: Pending
**Dependencies**: Phases 7-9 (Temporal wiring must be in place before removing old engine)

---

## Overview

Remove the old RabbitMQ-based workflow engine (~10,000 lines across 14 files) and wire Temporal as the sole execution path. This creates a `TemporalDelegateHandler` that bridges delegate events to `WorkflowTrigger.OnEntityEvent()`, adds cache invalidation to the TriggerProcessor, rewires `all.go`, and cleans up obsolete types from models/interfaces.

**Key Finding**: The workflow validator (`workflowsaveapp/graph.go`, `validation.go`) and seed data (`testutil.go`) are already 100% compatible with Temporal — no changes needed.

## Goals

1. Delete old engine files (7 source + 7 test files, ~10,000 lines)
2. Create `TemporalDelegateHandler` bridging delegate events → Temporal workflow dispatch
3. Rewire `all.go` with Temporal as the sole execution path (replace RabbitMQ engine block)

## Prerequisites

- Phases 7-9 completed (Temporal trigger, edge store, worker wiring in place)
- Understanding of delegate event pattern (`business/sdk/delegate`)
- Understanding of old engine pipeline: `DelegateHandler → EventPublisher → RabbitMQ → QueueManager → Engine`

### Key Signatures Reference

```go
// WorkflowTrigger (from Phase 7 - temporal/trigger.go)
func NewWorkflowTrigger(log *logger.Logger, starter WorkflowStarter, matcher RuleMatcher, store EdgeStore) *WorkflowTrigger
func (wt *WorkflowTrigger) OnEntityEvent(ctx context.Context, event workflow.TriggerEvent) error

// TriggerProcessor (from trigger.go)
func NewTriggerProcessor(log *logger.Logger, db *sqlx.DB, workflowBus *Business) *TriggerProcessor
func (tp *TriggerProcessor) Initialize(ctx context.Context) error
func (tp *TriggerProcessor) ProcessEvent(ctx context.Context, event TriggerEvent) (*ProcessingResult, error)

// Old DelegateHandler (being replaced - delegatehandler.go)
func NewDelegateHandler(log *logger.Logger, publisher *EventPublisher) *DelegateHandler
func (h *DelegateHandler) RegisterDomain(del *delegate.Delegate, domainName, entityName string)

// Old EventPublisher (being deleted - eventpublisher.go)
func (ep *EventPublisher) extractEntityData(result any) (uuid.UUID, map[string]any, error)
func (ep *EventPublisher) extractIDViaReflection(result any) uuid.UUID
```

---

## Task Breakdown

### Task 1: Move Shared Constants Out of `delegatehandler.go`

**Status**: Pending

**Description**: The old `delegatehandler.go` defines `ActionCreated`, `ActionUpdated`, `ActionDeleted` constants and `DelegateEventParams` struct. Move them to `event.go` so they survive the file deletion.

**IMPORTANT FINDING**: Each domain's `event.go` file defines its own local copies of `ActionCreated`/`ActionUpdated`/`ActionDeleted` (e.g., `ordersbus.ActionCreated = "created"`). They do NOT import from the workflow package. The constants in `delegatehandler.go` are only consumed by the `DelegateHandler.RegisterDomain` method itself and will be needed by the new `TemporalDelegateHandler`.

**Notes**:
- Move to `business/sdk/workflow/event.go` which currently holds rule-lifecycle constants (`ActionRuleCreated`, `ActionRuleUpdated`, etc.)
- `DelegateEventParams` struct needs to move too — it's the standard params structure for delegate event deserialization
- Verify no existing consumers break: `grep -r "workflow.ActionCreated" business/domain/` returns 0 matches (only `delegatehandler.go` and its test use them)

**Files**:
- `business/sdk/workflow/event.go` (ADD ~15 lines: 3 constants + DelegateEventParams struct)

**Implementation Guide**:

Add to `business/sdk/workflow/event.go` (after existing `ActionRuleDeactivated` constant, before `ActionRuleChangedParms`):

```go
// Standard action names matching what domain event.go files use.
// Used by DelegateHandler implementations to register for CRUD events.
const (
	ActionCreated = "created"
	ActionUpdated = "updated"
	ActionDeleted = "deleted"
)

// DelegateEventParams is the standard structure for delegate event parameters.
// Domain event.go files use this structure (or compatible layouts) for their
// ActionXxxParms types. The UserID field identifies who triggered the action.
type DelegateEventParams struct {
	EntityID uuid.UUID `json:"entityID"`
	UserID   uuid.UUID `json:"userID"`
	Entity   any       `json:"entity,omitempty"`
}
```

**Verification**:
```bash
grep -r "workflow\.ActionCreated\|workflow\.ActionUpdated\|workflow\.ActionDeleted\|workflow\.DelegateEventParams" business/
```
All consumers should still resolve — the package path doesn't change, only the file within the package.

---

### Task 2: Create `TemporalDelegateHandler`

**Status**: Pending

**Description**: New adapter that bridges delegate events to `WorkflowTrigger.OnEntityEvent()`. Replaces the old `DelegateHandler` which bridged to `EventPublisher` → RabbitMQ.

**Notes**:
- Same `RegisterDomain()` interface as old handler (compatible drop-in replacement in all.go)
- Calls `OnEntityEvent` in a goroutine (non-blocking, fail-open — matching old pattern)
- Copy `extractEntityData` and `extractIDViaReflection` from `eventpublisher.go` (source being deleted in Task 5)
- These helpers convert arbitrary entity structs to `(uuid.UUID, map[string]any)` via JSON roundtrip + reflection fallback
- Event type mapping: `ActionCreated` → `"on_create"`, `ActionUpdated` → `"on_update"`, `ActionDeleted` → `"on_delete"`

**Files**:
- `business/sdk/workflow/temporal/delegatehandler.go` (CREATE ~150 lines)

**Implementation Guide**:

```go
package temporal

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// DelegateHandler bridges domain delegate events to Temporal workflow dispatch.
// It registers handlers for domain CRUD events and converts them to TriggerEvents
// for WorkflowTrigger.OnEntityEvent().
type DelegateHandler struct {
	log     *logger.Logger
	trigger *WorkflowTrigger
}

// NewDelegateHandler creates a handler bridging delegate events to Temporal.
func NewDelegateHandler(log *logger.Logger, trigger *WorkflowTrigger) *DelegateHandler {
	return &DelegateHandler{
		log:     log,
		trigger: trigger,
	}
}

// RegisterDomain registers delegate handlers for a domain's CRUD events.
// entityName should match the workflow entity name (e.g., "orders").
func (h *DelegateHandler) RegisterDomain(del *delegate.Delegate, domainName, entityName string) {
	del.Register(domainName, workflow.ActionCreated, func(ctx context.Context, data delegate.Data) error {
		return h.handleEvent(ctx, "on_create", entityName, data)
	})
	del.Register(domainName, workflow.ActionUpdated, func(ctx context.Context, data delegate.Data) error {
		return h.handleEvent(ctx, "on_update", entityName, data)
	})
	del.Register(domainName, workflow.ActionDeleted, func(ctx context.Context, data delegate.Data) error {
		return h.handleEvent(ctx, "on_delete", entityName, data)
	})

	h.log.Info(context.Background(), "temporal delegate handler registered",
		"domain", domainName, "entity", entityName)
}

func (h *DelegateHandler) handleEvent(ctx context.Context, eventType, entityName string, data delegate.Data) error {
	var params workflow.DelegateEventParams
	if err := json.Unmarshal(data.RawParams, &params); err != nil {
		h.log.Error(ctx, "temporal delegate: unmarshal params failed",
			"entity", entityName, "event_type", eventType, "error", err)
		return nil // Don't fail the delegate chain
	}

	// Build TriggerEvent from delegate data.
	event := workflow.TriggerEvent{
		EventType:  eventType,
		EntityName: entityName,
		EntityID:   params.EntityID,
		Timestamp:  time.Now().UTC(),
		UserID:     params.UserID,
	}

	// Extract entity raw data if present.
	if params.Entity != nil {
		entityID, rawData, err := extractEntityData(params.Entity)
		if err == nil {
			event.RawData = rawData
			if event.EntityID == uuid.Nil {
				event.EntityID = entityID
			}
		}
	}

	// Fire in goroutine to avoid blocking the delegate chain.
	go func() {
		if err := h.trigger.OnEntityEvent(context.Background(), event); err != nil {
			h.log.Error(context.Background(), "temporal delegate: dispatch failed",
				"entity", entityName, "event_type", eventType, "error", err)
		}
	}()

	return nil
}

// extractEntityData extracts ID and raw data from an entity result.
// Copied from eventpublisher.go (source file being deleted in Task 5).
func extractEntityData(result any) (uuid.UUID, map[string]any, error) {
	if result == nil {
		return uuid.Nil, nil, fmt.Errorf("nil result")
	}

	data, err := json.Marshal(result)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("marshal result: %w", err)
	}

	var rawData map[string]any
	if err := json.Unmarshal(data, &rawData); err != nil {
		return uuid.Nil, nil, fmt.Errorf("unmarshal to map: %w", err)
	}

	// Extract ID from JSON (app layer uses string IDs).
	var entityID uuid.UUID
	if id, ok := rawData["id"].(string); ok {
		if parsed, err := uuid.Parse(id); err == nil {
			entityID = parsed
		}
	}

	// Fallback: reflection for struct field ID.
	if entityID == uuid.Nil {
		entityID = extractIDViaReflection(result)
	}

	return entityID, rawData, nil
}

func extractIDViaReflection(result any) uuid.UUID {
	val := reflect.ValueOf(result)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return uuid.Nil
	}

	idField := val.FieldByName("ID")
	if !idField.IsValid() {
		return uuid.Nil
	}

	switch id := idField.Interface().(type) {
	case uuid.UUID:
		return id
	case string:
		if parsed, err := uuid.Parse(id); err == nil {
			return parsed
		}
	}
	return uuid.Nil
}
```

---

### Task 3: Add `TriggerProcessor.RegisterCacheInvalidation()`

**Status**: Pending

**Description**: The old `engine.go` registered delegate handlers for rule lifecycle events (created/updated/deleted) to invalidate the rule cache. The standalone `TriggerProcessor` needs the same capability.

**IMPORTANT**: `RefreshRules()` already exists at `trigger.go:493-499` with correct locking semantics. It resets `lastLoadTime` (forcing cache expiry), releases the lock, then calls `loadMetadata` which acquires its own lock internally. **DO NOT rewrite RefreshRules** — its locking pattern is intentional. `loadMetadata` (line 99-100) acquires `tp.mu.Lock()` internally, so wrapping `loadMetadata` in a `Lock()/Unlock()` would **DEADLOCK**.

**Notes**:
- Only add `RegisterCacheInvalidation` method — `RefreshRules` already exists and works correctly
- Register handlers for `workflow.DomainName` + `ActionRuleCreated/Updated/Deleted/Activated/Deactivated`
- On any rule change event, call existing `RefreshRules()` which resets cache time and re-runs `loadMetadata`
- **Concurrency**: `TriggerProcessor` already has `tp.mu sync.RWMutex`. `RefreshRules` and `loadMetadata` handle locking internally. `ProcessEvent` uses `tp.mu.RLock()` for thread-safe reads.

**Files**:
- `business/sdk/workflow/trigger.go` (ADD ~20 lines: `RegisterCacheInvalidation` method only)

**Implementation Guide**:

```go
// RegisterCacheInvalidation registers delegate handlers for automation rule
// lifecycle events that invalidate the trigger processor's rule cache.
func (tp *TriggerProcessor) RegisterCacheInvalidation(del *delegate.Delegate) {
	handler := func(ctx context.Context, data delegate.Data) error {
		tp.log.Info(ctx, "trigger processor: refreshing rules", "action", data.Action)
		if err := tp.RefreshRules(ctx); err != nil {
			tp.log.Error(ctx, "trigger processor: refresh failed", "error", err)
		}
		return nil // Don't fail the delegate chain
	}

	del.Register(DomainName, ActionRuleCreated, handler)
	del.Register(DomainName, ActionRuleUpdated, handler)
	del.Register(DomainName, ActionRuleDeleted, handler)
	del.Register(DomainName, ActionRuleActivated, handler)
	del.Register(DomainName, ActionRuleDeactivated, handler)

	tp.log.Info(context.Background(), "trigger processor: cache invalidation registered")
}
```

**Existing RefreshRules reference** (DO NOT MODIFY — already at `trigger.go:493-499`):
```go
func (tp *TriggerProcessor) RefreshRules(ctx context.Context) error {
	tp.mu.Lock()
	tp.lastLoadTime = time.Time{} // Reset cache time to force reload
	tp.mu.Unlock()
	return tp.loadMetadata(ctx)   // loadMetadata acquires its own lock
}
```

---

### Task 4: Rewire `all.go`

**Status**: Pending

**Description**: Replace the old RabbitMQ workflow block (lines 438-565) with Temporal as the sole execution path. Make Temporal block log a warning if disabled (don't crash).

**Notes**:

**Remove** (lines 438-565 — old RabbitMQ workflow block):
- `var eventPublisher *workflow.EventPublisher`
- `var queueManager *workflow.QueueManager`
- `var workflowQueue *rabbitmq.WorkflowQueue`
- `if cfg.RabbitClient != nil && cfg.RabbitClient.IsConnected()` block
- `workflow.NewEngine()`, `workflow.NewQueueManager()`, `workflow.NewEventPublisher()`
- `workflowactions.RegisterAll()` with 9 bus dependencies
- `workflow.NewDelegateHandler()` and all 60 `RegisterDomain()` calls

**Remove** (lines 1164-1175 — WebSocket alert delivery block):
- `if queueManager != nil && workflowQueue != nil` block
- `alertws.NewAlertConsumer()` and `queueManager.SetHandlerRegistry()`

**Update** alertapi.Routes:
- Change `WorkflowQueue: workflowQueue` to `WorkflowQueue: nil`
- alertapi already handles nil gracefully (`alert.go` guards with `if h.workflowQueue != nil`)

**Replace** current Temporal block (lines 567-613) with unconditional version:
- Remove `if cfg.TemporalHostPort != ""` guard
- When empty, log info and skip (no crash, no dispatch)
- After creating `workflowTrigger`:
  - Call `triggerProcessor.RegisterCacheInvalidation(delegate)`
  - Create `temporalpkg.NewDelegateHandler(cfg.Log, workflowTrigger)`
  - Register same ~60 domains as old block
- Remove `_ = workflowTrigger` placeholder (line 610)

**Remove unused imports**: `workflow.NewEngine`, `workflow.NewQueueManager`, `workflow.NewEventPublisher`, `workflow.NewDelegateHandler`

**Files**:
- `api/cmd/services/ichor/build/all/all.go` (MODIFY: ~130 lines removed, ~80 lines added/changed)

**Implementation Guide**:

The new Temporal block (replacing both old RabbitMQ block AND current conditional Temporal block):

```go
// =========================================================================
// Initialize Temporal Workflow Infrastructure
// =========================================================================

if cfg.TemporalHostPort != "" {
	tc, err := client.Dial(client.Options{
		HostPort: cfg.TemporalHostPort,
	})
	if err != nil {
		cfg.Log.Error(context.Background(),
			"temporal: client creation failed, workflow dispatch disabled",
			"error", err,
		)
	} else {
		defer tc.Close()

		edgeStore := edgedb.NewStore(cfg.Log, cfg.DB)

		triggerProcessor := workflow.NewTriggerProcessor(cfg.Log, cfg.DB, workflowBus)
		if err := triggerProcessor.Initialize(context.Background()); err != nil {
			cfg.Log.Error(context.Background(),
				"temporal: trigger processor init failed", "error", err,
			)
		} else {
			workflowTrigger := temporalpkg.NewWorkflowTrigger(
				cfg.Log, tc, triggerProcessor, edgeStore,
			)

			// Register cache invalidation for rule lifecycle events.
			triggerProcessor.RegisterCacheInvalidation(delegate)

			// Create Temporal delegate handler and register all domains.
			delegateHandler := temporalpkg.NewDelegateHandler(cfg.Log, workflowTrigger)

			// Sales domain
			delegateHandler.RegisterDomain(delegate, ordersbus.DomainName, ordersbus.EntityName)
			delegateHandler.RegisterDomain(delegate, customersbus.DomainName, customersbus.EntityName)
			delegateHandler.RegisterDomain(delegate, orderlineitemsbus.DomainName, orderlineitemsbus.EntityName)
			delegateHandler.RegisterDomain(delegate, orderfulfillmentstatusbus.DomainName, orderfulfillmentstatusbus.EntityName)
			delegateHandler.RegisterDomain(delegate, lineitemfulfillmentstatusbus.DomainName, lineitemfulfillmentstatusbus.EntityName)

			// Assets domain
			delegateHandler.RegisterDomain(delegate, assetbus.DomainName, assetbus.EntityName)
			delegateHandler.RegisterDomain(delegate, validassetbus.DomainName, validassetbus.EntityName)
			delegateHandler.RegisterDomain(delegate, userassetbus.DomainName, userassetbus.EntityName)
			delegateHandler.RegisterDomain(delegate, assettypebus.DomainName, assettypebus.EntityName)
			delegateHandler.RegisterDomain(delegate, assetconditionbus.DomainName, assetconditionbus.EntityName)
			delegateHandler.RegisterDomain(delegate, assettagbus.DomainName, assettagbus.EntityName)
			delegateHandler.RegisterDomain(delegate, tagbus.DomainName, tagbus.EntityName)
			delegateHandler.RegisterDomain(delegate, approvalstatusbus.DomainName, approvalstatusbus.EntityName)
			delegateHandler.RegisterDomain(delegate, fulfillmentstatusbus.DomainName, fulfillmentstatusbus.EntityName)

			// Core domain
			delegateHandler.RegisterDomain(delegate, userbus.DomainName, userbus.EntityName)
			delegateHandler.RegisterDomain(delegate, rolebus.DomainName, rolebus.EntityName)
			delegateHandler.RegisterDomain(delegate, userrolebus.DomainName, userrolebus.EntityName)
			delegateHandler.RegisterDomain(delegate, tableaccessbus.DomainName, tableaccessbus.EntityName)
			delegateHandler.RegisterDomain(delegate, pagebus.DomainName, pagebus.EntityName)
			delegateHandler.RegisterDomain(delegate, paymenttermbus.DomainName, paymenttermbus.EntityName)
			delegateHandler.RegisterDomain(delegate, currencybus.DomainName, currencybus.EntityName)
			delegateHandler.RegisterDomain(delegate, rolepagebus.DomainName, rolepagebus.EntityName)
			delegateHandler.RegisterDomain(delegate, contactinfosbus.DomainName, contactinfosbus.EntityName)

			// HR domain
			delegateHandler.RegisterDomain(delegate, approvalbus.DomainName, approvalbus.EntityName)
			delegateHandler.RegisterDomain(delegate, commentbus.DomainName, commentbus.EntityName)
			delegateHandler.RegisterDomain(delegate, homebus.DomainName, homebus.EntityName)
			delegateHandler.RegisterDomain(delegate, officebus.DomainName, officebus.EntityName)
			delegateHandler.RegisterDomain(delegate, reportstobus.DomainName, reportstobus.EntityName)
			delegateHandler.RegisterDomain(delegate, titlebus.DomainName, titlebus.EntityName)

			// Geography domain (countrybus/regionbus read-only, no events)
			delegateHandler.RegisterDomain(delegate, citybus.DomainName, citybus.EntityName)
			delegateHandler.RegisterDomain(delegate, streetbus.DomainName, streetbus.EntityName)
			delegateHandler.RegisterDomain(delegate, timezonebus.DomainName, timezonebus.EntityName)

			// Products domain
			delegateHandler.RegisterDomain(delegate, productbus.DomainName, productbus.EntityName)
			delegateHandler.RegisterDomain(delegate, productcategorybus.DomainName, productcategorybus.EntityName)
			delegateHandler.RegisterDomain(delegate, brandbus.DomainName, brandbus.EntityName)
			delegateHandler.RegisterDomain(delegate, productcostbus.DomainName, productcostbus.EntityName)
			delegateHandler.RegisterDomain(delegate, costhistorybus.DomainName, costhistorybus.EntityName)
			delegateHandler.RegisterDomain(delegate, physicalattributebus.DomainName, physicalattributebus.EntityName)
			delegateHandler.RegisterDomain(delegate, metricsbus.DomainName, metricsbus.EntityName)

			// Procurement domain
			delegateHandler.RegisterDomain(delegate, supplierbus.DomainName, supplierbus.EntityName)
			delegateHandler.RegisterDomain(delegate, supplierproductbus.DomainName, supplierproductbus.EntityName)
			delegateHandler.RegisterDomain(delegate, purchaseorderbus.DomainName, purchaseorderbus.EntityName)
			delegateHandler.RegisterDomain(delegate, purchaseorderlineitembus.DomainName, purchaseorderlineitembus.EntityName)
			delegateHandler.RegisterDomain(delegate, purchaseorderstatusbus.DomainName, purchaseorderstatusbus.EntityName)
			delegateHandler.RegisterDomain(delegate, purchaseorderlineitemstatusbus.DomainName, purchaseorderlineitemstatusbus.EntityName)

			// Inventory domain
			delegateHandler.RegisterDomain(delegate, warehousebus.DomainName, warehousebus.EntityName)
			delegateHandler.RegisterDomain(delegate, zonebus.DomainName, zonebus.EntityName)
			delegateHandler.RegisterDomain(delegate, inventorylocationbus.DomainName, inventorylocationbus.EntityName)
			delegateHandler.RegisterDomain(delegate, inventoryitembus.DomainName, inventoryitembus.EntityName)
			delegateHandler.RegisterDomain(delegate, inventorytransactionbus.DomainName, inventorytransactionbus.EntityName)
			delegateHandler.RegisterDomain(delegate, inventoryadjustmentbus.DomainName, inventoryadjustmentbus.EntityName)
			delegateHandler.RegisterDomain(delegate, transferorderbus.DomainName, transferorderbus.EntityName)
			delegateHandler.RegisterDomain(delegate, inspectionbus.DomainName, inspectionbus.EntityName)
			delegateHandler.RegisterDomain(delegate, lottrackingsbus.DomainName, lottrackingsbus.EntityName)
			delegateHandler.RegisterDomain(delegate, serialnumberbus.DomainName, serialnumberbus.EntityName)

			// Config domain
			delegateHandler.RegisterDomain(delegate, formbus.DomainName, formbus.EntityName)
			delegateHandler.RegisterDomain(delegate, formfieldbus.DomainName, formfieldbus.EntityName)
			delegateHandler.RegisterDomain(delegate, pageconfigbus.DomainName, pageconfigbus.EntityName)
			delegateHandler.RegisterDomain(delegate, pagecontentbus.DomainName, pagecontentbus.EntityName)
			delegateHandler.RegisterDomain(delegate, pageactionbus.DomainName, pageactionbus.EntityName)

			cfg.Log.Info(context.Background(), "temporal workflow infrastructure initialized",
				"temporal_host", cfg.TemporalHostPort,
			)
		}
	}
} else {
	cfg.Log.Info(context.Background(),
		"temporal: disabled (ICHOR_TEMPORAL_HOSTPORT not set)")
}
```

---

### Task 5: Delete Old Engine Files

**Status**: Pending

**Description**: Delete 7 source files and 7 test files comprising the old RabbitMQ-based workflow engine (~10,000 lines total).

**Source files to DELETE** (7 files, ~4,100 lines):

| File | ~Lines | Purpose |
|------|--------|---------|
| `business/sdk/workflow/engine.go` | 558 | Old workflow engine with singleton pattern |
| `business/sdk/workflow/executor.go` | 750 | Old action executor with registry |
| `business/sdk/workflow/dependency.go` | 589 | Dependency resolution and batch ordering |
| `business/sdk/workflow/queue.go` | 1,129 | RabbitMQ queue manager with circuit breaker |
| `business/sdk/workflow/notificationQueue.go` | 749 | Notification queue processing |
| `business/sdk/workflow/eventpublisher.go` | 216 | Event publisher for RabbitMQ |
| `business/sdk/workflow/delegatehandler.go` | 113 | Old delegate event bridge |

**Test files to DELETE** (7 files, ~5,900 lines):

| File | Purpose |
|------|---------|
| `business/sdk/workflow/executor_test.go` | Executor unit tests |
| `business/sdk/workflow/executor_graph_test.go` | Executor graph tests |
| `business/sdk/workflow/queue_test.go` | Queue manager tests |
| `business/sdk/workflow/notificationQueue_test.go` | Notification queue tests |
| `business/sdk/workflow/eventpublisher_test.go` | Event publisher tests |
| `business/sdk/workflow/eventpublisher_integration_test.go` | Event publisher integration tests |
| `business/sdk/workflow/delegatehandler_test.go` | Delegate handler tests |

**Files NOT being deleted** (still needed):
- `actionservice.go` — independent of old engine, used by `actionapi`
- `trigger.go` — `TriggerProcessor` used by Temporal's `WorkflowTrigger`
- `template.go` + tests — template variable resolution, used by action handlers
- `interfaces.go` — `ActionHandler`, `ActionRegistry` (partial cleanup in Task 6)
- `models.go` — shared types (partial cleanup in Task 6)
- `event.go`, `filter.go`, `order.go`, `executionfilter.go`, `testutil.go`, `workflowbus.go`

**Implementation**: `rm` each file. No archiving needed — git history preserves everything.

---

### Task 6: Clean Up `models.go` and `interfaces.go`

**Status**: Pending

**Description**: Remove types only used by the old engine from `models.go` and `interfaces.go`. Remove `ProcessQueued` methods from async action handlers.

**From `models.go` remove** (~115 lines, lines 30-144):

| Type | Lines | Used By |
|------|-------|---------|
| `ExecutionBatch` | 30-37 | `dependency.go`, `engine.go` |
| `ExecutionPlan` | 39-48 | `engine.go` |
| `WorkflowExecution` | 50-62 | `engine.go`, `executor.go` |
| `BatchResult` | 89-97 | `engine.go` |
| `RuleResult` | 99-109 | `engine.go` |
| `ActionResult` | 111-123 | `executor.go` |
| `WorkflowConfig` | 125-133 | `engine.go` |
| `WorkflowStats` | 135-144 | `engine.go` |

**KEEP** in models.go: `TriggerEvent`, `FieldChange`, `ExecutionStatus` constants, `TriggerSource` constants, `EventType` constants, `ActionExecutionContext`, all CRUD models, `ActionEdge`, `ConditionResult`, notification models, allocation models, view types

**From `interfaces.go` remove** (~35 lines, lines 94-131):

| Type | Lines | Replaced By |
|------|-------|-------------|
| `AsyncActionHandler` interface | 98-113 | Temporal's `AsyncActivityHandler` |
| `QueuedPayload` struct | 115-130 | No longer needed (old queue pattern) |

**KEEP** in interfaces.go: `ActionHandler`, `ActionRegistry` + methods, `EntityModifier`, `EntityModification`

**From workflowactions remove** `ProcessQueued` methods:
- `business/sdk/workflow/workflowactions/inventory/allocate.go`: Remove `ProcessQueued()` method (~80 lines) and `fireAllocationResultEvent()` method (~30 lines)
- `business/sdk/workflow/workflowactions/inventory/allocate_test.go`: Remove `Test_ProcessQueued` and `Test_ProcessQueued_FiresEventOnFailure` test functions
- `communication/send_email.go`: CONFIRMED no `ProcessQueued` method exists (grep verified) — no changes needed

**Files**:
- `business/sdk/workflow/models.go` (MODIFY: remove ~115 lines)
- `business/sdk/workflow/interfaces.go` (MODIFY: remove ~35 lines)
- `business/sdk/workflow/workflowactions/inventory/allocate.go` (MODIFY: remove 2 methods)
- `business/sdk/workflow/workflowactions/inventory/allocate_test.go` (MODIFY: remove 2 test functions)

---

### Task 7: Fix Compilation and Run Tests

**Status**: Pending

**Description**: After all deletions and modifications, fix any remaining compilation errors and verify all tests pass.

**Steps**:
```bash
go build ./...                                    # Find broken imports
go vet ./...                                      # Static analysis
go test ./business/sdk/workflow/...               # Surviving workflow tests
go test ./business/sdk/workflow/temporal/...       # Temporal tests (unaffected)
```

**Known compilation breakpoints**:

1. **`api/sdk/http/apitest/workflow.go`** (117 lines): References `workflow.Engine`, `workflow.QueueManager`, `workflow.EventPublisher`, `rabbitmq.Client`. **Replace with stub** that compiles but skips tests needing old infra. Don't rewrite for Temporal — Phase 15 handles the full test infrastructure rewrite.

**Stub implementation** (replaces entire file):
```go
package apitest

import (
	"testing"

	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
)

// WorkflowInfra holds the workflow infrastructure components for tests.
// NOTE: Old RabbitMQ-based engine removed in Phase 13 (Temporal migration).
// Full Temporal test infrastructure will be added in Phase 15.
type WorkflowInfra struct {
	WorkflowBus *workflow.Business
}

// InitWorkflowInfra sets up minimal workflow infrastructure for testing.
// The old RabbitMQ-based engine/queue manager has been removed.
// Tests requiring workflow execution should use Temporal test server (Phase 15).
func InitWorkflowInfra(t *testing.T, db *dbtest.Database) *WorkflowInfra {
	t.Helper()

	workflowBus := workflow.NewBusiness(db.Log, db.BusDomain.Delegate, workflowdb.NewStore(db.Log, db.DB))

	t.Log("Workflow infrastructure initialized (Temporal migration stub - no execution engine)")

	return &WorkflowInfra{
		WorkflowBus: workflowBus,
	}
}
```

**Note**: Tests that call `InitWorkflowInfra` and then access `wf.QueueManager`, `wf.Engine`, or `wf.Client` will fail to compile. Those tests are rewritten in Phase 15. For Phase 13, the stub only needs to make `go build ./...` pass — the affected integration tests (`trigger_test.go`, `execution_seed_test.go`, `errors_test.go`, `workflow_test.go` in formdataapi/ordersapi) will need field access updates or temporary commenting.

2. **`allocate_test.go`**: May reference `EventPublisher` in remaining tests beyond the 2 being removed. Check and fix.

3. **Import cleanup**: Remove unused imports of deleted packages in any file.

**Files**:
- `api/sdk/http/apitest/workflow.go` (REWRITE: stub without old engine types)
- Various files as needed based on `go build` errors

---

## Validation Criteria

- [ ] `go build ./...` compiles cleanly (zero errors)
- [ ] `go vet ./...` passes
- [ ] `go test ./business/sdk/workflow/...` passes (surviving tests)
- [ ] `go test ./business/sdk/workflow/temporal/...` passes (all 117 Temporal tests)
- [ ] All ~60 domain `RegisterDomain` calls present in `all.go` Temporal block
- [ ] No references to `workflow.NewEngine`, `workflow.NewQueueManager`, `workflow.NewEventPublisher`, `workflow.NewDelegateHandler` in non-test code
- [ ] Old files deleted: 14 files, ~10,000 lines removed
- [ ] `TemporalDelegateHandler` created with `RegisterDomain` method
- [ ] `TriggerProcessor.RegisterCacheInvalidation` added and called in `all.go`
- [ ] `alertapi.Routes` uses `WorkflowQueue: nil`
- [ ] WebSocket alert delivery block removed

---

## Deliverables

- Old engine removed (~4,600 lines of source + ~5,400 lines of tests = 14 files total)
- `business/sdk/workflow/temporal/delegatehandler.go` created (~150 lines)
- `business/sdk/workflow/trigger.go` updated with `RegisterCacheInvalidation` (RefreshRules already exists)
- `business/sdk/workflow/event.go` updated with shared constants
- `api/cmd/services/ichor/build/all/all.go` rewired: Temporal as sole execution path
- `business/sdk/workflow/models.go` cleaned: ~115 lines of obsolete types removed
- `business/sdk/workflow/interfaces.go` cleaned: ~35 lines of obsolete types removed
- `allocate.go`: `ProcessQueued` methods removed

---

## Gotchas & Tips

### Common Pitfalls

1. **apitest/workflow.go breaks compilation**: This file references `workflow.Engine`, `QueueManager`, `EventPublisher`. Don't try to rewrite it now — stub it out so it compiles and leave the full rewrite for Phase 15 (Integration Verification). This is the biggest "surprise" compilation error.

2. **Alert WebSocket delivery stops**: The old engine delivered alerts to WebSocket clients via RabbitMQ queue → AlertConsumer. Removing the old engine means alerts from Temporal activities still write to DB (via `create_alert` action handler) but won't push to WebSocket in real-time. REST API polling still works. Fix in a future phase.

3. **Domain event.go files are safe**: Each domain defines its own `ActionCreated`/`ActionUpdated`/`ActionDeleted` constants locally. They don't import from `workflow.delegatehandler`. Grep confirms 0 cross-package references. Deleting `delegatehandler.go` won't break any domain package.

4. **`extractEntityData` must be copied first**: The old `EventPublisher.extractEntityData()` and `extractIDViaReflection()` methods are the only way to convert arbitrary entity structs to `(uuid.UUID, map[string]any)`. Copy them to `TemporalDelegateHandler` as package-level functions BEFORE deleting `eventpublisher.go`.

5. **TriggerProcessor concurrency**: `RegisterCacheInvalidation` handlers may fire concurrently with `ProcessEvent`. Verify `TriggerProcessor` has `mu sync.RWMutex` and that `loadMetadata` acquires it. If not, add proper locking.

6. **FormData event publishing is safe**: VERIFIED — `formdata_registry.go` has ZERO references to `EventPublisher`. FormData relies on business layer CRUD operations which fire delegate events independently. Deleting `EventPublisher` does NOT break FormData workflow triggers. The `PublishCreateEventsBlocking`/`PublishUpdateEventsBlocking` methods were a duplicate event path used only by `all.go`'s old RabbitMQ block.

7. **`workflowQueue` nil safety for alertapi**: `alertapi.Routes` accepts `WorkflowQueue: nil` safely. The `CreateAlertHandler` guards with `if h.workflowQueue != nil` at `alert.go:197`.

8. **ProcessQueued removal in allocate.go**: `ProcessQueued` is only called by the old `queue.go` consumer loop (being deleted). No external callers exist. Safe to remove both the method and its tests.

### Tips

- Delete files BEFORE fixing compilation — it's easier to see what's actually broken when the old code is gone
- Run `go build ./...` after each task to catch errors incrementally
- The 60 `RegisterDomain` calls can be copied verbatim from the old RabbitMQ block — same domain/entity name pairs
- Use `git diff --stat` after completion to verify the expected ~10K line reduction
- Run `go test -count=1 ./business/sdk/workflow/temporal/...` after Task 2 to verify TemporalDelegateHandler doesn't break existing tests
- Task execution order matters: Task 1 → Task 2 → Task 3 → Task 4 → Task 5 → Task 6 → Task 7 (create new code first, rewire all.go, then delete old files, then clean up types)

---

## Testing Strategy

### Unit Tests

- **TemporalDelegateHandler**: Not creating dedicated tests in Phase 13. The handler is simple (unmarshal → build event → goroutine dispatch). Integration coverage comes from Phase 15.
- **RegisterCacheInvalidation**: Verify 5 delegate registrations happen (functional test via Phase 15 integration)
- **RefreshRules**: Verify rule reload from DB (tested implicitly when cache invalidation fires)

### Regression Tests

- All 117 existing Temporal package tests must continue to pass (they don't depend on old engine)
- Surviving workflow package tests must pass: trigger, template, actionservice, models
- Integration tests that used old workflow infrastructure (`apitest/workflow.go`) will be stubbed — Phase 15 scope

### Verification Commands

```bash
# After all tasks complete:
go build ./...                                    # Must compile
go vet ./...                                      # Must pass
go test ./business/sdk/workflow/...               # Surviving tests
go test ./business/sdk/workflow/temporal/...       # All 117 Temporal tests
go test ./business/sdk/workflow/workflowactions/...  # Action handler tests
```

---

## Known Side Effects

1. **Alert WebSocket delivery**: `alertws.AlertConsumer` uses RabbitMQ for real-time alerts. With old engine removed, alerts from Temporal activities still write to DB but won't push to WebSocket in real-time. Polling via REST API still works. Fix in a future phase.

2. **allocate_test.go**: `Test_ProcessQueued` and `Test_ProcessQueued_FiresEventOnFailure` being removed. Other allocate tests should be unaffected.

3. **FormData event publishing**: VERIFIED SAFE. `formdata_registry.go` has no `EventPublisher` dependency. FormData CRUD fires delegate events through business layer operations, which TemporalDelegateHandler will pick up. No behavioral change.

---

## Commands Reference

```bash
# Start this phase
/workflow-temporal-next

# Validate this phase
/workflow-temporal-validate 13

# Review plan before implementing
/workflow-temporal-plan-review 13

# Review code after implementing
/workflow-temporal-review 13
```
