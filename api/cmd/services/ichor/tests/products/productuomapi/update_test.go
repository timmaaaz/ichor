package productuomapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/products/productuomapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/products/productuoms/%s", sd.ProductUOMs[1].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &productuomapp.UpdateProductUOM{
				Name:             dbtest.StringPointer("Dozen"),
				Abbreviation:     dbtest.StringPointer("DZ"),
				ConversionFactor: dbtest.Float64Pointer(12.0),
				IsBase:           dbtest.BoolPointer(false),
				IsApproximate:    dbtest.BoolPointer(false),
				Notes:            dbtest.StringPointer("twelve units"),
			},
			GotResp: &productuomapp.ProductUOM{},
			ExpResp: &productuomapp.ProductUOM{
				ID:               sd.ProductUOMs[1].ID,
				ProductID:        sd.ProductUOMs[1].ProductID,
				Name:             "Dozen",
				Abbreviation:     "DZ",
				ConversionFactor: "12",
				IsBase:           false,
				IsApproximate:    false,
				Notes:            "twelve units",
				CreatedDate:      sd.ProductUOMs[1].CreatedDate,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*productuomapp.ProductUOM)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*productuomapp.ProductUOM)
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "malformed-uom-uuid",
			URL:        fmt.Sprintf("/v1/products/productuoms/%s", "not-a-uuid"),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input:      &productuomapp.UpdateProductUOM{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "invalid UUID length: 10"),
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
			URL:        fmt.Sprintf("/v1/products/productuoms/%s", sd.ProductUOMs[0].ID),
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
			URL:        fmt.Sprintf("/v1/products/productuoms/%s", sd.ProductUOMs[0].ID),
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
			URL:        fmt.Sprintf("/v1/products/productuoms/%s", sd.ProductUOMs[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusForbidden,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.PermissionDenied, "user does not have permission UPDATE for table: products.product_uoms"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "uom-dne",
			URL:        fmt.Sprintf("/v1/products/productuoms/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input:      &productuomapp.UpdateProductUOM{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.NotFound, "product uom not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
