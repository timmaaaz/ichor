package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
)

// TestNewTriggerTypes is a helper method for testing.
func TestNewTriggerTypes(n int) []NewTriggerType {
	triggerTypes := make([]NewTriggerType, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		tt := NewTriggerType{
			Name:        fmt.Sprintf("TriggerType%d", idx),
			Description: fmt.Sprintf("Description for trigger type %d", idx),
		}

		triggerTypes[i] = tt
	}

	return triggerTypes
}

// TestSeedTriggerTypes is a helper method for testing.
func TestSeedTriggerTypes(ctx context.Context, n int, api *Business) ([]TriggerType, error) {
	newTriggerTypes := TestNewTriggerTypes(n)

	triggerTypes := make([]TriggerType, len(newTriggerTypes))
	for i, ntt := range newTriggerTypes {
		tt, err := api.CreateTriggerType(ctx, ntt)
		if err != nil {
			return nil, fmt.Errorf("seeding trigger type: idx: %d : %w", i, err)
		}

		triggerTypes[i] = tt
	}

	return triggerTypes, nil
}

// TestNewEntityTypes is a helper method for testing.
func TestNewEntityTypes(n int) []NewEntityType {
	entityTypes := make([]NewEntityType, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		et := NewEntityType{
			Name:        fmt.Sprintf("EntityType%d", idx),
			Description: fmt.Sprintf("Description for entity type %d", idx),
		}

		entityTypes[i] = et
	}

	return entityTypes
}

// TestSeedEntityTypes is a helper method for testing.
func TestSeedEntityTypes(ctx context.Context, n int, api *Business) ([]EntityType, error) {
	newEntityTypes := TestNewEntityTypes(n)

	entityTypes := make([]EntityType, len(newEntityTypes))
	for i, net := range newEntityTypes {
		et, err := api.CreateEntityType(ctx, net)
		if err != nil {
			return nil, fmt.Errorf("seeding entity type: idx: %d : %w", i, err)
		}

		entityTypes[i] = et
	}

	return entityTypes, nil
}

// TestNewEntities is a helper method for testing.
func TestNewEntities(n int, entityTypeIDs []uuid.UUID) []NewEntity {
	entities := make([]NewEntity, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		e := NewEntity{
			Name:         fmt.Sprintf("Entity%d", idx),
			EntityTypeID: entityTypeIDs[i%len(entityTypeIDs)],
			SchemaName:   "public",
			IsActive:     true,
		}

		entities[i] = e
	}

	return entities
}

// TestSeedEntities is a helper method for testing.
func TestSeedEntities(ctx context.Context, n int, entityTypeIDs []uuid.UUID, api *Business) ([]Entity, error) {
	newEntities := TestNewEntities(n, entityTypeIDs)

	entities := make([]Entity, len(newEntities))
	for i, ne := range newEntities {
		e, err := api.CreateEntity(ctx, ne)
		if err != nil {
			return nil, fmt.Errorf("seeding entity: idx: %d : %w", i, err)
		}

		entities[i] = e
	}

	return entities, nil
}

// TestNewAutomationRules is a helper method for testing.
func TestNewAutomationRules(n int, entityIDs, entityTypeIDs, triggerTypeIDs []uuid.UUID, createdBy uuid.UUID) []NewAutomationRule {
	rules := make([]NewAutomationRule, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		// Create sample trigger conditions
		conditions := map[string]interface{}{
			"field":    fmt.Sprintf("field_%d", idx),
			"operator": "equals",
			"value":    fmt.Sprintf("value_%d", idx),
		}
		conditionsJSON, _ := json.Marshal(conditions)

		rule := NewAutomationRule{
			Name:              fmt.Sprintf("Rule%d", idx),
			Description:       fmt.Sprintf("Description for rule %d", idx),
			EntityID:          entityIDs[i%len(entityIDs)],
			EntityTypeID:      entityTypeIDs[i%len(entityTypeIDs)],
			TriggerTypeID:     triggerTypeIDs[i%len(triggerTypeIDs)],
			TriggerConditions: conditionsJSON,
			IsActive:          true,
			CreatedBy:         createdBy,
		}

		rules[i] = rule
	}

	return rules
}

// TestSeedAutomationRules is a helper method for testing.
func TestSeedAutomationRules(ctx context.Context, n int, entityIDs, entityTypeIDs, triggerTypeIDs []uuid.UUID, createdBy uuid.UUID, api *Business) ([]AutomationRule, error) {
	newRules := TestNewAutomationRules(n, entityIDs, entityTypeIDs, triggerTypeIDs, createdBy)

	rules := make([]AutomationRule, len(newRules))
	for i, nar := range newRules {
		rule, err := api.CreateRule(ctx, nar)
		if err != nil {
			return nil, fmt.Errorf("seeding automation rule: idx: %d : %w", i, err)
		}

		rules[i] = rule
	}

	return rules, nil
}

