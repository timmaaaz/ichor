package brandapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/core/brandapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/inventory/core/brands/%s", sd.Brands[1].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &brandapp.UpdateBrand{
				Name:          dbtest.StringPointer("Updated Brand"),
				ContactInfoID: dbtest.StringPointer(sd.ContactInfo[1].ID),
			},
			GotResp: &brandapp.Brand{},
			ExpResp: &brandapp.Brand{
				ID:            sd.Brands[1].ID,
				Name:          "Updated Brand",
				ContactInfoID: sd.ContactInfo[1].ID,
				CreatedDate:   sd.Brands[1].CreatedDate,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*brandapp.Brand)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*brandapp.Brand)
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "malformed-contact-info-uuid",
			URL:        fmt.Sprintf("/v1/inventory/core/brands/%s", sd.Brands[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &brandapp.UpdateBrand{
				ContactInfoID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"contact_info_id","error":"contact_info_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-brand-uuid",
			URL:        fmt.Sprintf("/v1/inventory/core/brands/%s", "not-a-uuid"),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &brandapp.UpdateBrand{
				ContactInfoID: &sd.ContactInfo[0].ID,
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
	table := []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        fmt.Sprintf("/v1/inventory/core/brands/%s", sd.Brands[0].ID),
			Token:      "&nbsp",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "badsig",
			URL:        fmt.Sprintf("/v1/inventory/core/brands/%s", sd.Brands[0].ID),
			Token:      sd.Users[0].Token + "A",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "roleadminonly",
			URL:        fmt.Sprintf("/v1/inventory/core/brands/%s", sd.Brands[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[USER]] rule[rule_admin_only]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func update404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "brand-dne",
			URL:        fmt.Sprintf("/v1/inventory/core/brands/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &brandapp.UpdateBrand{
				ContactInfoID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "brand not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "contact-info-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/inventory/core/brands/%s", sd.Brands[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &brandapp.UpdateBrand{
				ContactInfoID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
