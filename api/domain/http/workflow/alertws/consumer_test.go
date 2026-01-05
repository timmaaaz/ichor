package alertws_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/alertws"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
	ws "github.com/timmaaaz/ichor/foundation/websocket"
)

// =============================================================================
// Consumer Integration Tests
// =============================================================================

// ConsumerTestServer wraps test infrastructure for consumer testing
type ConsumerTestServer struct {
	Server   *httptest.Server
	Hub      *ws.Hub
	AlertHub *alertws.AlertHub
	Log      *logger.Logger
	Clients  map[uuid.UUID]*websocket.Conn
}

// NewConsumerTestServer creates test server for consumer testing
func NewConsumerTestServer(t *testing.T, busDomain dbtest.BusDomain) *ConsumerTestServer {
	t.Helper()

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST",
		func(context.Context) string { return otel.GetTraceID(context.Background()) })
	hub := ws.NewHub(log)
	alertHub := alertws.NewAlertHub(hub, busDomain.UserRole, log)

	// Start hub
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}

		userIDStr := r.URL.Query().Get("user_id")
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			conn.Close(websocket.StatusInternalError, "invalid user id")
			return
		}

		client := ws.NewClient(hub, conn, log)

		if err := alertHub.RegisterClient(r.Context(), client, userID); err != nil {
			conn.Close(websocket.StatusInternalError, "registration failed")
			return
		}

		go client.WritePump(r.Context())
		client.ReadPump(r.Context())
	})

	server := httptest.NewServer(handler)

	t.Cleanup(func() {
		cancel()
		server.Close()
	})

	return &ConsumerTestServer{
		Server:   server,
		Hub:      hub,
		AlertHub: alertHub,
		Log:      log,
		Clients:  make(map[uuid.UUID]*websocket.Conn),
	}
}

// ConnectClient connects a user and tracks the connection
func (ts *ConsumerTestServer) ConnectClient(t *testing.T, userID uuid.UUID) *websocket.Conn {
	t.Helper()

	wsURL := "ws" + ts.Server.URL[4:] + "?user_id=" + userID.String()
	conn, _, err := websocket.Dial(context.Background(), wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	ts.Clients[userID] = conn
	return conn
}

// insertConsumerSeedData creates test users with roles
func insertConsumerSeedData(busDomain dbtest.BusDomain) (AlertHubSeedData, error) {
	ctx := context.Background()

	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 3, userbus.Roles.User, busDomain.User)
	if err != nil {
		return AlertHubSeedData{}, err
	}

	roles, err := rolebus.TestSeedRoles(ctx, 2, busDomain.Role)
	if err != nil {
		return AlertHubSeedData{}, err
	}

	// User 0: role 0
	_, err = busDomain.UserRole.Create(ctx, userrolebus.NewUserRole{
		UserID: users[0].ID,
		RoleID: roles[0].ID,
	})
	if err != nil {
		return AlertHubSeedData{}, err
	}

	// User 1: role 0, role 1
	_, err = busDomain.UserRole.Create(ctx, userrolebus.NewUserRole{
		UserID: users[1].ID,
		RoleID: roles[0].ID,
	})
	if err != nil {
		return AlertHubSeedData{}, err
	}

	_, err = busDomain.UserRole.Create(ctx, userrolebus.NewUserRole{
		UserID: users[1].ID,
		RoleID: roles[1].ID,
	})
	if err != nil {
		return AlertHubSeedData{}, err
	}

	return AlertHubSeedData{
		Users: users,
		Roles: roles,
	}, nil
}

func Test_AlertConsumer(t *testing.T) {
	// Get RabbitMQ test container
	container := rabbitmq.GetTestContainer(t)

	// Create client and connect
	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer client.Close()

	// Create logger
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST",
		func(context.Context) string { return otel.GetTraceID(context.Background()) })

	// Initialize workflow queue
	wq := rabbitmq.NewWorkflowQueue(client, log)
	ctx := context.Background()
	if err := wq.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize queues: %s", err)
	}

	// Get database
	db := dbtest.NewDatabase(t, "Test_AlertConsumer")

	// Seed data
	sd, err := insertConsumerSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// Run tests
	t.Run("user-targeted", func(t *testing.T) {
		wq.PurgeQueue(ctx, rabbitmq.QueueTypeAlert)
		testUserTargetedConsumer(t, wq, db.BusDomain, sd)
	})
	t.Run("role-targeted", func(t *testing.T) {
		wq.PurgeQueue(ctx, rabbitmq.QueueTypeAlert)
		testRoleTargetedConsumer(t, wq, db.BusDomain, sd)
	})
	t.Run("broadcast-all", func(t *testing.T) {
		wq.PurgeQueue(ctx, rabbitmq.QueueTypeAlert)
		testBroadcastAllConsumer(t, wq, db.BusDomain, sd)
	})
	t.Run("malformed-payload", func(t *testing.T) {
		wq.PurgeQueue(ctx, rabbitmq.QueueTypeAlert)
		testMalformedPayload(t, wq, db.BusDomain, sd)
	})
}

