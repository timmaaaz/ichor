package workflowsaveapi_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// =============================================================================
// Phase 8: End-to-End Trigger Integration Tests
// =============================================================================

// TriggerTestData extends ExecutionTestData with real domain business layers
// that fire events through the delegate pattern.
type TriggerTestData struct {
	ExecutionTestData

	// Domain business layers (connected to delegate)
	CustomersBus *customersbus.Business

	// Event bridge components
	EventPublisher  *workflow.EventPublisher
	DelegateHandler *workflow.DelegateHandler

	// Test entities (FK dependencies for customers)
	StreetID      uuid.UUID
	ContactInfoID uuid.UUID
}

// insertTriggerSeedData initializes trigger test infrastructure.
// It sets up the delegate handler to bridge domain events to workflow events.
func insertTriggerSeedData(t *testing.T, test *apitest.Test, esd ExecutionTestData) TriggerTestData {
	t.Helper()
	ctx := context.Background()

	db := test.DB
	busDomain := db.BusDomain

	// -------------------------------------------------------------------------
	// Create EventPublisher and wire up DelegateHandler
	// -------------------------------------------------------------------------

	eventPublisher := workflow.NewEventPublisher(db.Log, esd.WF.QueueManager)
	delegateHandler := workflow.NewDelegateHandler(db.Log, eventPublisher)

	// Register customers domain for event bridging
	delegateHandler.RegisterDomain(busDomain.Delegate, customersbus.DomainName, customersbus.EntityName)

	// Use the existing customersbus from BusDomain
	customersBus := busDomain.Customers

	// -------------------------------------------------------------------------
	// Seed FK dependencies for customers
	// -------------------------------------------------------------------------

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		t.Fatalf("querying regions: %s", err)
	}
	regionIDs := make([]uuid.UUID, len(regions))
	for i, r := range regions {
		regionIDs[i] = r.ID
	}

	cities, err := citybus.TestSeedCities(ctx, 1, regionIDs, busDomain.City)
	if err != nil {
		t.Fatalf("seeding cities: %s", err)
	}

	streets, err := streetbus.TestSeedStreets(ctx, 1, []uuid.UUID{cities[0].ID}, busDomain.Street)
	if err != nil {
		t.Fatalf("seeding streets: %s", err)
	}

	tzs, err := busDomain.Timezone.Query(ctx, timezonebus.QueryFilter{}, timezonebus.DefaultOrderBy, page.MustParse("1", "1"))
	if err != nil {
		t.Fatalf("querying timezones: %s", err)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 1, []uuid.UUID{streets[0].ID}, []uuid.UUID{tzs[0].ID}, busDomain.ContactInfos)
	if err != nil {
		t.Fatalf("seeding contact infos: %s", err)
	}

	return TriggerTestData{
		ExecutionTestData: esd,
		CustomersBus:      customersBus,
		EventPublisher:    eventPublisher,
		DelegateHandler:   delegateHandler,
		StreetID:          streets[0].ID,
		ContactInfoID:     contactInfos[0].ID,
	}
}

// runTriggerTests runs all end-to-end trigger tests as subtests.
func runTriggerTests(t *testing.T, tsd TriggerTestData) {
	// Entity Create Triggers (8a)
	t.Run("trigger-customer-create", func(t *testing.T) {
		testCustomerCreateTriggersWorkflow(t, tsd)
	})

	// Entity Update Triggers (8b)
	t.Run("trigger-customer-update", func(t *testing.T) {
		testCustomerUpdateTriggersWorkflow(t, tsd)
	})

	// Inactive Rule Tests (8e)
	t.Run("trigger-inactive-rule-no-trigger", func(t *testing.T) {
		testInactiveRuleNoTrigger(t, tsd)
	})
}

