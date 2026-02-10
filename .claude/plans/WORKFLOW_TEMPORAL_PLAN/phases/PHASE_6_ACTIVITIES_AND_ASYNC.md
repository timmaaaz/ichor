# Phase 6: Activities & Async

**Category**: backend
**Status**: Pending
**Dependencies**: Phase 3 (Core Models & Context - COMPLETED), Phase 4 (Graph Executor - COMPLETED, provides `GraphExecutor`, `GetNextActions`, edge type logic)

---

## Overview

Implement the Temporal activity layer that bridges the workflow orchestration (Phase 5) with existing action handlers. Activities are the "side effect" layer in Temporal - they perform real work (send emails, allocate inventory, evaluate conditions) while the workflow function coordinates execution order. This phase creates three files: `activities.go` (synchronous action dispatch), `activities_async.go` (async completion pattern for RabbitMQ-backed actions), and `async_completer.go` (external completion API for async activities).

The key design constraint is **compatibility with existing `workflow.ActionHandler` interface** (`business/sdk/workflow/interfaces.go`). The existing handlers already implement `Execute(ctx, config, context) (any, error)` - the Temporal activity wrapper dispatches to these handlers unchanged. No modifications to existing handler code are needed.

## Goals

1. **Implement the synchronous activity dispatcher (`ExecuteActionActivity`)** that wraps existing `workflow.ActionHandler` implementations as Temporal activities, using `ActionRegistry` for handler lookup and `toResultMap` for result normalization
2. **Implement the async activity pattern (`ExecuteAsyncActionActivity`)** using Temporal's `activity.ErrResultPending` and task tokens for actions that publish to RabbitMQ and wait for external completion (e.g., approval workflows, long-running inventory operations)
3. **Implement `AsyncCompleter` for external activity completion** - the API that RabbitMQ consumers call to signal activity success/failure, closing the async loop with `client.CompleteActivity`

## Prerequisites

- Phase 3 complete: `ActionActivityInput`, `ActionActivityOutput`, `ActionExecutionContext` (implementation plan reference), `MergedContext`
- Phase 4 complete: Edge type constants, `GraphExecutor`
- Existing `workflow.ActionHandler` interface in `business/sdk/workflow/interfaces.go`
- Existing handler implementations in `business/sdk/workflow/workflowactions/` (all implement `Execute(ctx, config, execCtx) (any, error)`)
- Temporal SDK available in `go.mod`/`vendor` (from Phase 1)

---

## Go Package Structure

```
business/sdk/workflow/temporal/
    models.go              <- Phase 3 (COMPLETED)
    models_test.go         <- Phase 3 (COMPLETED)
    graph_executor.go      <- Phase 4 (COMPLETED)
    graph_executor_test.go <- Phase 4 (COMPLETED)
    workflow.go            <- Phase 5
    activities.go          <- THIS PHASE (Task 1)
    activities_async.go    <- THIS PHASE (Task 2)
    async_completer.go     <- THIS PHASE (Task 3)
    trigger.go             <- Phase 7
```

---

## Existing Handler Interface (DO NOT MODIFY)

The existing `workflow.ActionHandler` interface in `business/sdk/workflow/interfaces.go`:

```go
type ActionHandler interface {
    Execute(ctx context.Context, config json.RawMessage, context ActionExecutionContext) (any, error)
    Validate(config json.RawMessage) error
    GetType() string
    SupportsManualExecution() bool
    IsAsync() bool
    GetDescription() string
}
```

**Existing `ActionExecutionContext`** in `business/sdk/workflow/models.go`:

```go
type ActionExecutionContext struct {
    EntityID      uuid.UUID              `json:"entity_id,omitempty"`
    EntityName    string                 `json:"entity_name"`
    EventType     string                 `json:"event_type"`
    FieldChanges  map[string]FieldChange `json:"field_changes,omitempty"`
    RawData       map[string]interface{} `json:"raw_data,omitempty"`
    Timestamp     time.Time              `json:"timestamp"`
    UserID        uuid.UUID              `json:"user_id,omitempty"`
    RuleID        *uuid.UUID             `json:"rule_id,omitempty"`
    RuleName      string                 `json:"rule_name"`
    ExecutionID   uuid.UUID              `json:"execution_id,omitempty"`
    TriggerSource string                 `json:"trigger_source"`
}
```

