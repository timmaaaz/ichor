package costhistoryapi_test

import (
	"net/http"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/products/costhistoryapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func create200(sd apitest.SeedData) []apitest.Table {

	now := time.Now().Format(timeutil.FORMAT)
	later := time.Now().AddDate(0, 3, 0).Format(timeutil.FORMAT)

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/products/cost-history",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &costhistoryapp.NewCostHistory{
				ProductID:     sd.Products[0].ProductID,
				CostType:      "TYPE 1",
				Currency:      "USD",
				Amount:        "33.33",
				EndDate:       later,
				EffectiveDate: now,
			},
			GotResp: &costhistoryapp.CostHistory{},
			ExpResp: &costhistoryapp.CostHistory{
				ProductID:     sd.Products[0].ProductID,
				CostType:      "TYPE 1",
				Currency:      "USD",
				Amount:        "33.33",
				EndDate:       later,
				EffectiveDate: now,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*costhistoryapp.CostHistory)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*costhistoryapp.CostHistory)
				expResp.CostHistoryID = gotResp.CostHistoryID
				expResp.UpdatedDate = gotResp.UpdatedDate
				expResp.CreatedDate = gotResp.CreatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create400(sd apitest.SeedData) []apitest.Table {
	now := time.Now().Format(timeutil.FORMAT)
	later := time.Now().AddDate(0, 3, 0).Format(timeutil.FORMAT)

	return []apitest.Table{
		{
			Name:       "missing-product-id",
			URL:        "/v1/products/cost-history",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &costhistoryapp.NewCostHistory{
				CostType:      "TYPE 1",
				Currency:      "USD",
				Amount:        "33.33",
				EndDate:       later,
				EffectiveDate: now,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"product_id\",\"error\":\"product_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-cost-type",
			URL:        "/v1/products/cost-history",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &costhistoryapp.NewCostHistory{
				ProductID:     sd.Products[0].ProductID,
				Currency:      "USD",
				Amount:        "33.33",
				EndDate:       later,
				EffectiveDate: now,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"cost_type\",\"error\":\"cost_type is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-currency",
			URL:        "/v1/products/cost-history",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &costhistoryapp.NewCostHistory{
				ProductID:     sd.Products[0].ProductID,
				CostType:      "TYPE 1",
				Amount:        "33.33",
				EndDate:       later,
				EffectiveDate: now,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"currency\",\"error\":\"currency is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-amount",
			URL:        "/v1/products/cost-history",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &costhistoryapp.NewCostHistory{
				ProductID:     sd.Products[0].ProductID,
				CostType:      "TYPE 1",
				Currency:      "USD",
				EndDate:       later,
				EffectiveDate: now,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"amount\",\"error\":\"amount is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-end-date",
			URL:        "/v1/products/cost-history",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &costhistoryapp.NewCostHistory{
				ProductID:     sd.Products[0].ProductID,
				CostType:      "TYPE 1",
				Currency:      "USD",
				Amount:        "33.33",
				EffectiveDate: now,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"end_date\",\"error\":\"end_date is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-effective-date",
			URL:        "/v1/products/cost-history",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &costhistoryapp.NewCostHistory{
				ProductID: sd.Products[0].ProductID,
				CostType:  "TYPE 1",
				Currency:  "USD",
				Amount:    "33.33",
				EndDate:   later,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"effective_date\",\"error\":\"effective_date is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},

		{
			Name:       "malformed-product-id",
			URL:        "/v1/products/cost-history",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &costhistoryapp.NewCostHistory{
				ProductID:     "not a uuid",
				CostType:      "TYPE 1",
				Currency:      "USD",
				Amount:        "33.33",
				EndDate:       later,
				EffectiveDate: now,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"product_id\",\"error\":\"product_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create409(sd apitest.SeedData) []apitest.Table {
	now := time.Now().Format(timeutil.FORMAT)
	later := time.Now().AddDate(0, 3, 0).Format(timeutil.FORMAT)
	return []apitest.Table{
		{
			Name:       "product-id-not-valid-fk",
			URL:        "/v1/products/cost-history",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &costhistoryapp.NewCostHistory{
				ProductID:     uuid.NewString(),
				CostType:      "TYPE 1",
				Currency:      "USD",
				Amount:        "33.33",
				EndDate:       later,
				EffectiveDate: now,
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
			URL:        "/v1/products/cost-history",
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
			URL:        "/v1/products/cost-history",
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
			URL:        "/v1/products/cost-history",
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
			URL:        "/v1/products/cost-history",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: cost_history"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
