package workflow_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus/stores/inventoryitemdb"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus/stores/inventorylocationdb"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus/stores/inventorytransactiondb"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus/stores/productdb"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

/*
Package workflow_test tests the graph-based action execution in ExecuteRuleActionsGraph.

WHAT THIS TESTS:
- Start edge handling (single/multiple entry points)
- Sequential execution via 'sequence' and 'always' edge types
- Branch execution based on condition results (true_branch/false_branch)
- Complex graph patterns (diamond, parallel branches, nested conditions)
- Cycle prevention (visited nodes not re-executed)
- Edge ordering for deterministic execution
- shouldFollowEdge logic for all edge types

WHAT THIS DOES NOT TEST:
- Full workflow engine integration (covered by E2E tests)
- Real action side effects (emails, alerts, etc.)
*/

// =============================================================================
// Test Setup Helpers
// =============================================================================

// graphTestSetup creates the test infrastructure needed for graph executor tests.
func graphTestSetup(t *testing.T) (*workflow.ActionExecutor, *workflow.Business, context.Context) {
	t.Helper()

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	ndb := dbtest.NewDatabase(t, "Test_GraphExecutor")
	db := ndb.DB
	ctx := context.Background()

	workflowBus := workflow.NewBusiness(log, nil, workflowdb.NewStore(log, db))

	// Seed base workflow data
	_, err := workflow.TestSeedFullWorkflow(ctx, uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"), workflowBus)
	if err != nil {
		t.Fatalf("seeding workflow data: %s", err)
	}

	// Get RabbitMQ container
	container := rabbitmq.GetTestContainer(t)

	// Create RabbitMQ client
	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	t.Cleanup(func() { client.Close() })

	// Create workflow queue for initialization
	queue := rabbitmq.NewTestWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	ae := workflow.NewActionExecutor(log, db, workflowBus)
	workflowactions.RegisterAll(
		ae.GetRegistry(),
		workflowactions.ActionConfig{
			Log:         log,
			DB:          db,
			QueueClient: queue,
			Buses: workflowactions.BusDependencies{
				InventoryItem:        inventoryitembus.NewBusiness(log, delegate.New(log), inventoryitemdb.NewStore(log, db)),
				InventoryLocation:    inventorylocationbus.NewBusiness(log, delegate.New(log), inventorylocationdb.NewStore(log, db)),
				InventoryTransaction: inventorytransactionbus.NewBusiness(log, delegate.New(log), inventorytransactiondb.NewStore(log, db)),
				Product:              productbus.NewBusiness(log, delegate.New(log), productdb.NewStore(log, db)),
			},
		},
	)

	return ae, workflowBus, ctx
}

// createTestRule creates a rule with actions and optionally edges for testing.
func createTestRule(
	t *testing.T,
	ctx context.Context,
	workflowBus *workflow.Business,
	userID uuid.UUID,
	actions []testAction,
	edges []testEdge,
) (uuid.UUID, []uuid.UUID) {
	t.Helper()

	// Get existing entity type, entity, and trigger type
	entityType, err := workflowBus.QueryEntityTypeByName(ctx, "table")
	if err != nil {
		t.Fatalf("Failed to query entity type: %v", err)
	}

	entity, err := workflowBus.QueryEntityByName(ctx, "customers")
	if err != nil {
		t.Fatalf("Failed to query entity: %v", err)
	}

	triggerType, err := workflowBus.QueryTriggerTypeByName(ctx, "on_create")
	if err != nil {
		t.Fatalf("Failed to query trigger type: %v", err)
	}

	// Create automation rule
	rule, err := workflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Test Rule " + uuid.New().String()[:8],
		Description:   "Test rule for graph execution",
		EntityID:      entity.ID,
		EntityTypeID:  entityType.ID,
		TriggerTypeID: triggerType.ID,
		IsActive:      true,
		CreatedBy:     userID,
	})
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	// Create action template for send_email (for regular actions)
	emailTemplate, err := workflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:        "send_email_template_" + uuid.New().String()[:8],
		Description: "Email template",
		ActionType:  "send_email",
		DefaultConfig: json.RawMessage(`{
            "recipients": ["default@example.com"],
            "subject": "Default Subject"
        }`),
		CreatedBy: userID,
	})
	if err != nil {
		t.Fatalf("Failed to create email template: %v", err)
	}

	// Create action template for evaluate_condition
	conditionTemplate, err := workflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:        "condition_template_" + uuid.New().String()[:8],
		Description: "Condition template",
		ActionType:  "evaluate_condition",
		DefaultConfig: json.RawMessage(`{
            "conditions": [],
            "logic_type": "and"
        }`),
		CreatedBy: userID,
	})
	if err != nil {
		t.Fatalf("Failed to create condition template: %v", err)
	}

	// Create actions
	actionIDs := make([]uuid.UUID, len(actions))
	for i, a := range actions {
		templateID := &emailTemplate.ID
		config := json.RawMessage(`{
            "recipients": ["test@example.com"],
            "subject": "Test Email"
        }`)

		if a.isCondition {
			templateID = &conditionTemplate.ID
			config = a.conditionConfig
		}

		action, err := workflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
			AutomationRuleID: rule.ID,
			Name:             a.name,
			ActionConfig:     config,
			IsActive:         a.isActive,
			TemplateID:       templateID,
		})
		if err != nil {
			t.Fatalf("Failed to create action %s: %v", a.name, err)
		}
		actionIDs[i] = action.ID
	}

	// Create edges
	for _, e := range edges {
		var sourceID *uuid.UUID
		if e.sourceIdx >= 0 {
			sourceID = &actionIDs[e.sourceIdx]
		}

		_, err := workflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
			RuleID:         rule.ID,
			SourceActionID: sourceID,
			TargetActionID: actionIDs[e.targetIdx],
			EdgeType:       e.edgeType,
			EdgeOrder:      e.edgeOrder,
		})
		if err != nil {
			t.Fatalf("Failed to create edge: %v", err)
		}
	}

	return rule.ID, actionIDs
}

