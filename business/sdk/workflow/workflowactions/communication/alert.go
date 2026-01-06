package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

// Ensure workflow import is used (for ActionExecutionContext)
var _ workflow.ActionExecutionContext

// templateVarPattern matches {{variable_name}} patterns for template substitution.
var templateVarPattern = regexp.MustCompile(`\{\{([^}]+)\}\}`)

// AlertConfig represents the configuration for a create_alert action.
type AlertConfig struct {
	AlertType  string `json:"alert_type"`
	Severity   string `json:"severity"`
	Title      string `json:"title"`
	Message    string `json:"message"`
	Recipients struct {
		Users []string `json:"users"`
		Roles []string `json:"roles"`
	} `json:"recipients"`
	Context json.RawMessage `json:"context"`
}

// CreateAlertHandler handles create_alert actions.
type CreateAlertHandler struct {
	log           *logger.Logger
	alertBus      *alertbus.Business
	workflowQueue *rabbitmq.WorkflowQueue
}

// NewCreateAlertHandler creates a new create alert handler.
func NewCreateAlertHandler(log *logger.Logger, alertBus *alertbus.Business, workflowQueue *rabbitmq.WorkflowQueue) *CreateAlertHandler {
	return &CreateAlertHandler{
		log:           log,
		alertBus:      alertBus,
		workflowQueue: workflowQueue,
	}
}

// GetType returns the action type this handler supports.
func (h *CreateAlertHandler) GetType() string {
	return "create_alert"
}

// Validate validates the action configuration before execution.
func (h *CreateAlertHandler) Validate(config json.RawMessage) error {
	var cfg AlertConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if cfg.Message == "" {
		return fmt.Errorf("alert message is required")
	}

	if len(cfg.Recipients.Users) == 0 && len(cfg.Recipients.Roles) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}

	validSeverities := map[string]bool{
		"low": true, "medium": true, "high": true, "critical": true,
	}
	if cfg.Severity != "" && !validSeverities[cfg.Severity] {
		return fmt.Errorf("invalid severity level: %s", cfg.Severity)
	}

	return nil
}

// Execute creates an alert via the business layer.
func (h *CreateAlertHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (interface{}, error) {
	var cfg AlertConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	now := time.Now()

	// Set default severity if not provided
	severity := cfg.Severity
	if severity == "" {
		severity = alertbus.SeverityMedium
	}

	// Default context to empty JSON object if not provided
	context := cfg.Context
	if len(context) == 0 {
		context = json.RawMessage(`{}`)
	}

	// Build Alert struct
	alert := alertbus.Alert{
		ID:               uuid.New(),
		AlertType:        cfg.AlertType,
		Severity:         severity,
		Title:            resolveTemplateVars(cfg.Title, execCtx.RawData),
		Message:          resolveTemplateVars(cfg.Message, execCtx.RawData),
		Context:          context,
		SourceEntityName: execCtx.EntityName,
		SourceEntityID:   execCtx.EntityID,
		SourceRuleID:     execCtx.RuleID,
		Status:           alertbus.StatusActive,
		CreatedDate:      now,
		UpdatedDate:      now,
	}

	// Build recipients slice - validate all UUIDs first (fail fast on invalid config)
	var recipients []alertbus.AlertRecipient

	for _, u := range cfg.Recipients.Users {
		uid, err := uuid.Parse(u)
		if err != nil {
			return nil, fmt.Errorf("invalid user UUID %q: %w", u, err)
		}
		recipients = append(recipients, alertbus.AlertRecipient{
			ID:            uuid.New(),
			AlertID:       alert.ID,
			RecipientType: "user",
			RecipientID:   uid,
			CreatedDate:   now,
		})
	}

	for _, r := range cfg.Recipients.Roles {
		rid, err := uuid.Parse(r)
		if err != nil {
			return nil, fmt.Errorf("invalid role UUID %q: %w", r, err)
		}
		recipients = append(recipients, alertbus.AlertRecipient{
			ID:            uuid.New(),
			AlertID:       alert.ID,
			RecipientType: "role",
			RecipientID:   rid,
			CreatedDate:   now,
		})
	}

	// Create alert via business layer
	if err := h.alertBus.Create(ctx, alert); err != nil {
		return nil, fmt.Errorf("create alert: %w", err)
	}

	// Create all recipients via batch insert
	if err := h.alertBus.CreateRecipients(ctx, recipients); err != nil {
		return nil, fmt.Errorf("create recipients: %w", err)
	}

	// Publish to RabbitMQ for WebSocket delivery
	if h.workflowQueue != nil {
		h.publishAlertToWebSocket(ctx, alert, recipients)
	}

	h.log.Info(ctx, "create_alert action executed",
		"alert_id", alert.ID,
		"entity_id", execCtx.EntityID,
		"rule_name", execCtx.RuleName,
		"recipients", len(recipients))

	return map[string]interface{}{
		"alert_id": alert.ID.String(),
		"status":   "created",
	}, nil
}

// resolveTemplateVars replaces {{variable_name}} patterns with values from the data map.
func resolveTemplateVars(template string, data map[string]interface{}) string {
	if data == nil {
		return template
	}

	return templateVarPattern.ReplaceAllStringFunc(template, func(match string) string {
		// Extract variable name from {{variable_name}}
		varName := match[2 : len(match)-2]
		if value, ok := data[varName]; ok {
			return fmt.Sprintf("%v", value)
		}
		return match // Keep original if not found
	})
}

// publishAlertToWebSocket publishes alert messages to RabbitMQ for WebSocket delivery.
// Each recipient gets a targeted message (user or role).
func (h *CreateAlertHandler) publishAlertToWebSocket(ctx context.Context, alert alertbus.Alert, recipients []alertbus.AlertRecipient) {
	// Build the alert payload once (shared across all messages)
	alertData := map[string]interface{}{
		"id":          alert.ID.String(),
		"alertType":   alert.AlertType,
		"severity":    alert.Severity,
		"title":       alert.Title,
		"message":     alert.Message,
		"status":      alert.Status,
		"createdDate": alert.CreatedDate.Format(time.RFC3339),
		"updatedDate": alert.UpdatedDate.Format(time.RFC3339),
	}

	for _, recipient := range recipients {
		// Build payload with alert data
		payload := map[string]interface{}{
			"alert": alertData,
		}

		msg := &rabbitmq.Message{
			Type:       "alert",
			EntityName: "workflow.alerts",
			EntityID:   alert.ID,
		}

		// Target by user or role
		if recipient.RecipientType == "user" {
			msg.UserID = recipient.RecipientID
		} else if recipient.RecipientType == "role" {
			payload["role_id"] = recipient.RecipientID.String()
		}

		msg.Payload = payload

		if err := h.workflowQueue.Publish(ctx, rabbitmq.QueueTypeAlert, msg); err != nil {
			h.log.Error(ctx, "failed to publish alert to rabbitmq",
				"alert_id", alert.ID,
				"recipient_type", recipient.RecipientType,
				"recipient_id", recipient.RecipientID,
				"error", err)
			// Continue - alert is already persisted, log error but don't fail
		}
	}
}
