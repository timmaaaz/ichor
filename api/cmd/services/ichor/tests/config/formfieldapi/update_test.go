package formfieldapi_test

import (
	"encoding/json"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/config/formfieldapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	updatedConfig := json.RawMessage(`{"updated":"true"}`)
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/form-fields/" + sd.FormFields[0].ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &formfieldapp.UpdateFormField{
				Label:      dbtest.StringPointer("Updated Label"),
				FieldOrder: dbtest.IntPointer(999),
				Config:     &updatedConfig,
			},
			GotResp: &formfieldapp.FormField{},
			ExpResp: &formfieldapp.FormField{
				ID:         sd.FormFields[0].ID,
				FormID:     sd.FormFields[0].FormID,
				Name:       sd.FormFields[0].Name,
				Label:      "Updated Label",
				FieldType:  sd.FormFields[0].FieldType,
				FieldOrder: 999,
				Required:   sd.FormFields[0].Required,
				Config:     json.RawMessage(`{"updated":"true"}`),
			},
			CmpFunc: func(got, exp any) string {
				dbtest.NormalizeJSONFields(got, exp)
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func update400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "invalid-field-order-negative",
			URL:        "/v1/config/form-fields/" + sd.FormFields[0].ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &formfieldapp.UpdateFormField{
				FieldOrder: dbtest.IntPointer(-1),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"field_order\",\"error\":\"field_order must be 0 or greater\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func update401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        "/v1/config/form-fields/" + sd.FormFields[0].ID,
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
			Name:       "token-too-short",
			URL:        "/v1/config/form-fields/" + sd.FormFields[0].ID,
			Token:      sd.Admins[0].Token[:10],
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
			URL:        "/v1/config/form-fields/" + sd.FormFields[0].ID,
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
			Name:       "roleadminonly",
			URL:        "/v1/config/form-fields/" + sd.FormFields[0].ID,
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: form_fields"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func update404(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "not-found",
			URL:        "/v1/config/form-fields/" + uuid.NewString(),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &formfieldapp.UpdateFormField{
				Label: dbtest.StringPointer("Updated Label"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "form field not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