// TestNewActionTemplates is a helper method for testing.
func TestNewActionTemplates(n int, createdBy uuid.UUID) []NewActionTemplate {
	templates := make([]NewActionTemplate, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		// Create sample default config
		config := map[string]interface{}{
			"endpoint": fmt.Sprintf("https://api.example.com/action%d", idx),
			"method":   "POST",
			"timeout":  30,
		}
		configJSON, _ := json.Marshal(config)

		template := NewActionTemplate{
			Name:          fmt.Sprintf("Template%d", idx),
			Description:   fmt.Sprintf("Description for template %d", idx),
			ActionType:    fmt.Sprintf("ActionType%d", idx),
			DefaultConfig: configJSON,
			CreatedBy:     createdBy,
		}

		templates[i] = template
	}

	return templates
}

// TestSeedActionTemplates is a helper method for testing.
func TestSeedActionTemplates(ctx context.Context, n int, createdBy uuid.UUID, api *Business) ([]ActionTemplate, error) {
	newTemplates := TestNewActionTemplates(n, createdBy)

	templates := make([]ActionTemplate, len(newTemplates))
	for i, nat := range newTemplates {
		template, err := api.CreateActionTemplate(ctx, nat)
		if err != nil {
			return nil, fmt.Errorf("seeding action template: idx: %d : %w", i, err)
		}

		templates[i] = template
	}

	return templates, nil
}

// TestNewRuleActions is a helper method for testing.
func TestNewRuleActions(n int, ruleIDs []uuid.UUID, templateIDs *[]uuid.UUID) []NewRuleAction {
	actions := make([]NewRuleAction, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		// Create sample action config
		config := map[string]interface{}{
			"action_param": fmt.Sprintf("param_%d", idx),
			"enabled":      true,
		}
		configJSON, _ := json.Marshal(config)

		action := NewRuleAction{
			AutomationRuleID: ruleIDs[i%len(ruleIDs)],
			Name:             fmt.Sprintf("Action%d", idx),
			Description:      fmt.Sprintf("Description for action %d", idx),
			ActionConfig:     configJSON,
			ExecutionOrder:   i + 1,
			IsActive:         true,
		}

		// Optionally add template ID
		if templateIDs != nil && len(*templateIDs) > 0 {
			tid := (*templateIDs)[i%len(*templateIDs)]
			action.TemplateID = &tid
		}

		actions[i] = action
	}

	return actions
}

// TestSeedRuleActions is a helper method for testing.
func TestSeedRuleActions(ctx context.Context, n int, ruleIDs []uuid.UUID, templateIDs *[]uuid.UUID, api *Business) ([]RuleAction, error) {
	newActions := TestNewRuleActions(n, ruleIDs, templateIDs)

	actions := make([]RuleAction, len(newActions))
	for i, nra := range newActions {
		action, err := api.CreateRuleAction(ctx, nra)
		if err != nil {
			return nil, fmt.Errorf("seeding rule action: idx: %d : %w", i, err)
		}

		actions[i] = action
	}

	return actions, nil
}

// TestNewRuleDependencies creates test rule dependencies.
func TestNewRuleDependencies(parentRuleIDs, childRuleIDs []uuid.UUID) []NewRuleDependency {
	// Create dependencies ensuring no self-references
	var dependencies []NewRuleDependency

	for i, parentID := range parentRuleIDs {
		for j, childID := range childRuleIDs {
			// Skip if parent and child are the same
			if parentID == childID {
				continue
			}
			// Create some selective dependencies (not all combinations)
			if (i+j)%3 == 0 {
				dep := NewRuleDependency{
					ParentRuleID: parentID,
					ChildRuleID:  childID,
				}
				dependencies = append(dependencies, dep)
			}
		}
	}

	return dependencies
}

// TestSeedRuleDependencies is a helper method for testing.
func TestSeedRuleDependencies(ctx context.Context, parentRuleIDs, childRuleIDs []uuid.UUID, api *Business) ([]RuleDependency, error) {
	newDeps := TestNewRuleDependencies(parentRuleIDs, childRuleIDs)

	dependencies := make([]RuleDependency, len(newDeps))
	for i, nrd := range newDeps {
		dep, err := api.CreateDependency(ctx, nrd)
		if err != nil {
			return nil, fmt.Errorf("seeding rule dependency: idx: %d : %w", i, err)
		}

		dependencies[i] = dep
	}

	return dependencies, nil
}

