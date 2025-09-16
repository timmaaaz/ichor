// Package workflow manages the CRUD operations for ALL workflow related functions
package workflow

// WORKFLOWBUS FILE

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// TODO: Implement delegate stuff here
// TODO: Make sure all time.now are utc

type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	CreateTriggerType(ctx context.Context, tt TriggerType) error
	UpdateTriggerType(ctx context.Context, tt TriggerType) error
	DeactivateTriggerType(ctx context.Context, tt TriggerType) error
	ActivateTriggerType(ctx context.Context, tt TriggerType) error
	QueryTriggerTypes(ctx context.Context) ([]TriggerType, error)
	QueryTriggerTypeByName(ctx context.Context, name string) (TriggerType, error)

	CreateEntityType(ctx context.Context, et EntityType) error
	UpdateEntityType(ctx context.Context, et EntityType) error
	DeactivateEntityType(ctx context.Context, et EntityType) error
	ActivateEntityType(ctx context.Context, et EntityType) error
	QueryEntityTypes(ctx context.Context) ([]EntityType, error)
	QueryEntityTypeByName(ctx context.Context, name string) (EntityType, error)

	CreateRule(ctx context.Context, rule AutomationRule) error
	UpdateRule(ctx context.Context, rule AutomationRule) error
	QueryRuleByID(ctx context.Context, id uuid.UUID) (AutomationRule, error)
	QueryRulesByEntity(ctx context.Context, entityid uuid.UUID) ([]AutomationRule, error)
	DeactivateRule(ctx context.Context, rule AutomationRule) error
	ActivateRule(ctx context.Context, rule AutomationRule) error
	QueryActiveRules(ctx context.Context) ([]AutomationRule, error)

	CreateRuleAction(ctx context.Context, action RuleAction) error
	UpdateRuleAction(ctx context.Context, action RuleAction) error
	QueryActionsByRule(ctx context.Context, ruleid uuid.UUID) ([]RuleAction, error)
	DeactivateRuleAction(ctx context.Context, action RuleAction) error
	ActivateRuleAction(ctx context.Context, action RuleAction) error

	CreateDependency(ctx context.Context, dep RuleDependency) error
	DeleteDependency(ctx context.Context, dep RuleDependency) error
	QueryDependencies(ctx context.Context) ([]RuleDependency, error)

	CreateActionTemplate(ctx context.Context, template ActionTemplate) error
	UpdateActionTemplate(ctx context.Context, template ActionTemplate) error
	DeactivateActionTemplate(ctx context.Context, templateID uuid.UUID, deactivatedBy uuid.UUID) error
	ActivateActionTemplate(ctx context.Context, templateID uuid.UUID, activatedBy uuid.UUID) error
	QueryTemplateByID(ctx context.Context, templateID uuid.UUID) (ActionTemplate, error)

	CreateEntity(ctx context.Context, entity Entity) error
	UpdateEntity(ctx context.Context, entity Entity) error
	QueryEntities(ctx context.Context) ([]Entity, error)
	DeactivateEntity(ctx context.Context, entity Entity) error
	ActivateEntity(ctx context.Context, entity Entity) error
	QueryEntitiesByType(ctx context.Context, entityTypeID uuid.UUID) ([]Entity, error)
	QueryEntityByName(ctx context.Context, name string) (Entity, error)

	// Notification Delivery Tracking
	CreateNotificationDelivery(ctx context.Context, delivery NotificationDelivery) error
	UpdateNotificationDelivery(ctx context.Context, delivery NotificationDelivery) error
	QueryDeliveriesByAutomationExecution(ctx context.Context, executionID uuid.UUID) ([]NotificationDelivery, error)

	CreateExecution(ctx context.Context, exec AutomationExecution) error
	QueryExecutionHistory(ctx context.Context, ruleid uuid.UUID, limit int) ([]AutomationExecution, error)

	QueryAutomationRulesView(ctx context.Context) ([]AutomationRuleView, error)
}

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("workflow item not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrInvalidDependency     = errors.New("invalid rule dependency")
	ErrCircularDependency    = errors.New("circular dependency detected")
)

