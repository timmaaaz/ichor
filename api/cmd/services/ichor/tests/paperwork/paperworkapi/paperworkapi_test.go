// Package paperworkapi_test verifies the paperwork API surface end-to-end:
//   - 200 with application/pdf bytes (valid bearer + Read on the underlying
//     domain table) — happy-path cases drive the mux directly via
//     test.ServeHTTP because apitest.Run() unmarshals JSON, while paperwork
//     handlers return raw application/pdf.
//   - 403 Forbidden — non-admin user whose paperwork RouteTables have been
//     downgraded by insertSeedData.
//   - 400 Bad Request — invalid UUID query param.
//   - 401 Unauthorized — malformed bearer token.
package paperworkapi_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

// pdfMagic is the 5-byte header every well-formed PDF begins with.
var pdfMagic = []byte("%PDF-")

func Test_Paperwork(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_Paperwork")
	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// 200 happy paths — PDF body assertions, driven directly through
	// the mux because apitest.Run() expects JSON responses.
	t.Run("auth-200-pick-sheet", func(t *testing.T) {
		assertPDFOK(t, test, sd.Admins[0].Token,
			fmt.Sprintf("/v1/paperwork/pick-sheet?order_id=%s", sd.OrderID),
			"SO-")
	})
	t.Run("auth-200-receive-cover", func(t *testing.T) {
		assertPDFOK(t, test, sd.Admins[0].Token,
			fmt.Sprintf("/v1/paperwork/receive-cover?po_id=%s", sd.PurchaseID),
			"PO-")
	})
	t.Run("auth-200-transfer-sheet", func(t *testing.T) {
		assertPDFOK(t, test, sd.Admins[0].Token,
			fmt.Sprintf("/v1/paperwork/transfer-sheet?transfer_id=%s", sd.TransferID),
			"XFER-")
	})

	// 403 / 400 / 401 — JSON error responses, driven via apitest.Run.
	test.Run(t, forbidden403(sd), "auth-403")
	test.Run(t, badRequest400(sd), "auth-400")
	test.Run(t, noAuth401(sd), "noauth-401")
}

// assertPDFOK drives a GET against the paperwork mux with the given bearer
// token and asserts the response is 200, Content-Type application/pdf, body
// begins with %PDF- magic, and the body contains the expected task-code
// prefix (SO-, PO-, or XFER-).
func assertPDFOK(t *testing.T, test *apitest.Test, token, url, taskCodePrefix string) {
	t.Helper()

	r := httptest.NewRequest(http.MethodGet, url, nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	test.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200; body: %s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("Content-Type: got %q, want application/pdf", ct)
	}
	body := w.Body.Bytes()
	if !bytes.HasPrefix(body, pdfMagic) {
		head := body
		if len(head) > 8 {
			head = head[:8]
		}
		t.Fatalf("body is not a PDF: first bytes = %q", head)
	}
	if !bytes.Contains(body, []byte(taskCodePrefix)) {
		t.Errorf("body missing taskCode prefix %q", taskCodePrefix)
	}
}

// forbidden403 verifies all three paperwork endpoints return 403 Forbidden
// when called with a non-admin bearer token whose table_access rows for
// sales.orders, procurement.purchase_orders, and inventory.transfer_orders
// have been downgraded to 0 by insertSeedData.
func forbidden403(sd PaperworkSeed) []apitest.Table {
	tok := sd.Users[0].Token
	return []apitest.Table{
		{
			Name:       "pick-sheet",
			URL:        fmt.Sprintf("/v1/paperwork/pick-sheet?order_id=%s", sd.OrderID),
			Token:      tok,
			Method:     http.MethodGet,
			StatusCode: http.StatusForbidden,
			GotResp:    &errs.Error{},
			ExpResp:    nil,
			CmpFunc:    func(got, exp any) string { return "" },
		},
		{
			Name:       "receive-cover",
			URL:        fmt.Sprintf("/v1/paperwork/receive-cover?po_id=%s", sd.PurchaseID),
			Token:      tok,
			Method:     http.MethodGet,
			StatusCode: http.StatusForbidden,
			GotResp:    &errs.Error{},
			ExpResp:    nil,
			CmpFunc:    func(got, exp any) string { return "" },
		},
		{
			Name:       "transfer-sheet",
			URL:        fmt.Sprintf("/v1/paperwork/transfer-sheet?transfer_id=%s", sd.TransferID),
			Token:      tok,
			Method:     http.MethodGet,
			StatusCode: http.StatusForbidden,
			GotResp:    &errs.Error{},
			ExpResp:    nil,
			CmpFunc:    func(got, exp any) string { return "" },
		},
	}
}

// badRequest400 verifies an admin-authorized request with a non-UUID query
// param surfaces 400 from the handler's uuid.Parse error mapping. One case
// is enough — the same uuid.Parse path is shared by all three handlers.
func badRequest400(sd PaperworkSeed) []apitest.Table {
	tok := sd.Admins[0].Token
	return []apitest.Table{
		{
			Name:       "pick-sheet-bad-uuid",
			URL:        "/v1/paperwork/pick-sheet?order_id=not-a-uuid",
			Token:      tok,
			Method:     http.MethodGet,
			StatusCode: http.StatusBadRequest,
			GotResp:    &errs.Error{},
			ExpResp:    nil,
			CmpFunc:    func(got, exp any) string { return "" },
		},
	}
}

// noAuth401 verifies one paperwork endpoint returns 401 without a valid
// bearer token. mid.Authenticate rejects unauthenticated requests before
// reaching the handler. Single case — one Authenticate middleware shared
// across all three routes; if it rejects on pick-sheet it rejects on the
// others.
//
// Token is "&nbsp;" (the labelapi pattern) — a malformed Authorization
// header that mid.Authenticate forwards to the auth service via
// authclient; the auth service rejects it and the rejection is mapped to
// 401 Unauthenticated.
func noAuth401(_ PaperworkSeed) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "pick-sheet",
			URL:        fmt.Sprintf("/v1/paperwork/pick-sheet?order_id=%s", uuid.New().String()),
			Token:      "&nbsp;",
			Method:     http.MethodGet,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    nil,
			CmpFunc:    func(got, exp any) string { return "" },
		},
	}
}
