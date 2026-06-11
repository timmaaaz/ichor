package temporal

import (
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/client"

	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// This file lives in package temporal (internal) so the runtime-loop-guard tests
// can set a parent lineage on a context via the unexported contextWithLineage.
// trigger_test.go's mocks are in package temporal_test and are unreachable here,
// so this file defines its own small recording starter / stub stores.

// =============================================================================
// Test doubles (leaf I/O only — the guard logic under test is the real code)
// =============================================================================

type stubRun struct{}

func (stubRun) GetID() string                  { return "id" }
func (stubRun) GetRunID() string               { return "run" }
func (stubRun) Get(context.Context, any) error { return nil }
func (stubRun) GetWithOptions(context.Context, any, client.WorkflowRunGetOptions) error {
	return nil
}

// recordingStarter records each dispatched WorkflowInput. The mutex makes it safe
// to read from a test goroutine while DelegateHandler dispatches from its own
// goroutine (the chokepoint end-to-end test).
type recordingStarter struct {
	mu    chan struct{} // 1-buffered, used as a mutex (avoids importing sync just for tests)
	calls []WorkflowInput
}

func newRecordingStarter() *recordingStarter {
	s := &recordingStarter{mu: make(chan struct{}, 1)}
	s.mu <- struct{}{}
	return s
}

func (s *recordingStarter) ExecuteWorkflow(_ context.Context, _ client.StartWorkflowOptions, _ any, args ...any) (client.WorkflowRun, error) {
	<-s.mu
	defer func() { s.mu <- struct{}{} }()
	if len(args) > 0 {
		if in, ok := args[0].(WorkflowInput); ok {
			s.calls = append(s.calls, in)
		}
	}
	return stubRun{}, nil
}

func (s *recordingStarter) count() int {
	<-s.mu
	defer func() { s.mu <- struct{}{} }()
	return len(s.calls)
}

func (s *recordingStarter) last() WorkflowInput {
	<-s.mu
	defer func() { s.mu <- struct{}{} }()
	return s.calls[len(s.calls)-1]
}

type stubEdgeStore struct {
	actions map[uuid.UUID][]ActionNode
	edges   map[uuid.UUID][]ActionEdge
}

func newStubEdgeStore() *stubEdgeStore {
	return &stubEdgeStore{
		actions: make(map[uuid.UUID][]ActionNode),
		edges:   make(map[uuid.UUID][]ActionEdge),
	}
}

func (s *stubEdgeStore) QueryActionsByRule(_ context.Context, ruleID uuid.UUID) ([]ActionNode, error) {
	return s.actions[ruleID], nil
}

func (s *stubEdgeStore) QueryEdgesByRule(_ context.Context, ruleID uuid.UUID) ([]ActionEdge, error) {
	return s.edges[ruleID], nil
}

// registerGraph gives a rule a minimal non-empty graph (one action + start edge)
// so startWorkflowForRule does not short-circuit on an empty graph.
func (s *stubEdgeStore) registerGraph(ruleID uuid.UUID) {
	actionID := uuid.New()
	s.actions[ruleID] = []ActionNode{
		{ID: actionID, Name: "noop", ActionType: "send_notification", Config: []byte(`{}`), IsActive: true},
	}
	s.edges[ruleID] = []ActionEdge{
		{ID: uuid.New(), TargetActionID: actionID, EdgeType: EdgeTypeStart, SortOrder: 1},
	}
}

type noopExecutionStore struct{}

func (noopExecutionStore) CreateExecution(context.Context, workflow.AutomationExecution) error {
	return nil
}

type settableMatcher struct {
	result *workflow.ProcessingResult
}

func (m *settableMatcher) ProcessEvent(context.Context, workflow.TriggerEvent) (*workflow.ProcessingResult, error) {
	return m.result, nil
}

// =============================================================================
// Test helpers
// =============================================================================

func guardLogger() *logger.Logger {
	return logger.New(io.Discard, logger.LevelError, "TEST", func(context.Context) string { return "" })
}

func matchOne(ruleID uuid.UUID, name string) *workflow.ProcessingResult {
	return &workflow.ProcessingResult{
		TotalRulesEvaluated: 1,
		MatchedRules: []workflow.RuleMatchResult{
			{Rule: workflow.AutomationRuleView{ID: ruleID, Name: name}, Matched: true},
		},
	}
}

func matchNone() *workflow.ProcessingResult {
	return &workflow.ProcessingResult{TotalRulesEvaluated: 1}
}

func eventFor(entityName string, entityID uuid.UUID) workflow.TriggerEvent {
	return workflow.TriggerEvent{
		EventType:  "on_update",
		EntityName: entityName,
		EntityID:   entityID,
		Timestamp:  time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC),
		UserID:     uuid.MustParse("22222222-2222-2222-2222-222222222222"),
	}
}

