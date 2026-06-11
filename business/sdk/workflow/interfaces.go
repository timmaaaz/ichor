package workflow

import (
	"context"
	"encoding/json"
)

// ActionHandler Registry Pattern - Type Safety vs Flexibility Trade-off
//
// This interface uses `any` as the return type for Execute, which sacrifices compile-time
// type safety for runtime flexibility. This is a fundamental limitation when building
// plugin/registry systems in Go where different handlers return different types.
//
// The Problem:
// - We need a single registry to hold all action handlers
// - Each handler may return a completely different type
// - Go's type system requires knowing types at compile time
// - Generic interfaces would require separate registries for each return type
//
// Common Solutions (all have trade-offs):
// 1. any/interface{} return (current approach) - Maximum flexibility, no compile-time safety
// 2. Generic interfaces - Type safe but requires multiple registries or type parameters everywhere
// 3. Common result wrapper - Type safe but requires unwrapping and marshaling overhead
// 4. Code generation - Type safe but adds build complexity
//
// Why we chose `any`:
// The workflow engine needs to handle arbitrary action types dynamically. The loss of type
// safety is localized to the framework boundary - individual handlers can still maintain
// internal type safety, and consumers should document their return types clearly.
//
// Best Practices:
// - Document your Execute return type clearly in comments
// - Consider adding a typed method that Execute calls internally
// - Validate types at runtime where results are consumed
// - Keep type assertions close to the framework boundary

// ActionHandler defines the interface for action type handlers.
// Handlers implement both automated workflow execution and manual execution capabilities.
type ActionHandler interface {
	// Execute performs the action with the given configuration and context.
	// Returns the result data (type varies by handler) and any error encountered.
	Execute(ctx context.Context, config json.RawMessage, context ActionExecutionContext) (any, error)

	// Validate validates the action configuration before execution.
	Validate(config json.RawMessage) error

	// GetType returns the unique identifier for this action type (e.g., "allocate_inventory").
	GetType() string

	// SupportsManualExecution returns true if this action can be triggered manually via API.
	// Returns false for actions like update_field that should only run via automation.
	SupportsManualExecution() bool

	// IsAsync returns true if this action queues work for async processing.
	// Async actions return immediately with tracking info; sync actions complete inline.
	IsAsync() bool

	// GetDescription returns a human-readable description for discovery APIs.
	GetDescription() string
}

// ActionRegistry manages action handlers
type ActionRegistry struct {
	handlers map[string]ActionHandler
}

// NewActionRegistry creates a new action registry
func NewActionRegistry() *ActionRegistry {
	return &ActionRegistry{
		handlers: make(map[string]ActionHandler),
	}
}

// Register adds a handler to the registry
func (ar *ActionRegistry) Register(handler ActionHandler) {
	ar.handlers[handler.GetType()] = handler
}

// Get retrieves a handler by type
func (ar *ActionRegistry) Get(actionType string) (ActionHandler, bool) {
	handler, exists := ar.handlers[actionType]
	return handler, exists
}

// GetAll returns all registered action types
func (ar *ActionRegistry) GetAll() []string {
	types := make([]string, 0, len(ar.handlers))
	for actionType := range ar.handlers {
		types = append(types, actionType)
	}
	return types
}

// =============================================================================
// Output Port Types (for hybrid output-based routing)
// =============================================================================

// OutputPort describes a named output from an action.
type OutputPort struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDefault   bool   `json:"is_default"` // True for the "happy path" output
}

// OutputPortProvider is implemented by action handlers that declare
// specific output ports beyond the default success/failure.
type OutputPortProvider interface {
	GetOutputPorts() []OutputPort
}

// DefaultOutputPorts returns the standard success/failure ports used by
// actions that don't implement OutputPortProvider.
func DefaultOutputPorts() []OutputPort {
	return []OutputPort{
		{Name: "success", Description: "Action completed successfully", IsDefault: true},
		{Name: "failure", Description: "Action failed"},
	}
}

