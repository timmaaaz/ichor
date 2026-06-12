package actionhandlers_test

// PL (Live-System Verification) shared rig.
//
// These helpers stand up the REAL cascade pipeline — a live Temporal worker bound to a
// caller-supplied ActionRegistry, plus the WorkflowTrigger + DelegateHandler wiring that
// turns a bus (or synthesized M1/M2) delegate event into a downstream rule dispatch. It
// mirrors the production wiring (all.go:574-681) and the one-hop rig in
// cascade_lineage_test.go, generalized so the PL suites (loop guard, M2 live, fan-out /
// idempotency) can each build a registry, seed rules, fire an event, and poll
// automation_executions for the cascade outcome + visited-set lineage.
//
// Rules are created via the workflow BUS (CreateRule/CreateRuleAction), which deliberately
// BYPASSES the P2 static cascade detector (that enforces only at the app layer —
// workflowsaveapp / ruleapp). This is intentional: PL exercises the RUNTIME visited-set
// guard in isolation, so the tests must be able to construct a real provable loop that the
// static layer would otherwise refuse.

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/api/cmd/services/ichor/build/all/workflowdomains"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	workflowtemporal "github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal/stores/edgedb"
	foundationtemporal "github.com/timmaaaz/ichor/foundation/temporal"
	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// cascadeRig holds the live worker-backed cascade pipeline for one test.
type cascadeRig struct {
	workflowBus      *workflow.Business
	triggerProcessor *workflow.TriggerProcessor
	workflowTrigger  *workflowtemporal.WorkflowTrigger
	delegateHandler  *workflowtemporal.DelegateHandler
	taskQueue        string
}

// startCascadeRig boots a real Temporal worker bound to registry, wires the
// WorkflowTrigger + DelegateHandler, and registers the FULL production domain set
// (workflowdomains.Registrations()) so any bus or synthesized event cascades — exactly
// as all.go does. Cleanup (worker stop + client close) is registered on t.
func startCascadeRig(t *testing.T, ctx context.Context, db *dbtest.Database, registry *workflow.ActionRegistry) *cascadeRig {
	t.Helper()

	container := foundationtemporal.GetTestContainer(t)
	tc, err := temporalclient.Dial(temporalclient.Options{HostPort: container.HostPort})
	if err != nil {
		t.Fatalf("connecting to Temporal: %v", err)
	}

	taskQueue := "test-cascade-pl-" + t.Name()
	w := worker.New(tc, taskQueue, worker.Options{})
	w.RegisterWorkflow(workflowtemporal.ExecuteGraphWorkflow)
	w.RegisterWorkflow(workflowtemporal.ExecuteBranchUntilConvergence)
	w.RegisterActivity(&workflowtemporal.Activities{
		Registry:      registry,
		AsyncRegistry: workflowtemporal.NewAsyncRegistry(),
	})
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

	delegateHandler := workflowtemporal.NewDelegateHandler(db.Log, workflowTrigger)
	for _, reg := range workflowdomains.Registrations() {
		delegateHandler.RegisterDomain(db.BusDomain.Delegate, reg.Domain, reg.Entity)
	}

	return &cascadeRig{
		workflowBus:      workflowBus,
		triggerProcessor: triggerProcessor,
		workflowTrigger:  workflowTrigger,
		delegateHandler:  delegateHandler,
		taskQueue:        taskQueue,
	}
}

// seedActiveRule creates an active rule with the given trigger conditions, one action
// (resolved to actionType via an action_template), and a start edge. Generalizes the
// local seedRule in cascade_lineage_test.go by accepting trigger conditions (nil = auto-
// match). triggerConditions, when non-nil, is marshaled as the rule's TriggerConditions
// JSON (use fieldConditions(...) to build the standard {"field_conditions":[...]} shape).
func seedActiveRule(t *testing.T, ctx context.Context, wb *workflow.Business, uid uuid.UUID,
	name string, entityID, entityTypeID, triggerTypeID uuid.UUID, triggerConditions any,
	actionType string, cfg map[string]any) workflow.AutomationRule {
	t.Helper()

	newRule := workflow.NewAutomationRule{
		Name: name, Description: name, EntityID: entityID, EntityTypeID: entityTypeID,
		TriggerTypeID: triggerTypeID, IsActive: true, CreatedBy: uid,
	}
	if triggerConditions != nil {
		b, err := json.Marshal(triggerConditions)
		if err != nil {
			t.Fatalf("marshal trigger conditions for %q: %v", name, err)
		}
		raw := json.RawMessage(b)
		newRule.TriggerConditions = &raw
	}
	rule, err := wb.CreateRule(ctx, newRule)
	if err != nil {
		t.Fatalf("creating rule %q: %v", name, err)
	}

	cfgBytes, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal action config for %q: %v", name, err)
	}
	tmpl, err := wb.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name: name + " template", Description: name, ActionType: actionType,
		DefaultConfig: json.RawMessage(cfgBytes), CreatedBy: uid,
	})
	if err != nil {
		t.Fatalf("creating template for %q: %v", name, err)
	}
	action, err := wb.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID, Name: name + " action",
		ActionConfig: json.RawMessage(cfgBytes), IsActive: true, TemplateID: &tmpl.ID,
	})
	if err != nil {
		t.Fatalf("creating action for %q: %v", name, err)
	}
	if _, err := wb.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID: rule.ID, SourceActionID: nil, TargetActionID: action.ID, EdgeType: "start", EdgeOrder: 0,
	}); err != nil {
		t.Fatalf("creating start edge for %q: %v", name, err)
	}
	return rule
}

