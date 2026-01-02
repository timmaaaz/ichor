package websocket_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
	ws "github.com/timmaaaz/ichor/foundation/websocket"
)

// =============================================================================
// Client Unit Tests
// =============================================================================

func TestClient_NewClient(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST",
		func(context.Context) string { return otel.GetTraceID(context.Background()) })
	hub := ws.NewHub(log)

	// Create test server for real WebSocket connection
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Logf("WebSocket upgrade failed: %v", err)
			return
		}

		client := ws.NewClient(hub, conn, log)
		if client == nil {
			t.Error("NewClient returned nil")
			return
		}

		// Verify initial state
		ids := client.IDs()
		if len(ids) != 0 {
			t.Errorf("Expected 0 initial IDs, got %d", len(ids))
		}

		// Close connection
		conn.Close(websocket.StatusNormalClosure, "test complete")
	}))
	defer server.Close()

	// Connect to trigger the handler
	wsURL := "ws" + server.URL[4:]
	conn, _, err := websocket.Dial(context.Background(), wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Give time for server handler to complete
	time.Sleep(50 * time.Millisecond)
}

func TestClient_IDs(t *testing.T) {
	ts := NewTestServer(t)

	// Connect client with specific IDs
	conn := ts.ConnectClient(t, "user:123", "role:admin")
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Wait for registration
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 1
	}) {
		t.Fatal("Timeout waiting for registration")
	}

	// Verify IDs are set (Hub side verification since we can't directly access client)
	if count := ts.Hub.ClientsForID("user:123"); count != 1 {
		t.Errorf("Expected client to be registered under user:123")
	}
	if count := ts.Hub.ClientsForID("role:admin"); count != 1 {
		t.Errorf("Expected client to be registered under role:admin")
	}
}

func TestClient_SetIDs(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST",
		func(context.Context) string { return otel.GetTraceID(context.Background()) })
	hub := ws.NewHub(log)

	// Create test server
	var serverClient *ws.Client
	clientReady := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Logf("WebSocket upgrade failed: %v", err)
			return
		}

		serverClient = ws.NewClient(hub, conn, log)

		// Set initial IDs
		serverClient.SetIDs([]string{"user:123", "role:admin"})

		// Verify IDs
		ids := serverClient.IDs()
		if len(ids) != 2 {
			t.Errorf("Expected 2 IDs, got %d", len(ids))
		}

		// Update IDs
		serverClient.SetIDs([]string{"user:123", "role:super-admin"})

		ids = serverClient.IDs()
		if len(ids) != 2 {
			t.Errorf("Expected 2 IDs after update, got %d", len(ids))
		}

		// Verify the ID change
		found := false
		for _, id := range ids {
			if id == "role:super-admin" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected role:super-admin in IDs")
		}

		close(clientReady)

		// Keep connection alive briefly
		time.Sleep(100 * time.Millisecond)
		conn.Close(websocket.StatusNormalClosure, "")
	}))
	defer server.Close()

	// Connect
	wsURL := "ws" + server.URL[4:]
	conn, _, err := websocket.Dial(context.Background(), wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Wait for server handler to complete
	select {
	case <-clientReady:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for client setup")
	}
}

