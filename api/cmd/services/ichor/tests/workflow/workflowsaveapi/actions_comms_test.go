// Package workflowsaveapi_test contains standalone integration tests for
// communication workflow actions using a real Temporal container.
//
// TestCallWebhookAction — spins up a local HTTP server and verifies the
// call_webhook handler delivers an outbound HTTP request end-to-end through
// Temporal.
//
// TestSendEmailAction / TestSendNotificationAction — verify that handlers
// registered with nil SMTP/queue clients do not panic; success is defined as
// the workflow completing (or timing out gracefully) without crashing the test.
package workflowsaveapi_test

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
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus/stores/alertdb"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
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

	adminID := uuid.New()

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

	adminID, rule, action := seedCommsActionRule(t, ctx, workflowBus, "send_email",
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

// TestSendNotificationAction verifies that the send_notification handler, when
// registered with a nil queue client, does not panic when a workflow triggers it.
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

	// send_notification with nil queue → logs a warning and returns without error.
	alertsStore := alertdb.NewStore(db.Log, db.DB)
	alertsBus := alertbus.NewBusiness(db.Log, alertsStore)

	registry := workflow.NewActionRegistry()
	registry.Register(communication.NewSendNotificationHandler(db.Log, nil))
	registry.Register(communication.NewCreateAlertHandler(db.Log, alertsBus, nil))

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

	adminID, rule, action := seedCommsActionRule(t, ctx, workflowBus, "send_notification",
		`{"recipients":{"users":["00000000-0000-0000-0000-000000000001"],"roles":[]},"title":"Test","message":"Hello"}`)

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
	t.Log("SUCCESS: send_notification with nil queue client completed without error")
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
	actionType string,
	configJSON string,
) (uuid.UUID, workflow.AutomationRule, workflow.RuleAction) {
	t.Helper()

	adminID := uuid.New()

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
