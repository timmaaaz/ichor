// Package actionhandlers_test contains standalone integration tests for
// communication workflow actions using a real Temporal container.
//
// TestCallWebhookAction — spins up a local HTTP server and verifies the
// call_webhook handler delivers an outbound HTTP request end-to-end through
// Temporal.
//
// TestSendEmailAction / TestSendNotificationAction — verify that handlers
// registered with nil SMTP/queue clients do not panic; success is defined as
// the workflow completing (or timing out gracefully) without crashing the test.
package actionhandlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.temporal.io/api/workflowservice/v1"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus/stores/alertdb"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	workflowtemporal "github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal/stores/edgedb"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/communication"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/integration"
	foundationtemporal "github.com/timmaaaz/ichor/foundation/temporal"
	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// =============================================================================
// call_webhook Action Test
// =============================================================================

// TestCallWebhookAction verifies the call_webhook handler end-to-end:
//  1. Spins up a local plain-HTTP test server to receive the webhook.
//  2. Creates a workflow rule with a call_webhook action pointing at that server.
//  3. Fires a trigger event.
//  4. Waits for the test server to receive the request (success) or times out (fail).
//
// The test server URL uses 127.0.0.1 (loopback), which the handler's Validate()
// method permits even for plain HTTP because isInternalHost returns true for
// loopback addresses.  A custom httptest server client is injected via
// NewCallWebhookHandlerWithClient so TLS is not required.
func TestCallWebhookAction(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_CallWebhookAction")
	ctx := context.Background()

	// ── Local HTTP server to capture the incoming webhook ─────────────────────

	received := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case received <- struct{}{}:
		default:
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	// ── Minimal Temporal worker with call_webhook (custom HTTP client) ─────────
	//
	// We use NewCallWebhookHandlerWithClient and pass the test server's own
	// http.Client so the handler reaches the local server without TLS issues.

	container := foundationtemporal.GetTestContainer(t)
	tc, err := temporalclient.Dial(temporalclient.Options{HostPort: container.HostPort})
	if err != nil {
		t.Fatalf("connecting to Temporal: %v", err)
	}

	taskQueue := "test-workflow-webhook-" + t.Name()
	w := worker.New(tc, taskQueue, worker.Options{})
	w.RegisterWorkflow(workflowtemporal.ExecuteGraphWorkflow)
	w.RegisterWorkflow(workflowtemporal.ExecuteBranchUntilConvergence)

	registry := workflow.NewActionRegistry()
	registry.Register(integration.NewCallWebhookHandlerWithClient(db.Log, server.Client()))

	activities := &workflowtemporal.Activities{
		Registry:      registry,
		AsyncRegistry: workflowtemporal.NewAsyncRegistry(),
	}
	w.RegisterActivity(activities)

	if err := w.Start(); err != nil {
		tc.Close()
		t.Fatalf("starting Temporal worker: %v", err)
	}
	t.Cleanup(func() { w.Stop(); tc.Close() })

	// ── Workflow infrastructure ────────────────────────────────────────────────

	workflowStore := workflowdb.NewStore(db.Log, db.DB)
	workflowBus := workflow.NewBusiness(db.Log, db.BusDomain.Delegate, workflowStore)
	edgeStore := edgedb.NewStore(db.Log, db.DB)
	triggerProcessor := workflow.NewTriggerProcessor(db.Log, db.DB, workflowBus)
	if err := triggerProcessor.Initialize(ctx); err != nil {
		t.Fatalf("initializing trigger processor: %v", err)
	}
	workflowTrigger := workflowtemporal.NewWorkflowTrigger(db.Log, tc, triggerProcessor, edgeStore, workflowStore).
		WithTaskQueue(taskQueue)

	// ── Seed workflow entity / trigger type ───────────────────────────────────
	//
	// Use the well-known "customers" entity + "on_create" trigger that are
	// present in every seeded test database (confirmed by actions_inventory_test.go).

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, db.BusDomain.User)
	if err != nil {
		t.Fatalf("seeding admin users: %v", err)
	}
	adminID := admins[0].ID

	customerEntity, err := workflowBus.QueryEntityByName(ctx, "customers")
	if err != nil {
		t.Fatalf("querying customers entity: %v", err)
	}
	entityType, err := workflowBus.QueryEntityTypeByName(ctx, "table")
	if err != nil {
		t.Fatalf("querying entity type: %v", err)
	}
	triggerTypeCreate, err := workflowBus.QueryTriggerTypeByName(ctx, "on_create")
	if err != nil {
		t.Fatalf("querying on_create trigger type: %v", err)
	}

	// ── Seed workflow rule + action + edge ────────────────────────────────────

	rule, err := workflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "CallWebhook-Test-" + uuid.New().String()[:8],
		Description:   "Integration test for call_webhook action",
		EntityID:      customerEntity.ID,
		EntityTypeID:  entityType.ID,
		TriggerTypeID: triggerTypeCreate.ID,
		IsActive:      true,
		CreatedBy:     adminID,
	})
	if err != nil {
		t.Fatalf("creating rule: %v", err)
	}

	actionConfig := map[string]any{
		"url":             server.URL + "/webhook",
		"method":          "POST",
		"timeout_seconds": 10,
	}
	configBytes, _ := json.Marshal(actionConfig)

	webhookTemplate, err := workflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:          "Call Webhook Template",
		Description:   "Template for call_webhook test",
		ActionType:    "call_webhook",
		DefaultConfig: json.RawMessage(configBytes),
		CreatedBy:     adminID,
	})
	if err != nil {
		t.Fatalf("creating call_webhook template: %v", err)
	}

	action, err := workflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Fire test webhook",
		Description:      "Sends POST to local test server",
		ActionConfig:     json.RawMessage(configBytes),
		IsActive:         true,
		TemplateID:       &webhookTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating call_webhook action: %v", err)
	}

	_, err = workflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: action.ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating start edge: %v", err)
	}

	// Refresh so TriggerProcessor sees the new rule.
	if err := triggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing trigger processor: %v", err)
	}

	// ── Fire trigger event ────────────────────────────────────────────────────

	event := workflow.TriggerEvent{
		EventType:  "on_create",
		EntityName: "customers",
		EntityID:   uuid.New(),
		Timestamp:  time.Now(),
		RawData:    map[string]any{"test": true},
		UserID:     adminID,
	}
	if err := workflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("firing trigger event: %v", err)
	}

	// ── Wait for webhook to arrive ────────────────────────────────────────────

	select {
	case <-received:
		t.Log("SUCCESS: webhook received by local test server")
	case <-time.After(15 * time.Second):
		t.Fatal("timeout: webhook was not received after 15s — call_webhook handler may have failed")
	}
}