// GetOutputPorts returns the output ports for the given action type.
// If the handler implements OutputPortProvider, its ports are returned.
// Otherwise, the default success/failure ports are returned.
func (ar *ActionRegistry) GetOutputPorts(actionType string) []OutputPort {
	handler, exists := ar.handlers[actionType]
	if !exists {
		return DefaultOutputPorts()
	}
	if provider, ok := handler.(OutputPortProvider); ok {
		return provider.GetOutputPorts()
	}
	return DefaultOutputPorts()
}

// =============================================================================
// Entity Modifier Interface (for Cascade Visualization)
// =============================================================================

// EntityModifier is an optional interface for action handlers that modify entities.
// Used for cascade visualization to determine which downstream workflows may trigger.
// Action handlers that modify entities should implement this interface to enable
// the cascade visualization API to show users what workflows will be triggered.
type EntityModifier interface {
	// GetEntityModifications returns what entities/fields this action modifies
	// based on the provided configuration. Returns nil if the action doesn't
	// modify entities (e.g., send_email, send_notification).
	GetEntityModifications(config json.RawMessage) []EntityModification
}

// EntityModification describes an entity modification caused by an action.
// Used to match against automation rules to determine which downstream
// workflows would be triggered when this action executes.
type EntityModification struct {
	// EntityName is the fully-qualified table name (e.g., "sales.orders", "inventory.inventory_items")
	EntityName string `json:"entity_name"`

	// EventType indicates what kind of event this modification triggers.
	// Valid values: "on_create", "on_update", "on_delete"
	EventType string `json:"event_type"`

	// Fields lists which fields are modified (optional, for on_update events).
	// Used for more precise cascade matching when rules have field-specific conditions.
	Fields []string `json:"fields,omitempty"`

	// Changes carries the statically-known produced VALUE for the fields this action
	// sets, enabling value-aware cascade-edge detection (does this modification's
	// output satisfy a downstream rule's trigger condition?). It is additive over
	// Fields: a field may appear in Fields with no Change (the value is not statically
	// knowable — dynamic/templated/computed), in which case the static detector must
	// treat the edge conservatively and defer to the runtime loop guard. Absence of a
	// Change for a field therefore means "indeterminate value", same as an explicit
	// Change with Indeterminate=true (the explicit form additionally records the operator).
	Changes []ProducedChange `json:"changes,omitempty"`
}

// Trigger / produced-change operators — the shared vocabulary used on BOTH sides of a
// cascade edge: matched by the trigger evaluator (see TriggerProcessor.evaluateFieldCondition)
// and declared by EntityModifier handlers via FieldChange.Operator. Static handlers that set
// a field to a fixed value emit OperatorChangedTo.
const (
	OperatorEquals      = "equals"
	OperatorNotEquals   = "not_equals"
	OperatorChangedFrom = "changed_from"
	OperatorChangedTo   = "changed_to"
	OperatorGreaterThan = "greater_than"
	OperatorLessThan    = "less_than"
	OperatorContains    = "contains"
	OperatorIn          = "in"
)

// ProducedChange describes a value an action produces for a single field, expressed in the
// SAME vocabulary as a trigger FieldCondition (FieldName/Operator/Value) so the static
// cascade detector can match a produced change against a downstream rule's trigger
// condition directly (condition-vs-condition). It is the manifest/produced counterpart to
// the runtime event delta type FieldChange (NewValue/OldValue) in models.go.
//
// Indeterminate marks a field the action writes with a value that is NOT statically
// knowable (dynamic/templated/computed, e.g. "{{trigger.x}}", a runtime user ID, or a
// computed quantity). When Indeterminate is true, Value is unset and the detector treats
// the edge conservatively, deferring to the runtime loop guard.
type ProducedChange struct {
	FieldName     string `json:"field_name"`
	Operator      string `json:"operator"`
	Value         any    `json:"value,omitempty"`
	Indeterminate bool   `json:"indeterminate,omitempty"`
}
