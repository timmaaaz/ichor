package workflow

// Hardening tests for the static cascade-loop detector. This is critical infrastructure, so
// beyond the example-based tests in cascade_detect_test.go we pin:
//
//	1. DIFFERENTIAL — the static edge logic (evalGate) must agree with the REAL runtime trigger
//	   evaluator (TriggerProcessor.evaluateFieldCondition) for the same produced value. The whole
//	   scheme's soundness rests on the static graph matching what actually fires; this test fails
//	   the moment the two drift.
//	2. MUST-NOT-BLOCK — a corpus of legitimate workflow shapes that must NEVER be blocked
//	   (false positives are the dangerous failure mode — they get the feature disabled).
//	3. TARJAN — property test against a brute-force reachability SCC oracle over random graphs.
//	4. RE-ARMABILITY invariants over generated cycles.

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/google/uuid"
)

// ---- 1. Differential: static evalGate vs runtime evaluateFieldCondition -----

func TestCascadeDetect_DifferentialAgainstRuntime(t *testing.T) {
	tp := &TriggerProcessor{} // evaluateFieldCondition is pure (no db/log/bus use)

	// A previous value distinct from every produced/condition value below, so a changed_to
	// producer models a real transition (prev != new).
	const prevSentinel = "<<prev>>"

	type pair struct{ produced, cond any }
	matrix := []struct {
		op    string
		pairs []pair
	}{
		{OperatorEquals, []pair{{"approved", "approved"}, {"approved", "shipped"}, {5, 5}, {5, 6}}},
		{OperatorNotEquals, []pair{{"approved", "approved"}, {"approved", "shipped"}}},
		{OperatorChangedTo, []pair{{"approved", "approved"}, {"approved", "shipped"}}},
		{OperatorGreaterThan, []pair{{10, 5}, {5, 10}, {5, 5}}},
		{OperatorLessThan, []pair{{5, 10}, {10, 5}, {5, 5}}},
		{OperatorContains, []pair{{"hello world", "world"}, {"hello", "xyz"}}},
		{OperatorIn, []pair{{"approved", []any{"approved", "shipped"}}, {"draft", []any{"approved", "shipped"}}}},
	}

	check := func(t *testing.T, op string, produced, condVal any, written bool) {
		t.Helper()
		cond := FieldCondition{FieldName: "f", Operator: op, Value: condVal}

		// Static side.
		prod := map[string]fieldProduction{}
		if written {
			prod["f"] = fieldProduction{writes: true, hasKnown: true, value: produced}
		}
		static := evalGate(cond, prod, "on_update")

		// Runtime side: model "producer set f = produced (from prevSentinel)".
		event := TriggerEvent{EventType: "on_update"}
		if written {
			event.FieldChanges = map[string]FieldChange{"f": {NewValue: produced, OldValue: prevSentinel}}
		}
		runtime := tp.evaluateFieldCondition(cond, event).Matched

		switch static {
		case gateYes:
			if !runtime {
				t.Errorf("op=%s produced=%v cond=%v written=%v: static=YES but runtime did NOT match", op, produced, condVal, written)
			}
		case gateNo:
			if runtime {
				t.Errorf("op=%s produced=%v cond=%v written=%v: static=NO but runtime MATCHED", op, produced, condVal, written)
			}
		case gateIndeterminate:
			// The 'cannot tell statically' band — runtime may go either way. No assertion, but
			// indeterminacy must be justified: with a known written value, only changed_from is
			// allowed to be indeterminate (it depends on the prior value).
			if written && op != OperatorChangedFrom {
				t.Errorf("op=%s produced=%v cond=%v: unexpected INDETERMINATE for a decidable known value", op, produced, condVal)
			}
		}
	}

	for _, m := range matrix {
		for _, p := range m.pairs {
			check(t, m.op, p.produced, p.cond, true)
		}
	}

	// changed_to / changed_from on a field the producer does NOT write must be a provable
	// no-match (no change → cannot fire), and agree with runtime.
	check(t, OperatorChangedTo, "approved", "approved", false)
	check(t, OperatorChangedFrom, "approved", "approved", false)

	// changed_from with a written value is the one operator that must stay indeterminate
	// (the producer determines the new value, not the prior one).
	cond := FieldCondition{FieldName: "f", Operator: OperatorChangedFrom, PreviousValue: "x", Value: "y"}
	if got := evalGate(cond, map[string]fieldProduction{"f": {writes: true, hasKnown: true, value: "y"}}, "on_update"); got != gateIndeterminate {
		t.Errorf("changed_from on a written field must be indeterminate, got %v", got)
	}
}

