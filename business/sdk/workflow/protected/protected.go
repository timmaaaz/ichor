// Package protected holds the registry of (entity, field) pairs that the generic
// raw-SQL workflow actions (update_field, create_entity, transition_status) must
// NOT write directly. Those fields carry business-layer invariants (state-machine
// guards, ledger append-only-ness) that live in typed action-verb bus methods
// (Approve/Reject/Claim/Execute/quantity verbs) and are bypassed by raw SQL.
//
// The registry is domain-declared and collected at startup (the delegate.Register
// pattern): the set of protected fields is the union of (1) fields a typed action
// claims via its GetEntityModifications manifest and (2) domain-model-declared
// protected fields (a `protected:"true"` struct tag) with no typed action yet.
// There is no central hand-maintained field list.
package protected

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// ErrProtectedField is the sentinel returned (wrapped) when a generic workflow
// action attempts to write a protected (entity, field). Callers use errors.Is to
// detect it; the app layer maps it to a 400-class response.
var ErrProtectedField = errors.New("workflow: protected field")

// Registry maps protected entities/fields to the typed action-verb that legitimately
// writes them (the route hint, empty when no typed action exists yet). It is populated
// once at startup and read-only thereafter (no locking on the read path).
type Registry struct {
	// fields[entity][field] = routeAction ("" when there is no typed action yet).
	fields map[string]map[string]string
	// entities[entity] = routeAction — whole-table protection (e.g. an append-only ledger).
	entities map[string]string
}

// New constructs an empty registry. An empty registry protects nothing.
func New() *Registry {
	return &Registry{
		fields:   make(map[string]map[string]string),
		entities: make(map[string]string),
	}
}

// ProtectField marks (entity, field) as protected, routing legitimate writes to
// routeAction (empty when no typed action exists). Re-registering keeps the first
// non-empty route so an auto-source (manifest) route is not clobbered by an empty one.
func (r *Registry) ProtectField(entity, field, routeAction string) {
	if entity == "" || field == "" {
		return
	}
	fm := r.fields[entity]
	if fm == nil {
		fm = make(map[string]string)
		r.fields[entity] = fm
	}
	if existing, ok := fm[field]; ok && existing != "" && routeAction == "" {
		return
	}
	fm[field] = routeAction
}

// ProtectEntity marks an entire entity as protected — every field is blocked.
func (r *Registry) ProtectEntity(entity, routeAction string) {
	if entity == "" {
		return
	}
	if existing, ok := r.entities[entity]; ok && existing != "" && routeAction == "" {
		return
	}
	r.entities[entity] = routeAction
}

// Check reports whether (entity, field) is protected and, if so, the route hint.
// Whole-table protection takes precedence over a field-specific route.
func (r *Registry) Check(entity, field string) (route string, blocked bool) {
	if route, ok := r.entities[entity]; ok {
		return route, true
	}
	if fm, ok := r.fields[entity]; ok {
		if route, ok := fm[field]; ok {
			return route, true
		}
	}
	return "", false
}

// Entry is one registered protection, returned by Entries for read-only introspection.
// Field is empty ("") for a whole-table protection (every column blocked).
type Entry struct {
	Entity string
	Field  string // "" → whole-table protection
	Route  string
}

// Entries returns every registered protection (whole-table and field-level) as a flat slice.
// It exists for verification/introspection — e.g. asserting each protected column resolves to a
// real DB column — and is NOT used on the enforcement hot path. Order is unspecified.
func (r *Registry) Entries() []Entry {
	out := make([]Entry, 0, len(r.entities)+len(r.fields))
	for entity, route := range r.entities {
		out = append(out, Entry{Entity: entity, Field: "", Route: route})
	}
	for entity, fm := range r.fields {
		for field, route := range fm {
			out = append(out, Entry{Entity: entity, Field: field, Route: route})
		}
	}
	return out
}

// CollectStructTags registers every field of model that carries a `protected:"true"`
// struct tag, keyed by the field's `db` tag — the authoritative column name the generic
// raw-SQL handlers target. (The db store model is the source of truth: bus-model json
// tags do NOT always equal column names, e.g. orders.fulfillment_status_id maps to the
// column order_fulfillment_status_id.) Fields without a usable db tag (missing or "-")
// are skipped. model may be a struct or a pointer to one.
func CollectStructTags(r *Registry, entity, routeAction string, model any) {
	t := reflect.TypeOf(model)
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if strings.ToLower(f.Tag.Get("protected")) != "true" {
			continue
		}
		col := strings.SplitN(f.Tag.Get("db"), ",", 2)[0]
		if col == "" || col == "-" {
			continue
		}
		r.ProtectField(entity, col, routeAction)
	}
}

// NewError builds a clear, wrapped ErrProtectedField for a blocked (entity, field).
// When routeAction is non-empty it names the typed action to use instead.
func NewError(entity, field, routeAction string) error {
	base := fmt.Sprintf("field %q on %q is protected and cannot be set by generic workflow actions", field, entity)
	if routeAction != "" {
		base += fmt.Sprintf("; use the %q action", routeAction)
	}
	return fmt.Errorf("%s: %w", base, ErrProtectedField)
}