// Business manages the set of APIs for workflow automation access.
type Business struct {
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs a workflow business API for use.
func NewBusiness(log *logger.Logger, storer Storer) *Business {
	return &Business{
		log:    log,
		storer: storer,
	}
}

// NewWithTx constructs a new business value that will use the
// specified transaction in any store related calls.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	bus := Business{
		log:    b.log,
		storer: storer,
	}

	return &bus, nil
}

// =============================================================================
// Trigger Types

// CreateTriggerType adds a new trigger type to the system.
func (b *Business) CreateTriggerType(ctx context.Context, ntt NewTriggerType) (TriggerType, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.createtriggertype")
	defer span.End()

	tt := TriggerType{
		ID:          uuid.New(),
		Name:        ntt.Name,
		Description: ntt.Description,
		IsActive:    ntt.IsActive,
	}

	if err := b.storer.CreateTriggerType(ctx, tt); err != nil {
		return TriggerType{}, fmt.Errorf("create: %w", err)
	}

	return tt, nil
}

// UpdateTriggerType modifies information about a trigger type.
func (b *Business) UpdateTriggerType(ctx context.Context, tt TriggerType, utt UpdateTriggerType) (TriggerType, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.updatetriggertype")
	defer span.End()

	if utt.Name != nil {
		tt.Name = *utt.Name
	}
	if utt.Description != nil {
		tt.Description = *utt.Description
	}

	if err := b.storer.UpdateTriggerType(ctx, tt); err != nil {
		return TriggerType{}, fmt.Errorf("update: %w", err)
	}

	return tt, nil
}

// DeactivateTriggerType deactivates a trigger type in the system.
func (b *Business) DeactivateTriggerType(ctx context.Context, tt TriggerType) error {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.deactivatetriggertype")
	defer span.End()

	if err := b.storer.DeactivateTriggerType(ctx, tt); err != nil {
		return fmt.Errorf("deactivate: %w", err)
	}

	return nil
}

// ActivateTriggerType reactivates a trigger type in the system.
func (b *Business) ActivateTriggerType(ctx context.Context, tt TriggerType) error {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.reactivatetriggertype")
	defer span.End()

	if err := b.storer.ActivateTriggerType(ctx, tt); err != nil {
		return fmt.Errorf("reactivate: %w", err)
	}

	return nil
}

// QueryTriggerTypes retrieves a list of existing trigger types.
func (b *Business) QueryTriggerTypes(ctx context.Context) ([]TriggerType, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.querytriggertypes")
	defer span.End()

	triggerTypes, err := b.storer.QueryTriggerTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return triggerTypes, nil
}

// QueryTriggerTypeByName retrieves a trigger type by its name.
func (b *Business) QueryTriggerTypeByName(ctx context.Context, name string) (TriggerType, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.querytriggertypebyname")
	defer span.End()

	tt, err := b.storer.QueryTriggerTypeByName(ctx, name)
	if err != nil {
		return TriggerType{}, fmt.Errorf("query: name[%s]: %w", name, err)
	}

	return tt, nil
}

// =============================================================================
// Entity Types

// CreateEntityType adds a new entity type to the system.
func (b *Business) CreateEntityType(ctx context.Context, net NewEntityType) (EntityType, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.createentitytype")
	defer span.End()

	et := EntityType{
		ID:          uuid.New(),
		Name:        net.Name,
		Description: net.Description,
		IsActive:    net.IsActive,
	}

	if err := b.storer.CreateEntityType(ctx, et); err != nil {
		return EntityType{}, fmt.Errorf("create: %w", err)
	}

	return et, nil
}

// UpdateEntityType modifies information about an entity type.
func (b *Business) UpdateEntityType(ctx context.Context, et EntityType, uet UpdateEntityType) (EntityType, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.updateentitytype")
	defer span.End()

	if uet.Name != nil {
		et.Name = *uet.Name
	}
	if uet.Description != nil {
		et.Description = *uet.Description
	}

	if err := b.storer.UpdateEntityType(ctx, et); err != nil {
		return EntityType{}, fmt.Errorf("update: %w", err)
	}

	return et, nil
}