type testAction struct {
	name            string
	isCondition     bool
	isActive        bool
	conditionConfig json.RawMessage
}

type testEdge struct {
	sourceIdx int // -1 for start edge
	targetIdx int
	edgeType  string
	edgeOrder int
}

// =============================================================================
// Start Edge Tests
// =============================================================================

func TestGraphExec_SingleStartEdge(t *testing.T) {
	ae, workflowBus, ctx := graphTestSetup(t)
	userID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	actions := []testAction{
		{name: "Start Action", isActive: true},
		{name: "Follow Up", isActive: true},
	}

	edges := []testEdge{
		{sourceIdx: -1, targetIdx: 0, edgeType: workflow.EdgeTypeStart, edgeOrder: 1},   // Start -> Action 0
		{sourceIdx: 0, targetIdx: 1, edgeType: workflow.EdgeTypeSequence, edgeOrder: 1}, // Action 0 -> Action 1
	}

	ruleID, _ := createTestRule(t, ctx, workflowBus, userID, actions, edges)

	entity, _ := workflowBus.QueryEntityByName(ctx, "customers")

	execContext := workflow.ActionExecutionContext{
		EntityID:      entity.ID,
		EntityName:    "customers",
		EventType:     "on_create",
		UserID:        userID,
		RuleID:        &ruleID,
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now(),
		TriggerSource: workflow.TriggerSourceAutomation,
	}

	result, err := ae.ExecuteRuleActionsGraph(ctx, ruleID, execContext)
	if err != nil {
		t.Fatalf("ExecuteRuleActionsGraph failed: %v", err)
	}

	if result.TotalActions != 2 {
		t.Errorf("TotalActions = %d, want 2", result.TotalActions)
	}

	// Both actions should execute
	if result.SuccessfulActions != 2 {
		t.Errorf("SuccessfulActions = %d, want 2", result.SuccessfulActions)
	}
}

func TestGraphExec_MultipleStartEdges(t *testing.T) {
	ae, workflowBus, ctx := graphTestSetup(t)
	userID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	actions := []testAction{
		{name: "Entry Point A", isActive: true},
		{name: "Entry Point B", isActive: true},
	}

	edges := []testEdge{
		{sourceIdx: -1, targetIdx: 0, edgeType: workflow.EdgeTypeStart, edgeOrder: 1}, // Start -> A
		{sourceIdx: -1, targetIdx: 1, edgeType: workflow.EdgeTypeStart, edgeOrder: 2}, // Start -> B
	}

	ruleID, _ := createTestRule(t, ctx, workflowBus, userID, actions, edges)

	entity, _ := workflowBus.QueryEntityByName(ctx, "customers")

	execContext := workflow.ActionExecutionContext{
		EntityID:      entity.ID,
		EntityName:    "customers",
		EventType:     "on_create",
		UserID:        userID,
		RuleID:        &ruleID,
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now(),
		TriggerSource: workflow.TriggerSourceAutomation,
	}

	result, err := ae.ExecuteRuleActionsGraph(ctx, ruleID, execContext)
	if err != nil {
		t.Fatalf("ExecuteRuleActionsGraph failed: %v", err)
	}

	// Both entry points should execute
	if result.TotalActions != 2 {
		t.Errorf("TotalActions = %d, want 2", result.TotalActions)
	}
	if result.SuccessfulActions != 2 {
		t.Errorf("SuccessfulActions = %d, want 2", result.SuccessfulActions)
	}
}

