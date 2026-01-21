package currencyapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/currencyapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/currencies",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: &currencyapp.NewCurrency{
				Code:          "TST",
				Name:          "Test Currency",
				Symbol:        "T$",
				Locale:        "en-US",
				DecimalPlaces: 2,
				IsActive:      true,
				SortOrder:     999,
			},
			GotResp: &currencyapp.Currency{},
			ExpResp: &currencyapp.Currency{
				Code:          "TST",
				Name:          "Test Currency",
				Symbol:        "T$",
				Locale:        "en-US",
				DecimalPlaces: 2,
				IsActive:      true,
				SortOrder:     999,
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*currencyapp.Currency)
				expResp := exp.(*currencyapp.Currency)

				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func create400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "missing-code",
			URL:        "/v1/core/currencies",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: &currencyapp.NewCurrency{
				Name:          "Test Currency",
				Symbol:        "T$",
				Locale:        "en-US",
				DecimalPlaces: 2,
				IsActive:      true,
				SortOrder:     999,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"code\",\"error\":\"code is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "invalid-code-length",
			URL:        "/v1/core/currencies",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: &currencyapp.NewCurrency{
				Code:          "TOOLONG",
				Name:          "Test Currency",
				Symbol:        "T$",
				Locale:        "en-US",
				DecimalPlaces: 2,
				IsActive:      true,
				SortOrder:     999,
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
