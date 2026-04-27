package labelapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/labels/labelapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "patch-code-and-type",
			URL:        fmt.Sprintf("/v1/labels/%s", sd.Labels[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &labelapp.UpdateLabel{
				Code: dbtest.StringPointer("UPDATED-001"),
				Type: dbtest.StringPointer("container"),
			},
			GotResp: &labelapp.Label{},
			ExpResp: &labelapp.Label{
				ID:          sd.Labels[0].ID,
				Code:        "UPDATED-001",
				Type:        "container",
				CreatedDate: sd.Labels[0].CreatedDate,
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "patch-entity-ref-only",
			URL:        fmt.Sprintf("/v1/labels/%s", sd.Labels[1].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &labelapp.UpdateLabel{
				EntityRef: dbtest.StringPointer("REF-XYZ"),
			},
			GotResp: &labelapp.Label{},
			ExpResp: &labelapp.Label{
				ID:          sd.Labels[1].ID,
				Code:        sd.Labels[1].Code,
				Type:        sd.Labels[1].Type,
				EntityRef:   "REF-XYZ",
				CreatedDate: sd.Labels[1].CreatedDate,
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "invalid-type",
			URL:        fmt.Sprintf("/v1/labels/%s", sd.Labels[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &labelapp.UpdateLabel{
				Type: dbtest.StringPointer("not_a_valid_type"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"type","error":"type must be one of [location container lot serial product receiving pick]"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-uuid",
			URL:        "/v1/labels/not-a-uuid",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &labelapp.UpdateLabel{
				Code: dbtest.StringPointer("X"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "invalid UUID length: 10"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-token",
			URL:        fmt.Sprintf("/v1/labels/%s", sd.Labels[0].ID),
			Token:      "&nbsp;",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-sig",
			URL:        fmt.Sprintf("/v1/labels/%s", sd.Labels[0].ID),
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "no-update-permission",
			URL:        fmt.Sprintf("/v1/labels/%s", sd.Labels[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusForbidden,
			Input: &labelapp.UpdateLabel{
				Code: dbtest.StringPointer("X"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.PermissionDenied, "user does not have permission UPDATE for table: inventory.label_catalog"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        fmt.Sprintf("/v1/labels/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &labelapp.UpdateLabel{
				Code: dbtest.StringPointer("MISSING"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "label not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
