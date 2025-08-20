// Package workflow manages the CRUD operations for ALL workflow related functions
package workflow

import (
	"context"

	"github.com/timmaaaz/ichor/business/sdk/sqldb"
)

type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	CreateTriggerType(ctx context.Context, tt TriggerType) error
	UpdateTriggerType(ctx context.Context, id string, tt TriggerType) error
	DeleteTriggerType(ctx context.Context, id string) error
	QueryTriggerTypes(ctx context.Context) ([]TriggerType, error)

	CreateEntityType(ctx context.Context, et EntityType) error
	UpdateEntityType(ctx context.Context, id string, et EntityType) error
	DeleteEntityType(ctx context.Context, id string) error
	QueryEntityTypes(ctx context.Context) ([]EntityType, error)

	CreateRule(ctx context.Context, rule AutomationRule) error
	UpdateRule(ctx context.Context, id string, rule AutomationRule) error
	DeleteRule(ctx context.Context, id string) error
	QueryRuleByID(ctx context.Context, id string) (AutomationRule, error)
	QueryRulesByEntity(ctx context.Context, entityID string) ([]AutomationRule, error)

	CreateRuleAction(ctx context.Context, action RuleAction) error
	UpdateRuleAction(ctx context.Context, id string, action RuleAction) error
	DeleteRuleAction(ctx context.Context, id string) error
	QueryActionsByRule(ctx context.Context, ruleID string) ([]RuleAction, error)

	CreateDependency(ctx context.Context, dep RuleDependency) error
	DeleteDependency(ctx context.Context, parentID, childID string) error
	QueryDependencies(ctx context.Context) ([]RuleDependency, error)

	CreateActionTemplate(ctx context.Context, template ActionTemplate) error
	UpdateActionTemplate(ctx context.Context, id string, template ActionTemplate) error
	DeleteActionTemplate(ctx context.Context, id string) error
	QueryTemplateByID(ctx context.Context, id string) (ActionTemplate, error)

	CreateEntity(ctx context.Context, entity Entity) error
	UpdateEntity(ctx context.Context, id string, entity Entity) error
	DeleteEntity(ctx context.Context, id string) error
	QueryEntities(ctx context.Context) ([]Entity, error)

	CreateExecution(ctx context.Context, exec AutomationExecution) error
	QueryExecutionHistory(ctx context.Context, ruleID string, limit int) ([]AutomationExecution, error)
}
