package actionhandlers_test

// PL — THE DECISIVE TEST. A real A→B→A loop across SYNTHESIZED (M1) update_field events,
// run against a live Temporal worker + real DB, is actually STOPPED by the runtime
// visited-set — the loop that could not fire before P4 (M1 synthesis was off, so nothing
// could loop). Plus A→B→C→done progresses (distinct (rule,entity) pairs never false-stop).
//
// Construction (clean products.product_categories target):
//
//	Rule A: on_update product_categories, gate description=="A" → update_field e2.description="B"
//	Rule B: on_update product_categories, gate description=="B" → update_field e1.description="A"
//	kick e1→"A": A fires → writes e2 → synth(e2,"B") → B fires → writes e1 → synth(e1,"A")
//	             → A matches e1 again → (ruleA,e1) ALREADY visited → REFUSED (no execution)
//
// `equals` gates always re-arm (no latch) and the values alternate, so absent the guard
// this loops forever. The explicit id= condition makes update_field synthesize the correct
// written-row EntityID (updatefield.go:226), so the guard keys on (rule, real-row) exactly.
// Each event matches exactly one rule (gate on the written description) — no self-fan-out.

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/data"

	"github.com/timmaaaz/ichor/business/sdk/workflowdomains"
)

const pcEntityTable = "products.product_categories"

