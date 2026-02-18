package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

// SendNotificationHandler handles send_notification actions
type SendNotificationHandler struct {
	log           *logger.Logger
	workflowQueue *rabbitmq.WorkflowQueue
}

// NewSendNotificationHandler creates a new send notification handler
func NewSendNotificationHandler(log *logger.Logger, workflowQueue *rabbitmq.WorkflowQueue) *SendNotificationHandler {
	return &SendNotificationHandler{
		log:           log,
		workflowQueue: workflowQueue,
	}
}

func (h *SendNotificationHandler) GetType() string {
	return "send_notification"
}

// SupportsManualExecution returns true - notifications can be sent manually
func (h *SendNotificationHandler) SupportsManualExecution() bool {
	return true
}

// IsAsync returns false - notification creation completes inline
func (h *SendNotificationHandler) IsAsync() bool {
	return false
}

// GetDescription returns a human-readable description for discovery APIs
func (h *SendNotificationHandler) GetDescription() string {
	return "Send an in-app notification"
}

func (h *SendNotificationHandler) Validate(config json.RawMessage) error {
	var cfg struct {
		Recipients []string `json:"recipients"`
		Priority   string   `json:"priority"`
		Message    string   `json:"message"`
	}

	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if len(cfg.Recipients) == 0 {
		return fmt.Errorf("recipients list is required and must not be empty")
	}

	if cfg.Message == "" {
		return fmt.Errorf("notification message is required")
	}

	validPriorities := map[string]bool{
		"low": true, "medium": true, "high": true, "critical": true,
	}
	if !validPriorities[cfg.Priority] {
		return fmt.Errorf("invalid priority level")
	}

	return nil
}

func (h *SendNotificationHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (interface{}, error) {
	var cfg struct {
		Recipients []string `json:"recipients"`
		Priority   string   `json:"priority"`
		Message    string   `json:"message"`
		Title      string   `json:"title,omitempty"`
	}
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse notification config: %w", err)
	}

	notificationID := uuid.New()
	now := time.Now()

	// Resolve template variables in message and title
	message := resolveTemplateVars(cfg.Message, execCtx.RawData)
	title := resolveTemplateVars(cfg.Title, execCtx.RawData)

	// Publish to RabbitMQ for WebSocket delivery (ephemeral — no DB persistence)
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
				EventType:  "send",
				UserID:     uid,
				Payload: map[string]interface{}{
					"notification_id": notificationID.String(),
					"title":           title,
					"message":         message,
					"priority":        cfg.Priority,
					"created_date":    now.Format(time.RFC3339),
				},
			}
			if err := h.workflowQueue.Publish(ctx, rabbitmq.QueueTypeNotification, msg); err != nil {
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
