package actionhandlers_test

// Fast-follow #5 — execution-record lifecycle, FAILURE path. Companion to
// cascade_lifecycle_test.go (which covers the success path). A cascaded workflow whose action
// errors must advance its automation_executions row to "failed" with the error recorded —
// exercising ExecuteGraphWorkflow's terminal `runErr != nil` branch (MarkExecutionFailed), which
// the success test cannot reach.
//
// Non-vacuity: if the failed terminal branch did NOT fire, the record would be stuck at
// "running" (MarkExecutionRunning fired, no terminal mark), so asserting "failed" genuinely
// catches the branch's absence.

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// alwaysFailHandler is a workflow.ActionHandler whose Execute always errors, used to drive a
// dispatched workflow to failure.
type alwaysFailHandler struct{}

func (alwaysFailHandler) Execute(ctx context.Context, config json.RawMessage, ec workflow.ActionExecutionContext) (any, error) {
	return nil, fmt.Errorf("intentional failure for lifecycle test")
}
func (alwaysFailHandler) Validate(config json.RawMessage) error { return nil }
func (alwaysFailHandler) GetType() string                       { return "lifecycle_fail" }
func (alwaysFailHandler) SupportsManualExecution() bool         { return true }
func (alwaysFailHandler) IsAsync() bool                         { return false }
func (alwaysFailHandler) GetDescription() string                { return "always fails (lifecycle test)" }

func TestCascade_ExecutionLifecycle_ReachesFailed(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_CascadeExecutionLifecycleFailed")
	ctx := context.Background()
	base := seedConsistencyBase(t, ctx, db)
	uid := base.userID

	registry := workflow.NewActionRegistry()
	registry.Register(alwaysFailHandler{})
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

	e1 := mustSeedCategory(t, ctx, db, "lifecycle-fail-e1")

	const marker = "LIFECYCLE_FAIL_MARKER"
	rule := seedActiveRule(t, ctx, rig.workflowBus, uid, "lifecycle-fail-rule",
		pcEntity.ID, entityType.ID, onUpdate.ID,
		fieldConditions(cond("description", "equals", marker)),
		"lifecycle_fail", map[string]any{})
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
	if err := rig.workflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("OnEntityEvent: %v", err)
	}

	// The action always errors; after Temporal exhausts the activity retries the workflow fails,
	// and the terminal branch must drive the record to "failed". 30s tolerates the retry backoff.
	if !eventually(30*time.Second, 200*time.Millisecond, func() bool {
		return executionStatus(t, ctx, db, rule.ID) == string(workflow.StatusFailed)
	}) {
		t.Fatalf("execution lifecycle (failure) FAILED: rule %s status = %q, want %q",
			rule.ID, executionStatus(t, ctx, db, rule.ID), workflow.StatusFailed)
	}

	// The workflow must thread the run error into the record (covers MarkExecutionFailedInput's
	// ErrorMessage wiring — distinct from the store round-trip, which passes a literal message).
	var errMsg sql.NullString
	if err := db.DB.QueryRowContext(ctx,
		`SELECT error_message FROM workflow.automation_executions WHERE automation_rules_id = $1
		 ORDER BY executed_at DESC LIMIT 1`, rule.ID).Scan(&errMsg); err != nil {
		t.Fatalf("reading error_message: %v", err)
	}
	if !errMsg.Valid || errMsg.String == "" {
		t.Errorf("failed execution must record a non-empty error_message, got %q", errMsg.String)
	}
	t.Logf("execution lifecycle (failure) proven: cascaded run reached status=%q with error_message=%q",
		workflow.StatusFailed, errMsg.String)
}