func TestCascade_LoopGuard(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_CascadeLoopGuard")
	ctx := context.Background()
	base := seedConsistencyBase(t, ctx, db)
	uid := base.userID

	// The only handler the rules use: update_field, wired to synthesize events AND emit
	// them to the outbox (data.WithOutbox) so the rig's relay dispatches each cascade hop.
	registry := workflow.NewActionRegistry()
	registry.Register(data.NewUpdateFieldHandler(db.Log, db.DB,
		data.WithDelegate(db.BusDomain.Delegate),
		data.WithEntityRegistry(workflowdomains.ReverseMap()),
		data.WithOutbox(db.BusDomain.OutboxWriter)))
	rig := startCascadeRig(t, ctx, db, registry)

	entityType, err := rig.workflowBus.QueryEntityTypeByName(ctx, "table")
	if err != nil {
		t.Fatalf("querying entity type: %v", err)
	}
	onUpdate, err := rig.workflowBus.QueryTriggerTypeByName(ctx, "on_update")
	if err != nil {
		t.Fatalf("querying on_update trigger type: %v", err)
	}
	pcEntity, err := rig.workflowBus.QueryEntityByName(ctx, "product_categories")
	if err != nil {
		t.Fatalf("querying product_categories entity: %v", err)
	}

	// ── The decisive A→B→A-stopped test ──────────────────────────────────────────────
	t.Run("loop_stopped_across_synthesized_events", func(t *testing.T) {
		const valA, valB = "LOOP_STATE_A", "LOOP_STATE_B"

		e1 := mustSeedCategory(t, ctx, db, "seed-e1")
		e2 := mustSeedCategory(t, ctx, db, "seed-e2")

		ruleA := seedActiveRule(t, ctx, rig.workflowBus, uid, "loop-A",
			pcEntity.ID, entityType.ID, onUpdate.ID,
			fieldConditions(cond("description", "equals", valA)),
			"update_field", updateFieldByID(pcEntityTable, "description", valB, e2))
		ruleB := seedActiveRule(t, ctx, rig.workflowBus, uid, "loop-B",
			pcEntity.ID, entityType.ID, onUpdate.ID,
			fieldConditions(cond("description", "equals", valB)),
			"update_field", updateFieldByID(pcEntityTable, "description", valA, e1))

		if err := rig.triggerProcessor.RefreshRules(ctx); err != nil {
			t.Fatalf("refreshing rules: %v", err)
		}

		// Kick: e1 enters state A (matches rule A). This is the chain root (ruleA, e1).
		if err := rig.workflowTrigger.OnEntityEvent(ctx, workflow.TriggerEvent{
			EventType: "on_update", EntityName: "product_categories", EntityID: e1,
			Timestamp: time.Now(), RawData: map[string]any{"description": valA}, UserID: uid,
		}); err != nil {
			t.Fatalf("firing kick event: %v", err)
		}

		// Hop 2 reached: rule B dispatched with BOTH pairs on its lineage.
		wantRoot := ruleA.ID.String() + ":" + e1.String()
		wantCascaded := ruleB.ID.String() + ":" + e2.String()
		if !eventually(20*time.Second, 400*time.Millisecond, func() bool {
			for _, vs := range executionVisitedSets(t, ctx, db, ruleB.ID) {
				if slices.Contains(vs, wantRoot) && slices.Contains(vs, wantCascaded) {
					return true
				}
			}
			return false
		}) {
			t.Fatalf("rule B never cascaded with lineage {%s, %s} — A→B hop did not happen", wantRoot, wantCascaded)
		}

		// Hop 2's action ran: e1 written back to state A by rule B. Once observed, rule B's
		// synthesized event has fired and the third hop (A on e1) has been attempted.
		if !eventually(20*time.Second, 400*time.Millisecond, func() bool {
			return categoryDescription(t, ctx, db, e1) == valA
		}) {
			t.Fatalf("rule B never wrote e1 back to %q — hop 2 action did not run", valA)
		}

		// Settle: let the (async) refused re-entry resolve, then prove no third hop ran.
		// Sized to comfortably outlast one would-be re-entry dispatch (OnEntityEvent +
		// startWorkflowForRule + execution-row insert) — so a DISABLED guard, which loops
		// unbounded, deterministically produces ruleA count>=2 here rather than a spurious 1
		// on a slow CI box.
		time.Sleep(5 * time.Second)

		// ── DECISIVE: the loop is stopped. Rule A executed exactly once (the kick); the
		// third hop's (ruleA, e1) re-entry is refused by the visited-set, creating NO
		// execution row. A broken guard loops unbounded → rule A would have many rows.
		if n := executionCount(t, ctx, db, ruleA.ID); n != 1 {
			t.Fatalf("LOOP GUARD FAILED: rule A executed %d times, want exactly 1 (the A→B→A re-entry must be refused)", n)
		}
		if n := executionCount(t, ctx, db, ruleB.ID); n != 1 {
			t.Fatalf("rule B executed %d times, want exactly 1", n)
		}

		// Final DB state is bounded, not flip-flopping: e1=A (last written by B), e2=B (by A).
		if got := categoryDescription(t, ctx, db, e1); got != valA {
			t.Errorf("e1 description = %q, want %q", got, valA)
		}
		if got := categoryDescription(t, ctx, db, e2); got != valB {
			t.Errorf("e2 description = %q, want %q", got, valB)
		}
		t.Logf("A→B→A stopped: ruleA execs=1, ruleB execs=1, final e1=%q e2=%q", valA, valB)
	})

	// ── A→B→C→done progresses (distinct pairs are never false-stopped) ────────────────
	t.Run("progression_does_not_false_stop", func(t *testing.T) {
		const valP, valQ, valR, valDone = "PROG_P", "PROG_Q", "PROG_R", "PROG_DONE"

		e1 := mustSeedCategory(t, ctx, db, "seed-p1")
		e2 := mustSeedCategory(t, ctx, db, "seed-p2")
		e3 := mustSeedCategory(t, ctx, db, "seed-p3")

		ruleA := seedActiveRule(t, ctx, rig.workflowBus, uid, "prog-A",
			pcEntity.ID, entityType.ID, onUpdate.ID,
			fieldConditions(cond("description", "equals", valP)),
			"update_field", updateFieldByID(pcEntityTable, "description", valQ, e2))
		ruleB := seedActiveRule(t, ctx, rig.workflowBus, uid, "prog-B",
			pcEntity.ID, entityType.ID, onUpdate.ID,
			fieldConditions(cond("description", "equals", valQ)),
			"update_field", updateFieldByID(pcEntityTable, "description", valR, e3))
		ruleC := seedActiveRule(t, ctx, rig.workflowBus, uid, "prog-C",
			pcEntity.ID, entityType.ID, onUpdate.ID,
			fieldConditions(cond("description", "equals", valR)),
			// C writes a terminal value back to e1 that matches NO rule's gate → chain ends.
			"update_field", updateFieldByID(pcEntityTable, "description", valDone, e1))

		if err := rig.triggerProcessor.RefreshRules(ctx); err != nil {
			t.Fatalf("refreshing rules: %v", err)
		}

		if err := rig.workflowTrigger.OnEntityEvent(ctx, workflow.TriggerEvent{
			EventType: "on_update", EntityName: "product_categories", EntityID: e1,
			Timestamp: time.Now(), RawData: map[string]any{"description": valP}, UserID: uid,
		}); err != nil {
			t.Fatalf("firing kick event: %v", err)
		}

		// The chain must REACH rule C (3 hops) carrying all three distinct pairs — proving
		// the visited-set never wrongly blocks a forward progression.
		wantC := []string{
			ruleA.ID.String() + ":" + e1.String(),
			ruleB.ID.String() + ":" + e2.String(),
			ruleC.ID.String() + ":" + e3.String(),
		}
		if !eventually(20*time.Second, 400*time.Millisecond, func() bool {
			for _, vs := range executionVisitedSets(t, ctx, db, ruleC.ID) {
				if slices.Contains(vs, wantC[0]) && slices.Contains(vs, wantC[1]) && slices.Contains(vs, wantC[2]) {
					return true
				}
			}
			return false
		}) {
			t.Fatalf("rule C never reached with full lineage %v — progression was wrongly stopped", wantC)
		}

		// Terminal: C wrote e1 to a no-match value, so each rule ran exactly once.
		if !eventually(20*time.Second, 400*time.Millisecond, func() bool {
			return categoryDescription(t, ctx, db, e1) == valDone
		}) {
			t.Fatalf("rule C never wrote e1 to terminal %q", valDone)
		}
		time.Sleep(3 * time.Second)
		for name, id := range map[string]workflow.AutomationRule{"A": ruleA, "B": ruleB, "C": ruleC} {
			if n := executionCount(t, ctx, db, id.ID); n != 1 {
				t.Errorf("prog rule %s executed %d times, want exactly 1", name, n)
			}
		}
		t.Logf("A→B→C→done progressed: all three rules executed once, chain terminated")
	})
}
