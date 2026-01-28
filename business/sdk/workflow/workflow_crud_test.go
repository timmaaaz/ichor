package workflow_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

func Test_Workflow(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_Workflow")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	unitest.Run(t, triggerTypeTests(db.BusDomain, sd), "triggerType")
	unitest.Run(t, entityTypeTests(db.BusDomain, sd), "entityType")
	unitest.Run(t, entityTests(db.BusDomain, sd), "entity")
	unitest.Run(t, automationRuleTests(db.BusDomain, sd), "automationRule")
	unitest.Run(t, actionTemplateTests(db.BusDomain, sd), "actionTemplate")
	unitest.Run(t, ruleActionTests(db.BusDomain, sd), "ruleAction")
	unitest.Run(t, ruleDependencyTests(db.BusDomain, sd), "ruleDependency")
	unitest.Run(t, automationExecutionTests(db.BusDomain, sd), "automationExecution")
	unitest.Run(t, notificationDeliveryTests(db.BusDomain, sd), "notificationDelivery")
}

// =============================================================================

type workflowSeedData struct {
	unitest.SeedData
	TriggerTypes           []workflow.TriggerType
	EntityTypes            []workflow.EntityType
	Entities               []workflow.Entity
	AutomationRules        []workflow.AutomationRule
	ActionTemplates        []workflow.ActionTemplate
	RuleActions            []workflow.RuleAction
	RuleDependencies       []workflow.RuleDependency
	AutomationExecutions   []workflow.AutomationExecution
	NotificationDeliveries []workflow.NotificationDelivery
}

func insertSeedData(busDomain dbtest.BusDomain) (workflowSeedData, error) {
	ctx := context.Background()

	// Seed users first (needed for created_by fields)
	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 2, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return workflowSeedData{}, fmt.Errorf("seeding users : %w", err)
	}
	userIDs := make([]uuid.UUID, len(usrs))
	for i, u := range usrs {
		userIDs[i] = u.ID
	}

	adminUser := usrs[0]

	// Seed trigger types
	triggerTypes, err := workflow.TestSeedTriggerTypes(ctx, 3, busDomain.Workflow)
	if err != nil {
		return workflowSeedData{}, fmt.Errorf("seeding trigger types : %w", err)
	}

	// Seed entity types
	entityTypes, err := workflow.GetEntityTypes(ctx, busDomain.Workflow)
	if err != nil {
		return workflowSeedData{}, fmt.Errorf("seeding entity types : %w", err)
	}

	// Extract entity type IDs
	entityTypeIDs := make([]uuid.UUID, len(entityTypes))
	for i, et := range entityTypes {
		entityTypeIDs[i] = et.ID
	}

	// Seed entities
	entities, err := workflow.GetEntities(ctx, busDomain.Workflow)
	if err != nil {
		return workflowSeedData{}, fmt.Errorf("seeding entities : %w", err)
	}

	// Extract IDs for rules
	entityIDs := make([]uuid.UUID, len(entities))
	for i, e := range entities {
		entityIDs[i] = e.ID
	}

	triggerTypeIDs := make([]uuid.UUID, len(triggerTypes))
	for i, tt := range triggerTypes {
		triggerTypeIDs[i] = tt.ID
	}

	// Seed automation rules
	rules, err := workflow.TestSeedAutomationRules(ctx, 5, entityIDs, entityTypeIDs, triggerTypeIDs, adminUser.ID, busDomain.Workflow)
	if err != nil {
		return workflowSeedData{}, fmt.Errorf("seeding automation rules : %w", err)
	}

	// Extract rule IDs
	ruleIDs := make([]uuid.UUID, len(rules))
	for i, r := range rules {
		ruleIDs[i] = r.ID
	}

	// Seed action templates
	templates, err := workflow.TestSeedActionTemplates(ctx, 3, adminUser.ID, busDomain.Workflow)
	if err != nil {
		return workflowSeedData{}, fmt.Errorf("seeding action templates : %w", err)
	}

	// Extract template IDs
	templateIDs := make([]uuid.UUID, len(templates))
	for i, t := range templates {
		templateIDs[i] = t.ID
	}

	// Seed rule actions
	actions, err := workflow.TestSeedRuleActions(ctx, 7, ruleIDs, &templateIDs, busDomain.Workflow)
	if err != nil {
		return workflowSeedData{}, fmt.Errorf("seeding rule actions : %w", err)
	}

	actionIDs := make([]uuid.UUID, len(actions))
	for i, a := range actions {
		actionIDs[i] = a.ID
	}

	// Seed rule dependencies
	var dependencies []workflow.RuleDependency
	if len(ruleIDs) >= 3 {
		parentIDs := ruleIDs[:2]
		childIDs := ruleIDs[2:4]
		dependencies, err = workflow.TestSeedRuleDependencies(ctx, parentIDs, childIDs, busDomain.Workflow)
		if err != nil {
			return workflowSeedData{}, fmt.Errorf("seeding rule dependencies : %w", err)
		}
	}

	// Seed automation executions
	executions, err := workflow.TestSeedAutomationExecutions(ctx, 10, ruleIDs, busDomain.Workflow)
	if err != nil {
		return workflowSeedData{}, fmt.Errorf("seeding automation executions : %w", err)
	}

	executionIDs := make([]uuid.UUID, len(executions))
	for i, e := range executions {
		executionIDs[i] = e.ID
	}

	// Seed notification deliveries
	notificationDeliveries, err := workflow.TestSeedNotificationDeliveries(ctx, 10, executionIDs, ruleIDs, actionIDs, userIDs, busDomain.Workflow)
	if err != nil {
		return workflowSeedData{}, fmt.Errorf("seeding notification deliveries : %w", err)
	}

	// -------------------------------------------------------------------------

	sd := workflowSeedData{
		SeedData: unitest.SeedData{
			Users:  []unitest.User{{User: usrs[0]}, {User: usrs[1]}},
			Admins: []unitest.User{{User: adminUser}},
		},
		TriggerTypes:           triggerTypes,
		EntityTypes:            entityTypes,
		Entities:               entities,
		AutomationRules:        rules,
		ActionTemplates:        templates,
		RuleActions:            actions,
		RuleDependencies:       dependencies,
		AutomationExecutions:   executions,
		NotificationDeliveries: notificationDeliveries,
	}

	return sd, nil
}

