package inventorytransactionapi_test

import (
	"net/http"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/foundation/timeutil"

	"github.com/timmaaaz/ichor/app/domain/inventory/inventorytransactionapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {

	now := time.Now()

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/inventory-transactions",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &inventorytransactionapp.NewInventoryTransaction{
				ProductID:       sd.Products[0].ProductID,
				LocationID:      sd.InventoryLocations[0].LocationID,
				UserID:          sd.Users[0].ID.String(),
				Quantity:        "10",
				TransactionType: "IN",
				ReferenceNumber: "ABC123",
				TransactionDate: now.Format(timeutil.FORMAT),
			},
			GotResp: &inventorytransactionapp.InventoryTransaction{},
			ExpResp: &inventorytransactionapp.InventoryTransaction{
				ProductID:       sd.Products[0].ProductID,
				LocationID:      sd.InventoryLocations[0].LocationID,
				UserID:          sd.Users[0].ID.String(),
				Quantity:        "10",
				TransactionType: "IN",
				ReferenceNumber: "ABC123",
				TransactionDate: now.Format(timeutil.FORMAT),
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*inventorytransactionapp.InventoryTransaction)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*inventorytransactionapp.InventoryTransaction)
				expResp.InventoryTransactionID = gotResp.InventoryTransactionID
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
			URL:        "/v1/inventory/inventory-transactions",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventorytransactionapp.NewInventoryTransaction{
				LocationID:      sd.InventoryLocations[0].LocationID,
				UserID:          sd.Users[0].ID.String(),
				Quantity:        "10",
				TransactionType: "IN",
				ReferenceNumber: "ABC123",
				TransactionDate: now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"product_id\",\"error\":\"product_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-location-id",
			URL:        "/v1/inventory/inventory-transactions",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventorytransactionapp.NewInventoryTransaction{
				ProductID:       sd.Products[0].ProductID,
				UserID:          sd.Users[0].ID.String(),
				Quantity:        "10",
				TransactionType: "IN",
				ReferenceNumber: "ABC123",
				TransactionDate: now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"location_id\",\"error\":\"location_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-user-id",
			URL:        "/v1/inventory/inventory-transactions",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventorytransactionapp.NewInventoryTransaction{
				ProductID:       sd.Products[0].ProductID,
				LocationID:      sd.InventoryLocations[0].LocationID,
				Quantity:        "10",
				TransactionType: "IN",
				ReferenceNumber: "ABC123",
				TransactionDate: now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"user_id\",\"error\":\"user_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-quantity",
			URL:        "/v1/inventory/inventory-transactions",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventorytransactionapp.NewInventoryTransaction{
				ProductID:       sd.Products[0].ProductID,
				LocationID:      sd.InventoryLocations[0].LocationID,
				UserID:          sd.Users[0].ID.String(),
				TransactionType: "IN",
				ReferenceNumber: "ABC123",
				TransactionDate: now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"quantity\",\"error\":\"quantity is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-transaction-type",
			URL:        "/v1/inventory/inventory-transactions",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventorytransactionapp.NewInventoryTransaction{
				ProductID:       sd.Products[0].ProductID,
				LocationID:      sd.InventoryLocations[0].LocationID,
				UserID:          sd.Users[0].ID.String(),
				Quantity:        "10",
				ReferenceNumber: "ABC123",
				TransactionDate: now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"transaction_type\",\"error\":\"transaction_type is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-reference-number",
			URL:        "/v1/inventory/inventory-transactions",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventorytransactionapp.NewInventoryTransaction{
				ProductID:       sd.Products[0].ProductID,
				LocationID:      sd.InventoryLocations[0].LocationID,
				UserID:          sd.Users[0].ID.String(),
				Quantity:        "10",
				TransactionType: "IN",
				TransactionDate: now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"reference_number\",\"error\":\"reference_number is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-transaction-date",
			URL:        "/v1/inventory/inventory-transactions",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventorytransactionapp.NewInventoryTransaction{
				ProductID:       sd.Products[0].ProductID,
				LocationID:      sd.InventoryLocations[0].LocationID,
				UserID:          sd.Users[0].ID.String(),
				Quantity:        "10",
				TransactionType: "IN",
				ReferenceNumber: "ABC123",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"transaction_date\",\"error\":\"transaction_date is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},

		{
			Name:       "malformed-product-id",
			URL:        "/v1/inventory/inventory-transactions",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventorytransactionapp.NewInventoryTransaction{
				ProductID:       "not-a-uuid",
				LocationID:      sd.InventoryLocations[0].LocationID,
				UserID:          sd.Users[0].ID.String(),
				Quantity:        "10",
				TransactionType: "IN",
				ReferenceNumber: "ABC123",
				TransactionDate: now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"product_id\",\"error\":\"product_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-location-id",
			URL:        "/v1/inventory/inventory-transactions",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventorytransactionapp.NewInventoryTransaction{
				ProductID:       sd.Products[0].ProductID,
				LocationID:      "not-a-uuid",
				UserID:          sd.Users[0].ID.String(),
				Quantity:        "10",
				TransactionType: "IN",
				ReferenceNumber: "ABC123",
				TransactionDate: now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"location_id\",\"error\":\"location_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-user-id",
			URL:        "/v1/inventory/inventory-transactions",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventorytransactionapp.NewInventoryTransaction{
				ProductID:       sd.Products[0].ProductID,
				LocationID:      sd.InventoryLocations[0].LocationID,
				UserID:          "not-a-uuid",
				Quantity:        "10",
				TransactionType: "IN",
				ReferenceNumber: "ABC123",
				TransactionDate: now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"user_id\",\"error\":\"user_id must be at least 36 characters in length\"}]"),
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
			URL:        "/v1/inventory/inventory-transactions",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &inventorytransactionapp.NewInventoryTransaction{
				ProductID:       uuid.NewString(),
				LocationID:      sd.InventoryLocations[0].LocationID,
				UserID:          sd.Users[0].ID.String(),
				Quantity:        "10",
				TransactionType: "IN",
				ReferenceNumber: "ABC123",
				TransactionDate: now.Format(timeutil.FORMAT),
			},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "location-id-not-valid-fk",
			URL:        "/v1/inventory/inventory-transactions",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &inventorytransactionapp.NewInventoryTransaction{
				ProductID:       sd.Products[0].ProductID,
				LocationID:      uuid.NewString(),
				UserID:          sd.Users[0].ID.String(),
				Quantity:        "10",
				TransactionType: "IN",
				ReferenceNumber: "ABC123",
				TransactionDate: now.Format(timeutil.FORMAT),
			},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "user-id-not-valid-fk",
			URL:        "/v1/inventory/inventory-transactions",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &inventorytransactionapp.NewInventoryTransaction{
				ProductID:       sd.Products[0].ProductID,
				LocationID:      sd.InventoryLocations[0].LocationID,
				UserID:          uuid.NewString(),
				Quantity:        "10",
				TransactionType: "IN",
				ReferenceNumber: "ABC123",
				TransactionDate: now.Format(timeutil.FORMAT),
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
			URL:        "/v1/inventory/inventory-transactions",
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
			URL:        "/v1/inventory/inventory-transactions",
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
			URL:        "/v1/inventory/inventory-transactions",
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
			URL:        "/v1/inventory/inventory-transactions",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: inventory_transactions"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