// dispatchedLineage extracts the cascade lineage the trigger stamped onto a
// dispatched WorkflowInput (same-process: stored as a WorkflowLineage value).
func dispatchedLineage(t *testing.T, in WorkflowInput) WorkflowLineage {
	t.Helper()
	v, ok := in.TriggerData[cascadeLineageKey]
	require.True(t, ok, "dispatched TriggerData must carry the cascade lineage")
	l, ok := v.(WorkflowLineage)
	require.True(t, ok, "lineage should be a WorkflowLineage value, got %T", v)
	return l
}

// nextHopCtx faithfully simulates how the dispatched lineage reaches the next
// hop in production: Temporal serializes WorkflowInput.TriggerData to JSON, the
// action activity decodes it from its Context map (lineageFromContextMap) and
// stamps it onto the Go context, which propagates through the bus write /
// delegate.Call into the next DelegateHandler -> OnEntityEvent. Round-tripping
// through JSON here also exercises lineageFromContextMap's decode path.
func nextHopCtx(t *testing.T, in WorkflowInput) context.Context {
	t.Helper()
	raw, err := json.Marshal(in.TriggerData)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))
	return contextWithLineage(context.Background(), lineageFromContextMap(decoded))
}

// =============================================================================
// Guard scenario tests (integration: real WorkflowTrigger.OnEntityEvent)
// =============================================================================

// Scenario: A->B->A stops after one hop. The third dispatch would re-enter
// (RuleA, order) which is already in the chain, so it is refused.
func TestGuard_ABA_StopsAfterOneHop(t *testing.T) {
	ruleA := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	ruleB := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	orderID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	lineItemID := uuid.MustParse("99999999-9999-9999-9999-999999999999")

	es := newStubEdgeStore()
	es.registerGraph(ruleA)
	es.registerGraph(ruleB)
	starter := newRecordingStarter()
	matcher := &settableMatcher{}
	trig := NewWorkflowTrigger(guardLogger(), starter, matcher, es, noopExecutionStore{})

	// Hop 0: human write on order -> matches A. No parent lineage.
	matcher.result = matchOne(ruleA, "Rule A")
	require.NoError(t, trig.OnEntityEvent(context.Background(), eventFor("orders", orderID)))
	require.Equal(t, 1, starter.count(), "hop 0 (A) should dispatch")
	l0 := dispatchedLineage(t, starter.last())
	require.True(t, l0.Contains(ruleA, orderID))

	// Hop 1: A's action wrote line_items -> matches B. Carry hop-0 lineage.
	matcher.result = matchOne(ruleB, "Rule B")
	require.NoError(t, trig.OnEntityEvent(nextHopCtx(t, starter.last()), eventFor("order_line_items", lineItemID)))
	require.Equal(t, 2, starter.count(), "hop 1 (B) should dispatch")
	l1 := dispatchedLineage(t, starter.last())
	require.True(t, l1.Contains(ruleA, orderID))
	require.True(t, l1.Contains(ruleB, lineItemID))

	// Hop 2: B's action wrote order -> matches A again on the SAME order. Loop.
	matcher.result = matchOne(ruleA, "Rule A")
	require.NoError(t, trig.OnEntityEvent(nextHopCtx(t, starter.last()), eventFor("orders", orderID)))
	require.Equal(t, 2, starter.count(), "hop 2 re-enters (A, order) and must be refused")
}

