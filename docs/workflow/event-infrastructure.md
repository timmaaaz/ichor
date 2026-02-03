# Event Infrastructure

This document covers the EventPublisher, DelegateHandler, and how workflow events flow from business layer operations to the workflow engine.

## Overview

The event infrastructure enables workflow automation by:

1. Capturing entity changes in the business layer
2. Converting them to workflow events
3. Queuing them for asynchronous processing
4. Executing matched automation rules

## Two Entry Points

Events can enter the workflow system through two paths:

### Path 1: FormData (App Layer)

For multi-entity transactions via the formdata API:

```
formdata API → UpsertFormData() → tx.Commit() → EventPublisher
```

Events fire after transaction commit, ensuring data consistency.

### Path 2: Delegate Pattern (Business Layer)

For direct API calls and all other entry points:

```
ordersapi → ordersapp → ordersbus.Create() → delegate.Call() → DelegateHandler → EventPublisher
```

Events fire within the business layer via the delegate pattern.

## EventPublisher

**Location**: `business/sdk/workflow/eventpublisher.go`

The EventPublisher provides non-blocking workflow event publishing.

### Structure

```go
type EventPublisher struct {
    log          *logger.Logger
    queueManager *QueueManager
}
```

### Methods

#### PublishCreateEvent

Fires an `on_create` event.

```go
func (ep *EventPublisher) PublishCreateEvent(
    ctx context.Context,
    entityName string,
    result interface{},
    userID uuid.UUID,
)
```

**Parameters:**
- `entityName`: Table name (e.g., "orders")
- `result`: Entity data (usually app layer response)
- `userID`: User who created the entity

#### PublishUpdateEvent

Fires an `on_update` event with field changes.

```go
func (ep *EventPublisher) PublishUpdateEvent(
    ctx context.Context,
    entityName string,
    result interface{},
    fieldChanges map[string]FieldChange,
    userID uuid.UUID,
)
```

**Parameters:**
- `fieldChanges`: Map of field names to old/new values

#### PublishDeleteEvent

Fires an `on_delete` event.

```go
func (ep *EventPublisher) PublishDeleteEvent(
    ctx context.Context,
    entityName string,
    entityID uuid.UUID,
    userID uuid.UUID,
)
```

### Non-Blocking Design

Events are queued in a goroutine to avoid blocking the primary operation:

```go
func (ep *EventPublisher) queueEventNonBlocking(ctx context.Context, event TriggerEvent) {
    go func() {
        queueCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        if err := ep.queueManager.QueueEvent(queueCtx, event); err != nil {
            ep.log.Error(queueCtx, "workflow event: queue failed", ...)
            // Never fail the primary operation
        }
    }()
}
```

**Key points:**
- 5-second timeout for queue operations
- Errors are logged but never returned
- Primary operation always completes

### Entity Data Extraction

The EventPublisher extracts entity ID and data from the result:

1. JSON marshal the result to get a map
2. Look for `id` field in the JSON
3. Fallback to reflection for struct field `ID`

```go
func (ep *EventPublisher) extractEntityData(result interface{}) (uuid.UUID, map[string]interface{}, error) {
    // JSON marshal/unmarshal for map representation
    data, _ := json.Marshal(result)
    var rawData map[string]interface{}
    json.Unmarshal(data, &rawData)

    // Extract ID from JSON
    if id, ok := rawData["id"].(string); ok {
        entityID, _ = uuid.Parse(id)
    }

    return entityID, rawData, nil
}
```

## DelegateHandler

**Location**: `business/sdk/workflow/delegatehandler.go`

Bridges the delegate pattern to workflow events.

### Structure

```go
type DelegateHandler struct {
    log            *logger.Logger
    eventPublisher *EventPublisher
    domainMappings map[string]string  // domain -> entity name
}
```

### Registration