func testUserTargetedConsumer(t *testing.T, wq *rabbitmq.WorkflowQueue, busDomain dbtest.BusDomain, sd AlertHubSeedData) {
	// Create test server
	ts := NewConsumerTestServer(t, busDomain)

	// Connect user
	user := sd.Users[0]
	conn := ts.ConnectClient(t, user.ID)
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Wait for registration
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 1
	}) {
		t.Fatal("Timeout waiting for registration")
	}

	// Create and start consumer
	consumer := alertws.NewAlertConsumer(ts.AlertHub, wq, ts.Log)
	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	defer consumerCancel()
	go consumer.Start(consumerCtx)

	// Give consumer time to start
	time.Sleep(100 * time.Millisecond)

	// Publish user-targeted alert
	ctx := context.Background()
	alertMsg := &rabbitmq.Message{
		Type:       "alert",
		EntityName: "workflow.alerts",
		EntityID:   uuid.New(),
		UserID:     user.ID, // Target specific user
		Payload: map[string]interface{}{
			"alert": map[string]interface{}{
				"type":     "test_alert",
				"severity": "high",
				"title":    "Test Alert",
				"message":  "This is a test alert",
			},
		},
	}

	if err := wq.Publish(ctx, rabbitmq.QueueTypeAlert, alertMsg); err != nil {
		t.Fatalf("Failed to publish alert: %s", err)
	}

	// Wait for message delivery
	received := make(chan []byte, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, msg, _ := conn.Read(ctx)
		received <- msg
	}()

	select {
	case msg := <-received:
		if len(msg) == 0 {
			t.Error("Received empty message")
		}
		// Verify it's a valid alert message
		if string(msg) == "" {
			t.Error("Message is empty")
		}
	case <-time.After(6 * time.Second):
		t.Fatal("Timeout waiting for alert message")
	}
}

func testRoleTargetedConsumer(t *testing.T, wq *rabbitmq.WorkflowQueue, busDomain dbtest.BusDomain, sd AlertHubSeedData) {
	ts := NewConsumerTestServer(t, busDomain)

	// Connect users with role 0
	conn0 := ts.ConnectClient(t, sd.Users[0].ID)
	defer conn0.Close(websocket.StatusNormalClosure, "")

	conn1 := ts.ConnectClient(t, sd.Users[1].ID)
	defer conn1.Close(websocket.StatusNormalClosure, "")

	// Connect user without role 0
	conn2 := ts.ConnectClient(t, sd.Users[2].ID)
	defer conn2.Close(websocket.StatusNormalClosure, "")

	// Wait for registrations
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 3
	}) {
		t.Fatalf("Timeout waiting for registrations, got %d", ts.Hub.ConnectionCount())
	}

	// Create and start consumer
	consumer := alertws.NewAlertConsumer(ts.AlertHub, wq, ts.Log)
	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	defer consumerCancel()
	go consumer.Start(consumerCtx)

	time.Sleep(100 * time.Millisecond)

	// Publish role-targeted alert
	ctx := context.Background()
	roleID := sd.Roles[0].ID
	alertMsg := &rabbitmq.Message{
		Type:       "alert",
		EntityName: "workflow.alerts",
		EntityID:   uuid.New(),
		Payload: map[string]interface{}{
			"role_id": roleID.String(), // Target role
			"alert": map[string]interface{}{
				"type":    "role_alert",
				"message": "Alert for role",
			},
		},
	}

	if err := wq.Publish(ctx, rabbitmq.QueueTypeAlert, alertMsg); err != nil {
		t.Fatalf("Failed to publish alert: %s", err)
	}

	// User 0 and 1 should receive (both have role 0)
	received0 := make(chan []byte, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, msg, _ := conn0.Read(ctx)
		received0 <- msg
	}()

	received1 := make(chan []byte, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, msg, _ := conn1.Read(ctx)
		received1 <- msg
	}()

	// User 2 should NOT receive
	received2 := make(chan []byte, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		_, msg, _ := conn2.Read(ctx)
		received2 <- msg
	}()

	// Check user 0 receives
	select {
	case msg := <-received0:
		if len(msg) == 0 {
			t.Error("User 0 received empty message")
		}
	case <-time.After(6 * time.Second):
		t.Fatal("Timeout waiting for user 0 message")
	}

	// Check user 1 receives
	select {
	case msg := <-received1:
		if len(msg) == 0 {
			t.Error("User 1 received empty message")
		}
	case <-time.After(6 * time.Second):
		t.Fatal("Timeout waiting for user 1 message")
	}

	// Check user 2 does NOT receive
	select {
	case msg := <-received2:
		if len(msg) > 0 {
			t.Errorf("User 2 should not receive role alert, got %s", msg)
		}
	case <-time.After(2 * time.Second):
		// Expected - user 2 doesn't have role 0
	}
}

