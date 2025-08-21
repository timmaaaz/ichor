// Package workflow manages the CRUD operations for ALL workflow related functions
package workflow

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
)

var (
	ErrNotFound              = errors.New("not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	CreateTriggerType(ctx context.Context, tt TriggerType) error
	UpdateTriggerType(ctx context.Context, tt TriggerType) error
	DeleteTriggerType(ctx context.Context, tt TriggerType) error
	QueryTriggerTypes(ctx context.Context) ([]TriggerType, error)

	CreateEntityType(ctx context.Context, et EntityType) error
	UpdateEntityType(ctx context.Context, et EntityType) error
	DeleteEntityType(ctx context.Context, et EntityType) error
	QueryEntityTypes(ctx context.Context) ([]EntityType, error)

	CreateRule(ctx context.Context, rule AutomationRule) error
	UpdateRule(ctx context.Context, rule AutomationRule) error
	DeleteRule(ctx context.Context, rule AutomationRule) error
	QueryRuleByID(ctx context.Context, id uuid.UUID) (AutomationRule, error)
	QueryRulesByEntity(ctx context.Context, entityid uuid.UUID) ([]AutomationRule, error)

	CreateRuleAction(ctx context.Context, action RuleAction) error
	UpdateRuleAction(ctx context.Context, action RuleAction) error
	DeleteRuleAction(ctx context.Context, action RuleAction) error
	QueryActionsByRule(ctx context.Context, ruleid uuid.UUID) ([]RuleAction, error)

	CreateDependency(ctx context.Context, dep RuleDependency) error
	DeleteDependency(ctx context.Context, dep RuleDependency) error
	QueryDependencies(ctx context.Context) ([]RuleDependency, error)

	CreateActionTemplate(ctx context.Context, template ActionTemplate) error
	UpdateActionTemplate(ctx context.Context, template ActionTemplate) error
	DeleteActionTemplate(ctx context.Context, template ActionTemplate) error
	QueryTemplateByID(ctx context.Context, id uuid.UUID) (ActionTemplate, error)

	CreateEntity(ctx context.Context, entity Entity) error
	UpdateEntity(ctx context.Context, entity Entity) error
	DeleteEntity(ctx context.Context, entity Entity) error
	QueryEntities(ctx context.Context) ([]Entity, error)

	CreateExecution(ctx context.Context, exec AutomationExecution) error
	QueryExecutionHistory(ctx context.Context, ruleid uuid.UUID, limit int) ([]AutomationExecution, error)
}
