package inspectionapi_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/foundation/timeutil"

	"github.com/timmaaaz/ichor/app/domain/inventory/inspectionapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {

	now := time.Now()
	later := now.Add(time.Hour * 24)

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/inventory/quality-inspections/%s", sd.Inspections[0].InspectionID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &inspectionapp.UpdateInspection{
				ProductID:          &sd.Products[3].ProductID,
				LotID:              &sd.LotTrackings[0].LotID,
				InspectorID:        dbtest.StringPointer(sd.Users[1].ID.String()),
				Status:             dbtest.StringPointer("passed"),
				Notes:              dbtest.StringPointer("Sample Notes"),
				InspectionDate:     dbtest.StringPointer(now.Format(timeutil.FORMAT)),
				NextInspectionDate: dbtest.StringPointer(later.Format(timeutil.FORMAT)),
			},
			GotResp: &inspectionapp.Inspection{},
			ExpResp: &inspectionapp.Inspection{
				ProductID:          sd.Products[3].ProductID,
				LotID:              sd.LotTrackings[0].LotID,
				InspectorID:        sd.Users[1].ID.String(),
				Status:             "passed",
				Notes:              "Sample Notes",
				InspectionDate:     now.Format(timeutil.FORMAT),
				NextInspectionDate: later.Format(timeutil.FORMAT),
				CreatedDate:        sd.Inspections[0].CreatedDate,
				InspectionID:       sd.Inspections[0].InspectionID,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*inspectionapp.Inspection)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*inspectionapp.Inspection)
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "malformed-product-id",
			URL:        fmt.Sprintf("/v1/inventory/quality-inspections/%s", sd.Inspections[0].InspectionID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &inspectionapp.UpdateInspection{
				ProductID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"product_id","error":"product_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},

		{
			Name:       "malformed-lot-id",
			URL:        fmt.Sprintf("/v1/inventory/quality-inspections/%s", sd.Inspections[0].InspectionID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &inspectionapp.UpdateInspection{
				LotID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"lot_id","error":"lot_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},

		{
			Name:       "malformed-inspector-id",
			URL:        fmt.Sprintf("/v1/inventory/quality-inspections/%s", sd.Inspections[0].InspectionID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &inspectionapp.UpdateInspection{
				InspectorID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"inspector_id","error":"inspector_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},

		{
			Name:       "malformed-inspection-id",
			URL:        fmt.Sprintf("/v1/inventory/quality-inspections/%s", "not-a-uuid"),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &inspectionapp.UpdateInspection{
				ProductID: &sd.Inspections[0].ProductID,
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
			URL:        fmt.Sprintf("/v1/inventory/quality-inspections/%s", sd.Inspections[0].InspectionID),
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
			URL:        fmt.Sprintf("/v1/inventory/quality-inspections/%s", sd.Inspections[0].InspectionID),
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
			URL:        fmt.Sprintf("/v1/inventory/quality-inspections/%s", sd.Inspections[0].InspectionID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: inventory.quality_inspections"),
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
			URL:        fmt.Sprintf("/v1/inventory/quality-inspections/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &inspectionapp.UpdateInspection{
				ProductID: &sd.Products[0].ProductID,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "queryByID: inspection not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "product-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/inventory/quality-inspections/%s", sd.Inspections[0].InspectionID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &inspectionapp.UpdateInspection{
				ProductID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "update: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "lot-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/inventory/quality-inspections/%s", sd.Inspections[0].InspectionID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &inspectionapp.UpdateInspection{
				LotID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "update: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "inspector-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/inventory/quality-inspections/%s", sd.Inspections[0].InspectionID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &inspectionapp.UpdateInspection{
				InspectorID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "update: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
