package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/outbox"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/protected"
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
	protected    *protected.Registry
	delegate     *delegate.Delegate
	entityMap    map[string]EntityRef
	outbox       *outbox.Writer
}

// NewTransitionStatusHandler creates a new transition status handler.
func NewTransitionStatusHandler(log *logger.Logger, db *sqlx.DB, opts ...Option) *TransitionStatusHandler {
	o := newOptions(opts)
	return &TransitionStatusHandler{
		log:          log,
		db:           db,
		templateProc: workflow.NewTemplateProcessor(workflow.DefaultTemplateProcessingOptions()),
		protected:    o.protected,
		delegate:     o.delegate,
		entityMap:    o.entityMap,
		outbox:       o.outbox,
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

	if !IsValidColumnName(cfg.StatusField) {
		return fmt.Errorf("invalid status_field: %s", cfg.StatusField)
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

	// transition_status is a generic handler: reject transitions of an invariant
	// (protected) status field — those belong to a typed action-verb (DESIGN §10).
	if err := checkProtectedField(h.protected, cfg.TargetEntity, cfg.StatusField); err != nil {
		return nil, err
	}

	templateContext := buildTemplateContext(execContext)

	// Resolve template values
	targetID := processTemplateValue(h.templateProc, ctx, h.log, cfg.TargetID, templateContext)
	toStatus := processTemplateValue(h.templateProc, ctx, h.log, cfg.ToStatus, templateContext)

	// Wrap the read-validate-update and the cascade-event emit in one transaction so
	// the outbox row commits or rolls back atomically with the status UPDATE (F2 Path
	// C). ctx carries the tx so fireSynthesizedEvent's Emit lands on it, and the
	// read-then-write sees a consistent snapshot. The invalid-transition / not-found
	// paths return early and the deferred rollback discards the (read-only) tx.
	tx, err := h.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("transition: begin tx: %w", err)
	}
	defer tx.Rollback()
	ctx = sqldb.WithTx(ctx, tx)

	// Read current status
	selectQuery := fmt.Sprintf("SELECT %s AS val FROM %s WHERE id = :target_id", cfg.StatusField, cfg.TargetEntity)
	var statusDest struct {
		Val string `db:"val"`
	}
	err = sqldb.NamedQueryStruct(ctx, h.log, tx, selectQuery, map[string]any{"target_id": targetID}, &statusDest)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return nil, fmt.Errorf("entity not found: %s with id %v", cfg.TargetEntity, targetID)
		}
		return nil, fmt.Errorf("failed to read current status: %w", err)
	}
	currentStatus := statusDest.Val

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

	rowsAffected, err := sqldb.NamedExecContextWithCount(ctx, h.log, tx, updateQuery, args)
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

	// Cascade (P4 M1): announce the transition (on_update) so it triggers any downstream
	// rule whose trigger matches — including value-aware changed_to, since we carry both
	// the prior status (currentStatus) and the new status (toStatus). Reached only on a
	// real write: the invalid-transition and zero-row paths return earlier.
	var entityID uuid.UUID
	if parsed, perr := uuid.Parse(fmt.Sprintf("%v", targetID)); perr == nil {
		entityID = parsed
	}
	if err := fireSynthesizedEvent(ctx, h.log, h.delegate, h.outbox, h.entityMap, cfg.TargetEntity, workflow.ActionUpdated,
		workflow.DelegateEventParams{
			EntityID:     entityID,
			UserID:       execContext.UserID,
			Entity:       map[string]any{"id": entityID, cfg.StatusField: toStatus},
			BeforeEntity: map[string]any{"id": entityID, cfg.StatusField: currentStatus},
		}); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("transition: commit: %w", err)
	}

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

	// to_status is the produced value; statically known unless it is templated ("{{...}}").
	change := workflow.ProducedChange{FieldName: cfg.StatusField, Operator: workflow.OperatorChangedTo}
	if cfg.ToStatus == "" || strings.Contains(cfg.ToStatus, "{{") {
		change.Indeterminate = true
	} else {
		change.Value = cfg.ToStatus
	}

	return []workflow.EntityModification{{
		EntityName: cfg.TargetEntity,
		EventType:  "on_update",
		Fields:     []string{cfg.StatusField},
		Changes:    []workflow.ProducedChange{change},
	}}
}