// DeactivateEntityType deactivates an entity type in the system.
func (b *Business) DeactivateEntityType(ctx context.Context, et EntityType) error {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.deactivateentitytype")
	defer span.End()

	if err := b.storer.DeactivateEntityType(ctx, et); err != nil {
		return fmt.Errorf("deactivate: %w", err)
	}

	return nil
}

// ActivateEntityType reactivates an entity type in the system.
func (b *Business) ActivateEntityType(ctx context.Context, et EntityType) error {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.reactivateentitytype")
	defer span.End()

	if err := b.storer.ActivateEntityType(ctx, et); err != nil {
		return fmt.Errorf("reactivate: %w", err)
	}

	return nil
}

// QueryEntityTypes retrieves a list of existing entity types.
func (b *Business) QueryEntityTypes(ctx context.Context) ([]EntityType, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.queryentitytypes")
	defer span.End()

	entityTypes, err := b.storer.QueryEntityTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return entityTypes, nil
}

// QueryEntityTypeByName retrieves an entity type by its name.
func (b *Business) QueryEntityTypeByName(ctx context.Context, name string) (EntityType, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.queryentitytypebyname")
	defer span.End()

	et, err := b.storer.QueryEntityTypeByName(ctx, name)
	if err != nil {
		return EntityType{}, fmt.Errorf("query: name[%s]: %w", name, err)
	}

	return et, nil
}

// QueryEntitiesByType retrieves a list of entities by their type.
func (b *Business) QueryEntitiesByType(ctx context.Context, entityTypeID uuid.UUID) ([]Entity, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.queryentitiesbytype")
	defer span.End()
	entities, err := b.storer.QueryEntitiesByType(ctx, entityTypeID)
	if err != nil {
		return nil, fmt.Errorf("query: entityTypeID[%s]: %w", entityTypeID, err)
	}
	return entities, nil
}

// QueryEntityByName retrieves an entity by its name.
func (b *Business) QueryEntityByName(ctx context.Context, name string) (Entity, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.queryentitybyname")
	defer span.End()

	entity, err := b.storer.QueryEntityByName(ctx, name)
	if err != nil {
		return Entity{}, fmt.Errorf("query: name[%s]: %w", name, err)
	}
	return entity, nil
}

// =============================================================================
// Entities

// CreateEntity adds a new entity to the system.
func (b *Business) CreateEntity(ctx context.Context, ne NewEntity) (Entity, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.createentity")
	defer span.End()

	now := time.Now().UTC()

	entity := Entity{
		ID:           uuid.New(),
		Name:         ne.Name,
		EntityTypeID: ne.EntityTypeID,
		SchemaName:   ne.SchemaName,
		IsActive:     ne.IsActive,
		CreatedDate:  now,
	}

	if err := b.storer.CreateEntity(ctx, entity); err != nil {
		return Entity{}, fmt.Errorf("create: %w", err)
	}

	return entity, nil
}

// UpdateEntity modifies information about an entity.
func (b *Business) UpdateEntity(ctx context.Context, entity Entity, ue UpdateEntity) (Entity, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.updateentity")
	defer span.End()

	if ue.Name != nil {
		entity.Name = *ue.Name
	}
	if ue.EntityTypeID != nil {
		entity.EntityTypeID = *ue.EntityTypeID
	}
	if ue.SchemaName != nil {
		entity.SchemaName = *ue.SchemaName
	}
	if ue.IsActive != nil {
		entity.IsActive = *ue.IsActive
	}

	if err := b.storer.UpdateEntity(ctx, entity); err != nil {
		return Entity{}, fmt.Errorf("update: %w", err)
	}

	return entity, nil
}

// DeactivateEntity deactivates an entity in the system.
func (b *Business) DeactivateEntity(ctx context.Context, entity Entity) error {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.deactivateentity")
	defer span.End()

	if err := b.storer.DeactivateEntity(ctx, entity); err != nil {
		return fmt.Errorf("deactivate: %w", err)
	}

	return nil
}

// ActivateEntity reactivates an entity in the system.
func (b *Business) ActivateEntity(ctx context.Context, entity Entity) error {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.activateentity")
	defer span.End()

	if err := b.storer.ActivateEntity(ctx, entity); err != nil {
		return fmt.Errorf("activate: %w", err)
	}

	return nil
}

