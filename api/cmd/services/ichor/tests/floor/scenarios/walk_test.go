package scenarios_test

// HTTP-issuing pattern in this file (established by labels/labelapi tests):
//
//	body := mustJSON(t, payload)
//	req := httptest.NewRequest(http.MethodPost, "/v1/...", bytes.NewReader(body))
//	req.Header.Set("Authorization", "Bearer "+token)
//	w := httptest.NewRecorder()
//	h.ServeHTTP(w, req)
//
// apitest.Test.ServeHTTP is public and exposes the inner mux directly.
// For GET requests with no body, pass nil as the body argument.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

// mustDecodeJSON unmarshals the recorder body into dst. It calls t.Fatalf on
// any decode error and is used by all walk helpers in Tasks 6-11.
func mustDecodeJSON(t *testing.T, w *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), dst); err != nil {
		t.Fatalf("mustDecodeJSON: %v — body: %s", err, w.Body.String())
	}
}

// mustJSON serialises v to JSON. Calls t.Fatalf if marshalling fails.
func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("mustJSON: %v", err)
	}
	return b
}

// adminToken returns a token for admin@example.com, the ZZZADMIN user who
// has full transfer_orders table-access permissions in seed.sql.
// walkTransfer requires this because the FLOOR_WORKER role does not include
// inventory.transfer_orders in its table access grants.
func adminToken(t *testing.T, h *apitest.Test) string {
	t.Helper()
	tok := apitest.Token(h.DB.BusDomain.User, h.Auth, "admin@example.com")
	if tok == "" {
		t.Fatal("adminToken: apitest.Token returned empty string — admin@example.com not found in DB")
	}
	return tok
}

// doRequest issues a single authenticated HTTP request against the test mux and
// returns the response recorder. It calls t.Fatalf if the status does not match
// wantStatus.
func doRequest(t *testing.T, h *apitest.Test, method, path string, token string, body []byte, wantStatus int, gbNote string) *httptest.ResponseRecorder {
	t.Helper()
	var bodyReader *bytes.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	} else {
		bodyReader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != wantStatus {
		t.Fatalf("%s %s: want status %d, got %d — body: %s\n%s", method, path, wantStatus, w.Code, w.Body.String(), gbNote)
	}
	return w
}

