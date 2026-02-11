package client

import (
	"context"
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

func TestClient_Get_CorrectPaths(t *testing.T) {
	tests := []struct {
		name     string
		callFunc func(c *Client) error
		wantPath string
	}{
		{"GetCatalog", func(c *Client) error { _, err := c.GetCatalog(context.Background()); return err }, "/v1/agent/catalog"},
		{"GetActionTypes", func(c *Client) error { _, err := c.GetActionTypes(context.Background()); return err }, "/v1/workflow/action-types"},
		{"GetFieldTypes", func(c *Client) error { _, err := c.GetFieldTypes(context.Background()); return err }, "/v1/config/form-field-types"},
		{"GetSchemas", func(c *Client) error { _, err := c.GetSchemas(context.Background()); return err }, "/v1/introspection/schemas"},
		{"GetTables", func(c *Client) error { _, err := c.GetTables(context.Background(), "core"); return err }, "/v1/introspection/schemas/core/tables"},
		{"GetColumns", func(c *Client) error { _, err := c.GetColumns(context.Background(), "core", "users"); return err }, "/v1/introspection/tables/core/users/columns"},
		{"GetEnumOptions", func(c *Client) error { _, err := c.GetEnumOptions(context.Background(), "core", "role_type"); return err }, "/v1/config/enums/core/role_type/options"},
		{"GetWorkflowRule", func(c *Client) error { _, err := c.GetWorkflowRule(context.Background(), "abc-123"); return err }, "/v1/workflow/rules/abc-123"},
		{"GetPageConfig", func(c *Client) error { _, err := c.GetPageConfig(context.Background(), "abc-123"); return err }, "/v1/config/page-configs/id/abc-123"},
		{"GetFormFull", func(c *Client) error { _, err := c.GetFormFull(context.Background(), "abc-123"); return err }, "/v1/config/forms/abc-123/full"},
		{"GetTableConfig", func(c *Client) error { _, err := c.GetTableConfig(context.Background(), "abc-123"); return err }, "/v1/data/id/abc-123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				w.Write([]byte(`{}`))
			}))
			defer srv.Close()

			c := New(srv.URL, "test-token")
			_ = tt.callFunc(c)

			if gotPath != tt.wantPath {
				t.Errorf("got path %q, want %q", gotPath, tt.wantPath)
			}
		})
	}
}
