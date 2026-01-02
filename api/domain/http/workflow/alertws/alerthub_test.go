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
	ws "github.com/timmaaaz/ichor/foundation/websocket"
)

// =============================================================================
// AlertHub Integration Tests
// =============================================================================

// SeedData holds test users and roles
type AlertHubSeedData struct {
	Users []userbus.User
	Roles []rolebus.Role
}

// insertSeedData creates test users with roles for AlertHub testing
func insertSeedData(busDomain dbtest.BusDomain) (AlertHubSeedData, error) {
	ctx := context.Background()

	// Create test users
	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 3, userbus.Roles.User, busDomain.User)
	if err != nil {
		return AlertHubSeedData{}, err
	}

	// Create test roles
	roles, err := rolebus.TestSeedRoles(ctx, 2, busDomain.Role)
	if err != nil {
		return AlertHubSeedData{}, err
	}

	// Assign roles to users
	// User 0: role 0
	// User 1: role 0, role 1
	// User 2: no roles
	_, err = busDomain.UserRole.Create(ctx, userrolebus.NewUserRole{
		UserID: users[0].ID,
		RoleID: roles[0].ID,
	})
	if err != nil {
		return AlertHubSeedData{}, err
	}

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

// AlertHubTestServer wraps httptest.Server with AlertHub for integration testing
type AlertHubTestServer struct {
	Server   *httptest.Server
	Hub      *ws.Hub
	AlertHub *alertws.AlertHub
	Log      *logger.Logger
}

// NewAlertHubTestServer creates a test server with real UserRoleBus
func NewAlertHubTestServer(t *testing.T, busDomain dbtest.BusDomain) *AlertHubTestServer {
	t.Helper()

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST",
		func(context.Context) string { return otel.GetTraceID(context.Background()) })
	hub := ws.NewHub(log)
	alertHub := alertws.NewAlertHub(hub, busDomain.UserRole, log)

	// Start hub in background
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Logf("WebSocket upgrade failed: %v", err)
			return
		}

		// Get user ID from query param (simulating auth)
		userIDStr := r.URL.Query().Get("user_id")
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			t.Logf("Invalid user ID: %v", err)
			conn.Close(websocket.StatusInternalError, "invalid user id")
			return
		}

		client := ws.NewClient(hub, conn, log)

		// Register via AlertHub (fetches roles)
		if err := alertHub.RegisterClient(r.Context(), client, userID); err != nil {
			t.Logf("Registration failed: %v", err)
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

	return &AlertHubTestServer{
		Server:   server,
		Hub:      hub,
		AlertHub: alertHub,
		Log:      log,
	}
}

// ConnectClient connects a WebSocket client for the given user
func (ts *AlertHubTestServer) ConnectClient(t *testing.T, userID uuid.UUID) *websocket.Conn {
	t.Helper()

	wsURL := "ws" + ts.Server.URL[4:] + "?user_id=" + userID.String()
	conn, _, err := websocket.Dial(context.Background(), wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	return conn
}

// waitForCondition polls until condition is true or timeout
func waitForCondition(t *testing.T, timeout time.Duration, condition func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func Test_AlertHub(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_AlertHub")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	t.Run("register-client-with-roles", func(t *testing.T) {
		testRegisterClientWithRoles(t, db.BusDomain, sd)
	})
	t.Run("register-client-no-roles", func(t *testing.T) {
		testRegisterClientNoRoles(t, db.BusDomain, sd)
	})
	t.Run("broadcast-to-user", func(t *testing.T) {
		testBroadcastToUser(t, db.BusDomain, sd)
	})
	t.Run("broadcast-to-role", func(t *testing.T) {
		testBroadcastToRole(t, db.BusDomain, sd)
	})
	t.Run("broadcast-all", func(t *testing.T) {
		testBroadcastAll(t, db.BusDomain, sd)
	})
	t.Run("user-isolation", func(t *testing.T) {
		testUserIsolation(t, db.BusDomain, sd)
	})
	t.Run("role-targeting", func(t *testing.T) {
		testRoleTargeting(t, db.BusDomain, sd)
	})
	t.Run("refresh-user-roles", func(t *testing.T) {
		testRefreshUserRoles(t, db.BusDomain, sd)
	})
}