**Existing `AsyncActionHandler`** in `business/sdk/workflow/interfaces.go`:

```go
type AsyncActionHandler interface {
    ActionHandler
    ProcessQueued(ctx context.Context, payload json.RawMessage, publisher *EventPublisher) error
}
```

**Existing handler implementations** (all in `business/sdk/workflow/workflowactions/`):

| Handler | Package | IsAsync | Signature |
|---------|---------|---------|-----------|
| `EvaluateConditionHandler` | `control/condition.go` | false | `Execute(ctx, config, execCtx) (interface{}, error)` |
| `UpdateFieldHandler` | `data/updatefield.go` | false | `Execute(ctx, config, execContext) (any, error)` |
| `SendEmailHandler` | `communication/email.go` | false | `Execute(ctx, config, context) (interface{}, error)` |
| `SendNotificationHandler` | `communication/notification.go` | false | `Execute(ctx, config, context) (interface{}, error)` |
| `CreateAlertHandler` | `communication/alert.go` | false | `Execute(ctx, config, execCtx) (interface{}, error)` |
| `SeekApprovalHandler` | `approval/seek.go` | true* | `Execute(ctx, config, context) (interface{}, error)` |
| `AllocateInventoryHandler` | `inventory/allocate.go` | true* | `Execute(ctx, config, execContext) (any, error)` |

*Note: `IsAsync()` returns true for approval and inventory handlers - these will use `ExecuteAsyncActionActivity` in Phase 9 wiring.

---

## Task Breakdown

### Task 1: Implement activities.go

**Status**: Pending

**Description**: Implement the synchronous activity dispatcher that wraps existing `workflow.ActionHandler` implementations as Temporal activities. This is the primary bridge between Temporal's activity execution model and our existing handler infrastructure. The `ActivityDependencies` struct holds all dependencies needed by activities (set during worker startup in Phase 9), and `ExecuteActionActivity` is registered as a Temporal activity function.

**Notes**:
- `ActivityDependencies` struct - holds `ActionRegistry` and future dependencies
- `SetActivityDependencies` - called during worker startup (Phase 9)
- `ExecuteActionActivity` - Temporal activity function, dispatches to registered handlers
- `toResultMap` - converts `any` handler return to `map[string]any` for context merging
- `buildExecContext` - converts `ActionActivityInput` + `WorkflowInput` metadata to existing `workflow.ActionExecutionContext`
- Must import `go.temporal.io/sdk/activity` for logging and info
- Must import `business/sdk/workflow` for `ActionHandler` and `ActionExecutionContext` types

**Files**:
- `business/sdk/workflow/temporal/activities.go`

**Implementation Guide**:

