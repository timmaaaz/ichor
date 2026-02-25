package lotlocationapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/lotlocationapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/inventory/lot-locations/%s", sd.LotLocations[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &lotlocationapp.UpdateLotLocation{
				Quantity: dbtest.StringPointer("99"),
			},
			GotResp: &lotlocationapp.LotLocation{},
			ExpResp: &lotlocationapp.LotLocation{
				ID:          sd.LotLocations[0].ID,
				LotID:       sd.LotLocations[0].LotID,
				LocationID:  sd.LotLocations[0].LocationID,
				Quantity:    "99",
				CreatedDate: sd.LotLocations[0].CreatedDate,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*lotlocationapp.LotLocation)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*lotlocationapp.LotLocation)
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "malformed-lot-id",
			URL:        fmt.Sprintf("/v1/inventory/lot-locations/%s", sd.LotLocations[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &lotlocationapp.UpdateLotLocation{
				LotID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"lot_id","error":"lot_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-lot-location-uuid",
			URL:        fmt.Sprintf("/v1/inventory/lot-locations/%s", "not-a-uuid"),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &lotlocationapp.UpdateLotLocation{
				Quantity: dbtest.StringPointer("10"),
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
	return []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        fmt.Sprintf("/v1/inventory/lot-locations/%s", sd.LotLocations[0].ID),
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
			URL:        fmt.Sprintf("/v1/inventory/lot-locations/%s", sd.LotLocations[0].ID),
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
			URL:        fmt.Sprintf("/v1/inventory/lot-locations/%s", sd.LotLocations[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: inventory.lot_locations"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "lot-location-dne",
			URL:        fmt.Sprintf("/v1/inventory/lot-locations/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &lotlocationapp.UpdateLotLocation{
				Quantity: dbtest.StringPointer("10"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "lot location not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "lot-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/inventory/lot-locations/%s", sd.LotLocations[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &lotlocationapp.UpdateLotLocation{
				LotID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "update: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