// =============================================================================
// Sequential Execution Tests
// =============================================================================

func TestGraphExec_LinearChain(t *testing.T) {
	ae, workflowBus, ctx := graphTestSetup(t)
	userID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	actions := []testAction{
		{name: "Step 1", isActive: true},
		{name: "Step 2", isActive: true},
		{name: "Step 3", isActive: true},
	}

	edges := []testEdge{
		{sourceIdx: -1, targetIdx: 0, edgeType: workflow.EdgeTypeStart, edgeOrder: 1},
		{sourceIdx: 0, targetIdx: 1, edgeType: workflow.EdgeTypeSequence, edgeOrder: 1},
		{sourceIdx: 1, targetIdx: 2, edgeType: workflow.EdgeTypeSequence, edgeOrder: 1},
	}

	ruleID, _ := createTestRule(t, ctx, workflowBus, userID, actions, edges)

	entity, _ := workflowBus.QueryEntityByName(ctx, "customers")

	execContext := workflow.ActionExecutionContext{
		EntityID:      entity.ID,
		EntityName:    "customers",
		EventType:     "on_create",
		UserID:        userID,
		RuleID:        &ruleID,
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now(),
		TriggerSource: workflow.TriggerSourceAutomation,
	}

	result, err := ae.ExecuteRuleActionsGraph(ctx, ruleID, execContext)
	if err != nil {
		t.Fatalf("ExecuteRuleActionsGraph failed: %v", err)
	}

	if result.TotalActions != 3 {
		t.Errorf("TotalActions = %d, want 3", result.TotalActions)
	}

	// Verify order
	expectedOrder := []string{"Step 1", "Step 2", "Step 3"}
	for i, ar := range result.ActionResults {
		if ar.ActionName != expectedOrder[i] {
			t.Errorf("ActionResults[%d].ActionName = %s, want %s", i, ar.ActionName, expectedOrder[i])
		}
	}
}

func TestGraphExec_AlwaysEdgeType(t *testing.T) {
	ae, workflowBus, ctx := graphTestSetup(t)
	userID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	actions := []testAction{
		{name: "First", isActive: true},
		{name: "Always Next", isActive: true},
	}

	edges := []testEdge{
		{sourceIdx: -1, targetIdx: 0, edgeType: workflow.EdgeTypeStart, edgeOrder: 1},
		{sourceIdx: 0, targetIdx: 1, edgeType: workflow.EdgeTypeAlways, edgeOrder: 1}, // "always" should always follow
	}

	ruleID, _ := createTestRule(t, ctx, workflowBus, userID, actions, edges)

	entity, _ := workflowBus.QueryEntityByName(ctx, "customers")

	execContext := workflow.ActionExecutionContext{
		EntityID:      entity.ID,
		EntityName:    "customers",
		EventType:     "on_create",
		UserID:        userID,
		RuleID:        &ruleID,
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now(),
		TriggerSource: workflow.TriggerSourceAutomation,
	}

	result, err := ae.ExecuteRuleActionsGraph(ctx, ruleID, execContext)
	if err != nil {
		t.Fatalf("ExecuteRuleActionsGraph failed: %v", err)
	}

	// Both actions should execute via "always" edge
	if result.SuccessfulActions != 2 {
		t.Errorf("SuccessfulActions = %d, want 2", result.SuccessfulActions)
	}
}

// =============================================================================
// Branch Execution Tests
// =============================================================================

func TestGraphExec_TrueBranch_WhenConditionTrue(t *testing.T) {
	ae, workflowBus, ctx := graphTestSetup(t)
	userID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	// Condition config that evaluates to TRUE when amount > 1000
	conditionConfig := json.RawMessage(`{
		"conditions": [{"field_name": "amount", "operator": "greater_than", "value": 1000}],
		"logic_type": "and"
	}`)

	actions := []testAction{
		{name: "Condition", isCondition: true, isActive: true, conditionConfig: conditionConfig},
		{name: "True Branch Action", isActive: true},
		{name: "False Branch Action", isActive: true},
	}

	edges := []testEdge{
		{sourceIdx: -1, targetIdx: 0, edgeType: workflow.EdgeTypeStart, edgeOrder: 1},
		{sourceIdx: 0, targetIdx: 1, edgeType: workflow.EdgeTypeTrueBranch, edgeOrder: 1},
		{sourceIdx: 0, targetIdx: 2, edgeType: workflow.EdgeTypeFalseBranch, edgeOrder: 2},
	}

	ruleID, _ := createTestRule(t, ctx, workflowBus, userID, actions, edges)

	entity, _ := workflowBus.QueryEntityByName(ctx, "customers")

	// Execute with amount = 1500 (condition is TRUE)
	execContext := workflow.ActionExecutionContext{
		EntityID:      entity.ID,
		EntityName:    "customers",
		EventType:     "on_create",
		UserID:        userID,
		RuleID:        &ruleID,
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now(),
		TriggerSource: workflow.TriggerSourceAutomation,
		RawData: map[string]interface{}{
			"amount": 1500, // > 1000, so condition is TRUE
		},
	}

	result, err := ae.ExecuteRuleActionsGraph(ctx, ruleID, execContext)
	if err != nil {
		t.Fatalf("ExecuteRuleActionsGraph failed: %v", err)
	}

	// Should execute: Condition + True Branch Action = 2 actions
	if result.TotalActions != 2 {
		t.Errorf("TotalActions = %d, want 2", result.TotalActions)
	}

	// Verify the condition took true_branch
	if result.ActionResults[0].BranchTaken != workflow.EdgeTypeTrueBranch {
		t.Errorf("Condition BranchTaken = %s, want %s", result.ActionResults[0].BranchTaken, workflow.EdgeTypeTrueBranch)
	}

	// Verify True Branch Action was executed
	if result.ActionResults[1].ActionName != "True Branch Action" {
		t.Errorf("Second action = %s, want True Branch Action", result.ActionResults[1].ActionName)
	}
}