// =============================================================================
// send_email Action Test
// =============================================================================

// TestSendEmailAction verifies that the send_email handler, when registered
// with a nil SMTP client, does not panic when a workflow triggers it.
// Success is defined as the test completing without a fatal error or panic —
// no actual email is delivered.
func TestSendEmailAction(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_SendEmailAction")
	ctx := context.Background()

	container := foundationtemporal.GetTestContainer(t)
	tc, err := temporalclient.Dial(temporalclient.Options{HostPort: container.HostPort})
	if err != nil {
		t.Fatalf("connecting to Temporal: %v", err)
	}

	taskQueue := "test-workflow-email-" + t.Name()
	w := worker.New(tc, taskQueue, worker.Options{})
	w.RegisterWorkflow(workflowtemporal.ExecuteGraphWorkflow)
	w.RegisterWorkflow(workflowtemporal.ExecuteBranchUntilConvergence)

	// send_email with nil SMTP client → logs a warning and returns without error.
	registry := workflow.NewActionRegistry()
	registry.Register(communication.NewSendEmailHandler(db.Log, db.DB, nil, ""))

	activities := &workflowtemporal.Activities{
		Registry:      registry,
		AsyncRegistry: workflowtemporal.NewAsyncRegistry(),
	}
	w.RegisterActivity(activities)

	if err := w.Start(); err != nil {
		tc.Close()
		t.Fatalf("starting Temporal worker: %v", err)
	}
	t.Cleanup(func() { w.Stop(); tc.Close() })

	workflowStore := workflowdb.NewStore(db.Log, db.DB)
	workflowBus := workflow.NewBusiness(db.Log, db.BusDomain.Delegate, workflowStore)
	edgeStore := edgedb.NewStore(db.Log, db.DB)
	triggerProcessor := workflow.NewTriggerProcessor(db.Log, db.DB, workflowBus)
	if err := triggerProcessor.Initialize(ctx); err != nil {
		t.Fatalf("initializing trigger processor: %v", err)
	}
	workflowTrigger := workflowtemporal.NewWorkflowTrigger(db.Log, tc, triggerProcessor, edgeStore, workflowStore).
		WithTaskQueue(taskQueue)

	adminID, rule, action := seedCommsActionRule(t, ctx, workflowBus, db.BusDomain.User, "send_email",
		`{"recipients":["test@example.com"],"subject":"Test","body":"Hello from workflow"}`)

	_, err = workflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: action.ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating start edge: %v", err)
	}

	if err := triggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing trigger processor: %v", err)
	}

	event := workflow.TriggerEvent{
		EventType:  "on_create",
		EntityName: "customers",
		EntityID:   uuid.New(),
		Timestamp:  time.Now(),
		RawData:    map[string]any{"test": true},
		UserID:     adminID,
	}
	if err := workflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("firing trigger event: %v", err)
	}

	var completed bool
	for i := 0; i < 30; i++ {
		resp, err := tc.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Namespace: "default",
			PageSize:  10,
			Query:     fmt.Sprintf(`TaskQueue = "%s" AND ExecutionStatus = "Completed"`, taskQueue),
		})
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if len(resp.Executions) > 0 {
			completed = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !completed {
		t.Fatal("timeout: send_email workflow did not complete within 15s")
	}
	t.Log("SUCCESS: send_email with nil SMTP client completed without error")
}