// Scenario: A->B->C->done. Three distinct (rule, entity) pairs all progress; the
// chain ends naturally when no rule matches.
func TestGuard_ABC_Progresses(t *testing.T) {
	ruleA := uuid.New()
	ruleB := uuid.New()
	ruleC := uuid.New()
	e0, e1, e2, e3 := uuid.New(), uuid.New(), uuid.New(), uuid.New()

	es := newStubEdgeStore()
	es.registerGraph(ruleA)
	es.registerGraph(ruleB)
	es.registerGraph(ruleC)
	starter := newRecordingStarter()
	matcher := &settableMatcher{}
	trig := NewWorkflowTrigger(guardLogger(), starter, matcher, es, noopExecutionStore{})

	matcher.result = matchOne(ruleA, "A")
	require.NoError(t, trig.OnEntityEvent(context.Background(), eventFor("orders", e0)))
	require.Equal(t, 1, starter.count())

	matcher.result = matchOne(ruleB, "B")
	require.NoError(t, trig.OnEntityEvent(nextHopCtx(t, starter.last()), eventFor("line_items", e1)))
	require.Equal(t, 2, starter.count())

	matcher.result = matchOne(ruleC, "C")
	require.NoError(t, trig.OnEntityEvent(nextHopCtx(t, starter.last()), eventFor("shipments", e2)))
	require.Equal(t, 3, starter.count())

	// Final hop matches nothing -> chain terminates.
	matcher.result = matchNone()
	require.NoError(t, trig.OnEntityEvent(nextHopCtx(t, starter.last()), eventFor("invoices", e3)))
	require.Equal(t, 3, starter.count(), "no further dispatch")

	// The full chain accumulated three distinct pairs.
	final := dispatchedLineage(t, starter.last())
	require.True(t, final.Contains(ruleA, e0))
	require.True(t, final.Contains(ruleB, e1))
	require.True(t, final.Contains(ruleC, e2))
}

// Scenario: convergent parent<->child status sync (DESIGN §4a — the
// ALLOCATED/PROCESSING pair) runs once and self-terminates.
//
// At runtime the real value-aware matcher's `changed_to` latch already disarms
// the second time B writes line_items=ALLOCATED (prev==new -> no match), so the
// chain stops without the visited-set even being consulted. This test models the
// stronger guarantee: even if that latch did NOT fire (matcher matches A again),
// the visited-set is the exact backstop — re-entry of (A, line_items) is refused.
func TestGuard_ConvergentSync_RunsOnce(t *testing.T) {
	ruleA := uuid.New() // on line_items ALLOCATED -> set orders PROCESSING
	ruleB := uuid.New() // on orders PROCESSING    -> set line_items ALLOCATED
	lineItemID := uuid.New()
	orderID := uuid.New()

	es := newStubEdgeStore()
	es.registerGraph(ruleA)
	es.registerGraph(ruleB)
	starter := newRecordingStarter()
	matcher := &settableMatcher{}
	trig := NewWorkflowTrigger(guardLogger(), starter, matcher, es, noopExecutionStore{})

	// line_items -> ALLOCATED matches A (sets orders PROCESSING).
	matcher.result = matchOne(ruleA, "Allocation Success - Update Order")
	require.NoError(t, trig.OnEntityEvent(context.Background(), eventFor("order_line_items", lineItemID)))
	require.Equal(t, 1, starter.count())

	// orders -> PROCESSING matches B (sets line_items ALLOCATED).
	matcher.result = matchOne(ruleB, "Order Processing - Sync Line Items")
	require.NoError(t, trig.OnEntityEvent(nextHopCtx(t, starter.last()), eventFor("orders", orderID)))
	require.Equal(t, 2, starter.count())

	// Worst case: latch does not disarm and A matches again on the same line_item.
	matcher.result = matchOne(ruleA, "Allocation Success - Update Order")
	require.NoError(t, trig.OnEntityEvent(nextHopCtx(t, starter.last()), eventFor("order_line_items", lineItemID)))
	require.Equal(t, 2, starter.count(), "convergent cycle must run once; re-entry refused")
}

