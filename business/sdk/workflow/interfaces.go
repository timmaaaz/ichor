package workflow

import (
	"context"
	"encoding/json"
)

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
