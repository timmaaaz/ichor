package costhistoryapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/finance/costhistoryapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/finance/cost-history/%s", sd.CostHistory[0].CostHistoryID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &costhistoryapp.UpdateCostHistory{
				ProductID:     &sd.CostHistory[1].ProductID,
				CostType:      &sd.CostHistory[2].CostType,
				Amount:        &sd.CostHistory[3].Amount,
				EndDate:       &sd.CostHistory[6].EndDate,
				Currency:      &sd.CostHistory[4].Currency,
				EffectiveDate: &sd.CostHistory[12].EffectiveDate,
			},
			GotResp: &costhistoryapp.CostHistory{},
			ExpResp: &costhistoryapp.CostHistory{
				ProductID:     sd.CostHistory[1].ProductID,
				CostType:      sd.CostHistory[2].CostType,
				Amount:        sd.CostHistory[3].Amount,
				EndDate:       sd.CostHistory[6].EndDate,
				Currency:      sd.CostHistory[4].Currency,
				EffectiveDate: sd.CostHistory[12].EffectiveDate,
				CreatedDate:   sd.CostHistory[0].CreatedDate,
				CostHistoryID: sd.CostHistory[0].CostHistoryID,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*costhistoryapp.CostHistory)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*costhistoryapp.CostHistory)
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "malformed-product-uuid",
			URL:        fmt.Sprintf("/v1/finance/cost-history/%s", sd.CostHistory[0].CostHistoryID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &costhistoryapp.UpdateCostHistory{
				ProductID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"product_id","error":"product_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-product-cost-uuid",
			URL:        fmt.Sprintf("/v1/finance/cost-history/%s", "not-a-uuid"),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &costhistoryapp.UpdateCostHistory{
				ProductID: &sd.CostHistory[0].CostHistoryID,
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
			URL:        fmt.Sprintf("/v1/finance/cost-history/%s", sd.CostHistory[0].CostHistoryID),
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
			URL:        fmt.Sprintf("/v1/finance/cost-history/%s", sd.CostHistory[0].CostHistoryID),
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
			URL:        fmt.Sprintf("/v1/finance/cost-history/%s", sd.CostHistory[0].CostHistoryID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: cost_history"),
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
			Name:       "product-dne",
			URL:        fmt.Sprintf("/v1/finance/cost-history/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &costhistoryapp.UpdateCostHistory{
				ProductID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "costHistory not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "product-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/finance/cost-history/%s", sd.CostHistory[0].CostHistoryID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &costhistoryapp.UpdateCostHistory{
				ProductID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
