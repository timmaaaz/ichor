package brandapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/core/brandapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/core/brands",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &brandapp.NewBrand{
				Name:          "Test Brand",
				ContactInfoID: sd.ContactInfo[0].ID,
			},
			GotResp: &brandapp.Brand{},
			ExpResp: &brandapp.Brand{
				Name:          "Test Brand",
				ContactInfoID: sd.ContactInfo[0].ID,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*brandapp.Brand)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*brandapp.Brand)
				expResp.ID = gotResp.ID
				expResp.UpdatedDate = gotResp.UpdatedDate
				expResp.CreatedDate = gotResp.CreatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing-name",
			URL:        "/v1/inventory/core/brands",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &brandapp.NewBrand{
				ContactInfoID: sd.ContactInfo[0].ID,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"name\",\"error\":\"name is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-contact-info-id",
			URL:        "/v1/inventory/core/brands",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &brandapp.NewBrand{
				Name: "test name",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"contact_infos_id\",\"error\":\"contact_infos_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-contact-info-id",
			URL:        "/v1/inventory/core/brands",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &brandapp.NewBrand{
				Name:          "test name",
				ContactInfoID: "not-a-uuid",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"contact_infos_id","error":"contact_infos_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "contact-info-not-valid-fk",
			URL:        "/v1/inventory/core/brands",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &brandapp.NewBrand{
				Name:          "test name",
				ContactInfoID: uuid.NewString(),
			},
			ExpResp: errs.Newf(errs.Aborted, "foreign key violation"),
			GotResp: &errs.Error{},
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
			URL:        "/v1/inventory/core/brands",
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
			URL:        "/v1/inventory/core/brands",
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
			URL:        "/v1/inventory/core/brands",
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
			URL:        "/v1/inventory/core/brands",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: brands"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
