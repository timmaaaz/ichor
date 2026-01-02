package websocket_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/timmaaaz/ichor/foundation/rabbitmq"
	"github.com/timmaaaz/ichor/foundation/websocket"
)

// =============================================================================
// Mock Handler for Testing
// =============================================================================

// mockHandler is a test implementation of MessageHandler.
type mockHandler struct {
	msgType     string
	handleFunc  func(ctx context.Context, msg *rabbitmq.Message) error
	handleCalls int
	mu          sync.Mutex
}

func newMockHandler(msgType string) *mockHandler {
	return &mockHandler{
		msgType: msgType,
		handleFunc: func(ctx context.Context, msg *rabbitmq.Message) error {
			return nil
		},
	}
}

func (m *mockHandler) MessageType() string {
	return m.msgType
}

func (m *mockHandler) HandleMessage(ctx context.Context, msg *rabbitmq.Message) error {
	m.mu.Lock()
	m.handleCalls++
	m.mu.Unlock()
	return m.handleFunc(ctx, msg)
}

func (m *mockHandler) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.handleCalls
}

// =============================================================================
// Registry Tests
// =============================================================================

func TestHandlerRegistry_RegisterAndGet(t *testing.T) {
	registry := websocket.NewHandlerRegistry()

	// Register a handler
	handler := newMockHandler("alert")
	registry.Register(handler)

	// Get the handler
	got, ok := registry.Get("alert")
	if !ok {
		t.Fatal("Expected to find handler for 'alert'")
	}
	if got.MessageType() != "alert" {
		t.Errorf("Expected MessageType 'alert', got '%s'", got.MessageType())
	}
}

func TestHandlerRegistry_GetNotFound(t *testing.T) {
	registry := websocket.NewHandlerRegistry()

	// Try to get a non-existent handler
	got, ok := registry.Get("nonexistent")
	if ok {
		t.Error("Expected not to find handler for 'nonexistent'")
	}
	if got != nil {
		t.Error("Expected nil handler for 'nonexistent'")
	}
}

func TestHandlerRegistry_RegisterReplaces(t *testing.T) {
	registry := websocket.NewHandlerRegistry()

	// Register first handler
	handler1 := newMockHandler("alert")
	handler1.handleFunc = func(ctx context.Context, msg *rabbitmq.Message) error {
		return errors.New("handler1")
	}
	registry.Register(handler1)

	// Register second handler with same type
	handler2 := newMockHandler("alert")
	handler2.handleFunc = func(ctx context.Context, msg *rabbitmq.Message) error {
		return errors.New("handler2")
	}
	registry.Register(handler2)

	// Get the handler - should be handler2
	got, ok := registry.Get("alert")
	if !ok {
		t.Fatal("Expected to find handler for 'alert'")
	}

	err := got.HandleMessage(context.Background(), &rabbitmq.Message{})
	if err == nil || err.Error() != "handler2" {
		t.Errorf("Expected error 'handler2', got '%v'", err)
	}
}

func TestHandlerRegistry_MultipleTypes(t *testing.T) {
	registry := websocket.NewHandlerRegistry()

	// Register multiple handlers
	alertHandler := newMockHandler("alert")
	inventoryHandler := newMockHandler("inventory")
	orderHandler := newMockHandler("order")

	registry.Register(alertHandler)
	registry.Register(inventoryHandler)
	registry.Register(orderHandler)

	// Verify all are retrievable
	testCases := []string{"alert", "inventory", "order"}
	for _, msgType := range testCases {
		got, ok := registry.Get(msgType)
		if !ok {
			t.Errorf("Expected to find handler for '%s'", msgType)
			continue
		}
		if got.MessageType() != msgType {
			t.Errorf("Expected MessageType '%s', got '%s'", msgType, got.MessageType())
		}
	}
}

func TestHandlerRegistry_Types(t *testing.T) {
	registry := websocket.NewHandlerRegistry()

	// Register handlers
	registry.Register(newMockHandler("alert"))
	registry.Register(newMockHandler("inventory"))
	registry.Register(newMockHandler("order"))

	// Get all types
	types := registry.Types()
	if len(types) != 3 {
		t.Errorf("Expected 3 types, got %d", len(types))
	}

	// Verify all expected types are present
	typeSet := make(map[string]bool)
	for _, t := range types {
		typeSet[t] = true
	}

	expected := []string{"alert", "inventory", "order"}
	for _, e := range expected {
		if !typeSet[e] {
			t.Errorf("Expected type '%s' not found in Types()", e)
		}
	}
}

func TestHandlerRegistry_ConcurrentAccess(t *testing.T) {
	registry := websocket.NewHandlerRegistry()

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent registrations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			handler := newMockHandler("concurrent")
			registry.Register(handler)
		}(i)
	}

	// Concurrent reads while writing
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			registry.Get("concurrent")
			registry.Types()
		}()
	}

	wg.Wait()

	// Should have exactly one handler for "concurrent" (last one wins)
	_, ok := registry.Get("concurrent")
	if !ok {
		t.Error("Expected to find handler for 'concurrent' after concurrent access")
	}
}

func TestHandlerRegistry_HandleMessage(t *testing.T) {
	registry := websocket.NewHandlerRegistry()

	handler := newMockHandler("alert")
	registry.Register(handler)

	// Get and call handler
	got, _ := registry.Get("alert")
	msg := &rabbitmq.Message{Type: "alert"}

	err := got.HandleMessage(context.Background(), msg)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if handler.CallCount() != 1 {
		t.Errorf("Expected 1 call to HandleMessage, got %d", handler.CallCount())
	}
}

func TestHandlerRegistry_EmptyRegistry(t *testing.T) {
	registry := websocket.NewHandlerRegistry()

	// Types should return empty slice, not nil
	types := registry.Types()
	if types == nil {
		t.Error("Expected non-nil slice from Types() on empty registry")
	}
	if len(types) != 0 {
		t.Errorf("Expected 0 types, got %d", len(types))
	}
}