func TestGraphExec_FalseBranch_WhenConditionFalse(t *testing.T) {
	ae, workflowBus, ctx := graphTestSetup(t)
	userID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	// Condition config that evaluates to FALSE when amount <= 1000
	conditionConfig := json.RawMessage(`{
		"conditions": [{"field_name": "amount", "operator": "greater_than", "value": 1000}],
		"logic_type": "and"
	}`)

	actions := []testAction{
		{name: "Condition", isCondition: true, isActive: true, conditionConfig: conditionConfig},
		{name: "True Branch Action", isActive: true},
		{name: "False Branch Action", isActive: true},
	}

	edges := []testEdge{
		{sourceIdx: -1, targetIdx: 0, edgeType: workflow.EdgeTypeStart, edgeOrder: 1},
		{sourceIdx: 0, targetIdx: 1, edgeType: workflow.EdgeTypeTrueBranch, edgeOrder: 1},
		{sourceIdx: 0, targetIdx: 2, edgeType: workflow.EdgeTypeFalseBranch, edgeOrder: 2},
	}

	ruleID, _ := createTestRule(t, ctx, workflowBus, userID, actions, edges)

	entity, _ := workflowBus.QueryEntityByName(ctx, "customers")

	// Execute with amount = 500 (condition is FALSE)
	execContext := workflow.ActionExecutionContext{
		EntityID:      entity.ID,
		EntityName:    "customers",
		EventType:     "on_create",
		UserID:        userID,
		RuleID:        &ruleID,
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now(),
		TriggerSource: workflow.TriggerSourceAutomation,
		RawData: map[string]interface{}{
			"amount": 500, // <= 1000, so condition is FALSE
		},
	}

	result, err := ae.ExecuteRuleActionsGraph(ctx, ruleID, execContext)
	if err != nil {
		t.Fatalf("ExecuteRuleActionsGraph failed: %v", err)
	}

	// Should execute: Condition + False Branch Action = 2 actions
	if result.TotalActions != 2 {
		t.Errorf("TotalActions = %d, want 2", result.TotalActions)
	}

	// Verify the condition took false_branch
	if result.ActionResults[0].BranchTaken != workflow.EdgeTypeFalseBranch {
		t.Errorf("Condition BranchTaken = %s, want %s", result.ActionResults[0].BranchTaken, workflow.EdgeTypeFalseBranch)
	}

	// Verify False Branch Action was executed
	if result.ActionResults[1].ActionName != "False Branch Action" {
		t.Errorf("Second action = %s, want False Branch Action", result.ActionResults[1].ActionName)
	}
}

