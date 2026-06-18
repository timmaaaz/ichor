package actionhandlers_test

// Fast-follow #5 — execution-record lifecycle. The decisive end-to-end proof that a CASCADED
// (Temporal-dispatched) automation advances its workflow.automation_executions row from
// pending → running → completed, instead of being stranded at "pending" forever.
//
// Before #5 the Temporal cascade path wrote the record at StatusPending (trigger.go
// CreateExecution) and NOTHING advanced it — only the synchronous manual path
// (actionservice.go recordExecution) set running/completed. Yet executionapi exposes status
// to users (and lets them filter on it), so every cascaded run displayed "pending" forever.
// #5 fires MarkExecutionRunning as the first workflow activity and MarkExecutionCompleted/
// Failed at the terminal points (gated behind a Temporal workflow-version bump for replay
// determinism), and wires the execution store into the rig's Activities so the worker can
// perform the write-back.
//
//	RED  (before fix): status stays "pending" forever → never reaches "completed".
//	GREEN (after fix): status reaches "completed" once the dispatched workflow finishes.

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/data"
)

func TestCascade_ExecutionLifecycle_ReachesCompleted(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_CascadeExecutionLifecycle")
	ctx := context.Background()
	base := seedConsistencyBase(t, ctx, db)
	uid := base.userID

	// One benign handler: update_field with NO outbox/delegate, so the dispatched workflow
	// performs a successful write and does NOT cascade onward (this test is about the record's
	// status lifecycle, not the cascade chain).
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
	e1 := mustSeedCategory(t, ctx, db, "lifecycle-e1")
	e2 := mustSeedCategory(t, ctx, db, "lifecycle-e2")

	const marker = "LIFECYCLE_MARKER"
	rule := seedActiveRule(t, ctx, rig.workflowBus, uid, "lifecycle-rule",
		pcEntity.ID, entityType.ID, onUpdate.ID,
		fieldConditions(cond("description", "equals", marker)),
		"update_field", updateFieldByID(pcEntityTable, "description", "LIFECYCLE_ACTION_RAN", e2))
	if err := rig.triggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing rules: %v", err)
	}

	event := workflow.TriggerEvent{
		EventType:  "on_update",
		EntityName: "product_categories",
		EntityID:   e1,
		EventID:    uuid.New(),
		Timestamp:  time.Now(),
		RawData:    map[string]any{"description": marker},
		UserID:     uid,
	}

	// Dispatch the workflow. CreateExecution writes the StatusPending row synchronously here;
	// the worker then runs the graph asynchronously and (post-#5) advances the status.
	if err := rig.workflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("OnEntityEvent: %v", err)
	}

	// The decisive assertion: the dispatched workflow must advance the record to "completed".
	// RED today (no lifecycle write-back → stuck "pending"); GREEN after #5.
	if !eventually(20*time.Second, 200*time.Millisecond, func() bool {
		return executionStatus(t, ctx, db, rule.ID) == string(workflow.StatusCompleted)
	}) {
		t.Fatalf("execution lifecycle FAILED: rule %s execution status = %q, want %q "+
			"(the cascade/Temporal path must advance pending → running → completed; before fast-follow #5 "+
			"nothing advanced the record so it stayed pending forever)",
			rule.ID, executionStatus(t, ctx, db, rule.ID), workflow.StatusCompleted)
	}
	t.Logf("execution lifecycle proven: cascaded run reached status=%q", workflow.StatusCompleted)
}