// Scenario: the same rule firing for a DIFFERENT entity must NOT be blocked
// (the guard keys on the (rule, entity) pair). E.g. a rule processing a batch of
// distinct rows is legitimate, not a loop.
func TestGuard_SameRuleDifferentEntity_Allowed(t *testing.T) {
	ruleA := uuid.New()
	e1 := uuid.New()
	e2 := uuid.New()

	es := newStubEdgeStore()
	es.registerGraph(ruleA)
	starter := newRecordingStarter()
	matcher := &settableMatcher{}
	trig := NewWorkflowTrigger(guardLogger(), starter, matcher, es, noopExecutionStore{})

	matcher.result = matchOne(ruleA, "A")
	require.NoError(t, trig.OnEntityEvent(context.Background(), eventFor("orders", e1)))
	require.Equal(t, 1, starter.count())

	// Same rule, different entity, carrying the prior lineage -> still dispatches.
	matcher.result = matchOne(ruleA, "A")
	require.NoError(t, trig.OnEntityEvent(nextHopCtx(t, starter.last()), eventFor("orders", e2)))
	require.Equal(t, 2, starter.count(), "(A, e2) is distinct from (A, e1) and must dispatch")

	final := dispatchedLineage(t, starter.last())
	require.True(t, final.Contains(ruleA, e1))
	require.True(t, final.Contains(ruleA, e2))
}

// Scenario: the visited-set survives Continue-As-New. Temporal serializes
// WorkflowInput (including ContinuationState=*MergedContext) to JSON across the
// CAN boundary; the lineage must still be readable from the restored Flattened
// map, which is what the action activity reads.
func TestGuard_VisitedSetSurvivesContinueAsNew(t *testing.T) {
	ruleA := uuid.New()
	entityID := uuid.New()
	execID := uuid.New()

	lineage := WorkflowLineage{
		Visited:                []string{lineagePairKey(ruleA, entityID)},
		OriginatingExecutionID: execID,
	}
	triggerData := map[string]any{cascadeLineageKey: lineage}

	// Before CAN: the activity reads the lineage from MergedContext.Flattened.
	mc := NewMergedContext(triggerData)
	require.True(t, lineageFromContextMap(mc.Flattened).Contains(ruleA, entityID))

	// CAN: workflow.go sets input.ContinuationState = mergedCtx, and Temporal
	// serializes WorkflowInput to JSON, then deserializes it for the new run.
	in := WorkflowInput{
		RuleID:            ruleA,
		ExecutionID:       execID,
		Graph:             GraphDefinition{Actions: []ActionNode{{ID: uuid.New()}}},
		TriggerData:       triggerData,
		ContinuationState: mc,
	}
	raw, err := json.Marshal(in)
	require.NoError(t, err)
	var restored WorkflowInput
	require.NoError(t, json.Unmarshal(raw, &restored))

	// After CAN: workflow.go sets mergedCtx = input.ContinuationState; the activity
	// reads its Flattened map, where the lineage must still resolve.
	require.NotNil(t, restored.ContinuationState)
	got := lineageFromContextMap(restored.ContinuationState.Flattened)
	require.True(t, got.Contains(ruleA, entityID), "lineage lost across Continue-As-New")
	require.Equal(t, execID, got.OriginatingExecutionID)
}

// Scenario: full chokepoint end-to-end. A real DelegateHandler reads the lineage
// stamped on ctx (as the action activity would), and after its dispatch goroutine
// the real WorkflowTrigger has seeded the next generation onto the dispatched
// WorkflowInput. Proves the lineage survives delegatehandler.go's goroutine.
func TestGuard_DelegateHandler_ThreadsAndSeedsLineage(t *testing.T) {
	ruleA := uuid.New()
	ruleB := uuid.New()
	orderID := uuid.New()
	lineItemID := uuid.New()

	es := newStubEdgeStore()
	es.registerGraph(ruleB)
	starter := newRecordingStarter()
	matcher := &settableMatcher{result: matchOne(ruleB, "Rule B")}
	trig := NewWorkflowTrigger(guardLogger(), starter, matcher, es, noopExecutionStore{})
	h := NewDelegateHandler(guardLogger(), trig)

	// Simulate the action activity having stamped a parent lineage {(A, order)}.
	parent := WorkflowLineage{Visited: []string{lineagePairKey(ruleA, orderID)}}
	ctx := contextWithLineage(context.Background(), parent)

	params := workflow.DelegateEventParams{EntityID: lineItemID, UserID: uuid.New()}
	raw, err := json.Marshal(params)
	require.NoError(t, err)

	require.NoError(t, h.handleEvent(ctx, "on_update", "order_line_items", delegate.Data{RawParams: raw}))

	require.Eventually(t, func() bool { return starter.count() == 1 }, 2*time.Second, 5*time.Millisecond,
		"DelegateHandler should dispatch one workflow")

	seeded := dispatchedLineage(t, starter.last())
	require.True(t, seeded.Contains(ruleA, orderID), "parent pair must survive the goroutine")
	require.True(t, seeded.Contains(ruleB, lineItemID), "new (B, line_item) pair must be seeded")
}

