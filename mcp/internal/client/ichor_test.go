package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Get_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("expected bearer token in Authorization header")
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Error("expected application/json Accept header")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "test-token")
	data, err := c.GetCatalog(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != `{"status":"ok"}` {
		t.Errorf("unexpected response: %s", string(data))
	}
}

func TestClient_Get_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "test-token")
	_, err := c.GetCatalog(context.Background())
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestClient_Post_SendsBody(t *testing.T) {
	var gotBody []byte
	var gotMethod string
	var gotContentType string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.Write([]byte(`{"id":"123"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "test-token")
	payload := json.RawMessage(`{"name":"test","value":42}`)
	_, err := c.CreatePageConfig(context.Background(), payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != "POST" {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", gotContentType)
	}
	if string(gotBody) != `{"name":"test","value":42}` {
		t.Errorf("body = %q, want %q", gotBody, `{"name":"test","value":42}`)
	}
}

func TestClient_Put_SendsBody(t *testing.T) {
	var gotMethod string
	var gotBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotBody, _ = io.ReadAll(r.Body)
		w.Write([]byte(`{"id":"abc","updated":true}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "test-token")
	payload := json.RawMessage(`{"name":"updated"}`)
	_, err := c.UpdatePageConfig(context.Background(), "abc-123", payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != "PUT" {
		t.Errorf("method = %q, want PUT", gotMethod)
	}
	if string(gotBody) != `{"name":"updated"}` {
		t.Errorf("body = %q, want %q", gotBody, `{"name":"updated"}`)
	}
}

func TestClient_CorrectPaths(t *testing.T) {
	tests := []struct {
		name       string
		callFunc   func(c *Client) error
		wantPath   string
		wantMethod string // defaults to GET if empty
	}{
		// ========== Discovery / Catalog Methods ==========
		{"GetCatalog", func(c *Client) error { _, err := c.GetCatalog(context.Background()); return err }, "/v1/agent/catalog", ""},
		{"GetActionTypes", func(c *Client) error { _, err := c.GetActionTypes(context.Background()); return err }, "/v1/workflow/action-types", ""},
		{"GetActionTypeSchema", func(c *Client) error {
			_, err := c.GetActionTypeSchema(context.Background(), "send_email")
			return err
		}, "/v1/workflow/action-types/send_email/schema", ""},
		{"GetFieldTypes", func(c *Client) error { _, err := c.GetFieldTypes(context.Background()); return err }, "/v1/config/form-field-types", ""},
		{"GetFieldTypeSchema", func(c *Client) error {
			_, err := c.GetFieldTypeSchema(context.Background(), "text")
			return err
		}, "/v1/config/form-field-types/text/schema", ""},
		{"GetTableConfigSchema", func(c *Client) error {
			_, err := c.GetTableConfigSchema(context.Background())
			return err
		}, "/v1/config/schemas/table-config", ""},
		{"GetLayoutSchema", func(c *Client) error { _, err := c.GetLayoutSchema(context.Background()); return err }, "/v1/config/schemas/layout", ""},
		{"GetContentTypes", func(c *Client) error { _, err := c.GetContentTypes(context.Background()); return err }, "/v1/config/schemas/content-types", ""},
		{"GetTriggerTypes", func(c *Client) error { _, err := c.GetTriggerTypes(context.Background()); return err }, "/v1/workflow/trigger-types", ""},
		{"GetEntityTypes", func(c *Client) error { _, err := c.GetEntityTypes(context.Background()); return err }, "/v1/workflow/entity-types", ""},
		{"GetEntities", func(c *Client) error { _, err := c.GetEntities(context.Background()); return err }, "/v1/workflow/entities", ""},

		// ========== Workflow Read Methods ==========
		{"GetWorkflowRules", func(c *Client) error { _, err := c.GetWorkflowRules(context.Background()); return err }, "/v1/workflow/rules", ""},
		{"GetWorkflowRule", func(c *Client) error {
			_, err := c.GetWorkflowRule(context.Background(), "abc-123")
			return err
		}, "/v1/workflow/rules/abc-123", ""},
		{"GetWorkflowRuleActions", func(c *Client) error {
			_, err := c.GetWorkflowRuleActions(context.Background(), "abc-123")
			return err
		}, "/v1/workflow/rules/abc-123/actions", ""},
		{"GetWorkflowRuleEdges", func(c *Client) error {
			_, err := c.GetWorkflowRuleEdges(context.Background(), "abc-123")
			return err
		}, "/v1/workflow/rules/abc-123/edges", ""},
		{"GetTemplates", func(c *Client) error { _, err := c.GetTemplates(context.Background()); return err }, "/v1/workflow/templates", ""},
		{"GetActiveTemplates", func(c *Client) error { _, err := c.GetActiveTemplates(context.Background()); return err }, "/v1/workflow/templates/active", ""},

		// ========== UI Read Methods ==========
		{"GetPageConfigs", func(c *Client) error { _, err := c.GetPageConfigs(context.Background()); return err }, "/v1/config/page-configs/all", ""},
		{"GetPageConfig", func(c *Client) error {
			_, err := c.GetPageConfig(context.Background(), "abc-123")
			return err
		}, "/v1/config/page-configs/id/abc-123", ""},
		{"GetPageConfigByName", func(c *Client) error {
			_, err := c.GetPageConfigByName(context.Background(), "dashboard")
			return err
		}, "/v1/config/page-configs/name/dashboard", ""},
		{"GetPageContent", func(c *Client) error {
			_, err := c.GetPageContent(context.Background(), "abc-123")
			return err
		}, "/v1/config/page-configs/content/abc-123", ""},
		{"GetForms", func(c *Client) error { _, err := c.GetForms(context.Background()); return err }, "/v1/config/forms", ""},
		{"GetFormFull", func(c *Client) error {
			_, err := c.GetFormFull(context.Background(), "abc-123")
			return err
		}, "/v1/config/forms/abc-123/full", ""},
		{"GetFormByNameFull", func(c *Client) error {
			_, err := c.GetFormByNameFull(context.Background(), "user_form")
			return err
		}, "/v1/config/forms/name/user_form/full", ""},
		{"GetTableConfigs", func(c *Client) error { _, err := c.GetTableConfigs(context.Background()); return err }, "/v1/data/configs/all", ""},
		{"GetTableConfig", func(c *Client) error {
			_, err := c.GetTableConfig(context.Background(), "abc-123")
			return err
		}, "/v1/data/id/abc-123", ""},
		{"GetTableConfigByName", func(c *Client) error {
			_, err := c.GetTableConfigByName(context.Background(), "my-table")
			return err
		}, "/v1/data/name/my-table", ""},

		// ========== Introspection Methods ==========
		{"GetSchemas", func(c *Client) error { _, err := c.GetSchemas(context.Background()); return err }, "/v1/introspection/schemas", ""},
		{"GetTables", func(c *Client) error {
			_, err := c.GetTables(context.Background(), "core")
			return err
		}, "/v1/introspection/schemas/core/tables", ""},
		{"GetColumns", func(c *Client) error {
			_, err := c.GetColumns(context.Background(), "core", "users")
			return err
		}, "/v1/introspection/tables/core/users/columns", ""},
		{"GetRelationships", func(c *Client) error {
			_, err := c.GetRelationships(context.Background(), "core", "users")
			return err
		}, "/v1/introspection/tables/core/users/relationships", ""},
		{"GetEnumTypes", func(c *Client) error {
			_, err := c.GetEnumTypes(context.Background(), "core")
			return err
		}, "/v1/introspection/enums/core", ""},
		{"GetEnumValues", func(c *Client) error {
			_, err := c.GetEnumValues(context.Background(), "core", "role_type")
			return err
		}, "/v1/introspection/enums/core/role_type", ""},
		{"GetEnumOptions", func(c *Client) error {
			_, err := c.GetEnumOptions(context.Background(), "core", "role_type")
			return err
		}, "/v1/config/enums/core/role_type/options", ""},

		// ========== Write Methods (POST) ==========
		{"CreateWorkflow", func(c *Client) error {
			_, err := c.CreateWorkflow(context.Background(), json.RawMessage(`{}`))
			return err
		}, "/v1/workflow/rules/full", "POST"},
		{"ValidateWorkflow", func(c *Client) error {
			_, err := c.ValidateWorkflow(context.Background(), json.RawMessage(`{}`))
			return err
		}, "/v1/workflow/rules/full", "POST"}, // Note: query param ?dry_run=true is handled separately
		{"CreatePageConfig", func(c *Client) error {
			_, err := c.CreatePageConfig(context.Background(), json.RawMessage(`{}`))
			return err
		}, "/v1/config/page-configs", "POST"},
		{"CreatePageContent", func(c *Client) error {
			_, err := c.CreatePageContent(context.Background(), json.RawMessage(`{}`))
			return err
		}, "/v1/config/page-content", "POST"},
		{"CreateForm", func(c *Client) error {
			_, err := c.CreateForm(context.Background(), json.RawMessage(`{}`))
			return err
		}, "/v1/config/forms", "POST"},
		{"CreateFormField", func(c *Client) error {
			_, err := c.CreateFormField(context.Background(), json.RawMessage(`{}`))
			return err
		}, "/v1/config/form-fields", "POST"},
		{"CreateTableConfig", func(c *Client) error {
			_, err := c.CreateTableConfig(context.Background(), json.RawMessage(`{}`))
			return err
		}, "/v1/data", "POST"},
		{"ValidateTableConfig", func(c *Client) error {
			_, err := c.ValidateTableConfig(context.Background(), json.RawMessage(`{}`))
			return err
		}, "/v1/data/validate", "POST"},

		// ========== Write Methods (PUT) ==========
		{"UpdateWorkflow", func(c *Client) error {
			_, err := c.UpdateWorkflow(context.Background(), "abc-123", json.RawMessage(`{}`))
			return err
		}, "/v1/workflow/rules/abc-123/full", "PUT"},
		{"UpdatePageConfig", func(c *Client) error {
			_, err := c.UpdatePageConfig(context.Background(), "abc-123", json.RawMessage(`{}`))
			return err
		}, "/v1/config/page-configs/id/abc-123", "PUT"},
		{"UpdatePageContent", func(c *Client) error {
			_, err := c.UpdatePageContent(context.Background(), "abc-123", json.RawMessage(`{}`))
			return err
		}, "/v1/config/page-content/abc-123", "PUT"},
		{"UpdateForm", func(c *Client) error {
			_, err := c.UpdateForm(context.Background(), "abc-123", json.RawMessage(`{}`))
			return err
		}, "/v1/config/forms/abc-123", "PUT"},
		{"UpdateFormField", func(c *Client) error {
			_, err := c.UpdateFormField(context.Background(), "abc-123", json.RawMessage(`{}`))
			return err
		}, "/v1/config/form-fields/abc-123", "PUT"},
		{"UpdateTableConfig", func(c *Client) error {
			_, err := c.UpdateTableConfig(context.Background(), "abc-123", json.RawMessage(`{}`))
			return err
		}, "/v1/data/abc-123", "PUT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			var gotMethod string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotMethod = r.Method
				w.Write([]byte(`{}`))
			}))
			defer srv.Close()

			c := New(srv.URL, "test-token")
			_ = tt.callFunc(c)

			wantMethod := tt.wantMethod
			if wantMethod == "" {
				wantMethod = "GET"
			}
			if gotMethod != wantMethod {
				t.Errorf("method = %q, want %q", gotMethod, wantMethod)
			}
			if gotPath != tt.wantPath {
				t.Errorf("path = %q, want %q", gotPath, tt.wantPath)
			}
		})
	}
}

func TestClient_ValidateWorkflow_QueryParam(t *testing.T) {
	var gotQuery string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Write([]byte(`{"valid":true}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "test-token")
	_, err := c.ValidateWorkflow(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotQuery != "dry_run=true" {
		t.Errorf("query = %q, want 'dry_run=true'", gotQuery)
	}
}
