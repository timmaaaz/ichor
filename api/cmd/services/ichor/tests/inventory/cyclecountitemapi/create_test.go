package cyclecountitemapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountitemapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/cycle-count-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &cyclecountitemapp.NewCycleCountItem{
				SessionID:      sd.CycleCountSessions[0].ID,
				ProductID:      sd.Products[0].ProductID,
				LocationID:     sd.InventoryLocations[0].LocationID,
				SystemQuantity: "50",
			},
			GotResp: &cyclecountitemapp.CycleCountItem{},
			ExpResp: &cyclecountitemapp.CycleCountItem{
				SessionID:       sd.CycleCountSessions[0].ID,
				ProductID:       sd.Products[0].ProductID,
				LocationID:      sd.InventoryLocations[0].LocationID,
				SystemQuantity:  "50",
				CountedQuantity: "",
				Variance:        "",
				Status:          "pending",
				CountedBy:       "",
				CountedDate:     "",
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*cyclecountitemapp.CycleCountItem)
				expResp := exp.(*cyclecountitemapp.CycleCountItem)

				// Server-assigned fields
				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing-session-id",
			URL:        "/v1/inventory/cycle-count-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &cyclecountitemapp.NewCycleCountItem{
				SessionID:      "",
				ProductID:      sd.Products[0].ProductID,
				LocationID:     sd.InventoryLocations[0].LocationID,
				SystemQuantity: "50",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"sessionId","error":"sessionId is a required field"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-product-id",
			URL:        "/v1/inventory/cycle-count-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &cyclecountitemapp.NewCycleCountItem{
				SessionID:      sd.CycleCountSessions[0].ID,
				ProductID:      "",
				LocationID:     sd.InventoryLocations[0].LocationID,
				SystemQuantity: "50",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"productId","error":"productId is a required field"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-location-id",
			URL:        "/v1/inventory/cycle-count-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &cyclecountitemapp.NewCycleCountItem{
				SessionID:      sd.CycleCountSessions[0].ID,
				ProductID:      sd.Products[0].ProductID,
				LocationID:     "",
				SystemQuantity: "50",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"locationId","error":"locationId is a required field"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-system-quantity",
			URL:        "/v1/inventory/cycle-count-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &cyclecountitemapp.NewCycleCountItem{
				SessionID:      sd.CycleCountSessions[0].ID,
				ProductID:      sd.Products[0].ProductID,
				LocationID:     sd.InventoryLocations[0].LocationID,
				SystemQuantity: "",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"systemQuantity","error":"systemQuantity is a required field"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-token",
			URL:        "/v1/inventory/cycle-count-items",
			Token:      "&nbsp;",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-sig",
			URL:        "/v1/inventory/cycle-count-items",
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "no-create-permission",
			URL:        "/v1/inventory/cycle-count-items",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusForbidden,
			Input: &cyclecountitemapp.NewCycleCountItem{
				SessionID:      sd.CycleCountSessions[0].ID,
				ProductID:      sd.Products[0].ProductID,
				LocationID:     sd.InventoryLocations[0].LocationID,
				SystemQuantity: "50",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.PermissionDenied, "user does not have permission CREATE for table: inventory.cycle_count_items"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "fk-violation-bad-session",
			URL:        "/v1/inventory/cycle-count-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &cyclecountitemapp.NewCycleCountItem{
				SessionID:      uuid.NewString(),
				ProductID:      sd.Products[0].ProductID,
				LocationID:     sd.InventoryLocations[0].LocationID,
				SystemQuantity: "50",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