// =============================================================================
// Trigger Type Tests

func triggerTypeTests(busDomain dbtest.BusDomain, sd workflowSeedData) []unitest.Table {
	return []unitest.Table{
		queryTriggerTypes(busDomain, sd),
		createTriggerType(busDomain),
		updateTriggerType(busDomain, sd),
		deactivateTriggerType(busDomain, sd),
		activateTriggerType(busDomain, sd),
	}
}

func createTriggerType(busDomain dbtest.BusDomain) unitest.Table {
	return unitest.Table{
		Name: "create",
		ExpResp: workflow.TriggerType{
			Name:        "on_custom_event",
			Description: "Triggers on custom event",
		},
		ExcFunc: func(ctx context.Context) any {
			ntt := workflow.NewTriggerType{
				Name:        "on_custom_event",
				Description: "Triggers on custom event",
			}

			resp, err := busDomain.Workflow.CreateTriggerType(ctx, ntt)
			if err != nil {
				return err
			}

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.(workflow.TriggerType)
			if !exists {
				return "error occurred"
			}

			expResp := exp.(workflow.TriggerType)
			expResp.ID = gotResp.ID

			return cmp.Diff(gotResp, expResp)
		},
	}
}

func queryTriggerTypes(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	sort.Slice(sd.TriggerTypes, func(i, j int) bool {
		return sd.TriggerTypes[i].Name < sd.TriggerTypes[j].Name
	})

	return unitest.Table{
		Name:    "query",
		ExpResp: sd.TriggerTypes,
		ExcFunc: func(ctx context.Context) any {
			resp, err := busDomain.Workflow.QueryTriggerTypes(ctx)
			if err != nil {
				return err
			}

			sort.Slice(resp, func(i, j int) bool {
				return resp[i].Name < resp[j].Name
			})

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.([]workflow.TriggerType)
			if !exists {
				return "error occurred"
			}

			// Filter to only compare seeded data
			var filtered []workflow.TriggerType
			for _, g := range gotResp {
				for _, e := range sd.TriggerTypes {
					if g.ID == e.ID {
						filtered = append(filtered, g)
						break
					}
				}
			}

			return cmp.Diff(filtered, exp)
		},
	}
}

func updateTriggerType(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	return unitest.Table{
		Name: "update",
		ExpResp: workflow.TriggerType{
			ID:          sd.TriggerTypes[0].ID,
			Name:        "updated_trigger",
			Description: "Updated description",
			IsActive:    true,
		},
		ExcFunc: func(ctx context.Context) any {
			utt := workflow.UpdateTriggerType{
				Name:        dbtest.StringPointer("updated_trigger"),
				Description: dbtest.StringPointer("Updated description"),
			}

			resp, err := busDomain.Workflow.UpdateTriggerType(ctx, sd.TriggerTypes[0], utt)
			if err != nil {
				return err
			}

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.(workflow.TriggerType)
			if !exists {
				return "error occurred"
			}

			return cmp.Diff(gotResp, exp)
		},
	}
}

func deactivateTriggerType(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	return unitest.Table{
		Name:    "deactivate",
		ExpResp: nil,
		ExcFunc: func(ctx context.Context) any {
			if err := busDomain.Workflow.DeactivateTriggerType(ctx, sd.TriggerTypes[len(sd.TriggerTypes)-1]); err != nil {
				return err
			}

			return nil
		},
		CmpFunc: func(got any, exp any) string {
			return cmp.Diff(got, exp)
		},
	}
}

func activateTriggerType(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	return unitest.Table{
		Name:    "activate",
		ExpResp: nil,
		ExcFunc: func(ctx context.Context) any {
			// First deactivate it
			if err := busDomain.Workflow.DeactivateTriggerType(ctx, sd.TriggerTypes[len(sd.TriggerTypes)-1]); err != nil {
				return err
			}

			// Then activate it
			if err := busDomain.Workflow.ActivateTriggerType(ctx, sd.TriggerTypes[len(sd.TriggerTypes)-1]); err != nil {
				return err
			}

			return nil
		},
		CmpFunc: func(got any, exp any) string {
			return cmp.Diff(got, exp)
		},
	}
}

// =============================================================================
// Entity Type Tests

func entityTypeTests(busDomain dbtest.BusDomain, sd workflowSeedData) []unitest.Table {
	return []unitest.Table{
		createEntityType(busDomain),
		queryEntityTypes(busDomain, sd),
		updateEntityType(busDomain, sd),
		deactivateEntityType(busDomain, sd),
		activateEntityType(busDomain, sd),
	}
}

func createEntityType(busDomain dbtest.BusDomain) unitest.Table {
	return unitest.Table{
		Name: "create",
		ExpResp: workflow.EntityType{
			Name:        "custom_entity",
			Description: "Custom entity type",
		},
		ExcFunc: func(ctx context.Context) any {
			net := workflow.NewEntityType{
				Name:        "custom_entity",
				Description: "Custom entity type",
			}

			resp, err := busDomain.Workflow.CreateEntityType(ctx, net)
			if err != nil {
				return err
			}

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.(workflow.EntityType)
			if !exists {
				return "error occurred"
			}

			expResp := exp.(workflow.EntityType)
			expResp.ID = gotResp.ID

			return cmp.Diff(gotResp, expResp)
		},
	}
}

func queryEntityTypes(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	sort.Slice(sd.EntityTypes, func(i, j int) bool {
		return sd.EntityTypes[i].Name < sd.EntityTypes[j].Name
	})

	return unitest.Table{
		Name:    "query",
		ExpResp: sd.EntityTypes,
		ExcFunc: func(ctx context.Context) any {
			resp, err := busDomain.Workflow.QueryEntityTypes(ctx)
			if err != nil {
				return err
			}

			sort.Slice(resp, func(i, j int) bool {
				return resp[i].Name < resp[j].Name
			})

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.([]workflow.EntityType)
			if !exists {
				return "error occurred"
			}

			// Filter to only compare seeded data
			var filtered []workflow.EntityType
			for _, g := range gotResp {
				for _, e := range sd.EntityTypes {
					if g.ID == e.ID {
						filtered = append(filtered, g)
						break
					}
				}
			}

			return cmp.Diff(filtered, exp)
		},
	}
}