```go
package temporal

import (
    "context"
    "encoding/json"
    "fmt"

    "go.temporal.io/sdk/activity"

    "github.com/timmaaaz/ichor/business/sdk/workflow"
)

// =============================================================================
// Activity Dependencies
// =============================================================================

// ActivityDependencies holds all dependencies needed by Temporal activities.
// Set during worker startup (Phase 9) via SetActivityDependencies.
type ActivityDependencies struct {
    Registry *workflow.ActionRegistry
}

// activityDeps is the package-level dependencies reference.
// Set once during worker initialization, read by all activity executions.
var activityDeps *ActivityDependencies

// SetActivityDependencies initializes activity dependencies.
// Called once during worker startup in Phase 9.
// Must be called before any workflow execution.
func SetActivityDependencies(deps *ActivityDependencies) {
    activityDeps = deps
}

// =============================================================================
// Synchronous Activity
// =============================================================================

// ExecuteActionActivity dispatches to the appropriate handler based on action type.
// This wraps existing workflow.ActionHandler implementations as Temporal activities.
//
// Registered as a Temporal activity function on the worker (Phase 9).
// Called from workflow.go's executeSingleAction via workflow.ExecuteActivity.
func ExecuteActionActivity(ctx context.Context, input ActionActivityInput) (ActionActivityOutput, error) {
    logger := activity.GetLogger(ctx)
    logger.Info("Executing action activity",
        "action_id", input.ActionID,
        "action_name", input.ActionName,
        "action_type", input.ActionType,
    )

    if err := input.Validate(); err != nil {
        return ActionActivityOutput{}, fmt.Errorf("invalid input: %w", err)
    }

    if activityDeps == nil || activityDeps.Registry == nil {
        return ActionActivityOutput{}, fmt.Errorf("activity dependencies not initialized")
    }

    // Look up handler from existing registry
    handler, exists := activityDeps.Registry.Get(input.ActionType)
    if !exists {
        return ActionActivityOutput{
            ActionID:   input.ActionID,
            ActionName: input.ActionName,
            Success:    false,
        }, fmt.Errorf("no handler registered for action type: %s", input.ActionType)
    }

    // Build the existing ActionExecutionContext from our input
    execCtx := buildExecContext(input)

    // Execute the action using the existing handler
    result, err := handler.Execute(ctx, input.Config, execCtx)
    if err != nil {
        logger.Error("Action execution failed",
            "action_id", input.ActionID,
            "action_type", input.ActionType,
            "error", err,
        )
        return ActionActivityOutput{
            ActionID:   input.ActionID,
            ActionName: input.ActionName,
            Success:    false,
        }, err
    }

    // Convert result to map for context merging
    resultMap := toResultMap(result)

    logger.Info("Action execution succeeded",
        "action_id", input.ActionID,
        "action_name", input.ActionName,
    )

    return ActionActivityOutput{
        ActionID:   input.ActionID,
        ActionName: input.ActionName,
        Result:     resultMap,
        Success:    true,
    }, nil
}

// =============================================================================
// Helpers
// =============================================================================

// buildExecContext converts ActionActivityInput into the existing
// workflow.ActionExecutionContext used by all action handlers.
//
// The Context map from ActionActivityInput (which is MergedContext.Flattened)
// is passed as RawData, making all prior action results available for
// template variable resolution in downstream handlers.
func buildExecContext(input ActionActivityInput) workflow.ActionExecutionContext {
    return workflow.ActionExecutionContext{
        RawData:       input.Context,
        ExecutionID:   input.ActionID, // Per-action execution tracking
        TriggerSource: "automation",
    }
}

// toResultMap converts any result type to a map for context merging.
// Handles: nil, map[string]any, map[string]interface{}, and struct types.
// Struct types are marshaled to JSON then unmarshaled into a map.
func toResultMap(result any) map[string]any {
    if result == nil {
        return map[string]any{}
    }

    // If already a map, return directly
    if m, ok := result.(map[string]any); ok {
        return m
    }

    // Marshal/unmarshal for struct types
    data, err := json.Marshal(result)
    if err != nil {
        return map[string]any{"raw": fmt.Sprintf("%v", result)}
    }

    var m map[string]any
    if err := json.Unmarshal(data, &m); err != nil {
        return map[string]any{"raw": fmt.Sprintf("%v", result)}
    }

    return m
}
```

**Design Decisions**:

1. **Reuses existing `workflow.ActionRegistry`** - No duplicate registry. The same registry from `business/sdk/workflow/interfaces.go` is used. Phase 9 initializes it with handlers and passes it via `ActivityDependencies`.

2. **Package-level `activityDeps`** - Temporal activities are plain functions (not methods). Dependencies must be accessible at the package level. Set once during worker startup, immutable thereafter.

3. **`buildExecContext` bridge** - Maps between the Temporal `ActionActivityInput` and the existing `workflow.ActionExecutionContext`. The `Context` map (MergedContext.Flattened) becomes `RawData` for template resolution in existing handlers.

4. **`toResultMap` defensive conversion** - Handlers return `any` which could be a struct, map, or nil. The marshal/unmarshal roundtrip safely converts structs to maps for `MergedContext.MergeResult`.

