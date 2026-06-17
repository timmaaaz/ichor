package actionhandlers_test

// F6 reliability verification — cascade lineage round-trip through the TRANSACTIONAL
// OUTBOX + RELAY (the post-F5 production cascade path).
//
// The sibling cascade_lineage_test.go proves the loop-guard lineage survives a real
// Temporal activity when cascades fire through the DELEGATE (it wires a DelegateHandler
// + RegisterDomain). F2 retired that path: a cascade-relevant bus write now persists an
// outbox row in its transaction, and a polling relay — not the delegate — is the SOLE
// dispatcher. This test exercises that new transport end to end:
//
//	human customers.on_create -> rule A (approve_transfer_order) runs in a REAL Temporal
//	activity -> transferorderbus.Approve writes the entity AND b.outbox.Emit persists a
//	transfer_orders.updated row carrying the serialized lineage -> the RELAY drains that
//	row (decode lineage -> contextWithLineage -> OnEntityEvent) -> rule B fires.
//
// It deliberately wires NO DelegateHandler, so the only way rule B can fire is the relay.
// That makes the test a no-double-dispatch trip-wire by construction: a cascade through a
// stray delegate subscriber would be a second, independent dispatch this wiring cannot
// produce.
//
// Assertions:
//  1. rule B fires at all  -> the relay is a sufficient, sole dispatcher (e2e Path).
//  2. rule B's execution carries the visited-set with BOTH the chain-root pair
//     (ruleA, customer) and the cascaded pair (ruleB, transfer_order) -> the loop-guard
//     lineage SURVIVED serialize-into-row + rehydrate-before-dispatch (the F2-specific
//     regression; the guard's refusal on that visited-set is unit-tested in
//     temporal/lineage_test.go and is unchanged by F2).
//  3. the cascade outbox row is gone afterward -> delete-on-publish.