// fieldConditions builds the standard rule TriggerConditions JSON shape:
// {"field_conditions":[{"field_name","operator","value"}, ...]}.
func fieldConditions(conds ...map[string]any) map[string]any {
	return map[string]any{"field_conditions": conds}
}

// cond is a single field condition for fieldConditions / action conditions.
func cond(field, operator string, value any) map[string]any {
	return map[string]any{"field_name": field, "operator": operator, "value": value}
}

// updateFieldByID builds an update_field action config that sets target.field on the row
// selected by an explicit id= condition (so the synthesized event carries that row's id
// as the EntityID — the loop guard keys on it).
func updateFieldByID(targetEntity, targetField string, newValue any, rowID uuid.UUID) map[string]any {
	return map[string]any{
		"target_entity": targetEntity,
		"target_field":  targetField,
		"new_value":     newValue,
		"conditions":    []map[string]any{cond("id", "equals", rowID.String())},
	}
}

// mustSeedCategory inserts a products.product_categories row (a clean generic-write target:
// no FKs, no typed action, no protected fields) and returns its id.
func mustSeedCategory(t *testing.T, ctx context.Context, db *dbtest.Database, initialDesc string) uuid.UUID {
	t.Helper()
	pc, err := db.BusDomain.ProductCategory.Create(ctx, productcategorybus.NewProductCategory{
		Name:        "pl-cat-" + uuid.NewString()[:8],
		Description: initialDesc,
	})
	if err != nil {
		t.Fatalf("seeding product category: %v", err)
	}
	return pc.ProductCategoryID
}

// categoryDescription reads the current description of a product_categories row.
func categoryDescription(t *testing.T, ctx context.Context, db *dbtest.Database, id uuid.UUID) string {
	t.Helper()
	var desc string
	err := db.DB.QueryRowContext(ctx,
		`SELECT description FROM products.product_categories WHERE id = $1`, id).Scan(&desc)
	if err != nil {
		t.Fatalf("reading product category %s description: %v", id, err)
	}
	return desc
}

// executionVisitedSets returns the cascade visited-set carried on every
// automation_executions row for the rule (one slice per execution row). It reads
// trigger_data via a scoped raw SELECT because workflowBus.QueryExecutionHistory cannot
// scan a freshly-pending execution (NULL actions_executed → "unsupported Scan into
// *json.RawMessage", a pre-existing workflowdb limitation orthogonal to the guard).
// An execution with no lineage key yields a nil entry (still counted).
func executionVisitedSets(t *testing.T, ctx context.Context, db *dbtest.Database, ruleID uuid.UUID) [][]string {
	t.Helper()
	const q = `SELECT trigger_data FROM workflow.automation_executions WHERE automation_rules_id = $1`
	rows, err := db.DB.QueryContext(ctx, q, ruleID)
	if err != nil {
		t.Fatalf("querying executions for rule %s: %v", ruleID, err)
	}
	defer rows.Close()

	var out [][]string
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			t.Fatalf("scanning trigger_data: %v", err)
		}
		var td map[string]json.RawMessage
		if err := json.Unmarshal(raw, &td); err != nil {
			out = append(out, nil)
			continue
		}
		lineageRaw, ok := td[workflowtemporal.CascadeLineageKey]
		if !ok {
			out = append(out, nil)
			continue
		}
		var lineage struct {
			Visited []string `json:"visited"`
		}
		if err := json.Unmarshal(lineageRaw, &lineage); err != nil {
			out = append(out, nil)
			continue
		}
		out = append(out, lineage.Visited)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterating executions for rule %s: %v", ruleID, err)
	}
	return out
}

// executionCount returns how many times a rule has been dispatched (one
// automation_executions row per dispatch). A re-entry refused by the visited-set creates
// NO row, so this is the load-bearing signal for "the loop was stopped".
func executionCount(t *testing.T, ctx context.Context, db *dbtest.Database, ruleID uuid.UUID) int {
	t.Helper()
	return len(executionVisitedSets(t, ctx, db, ruleID))
}

// eventually polls cond until it returns true or timeout elapses; returns the final result.
func eventually(timeout, interval time.Duration, cond func() bool) bool {
	deadline := time.Now().Add(timeout)
	for {
		if cond() {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(interval)
	}
}
