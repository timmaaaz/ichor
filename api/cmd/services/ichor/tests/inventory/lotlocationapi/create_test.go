package lotlocationapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/lotlocationapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/lot-locations",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &lotlocationapp.NewLotLocation{
				LotID:      sd.LotTrackings[0].LotID,
				LocationID: sd.InventoryLocations[0].LocationID,
				Quantity:   "10",
			},
			GotResp: &lotlocationapp.LotLocation{},
			ExpResp: &lotlocationapp.LotLocation{
				LotID:      sd.LotTrackings[0].LotID,
				LocationID: sd.InventoryLocations[0].LocationID,
				Quantity:   "10",
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*lotlocationapp.LotLocation)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*lotlocationapp.LotLocation)
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
			Name:       "missing-lot-id",
			URL:        "/v1/inventory/lot-locations",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &lotlocationapp.NewLotLocation{
				LocationID: sd.InventoryLocations[0].LocationID,
				Quantity:   "10",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"lot_id\",\"error\":\"lot_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-location-id",
			URL:        "/v1/inventory/lot-locations",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &lotlocationapp.NewLotLocation{
				LotID:    sd.LotTrackings[0].LotID,
				Quantity: "10",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"location_id\",\"error\":\"location_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-quantity",
			URL:        "/v1/inventory/lot-locations",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &lotlocationapp.NewLotLocation{
				LotID:      sd.LotTrackings[0].LotID,
				LocationID: sd.InventoryLocations[0].LocationID,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"quantity\",\"error\":\"quantity is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-lot-id",
			URL:        "/v1/inventory/lot-locations",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &lotlocationapp.NewLotLocation{
				LotID:      "not-a-uuid",
				LocationID: sd.InventoryLocations[0].LocationID,
				Quantity:   "10",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"lot_id\",\"error\":\"lot_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-location-id",
			URL:        "/v1/inventory/lot-locations",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &lotlocationapp.NewLotLocation{
				LotID:      sd.LotTrackings[0].LotID,
				LocationID: "not-a-uuid",
				Quantity:   "10",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"location_id\",\"error\":\"location_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty token",
			URL:        "/v1/inventory/lot-locations",
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
			URL:        "/v1/inventory/lot-locations",
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
			URL:        "/v1/inventory/lot-locations",
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
			URL:        "/v1/inventory/lot-locations",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: inventory.lot_locations"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "lot-id-not-valid-fk",
			URL:        "/v1/inventory/lot-locations",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &lotlocationapp.NewLotLocation{
				LotID:      uuid.New().String(),
				LocationID: sd.InventoryLocations[0].LocationID,
				Quantity:   "10",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
