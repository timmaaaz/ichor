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
// Async Action Handler Interface
// =============================================================================

// AsyncActionHandler extends ActionHandler for actions that queue work asynchronously.
// Execute() queues the work and returns immediately with tracking info.
// ProcessQueued() performs the actual async processing when dequeued.
//
// The handler is responsible for:
// - Deserializing the payload into its action-specific request type
// - Performing the async work
// - Firing result events via the publisher
type AsyncActionHandler interface {
	ActionHandler

	// ProcessQueued processes a queued message asynchronously.
	// payload contains the serialized QueuedPayload (use json.Unmarshal)
	// publisher is used to fire result events for downstream workflow rules
	ProcessQueued(ctx context.Context, payload json.RawMessage, publisher *EventPublisher) error
}

// QueuedPayload is a standard wrapper for async action payloads.
// All async handlers serialize their requests into this format when calling Execute(),
// and the queue manager passes this back to ProcessQueued().
type QueuedPayload struct {
	// RequestType identifies which handler should process this (e.g., "allocate_inventory")
	RequestType string `json:"request_type"`

	// RequestData contains the serialized action-specific request data
	RequestData json.RawMessage `json:"request_data"`

	// ExecutionContext contains workflow execution metadata
	ExecutionContext ActionExecutionContext `json:"execution_context"`

	// IdempotencyKey for deduplication
	IdempotencyKey string `json:"idempotency_key,omitempty"`
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
}
