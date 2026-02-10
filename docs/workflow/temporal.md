# Temporal Integration

This document describes how Ichor uses [Temporal](https://temporal.io) for durable workflow execution.

## Why Temporal

Temporal provides:
- **Durability**: Workflow state survives process crashes, restarts, and infrastructure failures
- **Visibility**: Inspect running workflows, view history, debug failures via Temporal UI
- **Structured retries**: Configurable per-activity retry policies with exponential backoff
- **Parallel execution**: Native support for child workflows and branch convergence
- **Crash recovery**: Workflows resume from the exact point of failure

This replaces the previous custom engine built on RabbitMQ (see [Migration from RabbitMQ](migration-from-rabbitmq.md)).

## Architecture

The system is split across two services:

```
┌──────────────────────┐          ┌──────────────────────┐
│   ichor (API)         │          │  workflow-worker      │
│                      │          │                      │
│  DelegateHandler     │  task    │  ExecuteGraphWorkflow│
│  WorkflowTrigger ────┼──queue──▶│  GraphExecutor       │
│  TriggerProcessor    │          │  Activities          │
│  EdgeStore           │          │  ActionHandlers      │
└──────────────────────┘          └──────────────────────┘
                    │                      │
                    └──────┬───────────────┘
                           │
                    ┌──────▼──────┐
                    │  Temporal    │
                    │  Server     │
                    └─────────────┘
```

**ichor (trigger side)**: Captures entity events, evaluates rules, dispatches workflows to Temporal.

**workflow-worker (execution side)**: Picks up workflow tasks, traverses action graphs, executes action handlers.

**Temporal Server**: Manages workflow state, task queues, history, and retry scheduling.

## Task Queue

All workflows use a single task queue: `ichor-workflow`

```go
const TaskQueue = "ichor-workflow"
```

Multiple workflow-worker instances can poll the same queue. Temporal handles task routing and load balancing automatically.

## GraphExecutor Design

**Location**: `business/sdk/workflow/temporal/graph_executor.go`

The GraphExecutor interprets visual workflow graphs stored in the database. It provides deterministic graph traversal that is safe for Temporal replay.

### Determinism

Temporal replays workflow history to reconstruct state. This means the graph traversal must produce identical results on every replay.

**Guarantees:**
- All map iterations use sorted keys (SortOrder primary, UUID secondary)
- No `time.Now()`, `rand`, or bare goroutines in workflow code
- `sort.Slice` with stable secondary sort prevents non-deterministic ordering

### Edge Types

| Type | Behavior |
|------|----------|
| `start` | Entry point — `source_action_id` is null, connects to first action(s) |
| `sequence` | Always followed after source action completes |
| `true_branch` | Followed when source action's result has `branch_taken: "true_branch"` |
| `false_branch` | Followed when source action's result has `branch_taken: "false_branch"` |
| `always` | Always followed from a condition action (runs regardless of branch) |

### Convergence Detection

When multiple actions follow a single action (parallel fork), the GraphExecutor detects where branches reconverge:

1. BFS from each branch to find all reachable nodes
2. Intersect reachable sets to find common nodes
3. Select the closest common node (minimum depth) as convergence point
4. `uuid.Nil` convergence = fire-and-forget (no waiting)

### Parallel Execution Patterns

**With convergence** (fork → branches → merge):
```
    [condition]
     /       \
[branchA] [branchB]
     \       /
     [merge]
```
Uses child workflows + Selector to wait for all branches.

**Fire-and-forget** (fork → branches, no merge):
```
    [action]
     /       \
[branchA] [branchB]
```
Uses child workflows with `PARENT_CLOSE_POLICY_ABANDON`. Parent continues immediately.

## Activity Types

### Sync Activities

Executed by `ExecuteActionActivity`. These are safe to retry.

| Action Type | Handler | Description |
|-------------|---------|-------------|
| `evaluate_condition` | ConditionHandler | Evaluates field conditions for branching |
| `update_field` | UpdateFieldHandler | Updates database fields dynamically |
| `send_notification` | SendNotificationHandler | Multi-channel notifications |
| `create_alert` | CreateAlertHandler | Creates in-app alerts |
| `seek_approval` | SeekApprovalHandler | Initiates approval workflows |

**Retry policy**: MaximumAttempts=3, exponential backoff.

### Async Activities

Executed by `ExecuteAsyncActionActivity`. These queue work externally and return `ErrResultPending`.

| Action Type | Handler | Description |
|-------------|---------|-------------|
| `send_email` | SendEmailHandler | Sends email via external service |
| `allocate_inventory` | AllocateInventoryHandler | Reserves inventory |

**Retry policy**: MaximumAttempts=1 (prevent duplicate side effects).

### Human Activities

Also use `ExecuteAsyncActionActivity`. These wait for human interaction.

| Action Type | Description |
|-------------|-------------|
| `manager_approval` | Requires manager sign-off |
| `manual_review` | Requires manual review |
| `human_verification` | Requires human verification |
| `approval_request` | General approval request |

**Retry policy**: MaximumAttempts=1.

## Context Propagation

### MergedContext

Each workflow maintains a `MergedContext` that accumulates results from all executed actions:

```go
type MergedContext struct {
    TriggerData   map[string]any  // Initial event data (entity fields, metadata)
    ActionResults map[string]any  // Results keyed by action name
}
```

**Flow:**
1. `TriggerData` populated from TriggerEvent (entity ID, name, event type, raw data)
2. Each action's result merged into `ActionResults` by action name
3. Template variables resolved against the full context
4. `ActionExecutionContext` built from context for each handler

### Payload Size Management

Results are sanitized to prevent oversized Temporal payloads:

- `MaxResultValueSize` = 50KB per value
- Strings truncated with `[truncated]` suffix
- Binary data (`[]byte`) truncated similarly
- Objects (structs/maps) serialized to JSON, truncated if needed

## Continue-As-New

Temporal workflows have finite history limits. For long-running workflows:

- **Threshold**: `HistoryLengthThreshold` = 10,000 history events
- **Trigger**: Checked after each action execution
- **Preservation**: Full `MergedContext` saved as `ContinuationState` on new `WorkflowInput`
- **Resumption**: New workflow instance continues from where the previous left off

```go
func shouldContinueAsNew(historyLength int) bool {
    return historyLength > HistoryLengthThreshold
}
```

## Error Handling

### Per-Action Errors

- Action failures wrapped with action name: `"action_0: handler error"`
- Regular activities retry up to 3 times before failing the workflow
- Async/human activities fail immediately (MaximumAttempts=1)

### Per-Rule Fail-Open

When multiple rules match an event, each rule dispatches independently:
- Rule A failure doesn't prevent Rule B from dispatching
- Failures logged with rule ID for debugging

### Non-Blocking Dispatch

The TemporalDelegateHandler dispatches in a goroutine:
- Business layer operation always completes
- Temporal dispatch errors logged but never returned to caller

## Workflow ID Format

```
workflow-{ruleID}-{entityID}-{executionID}
```

- **ruleID**: The automation rule being executed
- **entityID**: The entity that triggered the event
- **executionID**: Unique per execution (UUID)

The deterministic prefix enables searching by rule or entity in Temporal UI.

## Versioning

```go
workflow.GetVersion(ctx, "graph-interpreter", DefaultVersion, 1)
```

Enables safe schema evolution:
- In-flight workflows continue with their original version
- New workflows use the latest version
- No breaking changes to running workflows

## Key Files

| File | Purpose |
|------|---------|
| `temporal/models.go` | WorkflowInput, GraphDefinition, MergedContext, ActionNode, ActionEdge |
| `temporal/graph_executor.go` | Deterministic graph traversal, convergence detection |
| `temporal/workflow.go` | Temporal workflow implementation, parallel execution |
| `temporal/activities.go` | Activity wrappers, action handler dispatch |
| `temporal/activities_async.go` | Async activity handler, AsyncRegistry |
| `temporal/async_completer.go` | AsyncCompleter for external completion |
| `temporal/trigger.go` | WorkflowTrigger, rule matching, Temporal dispatch |
| `temporal/delegatehandler.go` | TemporalDelegateHandler, delegate bridge |
| `temporal/stores/edgedb/edgedb.go` | Edge store DB adapter |

All files are in `business/sdk/workflow/temporal/`.

## Related Documentation

- [Architecture](architecture.md) — System overview and all components
- [Worker Deployment](worker-deployment.md) — Worker service operations
- [Migration from RabbitMQ](migration-from-rabbitmq.md) — What changed and why
- [Testing](testing.md) — Test patterns including Temporal-specific tests
- [Branching](branching.md) — Graph-based execution patterns
