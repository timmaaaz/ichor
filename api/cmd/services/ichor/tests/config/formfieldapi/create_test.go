package formfieldapi_test

import (
	"encoding/json"
	"net/http"

	"github.com/google/go-cmp/cmp"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/config/formfieldapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func create200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/form-fields",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &formfieldapp.NewFormField{
				FormID:     sd.Forms[0].ID,
				EntityID:   sd.Entities[0].ID.String(),
				Name:       "test_field",
				Label:      "Test Field",
				FieldType:  "text",
				FieldOrder: 100,
				Required:   true,
				Config:     json.RawMessage(`{"placeholder":"Enter text"}`),
			},
			GotResp: &formfieldapp.FormField{},
			ExpResp: &formfieldapp.FormField{
				FormID:     sd.Forms[0].ID,
				EntityID:   sd.Entities[0].ID.String(),
				Name:       "test_field",
				Label:      "Test Field",
				FieldType:  "text",
				FieldOrder: 100,
				Required:   true,
				Config:     json.RawMessage(`{"placeholder":"Enter text"}`),
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*formfieldapp.FormField)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*formfieldapp.FormField)
				expResp.ID = gotResp.ID

				dbtest.NormalizeJSONFields(gotResp, expResp)
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing-name",
			URL:        "/v1/config/form-fields",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &formfieldapp.NewFormField{
				FormID:     sd.Forms[0].ID,
				EntityID:   sd.Entities[0].ID.String(),
				Label:      "Test Field",
				FieldType:  "text",
				FieldOrder: 1,
				Config:     json.RawMessage(`{}`),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"name\",\"error\":\"name is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "empty token",
			URL:        "/v1/config/form-fields",
			Token:      "&nbsp;",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad token",
			URL:        "/v1/config/form-fields",
			Token:      sd.Admins[0].Token[:10],
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad sig",
			URL:        "/v1/config/form-fields",
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "roleadminonly",
			URL:        "/v1/config/form-fields",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: form_fields"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
