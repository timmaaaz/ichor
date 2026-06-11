package data

import "github.com/timmaaaz/ichor/business/sdk/workflow/protected"

// Option configures a generic data action handler (update_field, create_entity,
// transition_status) at construction time.
type Option func(*options)

type options struct {
	protected *protected.Registry
}

// WithProtectedRegistry injects the protected-field registry so the handler rejects
// generic writes to fields that carry business-layer invariants (DESIGN §10 — the
// protected-list). A nil registry (the default) disables enforcement, preserving the
// previous behavior for callers that do not wire it.
func WithProtectedRegistry(r *protected.Registry) Option {
	return func(o *options) { o.protected = r }
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