func TestGraphExec_SkipsTrueBranch_WhenFalse(t *testing.T) {
	ae, workflowBus, ctx := graphTestSetup(t)
	userID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	conditionConfig := json.RawMessage(`{
		"conditions": [{"field_name": "status", "operator": "equals", "value": "premium"}],
		"logic_type": "and"
	}`)

	actions := []testAction{
		{name: "Condition", isCondition: true, isActive: true, conditionConfig: conditionConfig},
		{name: "Premium Only Action", isActive: true},
	}

	edges := []testEdge{
		{sourceIdx: -1, targetIdx: 0, edgeType: workflow.EdgeTypeStart, edgeOrder: 1},
		{sourceIdx: 0, targetIdx: 1, edgeType: workflow.EdgeTypeTrueBranch, edgeOrder: 1}, // Only follow if true
	}

	ruleID, _ := createTestRule(t, ctx, workflowBus, userID, actions, edges)

	entity, _ := workflowBus.QueryEntityByName(ctx, "customers")

	// Execute with status = "regular" (condition is FALSE)
	execContext := workflow.ActionExecutionContext{
		EntityID:      entity.ID,
		EntityName:    "customers",
		EventType:     "on_create",
		UserID:        userID,
		RuleID:        &ruleID,
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now(),
		TriggerSource: workflow.TriggerSourceAutomation,
		RawData: map[string]interface{}{
			"status": "regular", // != "premium", so condition is FALSE
		},
	}

	result, err := ae.ExecuteRuleActionsGraph(ctx, ruleID, execContext)
	if err != nil {
		t.Fatalf("ExecuteRuleActionsGraph failed: %v", err)
	}

	// Should only execute the condition (Premium Only Action skipped)
	if result.TotalActions != 1 {
		t.Errorf("TotalActions = %d, want 1 (only condition)", result.TotalActions)
	}

	if result.ActionResults[0].ActionName != "Condition" {
		t.Errorf("Only executed action should be Condition, got %s", result.ActionResults[0].ActionName)
	}
}

// =============================================================================
// Complex Graph Tests
// =============================================================================

func TestGraphExec_DiamondPattern(t *testing.T) {
	ae, workflowBus, ctx := graphTestSetup(t)
	userID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	// Diamond pattern: Condition -> Branch A/B -> Converge
	conditionConfig := json.RawMessage(`{
		"conditions": [{"field_name": "priority", "operator": "equals", "value": "high"}],
		"logic_type": "and"
	}`)

	actions := []testAction{
		{name: "Condition", isCondition: true, isActive: true, conditionConfig: conditionConfig}, // 0
		{name: "High Priority Path", isActive: true},                                             // 1
		{name: "Normal Priority Path", isActive: true},                                           // 2
		{name: "Convergence Point", isActive: true},                                              // 3
	}

	edges := []testEdge{
		{sourceIdx: -1, targetIdx: 0, edgeType: workflow.EdgeTypeStart, edgeOrder: 1},
		{sourceIdx: 0, targetIdx: 1, edgeType: workflow.EdgeTypeTrueBranch, edgeOrder: 1},
		{sourceIdx: 0, targetIdx: 2, edgeType: workflow.EdgeTypeFalseBranch, edgeOrder: 2},
		{sourceIdx: 1, targetIdx: 3, edgeType: workflow.EdgeTypeSequence, edgeOrder: 1},
		{sourceIdx: 2, targetIdx: 3, edgeType: workflow.EdgeTypeSequence, edgeOrder: 1},
	}

	ruleID, _ := createTestRule(t, ctx, workflowBus, userID, actions, edges)

	entity, _ := workflowBus.QueryEntityByName(ctx, "customers")

	// Execute with high priority (true branch)
	execContext := workflow.ActionExecutionContext{
		EntityID:      entity.ID,
		EntityName:    "customers",
		EventType:     "on_create",
		UserID:        userID,
		RuleID:        &ruleID,
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now(),
		TriggerSource: workflow.TriggerSourceAutomation,
		RawData: map[string]interface{}{
			"priority": "high",
		},
	}

	result, err := ae.ExecuteRuleActionsGraph(ctx, ruleID, execContext)
	if err != nil {
		t.Fatalf("ExecuteRuleActionsGraph failed: %v", err)
	}

	// Should execute: Condition -> High Priority Path -> Convergence Point = 3 actions
	if result.TotalActions != 3 {
		t.Errorf("TotalActions = %d, want 3", result.TotalActions)
	}

	// Verify "Normal Priority Path" was NOT executed
	for _, ar := range result.ActionResults {
		if ar.ActionName == "Normal Priority Path" {
			t.Error("Normal Priority Path should not have executed on high priority")
		}
	}

	// Verify convergence point executed exactly once (not duplicated from both branches)
	convergenceCount := 0
	for _, ar := range result.ActionResults {
		if ar.ActionName == "Convergence Point" {
			convergenceCount++
		}
	}
	if convergenceCount != 1 {
		t.Errorf("Convergence Point executed %d times, want 1", convergenceCount)
	}
}