// =============================================================================
// send_notification Action Test
// =============================================================================

// TestSendNotificationAction verifies that the send_notification handler, registered
// with a real alertBus and a nil queue client, completes when a workflow triggers it.
func TestSendNotificationAction(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_SendNotificationAction")
	ctx := context.Background()

	container := foundationtemporal.GetTestContainer(t)
	tc, err := temporalclient.Dial(temporalclient.Options{HostPort: container.HostPort})
	if err != nil {
		t.Fatalf("connecting to Temporal: %v", err)
	}

	taskQueue := "test-workflow-notif-" + t.Name()
	w := worker.New(tc, taskQueue, worker.Options{})
	w.RegisterWorkflow(workflowtemporal.ExecuteGraphWorkflow)
	w.RegisterWorkflow(workflowtemporal.ExecuteBranchUntilConvergence)

	// send_notification with real alertBus → persists notification alert row.
	alertsStore := alertdb.NewStore(db.Log, db.DB)
	alertsBus := alertbus.NewBusiness(db.Log, alertsStore)

	registry := workflow.NewActionRegistry()
	registry.Register(communication.NewSendNotificationHandler(db.Log, alertsBus, nil))
	registry.Register(communication.NewCreateAlertHandler(db.Log, alertsBus, nil, nil, nil))

	activities := &workflowtemporal.Activities{
		Registry:      registry,
		AsyncRegistry: workflowtemporal.NewAsyncRegistry(),
	}
	w.RegisterActivity(activities)

	if err := w.Start(); err != nil {
		tc.Close()
		t.Fatalf("starting Temporal worker: %v", err)
	}
	t.Cleanup(func() { w.Stop(); tc.Close() })

	workflowStore := workflowdb.NewStore(db.Log, db.DB)
	workflowBus := workflow.NewBusiness(db.Log, db.BusDomain.Delegate, workflowStore)
	edgeStore := edgedb.NewStore(db.Log, db.DB)
	triggerProcessor := workflow.NewTriggerProcessor(db.Log, db.DB, workflowBus)
	if err := triggerProcessor.Initialize(ctx); err != nil {
		t.Fatalf("initializing trigger processor: %v", err)
	}
	workflowTrigger := workflowtemporal.NewWorkflowTrigger(db.Log, tc, triggerProcessor, edgeStore, workflowStore).
		WithTaskQueue(taskQueue)

	adminID, rule, action := seedCommsActionRule(t, ctx, workflowBus, db.BusDomain.User, "send_notification",
		`{"recipients":["5cf37266-3473-4006-984f-9325122678b7"],"priority":"low","title":"Test","message":"Hello"}`)

	_, err = workflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: action.ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating start edge: %v", err)
	}

	if err := triggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing trigger processor: %v", err)
	}

	event := workflow.TriggerEvent{
		EventType:  "on_create",
		EntityName: "customers",
		EntityID:   uuid.New(),
		Timestamp:  time.Now(),
		RawData:    map[string]any{"test": true},
		UserID:     adminID,
	}
	if err := workflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("firing trigger event: %v", err)
	}

	var completed bool
	for i := 0; i < 30; i++ {
		resp, err := tc.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Namespace: "default",
			PageSize:  10,
			Query:     fmt.Sprintf(`TaskQueue = "%s" AND ExecutionStatus = "Completed"`, taskQueue),
		})
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if len(resp.Executions) > 0 {
			completed = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !completed {
		t.Fatal("timeout: send_notification workflow did not complete within 15s")
	}
	t.Log("SUCCESS: send_notification (real alertBus, nil queue client) completed without error")
}