func TestClient_Send(t *testing.T) {
	ts := NewTestServer(t)

	// Connect client
	conn := ts.ConnectClient(t, "user:123")
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Wait for registration
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 1
	}) {
		t.Fatal("Timeout waiting for registration")
	}

	// Send message via hub (which calls client.Send internally)
	testMsg := []byte(`{"type":"test","data":"hello"}`)
	delivered := ts.Hub.BroadcastToID("user:123", testMsg)

	if delivered != 1 {
		t.Errorf("Expected 1 delivery, got %d", delivered)
	}

	// Read the message from client side
	received := make(chan []byte, 1)
	errCh := make(chan error, 1)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, msg, err := conn.Read(ctx)
		if err != nil {
			errCh <- err
			return
		}
		received <- msg
	}()

	select {
	case msg := <-received:
		if string(msg) != string(testMsg) {
			t.Errorf("Expected %s, got %s", testMsg, msg)
		}
	case err := <-errCh:
		t.Fatalf("Failed to read message: %v", err)
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

func TestClient_Send_BufferFull(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST",
		func(context.Context) string { return otel.GetTraceID(context.Background()) })
	hub := ws.NewHub(log)

	// Create test server where we don't start the write pump
	// This simulates a slow client that can't keep up
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}

		client := ws.NewClient(hub, conn, log)
		hub.Register(r.Context(), client, []string{"user:slow"})

		// Don't start write pump - messages will queue up
		// This tests the buffer full scenario

		// Wait for test to send messages
		time.Sleep(500 * time.Millisecond)

		// Clean up
		hub.Unregister(r.Context(), client)
		conn.Close(websocket.StatusNormalClosure, "")
	}))
	defer server.Close()

	// Start hub
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	// Connect
	wsURL := "ws" + server.URL[4:]
	conn, _, err := websocket.Dial(context.Background(), wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Wait for registration
	if !waitForCondition(t, 2*time.Second, func() bool {
		return hub.ConnectionCount() == 1
	}) {
		t.Fatal("Timeout waiting for registration")
	}

	// Send many messages to fill the buffer (buffer size is 256)
	msg := []byte("test message")
	for i := 0; i < 300; i++ {
		hub.BroadcastToID("user:slow", msg)
	}

	// Test passes if no panic occurred
	// The messages beyond buffer capacity should be dropped with a log warning
}

func TestClient_Send_AfterClose(t *testing.T) {
	ts := NewTestServer(t)

	// Connect and immediately close
	conn := ts.ConnectClient(t, "user:123")

	// Wait for registration
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 1
	}) {
		t.Fatal("Timeout waiting for registration")
	}

	// Close the connection
	conn.Close(websocket.StatusNormalClosure, "")

	// Wait for unregistration
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 0
	}) {
		t.Fatal("Timeout waiting for unregistration")
	}

	// Try to send after close - should not panic
	delivered := ts.Hub.BroadcastToID("user:123", []byte("test"))
	if delivered != 0 {
		t.Errorf("Expected 0 deliveries after close, got %d", delivered)
	}
}

func TestClient_Close(t *testing.T) {
	ts := NewTestServer(t)

	// Connect client
	conn := ts.ConnectClient(t, "user:123")

	// Wait for registration
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 1
	}) {
		t.Fatal("Timeout waiting for registration")
	}

	// Close from client side
	err := conn.Close(websocket.StatusNormalClosure, "test close")
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}

	// Wait for unregistration
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 0
	}) {
		t.Fatal("Timeout waiting for unregistration after close")
	}
}

func TestClient_Close_Multiple(t *testing.T) {
	ts := NewTestServer(t)

	// Connect client
	conn := ts.ConnectClient(t, "user:123")

	// Wait for registration
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 1
	}) {
		t.Fatal("Timeout waiting for registration")
	}

	// Close multiple times - should not panic
	conn.Close(websocket.StatusNormalClosure, "first close")
	conn.Close(websocket.StatusNormalClosure, "second close")
	conn.Close(websocket.StatusNormalClosure, "third close")

	// Test passes if no panic occurred
}

func TestClient_ReadPump_NormalClose(t *testing.T) {
	ts := NewTestServer(t)

	// Connect client
	conn := ts.ConnectClient(t, "user:123")

	// Wait for registration
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 1
	}) {
		t.Fatal("Timeout waiting for registration")
	}

	// Close connection normally
	conn.Close(websocket.StatusNormalClosure, "normal close")

	// Verify client is unregistered (ReadPump exited and triggered Unregister)
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 0
	}) {
		t.Fatal("Timeout waiting for unregistration")
	}
}