// testCustomerCreateTriggersWorkflow tests that creating a customer via customersbus
// fires an on_create event that triggers a matching workflow.
func testCustomerCreateTriggersWorkflow(t *testing.T, tsd TriggerTestData) {
	ctx := context.Background()

	// -------------------------------------------------------------------------
	// 1. Create a workflow rule that triggers on customer creation
	// -------------------------------------------------------------------------

	// Query for customers entity
	customerEntity, err := tsd.WF.WorkflowBus.QueryEntityByName(ctx, "customers")
	if err != nil {
		t.Fatalf("querying customers entity: %s", err)
	}

	entityType, err := tsd.WF.WorkflowBus.QueryEntityTypeByName(ctx, "table")
	if err != nil {
		t.Fatalf("querying entity type: %s", err)
	}

	triggerTypeCreate, err := tsd.WF.WorkflowBus.QueryTriggerTypeByName(ctx, "on_create")
	if err != nil {
		t.Fatalf("querying on_create trigger type: %s", err)
	}

	// Create automation rule
	rule, err := tsd.WF.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Customer Created - Trigger Test " + uuid.New().String()[:8],
		Description:   "Tests delegate pattern fires on_create events",
		EntityID:      customerEntity.ID,
		EntityTypeID:  entityType.ID,
		TriggerTypeID: triggerTypeCreate.ID,
		IsActive:      true,
		CreatedBy:     tsd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating on_create rule: %s", err)
	}

	// Create action
	action, err := tsd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Create Customer Alert",
		ActionConfig: json.RawMessage(`{
			"alert_type": "customer_created",
			"severity": "medium",
			"title": "New Customer Created",
			"message": "Customer was created via trigger test",
			"recipients": {"users": ["` + tsd.Users[0].ID.String() + `"], "roles": []}
		}`),
		IsActive:       true,
		TemplateID:     &tsd.CreateAlertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating rule action: %s", err)
	}

	// Create start edge (required for workflow execution)
	_, err = tsd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: action.ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating start edge: %s", err)
	}

	// Rule cache is now automatically invalidated via delegate pattern
	// No need to manually reset the engine

	// -------------------------------------------------------------------------
	// 2. Create a customer (this should trigger the workflow via delegate)
	// -------------------------------------------------------------------------

	initialMetrics := tsd.WF.QueueManager.GetMetrics()

	newCustomer := customersbus.NewCustomers{
		Name:              "Trigger Test Customer " + uuid.New().String()[:8],
		DeliveryAddressID: tsd.StreetID,
		ContactID:         tsd.ContactInfoID,
		CreatedBy:         tsd.Users[0].ID,
	}

	customer, err := tsd.CustomersBus.Create(ctx, newCustomer)
	if err != nil {
		t.Fatalf("creating customer: %s", err)
	}

	t.Logf("Created customer: %s (ID: %s)", customer.Name, customer.ID)

	// -------------------------------------------------------------------------
	// 3. Wait for async workflow execution
	// -------------------------------------------------------------------------

	if !waitForProcessing(t, tsd.WF.QueueManager, initialMetrics, 5*time.Second) {
		t.Fatal("workflow did not process in time")
	}

	// -------------------------------------------------------------------------
	// 4. Verify workflow was triggered
	// -------------------------------------------------------------------------

	finalMetrics := tsd.WF.QueueManager.GetMetrics()
	if finalMetrics.TotalEnqueued <= initialMetrics.TotalEnqueued {
		t.Error("expected event to be enqueued")
	}

	if finalMetrics.TotalProcessed <= initialMetrics.TotalProcessed {
		t.Error("expected event to be processed")
	}

	// Verify execution history shows on_create event for customers
	history := tsd.WF.Engine.GetExecutionHistory(100)
	foundCreate := false
	for _, exec := range history {
		if exec.ExecutionPlan.TriggerEvent.EventType == "on_create" &&
			exec.ExecutionPlan.TriggerEvent.EntityName == "customers" {
			foundCreate = true

			// Verify at least one rule matched
			if exec.ExecutionPlan.MatchedRuleCount == 0 {
				t.Error("expected at least one matched rule")
			}

			// Verify execution completed
			if exec.Status != workflow.StatusCompleted {
				t.Errorf("expected completed status, got %s", exec.Status)
			}
			break
		}
	}

	if !foundCreate {
		t.Error("on_create event for customers not found in execution history")
	}

	t.Log("SUCCESS: Customer creation triggered workflow via delegate")
}

