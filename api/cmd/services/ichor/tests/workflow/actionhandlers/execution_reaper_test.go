package actionhandlers_test

// Fast-follow #5 — execution-record reaper + status round-trip (store-level).
//
// The reaper is the crash-safe backstop for orphaned StatusPending rows: if the process dies
// after the trigger writes the pending record but before ExecuteWorkflow returns, #3's
// delete-on-error can't run (its code died with the process), so the row is stranded. The
// reaper sweeps such rows — but ONLY pending rows with no children, so the DELETE is FK-safe
// (the "MarkExecutionRunning fires first" invariant guarantees a pending row has no children;
// the NOT EXISTS clauses are defense-in-depth for the shouldn't-happen violation).

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	workflowtemporal "github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
)

// insertPendingExecutionAt inserts a pending automation_executions row with an explicit
// executed_at (raw SQL — CreateExecution hardcodes executed_at = now(), so back-dating needs
// a direct insert). Returns its id.
func insertPendingExecutionAt(t *testing.T, ctx context.Context, db *dbtest.Database, executedAt time.Time) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := db.DB.ExecContext(ctx,
		`INSERT INTO workflow.automation_executions (id, entity_type, status, executed_at, trigger_source)
		 VALUES ($1, 'reaper_goroutine_test', 'pending', $2, 'automation')`, id, executedAt); err != nil {
		t.Fatalf("inserting pending execution at %s: %v", executedAt, err)
	}
	return id
}

// seedExecutionRow inserts a bare automation_executions row (nil rule = manual-style, avoids
// needing a seeded rule) at the given status and returns its id.
func seedExecutionRow(t *testing.T, ctx context.Context, store *workflowdb.Store, status workflow.ExecutionStatus) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if err := store.CreateExecution(ctx, workflow.AutomationExecution{
		ID:            id,
		EntityType:    "reaper_test",
		Status:        status,
		TriggerSource: "automation",
	}); err != nil {
		t.Fatalf("seeding %s execution: %v", status, err)
	}
	return id
}

// executionExists reports whether the automation_executions row still exists.
func executionExists(t *testing.T, ctx context.Context, db *dbtest.Database, id uuid.UUID) bool {
	t.Helper()
	var n int
	if err := db.DB.QueryRowContext(ctx,
		`SELECT count(*) FROM workflow.automation_executions WHERE id = $1`, id).Scan(&n); err != nil {
		t.Fatalf("checking execution %s: %v", id, err)
	}
	return n > 0
}

// executionStatusAndError reads a single execution row's status + error_message by id.
func executionStatusAndError(t *testing.T, ctx context.Context, db *dbtest.Database, id uuid.UUID) (string, string) {
	t.Helper()
	var status string
	var errMsg sql.NullString
	if err := db.DB.QueryRowContext(ctx,
		`SELECT status, error_message FROM workflow.automation_executions WHERE id = $1`, id).Scan(&status, &errMsg); err != nil {
		t.Fatalf("reading execution %s: %v", id, err)
	}
	return status, errMsg.String
}

func TestReapStaleExecutions_FKSafety(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDatabase(t, "Test_ReapStaleExecutions")
	ctx := context.Background()
	store := workflowdb.NewStore(db.Log, db.DB)

	pendingNoChild := seedExecutionRow(t, ctx, store, workflow.StatusPending)
	running := seedExecutionRow(t, ctx, store, workflow.StatusRunning)
	pendingWithChild := seedExecutionRow(t, ctx, store, workflow.StatusPending)

	// Give pendingWithChild a notification_deliveries child (FK → automation_executions),
	// simulating the shouldn't-happen case where a pending row has children. The reaper must
	// skip it rather than abort the whole sweep on the FK violation.
	if _, err := db.DB.ExecContext(ctx,
		`INSERT INTO workflow.notification_deliveries
			(notification_id, automation_execution_id, recipient_id, channel, status)
		 VALUES (gen_random_uuid(), $1, gen_random_uuid(), 'in_app', 'pending')`, pendingWithChild); err != nil {
		t.Fatalf("seeding notification_deliveries child: %v", err)
	}

	// Future cutoff so every freshly-inserted row qualifies by age — isolating the status +
	// FK-safety filtering from wall-clock timing (CreateExecution hardcodes executed_at = now()).
	n, err := store.ReapStaleExecutions(ctx, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("ReapStaleExecutions: %v", err)
	}

	if executionExists(t, ctx, db, pendingNoChild) {
		t.Errorf("stale pending row with no children should have been reaped, but it survived")
	}
	if !executionExists(t, ctx, db, running) {
		t.Errorf("running row must NOT be reaped (reaper deletes only pending)")
	}
	if !executionExists(t, ctx, db, pendingWithChild) {
		t.Errorf("pending row WITH a child must NOT be reaped (FK-safety via NOT EXISTS)")
	}
	if n < 1 {
		t.Errorf("expected at least 1 reaped row (the clean pending one), got %d", n)
	}
	t.Logf("reaper FK-safety proven: clean pending reaped (n=%d); running + pending-with-child survived", n)
}

