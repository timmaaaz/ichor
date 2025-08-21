package workflowdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for user database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (workflow.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// // stores/workflowdb/workflowdb.go
// // Trigger Types
// func (s *Store) CreateTriggerType(ctx context.Context, tt workflow.TriggerType) error
// func (s *Store) UpdateTriggerType(ctx context.Context, id string, tt workflow.TriggerType) error
// func (s *Store) DeleteTriggerType(ctx context.Context, id string) error
// func (s *Store) QueryTriggerTypes(ctx context.Context) ([]workflow.TriggerType, error)

// // Entity Types
// func (s *Store) CreateEntityType(ctx context.Context, et workflow.EntityType) error
// func (s *Store) UpdateEntityType(ctx context.Context, id string, et workflow.EntityType) error
// func (s *Store) DeleteEntityType(ctx context.Context, id string) error
// func (s *Store) QueryEntityTypes(ctx context.Context) ([]workflow.EntityType, error)

// // Rules
// func (s *Store) CreateRule(ctx context.Context, rule workflow.AutomationRule) error
// func (s *Store) UpdateRule(ctx context.Context, id string, rule workflow.AutomationRule) error
// func (s *Store) DeleteRule(ctx context.Context, id string) error
// func (s *Store) QueryRuleByID(ctx context.Context, id string) (workflow.AutomationRule, error)
// func (s *Store) QueryRulesByEntity(ctx context.Context, entityID string) ([]workflow.AutomationRule, error)

// // Actions
// func (s *Store) CreateRuleAction(ctx context.Context, action workflow.RuleAction) error
// func (s *Store) UpdateRuleAction(ctx context.Context, id string, action workflow.RuleAction) error
// func (s *Store) DeleteRuleAction(ctx context.Context, id string) error
// func (s *Store) QueryActionsByRule(ctx context.Context, ruleID string) ([]workflow.RuleAction, error)

// // Dependencies
// func (s *Store) CreateDependency(ctx context.Context, dep workflow.RuleDependency) error
// func (s *Store) DeleteDependency(ctx context.Context, parentID, childID string) error
// func (s *Store) QueryDependencies(ctx context.Context) ([]workflow.RuleDependency, error)

// // Templates
// func (s *Store) CreateActionTemplate(ctx context.Context, template workflow.ActionTemplate) error
// func (s *Store) UpdateActionTemplate(ctx context.Context, id string, template workflow.ActionTemplate) error
// func (s *Store) DeleteActionTemplate(ctx context.Context, id string) error
// func (s *Store) QueryTemplateByID(ctx context.Context, id string) (workflow.ActionTemplate, error)

// // Entities
// func (s *Store) CreateEntity(ctx context.Context, entity workflow.Entity) error
// func (s *Store) UpdateEntity(ctx context.Context, id string, entity workflow.Entity) error
// func (s *Store) DeleteEntity(ctx context.Context, id string) error
// func (s *Store) QueryEntities(ctx context.Context) ([]workflow.Entity, error)

// // Executions (mostly create/read)
// func (s *Store) CreateExecution(ctx context.Context, exec workflow.AutomationExecution) error
// func (s *Store) QueryExecutionHistory(ctx context.Context, ruleID string, limit int) ([]workflow.AutomationExecution, error)

// Package workflowdb contains workflow automation related CRUD functionality.

// =============================================================================
// Trigger Types

// CreateTriggerType inserts a new trigger type into the database.
func (s *Store) CreateTriggerType(ctx context.Context, tt workflow.TriggerType) error {
	const q = `
	INSERT INTO trigger_types (
		id, name, description
	) VALUES (
		:id, :name, :description
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBTriggerType(tt)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// UpdateTriggerType replaces a trigger type document in the database.
func (s *Store) UpdateTriggerType(ctx context.Context, tt workflow.TriggerType) error {
	const q = `
	UPDATE
		trigger_types
	SET 
		name = :name,
		description = :description
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBTriggerType(tt)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// DeactivateTriggerType deactivates a trigger type in the database.
func (s *Store) DeactivateTriggerType(ctx context.Context, tt workflow.TriggerType) error {
	const q = `
	UPDATE
		trigger_types
	SET
		deactivated_by = :deactivated_by,
		is_active = false
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBTriggerType(tt)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// ActivateTriggerType reactivates a trigger type in the database.
func (s *Store) ActivateTriggerType(ctx context.Context, tt workflow.TriggerType) error {
	const q = `
	UPDATE
		trigger_types
	SET
		deactivated_by = NULL,
		is_active = true
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBTriggerType(tt)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// QueryTriggerTypes retrieves a list of existing trigger types from the database.
func (s *Store) QueryTriggerTypes(ctx context.Context) ([]workflow.TriggerType, error) {
	const q = `
	SELECT
		id, name, description
	FROM
		trigger_types`

	var dbTriggerTypes []triggerType
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, struct{}{}, &dbTriggerTypes); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	triggerTypes := make([]workflow.TriggerType, len(dbTriggerTypes))
	for i, dbTT := range dbTriggerTypes {
		triggerTypes[i] = toCoreTriggerType(dbTT)
	}

	return triggerTypes, nil
}

// =============================================================================
// Entity Types

// CreateEntityType inserts a new entity type into the database.
func (s *Store) CreateEntityType(ctx context.Context, et workflow.EntityType) error {
	const q = `
	INSERT INTO entity_types (
		id, name, description
	) VALUES (
		:id, :name, :description
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBEntityType(et)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// UpdateEntityType replaces an entity type document in the database.
func (s *Store) UpdateEntityType(ctx context.Context, et workflow.EntityType) error {
	const q = `
	UPDATE
		entity_types
	SET 
		name = :name,
		description = :description
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBEntityType(et)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// DeactivateEntityType deactivates an entity type from the database.
func (s *Store) DeactivateEntityType(ctx context.Context, et workflow.EntityType) error {
	const q = `
	UPDATE
		entity_types
	SET
		deactivated_by = :deactivated_by,
		is_active = false
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBEntityType(et)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// ActivateEntityType reactivates an entity type in the database.
func (s *Store) ActivateEntityType(ctx context.Context, et workflow.EntityType) error {
	const q = `
	UPDATE
		entity_types
	SET
		deactivated_by = NULL,
		is_active = true
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBEntityType(et)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// QueryEntityTypes retrieves a list of existing entity types from the database.
func (s *Store) QueryEntityTypes(ctx context.Context) ([]workflow.EntityType, error) {
	const q = `
	SELECT
		id, name, description
	FROM
		entity_types`

	var dbEntityTypes []entityType
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, struct{}{}, &dbEntityTypes); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreEntityTypeSlice(dbEntityTypes), nil
}

// =============================================================================
// Entities

// CreateEntity inserts a new entity into the database.
func (s *Store) CreateEntity(ctx context.Context, entity workflow.Entity) error {
	const q = `
	INSERT INTO entities (
		id, name, entity_type_id, schema_name, is_active, created_date
	) VALUES (
		:id, :name, :entity_type_id, :schema_name, :is_active, :created_date
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBEntity(entity)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// UpdateEntity replaces an entity document in the database.
func (s *Store) UpdateEntity(ctx context.Context, entity workflow.Entity) error {
	const q = `
	UPDATE
		entities
	SET 
		name = :name,
		entity_type_id = :entity_type_id,
		schema_name = :schema_name,
		is_active = :is_active
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBEntity(entity)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// DeactivateEntity deactivates an entity in the database.
func (s *Store) DeactivateEntity(ctx context.Context, entity workflow.Entity) error {
	const q = `
	UPDATE
		entities
	SET
		deactivated_by = :deactivated_by,
		is_active = false
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBEntity(entity)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// ActivateEntity reactivates an entity in the database.
func (s *Store) ActivateEntity(ctx context.Context, entity workflow.Entity) error {
	const q = `
	UPDATE
		entities
	SET
		deactivated_by = NULL,
		is_active = true
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBEntity(entity)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// QueryEntities retrieves a list of existing entities from the database.
func (s *Store) QueryEntities(ctx context.Context) ([]workflow.Entity, error) {
	const q = `
	SELECT
		id, name, entity_type_id, schema_name, is_active, created_date
	FROM
		entities`

	var dbEntities []entity
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, struct{}{}, &dbEntities); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}
	return toCoreEntitySlice(dbEntities), nil
}

// =============================================================================
// Automation Rules

// CreateRule inserts a new automation rule into the database.
func (s *Store) CreateRule(ctx context.Context, rule workflow.AutomationRule) error {
	const q = `
	INSERT INTO automation_rules (
		id, name, description, entity_id, entity_type_id, trigger_type_id,
		trigger_conditions, is_active, created_date, updated_date,
		created_by, updated_by
	) VALUES (
		:id, :name, :description, :entity_id, :entity_type_id, :trigger_type_id,
		:trigger_conditions, :is_active, :created_date, :updated_date,
		:created_by, :updated_by
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAutomationRule(rule)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// UpdateRule replaces an automation rule document in the database.
func (s *Store) UpdateRule(ctx context.Context, rule workflow.AutomationRule) error {
	const q = `
	UPDATE
		automation_rules
	SET 
		name = :name,
		description = :description,
		entity_id = :entity_id,
		entity_type_id = :entity_type_id,
		trigger_type_id = :trigger_type_id,
		trigger_conditions = :trigger_conditions,
		is_active = :is_active,
		updated_date = :updated_date,
		updated_by = :updated_by
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAutomationRule(rule)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// DeactivateRule deactivates an automation rule in the database.
func (s *Store) DeactivateRule(ctx context.Context, rule workflow.AutomationRule) error {
	const q = `
	UPDATE
		automation_rules
	SET
		deactivated_by = :deactivated_by,
		is_active = false
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAutomationRule(rule)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// ActivateRule activates an automation rule in the database.
func (s *Store) ActivateRule(ctx context.Context, rule workflow.AutomationRule) error {
	const q = `
	UPDATE
		automation_rules
	SET
		deactivated_by = NULL,
		is_active = true
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAutomationRule(rule)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// QueryRuleByID gets the specified automation rule from the database.
func (s *Store) QueryRuleByID(ctx context.Context, userID uuid.UUID) (workflow.AutomationRule, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: userID.String(),
	}

	const q = `
	SELECT
		id, name, description, entity_id, entity_type_id, trigger_type_id,
		trigger_conditions, is_active, created_date, updated_date,
		created_by, updated_by
	FROM
		automation_rules
	WHERE 
		id = :id`

	var dbRule automationRule
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbRule); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return workflow.AutomationRule{}, fmt.Errorf("db: %w", workflow.ErrNotFound)
		}
		return workflow.AutomationRule{}, fmt.Errorf("db: %w", err)
	}

	return toCoreAutomationRule(dbRule), nil
}

// QueryRulesByEntity gets automation rules for the specified entity from the database.
func (s *Store) QueryRulesByEntity(ctx context.Context, entityID uuid.UUID) ([]workflow.AutomationRule, error) {
	data := struct {
		EntityID string `db:"entity_id"`
	}{
		EntityID: entityID.String(),
	}

	const q = `
	SELECT
		id, name, description, entity_id, entity_type_id, trigger_type_id,
		trigger_conditions, is_active, created_date, updated_date,
		created_by, updated_by
	FROM
		automation_rules
	WHERE 
		entity_id = :entity_id`

	var dbRules []automationRule
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbRules); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreAutomationRuleSlice(dbRules), nil
}