Domains register with the delegate handler:

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
    // 1. Parse params from delegate data
    var params struct {
        EntityID uuid.UUID       `json:"entityID"`
        UserID   uuid.UUID       `json:"userID"`
        Entity   json.RawMessage `json:"entity"`
    }
    json.Unmarshal(data.RawParams, &params)

    // 2. Get entity name from mapping
    entityName := h.domainMappings[data.Domain]

    // 3. Publish appropriate event
    switch data.Action {
    case "created":
        h.eventPublisher.PublishCreateEvent(ctx, entityName, params.Entity, params.UserID)
    case "updated":
        h.eventPublisher.PublishUpdateEvent(ctx, entityName, params.Entity, nil, params.UserID)
    case "deleted":
        h.eventPublisher.PublishDeleteEvent(ctx, entityName, params.EntityID, params.UserID)
    }

    return nil
}
```

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

Register domains with the delegate handler during startup:

```go
// In all.go Add() function

// Create delegate handler
delegateHandler := workflow.NewDelegateHandler(cfg.Log, eventPublisher)

// Register domains
delegateHandler.RegisterDomain(delegate, ordersbus.DomainName, ordersbus.EntityName)
delegateHandler.RegisterDomain(delegate, customersbus.DomainName, customersbus.EntityName)
delegateHandler.RegisterDomain(delegate, productbus.DomainName, productbus.EntityName)
// ... register other domains
```

## TriggerEvent Structure

The final event sent to the workflow engine:

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
│  3. Dispatches to DelegateHandler                                           │
│       │                                                                      │
│       └── Receives: domain="order", action="created", params={...}          │
│                    │                                                         │
└────────────────────┼─────────────────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        DelegateHandler                                       │
│                                                                             │
│  4. Maps domain to entity name                                              │
│       │                                                                      │
│       └── eventPublisher.PublishCreateEvent("orders", entity, userID)       │
│                    │                                                         │
└────────────────────┼─────────────────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        EventPublisher (goroutine)                            │
│                                                                             │
│  5. Constructs TriggerEvent                                                 │
│       │                                                                      │
│       └── queueManager.QueueEvent(ctx, event)                               │
│                    │                                                         │
└────────────────────┼─────────────────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           RabbitMQ                                           │
│                                                                             │
│  6. Event queued for async processing                                       │
│                    │                                                         │
└────────────────────┼─────────────────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        QueueManager Consumer                                 │
│                                                                             │
│  7. Picks up message                                                        │
│       │                                                                      │
│       └── engine.ExecuteWorkflow(ctx, event)                                │
│                    │                                                         │
└────────────────────┼─────────────────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        WorkflowEngine                                        │
│                                                                             │
│  8. Evaluates rules, executes matching actions                              │
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
2. DelegateHandler: event processing failures
3. EventPublisher: queue failures
4. QueueManager: processing failures

## Cascade Visualization and the Event System

The cascade visualization feature leverages the event infrastructure to show downstream workflow effects. When an action handler implements the `EntityModifier` interface (see [Cascade Visualization](cascade-visualization.md)), it declares which entities it modifies and what events those modifications produce.

The cascade map endpoint (`/workflow/rules/{id}/cascade-map`) uses this information to:
1. Identify which events an action will emit (e.g., `update_field` on "orders" emits `on_update` for "orders")
2. Find other rules that trigger on those events
3. Build a graph of potential downstream workflows

This helps operators understand the full impact of enabling a workflow rule before it runs.

## Condition Node Results

When workflows include `evaluate_condition` actions, the condition result flows through the system:

1. The condition handler evaluates field values and returns a `ConditionResult` with `BranchTaken` ("true" or "false")
2. The graph executor examines this result to determine which edges to follow (`true_branch` or `false_branch`)
3. The execution log records which branch was taken for debugging

See [Branching](branching.md) for details on how condition results control execution flow.

## Related Documentation

- [Architecture](architecture.md) - System overview and component details
- [Adding Domains](adding-domains.md) - Step-by-step guide for adding workflow events
- [Testing](testing.md) - Testing patterns for event infrastructure
- [Cascade Visualization](cascade-visualization.md) - Understanding downstream workflow effects
- [Branching](branching.md) - Graph-based execution with conditional paths

## Key Files

| File | Purpose |
|------|---------|
| `business/sdk/workflow/eventpublisher.go` | Event publishing |
| `business/sdk/workflow/delegatehandler.go` | Delegate bridge |
| `business/sdk/delegate/delegate.go` | Core delegate system |
| `business/domain/sales/ordersbus/event.go` | Example domain events |
| `api/cmd/services/ichor/build/all/all.go` | Registration |