func TestGraphExec_NestedConditions(t *testing.T) {
	ae, workflowBus, ctx := graphTestSetup(t)
	userID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	// Nested: Cond1 (urgent?) -> TRUE -> Cond2 (priority > 5?) -> TRUE -> Escalate
	cond1Config := json.RawMessage(`{
		"conditions": [{"field_name": "type", "operator": "equals", "value": "urgent"}],
		"logic_type": "and"
	}`)
	cond2Config := json.RawMessage(`{
		"conditions": [{"field_name": "priority", "operator": "greater_than", "value": 5}],
		"logic_type": "and"
	}`)

	actions := []testAction{
		{name: "Is Urgent?", isCondition: true, isActive: true, conditionConfig: cond1Config}, // 0
		{name: "Priority Check", isCondition: true, isActive: true, conditionConfig: cond2Config}, // 1
		{name: "Escalate", isActive: true},                   // 2
		{name: "Standard Process", isActive: true},           // 3
		{name: "Queue for Later", isActive: true},            // 4
	}

	edges := []testEdge{
		{sourceIdx: -1, targetIdx: 0, edgeType: workflow.EdgeTypeStart, edgeOrder: 1},
		{sourceIdx: 0, targetIdx: 1, edgeType: workflow.EdgeTypeTrueBranch, edgeOrder: 1},  // Urgent -> Priority Check
		{sourceIdx: 0, targetIdx: 4, edgeType: workflow.EdgeTypeFalseBranch, edgeOrder: 2}, // Not urgent -> Queue
		{sourceIdx: 1, targetIdx: 2, edgeType: workflow.EdgeTypeTrueBranch, edgeOrder: 1},  // High priority -> Escalate
		{sourceIdx: 1, targetIdx: 3, edgeType: workflow.EdgeTypeFalseBranch, edgeOrder: 2}, // Low priority -> Standard
	}

	ruleID, _ := createTestRule(t, ctx, workflowBus, userID, actions, edges)

	entity, _ := workflowBus.QueryEntityByName(ctx, "customers")

	// Test case: urgent + high priority -> should escalate
	execContext := workflow.ActionExecutionContext{
		EntityID:      entity.ID,
		EntityName:    "customers",
		EventType:     "on_create",
		UserID:        userID,
		RuleID:        &ruleID,
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now(),
		TriggerSource: workflow.TriggerSourceAutomation,
		RawData: map[string]interface{}{
			"type":     "urgent",
			"priority": 8,
		},
	}

	result, err := ae.ExecuteRuleActionsGraph(ctx, ruleID, execContext)
	if err != nil {
		t.Fatalf("ExecuteRuleActionsGraph failed: %v", err)
	}

	// Should execute: Is Urgent? -> Priority Check -> Escalate = 3 actions
	if result.TotalActions != 3 {
		t.Errorf("TotalActions = %d, want 3", result.TotalActions)
	}

	// Verify "Escalate" was executed
	found := false
	for _, ar := range result.ActionResults {
		if ar.ActionName == "Escalate" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Escalate action should have been executed")
	}
}

// =============================================================================
// Cycle Prevention Tests
// =============================================================================

func TestGraphExec_NoCycleInfiniteLoop(t *testing.T) {
	ae, workflowBus, ctx := graphTestSetup(t)
	userID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	// Create a graph that could cycle: A -> B -> A
	// The executor should prevent infinite loops
	actions := []testAction{
		{name: "Action A", isActive: true},
		{name: "Action B", isActive: true},
	}

	edges := []testEdge{
		{sourceIdx: -1, targetIdx: 0, edgeType: workflow.EdgeTypeStart, edgeOrder: 1},
		{sourceIdx: 0, targetIdx: 1, edgeType: workflow.EdgeTypeSequence, edgeOrder: 1},
		{sourceIdx: 1, targetIdx: 0, edgeType: workflow.EdgeTypeSequence, edgeOrder: 1}, // Back to A (cycle)
	}

	ruleID, _ := createTestRule(t, ctx, workflowBus, userID, actions, edges)

	entity, _ := workflowBus.QueryEntityByName(ctx, "customers")

	execContext := workflow.ActionExecutionContext{
		EntityID:      entity.ID,
		EntityName:    "customers",
		EventType:     "on_create",
		UserID:        userID,
		RuleID:        &ruleID,
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now(),
		TriggerSource: workflow.TriggerSourceAutomation,
	}

	result, err := ae.ExecuteRuleActionsGraph(ctx, ruleID, execContext)
	if err != nil {
		t.Fatalf("ExecuteRuleActionsGraph failed: %v", err)
	}

	// Should execute each action exactly once (no infinite loop)
	if result.TotalActions != 2 {
		t.Errorf("TotalActions = %d, want 2 (visited nodes not re-executed)", result.TotalActions)
	}
}

