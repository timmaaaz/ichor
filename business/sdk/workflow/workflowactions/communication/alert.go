package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
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
	Context      json.RawMessage `json:"context"`
	ActionURL    string          `json:"action_url"`
	ResolvePrior bool            `json:"resolve_prior"` // When true, resolves prior alerts with same source_entity_id + alert_type
}

// CreateAlertHandler handles create_alert actions.
type CreateAlertHandler struct {
	log           *logger.Logger
	alertBus      *alertbus.Business
	workflowQueue *rabbitmq.WorkflowQueue
	// ordersBus and productBus are optional (nil-tolerant): when wired, the handler
	// resolves order_id -> order_number and product_id -> product_name so alert
	// messages render human labels instead of UUIDs.
	ordersBus  *ordersbus.Business
	productBus *productbus.Business
}

// NewCreateAlertHandler creates a new create alert handler. ordersBus/productBus
// may be nil (e.g. core/test registration); FK-label resolution is then skipped.
func NewCreateAlertHandler(log *logger.Logger, alertBus *alertbus.Business, workflowQueue *rabbitmq.WorkflowQueue, ordersBus *ordersbus.Business, productBus *productbus.Business) *CreateAlertHandler {
	return &CreateAlertHandler{
		log:           log,
		alertBus:      alertBus,
		workflowQueue: workflowQueue,
		ordersBus:     ordersBus,
		productBus:    productBus,
	}
}

// GetType returns the action type this handler supports.
func (h *CreateAlertHandler) GetType() string {
	return "create_alert"
}

// SupportsManualExecution returns true - alerts can be triggered manually
func (h *CreateAlertHandler) SupportsManualExecution() bool {
	return true
}

// IsAsync returns false - alert creation completes inline
func (h *CreateAlertHandler) IsAsync() bool {
	return false
}