func testRegisterClientWithRoles(t *testing.T, busDomain dbtest.BusDomain, sd AlertHubSeedData) {
	ts := NewAlertHubTestServer(t, busDomain)

	// User with roles
	user := sd.Users[0] // Has 1 role
	conn := ts.ConnectClient(t, user.ID)
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Wait for registration
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 1
	}) {
		t.Fatal("Timeout waiting for registration")
	}

	// Should be registered under user ID
	userIDStr := "user:" + user.ID.String()
	if ts.Hub.ClientsForID(userIDStr) != 1 {
		t.Errorf("Expected 1 client for user ID")
	}

	// Should be registered under role ID
	roleIDStr := "role:" + sd.Roles[0].ID.String()
	if ts.Hub.ClientsForID(roleIDStr) != 1 {
		t.Errorf("Expected 1 client for role ID")
	}
}

func testRegisterClientNoRoles(t *testing.T, busDomain dbtest.BusDomain, sd AlertHubSeedData) {
	ts := NewAlertHubTestServer(t, busDomain)

	// User without roles
	user := sd.Users[2]
	conn := ts.ConnectClient(t, user.ID)
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Wait for registration
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 1
	}) {
		t.Fatal("Timeout waiting for registration")
	}

	// Should be registered only under user ID
	userIDStr := "user:" + user.ID.String()
	if ts.Hub.ClientsForID(userIDStr) != 1 {
		t.Errorf("Expected 1 client for user ID")
	}
}

func testBroadcastToUser(t *testing.T, busDomain dbtest.BusDomain, sd AlertHubSeedData) {
	ts := NewAlertHubTestServer(t, busDomain)

	user := sd.Users[0]
	conn := ts.ConnectClient(t, user.ID)
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Wait for registration
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 1
	}) {
		t.Fatal("Timeout waiting for registration")
	}

	// Broadcast to user
	testMsg := []byte(`{"type":"alert","target":"user"}`)
	delivered := ts.AlertHub.BroadcastToUser(user.ID, testMsg)

	if delivered != 1 {
		t.Errorf("Expected 1 delivery, got %d", delivered)
	}

	// Read message
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
		t.Fatal("Timeout waiting for message")
	}
}

func testBroadcastToRole(t *testing.T, busDomain dbtest.BusDomain, sd AlertHubSeedData) {
	ts := NewAlertHubTestServer(t, busDomain)

	// Connect user with role
	user := sd.Users[0]
	conn := ts.ConnectClient(t, user.ID)
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Wait for registration
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 1
	}) {
		t.Fatal("Timeout waiting for registration")
	}

	// Broadcast to role
	testMsg := []byte(`{"type":"alert","target":"role"}`)
	roleID := sd.Roles[0].ID
	delivered := ts.AlertHub.BroadcastToRole(roleID, testMsg)

	if delivered != 1 {
		t.Errorf("Expected 1 delivery, got %d", delivered)
	}

	// Read message
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
		t.Fatal("Timeout waiting for message")
	}
}

func testBroadcastAll(t *testing.T, busDomain dbtest.BusDomain, sd AlertHubSeedData) {
	ts := NewAlertHubTestServer(t, busDomain)

	// Connect multiple users
	conn1 := ts.ConnectClient(t, sd.Users[0].ID)
	defer conn1.Close(websocket.StatusNormalClosure, "")

	conn2 := ts.ConnectClient(t, sd.Users[1].ID)
	defer conn2.Close(websocket.StatusNormalClosure, "")

	// Wait for registrations
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 2
	}) {
		t.Fatalf("Timeout waiting for registrations, got %d", ts.Hub.ConnectionCount())
	}

	// Broadcast to all
	testMsg := []byte(`{"type":"broadcast"}`)
	delivered := ts.AlertHub.BroadcastAll(testMsg)

	if delivered != 2 {
		t.Errorf("Expected 2 deliveries, got %d", delivered)
	}
}

