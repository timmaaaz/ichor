package alertws_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

// =============================================================================
// End-to-End Tests
// These tests verify the complete alert delivery pipeline through:
// 1. WebSocket client connects with JWT authentication
// 2. Alert is published to RabbitMQ
// 3. Consumer processes the message and routes via AlertHub
// 4. WebSocket client receives the alert
// =============================================================================

// testE2EUserTargetedDelivery tests the complete alert delivery pipeline:
// 1. Connect WebSocket client with JWT auth
// 2. Publish alert to RabbitMQ targeting specific user
// 3. Verify alert is delivered to WebSocket client
// 4. Verify message format matches frontend expectations
func testE2EUserTargetedDelivery(t *testing.T, test *apitest.WSTest, sd AlertWSSeedData) {
	ctx := context.Background()

	// Connect WebSocket client with valid JWT
	conn := test.ConnectClient(t, sd.UserToken(0))
	defer conn.Close(websocket.StatusNormalClosure, "test complete")

	// Give time for connection to be fully registered and consumer to be ready.
	// The Hub/AlertHub are internal to the mux, we can't observe directly.
	// Use longer delay on first test as consumer may still be starting up.
	time.Sleep(500 * time.Millisecond)

	// Publish user-targeted alert via RabbitMQ
	alertID := uuid.New()
	alertPayload := map[string]interface{}{
		"alert": map[string]interface{}{
			"id":        alertID.String(),
			"type":      "order_update",
			"severity":  "medium",
			"title":     "Order Status Changed",
			"message":   "Order #12345 has been shipped",
			"entity_id": "order-uuid-here",
		},
	}

	alertMsg := &rabbitmq.Message{
		Type:       "alert",
		EntityName: "workflow.alerts",
		EntityID:   alertID,
		UserID:     sd.UserID(0),
		Payload:    alertPayload,
	}

	if err := test.WorkflowQueue.Publish(ctx, rabbitmq.QueueTypeAlert, alertMsg); err != nil {
		t.Fatalf("Failed to publish alert: %v", err)
	}

	// Wait for WebSocket message delivery
	readCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, msg, err := conn.Read(readCtx)
	if err != nil {
		t.Fatalf("Failed to receive alert: %v", err)
	}

	// Verify message content
	if len(msg) == 0 {
		t.Fatal("Received empty message")
	}

	// Parse message and verify structure
	var received map[string]interface{}
	if err := json.Unmarshal(msg, &received); err != nil {
		t.Fatalf("Failed to parse received message: %v", err)
	}

	// Verify message type
	msgType, ok := received["type"].(string)
	if !ok || msgType != "alert" {
		t.Errorf("Expected message type 'alert', got %v", received["type"])
	}

	// Verify payload contains alert data
	if received["payload"] == nil {
		t.Error("Expected payload in message")
	}

	// Verify timestamp is present
	if received["timestamp"] == nil {
		t.Error("Expected timestamp in message")
	}

	t.Logf("E2E user-targeted alert delivered successfully: %s", string(msg))
}

// testE2ERoleBasedDelivery tests role-based alert delivery end-to-end:
// 1. Connect multiple users with different roles
// 2. Publish role-targeted alert
// 3. Verify only users with the target role receive the alert
func testE2ERoleBasedDelivery(t *testing.T, test *apitest.WSTest, sd AlertWSSeedData) {
	ctx := context.Background()

	// Connect user 0 (has role 0)
	conn0 := test.ConnectClient(t, sd.UserToken(0))
	defer conn0.Close(websocket.StatusNormalClosure, "test complete")

	// Connect user 1 (has role 0 and role 1)
	conn1 := test.ConnectClient(t, sd.UserToken(1))
	defer conn1.Close(websocket.StatusNormalClosure, "test complete")

	// Connect user 2 (no roles)
	conn2 := test.ConnectClient(t, sd.UserToken(2))
	defer conn2.Close(websocket.StatusNormalClosure, "test complete")

	// Give time for connections to be fully registered
	time.Sleep(300 * time.Millisecond)

	// Publish role-targeted alert (targeting role 1, only user 1 has this)
	alertID := uuid.New()
	alertMsg := &rabbitmq.Message{
		Type:       "alert",
		EntityName: "workflow.alerts",
		EntityID:   alertID,
		Payload: map[string]interface{}{
			"role_id": sd.RoleID(1).String(),
			"alert": map[string]interface{}{
				"id":       alertID.String(),
				"type":     "role_notification",
				"severity": "low",
				"title":    "Role-Specific Alert",
				"message":  "This is for role 1 members only",
			},
		},
	}

	if err := test.WorkflowQueue.Publish(ctx, rabbitmq.QueueTypeAlert, alertMsg); err != nil {
		t.Fatalf("Failed to publish alert: %v", err)
	}

	// User 1 should receive (has role 1)
	received1 := make(chan []byte, 1)
	go func() {
		readCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, msg, _ := conn1.Read(readCtx)
		received1 <- msg
	}()

	// User 0 should NOT receive (doesn't have role 1)
	received0 := make(chan []byte, 1)
	go func() {
		readCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		_, msg, _ := conn0.Read(readCtx)
		received0 <- msg
	}()

	// User 2 should NOT receive (no roles)
	received2 := make(chan []byte, 1)
	go func() {
		readCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		_, msg, _ := conn2.Read(readCtx)
		received2 <- msg
	}()

	// Verify user 1 receives
	select {
	case msg := <-received1:
		if len(msg) == 0 {
			t.Error("User 1 received empty message")
		} else {
			// Verify it contains role notification content
			if !strings.Contains(string(msg), "role_notification") {
				t.Errorf("Expected role_notification in message, got: %s", string(msg))
			}
			t.Logf("User 1 received role alert: %s", string(msg))
		}
	case <-time.After(6 * time.Second):
		t.Fatal("Timeout waiting for user 1 message")
	}

	// Verify user 0 does NOT receive
	select {
	case msg := <-received0:
		if len(msg) > 0 {
			t.Errorf("User 0 should not receive role 1 alert, got: %s", string(msg))
		}
	case <-time.After(2 * time.Second):
		// Expected - user 0 doesn't have role 1
		t.Log("User 0 correctly did not receive role 1 alert")
	}

	// Verify user 2 does NOT receive
	select {
	case msg := <-received2:
		if len(msg) > 0 {
			t.Errorf("User 2 should not receive role alert, got: %s", string(msg))
		}
	case <-time.After(2 * time.Second):
		// Expected - user 2 has no roles
		t.Log("User 2 correctly did not receive role alert")
	}
}

