// Package paperworkapi_test verifies the paperwork API surface returns
// 501 Not Implemented while authenticated and 401 Unauthorized without.
package paperworkapi_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func Test_Paperwork(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_Paperwork")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	test.Run(t, authed501(sd), "auth-501")
	test.Run(t, noAuth401(sd), "noauth-401")
}

// authed501 verifies all three paperwork endpoints return 501 Not Implemented
// when called with a valid admin bearer token. The handlers map ErrNotImplemented
// (from the bus stub) to errs.Unimplemented (HTTP 501).
func authed501(sd apitest.SeedData) []apitest.Table {
	tok := sd.Admins[0].Token
	url := func(path, key string) string {
		return fmt.Sprintf("%s?%s=%s", path, key, uuid.New().String())
	}
	return []apitest.Table{
		{
			Name:       "pick-sheet",
			URL:        url("/v1/paperwork/pick-sheet", "order_id"),
			Token:      tok,
			Method:     http.MethodGet,
			StatusCode: http.StatusNotImplemented,
			GotResp:    &errs.Error{},
			ExpResp:    nil,
			CmpFunc:    func(got, exp any) string { return "" },
		},
		{
			Name:       "receive-cover",
			URL:        url("/v1/paperwork/receive-cover", "po_id"),
			Token:      tok,
			Method:     http.MethodGet,
			StatusCode: http.StatusNotImplemented,
			GotResp:    &errs.Error{},
			ExpResp:    nil,
			CmpFunc:    func(got, exp any) string { return "" },
		},
		{
			Name:       "transfer-sheet",
			URL:        url("/v1/paperwork/transfer-sheet", "transfer_id"),
			Token:      tok,
			Method:     http.MethodGet,
			StatusCode: http.StatusNotImplemented,
			GotResp:    &errs.Error{},
			ExpResp:    nil,
			CmpFunc:    func(got, exp any) string { return "" },
		},
	}
}

// noAuth401 verifies one paperwork endpoint returns 401 without a valid bearer
// token. mid.Authenticate rejects unauthenticated requests before reaching
// the handler. Single case — one Authenticate middleware shared across all
// three routes; if it rejects on pick-sheet it rejects on the others.
//
// Token is "&nbsp;" (the labelapi pattern) — a malformed Authorization header
// that mid.Authenticate forwards to the auth service via authclient; the auth
// service rejects it and the rejection is mapped to 401 Unauthenticated.
func noAuth401(_ apitest.SeedData) []apitest.Table {
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
