# Testing

Testing patterns and examples for the workflow system.

## Overview

Workflow tests use a mix of approaches:
- **Temporal SDK test suite** for workflow and activity unit tests (no external services)
- **Real PostgreSQL** via testcontainers (`dbtest.NewDatabase`) for database integration tests
- **Real Temporal server** via testcontainers (`foundation/temporal`) for replay and integration tests
- **Real action handlers** for end-to-end verification

## Test Categories

| Category | Location | Purpose |
|----------|----------|---------|
| Temporal Models | `business/sdk/workflow/temporal/models_test.go` | MergedContext, sanitizeResult |
| Graph Executor | `business/sdk/workflow/temporal/graph_executor_test.go` | Graph traversal, edge types |
| Graph Determinism | `business/sdk/workflow/temporal/graph_executor_determinism_test.go` | 1000-iteration stress tests |
| Graph Edge Types | `business/sdk/workflow/temporal/graph_executor_edges_test.go` | All 5 edge type combinations |
| Graph Convergence | `business/sdk/workflow/temporal/graph_executor_convergence_test.go` | Convergence detection, cycles |
| Workflow Tests | `business/sdk/workflow/temporal/workflow_test.go` | Sequential, branching, validation |
| Parallel Tests | `business/sdk/workflow/temporal/workflow_parallel_test.go` | Convergence, fire-and-forget |
| Replay Tests | `business/sdk/workflow/temporal/workflow_replay_test.go` | Determinism via history replay |
| Continue-As-New | `business/sdk/workflow/temporal/workflow_continueasnew_test.go` | Threshold, state preservation |
| Payload Limits | `business/sdk/workflow/temporal/workflow_payload_test.go` | 50KB truncation boundaries |
| Async Activities | `business/sdk/workflow/temporal/activities_async_test.go` | Async handler, completer |
| Error Handling | `business/sdk/workflow/temporal/workflow_errors_test.go` | Failures, retries, isolation |
| Trigger Tests | `business/sdk/workflow/temporal/trigger_test.go` | Rule matching, Temporal dispatch |
| Edge Store | `business/sdk/workflow/temporal/stores/edgedb/edgedb_test.go` | DB adapter integration |
| Condition Handler | `business/sdk/workflow/workflowactions/control/condition_test.go` | Condition evaluation |
| Action Handlers | `business/sdk/workflow/workflowactions/*/` | Individual handler testing |
| Alert API Tests | `api/cmd/services/ichor/tests/workflow/alertapi/` | Alert HTTP endpoints |
| Edge API Tests | `api/cmd/services/ichor/tests/workflow/edgeapi/` | Edge CRUD endpoints |
| Cascade API Tests | `api/cmd/services/ichor/tests/workflow/ruleapi/cascade_test.go` | Cascade visualization |

**Total**: 155+ tests in the Temporal package alone.

## Temporal Test Patterns

### SDK TestWorkflowEnvironment (Unit Tests)

For testing workflows without a real Temporal server:

```go
func TestWorkflow(t *testing.T) {
    suite := testsuite.WorkflowTestSuite{}
    env := suite.NewTestWorkflowEnvironment()

    // Register activities
    handler := &testActionHandler{result: map[string]any{"status": "ok"}}
    registry := workflowactions.NewActionRegistry()
    registry.Register(handler)
    activities := &temporal.Activities{
        Registry:      registry,
        AsyncRegistry: temporal.NewAsyncRegistry(),
    }
    env.RegisterActivity(activities)

    // Execute workflow
    input := temporal.WorkflowInput{
        RuleID: uuid.New(),
        Graph:  buildTestGraph(),
        Context: temporal.NewMergedContext(map[string]any{
            "entity_id": uuid.New().String(),
        }),
    }
    env.ExecuteWorkflow(temporal.ExecuteGraphWorkflow, input)

    // Assert
    require.True(t, env.IsWorkflowCompleted())
    require.NoError(t, env.GetWorkflowError())
}
```

### SDK TestActivityEnvironment (Activity Tests)

For testing activities in isolation:

