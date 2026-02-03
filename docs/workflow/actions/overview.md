# Action Handlers Overview

Action handlers execute specific behaviors when automation rules fire. Each handler type has its own configuration schema and execution logic.

## Available Actions

| Action Type | Description | Documentation |
|-------------|-------------|---------------|
| `create_alert` | Creates in-app alerts | [create-alert.md](create-alert.md) |
| `update_field` | Updates database fields | [update-field.md](update-field.md) |
| `send_email` | Sends email notifications | [send-email.md](send-email.md) |
| `send_notification` | Multi-channel notifications | [send-notification.md](send-notification.md) |
| `seek_approval` | Initiates approval workflows | [seek-approval.md](seek-approval.md) |
| `allocate_inventory` | Reserves/allocates inventory | [allocate-inventory.md](allocate-inventory.md) |
| `evaluate_condition` | Evaluates conditions for branching | [evaluate-condition.md](evaluate-condition.md) |

## ActionHandler Interface

All action handlers implement this interface:

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

**Source**: `business/sdk/workflow/interfaces.go:38-60`

### Methods

| Method | Purpose |
|--------|---------|
| `GetType()` | Returns the action type identifier (e.g., `"create_alert"`) |
| `Validate()` | Validates configuration before rule creation |
| `Execute()` | Executes the action with given config and context |
| `SupportsManualExecution()` | Returns true if action can be triggered via API |
| `IsAsync()` | Returns true if action queues work for async processing |
| `GetDescription()` | Returns human-readable description for discovery APIs |

### Return Type

The `Execute()` method returns `(any, error)` for maximum flexibility - each handler returns its own result type. Common patterns:

- `map[string]interface{}` - Simple key-value results
- Custom structs - Type-safe results (e.g., `QueuedAllocationResponse`)
- `ConditionResult` - For `evaluate_condition` actions (includes `BranchTaken` field)

### Manual Execution Support

Some actions can be executed manually via API:

| Action Type | Supports Manual | Reason |
|-------------|-----------------|--------|
| `create_alert` | Yes | Useful for testing, manual notifications |
| `send_email` | Yes | Manual email sends |
| `send_notification` | Yes | Manual notifications |
| `allocate_inventory` | Yes | Manual inventory operations |
| `update_field` | **No** | Use entity CRUD endpoints instead |
| `evaluate_condition` | **No** | Only makes sense in workflow context |
| `seek_approval` | Yes | Manual approval requests |

### Async vs Sync Actions

| Action Type | Async | Description |
|-------------|-------|-------------|
| `create_alert` | No | Creates alert inline |
| `send_email` | Yes | Queues for email delivery |
| `send_notification` | Yes | Queues for notification delivery |
| `update_field` | No | Updates database inline |
| `allocate_inventory` | Yes | Queues for inventory processing |
| `evaluate_condition` | No | Evaluates inline (decision node) |
| `seek_approval` | Yes | Queues approval request |

## ActionExecutionContext

Context passed to action handlers:

```go
type ActionExecutionContext struct {
    EntityID      uuid.UUID              // The entity's UUID
    EntityName    string                 // Table name (e.g., "sales.orders")
    EventType     string                 // "on_create", "on_update", "on_delete"
    UserID        uuid.UUID              // User who triggered
    RuleID        *uuid.UUID             // The matched rule ID (nil for manual executions)
    RuleName      string                 // Rule name for logging
    ExecutionID   uuid.UUID              // Unique execution ID
    Timestamp     time.Time              // When the event occurred
    RawData       map[string]interface{} // Entity data for templates
    FieldChanges  map[string]FieldChange // Changed fields (update events)
    TriggerSource string                 // "automation" or "manual"
}
```

**Source**: `business/sdk/workflow/models.go`

Handlers can access:
- `context.EntityID` - The entity's UUID
- `context.EntityName` - Table name
- `context.UserID` - User who triggered
- `context.RawData["field"]` - Entity field values
- `context.FieldChanges["field"]` - Old and new values (for updates)
- `context.TriggerSource` - Whether triggered by automation or manual API call

## ActionResult and BranchTaken

Action execution returns an `ActionResult`:

```go
type ActionResult struct {
    ActionID   uuid.UUID
    ActionName string
    Success    bool
    Error      string
    Data       any          // Handler-specific result data
    BranchTaken string      // For condition nodes: "true_branch" or "false_branch"
}
```

**Source**: `business/sdk/workflow/models.go:116-123`

The `BranchTaken` field is set by `evaluate_condition` actions to indicate which path to follow in graph execution. See [branching.md](../branching.md) for details on how edges use this field.

## EntityModifier Interface

