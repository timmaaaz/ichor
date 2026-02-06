package edge_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/edgeapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// EdgeSeedData holds test data for edge API tests.
type EdgeSeedData struct {
	apitest.SeedData
	TriggerTypes []workflow.TriggerType
	EntityTypes  []workflow.EntityType
	Entities     []workflow.Entity
	Rules        []workflow.AutomationRule
	Actions      []workflow.RuleAction
	Edges        []workflow.ActionEdge
	// OtherRule is used for cross-rule validation tests (action belongs to different rule)
	OtherRule workflow.AutomationRule
	// OtherAction is an action belonging to OtherRule
	OtherAction workflow.RuleAction
}

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (EdgeSeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	// =========================================================================
	// Create admin user with full permissions
	// =========================================================================
	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return EdgeSeedData{}, fmt.Errorf("seeding admin users: %w", err)
	}
	adminUser := apitest.User{
		User:  admins[0],
		Token: apitest.Token(db.BusDomain.User, ath, admins[0].Email.Address),
	}

	// Create regular user with limited permissions (for 401 tests)
	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return EdgeSeedData{}, fmt.Errorf("seeding users: %w", err)
	}
	regularUser := apitest.User{
		User:  users[0],
		Token: apitest.Token(db.BusDomain.User, ath, users[0].Email.Address),
	}

	// =========================================================================
	// Seed workflow reference data
	// =========================================================================

	// Seed trigger types
	triggerTypes, err := workflow.TestSeedTriggerTypes(ctx, 3, busDomain.Workflow)
	if err != nil {
		return EdgeSeedData{}, fmt.Errorf("seeding trigger types: %w", err)
	}

	// Get entity types from migrations
	entityTypes, err := workflow.GetEntityTypes(ctx, busDomain.Workflow)
	if err != nil {
		return EdgeSeedData{}, fmt.Errorf("getting entity types: %w", err)
	}

	// Get entities from migrations
	entities, err := workflow.GetEntities(ctx, busDomain.Workflow)
	if err != nil {
		return EdgeSeedData{}, fmt.Errorf("getting entities: %w", err)
	}

	// =========================================================================
	// Seed automation rules (need 2: one for primary tests, one for cross-rule validation)
	// =========================================================================
	entityIDs := make([]uuid.UUID, len(entities))
	for i, e := range entities {
		entityIDs[i] = e.ID
	}
	entityTypeIDs := make([]uuid.UUID, len(entityTypes))
	for i, et := range entityTypes {
		entityTypeIDs[i] = et.ID
	}
	triggerTypeIDs := make([]uuid.UUID, len(triggerTypes))
	for i, tt := range triggerTypes {
		triggerTypeIDs[i] = tt.ID
	}

	// Create 2 automation rules
	rules, err := workflow.TestSeedAutomationRules(ctx, 2, entityIDs, entityTypeIDs, triggerTypeIDs, admins[0].ID, busDomain.Workflow)
	if err != nil {
		return EdgeSeedData{}, fmt.Errorf("seeding rules: %w", err)
	}

	primaryRule := rules[0]
	otherRule := rules[1]

	// =========================================================================
	// Seed rule actions - 4 for primary rule, 1 for other rule
	// =========================================================================

	// Create 4 actions for the primary rule (for graph/branching tests)
	primaryActions, err := workflow.TestSeedRuleActions(ctx, 4, []uuid.UUID{primaryRule.ID}, nil, busDomain.Workflow)
	if err != nil {
		return EdgeSeedData{}, fmt.Errorf("seeding primary rule actions: %w", err)
	}

	// Create 1 action for the other rule (for cross-rule validation tests)
	otherActions, err := workflow.TestSeedRuleActions(ctx, 1, []uuid.UUID{otherRule.ID}, nil, busDomain.Workflow)
	if err != nil {
		return EdgeSeedData{}, fmt.Errorf("seeding other rule action: %w", err)
	}

	// =========================================================================
	// Query existing edges created by TestSeedRuleActions (for query/delete tests)
	// Note: TestSeedRuleActions automatically creates edges (start + sequence chain)
	// =========================================================================
	edges, err := busDomain.Workflow.QueryEdgesByRuleID(ctx, primaryRule.ID)
	if err != nil {
		return EdgeSeedData{}, fmt.Errorf("querying edges: %w", err)
	}

	// =========================================================================
	// Table Permissions - Grant admin user full access to workflow.automation_rules
	// =========================================================================

	// Create a test role
	roles, err := rolebus.TestSeedRoles(ctx, 1, busDomain.Role)
	if err != nil {
		return EdgeSeedData{}, fmt.Errorf("seeding roles: %w", err)
	}

	// Assign admin user to role
	_, err = userrolebus.TestSeedUserRoles(ctx, uuid.UUIDs{admins[0].ID}, uuid.UUIDs{roles[0].ID}, busDomain.UserRole)
	if err != nil {
		return EdgeSeedData{}, fmt.Errorf("seeding user roles: %w", err)
	}

	// Create table access entry for workflow.automation_rules
	_, err = busDomain.TableAccess.Create(ctx, tableaccessbus.NewTableAccess{
		RoleID:    roles[0].ID,
		TableName: edgeapi.RouteTable, // "workflow.automation_rules"
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	})
	if err != nil {
		return EdgeSeedData{}, fmt.Errorf("creating table access for workflow.automation_rules: %w", err)
	}

	return EdgeSeedData{
		SeedData: apitest.SeedData{
			// Note: Tests use Users[0] for authentication, so admin user goes here
			Users:  []apitest.User{adminUser},
			Admins: []apitest.User{regularUser},
		},
		TriggerTypes: triggerTypes,
		EntityTypes:  entityTypes,
		Entities:     entities,
		Rules:        []workflow.AutomationRule{primaryRule},
		Actions:      primaryActions,
		Edges:        edges,
		OtherRule:    otherRule,
		OtherAction:  otherActions[0],
	}, nil
}

// toEdgeResponse converts a workflow.ActionEdge to an edgeapi.EdgeResponse for comparison.
func toEdgeResponse(edge workflow.ActionEdge) edgeapi.EdgeResponse {
	return edgeapi.EdgeResponse{
		ID:             edge.ID,
		RuleID:         edge.RuleID,
		SourceActionID: edge.SourceActionID,
		TargetActionID: edge.TargetActionID,
		EdgeType:       edge.EdgeType,
		EdgeOrder:      edge.EdgeOrder,
		CreatedDate:    edge.CreatedDate,
	}
}
