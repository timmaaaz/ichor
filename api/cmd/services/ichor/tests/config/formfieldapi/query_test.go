package formfieldapi_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/config/formfieldapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/form-fields?page=1&rows=10",
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &query.Result[formfieldapp.FormField]{},
			ExpResp: &query.Result[formfieldapp.FormField]{
				Page:        1,
				RowsPerPage: 10,
				Total:       20,
				Items:       sd.FormFields[:10],
			},
			CmpFunc: func(got, exp any) string {
				expResp, ok := exp.(*query.Result[formfieldapp.FormField])
				if !ok {
					return "error occurred"
				}
				gotResp, ok := got.(*query.Result[formfieldapp.FormField])
				if !ok {
					return "error occurred"
				}

				dbtest.NormalizeJSONFields(gotResp, expResp)
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
	return table
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/form-fields/" + sd.FormFields[0].ID,
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &formfieldapp.FormField{},
			ExpResp:    &sd.FormFields[0],
			CmpFunc: func(got, exp any) string {
				expResp, ok := exp.(*formfieldapp.FormField)
				if !ok {
					return "error occurred"
				}
				gotResp, ok := got.(*formfieldapp.FormField)
				if !ok {
					return "error occurred"
				}

				dbtest.NormalizeJSONFields(gotResp, expResp)
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
	return table
}

func queryByFormID200(sd apitest.SeedData) []apitest.Table {
	// Find all form fields that belong to the first form
	var expectedFields []formfieldapp.FormField
	for _, ff := range sd.FormFields {
		if ff.FormID == sd.Forms[0].ID {
			expectedFields = append(expectedFields, ff)
		}
	}

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/forms/" + sd.Forms[0].ID + "/form-fields",
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &formfieldapp.FormFields{},
			ExpResp:    &formfieldapp.FormFields{Fields: expectedFields},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*formfieldapp.FormFields)
				if !ok {
					return "error occurred"
				}
				expResp, ok := exp.(*formfieldapp.FormFields)
				if !ok {
					return "error occurred"
				}

				dbtest.NormalizeJSONFields(gotResp, expResp)
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
	return table
}