// QueryEntities retrieves a list of existing entities.
func (b *Business) QueryEntities(ctx context.Context) ([]Entity, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.queryentities")
	defer span.End()

	entities, err := b.storer.QueryEntities(ctx)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return entities, nil
}

// =============================================================================
// Automation Rules

// CreateRule adds a new automation rule to the system.
func (b *Business) CreateRule(ctx context.Context, nar NewAutomationRule) (AutomationRule, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.createrule")
	defer span.End()

	now := time.Now().UTC()

	rule := AutomationRule{
		ID:                uuid.New(),
		Name:              nar.Name,
		Description:       nar.Description,
		EntityID:          nar.EntityID,
		EntityTypeID:      nar.EntityTypeID,
		TriggerTypeID:     nar.TriggerTypeID,
		TriggerConditions: nar.TriggerConditions,
		IsActive:          nar.IsActive,
		CreatedDate:       now,
		UpdatedDate:       now,
		CreatedBy:         nar.CreatedBy,
		UpdatedBy:         nar.CreatedBy,
	}

	if err := b.storer.CreateRule(ctx, rule); err != nil {
		return AutomationRule{}, fmt.Errorf("create: %w", err)
	}

	return rule, nil
}

// UpdateRule modifies information about an automation rule.
func (b *Business) UpdateRule(ctx context.Context, rule AutomationRule, uar UpdateAutomationRule) (AutomationRule, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.updaterule")
	defer span.End()

	if uar.Name != nil {
		rule.Name = *uar.Name
	}
	if uar.Description != nil {
		rule.Description = *uar.Description
	}
	if uar.EntityID != nil {
		rule.EntityID = *uar.EntityID
	}
	if uar.EntityTypeID != nil {
		rule.EntityTypeID = *uar.EntityTypeID
	}
	if uar.TriggerTypeID != nil {
		rule.TriggerTypeID = *uar.TriggerTypeID
	}
	if uar.TriggerConditions != nil {
		rule.TriggerConditions = uar.TriggerConditions
	}
	if uar.IsActive != nil {
		rule.IsActive = *uar.IsActive
	}
	if uar.UpdatedBy != nil {
		rule.UpdatedBy = *uar.UpdatedBy
	}

	rule.UpdatedDate = time.Now().UTC()

	if err := b.storer.UpdateRule(ctx, rule); err != nil {
		return AutomationRule{}, fmt.Errorf("update: %w", err)
	}

	return rule, nil
}

// DeactivateRule deactivates the specified automation rule.
func (b *Business) DeactivateRule(ctx context.Context, rule AutomationRule) error {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.deactivaterule")
	defer span.End()

	if err := b.storer.DeactivateRule(ctx, rule); err != nil {
		return fmt.Errorf("deactivate: %w", err)
	}

	return nil
}

// ActivateRule reactivates the specified automation rule.
func (b *Business) ActivateRule(ctx context.Context, rule AutomationRule) error {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.activaterule")
	defer span.End()

	if err := b.storer.ActivateRule(ctx, rule); err != nil {
		return fmt.Errorf("activate: %w", err)
	}

	return nil
}

// QueryActiveRules retrieves all active automation rules.
func (b *Business) QueryActiveRules(ctx context.Context) ([]AutomationRule, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.queryactiverules")
	defer span.End()
	rules, err := b.storer.QueryActiveRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("queryactiverules: %w", err)
	}
	return rules, nil
}

// QueryRuleByID finds the automation rule by the specified ID.
func (b *Business) QueryRuleByID(ctx context.Context, ruleID uuid.UUID) (AutomationRule, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.queryrulebyid")
	defer span.End()

	rule, err := b.storer.QueryRuleByID(ctx, ruleID)
	if err != nil {
		return AutomationRule{}, fmt.Errorf("query: ruleID[%s]: %w", ruleID, err)
	}

	return rule, nil
}

// QueryRulesByEntity finds automation rules by the specified entity ID.
func (b *Business) QueryRulesByEntity(ctx context.Context, entityID uuid.UUID) ([]AutomationRule, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.queryrulesbyentity")
	defer span.End()

	rules, err := b.storer.QueryRulesByEntity(ctx, entityID)
	if err != nil {
		return nil, fmt.Errorf("query: entityID[%s]: %w", entityID, err)
	}

	return rules, nil
}

