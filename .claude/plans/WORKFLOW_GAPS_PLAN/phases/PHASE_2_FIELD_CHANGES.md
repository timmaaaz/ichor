# Phase 2: Fix FieldChanges in DelegateHandler

**Category**: Backend
**Status**: Pending
**Dependencies**: None
**Effort**: Medium

---

## Overview

The `TriggerEvent` struct has a `FieldChanges map[string]FieldChange` field that captures old and new values for fields that changed during an update event. This is used by:

1. **`evaluate_condition` handler** — supports `changed_from` and `changed_to` operators that compare against `FieldChanges` entries
2. **`trigger.go` TriggerProcessor** — matches rules that have conditions using those same operators

The problem: the `DelegateHandler.handleEvent()` function in `business/sdk/workflow/temporal/delegatehandler.go` builds a `TriggerEvent` but **never sets `FieldChanges`**. It only sets `RawData` (the post-update snapshot). As a result, any rule condition using `changed_from` or `changed_to` never fires.

---

## Goals

1. Extend `DelegateEventParams` to carry the entity state before the update
2. Update domain bus `Update` methods to emit the before-state
3. Update `DelegateHandler` to compute and populate `FieldChanges` from the diff

---

## Architecture Decision

**Option A: Compute diff at delegate handler** — Domain buses emit `BeforeEntity` in the delegate params, handler diffs before/after. Simpler for bus authors.

**Option B: Compute diff in domain buses** — Each bus computes `FieldChanges` itself and passes a pre-built map. More precise (can exclude computed fields), but requires more work per bus.

**Decision: Use Option A.** The delegate handler already has `extractEntityData()` which marshals entities to `map[string]any`. We can diff two such maps efficiently. Bus authors just need to pass the before-entity alongside the after-entity.

---

## Task Breakdown

### Task 1: Extend DelegateEventParams

**File**: `business/sdk/workflow/event.go`

The `DelegateEventParams` struct currently has `EntityID`, `UserID`, and `Entity`. Add `BeforeEntity`:

```go
// DelegateEventParams holds the parameters passed through the delegate event system.
type DelegateEventParams struct {
    EntityID     uuid.UUID `json:"entity_id"`
    UserID       uuid.UUID `json:"user_id"`
    Entity       any       `json:"entity"`       // Post-update entity state
    BeforeEntity any       `json:"before_entity"` // Pre-update entity state (for on_update events only)
}
```

Update the `ActionUpdatedData` helper to accept a `before` parameter:

```go
// ActionUpdatedData creates the delegate params for an entity update event.
// before is the entity state BEFORE the update (used for FieldChanges diff).
func ActionUpdatedData(before, after any, entityID, userID uuid.UUID) DelegateEventParams {
    return DelegateEventParams{
        EntityID:     entityID,
        UserID:       userID,
        Entity:       after,
        BeforeEntity: before,
    }
}
```

Note: `ActionCreatedData` and `ActionDeletedData` do not need `BeforeEntity`.

### Task 2: Update High-Value Domain Buses

For each domain bus `Update` method, load the current state before updating and pass it to `ActionUpdatedData`.

**Pattern** (using `ordersbus` as example):

```go
// Current code (in ordersbus.go):
func (b *Business) Update(ctx context.Context, ord Order, uord UpdateOrder) (Order, error) {
    // ... update logic ...
    b.delegate.Call(ctx, workflow.ActionUpdatedData(order, ord.ID, claims.UserID))
    return order, nil
}

// Updated code:
func (b *Business) Update(ctx context.Context, ord Order, uord UpdateOrder) (Order, error) {
    // Capture before-state for FieldChanges diff
    before := ord  // ord is passed in — already the current state

    // ... update logic to produce updated order ...
    b.delegate.Call(ctx, workflow.ActionUpdatedData(before, order, ord.ID, claims.UserID))
    return order, nil
}
```

**Key insight**: Most `Update` methods already receive the current entity as a parameter (the `ord Order` argument). Use that as `before`.

Target buses for initial implementation (highest workflow value):
- `business/domain/sales/ordersbus/ordersbus.go`
- `business/domain/inventory/inventoryitembus/inventoryitembus.go`
- `business/domain/procurement/purchaseorderbus/purchaseorderbus.go`
- `business/domain/procurement/purchaseorderstatusbus/purchaseorderstatusbus.go`

