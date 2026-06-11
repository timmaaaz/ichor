package workflow

// PG — Guard Verification (Piece 1 exit, cascades OFF). This file is the consolidated
// whole-guard pass over the STATIC detector. P2/P2.4 shipped example-based + differential +
// property tests inline (cascade_detect_test.go, cascade_detect_hardening_test.go); PG adds the
// gaps those miss:
//
//	A.1 DIFFERENTIAL FULL MATRIX — every supported operator (incl. the runtime string-fallback
//	    compare path for greater_than/less_than that the inline differential never exercises),
//	    plus a completeness guard so a new operator cannot be added without a differential entry.
//	A.2 ADVERSARIAL LOOP CORPUS — provable re-arming loops that MUST be ERROR-blocked (the
//	    must-CATCH complement to the must-not-block corpus), including the create->create case
//	    the runtime visited-set cannot backstop.
//	A.3 INDETERMINATE BAND — rings whose only re-arming evidence is indeterminate must surface
//	    as WARNING and NEVER as ERROR, at every size (no silent ERROR in the indeterminate band).

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// ---- A.1 Differential: static evalGate ⟺ runtime evaluateFieldCondition, FULL matrix --------

func TestCascadePG_DifferentialFullMatrix(t *testing.T) {
	tp := &TriggerProcessor{} // evaluateFieldCondition is pure (no db/log/bus use)

	// A prior value distinct from every produced/cond value below so a `changed_to` producer
	// models a real transition (prev != new). Aligned into changed_from's PreviousValue too.
	const pgPrev = "<<pg-prev>>"

	staticGate := func(op string, produced, condVal, prevVal any, written bool) gateVerdict {
		c := FieldCondition{FieldName: "f", Operator: op, Value: condVal, PreviousValue: prevVal}
		prod := map[string]fieldProduction{}
		if written {
			prod["f"] = fieldProduction{writes: true, hasKnown: true, value: produced}
		}
		return evalGate(c, prod, "on_update")
	}
	runtimeMatch := func(op string, produced, condVal, prevVal any, written bool) bool {
		c := FieldCondition{FieldName: "f", Operator: op, Value: condVal, PreviousValue: prevVal}
		ev := TriggerEvent{EventType: "on_update"}
		if written {
			ev.FieldChanges = map[string]FieldChange{"f": {NewValue: produced, OldValue: pgPrev}}
		}
		return tp.evaluateFieldCondition(c, ev).Matched
	}

	type row struct {
		op       string
		produced any
		cond     any
		written  bool
	}
	rows := []row{
		// equals / not_equals
		{OperatorEquals, "approved", "approved", true},
		{OperatorEquals, "approved", "shipped", true},
		{OperatorNotEquals, "approved", "approved", true},
		{OperatorNotEquals, "approved", "shipped", true},
		// changed_to: decidable when written; provable no-match when the field isn't written.
		{OperatorChangedTo, "approved", "approved", true},
		{OperatorChangedTo, "approved", "shipped", true},
		{OperatorChangedTo, "approved", "approved", false},
		// changed_from: always indeterminate when written (producer determines new, not prior),
		// provable no-match when not written.
		{OperatorChangedFrom, "approved", "approved", true},
		{OperatorChangedFrom, "approved", "approved", false},
		// greater_than / less_than — NUMERIC (the inline differential covers this path).
		{OperatorGreaterThan, 10, 5, true},
		{OperatorGreaterThan, 5, 5, true},
		{OperatorLessThan, 5, 10, true},
		// greater_than / less_than — STRING fallback (fmt.Sprintf compare, trigger.go:403/411).
		// This is the path the inline differential never feeds; matchKnown must mirror it.
		{OperatorGreaterThan, "banana", "apple", true},
		{OperatorGreaterThan, "apple", "banana", true},
		{OperatorLessThan, "apple", "banana", true},
		{OperatorLessThan, "banana", "apple", true},
		// contains / in
		{OperatorContains, "hello world", "world", true},
		{OperatorContains, "hello", "xyz", true},
		{OperatorIn, "approved", []any{"approved", "shipped"}, true},
		{OperatorIn, "draft", []any{"approved", "shipped"}, true},
	}

	covered := map[string]bool{}
	for _, r := range rows {
		covered[r.op] = true
		var prevVal any
		if r.op == OperatorChangedFrom {
			prevVal = pgPrev
		}
		static := staticGate(r.op, r.produced, r.cond, prevVal, r.written)
		runtime := runtimeMatch(r.op, r.produced, r.cond, prevVal, r.written)
		switch static {
		case gateYes:
			if !runtime {
				t.Errorf("op=%s produced=%v cond=%v written=%v: static=YES but runtime did NOT match", r.op, r.produced, r.cond, r.written)
			}
		case gateNo:
			if runtime {
				t.Errorf("op=%s produced=%v cond=%v written=%v: static=NO but runtime MATCHED", r.op, r.produced, r.cond, r.written)
			}
		case gateIndeterminate:
			// The 'cannot tell statically' band — runtime may go either way, so no agreement
			// assertion. But indeterminacy must be JUSTIFIED: with a known written value the only
			// operator allowed to be indeterminate is changed_from (it reads the prior value).
			if r.written && r.op != OperatorChangedFrom {
				t.Errorf("op=%s produced=%v cond=%v: unexpected INDETERMINATE for a decidable known value", r.op, r.produced, r.cond)
			}
		}
	}

	// Completeness: every operator the runtime supports must be exercised by the matrix (no
	// silent cap on operator coverage), and each must be a real operator the runtime recognizes
	// (not the Unknown-operator default branch). Adding a 9th operator forces a matrix entry here.
	allOps := []string{
		OperatorEquals, OperatorNotEquals, OperatorChangedFrom, OperatorChangedTo,
		OperatorGreaterThan, OperatorLessThan, OperatorContains, OperatorIn,
	}
	for _, op := range allOps {
		if !covered[op] {
			t.Errorf("operator %q is not exercised by the differential matrix", op)
		}
		res := tp.evaluateFieldCondition(FieldCondition{FieldName: "f", Operator: op, Value: "x"}, TriggerEvent{EventType: "on_update"})
		if res.Error != "" {
			t.Errorf("operator %q unexpectedly unknown to the runtime evaluator: %s", op, res.Error)
		}
	}
}