func testUserIsolation(t *testing.T, busDomain dbtest.BusDomain, sd AlertHubSeedData) {
	ts := NewAlertHubTestServer(t, busDomain)

	// Connect two users
	user1 := sd.Users[0]
	user2 := sd.Users[2]

	conn1 := ts.ConnectClient(t, user1.ID)
	defer conn1.Close(websocket.StatusNormalClosure, "")

	conn2 := ts.ConnectClient(t, user2.ID)
	defer conn2.Close(websocket.StatusNormalClosure, "")

	// Wait for registrations
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 2
	}) {
		t.Fatalf("Timeout waiting for registrations, got %d", ts.Hub.ConnectionCount())
	}

	// Broadcast to user1 only
	testMsg := []byte(`{"type":"alert","for":"user1"}`)
	delivered := ts.AlertHub.BroadcastToUser(user1.ID, testMsg)

	if delivered != 1 {
		t.Errorf("Expected 1 delivery, got %d", delivered)
	}

	// User1 should receive
	received1 := make(chan []byte, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, msg, _ := conn1.Read(ctx)
		received1 <- msg
	}()

	// User2 should NOT receive
	received2 := make(chan []byte, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		_, msg, _ := conn2.Read(ctx)
		received2 <- msg
	}()

	select {
	case msg := <-received1:
		if string(msg) != string(testMsg) {
			t.Errorf("User1 expected %s, got %s", testMsg, msg)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for user1 message")
	}

	// User2 should timeout (not receive)
	select {
	case msg := <-received2:
		if len(msg) > 0 {
			t.Errorf("User2 should not have received message, got %s", msg)
		}
	case <-time.After(1 * time.Second):
		// Expected - user2 didn't receive
	}
}

func testRoleTargeting(t *testing.T, busDomain dbtest.BusDomain, sd AlertHubSeedData) {
	ts := NewAlertHubTestServer(t, busDomain)

	// User 0 and 1 both have role 0
	// User 2 has no roles
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

	// Broadcast to role 0
	testMsg := []byte(`{"type":"role-alert"}`)
	roleID := sd.Roles[0].ID
	delivered := ts.AlertHub.BroadcastToRole(roleID, testMsg)

	// Should deliver to user 0 and 1 (both have role 0)
	if delivered != 2 {
		t.Errorf("Expected 2 deliveries for role 0, got %d", delivered)
	}

	// User 2 should NOT receive (no roles)
	received2 := make(chan []byte, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		_, msg, _ := conn2.Read(ctx)
		received2 <- msg
	}()

	select {
	case msg := <-received2:
		if len(msg) > 0 {
			t.Errorf("User2 should not have received role message, got %s", msg)
		}
	case <-time.After(1 * time.Second):
		// Expected - user2 didn't receive
	}
}

func testRefreshUserRoles(t *testing.T, busDomain dbtest.BusDomain, sd AlertHubSeedData) {
	ts := NewAlertHubTestServer(t, busDomain)

	// Connect user with no roles
	user := sd.Users[2]
	conn := ts.ConnectClient(t, user.ID)
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Wait for registration
	if !waitForCondition(t, 2*time.Second, func() bool {
		return ts.Hub.ConnectionCount() == 1
	}) {
		t.Fatal("Timeout waiting for registration")
	}

	// Verify not registered under role 1
	roleIDStr := "role:" + sd.Roles[1].ID.String()
	if ts.Hub.ClientsForID(roleIDStr) != 0 {
		t.Errorf("User should not be under role 1 initially")
	}

	// Add role to user (directly via business layer)
	ctx := context.Background()
	_, err := busDomain.UserRole.Create(ctx, userrolebus.NewUserRole{
		UserID: user.ID,
		RoleID: sd.Roles[1].ID,
	})
	if err != nil {
		t.Fatalf("Failed to add role: %v", err)
	}

	// Refresh roles for user
	err = ts.AlertHub.RefreshUserRoles(ctx, user.ID)
	if err != nil {
		t.Fatalf("RefreshUserRoles failed: %v", err)
	}

	// Give time for update
	time.Sleep(50 * time.Millisecond)

	// Now user should be registered under role 1
	if ts.Hub.ClientsForID(roleIDStr) != 1 {
		t.Errorf("User should now be under role 1 after refresh")
	}

	// Clean up - remove the role we added
	userRoles, err := busDomain.UserRole.QueryByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to query user roles: %v", err)
	}
	for _, ur := range userRoles {
		if ur.RoleID == sd.Roles[1].ID {
			busDomain.UserRole.Delete(ctx, ur)
		}
	}
}
