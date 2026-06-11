package temporal

import (
	"context"
	"encoding/json"
	"slices"

	"github.com/google/uuid"
)

// =============================================================================
// Cascade lineage carrier (P1 runtime loop guard)
// =============================================================================
//
// A workflow action can write the database; that write fires a delegate event;
// that event can match *another* automation rule and dispatch a new workflow —
// a cascade. Without a guard, an A->B->A rule chain loops forever: Temporal
// gives zero incidental dedup (every dispatch gets a fresh executionID), and
// nothing else tracks "has this (rule, entity) already fired in this chain?".
//
// WorkflowLineage is that tracker. It carries a visited-set of (ruleID, entityID)
// pairs along a cascade chain. Before dispatching rule R for entity E, the trigger
// checks the set: if (R,E) is already present the chain has looped back and the
// dispatch is refused; otherwise (R,E) is added and the extended set rides the new
// workflow so the next hop can do the same check.
//
// Carrier path (the "chokepoint" — DESIGN §5/§10): the lineage rides
// WorkflowInput.TriggerData (Continue-As-New-safe). It flows for free into the
// activity via TriggerData -> NewMergedContext -> Flattened -> ActionActivityInput.Context.
// The activity stamps it onto the Go context (contextWithLineage); the handler's
// bus write propagates that context through delegate.Call into
// DelegateHandler.handleEvent, which reads it back (lineageFromContext) and seeds
// the next generation. This avoids threading a token through the 205 delegate.Call
// sites or changing the WorkflowInput struct shape.

// CascadeLineageKey is the reserved key under which the lineage rides
// WorkflowInput.TriggerData and (after flowing through MergedContext.Flattened)
// the action activity Context map. The "__" prefix keeps it clear of any user
// template variable. Exported so out-of-package tests (e.g. the actionhandlers
// real-Temporal round-trip) reference the one true key instead of a drifting copy.
const CascadeLineageKey = "__cascade_lineage"

// WorkflowLineage carries cross-rule cascade provenance along a cascade chain.
//
// It is intentionally extensible: the visited-set powers the P1 runtime loop
// guard today, and OriginatingExecutionID identifies the chain root. Future
// observability work (F8 — distributed tracing) can add fields such as a
// traceparent or correlation ID here without re-plumbing the carrier.
type WorkflowLineage struct {
	// Visited is the set of (ruleID, entityID) pairs already fired in this
	// cascade chain, encoded as "ruleID:entityID" strings. A slice of strings
	// (rather than a map or struct slice) round-trips cleanly through Temporal's
	// JSON serialization with no decode ambiguity.
	Visited []string `json:"visited,omitempty"`

	// OriginatingExecutionID is the execution that started this cascade chain.
	// Set once, on the first dispatched hop, and preserved across every
	// subsequent hop. Useful for correlating an entire chain (F8).
	OriginatingExecutionID uuid.UUID `json:"originating_execution_id,omitempty"`
}

// lineagePairKey encodes a (ruleID, entityID) pair as a stable string key.
func lineagePairKey(ruleID, entityID uuid.UUID) string {
	return ruleID.String() + ":" + entityID.String()
}

// Contains reports whether (ruleID, entityID) has already fired in this chain.
func (l WorkflowLineage) Contains(ruleID, entityID uuid.UUID) bool {
	return slices.Contains(l.Visited, lineagePairKey(ruleID, entityID))
}

// With returns a copy of the lineage extended with (ruleID, entityID).
// The receiver is never mutated, so a parent lineage can safely seed multiple
// child dispatches (multiple matched rules off one event).
func (l WorkflowLineage) With(ruleID, entityID uuid.UUID) WorkflowLineage {
	next := WorkflowLineage{
		Visited:                make([]string, len(l.Visited), len(l.Visited)+1),
		OriginatingExecutionID: l.OriginatingExecutionID,
	}
	copy(next.Visited, l.Visited)
	next.Visited = append(next.Visited, lineagePairKey(ruleID, entityID))
	return next
}

// =============================================================================
// Context transport
// =============================================================================

// lineageCtxKeyType is a private context key type, avoiding collisions with any
// other package's context values.
type lineageCtxKeyType struct{}

var lineageCtxKey = lineageCtxKeyType{}

// contextWithLineage returns a context carrying the cascade lineage. The action
// activity stamps it so the handler's bus write (and its delegate.Call) carry the
// lineage forward; DelegateHandler re-stamps it onto the dispatch context.
func contextWithLineage(ctx context.Context, l WorkflowLineage) context.Context {
	return context.WithValue(ctx, lineageCtxKey, l)
}

// lineageFromContext extracts the cascade lineage from a context, returning the
// zero value (empty set) when none is present — e.g. a human-originated write or
// a manually dispatched workflow, which correctly start a fresh chain.
func lineageFromContext(ctx context.Context) WorkflowLineage {
	if l, ok := ctx.Value(lineageCtxKey).(WorkflowLineage); ok {
		return l
	}
	return WorkflowLineage{}
}

// lineageFromContextMap extracts the cascade lineage from an action activity
// Context map (WorkflowInput.TriggerData -> MergedContext.Flattened ->
// ActionActivityInput.Context). The stored value may be a WorkflowLineage
// (same-process dispatch) or a map[string]any (after a Temporal JSON round-trip);
// both forms decode to the same lineage. Missing/nil/garbage degrade to the zero
// value (empty set).
func lineageFromContextMap(m map[string]any) WorkflowLineage {
	v, ok := m[CascadeLineageKey]
	if !ok || v == nil {
		return WorkflowLineage{}
	}

	if l, ok := v.(WorkflowLineage); ok {
		return l
	}

	// JSON round-trip form (Temporal deserialized TriggerData): re-marshal the
	// generic value and decode it into the struct.
	data, err := json.Marshal(v)
	if err != nil {
		return WorkflowLineage{}
	}
	var l WorkflowLineage
	if err := json.Unmarshal(data, &l); err != nil {
		return WorkflowLineage{}
	}
	return l
}
