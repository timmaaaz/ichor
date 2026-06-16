package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
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

// CreateEntityConfig represents configuration for entity creation.
type CreateEntityConfig struct {
	TargetEntity string         `json:"target_entity"`
	Fields       map[string]any `json:"fields"`
}

// CreateEntityHandler handles create_entity actions.
type CreateEntityHandler struct {
	log          *logger.Logger
	db           *sqlx.DB
	templateProc *workflow.TemplateProcessor
	protected    *protected.Registry
	delegate     *delegate.Delegate
	entityMap    map[string]EntityRef
	outbox       *outbox.Writer
}

// NewCreateEntityHandler creates a new create entity handler.
func NewCreateEntityHandler(log *logger.Logger, db *sqlx.DB, opts ...Option) *CreateEntityHandler {
	o := newOptions(opts)
	return &CreateEntityHandler{
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
func (h *CreateEntityHandler) GetType() string {
	return "create_entity"
}

// SupportsManualExecution returns true - entity creation can be run manually.
func (h *CreateEntityHandler) SupportsManualExecution() bool {
	return true
}

// IsAsync returns false - entity creation completes inline.
func (h *CreateEntityHandler) IsAsync() bool {
	return false
}

// GetDescription returns a human-readable description.
func (h *CreateEntityHandler) GetDescription() string {
	return "Create a new entity record in the database"
}

// Validate validates the create entity configuration.
func (h *CreateEntityHandler) Validate(config json.RawMessage) error {
	var cfg CreateEntityConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if cfg.TargetEntity == "" {
		return errors.New("target_entity is required")
	}

	if !IsValidTableName(cfg.TargetEntity) {
		return fmt.Errorf("invalid target_entity: %s", cfg.TargetEntity)
	}

	if len(cfg.Fields) == 0 {
		return errors.New("fields is required and must not be empty")
	}

	return nil
}

// Execute executes the entity creation.
func (h *CreateEntityHandler) Execute(ctx context.Context, config json.RawMessage, execContext workflow.ActionExecutionContext) (any, error) {
	var cfg CreateEntityConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Reject creates into a whole-table-protected entity (e.g. an append-only ledger)
	// or that set any protected field directly (DESIGN §10 protected-list).
	fieldNames := make([]string, 0, len(cfg.Fields))
	for k := range cfg.Fields {
		fieldNames = append(fieldNames, k)
	}
	if err := checkProtectedEntity(h.protected, cfg.TargetEntity, fieldNames); err != nil {
		return nil, err
	}

	templateContext := buildTemplateContext(execContext)

	// Process template variables in all field values
	processedFields := make(map[string]any, len(cfg.Fields))
	for k, v := range cfg.Fields {
		processedFields[k] = processTemplateValue(h.templateProc, ctx, h.log, v, templateContext)
	}

	// Auto-generate ID if not provided
	if _, hasID := processedFields["id"]; !hasID {
		processedFields["id"] = uuid.New()
	}

	// Build INSERT query with sorted columns for deterministic output
	columns := make([]string, 0, len(processedFields))
	for col := range processedFields {
		columns = append(columns, col)
	}
	sort.Strings(columns)

	placeholders := make([]string, len(columns))
	for i, col := range columns {
		placeholders[i] = ":" + col
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		cfg.TargetEntity,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	// Wrap the write and the cascade-event emit in one transaction so the outbox row
	// commits or rolls back atomically with the INSERT (F2 Path C). ctx carries the tx
	// so fireSynthesizedEvent's Emit lands on it (read-your-writes correct).
	tx, err := h.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("entity creation: begin tx: %w", err)
	}
	defer tx.Rollback()
	ctx = sqldb.WithTx(ctx, tx)

	if _, err := sqldb.NamedExecContextWithCount(ctx, h.log, tx, query, processedFields); err != nil {
		return nil, fmt.Errorf("entity creation failed: %w", err)
	}

	createdID := processedFields["id"]

	result := map[string]any{
		"created_id":    createdID,
		"target_entity": cfg.TargetEntity,
		"status":        "success",
	}

	// Cascade (P4 M1): announce the new record (on_create) so it triggers any downstream
	// rule whose trigger matches. INSERT always wrote one row, so this always fires. The
	// full created record rides Entity for downstream field access.
	var entityID uuid.UUID
	switch v := createdID.(type) {
	case uuid.UUID:
		entityID = v
	case string:
		if parsed, perr := uuid.Parse(v); perr == nil {
			entityID = parsed
		}
	}
	if err := fireSynthesizedEvent(ctx, h.log, h.delegate, h.outbox, h.entityMap, cfg.TargetEntity, workflow.ActionCreated,
		workflow.DelegateEventParams{
			EntityID: entityID,
			UserID:   execContext.UserID,
			Entity:   processedFields,
		}); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("entity creation: commit: %w", err)
	}

	h.log.Info(ctx, "create_entity completed",
		"entity", cfg.TargetEntity,
		"created_id", createdID)

	return result, nil
}

// GetEntityModifications implements workflow.EntityModifier for cascade visualization.
func (h *CreateEntityHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
	var cfg CreateEntityConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil
	}

	return []workflow.EntityModification{{
		EntityName: cfg.TargetEntity,
		EventType:  "on_create",
	}}
}