---

### Task 2: Implement activities_async.go

**Status**: Pending

**Description**: Implement the async activity pattern for actions that publish work to RabbitMQ and wait for external completion. Uses Temporal's `activity.ErrResultPending` to signal that the activity will complete later via the CompleteActivity API. The task token (from `activity.GetInfo(ctx).TaskToken`) is passed to the async handler so it can be forwarded to the external system for later completion.

**Notes**:
- `ExecuteAsyncActionActivity` - Returns `activity.ErrResultPending` after starting async work
- Uses `activity.GetInfo(ctx).TaskToken` for later completion
- Async handlers implement `StartAsync(ctx, config, execCtx, taskToken) error`
- The existing `workflow.AsyncActionHandler.ProcessQueued` is for the RabbitMQ consumer side
- `ExecuteAsyncActionActivity` is the Temporal-side entry point (different from `ProcessQueued`)
- Existing `SeekApprovalHandler` and `AllocateInventoryHandler` will need a `StartAsync` adapter in Phase 9

**Files**:
- `business/sdk/workflow/temporal/activities_async.go`

**Implementation Guide**:

```go
package temporal

import (
    "context"
    "encoding/json"
    "fmt"

    "go.temporal.io/sdk/activity"

    "github.com/timmaaaz/ichor/business/sdk/workflow"
)

// =============================================================================
// Async Activity Handler Interface
// =============================================================================

// AsyncActivityHandler defines the interface for actions that complete
// asynchronously via Temporal's async completion pattern.
//
// Unlike workflow.AsyncActionHandler (which uses ProcessQueued for RabbitMQ),
// this interface is specifically for the Temporal activity side:
// 1. StartAsync publishes work with the task token
// 2. Activity returns ErrResultPending
// 3. External system calls AsyncCompleter.Complete with the task token
//
// Existing async handlers (SeekApprovalHandler, AllocateInventoryHandler)
// will be adapted to implement this interface in Phase 9.
type AsyncActivityHandler interface {
    workflow.ActionHandler

    // StartAsync initiates the async operation.
    // taskToken must be forwarded to the external system for later completion.
    // The handler should publish to RabbitMQ (or other queue) with the task token
    // as a correlation identifier.
    StartAsync(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext, taskToken []byte) error
}

// =============================================================================
// Async Activity Dependencies
// =============================================================================

// AsyncRegistry holds async action handlers separate from synchronous ones.
// Used by ExecuteAsyncActionActivity to dispatch async actions.
type AsyncRegistry struct {
    handlers map[string]AsyncActivityHandler
}

// NewAsyncRegistry creates a new async handler registry.
func NewAsyncRegistry() *AsyncRegistry {
    return &AsyncRegistry{
        handlers: make(map[string]AsyncActivityHandler),
    }
}

// Register adds an async handler for an action type.
func (r *AsyncRegistry) Register(actionType string, handler AsyncActivityHandler) {
    r.handlers[actionType] = handler
}

// Get retrieves an async handler by action type.
func (r *AsyncRegistry) Get(actionType string) (AsyncActivityHandler, bool) {
    h, ok := r.handlers[actionType]
    return h, ok
}

// =============================================================================
// Async Activity Dependencies Extension
// =============================================================================

// AsyncActivityDependencies extends ActivityDependencies with async registry.
// Set during worker startup (Phase 9) via SetAsyncActivityDependencies.
type AsyncActivityDependencies struct {
    AsyncRegistry *AsyncRegistry
}

var asyncDeps *AsyncActivityDependencies

// SetAsyncActivityDependencies initializes async activity dependencies.
func SetAsyncActivityDependencies(deps *AsyncActivityDependencies) {
    asyncDeps = deps
}

// =============================================================================
// Async Activity Function
// =============================================================================

// ExecuteAsyncActionActivity handles actions that complete asynchronously.
// The activity returns ErrResultPending; completion happens via AsyncCompleter.
//
// Flow:
// 1. Get task token from activity context
// 2. Call handler.StartAsync with task token
// 3. Return ErrResultPending (activity stays open)
// 4. External system processes work
// 5. External system calls AsyncCompleter.Complete(taskToken, result)
// 6. Temporal resumes workflow with the result
//
// Registered as a separate Temporal activity function on the worker (Phase 9).
// Called from workflow.go when action handler reports IsAsync() == true.
func ExecuteAsyncActionActivity(ctx context.Context, input ActionActivityInput) (ActionActivityOutput, error) {
    logger := activity.GetLogger(ctx)
    logger.Info("Starting async action activity",
        "action_id", input.ActionID,
        "action_name", input.ActionName,
        "action_type", input.ActionType,
    )

    if err := input.Validate(); err != nil {
        return ActionActivityOutput{}, fmt.Errorf("invalid input: %w", err)
    }

    if asyncDeps == nil || asyncDeps.AsyncRegistry == nil {
        return ActionActivityOutput{}, fmt.Errorf("async activity dependencies not initialized")
    }

    // Get task token for async completion
    activityInfo := activity.GetInfo(ctx)
    taskToken := activityInfo.TaskToken

    // Build execution context
    execCtx := buildExecContext(input)

    // Get async handler
    handler, exists := asyncDeps.AsyncRegistry.Get(input.ActionType)
    if !exists {
        return ActionActivityOutput{
            ActionID:   input.ActionID,
            ActionName: input.ActionName,
            Success:    false,
        }, fmt.Errorf("no async handler registered for action type: %s", input.ActionType)
    }

    // Start the async operation (e.g., publish to RabbitMQ with task token)
    if err := handler.StartAsync(ctx, input.Config, execCtx, taskToken); err != nil {
        logger.Error("Async action start failed",
            "action_id", input.ActionID,
            "action_type", input.ActionType,
            "error", err,
        )
        return ActionActivityOutput{}, fmt.Errorf("start async action %s: %w", input.ActionType, err)
    }

    logger.Info("Async action started, awaiting external completion",
        "action_id", input.ActionID,
        "action_name", input.ActionName,
    )

    // Return pending - activity does NOT complete yet.
    // Completion will happen when external system calls AsyncCompleter.Complete.
    return ActionActivityOutput{}, activity.ErrResultPending
}
```