func updateEntityType(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	return unitest.Table{
		Name: "update",
		ExpResp: workflow.EntityType{
			ID:          sd.EntityTypes[0].ID,
			Name:        "updated_entity_type",
			Description: "Updated entity description",
			IsActive:    true,
		},
		ExcFunc: func(ctx context.Context) any {
			uet := workflow.UpdateEntityType{
				Name:        dbtest.StringPointer("updated_entity_type"),
				Description: dbtest.StringPointer("Updated entity description"),
			}

			resp, err := busDomain.Workflow.UpdateEntityType(ctx, sd.EntityTypes[0], uet)
			if err != nil {
				return err
			}

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.(workflow.EntityType)
			if !exists {
				return "error occurred"
			}

			return cmp.Diff(gotResp, exp)
		},
	}
}

func deactivateEntityType(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	return unitest.Table{
		Name:    "deactivate",
		ExpResp: nil,
		ExcFunc: func(ctx context.Context) any {
			if err := busDomain.Workflow.DeactivateEntityType(ctx, sd.EntityTypes[len(sd.EntityTypes)-1]); err != nil {
				return err
			}

			return nil
		},
		CmpFunc: func(got any, exp any) string {
			return cmp.Diff(got, exp)
		},
	}
}

func activateEntityType(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	return unitest.Table{
		Name:    "activate",
		ExpResp: nil,
		ExcFunc: func(ctx context.Context) any {
			// First deactivate it
			if err := busDomain.Workflow.DeactivateEntityType(ctx, sd.EntityTypes[len(sd.EntityTypes)-1]); err != nil {
				return err
			}

			// Then activate it
			if err := busDomain.Workflow.ActivateEntityType(ctx, sd.EntityTypes[len(sd.EntityTypes)-1]); err != nil {
				return err
			}

			return nil
		},
		CmpFunc: func(got any, exp any) string {
			return cmp.Diff(got, exp)
		},
	}
}

// =============================================================================
// Entity Tests

func entityTests(busDomain dbtest.BusDomain, sd workflowSeedData) []unitest.Table {
	return []unitest.Table{
		createEntity(busDomain, sd),
		queryEntities(busDomain, sd),
		updateEntity(busDomain, sd),
		deactivateEntity(busDomain, sd),
		activateEntity(busDomain, sd),
	}
}

func createEntity(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	return unitest.Table{
		Name: "create",
		ExpResp: workflow.Entity{
			Name:         "test_entity",
			EntityTypeID: sd.EntityTypes[0].ID,
			SchemaName:   "public",
			IsActive:     true,
		},
		ExcFunc: func(ctx context.Context) any {
			ne := workflow.NewEntity{
				Name:         "test_entity",
				EntityTypeID: sd.EntityTypes[0].ID,
				SchemaName:   "public",
				IsActive:     true,
			}

			resp, err := busDomain.Workflow.CreateEntity(ctx, ne)
			if err != nil {
				return err
			}

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.(workflow.Entity)
			if !exists {
				return "error occurred"
			}

			expResp := exp.(workflow.Entity)
			expResp.ID = gotResp.ID
			expResp.CreatedDate = gotResp.CreatedDate.Round(0).Truncate(time.Microsecond)
			gotResp.CreatedDate = gotResp.CreatedDate.Round(0).Truncate(time.Microsecond)

			return cmp.Diff(gotResp, expResp)
		},
	}
}

func queryEntities(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	sort.Slice(sd.Entities, func(i, j int) bool {
		return sd.Entities[i].Name < sd.Entities[j].Name
	})

	return unitest.Table{
		Name:    "query",
		ExpResp: sd.Entities,
		ExcFunc: func(ctx context.Context) any {
			resp, err := busDomain.Workflow.QueryEntities(ctx)
			if err != nil {
				return err
			}

			sort.Slice(resp, func(i, j int) bool {
				return resp[i].Name < resp[j].Name
			})

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.([]workflow.Entity)
			if !exists {
				return "error occurred"
			}

			// Filter to only compare seeded data
			var filtered []workflow.Entity
			for _, g := range gotResp {
				for _, e := range sd.Entities {
					if g.ID == e.ID {
						filtered = append(filtered, g)
						break
					}
				}
			}

			expResp := exp.([]workflow.Entity)
			for i := range filtered {
				for j := range expResp {
					if filtered[i].ID == expResp[j].ID {
						if filtered[i].CreatedDate.Format(time.RFC3339) == expResp[j].CreatedDate.Format(time.RFC3339) {
							expResp[j].CreatedDate = filtered[i].CreatedDate
						}
						break
					}
				}
			}

			return cmp.Diff(filtered, expResp)
		},
	}
}

func updateEntity(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	return unitest.Table{
		Name: "update",
		ExpResp: workflow.Entity{
			ID:           sd.Entities[0].ID,
			Name:         "updated_entity",
			EntityTypeID: sd.Entities[0].EntityTypeID,
			SchemaName:   "custom_schema",
			IsActive:     false,
			CreatedDate:  sd.Entities[0].CreatedDate,
		},
		ExcFunc: func(ctx context.Context) any {
			ue := workflow.UpdateEntity{
				Name:       dbtest.StringPointer("updated_entity"),
				SchemaName: dbtest.StringPointer("custom_schema"),
				IsActive:   dbtest.BoolPointer(false),
			}

			resp, err := busDomain.Workflow.UpdateEntity(ctx, sd.Entities[0], ue)
			if err != nil {
				return err
			}

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.(workflow.Entity)
			if !exists {
				return "error occurred"
			}
			expResp, exists := exp.(workflow.Entity)
			if !exists {
				return "error occurred"
			}

			expResp.CreatedDate = gotResp.CreatedDate.Round(0).Truncate(time.Microsecond)
			gotResp.CreatedDate = gotResp.CreatedDate.Round(0).Truncate(time.Microsecond)

			return cmp.Diff(gotResp, expResp)
		},
	}
}

