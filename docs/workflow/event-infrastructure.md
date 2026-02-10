# Event Infrastructure

This document covers the TemporalDelegateHandler, delegate pattern, and how workflow events flow from business layer operations to Temporal for execution.

## Overview

The event infrastructure enables workflow automation by:

1. Capturing entity changes in the business layer via delegate events
2. Converting them to TriggerEvents via the TemporalDelegateHandler
3. Evaluating matching automation rules via WorkflowTrigger
4. Dispatching matched workflows to Temporal for durable execution

All events enter the workflow system through a single path: the delegate pattern.

```
Business Layer → delegate.Call() → TemporalDelegateHandler → WorkflowTrigger → Temporal
```

## TemporalDelegateHandler

**Location**: `business/sdk/workflow/temporal/delegatehandler.go`

Bridges the delegate pattern to the Temporal workflow system. Replaces the old `DelegateHandler` → `EventPublisher` → `RabbitMQ` pipeline with direct Temporal dispatch.

### Structure

```go
type DelegateHandler struct {
    log     *logger.Logger
    trigger *WorkflowTrigger
}
```

### Registration

Domains register with the delegate handler during startup:

```go
func (h *DelegateHandler) RegisterDomain(
    delegate *delegate.Delegate,
    domainName string,
    entityName string,
)
```

This registers listeners for:
- `{domainName}/created`
- `{domainName}/updated`
- `{domainName}/deleted`

### Action Mapping

| Delegate Action | Workflow Event Type |
|-----------------|---------------------|
| `created` | `on_create` |
| `updated` | `on_update` |
| `deleted` | `on_delete` |

### Handler Implementation

When a delegate event fires:

```go
func (h *DelegateHandler) handleEvent(ctx context.Context, data delegate.Data) error {
    // 1. Extract entity data from params (JSON + reflection)
    entityID, rawData, _ := extractEntityData(data.RawParams)

    // 2. Construct TriggerEvent
    event := workflow.TriggerEvent{
        EventType:  mapAction(data.Action),  // created → on_create
        EntityName: entityName,
        EntityID:   entityID,
        RawData:    rawData,
        // ...
    }

    // 3. Dispatch in goroutine (non-blocking, fail-open)
    go func() {
        if err := h.trigger.OnEntityEvent(ctx, event); err != nil {
            h.log.Error(ctx, "workflow trigger failed", "err", err)
        }
    }()

    return nil
}
```

**Key behaviors:**
- Non-blocking: dispatches in a goroutine so the primary operation always completes
- Fail-open: errors are logged but never returned to the caller
- Entity data extracted via JSON marshal/unmarshal with reflection fallback for ID

## Domain Event Pattern

Each domain that fires workflow events needs an `event.go` file.

### Structure

**Location**: `business/domain/{area}/{entity}bus/event.go`

```go
package ordersbus

import (
    "encoding/json"
    "github.com/google/uuid"
    "github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName for delegate routing
const DomainName = "order"

// EntityName for workflow entity matching (table name only)
const EntityName = "orders"

// Action constants
const (
    ActionCreated = "created"
    ActionUpdated = "updated"
    ActionDeleted = "deleted"
)

// ActionCreatedData constructs delegate data for creation
func ActionCreatedData(order Order) delegate.Data {
    params := ActionCreatedParms{
        EntityID: order.ID,
        UserID:   order.CreatedBy,
        Entity:   order,
    }
    rawParams, _ := params.Marshal()

    return delegate.Data{
        Domain:    DomainName,
        Action:    ActionCreated,
        RawParams: rawParams,
    }
}
```

### Params Structure

```go
type ActionCreatedParms struct {
    EntityID uuid.UUID `json:"entityID"`
    UserID   uuid.UUID `json:"userID"`
    Entity   Order     `json:"entity"`
}

func (p *ActionCreatedParms) Marshal() ([]byte, error) {
    return json.Marshal(p)
}
```

## Integration in Business Layer

Add delegate calls in the bus CRUD methods:

### Create

```go
func (b *Business) Create(ctx context.Context, no NewOrder) (Order, error) {
    order := Order{
        ID: b.delegate.GenerateUUID(),
        // ... other fields
    }

    if err := b.storer.Create(ctx, order); err != nil {
        return Order{}, fmt.Errorf("create: %w", err)
    }

    // Fire delegate event AFTER successful database operation
    if err := b.delegate.Call(ctx, ActionCreatedData(order)); err != nil {
        b.log.Error(ctx, "ordersbus: delegate call failed", "action", ActionCreated, "err", err)
    }

    return order, nil
}
```

### Update

```go
func (b *Business) Update(ctx context.Context, order Order, uo UpdateOrder) (Order, error) {
    // ... update logic ...

    if err := b.storer.Update(ctx, order); err != nil {
        return Order{}, fmt.Errorf("update: %w", err)
    }

    // Fire delegate event
    if err := b.delegate.Call(ctx, ActionUpdatedData(order)); err != nil {
        b.log.Error(ctx, "ordersbus: delegate call failed", "action", ActionUpdated, "err", err)
    }

    return order, nil
}
```

### Delete

