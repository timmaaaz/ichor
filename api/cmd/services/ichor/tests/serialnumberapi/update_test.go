package serialnumber_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"

	"github.com/timmaaaz/ichor/app/domain/lots/serialnumberapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/lots/serialnumber/%s", sd.SerialNumbers[0].SerialID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &serialnumberapp.UpdateSerialNumber{
				LotID:        &sd.LotTrackings[0].LotID,
				ProductID:    &sd.Products[0].ProductID,
				LocationID:   &sd.InventoryLocations[0].LocationID,
				SerialNumber: dbtest.StringPointer("UpdateSerialNumber"),
				Status:       dbtest.StringPointer("active"),
			},
			GotResp: &serialnumberapp.SerialNumber{},
			ExpResp: &serialnumberapp.SerialNumber{
				LotID:        sd.LotTrackings[0].LotID,
				ProductID:    sd.Products[0].ProductID,
				LocationID:   sd.InventoryLocations[0].LocationID,
				SerialNumber: "UpdateSerialNumber",
				Status:       "active",
				CreatedDate:  sd.SerialNumbers[0].CreatedDate,
				SerialID:     sd.SerialNumbers[0].SerialID,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*serialnumberapp.SerialNumber)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*serialnumberapp.SerialNumber)
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "malformed-lot-uuid",
			URL:        fmt.Sprintf("/v1/lots/serialnumber/%s", sd.SerialNumbers[0].SerialID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &serialnumberapp.UpdateSerialNumber{
				LotID:        dbtest.StringPointer("not-a-uuid"),
				ProductID:    &sd.Products[0].ProductID,
				LocationID:   &sd.InventoryLocations[0].LocationID,
				SerialNumber: dbtest.StringPointer("UpdateSerialNumber"),
				Status:       dbtest.StringPointer("active"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"lot_id","error":"lot_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-product-uuid",
			URL:        fmt.Sprintf("/v1/lots/serialnumber/%s", sd.SerialNumbers[0].SerialID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &serialnumberapp.UpdateSerialNumber{
				LotID:        &sd.LotTrackings[0].LotID,
				ProductID:    dbtest.StringPointer("not-a-uuid"),
				LocationID:   &sd.InventoryLocations[0].LocationID,
				SerialNumber: dbtest.StringPointer("UpdateSerialNumber"),
				Status:       dbtest.StringPointer("active"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"product_id","error":"product_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-location-uuid",
			URL:        fmt.Sprintf("/v1/lots/serialnumber/%s", sd.SerialNumbers[0].SerialID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &serialnumberapp.UpdateSerialNumber{
				LotID:        &sd.LotTrackings[0].LotID,
				ProductID:    &sd.Products[0].ProductID,
				LocationID:   dbtest.StringPointer("not-a-uuid"),
				SerialNumber: dbtest.StringPointer("UpdateSerialNumber"),
				Status:       dbtest.StringPointer("active"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"location_id","error":"location_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-serial-uuid",
			URL:        fmt.Sprintf("/v1/lots/serialnumber/%s", "not-a-uuid"),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &serialnumberapp.UpdateSerialNumber{
				LotID: &sd.LotTrackings[0].LotID,
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
			URL:        fmt.Sprintf("/v1/lots/serialnumber/%s", sd.SerialNumbers[0].SerialID),
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
			URL:        fmt.Sprintf("/v1/lots/serialnumber/%s", sd.SerialNumbers[0].SerialID),
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
			URL:        fmt.Sprintf("/v1/lots/serialnumber/%s", sd.SerialNumbers[0].SerialID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: serial_numbers"),
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
			Name:       "supplier-dne",
			URL:        fmt.Sprintf("/v1/lots/serialnumber/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &serialnumberapp.UpdateSerialNumber{
				LotID:        &sd.LotTrackings[0].LotID,
				ProductID:    &sd.Products[0].ProductID,
				LocationID:   &sd.InventoryLocations[0].LocationID,
				SerialNumber: dbtest.StringPointer("UpdateSerialNumber"),
				Status:       dbtest.StringPointer("active"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "querying by ID: namedquerystruct: serialNumber not found"),
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
			URL:        fmt.Sprintf("/v1/lots/serialnumber/%s", sd.SerialNumbers[0].SerialID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &serialnumberapp.UpdateSerialNumber{
				LotID:        dbtest.StringPointer(uuid.NewString()),
				ProductID:    &sd.Products[0].ProductID,
				LocationID:   &sd.InventoryLocations[0].LocationID,
				SerialNumber: dbtest.StringPointer("UpdateSerialNumber"),
				Status:       dbtest.StringPointer("active"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "update: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "product-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/lots/serialnumber/%s", sd.SerialNumbers[0].SerialID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &serialnumberapp.UpdateSerialNumber{
				LotID:        &sd.LotTrackings[0].LotID,
				ProductID:    dbtest.StringPointer(uuid.NewString()),
				LocationID:   &sd.InventoryLocations[0].LocationID,
				SerialNumber: dbtest.StringPointer("UpdateSerialNumber"),
				Status:       dbtest.StringPointer("active"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "update: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "location-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/lots/serialnumber/%s", sd.SerialNumbers[0].SerialID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &serialnumberapp.UpdateSerialNumber{
				LotID:        &sd.LotTrackings[0].LotID,
				ProductID:    &sd.Products[0].ProductID,
				LocationID:   dbtest.StringPointer(uuid.NewString()),
				SerialNumber: dbtest.StringPointer("UpdateSerialNumber"),
				Status:       dbtest.StringPointer("active"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "update: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
