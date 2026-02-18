package approval

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// seekApprovalConfig represents the configuration for a seek_approval action.
type seekApprovalConfig struct {
	Approvers       []string `json:"approvers"`
	ApprovalType    string   `json:"approval_type"`
	TimeoutHours    int      `json:"timeout_hours"`
	ApprovalMessage string   `json:"approval_message"`
}

// SeekApprovalHandler handles seek_approval actions.
type SeekApprovalHandler struct {
	log                *logger.Logger
	db                 *sqlx.DB
	approvalRequestBus *approvalrequestbus.Business
	alertBus           *alertbus.Business
}

// NewSeekApprovalHandler creates a new seek approval handler.
// approvalRequestBus and alertBus can be nil for core registration (graceful degradation).
func NewSeekApprovalHandler(log *logger.Logger, db *sqlx.DB, approvalRequestBus *approvalrequestbus.Business, alertBus *alertbus.Business) *SeekApprovalHandler {
	return &SeekApprovalHandler{
		log:                log,
		db:                 db,
		approvalRequestBus: approvalRequestBus,
		alertBus:           alertBus,
	}
}

func (h *SeekApprovalHandler) GetType() string {
	return "seek_approval"
}

// SupportsManualExecution returns true - approval requests can be initiated manually.
func (h *SeekApprovalHandler) SupportsManualExecution() bool {
	return true
}

// IsAsync returns true - this handler completes via Temporal async completion.
func (h *SeekApprovalHandler) IsAsync() bool {
	return true
}

// GetDescription returns a human-readable description for discovery APIs.
func (h *SeekApprovalHandler) GetDescription() string {
	return "Request approval from specified approvers"
}

func (h *SeekApprovalHandler) Validate(config json.RawMessage) error {
	var cfg seekApprovalConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if len(cfg.Approvers) == 0 {
		return fmt.Errorf("approvers list is required and must not be empty")
	}

	validTypes := map[string]bool{"any": true, "all": true, "majority": true}
	if !validTypes[cfg.ApprovalType] {
		return fmt.Errorf("invalid approval_type, must be: any, all, or majority")
	}

	return nil
}

// GetOutputPorts implements workflow.OutputPortProvider.
func (h *SeekApprovalHandler) GetOutputPorts() []workflow.OutputPort {
	return []workflow.OutputPort{
		{Name: "approved", Description: "Approval was granted", IsDefault: true},
		{Name: "rejected", Description: "Approval was denied"},
		{Name: "timed_out", Description: "Approval request timed out"},
	}
}

// StartAsync initiates the async approval operation. It creates an approval
// request record with the Temporal task token, then creates an alert for approvers.
// The Temporal activity will return ErrResultPending after this method completes.
func (h *SeekApprovalHandler) StartAsync(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext, taskToken []byte) error {
	if h.approvalRequestBus == nil || h.alertBus == nil {
		return fmt.Errorf("seek_approval requires approval request bus and alert bus (not available in core registration)")
	}

	var cfg seekApprovalConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Parse approver UUIDs.
	approvers := make([]uuid.UUID, len(cfg.Approvers))
	for i, s := range cfg.Approvers {
		id, err := uuid.Parse(s)
		if err != nil {
			return fmt.Errorf("invalid approver UUID %q at index %d: %w", s, i, err)
		}
		approvers[i] = id
	}

	// Log warning for unimplemented approval types.
	if cfg.ApprovalType != approvalrequestbus.ApprovalTypeAny {
		h.log.Warn(ctx, "approval_type not fully implemented, treating as 'any'",
			"approval_type", cfg.ApprovalType)
	}

	// Default timeout.
	timeoutHours := cfg.TimeoutHours
	if timeoutHours <= 0 {
		timeoutHours = 72
	}

	// Determine rule ID.
	ruleID := uuid.Nil
	if execCtx.RuleID != nil {
		ruleID = *execCtx.RuleID
	}

	// Create approval request with base64-encoded task token.
	req, err := h.approvalRequestBus.Create(ctx, approvalrequestbus.NewApprovalRequest{
		ExecutionID:     execCtx.ExecutionID,
		RuleID:          ruleID,
		ActionName:      execCtx.ActionName,
		Approvers:       approvers,
		ApprovalType:    cfg.ApprovalType,
		TimeoutHours:    timeoutHours,
		TaskToken:       base64.StdEncoding.EncodeToString(taskToken),
		ApprovalMessage: cfg.ApprovalMessage,
	})
	if err != nil {
		return fmt.Errorf("create approval request: %w", err)
	}

	// Create alert notification for approvers.
	h.createApprovalAlert(ctx, req, execCtx)

	h.log.Info(ctx, "seek_approval action started",
		"approval_request_id", req.ID,
		"execution_id", execCtx.ExecutionID,
		"approvers", len(approvers),
		"rule_name", execCtx.RuleName)

	return nil
}

