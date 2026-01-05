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
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
	ws "github.com/timmaaaz/ichor/foundation/websocket"
)

// =============================================================================
// Handler Tests - Additional coverage for alerthub.go
// =============================================================================

func Test_AlertHub_Hub(t *testing.T) {
	// Test the Hub() accessor method
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST",
		func(context.Context) string { return otel.GetTraceID(context.Background()) })
	hub := ws.NewHub(log)

	db := dbtest.NewDatabase(t, "Test_AlertHub_Hub")
	alertHub := alertws.NewAlertHub(hub, db.BusDomain.UserRole, log)

	// Verify Hub() returns the correct hub
	if alertHub.Hub() != hub {
		t.Error("Hub() should return the underlying hub")
	}
}

func Test_RegisterClient_NoRoles(t *testing.T) {
	// Test RegisterClient with a user that has no roles
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST",
		func(context.Context) string { return otel.GetTraceID(context.Background()) })
	hub := ws.NewHub(log)

	db := dbtest.NewDatabase(t, "Test_RegisterClient_NoRoles")
	alertHub := alertws.NewAlertHub(hub, db.BusDomain.UserRole, log)

	// Start hub
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	// Create test server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}

		userID := uuid.New() // Non-existent user - will have no roles
		client := ws.NewClient(hub, conn, log)

		// This should work - empty roles, but still registers under user ID
		err = alertHub.RegisterClient(r.Context(), client, userID)
		if err != nil {
			t.Logf("RegisterClient failed: %v", err)
			conn.Close(websocket.StatusInternalError, "registration failed")
			return
		}

		go client.WritePump(r.Context())
		client.ReadPump(r.Context())
	})

	server := httptest.NewServer(handler)
	defer server.Close()

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
}
