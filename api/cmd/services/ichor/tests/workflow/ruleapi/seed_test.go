package rule_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/ruleapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// RuleSeedData holds test data for rule API tests.
type RuleSeedData struct {
	apitest.SeedData
	TriggerTypes []workflow.TriggerType
	EntityTypes  []workflow.EntityType
	Entities     []workflow.Entity
	Rules        []workflow.AutomationRule
	Actions      []workflow.RuleAction
}

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (RuleSeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	// Create admin user with full permissions
	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return RuleSeedData{}, fmt.Errorf("seeding admin users: %w", err)
	}
	adminUser := apitest.User{
		User:  admins[0],
		Token: apitest.Token(db.BusDomain.User, ath, admins[0].Email.Address),
	}

	// Create regular user with limited permissions
	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return RuleSeedData{}, fmt.Errorf("seeding users: %w", err)
	}
	regularUser := apitest.User{
		User:  users[0],
		Token: apitest.Token(db.BusDomain.User, ath, users[0].Email.Address),
	}

	// Seed trigger types
	triggerTypes, err := workflow.TestSeedTriggerTypes(ctx, 3, busDomain.Workflow)
	if err != nil {
		return RuleSeedData{}, fmt.Errorf("seeding trigger types: %w", err)
	}

	// Get entity types from migrations
	entityTypes, err := workflow.GetEntityTypes(ctx, busDomain.Workflow)
	if err != nil {
		return RuleSeedData{}, fmt.Errorf("getting entity types: %w", err)
	}

	// Get entities from migrations
	entities, err := workflow.GetEntities(ctx, busDomain.Workflow)
	if err != nil {
		return RuleSeedData{}, fmt.Errorf("getting entities: %w", err)
	}

	// Seed automation rules using the admin user
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

	rules, err := workflow.TestSeedAutomationRules(ctx, 3, entityIDs, entityTypeIDs, triggerTypeIDs, admins[0].ID, busDomain.Workflow)
	if err != nil {
		return RuleSeedData{}, fmt.Errorf("seeding rules: %w", err)
	}

	// Seed rule actions
	ruleIDs := make([]uuid.UUID, len(rules))
	for i, r := range rules {
		ruleIDs[i] = r.ID
	}

	actions, err := workflow.TestSeedRuleActions(ctx, 5, ruleIDs, nil, busDomain.Workflow)
	if err != nil {
		return RuleSeedData{}, fmt.Errorf("seeding actions: %w", err)
	}

	// =========================================================================
	// Table Permissions - Grant admin user full access to workflow.automation_rules
	// =========================================================================

	// Create a test role
	roles, err := rolebus.TestSeedRoles(ctx, 1, busDomain.Role)
	if err != nil {
		return RuleSeedData{}, fmt.Errorf("seeding roles: %w", err)
	}

	// Assign admin user to role
	_, err = userrolebus.TestSeedUserRoles(ctx, uuid.UUIDs{admins[0].ID}, uuid.UUIDs{roles[0].ID}, busDomain.UserRole)
	if err != nil {
		return RuleSeedData{}, fmt.Errorf("seeding user roles: %w", err)
	}

	// Create table access entry directly for workflow.automation_rules
	// Note: TestSeedTableAccess doesn't include workflow tables, so we create it manually
	_, err = busDomain.TableAccess.Create(ctx, tableaccessbus.NewTableAccess{
		RoleID:    roles[0].ID,
		TableName: ruleapi.RouteTable, // "workflow.automation_rules"
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	})
	if err != nil {
		return RuleSeedData{}, fmt.Errorf("creating table access for workflow.automation_rules: %w", err)
	}

	return RuleSeedData{
		SeedData: apitest.SeedData{
			// Note: Tests use Users[0] for authentication, so admin user goes here
			Users:  []apitest.User{adminUser},
			Admins: []apitest.User{regularUser}, // Not currently used in tests
		},
		TriggerTypes: triggerTypes,
		EntityTypes:  entityTypes,
		Entities:     entities,
		Rules:        rules,
		Actions:      actions,
	}, nil
}

// toAppRule converts a workflow.AutomationRule to a simplified format for comparison.
func toAppRule(rule workflow.AutomationRule) ruleapi.RuleResponse {
	return ruleapi.RuleResponse{
		ID:       rule.ID,
		Name:     rule.Name,
		IsActive: rule.IsActive,
	}
}