// =============================================================================
// Rule Actions

// CreateRuleAction inserts a new rule action into the database.
func (s *Store) CreateRuleAction(ctx context.Context, action workflow.RuleAction) error {
	const q = `
	INSERT INTO rule_actions (
		id, automation_rules_id, name, description, action_config,
		execution_order, is_active, template_id
	) VALUES (
		:id, :automation_rules_id, :name, :description, :action_config,
		:execution_order, :is_active, :template_id
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBRuleAction(action)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// UpdateRuleAction replaces a rule action document in the database.
func (s *Store) UpdateRuleAction(ctx context.Context, action workflow.RuleAction) error {
	const q = `
	UPDATE
		rule_actions
	SET 
		automation_rules_id = :automation_rules_id,
		name = :name,
		description = :description,
		action_config = :action_config,
		execution_order = :execution_order,
		is_active = :is_active,
		template_id = :template_id
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBRuleAction(action)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// DeleteRuleAction removes a rule action from the database.
func (s *Store) DeleteRuleAction(ctx context.Context, action workflow.RuleAction) error {
	const q = `
	DELETE FROM
		rule_actions
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBRuleAction(action)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// QueryActionsByRule gets rule actions for the specified automation rule from the database.
func (s *Store) QueryActionsByRule(ctx context.Context, ruleID uuid.UUID) ([]workflow.RuleAction, error) {
	data := struct {
		RuleID string `db:"automation_rules_id"`
	}{
		RuleID: ruleID.String(),
	}

	const q = `
	SELECT
		id, automation_rules_id, name, description, action_config,
		execution_order, is_active, template_id
	FROM
		rule_actions
	WHERE 
		automation_rules_id = :automation_rules_id`

	var dbActions []ruleAction
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbActions); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreRuleActionSlice(dbActions), nil
}

// =============================================================================
// Rule Dependencies

// CreateDependency inserts a new rule dependency into the database.
func (s *Store) CreateDependency(ctx context.Context, dep workflow.RuleDependency) error {
	const q = `
	INSERT INTO rule_dependencies (
		parent_rule_id, child_rule_id
	) VALUES (
		:parent_rule_id, :child_rule_id
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBRuleDependency(dep)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// DeleteDependency removes a rule dependency from the database.
func (s *Store) DeleteDependency(ctx context.Context, dep workflow.RuleDependency) error {
	const q = `
	DELETE FROM
		rule_dependencies
	WHERE
		parent_rule_id = :parent_rule_id AND child_rule_id = :child_rule_id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBRuleDependency(dep)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// QueryDependencies retrieves a list of existing rule dependencies from the database.
func (s *Store) QueryDependencies(ctx context.Context) ([]workflow.RuleDependency, error) {
	const q = `
	SELECT
		parent_rule_id, child_rule_id
	FROM
		rule_dependencies`

	var dbDependencies []ruleDependency
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, struct{}{}, &dbDependencies); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreRuleDependencySlice(dbDependencies), nil
}

// =============================================================================
// Action Templates

// CreateActionTemplate inserts a new action template into the database.
func (s *Store) CreateActionTemplate(ctx context.Context, template workflow.ActionTemplate) error {
	const q = `
	INSERT INTO action_templates (
		id, name, description, action_type, default_config,
		created_date, created_by
	) VALUES (
		:id, :name, :description, :action_type, :default_config,
		:created_date, :created_by
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBActionTemplate(template)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// UpdateActionTemplate replaces an action template document in the database.
func (s *Store) UpdateActionTemplate(ctx context.Context, template workflow.ActionTemplate) error {
	const q = `
	UPDATE
		action_templates
	SET 
		name = :name,
		description = :description,
		action_type = :action_type,
		default_config = :default_config
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBActionTemplate(template)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// DeactivateActionTemplate deactivates an action template in the database.
func (s *Store) DeactivateActionTemplate(ctx context.Context, templateID uuid.UUID, deactivatedBy uuid.UUID) error {
	const q = `
	UPDATE
		action_templates
	SET
		is_active = false,
		deactivated_by = :deactivated_by
	WHERE
		id = :id`

	data := struct {
		ID            string `db:"id"`
		DeactivatedBy string `db:"deactivated_by"`
	}{
		ID:            templateID.String(),
		DeactivatedBy: deactivatedBy.String(),
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// ActivateActionTemplate activates an action template in the database.
func (s *Store) ActivateActionTemplate(ctx context.Context, templateID uuid.UUID, activatedBy uuid.UUID) error {
	const q = `
	UPDATE
		action_templates
	SET
		is_active = true,
		deactivated_by = NULL
	WHERE
		id = :id`

	data := struct {
		ID          string `db:"id"`
		ActivatedBy string `db:"activated_by"`
	}{
		ID:          templateID.String(),
		ActivatedBy: activatedBy.String(),
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// QueryTemplateByID gets the specified action template from the database.
func (s *Store) QueryTemplateByID(ctx context.Context, templateID uuid.UUID) (workflow.ActionTemplate, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: templateID.String(),
	}

	const q = `
	SELECT
		id, name, description, action_type, default_config,
		created_date, created_by
	FROM
		action_templates
	WHERE 
		id = :id`

	var dbTemplate actionTemplate
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbTemplate); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return workflow.ActionTemplate{}, fmt.Errorf("db: %w", workflow.ErrNotFound)
		}
		return workflow.ActionTemplate{}, fmt.Errorf("db: %w", err)
	}

	return toCoreActionTemplate(dbTemplate), nil
}

// =============================================================================
// Automation Executions

// CreateExecution inserts a new automation execution into the database.
func (s *Store) CreateExecution(ctx context.Context, exec workflow.AutomationExecution) error {
	const q = `
	INSERT INTO automation_executions (
		id, automation_rules_id, entity_type, trigger_data, actions_executed,
		status, error_message, execution_time_ms, executed_at
	) VALUES (
		:id, :automation_rules_id, :entity_type, :trigger_data, :actions_executed,
		:status, :error_message, :execution_time_ms, :executed_at
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAutomationExecution(exec)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// QueryExecutionHistory gets execution history for the specified automation rule from the database.
func (s *Store) QueryExecutionHistory(ctx context.Context, ruleID uuid.UUID, limit int) ([]workflow.AutomationExecution, error) {
	data := struct {
		RuleID string `db:"automation_rules_id"`
		Limit  int    `db:"limit"`
	}{
		RuleID: ruleID.String(),
		Limit:  limit,
	}

	const q = `
	SELECT
		id, automation_rules_id, entity_type, trigger_data, actions_executed,
		status, error_message, execution_time_ms, executed_at
	FROM
		automation_executions
	WHERE 
		automation_rules_id = :automation_rules_id
	LIMIT :limit`

	var dbExecutions []automationExecution
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbExecutions); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreAutomationExecutionSlice(dbExecutions), nil
}