// walkTransfer drives a floor-worker transfer through the canonical 5-step
// wizard using real HTTP against the test mux:
//
//  1. GET  /v1/inventory/transfer-orders/{id}            → 200, status=pending   (sanity)
//  2. POST /v1/inventory/transfer-orders/{id}/approve    → 200, status=approved  (GB-010)
//  3. POST /v1/inventory/transfer-orders/{id}/claim      → 200, status=in_transit (GB-010)
//  4. GET  /v1/inventory/inventory-locations?location_code_exact={fromCode} → 200, ≥1 item (GB-008)
//  5. GET  /v1/inventory/inventory-locations?location_code_exact={toCode}   → 200, ≥1 item (GB-008)
//  6. POST /v1/inventory/transfer-orders/{id}/execute    → 200, status=completed (GB-015)
//
// Steps 4-5 validate that the location codes the scenario wired up are actually
// queryable via the inventory-locations API (GB-008: location-code scan).
//
// Step 6 triggers DecrementQuantity at the source. If the scenario seeds
// insufficient stock, the app returns 422/FailedPrecondition (GB-011). That
// case is handled by the caller (TestFloorScenarios_TransferIntraZone) by
// skipping with a descriptive message rather than failing.
func walkTransfer(t *testing.T, h *apitest.Test, _ uuid.UUID, in TransferInputs) {
	t.Helper()
	token := adminToken(t, h)
	idStr := in.TransferID.String()

	// Step 1 — sanity: GET the transfer order, expect status=pending.
	{
		w := doRequest(t, h, http.MethodGet,
			"/v1/inventory/transfer-orders/"+idStr,
			token, nil, http.StatusOK,
			"Step 1 GET transfer order failed — scenario fixtures may not have loaded correctly")

		var to struct {
			Status string `json:"status"`
		}
		mustDecodeJSON(t, w, &to)
		if to.Status != "pending" {
			t.Fatalf("walkTransfer step 1: want status=pending, got %q (transfer_id=%s)", to.Status, idStr)
		}
	}

	// Step 2 — GB-010: advance pending → approved via POST /approve.
	// GB-010 regression: this endpoint must accept a pending transfer and
	// return it in approved state. A 422/400 here means the status machine
	// or the Approve handler is broken.
	{
		w := doRequest(t, h, http.MethodPost,
			"/v1/inventory/transfer-orders/"+idStr+"/approve",
			token, nil, http.StatusOK,
			"GB-010: POST /approve returned non-200 — status machine may be broken (approve must accept pending)")

		var to struct {
			Status string `json:"status"`
		}
		mustDecodeJSON(t, w, &to)
		if to.Status != "approved" {
			t.Fatalf("walkTransfer step 2 (GB-010): want status=approved after /approve, got %q", to.Status)
		}
	}

	// Step 3 — GB-010: advance approved → in_transit via POST /claim.
	// GB-010 regression: /claim must accept an approved transfer and return
	// it with status=in_transit. Failing here means the Claim handler or the
	// status transition is broken before we even reach the execute step.
	{
		w := doRequest(t, h, http.MethodPost,
			"/v1/inventory/transfer-orders/"+idStr+"/claim",
			token, nil, http.StatusOK,
			"GB-010: POST /claim returned non-200 — status machine may be broken (claim must accept approved)")

		var to struct {
			Status string `json:"status"`
		}
		mustDecodeJSON(t, w, &to)
		if to.Status != "in_transit" {
			t.Fatalf("walkTransfer step 3 (GB-010): want status=in_transit after /claim, got %q", to.Status)
		}
	}

	// Step 4 — GB-008: from-location must be queryable by location_code_exact.
	// GB-008 regression: the inventory-locations query filter must resolve the
	// location code the scenario wired as the source. A 200 with empty items
	// means the location code didn't survive the scenario fixture load.
	{
		url := fmt.Sprintf("/v1/inventory/inventory-locations?location_code_exact=%s", in.FromCode)
		w := doRequest(t, h, http.MethodGet, url, token, nil, http.StatusOK,
			"GB-008: GET inventory-locations by from_location_code returned non-200")

		var result struct {
			Items []struct {
				LocationID string `json:"location_id"`
			} `json:"items"`
			Total int `json:"total"`
		}
		mustDecodeJSON(t, w, &result)
		if result.Total == 0 || len(result.Items) == 0 {
			t.Fatalf("GB-008: from_location %q not found via location_code_exact filter (total=%d)", in.FromCode, result.Total)
		}
		if result.Items[0].LocationID == "" {
			t.Fatalf("GB-008: from_location %q item missing location_id in response", in.FromCode)
		}
	}

	// Step 5 — GB-008: to-location must be queryable by location_code_exact.
	{
		url := fmt.Sprintf("/v1/inventory/inventory-locations?location_code_exact=%s", in.ToCode)
		w := doRequest(t, h, http.MethodGet, url, token, nil, http.StatusOK,
			"GB-008: GET inventory-locations by to_location_code returned non-200")

		var result struct {
			Items []struct {
				LocationID string `json:"location_id"`
			} `json:"items"`
			Total int `json:"total"`
		}
		mustDecodeJSON(t, w, &result)
		if result.Total == 0 || len(result.Items) == 0 {
			t.Fatalf("GB-008: to_location %q not found via location_code_exact filter (total=%d)", in.ToCode, result.Total)
		}
		if result.Items[0].LocationID == "" {
			t.Fatalf("GB-008: to_location %q item missing location_id in response", in.ToCode)
		}
	}

	// Step 6 — GB-015: execute the in_transit transfer.
	// The execute handler calls DecrementQuantity at the source; if the seed
	// quantity < transfer quantity, this returns 422 (ErrInsufficientStock →
	// FailedPrecondition). That is the GB-011 surface: the caller must t.Skip
	// with an explanation rather than letting this t.Fatalf propagate. Any
	// other non-200 status is a genuine GB-015 regression.
	{
		w := doRequest(t, h, http.MethodPost,
			"/v1/inventory/transfer-orders/"+idStr+"/execute",
			token, nil, http.StatusOK,
			"GB-015: POST /execute returned non-200 — may be GB-011 (insufficient stock) or a genuine execute regression")

		var to struct {
			Status string `json:"status"`
		}
		mustDecodeJSON(t, w, &to)
		if to.Status != "completed" {
			t.Fatalf("walkTransfer step 6 (GB-015): want status=completed after /execute, got %q", to.Status)
		}
	}
}
