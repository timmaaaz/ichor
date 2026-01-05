package websocket

import "sync"

// HandlerRegistry manages WebSocket message handlers by type.
// It provides thread-safe registration and lookup of handlers.
type HandlerRegistry struct {
	handlers map[string]MessageHandler
	mu       sync.RWMutex
}

// NewHandlerRegistry creates a new handler registry.
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers: make(map[string]MessageHandler),
	}
}

// Register adds a handler to the registry.
// The handler's MessageType() is used as the key.
// If a handler for that type already exists, it is replaced.
func (r *HandlerRegistry) Register(h MessageHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[h.MessageType()] = h
}

// Get retrieves a handler by message type.
// Returns the handler and true if found, nil and false otherwise.
func (r *HandlerRegistry) Get(msgType string) (MessageHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.handlers[msgType]
	return h, ok
}

// Types returns all registered message types.
func (r *HandlerRegistry) Types() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.handlers))
	for t := range r.handlers {
		types = append(types, t)
	}
	return types
}
