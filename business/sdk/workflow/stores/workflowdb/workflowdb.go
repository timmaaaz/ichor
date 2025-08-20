package workflowdb

import (
	"context"

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

// stores/workflowdb/workflowdb.go
// Trigger Types
func (s *Store) CreateTriggerType(ctx context.Context, tt workflow.TriggerType) error
func (s *Store) UpdateTriggerType(ctx context.Context, id string, tt workflow.TriggerType) error
func (s *Store) DeleteTriggerType(ctx context.Context, id string) error
func (s *Store) QueryTriggerTypes(ctx context.Context) ([]workflow.TriggerType, error)

// Entity Types
func (s *Store) CreateEntityType(ctx context.Context, et workflow.EntityType) error
func (s *Store) UpdateEntityType(ctx context.Context, id string, et workflow.EntityType) error
func (s *Store) DeleteEntityType(ctx context.Context, id string) error
func (s *Store) QueryEntityTypes(ctx context.Context) ([]workflow.EntityType, error)

// Rules
func (s *Store) CreateRule(ctx context.Context, rule workflow.AutomationRule) error
func (s *Store) UpdateRule(ctx context.Context, id string, rule workflow.AutomationRule) error
func (s *Store) DeleteRule(ctx context.Context, id string) error
func (s *Store) QueryRuleByID(ctx context.Context, id string) (workflow.AutomationRule, error)
func (s *Store) QueryRulesByEntity(ctx context.Context, entityID string) ([]workflow.AutomationRule, error)

// Actions
func (s *Store) CreateRuleAction(ctx context.Context, action workflow.RuleAction) error
func (s *Store) UpdateRuleAction(ctx context.Context, id string, action workflow.RuleAction) error
func (s *Store) DeleteRuleAction(ctx context.Context, id string) error
func (s *Store) QueryActionsByRule(ctx context.Context, ruleID string) ([]workflow.RuleAction, error)

// Dependencies
func (s *Store) CreateDependency(ctx context.Context, dep workflow.RuleDependency) error
func (s *Store) DeleteDependency(ctx context.Context, parentID, childID string) error
func (s *Store) QueryDependencies(ctx context.Context) ([]workflow.RuleDependency, error)

// Templates
func (s *Store) CreateActionTemplate(ctx context.Context, template workflow.ActionTemplate) error
func (s *Store) UpdateActionTemplate(ctx context.Context, id string, template workflow.ActionTemplate) error
func (s *Store) DeleteActionTemplate(ctx context.Context, id string) error
func (s *Store) QueryTemplateByID(ctx context.Context, id string) (workflow.ActionTemplate, error)

// Entities
func (s *Store) CreateEntity(ctx context.Context, entity workflow.Entity) error
func (s *Store) UpdateEntity(ctx context.Context, id string, entity workflow.Entity) error
func (s *Store) DeleteEntity(ctx context.Context, id string) error
func (s *Store) QueryEntities(ctx context.Context) ([]workflow.Entity, error)

// Executions (mostly create/read)
func (s *Store) CreateExecution(ctx context.Context, exec workflow.AutomationExecution) error
func (s *Store) QueryExecutionHistory(ctx context.Context, ruleID string, limit int) ([]workflow.AutomationExecution, error)
