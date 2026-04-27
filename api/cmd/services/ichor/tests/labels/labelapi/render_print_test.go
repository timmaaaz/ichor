package labelapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/labels/labelapp"
)

// runRenderPrintTests covers POST /v1/labels/render-print. Like print_test,
// these subtests inspect the recorded ZPL between calls so they bypass the
// table-driven Run() helper and drive the mux directly.
func runRenderPrintTests(t *testing.T, test *apitest.Test, printer *apitest.RecPrinter, sd apitest.SeedData) {
	t.Run("render-print-200-product", func(t *testing.T) {
		printer.Reset()

		payload := mustJSON(t, map[string]any{
			"productName": "Widget",
			"sku":         "SKU-RP-1",
			"upc":         "012345678905",
			"lotNumber":   nil,
		})

		body := mustJSON(t, labelapp.RenderPrintRequest{
			Type:    "product",
			Payload: json.RawMessage(payload),
			Copies:  2,
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/labels/render-print", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+sd.Admins[0].Token)
		w := httptest.NewRecorder()
		test.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("status: got %d body=%s", w.Code, w.Body.String())
		}
		calls := printer.Calls()
		if len(calls) != 2 {
			t.Fatalf("printer calls: got %d want 2", len(calls))
		}
		if !bytes.Contains(calls[0], []byte("SKU: SKU-RP-1")) {
			t.Fatalf("ZPL did not contain SKU: SKU-RP-1: %s", calls[0])
		}
	})

	t.Run("render-print-400-bad-type", func(t *testing.T) {
		printer.Reset()

		body := mustJSON(t, labelapp.RenderPrintRequest{
			Type:    "not_a_valid_type",
			Payload: json.RawMessage(`{}`),
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/labels/render-print", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+sd.Admins[0].Token)
		w := httptest.NewRecorder()
		test.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d want %d body=%s", w.Code, http.StatusBadRequest, w.Body.String())
		}
		if got := len(printer.Calls()); got != 0 {
			t.Fatalf("printer should not be called on validation failure; got %d", got)
		}
	})

	t.Run("render-print-400-missing-payload", func(t *testing.T) {
		printer.Reset()

		// Sending raw JSON without a "payload" key triggers the
		// validate:"required" tag on RenderPrintRequest.Payload.
		body := []byte(`{"type":"product"}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/labels/render-print", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+sd.Admins[0].Token)
		w := httptest.NewRecorder()
		test.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d want %d body=%s", w.Code, http.StatusBadRequest, w.Body.String())
		}
		if got := len(printer.Calls()); got != 0 {
			t.Fatalf("printer should not be called on validation failure; got %d", got)
		}
	})

	t.Run("render-print-401-empty-token", func(t *testing.T) {
		printer.Reset()

		body := mustJSON(t, labelapp.RenderPrintRequest{
			Type:    "product",
			Payload: json.RawMessage(`{"productName":"X"}`),
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/labels/render-print", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer &nbsp;")
		w := httptest.NewRecorder()
		test.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status: got %d body=%s", w.Code, w.Body.String())
		}
		if got := len(printer.Calls()); got != 0 {
			t.Fatalf("printer should not be called on auth failure; got %d", got)
		}
	})

	t.Run("render-print-403-no-read-permission", func(t *testing.T) {
		printer.Reset()

		body := mustJSON(t, labelapp.RenderPrintRequest{
			Type:    "product",
			Payload: json.RawMessage(`{"productName":"X"}`),
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/labels/render-print", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+sd.Users[0].Token)
		w := httptest.NewRecorder()
		test.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Fatalf("status: got %d body=%s", w.Code, w.Body.String())
		}
		if got := len(printer.Calls()); got != 0 {
			t.Fatalf("printer should not be called on permission failure; got %d", got)
		}
	})
}
