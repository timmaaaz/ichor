package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// LookupEntityConfig represents configuration for entity lookups.
type LookupEntityConfig struct {
	Entity    string            `json:"entity"`
	Filter    map[string]string `json:"filter"`
	Fields    []string          `json:"fields"`
	OutputKey string            `json:"output_key"`
}

// LookupEntityHandler handles lookup_entity actions.
type LookupEntityHandler struct {
	log          *logger.Logger
	db           *sqlx.DB
	templateProc *workflow.TemplateProcessor
}

// NewLookupEntityHandler creates a new lookup entity handler.
func NewLookupEntityHandler(log *logger.Logger, db *sqlx.DB) *LookupEntityHandler {
	return &LookupEntityHandler{
		log:          log,
		db:           db,
		templateProc: workflow.NewTemplateProcessor(workflow.DefaultTemplateProcessingOptions()),
	}
}

// GetType returns the action type.
func (h *LookupEntityHandler) GetType() string {
	return "lookup_entity"
}

// SupportsManualExecution returns true - lookups can be run manually.
func (h *LookupEntityHandler) SupportsManualExecution() bool {
	return true
}

// IsAsync returns false - lookups complete inline.
func (h *LookupEntityHandler) IsAsync() bool {
	return false
}

// GetDescription returns a human-readable description.
func (h *LookupEntityHandler) GetDescription() string {
	return "Look up entity data by filter criteria for use in downstream actions"
}

// Validate validates the lookup entity configuration.
func (h *LookupEntityHandler) Validate(config json.RawMessage) error {
	var cfg LookupEntityConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if cfg.Entity == "" {
		return errors.New("entity is required")
	}

	if !IsValidTableName(cfg.Entity) {
		return fmt.Errorf("invalid entity: %s", cfg.Entity)
	}

	if len(cfg.Filter) == 0 {
		return errors.New("filter is required and must not be empty")
	}

	if len(cfg.Fields) == 0 {
		return errors.New("fields is required and must not be empty")
	}

	if cfg.OutputKey == "" {
		return errors.New("output_key is required")
	}

	return nil
}

// GetOutputPorts implements workflow.OutputPortProvider.
func (h *LookupEntityHandler) GetOutputPorts() []workflow.OutputPort {
	return []workflow.OutputPort{
		{Name: "found", Description: "Entity found matching filter criteria", IsDefault: true},
		{Name: "not_found", Description: "No entity found matching filter criteria"},
		{Name: "failure", Description: "Lookup query failed"},
	}
}

// Execute executes the entity lookup.
func (h *LookupEntityHandler) Execute(ctx context.Context, config json.RawMessage, execContext workflow.ActionExecutionContext) (any, error) {
	var cfg LookupEntityConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	templateContext := buildTemplateContext(execContext)

	// Build SELECT query
	fields := strings.Join(cfg.Fields, ", ")
	query := fmt.Sprintf("SELECT %s FROM %s", fields, cfg.Entity)

	// Build WHERE clause from filter
	args := make(map[string]any)
	whereClauses := make([]string, 0, len(cfg.Filter))
	i := 0
	for col, val := range cfg.Filter {
		argName := fmt.Sprintf("filter_%d", i)
		resolvedVal := processTemplateValue(h.templateProc, ctx, h.log, val, templateContext)
		whereClauses = append(whereClauses, fmt.Sprintf("%s = :%s", col, argName))
		args[argName] = resolvedVal
		i++
	}
	query += " WHERE " + strings.Join(whereClauses, " AND ")

	// Execute query
	rows, err := func() (*sqlx.Rows, error) {
		namedQuery, qArgs, err := sqlx.Named(query, args)
		if err != nil {
			return nil, err
		}
		namedQuery = h.db.Rebind(namedQuery)
		return h.db.QueryxContext(ctx, namedQuery, qArgs...)
	}()
	if err != nil {
		return nil, fmt.Errorf("lookup query failed: %w", err)
	}
	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		row := make(map[string]any)
		if err := rows.MapScan(row); err != nil {
			return nil, fmt.Errorf("scanning lookup result: %w", err)
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating lookup results: %w", err)
	}

	// Build result
	result := map[string]any{}

	switch len(results) {
	case 0:
		result["found"] = false
		result[cfg.OutputKey] = nil
		result["output"] = "not_found"
	case 1:
		result["found"] = true
		result[cfg.OutputKey] = results[0]
		result["output"] = "found"
	default:
		result["found"] = true
		result[cfg.OutputKey] = results
		result[cfg.OutputKey+"_count"] = len(results)
		result["output"] = "found"
	}

	h.log.Info(ctx, "lookup_entity completed",
		"entity", cfg.Entity,
		"output_key", cfg.OutputKey,
		"found", result["found"],
		"result_count", len(results))

	return result, nil
}

// =============================================================================
// Shared Template Helpers
// =============================================================================

// buildTemplateContext creates template context from execution context.
// Shared across data action handlers.
func buildTemplateContext(execContext workflow.ActionExecutionContext) workflow.TemplateContext {
	context := make(workflow.TemplateContext)

	context["entity_id"] = execContext.EntityID
	context["entity_name"] = execContext.EntityName
	context["event_type"] = execContext.EventType
	context["timestamp"] = execContext.Timestamp
	context["user_id"] = execContext.UserID
	if execContext.RuleID != nil {
		context["rule_id"] = *execContext.RuleID
	}
	context["rule_name"] = execContext.RuleName
	context["execution_id"] = execContext.ExecutionID

	if execContext.RawData != nil {
		for k, v := range execContext.RawData {
			context[k] = v
		}
	}

	if execContext.FieldChanges != nil {
		for fieldName, change := range execContext.FieldChanges {
			context["old_"+fieldName] = change.OldValue
			context["new_"+fieldName] = change.NewValue
		}
	}

	return context
}

// processTemplateValue processes template variables in a string value.
// Shared across data action handlers.
func processTemplateValue(proc *workflow.TemplateProcessor, ctx context.Context, log *logger.Logger, value any, templateContext workflow.TemplateContext) any {
	strVal, ok := value.(string)
	if !ok {
		return value
	}

	result := proc.ProcessTemplate(strVal, templateContext)
	if len(result.Errors) > 0 {
		log.Error(ctx, "Template processing errors", "errors", result.Errors)
		return value
	}

	return result.Processed
}