func TestClient_WritePump_MessageDelivery(t *testing.T) {
	ts := NewTestServer(t)

	// Connect client
	conn := ts.ConnectClient(t, "user:123")
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Wait for registration
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 1
	}) {
		t.Fatal("Timeout waiting for registration")
	}

	// Send multiple messages
	messages := []string{
		`{"type":"msg1"}`,
		`{"type":"msg2"}`,
		`{"type":"msg3"}`,
	}

	for _, msg := range messages {
		ts.Hub.BroadcastToID("user:123", []byte(msg))
	}

	// Read all messages
	for i, expected := range messages {
		received := make(chan []byte, 1)
		errCh := make(chan error, 1)

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_, msg, err := conn.Read(ctx)
			if err != nil {
				errCh <- err
				return
			}
			received <- msg
		}()

		select {
		case msg := <-received:
			if string(msg) != expected {
				t.Errorf("Message %d: expected %s, got %s", i+1, expected, msg)
			}
		case err := <-errCh:
			t.Fatalf("Message %d: failed to read: %v", i+1, err)
		case <-time.After(3 * time.Second):
			t.Fatalf("Message %d: timeout waiting for message", i+1)
		}
	}
}

func TestClient_WritePump_Ping(t *testing.T) {
	// This test verifies the ping/pong mechanism works
	// The write pump sends pings periodically (every 54 seconds by default)
	// We can't easily test this without waiting 54 seconds, so we just
	// verify the connection stays alive during normal operation

	ts := NewTestServer(t)

	// Connect client
	conn := ts.ConnectClient(t, "user:123")
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Wait for registration
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 1
	}) {
		t.Fatal("Timeout waiting for registration")
	}

	// Connection should stay alive during message exchange
	testMsg := []byte(`{"type":"test"}`)
	delivered := ts.Hub.BroadcastToID("user:123", testMsg)

	if delivered != 1 {
		t.Errorf("Expected 1 delivery, got %d", delivered)
	}

	// Read message to confirm connection is healthy
	received := make(chan []byte, 1)
	errCh := make(chan error, 1)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, msg, err := conn.Read(ctx)
		if err != nil {
			errCh <- err
			return
		}
		received <- msg
	}()

	select {
	case <-received:
		// Success - connection is healthy
	case err := <-errCh:
		t.Fatalf("Connection error: %v", err)
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout - connection may be unhealthy")
	}
}

func TestClient_Lifecycle(t *testing.T) {
	ts := NewTestServer(t)

	// Connect
	conn := ts.ConnectClient(t, "user:lifecycle", "role:test")

	// Wait for registration
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 1
	}) {
		t.Fatal("Timeout waiting for registration")
	}

	// Verify connected under both IDs
	if ts.Hub.ClientsForID("user:lifecycle") != 1 {
		t.Error("Expected client registered under user:lifecycle")
	}
	if ts.Hub.ClientsForID("role:test") != 1 {
		t.Error("Expected client registered under role:test")
	}

	// Send and receive a message
	testMsg := []byte(`{"lifecycle":"test"}`)
	ts.Hub.BroadcastToID("user:lifecycle", testMsg)

	received := make(chan []byte, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, msg, _ := conn.Read(ctx)
		received <- msg
	}()

	select {
	case msg := <-received:
		if string(msg) != string(testMsg) {
			t.Errorf("Expected %s, got %s", testMsg, msg)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout receiving message")
	}

	// Close connection
	conn.Close(websocket.StatusNormalClosure, "lifecycle complete")

	// Wait for unregistration
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 0
	}) {
		t.Fatal("Timeout waiting for unregistration")
	}

	// Verify unregistered from both IDs
	if ts.Hub.ClientsForID("user:lifecycle") != 0 {
		t.Error("Expected client unregistered from user:lifecycle")
	}
	if ts.Hub.ClientsForID("role:test") != 0 {
		t.Error("Expected client unregistered from role:test")
	}
}