func TestGraphExec_SelfLoop_Ignored(t *testing.T) {
	ae, workflowBus, ctx := graphTestSetup(t)
	userID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	// Self-loop: A -> A
	actions := []testAction{
		{name: "Self Looper", isActive: true},
	}

	edges := []testEdge{
		{sourceIdx: -1, targetIdx: 0, edgeType: workflow.EdgeTypeStart, edgeOrder: 1},
		{sourceIdx: 0, targetIdx: 0, edgeType: workflow.EdgeTypeSequence, edgeOrder: 1}, // Self-loop
	}

	ruleID, _ := createTestRule(t, ctx, workflowBus, userID, actions, edges)

	entity, _ := workflowBus.QueryEntityByName(ctx, "customers")

	execContext := workflow.ActionExecutionContext{
		EntityID:      entity.ID,
		EntityName:    "customers",
		EventType:     "on_create",
		UserID:        userID,
		RuleID:        &ruleID,
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now(),
		TriggerSource: workflow.TriggerSourceAutomation,
	}

	result, err := ae.ExecuteRuleActionsGraph(ctx, ruleID, execContext)
	if err != nil {
		t.Fatalf("ExecuteRuleActionsGraph failed: %v", err)
	}

	// Should execute exactly once (self-loop ignored)
	if result.TotalActions != 1 {
		t.Errorf("TotalActions = %d, want 1 (self-loop ignored)", result.TotalActions)
	}
}

// =============================================================================
// Edge Order Tests
// =============================================================================

func TestGraphExec_EdgeOrderRespected(t *testing.T) {
	ae, workflowBus, ctx := graphTestSetup(t)
	userID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	// Multiple start edges with different orders
	actions := []testAction{
		{name: "First (order 1)", isActive: true},
		{name: "Second (order 2)", isActive: true},
		{name: "Third (order 3)", isActive: true},
	}

	// Start edges in reverse order to verify sorting
	edges := []testEdge{
		{sourceIdx: -1, targetIdx: 2, edgeType: workflow.EdgeTypeStart, edgeOrder: 3},
		{sourceIdx: -1, targetIdx: 0, edgeType: workflow.EdgeTypeStart, edgeOrder: 1},
		{sourceIdx: -1, targetIdx: 1, edgeType: workflow.EdgeTypeStart, edgeOrder: 2},
	}

	ruleID, _ := createTestRule(t, ctx, workflowBus, userID, actions, edges)

	entity, _ := workflowBus.QueryEntityByName(ctx, "customers")

	execContext := workflow.ActionExecutionContext{
		EntityID:      entity.ID,
		EntityName:    "customers",
		EventType:     "on_create",
		UserID:        userID,
		RuleID:        &ruleID,
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now(),
		TriggerSource: workflow.TriggerSourceAutomation,
	}

	result, err := ae.ExecuteRuleActionsGraph(ctx, ruleID, execContext)
	if err != nil {
		t.Fatalf("ExecuteRuleActionsGraph failed: %v", err)
	}

	// All 3 should execute
	if result.TotalActions != 3 {
		t.Errorf("TotalActions = %d, want 3", result.TotalActions)
	}

	// Verify execution order matches edge_order
	expectedOrder := []string{"First (order 1)", "Second (order 2)", "Third (order 3)"}
	for i, ar := range result.ActionResults {
		if ar.ActionName != expectedOrder[i] {
			t.Errorf("ActionResults[%d].ActionName = %s, want %s", i, ar.ActionName, expectedOrder[i])
		}
	}
}

// =============================================================================
// shouldFollowEdge Unit Tests (Pure unit tests - no database required)
// =============================================================================

func TestShouldFollowEdge_Always(t *testing.T) {
	t.Parallel()

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})
	ae := workflow.NewActionExecutor(log, nil, nil)

	edge := workflow.ActionEdge{EdgeType: workflow.EdgeTypeAlways}
	result := workflow.ActionResult{Status: "success"}

	if !ae.ShouldFollowEdge(edge, result) {
		t.Error("EdgeTypeAlways should always return true")
	}

	// Even with branch taken set
	result.BranchTaken = workflow.EdgeTypeFalseBranch
	if !ae.ShouldFollowEdge(edge, result) {
		t.Error("EdgeTypeAlways should return true regardless of BranchTaken")
	}
}

func TestShouldFollowEdge_Sequence(t *testing.T) {
	t.Parallel()

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})
	ae := workflow.NewActionExecutor(log, nil, nil)

	edge := workflow.ActionEdge{EdgeType: workflow.EdgeTypeSequence}
	result := workflow.ActionResult{Status: "success"}

	if !ae.ShouldFollowEdge(edge, result) {
		t.Error("EdgeTypeSequence should always return true")
	}
}

func TestShouldFollowEdge_Start(t *testing.T) {
	t.Parallel()

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})
	ae := workflow.NewActionExecutor(log, nil, nil)

	edge := workflow.ActionEdge{EdgeType: workflow.EdgeTypeStart}
	result := workflow.ActionResult{Status: "success"}

	if ae.ShouldFollowEdge(edge, result) {
		t.Error("EdgeTypeStart should return false (handled separately)")
	}
}