Remaining buses can be updated in a follow-up or all at once — check each bus's `Update` signature.

### Task 3: Update DelegateHandler to Compute FieldChanges

**File**: `business/sdk/workflow/temporal/delegatehandler.go`

In `handleEvent()`, after extracting `entityID` and `rawData` from `params.Entity`, also extract data from `params.BeforeEntity` and compute the diff:

```go
func (h *DelegateHandler) handleEvent(ctx context.Context, eventType, entityName string, data delegate.Data) error {
    var params workflow.DelegateEventParams
    if err := json.Unmarshal(data.RawParams, &params); err != nil {
        // ...
        return nil
    }

    event := workflow.TriggerEvent{
        EventType:  eventType,
        EntityName: entityName,
        EntityID:   params.EntityID,
        Timestamp:  time.Now().UTC(),
        UserID:     params.UserID,
    }

    // Extract post-update entity data
    if params.Entity != nil {
        entityID, rawData, err := extractEntityData(params.Entity)
        if err == nil {
            event.RawData = rawData
            if event.EntityID == uuid.Nil {
                event.EntityID = entityID
            }
        }
    }

    // Compute FieldChanges for on_update events
    if eventType == "on_update" && params.BeforeEntity != nil && event.RawData != nil {
        _, beforeData, err := extractEntityData(params.BeforeEntity)
        if err == nil {
            event.FieldChanges = computeFieldChanges(beforeData, event.RawData)
        }
    }

    go func() {
        if err := h.trigger.OnEntityEvent(context.Background(), event); err != nil {
            // ...
        }
    }()

    return nil
}

// computeFieldChanges builds a FieldChange map by diffing before and after snapshots.
func computeFieldChanges(before, after map[string]any) map[string]workflow.FieldChange {
    changes := make(map[string]workflow.FieldChange)
    for key, afterVal := range after {
        beforeVal := before[key]
        // Use fmt.Sprintf for consistent comparison (handles nested types)
        if fmt.Sprintf("%v", beforeVal) != fmt.Sprintf("%v", afterVal) {
            changes[key] = workflow.FieldChange{
                OldValue: beforeVal,
                NewValue: afterVal,
            }
        }
    }
    return changes
}
```

Check the `FieldChange` struct definition in `business/sdk/workflow/models.go` to confirm field names (`OldValue`/`NewValue` or `Old`/`New`).

### Task 4: Add Integration Test

**File**: `api/cmd/services/ichor/tests/workflow/workflowsaveapi/trigger_test.go`

Add a test case that:
1. Creates a workflow rule with trigger `on_update` for `orders` entity
2. Adds a condition using `changed_from` operator (e.g., status changed from `pending`)
3. Calls the orders update endpoint to change the status
4. Verifies the workflow was triggered (check `workflow.automation_executions` table)

---

## Validation

```bash
go build ./...

# Check FieldChange struct fields
grep -A 5 "type FieldChange struct" business/sdk/workflow/models.go

# Check ActionUpdatedData signature
grep -A 5 "func ActionUpdatedData" business/sdk/workflow/event.go
```

---

## Gotchas

- **`BeforeEntity` is nil for `on_create` and `on_delete`** — the `computeFieldChanges` call is guarded by `eventType == "on_update"`. Don't break the on_create/on_delete path.
- **JSON marshaling may change values** — `extractEntityData` uses JSON marshal/unmarshal which converts `time.Time` to string and `uuid.UUID` to string. The comparison should still work for equality since both before and after go through the same conversion.
- **Float64 precision** — JSON numbers become `float64` in Go. `100` and `100.0` will compare equal after JSON round-trip. This is acceptable.
- **Check `ActionUpdatedData` call sites** — if some buses call it with only 3 args (before this change), they'll fail to compile. Search for `ActionUpdatedData` across the codebase.
- **`BeforeEntity` in delegate.Data.RawParams** — `DelegateEventParams` is JSON-marshaled before being passed through the delegate chain. The `BeforeEntity any` field will be JSON-serialized as the entity struct. This is fine since `extractEntityData` already handles `any` → JSON → `map[string]any`.