Optional interface for action handlers that modify database entities. Enables cascade visualization - showing which downstream workflows will trigger when an action modifies an entity.

```go
type EntityModifier interface {
    GetEntityModifications(config json.RawMessage) []EntityModification
}

type EntityModification struct {
    EntityName string   // Fully-qualified table name (e.g., "sales.orders")
    EventType  string   // "on_create", "on_update", or "on_delete"
    Fields     []string // Modified fields (for on_update events)
}
```

**Source**: `business/sdk/workflow/interfaces.go:140-161`

**Handlers implementing EntityModifier:**

| Handler | Implements | Entity Modified |
|---------|------------|-----------------|
| `update_field` | Yes | Configured `target_entity` |
| `create_alert` | No | Alerts are not cascade targets |
| `send_email` | No | No entity modification |
| `send_notification` | No | No entity modification |
| `allocate_inventory` | No | Future enhancement |
| `evaluate_condition` | No | Decision node only |
| `seek_approval` | No | No entity modification |

See [cascade-visualization.md](../cascade-visualization.md) for how this enables downstream workflow detection.

## Action Registry

Handlers are registered at application startup:

```go
// In all.go
registry := workflowEngine.GetRegistry()
registry.Register(communication.NewCreateAlertHandler(log, db, alertBus))
registry.Register(communication.NewSendEmailHandler(log, db))
registry.Register(communication.NewSendNotificationHandler(log, db))
registry.Register(data.NewUpdateFieldHandler(log, db))
registry.Register(approval.NewSeekApprovalHandler(log, db))
registry.Register(inventory.NewAllocateInventoryHandler(log, db, ...))
registry.Register(control.NewConditionHandler(log))  // Branching support
```

**Source**: `business/sdk/workflow/workflowactions/register.go`

## Execution Order

### Linear Execution (Default)

Actions execute based on their `execution_order` in the rule:

```
Order 1: [action_a, action_b]  ← Parallel (same order)
Order 2: [action_c]            ← Sequential (waits for order 1)
Order 3: [action_d, action_e]  ← Parallel (waits for order 2)
```

Actions with the same execution order run concurrently.

### Graph-Based Execution (Branching)

When `action_edges` are defined for a rule, the executor uses BFS graph traversal instead of `execution_order`:

```
[Start] ──start──▶ [Condition] ──true_branch──▶ [Action A]
                              └──false_branch─▶ [Action B]
```

Graph execution enables:
- Conditional branching with `evaluate_condition` actions
- Multiple entry points (parallel start edges)
- Converging branches (diamond patterns)
- Nested conditions

Rules WITHOUT edges automatically fall back to linear `execution_order` execution (backwards compatible).

See [branching.md](../branching.md) for detailed patterns and the [Edge API](../api-reference.md#edge-api-graph-based-execution) for creating edges.

## Template Support

Most action configurations support [template variables](../configuration/templates.md):

```json
{
  "title": "Order {{number}} - {{status | capitalize}}",
  "message": "Total: {{total | currency:USD}}"
}
```

Templates are processed before action execution.

## Error Handling

- Actions that fail don't block other actions
- Failures are logged and recorded in execution history
- Each action result includes error details if failed
- The overall execution continues even if individual actions fail

## Validation

All handlers validate their configuration:

1. **Required fields** - Must be present
2. **Value constraints** - Enums, formats, ranges
3. **Reference validity** - UUIDs, table names
4. **Template syntax** - Valid variable patterns

Invalid configurations are rejected when creating/updating rules.

## Adding New Action Types

To add a new action handler:

1. Create handler in `business/sdk/workflow/workflowactions/{category}/`
2. Implement `ActionHandler` interface
3. Register in `all.go` during startup
4. Add documentation

Example structure:
```go
package myaction

type MyActionHandler struct {
    log *logger.Logger
    db  *sqlx.DB
}

func NewMyActionHandler(log *logger.Logger, db *sqlx.DB) *MyActionHandler {
    return &MyActionHandler{log: log, db: db}
}

func (h *MyActionHandler) GetType() string {
    return "my_action"
}

func (h *MyActionHandler) Validate(config json.RawMessage) error {
    var cfg MyActionConfig
    if err := json.Unmarshal(config, &cfg); err != nil {
        return fmt.Errorf("invalid config: %w", err)
    }
    // Validate fields...
    return nil
}

func (h *MyActionHandler) Execute(ctx context.Context, config json.RawMessage, context workflow.ActionExecutionContext) (any, error) {
    var cfg MyActionConfig
    json.Unmarshal(config, &cfg)

    // Execute action...

    return map[string]interface{}{
        "status":  "success",
        "message": "Action completed",
    }, nil
}
```
