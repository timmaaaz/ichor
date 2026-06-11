package temporal

// PG — Guard Verification (Piece 1 exit, cascades OFF): the whole-guard pass over the RUNTIME
// visited-set guard. P1.3 shipped the sequential DESIGN scenarios + Continue-As-New survival +
// the DelegateHandler chokepoint (lineage_test.go). PG adds the one realistic CONCURRENCY path
// those sequential scenarios don't hit: in production a single bus write can fire several
// delegate events, each handled in its OWN goroutine reading the SAME parent lineage off the
// shared activity ctx. This drives N concurrent OnEntityEvent calls sharing one parent-lineage
// ctx and asserts each dispatched child = parent ∪ {its own pair} with no cross-contamination —
// exercising WorkflowLineage.With() immutability + concurrent ctx reads under `-race`.

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// all returns a snapshot of every recorded WorkflowInput (mutex-guarded — safe to call after a
// concurrent fan-out). Extends the recordingStarter double defined in lineage_test.go.
func (s *recordingStarter) all() []WorkflowInput {
	<-s.mu
	defer func() { s.mu <- struct{}{} }()
	out := make([]WorkflowInput, len(s.calls))
	copy(out, s.calls)
	return out
}

// Scenario: N delegate events fan out concurrently, each carrying the SAME parent lineage on the
// shared ctx, each matching the same rule on a DISTINCT entity. Every dispatch must preserve the
// parent pair and add only its own (rule, entity) — never another sibling's. Run with `-race`.
func TestGuardPG_ConcurrentSiblings_NoSharedStateRace(t *testing.T) {
	const n = 16

	ruleP := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd") // parent (chain root) rule
	entityP := uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
	ruleA := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa") // the rule every sibling matches

	es := newStubEdgeStore()
	es.registerGraph(ruleA)
	starter := newRecordingStarter()
	matcher := &settableMatcher{result: matchOne(ruleA, "Rule A")}
	trig := NewWorkflowTrigger(guardLogger(), starter, matcher, es, noopExecutionStore{})

	// One shared parent lineage on one shared ctx — exactly what concurrent delegate goroutines
	// off a single activity ctx would read.
	parent := WorkflowLineage{Visited: []string{lineagePairKey(ruleP, entityP)}}
	parentCtx := contextWithLineage(context.Background(), parent)

	entities := make([]uuid.UUID, n)
	for i := range entities {
		entities[i] = uuid.New()
	}

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(entityID uuid.UUID) {
			defer wg.Done()
			// Each sibling is a distinct (ruleA, entityID) pair → none is a re-entry → all dispatch.
			require.NoError(t, trig.OnEntityEvent(parentCtx, eventFor("orders", entityID)))
		}(entities[i])
	}
	wg.Wait()

	calls := starter.all()
	require.Len(t, calls, n, "every distinct (rule, entity) sibling must dispatch")

	// Each dispatched child must carry the parent pair + exactly one sibling pair, and the set of
	// sibling pairs across all children must be exactly {(ruleA, entities[i])} with no duplicates
	// and no cross-contamination.
	wantEntity := make(map[uuid.UUID]bool, n)
	for _, e := range entities {
		wantEntity[e] = true
	}
	seen := make(map[uuid.UUID]int, n)
	for _, in := range calls {
		l := dispatchedLineage(t, in)
		require.True(t, l.Contains(ruleP, entityP), "parent pair must survive every concurrent dispatch")

		var mine uuid.UUID
		matches := 0
		for e := range wantEntity {
			if l.Contains(ruleA, e) {
				mine = e
				matches++
			}
		}
		require.Equal(t, 1, matches, "child must contain exactly one sibling pair (no cross-contamination): %v", l.Visited)
		seen[mine]++
		// The visited-set is exactly {parent, mine} — no extra pairs leaked in.
		require.Len(t, l.Visited, 2, "child lineage must be parent ∪ {its own pair} only: %v", l.Visited)
	}

	for _, e := range entities {
		require.Equal(t, 1, seen[e], fmt.Sprintf("entity %s should be dispatched exactly once", e))
	}
}
