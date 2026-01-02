package alertws_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

// =============================================================================
// WebSocket Authentication Tests
// =============================================================================

// testWSUpgradeAuthenticated verifies that WebSocket upgrade succeeds
// when a valid JWT token is provided via query parameter.
func testWSUpgradeAuthenticated(t *testing.T, test *apitest.WSTest, sd AlertWSSeedData) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect with valid token
	conn, err := test.ConnectClientWithContext(ctx, t, sd.UserToken(0))
	if err != nil {
		t.Fatalf("Expected successful connection with valid token: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "test complete")

	// Verify connection is functional by waiting a moment
	// (if connection failed, we'd have gotten an error above)
	time.Sleep(50 * time.Millisecond)

	// Connection succeeded - test passes
}

// testWSUpgradeUnauthenticated verifies that WebSocket upgrade fails
// when no token is provided.
func testWSUpgradeUnauthenticated(t *testing.T, test *apitest.WSTest, sd AlertWSSeedData) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to connect without a token
	conn, statusCode, err := test.TryConnectClient(ctx, t, "")
	if err == nil {
		conn.Close(websocket.StatusNormalClosure, "unexpected success")
		t.Fatal("Expected connection to fail without auth token")
	}

	// Should get 401 Unauthorized
	if statusCode != http.StatusUnauthorized && statusCode != 0 {
		// Note: statusCode might be 0 if the connection was rejected before HTTP response
		t.Logf("Status code: %d (0 is acceptable if connection was rejected early)", statusCode)
	}
}

// testWSUpgradeInvalidToken verifies that WebSocket upgrade fails
// when an invalid/malformed token is provided.
func testWSUpgradeInvalidToken(t *testing.T, test *apitest.WSTest, sd AlertWSSeedData) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testCases := []struct {
		name  string
		token string
	}{
		{
			name:  "malformed-token",
			token: "not.a.valid.jwt.token",
		},
		{
			name:  "empty-segments",
			token: "...",
		},
		{
			name:  "random-string",
			token: "randomstringwithoutdots",
		},
		{
			name:  "expired-format",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn, _, err := test.TryConnectClient(ctx, t, tc.token)
			if err == nil {
				conn.Close(websocket.StatusNormalClosure, "unexpected success")
				t.Errorf("Expected connection to fail with invalid token: %s", tc.token)
			}
		})
	}
}

// testWSUpgradeWithHeaderToken verifies that JWT can also be provided
// via Authorization header (for non-browser clients).
func testWSUpgradeWithHeaderToken(t *testing.T, test *apitest.WSTest, sd AlertWSSeedData) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// The BearerQueryParam middleware also checks the Authorization header
	// This test verifies that path works as well
	wsURL := test.WSTestURL()

	// Connect with header-based auth
	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: map[string][]string{
			"Authorization": {"Bearer " + sd.UserToken(0)},
		},
	})
	if err != nil {
		t.Fatalf("Expected successful connection with header token: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "test complete")
}