// ---- 2. Must-not-block corpus ---------------------------------------------

func TestCascadeDetect_LegitimateWorkflowsNotBlocked(t *testing.T) {
	tests := []struct {
		name  string
		build func() ([]ruleNode, uuid.UUID)
	}{
		{
			name: "linear state machine pending->approved->shipped",
			build: func() ([]ruleNode, uuid.UUID) {
				a, b, c := uuid.New(), uuid.New(), uuid.New()
				ra := node(a, "A", "e1", "on_update", []FieldCondition{cond("status", OperatorChangedTo, "pending")}, setMod("e2", "status", "approved"))
				rb := node(b, "B", "e2", "on_update", []FieldCondition{cond("status", OperatorChangedTo, "approved")}, setMod("e3", "status", "shipped"))
				rc := node(c, "C", "e3", "on_update", []FieldCondition{cond("status", OperatorChangedTo, "shipped")})
				return []ruleNode{ra, rb, rc}, a
			},
		},
		{
			name: "fan-out: one writer, three independent consumers",
			build: func() ([]ruleNode, uuid.UUID) {
				a, b, c, d := uuid.New(), uuid.New(), uuid.New(), uuid.New()
				ra := node(a, "A", "src", "on_update", nil, setMod("hub", "status", "ready"))
				rb := node(b, "B", "hub", "on_update", []FieldCondition{cond("status", OperatorChangedTo, "ready")}, setMod("leaf1", "x", "1"))
				rc := node(c, "C", "hub", "on_update", []FieldCondition{cond("status", OperatorChangedTo, "ready")}, setMod("leaf2", "x", "2"))
				rd := node(d, "D", "hub", "on_update", []FieldCondition{cond("status", OperatorChangedTo, "ready")}, setMod("leaf3", "x", "3"))
				return []ruleNode{ra, rb, rc, rd}, a
			},
		},
		{
			name: "convergent sync (changed_to latches, never reset)",
			build: func() ([]ruleNode, uuid.UUID) {
				a, b := uuid.New(), uuid.New()
				ra := node(a, "A", "line_items", "on_update", []FieldCondition{cond("status", OperatorChangedTo, "ALLOCATED")}, setMod("orders", "status", "PROCESSING"))
				rb := node(b, "B", "orders", "on_update", []FieldCondition{cond("status", OperatorChangedTo, "PROCESSING")}, setMod("line_items", "status", "ALLOCATED"))
				return []ruleNode{ra, rb}, a
			},
		},
		{
			name: "two rules touch same field but with non-matching values (no edge)",
			build: func() ([]ruleNode, uuid.UUID) {
				a, b := uuid.New(), uuid.New()
				ra := node(a, "A", "orders", "on_update", []FieldCondition{cond("status", OperatorChangedTo, "approved")}, setMod("orders", "status", "shipped"))
				rb := node(b, "B", "orders", "on_update", []FieldCondition{cond("status", OperatorChangedTo, "rejected")}, setMod("orders", "status", "closed"))
				return []ruleNode{ra, rb}, a
			},
		},
		{
			name: "conditioned self-reference that converges",
			build: func() ([]ruleNode, uuid.UUID) {
				a := uuid.New()
				ra := node(a, "A", "orders", "on_update", []FieldCondition{cond("status", OperatorChangedTo, "pending")}, setMod("orders", "status", "approved"))
				return []ruleNode{ra}, a
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nodes, candID := tc.build()
			res := classifyCascades(nodes, candID)
			if res.HasErrors() {
				t.Fatalf("legitimate workflow must NOT be blocked, got errors: %+v", res.Errors)
			}
		})
	}
}

// ---- 3. Tarjan property test vs brute-force reachability oracle -----------

