package currencyapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/currencyapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/currencies/" + sd.Currencies[0].ID.String(),
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPut,
			Input: &currencyapp.UpdateCurrency{
				Name: dbtest.StringPointer("Updated Currency Name"),
			},
			GotResp: &currencyapp.Currency{},
			ExpResp: &currencyapp.Currency{
				ID:            sd.Currencies[0].ID.String(),
				Code:          sd.Currencies[0].Code,
				Name:          "Updated Currency Name",
				Symbol:        sd.Currencies[0].Symbol,
				Locale:        sd.Currencies[0].Locale,
				DecimalPlaces: sd.Currencies[0].DecimalPlaces,
				IsActive:      sd.Currencies[0].IsActive,
				SortOrder:     sd.Currencies[0].SortOrder,
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*currencyapp.Currency)
				expResp := exp.(*currencyapp.Currency)

				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate
				expResp.CreatedBy = gotResp.CreatedBy
				expResp.UpdatedBy = gotResp.UpdatedBy

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func update400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "invalid-code-length",
			URL:        "/v1/core/currencies/" + sd.Currencies[0].ID.String(),
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPut,
			Input: &currencyapp.UpdateCurrency{
				Code: dbtest.StringPointer("TOOLONG"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"code\",\"error\":\"code must be 3 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