import (
	"context"
	"encoding/json"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus/stores/transferorderdb"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/outbox"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	workflowtemporal "github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal/stores/edgedb"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/control"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory"
	"github.com/timmaaaz/ichor/business/sdk/workflowdomains"
	foundationtemporal "github.com/timmaaaz/ichor/foundation/temporal"
	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func TestCascade_OutboxRoundTrip_LineageSurvivesViaRelay(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_CascadeOutboxRoundTrip")
	ctx := context.Background()

	base := seedConsistencyBase(t, ctx, db)
	uid := base.userID

	// Seed a pending transfer order — the entity rule A approves (Approve fires
	// transfer_orders.updated and, on the outbox-wired bus below, emits the cascade row).
	to, err := db.BusDomain.TransferOrder.Create(ctx, transferorderbus.NewTransferOrder{
		ProductID: base.productIDs[0], FromLocationID: base.loc0, ToLocationID: base.loc1,
		RequestedByID: uid, Quantity: 5, Status: transferorderbus.StatusPending, TransferDate: time.Now(),
	})
	if err != nil {
		t.Fatalf("seeding transfer order: %v", err)
	}

	// ── Real Temporal worker (manual — Category-B handlers need bus deps) ──
	container := foundationtemporal.GetTestContainer(t)
	tc, err := temporalclient.Dial(temporalclient.Options{HostPort: container.HostPort})
	if err != nil {
		t.Fatalf("connecting to Temporal: %v", err)
	}

	taskQueue := "test-cascade-outbox-" + t.Name()
	w := worker.New(tc, taskQueue, worker.Options{})
	w.RegisterWorkflow(workflowtemporal.ExecuteGraphWorkflow)
	w.RegisterWorkflow(workflowtemporal.ExecuteBranchUntilConvergence)

	// ── The outbox Writer, built exactly as the composition root does (DESIGN §3 / F5):
	// entityForDomain from workflowdomains.Registrations(); the lineage extractor is the
	// exported temporal helper the Writer needs (outbox cannot import temporal). ──
	entityForDomain := make(map[string]string)
	for _, r := range workflowdomains.Registrations() {
		entityForDomain[r.Domain] = r.Entity
	}
	outboxWriter := outbox.NewWriter(db.Log, db.DB, entityForDomain, workflowtemporal.MarshalLineageFromContext)

	registry := workflow.NewActionRegistry()
	// Rule A's action runs against an OUTBOX-WIRED transfer order bus, so its Approve both
	// writes the entity and emits the transfer_orders.updated cascade row (the only thing
	// that can drive hop 2 — there is no DelegateHandler here).
	toBus := transferorderbus.NewBusiness(db.Log, db.BusDomain.Delegate, transferorderdb.NewStore(db.Log, db.DB)).
		WithOutbox(outboxWriter)
	registry.Register(inventory.NewApproveTransferOrderHandler(db.Log, toBus))
	// Rule B's action: a benign sink that writes nothing (the chain terminates cleanly).
	registry.Register(control.NewEvaluateConditionHandler(db.Log))
	w.RegisterActivity(&workflowtemporal.Activities{
		Registry:      registry,
		AsyncRegistry: workflowtemporal.NewAsyncRegistry(),
	})
	if err := w.Start(); err != nil {
		tc.Close()
		t.Fatalf("starting Temporal worker: %v", err)
	}
	t.Cleanup(func() { w.Stop(); tc.Close() })

	// ── Workflow infra. NOTE: NO DelegateHandler / RegisterDomain. The relay below is the
	// SOLE cascade dispatcher — exactly the post-F5 production shape. ──
	workflowStore := workflowdb.NewStore(db.Log, db.DB)
	workflowBus := workflow.NewBusiness(db.Log, db.BusDomain.Delegate, workflowStore)
	edgeStore := edgedb.NewStore(db.Log, db.DB)
	triggerProcessor := workflow.NewTriggerProcessor(db.Log, db.DB, workflowBus)
	if err := triggerProcessor.Initialize(ctx); err != nil {
		t.Fatalf("initializing trigger processor: %v", err)
	}
	workflowTrigger := workflowtemporal.NewWorkflowTrigger(db.Log, tc, triggerProcessor, edgeStore, workflowStore).
		WithTaskQueue(taskQueue)

	// ── Start the relay — the production cascade dispatcher (server-only in prod; here it
	// drains this test's isolated outbox). A short poll keeps the test snappy. ──
	relay := workflowtemporal.NewRelay(db.Log, db.DB, workflowTrigger, workflowtemporal.RelayConfig{
		PollInterval: 200 * time.Millisecond,
	})
	relayCtx, cancelRelay := context.WithCancel(ctx)
	t.Cleanup(cancelRelay)
	go func() { _ = relay.Run(relayCtx) }()

	// ── Build rule A (customers.on_create -> approve TO) and rule B (transfer_orders.on_update -> sink) ──
	entityType, err := workflowBus.QueryEntityTypeByName(ctx, "table")
	if err != nil {
		t.Fatalf("querying entity type: %v", err)
	}
	onCreate, err := workflowBus.QueryTriggerTypeByName(ctx, "on_create")
	if err != nil {
		t.Fatalf("querying on_create: %v", err)
	}
	onUpdate, err := workflowBus.QueryTriggerTypeByName(ctx, "on_update")
	if err != nil {
		t.Fatalf("querying on_update: %v", err)
	}
	customersEntity, err := workflowBus.QueryEntityByName(ctx, "customers")
	if err != nil {
		t.Fatalf("querying customers entity: %v", err)
	}
	transferEntity, err := workflowBus.QueryEntityByName(ctx, "transfer_orders")
	if err != nil {
		t.Fatalf("querying transfer_orders entity: %v", err)
	}

	seedRule := func(name string, entityID, triggerTypeID uuid.UUID, actionType string, cfg map[string]any) workflow.AutomationRule {
		t.Helper()
		rule, err := workflowBus.CreateRule(ctx, workflow.NewAutomationRule{
			Name: name, Description: name, EntityID: entityID, EntityTypeID: entityType.ID,
			TriggerTypeID: triggerTypeID, IsActive: true, CreatedBy: uid,
		})
		if err != nil {
			t.Fatalf("creating rule %q: %v", name, err)
		}
		cfgBytes, _ := json.Marshal(cfg)
		tmpl, err := workflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
			Name: name + " template", Description: name, ActionType: actionType,
			DefaultConfig: json.RawMessage(cfgBytes), CreatedBy: uid,
		})
		if err != nil {
			t.Fatalf("creating template for %q: %v", name, err)
		}
		action, err := workflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
			AutomationRuleID: rule.ID, Name: name + " action",
			ActionConfig: json.RawMessage(cfgBytes), IsActive: true, TemplateID: &tmpl.ID,
		})
		if err != nil {
			t.Fatalf("creating action for %q: %v", name, err)
		}
		if _, err := workflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
			RuleID: rule.ID, SourceActionID: nil, TargetActionID: action.ID, EdgeType: "start", EdgeOrder: 0,
		}); err != nil {
			t.Fatalf("creating start edge for %q: %v", name, err)
		}
		return rule
	}

	ruleA := seedRule("outbox-A-approve-TO", customersEntity.ID, onCreate.ID, "approve_transfer_order",
		map[string]any{"transfer_order_id": to.TransferID.String(), "approval_reason": "outbox cascade test"})
	ruleB := seedRule("outbox-B-sink", transferEntity.ID, onUpdate.ID, "evaluate_condition",
		map[string]any{"conditions": []map[string]any{{"field_name": "status", "operator": "equals", "value": "approved"}}})

	if err := triggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing rules: %v", err)
	}

	// ── Fire the chain root: a human customer creation (ruleA, customer) ──
	customerID := uuid.New()
	if err := workflowTrigger.OnEntityEvent(ctx, workflow.TriggerEvent{
		EventType: "on_create", EntityName: "customers", EntityID: customerID,
		Timestamp: time.Now(), RawData: map[string]any{"test": true}, UserID: uid,
	}); err != nil {
		t.Fatalf("firing initial event: %v", err)
	}

	// ── Poll rule B's execution history; the cascade must thread the lineage onto its
	// record. Expected visited-set: {(ruleA, customer), (ruleB, transfer_order)}. ──
	wantRoot := ruleA.ID.String() + ":" + customerID.String()
	wantCascaded := ruleB.ID.String() + ":" + to.TransferID.String()

	const trigDataQ = `SELECT trigger_data FROM workflow.automation_executions WHERE automation_rules_id = $1`

	var gotVisited []string
	for i := 0; i < 50; i++ { // up to ~25s — first hop (worker) + relay poll + second dispatch
		rows, err := db.DB.QueryContext(ctx, trigDataQ, ruleB.ID)
		if err != nil {
			t.Fatalf("querying rule B trigger_data: %v", err)
		}
		for rows.Next() {
			var raw []byte
			if err := rows.Scan(&raw); err != nil {
				rows.Close()
				t.Fatalf("scanning trigger_data: %v", err)
			}
			var td map[string]json.RawMessage
			if err := json.Unmarshal(raw, &td); err != nil {
				continue
			}
			lineageRaw, ok := td[workflowtemporal.CascadeLineageKey]
			if !ok {
				continue
			}
			var lineage struct {
				Visited []string `json:"visited"`
			}
			if err := json.Unmarshal(lineageRaw, &lineage); err != nil {
				continue
			}
			if slices.Contains(lineage.Visited, wantRoot) && slices.Contains(lineage.Visited, wantCascaded) {
				gotVisited = lineage.Visited
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			t.Fatalf("iterating rule B trigger_data: %v", err)
		}
		rows.Close()
		if gotVisited != nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if gotVisited == nil {
		t.Fatalf("timeout: rule B never received a relay-dispatched execution carrying lineage {%s, %s} "+
			"(key %q) — the visited-set did not survive the outbox row serialize/rehydrate, or the relay "+
			"never dispatched (no double-dispatch path exists in this wiring)",
			wantRoot, wantCascaded, workflowtemporal.CascadeLineageKey)
	}
	t.Logf("cascade lineage survived the outbox round-trip via the relay: visited=%v", gotVisited)

	// ── delete-on-publish: the transfer_orders.updated cascade row the relay dispatched
	// must be gone. (The seed Create used db.BusDomain.TransferOrder, which has no Writer,
	// so the ONLY emitted row is Approve's — its absence proves delete-on-publish.) ──
	const pendingQ = `SELECT count(*) FROM workflow.cascade_outbox WHERE domain = $1`
	var remaining int
	for i := 0; i < 20; i++ {
		if err := db.DB.QueryRowContext(ctx, pendingQ, transferorderbus.DomainName).Scan(&remaining); err != nil {
			t.Fatalf("counting outbox rows: %v", err)
		}
		if remaining == 0 {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	if remaining != 0 {
		t.Fatalf("expected the dispatched cascade row to be deleted-on-publish, found %d %s rows remaining",
			remaining, transferorderbus.DomainName)
	}
}