func deactivateEntity(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	return unitest.Table{
		Name:    "deactivate",
		ExpResp: nil,
		ExcFunc: func(ctx context.Context) any {
			if err := busDomain.Workflow.DeactivateEntity(ctx, sd.Entities[len(sd.Entities)-1]); err != nil {
				return err
			}
			return nil
		},
		CmpFunc: func(got any, exp any) string {
			return cmp.Diff(got, exp)
		},
	}
}

func activateEntity(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	return unitest.Table{
		Name:    "activate",
		ExpResp: nil,
		ExcFunc: func(ctx context.Context) any {
			if err := busDomain.Workflow.ActivateEntity(ctx, sd.Entities[len(sd.Entities)-1]); err != nil {
				return err
			}
			return nil
		},
		CmpFunc: func(got any, exp any) string {
			return cmp.Diff(got, exp)
		},
	}
}

// =============================================================================
// Automation Rule Tests

func automationRuleTests(busDomain dbtest.BusDomain, sd workflowSeedData) []unitest.Table {
	return []unitest.Table{
		createAutomationRule(busDomain, sd),
		queryAutomationRuleByID(busDomain, sd),
		queryAutomationRulesByEntity(busDomain, sd),
		updateAutomationRule(busDomain, sd),
		deactivateAutomationRule(busDomain, sd),
		activateAutomationRule(busDomain, sd),
	}
}

func createAutomationRule(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	conditions := map[string]interface{}{
		"field": "status",
		"value": "active",
	}
	conditionsJSON, _ := json.Marshal(conditions)
	triggerConditions := json.RawMessage(conditionsJSON)

	return unitest.Table{
		Name: "create",
		ExpResp: workflow.AutomationRule{
			Name:              "test_rule",
			Description:       "Test automation rule",
			EntityID:          sd.Entities[1].ID,
			EntityTypeID:      sd.EntityTypes[0].ID,
			TriggerTypeID:     sd.TriggerTypes[0].ID,
			TriggerConditions: &triggerConditions,
			IsActive:          true,
			CreatedBy:         sd.Admins[0].ID,
			UpdatedBy:         sd.Admins[0].ID,
		},
		ExcFunc: func(ctx context.Context) any {
			nar := workflow.NewAutomationRule{
				Name:              "test_rule",
				Description:       "Test automation rule",
				EntityID:          sd.Entities[1].ID,
				EntityTypeID:      sd.EntityTypes[0].ID,
				TriggerTypeID:     sd.TriggerTypes[0].ID,
				TriggerConditions: &triggerConditions,
				IsActive:          true,
				CreatedBy:         sd.Admins[0].ID,
			}

			resp, err := busDomain.Workflow.CreateRule(ctx, nar)
			if err != nil {
				return err
			}

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.(workflow.AutomationRule)
			if !exists {
				return "error occurred"
			}

			expResp := exp.(workflow.AutomationRule)
			expResp.ID = gotResp.ID
			expResp.CreatedDate = gotResp.CreatedDate
			expResp.UpdatedDate = gotResp.UpdatedDate

			return cmp.Diff(gotResp, expResp)
		},
	}
}

func queryAutomationRuleByID(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	return unitest.Table{
		Name:    "queryByID",
		ExpResp: sd.AutomationRules[0],
		ExcFunc: func(ctx context.Context) any {
			resp, err := busDomain.Workflow.QueryRuleByID(ctx, sd.AutomationRules[0].ID)
			if err != nil {
				return err
			}

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.(workflow.AutomationRule)
			if !exists {
				return "error occurred"
			}

			expResp := exp.(workflow.AutomationRule)

			if gotResp.CreatedDate.Format(time.RFC3339) == expResp.CreatedDate.Format(time.RFC3339) {
				expResp.CreatedDate = gotResp.CreatedDate
			}

			if gotResp.UpdatedDate.Format(time.RFC3339) == expResp.UpdatedDate.Format(time.RFC3339) {
				expResp.UpdatedDate = gotResp.UpdatedDate
			}

			dbtest.NormalizeJSONFields(gotResp, &expResp)

			return cmp.Diff(gotResp, expResp)
		},
	}
}

func queryAutomationRulesByEntity(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	// Find rules for first entity
	// Initialize as empty slice to match QueryRulesByEntity return type
	expectedRules := []workflow.AutomationRule{}
	for _, rule := range sd.AutomationRules {
		if rule.EntityID == sd.Entities[0].ID {
			expectedRules = append(expectedRules, rule)
		}
	}

	return unitest.Table{
		Name:    "queryByEntity",
		ExpResp: expectedRules,
		ExcFunc: func(ctx context.Context) any {
			resp, err := busDomain.Workflow.QueryRulesByEntity(ctx, sd.Entities[0].ID)
			if err != nil {
				return err
			}

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.([]workflow.AutomationRule)
			if !exists {
				return "error occurred"
			}

			expResp := exp.([]workflow.AutomationRule)

			for i := range gotResp {
				for j := range expResp {
					if gotResp[i].ID == expResp[j].ID {
						if gotResp[i].CreatedDate.Format(time.RFC3339) == expResp[j].CreatedDate.Format(time.RFC3339) {
							expResp[j].CreatedDate = gotResp[i].CreatedDate
						}
						if gotResp[i].UpdatedDate.Format(time.RFC3339) == expResp[j].UpdatedDate.Format(time.RFC3339) {
							expResp[j].UpdatedDate = gotResp[i].UpdatedDate
						}
						break
					}
				}
			}

			dbtest.NormalizeJSONFields(gotResp, &expResp)

			return cmp.Diff(gotResp, expResp)
		},
	}
}

