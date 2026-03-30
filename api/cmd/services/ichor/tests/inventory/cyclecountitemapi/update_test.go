package cyclecountitemapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountitemapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "count-item",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", sd.CycleCountItems[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &cyclecountitemapp.UpdateCycleCountItem{
				CountedQuantity: dbtest.StringPointer("8"),
			},
			GotResp: &cyclecountitemapp.CycleCountItem{},
			ExpResp: &cyclecountitemapp.CycleCountItem{
				ID:              sd.CycleCountItems[0].ID,
				SessionID:       sd.CycleCountItems[0].SessionID,
				ProductID:       sd.CycleCountItems[0].ProductID,
				LocationID:      sd.CycleCountItems[0].LocationID,
				SystemQuantity:  sd.CycleCountItems[0].SystemQuantity,
				CountedQuantity: "8",
				Status:          sd.CycleCountItems[0].Status,
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*cyclecountitemapp.CycleCountItem)
				expResp := exp.(*cyclecountitemapp.CycleCountItem)

				// Server-assigned fields
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				// Auto-injected fields when counted_quantity is set
				expResp.CountedBy = gotResp.CountedBy
				expResp.CountedDate = gotResp.CountedDate

				// Auto-computed variance
				expResp.Variance = gotResp.Variance

				return cmp.Diff(gotResp, expResp)
			},
		},
		{
			Name:       "status-to-counted",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", sd.CycleCountItems[1].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &cyclecountitemapp.UpdateCycleCountItem{
				Status: dbtest.StringPointer("counted"),
			},
			GotResp: &cyclecountitemapp.CycleCountItem{},
			ExpResp: &cyclecountitemapp.CycleCountItem{
				ID:              sd.CycleCountItems[1].ID,
				SessionID:       sd.CycleCountItems[1].SessionID,
				ProductID:       sd.CycleCountItems[1].ProductID,
				LocationID:      sd.CycleCountItems[1].LocationID,
				SystemQuantity:  sd.CycleCountItems[1].SystemQuantity,
				CountedQuantity: "",
				Variance:        "",
				Status:          "counted",
				CountedBy:       "",
				CountedDate:     "",
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*cyclecountitemapp.CycleCountItem)
				expResp := exp.(*cyclecountitemapp.CycleCountItem)
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "invalid-status",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", sd.CycleCountItems[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &cyclecountitemapp.UpdateCycleCountItem{
				Status: dbtest.StringPointer("not_a_valid_status"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `parse status: invalid status "not_a_valid_status"`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-token",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", sd.CycleCountItems[0].ID),
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
			Name:       "bad-sig",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", sd.CycleCountItems[0].ID),
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
			Name:       "no-update-permission",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", sd.CycleCountItems[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusForbidden,
			Input: &cyclecountitemapp.UpdateCycleCountItem{
				CountedQuantity: dbtest.StringPointer("5"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.PermissionDenied, "user does not have permission UPDATE for table: inventory.cycle_count_items"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &cyclecountitemapp.UpdateCycleCountItem{
				CountedQuantity: dbtest.StringPointer("5"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "cycle count item not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