**Design Decisions**:

1. **Separate `AsyncActivityHandler` interface** - Not the same as `workflow.AsyncActionHandler`. The existing interface has `ProcessQueued` (for RabbitMQ consumer side). The new interface has `StartAsync` (for Temporal activity side). Phase 9 adapters will bridge the two.

2. **Separate `AsyncRegistry`** - Keeps sync and async handlers distinct. The workflow layer (Phase 5) uses `handler.IsAsync()` to decide which activity function to call.

3. **`activity.ErrResultPending`** - Temporal's built-in signal that the activity will complete asynchronously. The worker doesn't wait; it frees the activity slot. The workflow stays blocked until `CompleteActivity` is called.

4. **Task token forwarding** - The task token is Temporal's correlation ID. It must be forwarded to the external system (via RabbitMQ message, database record, etc.) and returned when completing the activity.

---

### Task 3: Implement async_completer.go

**Status**: Pending

**Description**: Implement the `AsyncCompleter` that external systems (RabbitMQ consumers, webhook handlers) use to complete async activities. This is the other half of the async activity pattern: the activity starts in `ExecuteAsyncActionActivity`, and the completer finishes it from outside the Temporal worker.

**Notes**:
- `AsyncCompleter` wraps `client.Client.CompleteActivity`
- `Complete(ctx, taskToken, result)` - finishes with success
- `Fail(ctx, taskToken, err)` - finishes with error
- Used by RabbitMQ consumer in Phase 9 wiring
- Task token comes from the message published by `StartAsync`

**Files**:
- `business/sdk/workflow/temporal/async_completer.go`

**Implementation Guide**:

