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

// ActionHandler defines the interface for action type handlers
type ActionHandler interface {
	Execute(ctx context.Context, config json.RawMessage, context ActionExecutionContext) (any, error)
	Validate(config json.RawMessage) error
	GetType() string
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
