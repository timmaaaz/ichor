package lottrackingsapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/foundation/timeutil"

	"github.com/timmaaaz/ichor/app/domain/lots/lottrackingsapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/lot/lottrackingsbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

var (
	md = lottrackingsbus.RandomDate()
	ed = lottrackingsbus.RandomDate()
	rd = lottrackingsbus.RandomDate()
)

func update200(sd apitest.SeedData) []apitest.Table {

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/lots/lot-trackings/%s", sd.LotTrackings[0].LotID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &lottrackingsapp.UpdateLotTrackings{
				SupplierProductID: &sd.SupplierProducts[0].SupplierProductID,
				LotNumber:         dbtest.StringPointer("UpdateLotNumber"),
				ManufactureDate:   dbtest.StringPointer(md.Format(timeutil.FORMAT)),
				ExpirationDate:    dbtest.StringPointer(ed.Format(timeutil.FORMAT)),
				RecievedDate:      dbtest.StringPointer(rd.Format(timeutil.FORMAT)),
				Quantity:          dbtest.StringPointer("15"),
				QualityStatus:     dbtest.StringPointer("perfect"),
			},
			GotResp: &lottrackingsapp.LotTrackings{},
			ExpResp: &lottrackingsapp.LotTrackings{
				SupplierProductID: sd.SupplierProducts[0].SupplierProductID,
				LotNumber:         "UpdateLotNumber",
				ManufactureDate:   md.Format(timeutil.FORMAT),
				ExpirationDate:    ed.Format(timeutil.FORMAT),
				RecievedDate:      rd.Format(timeutil.FORMAT),
				Quantity:          "15",
				QualityStatus:     "perfect",
				LotID:             sd.LotTrackings[0].LotID,
				CreatedDate:       sd.LotTrackings[0].CreatedDate,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*lottrackingsapp.LotTrackings)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*lottrackingsapp.LotTrackings)
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "malformed-supplier-product-id",
			URL:        fmt.Sprintf("/v1/lots/lot-trackings/%s", sd.LotTrackings[0].LotID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &lottrackingsapp.UpdateLotTrackings{
				SupplierProductID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"supplier_product_id","error":"supplier_product_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-lot-uuid",
			URL:        fmt.Sprintf("/v1/lots/lot-trackings/%s", "not-a-uuid"),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &lottrackingsapp.UpdateLotTrackings{
				Quantity:      dbtest.StringPointer("15"),
				QualityStatus: dbtest.StringPointer("perfect"),
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
			URL:        fmt.Sprintf("/v1/lots/lot-trackings/%s", sd.LotTrackings[0].LotID),
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
			URL:        fmt.Sprintf("/v1/lots/lot-trackings/%s", sd.LotTrackings[0].LotID),
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
			URL:        fmt.Sprintf("/v1/lots/lot-trackings/%s", sd.LotTrackings[0].LotID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: lot_trackings"),
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
			URL:        fmt.Sprintf("/v1/lots/lot-trackings/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &lottrackingsapp.UpdateLotTrackings{
				Quantity:      dbtest.StringPointer("15"),
				QualityStatus: dbtest.StringPointer("perfect"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "lot not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "contact-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/lots/lot-trackings/%s", sd.LotTrackings[0].LotID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &lottrackingsapp.UpdateLotTrackings{
				SupplierProductID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "update: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
