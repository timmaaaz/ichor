package putawaytaskapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/putawaytaskapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/put-away-tasks",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &putawaytaskapp.NewPutAwayTask{
				ProductID:       sd.Products[0].ProductID,
				LocationID:      sd.InventoryLocations[0].LocationID,
				Quantity:        "5",
				ReferenceNumber: "PO-CREATE-TEST-001",
			},
			GotResp: &putawaytaskapp.PutAwayTask{},
			ExpResp: &putawaytaskapp.PutAwayTask{
				ProductID:       sd.Products[0].ProductID,
				LocationID:      sd.InventoryLocations[0].LocationID,
				Quantity:        "5",
				ReferenceNumber: "PO-CREATE-TEST-001",
				Status:          "pending",
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*putawaytaskapp.PutAwayTask)
				if !exists {
					return "error occurred"
				}
				expResp := exp.(*putawaytaskapp.PutAwayTask)
				expResp.ID = gotResp.ID
				expResp.CreatedBy = gotResp.CreatedBy
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
			Name:       "missing-product-id",
			URL:        "/v1/inventory/put-away-tasks",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &putawaytaskapp.NewPutAwayTask{
				LocationID: sd.InventoryLocations[0].LocationID,
				Quantity:   "5",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"product_id","error":"product_id is a required field"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-location-id",
			URL:        "/v1/inventory/put-away-tasks",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &putawaytaskapp.NewPutAwayTask{
				ProductID: sd.Products[0].ProductID,
				Quantity:  "5",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"location_id","error":"location_id is a required field"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-quantity",
			URL:        "/v1/inventory/put-away-tasks",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &putawaytaskapp.NewPutAwayTask{
				ProductID:  sd.Products[0].ProductID,
				LocationID: sd.InventoryLocations[0].LocationID,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"quantity","error":"quantity is a required field"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "invalid-product-fk",
			URL:        "/v1/inventory/put-away-tasks",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &putawaytaskapp.NewPutAwayTask{
				ProductID:  uuid.NewString(),
				LocationID: sd.InventoryLocations[0].LocationID,
				Quantity:   "5",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "invalid-location-fk",
			URL:        "/v1/inventory/put-away-tasks",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &putawaytaskapp.NewPutAwayTask{
				ProductID:  sd.Products[0].ProductID,
				LocationID: uuid.NewString(),
				Quantity:   "5",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
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
			URL:        "/v1/inventory/put-away-tasks",
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
			URL:        "/v1/inventory/put-away-tasks",
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
			URL:        "/v1/inventory/put-away-tasks",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: inventory.put_away_tasks"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