// =============================================================================
// send_notification Persistence Test
// =============================================================================

// TestSendNotificationPersistence is the DB-backed behavioral proof that
// send_notification persists a real alert row of type "notification" via the
// alertBus, including template substitution in the message field.
// It does NOT use Temporal — it calls the handler directly — to keep the test
// focused on the persistence guarantee rather than the full execution pipeline.
func TestSendNotificationPersistence(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_SendNotificationPersistence")
	ctx := context.Background()

	// Seed one real user to act as the notification recipient.
	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, db.BusDomain.User)
	if err != nil {
		t.Fatalf("seeding users: %v", err)
	}
	recipientID := users[0].ID

	// Build a real alertBus backed by the test DB.
	alertsStore := alertdb.NewStore(db.Log, db.DB)
	alertsBus := alertbus.NewBusiness(db.Log, alertsStore)

	// Construct handler with real alertBus; nil workflowQueue is intentional —
	// this test proves persistence, not WebSocket delivery.
	handler := communication.NewSendNotificationHandler(db.Log, alertsBus, nil)

	// Seed a real rule so source_rule_id satisfies the FK on workflow.alerts.
	workflowStore := workflowdb.NewStore(db.Log, db.DB)
	workflowBus := workflow.NewBusiness(db.Log, db.BusDomain.Delegate, workflowStore)
	_, rule, _ := seedCommsActionRule(t, ctx, workflowBus, db.BusDomain.User, "send_notification",
		fmt.Sprintf(`{"recipients":["%s"],"priority":"high","message":"Order {{order_number}} shipped","title":"Shipped"}`, recipientID))
	ruleID := rule.ID

	config := json.RawMessage(fmt.Sprintf(`{
		"recipients": ["%s"],
		"priority": "high",
		"message": "Order {{order_number}} shipped",
		"title": "Shipped"
	}`, recipientID))

	execCtx := workflow.ActionExecutionContext{
		EntityID:    uuid.New(),
		EntityName:  "orders",
		EventType:   "on_update",
		UserID:      recipientID,
		RuleID:      &ruleID,
		RuleName:    rule.Name,
		ExecutionID: uuid.New(),
		RawData: map[string]interface{}{
			"order_number": "ORD-7",
		},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected result type: %T", result)
	}

	if resultMap["status"] != "sent" {
		t.Fatalf("expected status=sent, got %v", resultMap["status"])
	}

	// Query alerts scoped to the rule — proves the row was persisted.
	alerts, err := alertsBus.Query(
		ctx,
		alertbus.QueryFilter{SourceRuleID: &ruleID},
		order.NewBy(alertbus.OrderByCreatedDate, order.DESC),
		page.MustParse("1", "10"),
	)
	if err != nil {
		t.Fatalf("querying alerts: %v", err)
	}

	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}

	got := alerts[0]

	if got.AlertType != "notification" {
		t.Errorf("AlertType: got %q, want %q", got.AlertType, "notification")
	}
	if got.Severity != alertbus.SeverityHigh {
		t.Errorf("Severity: got %q, want %q", got.Severity, alertbus.SeverityHigh)
	}
	if got.Status != alertbus.StatusActive {
		t.Errorf("Status: got %q, want %q", got.Status, alertbus.StatusActive)
	}
	// Proves template substitution: {{order_number}} → ORD-7
	wantMsg := "Order ORD-7 shipped"
	if got.Message != wantMsg {
		t.Errorf("Message: got %q, want %q", got.Message, wantMsg)
	}
	// Carries the triggering entity's name (like create_alert) so the frontend
	// can show context. SourceEntityID stays nil on purpose (bundling).
	if got.SourceEntityName != execCtx.EntityName {
		t.Errorf("SourceEntityName: got %q, want %q", got.SourceEntityName, execCtx.EntityName)
	}

	// Verify the recipient row was created for the seeded user.
	recipients, err := alertsBus.QueryRecipientsByAlertID(ctx, got.ID)
	if err != nil {
		t.Fatalf("querying recipients: %v", err)
	}
	if len(recipients) != 1 {
		t.Fatalf("expected 1 recipient row, got %d", len(recipients))
	}
	if recipients[0].RecipientID != recipientID {
		t.Errorf("RecipientID: got %v, want %v", recipients[0].RecipientID, recipientID)
	}
	if recipients[0].RecipientType != "user" {
		t.Errorf("RecipientType: got %q, want %q", recipients[0].RecipientType, "user")
	}

	t.Logf("SUCCESS: send_notification persisted alert %s with message %q and 1 recipient", got.ID, got.Message)
}