func updateAutomationRule(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	newConditions := map[string]interface{}{
		"field": "type",
		"value": "updated",
	}
	newConditionsJSON, _ := json.Marshal(newConditions)
	raw := json.RawMessage(newConditionsJSON)

	return unitest.Table{
		Name: "update",
		ExpResp: workflow.AutomationRule{
			ID:                sd.AutomationRules[0].ID,
			Name:              "updated_rule",
			Description:       "Updated description",
			EntityID:          sd.AutomationRules[0].EntityID,
			EntityTypeID:      sd.AutomationRules[0].EntityTypeID,
			TriggerTypeID:     sd.AutomationRules[0].TriggerTypeID,
			TriggerConditions: &raw,
			IsActive:          false,
			CreatedDate:       sd.AutomationRules[0].CreatedDate,
			CreatedBy:         sd.AutomationRules[0].CreatedBy,
			UpdatedBy:         sd.Admins[0].ID,
		},
		ExcFunc: func(ctx context.Context) any {
			uar := workflow.UpdateAutomationRule{
				Name:              dbtest.StringPointer("updated_rule"),
				Description:       dbtest.StringPointer("Updated description"),
				TriggerConditions: &raw,
				IsActive:          dbtest.BoolPointer(false),
				UpdatedBy:         &sd.Admins[0].ID,
			}

			resp, err := busDomain.Workflow.UpdateRule(ctx, sd.AutomationRules[0], uar)
			if err != nil {
				return err
			}

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.(workflow.AutomationRule)
			if !exists {
				return "error occurred"
			}

			expResp := exp.(workflow.AutomationRule)
			expResp.UpdatedDate = gotResp.UpdatedDate

			return cmp.Diff(gotResp, expResp)
		},
	}
}

func deactivateAutomationRule(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	return unitest.Table{
		Name:    "deactivate",
		ExpResp: nil,
		ExcFunc: func(ctx context.Context) any {
			if err := busDomain.Workflow.DeactivateRule(ctx, sd.AutomationRules[0]); err != nil {
				return err
			}
			return nil
		},
		CmpFunc: func(got any, exp any) string {
			return cmp.Diff(got, exp)
		},
	}
}

func activateAutomationRule(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	return unitest.Table{
		Name:    "activate",
		ExpResp: nil,
		ExcFunc: func(ctx context.Context) any {
			if err := busDomain.Workflow.ActivateRule(ctx, sd.AutomationRules[0]); err != nil {
				return err
			}
			return nil
		},
		CmpFunc: func(got any, exp any) string {
			return cmp.Diff(got, exp)
		},
	}
}

// =============================================================================
// Action Template Tests

func actionTemplateTests(busDomain dbtest.BusDomain, sd workflowSeedData) []unitest.Table {
	return []unitest.Table{
		createActionTemplate(busDomain, sd),
		queryActionTemplateByID(busDomain, sd),
		updateActionTemplate(busDomain, sd),
		deactivateActionTemplate(busDomain, sd),
		activateActionTemplate(busDomain, sd),
	}
}

func createActionTemplate(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	config := map[string]interface{}{
		"endpoint": "https://api.test.com",
		"method":   "POST",
	}
	configJSON, _ := json.Marshal(config)

	return unitest.Table{
		Name: "create",
		ExpResp: workflow.ActionTemplate{
			Name:          "test_template",
			Description:   "Test action template",
			ActionType:    "webhook",
			DefaultConfig: configJSON,
			CreatedBy:     sd.Admins[0].ID,
		},
		ExcFunc: func(ctx context.Context) any {
			nat := workflow.NewActionTemplate{
				Name:          "test_template",
				Description:   "Test action template",
				ActionType:    "webhook",
				DefaultConfig: configJSON,
				CreatedBy:     sd.Admins[0].ID,
			}

			resp, err := busDomain.Workflow.CreateActionTemplate(ctx, nat)
			if err != nil {
				return err
			}

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.(workflow.ActionTemplate)
			if !exists {
				return "error occurred"
			}

			expResp := exp.(workflow.ActionTemplate)
			expResp.ID = gotResp.ID
			expResp.CreatedDate = gotResp.CreatedDate

			return cmp.Diff(gotResp, expResp)
		},
	}
}

func queryActionTemplateByID(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	return unitest.Table{
		Name:    "queryByID",
		ExpResp: sd.ActionTemplates[2],
		ExcFunc: func(ctx context.Context) any {
			resp, err := busDomain.Workflow.QueryTemplateByID(ctx, sd.ActionTemplates[2].ID)
			if err != nil {
				return err
			}

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.(workflow.ActionTemplate)
			if !exists {
				return "error occurred"
			}

			expResp := exp.(workflow.ActionTemplate)

			if gotResp.CreatedDate.Format(time.RFC3339) == expResp.CreatedDate.Format(time.RFC3339) {
				expResp.CreatedDate = gotResp.CreatedDate
			}

			dbtest.NormalizeJSONFields(gotResp, &expResp)

			return cmp.Diff(gotResp, expResp)
		},
	}
}

func updateActionTemplate(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	newConfig := map[string]interface{}{
		"endpoint": "https://api.updated.com",
		"method":   "PUT",
	}
	newConfigJSON, _ := json.Marshal(newConfig)
	raw := json.RawMessage(newConfigJSON)

	return unitest.Table{
		Name: "update",
		ExpResp: workflow.ActionTemplate{
			ID:            sd.ActionTemplates[0].ID,
			Name:          "updated_template",
			Description:   "Updated template description",
			ActionType:    "api_call",
			DefaultConfig: newConfigJSON,
			CreatedDate:   sd.ActionTemplates[0].CreatedDate,
			CreatedBy:     sd.ActionTemplates[0].CreatedBy,
		},
		ExcFunc: func(ctx context.Context) any {
			uat := workflow.UpdateActionTemplate{
				Name:          dbtest.StringPointer("updated_template"),
				Description:   dbtest.StringPointer("Updated template description"),
				ActionType:    dbtest.StringPointer("api_call"),
				DefaultConfig: &raw,
			}

			resp, err := busDomain.Workflow.UpdateActionTemplate(ctx, sd.ActionTemplates[0], uat)
			if err != nil {
				return err
			}

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.(workflow.ActionTemplate)
			if !exists {
				return "error occurred"
			}

			return cmp.Diff(gotResp, exp)
		},
	}
}