```go
package temporal

import (
    "context"
    "fmt"

    "go.temporal.io/sdk/client"
)

// =============================================================================
// Async Completer
// =============================================================================

// AsyncCompleter completes Temporal activities from external systems.
//
// Usage pattern:
// 1. Async activity starts (ExecuteAsyncActionActivity) and publishes work with task token
// 2. External system processes the work (e.g., RabbitMQ consumer, webhook)
// 3. External system calls Complete or Fail with the task token
// 4. Temporal resumes the workflow with the result
//
// Example: RabbitMQ consumer receives approval response:
//
//     completer := temporal.NewAsyncCompleter(temporalClient)
//     output := temporal.ActionActivityOutput{
//         ActionID:   approvalActionID,
//         ActionName: "seek_approval",
//         Result:     map[string]any{"approved": true, "approver": "admin@example.com"},
//         Success:    true,
//     }
//     completer.Complete(ctx, msg.TaskToken, output)
type AsyncCompleter struct {
    client client.Client
}

// NewAsyncCompleter creates a completer for async activities.
func NewAsyncCompleter(c client.Client) *AsyncCompleter {
    return &AsyncCompleter{client: c}
}

// Complete finishes an async activity with a successful result.
// taskToken is the correlation ID from activity.GetInfo(ctx).TaskToken,
// forwarded by the async handler to the external system.
func (c *AsyncCompleter) Complete(ctx context.Context, taskToken []byte, result ActionActivityOutput) error {
    if err := c.client.CompleteActivity(ctx, taskToken, result, nil); err != nil {
        return fmt.Errorf("complete async activity: %w", err)
    }
    return nil
}

// Fail finishes an async activity with an error.
// The workflow will see this as an activity failure and may retry
// depending on the retry policy configured in workflow.go.
func (c *AsyncCompleter) Fail(ctx context.Context, taskToken []byte, activityErr error) error {
    if err := c.client.CompleteActivity(ctx, taskToken, nil, activityErr); err != nil {
        return fmt.Errorf("fail async activity: %w", err)
    }
    return nil
}
```

---

### Task 4: Adapt Existing Action Handlers for Temporal (Assessment Only)

**Status**: Pending

**Description**: Review existing action handlers and document what adaptations are needed for Temporal integration. **No code changes in this phase** - this is an assessment that informs Phase 9 wiring. The key finding is that existing handlers already implement `workflow.ActionHandler` and can be used directly with `ExecuteActionActivity`. Async handlers need a `StartAsync` adapter wrapper.

**Notes**:
- All existing handlers implement `workflow.ActionHandler` - compatible with `ExecuteActionActivity` as-is
- `EvaluateConditionHandler` returns `ConditionResult` with `BranchTaken` field - `toResultMap` handles conversion
- `SeekApprovalHandler` (`IsAsync() == true`) needs a `StartAsync` adapter for Temporal async pattern
- `AllocateInventoryHandler` (`IsAsync() == true`) needs a `StartAsync` adapter for Temporal async pattern
- Sync handlers (email, notification, alert, update_field, condition) work unchanged through `ExecuteActionActivity`
- Phase 9 creates adapter wrappers, not this phase

**Files**: (none - assessment only)

**Handler Compatibility Matrix**:

| Handler | IsAsync | Temporal Activity | Adapter Needed? | Notes |
|---------|---------|-------------------|-----------------|-------|
| `EvaluateConditionHandler` | false | `ExecuteActionActivity` | No | Returns `ConditionResult{BranchTaken: "true"/"false"}` - `toResultMap` marshals to `{"branch_taken": "true"}` for edge type dispatch |
| `UpdateFieldHandler` | false | `ExecuteActionActivity` | No | Direct execution, result becomes context for downstream actions |
| `SendEmailHandler` | false | `ExecuteActionActivity` | No | Side-effect only, result is informational |
| `SendNotificationHandler` | false | `ExecuteActionActivity` | No | Side-effect only, result is informational |
| `CreateAlertHandler` | false | `ExecuteActionActivity` | No | Creates alert record, returns alert info |
| `SeekApprovalHandler` | true | `ExecuteAsyncActionActivity` | **Yes** | Needs `StartAsync` adapter wrapping `Execute` + task token forwarding via RabbitMQ |
| `AllocateInventoryHandler` | true | `ExecuteAsyncActionActivity` | **Yes** | Needs `StartAsync` adapter wrapping `Execute` + task token forwarding via RabbitMQ |