func TestCascadeDetect_TarjanMatchesBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(1))

	for iter := 0; iter < 600; iter++ {
		n := 1 + rng.Intn(8)
		nodes := make([]uuid.UUID, n)
		for i := range nodes {
			nodes[i] = uuid.New()
		}
		adj := map[uuid.UUID][]uuid.UUID{}
		for _, u := range nodes {
			for _, v := range nodes {
				if rng.Intn(4) == 0 { // ~25% edge density, self-loops allowed
					adj[u] = append(adj[u], v)
				}
			}
		}

		tarjanComp := compIDsFromSCCs(tarjanSCC(nodes, adj))
		bruteComp := bruteForceSCC(nodes, adj)

		// Compare the two as PARTITIONS: every pair of nodes must agree on same-component-ness.
		for _, u := range nodes {
			for _, v := range nodes {
				tSame := tarjanComp[u] == tarjanComp[v]
				bSame := bruteComp[u] == bruteComp[v]
				if tSame != bSame {
					t.Fatalf("iter %d: tarjan/brute disagree for (%s,%s): tarjan-same=%v brute-same=%v\nadj=%v",
						iter, short(u), short(v), tSame, bSame, adj)
				}
			}
		}
	}
}

func compIDsFromSCCs(sccs [][]uuid.UUID) map[uuid.UUID]int {
	out := map[uuid.UUID]int{}
	for i, scc := range sccs {
		for _, id := range scc {
			out[id] = i
		}
	}
	return out
}

// bruteForceSCC computes SCC membership via transitive-closure reachability:
// u,v share a component iff each reaches the other.
func bruteForceSCC(nodes []uuid.UUID, adj map[uuid.UUID][]uuid.UUID) map[uuid.UUID]int {
	reach := map[uuid.UUID]map[uuid.UUID]bool{}
	for _, u := range nodes {
		reach[u] = map[uuid.UUID]bool{u: true}
		for _, w := range adj[u] {
			reach[u][w] = true
		}
	}
	for _, k := range nodes {
		for _, i := range nodes {
			if reach[i][k] {
				for _, j := range nodes {
					if reach[k][j] {
						reach[i][j] = true
					}
				}
			}
		}
	}
	comp := map[uuid.UUID]int{}
	next := 0
	for _, u := range nodes {
		if _, ok := comp[u]; ok {
			continue
		}
		comp[u] = next
		for _, v := range nodes {
			if v != u && reach[u][v] && reach[v][u] {
				comp[v] = next
			}
		}
		next++
	}
	return comp
}

func short(id uuid.UUID) string { return id.String()[:8] }

// ---- 4. Re-armability invariants over generated cycles --------------------

func TestCascadeDetect_ReArmInvariants(t *testing.T) {
	// Invariant A: a cycle whose gates are ALL auto-match always re-arms → ERROR, at any size.
	for size := 1; size <= 5; size++ {
		ids := make([]uuid.UUID, size)
		for i := range ids {
			ids[i] = uuid.New()
		}
		nodes := make([]ruleNode, size)
		for i := range ids {
			entity := fmt.Sprintf("e%d", i)
			nextEntity := fmt.Sprintf("e%d", (i+1)%size)
			// auto-match on its own entity, writes the next entity (closing the ring).
			nodes[i] = node(ids[i], fmt.Sprintf("R%d", i), entity, "on_update", nil, setMod(nextEntity, "status", "x"))
		}
		res := classifyCascades(nodes, ids[0])
		if !res.HasErrors() {
			t.Fatalf("size=%d all-auto-match ring must ERROR, got %+v", size, res)
		}
	}

	// Invariant B: a ring of changed_to latches where NO node resets the gate field off its
	// value converges → never ERROR (the latch sticks), at any size.
	for size := 2; size <= 5; size++ {
		ids := make([]uuid.UUID, size)
		for i := range ids {
			ids[i] = uuid.New()
		}
		nodes := make([]ruleNode, size)
		for i := range ids {
			entity := fmt.Sprintf("e%d", i)
			nextEntity := fmt.Sprintf("e%d", (i+1)%size)
			// gate: this entity status changed_to V; action: set NEXT entity status = V (same V,
			// so the next gate's field is only ever written TO its latch value → never reset).
			nodes[i] = node(ids[i], fmt.Sprintf("R%d", i), entity, "on_update",
				[]FieldCondition{cond("status", OperatorChangedTo, "V")}, setMod(nextEntity, "status", "V"))
		}
		res := classifyCascades(nodes, ids[0])
		if res.HasErrors() {
			t.Fatalf("size=%d convergent changed_to ring must NOT ERROR, got errors: %+v", size, res.Errors)
		}
	}
}