func deactivateActionTemplate(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	return unitest.Table{
		Name:    "deactivate",
		ExpResp: nil,
		ExcFunc: func(ctx context.Context) any {
			if err := busDomain.Workflow.DeactivateActionTemplate(ctx, sd.ActionTemplates[0].ID, sd.Admins[0].ID); err != nil {
				return err
			}
			return nil
		},
		CmpFunc: func(got any, exp any) string {
			return cmp.Diff(got, exp)
		},
	}
}

func activateActionTemplate(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	return unitest.Table{
		Name:    "activate",
		ExpResp: nil,
		ExcFunc: func(ctx context.Context) any {
			if err := busDomain.Workflow.ActivateActionTemplate(ctx, sd.ActionTemplates[0].ID, sd.Admins[0].ID); err != nil {
				return err
			}
			return nil
		},
		CmpFunc: func(got any, exp any) string {
			return cmp.Diff(got, exp)
		},
	}
}

// =============================================================================
// Rule Action Tests

func ruleActionTests(busDomain dbtest.BusDomain, sd workflowSeedData) []unitest.Table {
	return []unitest.Table{
		createRuleAction(busDomain, sd),
		queryActionsByRule(busDomain, sd),
		updateRuleAction(busDomain, sd),
		deactivateRuleAction(busDomain, sd),
		activateRuleAction(busDomain, sd),
	}
}

func createRuleAction(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	config := map[string]interface{}{
		"action": "send_email",
		"to":     "test@example.com",
	}
	configJSON, _ := json.Marshal(config)

	return unitest.Table{
		Name: "create",
		ExpResp: workflow.RuleAction{
			AutomationRuleID: sd.AutomationRules[1].ID,
			Name:             "test_action",
			Description:      "Test rule action",
			ActionConfig:     configJSON,
			ExecutionOrder:   99,
			IsActive:         true,
			TemplateID:       &sd.ActionTemplates[0].ID,
		},
		ExcFunc: func(ctx context.Context) any {
			nra := workflow.NewRuleAction{
				AutomationRuleID: sd.AutomationRules[1].ID,
				Name:             "test_action",
				Description:      "Test rule action",
				ActionConfig:     configJSON,
				ExecutionOrder:   99,
				IsActive:         true,
				TemplateID:       &sd.ActionTemplates[0].ID,
			}

			resp, err := busDomain.Workflow.CreateRuleAction(ctx, nra)
			if err != nil {
				return err
			}

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.(workflow.RuleAction)
			if !exists {
				return "error occurred"
			}

			expResp := exp.(workflow.RuleAction)
			expResp.ID = gotResp.ID

			return cmp.Diff(gotResp, expResp)
		},
	}
}

func queryActionsByRule(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	// Find actions for first rule
	var expectedActions []workflow.RuleAction
	for _, action := range sd.RuleActions {
		if action.AutomationRuleID == sd.AutomationRules[0].ID {
			expectedActions = append(expectedActions, action)
		}
	}

	sort.Slice(expectedActions, func(i, j int) bool {
		return expectedActions[i].ExecutionOrder < expectedActions[j].ExecutionOrder
	})

	return unitest.Table{
		Name:    "queryByRule",
		ExpResp: expectedActions,
		ExcFunc: func(ctx context.Context) any {
			resp, err := busDomain.Workflow.QueryActionsByRule(ctx, sd.AutomationRules[0].ID)
			if err != nil {
				return err
			}

			sort.Slice(resp, func(i, j int) bool {
				return resp[i].ExecutionOrder < resp[j].ExecutionOrder
			})

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.([]workflow.RuleAction)
			if !exists {
				return "error occurred"
			}
			expResp := exp.([]workflow.RuleAction)

			dbtest.NormalizeJSONFields(gotResp, &expResp)

			return cmp.Diff(gotResp, expResp)
		},
	}
}

func updateRuleAction(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	newConfig := map[string]interface{}{
		"action": "webhook",
		"url":    "https://updated.example.com",
	}
	newConfigJSON, _ := json.Marshal(newConfig)
	raw := json.RawMessage(newConfigJSON)

	return unitest.Table{
		Name: "update",
		ExpResp: workflow.RuleAction{
			ID:               sd.RuleActions[0].ID,
			AutomationRuleID: sd.RuleActions[0].AutomationRuleID,
			Name:             "updated_action",
			Description:      "Updated action description",
			ActionConfig:     newConfigJSON,
			ExecutionOrder:   50,
			IsActive:         false,
			TemplateID:       dbtest.UUIDPointer(sd.ActionTemplates[1].ID),
		},
		ExcFunc: func(ctx context.Context) any {
			ura := workflow.UpdateRuleAction{
				Name:           dbtest.StringPointer("updated_action"),
				Description:    dbtest.StringPointer("Updated action description"),
				ActionConfig:   &raw,
				ExecutionOrder: dbtest.IntPointer(50),
				IsActive:       dbtest.BoolPointer(false),
				TemplateID:     dbtest.UUIDPointer(sd.ActionTemplates[1].ID), //
			}

			resp, err := busDomain.Workflow.UpdateRuleAction(ctx, sd.RuleActions[0], ura)
			if err != nil {
				return err
			}

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.(workflow.RuleAction)
			if !exists {
				return "error occurred"
			}

			return cmp.Diff(gotResp, exp)
		},
	}
}

func deactivateRuleAction(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	return unitest.Table{
		Name:    "deactivate",
		ExpResp: nil,
		ExcFunc: func(ctx context.Context) any {
			if err := busDomain.Workflow.DeactivateRuleAction(ctx, sd.RuleActions[1]); err != nil {
				return err
			}
			return nil
		},
		CmpFunc: func(got any, exp any) string {
			return cmp.Diff(got, exp)
		},
	}
}

func activateRuleAction(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	return unitest.Table{
		Name:    "activate",
		ExpResp: nil,
		ExcFunc: func(ctx context.Context) any {
			if err := busDomain.Workflow.ActivateRuleAction(ctx, sd.RuleActions[1]); err != nil {
				return err
			}
			return nil
		},
		CmpFunc: func(got any, exp any) string {
			return cmp.Diff(got, exp)
		},
	}
}

