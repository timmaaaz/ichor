package inventoryadjustmentapi_test

import (
	"net/http"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/foundation/timeutil"

	"github.com/timmaaaz/ichor/app/domain/movement/inventoryadjustmentapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {

	now := time.Now()

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/movement/inventoryadjustment",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &inventoryadjustmentapp.NewInventoryAdjustment{
				ProductID:      sd.Products[0].ProductID,
				LocationID:     sd.InventoryLocations[0].LocationID,
				AdjustedBy:     sd.InventoryAdjustments[0].AdjustedBy,
				ApprovedBy:     sd.InventoryAdjustments[0].ApprovedBy,
				QuantityChange: "10",
				ReasonCode:     "Purchase",
				Notes:          "New purchase",
				AdjustmentDate: now.Format(timeutil.FORMAT),
			},
			GotResp: &inventoryadjustmentapp.InventoryAdjustment{},
			ExpResp: &inventoryadjustmentapp.InventoryAdjustment{
				ProductID:      sd.Products[0].ProductID,
				LocationID:     sd.InventoryLocations[0].LocationID,
				AdjustedBy:     sd.InventoryAdjustments[0].AdjustedBy,
				ApprovedBy:     sd.InventoryAdjustments[0].ApprovedBy,
				QuantityChange: "10",
				ReasonCode:     "Purchase",
				Notes:          "New purchase",
				AdjustmentDate: now.Format(timeutil.FORMAT),
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*inventoryadjustmentapp.InventoryAdjustment)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*inventoryadjustmentapp.InventoryAdjustment)
				expResp.InventoryAdjustmentID = gotResp.InventoryAdjustmentID
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
			URL:        "/v1/movement/inventoryadjustment",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryadjustmentapp.NewInventoryAdjustment{
				LocationID:     sd.InventoryLocations[0].LocationID,
				AdjustedBy:     sd.InventoryAdjustments[0].AdjustedBy,
				ApprovedBy:     sd.InventoryAdjustments[0].ApprovedBy,
				QuantityChange: "10",
				ReasonCode:     "Purchase",
				Notes:          "New purchase",
				AdjustmentDate: now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"product_id\",\"error\":\"product_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-location-id",
			URL:        "/v1/movement/inventoryadjustment",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryadjustmentapp.NewInventoryAdjustment{
				ProductID:      sd.Products[0].ProductID,
				AdjustedBy:     sd.InventoryAdjustments[0].AdjustedBy,
				ApprovedBy:     sd.InventoryAdjustments[0].ApprovedBy,
				QuantityChange: "10",
				ReasonCode:     "Purchase",
				Notes:          "New purchase",
				AdjustmentDate: now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"location_id\",\"error\":\"location_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-adjusted-by",
			URL:        "/v1/movement/inventoryadjustment",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryadjustmentapp.NewInventoryAdjustment{
				ProductID:      sd.Products[0].ProductID,
				LocationID:     sd.InventoryLocations[0].LocationID,
				ApprovedBy:     sd.InventoryAdjustments[0].ApprovedBy,
				QuantityChange: "10",
				ReasonCode:     "Purchase",
				Notes:          "New purchase",
				AdjustmentDate: now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"adjusted_by\",\"error\":\"adjusted_by is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-approved-by",
			URL:        "/v1/movement/inventoryadjustment",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryadjustmentapp.NewInventoryAdjustment{
				ProductID:      sd.Products[0].ProductID,
				LocationID:     sd.InventoryLocations[0].LocationID,
				AdjustedBy:     sd.InventoryAdjustments[0].AdjustedBy,
				QuantityChange: "10",
				ReasonCode:     "Purchase",
				Notes:          "New purchase",
				AdjustmentDate: now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"approved_by\",\"error\":\"approved_by is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-quantity-change",
			URL:        "/v1/movement/inventoryadjustment",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryadjustmentapp.NewInventoryAdjustment{
				ProductID:      sd.Products[0].ProductID,
				LocationID:     sd.InventoryLocations[0].LocationID,
				AdjustedBy:     sd.InventoryAdjustments[0].AdjustedBy,
				ApprovedBy:     sd.InventoryAdjustments[0].ApprovedBy,
				ReasonCode:     "Purchase",
				Notes:          "New purchase",
				AdjustmentDate: now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"quantity_change\",\"error\":\"quantity_change is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-reason-code",
			URL:        "/v1/movement/inventoryadjustment",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryadjustmentapp.NewInventoryAdjustment{
				ProductID:      sd.Products[0].ProductID,
				LocationID:     sd.InventoryLocations[0].LocationID,
				AdjustedBy:     sd.InventoryAdjustments[0].AdjustedBy,
				ApprovedBy:     sd.InventoryAdjustments[0].ApprovedBy,
				QuantityChange: "10",
				Notes:          "New purchase",
				AdjustmentDate: now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"reason_code\",\"error\":\"reason_code is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-notes",
			URL:        "/v1/movement/inventoryadjustment",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryadjustmentapp.NewInventoryAdjustment{
				ProductID:      sd.Products[0].ProductID,
				LocationID:     sd.InventoryLocations[0].LocationID,
				AdjustedBy:     sd.InventoryAdjustments[0].AdjustedBy,
				ApprovedBy:     sd.InventoryAdjustments[0].ApprovedBy,
				QuantityChange: "10",
				ReasonCode:     "Purchase",
				AdjustmentDate: now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"notes\",\"error\":\"notes is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-adjustment-date",
			URL:        "/v1/movement/inventoryadjustment",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryadjustmentapp.NewInventoryAdjustment{
				ProductID:      sd.Products[0].ProductID,
				LocationID:     sd.InventoryLocations[0].LocationID,
				AdjustedBy:     sd.InventoryAdjustments[0].AdjustedBy,
				ApprovedBy:     sd.InventoryAdjustments[0].ApprovedBy,
				QuantityChange: "10",
				ReasonCode:     "Purchase",
				Notes:          "New purchase",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"adjustment_date\",\"error\":\"adjustment_date is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},

		{
			Name:       "malformed-product-id",
			URL:        "/v1/movement/inventoryadjustment",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryadjustmentapp.NewInventoryAdjustment{
				ProductID:      "not-a-uuid",
				LocationID:     sd.InventoryLocations[0].LocationID,
				AdjustedBy:     sd.InventoryAdjustments[0].AdjustedBy,
				ApprovedBy:     sd.InventoryAdjustments[0].ApprovedBy,
				QuantityChange: "10",
				ReasonCode:     "Purchase",
				Notes:          "New purchase",
				AdjustmentDate: now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"product_id\",\"error\":\"product_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-location-id",
			URL:        "/v1/movement/inventoryadjustment",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryadjustmentapp.NewInventoryAdjustment{
				ProductID:      sd.Products[0].ProductID,
				LocationID:     "not-a-uuid",
				AdjustedBy:     sd.InventoryAdjustments[0].AdjustedBy,
				ApprovedBy:     sd.InventoryAdjustments[0].ApprovedBy,
				QuantityChange: "10",
				ReasonCode:     "Purchase",
				Notes:          "New purchase",
				AdjustmentDate: now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"location_id\",\"error\":\"location_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-adjusted-by",
			URL:        "/v1/movement/inventoryadjustment",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryadjustmentapp.NewInventoryAdjustment{
				ProductID:      sd.Products[0].ProductID,
				LocationID:     sd.InventoryLocations[0].LocationID,
				AdjustedBy:     "not-a-uuid",
				ApprovedBy:     sd.InventoryAdjustments[0].ApprovedBy,
				QuantityChange: "10",
				ReasonCode:     "Purchase",
				Notes:          "New purchase",
				AdjustmentDate: now.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"adjusted_by\",\"error\":\"adjusted_by must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-approved-by",
			URL:        "/v1/movement/inventoryadjustment",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryadjustmentapp.NewInventoryAdjustment{
				ProductID:      sd.Products[0].ProductID,
				LocationID:     sd.InventoryLocations[0].LocationID,
				AdjustedBy:     sd.InventoryAdjustments[0].AdjustedBy,
				ApprovedBy:     "not-a-uuid",
				QuantityChange: "10",
				ReasonCode:     "Purchase",
				Notes:          "New purchase",
				AdjustmentDate: now.Format(timeutil.FORMAT),
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
			URL:        "/v1/movement/inventoryadjustment",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &inventoryadjustmentapp.NewInventoryAdjustment{
				ProductID:      uuid.New().String(),
				LocationID:     sd.InventoryLocations[0].LocationID,
				AdjustedBy:     sd.InventoryAdjustments[0].AdjustedBy,
				ApprovedBy:     sd.InventoryAdjustments[0].ApprovedBy,
				QuantityChange: "10",
				ReasonCode:     "Purchase",
				Notes:          "New purchase",
				AdjustmentDate: now.Format(timeutil.FORMAT),
			},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "location-id-not-valid-fk",
			URL:        "/v1/movement/inventoryadjustment",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &inventoryadjustmentapp.NewInventoryAdjustment{
				ProductID:      sd.Products[0].ProductID,
				LocationID:     uuid.NewString(),
				AdjustedBy:     sd.InventoryAdjustments[0].AdjustedBy,
				ApprovedBy:     sd.InventoryAdjustments[0].ApprovedBy,
				QuantityChange: "10",
				ReasonCode:     "Purchase",
				Notes:          "New purchase",
				AdjustmentDate: now.Format(timeutil.FORMAT),
			},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "approved-by-not-valid-fk",
			URL:        "/v1/movement/inventoryadjustment",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &inventoryadjustmentapp.NewInventoryAdjustment{
				ProductID:      sd.Products[0].ProductID,
				LocationID:     sd.InventoryLocations[0].LocationID,
				AdjustedBy:     sd.InventoryAdjustments[0].AdjustedBy,
				ApprovedBy:     uuid.NewString(),
				QuantityChange: "10",
				ReasonCode:     "Purchase",
				Notes:          "New purchase",
				AdjustmentDate: now.Format(timeutil.FORMAT),
			},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "adjusted-by-not-valid-fk",
			URL:        "/v1/movement/inventoryadjustment",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &inventoryadjustmentapp.NewInventoryAdjustment{
				ProductID:      sd.Products[0].ProductID,
				LocationID:     sd.InventoryLocations[0].LocationID,
				AdjustedBy:     uuid.NewString(),
				ApprovedBy:     sd.InventoryAdjustments[0].ApprovedBy,
				QuantityChange: "10",
				ReasonCode:     "Purchase",
				Notes:          "New purchase",
				AdjustmentDate: now.Format(timeutil.FORMAT),
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
			URL:        "/v1/movement/inventoryadjustment",
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
			URL:        "/v1/movement/inventoryadjustment",
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
			URL:        "/v1/movement/inventoryadjustment",
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
			URL:        "/v1/movement/inventoryadjustment",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: inventory_adjustments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
