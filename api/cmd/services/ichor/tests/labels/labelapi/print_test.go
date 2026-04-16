package labelapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/labels/labelapp"
)

// runPrintTests covers POST /v1/labels/print. Unlike the table-driven Run()
// helper, these tests need to inspect the RecPrinter between calls, so they
// drive the mux directly via httptest.
func runPrintTests(t *testing.T, test *apitest.Test, printer *apitest.RecPrinter, sd apitest.SeedData) {
	t.Run("print-200-single-copy", func(t *testing.T) {
		printer.Reset()

		// sd.Labels[2] is untouched by the upstream update/delete cases;
		// using sd.Labels[0] here would assert on a captured-at-seed-time
		// Code value that the update-200 test has since patched.
		body := mustJSON(t, labelapp.PrintRequest{
			LabelID: sd.Labels[2].ID,
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/labels/print", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+sd.Admins[0].Token)
		w := httptest.NewRecorder()
		test.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("status: got %d want %d body=%s", w.Code, http.StatusNoContent, w.Body.String())
		}
		calls := printer.Calls()
		if len(calls) != 1 {
			t.Fatalf("printer calls: got %d want 1", len(calls))
		}
		if !bytes.Contains(calls[0], []byte(sd.Labels[2].Code)) {
			t.Fatalf("ZPL did not contain label code %q; got %s", sd.Labels[2].Code, calls[0])
		}
	})

	t.Run("print-200-three-copies", func(t *testing.T) {
		printer.Reset()

		body := mustJSON(t, labelapp.PrintRequest{
			LabelID: sd.Labels[1].ID,
			Copies:  3,
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/labels/print", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+sd.Admins[0].Token)
		w := httptest.NewRecorder()
		test.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("status: got %d body=%s", w.Code, w.Body.String())
		}
		if got := len(printer.Calls()); got != 3 {
			t.Fatalf("printer calls: got %d want 3", got)
		}
	})

	t.Run("print-400-bad-uuid", func(t *testing.T) {
		printer.Reset()

		body := mustJSON(t, labelapp.PrintRequest{LabelID: "not-a-uuid"})
		req := httptest.NewRequest(http.MethodPost, "/v1/labels/print", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+sd.Admins[0].Token)
		w := httptest.NewRecorder()
		test.ServeHTTP(w, req)

		// validate:"required,min=36,max=36" rejects strings that aren't 36 chars.
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d want %d body=%s", w.Code, http.StatusBadRequest, w.Body.String())
		}
		if got := len(printer.Calls()); got != 0 {
			t.Fatalf("printer should not be called on validation failure; got %d calls", got)
		}
	})

	t.Run("print-401-empty-token", func(t *testing.T) {
		printer.Reset()

		body := mustJSON(t, labelapp.PrintRequest{LabelID: sd.Labels[0].ID})
		req := httptest.NewRequest(http.MethodPost, "/v1/labels/print", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer &nbsp;")
		w := httptest.NewRecorder()
		test.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status: got %d want %d body=%s", w.Code, http.StatusUnauthorized, w.Body.String())
		}
		if got := len(printer.Calls()); got != 0 {
			t.Fatalf("printer should not be called on auth failure; got %d calls", got)
		}
	})

	t.Run("print-403-no-read-permission", func(t *testing.T) {
		printer.Reset()

		body := mustJSON(t, labelapp.PrintRequest{LabelID: sd.Labels[0].ID})
		req := httptest.NewRequest(http.MethodPost, "/v1/labels/print", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+sd.Users[0].Token)
		w := httptest.NewRecorder()
		test.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Fatalf("status: got %d want %d body=%s", w.Code, http.StatusForbidden, w.Body.String())
		}
		if got := len(printer.Calls()); got != 0 {
			t.Fatalf("printer should not be called on permission failure; got %d calls", got)
		}
	})

	t.Run("print-404-missing-label", func(t *testing.T) {
		printer.Reset()

		body := mustJSON(t, labelapp.PrintRequest{LabelID: uuid.NewString()})
		req := httptest.NewRequest(http.MethodPost, "/v1/labels/print", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+sd.Admins[0].Token)
		w := httptest.NewRecorder()
		test.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("status: got %d want %d body=%s", w.Code, http.StatusNotFound, w.Body.String())
		}
		if got := len(printer.Calls()); got != 0 {
			t.Fatalf("printer should not be called when label is missing; got %d calls", got)
		}
	})
}

// mustJSON is a tiny helper so each subtest reads as a single statement.
func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %s", err)
	}
	return b
}

