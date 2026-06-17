package actionhandlers_test

// PL — fan-out / depth-cap re-evaluation against REAL behavior.
//
// The P1 depth-cap was HELD on the reasoning that no unbounded 1→N fan-out source exists.
// This test pins that live: update_field is the ONLY handler that can touch N rows in one
// call, and per the P4 accepted limitation a multi-row (non-id condition) update fires ONE
// synthesized event with a zero EntityID — NOT one event per row. create_entity and allocate
// are single-row. So there is nothing for a depth-cap to bound, and HELD remains correct.
//
// Convergence / idempotency and sinks-stay-silent are intentionally NOT re-tested here — they
// are already covered, and re-proving them would be redundant:
//   - the runtime visited-set stopping a re-arming cycle is proven decisively by
//     TestCascade_LoopGuard/loop_stopped; the static convergent-cycle classification and the
//     static⟺runtime matcher agreement are proven by the PG differential / loop-corpus tests.
//   - the three sinks (log_audit_entry / create_alert / seek_approval) carry NO delegate
//     channel by construction (no WithDelegate, not in the M1 generic set) and are enumerated
//     as non-firing by the manifest consistency tests; constructing their rabbitmq/alert deps
//     to assert a negative that holds by construction would add cost without coverage.

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/data"

	"github.com/timmaaaz/ichor/business/sdk/workflowdomains"
)

func TestCascade_MultiRowUpdateFiresSingleEvent(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_CascadeFanout")
	ctx := context.Background()
	base := seedConsistencyBase(t, ctx, db)
	uid := base.userID

	const srcDesc, dstDesc = "FANOUT_SRC", "FANOUT_DST"
	const n = 3
	for range n {
		mustSeedCategory(t, ctx, db, srcDesc)
	}

	// Recorder registered AFTER seeding (the create events are out of window) so only the
	// update's synthesized event is observed.
	rec := &delegateRecorder{}
	rec.registerOn(db.BusDomain.Delegate, productcategorybus.DomainName)

	h := data.NewUpdateFieldHandler(db.Log, db.DB,
		data.WithDelegate(db.BusDomain.Delegate),
		data.WithEntityRegistry(workflowdomains.ReverseMap()))
	cfg := mustJSON(t, map[string]any{
		"target_entity": "products.product_categories",
		"target_field":  "description",
		"new_value":     dstDesc,
		"conditions":    []map[string]any{cond("description", "equals", srcDesc)},
	})

	mark := rec.mark()
	if _, err := h.Execute(ctx, cfg, workflow.ActionExecutionContext{UserID: uid}); err != nil {
		t.Fatalf("update_field Execute: %v", err)
	}

	// A genuine multi-row write: all N rows updated.
	if got := categoriesWithDescription(t, ctx, db, dstDesc); got != n {
		t.Fatalf("expected %d rows updated to %q, got %d", n, dstDesc, got)
	}

	// ...but exactly ONE synthesized event fired, with a zero EntityID.
	var updates []firedEvent
	for _, e := range rec.since(mark) {
		if e.domain == productcategorybus.DomainName && e.action == workflow.ActionUpdated {
			updates = append(updates, e)
		}
	}
	if len(updates) != 1 {
		t.Fatalf("multi-row update fired %d events, want exactly 1 (one event for N rows, not per-row)", len(updates))
	}
	if updates[0].entityID != uuid.Nil {
		t.Errorf("multi-row update event EntityID = %s, want zero (written rows unknown without RETURNING)", updates[0].entityID)
	}
	t.Logf("fan-out re-eval: %d-row update fired exactly 1 event (EntityID=nil); no unbounded 1→N source exists → depth-cap HELD remains correct", n)
}

// categoriesWithDescription counts product_categories rows with a given description.
func categoriesWithDescription(t *testing.T, ctx context.Context, db *dbtest.Database, desc string) int {
	t.Helper()
	var c int
	if err := db.DB.QueryRowContext(ctx,
		`SELECT count(*) FROM products.product_categories WHERE description = $1`, desc).Scan(&c); err != nil {
		t.Fatalf("counting product categories with description %q: %v", desc, err)
	}
	return c
}
