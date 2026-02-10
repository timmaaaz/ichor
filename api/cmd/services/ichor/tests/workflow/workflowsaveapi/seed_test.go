package workflowsaveapi_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/workflowsaveapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// SaveSeedData holds test data for workflow save API tests.
type SaveSeedData struct {
	apitest.SeedData
	TriggerTypes []workflow.TriggerType
	EntityTypes  []workflow.EntityType
	Entities     []workflow.Entity

	// Pre-existing rules for update tests
	ExistingRule    workflow.AutomationRule
	ExistingActions []workflow.RuleAction
	ExistingEdges   []workflow.ActionEdge
}

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (SaveSeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	// Create admin user with full permissions
	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return SaveSeedData{}, fmt.Errorf("seeding admin users: %w", err)
	}
	adminUser := apitest.User{
		User:  admins[0],
		Token: apitest.Token(db.BusDomain.User, ath, admins[0].Email.Address),
	}

	// Create regular user with limited permissions (for 401 tests)
	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return SaveSeedData{}, fmt.Errorf("seeding users: %w", err)
	}
	regularUser := apitest.User{
		User:  users[0],
		Token: apitest.Token(db.BusDomain.User, ath, users[0].Email.Address),
	}

	// Seed trigger types (at least 3 for various tests)
	triggerTypes, err := workflow.TestSeedTriggerTypes(ctx, 3, busDomain.Workflow)
	if err != nil {
		return SaveSeedData{}, fmt.Errorf("seeding trigger types: %w", err)
	}

	// Get entity types from migrations
	entityTypes, err := workflow.GetEntityTypes(ctx, busDomain.Workflow)
	if err != nil {
		return SaveSeedData{}, fmt.Errorf("getting entity types: %w", err)
	}

	// Get entities from migrations
	entities, err := workflow.GetEntities(ctx, busDomain.Workflow)
	if err != nil {
		return SaveSeedData{}, fmt.Errorf("getting entities: %w", err)
	}

	// Build ID arrays for seeding
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

	// Create an existing rule for update tests
	existingRules, err := workflow.TestSeedAutomationRules(ctx, 1, entityIDs, entityTypeIDs, triggerTypeIDs, admins[0].ID, busDomain.Workflow)
	if err != nil {
		return SaveSeedData{}, fmt.Errorf("seeding existing rule: %w", err)
	}
	existingRule := existingRules[0]

	// Seed actions for the existing rule
	ruleIDs := []uuid.UUID{existingRule.ID}
	existingActions, err := workflow.TestSeedRuleActions(ctx, 3, ruleIDs, nil, busDomain.Workflow)
	if err != nil {
		return SaveSeedData{}, fmt.Errorf("seeding existing actions: %w", err)
	}

	// Query edges already created by TestSeedRuleActions
	existingEdges, err := busDomain.Workflow.QueryEdgesByRuleID(ctx, existingRule.ID)
	if err != nil {
		return SaveSeedData{}, fmt.Errorf("querying existing edges: %w", err)
	}

	// =========================================================================
	// Table Permissions - Grant admin user full access to workflow.automation_rules
	// =========================================================================

	// Create a test role
	roles, err := rolebus.TestSeedRoles(ctx, 1, busDomain.Role)
	if err != nil {
		return SaveSeedData{}, fmt.Errorf("seeding roles: %w", err)
	}

	// Assign admin user to role
	_, err = userrolebus.TestSeedUserRoles(ctx, uuid.UUIDs{admins[0].ID}, uuid.UUIDs{roles[0].ID}, busDomain.UserRole)
	if err != nil {
		return SaveSeedData{}, fmt.Errorf("seeding user roles: %w", err)
	}

	// Create table access entry for workflow.automation_rules
	_, err = busDomain.TableAccess.Create(ctx, tableaccessbus.NewTableAccess{
		RoleID:    roles[0].ID,
		TableName: workflowsaveapi.RouteTable,
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	})
	if err != nil {
		return SaveSeedData{}, fmt.Errorf("creating table access for workflow.automation_rules: %w", err)
	}

	return SaveSeedData{
		SeedData: apitest.SeedData{
			Users:  []apitest.User{adminUser},
			Admins: []apitest.User{regularUser},
		},
		TriggerTypes:    triggerTypes,
		EntityTypes:     entityTypes,
		Entities:        entities,
		ExistingRule:    existingRule,
		ExistingActions: existingActions,
		ExistingEdges:   existingEdges,
	}, nil
}

