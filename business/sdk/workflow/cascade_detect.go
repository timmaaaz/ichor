package workflow

// STATIC CASCADE-LOOP DETECTOR (P2)
//
// Builds one inter-rule graph over the ACTIVE rule set plus the rule currently being
// saved/activated (the "candidate"), and reports cascade relationships in three tiers:
//
//	ERROR   — a PROVABLE re-arming loop. The save/activation is blocked.
//	WARNING — a POSSIBLE loop we cannot prove (dynamic/templated values, or a reset we
//	          cannot evaluate statically). Surfaced, never blocked. The P1 runtime
//	          visited-set guard is the backstop for these.
//	INFO    — a cascade-awareness datapoint ("this rule can trigger / be triggered by
//	          rule X"). Pure awareness, never a prohibition.
//
// An edge X -> Y exists iff X produces a DB change whose delegate event would satisfy Y's
// trigger. Edges are VALUE-AWARE (DESIGN §4): "A & B both touch status" is NOT an edge;
// "A sets status='approved' and B triggers on status changed_to 'approved'" IS. We mirror
// the runtime trigger evaluator (TriggerProcessor.evaluateFieldCondition) exactly so the
// static graph agrees with what fires at runtime.
//
// Re-armability (DESIGN §4a): a bare cycle is not enough to block. `changed_to V` is a
// fixed-point latch — once the field sits at V it is disarmed until moved off V. A cycle
// only loops forever if every gate can re-arm. A cycle with a `changed_to` latch that the
// cycle never resets off V is convergent (self-terminating) and must NOT be blocked.
//
// Conservatism principle: only ERROR-block what we can PROVE re-arms; everything uncertain
// is WARN. False negatives are backstopped by the P1 runtime guard; false positives (wrongly
// blocking a legitimate workflow) are the dangerous failure mode.

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// =============================================================================
// Public output types
// =============================================================================

// CascadeFinding describes one cascade relationship discovered by the static detector.
type CascadeFinding struct {
	// RuleIDs / RuleNames are the ordered participants. For a loop finding this is the
	// cycle path (A -> B -> A); for an INFO datapoint it is the [source, target] pair.
	RuleIDs   []uuid.UUID `json:"rule_ids,omitempty"`
	RuleNames []string    `json:"rule_names,omitempty"`
	// Reason is a human-readable explanation of the finding (the path + why it is an
	// error/warning, or what cascade the datapoint represents).
	Reason string `json:"reason"`
}

// CascadeAnalysis is the three-tier output of the static loop detector.
type CascadeAnalysis struct {
	Errors   []CascadeFinding `json:"errors,omitempty"`   // provable re-arming loops — BLOCK
	Warnings []CascadeFinding `json:"warnings,omitempty"` // possible loops (indeterminate) — surface only
	Info     []CascadeFinding `json:"info,omitempty"`     // cascade-awareness datapoints
}

// HasErrors reports whether the analysis found a provable loop that must block the save.
func (a CascadeAnalysis) HasErrors() bool { return len(a.Errors) > 0 }

// ErrorSummary renders the blocking findings into a single message suitable for an
// InvalidArgument error returned to the caller.
func (a CascadeAnalysis) ErrorSummary() string {
	parts := make([]string, 0, len(a.Errors))
	for _, f := range a.Errors {
		parts = append(parts, f.Reason)
	}
	return strings.Join(parts, "; ")
}

// =============================================================================
// Candidate input (built by each enforcement hook from its in-flight data)
// =============================================================================

// CandidateRule is the rule under evaluation as it WOULD be after the save/activation,
// expressed from the in-flight request (NOT re-queried — its actions may not be persisted
// yet at save time).
type CandidateRule struct {
	RuleID            uuid.UUID
	Name              string
	IsActive          bool
	EntityID          uuid.UUID
	TriggerTypeID     uuid.UUID
	TriggerConditions *json.RawMessage
	Actions           []CandidateAction
}

// CandidateAction is one action of the candidate rule (its type + config).
type CandidateAction struct {
	ActionType string
	Config     json.RawMessage
}

