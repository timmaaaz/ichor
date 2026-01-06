package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// UpdateFieldConfig represents configuration for field updates
type UpdateFieldConfig struct {
	TargetEntity     string                    `json:"target_entity"`
	TargetField      string                    `json:"target_field"`
	NewValue         any                       `json:"new_value"`
	FieldType        string                    `json:"field_type,omitempty"`
	ForeignKeyConfig *ForeignKeyConfig         `json:"foreign_key_config,omitempty"`
	Conditions       []workflow.FieldCondition `json:"conditions,omitempty"`
	TimeoutMs        int                       `json:"timeout_ms,omitempty"`
}

// ForeignKeyConfig handles foreign key resolution
type ForeignKeyConfig struct {
	ReferenceTable string         `json:"reference_table"`
	LookupField    string         `json:"lookup_field"`
	IDField        string         `json:"id_field,omitempty"`
	AllowCreate    bool           `json:"allow_create,omitempty"`
	CreateConfig   map[string]any `json:"create_config,omitempty"`
}

// UpdateFieldHandler handles update_field actions
type UpdateFieldHandler struct {
	log          *logger.Logger
	db           *sqlx.DB
	templateProc *workflow.TemplateProcessor
}

// NewUpdateFieldHandler creates a new update field handler
func NewUpdateFieldHandler(log *logger.Logger, db *sqlx.DB) *UpdateFieldHandler {
	return &UpdateFieldHandler{
		log:          log,
		db:           db,
		templateProc: workflow.NewTemplateProcessor(workflow.DefaultTemplateProcessingOptions()),
	}
}

// GetType returns the action type
func (h *UpdateFieldHandler) GetType() string {
	return "update_field"
}

// Validate validates the update field configuration
func (h *UpdateFieldHandler) Validate(config json.RawMessage) error {
	var cfg UpdateFieldConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if cfg.TargetEntity == "" {
		return errors.New("target_entity is required")
	}

	if cfg.TargetField == "" {
		return errors.New("target_field is required")
	}

	if cfg.NewValue == nil {
		return errors.New("new_value is required")
	}

	// Validate table name against whitelist
	if !h.isValidTableName(cfg.TargetEntity) {
		return fmt.Errorf("invalid target_entity: %s", cfg.TargetEntity)
	}

	// Validate foreign key config if present
	if cfg.FieldType == "foreign_key" {
		if cfg.ForeignKeyConfig == nil {
			return errors.New("foreign_key_config is required when field_type is foreign_key")
		}
		if cfg.ForeignKeyConfig.ReferenceTable == "" {
			return errors.New("foreign_key_config.reference_table is required")
		}
		if cfg.ForeignKeyConfig.LookupField == "" {
			return errors.New("foreign_key_config.lookup_field is required")
		}
		if !h.isValidTableName(cfg.ForeignKeyConfig.ReferenceTable) {
			return fmt.Errorf("invalid reference_table: %s", cfg.ForeignKeyConfig.ReferenceTable)
		}
	}

	// Validate conditions
	for i, condition := range cfg.Conditions {
		if condition.FieldName == "" {
			return fmt.Errorf("condition %d: field_name is required", i)
		}
		if condition.Operator == "" {
			return fmt.Errorf("condition %d: operator is required", i)
		}
		if !h.isValidOperator(condition.Operator) {
			return fmt.Errorf("condition %d: invalid operator %s", i, condition.Operator)
		}
	}

	return nil
}

