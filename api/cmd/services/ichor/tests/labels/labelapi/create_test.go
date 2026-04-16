package labelapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/labels/labelapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/labels",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &labelapp.NewLabel{
				Code: "NEW-CREATE-001",
				Type: "location",
			},
			GotResp: &labelapp.Label{},
			ExpResp: &labelapp.Label{
				Code: "NEW-CREATE-001",
				Type: "location",
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*labelapp.Label)
				expResp := exp.(*labelapp.Label)

				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
		{
			Name:       "with-entity-ref-and-payload",
			URL:        "/v1/labels",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &labelapp.NewLabel{
				Code:        "NEW-CREATE-002",
				Type:        "tote",
				EntityRef:   "TOTE-A1",
				PayloadJSON: `{"capacity":50}`,
			},
			GotResp: &labelapp.Label{},
			ExpResp: &labelapp.Label{
				Code:        "NEW-CREATE-002",
				Type:        "tote",
				EntityRef:   "TOTE-A1",
				PayloadJSON: `{"capacity":50}`,
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*labelapp.Label)
				expResp := exp.(*labelapp.Label)

				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing-code",
			URL:        "/v1/labels",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &labelapp.NewLabel{
				Type: "location",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"code","error":"code is a required field"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-type",
			URL:        "/v1/labels",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &labelapp.NewLabel{
				Code: "NEW-BAD-001",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"type","error":"type is a required field"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "invalid-type",
			URL:        "/v1/labels",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &labelapp.NewLabel{
				Code: "NEW-BAD-002",
				Type: "not_a_valid_type",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"type","error":"type must be one of [location tote lot serial product receiving pick]"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-token",
			URL:        "/v1/labels",
			Token:      "&nbsp;",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-sig",
			URL:        "/v1/labels",
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "no-create-permission",
			URL:        "/v1/labels",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusForbidden,
			Input: &labelapp.NewLabel{
				Code: "FORBIDDEN-001",
				Type: "location",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.PermissionDenied, "user does not have permission CREATE for table: inventory.label_catalog"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "duplicate-code",
			URL:        "/v1/labels",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &labelapp.NewLabel{
				Code: sd.Labels[0].Code,
				Type: "location",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "label code already exists"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
