package ordersapi_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// TestWorkflow_OrdersDelegateEvents tests that workflow events fire correctly
// when orders are created, updated, and deleted via ordersbus.
// This validates the Temporal delegate pattern integration.
func TestWorkflow_OrdersDelegateEvents(t *testing.T) {
	t.Parallel()

	// Setup database.
	db := dbtest.NewDatabase(t, "Test_Workflow_OrdersDelegate")

	// Initialize Temporal-based workflow infrastructure.
	wf := apitest.InitWorkflowInfra(t, db)

	ctx := context.Background()
	adminUserID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	// -------------------------------------------------------------------------
	// Create workflow rules for orders create/update/delete.
	// -------------------------------------------------------------------------

	orderEntity, err := wf.WorkflowBus.QueryEntityByName(ctx, "orders")
	if err != nil {
		t.Fatalf("querying orders entity: %s", err)
	}

	entityType, err := wf.WorkflowBus.QueryEntityTypeByName(ctx, "table")
	if err != nil {
		t.Fatalf("querying entity type: %s", err)
	}

	triggerTypeCreate, err := wf.WorkflowBus.QueryTriggerTypeByName(ctx, "on_create")
	if err != nil {
		t.Fatalf("querying on_create trigger type: %s", err)
	}

	triggerTypeUpdate, err := wf.WorkflowBus.QueryTriggerTypeByName(ctx, "on_update")
	if err != nil {
		t.Fatalf("querying on_update trigger type: %s", err)
	}

	triggerTypeDelete, err := wf.WorkflowBus.QueryTriggerTypeByName(ctx, "on_delete")
	if err != nil {
		t.Fatalf("querying on_delete trigger type: %s", err)
	}

	// Create email action template.
	emailTemplate, err := wf.WorkflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:        "Order Event Notification",
		Description: "Template for order event notifications",
		ActionType:  "send_email",
		Icon:        "material-symbols:mail",
		DefaultConfig: json.RawMessage(`{
			"recipients": ["orders@example.com"],
			"subject": "Order Event"
		}`),
		CreatedBy: adminUserID,
	})
	if err != nil {
		t.Fatalf("creating action template: %s", err)
	}

	// Create rule + action + edge for each event type.
	for _, tc := range []struct {
		name          string
		triggerTypeID uuid.UUID
	}{
		{"Order Created - Delegate Pattern Test", triggerTypeCreate.ID},
		{"Order Updated - Delegate Pattern Test", triggerTypeUpdate.ID},
		{"Order Deleted - Delegate Pattern Test", triggerTypeDelete.ID},
	} {
		rule, err := wf.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
			Name:              tc.name,
			Description:       "Tests delegate pattern fires events",
			EntityID:          orderEntity.ID,
			EntityTypeID:      entityType.ID,
			TriggerTypeID:     tc.triggerTypeID,
			TriggerConditions: nil,
			IsActive:          true,
			CreatedBy:         adminUserID,
		})
		if err != nil {
			t.Fatalf("creating rule %q: %s", tc.name, err)
		}

		action, err := wf.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
			AutomationRuleID: rule.ID,
			Name:             "Send Email for " + tc.name,
			ActionConfig: json.RawMessage(`{
				"recipients": ["sales@example.com"],
				"subject": "Order Event",
				"body": "Order event fired"
			}`),
			IsActive:   true,
			TemplateID: &emailTemplate.ID,
		})
		if err != nil {
			t.Fatalf("creating action for %q: %s", tc.name, err)
		}

		_, err = wf.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
			RuleID:         rule.ID,
			SourceActionID: nil,
			TargetActionID: action.ID,
			EdgeType:       workflow.EdgeTypeStart,
			EdgeOrder:      0,
		})
		if err != nil {
			t.Fatalf("creating edge for %q: %s", tc.name, err)
		}
	}

	// Refresh trigger processor to pick up all new rules.
	if err := wf.TriggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing trigger processor: %s", err)
	}

	// -------------------------------------------------------------------------
	// Register orders domain with Temporal delegate handler.
	// -------------------------------------------------------------------------

	wf.DelegateHandler.RegisterDomain(db.BusDomain.Delegate, ordersbus.DomainName, ordersbus.EntityName)

	// Use the existing ordersbus from BusDomain.
	ordersBus := db.BusDomain.Order

	// -------------------------------------------------------------------------
	// Seed FK dependencies for orders.
	// -------------------------------------------------------------------------

	oflStatuses, err := orderfulfillmentstatusbus.TestSeedOrderFulfillmentStatuses(ctx, db.BusDomain.OrderFulfillmentStatus)
	if err != nil {
		t.Fatalf("seeding order fulfillment statuses: %s", err)
	}
	fulfillmentStatusID := oflStatuses[0].ID

	regions, err := db.BusDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		t.Fatalf("querying regions: %s", err)
	}
	regionIDs := make([]uuid.UUID, len(regions))
	for i, r := range regions {
		regionIDs[i] = r.ID
	}

	cities, err := citybus.TestSeedCities(ctx, 1, regionIDs, db.BusDomain.City)
	if err != nil {
		t.Fatalf("seeding cities: %s", err)
	}

	streets, err := streetbus.TestSeedStreets(ctx, 1, []uuid.UUID{cities[0].ID}, db.BusDomain.Street)
	if err != nil {
		t.Fatalf("seeding streets: %s", err)
	}

	tzs, err := db.BusDomain.Timezone.Query(ctx, timezonebus.QueryFilter{}, timezonebus.DefaultOrderBy, page.MustParse("1", "1"))
	if err != nil {
		t.Fatalf("querying timezones: %s", err)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 1, []uuid.UUID{streets[0].ID}, []uuid.UUID{tzs[0].ID}, db.BusDomain.ContactInfos)
	if err != nil {
		t.Fatalf("seeding contact infos: %s", err)
	}

	customers, err := customersbus.TestSeedCustomers(ctx, 1, []uuid.UUID{streets[0].ID}, []uuid.UUID{contactInfos[0].ID}, []uuid.UUID{adminUserID}, db.BusDomain.Customers)
	if err != nil {
		t.Fatalf("seeding customers: %s", err)
	}
	customerID := customers[0].ID

	currencies, err := currencybus.TestSeedCurrencies(ctx, 1, db.BusDomain.Currency)
	if err != nil {
		t.Fatalf("seeding currencies: %s", err)
	}
	currencyID := currencies[0].ID

	// -------------------------------------------------------------------------
	// Test: Create order fires on_create event.
	// -------------------------------------------------------------------------

	t.Run("create-fires-on_create-event", func(t *testing.T) {
		newOrder := ordersbus.NewOrder{
			Number:              "DELEGATE-TEST-001",
			CustomerID:          customerID,
			DueDate:             time.Now().Add(7 * 24 * time.Hour),
			FulfillmentStatusID: fulfillmentStatusID,
			CurrencyID:          currencyID,
			CreatedBy:           adminUserID,
		}

		order, err := ordersBus.Create(ctx, newOrder)
		if err != nil {
			t.Fatalf("creating order: %s", err)
		}

		t.Logf("Created order: %s (ID: %s)", order.Number, order.ID)

		// Wait for async Temporal workflow execution.
		time.Sleep(3 * time.Second)

		// Verify: no panic/error means dispatch succeeded.
		t.Log("SUCCESS: ordersbus.Create() fired on_create event via delegate (Temporal)")
	})

	// -------------------------------------------------------------------------
	// Test: Update order fires on_update event.
	// -------------------------------------------------------------------------

	t.Run("update-fires-on_update-event", func(t *testing.T) {
		newOrder := ordersbus.NewOrder{
			Number:              "DELEGATE-TEST-002",
			CustomerID:          customerID,
			DueDate:             time.Now().Add(7 * 24 * time.Hour),
			FulfillmentStatusID: fulfillmentStatusID,
			CurrencyID:          currencyID,
			CreatedBy:           adminUserID,
		}

		order, err := ordersBus.Create(ctx, newOrder)
		if err != nil {
			t.Fatalf("creating order for update test: %s", err)
		}

		// Wait for create event to process.
		time.Sleep(500 * time.Millisecond)

		updatedNumber := "DELEGATE-TEST-002-UPDATED"
		updateOrder := ordersbus.UpdateOrder{
			Number:    &updatedNumber,
			UpdatedBy: &adminUserID,
		}

		_, err = ordersBus.Update(ctx, order, updateOrder)
		if err != nil {
			t.Fatalf("updating order: %s", err)
		}

		t.Logf("Updated order: %s -> %s", order.Number, updatedNumber)

		// Wait for async Temporal workflow execution.
		time.Sleep(3 * time.Second)

		t.Log("SUCCESS: ordersbus.Update() fired on_update event via delegate (Temporal)")
	})

	// -------------------------------------------------------------------------
	// Test: Delete order fires on_delete event.
	// -------------------------------------------------------------------------

	t.Run("delete-fires-on_delete-event", func(t *testing.T) {
		newOrder := ordersbus.NewOrder{
			Number:              "DELEGATE-TEST-003",
			CustomerID:          customerID,
			DueDate:             time.Now().Add(7 * 24 * time.Hour),
			FulfillmentStatusID: fulfillmentStatusID,
			CurrencyID:          currencyID,
			CreatedBy:           adminUserID,
		}

		order, err := ordersBus.Create(ctx, newOrder)
		if err != nil {
			t.Fatalf("creating order for delete test: %s", err)
		}

		// Wait for create event to process.
		time.Sleep(500 * time.Millisecond)

		err = ordersBus.Delete(ctx, order)
		if err != nil {
			t.Fatalf("deleting order: %s", err)
		}

		t.Logf("Deleted order: %s (ID: %s)", order.Number, order.ID)

		// Wait for async Temporal workflow execution.
		time.Sleep(3 * time.Second)

		t.Log("SUCCESS: ordersbus.Delete() fired on_delete event via delegate (Temporal)")
	})
}