// testCustomerUpdateTriggersWorkflow tests that updating a customer via customersbus
// fires an on_update event that triggers a matching workflow.
func testCustomerUpdateTriggersWorkflow(t *testing.T, tsd TriggerTestData) {
	ctx := context.Background()

	// -------------------------------------------------------------------------
	// 1. Create a workflow rule that triggers on customer update
	// -------------------------------------------------------------------------

	customerEntity, err := tsd.WF.WorkflowBus.QueryEntityByName(ctx, "customers")
	if err != nil {
		t.Fatalf("querying customers entity: %s", err)
	}

	entityType, err := tsd.WF.WorkflowBus.QueryEntityTypeByName(ctx, "table")
	if err != nil {
		t.Fatalf("querying entity type: %s", err)
	}

	triggerTypeUpdate, err := tsd.WF.WorkflowBus.QueryTriggerTypeByName(ctx, "on_update")
	if err != nil {
		t.Fatalf("querying on_update trigger type: %s", err)
	}

	// Create automation rule for updates
	rule, err := tsd.WF.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Customer Updated - Trigger Test " + uuid.New().String()[:8],
		Description:   "Tests delegate pattern fires on_update events",
		EntityID:      customerEntity.ID,
		EntityTypeID:  entityType.ID,
		TriggerTypeID: triggerTypeUpdate.ID,
		IsActive:      true,
		CreatedBy:     tsd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating on_update rule: %s", err)
	}

	action, err := tsd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Update Customer Alert",
		ActionConfig: json.RawMessage(`{
			"alert_type": "customer_updated",
			"severity": "medium",
			"title": "Customer Updated",
			"message": "Customer was updated via trigger test",
			"recipients": {"users": ["` + tsd.Users[0].ID.String() + `"], "roles": []}
		}`),
		IsActive:       true,
		TemplateID:     &tsd.CreateAlertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating rule action: %s", err)
	}

	// Create start edge
	_, err = tsd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: action.ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating start edge: %s", err)
	}

	// Rule cache is now automatically invalidated via delegate pattern
	// No need to manually reset the engine

	// -------------------------------------------------------------------------
	// 2. Create a customer first
	// -------------------------------------------------------------------------

	newCustomer := customersbus.NewCustomers{
		Name:              "Update Test Customer " + uuid.New().String()[:8],
		DeliveryAddressID: tsd.StreetID,
		ContactID:         tsd.ContactInfoID,
		CreatedBy:         tsd.Users[0].ID,
	}

	customer, err := tsd.CustomersBus.Create(ctx, newCustomer)
	if err != nil {
		t.Fatalf("creating customer for update test: %s", err)
	}

	// Wait for create event to process
	time.Sleep(500 * time.Millisecond)

	// -------------------------------------------------------------------------
	// 3. Update the customer
	// -------------------------------------------------------------------------

	initialMetrics := tsd.WF.QueueManager.GetMetrics()

	updatedName := customer.Name + "-UPDATED"
	userID := tsd.Users[0].ID
	updateCustomer := customersbus.UpdateCustomers{
		Name:      &updatedName,
		UpdatedBy: &userID,
	}

	_, err = tsd.CustomersBus.Update(ctx, customer, updateCustomer)
	if err != nil {
		t.Fatalf("updating customer: %s", err)
	}

	t.Logf("Updated customer: %s -> %s", customer.Name, updatedName)

	// -------------------------------------------------------------------------
	// 4. Wait and verify
	// -------------------------------------------------------------------------

	if !waitForProcessing(t, tsd.WF.QueueManager, initialMetrics, 5*time.Second) {
		t.Fatal("workflow did not process in time")
	}

	finalMetrics := tsd.WF.QueueManager.GetMetrics()
	if finalMetrics.TotalEnqueued <= initialMetrics.TotalEnqueued {
		t.Error("expected event to be enqueued")
	}

	// Verify on_update event in history
	history := tsd.WF.Engine.GetExecutionHistory(50)
	foundUpdate := false
	for _, exec := range history {
		if exec.ExecutionPlan.TriggerEvent.EventType == "on_update" &&
			exec.ExecutionPlan.TriggerEvent.EntityName == "customers" {
			foundUpdate = true
			break
		}
	}

	if !foundUpdate {
		t.Error("on_update event for customers not found in execution history")
	}

	t.Log("SUCCESS: Customer update triggered workflow via delegate")
}

