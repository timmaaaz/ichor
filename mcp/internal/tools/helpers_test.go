package tools_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
)

// setupToolTest creates a mock HTTP server, registers tools on an MCP server,
// and returns a connected MCP client session for testing.
func setupToolTest(t *testing.T, handler http.Handler, register func(s *mcp.Server, c *client.Client)) (*mcp.ClientSession, context.Context) {
	t.Helper()
	mock := httptest.NewServer(handler)
	t.Cleanup(mock.Close)

	ichorClient := client.New(mock.URL, "test-token")

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test",
		Version: "0.0.1",
	}, nil)
	register(server, ichorClient)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		server.Connect(ctx, serverTransport, nil)
	}()

	mcpClient := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	session, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { session.Close() })

	return session, ctx
}

// pathRouter creates an HTTP handler that routes based on request path.
// Routes are matched by path (with optional query string). Unmatched paths return 404.
func pathRouter(routes map[string]string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try path + query first.
		pathWithQuery := r.URL.Path
		if r.URL.RawQuery != "" {
			pathWithQuery += "?" + r.URL.RawQuery
		}
		if resp, ok := routes[pathWithQuery]; ok {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(resp))
			return
		}
		// Fall back to path only.
		if resp, ok := routes[r.URL.Path]; ok {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(resp))
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	})
}

// staticHandler returns the same JSON response for all requests.
func staticHandler(response string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(response))
	})
}

// errorHandler returns an HTTP error status for all requests.
func errorHandler(status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		w.Write([]byte(`{"error":"server error"}`))
	})
}

// getTextContent extracts the text from the first content item in a CallToolResult.
func getTextContent(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("no content returned")
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *TextContent, got %T", result.Content[0])
	}
	return tc.Text
}

// callTool is a convenience wrapper for session.CallTool.
func callTool(t *testing.T, session *mcp.ClientSession, ctx context.Context, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		t.Fatalf("CallTool(%s): %v", name, err)
	}
	return result
}
