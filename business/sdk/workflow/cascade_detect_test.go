package workflow

import (
	"testing"

	"github.com/google/uuid"
)

// ---- builders -------------------------------------------------------------

func cond(field, op string, value any) FieldCondition {
	return FieldCondition{FieldName: field, Operator: op, Value: value}
}

// setMod: an on_update modification that sets entity.field to a known literal value.
func setMod(entity, field string, value any) EntityModification {
	return EntityModification{
		EntityName: entity,
		EventType:  "on_update",
		Fields:     []string{field},
		Changes:    []ProducedChange{{FieldName: field, Operator: OperatorChangedTo, Value: value}},
	}
}

// setModIndet: an on_update modification that sets entity.field to an unknown (templated) value.
func setModIndet(entity, field string) EntityModification {
	return EntityModification{
		EntityName: entity,
		EventType:  "on_update",
		Fields:     []string{field},
		Changes:    []ProducedChange{{FieldName: field, Operator: OperatorChangedTo, Indeterminate: true}},
	}
}

func createMod(entity string) EntityModification {
	return EntityModification{EntityName: entity, EventType: "on_create"}
}

func node(id uuid.UUID, name, entity, event string, conds []FieldCondition, mods ...EntityModification) ruleNode {
	return ruleNode{id: id, name: name, entity: entity, event: event, conds: conds, mods: mods}
}

// ---- classifyEdge -------------------------------------------------------

func TestCascadeDetect_classifyEdge(t *testing.T) {
	x := uuid.New()
	y := uuid.New()

	tests := []struct {
		name string
		from ruleNode
		to   ruleNode
		want edgeKind
	}{
		{
			name: "value-match changed_to → definite",
			from: node(x, "X", "li", "on_update", nil, setMod("orders", "status", "approved")),
			to:   node(y, "Y", "orders", "on_update", []FieldCondition{cond("status", OperatorChangedTo, "approved")}),
			want: edgeDefinite,
		},
		{
			name: "value-mismatch changed_to → no edge (state machine, not a loop)",
			from: node(x, "X", "li", "on_update", nil, setMod("orders", "status", "shipped")),
			to:   node(y, "Y", "orders", "on_update", []FieldCondition{cond("status", OperatorChangedTo, "approved")}),
			want: edgeNone,
		},
		{
			name: "auto-match consumer → definite",
			from: node(x, "X", "li", "on_update", nil, setMod("orders", "status", "anything")),
			to:   node(y, "Y", "orders", "on_update", nil),
			want: edgeDefinite,
		},
		{
			name: "different entity → no edge",
			from: node(x, "X", "li", "on_update", nil, setMod("orders", "status", "approved")),
			to:   node(y, "Y", "shipments", "on_update", []FieldCondition{cond("status", OperatorChangedTo, "approved")}),
			want: edgeNone,
		},
		{
			name: "changed_to on on_create consumer → no edge (cannot match on create)",
			from: node(x, "X", "li", "on_create", nil, createMod("orders")),
			to:   node(y, "Y", "orders", "on_create", []FieldCondition{cond("status", OperatorChangedTo, "approved")}),
			want: edgeNone,
		},
		{
			name: "indeterminate produced value on gate field → indeterminate",
			from: node(x, "X", "li", "on_update", nil, setModIndet("orders", "status")),
			to:   node(y, "Y", "orders", "on_update", []FieldCondition{cond("status", OperatorChangedTo, "approved")}),
			want: edgeIndeterminate,
		},
		{
			name: "producer omits a gate field (equals) → indeterminate (pre-existing state)",
			from: node(x, "X", "li", "on_update", nil, setMod("orders", "status", "approved")),
			to: node(y, "Y", "orders", "on_update", []FieldCondition{
				cond("status", OperatorChangedTo, "approved"),
				cond("priority", OperatorEquals, "HIGH"),
			}),
			want: edgeIndeterminate,
		},
		{
			name: "producer also writes a non-gate field indeterminately → still definite",
			from: node(x, "X", "li", "on_update", nil, EntityModification{
				EntityName: "orders", EventType: "on_update",
				Fields: []string{"status", "updated_by"},
				Changes: []ProducedChange{
					{FieldName: "status", Operator: OperatorChangedTo, Value: "approved"},
					{FieldName: "updated_by", Operator: OperatorChangedTo, Indeterminate: true},
				},
			}),
			to:   node(y, "Y", "orders", "on_update", []FieldCondition{cond("status", OperatorChangedTo, "approved")}),
			want: edgeDefinite,
		},
		{
			name: "equals operator with known value → definite",
			from: node(x, "X", "li", "on_update", nil, setMod("orders", "status", "approved")),
			to:   node(y, "Y", "orders", "on_update", []FieldCondition{cond("status", OperatorEquals, "approved")}),
			want: edgeDefinite,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifyEdge(tc.from, tc.to); got != tc.want {
				t.Fatalf("classifyEdge = %v, want %v", got, tc.want)
			}
		})
	}
}

// ---- classifyCascades (three-tier) --------------------------------------

func TestCascadeDetect_AutoMatchSelfLoop_Error(t *testing.T) {
	a := uuid.New()
	// auto-match on orders + writes orders → self-loop that always re-arms.
	cand := node(a, "Self", "orders", "on_update", nil, setMod("orders", "status", "x"))
	res := classifyCascades([]ruleNode{cand}, a)
	if !res.HasErrors() {
		t.Fatalf("auto-match self-loop must be ERROR, got %+v", res)
	}
}