```go
func TestAsyncActivity(t *testing.T) {
    suite := testsuite.WorkflowTestSuite{}
    env := suite.NewTestActivityEnvironment()

    handler := &mockAsyncHandler{}
    asyncRegistry := temporal.NewAsyncRegistry()
    asyncRegistry.Register("test_action", handler)
    activities := &temporal.Activities{
        Registry:      workflowactions.NewActionRegistry(),
        AsyncRegistry: asyncRegistry,
    }
    env.RegisterActivity(activities)

    input := temporal.ActionActivityInput{
        ActionName: "test",
        ActionType: "test_action",
        Config:     json.RawMessage(`{}`),
    }
    _, err := env.ExecuteActivity(activities.ExecuteAsyncActionActivity, input)
    // ErrResultPending is expected for async activities
    require.Error(t, err)
    require.Contains(t, err.Error(), "CompleteActivity")
}
```

### Real Temporal Container (Integration Tests)

For replay and full integration tests:

```go
func TestWorkflowReplay(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Get shared test container
    container := foundationtemporal.GetTestContainer(t)
    c, _ := client.Dial(client.Options{HostPort: container.HostPort})
    defer c.Close()

    // Create worker with unique task queue
    taskQueue := fmt.Sprintf("test-%s", uuid.New().String()[:8])
    w := worker.New(c, taskQueue, worker.Options{})
    w.RegisterWorkflow(temporal.ExecuteGraphWorkflow)
    w.RegisterActivity(&temporal.Activities{Registry: registry})
    go w.Run(worker.InterruptCh())

    // Execute workflow
    run, _ := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
        TaskQueue: taskQueue,
    }, temporal.ExecuteGraphWorkflow, input)
    run.Get(ctx, nil)

    // Fetch history and replay
    history := fetchHistory(ctx, c, run.GetID(), run.GetRunID())
    replayer := worker.NewWorkflowReplayer()
    replayer.RegisterWorkflow(temporal.ExecuteGraphWorkflow)
    err := replayer.ReplayWorkflowHistory(nil, history)
    require.NoError(t, err) // No non-determinism errors
}
```

### Database Integration (Edge Store Tests)

For testing the edge store adapter:

```go
func TestEdgeStore(t *testing.T) {
    db := dbtest.NewDatabase(t, "Test_EdgeStore")
    ctx := context.Background()

    // Seed workflow data
    adminUserID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")
    workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db.DB))
    sd := workflow.TestSeedFullWorkflow(ctx, adminUserID, workflowBus)

    store := edgedb.NewStore(log, db.DB)

    // Query actions for a rule
    actions, err := store.QueryActionsByRule(ctx, sd.Rules[0].ID)
    require.NoError(t, err)
    require.NotEmpty(t, actions)
}
```

## Action Handler Tests

### Alert Handler Test

**File**: `business/sdk/workflow/workflowactions/communication/alert_test.go`

