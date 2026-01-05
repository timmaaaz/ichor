package apitest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coder/websocket"
	authbuild "github.com/timmaaaz/ichor/api/cmd/services/auth/build/all"
	ichorbuild "github.com/timmaaaz/ichor/api/cmd/services/ichor/build/all"
	"github.com/timmaaaz/ichor/api/sdk/http/mux"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

// WSTest extends Test with WebSocket infrastructure for integration testing.
// Unlike unit tests that create their own Hub, E2E tests go through the full
// HTTP stack where the mux creates and manages its own Hub internally.
type WSTest struct {
	*Test
	Server        *httptest.Server        // Real HTTP server for WS connections
	WorkflowQueue *rabbitmq.WorkflowQueue // For publishing alerts in E2E tests
	RabbitClient  *rabbitmq.Client        // Optional, nil if no RabbitMQ
}

// StartWSTest initializes the system for WebSocket integration tests.
// It creates the full HTTP stack including WebSocket routes with JWT auth.
// Unlike StartTest(), this version passes Auth to the mux so BearerQueryParam works.
//
// Note: The Hub/AlertHub are created internally by the mux (all.go) and cannot
// be accessed directly. Use time-based waiting and observable behavior in tests.
func StartWSTest(t *testing.T, testName string) *WSTest {
	t.Helper()

	// Create database infrastructure
	db := dbtest.NewDatabase(t, testName)

	// Create auth instance with test key store
	ath, err := auth.New(auth.Config{
		Log:       db.Log,
		DB:        db.DB,
		KeyLookup: &KeyStore{},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create auth service server
	authServer := httptest.NewServer(mux.WebAPI(mux.Config{
		Log:  db.Log,
		Auth: ath,
		DB:   db.DB,
	}, authbuild.Routes()))

	authClient := authclient.New(db.Log, authServer.URL)

	// Create ichor service mux WITH Auth (for BearerQueryParam WebSocket auth)
	// The mux creates its own Hub and AlertHub internally via all.go
	ichorMux := mux.WebAPI(mux.Config{
		Log:        db.Log,
		Auth:       ath, // Required for BearerQueryParam middleware
		AuthClient: authClient,
		DB:         db.DB,
	}, ichorbuild.Routes())

	// Create real HTTP server with the mux
	server := httptest.NewServer(ichorMux)

	// Create the base Test struct
	test := New(db, ath, ichorMux)

	t.Cleanup(func() {
		server.Close()
		authServer.Close()
	})

	return &WSTest{
		Test:   test,
		Server: server,
	}
}

// Mux returns the HTTP handler for custom test scenarios.
func (wt *WSTest) Mux() http.Handler {
	return wt.mux
}

// StartWSTestWithRabbitMQ initializes WebSocket tests with RabbitMQ support.
// The RabbitMQ client is passed to the mux config so that all.go starts the
// AlertConsumer internally. This enables true E2E testing of the alert pipeline.
func StartWSTestWithRabbitMQ(t *testing.T, testName string) *WSTest {
	t.Helper()

	// Create database infrastructure
	db := dbtest.NewDatabase(t, testName)

	// Create auth instance with test key store
	ath, err := auth.New(auth.Config{
		Log:       db.Log,
		DB:        db.DB,
		KeyLookup: &KeyStore{},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create auth service server
	authServer := httptest.NewServer(mux.WebAPI(mux.Config{
		Log:  db.Log,
		Auth: ath,
		DB:   db.DB,
	}, authbuild.Routes()))

	authClient := authclient.New(db.Log, authServer.URL)

	// Get RabbitMQ test container
	container := rabbitmq.GetTestContainer(t)
	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("RabbitMQ connection failed: %v", err)
	}

	// Create ichor service mux WITH Auth AND RabbitMQ
	// The mux creates its own Hub, AlertHub, and AlertConsumer internally
	ichorMux := mux.WebAPI(mux.Config{
		Log:          db.Log,
		Auth:         ath, // Required for BearerQueryParam middleware
		AuthClient:   authClient,
		DB:           db.DB,
		RabbitClient: client, // Enables AlertConsumer in all.go
	}, ichorbuild.Routes())

	// Create real HTTP server with the mux
	server := httptest.NewServer(ichorMux)

	// Create the base Test struct
	test := New(db, ath, ichorMux)

	// Create workflow queue for publishing test alerts
	wq := rabbitmq.NewWorkflowQueue(client, db.Log)
	if err := wq.Initialize(context.Background()); err != nil {
		t.Fatalf("Queue init failed: %v", err)
	}

	t.Cleanup(func() {
		server.Close()
		authServer.Close()
		client.Close()
	})

	return &WSTest{
		Test:          test,
		Server:        server,
		WorkflowQueue: wq,
		RabbitClient:  client,
	}
}

// ConnectClient creates a WebSocket connection with JWT auth via query parameter.
// This simulates how the frontend connects to the WebSocket endpoint.
func (wt *WSTest) ConnectClient(t *testing.T, token string) *websocket.Conn {
	t.Helper()

	wsURL := "ws" + wt.Server.URL[4:] + "/v1/workflow/alerts/ws?token=" + token
	conn, _, err := websocket.Dial(context.Background(), wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket connect failed: %v", err)
	}
	return conn
}

// ConnectClientWithContext creates a WebSocket connection with context for timeout control.
func (wt *WSTest) ConnectClientWithContext(ctx context.Context, t *testing.T, token string) (*websocket.Conn, error) {
	t.Helper()

	wsURL := "ws" + wt.Server.URL[4:] + "/v1/workflow/alerts/ws?token=" + token
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	return conn, err
}

// TryConnectClient attempts to connect and returns the connection and response.
// Use this for testing authentication failures where you need access to the HTTP response.
func (wt *WSTest) TryConnectClient(ctx context.Context, t *testing.T, token string) (*websocket.Conn, int, error) {
	t.Helper()

	wsURL := "ws" + wt.Server.URL[4:] + "/v1/workflow/alerts/ws"
	if token != "" {
		wsURL += "?token=" + token
	}

	conn, resp, err := websocket.Dial(ctx, wsURL, nil)
	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}
	return conn, statusCode, err
}

// WSTestURL returns the WebSocket URL for the alerts endpoint.
func (wt *WSTest) WSTestURL() string {
	return "ws" + wt.Server.URL[4:] + "/v1/workflow/alerts/ws"
}
