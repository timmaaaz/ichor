package workflowdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
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
	INSERT INTO workflow.trigger_types (
		id, name, description, is_active
	) VALUES (
		:id, :name, :description, :is_active
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
		workflow.trigger_types
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
		workflow.trigger_types
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
		workflow.trigger_types
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
		id, name, description, is_active
	FROM
		workflow.trigger_types`

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

// QueryTriggerTypeByName retrieves a single trigger type by name from the database.
func (s *Store) QueryTriggerTypeByName(ctx context.Context, name string) (workflow.TriggerType, error) {
	data := struct {
		Name string `db:"name"`
	}{
		Name: name,
	}

	const q = `
	SELECT
		id, name, description, is_active
	FROM
		workflow.trigger_types
	WHERE
		name = :name`

	var dbTriggerType triggerType

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbTriggerType); err != nil {
		return workflow.TriggerType{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toCoreTriggerType(dbTriggerType), nil
}

// =============================================================================
// Entity Types

// CreateEntityType inserts a new entity type into the database.
func (s *Store) CreateEntityType(ctx context.Context, et workflow.EntityType) error {
	const q = `
	INSERT INTO workflow.entity_types (
		id, name, description, is_active
	) VALUES (
		:id, :name, :description, :is_active
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
		workflow.entity_types
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
		workflow.entity_types
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
		workflow.entity_types
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
		id, name, description, is_active
	FROM
		workflow.entity_types`

	var dbEntityTypes []entityType
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, struct{}{}, &dbEntityTypes); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreEntityTypeSlice(dbEntityTypes), nil
}

// QueryEntityTypeByName retrieves a single entity type by name from the database.
func (s *Store) QueryEntityTypeByName(ctx context.Context, name string) (workflow.EntityType, error) {
	data := struct {
		Name string `db:"name"`
	}{
		Name: name,
	}

	const q = `
	SELECT
		id, name, description, is_active
	FROM
		workflow.entity_types
	WHERE
		name = :name`

	var dbEntityType entityType

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbEntityType); err != nil {
		return workflow.EntityType{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toCoreEntityType(dbEntityType), nil
}

// =============================================================================
// Entities

// CreateEntity inserts a new entity into the database.
func (s *Store) CreateEntity(ctx context.Context, entity workflow.Entity) error {
	const q = `
	INSERT INTO workflow.entities (
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
		workflow.entities
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
		workflow.entities
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
		workflow.entities
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

// QueryEntities retrieves a list of existing workflow.entities from the database.
func (s *Store) QueryEntities(ctx context.Context) ([]workflow.Entity, error) {
	const q = `
	SELECT
		id, name, entity_type_id, schema_name, is_active, created_date
	FROM
		workflow.entities`

	var dbEntities []entity
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, struct{}{}, &dbEntities); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}
	return toCoreEntitySlice(dbEntities), nil
}

// QueryEntityByName retrieves a single entity by name from the database.
func (s *Store) QueryEntityByName(ctx context.Context, name string) (workflow.Entity, error) {
	data := struct {
		Name string `db:"name"`
	}{
		Name: name,
	}

	const q = `
	SELECT
		id, name, entity_type_id, schema_name, is_active, created_date
	FROM
		workflow.entities
	WHERE
		name = :name`

	var dbEntity entity

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbEntity); err != nil {
		return workflow.Entity{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toCoreEntity(dbEntity), nil
}

// QueryEntitiesByType retrieves a list of entities for the specified entity type from the database.
func (s *Store) QueryEntitiesByType(ctx context.Context, entityTypeID uuid.UUID) ([]workflow.Entity, error) {
	data := struct {
		EntityTypeID string `db:"entity_type_id"`
	}{
		EntityTypeID: entityTypeID.String(),
	}

	const q = `
	SELECT
		id, name, entity_type_id, schema_name, is_active, created_date
	FROM
		workflow.entities
	WHERE
		entity_type_id = :entity_type_id`

	var dbEntities []entity

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbEntities); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreEntitySlice(dbEntities), nil
}

// =============================================================================
// Automation Rules

// CreateRule inserts a new automation rule into the database.
func (s *Store) CreateRule(ctx context.Context, rule workflow.AutomationRule) error {
	const q = `
	INSERT INTO workflow.automation_rules (
		id, name, description, entity_id, entity_type_id, trigger_type_id,
		trigger_conditions, canvas_layout, is_active, created_date, updated_date,
		created_by, updated_by
	) VALUES (
		:id, :name, :description, :entity_id, :entity_type_id, :trigger_type_id,
		:trigger_conditions, :canvas_layout, :is_active, :created_date, :updated_date,
		:created_by, :updated_by
	)`

	dbAR, err := toDBAutomationRule(rule)
	if err != nil {
		return fmt.Errorf("toDBAutomationRule: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbAR); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// UpdateRule replaces an automation rule document in the database.
func (s *Store) UpdateRule(ctx context.Context, rule workflow.AutomationRule) error {
	const q = `
	UPDATE
		workflow.automation_rules
	SET
		name = :name,
		description = :description,
		entity_id = :entity_id,
		entity_type_id = :entity_type_id,
		trigger_type_id = :trigger_type_id,
		trigger_conditions = :trigger_conditions,
		canvas_layout = :canvas_layout,
		is_active = :is_active,
		updated_date = :updated_date,
		updated_by = :updated_by
	WHERE
		id = :id`

	dbAR, err := toDBAutomationRule(rule)
	if err != nil {
		return fmt.Errorf("toDBAutomationRule: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbAR); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// DeactivateRule deactivates an automation rule in the database.
func (s *Store) DeactivateRule(ctx context.Context, rule workflow.AutomationRule) error {
	const q = `
	UPDATE
		workflow.automation_rules
	SET
		deactivated_by = :deactivated_by,
		is_active = false
	WHERE
		id = :id`

	dbAR, err := toDBAutomationRule(rule)
	if err != nil {
		return fmt.Errorf("toDBAutomationRule: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbAR); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// ActivateRule activates an automation rule in the database.
func (s *Store) ActivateRule(ctx context.Context, rule workflow.AutomationRule) error {
	const q = `
	UPDATE
		workflow.automation_rules
	SET
		deactivated_by = NULL,
		is_active = true
	WHERE
		id = :id`

	dbAR, err := toDBAutomationRule(rule)
	if err != nil {
		return fmt.Errorf("toDBAutomationRule: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbAR); err != nil {
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
		trigger_conditions, canvas_layout, is_active, created_date, updated_date,
		created_by, updated_by
	FROM
		workflow.automation_rules
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
		trigger_conditions, canvas_layout, is_active, created_date, updated_date,
		created_by, updated_by
	FROM
		workflow.automation_rules
	WHERE
		entity_id = :entity_id`

	var dbRules []automationRule
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbRules); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreAutomationRuleSlice(dbRules), nil
}

// QueryActiveRules gets all active automation rules from the database.
func (s *Store) QueryActiveRules(ctx context.Context) ([]workflow.AutomationRule, error) {
	const q = `
	SELECT
		id, name, description, entity_id, entity_type_id, trigger_type_id,
		trigger_conditions, canvas_layout, is_active, created_date, updated_date,
		created_by, updated_by
	FROM
		workflow.automation_rules
	WHERE
		is_active = true`

	var dbRules []automationRule

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, struct{}{}, &dbRules); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}
	return toCoreAutomationRuleSlice(dbRules), nil
}

// =============================================================================
// Rule Actions

// CreateRuleAction inserts a new rule action into the database.
func (s *Store) CreateRuleAction(ctx context.Context, action workflow.RuleAction) error {
	const q = `
	INSERT INTO workflow.rule_actions (
		id, automation_rules_id, name, description, action_config,
		is_active, template_id, deactivated_by
	) VALUES (
		:id, :automation_rules_id, :name, :description, :action_config,
		:is_active, :template_id, :deactivated_by
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
		workflow.rule_actions
	SET 
		automation_rules_id = :automation_rules_id,
		name = :name,
		description = :description,
		action_config = :action_config,
		is_active = :is_active,
		template_id = :template_id
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBRuleAction(action)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// DeactivateRuleAction sets a rule action's is_active flag to false.
func (s *Store) DeactivateRuleAction(ctx context.Context, action workflow.RuleAction) error {
	const q = `
	UPDATE
		workflow.rule_actions
	SET
		is_active = false,
		deactivated_by = NULL
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBRuleAction(action)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// ActivateRuleAction sets a rule action's is_active flag to true.
func (s *Store) ActivateRuleAction(ctx context.Context, action workflow.RuleAction) error {
	const q = `
	UPDATE
		workflow.rule_actions
	SET
		is_active = true,
		deactivated_by = NULL
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
		is_active, template_id
	FROM
		workflow.rule_actions
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
	INSERT INTO workflow.rule_dependencies (
		id, parent_rule_id, child_rule_id
	) VALUES (
		:id, :parent_rule_id, :child_rule_id
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
		workflow.rule_dependencies
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBRuleDependency(dep)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// QueryDependencies retrieves a list of existing rule dependencies from the database.
func (s *Store) QueryDependencies(ctx context.Context) ([]workflow.RuleDependency, error) {
	const q = `
	SELECT
		id, parent_rule_id, child_rule_id
	FROM
		workflow.rule_dependencies`

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
	INSERT INTO workflow.action_templates (
		id, name, description, action_type, icon, default_config,
		created_date, created_by
	) VALUES (
		:id, :name, :description, :action_type, :icon, :default_config,
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
		workflow.action_templates
	SET
		name = :name,
		description = :description,
		action_type = :action_type,
		icon = :icon,
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
		workflow.action_templates
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
		workflow.action_templates
	SET
		is_active = true,
		deactivated_by = NULL
	WHERE
		id = :id`

	data := struct {
		ID string `db:"id"`
	}{
		ID: templateID.String(),
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
		id, name, description, action_type, icon, default_config,
		created_date, created_by, is_active, deactivated_by
	FROM
		workflow.action_templates
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

// QueryAllTemplates retrieves all action templates from the database.
func (s *Store) QueryAllTemplates(ctx context.Context) ([]workflow.ActionTemplate, error) {
	const q = `
	SELECT
		id, name, description, action_type, icon, default_config,
		created_date, created_by, is_active, deactivated_by
	FROM
		workflow.action_templates`

	var dbTemplates []actionTemplate
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, struct{}{}, &dbTemplates); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	templates := make([]workflow.ActionTemplate, len(dbTemplates))
	for i, dbt := range dbTemplates {
		templates[i] = toCoreActionTemplate(dbt)
	}

	return templates, nil
}

// QueryActiveTemplates retrieves only active action templates from the database.
func (s *Store) QueryActiveTemplates(ctx context.Context) ([]workflow.ActionTemplate, error) {
	const q = `
	SELECT
		id, name, description, action_type, icon, default_config,
		created_date, created_by, is_active, deactivated_by
	FROM
		workflow.action_templates
	WHERE
		is_active = true`

	var dbTemplates []actionTemplate
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, struct{}{}, &dbTemplates); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	templates := make([]workflow.ActionTemplate, len(dbTemplates))
	for i, dbt := range dbTemplates {
		templates[i] = toCoreActionTemplate(dbt)
	}

	return templates, nil
}

// =============================================================================
// Automation Executions

// CreateExecution inserts a new automation execution into the database.
func (s *Store) CreateExecution(ctx context.Context, exec workflow.AutomationExecution) error {
	const q = `
	INSERT INTO workflow.automation_executions (
		id, automation_rules_id, entity_type, trigger_data, actions_executed,
		status, error_message, execution_time_ms, executed_at,
		trigger_source, executed_by, action_type
	) VALUES (
		:id, :automation_rules_id, :entity_type, :trigger_data, :actions_executed,
		:status, :error_message, :execution_time_ms, :executed_at,
		:trigger_source, :executed_by, :action_type
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
		status, error_message, execution_time_ms, executed_at,
		trigger_source, executed_by, action_type
	FROM
		workflow.automation_executions
	WHERE
		automation_rules_id = :automation_rules_id
	LIMIT :limit`

	var dbExecutions []automationExecution
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbExecutions); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreAutomationExecutionSlice(dbExecutions), nil
}

// CreateNotificationDelivery inserts a notification delivery record
func (s *Store) CreateNotificationDelivery(ctx context.Context, delivery workflow.NotificationDelivery) error {
	const q = `
	INSERT INTO workflow.notification_deliveries (
		id, notification_id, automation_execution_id, rule_id, action_id,
		recipient_id, channel, status, attempts,
		sent_at, delivered_at, failed_at, error_message,
		provider_response, created_date, updated_date
	) VALUES (
		:id, :notification_id, :automation_execution_id, :rule_id, :action_id,
		:recipient_id, :channel, :status, :attempts,
		:sent_at, :delivered_at, :failed_at, :error_message,
		:provider_response, :created_date, :updated_date
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBNotificationDelivery(delivery)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// UpdateNotificationDelivery updates a delivery record (for retries, status changes)
func (s *Store) UpdateNotificationDelivery(ctx context.Context, delivery workflow.NotificationDelivery) error {
	const q = `
	UPDATE workflow.notification_deliveries 
	SET 
		status = :status,
		attempts = :attempts,
		delivered_at = :delivered_at,
		failed_at = :failed_at,
		error_message = :error_message,
		provider_response = :provider_response,
		updated_date = :updated_date
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBNotificationDelivery(delivery)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// QueryDeliveriesByAutomationExecution gets notification deliveries for the specified automation execution from the database.
func (s *Store) QueryDeliveriesByAutomationExecution(ctx context.Context, executionID uuid.UUID) ([]workflow.NotificationDelivery, error) {
	data := struct {
		ExecutionID string `db:"automation_execution_id"`
	}{
		ExecutionID: executionID.String(),
	}

	const q = `
	SELECT
		id, notification_id, automation_execution_id, rule_id, action_id,
		recipient_id, channel, status, attempts,
		sent_at, delivered_at, failed_at, error_message,
		provider_response, created_date, updated_date
	FROM
		workflow.notification_deliveries
	WHERE
		automation_execution_id = :automation_execution_id`

	var dbDeliveries []notificationDelivery
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbDeliveries); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreNotificationDeliverySlice(dbDeliveries), nil
}

// QueryAllDeliveries gets all notification deliveries from the database.
func (s *Store) QueryAllDeliveries(ctx context.Context) ([]workflow.NotificationDelivery, error) {
	const q = `
	SELECT
		id, notification_id, automation_execution_id, rule_id, action_id,
		recipient_id, channel, status, attempts,
		sent_at, delivered_at, failed_at, error_message,
		provider_response, created_date, updated_date
	FROM
		workflow.notification_deliveries`

	var dbDeliveries []notificationDelivery

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, struct{}{}, &dbDeliveries); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreNotificationDeliverySlice(dbDeliveries), nil
}

// =============================================================================
// Allocation Results

func (s *Store) CreateAllocationResult(ctx context.Context, result workflow.AllocationResult) error {
	const q = `
	INSERT INTO 
		workflow.allocation_results (id, idempotency_key, allocation_data, created_date)
	VALUES 
		(:id, :idempotency_key, :allocation_data, :created_date)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAllocationResult(result)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) QueryAllocationResultByIdempotencyKey(ctx context.Context, idempotencyKey string) (workflow.AllocationResult, workflow.IdempotencyResult, error) {
	data := struct {
		IdempotencyKey string `db:"idempotency_key"`
	}{
		IdempotencyKey: idempotencyKey,
	}

	const q = `	
	SELECT
		id, idempotency_key, allocation_data, created_date
	FROM
		workflow.allocation_results
	WHERE 
		idempotency_key = :idempotency_key`

	var dbResult allocationResult
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbResult); err != nil {
		// NOTE: Because we are checking for idempotency, the ErrNotFound is the
		// case we WANT, there is a problem if we find a record here.
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return workflow.AllocationResult{}, workflow.IdempotencyNotFound, nil
		}
		return workflow.AllocationResult{}, 0, fmt.Errorf("db: %w", err)
	}

	return toCoreAllocationResult(dbResult), workflow.IdempotencyExists, nil
}

// =============================================================================
// VIEWS
// =============================================================================
func (s *Store) QueryAutomationRulesView(ctx context.Context) ([]workflow.AutomationRuleView, error) {
	const q = `
	SELECT
		id, name, description, trigger_conditions, canvas_layout, is_active, created_date, updated_date,
		created_by, updated_by, trigger_type_id, trigger_type_name, trigger_type_description,
		entity_type_id, entity_type_name, entity_type_description, entity_name, entity_id
	FROM
		workflow.automation_rules_view
	WHERE
		is_active = true`

	var dbAutomationRulesViews []automationRulesView
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, struct{}{}, &dbAutomationRulesViews); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreAutomationRuleViews(dbAutomationRulesViews), nil
}

func (s *Store) QueryRoleActionsViewByRuleID(ctx context.Context, ruleID uuid.UUID) ([]workflow.RuleActionView, error) {
	data := struct {
		RuleID string `db:"automation_rules_id"`
	}{
		RuleID: ruleID.String(),
	}

	const q = `
	SELECT
		id, automation_rules_id, name, description, action_config,
		is_active, template_id, template_name,
		template_action_type, template_default_config
	FROM
		workflow.rule_actions_view
	WHERE
		automation_rules_id = :automation_rules_id`

	var dbRuleActionsViews []ruleActionView
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbRuleActionsViews); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreRuleActionViews(dbRuleActionsViews), nil
}

// QueryAutomationRulesViewPaginated retrieves rules with pagination, filtering, and ordering.
func (s *Store) QueryAutomationRulesViewPaginated(
	ctx context.Context,
	filter workflow.AutomationRuleFilter,
	orderBy order.By,
	pg page.Page,
) ([]workflow.AutomationRuleView, error) {
	data := map[string]any{
		"offset":        (pg.Number() - 1) * pg.RowsPerPage(),
		"rows_per_page": pg.RowsPerPage(),
	}

	const baseQuery = `
	SELECT
		ar.id,
		ar.name,
		ar.description,
		ar.entity_id,
		ar.trigger_conditions,
		ar.canvas_layout,
		ar.is_active,
		ar.created_date,
		ar.updated_date,
		ar.created_by,
		ar.updated_by,
		ar.trigger_type_id,
		COALESCE(tt.name, '') AS trigger_type_name,
		COALESCE(tt.description, '') AS trigger_type_description,
		ar.entity_type_id,
		COALESCE(et.name, '') AS entity_type_name,
		COALESCE(et.description, '') AS entity_type_description,
		COALESCE(e.name, '') AS entity_name,
		COALESCE(e.schema_name, '') AS entity_schema_name,
		COALESCE((
			SELECT json_agg(json_build_object(
				'id', ra.id,
				'name', ra.name,
				'description', ra.description,
				'action_config', ra.action_config,
				'is_active', ra.is_active,
				'template_id', ra.template_id,
				'deactivated_by', ra.deactivated_by
			) ORDER BY ra.name)
			FROM workflow.rule_actions ra
			WHERE ra.automation_rules_id = ar.id
		), '[]'::::json) AS actions
	FROM
		workflow.automation_rules ar
	LEFT JOIN
		workflow.trigger_types tt ON ar.trigger_type_id = tt.id
	LEFT JOIN
		workflow.entity_types et ON ar.entity_type_id = et.id
	LEFT JOIN
		workflow.entities e ON ar.entity_id = e.id`

	buf := bytes.NewBufferString(baseQuery)

	applyAutomationRuleFilter(filter, data, buf)

	orderByClause, err := orderByClauseAutomationRule(orderBy)
	if err != nil {
		return nil, fmt.Errorf("orderby: %w", err)
	}
	buf.WriteString(orderByClause)

	buf.WriteString(" LIMIT :rows_per_page OFFSET :offset")

	var dbRules []automationRulesView
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbRules); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreAutomationRuleViews(dbRules), nil
}

// CountAutomationRulesView counts rules matching the filter.
func (s *Store) CountAutomationRulesView(
	ctx context.Context,
	filter workflow.AutomationRuleFilter,
) (int, error) {
	data := map[string]any{}

	const baseQuery = `
	SELECT COUNT(ar.id) AS count
	FROM workflow.automation_rules ar`

	buf := bytes.NewBufferString(baseQuery)

	applyAutomationRuleFilter(filter, data, buf)

	var result struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &result); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return result.Count, nil
}

// QueryActionByID retrieves a single action by its ID.
func (s *Store) QueryActionByID(ctx context.Context, actionID uuid.UUID) (workflow.RuleAction, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: actionID.String(),
	}

	const q = `
	SELECT
		id, automation_rules_id, name, description, action_config,
		is_active, template_id
	FROM workflow.rule_actions
	WHERE id = :id`

	var dbAction ruleAction
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbAction); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return workflow.RuleAction{}, workflow.ErrNotFound
		}
		return workflow.RuleAction{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toCoreRuleAction(dbAction), nil
}

// QueryActionViewByID retrieves a single action view by its ID (with template info).
func (s *Store) QueryActionViewByID(ctx context.Context, actionID uuid.UUID) (workflow.RuleActionView, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: actionID.String(),
	}

	const q = `
	SELECT
		id, automation_rules_id, name, description, action_config,
		is_active, template_id, template_name,
		template_action_type, template_default_config
	FROM workflow.rule_actions_view
	WHERE id = :id`

	var dbActionView ruleActionView
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbActionView); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return workflow.RuleActionView{}, workflow.ErrNotFound
		}
		return workflow.RuleActionView{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toCoreRuleActionView(dbActionView), nil
}

// =============================================================================
// Execution Paginated Queries

// QueryExecutionsPaginated retrieves executions with pagination, filtering, and ordering.
func (s *Store) QueryExecutionsPaginated(
	ctx context.Context,
	filter workflow.ExecutionFilter,
	orderBy order.By,
	pg page.Page,
) ([]workflow.AutomationExecution, error) {
	data := map[string]any{
		"offset":        (pg.Number() - 1) * pg.RowsPerPage(),
		"rows_per_page": pg.RowsPerPage(),
	}

	const baseQuery = `
	SELECT
		ae.id,
		ae.automation_rules_id,
		ar.name AS rule_name,
		ae.entity_type,
		ae.trigger_data,
		ae.actions_executed,
		ae.status,
		ae.error_message,
		ae.execution_time_ms,
		ae.executed_at,
		ae.trigger_source,
		ae.executed_by,
		ae.action_type
	FROM
		workflow.automation_executions ae
	LEFT JOIN
		workflow.automation_rules ar ON ae.automation_rules_id = ar.id`

	buf := bytes.NewBufferString(baseQuery)

	applyExecutionFilter(filter, data, buf)

	orderByClause, err := orderByClauseExecution(orderBy)
	if err != nil {
		return nil, fmt.Errorf("orderby: %w", err)
	}
	buf.WriteString(orderByClause)

	buf.WriteString(" LIMIT :rows_per_page OFFSET :offset")

	var dbExecutions []automationExecution
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbExecutions); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreAutomationExecutionSlice(dbExecutions), nil
}

// CountExecutions counts executions matching the filter.
func (s *Store) CountExecutions(
	ctx context.Context,
	filter workflow.ExecutionFilter,
) (int, error) {
	data := map[string]any{}

	const baseQuery = `
	SELECT COUNT(ae.id) AS count
	FROM workflow.automation_executions ae`

	buf := bytes.NewBufferString(baseQuery)

	applyExecutionFilter(filter, data, buf)

	var result struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &result); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return result.Count, nil
}

// QueryExecutionByID retrieves a single execution by its ID.
func (s *Store) QueryExecutionByID(ctx context.Context, id uuid.UUID) (workflow.AutomationExecution, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: id.String(),
	}

	const q = `
	SELECT
		ae.id,
		ae.automation_rules_id,
		ar.name AS rule_name,
		ae.entity_type,
		ae.trigger_data,
		ae.actions_executed,
		ae.status,
		ae.error_message,
		ae.execution_time_ms,
		ae.executed_at,
		ae.trigger_source,
		ae.executed_by,
		ae.action_type
	FROM workflow.automation_executions ae
	LEFT JOIN workflow.automation_rules ar ON ae.automation_rules_id = ar.id
	WHERE ae.id = :id`

	var dbExecution automationExecution
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbExecution); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return workflow.AutomationExecution{}, workflow.ErrNotFound
		}
		return workflow.AutomationExecution{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toCoreAutomationExecution(dbExecution), nil
}

// =============================================================================
// Action Edges (for workflow branching/condition nodes)

// CreateActionEdge inserts a new action edge into the database.
func (s *Store) CreateActionEdge(ctx context.Context, edge workflow.NewActionEdge) (workflow.ActionEdge, error) {
	dbEdge := toDBNewActionEdge(edge)

	const q = `
	INSERT INTO workflow.action_edges (
		id, rule_id, source_action_id, target_action_id, edge_type, source_output, edge_order, created_date
	) VALUES (
		:id, :rule_id, :source_action_id, :target_action_id, :edge_type, :source_output, :edge_order, :created_date
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbEdge); err != nil {
		return workflow.ActionEdge{}, fmt.Errorf("namedexeccontext: %w", err)
	}

	return toCoreActionEdge(dbEdge), nil
}

// QueryEdgesByRuleID returns all edges for a rule, ordered by edge_order.
func (s *Store) QueryEdgesByRuleID(ctx context.Context, ruleID uuid.UUID) ([]workflow.ActionEdge, error) {
	data := struct {
		RuleID string `db:"rule_id"`
	}{
		RuleID: ruleID.String(),
	}

	const q = `
	SELECT
		id, rule_id, source_action_id, target_action_id, edge_type, source_output, edge_order, created_date
	FROM workflow.action_edges
	WHERE rule_id = :rule_id
	ORDER BY edge_order ASC`

	var dbEdges []actionEdge
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbEdges); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreActionEdgeSlice(dbEdges), nil
}

// QueryEdgeByID retrieves a single edge by its ID.
func (s *Store) QueryEdgeByID(ctx context.Context, edgeID uuid.UUID) (workflow.ActionEdge, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: edgeID.String(),
	}

	const q = `
	SELECT
		id, rule_id, source_action_id, target_action_id, edge_type, source_output, edge_order, created_date
	FROM workflow.action_edges
	WHERE id = :id`

	var dbEdge actionEdge
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbEdge); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return workflow.ActionEdge{}, workflow.ErrNotFound
		}
		return workflow.ActionEdge{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toCoreActionEdge(dbEdge), nil
}

// DeleteActionEdge deletes an edge by ID.
func (s *Store) DeleteActionEdge(ctx context.Context, edgeID uuid.UUID) error {
	data := struct {
		ID string `db:"id"`
	}{
		ID: edgeID.String(),
	}

	const q = `DELETE FROM workflow.action_edges WHERE id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// DeleteEdgesByRuleID deletes all edges for a rule.
func (s *Store) DeleteEdgesByRuleID(ctx context.Context, ruleID uuid.UUID) error {
	data := struct {
		RuleID string `db:"rule_id"`
	}{
		RuleID: ruleID.String(),
	}

	const q = `DELETE FROM workflow.action_edges WHERE rule_id = :rule_id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}