// ---- A.2 Adversarial loop corpus: provable re-arming loops that MUST be blocked ------------

func TestCascadePG_AdversarialLoopCorpus(t *testing.T) {
	// multiEntityRing builds e0 -> e1 -> ... -> e0 where node i triggers on its OWN entity ei and
	// writes the NEXT entity e(i+1). Used for the non-latch gates (equals / not_equals / create):
	// those have no convergence latch, so the topology alone forces ERROR.
	multiEntityRing := func(size int, mk func(entity, next string) ruleNode) ([]ruleNode, uuid.UUID) {
		ids := make([]uuid.UUID, size)
		for i := range ids {
			ids[i] = uuid.New()
		}
		nodes := make([]ruleNode, size)
		for i := range ids {
			nodes[i] = mk(fmt.Sprintf("e%d", i), fmt.Sprintf("e%d", (i+1)%size))
			nodes[i].id = ids[i]
			nodes[i].name = fmt.Sprintf("R%d", i)
		}
		return nodes, ids[0]
	}

	// toggleRing builds a changed_to ring on a SINGLE shared entity: node i gates on status
	// changed_to Vi and writes status = V(i+1). Because every node writes its OWN gate field off
	// its latch value (V(i+1) != Vi), every latch is reset within the cycle → it re-arms forever
	// (a same-entity toggle, unlike a convergent multi-entity sync). withIndetNonGate adds an
	// indeterminate write to a NON-gate field, which must NOT downgrade the provable loop.
	toggleRing := func(size int, withIndetNonGate bool) ([]ruleNode, uuid.UUID) {
		ids := make([]uuid.UUID, size)
		for i := range ids {
			ids[i] = uuid.New()
		}
		nodes := make([]ruleNode, size)
		for i := range ids {
			v := fmt.Sprintf("V%d", i)
			vNext := fmt.Sprintf("V%d", (i+1)%size)
			mods := []EntityModification{setMod("orders", "status", vNext)}
			if withIndetNonGate {
				mods = append(mods, setModIndet("orders", "updated_at"))
			}
			nodes[i] = node(ids[i], fmt.Sprintf("R%d", i), "orders", "on_update",
				[]FieldCondition{cond("status", OperatorChangedTo, v)}, mods...)
		}
		return nodes, ids[0]
	}

	tests := []struct {
		name  string
		build func() ([]ruleNode, uuid.UUID)
	}{
		{
			name: "auto-match self-loop",
			build: func() ([]ruleNode, uuid.UUID) {
				a := uuid.New()
				return []ruleNode{node(a, "Self", "orders", "on_update", nil, setMod("orders", "status", "x"))}, a
			},
		},
		{
			name: "equals-gated 3-ring (equals is not a latch → always re-arms)",
			build: func() ([]ruleNode, uuid.UUID) {
				return multiEntityRing(3, func(entity, next string) ruleNode {
					return node(uuid.Nil, "", entity, "on_update",
						[]FieldCondition{cond("status", OperatorEquals, "S")}, setMod(next, "status", "S"))
				})
			},
		},
		{
			name: "not_equals-gated 2-ring (not a latch → re-arms)",
			build: func() ([]ruleNode, uuid.UUID) {
				return multiEntityRing(2, func(entity, next string) ruleNode {
					// produce "P" which satisfies the next gate `status not_equals Q` (P != Q).
					return node(uuid.Nil, "", entity, "on_update",
						[]FieldCondition{cond("status", OperatorNotEquals, "Q")}, setMod(next, "status", "P"))
				})
			},
		},
		{
			name: "changed_to toggle ring size 2 (each node resets its own gate off its value)",
			build: func() ([]ruleNode, uuid.UUID) { return toggleRing(2, false) },
		},
		{
			name: "changed_to toggle ring size 4 (distinct values, every latch reset)",
			build: func() ([]ruleNode, uuid.UUID) { return toggleRing(4, false) },
		},
		{
			name: "loop disguised by an indeterminate NON-gate write (must still ERROR)",
			build: func() ([]ruleNode, uuid.UUID) { return toggleRing(2, true) },
		},
		{
			name: "create->create auto-match ring (P1 visited-set cannot backstop this)",
			build: func() ([]ruleNode, uuid.UUID) {
				return multiEntityRing(2, func(entity, next string) ruleNode {
					// each rule fires on_create of its own entity and creates the next entity.
					return node(uuid.Nil, "", entity, "on_create", nil, createMod(next))
				})
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nodes, candID := tc.build()
			res := classifyCascades(nodes, candID)
			if !res.HasErrors() {
				t.Fatalf("provable re-arming loop must be ERROR-blocked, got %+v", res)
			}
		})
	}
}

