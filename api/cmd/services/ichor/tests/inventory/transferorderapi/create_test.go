package transferorderapi_test

import (
	"net/http"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/foundation/timeutil"

	"github.com/timmaaaz/ichor/app/domain/inventory/transferorderapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {

	now := time.Now()

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/transfer-orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &transferorderapp.NewTransferOrder{
				ProductID:      sd.Products[0].ProductID,
				FromLocationID: sd.InventoryLocations[0].LocationID,
				ToLocationID:   sd.InventoryLocations[2].LocationID,
				RequestedByID:  sd.TransferOrders[0].RequestedByID,
				ApprovedByID:   sd.TransferOrders[1].ApprovedByID,
				Quantity:       "10",
				Status:         "pending",
				TransferDate:   now.Format(timeutil.FORMAT),
			},
			GotResp: &transferorderapp.TransferOrder{},
			ExpResp: &transferorderapp.TransferOrder{
				ProductID:      sd.Products[0].ProductID,
				FromLocationID: sd.InventoryLocations[0].LocationID,
				ToLocationID:   sd.InventoryLocations[2].LocationID,
				RequestedByID:  sd.TransferOrders[0].RequestedByID,
				ApprovedByID:   sd.TransferOrders[1].ApprovedByID,
				Quantity:       "10",
				Status:         "pending",
				TransferDate:   now.Format(timeutil.FORMAT),
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*transferorderapp.TransferOrder)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*transferorderapp.TransferOrder)
				expResp.TransferID = gotResp.TransferID
				expResp.UpdatedDate = gotResp.UpdatedDate
				expResp.CreatedDate = gotResp.CreatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create400(sd apitest.SeedData) []apitest.Table {

	now := time.Now()
	return []apitest.Table{
		{
			Name:       "missing-product-id",
			URL:        "/v1/inventory/transfer-orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &transferorderapp.NewTransferOrder{
				FromLocationID: sd.InventoryLocations[0].LocationID,
				ToLocationID:   sd.InventoryLocations[2].LocationID,
				RequestedByID:  sd.TransferOrders[0].RequestedByID,
				ApprovedByID:   sd.TransferOrders[1].ApprovedByID,
				Quantity:       "10",
				Status:         "pending",
				TransferDate:   now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"product_id\",\"error\":\"product_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-from-location-id",
			URL:        "/v1/inventory/transfer-orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &transferorderapp.NewTransferOrder{
				ProductID:     sd.Products[0].ProductID,
				ToLocationID:  sd.InventoryLocations[2].LocationID,
				RequestedByID: sd.TransferOrders[0].RequestedByID,
				ApprovedByID:  sd.TransferOrders[1].ApprovedByID,
				Quantity:      "10",
				Status:        "pending",
				TransferDate:  now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"from_location_id\",\"error\":\"from_location_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-to-location-id",
			URL:        "/v1/inventory/transfer-orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &transferorderapp.NewTransferOrder{
				ProductID:      sd.Products[0].ProductID,
				FromLocationID: sd.InventoryLocations[0].LocationID,
				RequestedByID:  sd.TransferOrders[0].RequestedByID,
				ApprovedByID:   sd.TransferOrders[1].ApprovedByID,
				Quantity:       "10",
				Status:         "pending",
				TransferDate:   now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"to_location_id\",\"error\":\"to_location_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-requested-by-id",
			URL:        "/v1/inventory/transfer-orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &transferorderapp.NewTransferOrder{
				ProductID:      sd.Products[0].ProductID,
				FromLocationID: sd.InventoryLocations[0].LocationID,
				ToLocationID:   sd.InventoryLocations[2].LocationID,
				ApprovedByID:   sd.TransferOrders[1].ApprovedByID,
				Quantity:       "10",
				Status:         "pending",
				TransferDate:   now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"requested_by\",\"error\":\"requested_by is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-approved-by-id",
			URL:        "/v1/inventory/transfer-orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &transferorderapp.NewTransferOrder{
				ProductID:      sd.Products[0].ProductID,
				FromLocationID: sd.InventoryLocations[0].LocationID,
				ToLocationID:   sd.InventoryLocations[2].LocationID,
				RequestedByID:  sd.TransferOrders[0].RequestedByID,
				Quantity:       "10",
				Status:         "pending",
				TransferDate:   now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"approved_by\",\"error\":\"approved_by is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-quantity",
			URL:        "/v1/inventory/transfer-orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &transferorderapp.NewTransferOrder{
				ProductID:      sd.Products[0].ProductID,
				FromLocationID: sd.InventoryLocations[0].LocationID,
				ToLocationID:   sd.InventoryLocations[2].LocationID,
				RequestedByID:  sd.TransferOrders[0].RequestedByID,
				ApprovedByID:   sd.TransferOrders[1].ApprovedByID,
				Status:         "pending",
				TransferDate:   now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"quantity\",\"error\":\"quantity is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-status",
			URL:        "/v1/inventory/transfer-orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &transferorderapp.NewTransferOrder{
				ProductID:      sd.Products[0].ProductID,
				FromLocationID: sd.InventoryLocations[0].LocationID,
				ToLocationID:   sd.InventoryLocations[2].LocationID,
				RequestedByID:  sd.TransferOrders[0].RequestedByID,
				ApprovedByID:   sd.TransferOrders[1].ApprovedByID,
				Quantity:       "10",
				TransferDate:   now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"status\",\"error\":\"status is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-transfer-date",
			URL:        "/v1/inventory/transfer-orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &transferorderapp.NewTransferOrder{
				ProductID:      sd.Products[0].ProductID,
				FromLocationID: sd.InventoryLocations[0].LocationID,
				ToLocationID:   sd.InventoryLocations[2].LocationID,
				RequestedByID:  sd.TransferOrders[0].RequestedByID,
				ApprovedByID:   sd.TransferOrders[1].ApprovedByID,
				Quantity:       "10",
				Status:         "pending",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"transfer_date\",\"error\":\"transfer_date is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},

		{
			Name:       "malformed-product-id",
			URL:        "/v1/inventory/transfer-orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &transferorderapp.NewTransferOrder{
				ProductID:      "not-a-uuid",
				FromLocationID: sd.InventoryLocations[0].LocationID,
				ToLocationID:   sd.InventoryLocations[2].LocationID,
				RequestedByID:  sd.TransferOrders[0].RequestedByID,
				ApprovedByID:   sd.TransferOrders[1].ApprovedByID,
				Quantity:       "10",
				Status:         "pending",
				TransferDate:   now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"product_id\",\"error\":\"product_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-from-location-id",
			URL:        "/v1/inventory/transfer-orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &transferorderapp.NewTransferOrder{
				ProductID:      sd.Products[0].ProductID,
				FromLocationID: "not-a-uuid",
				ToLocationID:   sd.InventoryLocations[2].LocationID,
				RequestedByID:  sd.TransferOrders[0].RequestedByID,
				ApprovedByID:   sd.TransferOrders[1].ApprovedByID,
				Quantity:       "10",
				Status:         "pending",
				TransferDate:   now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"from_location_id\",\"error\":\"from_location_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-to-location-id",
			URL:        "/v1/inventory/transfer-orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &transferorderapp.NewTransferOrder{
				ProductID:      sd.Products[0].ProductID,
				FromLocationID: sd.InventoryLocations[0].LocationID,
				ToLocationID:   "not-a-uuid",
				RequestedByID:  sd.TransferOrders[0].RequestedByID,
				ApprovedByID:   sd.TransferOrders[1].ApprovedByID,
				Quantity:       "10",
				Status:         "pending",
				TransferDate:   now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"to_location_id\",\"error\":\"to_location_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-request-by-id",
			URL:        "/v1/inventory/transfer-orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &transferorderapp.NewTransferOrder{
				ProductID:      sd.Products[0].ProductID,
				FromLocationID: sd.InventoryLocations[0].LocationID,
				ToLocationID:   sd.InventoryLocations[2].LocationID,
				RequestedByID:  "not-a-uuid",
				ApprovedByID:   sd.TransferOrders[1].ApprovedByID,
				Quantity:       "10",
				Status:         "pending",
				TransferDate:   now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"requested_by\",\"error\":\"requested_by must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-approved-by-id",
			URL:        "/v1/inventory/transfer-orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &transferorderapp.NewTransferOrder{
				ProductID:      sd.Products[0].ProductID,
				FromLocationID: sd.InventoryLocations[0].LocationID,
				ToLocationID:   sd.InventoryLocations[2].LocationID,
				RequestedByID:  sd.TransferOrders[0].RequestedByID,
				ApprovedByID:   "not-a-uuid",
				Quantity:       "10",
				Status:         "pending",
				TransferDate:   now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"approved_by\",\"error\":\"approved_by must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create409(sd apitest.SeedData) []apitest.Table {

	now := time.Now()

	return []apitest.Table{
		{
			Name:       "product-id-not-valid-fk",
			URL:        "/v1/inventory/transfer-orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &transferorderapp.NewTransferOrder{
				ProductID:      uuid.NewString(),
				FromLocationID: sd.InventoryLocations[0].LocationID,
				ToLocationID:   sd.InventoryLocations[2].LocationID,
				RequestedByID:  sd.TransferOrders[0].RequestedByID,
				ApprovedByID:   sd.TransferOrders[1].ApprovedByID,
				Quantity:       "10",
				Status:         "pending",
				TransferDate:   now.Format(timeutil.FORMAT),
			},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext foreign key violation"),
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "from-location-id-not-valid-fk",
			URL:        "/v1/inventory/transfer-orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &transferorderapp.NewTransferOrder{
				ProductID:      sd.Products[0].ProductID,
				FromLocationID: uuid.NewString(),
				ToLocationID:   sd.InventoryLocations[2].LocationID,
				RequestedByID:  sd.TransferOrders[0].RequestedByID,
				ApprovedByID:   sd.TransferOrders[1].ApprovedByID,
				Quantity:       "10",
				Status:         "pending",
				TransferDate:   now.Format(timeutil.FORMAT),
			},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext foreign key violation"),
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "to-location-id-not-valid-fk",
			URL:        "/v1/inventory/transfer-orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &transferorderapp.NewTransferOrder{
				ProductID:      sd.Products[0].ProductID,
				FromLocationID: sd.InventoryLocations[0].LocationID,
				ToLocationID:   uuid.NewString(),
				RequestedByID:  sd.TransferOrders[0].RequestedByID,
				ApprovedByID:   sd.TransferOrders[1].ApprovedByID,
				Quantity:       "10",
				Status:         "pending",
				TransferDate:   now.Format(timeutil.FORMAT),
			},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext foreign key violation"),
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "requested-by-id-not-valid-fk",
			URL:        "/v1/inventory/transfer-orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &transferorderapp.NewTransferOrder{
				ProductID:      sd.Products[0].ProductID,
				FromLocationID: sd.InventoryLocations[0].LocationID,
				ToLocationID:   sd.InventoryLocations[2].LocationID,
				RequestedByID:  uuid.NewString(),
				ApprovedByID:   sd.TransferOrders[1].ApprovedByID,
				Quantity:       "10",
				Status:         "pending",
				TransferDate:   now.Format(timeutil.FORMAT),
			},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext foreign key violation"),
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "approved-by-not-valid-fk",
			URL:        "/v1/inventory/transfer-orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &transferorderapp.NewTransferOrder{
				ProductID:      sd.Products[0].ProductID,
				FromLocationID: sd.InventoryLocations[0].LocationID,
				ToLocationID:   sd.InventoryLocations[2].LocationID,
				RequestedByID:  sd.TransferOrders[0].RequestedByID,
				ApprovedByID:   uuid.NewString(),
				Quantity:       "10",
				Status:         "pending",
				TransferDate:   now.Format(timeutil.FORMAT),
			},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext foreign key violation"),
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
			URL:        "/v1/inventory/transfer-orders",
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
			URL:        "/v1/inventory/transfer-orders",
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
			URL:        "/v1/inventory/transfer-orders",
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
			URL:        "/v1/inventory/transfer-orders",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: transfer_orders"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