func TestShouldFollowEdge_TrueBranch_Match(t *testing.T) {
	t.Parallel()

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})
	ae := workflow.NewActionExecutor(log, nil, nil)

	edge := workflow.ActionEdge{EdgeType: workflow.EdgeTypeTrueBranch}
	result := workflow.ActionResult{
		Status:      "success",
		BranchTaken: workflow.EdgeTypeTrueBranch,
	}

	if !ae.ShouldFollowEdge(edge, result) {
		t.Error("EdgeTypeTrueBranch should return true when BranchTaken matches")
	}
}

func TestShouldFollowEdge_TrueBranch_NoMatch(t *testing.T) {
	t.Parallel()

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})
	ae := workflow.NewActionExecutor(log, nil, nil)

	edge := workflow.ActionEdge{EdgeType: workflow.EdgeTypeTrueBranch}
	result := workflow.ActionResult{
		Status:      "success",
		BranchTaken: workflow.EdgeTypeFalseBranch,
	}

	if ae.ShouldFollowEdge(edge, result) {
		t.Error("EdgeTypeTrueBranch should return false when BranchTaken doesn't match")
	}
}

func TestShouldFollowEdge_FalseBranch_Match(t *testing.T) {
	t.Parallel()

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})
	ae := workflow.NewActionExecutor(log, nil, nil)

	edge := workflow.ActionEdge{EdgeType: workflow.EdgeTypeFalseBranch}
	result := workflow.ActionResult{
		Status:      "success",
		BranchTaken: workflow.EdgeTypeFalseBranch,
	}

	if !ae.ShouldFollowEdge(edge, result) {
		t.Error("EdgeTypeFalseBranch should return true when BranchTaken matches")
	}
}

func TestShouldFollowEdge_FalseBranch_NoMatch(t *testing.T) {
	t.Parallel()

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})
	ae := workflow.NewActionExecutor(log, nil, nil)

	edge := workflow.ActionEdge{EdgeType: workflow.EdgeTypeFalseBranch}
	result := workflow.ActionResult{
		Status:      "success",
		BranchTaken: workflow.EdgeTypeTrueBranch,
	}

	if ae.ShouldFollowEdge(edge, result) {
		t.Error("EdgeTypeFalseBranch should return false when BranchTaken doesn't match")
	}
}

func TestShouldFollowEdge_UnknownType(t *testing.T) {
	t.Parallel()

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})
	ae := workflow.NewActionExecutor(log, nil, nil)

	edge := workflow.ActionEdge{EdgeType: "unknown_type"}
	result := workflow.ActionResult{Status: "success"}

	if ae.ShouldFollowEdge(edge, result) {
		t.Error("Unknown edge type should return false")
	}
}

// =============================================================================
// Skipped/Inactive Action Tests
// =============================================================================

func TestGraphExec_SkipsInactiveActions(t *testing.T) {
	ae, workflowBus, ctx := graphTestSetup(t)
	userID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	actions := []testAction{
		{name: "Active Start", isActive: true},
		{name: "Inactive Middle", isActive: false}, // Will be skipped
		{name: "Active End", isActive: true},
	}

	edges := []testEdge{
		{sourceIdx: -1, targetIdx: 0, edgeType: workflow.EdgeTypeStart, edgeOrder: 1},
		{sourceIdx: 0, targetIdx: 1, edgeType: workflow.EdgeTypeSequence, edgeOrder: 1},
		{sourceIdx: 1, targetIdx: 2, edgeType: workflow.EdgeTypeSequence, edgeOrder: 1},
	}

	ruleID, _ := createTestRule(t, ctx, workflowBus, userID, actions, edges)

	entity, _ := workflowBus.QueryEntityByName(ctx, "customers")

	execContext := workflow.ActionExecutionContext{
		EntityID:      entity.ID,
		EntityName:    "customers",
		EventType:     "on_create",
		UserID:        userID,
		RuleID:        &ruleID,
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now(),
		TriggerSource: workflow.TriggerSourceAutomation,
	}

	result, err := ae.ExecuteRuleActionsGraph(ctx, ruleID, execContext)
	if err != nil {
		t.Fatalf("ExecuteRuleActionsGraph failed: %v", err)
	}

	// All 3 actions should be processed (even if one is skipped)
	if result.TotalActions != 3 {
		t.Errorf("TotalActions = %d, want 3", result.TotalActions)
	}

	// 2 successful, 1 skipped
	if result.SuccessfulActions != 2 {
		t.Errorf("SuccessfulActions = %d, want 2", result.SuccessfulActions)
	}
	if result.SkippedActions != 1 {
		t.Errorf("SkippedActions = %d, want 1", result.SkippedActions)
	}

	// Verify the middle action was marked as skipped
	if result.ActionResults[1].Status != "skipped" {
		t.Errorf("Middle action status = %s, want skipped", result.ActionResults[1].Status)
	}
}