func TestUpdateExecutionStatus_RoundTrip(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDatabase(t, "Test_UpdateExecutionStatus")
	ctx := context.Background()
	store := workflowdb.NewStore(db.Log, db.DB)

	id := seedExecutionRow(t, ctx, store, workflow.StatusPending)
	if s, _ := executionStatusAndError(t, ctx, db, id); s != string(workflow.StatusPending) {
		t.Fatalf("seeded status = %q, want pending", s)
	}

	mustUpdate := func(s workflow.ExecutionStatus, msg string) {
		if err := store.UpdateExecutionStatus(ctx, id, s, msg); err != nil {
			t.Fatalf("UpdateExecutionStatus(%s): %v", s, err)
		}
	}

	mustUpdate(workflow.StatusRunning, "")
	if s, _ := executionStatusAndError(t, ctx, db, id); s != string(workflow.StatusRunning) {
		t.Errorf("after running: status = %q, want running", s)
	}

	mustUpdate(workflow.StatusCompleted, "")
	if s, _ := executionStatusAndError(t, ctx, db, id); s != string(workflow.StatusCompleted) {
		t.Errorf("after completed: status = %q, want completed", s)
	}

	mustUpdate(workflow.StatusFailed, "boom: downstream exploded")
	s, msg := executionStatusAndError(t, ctx, db, id)
	if s != string(workflow.StatusFailed) {
		t.Errorf("after failed: status = %q, want failed", s)
	}
	if msg != "boom: downstream exploded" {
		t.Errorf("after failed: error_message = %q, want the recorded error", msg)
	}
	t.Logf("status round-trip proven: pending → running → completed → failed(+error)")
}

// TestExecutionReaper_Goroutine_ReapsStaleNotFresh exercises the ExecutionReaper.Run ticker end
// to end (the goroutine the composition root launches). It also guards the cutoff arithmetic:
// the reaper computes `time.Now().Add(-Window)`, and a flipped sign would reap EVERY pending row
// — including live in-flight ones — so the "fresh survives" assertion is the real safety net the
// direct ReapStaleExecutions test (which passes an explicit cutoff) cannot provide.
func TestExecutionReaper_Goroutine_ReapsStaleNotFresh(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDatabase(t, "Test_ExecutionReaperGoroutine")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := workflowdb.NewStore(db.Log, db.DB)

	stale := insertPendingExecutionAt(t, ctx, db, time.Now().Add(-48*time.Hour))
	fresh := insertPendingExecutionAt(t, ctx, db, time.Now())

	reaper := workflowtemporal.NewExecutionReaper(db.Log, store, workflowtemporal.ExecutionReaperConfig{
		Interval: 50 * time.Millisecond,
		Window:   24 * time.Hour,
	})
	go func() { _ = reaper.Run(ctx) }()

	if !eventually(5*time.Second, 50*time.Millisecond, func() bool {
		return !executionExists(t, ctx, db, stale)
	}) {
		t.Fatalf("stale (48h-old) pending row was not reaped by the reaper goroutine")
	}
	// Checked AFTER the stale row is gone, so the reaper has provably run a sweep: if the cutoff
	// sign were wrong, this same sweep would have taken the fresh row too.
	if !executionExists(t, ctx, db, fresh) {
		t.Errorf("fresh pending row (within the 24h window) must NOT be reaped — check the now-Window cutoff sign")
	}
	t.Logf("reaper goroutine proven: stale (48h) reaped via ticker; fresh (within 24h window) survived")
}
