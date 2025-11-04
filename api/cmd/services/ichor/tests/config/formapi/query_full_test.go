package formapi_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/config/formapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func queryFull200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/forms/" + sd.Forms[0].ID + "/full",
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &formapp.FormFull{},
			ExpResp: &formapp.FormFull{
				ID:     sd.Forms[0].ID,
				Name:   sd.Forms[0].Name,
				Fields: sd.FormFields,
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func queryFull400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "invalid-uuid",
			URL:        "/v1/config/forms/invalid-uuid/full",
			Token:      sd.Users[0].Token,
			StatusCode: 400,
			Method:     "GET",
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "invalid UUID length: 12"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func queryFull401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "unauthorized",
			URL:        "/v1/config/forms/" + sd.Forms[0].ID + "/full",
			Token:      "&nbsp;",
			StatusCode: 401,
			Method:     "GET",
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func queryFull404(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "not-found",
			URL:        "/v1/config/forms/00000000-0000-0000-0000-000000000000/full",
			Token:      sd.Users[0].Token,
			StatusCode: 404,
			Method:     "GET",
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.NotFound, "form not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
