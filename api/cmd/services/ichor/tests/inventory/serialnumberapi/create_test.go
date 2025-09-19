package serialnumber_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"

	"github.com/timmaaaz/ichor/app/domain/inventory/serialnumberapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/lots/serial-number",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &serialnumberapp.NewSerialNumber{
				LotID:        sd.LotTrackings[0].LotID,
				ProductID:    sd.Products[0].ProductID,
				LocationID:   sd.InventoryLocations[0].LocationID,
				SerialNumber: "SN-1234567890",
				Status:       "active",
			},
			GotResp: &serialnumberapp.SerialNumber{},
			ExpResp: &serialnumberapp.SerialNumber{
				LotID:        sd.LotTrackings[0].LotID,
				ProductID:    sd.Products[0].ProductID,
				LocationID:   sd.InventoryLocations[0].LocationID,
				SerialNumber: "SN-1234567890",
				Status:       "active",
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*serialnumberapp.SerialNumber)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*serialnumberapp.SerialNumber)
				expResp.SerialID = gotResp.SerialID
				expResp.UpdatedDate = gotResp.UpdatedDate
				expResp.CreatedDate = gotResp.CreatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing-lot-id",
			URL:        "/v1/lots/serial-number",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &serialnumberapp.NewSerialNumber{

				ProductID:    sd.Products[0].ProductID,
				LocationID:   sd.InventoryLocations[0].LocationID,
				SerialNumber: "SN-1234567890",
				Status:       "active",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"lot_id\",\"error\":\"lot_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-product-id",
			URL:        "/v1/lots/serial-number",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &serialnumberapp.NewSerialNumber{
				LotID:        sd.LotTrackings[0].LotID,
				LocationID:   sd.InventoryLocations[0].LocationID,
				SerialNumber: "SN-1234567890",
				Status:       "active",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"product_id\",\"error\":\"product_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-location-id",
			URL:        "/v1/lots/serial-number",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &serialnumberapp.NewSerialNumber{
				LotID:     sd.LotTrackings[0].LotID,
				ProductID: sd.Products[0].ProductID,

				SerialNumber: "SN-1234567890",
				Status:       "active",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"location_id\",\"error\":\"location_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-serial-number",
			URL:        "/v1/lots/serial-number",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &serialnumberapp.NewSerialNumber{
				LotID:      sd.LotTrackings[0].LotID,
				ProductID:  sd.Products[0].ProductID,
				LocationID: sd.InventoryLocations[0].LocationID,

				Status: "active",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"serial_number\",\"error\":\"serial_number is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-status",
			URL:        "/v1/lots/serial-number",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &serialnumberapp.NewSerialNumber{
				LotID:        sd.LotTrackings[0].LotID,
				ProductID:    sd.Products[0].ProductID,
				LocationID:   sd.InventoryLocations[0].LocationID,
				SerialNumber: "SN-1234567890",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"status\",\"error\":\"status is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},

		{
			Name:       "malformed-lot-id",
			URL:        "/v1/lots/serial-number",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &serialnumberapp.NewSerialNumber{
				LotID:        "not-a-uuid",
				ProductID:    sd.Products[0].ProductID,
				LocationID:   sd.InventoryLocations[0].LocationID,
				SerialNumber: "SN-1234567890",
				Status:       "active",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"lot_id\",\"error\":\"lot_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-location-id",
			URL:        "/v1/lots/serial-number",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &serialnumberapp.NewSerialNumber{
				LotID:        sd.LotTrackings[0].LotID,
				ProductID:    sd.Products[0].ProductID,
				LocationID:   "not-a-uuid",
				SerialNumber: "SN-1234567890",
				Status:       "active",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"location_id\",\"error\":\"location_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-product-id",
			URL:        "/v1/lots/serial-number",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &serialnumberapp.NewSerialNumber{
				LotID:        sd.LotTrackings[0].LotID,
				ProductID:    "not-a-uuid",
				LocationID:   sd.InventoryLocations[0].LocationID,
				SerialNumber: "SN-1234567890",
				Status:       "active",
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
	return []apitest.Table{
		{
			Name:       "lot-id-not-valid-fk",
			URL:        "/v1/lots/serial-number",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &serialnumberapp.NewSerialNumber{
				LotID:        uuid.NewString(),
				ProductID:    sd.Products[0].ProductID,
				LocationID:   sd.InventoryLocations[0].LocationID,
				SerialNumber: "SN-1234567890",
				Status:       "active",
			},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "product-id-not-valid-fk",
			URL:        "/v1/lots/serial-number",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &serialnumberapp.NewSerialNumber{
				LotID:        sd.LotTrackings[0].LotID,
				ProductID:    uuid.NewString(),
				LocationID:   sd.InventoryLocations[0].LocationID,
				SerialNumber: "SN-1234567890",
				Status:       "active",
			},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "location-id-not-valid-fk",
			URL:        "/v1/lots/serial-number",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &serialnumberapp.NewSerialNumber{
				LotID:        sd.LotTrackings[0].LotID,
				ProductID:    sd.Products[0].ProductID,
				LocationID:   uuid.NewString(),
				SerialNumber: "SN-1234567890",
				Status:       "active",
			},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
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
			URL:        "/v1/lots/serial-number",
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
			URL:        "/v1/lots/serial-number",
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
			URL:        "/v1/lots/serial-number",
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
			URL:        "/v1/lots/serial-number",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: serial_numbers"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