// =============================================================================
// Rule Actions

// CreateRuleAction adds a new rule action to the system.
func (b *Business) CreateRuleAction(ctx context.Context, nra NewRuleAction) (RuleAction, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.createruleaction")
	defer span.End()

	action := RuleAction{
		ID:               uuid.New(),
		AutomationRuleID: nra.AutomationRuleID,
		Name:             nra.Name,
		Description:      nra.Description,
		ActionConfig:     nra.ActionConfig,
		ExecutionOrder:   nra.ExecutionOrder,
		IsActive:         nra.IsActive,
		TemplateID:       nra.TemplateID,
	}

	if err := b.storer.CreateRuleAction(ctx, action); err != nil {
		return RuleAction{}, fmt.Errorf("create: %w", err)
	}

	return action, nil
}

// UpdateRuleAction modifies information about a rule action.
func (b *Business) UpdateRuleAction(ctx context.Context, action RuleAction, ura UpdateRuleAction) (RuleAction, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.updateruleaction")
	defer span.End()

	if ura.Name != nil {
		action.Name = *ura.Name
	}
	if ura.Description != nil {
		action.Description = *ura.Description
	}
	if ura.ActionConfig != nil {
		action.ActionConfig = *ura.ActionConfig
	}
	if ura.ExecutionOrder != nil {
		action.ExecutionOrder = *ura.ExecutionOrder
	}
	if ura.IsActive != nil {
		action.IsActive = *ura.IsActive
	}
	if ura.TemplateID != nil {
		action.TemplateID = ura.TemplateID
	}

	if err := b.storer.UpdateRuleAction(ctx, action); err != nil {
		return RuleAction{}, fmt.Errorf("update: %w", err)
	}

	return action, nil
}

// DeactivateRuleAction deactivates the specified rule action.
func (b *Business) DeactivateRuleAction(ctx context.Context, action RuleAction) error {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.deactivateruleaction")
	defer span.End()

	if err := b.storer.DeactivateRuleAction(ctx, action); err != nil {
		return fmt.Errorf("deactivate: %w", err)
	}

	return nil
}

// ActivateRuleAction reactivates the specified rule action.
func (b *Business) ActivateRuleAction(ctx context.Context, action RuleAction) error {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.activateruleaction")
	defer span.End()

	if err := b.storer.ActivateRuleAction(ctx, action); err != nil {
		return fmt.Errorf("activate: %w", err)
	}

	return nil
}

// QueryActionsByRule finds rule actions by the specified rule ID.
func (b *Business) QueryActionsByRule(ctx context.Context, ruleID uuid.UUID) ([]RuleAction, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.queryactionsbyrule")
	defer span.End()

	actions, err := b.storer.QueryActionsByRule(ctx, ruleID)
	if err != nil {
		return nil, fmt.Errorf("query: ruleID[%s]: %w", ruleID, err)
	}

	return actions, nil
}

// =============================================================================
// Rule Dependencies

// CreateDependency adds a new rule dependency to the system.
func (b *Business) CreateDependency(ctx context.Context, nrd NewRuleDependency) (RuleDependency, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.createdependency")
	defer span.End()

	// Validate that we're not creating a circular dependency
	if nrd.ParentRuleID == nrd.ChildRuleID {
		return RuleDependency{}, ErrInvalidDependency
	}

	// TODO: Add more sophisticated circular dependency detection
	// This would require traversing the dependency graph

	dependency := RuleDependency{
		ID:           uuid.New(),
		ParentRuleID: nrd.ParentRuleID,
		ChildRuleID:  nrd.ChildRuleID,
	}

	if err := b.storer.CreateDependency(ctx, dependency); err != nil {
		return RuleDependency{}, fmt.Errorf("create: %w", err)
	}

	return dependency, nil
}