// Execute executes the field update
func (h *UpdateFieldHandler) Execute(ctx context.Context, config json.RawMessage, execContext workflow.ActionExecutionContext) (any, error) {
	startTime := time.Now()
	updateID := fmt.Sprintf("update_%d_%s", time.Now().Unix(), uuid.New().String()[:8])

	// Parse configuration
	var cfg UpdateFieldConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Initialize result
	result := map[string]any{
		"update_id":        updateID,
		"status":           "failed",
		"target_entity":    cfg.TargetEntity,
		"target_field":     cfg.TargetField,
		"new_value":        cfg.NewValue,
		"records_affected": int64(0),
		"warnings":         []string{},
		"created_at":       startTime,
	}

	// Process template variables
	templateContext := h.buildTemplateContext(execContext)
	processedValue := h.processTemplateValue(ctx, cfg.NewValue, templateContext)
	resolvedValue := processedValue

	// Handle foreign key resolution
	if cfg.FieldType == "foreign_key" && cfg.ForeignKeyConfig != nil {
		var err error
		resolvedValue, err = h.resolveForeignKey(ctx, processedValue, cfg.ForeignKeyConfig, templateContext)
		if err != nil {
			result["error_message"] = fmt.Sprintf("foreign key resolution failed: %v", err)
			result["completed_at"] = time.Now()
			result["execution_time_ms"] = time.Since(startTime).Milliseconds()
			return result, err
		}
		if resolvedValue != processedValue {
			result["resolved_value"] = resolvedValue
		}
	}

	// Execute update - use h.db directly, no transaction support for now
	recordsAffected, err := h.executeUpdate(ctx, h.db, cfg, resolvedValue, templateContext)
	if err != nil {
		result["error_message"] = err.Error()
		result["completed_at"] = time.Now()
		result["execution_time_ms"] = time.Since(startTime).Milliseconds()
		return result, err
	}

	// Success
	result["status"] = "success"
	result["records_affected"] = recordsAffected
	result["completed_at"] = time.Now()
	result["execution_time_ms"] = time.Since(startTime).Milliseconds()

	// If single record update, set the target record ID
	if len(cfg.Conditions) == 1 && cfg.Conditions[0].FieldName == "id" {
		if id, err := uuid.Parse(fmt.Sprintf("%v", cfg.Conditions[0].Value)); err == nil {
			result["target_record_id"] = id
		}
	}

	h.log.Info(ctx, "Field update completed",
		"updateID", updateID,
		"entity", cfg.TargetEntity,
		"field", cfg.TargetField,
		"recordsAffected", recordsAffected,
		"duration", result["execution_time_ms"])

	return result, nil
}

// executeUpdate performs the actual database update
func (h *UpdateFieldHandler) executeUpdate(ctx context.Context, execer sqlx.ExtContext, cfg UpdateFieldConfig, value any, templateContext workflow.TemplateContext) (int64, error) {
	// Build UPDATE query
	query := fmt.Sprintf("UPDATE %s SET %s = :value", cfg.TargetEntity, cfg.TargetField)
	args := map[string]any{
		"value": value,
	}

	// Add WHERE conditions
	if len(cfg.Conditions) > 0 {
		whereClauses := make([]string, 0, len(cfg.Conditions))
		for i, condition := range cfg.Conditions {
			processedValue := h.processTemplateValue(ctx, condition.Value, templateContext)
			whereClause, argName := h.buildWhereClause(condition, processedValue, i)
			whereClauses = append(whereClauses, whereClause)
			if condition.Operator != "is_null" && condition.Operator != "is_not_null" {
				args[argName] = processedValue
			}
		}
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	} else if entityID := templateContext["entity_id"]; entityID != nil {
		// Default to updating the triggering entity
		query += " WHERE id = :entity_id"
		args["entity_id"] = entityID
	} else {
		return 0, errors.New("no conditions specified and no entity_id in context")
	}

	// Execute query
	rowsAffected, err := sqldb.NamedExecContextWithCount(ctx, h.log, execer, query, args)
	if err != nil {
		return 0, fmt.Errorf("update failed: %w, query: %s, args: %+v", err, query, args)
	}

	return rowsAffected, nil
}

// buildWhereClause builds a WHERE clause for a condition
func (h *UpdateFieldHandler) buildWhereClause(condition workflow.FieldCondition, value any, index int) (string, string) {
	argName := fmt.Sprintf("cond_%d", index)

	switch condition.Operator {
	case "equals":
		return fmt.Sprintf("%s = :%s", condition.FieldName, argName), argName
	case "not_equals":
		return fmt.Sprintf("%s != :%s", condition.FieldName, argName), argName
	case "greater_than":
		return fmt.Sprintf("%s > :%s", condition.FieldName, argName), argName
	case "less_than":
		return fmt.Sprintf("%s < :%s", condition.FieldName, argName), argName
	case "contains":
		return fmt.Sprintf("%s LIKE '%%' || :%s || '%%'", condition.FieldName, argName), argName
	case "is_null":
		return fmt.Sprintf("%s IS NULL", condition.FieldName), argName
	case "is_not_null":
		return fmt.Sprintf("%s IS NOT NULL", condition.FieldName), argName
	case "in":
		return fmt.Sprintf("%s = ANY(:%s)", condition.FieldName, argName), argName
	default:
		return fmt.Sprintf("%s = :%s", condition.FieldName, argName), argName
	}
}