func TestCascadeDetect_ConvergentSync_Info_NotBlocked(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	// §4a convergent sync: both changed_to, neither resets the other's gate field.
	ruleA := node(a, "A", "line_items", "on_update",
		[]FieldCondition{cond("status", OperatorChangedTo, "ALLOCATED")},
		setMod("orders", "status", "PROCESSING"))
	ruleB := node(b, "B", "orders", "on_update",
		[]FieldCondition{cond("status", OperatorChangedTo, "PROCESSING")},
		setMod("line_items", "status", "ALLOCATED"))

	res := classifyCascades([]ruleNode{ruleA, ruleB}, a)
	if res.HasErrors() {
		t.Fatalf("convergent sync must NOT be blocked, got errors: %+v", res.Errors)
	}
	if len(res.Info) == 0 {
		t.Fatalf("convergent sync should produce an INFO finding, got %+v", res)
	}
}

func TestCascadeDetect_EqualsTwoRuleLoop_Error(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	// Both gates use `equals` (no latch) → always re-arms → provable loop.
	ruleA := node(a, "A", "orders", "on_update",
		[]FieldCondition{cond("status", OperatorEquals, "PROCESSING")},
		setMod("line_items", "status", "ALLOCATED"))
	ruleB := node(b, "B", "line_items", "on_update",
		[]FieldCondition{cond("status", OperatorEquals, "ALLOCATED")},
		setMod("orders", "status", "PROCESSING"))

	res := classifyCascades([]ruleNode{ruleA, ruleB}, a)
	if !res.HasErrors() {
		t.Fatalf("equals-gated loop must be ERROR (equals is not a latch), got %+v", res)
	}
}

func TestCascadeDetect_TwoValueToggle_Error(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	// changed_to gates but each side resets the other's field off its value → re-arms.
	ruleA := node(a, "A", "orders", "on_update",
		[]FieldCondition{cond("flag", OperatorChangedTo, "X")},
		setMod("orders", "flag", "Y"))
	ruleB := node(b, "B", "orders", "on_update",
		[]FieldCondition{cond("flag", OperatorChangedTo, "Y")},
		setMod("orders", "flag", "X"))

	res := classifyCascades([]ruleNode{ruleA, ruleB}, a)
	if !res.HasErrors() {
		t.Fatalf("two-value toggle must be ERROR, got %+v", res)
	}
}

func TestCascadeDetect_IndeterminateCycle_Warn_NotError(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	// candidate -> B is indeterminate (B has an extra equals gate on a field A doesn't write);
	// B -> candidate is definite. No D-cycle, but a P-cycle through candidate → WARN.
	cand := node(a, "A", "orders", "on_update",
		[]FieldCondition{cond("status", OperatorChangedTo, "PROCESSING")},
		setMod("line_items", "status", "ALLOCATED"))
	ruleB := node(b, "B", "line_items", "on_update",
		[]FieldCondition{
			cond("status", OperatorChangedTo, "ALLOCATED"),
			cond("priority", OperatorEquals, "HIGH"),
		},
		setMod("orders", "status", "PROCESSING"))

	res := classifyCascades([]ruleNode{cand, ruleB}, a)
	if res.HasErrors() {
		t.Fatalf("indeterminate cycle must NOT be ERROR, got %+v", res.Errors)
	}
	if len(res.Warnings) == 0 {
		t.Fatalf("indeterminate cycle should WARN, got %+v", res)
	}
}

func TestCascadeDetect_ThreeNodeCycle_Error(t *testing.T) {
	a, b, c := uuid.New(), uuid.New(), uuid.New()
	// A→B→C→A, all auto-match on_update writing the next rule's entity.
	ruleA := node(a, "A", "e1", "on_update", nil, setMod("e2", "f", "v"))
	ruleB := node(b, "B", "e2", "on_update", nil, setMod("e3", "f", "v"))
	ruleC := node(c, "C", "e3", "on_update", nil, setMod("e1", "f", "v"))

	res := classifyCascades([]ruleNode{ruleA, ruleB, ruleC}, a)
	if !res.HasErrors() {
		t.Fatalf("A→B→C→A must be ERROR, got %+v", res)
	}
	// The blocking path should include all three rules.
	if len(res.Errors[0].RuleIDs) < 3 {
		t.Fatalf("error path should traverse the 3-node cycle, got %v", res.Errors[0].RuleNames)
	}
}

func TestCascadeDetect_ForwardOnly_InfoOnly(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	// candidate sets orders.status='approved'; B triggers on that but writes nothing cyclic.
	cand := node(a, "A", "purchase_orders", "on_update", nil, setMod("orders", "status", "approved"))
	ruleB := node(b, "B", "orders", "on_update",
		[]FieldCondition{cond("status", OperatorChangedTo, "approved")},
		setMod("shipments", "status", "ready")) // shipments triggers nothing here

	res := classifyCascades([]ruleNode{cand, ruleB}, a)
	if res.HasErrors() || len(res.Warnings) != 0 {
		t.Fatalf("forward-only chain must be clean, got %+v", res)
	}
	if len(res.Info) == 0 {
		t.Fatalf("forward-only chain should surface an INFO datapoint, got %+v", res)
	}
}

func TestCascadeDetect_ConditionedSelfReference_NotBlocked(t *testing.T) {
	a := uuid.New()
	// A conditioned changed_to self-reference that converges (sets a different value than the
	// gate, and the gate value is never re-produced) must NOT be hard-blocked like auto-match.
	cand := node(a, "A", "orders", "on_update",
		[]FieldCondition{cond("status", OperatorChangedTo, "pending")},
		setMod("orders", "status", "approved")) // gate 'pending' is never re-written → converges
	res := classifyCascades([]ruleNode{cand}, a)
	if res.HasErrors() {
		t.Fatalf("convergent conditioned self-reference must NOT be ERROR, got %+v", res.Errors)
	}
}
