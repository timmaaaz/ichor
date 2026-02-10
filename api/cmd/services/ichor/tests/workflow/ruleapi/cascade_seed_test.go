package rule_test

import (
	"context"
	"encoding/json"
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

// CascadeSeedData holds test data specifically for cascade API tests.
// These tests require specific rules and actions to test downstream workflow detection.
type CascadeSeedData struct {
	apitest.SeedData
	TriggerTypes []workflow.TriggerType
	EntityTypes  []workflow.EntityType
	Entities     []workflow.Entity

	// PrimaryRule has an update_field action that modifies an entity
	// that other rules listen to
	PrimaryRule        workflow.AutomationRule
	PrimaryRuleActions []workflow.RuleAction

	// DownstreamTriggerRules are rules that listen to the same entity
	// that the PrimaryRule modifies
	DownstreamTriggerRules []workflow.AutomationRule

	// NonModifyingRule has actions that don't modify any entities (e.g., generic actions)
	NonModifyingRule        workflow.AutomationRule
	NonModifyingRuleActions []workflow.RuleAction

	// MixedActionsRule has both modifying and non-modifying actions
	MixedActionsRule        workflow.AutomationRule
	MixedActionsRuleActions []workflow.RuleAction

	// SelfTriggerRule listens to the same entity it modifies (should exclude itself)
	SelfTriggerRule        workflow.AutomationRule
	SelfTriggerRuleActions []workflow.RuleAction

	// InactiveDownstreamRule is an inactive rule that listens to the same entity
	// as the PrimaryRule - should be excluded from cascade results
	InactiveDownstreamRule workflow.AutomationRule

	// TargetEntity is the entity used for testing (what PrimaryRule modifies)
	TargetEntity workflow.Entity
}

// insertCascadeSeedData sets up the test data for cascade API tests.
func insertCascadeSeedData(db *dbtest.Database, ath *auth.Auth) (CascadeSeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	// =========================================================================
	// Create admin user with full permissions
	// =========================================================================
	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("seeding admin users: %w", err)
	}
	adminUser := apitest.User{
		User:  admins[0],
		Token: apitest.Token(db.BusDomain.User, ath, admins[0].Email.Address),
	}

	// =========================================================================
	// Seed workflow reference data
	// =========================================================================

	// Seed trigger types (on_create, on_update, on_delete)
	triggerTypes, err := workflow.TestSeedTriggerTypes(ctx, 3, busDomain.Workflow)
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("seeding trigger types: %w", err)
	}

	// Get entity types from migrations
	entityTypes, err := workflow.GetEntityTypes(ctx, busDomain.Workflow)
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("getting entity types: %w", err)
	}

	// Get entities from migrations
	entities, err := workflow.GetEntities(ctx, busDomain.Workflow)
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("getting entities: %w", err)
	}

	if len(entities) == 0 || len(entityTypes) == 0 || len(triggerTypes) < 2 {
		return CascadeSeedData{}, fmt.Errorf("insufficient seed data: entities=%d, entityTypes=%d, triggerTypes=%d",
			len(entities), len(entityTypes), len(triggerTypes))
	}

	// Find on_update trigger type
	var onUpdateTriggerType workflow.TriggerType
	for _, tt := range triggerTypes {
		if tt.Name == "on_update" {
			onUpdateTriggerType = tt
			break
		}
	}
	if onUpdateTriggerType.ID == uuid.Nil {
		return CascadeSeedData{}, fmt.Errorf("on_update trigger type not found")
	}

	// Use the first entity as our target entity
	targetEntity := entities[0]

	// Construct full table name from schema and entity name
	fullTableName := targetEntity.SchemaName + "." + targetEntity.Name

	// =========================================================================
	// Create Action Templates - needed for TemplateActionType to be populated
	// =========================================================================

	// Create update_field action template
	updateFieldTemplate, err := busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:          "Update Field Template",
		Description:   "Template for update_field actions",
		ActionType:    "update_field",
		DefaultConfig: createUpdateFieldActionConfig(fullTableName, "status", "default"),
		CreatedBy:     admins[0].ID,
	})
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("creating update_field template: %w", err)
	}

	// Create generic action template (for non-modifying actions)
	genericTemplate, err := busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:          "Generic Action Template",
		Description:   "Template for generic non-modifying actions",
		ActionType:    "generic",
		DefaultConfig: createGenericActionConfig(),
		CreatedBy:     admins[0].ID,
	})
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("creating generic template: %w", err)
	}

	// =========================================================================
	// Create PrimaryRule - has an update_field action that modifies targetEntity
	// =========================================================================
	primaryRuleNew := workflow.NewAutomationRule{
		Name:          "Primary Cascade Test Rule",
		Description:   "Rule with update_field action for cascade testing",
		EntityID:      targetEntity.ID,
		EntityTypeID:  entityTypes[0].ID,
		TriggerTypeID: onUpdateTriggerType.ID,
		IsActive:      true,
		CreatedBy:     admins[0].ID,
	}
	primaryRule, err := busDomain.Workflow.CreateRule(ctx, primaryRuleNew)
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("creating primary rule: %w", err)
	}

	// Create update_field action for primary rule (linked to template)
	updateFieldConfig := createUpdateFieldActionConfig(fullTableName, "status", "processed")
	primaryAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: primaryRule.ID,
		Name:             "Update Status Action",
		Description:      "Updates entity status field",
		ActionConfig:     updateFieldConfig,
		IsActive:         true,
		TemplateID:       &updateFieldTemplate.ID,
	})
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("creating primary rule action: %w", err)
	}

	_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         primaryRule.ID,
		SourceActionID: nil,
		TargetActionID: primaryAction.ID,
		EdgeType:       workflow.EdgeTypeStart,
		EdgeOrder:      0,
	})
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("creating edge for primary rule action: %w", err)
	}

	// =========================================================================
	// Create DownstreamTriggerRules - listen to the same entity
	// =========================================================================
	downstreamRules := make([]workflow.AutomationRule, 2)

	for i := 0; i < 2; i++ {
		downstreamNew := workflow.NewAutomationRule{
			Name:          fmt.Sprintf("Downstream Rule %d", i+1),
			Description:   fmt.Sprintf("Listens to %s for cascade testing", fullTableName),
			EntityID:      targetEntity.ID,
			EntityTypeID:  entityTypes[0].ID,
			TriggerTypeID: onUpdateTriggerType.ID,
			IsActive:      true,
			CreatedBy:     admins[0].ID,
		}
		rule, err := busDomain.Workflow.CreateRule(ctx, downstreamNew)
		if err != nil {
			return CascadeSeedData{}, fmt.Errorf("creating downstream rule %d: %w", i+1, err)
		}
		downstreamRules[i] = rule

		// Add a simple action to each downstream rule (linked to generic template)
		downstreamAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
			AutomationRuleID: rule.ID,
			Name:             fmt.Sprintf("Downstream Action %d", i+1),
			Description:      "Action for downstream rule",
			ActionConfig:     createGenericActionConfig(),
			IsActive:         true,
			TemplateID:       &genericTemplate.ID,
		})
		if err != nil {
			return CascadeSeedData{}, fmt.Errorf("creating downstream rule %d action: %w", i+1, err)
		}

		_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
			RuleID:         rule.ID,
			SourceActionID: nil,
			TargetActionID: downstreamAction.ID,
			EdgeType:       workflow.EdgeTypeStart,
			EdgeOrder:      0,
		})
		if err != nil {
			return CascadeSeedData{}, fmt.Errorf("creating edge for downstream rule %d action: %w", i+1, err)
		}
	}

	// =========================================================================
	// Create NonModifyingRule - has actions that don't modify entities
	// =========================================================================
	nonModifyingRuleNew := workflow.NewAutomationRule{
		Name:          "Non-Modifying Rule",
		Description:   "Rule with actions that don't modify entities",
		EntityID:      targetEntity.ID,
		EntityTypeID:  entityTypes[0].ID,
		TriggerTypeID: onUpdateTriggerType.ID,
		IsActive:      true,
		CreatedBy:     admins[0].ID,
	}
	nonModifyingRule, err := busDomain.Workflow.CreateRule(ctx, nonModifyingRuleNew)
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("creating non-modifying rule: %w", err)
	}

	// Create generic action (doesn't modify entities, linked to generic template)
	nonModifyingAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: nonModifyingRule.ID,
		Name:             "Generic Action",
		Description:      "Generic action that doesn't modify entities",
		ActionConfig:     createGenericActionConfig(),
		IsActive:         true,
		TemplateID:       &genericTemplate.ID,
	})
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("creating non-modifying action: %w", err)
	}

	_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         nonModifyingRule.ID,
		SourceActionID: nil,
		TargetActionID: nonModifyingAction.ID,
		EdgeType:       workflow.EdgeTypeStart,
		EdgeOrder:      0,
	})
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("creating edge for non-modifying action: %w", err)
	}

	// =========================================================================
	// Create MixedActionsRule - has both modifying and non-modifying actions
	// =========================================================================
	mixedRuleNew := workflow.NewAutomationRule{
		Name:          "Mixed Actions Rule",
		Description:   "Rule with both modifying and non-modifying actions",
		EntityID:      targetEntity.ID,
		EntityTypeID:  entityTypes[0].ID,
		TriggerTypeID: onUpdateTriggerType.ID,
		IsActive:      true,
		CreatedBy:     admins[0].ID,
	}
	mixedRule, err := busDomain.Workflow.CreateRule(ctx, mixedRuleNew)
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("creating mixed rule: %w", err)
	}

	// Create update_field action (modifies entities, linked to update_field template)
	mixedModifyingAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: mixedRule.ID,
		Name:             "Mixed Modifying Action",
		Description:      "Updates entity field",
		ActionConfig:     createUpdateFieldActionConfig(fullTableName, "status", "mixed"),
		IsActive:         true,
		TemplateID:       &updateFieldTemplate.ID,
	})
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("creating mixed modifying action: %w", err)
	}

	// Create generic action (doesn't modify entities, linked to generic template)
	mixedNonModifyingAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: mixedRule.ID,
		Name:             "Mixed Non-Modifying Action",
		Description:      "Generic action",
		ActionConfig:     createGenericActionConfig(),
		IsActive:         true,
		TemplateID:       &genericTemplate.ID,
	})
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("creating mixed non-modifying action: %w", err)
	}

	// Create edge chain for mixed rule: start -> modifying -> non-modifying
	_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         mixedRule.ID,
		SourceActionID: nil,
		TargetActionID: mixedModifyingAction.ID,
		EdgeType:       workflow.EdgeTypeStart,
		EdgeOrder:      0,
	})
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("creating start edge for mixed rule: %w", err)
	}

	mixedModifyingActionID := mixedModifyingAction.ID
	_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         mixedRule.ID,
		SourceActionID: &mixedModifyingActionID,
		TargetActionID: mixedNonModifyingAction.ID,
		EdgeType:       workflow.EdgeTypeSequence,
		EdgeOrder:      1,
	})
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("creating sequence edge for mixed rule: %w", err)
	}

	// =========================================================================
	// Create SelfTriggerRule - listens to and modifies the same entity
	// =========================================================================
	selfTriggerRuleNew := workflow.NewAutomationRule{
		Name:          "Self Trigger Rule",
		Description:   "Rule that listens to and modifies the same entity",
		EntityID:      targetEntity.ID,
		EntityTypeID:  entityTypes[0].ID,
		TriggerTypeID: onUpdateTriggerType.ID,
		IsActive:      true,
		CreatedBy:     admins[0].ID,
	}
	selfTriggerRule, err := busDomain.Workflow.CreateRule(ctx, selfTriggerRuleNew)
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("creating self trigger rule: %w", err)
	}

	// Create update_field action that modifies the same entity it listens to (linked to template)
	selfTriggerAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: selfTriggerRule.ID,
		Name:             "Self Trigger Action",
		Description:      "Updates entity that rule is listening to",
		ActionConfig:     createUpdateFieldActionConfig(fullTableName, "status", "self_trigger"),
		IsActive:         true,
		TemplateID:       &updateFieldTemplate.ID,
	})
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("creating self trigger action: %w", err)
	}

	_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         selfTriggerRule.ID,
		SourceActionID: nil,
		TargetActionID: selfTriggerAction.ID,
		EdgeType:       workflow.EdgeTypeStart,
		EdgeOrder:      0,
	})
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("creating edge for self trigger action: %w", err)
	}

	// =========================================================================
	// Create InactiveDownstreamRule - inactive rule that listens to same entity
	// =========================================================================
	inactiveRuleNew := workflow.NewAutomationRule{
		Name:          "Inactive Downstream Rule",
		Description:   "Inactive rule that listens to target entity",
		EntityID:      targetEntity.ID,
		EntityTypeID:  entityTypes[0].ID,
		TriggerTypeID: onUpdateTriggerType.ID,
		IsActive:      false, // INACTIVE
		CreatedBy:     admins[0].ID,
	}
	inactiveRule, err := busDomain.Workflow.CreateRule(ctx, inactiveRuleNew)
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("creating inactive rule: %w", err)
	}

	// =========================================================================
	// Table Permissions - Grant admin user full access
	// =========================================================================

	// Create a test role
	roles, err := rolebus.TestSeedRoles(ctx, 1, busDomain.Role)
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("seeding roles: %w", err)
	}

	// Assign admin user to role
	_, err = userrolebus.TestSeedUserRoles(ctx, uuid.UUIDs{admins[0].ID}, uuid.UUIDs{roles[0].ID}, busDomain.UserRole)
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("seeding user roles: %w", err)
	}

	// Create table access entry for workflow.automation_rules
	_, err = busDomain.TableAccess.Create(ctx, tableaccessbus.NewTableAccess{
		RoleID:    roles[0].ID,
		TableName: ruleapi.RouteTable,
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	})
	if err != nil {
		return CascadeSeedData{}, fmt.Errorf("creating table access: %w", err)
	}

	return CascadeSeedData{
		SeedData: apitest.SeedData{
			Users: []apitest.User{adminUser},
		},
		TriggerTypes:            triggerTypes,
		EntityTypes:             entityTypes,
		Entities:                entities,
		PrimaryRule:             primaryRule,
		PrimaryRuleActions:      []workflow.RuleAction{primaryAction},
		DownstreamTriggerRules:  downstreamRules,
		NonModifyingRule:        nonModifyingRule,
		NonModifyingRuleActions: []workflow.RuleAction{nonModifyingAction},
		MixedActionsRule:        mixedRule,
		MixedActionsRuleActions: []workflow.RuleAction{mixedModifyingAction, mixedNonModifyingAction},
		SelfTriggerRule:         selfTriggerRule,
		SelfTriggerRuleActions:  []workflow.RuleAction{selfTriggerAction},
		InactiveDownstreamRule:  inactiveRule,
		TargetEntity:            targetEntity,
	}, nil
}

// createUpdateFieldActionConfig creates an update_field action configuration.
// This is the format expected by the UpdateFieldHandler.
func createUpdateFieldActionConfig(targetEntity, targetField, newValue string) json.RawMessage {
	config := map[string]interface{}{
		"target_entity": targetEntity,
		"target_field":  targetField,
		"new_value":     newValue,
	}
	data, _ := json.Marshal(config)
	return data
}

// createGenericActionConfig creates a generic action config for non-modifying actions.
func createGenericActionConfig() json.RawMessage {
	config := map[string]interface{}{
		"action_param": "test_value",
		"enabled":      true,
	}
	data, _ := json.Marshal(config)
	return data
}