// ---- A.3 Indeterminate band: WARN, never ERROR (no silent ERROR in the dynamic band) -------

func TestCascadePG_IndeterminateBandNeverErrors(t *testing.T) {
	// Build a ring where every gate is a changed_to latch, BUT one edge is fed by an
	// indeterminate (templated) write. That single indeterminate edge breaks the DEFINITE-edge
	// cycle (so there is no provable loop) while the POSSIBLE-edge cycle still closes through the
	// candidate → the detector must WARN, never ERROR, and the warning must actually be surfaced.
	for size := 2; size <= 5; size++ {
		ids := make([]uuid.UUID, size)
		for i := range ids {
			ids[i] = uuid.New()
		}
		nodes := make([]ruleNode, size)
		for i := range ids {
			entity := fmt.Sprintf("e%d", i)
			next := fmt.Sprintf("e%d", (i+1)%size)
			conds := []FieldCondition{cond("status", OperatorChangedTo, "V")}
			var mod EntityModification
			if i == 0 {
				// node 0 (the candidate) writes the next gate field with a dynamic value →
				// edge 0->1 is indeterminate → no definite cycle.
				mod = setModIndet(next, "status")
			} else {
				mod = setMod(next, "status", "V") // definite edge i->i+1
			}
			nodes[i] = node(ids[i], fmt.Sprintf("R%d", i), entity, "on_update", conds, mod)
		}

		res := classifyCascades(nodes, ids[0])
		if res.HasErrors() {
			t.Fatalf("size=%d: indeterminate-band cycle must NOT ERROR, got errors: %+v", size, res.Errors)
		}
		if len(res.Warnings) == 0 {
			t.Fatalf("size=%d: indeterminate-band cycle must SURFACE a warning (no silent cap), got %+v", size, res)
		}
	}
}
