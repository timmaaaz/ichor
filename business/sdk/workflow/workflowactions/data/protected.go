package data

import (
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/workflow/protected"
)

// EntityRef is the delegate target for a generically-written entity: the bus
// DomainName the synthesized event fires under and the bare entity name the
// trigger matches on. A reverse map keyed by the schema-qualified table
// (e.g. "inventory.inventory_items" → {Domain:"inventoryitem", Entity:"inventory_items"})
// is injected into the generic handlers so an M1 write can announce itself on the
// same channel a real bus write would (DESIGN §6 / P4 §E.2).
type EntityRef struct {
	Domain string
	Entity string
}

// Option configures a generic data action handler (update_field, create_entity,
// transition_status) at construction time.
type Option func(*options)

type options struct {
	protected *protected.Registry
	delegate  *delegate.Delegate
	entityMap map[string]EntityRef
}

// WithProtectedRegistry injects the protected-field registry so the handler rejects
// generic writes to fields that carry business-layer invariants (DESIGN §10 — the
// protected-list). A nil registry (the default) disables enforcement, preserving the
// previous behavior for callers that do not wire it.
func WithProtectedRegistry(r *protected.Registry) Option {
	return func(o *options) { o.protected = r }
}

// WithDelegate injects the delegate the handler fires a synthesized event on after a
// successful raw-SQL write, so the write cascades to downstream automation (P4 M1).
// A nil delegate (the default) disables synthesis, preserving cascades-OFF behavior
// for callers that do not wire it (e.g. RegisterCoreActions, unit tests).
func WithDelegate(d *delegate.Delegate) Option {
	return func(o *options) { o.delegate = d }
}

// WithEntityRegistry injects the reverse map (schema-qualified table → EntityRef) the
// handler uses to resolve which delegate domain + bare entity name to fire under. A
// target absent from the map degrades safely: the write succeeds but no event fires
// (logged). Wired only alongside WithDelegate.
func WithEntityRegistry(m map[string]EntityRef) Option {
	return func(o *options) { o.entityMap = m }
}

func newOptions(opts []Option) options {
	var o options
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

// checkProtectedField returns a wrapped protected.ErrProtectedField when (entity, field)
// is protected, and nil when the registry is unset or the field is writable.
func checkProtectedField(reg *protected.Registry, entity, field string) error {
	if reg == nil {
		return nil
	}
	if route, blocked := reg.Check(entity, field); blocked {
		return protected.NewError(entity, field, route)
	}
	return nil
}

// checkProtectedEntity guards a create against whole-table protection and against any
// protected field present in the payload.
func checkProtectedEntity(reg *protected.Registry, entity string, fields []string) error {
	if reg == nil {
		return nil
	}
	// Whole-table protection (e.g. an append-only ledger) — Check ignores the field.
	if route, blocked := reg.Check(entity, ""); blocked {
		return protected.NewError(entity, "", route)
	}
	for _, f := range fields {
		if route, blocked := reg.Check(entity, f); blocked {
			return protected.NewError(entity, f, route)
		}
	}
	return nil
}