```go
func TestCreateAlertHandler_Validate(t *testing.T) {
    handler := NewCreateAlertHandler(log, db, alertBus)

    tests := []struct {
        name    string
        config  string
        wantErr bool
    }{
        {
            name: "valid config",
            config: `{
                "message": "Test alert",
                "recipients": {"users": ["5cf37266-3473-4006-984f-9325122678b7"]}
            }`,
            wantErr: false,
        },
        {
            name:    "missing message",
            config:  `{"recipients": {"users": ["uuid"]}}`,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := handler.Validate(json.RawMessage(tt.config))
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Condition Handler Tests

**File**: `business/sdk/workflow/workflowactions/control/condition_test.go`

Pure unit tests for the `evaluate_condition` action handler. No external services required.

### What These Tests Cover

- Validation of condition configurations
- All 10 operators: `equals`, `not_equals`, `greater_than`, `less_than`, `contains`, `in`, `is_null`, `is_not_null`, `changed_from`, `changed_to`
- Logic combinations (`AND`/`OR`)
- Branch result generation (`true_branch`/`false_branch`)
- Edge cases (nil data, missing fields, type mismatches, json.Number handling)

## Edge API Tests

**Directory**: `api/cmd/services/ichor/tests/workflow/edgeapi/`

HTTP endpoint tests for edge CRUD operations.

### Test Files

| File | Purpose |
|------|---------|
| `edge_test.go` | Main test entry point |
| `create_test.go` | Create edge tests |
| `query_test.go` | Query edges tests |
| `delete_test.go` | Delete edge tests |
| `seed_test.go` | Test data seeding |

## Cascade API Tests

**Files**:
- `api/cmd/services/ichor/tests/workflow/ruleapi/cascade_test.go`
- `api/cmd/services/ichor/tests/workflow/ruleapi/cascade_seed_test.go`

Tests the cascade visualization endpoint that shows downstream workflows.

## Running Tests

### Run All Temporal Package Tests

```bash
go test -v ./business/sdk/workflow/temporal/...
```

### Run Temporal Tests (Skip Integration)

```bash
go test -short ./business/sdk/workflow/temporal/...
```

### Run All Workflow Tests

```bash
go test -v ./business/sdk/workflow/...
```

### Run with Race Detector

```bash
go test -race ./business/sdk/workflow/temporal/...
```

### Run Determinism Stress Tests

```bash
go test -v ./business/sdk/workflow/temporal/... -run Determinism -count=10
```

### Run Edge Store Integration Tests

```bash
go test -v ./business/sdk/workflow/temporal/stores/edgedb/...
```

### Run Alert API Tests

```bash
go test -v ./api/cmd/services/ichor/tests/workflow/alertapi/...
```

### Run Edge API Tests

```bash
go test -v ./api/cmd/services/ichor/tests/workflow/edgeapi/...
```

### Run Condition Handler Tests

```bash
go test -v ./business/sdk/workflow/workflowactions/control/...
```

### Run Cascade Tests

```bash
go test -v ./api/cmd/services/ichor/tests/workflow/ruleapi/... -run Cascade
```

## Manual Testing

### Start Services

```bash
make dev-up
```

### Check Temporal UI

Access the Temporal UI to inspect running workflows:

```bash
make temporal-ui
# Opens port-forward to http://localhost:8280
```

### Check Worker Logs

```bash
make dev-logs-workflow-worker
```

### Create Order via FormData

```bash
curl -X POST /v1/formdata/{form_id}/upsert \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "operations": {"sales.orders": {"operation": "create", "order": 1}},
    "data": {"sales.orders": {"number": "ORD-001"}}
  }'
```

### Check Service Logs

```bash
make dev-logs | grep "workflow"
```

### Query Executions

```bash
make pgcli
SELECT * FROM workflow.automation_executions ORDER BY executed_at DESC LIMIT 5;
```

## Test Data Seeding

### Seed Workflow Infrastructure

```go
adminUserID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")
workflow.TestSeedFullWorkflow(ctx, adminUserID, workflowBus)
```

This creates:
- Standard trigger types (on_create, on_update, on_delete, scheduled)
- Standard entity types (table, view)
- Common entities (orders, customers, products, etc.)
- 5 rules, 3 templates, 10 actions, start + sequence edges

### Create Test Rule

```go
rule, _ := workflowBus.CreateRule(ctx, workflow.NewAutomationRule{
    Name:          "Test Rule",
    EntityID:      entityID,
    EntityTypeID:  entityTypeID,
    TriggerTypeID: triggerTypeID,
    IsActive:      true,
    CreatedBy:     adminUserID,
})
```

## Troubleshooting Tests

### Temporal Container Not Starting

1. Ensure Docker is running
2. Check `temporalio/temporal:latest` image is available: `docker pull temporalio/temporal:latest`
3. Increase container startup timeout if on slow machine

### Database Errors

1. Ensure migrations are current
2. Check foreign key constraints
3. Verify test data seeding completed

### Determinism Failures

If replay tests fail with non-determinism errors:
1. Check for `time.Now()`, `rand`, or bare goroutines in workflow code
2. Ensure map iteration uses sorted keys
3. Verify no new side effects added to workflow functions

### Integration Tests Skipped

Integration tests (replay, real Temporal) skip when `-short` flag is used:
```bash
# Run with integration tests
go test -v ./business/sdk/workflow/temporal/...

# Skip integration tests
go test -short ./business/sdk/workflow/temporal/...
```

## Related Documentation

- [Architecture](architecture.md) — System overview and component details
- [Temporal Integration](temporal.md) — Temporal architecture and design
- [Branching](branching.md) — Graph-based execution and conditional workflows
- [Event Infrastructure](event-infrastructure.md) — Delegate pattern and workflow dispatch
- [Actions Overview](actions/overview.md) — Action handler testing
- [Evaluate Condition](actions/evaluate-condition.md) — Condition action handler
