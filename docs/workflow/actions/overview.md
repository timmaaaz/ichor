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

## ActionHandler Interface

All action handlers implement this interface:

```go
type ActionHandler interface {
    Type() string
    Validate(config json.RawMessage) error
    Execute(ctx context.Context, config json.RawMessage, execCtx ExecutionContext) (*ActionResult, error)
}
```

**Source**: `business/sdk/workflow/interfaces.go`

### Methods

| Method | Purpose |
|--------|---------|
| `Type()` | Returns the action type identifier (e.g., `"create_alert"`) |
| `Validate()` | Validates configuration before rule creation |
| `Execute()` | Executes the action with given config and context |

## ActionResult

Actions return an `ActionResult` with execution details:

```go
type ActionResult struct {
    Status       string                 // "success", "failed", "skipped"
    ActionType   string                 // The action type
    Message      string                 // Human-readable result
    Data         map[string]interface{} // Action-specific data
    Error        string                 // Error message if failed
    Duration     time.Duration          // Execution time
}
```

## ExecutionContext

Context passed to action handlers:

```go
type ExecutionContext struct {
    Event       TriggerEvent           // The triggering event
    Rule        AutomationRule         // The matched rule
    ExecutionID uuid.UUID              // Unique execution ID
    RawData     map[string]interface{} // Entity data for templates
}
```

Handlers can access:
- `execCtx.Event.EntityID` - The entity's UUID
- `execCtx.Event.EntityName` - Table name
- `execCtx.Event.UserID` - User who triggered
- `execCtx.RawData["field"]` - Entity field values

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
```

**Source**: `business/sdk/workflow/workflowactions/register.go`

## Execution Order

Actions execute based on their `execution_order` in the rule:

```
Order 1: [action_a, action_b]  ← Parallel (same order)
Order 2: [action_c]            ← Sequential (waits for order 1)
Order 3: [action_d, action_e]  ← Parallel (waits for order 2)
```

Actions with the same execution order run concurrently.

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

func (h *MyActionHandler) Type() string {
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

func (h *MyActionHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ExecutionContext) (*workflow.ActionResult, error) {
    var cfg MyActionConfig
    json.Unmarshal(config, &cfg)

    // Execute action...

    return &workflow.ActionResult{
        Status:     "success",
        ActionType: h.Type(),
        Message:    "Action completed",
    }, nil
}
```
