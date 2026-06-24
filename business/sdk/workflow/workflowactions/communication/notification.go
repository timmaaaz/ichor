package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

// SendNotificationHandler handles send_notification actions. A notification is a
// single-user, informational alert that rides the durable alert pipeline.
type SendNotificationHandler struct {
	log           *logger.Logger
	alertBus      *alertbus.Business
	workflowQueue *rabbitmq.WorkflowQueue
}

// NewSendNotificationHandler creates a new send notification handler.
// alertBus persists the notification as a single-user alert (nil = no
// persistence, e.g. validation-only registries).
func NewSendNotificationHandler(log *logger.Logger, alertBus *alertbus.Business, workflowQueue *rabbitmq.WorkflowQueue) *SendNotificationHandler {
	return &SendNotificationHandler{
		log:           log,
		alertBus:      alertBus,
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

	now := time.Now()

	// A notification is a single-user informational alert riding the same durable
	// pipeline (alertbus -> RabbitMQ -> AlertHub -> WebSocket) the frontend already
	// consumes. Severity carries the config priority: low = silent inbox entry,
	// high/critical interrupt (frontend toast gating). alert_type "notification"
	// lets the UI treat these distinctly later. SourceEntityID is deliberately
	// left nil so personal notifications are NOT collapsed in the frontend's alert
	// bundling view (bundle key = source_entity_id:alert_type).
	severity := cfg.Priority
	if severity == "" {
		severity = alertbus.SeverityLow
	}

	sourceRuleID := uuid.Nil
	if execCtx.RuleID != nil {
		sourceRuleID = *execCtx.RuleID
	}

	alert := alertbus.Alert{
		ID:           uuid.New(),
		AlertType:    "notification",
		Severity:     severity,
		Title:        resolveTemplateVars(cfg.Title, execCtx.RawData),
		Message:      resolveTemplateVars(cfg.Message, execCtx.RawData),
		Context:      json.RawMessage(`{}`),
		SourceRuleID: sourceRuleID,
		Status:       alertbus.StatusActive,
		CreatedDate:  now,
		UpdatedDate:  now,
	}

	// Notifications target users only (no role fan-out). Validate all UUIDs first.
	var recipients []alertbus.AlertRecipient
	for _, u := range cfg.Recipients {
		uid, err := uuid.Parse(u)
		if err != nil {
			return nil, fmt.Errorf("invalid recipient UUID %q: %w", u, err)
		}
		recipients = append(recipients, alertbus.AlertRecipient{
			ID:            uuid.New(),
			AlertID:       alert.ID,
			RecipientType: "user",
			RecipientID:   uid,
			CreatedDate:   now,
		})
	}

	// Graceful degradation when no alert bus is wired (validation-only registries).
	if h.alertBus == nil {
		h.log.Warn(ctx, "send_notification: alertBus not configured, skipping notification")
		return map[string]interface{}{
			"notification_id": uuid.Nil.String(),
			"status":          "skipped",
			"recipients":      0,
		}, nil
	}

	if err := h.alertBus.Create(ctx, alert); err != nil {
		return nil, fmt.Errorf("create notification alert: %w", err)
	}
	if err := h.alertBus.CreateRecipients(ctx, recipients); err != nil {
		return nil, fmt.Errorf("create notification recipients: %w", err)
	}

	// Publish for real-time WebSocket delivery via the shared alert publish seam.
	if h.workflowQueue != nil {
		PublishAlertToRecipients(ctx, h.workflowQueue, h.log, alert, recipients)
	}

	h.log.Info(ctx, "send_notification executed",
		"notification_id", alert.ID,
		"recipients", len(recipients),
		"severity", severity)

	return map[string]interface{}{
		"notification_id": alert.ID.String(),
		"status":          "sent",
		"recipients":      len(recipients),
	}, nil
}