// DeleteDependency removes the specified rule dependency.
func (b *Business) DeleteDependency(ctx context.Context, dependency RuleDependency) error {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.deletedependency")
	defer span.End()

	if err := b.storer.DeleteDependency(ctx, dependency); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// QueryDependencies retrieves a list of existing rule dependencies.
func (b *Business) QueryDependencies(ctx context.Context) ([]RuleDependency, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.querydependencies")
	defer span.End()

	dependencies, err := b.storer.QueryDependencies(ctx)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return dependencies, nil
}

// =============================================================================
// Action Templates

// CreateActionTemplate adds a new action template to the system.
func (b *Business) CreateActionTemplate(ctx context.Context, nat NewActionTemplate) (ActionTemplate, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.createactiontemplate")
	defer span.End()

	now := time.Now().UTC()

	template := ActionTemplate{
		ID:            uuid.New(),
		Name:          nat.Name,
		Description:   nat.Description,
		ActionType:    nat.ActionType,
		DefaultConfig: nat.DefaultConfig,
		CreatedDate:   now,
		CreatedBy:     nat.CreatedBy,
	}

	if err := b.storer.CreateActionTemplate(ctx, template); err != nil {
		return ActionTemplate{}, fmt.Errorf("create: %w", err)
	}

	return template, nil
}

// UpdateActionTemplate modifies information about an action template.
func (b *Business) UpdateActionTemplate(ctx context.Context, template ActionTemplate, uat UpdateActionTemplate) (ActionTemplate, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.updateactiontemplate")
	defer span.End()

	if uat.Name != nil {
		template.Name = *uat.Name
	}
	if uat.Description != nil {
		template.Description = *uat.Description
	}
	if uat.ActionType != nil {
		template.ActionType = *uat.ActionType
	}
	if uat.DefaultConfig != nil {
		template.DefaultConfig = *uat.DefaultConfig
	}

	if err := b.storer.UpdateActionTemplate(ctx, template); err != nil {
		return ActionTemplate{}, fmt.Errorf("update: %w", err)
	}

	return template, nil
}

// DeactivateActionTemplate deactivates an action template in the system.
func (b *Business) DeactivateActionTemplate(ctx context.Context, templateID uuid.UUID, deactivatedBy uuid.UUID) error {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.deactivateactiontemplate")
	defer span.End()

	if err := b.storer.DeactivateActionTemplate(ctx, templateID, deactivatedBy); err != nil {
		return fmt.Errorf("deactivate: %w", err)
	}

	return nil
}

// ActivateActionTemplate reactivates an action template in the system.
func (b *Business) ActivateActionTemplate(ctx context.Context, templateID uuid.UUID, activatedBy uuid.UUID) error {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.activateactiontemplate")
	defer span.End()

	if err := b.storer.ActivateActionTemplate(ctx, templateID, activatedBy); err != nil {
		return fmt.Errorf("activate: %w", err)
	}

	return nil
}

// QueryTemplateByID finds the action template by the specified ID.
func (b *Business) QueryTemplateByID(ctx context.Context, templateID uuid.UUID) (ActionTemplate, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.querytemplatebyid")
	defer span.End()

	template, err := b.storer.QueryTemplateByID(ctx, templateID)
	if err != nil {
		return ActionTemplate{}, fmt.Errorf("query: templateID[%s]: %w", templateID, err)
	}

	return template, nil
}

// CreateNotificationDelivery creates a new notification delivery record.
func (b *Business) CreateNotificationDelivery(ctx context.Context, nd NewNotificationDelivery) (NotificationDelivery, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.createnotificationdelivery")
	defer span.End()

	now := time.Now().UTC()

	delivery := NotificationDelivery{
		ID:                    uuid.New(),
		NotificationID:        nd.NotificationID,
		AutomationExecutionID: nd.AutomationExecutionID,
		RuleID:                nd.RuleID,
		ActionID:              nd.ActionID,
		RecipientID:           nd.RecipientID,
		Channel:               nd.Channel,
		Status:                nd.Status,
		Attempts:              nd.Attempts,
		SentAt:                nd.SentAt,
		DeliveredAt:           nd.DeliveredAt,
		FailedAt:              nd.FailedAt,
		ErrorMessage:          nd.ErrorMessage,
		ProviderResponse:      nd.ProviderResponse,
		CreatedDate:           now,
		UpdatedDate:           now,
	}

	if err := b.storer.CreateNotificationDelivery(ctx, delivery); err != nil {
		return NotificationDelivery{}, fmt.Errorf("create: %w", err)
	}

	return delivery, nil
}

func (b *Business) UpdateNotificationDelivery(ctx context.Context, nd NotificationDelivery, und UpdateNotificationDelivery) (NotificationDelivery, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.updatenotificationdelivery")
	defer span.End()

	now := time.Now().UTC()

	if und.NotificationID != nil {
		nd.NotificationID = *und.NotificationID
	}
	if und.AutomationExecutionID != nil {
		nd.AutomationExecutionID = *und.AutomationExecutionID
	}
	if und.RuleID != nil {
		nd.RuleID = *und.RuleID
	}
	if und.ActionID != nil {
		nd.ActionID = *und.ActionID
	}
	if und.RecipientID != nil {
		nd.RecipientID = *und.RecipientID
	}
	if und.Channel != nil {
		nd.Channel = *und.Channel
	}
	if und.Status != nil {
		nd.Status = *und.Status
	}
	if und.Attempts != nil {
		nd.Attempts = *und.Attempts
	}
	if und.SentAt != nil {
		nd.SentAt = und.SentAt
	}
	if und.DeliveredAt != nil {
		nd.DeliveredAt = und.DeliveredAt
	}
	if und.FailedAt != nil {
		nd.FailedAt = und.FailedAt
	}
	if und.ErrorMessage != nil {
		nd.ErrorMessage = und.ErrorMessage
	}
	if und.ProviderResponse != nil {
		nd.ProviderResponse = *und.ProviderResponse
	}
	if und.UpdatedDate != nil {
		nd.UpdatedDate = now
	}

	if err := b.storer.UpdateNotificationDelivery(ctx, nd); err != nil {
		return NotificationDelivery{}, fmt.Errorf("update: %w", err)
	}

	return nd, nil
}

// QueryDeliveriesByAutomationExecution retrieves notification deliveries for a specific automation execution.
func (b *Business) QueryDeliveriesByAutomationExecution(ctx context.Context, executionID uuid.UUID) ([]NotificationDelivery, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.querydeliveriesbyautomationexecution")
	defer span.End()

	deliveries, err := b.storer.QueryDeliveriesByAutomationExecution(ctx, executionID)
	if err != nil {
		return nil, fmt.Errorf("query: executionID[%s]: %w", executionID, err)
	}

	return deliveries, nil
}

// =============================================================================
// Automation Executions

// CreateExecution records a new automation execution in the system.
func (b *Business) CreateExecution(ctx context.Context, nae NewAutomationExecution) (AutomationExecution, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.createexecution")
	defer span.End()

	now := time.Now().UTC()

	execution := AutomationExecution{
		ID:               uuid.New(),
		AutomationRuleID: nae.AutomationRuleID,
		EntityType:       nae.EntityType,
		TriggerData:      nae.TriggerData,
		ActionsExecuted:  nae.ActionsExecuted,
		Status:           nae.Status,
		ErrorMessage:     nae.ErrorMessage,
		ExecutionTimeMs:  nae.ExecutionTimeMs,
		ExecutedAt:       now,
	}

	if err := b.storer.CreateExecution(ctx, execution); err != nil {
		return AutomationExecution{}, fmt.Errorf("create: %w", err)
	}

	return execution, nil
}

// QueryExecutionHistory retrieves execution history for the specified rule.
func (b *Business) QueryExecutionHistory(ctx context.Context, ruleID uuid.UUID, limit int) ([]AutomationExecution, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.queryexecutionhistory")
	defer span.End()

	executions, err := b.storer.QueryExecutionHistory(ctx, ruleID, limit)
	if err != nil {
		return nil, fmt.Errorf("query: ruleID[%s]: %w", ruleID, err)
	}

	return executions, nil
}

// QueryAutomationRulesView retrieves a comprehensive view of automation rules.
func (b *Business) QueryAutomationRulesView(ctx context.Context) ([]AutomationRuleView, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.queryautomationrulesview")
	defer span.End()

	rulesView, err := b.storer.QueryAutomationRulesView(ctx)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return rulesView, nil
}
