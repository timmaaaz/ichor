package approval

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// ResolveApprovalConfig holds the configuration for the resolve_approval_request action.
type ResolveApprovalConfig struct {
	ApprovalRequestID string `json:"approval_request_id"` // required
	Resolution        string `json:"resolution"`           // required: "approved" | "rejected"
	Reason            string `json:"reason,omitempty"`     // optional audit trail
}

// ResolveApprovalHandler programmatically resolves an open approval request.
type ResolveApprovalHandler struct {
	log                *logger.Logger
	approvalRequestBus *approvalrequestbus.Business
}

// NewResolveApprovalHandler constructs a new resolve approval handler.
func NewResolveApprovalHandler(log *logger.Logger, approvalRequestBus *approvalrequestbus.Business) *ResolveApprovalHandler {
	return &ResolveApprovalHandler{log: log, approvalRequestBus: approvalRequestBus}
}

// GetType returns the action type identifier.
func (h *ResolveApprovalHandler) GetType() string { return "resolve_approval_request" }

// IsAsync returns true because resolution may complete a Temporal async workflow.
func (h *ResolveApprovalHandler) IsAsync() bool { return true }

// SupportsManualExecution returns true to allow manual invocation via the API.
func (h *ResolveApprovalHandler) SupportsManualExecution() bool { return true }

// GetDescription returns a human-readable description of this handler.
func (h *ResolveApprovalHandler) GetDescription() string {
	return "Programmatically resolve an open approval request — enables cross-workflow orchestration where one workflow closes an approval another workflow is waiting on"
}

// Validate checks that the config contains a valid approval_request_id and resolution.
func (h *ResolveApprovalHandler) Validate(config json.RawMessage) error {
	var cfg ResolveApprovalConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	if cfg.ApprovalRequestID == "" {
		return fmt.Errorf("approval_request_id is required")
	}
	if _, err := uuid.Parse(cfg.ApprovalRequestID); err != nil {
		return fmt.Errorf("invalid approval_request_id: %w", err)
	}
	if cfg.Resolution != approvalrequestbus.StatusApproved && cfg.Resolution != approvalrequestbus.StatusRejected {
		return fmt.Errorf("resolution must be %q or %q, got %q", approvalrequestbus.StatusApproved, approvalrequestbus.StatusRejected, cfg.Resolution)
	}
	return nil
}

// GetOutputPorts returns the set of named output ports this handler can produce.
func (h *ResolveApprovalHandler) GetOutputPorts() []workflow.OutputPort {
	return []workflow.OutputPort{
		{Name: "resolved_approved", Description: "Approval request was resolved as approved"},
		{Name: "resolved_rejected", Description: "Approval request was resolved as rejected"},
		{Name: "not_found", Description: "Approval request not found"},
		{Name: "already_resolved", Description: "Approval request was already resolved"},
		{Name: "failure", Description: "Unexpected error"},
	}
}

// GetEntityModifications declares which entities this handler modifies.
func (h *ResolveApprovalHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
	return []workflow.EntityModification{
		{EntityName: "workflow.approval_requests", EventType: "on_update", Fields: []string{"status"}},
	}
}

// Execute resolves the approval request identified by approval_request_id.
func (h *ResolveApprovalHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
	if h.approvalRequestBus == nil {
		return map[string]any{"output": "failure", "error": "approval request bus not available"}, nil
	}

	var cfg ResolveApprovalConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return map[string]any{"output": "failure", "error": err.Error()}, nil
	}

	id, _ := uuid.Parse(cfg.ApprovalRequestID)

	req, err := h.approvalRequestBus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, approvalrequestbus.ErrNotFound) {
			return map[string]any{"output": "not_found", "approval_request_id": cfg.ApprovalRequestID}, nil
		}
		return nil, fmt.Errorf("query approval request: %w", err)
	}

	if req.Status != approvalrequestbus.StatusPending {
		return map[string]any{
			"output":              "already_resolved",
			"approval_request_id": cfg.ApprovalRequestID,
			"current_status":      req.Status,
		}, nil
	}

	resolved, err := h.approvalRequestBus.Resolve(ctx, id, execCtx.UserID, cfg.Resolution, cfg.Reason)
	if err != nil {
		if errors.Is(err, approvalrequestbus.ErrAlreadyResolved) {
			return map[string]any{"output": "already_resolved", "approval_request_id": cfg.ApprovalRequestID}, nil
		}
		return nil, fmt.Errorf("resolve approval request: %w", err)
	}

	outputPort := "resolved_approved"
	if cfg.Resolution == approvalrequestbus.StatusRejected {
		outputPort = "resolved_rejected"
	}

	return map[string]any{
		"output":              outputPort,
		"approval_request_id": resolved.ID.String(),
		"resolution":          cfg.Resolution,
	}, nil
}
