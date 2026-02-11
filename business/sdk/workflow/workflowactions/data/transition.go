package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// TransitionStatusConfig represents configuration for status transitions.
type TransitionStatusConfig struct {
	TargetEntity      string   `json:"target_entity"`
	TargetID          string   `json:"target_id"`
	StatusField       string   `json:"status_field"`
	ToStatus          string   `json:"to_status"`
	ValidFromStatuses []string `json:"valid_from_statuses"`
}

// TransitionStatusHandler handles transition_status actions.
type TransitionStatusHandler struct {
	log          *logger.Logger
	db           *sqlx.DB
	templateProc *workflow.TemplateProcessor
}

// NewTransitionStatusHandler creates a new transition status handler.
func NewTransitionStatusHandler(log *logger.Logger, db *sqlx.DB) *TransitionStatusHandler {
	return &TransitionStatusHandler{
		log:          log,
		db:           db,
		templateProc: workflow.NewTemplateProcessor(workflow.DefaultTemplateProcessingOptions()),
	}
}

// GetType returns the action type.
func (h *TransitionStatusHandler) GetType() string {
	return "transition_status"
}

// SupportsManualExecution returns true - transitions can be run manually.
func (h *TransitionStatusHandler) SupportsManualExecution() bool {
	return true
}

// IsAsync returns false - transitions complete inline.
func (h *TransitionStatusHandler) IsAsync() bool {
	return false
}

// GetDescription returns a human-readable description.
func (h *TransitionStatusHandler) GetDescription() string {
	return "Transition an entity field from one status to another with validation"
}

// Validate validates the transition status configuration.
func (h *TransitionStatusHandler) Validate(config json.RawMessage) error {
	var cfg TransitionStatusConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if cfg.TargetEntity == "" {
		return errors.New("target_entity is required")
	}

	if !IsValidTableName(cfg.TargetEntity) {
		return fmt.Errorf("invalid target_entity: %s", cfg.TargetEntity)
	}

	if cfg.StatusField == "" {
		return errors.New("status_field is required")
	}

	if cfg.ToStatus == "" {
		return errors.New("to_status is required")
	}

	if len(cfg.ValidFromStatuses) == 0 {
		return errors.New("valid_from_statuses is required and must not be empty")
	}

	return nil
}

// GetOutputPorts implements workflow.OutputPortProvider.
func (h *TransitionStatusHandler) GetOutputPorts() []workflow.OutputPort {
	return []workflow.OutputPort{
		{Name: "success", Description: "Status transition completed", IsDefault: true},
		{Name: "invalid_transition", Description: "Current status not in allowed from-statuses"},
		{Name: "failure", Description: "Transition failed due to error"},
	}
}

// Execute executes the status transition.
func (h *TransitionStatusHandler) Execute(ctx context.Context, config json.RawMessage, execContext workflow.ActionExecutionContext) (any, error) {
	var cfg TransitionStatusConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	templateContext := buildTemplateContext(execContext)

	// Resolve template values
	targetID := processTemplateValue(h.templateProc, ctx, h.log, cfg.TargetID, templateContext)
	toStatus := processTemplateValue(h.templateProc, ctx, h.log, cfg.ToStatus, templateContext)

	// Read current status
	selectQuery := fmt.Sprintf("SELECT %s FROM %s WHERE id = $1", cfg.StatusField, cfg.TargetEntity)
	var currentStatus string
	err := h.db.GetContext(ctx, &currentStatus, selectQuery, targetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("entity not found: %s with id %v", cfg.TargetEntity, targetID)
		}
		return nil, fmt.Errorf("failed to read current status: %w", err)
	}

	// Validate transition
	validFrom := false
	for _, allowed := range cfg.ValidFromStatuses {
		if currentStatus == allowed {
			validFrom = true
			break
		}
	}

	if !validFrom {
		return map[string]any{
			"transitioned": false,
			"from_status":  currentStatus,
			"to_status":    toStatus,
			"output":       "invalid_transition",
		}, nil
	}

	// Execute update
	updateQuery := fmt.Sprintf("UPDATE %s SET %s = :to_status WHERE id = :target_id", cfg.TargetEntity, cfg.StatusField)
	args := map[string]any{
		"to_status": toStatus,
		"target_id": targetID,
	}

	rowsAffected, err := sqldb.NamedExecContextWithCount(ctx, h.log, h.db, updateQuery, args)
	if err != nil {
		return nil, fmt.Errorf("transition update failed: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("no rows updated for %s with id %v", cfg.TargetEntity, targetID)
	}

	h.log.Info(ctx, "transition_status completed",
		"entity", cfg.TargetEntity,
		"target_id", targetID,
		"from", currentStatus,
		"to", toStatus)

	return map[string]any{
		"transitioned": true,
		"from_status":  currentStatus,
		"to_status":    toStatus,
		"output":       "success",
	}, nil
}

// GetEntityModifications implements workflow.EntityModifier for cascade visualization.
func (h *TransitionStatusHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
	var cfg TransitionStatusConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil
	}

	return []workflow.EntityModification{{
		EntityName: cfg.TargetEntity,
		EventType:  "on_update",
		Fields:     []string{cfg.StatusField},
	}}
}