// =============================================================================
// lineage.go unit tests
// =============================================================================

func TestWorkflowLineage_Contains(t *testing.T) {
	r1, r2 := uuid.New(), uuid.New()
	e1, e2 := uuid.New(), uuid.New()

	var empty WorkflowLineage
	require.False(t, empty.Contains(r1, e1))

	l := empty.With(r1, e1)
	require.True(t, l.Contains(r1, e1))
	require.False(t, l.Contains(r2, e1), "different rule")
	require.False(t, l.Contains(r1, e2), "different entity")
}

func TestWorkflowLineage_With_IsImmutable(t *testing.T) {
	r1, r2 := uuid.New(), uuid.New()
	e1, e2 := uuid.New(), uuid.New()

	parent := WorkflowLineage{
		Visited:                []string{lineagePairKey(r1, e1)},
		OriginatingExecutionID: uuid.New(),
	}
	child := parent.With(r2, e2)

	// Parent unchanged — a parent can seed multiple children safely.
	require.Len(t, parent.Visited, 1)
	require.False(t, parent.Contains(r2, e2))

	// Child has both pairs and preserves the chain root.
	require.Len(t, child.Visited, 2)
	require.True(t, child.Contains(r1, e1))
	require.True(t, child.Contains(r2, e2))
	require.Equal(t, parent.OriginatingExecutionID, child.OriginatingExecutionID)
}

func TestLineageContextRoundTrip(t *testing.T) {
	r, e := uuid.New(), uuid.New()
	l := WorkflowLineage{Visited: []string{lineagePairKey(r, e)}, OriginatingExecutionID: uuid.New()}

	got := lineageFromContext(contextWithLineage(context.Background(), l))
	require.Equal(t, l, got)

	// Missing -> zero value (empty set), never a panic.
	require.Empty(t, lineageFromContext(context.Background()).Visited)
}

func TestLineageFromContextMap(t *testing.T) {
	r, e := uuid.New(), uuid.New()
	execID := uuid.New()
	l := WorkflowLineage{Visited: []string{lineagePairKey(r, e)}, OriginatingExecutionID: execID}

	t.Run("struct value (same-process)", func(t *testing.T) {
		got := lineageFromContextMap(map[string]any{cascadeLineageKey: l})
		require.Equal(t, l, got)
	})

	t.Run("json-roundtrip map (Temporal-deserialized)", func(t *testing.T) {
		// Marshal a TriggerData carrying the struct, then unmarshal to map[string]any
		// — exactly what the activity sees after Temporal serialization.
		raw, err := json.Marshal(map[string]any{cascadeLineageKey: l})
		require.NoError(t, err)
		var decoded map[string]any
		require.NoError(t, json.Unmarshal(raw, &decoded))

		got := lineageFromContextMap(decoded)
		require.True(t, got.Contains(r, e))
		require.Equal(t, execID, got.OriginatingExecutionID)
	})

	t.Run("missing key -> empty", func(t *testing.T) {
		require.Empty(t, lineageFromContextMap(map[string]any{"other": 1}).Visited)
	})

	t.Run("nil value -> empty", func(t *testing.T) {
		require.Empty(t, lineageFromContextMap(map[string]any{cascadeLineageKey: nil}).Visited)
	})

	t.Run("nil map -> empty", func(t *testing.T) {
		require.Empty(t, lineageFromContextMap(nil).Visited)
	})

	t.Run("wrong type -> empty", func(t *testing.T) {
		require.Empty(t, lineageFromContextMap(map[string]any{cascadeLineageKey: "garbage"}).Visited)
	})
}
