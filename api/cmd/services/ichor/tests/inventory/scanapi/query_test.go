package scan_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

// scanTypeOnly captures just the type field for assertions where Data content
// is dynamic and not predictable from seed alone.
type scanTypeOnly struct {
	Type string `json:"type"`
}

// scanResult captures the full top-level scan response for unknown/null checks.
type scanResult struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

func scanSerial200(sd apitest.SeedData) []apitest.Table {
	sn := sd.SerialNumbers[0]
	return []apitest.Table{
		{
			Name:       "by-serial-number",
			URL:        "/v1/inventory/scan?barcode=" + sn.SerialNumber,
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &scanTypeOnly{},
			ExpResp:    &scanTypeOnly{Type: "serial"},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got.(*scanTypeOnly).Type, exp.(*scanTypeOnly).Type)
			},
		},
	}
}

func scanLot200(sd apitest.SeedData) []apitest.Table {
	lt := sd.LotTrackings[0]
	return []apitest.Table{
		{
			Name:       "by-lot-number",
			URL:        "/v1/inventory/scan?barcode=" + lt.LotNumber,
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &scanTypeOnly{},
			ExpResp:    &scanTypeOnly{Type: "lot"},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got.(*scanTypeOnly).Type, exp.(*scanTypeOnly).Type)
			},
		},
	}
}

func scanProduct200(sd apitest.SeedData) []apitest.Table {
	p := sd.Products[0]
	return []apitest.Table{
		{
			Name:       "by-upc-code",
			URL:        "/v1/inventory/scan?barcode=" + p.UpcCode,
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &scanTypeOnly{},
			ExpResp:    &scanTypeOnly{Type: "product"},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got.(*scanTypeOnly).Type, exp.(*scanTypeOnly).Type)
			},
		},
	}
}

func scanUnknown200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "unknown-barcode",
			URL:        "/v1/inventory/scan?barcode=BARCODE-UNKNOWN-XYZ-999",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &scanResult{},
			ExpResp:    &scanResult{Type: "unknown", Data: nil},
			CmpFunc: func(got, exp any) string {
				g := got.(*scanResult)
				e := exp.(*scanResult)
				if g.Type != e.Type {
					return cmp.Diff(g.Type, e.Type)
				}
				if g.Data != e.Data {
					return cmp.Diff(g.Data, e.Data)
				}
				return ""
			},
		},
	}
}

func scan400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing-barcode-param",
			URL:        "/v1/inventory/scan",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodGet,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "barcode query parameter is required"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func scan401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-token",
			URL:        "/v1/inventory/scan?barcode=SN-1",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
