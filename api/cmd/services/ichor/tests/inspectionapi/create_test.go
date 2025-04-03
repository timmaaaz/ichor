package inspectionapi_test

import (
	"net/http"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/foundation/timeutil"

	"github.com/timmaaaz/ichor/app/domain/quality/inspectionapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {

	now := time.Now()
	later := now.Add(time.Hour * 24)

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/quality/inspections",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &inspectionapp.NewInspection{
				ProductID:          sd.Products[0].ProductID,
				InspectorID:        sd.Users[0].ID.String(),
				LotID:              sd.LotTracking[0].LotID,
				Status:             "pending",
				Notes:              "Sample Notes",
				InspectionDate:     now.Format(timeutil.FORMAT),
				NextInspectionDate: later.Format(timeutil.FORMAT),
			},
			GotResp: &inspectionapp.Inspection{},
			ExpResp: &inspectionapp.Inspection{
				ProductID:          sd.Products[0].ProductID,
				InspectorID:        sd.Users[0].ID.String(),
				LotID:              sd.LotTracking[0].LotID,
				Status:             "pending",
				Notes:              "Sample Notes",
				InspectionDate:     now.Format(timeutil.FORMAT),
				NextInspectionDate: later.Format(timeutil.FORMAT),
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*inspectionapp.Inspection)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*inspectionapp.Inspection)
				expResp.InspectionID = gotResp.InspectionID
				expResp.UpdatedDate = gotResp.UpdatedDate
				expResp.CreatedDate = gotResp.CreatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create400(sd apitest.SeedData) []apitest.Table {

	now := time.Now()
	later := now.Add(time.Hour * 24)

	return []apitest.Table{
		{
			Name:       "missing-product-id",
			URL:        "/v1/quality/inspections",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inspectionapp.NewInspection{
				InspectorID:        sd.Users[0].ID.String(),
				LotID:              sd.LotTracking[0].LotID,
				Status:             "pending",
				Notes:              "Sample Notes",
				InspectionDate:     now.Format(timeutil.FORMAT),
				NextInspectionDate: later.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"product_id\",\"error\":\"product_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-inspector-id",
			URL:        "/v1/quality/inspections",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inspectionapp.NewInspection{
				ProductID:          sd.Products[0].ProductID,
				LotID:              sd.LotTracking[0].LotID,
				Status:             "pending",
				Notes:              "Sample Notes",
				InspectionDate:     now.Format(timeutil.FORMAT),
				NextInspectionDate: later.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"inspector_id\",\"error\":\"inspector_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-lot-id",
			URL:        "/v1/quality/inspections",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inspectionapp.NewInspection{
				ProductID:          sd.Products[0].ProductID,
				InspectorID:        sd.Users[0].ID.String(),
				Status:             "pending",
				Notes:              "Sample Notes",
				InspectionDate:     now.Format(timeutil.FORMAT),
				NextInspectionDate: later.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"lot_id\",\"error\":\"lot_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-status",
			URL:        "/v1/quality/inspections",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inspectionapp.NewInspection{
				ProductID:          sd.Products[0].ProductID,
				InspectorID:        sd.Users[0].ID.String(),
				LotID:              sd.LotTracking[0].LotID,
				Notes:              "Sample Notes",
				InspectionDate:     now.Format(timeutil.FORMAT),
				NextInspectionDate: later.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"status\",\"error\":\"status is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-notes",
			URL:        "/v1/quality/inspections",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inspectionapp.NewInspection{
				ProductID:          sd.Products[0].ProductID,
				InspectorID:        sd.Users[0].ID.String(),
				LotID:              sd.LotTracking[0].LotID,
				Status:             "pending",
				InspectionDate:     now.Format(timeutil.FORMAT),
				NextInspectionDate: later.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"notes\",\"error\":\"notes is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-inspection-date",
			URL:        "/v1/quality/inspections",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inspectionapp.NewInspection{
				ProductID:          sd.Products[0].ProductID,
				InspectorID:        sd.Users[0].ID.String(),
				LotID:              sd.LotTracking[0].LotID,
				Status:             "pending",
				Notes:              "Sample Notes",
				NextInspectionDate: later.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"inspection_date\",\"error\":\"inspection_date is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-next-inspection-date",
			URL:        "/v1/quality/inspections",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inspectionapp.NewInspection{
				ProductID:      sd.Products[0].ProductID,
				InspectorID:    sd.Users[0].ID.String(),
				LotID:          sd.LotTracking[0].LotID,
				Status:         "pending",
				Notes:          "Sample Notes",
				InspectionDate: later.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"next_inspection_date\",\"error\":\"next_inspection_date is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},

		{
			Name:       "malformed-product-id",
			URL:        "/v1/quality/inspections",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inspectionapp.NewInspection{
				ProductID:          "not-a-uuid",
				InspectorID:        sd.Users[0].ID.String(),
				LotID:              sd.LotTracking[0].LotID,
				Status:             "pending",
				Notes:              "Sample Notes",
				InspectionDate:     now.Format(timeutil.FORMAT),
				NextInspectionDate: later.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"product_id\",\"error\":\"product_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-inspector-id",
			URL:        "/v1/quality/inspections",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inspectionapp.NewInspection{
				ProductID:          sd.Products[0].ProductID,
				InspectorID:        "not-a-uuid",
				LotID:              sd.LotTracking[0].LotID,
				Status:             "pending",
				Notes:              "Sample Notes",
				InspectionDate:     now.Format(timeutil.FORMAT),
				NextInspectionDate: later.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"inspector_id\",\"error\":\"inspector_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-lot-id",
			URL:        "/v1/quality/inspections",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inspectionapp.NewInspection{
				ProductID:          sd.Products[0].ProductID,
				InspectorID:        sd.Users[0].ID.String(),
				LotID:              "not-a-uuid",
				Status:             "pending",
				Notes:              "Sample Notes",
				InspectionDate:     now.Format(timeutil.FORMAT),
				NextInspectionDate: later.Format(timeutil.FORMAT),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"lot_id\",\"error\":\"lot_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create409(sd apitest.SeedData) []apitest.Table {
	now := time.Now()
	later := now.Add(time.Hour * 24)
	return []apitest.Table{
		{
			Name:       "product-id-not-valid-fk",
			URL:        "/v1/quality/inspections",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &inspectionapp.NewInspection{
				ProductID:          uuid.NewString(),
				InspectorID:        sd.Users[0].ID.String(),
				LotID:              sd.LotTracking[0].LotID,
				Status:             "pending",
				Notes:              "Sample Notes",
				InspectionDate:     now.Format(timeutil.FORMAT),
				NextInspectionDate: later.Format(timeutil.FORMAT),
			},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "inspector-id-not-valid-fk",
			URL:        "/v1/quality/inspections",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &inspectionapp.NewInspection{
				ProductID:          sd.Products[0].ProductID,
				InspectorID:        uuid.NewString(),
				LotID:              sd.LotTracking[0].LotID,
				Status:             "pending",
				Notes:              "Sample Notes",
				InspectionDate:     now.Format(timeutil.FORMAT),
				NextInspectionDate: later.Format(timeutil.FORMAT),
			},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "lot-id-not-valid-fk",
			URL:        "/v1/quality/inspections",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &inspectionapp.NewInspection{
				ProductID:          sd.Products[0].ProductID,
				InspectorID:        sd.Users[0].ID.String(),
				LotID:              uuid.NewString(),
				Status:             "pending",
				Notes:              "Sample Notes",
				InspectionDate:     now.Format(timeutil.FORMAT),
				NextInspectionDate: later.Format(timeutil.FORMAT),
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
			URL:        "/v1/quality/inspections",
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
			URL:        "/v1/quality/inspections",
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
			URL:        "/v1/quality/inspections",
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
			URL:        "/v1/quality/inspections",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: quality_inspections"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