**Condition Handler Edge Dispatch**:

The `EvaluateConditionHandler` returns a `ConditionResult` struct:

```go
type ConditionResult struct {
    BranchTaken string `json:"branch_taken"` // "true" or "false"
}
```

After `toResultMap` conversion, this becomes `map[string]any{"branch_taken": "true"}`. The workflow layer (Phase 5) reads `result["branch_taken"]` to determine which edge type to follow:
- `"true"` → follow `true_branch` and `always` edges
- `"false"` → follow `false_branch` and `always` edges

This mapping happens in `GraphExecutor.GetNextActions` (Phase 4) via the `ActionActivityOutput.Result` map.

---

## Validation Criteria

- [ ] `go build ./business/sdk/workflow/temporal/...` passes
- [ ] `activities.go` compiles with correct imports (`go.temporal.io/sdk/activity`, `business/sdk/workflow`)
- [ ] `activities_async.go` compiles with `activity.ErrResultPending` usage
- [ ] `async_completer.go` compiles with `client.Client.CompleteActivity`
- [ ] `ExecuteActionActivity` accepts `ActionActivityInput` (from Phase 3 models.go) and returns `ActionActivityOutput`
- [ ] `ExecuteAsyncActionActivity` returns `activity.ErrResultPending` on success
- [ ] `toResultMap` handles nil, map, and struct types
- [ ] `buildExecContext` produces valid `workflow.ActionExecutionContext`
- [ ] `AsyncCompleter.Complete` and `Fail` call `client.CompleteActivity` correctly
- [ ] No import cycles between `temporal` package and `workflow` package
- [ ] Handler compatibility matrix documented for all 7 existing handlers

---

## Deliverables

- `business/sdk/workflow/temporal/activities.go`
- `business/sdk/workflow/temporal/activities_async.go`
- `business/sdk/workflow/temporal/async_completer.go`

---

## Gotchas & Tips

### Common Pitfalls

- **Import cycle risk**: `temporal` package imports `workflow` package for `ActionHandler` interface. Make sure `workflow` package does NOT import `temporal` package. This is one-directional: `temporal` → `workflow`, never `workflow` → `temporal`.

- **`any` vs `interface{}`**: Existing handlers use a mix of `any` and `interface{}` return types. Both are the same in Go 1.18+. Use `any` consistently in new code.

- **`toResultMap` JSON roundtrip loses numeric precision**: The marshal/unmarshal roundtrip converts ALL numeric types to `float64`. This means `int64` values (e.g., database IDs, timestamps) lose precision for values > 2^53. This is acceptable for context merging where values are primarily used for template resolution (string interpolation). Handlers that need precise large integers should return string representations.

- **Package-level deps are not thread-safe to SET**: `SetActivityDependencies` and `SetAsyncActivityDependencies` must be called once during initialization, before any workflow execution. They are safe to READ concurrently (immutable after init).

- **Task token lifecycle**: Task tokens are opaque byte slices. They must not be modified, only forwarded. They are valid until the activity times out. If the external system takes too long, the activity will time out and the task token becomes invalid.

