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
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
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
}

// NewCreateEntityHandler creates a new create entity handler.
func NewCreateEntityHandler(log *logger.Logger, db *sqlx.DB) *CreateEntityHandler {
	return &CreateEntityHandler{
		log:          log,
		db:           db,
		templateProc: workflow.NewTemplateProcessor(workflow.DefaultTemplateProcessingOptions()),
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

	if _, err := sqldb.NamedExecContextWithCount(ctx, h.log, h.db, query, processedFields); err != nil {
		return nil, fmt.Errorf("entity creation failed: %w", err)
	}

	createdID := processedFields["id"]

	result := map[string]any{
		"created_id":    createdID,
		"target_entity": cfg.TargetEntity,
		"status":        "success",
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