func testBroadcastAllConsumer(t *testing.T, wq *rabbitmq.WorkflowQueue, busDomain dbtest.BusDomain, sd AlertHubSeedData) {
	ts := NewConsumerTestServer(t, busDomain)

	// Connect all users
	conn0 := ts.ConnectClient(t, sd.Users[0].ID)
	defer conn0.Close(websocket.StatusNormalClosure, "")

	conn1 := ts.ConnectClient(t, sd.Users[1].ID)
	defer conn1.Close(websocket.StatusNormalClosure, "")

	conn2 := ts.ConnectClient(t, sd.Users[2].ID)
	defer conn2.Close(websocket.StatusNormalClosure, "")

	// Wait for registrations
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 3
	}) {
		t.Fatalf("Timeout waiting for registrations, got %d", ts.Hub.ConnectionCount())
	}

	// Create and start consumer
	consumer := alertws.NewAlertConsumer(ts.AlertHub, wq, ts.Log)
	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	defer consumerCancel()
	go consumer.Start(consumerCtx)

	time.Sleep(100 * time.Millisecond)

	// Publish broadcast alert (no user_id, no role_id)
	ctx := context.Background()
	alertMsg := &rabbitmq.Message{
		Type:       "alert",
		EntityName: "workflow.alerts",
		EntityID:   uuid.New(),
		Payload: map[string]interface{}{
			"alert": map[string]interface{}{
				"type":    "broadcast_alert",
				"message": "Alert for everyone",
			},
		},
	}

	if err := wq.Publish(ctx, rabbitmq.QueueTypeAlert, alertMsg); err != nil {
		t.Fatalf("Failed to publish alert: %s", err)
	}

	// All users should receive
	receivedCount := 0
	for i, conn := range []*websocket.Conn{conn0, conn1, conn2} {
		received := make(chan []byte, 1)
		go func(c *websocket.Conn) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, msg, _ := c.Read(ctx)
			received <- msg
		}(conn)

		select {
		case msg := <-received:
			if len(msg) > 0 {
				receivedCount++
			}
		case <-time.After(6 * time.Second):
			t.Errorf("Timeout waiting for user %d message", i)
		}
	}

	if receivedCount != 3 {
		t.Errorf("Expected 3 users to receive broadcast, got %d", receivedCount)
	}
}

func testMalformedPayload(t *testing.T, wq *rabbitmq.WorkflowQueue, busDomain dbtest.BusDomain, sd AlertHubSeedData) {
	ts := NewConsumerTestServer(t, busDomain)

	// Connect user
	conn := ts.ConnectClient(t, sd.Users[0].ID)
	defer conn.Close(websocket.StatusNormalClosure, "")

	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 1
	}) {
		t.Fatal("Timeout waiting for registration")
	}

	// Create and start consumer
	consumer := alertws.NewAlertConsumer(ts.AlertHub, wq, ts.Log)
	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	defer consumerCancel()
	go consumer.Start(consumerCtx)

	time.Sleep(100 * time.Millisecond)

	// Publish message with invalid role_id
	ctx := context.Background()
	alertMsg := &rabbitmq.Message{
		Type:       "alert",
		EntityName: "workflow.alerts",
		EntityID:   uuid.New(),
		Payload: map[string]interface{}{
			"role_id": "invalid-uuid", // Invalid role ID
			"alert": map[string]interface{}{
				"message": "Bad role ID",
			},
		},
	}

	if err := wq.Publish(ctx, rabbitmq.QueueTypeAlert, alertMsg); err != nil {
		t.Fatalf("Failed to publish alert: %s", err)
	}

	// Consumer should handle error gracefully (log and not retry)
	// Give time for processing
	time.Sleep(500 * time.Millisecond)

	// Test passes if no panic occurred
}