// =============================================================================
// Rule Dependency Tests

func ruleDependencyTests(busDomain dbtest.BusDomain, sd workflowSeedData) []unitest.Table {
	return []unitest.Table{
		createRuleDependency(busDomain, sd),
		queryRuleDependencies(busDomain, sd),
		deleteRuleDependency(busDomain, sd),
	}
}

func createRuleDependency(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	// Use rules that don't have dependencies yet
	parentID := sd.AutomationRules[3].ID
	childID := sd.AutomationRules[4].ID

	return unitest.Table{
		Name: "create",
		ExpResp: workflow.RuleDependency{
			ParentRuleID: parentID,
			ChildRuleID:  childID,
		},
		ExcFunc: func(ctx context.Context) any {
			nrd := workflow.NewRuleDependency{
				ParentRuleID: parentID,
				ChildRuleID:  childID,
			}

			resp, err := busDomain.Workflow.CreateDependency(ctx, nrd)
			if err != nil {
				return err
			}

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.(workflow.RuleDependency)
			if !exists {
				return "error occurred"
			}
			expResp, exists := exp.(workflow.RuleDependency)
			if !exists {
				return "error occurred"
			}

			expResp.ID = gotResp.ID

			return cmp.Diff(gotResp, expResp)
		},
	}
}

func queryRuleDependencies(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	return unitest.Table{
		Name:    "query",
		ExpResp: sd.RuleDependencies,
		ExcFunc: func(ctx context.Context) any {
			resp, err := busDomain.Workflow.QueryDependencies(ctx)
			if err != nil {
				return err
			}

			// Filter to only seeded dependencies
			var filtered []workflow.RuleDependency
			for _, r := range resp {
				for _, s := range sd.RuleDependencies {
					if r.ParentRuleID == s.ParentRuleID && r.ChildRuleID == s.ChildRuleID {
						filtered = append(filtered, r)
						break
					}
				}
			}

			return filtered
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.([]workflow.RuleDependency)
			if !exists {
				return "error occurred"
			}

			return cmp.Diff(gotResp, exp)
		},
	}
}

func deleteRuleDependency(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	return unitest.Table{
		Name:    "delete",
		ExpResp: nil,
		ExcFunc: func(ctx context.Context) any {
			if len(sd.RuleDependencies) > 0 {
				if err := busDomain.Workflow.DeleteDependency(ctx, sd.RuleDependencies[0]); err != nil {
					return err
				}
			}
			return nil
		},
		CmpFunc: func(got any, exp any) string {
			return cmp.Diff(got, exp)
		},
	}
}

// =============================================================================
// Automation Execution Tests

func automationExecutionTests(busDomain dbtest.BusDomain, sd workflowSeedData) []unitest.Table {
	return []unitest.Table{
		createAutomationExecution(busDomain, sd),
		queryExecutionHistory(busDomain, sd),
	}
}

func createAutomationExecution(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	triggerData := map[string]interface{}{
		"entity_id": "test_entity",
		"event":     "created",
	}
	triggerDataJSON, _ := json.Marshal(triggerData)

	actionsExecuted := []map[string]interface{}{
		{
			"action_id": "action_1",
			"status":    "completed",
		},
	}
	actionsExecutedJSON, _ := json.Marshal(actionsExecuted)

	ruleID := sd.AutomationRules[0].ID
	return unitest.Table{
		Name: "create",
		ExpResp: workflow.AutomationExecution{
			AutomationRuleID: &ruleID,
			EntityType:       "test_entity_type",
			TriggerData:      triggerDataJSON,
			ActionsExecuted:  actionsExecutedJSON,
			Status:           workflow.StatusCompleted,
			ErrorMessage:     "",
			ExecutionTimeMs:  250,
			TriggerSource:    workflow.TriggerSourceAutomation,
		},
		ExcFunc: func(ctx context.Context) any {
			nae := workflow.NewAutomationExecution{
				AutomationRuleID: &ruleID,
				EntityType:       "test_entity_type",
				TriggerData:      triggerDataJSON,
				ActionsExecuted:  actionsExecutedJSON,
				Status:           workflow.StatusCompleted,
				ErrorMessage:     "",
				ExecutionTimeMs:  250,
				TriggerSource:    workflow.TriggerSourceAutomation,
			}

			resp, err := busDomain.Workflow.CreateExecution(ctx, nae)
			if err != nil {
				return err
			}

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.(workflow.AutomationExecution)
			if !exists {
				return "error occurred"
			}

			expResp := exp.(workflow.AutomationExecution)
			expResp.ID = gotResp.ID
			expResp.ExecutedAt = gotResp.ExecutedAt

			return cmp.Diff(gotResp, expResp)
		},
	}
}

func queryExecutionHistory(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	// Find executions for first rule
	targetRuleID := sd.AutomationRules[0].ID
	var expectedExecutions []workflow.AutomationExecution
	for _, exec := range sd.AutomationExecutions {
		if exec.AutomationRuleID != nil && *exec.AutomationRuleID == targetRuleID {
			expectedExecutions = append(expectedExecutions, exec)
			if len(expectedExecutions) >= 5 {
				break
			}
		}
	}

	return unitest.Table{
		Name:    "queryHistory",
		ExpResp: expectedExecutions,
		ExcFunc: func(ctx context.Context) any {
			resp, err := busDomain.Workflow.QueryExecutionHistory(ctx, sd.AutomationRules[0].ID, 5)
			if err != nil {
				return err
			}

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.([]workflow.AutomationExecution)
			if !exists {
				return "error occurred"
			}

			expResp := exp.([]workflow.AutomationExecution)

			// Match up the executions by ID and update timestamps
			for i := range gotResp {
				for j := range expResp {
					if gotResp[i].ID == expResp[j].ID {
						if gotResp[i].ExecutedAt.Format(time.RFC3339) == expResp[j].ExecutedAt.Format(time.RFC3339) {
							expResp[j].ExecutedAt = gotResp[i].ExecutedAt
						}
						break
					}
				}
			}

			// Only compare what we expect
			if len(gotResp) > len(expResp) {
				gotResp = gotResp[:len(expResp)]
			}

			dbtest.NormalizeJSONFields(gotResp, &expResp)

			return cmp.Diff(gotResp, expResp)
		},
	}
}