```go
func (b *Business) Delete(ctx context.Context, order Order) error {
    if err := b.storer.Delete(ctx, order); err != nil {
        return fmt.Errorf("delete: %w", err)
    }

    // Fire delegate event
    if err := b.delegate.Call(ctx, ActionDeletedData(order)); err != nil {
        b.log.Error(ctx, "ordersbus: delegate call failed", "action", ActionDeleted, "err", err)
    }

    return nil
}
```

## Registration in all.go

Register domains with the TemporalDelegateHandler during startup:

```go
// In all.go Add() function (inside TemporalHostPort guard)

// Create delegate handler
delegateHandler := temporal.NewDelegateHandler(cfg.Log, workflowTrigger)

// Register domains
delegateHandler.RegisterDomain(delegate, ordersbus.DomainName, ordersbus.EntityName)
delegateHandler.RegisterDomain(delegate, customersbus.DomainName, customersbus.EntityName)
delegateHandler.RegisterDomain(delegate, productbus.DomainName, productbus.EntityName)
// ... register ~60 domains total
```

## TriggerEvent Structure

The event sent to the WorkflowTrigger:

```go
type TriggerEvent struct {
    EventType    string                    // "on_create", "on_update", "on_delete"
    EntityName   string                    // "orders", "customers", etc.
    EntityID     uuid.UUID                 // Entity's UUID
    FieldChanges map[string]FieldChange    // For on_update events
    Timestamp    time.Time                 // Event timestamp (UTC)
    RawData      map[string]interface{}    // Entity data snapshot
    UserID       uuid.UUID                 // User who triggered
}

type FieldChange struct {
    OldValue interface{}
    NewValue interface{}
}
```

## Event Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Business Layer                                     │
│                                                                             │
│  ordersbus.Create()                                                         │
│       │                                                                      │
│       ├── 1. Save to database                                               │
│       │                                                                      │
│       └── 2. delegate.Call(ActionCreatedData(order))                        │
│                    │                                                         │
└────────────────────┼─────────────────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Delegate System                                    │
│                                                                             │
│  3. Dispatches to TemporalDelegateHandler                                   │
│       │                                                                      │
│       └── Receives: domain="order", action="created", params={...}          │
│                    │                                                         │
└────────────────────┼─────────────────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                   TemporalDelegateHandler (goroutine)                        │
│                                                                             │
│  4. Extracts entity data, constructs TriggerEvent                           │
│       │                                                                      │
│       └── workflowTrigger.OnEntityEvent(ctx, triggerEvent)                  │
│                    │                                                         │
└────────────────────┼─────────────────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        WorkflowTrigger                                       │
│                                                                             │
│  5. Evaluates matching rules via TriggerProcessor                           │
│  6. Loads graph definition from EdgeStore for each match                    │
│       │                                                                      │
│       └── temporalClient.ExecuteWorkflow(ctx, workflowInput)                │
│                    │                                                         │
└────────────────────┼─────────────────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        Temporal → workflow-worker                             │
│                                                                             │
│  7. GraphExecutor traverses graph, executes actions as activities            │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Error Handling

### Non-Blocking Philosophy

Workflow failures should **never** block the primary operation:

```go
// In business layer
if err := b.delegate.Call(ctx, ActionCreatedData(order)); err != nil {
    b.log.Error(ctx, "ordersbus: delegate call failed", "err", err)
    // DON'T return error - continue with normal operation
}
return order, nil
```

### Error Logging

Errors are logged at multiple levels:
1. Business layer: delegate.Call() failures
2. TemporalDelegateHandler: event extraction or dispatch failures
3. WorkflowTrigger: rule matching or Temporal dispatch failures
4. Temporal: activity execution failures (with configurable retries)

## Cascade Visualization and the Event System

The cascade visualization feature leverages the event infrastructure to show downstream workflow effects. When an action handler implements the `EntityModifier` interface (see [Cascade Visualization](cascade-visualization.md)), it declares which entities it modifies and what events those modifications produce.

The cascade map endpoint (`/workflow/rules/{id}/cascade-map`) uses this information to:
1. Identify which events an action will emit (e.g., `update_field` on "orders" emits `on_update` for "orders")
2. Find other rules that trigger on those events
3. Build a graph of potential downstream workflows

## Condition Node Results

When workflows include `evaluate_condition` actions, the condition result flows through the system:

1. The condition handler evaluates field values and returns a result with `BranchTaken` ("true_branch" or "false_branch")
2. The GraphExecutor examines this result to determine which edges to follow
3. Temporal records the execution path for visibility

See [Branching](branching.md) for details on how condition results control execution flow.

## Related Documentation

- [Architecture](architecture.md) — System overview and component details
- [Adding Domains](adding-domains.md) — Step-by-step guide for adding workflow events
- [Testing](testing.md) — Testing patterns for event infrastructure
- [Cascade Visualization](cascade-visualization.md) — Understanding downstream workflow effects
- [Branching](branching.md) — Graph-based execution with conditional paths

## Key Files

| File | Purpose |
|------|---------|
| `business/sdk/workflow/temporal/delegatehandler.go` | Temporal delegate bridge |
| `business/sdk/workflow/temporal/trigger.go` | Rule matching and Temporal dispatch |
| `business/sdk/delegate/delegate.go` | Core delegate system |
| `business/domain/sales/ordersbus/event.go` | Example domain events |
| `api/cmd/services/ichor/build/all/all.go` | Registration |
