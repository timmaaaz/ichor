# Migration from RabbitMQ

This document describes the migration from the RabbitMQ-based workflow engine to Temporal, what changed, and what stayed the same.

## Why We Migrated

The old engine was built on RabbitMQ with custom orchestration code (~10K lines). Temporal replaces all of this with a proven workflow orchestration platform:

| Concern | Old (RabbitMQ) | New (Temporal) |
|---------|----------------|----------------|
| **Durability** | Custom dead-letter queues | Built-in workflow state persistence |
| **Visibility** | Log parsing only | Web UI for inspecting running workflows |
| **Retries** | Custom circuit breakers | Configurable per-activity retry policies |
| **Parallel execution** | Not supported | Child workflows with convergence |
| **Crash recovery** | Message redelivery only | Full workflow replay from history |
| **Code complexity** | ~10K lines (engine, queue, executor, publisher) | ~3K lines (temporal package) |

## What Changed

### Removed Components (~10K lines deleted)

| Component | File | Purpose |
|-----------|------|---------|
| Engine | `engine.go` | Workflow execution singleton |
| ActionExecutor | `executor.go` | BFS graph traversal |
| DependencyResolver | `dependency.go` | Rule dependency resolution |
| QueueManager | `queue.go` | RabbitMQ queue management |
| NotificationQueue | `notificationQueue.go` | Notification delivery queue |
| EventPublisher | `eventpublisher.go` | Event queuing to RabbitMQ |
| DelegateHandler | `delegatehandler.go` | Old delegate → EventPublisher bridge |

All corresponding test files were also removed (~7 test files).

### Added Components (~3K lines)

| Component | File | Purpose |
|-----------|------|---------|
| Models | `temporal/models.go` | WorkflowInput, GraphDefinition, MergedContext |
| GraphExecutor | `temporal/graph_executor.go` | Deterministic graph traversal |
| Workflow | `temporal/workflow.go` | Temporal workflow implementation |
| Activities | `temporal/activities.go` | Activity wrappers for action handlers |
| AsyncActivities | `temporal/activities_async.go` | Async handler pattern |
| AsyncCompleter | `temporal/async_completer.go` | External activity completion |
| Trigger | `temporal/trigger.go` | Rule matching → Temporal dispatch |
| DelegateHandler | `temporal/delegatehandler.go` | New delegate → Temporal bridge |
| EdgeStore | `temporal/stores/edgedb/edgedb.go` | DB adapter for graph loading |

Also added: `workflow-worker` service (`api/cmd/services/workflow-worker/main.go`).

### Replaced Pipeline

**Old pipeline:**
```
Business Layer → delegate.Call()
  → DelegateHandler → EventPublisher → QueueManager → RabbitMQ
  → QueueManager consumer → Engine → ActionExecutor → handlers
```

**New pipeline:**
```
Business Layer → delegate.Call()
  → TemporalDelegateHandler → WorkflowTrigger → Temporal
  → workflow-worker → ExecuteGraphWorkflow → GraphExecutor → Activities → handlers
```

### Removed Types

From `models.go`:
- `ExecutionBatch`, `ExecutionPlan`, `WorkflowExecution`
- `BatchResult`, `RuleResult`, `ActionResult`
- `WorkflowConfig`, `WorkflowStats`

From `interfaces.go`:
- `AsyncActionHandler` interface (replaced by `temporal.AsyncActivityHandler`)
- `QueuedPayload` struct

### Modified Components

| Component | Change |
|-----------|--------|
| `allocate.go` | Removed `ProcessQueued`, `fireAllocationResultEvent`, `queueClient` field. `Execute` now calls `ProcessAllocation` directly |
| `register.go` | Removed `queueClient` param from `NewAllocateInventoryHandler` |
| `all.go` | Unified Temporal block replaces RabbitMQ + old Temporal blocks |

## What Didn't Change

These components are identical before and after migration:

| Component | Why Unchanged |
|-----------|---------------|
| **Database tables** | Same schema: `automation_rules`, `rule_actions`, `action_edges`, `action_templates`, etc. |
| **Action handlers** | All 7 still implement `ActionHandler` interface (Execute, Validate, GetType, etc.) |
| **Delegate pattern** | Business layer still calls `delegate.Call()` with same `ActionCreatedData` etc. |
| **Save API** | `workflowsaveapp` validation unchanged (edge requirements, graph validation) |
| **Template resolution** | Same `{{variable}}` syntax, same template processor |
| **TriggerProcessor** | Same rule matching logic, same condition operators |
| **Domain event.go files** | Same DomainName/EntityName constants, same ActionCreatedParms |
| **ActionRegistry** | Same Register/Get pattern |

## Configuration Changes

### New

| Setting | Description |
|---------|-------------|
| `ICHOR_TEMPORAL_HOSTPORT` | Temporal server address. Empty string = disabled (Temporal features off) |

### Unchanged

| Setting | Notes |
|---------|-------|
| RabbitMQ connection | Still used for alert WebSocket delivery (`alertws.AlertConsumer`) — separate concern |
| Database settings | Same database, same tables |

### Removed

| Setting | Notes |
|---------|-------|
| Engine singleton config | No longer exists |
| QueueManager config | No longer exists |
| Circuit breaker settings | Temporal handles retries natively |
| Queue type definitions | Single Temporal task queue replaces 6 RabbitMQ queues |

## Known Limitations

### Alert WebSocket Delivery

Alerts created by `CreateAlertHandler` are still written to the database. However, real-time WebSocket push delivery relied on RabbitMQ's alert queue. Post-migration:
- Alerts are created in the DB (works)
- REST API polling for alerts (works)
- Real-time WebSocket push (stopped — `alertws.AlertConsumer` has nil guard)

### Async Handler Adapters

`SendEmailHandler` and `AllocateInventoryHandler` need `StartAsync` adapters to fully support the Temporal async completion pattern. Currently, the worker registers an empty `AsyncRegistry`.

### Integration Tests

Most integration tests have been restored using Temporal test infrastructure. Three test files remain excluded with `//go:build ignore` because they deeply use the deleted `Engine.ExecuteWorkflow()` synchronous API:
- `actions_test.go` — action handler integration tests
- `errors_test.go` — error propagation tests
- `execution_test.go` — synchronous execution tests

The following were restored in Phase 15:
- `execution_seed_test.go`, `trigger_test.go`
- `formdataapi/workflow_test.go`, `ordersapi/workflow_test.go`

## Mapping Table

Quick reference for developers familiar with the old system:

| Old Concept | New Concept |
|-------------|-------------|
| `workflow.NewEngine()` | Not needed (Temporal manages execution) |
| `workflow.NewQueueManager()` | Not needed (Temporal task queue) |
| `workflow.NewEventPublisher()` | Not needed (direct Temporal dispatch) |
| `workflow.NewDelegateHandler()` | `temporal.NewDelegateHandler()` |
| `engine.ExecuteWorkflow()` | Temporal dispatches `ExecuteGraphWorkflow` |
| `engine.GetRegistry()` | Registry passed to `temporal.Activities` struct |
| `executor.ExecuteRuleActionsGraph()` | `GraphExecutor.GetStartActions()` + `GetNextActions()` |
| `qm.Start(ctx)` / `qm.Stop(ctx)` | `worker.Run()` / signal handling |
| `qm.GetMetrics()` | Temporal UI visibility |
| Dead letter queue | Temporal retry policy + failed workflow visibility |
| Circuit breakers | Temporal retry policy (per-activity) |
| `QueueTypeWorkflow`, `QueueTypeEmail`, etc. | Single `ichor-workflow` task queue |
| `AsyncActionHandler.ProcessQueued()` | `AsyncActivityHandler.StartAsync()` + `AsyncCompleter` |

## Related Documentation

- [Architecture](architecture.md) — Current system architecture
- [Temporal Integration](temporal.md) — Temporal design details
- [Worker Deployment](worker-deployment.md) — Worker service operations
