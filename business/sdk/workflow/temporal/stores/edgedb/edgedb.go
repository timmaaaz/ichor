// Package edgedb implements the EdgeStore interface for loading graph
// definitions from PostgreSQL. Used by the Temporal trigger system
// to build GraphDefinition from rule_actions and action_edges tables.
package edgedb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Verify Store implements EdgeStore interface at compile time.
var _ temporal.EdgeStore = (*Store)(nil)

// Store implements the temporal.EdgeStore interface by loading
// graph definitions from PostgreSQL.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore creates a new edge store.
func NewStore(log *logger.Logger, db sqlx.ExtContext) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// =============================================================================
// Database Models
// =============================================================================

// dbAction represents the database model for rule_actions joined with
// action_templates to resolve ActionType.
//
// Uses sql.NullString for nullable columns:
//   - deactivated_by: NULL when action is active
//   - template_action_type: NULL when no template is linked (template_id is NULL)
type dbAction struct {
	ID                 string          `db:"id"`
	Name               string          `db:"name"`
	Description        sql.NullString  `db:"description"`
	ActionConfig       json.RawMessage `db:"action_config"`
	IsActive           bool            `db:"is_active"`
	DeactivatedBy      sql.NullString  `db:"deactivated_by"`
	TemplateActionType sql.NullString  `db:"template_action_type"`
}

// dbEdge represents the database model for action_edges.
//
// Uses sql.NullString for source_action_id which is NULL for start edges.
type dbEdge struct {
	ID             string         `db:"id"`
	SourceActionID sql.NullString `db:"source_action_id"`
	TargetActionID string         `db:"target_action_id"`
	EdgeType       string         `db:"edge_type"`
	EdgeOrder      int            `db:"edge_order"`
}

// =============================================================================
// EdgeStore Interface Implementation
// =============================================================================

// QueryActionsByRule returns all action nodes for a given automation rule,
// with ActionType resolved from the linked action_template.
//
// The caller should set a context deadline to prevent indefinite blocking
// on slow or locked database queries.
func (s *Store) QueryActionsByRule(ctx context.Context, ruleID uuid.UUID) ([]temporal.ActionNode, error) {
	data := struct {
		RuleID string `db:"automation_rules_id"`
	}{
		RuleID: ruleID.String(),
	}

	const q = `
	SELECT
		ra.id,
		ra.name,
		ra.description,
		ra.action_config,
		ra.is_active,
		ra.deactivated_by,
		at.action_type AS template_action_type
	FROM
		workflow.rule_actions ra
	LEFT JOIN
		workflow.action_templates at ON ra.template_id = at.id
	WHERE
		ra.automation_rules_id = :automation_rules_id`

	var dbActions []dbAction
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbActions); err != nil {
		return nil, fmt.Errorf("namedqueryslice[actions]: %w", err)
	}

	s.log.Info(ctx, "edgedb.QueryActionsByRule", "rule_id", ruleID, "count", len(dbActions))

	actions := make([]temporal.ActionNode, len(dbActions))
	for i, dba := range dbActions {
		actions[i] = toActionNode(dba)
	}

	return actions, nil
}

// QueryEdgesByRule returns all action edges for a given automation rule,
// ordered by edge_order for deterministic graph traversal.
//
// The caller should set a context deadline to prevent indefinite blocking
// on slow or locked database queries.
func (s *Store) QueryEdgesByRule(ctx context.Context, ruleID uuid.UUID) ([]temporal.ActionEdge, error) {
	data := struct {
		RuleID string `db:"rule_id"`
	}{
		RuleID: ruleID.String(),
	}

	const q = `
	SELECT
		id,
		source_action_id,
		target_action_id,
		edge_type,
		edge_order
	FROM
		workflow.action_edges
	WHERE
		rule_id = :rule_id
	ORDER BY
		edge_order ASC`

	var dbEdges []dbEdge
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbEdges); err != nil {
		return nil, fmt.Errorf("namedqueryslice[edges]: %w", err)
	}

	s.log.Info(ctx, "edgedb.QueryEdgesByRule", "rule_id", ruleID, "count", len(dbEdges))

	edges := make([]temporal.ActionEdge, len(dbEdges))
	for i, dbe := range dbEdges {
		edges[i] = toActionEdge(dbe)
	}

	return edges, nil
}

// =============================================================================
// Conversion Functions
// =============================================================================

// toActionNode converts a database action row to a temporal.ActionNode.
func toActionNode(dba dbAction) temporal.ActionNode {
	node := temporal.ActionNode{
		ID:       uuid.MustParse(dba.ID),
		Name:     dba.Name,
		Config:   dba.ActionConfig,
		IsActive: dba.IsActive,
	}

	if dba.Description.Valid {
		node.Description = dba.Description.String
	}

	if dba.DeactivatedBy.Valid {
		node.DeactivatedBy = uuid.MustParse(dba.DeactivatedBy.String)
	}

	if dba.TemplateActionType.Valid {
		node.ActionType = dba.TemplateActionType.String
	}

	return node
}

// toActionEdge converts a database edge row to a temporal.ActionEdge.
func toActionEdge(dbe dbEdge) temporal.ActionEdge {
	edge := temporal.ActionEdge{
		ID:             uuid.MustParse(dbe.ID),
		TargetActionID: uuid.MustParse(dbe.TargetActionID),
		EdgeType:       dbe.EdgeType,
		SortOrder:      dbe.EdgeOrder,
	}

	// source_action_id is NULL for start edges.
	if dbe.SourceActionID.Valid {
		id := uuid.MustParse(dbe.SourceActionID.String)
		edge.SourceActionID = &id
	}

	return edge
}
