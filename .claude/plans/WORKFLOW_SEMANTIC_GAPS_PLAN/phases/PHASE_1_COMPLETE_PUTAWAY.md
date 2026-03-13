# Phase 1: Putaway Completion Integration Tests

## REVISED SCOPE

The `complete_putaway` workflow action handler has been **removed from scope** after critical
review. Reason: completing a putaway task is reporting that a physical act occurred (warehouse
worker places product in bin). A workflow engine invoking this would fabricate the physical
confirmation. The delegate-reactive pattern is architecturally correct here — the worker
completes the task → delegate fires → workflow reacts.

`transition_status(to: completed)` would be even more dangerous: it flips the status without
executing the 3-way write, silently corrupting inventory records.

**This phase is now only the integration tests for the existing putaway completion HTTP endpoint.**
The 3-way transaction atomicity is completely untested regardless of the handler decision.

## Objective

Write integration tests that verify the 3-way atomic transaction in the putaway completion
endpoint: task status update + inventory item quantity increment + inventory transaction record
creation must all succeed or all roll back as a unit.

## Pre-Work: Read Before Implementing

1. Read `business/domain/inventory/putawaytaskbus/putawaytaskbus.go` — find the Complete() method signature
2. Read `app/domain/inventory/putawaytaskapp/putawaytaskapp.go` — find the complete() app method
3. Read `business/domain/inventory/putawaytaskbus/status.go` — understand the Status type
4. Read `api/cmd/services/ichor/tests/inventory/putawaytaskapi/` — check what tests exist
5. Read `business/sdk/workflow/workflowactions/inventory/receive.go` — use as handler template

## Step 1: Business Layer Investigation

Confirm:
- Does `putawaytaskbus.Business` have a `Complete(ctx, task)` method, or is completion done via `Update()` with a status change?
- What bus dependencies does completion require? (putawaytaskbus, inventoryitembus, inventorytransactionbus)
- Does the 3-way transaction live in the bus layer or app layer?

## Step 2: Create Handler File

**File**: `business/sdk/workflow/workflowactions/inventory/complete_putaway.go`

```
Package: inventory

Handler struct: CompletePutawayHandler
  - log *logger.Logger
  - db *sqlx.DB
  - putawayTaskBus *putawaytaskbus.Business
  - inventoryItemBus *inventoryitembus.Business
  - transactionBus *inventorytransactionbus.Business

Config struct: CompletePutawayConfig
  - TaskID string `json:"task_id"` (required)
  - Notes string  `json:"notes,omitempty"`

Methods:
  - GetType() string → "complete_putaway"
  - IsAsync() bool → false
  - SupportsManualExecution() bool → true
  - GetDescription() string → "Complete a putaway task: update task status, increase inventory quantity, and record inbound transaction"
  - Validate(config json.RawMessage) error — validate task_id is non-empty UUID
  - GetOutputPorts() []workflow.OutputPort → completed, task_not_found, already_completed, failure
  - GetEntityModifications() → inventory.put_away_tasks (on_update), inventory.inventory_items (on_update), inventory.inventory_transactions (on_create)
  - Execute(ctx, config, execCtx) — call the bus Complete() or replicate the 3-way txn
```

**Execute logic**:
1. Parse config, validate task_id UUID
2. QueryByID the putaway task — route to `task_not_found` if not found
3. If task.Status == completed, route to `already_completed` (idempotent)
4. Call the appropriate business method (Complete() or equivalent) in a transaction
5. Commit → route to `completed` with task_id, inventory_item_id, quantity_received, transaction_id
6. Any error → route to `failure`

## Step 3: Register Handler

**File**: `business/sdk/workflow/workflowactions/register.go`

Add to `BusDependencies`:
```go
PutawayTask *putawaytaskbus.Business
```

Add to `RegisterGranularInventoryActions()`:
```go
registry.Register(inventory.NewCompletePutawayHandler(
    config.Log,
    config.DB,
    config.Buses.PutawayTask,
    config.Buses.InventoryItem,
    config.Buses.InventoryTransaction,
))
```

## Step 4: Wire in all.go

**File**: `api/cmd/services/ichor/build/all/all.go`

Add `putawaytaskBus` to the `ActionConfig.Buses.PutawayTask` field. Verify it's already
instantiated in all.go (it should be for the existing putaway API routes). Just add the
field reference — no new bus construction.

## Step 5: Integration Tests for Putaway Completion

**Directory**: `api/cmd/services/ichor/tests/inventory/putawaytaskapi/`

Check if `complete_test.go` exists. If not, create it using the `apitest.Table` pattern.

Test cases required:
1. **Happy path**: PUT /putaway-tasks/{id}/complete → 200, task status = completed, inventory quantity incremented, transaction record created
2. **Already completed**: PUT /putaway-tasks/{id}/complete on a completed task → appropriate response (idempotent or 409)
3. **Not found**: PUT /putaway-tasks/{id}/complete with unknown ID → 404
4. **Wrong status**: PUT /putaway-tasks/{id}/complete on cancelled task → 422

Also verify the 3-way atomicity: if inventory update fails, task status must NOT be updated
(requires inspecting the transaction rollback path).

## Step 6: Unit Test for Handler

**File**: `business/sdk/workflow/workflowactions/inventory/complete_putaway_test.go`

Test:
- Valid config validates
- Invalid task_id UUID fails validation
- Execute with valid task → `completed` output port
- Execute with already-completed task → `already_completed` output port
- Execute with not-found task → `task_not_found` output port

## Verification

```bash
go build ./business/sdk/workflow/workflowactions/...
go build ./api/cmd/services/ichor/...
go test ./business/sdk/workflow/workflowactions/inventory/...
go test ./api/cmd/services/ichor/tests/inventory/putawaytaskapi/...
```

## Definition of Done

- [ ] Handler file created and compiles
- [ ] Registered in RegisterGranularInventoryActions
- [ ] BusDependencies.PutawayTask field added
- [ ] Wired in all.go
- [ ] Integration tests for putaway completion endpoint
- [ ] Unit tests for handler
- [ ] `go build` passes on all affected packages
- [ ] `GET /v1/workflow/action-types` returns `complete_putaway` with description and output ports
