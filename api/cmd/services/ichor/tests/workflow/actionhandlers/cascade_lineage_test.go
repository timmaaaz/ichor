package actionhandlers_test

// PG bridge — real-Temporal cascade lineage round-trip.
//
// The stub guard tests (business/sdk/workflow/temporal/lineage_test.go) drive the real
// WorkflowTrigger but fake all leaf I/O (recordingStarter never starts a workflow) and only
// MODEL the activity step via nextHopCtx. The one seam they never execute is the lineage
// actually riding through a REAL Temporal activity:
//
//	lineage on TriggerData -> Temporal serialize -> real MergedContext.Flattened ->
//	real activity stamping (activities.go: ctx = contextWithLineage(ctx, lineageFromContextMap(input.Context))) ->
//	real handler ctx -> real bus write -> real delegate.Call -> next DelegateHandler.handleEvent -> next OnEntityEvent
//
// This test exercises exactly that, with cascades still OFF: it builds a one-hop A->B cascade
// from an ALREADY-FIRING Category-B handler (approve_transfer_order — which fires
// transfer_orders.updated today, independent of the M1/M2 gate). It does NOT touch M1/M2
// synthesis and cannot form a loop (the sink writes nothing). The decisive loop-STOPPED test
// (A->B->A across SYNTHESIZED events) still requires P4 and remains PL's job.
//
// Assertion: rule B's automation_executions row carries the visited-set lineage with BOTH the
// chain-root pair (ruleA, customer) and the cascaded pair (ruleB, transfer_order) — proving the
// carrier survived a real activity + real delegate propagation.

import (
	"context"
	"encoding/json"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	workflowtemporal "github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal/stores/edgedb"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/control"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory"
	foundationtemporal "github.com/timmaaaz/ichor/foundation/temporal"
	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func TestCascade_LineageThreadsThroughRealTemporalActivity(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_CascadeLineageThreads")
	ctx := context.Background()

	base := seedConsistencyBase(t, ctx, db)
	uid := base.userID

	// Seed a pending transfer order — the entity rule A approves (firing transfer_orders.updated).
	to, err := db.BusDomain.TransferOrder.Create(ctx, transferorderbus.NewTransferOrder{
		ProductID: base.productIDs[0], FromLocationID: base.loc0, ToLocationID: base.loc1,
		RequestedByID: uid, Quantity: 5, Status: transferorderbus.StatusPending, TransferDate: time.Now(),
	})
	if err != nil {
		t.Fatalf("seeding transfer order: %v", err)
	}

	// ── Real Temporal worker (manual — Category-B handlers need bus deps InitWorkflowInfra lacks) ──
	container := foundationtemporal.GetTestContainer(t)
	tc, err := temporalclient.Dial(temporalclient.Options{HostPort: container.HostPort})
	if err != nil {
		t.Fatalf("connecting to Temporal: %v", err)
	}

	taskQueue := "test-cascade-lineage-" + t.Name()
	w := worker.New(tc, taskQueue, worker.Options{})
	w.RegisterWorkflow(workflowtemporal.ExecuteGraphWorkflow)
	w.RegisterWorkflow(workflowtemporal.ExecuteBranchUntilConvergence)

	registry := workflow.NewActionRegistry()
	// Rule A's action: approve_transfer_order, bound to the SAME bus that is wired to
	// db.BusDomain.Delegate, so its Approve fires a delegate the DelegateHandler will receive.
	registry.Register(inventory.NewApproveTransferOrderHandler(db.Log, db.BusDomain.TransferOrder))
	// Rule B's action: a benign sink that writes nothing (so the chain terminates, no loop).
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

	// ── Workflow infra + the DelegateHandler that turns rule A's bus write into rule B's trigger ──
	workflowStore := workflowdb.NewStore(db.Log, db.DB)
	workflowBus := workflow.NewBusiness(db.Log, db.BusDomain.Delegate, workflowStore)
	edgeStore := edgedb.NewStore(db.Log, db.DB)
	triggerProcessor := workflow.NewTriggerProcessor(db.Log, db.DB, workflowBus)
	if err := triggerProcessor.Initialize(ctx); err != nil {
		t.Fatalf("initializing trigger processor: %v", err)
	}
	workflowTrigger := workflowtemporal.NewWorkflowTrigger(db.Log, tc, triggerProcessor, edgeStore, workflowStore).
		WithTaskQueue(taskQueue)

	// This is the wiring that makes a REAL cascade fire (mirrors all.go:664): without it, rule A's
	// approve fires a delegate that nothing listens to and there is no second hop.
	delegateHandler := workflowtemporal.NewDelegateHandler(db.Log, workflowTrigger)
	delegateHandler.RegisterDomain(db.BusDomain.Delegate, transferorderbus.DomainName, transferorderbus.EntityName)

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

	// seedRule creates an active rule with one action + start edge, returning the rule. The
	// action's type is resolved by the edge store via a LEFT JOIN on action_templates, so the
	// action MUST be linked to a template carrying actionType (a bare action has empty type and
	// the worker cannot resolve a handler).
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

	ruleA := seedRule("cascade-A-approve-TO", customersEntity.ID, onCreate.ID, "approve_transfer_order",
		map[string]any{"transfer_order_id": to.TransferID.String(), "approval_reason": "cascade lineage test"})
	ruleB := seedRule("cascade-B-sink", transferEntity.ID, onUpdate.ID, "evaluate_condition",
		map[string]any{"conditions": []map[string]any{{"field_name": "status", "operator": "equals", "value": "approved"}}})

	if err := triggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing rules: %v", err)
	}

	// ── Fire the initial (human) event: customer creation is the chain root (ruleA, customerID) ──
	customerID := uuid.New()
	if err := workflowTrigger.OnEntityEvent(ctx, workflow.TriggerEvent{
		EventType: "on_create", EntityName: "customers", EntityID: customerID,
		Timestamp: time.Now(), RawData: map[string]any{"test": true}, UserID: uid,
	}); err != nil {
		t.Fatalf("firing initial event: %v", err)
	}

	// ── Poll rule B's execution history; the cascade must thread the lineage onto its record ──
	// Expected visited-set on rule B's execution: {(ruleA, customer), (ruleB, transfer_order)}.
	wantRoot := ruleA.ID.String() + ":" + customerID.String()
	wantCascaded := ruleB.ID.String() + ":" + to.TransferID.String()

	// Read trigger_data directly: workflowBus.QueryExecutionHistory cannot scan a freshly-pending
	// execution (its actions_executed column is NULL → "unsupported Scan into *json.RawMessage",
	// a pre-existing workflowdb limitation, orthogonal to the guard). A scoped raw SELECT of just
	// trigger_data sidesteps it.
	const trigDataQ = `SELECT trigger_data FROM workflow.automation_executions WHERE automation_rules_id = $1`

	var gotVisited []string
	for i := 0; i < 40; i++ { // up to 20s — real worker run + delegate goroutine + second dispatch
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
			// Key off the exported temporal.CascadeLineageKey (not a hardcoded literal) so this
			// test breaks at compile time if the carrier key is ever renamed.
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
		t.Fatalf("timeout: rule B never received a cascaded execution carrying lineage {%s, %s} "+
			"(key %q) — the visited-set did not survive the real Temporal activity / delegate round-trip",
			wantRoot, wantCascaded, workflowtemporal.CascadeLineageKey)
	}
	t.Logf("cascade lineage threaded through real Temporal: visited=%v", gotVisited)
}
