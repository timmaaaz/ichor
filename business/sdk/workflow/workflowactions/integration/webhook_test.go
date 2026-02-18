package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// ---------------------------------------------------------------------------
// Validate tests
// ---------------------------------------------------------------------------

func TestValidate(t *testing.T) {
	h := NewCallWebhookHandler(nil)

	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{
			name:   "valid https POST",
			config: `{"url":"https://example.com/hook","method":"POST","body":"{}"}`,
		},
		{
			name:   "valid localhost http",
			config: `{"url":"http://localhost:8080/hook","method":"GET"}`,
		},
		{
			name:   "valid private IP http",
			config: `{"url":"http://10.0.0.5/hook","method":"POST"}`,
		},
		{
			name:   "valid 192.168 http",
			config: `{"url":"http://192.168.1.100/hook","method":"PUT"}`,
		},
		{
			name:    "missing url",
			config:  `{"method":"POST"}`,
			wantErr: "url is required",
		},
		{
			name:    "non-https external",
			config:  `{"url":"http://external.com/hook","method":"POST"}`,
			wantErr: "only https URLs are allowed",
		},
		{
			name:    "invalid method",
			config:  `{"url":"https://example.com/hook","method":"CONNECT"}`,
			wantErr: "invalid method",
		},
		{
			name:    "timeout too high",
			config:  `{"url":"https://example.com/hook","method":"POST","timeout_seconds":999}`,
			wantErr: "timeout_seconds must be <= 120",
		},
		{
			name:   "template URL skips scheme check",
			config: `{"url":"{{webhook_url}}","method":"POST"}`,
		},
		{
			name:   "empty method defaults to POST (valid)",
			config: `{"url":"https://example.com/hook"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := h.Validate(json.RawMessage(tt.config))
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// isInternalHost tests
// ---------------------------------------------------------------------------

func TestIsInternalHost(t *testing.T) {
	tests := []struct {
		host     string
		expected bool
	}{
		{"localhost", true},
		{"localhost:8080", true},
		{"127.0.0.1", true},
		{"127.0.0.1:9090", true},
		{"10.0.0.5", true},
		{"10.0.0.5:443", true},
		{"192.168.1.1", true},
		{"192.168.1.1:3000", true},
		{"example.com", false},
		{"example.com:443", false},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"172.15.255.255", false},
		{"[::1]:8080", true},
		{"::1", true},
		{"169.254.1.1", true},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			assert.Equal(t, tt.expected, isInternalHost(tt.host))
		})
	}
}

// ---------------------------------------------------------------------------
// determineWebhookOutput tests
// ---------------------------------------------------------------------------

func TestDetermineWebhookOutput(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		expected []int
		want     string
	}{
		{"200 in defaults", 200, nil, "success"},
		{"201 in defaults", 201, nil, "success"},
		{"204 in defaults", 204, nil, "success"},
		{"301 not in expected", 301, []int{200}, "unexpected_response"},
		{"400 client error", 400, nil, "client_error"},
		{"404 client error", 404, nil, "client_error"},
		{"500 server error", 500, nil, "server_error"},
		{"503 server error", 503, nil, "server_error"},
		{"custom expected 202", 202, []int{202}, "success"},
		{"200 not in custom expected", 200, []int{202}, "unexpected_response"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := tt.expected
			if expected == nil {
				expected = []int{200, 201, 202, 204}
			}
			assert.Equal(t, tt.want, determineWebhookOutput(tt.status, expected))
		})
	}
}

// ---------------------------------------------------------------------------
// resolveTemplateVars tests
// ---------------------------------------------------------------------------

func TestResolveTemplateVars(t *testing.T) {
	data := map[string]any{
		"entity_id":    "abc-123",
		"product_name": "Widget",
		"quantity":     42,
	}

	tests := []struct {
		name     string
		template string
		want     string
	}{
		{"simple var", "ID: {{entity_id}}", "ID: abc-123"},
		{"multiple vars", "{{product_name}} x {{quantity}}", "Widget x 42"},
		{"unknown var kept", "{{unknown}}", "{{unknown}}"},
		{"no vars", "plain text", "plain text"},
		{"nil data", "{{entity_id}}", "{{entity_id}}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := data
			if tt.name == "nil data" {
				d = nil
			}
			assert.Equal(t, tt.want, resolveTemplateVars(tt.template, d))
		})
	}
}

// ---------------------------------------------------------------------------
// Execute tests (using httptest)
// ---------------------------------------------------------------------------

func TestExecuteSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	h := NewCallWebhookHandlerWithClient(nil, server.Client())

	config := json.RawMessage(fmt.Sprintf(`{"url":"%s","method":"POST","body":"{\"id\":\"{{entity_id}}\"}"}`, server.URL))

	execCtx := workflow.ActionExecutionContext{
		EntityID: uuid.New(),
		RawData: map[string]any{
			"entity_id": "test-123",
		},
	}

	result, err := h.Execute(context.Background(), config, execCtx)
	require.NoError(t, err)

	m := result.(map[string]any)
	assert.Equal(t, "success", m["output"])
	assert.Equal(t, 200, m["status_code"])
	assert.Contains(t, m["body"], `"ok":true`)
	assert.Equal(t, "application/json", m["content_type"])
}

func TestExecuteClientError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	h := NewCallWebhookHandlerWithClient(nil, server.Client())
	config := json.RawMessage(fmt.Sprintf(`{"url":"%s","method":"GET"}`, server.URL))

	result, err := h.Execute(context.Background(), config, workflow.ActionExecutionContext{})
	require.NoError(t, err)

	m := result.(map[string]any)
	assert.Equal(t, "client_error", m["output"])
	assert.Equal(t, 404, m["status_code"])
}

func TestExecuteServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	h := NewCallWebhookHandlerWithClient(nil, server.Client())
	config := json.RawMessage(fmt.Sprintf(`{"url":"%s","method":"POST"}`, server.URL))

	result, err := h.Execute(context.Background(), config, workflow.ActionExecutionContext{})
	require.NoError(t, err)

	m := result.(map[string]any)
	assert.Equal(t, "server_error", m["output"])
	assert.Equal(t, 500, m["status_code"])
}

func TestExecuteTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	h := NewCallWebhookHandlerWithClient(nil, server.Client())
	config := json.RawMessage(fmt.Sprintf(`{"url":"%s","method":"GET","timeout_seconds":1}`, server.URL))

	result, err := h.Execute(context.Background(), config, workflow.ActionExecutionContext{})
	// Timeout returns a result with "timeout" output, no error.
	require.NoError(t, err)

	m := result.(map[string]any)
	assert.Equal(t, "timeout", m["output"])
}

func TestExecuteTemplateResolution(t *testing.T) {
	var receivedBody string
	var receivedAuthHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		receivedBody = string(b)
		receivedAuthHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	h := NewCallWebhookHandlerWithClient(nil, server.Client())

	config := json.RawMessage(fmt.Sprintf(`{
		"url": "%s",
		"method": "POST",
		"headers": {"Authorization": "Bearer {{api_token}}"},
		"body": "{\"order_id\": \"{{order_id}}\"}"
	}`, server.URL))

	execCtx := workflow.ActionExecutionContext{
		RawData: map[string]any{
			"api_token": "tok_abc123",
			"order_id":  "ORD-999",
		},
	}

	result, err := h.Execute(context.Background(), config, execCtx)
	require.NoError(t, err)

	m := result.(map[string]any)
	assert.Equal(t, "success", m["output"])
	assert.Equal(t, `{"order_id": "ORD-999"}`, receivedBody)
	assert.Equal(t, "Bearer tok_abc123", receivedAuthHeader)
}

func TestGetType(t *testing.T) {
	h := NewCallWebhookHandler(nil)
	assert.Equal(t, "call_webhook", h.GetType())
}

func TestGetOutputPorts(t *testing.T) {
	h := NewCallWebhookHandler(nil)
	ports := h.GetOutputPorts()
	assert.Len(t, ports, 6)

	names := make([]string, len(ports))
	for i, p := range ports {
		names[i] = p.Name
	}
	assert.Contains(t, names, "success")
	assert.Contains(t, names, "client_error")
	assert.Contains(t, names, "server_error")
	assert.Contains(t, names, "unexpected_response")
	assert.Contains(t, names, "timeout")
	assert.Contains(t, names, "failure")
}