// GetDescription returns a human-readable description for discovery APIs
func (h *CreateAlertHandler) GetDescription() string {
	return "Create an alert notification for users or roles"
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

	// Handle pointer-based RuleID (nil for manual executions)
	sourceRuleID := uuid.Nil
	if execCtx.RuleID != nil {
		sourceRuleID = *execCtx.RuleID
	}

	// Augment template data + context with execution_id / rule_id (deep-linking).
	tmplData := buildAlertTemplateData(execCtx.RawData, execCtx.ExecutionID, execCtx.RuleID)

	// Resolve FK ids referenced by the templates into human labels (order_number,
	// product_name) so messages render names instead of UUIDs. Mirrors the alert
	// domain's rule/acknowledger-name resolution; fail-open (leaves the literal id).
	h.resolveEntityLabels(ctx, tmplData, cfg.Title, cfg.Message, cfg.ActionURL)

	rawContext := cfg.Context
	if len(rawContext) == 0 {
		rawContext = json.RawMessage(`{}`)
	}
	enrichedContext, err := enrichAlertContext(rawContext, execCtx.ExecutionID, execCtx.RuleID)
	if err != nil {
		return nil, err
	}

	// Build Alert struct
	alert := alertbus.Alert{
		ID:               uuid.New(),
		AlertType:        cfg.AlertType,
		Severity:         severity,
		Title:            resolveTemplateVars(cfg.Title, tmplData),
		Message:          resolveTemplateVars(cfg.Message, tmplData),
		Context:          enrichedContext,
		SourceEntityName: execCtx.EntityName,
		SourceEntityID:   execCtx.EntityID,
		SourceRuleID:     sourceRuleID,
		Status:           alertbus.StatusActive,
		CreatedDate:      now,
		UpdatedDate:      now,
	}

	// Set action URL with template variable substitution
	if cfg.ActionURL != "" {
		alert.ActionURL = resolveTemplateVars(cfg.ActionURL, tmplData)
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
	if h.alertBus == nil {
		h.log.Warn(ctx, "create_alert: alertBus not configured, skipping alert creation")
		return map[string]interface{}{
			"alert_id":       uuid.Nil.String(),
			"status":         "skipped",
			"resolved_count": 0,
		}, nil
	}
	if err := h.alertBus.Create(ctx, alert); err != nil {
		return nil, fmt.Errorf("create alert: %w", err)
	}

	// Create all recipients via batch insert
	if err := h.alertBus.CreateRecipients(ctx, recipients); err != nil {
		return nil, fmt.Errorf("create recipients: %w", err)
	}

	// Auto-resolve prior related alerts if configured
	var resolvedCount int
	if cfg.ResolvePrior && execCtx.EntityID != uuid.Nil && cfg.AlertType != "" {
		var err error
		resolvedCount, err = h.alertBus.ResolveRelatedAlerts(ctx, execCtx.EntityID, cfg.AlertType, alert.ID, now)
		if err != nil {
			// Log error but don't fail the alert creation
			h.log.Error(ctx, "failed to resolve prior alerts",
				"error", err,
				"source_entity_id", execCtx.EntityID,
				"alert_type", cfg.AlertType)
		}
	}

	// Publish to RabbitMQ for WebSocket delivery
	if h.workflowQueue != nil {
		h.publishAlertToWebSocket(ctx, alert, recipients)
	}

	h.log.Info(ctx, "create_alert action executed",
		"alert_id", alert.ID,
		"entity_id", execCtx.EntityID,
		"rule_name", execCtx.RuleName,
		"recipients", len(recipients),
		"resolved_prior", resolvedCount)

	return map[string]interface{}{
		"alert_id":       alert.ID.String(),
		"status":         "created",
		"resolved_count": resolvedCount,
	}, nil
}

// buildAlertTemplateData returns a copy of rawData with execution_id and
// rule_id added, so {{execution_id}} / {{rule_id}} resolve in alert templates.
// It never mutates the caller's map.
func buildAlertTemplateData(rawData map[string]any, execID uuid.UUID, ruleID *uuid.UUID) map[string]any {
	out := make(map[string]any, len(rawData)+2)
	for k, v := range rawData {
		out[k] = v
	}
	if execID != uuid.Nil {
		out["execution_id"] = execID.String()
	}
	if ruleID != nil {
		out["rule_id"] = ruleID.String()
	}
	return out
}

// resolveEntityLabels enriches the template data with human-readable labels for FK
// ids that the alert templates reference — order_id -> order_number and
// product_id -> product_name. A lookup runs only when (a) the relevant bus is
// wired, (b) some template actually references the label key (so non-over-order
// alerts incur no DB cost), and (c) the id is present and parseable. Lookup
// failures are logged and skipped, leaving the literal placeholder — they never
// fail the alert. Resolved values are written onto data (a copy of RawData).
func (h *CreateAlertHandler) resolveEntityLabels(ctx context.Context, data map[string]any, templates ...string) {
	referenced := func(key string) bool {
		needle := "{{" + key + "}}"
		for _, t := range templates {
			if strings.Contains(t, needle) {
				return true
			}
		}
		return false
	}

	if h.ordersBus != nil && referenced("order_number") {
		if id, ok := uuidFromData(data, "order_id"); ok {
			if order, err := h.ordersBus.QueryByID(ctx, id); err != nil {
				h.log.Error(ctx, "create_alert: resolve order_number failed", "order_id", id, "error", err)
			} else {
				data["order_number"] = order.Number
			}
		}
	}

	if h.productBus != nil && referenced("product_name") {
		if id, ok := uuidFromData(data, "product_id"); ok {
			if product, err := h.productBus.QueryByID(ctx, id); err != nil {
				h.log.Error(ctx, "create_alert: resolve product_name failed", "product_id", id, "error", err)
			} else {
				data["product_name"] = product.Name
			}
		}
	}
}

// uuidFromData parses data[key] as a UUID. Trigger-data ids arrive as strings.
func uuidFromData(data map[string]any, key string) (uuid.UUID, bool) {
	s, ok := data[key].(string)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

// enrichAlertContext merges execution_id and rule_id into the alert's context
// JSON object so the frontend can deep-link an alert to its execution.
func enrichAlertContext(ctx json.RawMessage, execID uuid.UUID, ruleID *uuid.UUID) (json.RawMessage, error) {
	m := map[string]any{}
	if len(ctx) > 0 {
		if err := json.Unmarshal(ctx, &m); err != nil {
			return nil, fmt.Errorf("parse alert context: %w", err)
		}
	}
	if execID != uuid.Nil {
		m["execution_id"] = execID.String()
	}
	if ruleID != nil {
		m["rule_id"] = ruleID.String()
	}
	return json.Marshal(m)
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

// publishAlertToWebSocket delegates to the shared PublishAlertToRecipients
// helper so this handler and other alert emitters (e.g. seek_approval) share a
// single publishing path and payload shape.
func (h *CreateAlertHandler) publishAlertToWebSocket(ctx context.Context, alert alertbus.Alert, recipients []alertbus.AlertRecipient) {
	PublishAlertToRecipients(ctx, h.workflowQueue, h.log, alert, recipients)
}
