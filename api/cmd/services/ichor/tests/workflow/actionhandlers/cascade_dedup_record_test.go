package actionhandlers_test

// Fast-follow #3 — execution-record dedup (run-level → record-level). The I4 end-to-end the
// F9 spec wanted, now backed by a real fix.
//
// The trigger writes the workflow.automation_executions record (CreateExecution) BEFORE it
// starts the Temporal workflow (ExecuteWorkflow), and the workflow id is keyed on the durable
// outbox EventID with WorkflowIDReusePolicy=REJECT_DUPLICATE. So a relay re-delivery of the
// SAME outbox row gives effectively-once *workflow execution* (Temporal rejects the dup), but —
// before the fix — the trigger has already written a SECOND StatusPending *record*. Dedup is
// run-level, not record-level.
//
// We simulate a relay re-delivery by feeding the SAME EventID through WorkflowTrigger.OnEntityEvent
// twice (the rig's relay deletes-on-publish, so re-driving it deterministically is awkward; the
// trigger entry point is exactly where the bug lives, and ExecuteWorkflow blocks until the server
// accepts the start, so the 2nd call deterministically hits REJECT_DUPLICATE — no sleeps needed).
//
//	RED  (before fix): two StatusPending rows for the rule.
//	GREEN (after fix): the rejected 2nd dispatch's orphan row is deleted → exactly one row.

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/data"
)

func TestCascade_DedupExecutionRecord(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_CascadeDedupRecord")
	ctx := context.Background()
	base := seedConsistencyBase(t, ctx, db)
	uid := base.userID

	// One harmless handler: update_field with NO outbox/delegate, so the dispatched workflow
	// performs a benign write and does NOT cascade onward (this test is about the record table,
	// not the cascade chain).
	registry := workflow.NewActionRegistry()
	registry.Register(data.NewUpdateFieldHandler(db.Log, db.DB))
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

	// e1 is the event's entity; e2 is the (benign) write target so the action never touches e1.
	e1 := mustSeedCategory(t, ctx, db, "dedup-e1")
	e2 := mustSeedCategory(t, ctx, db, "dedup-e2")

	const marker = "DEDUP_MARKER"
	rule := seedActiveRule(t, ctx, rig.workflowBus, uid, "dedup-record-rule",
		pcEntity.ID, entityType.ID, onUpdate.ID,
		fieldConditions(cond("description", "equals", marker)),
		"update_field", updateFieldByID(pcEntityTable, "description", "DEDUP_ACTION_RAN", e2))
	if err := rig.triggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing rules: %v", err)
	}

	// A single durable EventID, re-delivered: the same id both dispatches ⇒ the same
	// workflow-{ruleID}-{eventID} ⇒ REJECT_DUPLICATE on the 2nd start.
	eventID := uuid.New()
	event := workflow.TriggerEvent{
		EventType:  "on_update",
		EntityName: "product_categories",
		EntityID:   e1,
		EventID:    eventID,
		Timestamp:  time.Now(),
		RawData:    map[string]any{"description": marker},
		UserID:     uid,
	}

	// First delivery: creates the record and starts the workflow.
	if err := rig.workflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("first OnEntityEvent: %v", err)
	}
	// Re-delivery of the same outbox row: REJECT_DUPLICATE rejects the duplicate workflow.
	// OnEntityEvent is fail-open per rule, so it returns nil even though the start was rejected.
	if err := rig.workflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("second OnEntityEvent (re-delivery): %v", err)
	}

	// CreateExecution is synchronous and precedes ExecuteWorkflow, so the record count is
	// settled the moment both dispatches return — no need to wait on workflow completion.
	if n := executionCount(t, ctx, db, rule.ID); n != 1 {
		t.Fatalf("execution-record dedup FAILED: rule has %d automation_executions rows, want exactly 1 "+
			"(a re-delivered EventID is REJECT_DUPLICATE'd at the workflow layer; the trigger must delete "+
			"the orphaned StatusPending record so dedup is record-level, not just run-level)", n)
	}
	t.Logf("record-level dedup proven: same EventID re-delivered twice → exactly 1 automation_executions row")
}
