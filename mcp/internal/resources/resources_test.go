package resources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
)

// ========== Existing URI parsing tests ==========

func TestParseDBResourceURI(t *testing.T) {
	tests := []struct {
		uri        string
		wantSchema string
		wantTable  string
		wantErr    bool
	}{
		{"config://db/core/users", "core", "users", false},
		{"config://db/inventory/locations", "inventory", "locations", false},
		{"config://db/", "", "", true},
		{"config://db/core/", "", "", true},
		{"config://db/core", "", "", true},
		{"config://other", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			schema, table, err := parseDBResourceURI(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseDBResourceURI(%q) error = %v, wantErr %v", tt.uri, err, tt.wantErr)
			}
			if schema != tt.wantSchema {
				t.Errorf("schema = %q, want %q", schema, tt.wantSchema)
			}
			if table != tt.wantTable {
				t.Errorf("table = %q, want %q", table, tt.wantTable)
			}
		})
	}
}

func TestParseEnumResourceURI(t *testing.T) {
	tests := []struct {
		uri        string
		wantSchema string
		wantName   string
		wantErr    bool
	}{
		{"config://enums/core/role_type", "core", "role_type", false},
		{"config://enums/workflow/trigger_type", "workflow", "trigger_type", false},
		{"config://enums/", "", "", true},
		{"config://enums/core/", "", "", true},
		{"config://enums/core", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			schema, name, err := parseEnumResourceURI(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseEnumResourceURI(%q) error = %v, wantErr %v", tt.uri, err, tt.wantErr)
			}
			if schema != tt.wantSchema {
				t.Errorf("schema = %q, want %q", schema, tt.wantSchema)
			}
			if name != tt.wantName {
				t.Errorf("name = %q, want %q", name, tt.wantName)
			}
		})
	}
}

// ========== Resource handler tests ==========

// setupResourceTest creates a mock HTTP server, registers resources, and returns an MCP session.
func setupResourceTest(t *testing.T, handler http.Handler) (*mcp.ClientSession, context.Context) {
	t.Helper()
	mock := httptest.NewServer(handler)
	t.Cleanup(mock.Close)

	ichorClient := client.New(mock.URL, "test-token")

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test",
		Version: "0.0.1",
	}, nil)
	RegisterResources(server, ichorClient)

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
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	})
}

func TestStaticResources_Success(t *testing.T) {
	tests := []struct {
		name         string
		uri          string
		apiPath      string
		mockResponse string
	}{
		{"catalog", "config://catalog", "/v1/agent/catalog", `[{"name":"page_configs"}]`},
		{"action-types", "config://action-types", "/v1/workflow/action-types", `[{"type":"send_email"}]`},
		{"field-types", "config://field-types", "/v1/config/form-field-types", `[{"type":"text"}]`},
		{"table-config-schema", "config://table-config-schema", "/v1/config/schemas/table-config", `{"type":"object","properties":{}}`},
		{"layout-schema", "config://layout-schema", "/v1/config/schemas/layout", `{"type":"object","properties":{}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, ctx := setupResourceTest(t, pathRouter(map[string]string{
				tt.apiPath: tt.mockResponse,
			}))

			result, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
				URI: tt.uri,
			})
			if err != nil {
				t.Fatalf("ReadResource(%s): %v", tt.uri, err)
			}

			if len(result.Contents) == 0 {
				t.Fatal("no contents returned")
			}

			content := result.Contents[0]
			if content.URI != tt.uri {
				t.Errorf("URI = %q, want %q", content.URI, tt.uri)
			}
			if content.MIMEType != "application/json" {
				t.Errorf("MIMEType = %q, want application/json", content.MIMEType)
			}
			if content.Text != tt.mockResponse {
				t.Errorf("Text = %q, want %q", content.Text, tt.mockResponse)
			}
		})
	}
}

func TestStaticResources_APIError(t *testing.T) {
	errorServer := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"server down"}`))
	})

	session, ctx := setupResourceTest(t, errorServer)

	uris := []string{
		"config://catalog",
		"config://action-types",
		"config://field-types",
		"config://table-config-schema",
		"config://layout-schema",
	}

	for _, uri := range uris {
		t.Run(uri, func(t *testing.T) {
			_, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: uri})
			if err == nil {
				t.Errorf("ReadResource(%s) should return error on API failure", uri)
			}
		})
	}
}

func TestTemplateResource_DBSchema(t *testing.T) {
	columnsJSON := `[{"name":"id","type":"uuid"},{"name":"email","type":"text"}]`

	session, ctx := setupResourceTest(t, pathRouter(map[string]string{
		"/v1/introspection/tables/core/users/columns": columnsJSON,
	}))

	result, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
		URI: "config://db/core/users",
	})
	if err != nil {
		t.Fatalf("ReadResource: %v", err)
	}

	if len(result.Contents) == 0 {
		t.Fatal("no contents returned")
	}

	content := result.Contents[0]
	if content.URI != "config://db/core/users" {
		t.Errorf("URI = %q, want 'config://db/core/users'", content.URI)
	}
	if content.MIMEType != "application/json" {
		t.Errorf("MIMEType = %q, want application/json", content.MIMEType)
	}
	if content.Text != columnsJSON {
		t.Errorf("Text = %q, want %q", content.Text, columnsJSON)
	}
}

func TestTemplateResource_Enums(t *testing.T) {
	enumsJSON := `[{"value":"admin","label":"Administrator"}]`

	session, ctx := setupResourceTest(t, pathRouter(map[string]string{
		"/v1/config/enums/core/role_type/options": enumsJSON,
	}))

	result, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
		URI: "config://enums/core/role_type",
	})
	if err != nil {
		t.Fatalf("ReadResource: %v", err)
	}

	if len(result.Contents) == 0 {
		t.Fatal("no contents returned")
	}

	content := result.Contents[0]
	if content.URI != "config://enums/core/role_type" {
		t.Errorf("URI = %q, want 'config://enums/core/role_type'", content.URI)
	}
	if content.MIMEType != "application/json" {
		t.Errorf("MIMEType = %q, want application/json", content.MIMEType)
	}
	if content.Text != enumsJSON {
		t.Errorf("Text = %q, want %q", content.Text, enumsJSON)
	}
}

func TestTemplateResource_DBSchema_APIError(t *testing.T) {
	errorServer := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"server error"}`))
	})

	session, ctx := setupResourceTest(t, errorServer)

	_, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
		URI: "config://db/core/users",
	})
	if err == nil {
		t.Error("ReadResource should return error on API failure")
	}
}

func TestTemplateResource_Enums_APIError(t *testing.T) {
	errorServer := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"server error"}`))
	})

	session, ctx := setupResourceTest(t, errorServer)

	_, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
		URI: "config://enums/core/role_type",
	})
	if err == nil {
		t.Error("ReadResource should return error on API failure")
	}
}