// TestNewAutomationExecutions creates test automation executions.
func TestNewAutomationExecutions(n int, ruleIDs []uuid.UUID) []NewAutomationExecution {
	executions := make([]NewAutomationExecution, n)

	idx := rand.Intn(10000)
	statuses := []ExecutionStatus{StatusCompleted, StatusFailed, StatusPending, StatusRunning}

	for i := 0; i < n; i++ {
		idx++

		// Create sample trigger data
		triggerData := map[string]interface{}{
			"entity_id":      fmt.Sprintf("entity_%d", idx),
			"event_type":     "update",
			"changed_fields": []string{"field1", "field2"},
		}
		triggerDataJSON, _ := json.Marshal(triggerData)

		// Create sample actions executed data
		actionsExecuted := []map[string]interface{}{
			{
				"action_id": fmt.Sprintf("action_%d", idx),
				"status":    "completed",
				"duration":  123,
			},
		}
		actionsExecutedJSON, _ := json.Marshal(actionsExecuted)

		exec := NewAutomationExecution{
			AutomationRuleID: ruleIDs[i%len(ruleIDs)],
			EntityType:       fmt.Sprintf("EntityType%d", idx),
			TriggerData:      triggerDataJSON,
			ActionsExecuted:  actionsExecutedJSON,
			Status:           statuses[i%len(statuses)],
			ErrorMessage:     "",
			ExecutionTimeMs:  100 + rand.Intn(900), // Random time between 100-1000ms
		}

		// Add error message for failed executions
		if exec.Status == StatusFailed {
			exec.ErrorMessage = fmt.Sprintf("Execution failed: error %d", idx)
		}

		executions[i] = exec
	}

	return executions
}

// TestSeedAutomationExecutions is a helper method for testing.
func TestSeedAutomationExecutions(ctx context.Context, n int, ruleIDs []uuid.UUID, api *Business) ([]AutomationExecution, error) {
	newExecutions := TestNewAutomationExecutions(n, ruleIDs)

	executions := make([]AutomationExecution, len(newExecutions))
	for i, nae := range newExecutions {
		exec, err := api.CreateExecution(ctx, nae)
		if err != nil {
			return nil, fmt.Errorf("seeding automation execution: idx: %d : %w", i, err)
		}

		executions[i] = exec
	}

	return executions, nil
}

// TestSeedFullWorkflow seeds a complete workflow setup for testing.
func TestSeedFullWorkflow(ctx context.Context, userID uuid.UUID, api *Business) (*TestWorkflowData, error) {
	data := &TestWorkflowData{}

	// Seed trigger types
	triggerTypes, err := TestSeedTriggerTypes(ctx, 3, api)
	if err != nil {
		return nil, fmt.Errorf("seeding trigger types: %w", err)
	}
	data.TriggerTypes = triggerTypes

	// Seed entity types
	entityTypes, err := TestSeedEntityTypes(ctx, 2, api)
	if err != nil {
		return nil, fmt.Errorf("seeding entity types: %w", err)
	}
	data.EntityTypes = entityTypes

	// Extract entity type IDs
	entityTypeIDs := make([]uuid.UUID, len(entityTypes))
	for i, et := range entityTypes {
		entityTypeIDs[i] = et.ID
	}

	// Seed entities
	entities, err := TestSeedEntities(ctx, 4, entityTypeIDs, api)
	if err != nil {
		return nil, fmt.Errorf("seeding entities: %w", err)
	}
	data.Entities = entities

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
	rules, err := TestSeedAutomationRules(ctx, 5, entityIDs, entityTypeIDs, triggerTypeIDs, userID, api)
	if err != nil {
		return nil, fmt.Errorf("seeding automation rules: %w", err)
	}
	data.AutomationRules = rules

	// Extract rule IDs
	ruleIDs := make([]uuid.UUID, len(rules))
	for i, r := range rules {
		ruleIDs[i] = r.ID
	}

	// Seed action templates
	templates, err := TestSeedActionTemplates(ctx, 3, userID, api)
	if err != nil {
		return nil, fmt.Errorf("seeding action templates: %w", err)
	}
	data.ActionTemplates = templates

	// Extract template IDs
	templateIDs := make([]uuid.UUID, len(templates))
	for i, t := range templates {
		templateIDs[i] = t.ID
	}

	// Seed rule actions
	actions, err := TestSeedRuleActions(ctx, 10, ruleIDs, &templateIDs, api)
	if err != nil {
		return nil, fmt.Errorf("seeding rule actions: %w", err)
	}
	data.RuleActions = actions

	// Seed rule dependencies (using first 3 rules as parents and last 2 as children)
	if len(ruleIDs) >= 3 {
		parentIDs := ruleIDs[:3]
		childIDs := ruleIDs[len(ruleIDs)-2:]
		dependencies, err := TestSeedRuleDependencies(ctx, parentIDs, childIDs, api)
		if err != nil {
			return nil, fmt.Errorf("seeding rule dependencies: %w", err)
		}
		data.RuleDependencies = dependencies
	}

	// Seed automation executions
	executions, err := TestSeedAutomationExecutions(ctx, 15, ruleIDs, api)
	if err != nil {
		return nil, fmt.Errorf("seeding automation executions: %w", err)
	}
	data.AutomationExecutions = executions

	return data, nil
}

// TestWorkflowData holds all seeded workflow data for testing.
type TestWorkflowData struct {
	TriggerTypes         []TriggerType
	EntityTypes          []EntityType
	Entities             []Entity
	AutomationRules      []AutomationRule
	ActionTemplates      []ActionTemplate
	RuleActions          []RuleAction
	RuleDependencies     []RuleDependency
	AutomationExecutions []AutomationExecution
}
