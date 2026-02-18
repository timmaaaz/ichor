# Phase 3: Implement send_notification

**Category**: Backend
**Status**: Pending
**Dependencies**: None
**Effort**: Low

---

## Overview

`send_notification` is currently a stub — its `Execute()` returns a fake `notification_id` without delivering anything. The infrastructure for real-time in-app notification delivery already exists: `create_alert` uses `rabbitmq.WorkflowQueue` to publish messages that the WebSocket layer delivers to connected clients.

`send_notification` should use the same RabbitMQ path but **without persisting** an alert record — making it an ephemeral notification vs. a persistent dismissible alert.

---

## Goals

1. Wire `SendNotificationHandler` to the `rabbitmq.WorkflowQueue`
2. Publish ephemeral notification messages (same queue, different message type)
3. Resolve `{{template_vars}}` in the notification message

---

## Current State

```go
// notification.go Execute() — current stub:
func (h *SendNotificationHandler) Execute(...) (interface{}, error) {
    h.log.Info(ctx, "Executing send_notification action", ...)
    result := map[string]interface{}{
        "notification_id": fmt.Sprintf("notif_%d", time.Now().Unix()),
        "status":          "sent",
        "sent_at":         time.Now().Format(time.RFC3339),
    }
    return result, nil
}
```

---

## Task Breakdown

### Task 1: Add workflowQueue to Handler

**File**: `business/sdk/workflow/workflowactions/communication/notification.go`

```go
type SendNotificationHandler struct {
    log           *logger.Logger
    db            *sqlx.DB
    workflowQueue *rabbitmq.WorkflowQueue // ADD THIS
}

func NewSendNotificationHandler(log *logger.Logger, db *sqlx.DB, workflowQueue *rabbitmq.WorkflowQueue) *SendNotificationHandler {
    return &SendNotificationHandler{
        log:           log,
        db:            db,
        workflowQueue: workflowQueue,
    }
}
```

### Task 2: Update register.go

**File**: `business/sdk/workflow/workflowactions/register.go`

In `RegisterAll`:
```go
registry.Register(communication.NewSendNotificationHandler(config.Log, config.DB, config.QueueClient))
```

In `RegisterCoreActions` (test environments — no queue):
```go
registry.Register(communication.NewSendNotificationHandler(config.Log, config.DB, nil))
```

### Task 3: Implement Execute with RabbitMQ Delivery

**File**: `business/sdk/workflow/workflowactions/communication/notification.go`

```go
func (h *SendNotificationHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (interface{}, error) {
    var cfg struct {
        Recipients []string `json:"recipients"`
        Channels   []struct {
            Type string `json:"type"`
        } `json:"channels"`
        Priority string `json:"priority"`
        Message  string `json:"message"`
        Title    string `json:"title,omitempty"`
    }
    if err := json.Unmarshal(config, &cfg); err != nil {
        return nil, fmt.Errorf("failed to parse notification config: %w", err)
    }

    notificationID := uuid.New()
    now := time.Now()

    // Resolve template variables
    message := resolveTemplateVars(cfg.Message, execCtx.RawData)
    title := resolveTemplateVars(cfg.Title, execCtx.RawData)

    if h.workflowQueue != nil {
        for _, recipientID := range cfg.Recipients {
            uid, err := uuid.Parse(recipientID)
            if err != nil {
                h.log.Warn(ctx, "send_notification: invalid recipient UUID", "recipient", recipientID)
                continue
            }

            msg := &rabbitmq.Message{
                Type:       "notification",
                EntityName: "workflow.notifications",
                EntityID:   notificationID,
                UserID:     uid,
                Payload: map[string]interface{}{
                    "notification_id": notificationID.String(),
                    "title":           title,
                    "message":         message,
                    "priority":        cfg.Priority,
                    "created_date":    now.Format(time.RFC3339),
                },
            }
            if err := h.workflowQueue.Publish(ctx, rabbitmq.QueueTypeAlert, msg); err != nil {
                h.log.Error(ctx, "send_notification: publish failed",
                    "recipient", recipientID, "error", err)
            }
        }
    }

    h.log.Info(ctx, "send_notification executed",
        "notification_id", notificationID,
        "recipients", len(cfg.Recipients),
        "priority", cfg.Priority)

    return map[string]interface{}{
        "notification_id": notificationID.String(),
        "status":          "sent",
        "sent_at":         now.Format(time.RFC3339),
        "recipients":      len(cfg.Recipients),
    }, nil
}
```

Note: `resolveTemplateVars` is defined in `alert.go` in the same package — reuse it directly.

Also add `message` and `title` to the `Validate()` function:
```go
if cfg.Message == "" {
    return fmt.Errorf("notification message is required")
}
```

---

## Validation

```bash
go build ./...

# Check rabbitmq.Message type
grep -A 10 "type Message struct" foundation/rabbitmq/

# Check QueueType constants
grep "QueueType" foundation/rabbitmq/

# Check resolveTemplateVars is accessible (same package)
grep "func resolveTemplateVars" business/sdk/workflow/workflowactions/communication/
```

---

## Gotchas

- **`resolveTemplateVars` is in `alert.go`** — it's unexported but in the same `communication` package, so it's accessible directly without import.
- **`uuid.Parse` vs string UUID** — `cfg.Recipients` are user UUIDs as strings. Parse them before passing to RabbitMQ message.
- **`RegisterCoreActions` passes nil** — guard with `if h.workflowQueue != nil` before publishing. The stub behavior (just log + return) is acceptable in test environments.
- **No `message` field in current Validate()** — the current config struct only validates `recipients`, `channels`, `priority`. Add `message` as required.
- **`db` field may become unused** — if the new implementation only uses `workflowQueue`, the `db` field becomes dead weight. Keep it for now (removing it would change the constructor signature and may be needed for future features like notification history).