// Execute provides a synchronous fallback for manual execution.
func (h *SeekApprovalHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (interface{}, error) {
	if h.approvalRequestBus == nil {
		// Graceful degradation: return stub result.
		return map[string]interface{}{
			"output": "approved",
			"status": "stub",
		}, nil
	}

	var cfg seekApprovalConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	approvers := make([]uuid.UUID, len(cfg.Approvers))
	for i, s := range cfg.Approvers {
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("invalid approver UUID %q: %w", s, err)
		}
		approvers[i] = id
	}

	timeoutHours := cfg.TimeoutHours
	if timeoutHours <= 0 {
		timeoutHours = 72
	}

	ruleID := uuid.Nil
	if execCtx.RuleID != nil {
		ruleID = *execCtx.RuleID
	}

	// Create approval request with empty task token (manual execution, no Temporal).
	req, err := h.approvalRequestBus.Create(ctx, approvalrequestbus.NewApprovalRequest{
		ExecutionID:     execCtx.ExecutionID,
		RuleID:          ruleID,
		ActionName:      execCtx.ActionName,
		Approvers:       approvers,
		ApprovalType:    cfg.ApprovalType,
		TimeoutHours:    timeoutHours,
		TaskToken:       "",
		ApprovalMessage: cfg.ApprovalMessage,
	})
	if err != nil {
		return nil, fmt.Errorf("create approval request: %w", err)
	}

	return map[string]interface{}{
		"approval_id": req.ID.String(),
		"output":      "pending",
		"status":      "pending",
	}, nil
}

// createApprovalAlert creates an alert with per-approver recipients.
func (h *SeekApprovalHandler) createApprovalAlert(ctx context.Context, req approvalrequestbus.ApprovalRequest, execCtx workflow.ActionExecutionContext) {
	if h.alertBus == nil {
		return
	}

	now := time.Now()

	sourceRuleID := uuid.Nil
	if execCtx.RuleID != nil {
		sourceRuleID = *execCtx.RuleID
	}

	title := fmt.Sprintf("Approval Required: %s", execCtx.RuleName)
	message := req.ApprovalMessage
	if message == "" {
		message = fmt.Sprintf("An approval request has been created for workflow '%s'", execCtx.RuleName)
	}

	alert := alertbus.Alert{
		ID:               uuid.New(),
		AlertType:        "approval_request",
		Severity:         alertbus.SeverityHigh,
		Title:            title,
		Message:          message,
		Context:          json.RawMessage(`{}`),
		SourceEntityName: execCtx.EntityName,
		SourceEntityID:   execCtx.EntityID,
		SourceRuleID:     sourceRuleID,
		Status:           alertbus.StatusActive,
		CreatedDate:      now,
		UpdatedDate:      now,
	}

	if err := h.alertBus.Create(ctx, alert); err != nil {
		h.log.Error(ctx, "failed to create approval alert",
			"approval_request_id", req.ID,
			"error", err)
		return
	}

	var recipients []alertbus.AlertRecipient
	for _, approverID := range req.Approvers {
		recipients = append(recipients, alertbus.AlertRecipient{
			ID:            uuid.New(),
			AlertID:       alert.ID,
			RecipientType: "user",
			RecipientID:   approverID,
			CreatedDate:   now,
		})
	}

	if err := h.alertBus.CreateRecipients(ctx, recipients); err != nil {
		h.log.Error(ctx, "failed to create approval alert recipients",
			"approval_request_id", req.ID,
			"error", err)
	}
}