- **`buildExecContext` extracts typed fields from Context map**: The implementation extracts EntityID, EntityName, EventType, UserID, and Timestamp from the Context map (which contains trigger data from Phase 7's buildTriggerData). This avoids adding new fields to ActionActivityInput while still populating the typed fields in ActionExecutionContext that handlers may rely on.

### Tips

- Start with Task 1 (`activities.go`) since it's the most critical and has no dependencies on Tasks 2-3
- Task 3 (`async_completer.go`) is the simplest file - can be written quickly
- Task 4 is assessment only - read existing handlers but don't modify them
- Test compilation with `go build ./business/sdk/workflow/temporal/...` after each file
- The `buildExecContext` function is a bridge point that will likely grow in Phase 9 when more metadata flows through
- `toResultMap` should handle the `ConditionResult` struct correctly via JSON marshal - verify this mentally or with a quick test

### Activity Idempotency

Temporal may retry activities on failure. Handlers MUST account for this:

- **Sync activities**: Retried up to 3 times (default retry policy). Handlers should be idempotent:
  - `UpdateFieldHandler`: Check if field already has target value before writing
  - `SendEmailHandler`: Use external deduplication (e.g., email provider's idempotency keys)
  - `EvaluateConditionHandler`: Pure function, naturally idempotent
  - `CreateAlertHandler`: Use execution ID for deduplication
- **Async activities**: `MaximumAttempts=1` (no retries) to prevent duplicate queue publishes.
  If `StartAsync` succeeds but the activity times out waiting for external completion,
  the workflow will see an activity failure. Recovery requires manual intervention or
  a compensating activity.
- **AsyncCompleter**: `CompleteActivity` with same token twice returns success (Temporal handles idempotency).
  Callers don't need to deduplicate completion calls.

### Determinism Notes

Activities are NOT subject to Temporal's determinism constraints. Unlike workflow code:
- Activities CAN use `time.Now()`, random numbers, network I/O
- Activities CAN use goroutines, channels, mutexes
- Activities CAN make database calls, HTTP requests, etc.
- Activities ARE retried on failure (retry policy configured in workflow.go activityOptions)

The determinism rules from Phase 5 (workflow.go) do NOT apply to activities.go.

---

## Testing Strategy

### Unit Tests (Deferred to Phase 12)

Activity tests require Temporal's test framework (`testsuite.WorkflowTestSuite` with mocked activities). These are deferred to Phase 12 (Edge Case & Limit Tests) where async completion tests are already planned.

### Compile-Time Validation (This Phase)

For this phase, validation is:
1. `go build` passes with no errors
2. All imports resolve correctly
3. Function signatures match Phase 3 model types
4. No import cycles

### Quick Smoke Tests (Phase 9)

When Phase 9 wires everything together, the first end-to-end test will exercise all three files.

---

## Out of Scope for Phase 6

The following items are explicitly deferred:

- **Activity registration with Temporal worker** - Phase 9 calls `worker.RegisterActivity(ExecuteActionActivity)` and `worker.RegisterActivity(ExecuteAsyncActionActivity)` during worker startup
- **`StartAsync` adapter wrappers** - Phase 9 creates adapters for `SendEmailHandler` and `AllocateInventoryHandler` to implement `AsyncActivityHandler.StartAsync`
- **Full ActionExecutionContext population** - `buildExecContext` now extracts typed fields from Context map. Phase 9 may pass additional metadata (RuleID, RuleName) if handlers need it
- **Unit tests** - Deferred to Phase 12 (Edge Case & Limit Tests) where async completion tests are planned

---

## Completion Checklist

- [x] `activities.go` implements `ExecuteActionActivity`
- [x] `activities_async.go` implements `ExecuteAsyncActionActivity`
- [x] `async_completer.go` implements `AsyncCompleter`
- [x] `ActivityDependencies` struct defined with `SetActivityDependencies`
- [x] `AsyncActivityDependencies` struct defined with `SetAsyncActivityDependencies`
- [x] `buildExecContext` extracts typed fields from Context map
- [x] Error messages include action name and ID for debugging
- [x] Log entries include workflow_id for cross-system traceability
- [x] `toResultMap` documents int64→float64 JSON roundtrip behavior
- [x] `selectActivityFunc` routes async/human actions to `ExecuteAsyncActionActivity`
- [x] `activityOptions` sets `MaximumAttempts=1` for async/human actions (idempotency)
- [x] Handler compatibility matrix documented for all 7 existing handlers
- [x] Activity idempotency guidance documented
- [x] All validation criteria pass

---

## Commands Reference

```bash
# Start this phase
/workflow-temporal-next

# Validate this phase
/workflow-temporal-validate 6

# Review plan before implementing
/workflow-temporal-plan-review 6

# Review code after implementing
/workflow-temporal-review 6
```