// resolveForeignKey resolves a foreign key value
func (h *UpdateFieldHandler) resolveForeignKey(ctx context.Context, value any, fkConfig *ForeignKeyConfig, templateContext workflow.TemplateContext) (any, error) {
	// If value is already a UUID, validate it exists
	if strVal, ok := value.(string); ok {
		if id, err := uuid.Parse(strVal); err == nil {
			idField := fkConfig.IDField
			if idField == "" {
				idField = "id"
			}

			query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", idField, fkConfig.ReferenceTable, idField)
			var exists uuid.UUID
			err := h.db.GetContext(ctx, &exists, query, id)
			if err == nil {
				return id, nil
			}
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("failed to validate foreign key: %w", err)
			}
		}
	}

	// Lookup by display value
	idField := fkConfig.IDField
	if idField == "" {
		idField = "id"
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", idField, fkConfig.ReferenceTable, fkConfig.LookupField)
	var resolvedID uuid.UUID
	err := h.db.GetContext(ctx, &resolvedID, query, value)

	if err == nil {
		return resolvedID, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("lookup query failed: %w", err)
	}

	// Record not found - create if allowed
	if fkConfig.AllowCreate && fkConfig.CreateConfig != nil {
		createData := make(map[string]any)
		for k, v := range fkConfig.CreateConfig {
			createData[k] = v
		}
		createData[fkConfig.LookupField] = value
		createData["id"] = uuid.New()

		columns := make([]string, 0, len(createData))
		placeholders := make([]string, 0, len(createData))
		for col := range createData {
			columns = append(columns, col)
			placeholders = append(placeholders, ":"+col)
		}

		insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id",
			fkConfig.ReferenceTable,
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "))

		var newID uuid.UUID
		err = sqldb.NamedQueryStruct(ctx, h.log, h.db, insertQuery, createData, &newID)
		if err != nil {
			return nil, fmt.Errorf("failed to create referenced record: %w", err)
		}

		return newID, nil
	}

	return nil, fmt.Errorf("referenced record not found: %v in %s.%s", value, fkConfig.ReferenceTable, fkConfig.LookupField)
}

// processTemplateValue processes template variables in a value
func (h *UpdateFieldHandler) processTemplateValue(ctx context.Context, value any, templateContext workflow.TemplateContext) any {
	strVal, ok := value.(string)
	if !ok {
		return value
	}

	result := h.templateProc.ProcessTemplate(strVal, templateContext)
	if len(result.Errors) > 0 {
		h.log.Error(ctx, "Template processing errors", "errors", result.Errors)
		return value
	}

	return result.Processed
}

// buildTemplateContext creates template context from execution context
func (h *UpdateFieldHandler) buildTemplateContext(execContext workflow.ActionExecutionContext) workflow.TemplateContext {
	context := make(workflow.TemplateContext)

	context["entity_id"] = execContext.EntityID
	context["entity_name"] = execContext.EntityName
	context["event_type"] = execContext.EventType
	context["timestamp"] = execContext.Timestamp
	context["user_id"] = execContext.UserID
	context["rule_id"] = execContext.RuleID
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

// isValidTableName validates table names against whitelist
func (h *UpdateFieldHandler) isValidTableName(tableName string) bool {
	validTables := []string{
		// sales schema
		"sales.customers", "sales.orders", "sales.order_line_items",
		"sales.order_fulfillment_statuses", "sales.line_item_fulfillment_statuses",
		// products schema
		"products.products", "products.brands", "products.product_categories",
		"products.physical_attributes", "products.product_costs", "products.cost_history",
		"products.quality_metrics",
		// inventory schema
		"inventory.inventory_items", "inventory.inventory_locations", "inventory.inventory_transactions",
		"inventory.warehouses", "inventory.zones", "inventory.lot_trackings",
		"inventory.serial_numbers", "inventory.inspections", "inventory.inventory_adjustments",
		"inventory.transfer_orders",
		// procurement schema
		"procurement.suppliers", "procurement.supplier_products",
		// core schema
		"core.users", "core.roles", "core.user_roles", "core.contact_infos", "core.table_access",
		// hr schema
		"hr.offices",
		// geography schema
		"geography.countries", "geography.regions", "geography.cities", "geography.streets",
		// assets schema
		"assets.assets", "assets.valid_assets",
		// config schema
		"config.table_configs",
		// workflow schema
		"workflow.automation_rules", "workflow.rule_actions", "workflow.action_templates",
		"workflow.rule_dependencies", "workflow.trigger_types", "workflow.entity_types",
		"workflow.entities", "workflow.automation_executions", "workflow.notification_deliveries",
	}

	for _, valid := range validTables {
		if tableName == valid {
			return true
		}
	}
	return false
}

// isValidOperator validates condition operators
func (h *UpdateFieldHandler) isValidOperator(operator string) bool {
	validOperators := []string{
		"equals", "not_equals", "greater_than", "less_than",
		"contains", "is_null", "is_not_null", "in", "not_in",
	}

	for _, valid := range validOperators {
		if operator == valid {
			return true
		}
	}
	return false
}

// // Context returns a context for template processing
// func (tc workflow.TemplateContext) Context() context.Context {
// 	return context.Background()
// }
