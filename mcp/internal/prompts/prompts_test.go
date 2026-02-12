package prompts_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
	"github.com/timmaaaz/ichor/mcp/internal/prompts"
)

// setupPromptTest creates a mock HTTP server, registers prompts, and returns an MCP session.
func setupPromptTest(t *testing.T, handler http.Handler) (*mcp.ClientSession, context.Context) {
	t.Helper()
	mock := httptest.NewServer(handler)
	t.Cleanup(mock.Close)

	ichorClient := client.New(mock.URL, "test-token")

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test",
		Version: "0.0.1",
	}, nil)
	prompts.RegisterPrompts(server, ichorClient)

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
func pathRouter(routes map[string]string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if resp, ok := routes[r.URL.Path]; ok {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(resp))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	})
}

func TestBuildWorkflow_Success(t *testing.T) {
	mockRoutes := map[string]string{
		"/v1/workflow/action-types":  `[{"type":"send_email"},{"type":"evaluate_condition"}]`,
		"/v1/workflow/trigger-types": `["on_create","on_update","on_delete"]`,
	}

	session, ctx := setupPromptTest(t, pathRouter(mockRoutes))

	result, err := session.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "build-workflow",
		Arguments: map[string]string{
			"trigger": "on_create",
			"entity":  "orders",
		},
	})
	if err != nil {
		t.Fatalf("GetPrompt: %v", err)
	}

	if result.Description == "" {
		t.Error("expected non-empty description")
	}
	if !strings.Contains(result.Description, "on_create") {
		t.Error("description should mention trigger type")
	}
	if !strings.Contains(result.Description, "orders") {
		t.Error("description should mention entity")
	}

	if len(result.Messages) == 0 {
		t.Fatal("expected at least one message")
	}

	msg := result.Messages[0]
	if msg.Role != "user" {
		t.Errorf("role = %q, want 'user'", msg.Role)
	}

	tc, ok := msg.Content.(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *TextContent, got %T", msg.Content)
	}

	// Verify the prompt contains context data.
	if !strings.Contains(tc.Text, "on_create") {
		t.Error("prompt text should contain trigger argument")
	}
	if !strings.Contains(tc.Text, "orders") {
		t.Error("prompt text should contain entity argument")
	}
	if !strings.Contains(tc.Text, "send_email") {
		t.Error("prompt text should contain action types from API")
	}
	if !strings.Contains(tc.Text, "on_delete") {
		t.Error("prompt text should contain trigger types from API")
	}
}

func TestConfigurePage_Success(t *testing.T) {
	mockRoutes := map[string]string{
		"/v1/config/schemas/content-types": `[{"type":"table"},{"type":"form"},{"type":"chart"}]`,
		"/v1/config/form-field-types":      `[{"type":"text"},{"type":"number"}]`,
	}

	session, ctx := setupPromptTest(t, pathRouter(mockRoutes))

	result, err := session.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "configure-page",
		Arguments: map[string]string{
			"entity": "products",
		},
	})
	if err != nil {
		t.Fatalf("GetPrompt: %v", err)
	}

	if result.Description == "" {
		t.Error("expected non-empty description")
	}
	if !strings.Contains(result.Description, "products") {
		t.Error("description should mention entity")
	}

	if len(result.Messages) == 0 {
		t.Fatal("expected at least one message")
	}

	tc, ok := result.Messages[0].Content.(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *TextContent, got %T", result.Messages[0].Content)
	}

	if !strings.Contains(tc.Text, "products") {
		t.Error("prompt text should contain entity argument")
	}
	if !strings.Contains(tc.Text, "table") {
		t.Error("prompt text should contain content types from API")
	}
	if !strings.Contains(tc.Text, "text") {
		t.Error("prompt text should contain field types from API")
	}
}

func TestDesignForm_Success(t *testing.T) {
	mockRoutes := map[string]string{
		"/v1/config/form-field-types": `[{"type":"text"},{"type":"dropdown"},{"type":"date"}]`,
	}

	session, ctx := setupPromptTest(t, pathRouter(mockRoutes))

	result, err := session.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "design-form",
		Arguments: map[string]string{
			"entity": "invoices",
		},
	})
	if err != nil {
		t.Fatalf("GetPrompt: %v", err)
	}

	if result.Description == "" {
		t.Error("expected non-empty description")
	}
	if !strings.Contains(result.Description, "invoices") {
		t.Error("description should mention entity")
	}

	if len(result.Messages) == 0 {
		t.Fatal("expected at least one message")
	}

	tc, ok := result.Messages[0].Content.(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *TextContent, got %T", result.Messages[0].Content)
	}

	if !strings.Contains(tc.Text, "invoices") {
		t.Error("prompt text should contain entity argument")
	}
	if !strings.Contains(tc.Text, "dropdown") {
		t.Error("prompt text should contain field types from API")
	}
}

func TestPrompts_APIError_StillReturnsPrompt(t *testing.T) {
	// When API calls fail, prompts should still return (with empty data sections).
	errorServer := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"server down"}`))
	})

	session, ctx := setupPromptTest(t, errorServer)

	// build-workflow ignores API errors (uses _ for error return).
	result, err := session.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "build-workflow",
		Arguments: map[string]string{
			"trigger": "on_create",
			"entity":  "orders",
		},
	})
	if err != nil {
		t.Fatalf("GetPrompt should not return error even if API fails: %v", err)
	}
	if len(result.Messages) == 0 {
		t.Error("prompt should still return messages even if API fails")
	}
}