// =============================================================================
// Helpers
// =============================================================================

// seedCommsActionRule seeds a minimal workflow rule with a single action of
// the given type and returns (createdByID, rule, action).
func seedCommsActionRule(
	t *testing.T,
	ctx context.Context,
	workflowBus *workflow.Business,
	userBus *userbus.Business,
	actionType string,
	configJSON string,
) (uuid.UUID, workflow.AutomationRule, workflow.RuleAction) {
	t.Helper()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, userBus)
	if err != nil {
		t.Fatalf("seeding admin users: %v", err)
	}
	adminID := admins[0].ID

	customerEntity, err := workflowBus.QueryEntityByName(ctx, "customers")
	if err != nil {
		t.Fatalf("querying customers entity: %v", err)
	}
	entityType, err := workflowBus.QueryEntityTypeByName(ctx, "table")
	if err != nil {
		t.Fatalf("querying entity type: %v", err)
	}
	triggerTypeCreate, err := workflowBus.QueryTriggerTypeByName(ctx, "on_create")
	if err != nil {
		t.Fatalf("querying on_create trigger type: %v", err)
	}

	rule, err := workflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          actionType + "-Test-" + uuid.New().String()[:8],
		Description:   "Integration test for " + actionType + " action",
		EntityID:      customerEntity.ID,
		EntityTypeID:  entityType.ID,
		TriggerTypeID: triggerTypeCreate.ID,
		IsActive:      true,
		CreatedBy:     adminID,
	})
	if err != nil {
		t.Fatalf("creating rule: %v", err)
	}

	tmpl, err := workflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:          actionType + " Template",
		Description:   "Template for " + actionType + " test",
		ActionType:    actionType,
		DefaultConfig: json.RawMessage(configJSON),
		CreatedBy:     adminID,
	})
	if err != nil {
		t.Fatalf("creating %s template: %v", actionType, err)
	}

	action, err := workflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             actionType + " action",
		Description:      "Test " + actionType,
		ActionConfig:     json.RawMessage(configJSON),
		IsActive:         true,
		TemplateID:       &tmpl.ID,
	})
	if err != nil {
		t.Fatalf("creating %s action: %v", actionType, err)
	}

	return adminID, rule, action
}