// =============================================================================
// Internal graph model (pure — unit-testable with no DB / registry)
// =============================================================================

// ruleNode is one node in the inter-rule graph.
type ruleNode struct {
	id          uuid.UUID
	name        string
	isCandidate bool
	entity      string // schema-stripped table name the rule's trigger listens to
	event       string // on_create / on_update / on_delete
	conds       []FieldCondition
	mods        []EntityModification
}

type edgeKind int

const (
	edgeNone edgeKind = iota
	edgeIndeterminate
	edgeDefinite
)

type gateVerdict int

const (
	gateNo gateVerdict = iota // condition provably cannot match
	gateIndeterminate         // cannot tell statically
	gateYes                   // condition provably matches
)

// fieldProduction summarizes what a node produces for a single field on a given (entity,event).
type fieldProduction struct {
	writes        bool // the node writes this field at all
	hasKnown      bool // a statically-known value exists
	value         any  // the known value (valid only when hasKnown && !indeterminate)
	indeterminate bool // at least one write of this field has an unknown/conflicting value
}

// =============================================================================
// Value comparison — mirrors TriggerProcessor.compareValues / toFloat64 exactly
// =============================================================================

func valuesEqual(a, b any) bool {
	if a == nil || b == nil {
		return a == b
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// numFloat64 mirrors TriggerProcessor.toFloat64 (numeric types only, no string parsing),
// distinct from the package's template toFloat64 which coerces strings.
func numFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

// matchKnown evaluates a non-change operator against a statically-known produced value,
// mirroring the runtime evaluator's per-operator semantics.
func matchKnown(op string, produced, condVal any) bool {
	switch op {
	case OperatorEquals:
		return valuesEqual(produced, condVal)
	case OperatorNotEquals:
		return !valuesEqual(produced, condVal)
	case OperatorGreaterThan, OperatorLessThan:
		af, aok := numFloat64(produced)
		bf, bok := numFloat64(condVal)
		if aok && bok {
			if op == OperatorGreaterThan {
				return af > bf
			}
			return af < bf
		}
		if op == OperatorGreaterThan {
			return fmt.Sprintf("%v", produced) > fmt.Sprintf("%v", condVal)
		}
		return fmt.Sprintf("%v", produced) < fmt.Sprintf("%v", condVal)
	case OperatorContains:
		ps, ok1 := produced.(string)
		cs, ok2 := condVal.(string)
		return ok1 && ok2 && strings.Contains(ps, cs)
	case OperatorIn:
		if vals, ok := condVal.([]any); ok {
			for _, v := range vals {
				if valuesEqual(produced, v) {
					return true
				}
			}
		}
		return false
	default:
		return false
	}
}

// isConvergentLatch reports whether an operator is a fixed-point latch — the only operators
// whose gate can self-disarm (DESIGN §4a). `changed_to`/`changed_from` require the field to
// transition; every other operator re-fires idempotently and therefore always re-arms.
func isConvergentLatch(op string) bool {
	return op == OperatorChangedTo || op == OperatorChangedFrom
}

// =============================================================================
// Edge classification (pure)
// =============================================================================

func schemaStripTable(name string) string {
	if idx := strings.LastIndex(name, "."); idx != -1 {
		return name[idx+1:]
	}
	return name
}

// producesOnEvent reports whether node mods include any modification to (entity,event),
// even one with no field-level detail (e.g. create_entity emits on_create with no fields).
func producesOnEvent(mods []EntityModification, entity, event string) bool {
	for _, m := range mods {
		if schemaStripTable(m.EntityName) == entity && m.EventType == event {
			return true
		}
	}
	return false
}

// producedOnEvent collapses every modification the node makes to (entity,event) into a
// per-field production summary. Multiple writers of the same field aggregate: any
// indeterminate or conflicting writer makes the field's produced value indeterminate.
func producedOnEvent(mods []EntityModification, entity, event string) map[string]fieldProduction {
	out := map[string]fieldProduction{}
	for _, m := range mods {
		if schemaStripTable(m.EntityName) != entity || m.EventType != event {
			continue
		}
		covered := map[string]bool{}
		for _, ch := range m.Changes {
			covered[ch.FieldName] = true
			fp := out[ch.FieldName]
			fp.writes = true
			switch {
			case ch.Indeterminate || ch.Value == nil:
				fp.indeterminate = true
			case fp.hasKnown && !valuesEqual(fp.value, ch.Value):
				fp.indeterminate = true // conflicting known writers
			default:
				fp.value = ch.Value
				fp.hasKnown = true
			}
			out[ch.FieldName] = fp
		}
		// Fields listed without a corresponding Change carry an unknown value.
		for _, f := range m.Fields {
			if covered[f] {
				continue
			}
			fp := out[f]
			fp.writes = true
			fp.indeterminate = true
			out[f] = fp
		}
	}
	return out
}

// evalGate evaluates a single trigger condition of the consumer against the producer's
// per-field production, mirroring runtime semantics.
func evalGate(c FieldCondition, prod map[string]fieldProduction, consumerEvent string) gateVerdict {
	fp := prod[c.FieldName]

	switch c.Operator {
	case OperatorChangedTo:
		// changed_to matches only on_update, requires the field to change (so the producer must
		// write it), and the new value to equal the gate value.
		if consumerEvent != "on_update" {
			return gateNo
		}
		if !fp.writes {
			return gateNo // field not written → no change → cannot match
		}
		if fp.indeterminate || !fp.hasKnown {
			return gateIndeterminate
		}
		if valuesEqual(fp.value, c.Value) {
			return gateYes
		}
		return gateNo

	case OperatorChangedFrom:
		// changed_from matches on the PRIOR value (prev == PreviousValue) and does NOT require a
		// change. The producer never determines the prior value, so this is always indeterminate
		// on on_update (and cannot match on create/delete). It therefore never yields a definite
		// edge — cascade loops through a changed_from gate surface as WARN, not a hard block.
		if consumerEvent != "on_update" {
			return gateNo
		}
		return gateIndeterminate

	default:
		// equals / not_equals / greater_than / less_than / contains / in read the current
		// value, which may be pre-existing state when the producer doesn't write the field.
		if !fp.writes {
			return gateIndeterminate
		}
		if fp.indeterminate || !fp.hasKnown {
			return gateIndeterminate
		}
		if matchKnown(c.Operator, fp.value, c.Value) {
			return gateYes
		}
		return gateNo
	}
}

// classifyEdge returns whether x's mutations provably (definite), possibly (indeterminate),
// or never (none) satisfy y's trigger.
func classifyEdge(x, y ruleNode) edgeKind {
	if !producesOnEvent(x.mods, y.entity, y.event) {
		return edgeNone
	}
	// Auto-match consumer: any matching event triggers it.
	if len(y.conds) == 0 {
		return edgeDefinite
	}

	prod := producedOnEvent(x.mods, y.entity, y.event)
	verdict := gateYes // AND across all conditions
	for _, c := range y.conds {
		switch evalGate(c, prod, y.event) {
		case gateNo:
			return edgeNone // one condition provably fails → trigger cannot fire
		case gateIndeterminate:
			if verdict == gateYes {
				verdict = gateIndeterminate
			}
		}
	}
	if verdict == gateYes {
		return edgeDefinite
	}
	return edgeIndeterminate
}

// =============================================================================
// Graph assembly + cycle classification (pure)
// =============================================================================

// edgeMatrix is edges[from][to] = kind for every ordered pair with a non-none edge.
type edgeMatrix map[uuid.UUID]map[uuid.UUID]edgeKind

func buildEdges(nodes []ruleNode) edgeMatrix {
	edges := edgeMatrix{}
	for _, x := range nodes {
		for _, y := range nodes {
			k := classifyEdge(x, y)
			if k == edgeNone {
				continue
			}
			if edges[x.id] == nil {
				edges[x.id] = map[uuid.UUID]edgeKind{}
			}
			edges[x.id][y.id] = k
		}
	}
	return edges
}

// adjacency builds a successor map keeping edges of at least the given minimum kind.
func (edges edgeMatrix) adjacency(min edgeKind) map[uuid.UUID][]uuid.UUID {
	adj := map[uuid.UUID][]uuid.UUID{}
	for from, tos := range edges {
		for to, k := range tos {
			if k >= min {
				adj[from] = append(adj[from], to)
			}
		}
	}
	for from := range adj {
		sort.Slice(adj[from], func(i, j int) bool { return adj[from][i].String() < adj[from][j].String() })
	}
	return adj
}

// tarjanSCC returns the strongly-connected components of the graph. nodes must be the full,
// stably-ordered node-id list (so output is deterministic).
func tarjanSCC(nodes []uuid.UUID, adj map[uuid.UUID][]uuid.UUID) [][]uuid.UUID {
	index := 0
	indices := map[uuid.UUID]int{}
	low := map[uuid.UUID]int{}
	onStack := map[uuid.UUID]bool{}
	var stack []uuid.UUID
	var out [][]uuid.UUID

	var strongConnect func(v uuid.UUID)
	strongConnect = func(v uuid.UUID) {
		indices[v] = index
		low[v] = index
		index++
		stack = append(stack, v)
		onStack[v] = true

		for _, w := range adj[v] {
			if _, seen := indices[w]; !seen {
				strongConnect(w)
				if low[w] < low[v] {
					low[v] = low[w]
				}
			} else if onStack[w] {
				if indices[w] < low[v] {
					low[v] = indices[w]
				}
			}
		}

		if low[v] == indices[v] {
			var comp []uuid.UUID
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				comp = append(comp, w)
				if w == v {
					break
				}
			}
			out = append(out, comp)
		}
	}

	for _, v := range nodes {
		if _, seen := indices[v]; !seen {
			strongConnect(v)
		}
	}
	return out
}

// findCyclePath returns an ordered cycle through start within the SCC node set (start -> ... ->
// last, where last has an edge back to start), or nil. Used to render a readable path.
func findCyclePath(start uuid.UUID, adj map[uuid.UUID][]uuid.UUID, inSCC map[uuid.UUID]bool) []uuid.UUID {
	var path []uuid.UUID
	visited := map[uuid.UUID]bool{}

	var dfs func(u uuid.UUID) bool
	dfs = func(u uuid.UUID) bool {
		path = append(path, u)
		visited[u] = true
		for _, w := range adj[u] {
			if !inSCC[w] {
				continue
			}
			if w == start {
				return true
			}
			if !visited[w] {
				if dfs(w) {
					return true
				}
			}
		}
		path = path[:len(path)-1]
		visited[u] = false
		return false
	}

	if dfs(start) {
		return path
	}
	return nil
}

// reArm classifies the re-armability of a definite-edge SCC (DESIGN §4a).
//
//	"error"   — provably re-arming: no convergence latch survives the cycle.
//	"info"    — convergent: a changed_to latch the cycle never resets off its value.
//	"warning" — a needed reset is indeterminate (cannot prove convergence or re-arming).
func reArm(scc []uuid.UUID, nodeByID map[uuid.UUID]ruleNode, dAdj map[uuid.UUID][]uuid.UUID) (tier string, reason string) {
	inSCC := map[uuid.UUID]bool{}
	for _, id := range scc {
		inSCC[id] = true
	}

	// Collect the changed_to latches gating the intra-SCC definite edges. A latch is
	// {entity, field, value} taken from a consumer Y that is actually fed within the SCC.
	type latch struct {
		entity string
		field  string
		value  any
	}
	var latches []latch
	for _, yid := range scc {
		y := nodeByID[yid]
		// y is fed within the SCC iff some SCC node has a definite edge into y.
		fed := false
		for _, xid := range scc {
			if slices.Contains(dAdj[xid], yid) {
				fed = true
				break
			}
		}
		if !fed {
			continue
		}
		for _, c := range y.conds {
			if !isConvergentLatch(c.Operator) {
				continue
			}
			// The value the gate latches at: changed_to gates on the new value, changed_from
			// on the prior value. (changed_from never yields a definite edge so it won't
			// actually reach here, but handle it for correctness.)
			v := c.Value
			if c.Operator == OperatorChangedFrom {
				v = c.PreviousValue
			}
			if v != nil {
				latches = append(latches, latch{entity: y.entity, field: c.FieldName, value: v})
			}
		}
	}

	// No convergence latch anywhere → every gate always re-arms → provably looping.
	if len(latches) == 0 {
		return "error", "every trigger in the cycle re-arms on each pass (no self-terminating condition)"
	}

	// For each latch, scan all SCC nodes for a reset (a write of (entity,field) to a value
	// other than the latch value). A known reset re-arms it; an indeterminate write makes
	// convergence unprovable.
	anyUncertain := false
	for _, l := range latches {
		knownReset := false
		indetReset := false
		for _, nid := range scc {
			for _, m := range nodeByID[nid].mods {
				if schemaStripTable(m.EntityName) != l.entity {
					continue
				}
				covered := map[string]bool{}
				for _, ch := range m.Changes {
					if ch.FieldName != l.field {
						continue
					}
					covered[ch.FieldName] = true
					if ch.Indeterminate || ch.Value == nil {
						indetReset = true
					} else if !valuesEqual(ch.Value, l.value) {
						knownReset = true
					}
				}
				for _, f := range m.Fields {
					if f == l.field && !covered[f] {
						indetReset = true
					}
				}
			}
		}
		if !knownReset && !indetReset {
			// This latch is never moved off its value → it sticks → the cycle converges.
			return "info", fmt.Sprintf("converges: %s.%s latches at %v and is never reset within the cycle", l.entity, l.field, l.value)
		}
		if indetReset && !knownReset {
			anyUncertain = true
		}
	}

	if anyUncertain {
		return "warning", "a field that gates the cycle may be reset by a dynamic/templated value — cannot prove the loop terminates"
	}
	return "error", "every gating condition in the cycle is reset on each pass, so the loop re-arms indefinitely"
}

// classifyCascades is the pure entry point: builds the graph and emits the three-tier
// analysis, with blocking restricted to cycles that involve the candidate.
func classifyCascades(nodes []ruleNode, candidateID uuid.UUID) CascadeAnalysis {
	var analysis CascadeAnalysis

	nodeByID := make(map[uuid.UUID]ruleNode, len(nodes))
	ids := make([]uuid.UUID, 0, len(nodes))
	for _, n := range nodes {
		nodeByID[n.id] = n
		ids = append(ids, n.id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })

	edges := buildEdges(nodes)
	dAdj := edges.adjacency(edgeDefinite)
	pAdj := edges.adjacency(edgeIndeterminate)

	names := func(path []uuid.UUID) []string {
		out := make([]string, len(path))
		for i, id := range path {
			out[i] = nodeByID[id].name
		}
		return out
	}
	hasEdge := func(adj map[uuid.UUID][]uuid.UUID, from, to uuid.UUID) bool {
		return slices.Contains(adj[from], to)
	}
	isCycle := func(scc []uuid.UUID, adj map[uuid.UUID][]uuid.UUID) bool {
		if len(scc) > 1 {
			return true
		}
		return hasEdge(adj, scc[0], scc[0]) // self-loop
	}

	// --- Definite-edge SCCs: the provable-loop tier. ---
	candInDCycle := false
	for _, scc := range tarjanSCC(ids, dAdj) {
		if !isCycle(scc, dAdj) {
			continue
		}
		inSCC := map[uuid.UUID]bool{}
		containsCand := false
		for _, id := range scc {
			inSCC[id] = true
			if id == candidateID {
				containsCand = true
			}
		}

		tier, why := reArm(scc, nodeByID, dAdj)
		path := findCyclePath(scc[0], dAdj, inSCC)
		if containsCand {
			path = findCyclePath(candidateID, dAdj, inSCC) // anchor the path on the candidate
		}

		finding := CascadeFinding{RuleIDs: path, RuleNames: names(path)}

		if !containsCand {
			// A pre-existing loop not created by this save. Don't block the candidate, but
			// surface it as a warning safety-net (it slipped past the per-save check, e.g. seed).
			if tier == "error" {
				finding.Reason = "pre-existing cascade loop not involving this rule: " + why
				analysis.Warnings = append(analysis.Warnings, finding)
			}
			continue
		}

		candInDCycle = true
		finding.Reason = pathString(names(path)) + " — " + why
		switch tier {
		case "error":
			analysis.Errors = append(analysis.Errors, finding)
		case "warning":
			analysis.Warnings = append(analysis.Warnings, finding)
		default: // info: convergent cycle, allowed
			analysis.Info = append(analysis.Info, finding)
		}
	}

	// --- Possible-edge SCCs the candidate is in but that are NOT provable: WARN. ---
	if !candInDCycle {
		for _, scc := range tarjanSCC(ids, pAdj) {
			if !isCycle(scc, pAdj) {
				continue
			}
			inSCC := map[uuid.UUID]bool{}
			containsCand := false
			for _, id := range scc {
				inSCC[id] = true
				if id == candidateID {
					containsCand = true
				}
			}
			if !containsCand {
				continue
			}
			path := findCyclePath(candidateID, pAdj, inSCC)
			analysis.Warnings = append(analysis.Warnings, CascadeFinding{
				RuleIDs:   path,
				RuleNames: names(path),
				Reason:    pathString(names(path)) + " — a cascade loop is possible but its values are dynamic/templated, so it cannot be proven; the runtime guard will stop it if it occurs",
			})
		}
	}

	// --- INFO datapoints: every edge incident to the candidate (awareness, not a loop). ---
	for _, y := range nodes {
		if y.id == candidateID {
			continue
		}
		if k, ok := edges[candidateID][y.id]; ok {
			analysis.Info = append(analysis.Info, CascadeFinding{
				RuleIDs:   []uuid.UUID{candidateID, y.id},
				RuleNames: []string{nodeByID[candidateID].name, y.name},
				Reason:    fmt.Sprintf("this rule can trigger %q (%s cascade)", y.name, edgeKindLabel(k)),
			})
		}
		if k, ok := edges[y.id][candidateID]; ok {
			analysis.Info = append(analysis.Info, CascadeFinding{
				RuleIDs:   []uuid.UUID{y.id, candidateID},
				RuleNames: []string{y.name, nodeByID[candidateID].name},
				Reason:    fmt.Sprintf("%q can trigger this rule (%s cascade)", y.name, edgeKindLabel(k)),
			})
		}
	}

	return analysis
}

func edgeKindLabel(k edgeKind) string {
	if k == edgeDefinite {
		return "definite"
	}
	return "possible"
}

func pathString(names []string) string {
	if len(names) == 0 {
		return "(cycle)"
	}
	// Close the loop visually: A -> B -> A.
	return strings.Join(append(append([]string{}, names...), names[0]), " -> ")
}

// =============================================================================
// DB / registry assembly (the impure boundary)
// =============================================================================

// DetectCascadeLoops builds the inter-rule cascade graph over the currently-active rules
// overlaid with the candidate (the rule being saved/activated) and returns the three-tier
// analysis. Active-only scope (DESIGN §10): an inactive candidate is a no-op. The registry
// supplies each action's mutation manifest (GetEntityModifications).
func (b *Business) DetectCascadeLoops(ctx context.Context, reg *ActionRegistry, cand CandidateRule) (CascadeAnalysis, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.detectcascadeloops")
	defer span.End()

	// Inactive candidate (draft save / deactivation) cannot close a loop against the active set.
	if !cand.IsActive {
		return CascadeAnalysis{}, nil
	}
	if reg == nil {
		// No registry → no manifests → no edges to analyze. Fail-soft (don't block saves).
		return CascadeAnalysis{}, nil
	}

	views, err := b.QueryAutomationRulesView(ctx) // active-only (WHERE is_active = true)
	if err != nil {
		return CascadeAnalysis{}, fmt.Errorf("detectcascadeloops: query active rules: %w", err)
	}

	nodes := make([]ruleNode, 0, len(views)+1)
	for _, v := range views {
		if v.ID == cand.RuleID {
			continue // overlay: the candidate replaces its persisted version
		}
		n, err := b.nodeFromView(ctx, reg, v)
		if err != nil {
			return CascadeAnalysis{}, err
		}
		nodes = append(nodes, n)
	}

	candNode, err := b.candidateNode(ctx, reg, cand)
	if err != nil {
		return CascadeAnalysis{}, err
	}
	nodes = append(nodes, candNode)

	return classifyCascades(nodes, cand.RuleID), nil
}

// nodeFromView resolves an active rule view into a graph node (trigger + outgoing manifests).
func (b *Business) nodeFromView(ctx context.Context, reg *ActionRegistry, v AutomationRuleView) (ruleNode, error) {
	actions, err := b.QueryRoleActionsViewByRuleID(ctx, v.ID)
	if err != nil {
		return ruleNode{}, fmt.Errorf("detectcascadeloops: query actions for rule[%s]: %w", v.ID, err)
	}

	var mods []EntityModification
	for _, a := range actions {
		if !a.IsActive {
			continue
		}
		mods = append(mods, modsFor(reg, a.TemplateActionType, a.ActionConfig)...)
	}

	return ruleNode{
		id:     v.ID,
		name:   v.Name,
		entity: schemaStripTable(v.EntityName),
		event:  v.TriggerTypeName,
		conds:  parseConds(v.TriggerConditions),
		mods:   mods,
	}, nil
}

// candidateNode resolves the in-flight candidate into a graph node.
func (b *Business) candidateNode(ctx context.Context, reg *ActionRegistry, cand CandidateRule) (ruleNode, error) {
	entityName, err := b.entityTableName(ctx, cand.EntityID)
	if err != nil {
		return ruleNode{}, err
	}
	eventName, err := b.triggerTypeName(ctx, cand.TriggerTypeID)
	if err != nil {
		return ruleNode{}, err
	}

	var mods []EntityModification
	for _, a := range cand.Actions {
		mods = append(mods, modsFor(reg, a.ActionType, a.Config)...)
	}

	return ruleNode{
		id:          cand.RuleID,
		name:        cand.Name,
		isCandidate: true,
		entity:      schemaStripTable(entityName),
		event:       eventName,
		conds:       parseConds(cand.TriggerConditions),
		mods:        mods,
	}, nil
}

func (b *Business) entityTableName(ctx context.Context, id uuid.UUID) (string, error) {
	entities, err := b.QueryEntities(ctx)
	if err != nil {
		return "", fmt.Errorf("detectcascadeloops: query entities: %w", err)
	}
	for _, e := range entities {
		if e.ID == id {
			return e.Name, nil
		}
	}
	return "", nil // unresolved entity → no matching edges (fail-soft)
}

func (b *Business) triggerTypeName(ctx context.Context, id uuid.UUID) (string, error) {
	types, err := b.QueryTriggerTypes(ctx)
	if err != nil {
		return "", fmt.Errorf("detectcascadeloops: query trigger types: %w", err)
	}
	for _, t := range types {
		if t.ID == id {
			return t.Name, nil
		}
	}
	return "", nil // unresolved trigger type → no matching edges (fail-soft)
}

// modsFor returns the entity modifications an action handler declares for the given config.
func modsFor(reg *ActionRegistry, actionType string, config json.RawMessage) []EntityModification {
	handler, ok := reg.Get(actionType)
	if !ok {
		return nil
	}
	modifier, ok := handler.(EntityModifier)
	if !ok {
		return nil
	}
	return modifier.GetEntityModifications(config)
}

func parseConds(raw *json.RawMessage) []FieldCondition {
	if raw == nil || len(*raw) == 0 {
		return nil
	}
	var tc TriggerConditions
	if err := json.Unmarshal(*raw, &tc); err != nil {
		return nil
	}
	return tc.FieldConditions
}