// testE2EBroadcastDelivery tests that alerts without user_id or role_id
// are broadcast to all connected clients.
func testE2EBroadcastDelivery(t *testing.T, test *apitest.WSTest, sd AlertWSSeedData) {
	ctx := context.Background()

	// Connect all users with small delays between connections
	conn0 := test.ConnectClient(t, sd.UserToken(0))
	defer conn0.Close(websocket.StatusNormalClosure, "test complete")
	time.Sleep(100 * time.Millisecond)

	conn1 := test.ConnectClient(t, sd.UserToken(1))
	defer conn1.Close(websocket.StatusNormalClosure, "test complete")
	time.Sleep(100 * time.Millisecond)

	conn2 := test.ConnectClient(t, sd.UserToken(2))
	defer conn2.Close(websocket.StatusNormalClosure, "test complete")

	// Give time for all connections to be fully registered
	time.Sleep(500 * time.Millisecond)

	// Publish broadcast alert (no user_id, no role_id)
	alertID := uuid.New()
	alertMsg := &rabbitmq.Message{
		Type:       "alert",
		EntityName: "workflow.alerts",
		EntityID:   alertID,
		Payload: map[string]interface{}{
			"alert": map[string]interface{}{
				"id":      alertID.String(),
				"type":    "system_notification",
				"message": "Alert for everyone",
			},
		},
	}

	if err := test.WorkflowQueue.Publish(ctx, rabbitmq.QueueTypeAlert, alertMsg); err != nil {
		t.Fatalf("Failed to publish alert: %s", err)
	}

	// All users should receive
	receivedCount := 0
	for i, conn := range []*websocket.Conn{conn0, conn1, conn2} {
		received := make(chan []byte, 1)
		go func(c *websocket.Conn) {
			readCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, msg, _ := c.Read(readCtx)
			received <- msg
		}(conn)

		select {
		case msg := <-received:
			if len(msg) > 0 {
				receivedCount++
				t.Logf("User %d received broadcast: %s", i, string(msg))
			}
		case <-time.After(6 * time.Second):
			t.Errorf("Timeout waiting for user %d message", i)
		}
	}

	if receivedCount != 3 {
		t.Errorf("Expected 3 users to receive broadcast, got %d", receivedCount)
	}
}

// testE2EUserIsolation verifies that user-targeted alerts don't leak to other users.
func testE2EUserIsolation(t *testing.T, test *apitest.WSTest, sd AlertWSSeedData) {
	ctx := context.Background()

	// Connect two users
	conn0 := test.ConnectClient(t, sd.UserToken(0))
	defer conn0.Close(websocket.StatusNormalClosure, "test complete")

	conn2 := test.ConnectClient(t, sd.UserToken(2))
	defer conn2.Close(websocket.StatusNormalClosure, "test complete")

	// Give time for connections to be fully registered
	time.Sleep(300 * time.Millisecond)

	// Publish user-targeted alert (only for user 0)
	alertID := uuid.New()
	alertMsg := &rabbitmq.Message{
		Type:       "alert",
		EntityName: "workflow.alerts",
		EntityID:   alertID,
		UserID:     sd.UserID(0), // Target only user 0
		Payload: map[string]interface{}{
			"alert": map[string]interface{}{
				"id":      alertID.String(),
				"type":    "private_alert",
				"message": "This is for user 0 only",
			},
		},
	}

	if err := test.WorkflowQueue.Publish(ctx, rabbitmq.QueueTypeAlert, alertMsg); err != nil {
		t.Fatalf("Failed to publish alert: %v", err)
	}

	// User 0 should receive
	received0 := make(chan []byte, 1)
	go func() {
		readCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, msg, _ := conn0.Read(readCtx)
		received0 <- msg
	}()

	// User 2 should NOT receive
	received2 := make(chan []byte, 1)
	go func() {
		readCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		_, msg, _ := conn2.Read(readCtx)
		received2 <- msg
	}()

	// Verify user 0 receives
	select {
	case msg := <-received0:
		if len(msg) == 0 {
			t.Error("User 0 received empty message")
		} else {
			if !strings.Contains(string(msg), "private_alert") {
				t.Errorf("Expected private_alert in message, got: %s", string(msg))
			}
			t.Logf("User 0 received private alert: %s", string(msg))
		}
	case <-time.After(6 * time.Second):
		t.Fatal("Timeout waiting for user 0 message")
	}

	// Verify user 2 does NOT receive
	select {
	case msg := <-received2:
		if len(msg) > 0 {
			t.Errorf("User 2 should not receive user 0's alert, got: %s", string(msg))
		}
	case <-time.After(2 * time.Second):
		// Expected - user 2 should not receive user 0's message
		t.Log("User 2 correctly did not receive user 0's private alert")
	}
}