// =============================================================================
// Notification Delivery Tests
func notificationDeliveryTests(busDomain dbtest.BusDomain, sd workflowSeedData) []unitest.Table {
	return []unitest.Table{
		createNotificationDelivery(busDomain, sd),
		updateNotificationDelivery(busDomain, sd),
		queryNotificationDeliveriesByAutomationExecution(busDomain, sd),
	}
}

func createNotificationDelivery(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	return unitest.Table{

		Name: "create",
		ExpResp: workflow.NotificationDelivery{

			NotificationID:        uuid.New(),
			AutomationExecutionID: sd.AutomationExecutions[0].ID,
			RuleID:                sd.AutomationRules[0].ID,
			ActionID:              sd.RuleActions[0].ID,
			RecipientID:           sd.Users[0].ID,
			Channel:               "email",   // TODO: Make these constants
			Status:                "pending", // TODO: Make these constants
			Attempts:              0,
			SentAt:                nil,
			DeliveredAt:           nil,
			FailedAt:              nil,
			ErrorMessage:          dbtest.StringPointer(""),
			ProviderResponse:      json.RawMessage(`{}`),
		},
		ExcFunc: func(ctx context.Context) any {
			nd := workflow.NewNotificationDelivery{
				AutomationExecutionID: sd.AutomationExecutions[0].ID,
				RuleID:                sd.AutomationRules[0].ID,
				ActionID:              sd.RuleActions[0].ID,
				RecipientID:           sd.Users[0].ID,
				Channel:               "email",   // TODO: Make these constants
				Status:                "pending", // TODO: Make these constants
				Attempts:              0,
				SentAt:                nil,
				DeliveredAt:           nil,
				FailedAt:              nil,
				ErrorMessage:          dbtest.StringPointer(""),
				ProviderResponse:      json.RawMessage(`{}`),
			}

			resp, err := busDomain.Workflow.CreateNotificationDelivery(ctx, nd)
			if err != nil {
				return err
			}

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.(workflow.NotificationDelivery)
			if !exists {
				return "error occurred"
			}

			expResp := exp.(workflow.NotificationDelivery)
			expResp.ID = gotResp.ID
			expResp.DeliveredAt = gotResp.DeliveredAt
			expResp.NotificationID = gotResp.NotificationID
			expResp.CreatedDate = gotResp.CreatedDate.Round(0).Truncate(time.Microsecond)
			expResp.UpdatedDate = gotResp.UpdatedDate.Round(0).Truncate(time.Microsecond)

			return cmp.Diff(gotResp, expResp)
		},
	}
}

func updateNotificationDelivery(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {
	now := time.Now().UTC()
	return unitest.Table{
		Name: "update",
		ExpResp: workflow.NotificationDelivery{
			ID:                    sd.NotificationDeliveries[0].ID,
			NotificationID:        sd.NotificationDeliveries[0].NotificationID,
			AutomationExecutionID: sd.NotificationDeliveries[0].AutomationExecutionID,
			RuleID:                sd.NotificationDeliveries[0].RuleID,
			ActionID:              sd.NotificationDeliveries[0].ActionID,
			RecipientID:           sd.NotificationDeliveries[0].RecipientID,
			Channel:               sd.NotificationDeliveries[0].Channel,
			Status:                "sent",
			Attempts:              1,
			SentAt:                &now,
			CreatedDate:           sd.NotificationDeliveries[0].CreatedDate,
			ErrorMessage:          sd.NotificationDeliveries[0].ErrorMessage,
			ProviderResponse:      sd.NotificationDeliveries[0].ProviderResponse,
		},
		ExcFunc: func(ctx context.Context) any {
			ud := workflow.UpdateNotificationDelivery{
				Status:   (*workflow.DeliveryStatus)(dbtest.StringPointer("sent")),
				Attempts: dbtest.IntPointer(1),
				SentAt:   &now,
			}

			ret, err := busDomain.Workflow.UpdateNotificationDelivery(ctx, sd.NotificationDeliveries[0], ud)
			if err != nil {
				return err
			}

			return ret
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, ok := got.(workflow.NotificationDelivery)
			if !ok {
				return "type assertion failed"
			}

			expResp, ok := exp.(workflow.NotificationDelivery)
			if !ok {
				return "type assertion failed"
			}

			// Normalize timestamps for comparison
			expResp.CreatedDate = gotResp.CreatedDate
			expResp.UpdatedDate = gotResp.UpdatedDate

			return cmp.Diff(gotResp, expResp)
		},
	}
}

func queryNotificationDeliveriesByAutomationExecution(busDomain dbtest.BusDomain, sd workflowSeedData) unitest.Table {

	var expectedDeliveries []workflow.NotificationDelivery
	for i := range sd.NotificationDeliveries {
		// Use index instead of value to avoid reuse issues
		if sd.NotificationDeliveries[i].AutomationExecutionID == sd.AutomationExecutions[2].ID {
			// Make a proper copy
			delivery := sd.NotificationDeliveries[i]
			expectedDeliveries = append(expectedDeliveries, delivery)
		}
	}

	return unitest.Table{
		Name:    "queryByAutomationExecution",
		ExpResp: expectedDeliveries,
		ExcFunc: func(ctx context.Context) any {
			resp, err := busDomain.Workflow.QueryDeliveriesByAutomationExecution(ctx, sd.AutomationExecutions[2].ID)
			if err != nil {
				return err
			}

			return resp
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, exists := got.([]workflow.NotificationDelivery)
			if !exists {
				return "error occurred"
			}

			expResp := exp.([]workflow.NotificationDelivery)

			// Only compare what we expect
			if len(gotResp) > len(expResp) {
				gotResp = gotResp[:len(expResp)]
			}

			dbtest.NormalizeJSONFields(gotResp, &expResp)

			return cmp.Diff(gotResp, expResp)
		},
	}
}