// testInactiveRuleNoTrigger tests that inactive rules don't trigger workflows.
func testInactiveRuleNoTrigger(t *testing.T, tsd TriggerTestData) {
	ctx := context.Background()

	// -------------------------------------------------------------------------
	// 1. Create an INACTIVE workflow rule
	// -------------------------------------------------------------------------

	customerEntity, err := tsd.WF.WorkflowBus.QueryEntityByName(ctx, "customers")
	if err != nil {
		t.Fatalf("querying customers entity: %s", err)
	}

	entityType, err := tsd.WF.WorkflowBus.QueryEntityTypeByName(ctx, "table")
	if err != nil {
		t.Fatalf("querying entity type: %s", err)
	}

	triggerTypeCreate, err := tsd.WF.WorkflowBus.QueryTriggerTypeByName(ctx, "on_create")
	if err != nil {
		t.Fatalf("querying on_create trigger type: %s", err)
	}

	rule, err := tsd.WF.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Inactive Customer Rule - Should Not Trigger " + uuid.New().String()[:8],
		Description:   "This rule is inactive and should not trigger",
		EntityID:      customerEntity.ID,
		EntityTypeID:  entityType.ID,
		TriggerTypeID: triggerTypeCreate.ID,
		IsActive:      false, // INACTIVE
		CreatedBy:     tsd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating inactive rule: %s", err)
	}

	action, err := tsd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Should Not Execute",
		ActionConfig: json.RawMessage(`{
			"alert_type": "inactive_test",
			"severity": "high",
			"title": "This should not execute",
			"message": "If you see this, the inactive rule was incorrectly triggered",
			"recipients": {"users": ["` + tsd.Users[0].ID.String() + `"], "roles": []}
		}`),
		IsActive:       true,
		TemplateID:     &tsd.CreateAlertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating rule action: %s", err)
	}

	// Create start edge
	_, err = tsd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: action.ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating start edge: %s", err)
	}

	// Rule cache is now automatically invalidated via delegate pattern
	// No need to manually reset the engine

	// -------------------------------------------------------------------------
	// 2. Get baseline and create a customer
	// -------------------------------------------------------------------------

	// Count current matching executions for this specific rule
	history := tsd.WF.Engine.GetExecutionHistory(100)
	initialRuleMatchCount := countRuleMatches(history, rule.ID)

	newCustomer := customersbus.NewCustomers{
		Name:              "Inactive Rule Test Customer " + uuid.New().String()[:8],
		DeliveryAddressID: tsd.StreetID,
		ContactID:         tsd.ContactInfoID,
		CreatedBy:         tsd.Users[0].ID,
	}

	_, err = tsd.CustomersBus.Create(ctx, newCustomer)
	if err != nil {
		t.Fatalf("creating customer: %s", err)
	}

	// Wait for potential processing
	time.Sleep(1 * time.Second)

	// -------------------------------------------------------------------------
	// 3. Verify the inactive rule was NOT triggered
	// -------------------------------------------------------------------------

	finalHistory := tsd.WF.Engine.GetExecutionHistory(100)
	finalRuleMatchCount := countRuleMatches(finalHistory, rule.ID)

	if finalRuleMatchCount > initialRuleMatchCount {
		t.Error("inactive rule should NOT have triggered")
	}

	t.Log("SUCCESS: Inactive rule did not trigger")
}

// countRuleMatches counts how many times a specific rule was matched in execution history.
func countRuleMatches(history []*workflow.WorkflowExecution, ruleID uuid.UUID) int {
	count := 0
	for _, exec := range history {
		for _, batch := range exec.BatchResults {
			for _, ruleResult := range batch.RuleResults {
				if ruleResult.RuleID == ruleID {
					count++
				}
			}
		}
	}
	return count
}
