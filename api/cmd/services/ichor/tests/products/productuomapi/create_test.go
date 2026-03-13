package productuomapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/products/productuomapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/products/productuoms",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &productuomapp.NewProductUOM{
				ProductID:        sd.Products[0].ProductID,
				Name:             "Each",
				Abbreviation:     "EA",
				ConversionFactor: 1.0,
				IsBase:           false,
				IsApproximate:    false,
				Notes:            "base unit",
			},
			GotResp: &productuomapp.ProductUOM{},
			ExpResp: &productuomapp.ProductUOM{
				ProductID:        sd.Products[0].ProductID,
				Name:             "Each",
				Abbreviation:     "EA",
				ConversionFactor: "1",
				IsBase:           false,
				IsApproximate:    false,
				Notes:            "base unit",
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*productuomapp.ProductUOM)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*productuomapp.ProductUOM)
				expResp.ID = gotResp.ID
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
			Name:       "missing-product-id",
			URL:        "/v1/products/productuoms",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productuomapp.NewProductUOM{
				Name:             "Each",
				ConversionFactor: 1.0,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"product_id","error":"product_id is a required field"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-name",
			URL:        "/v1/products/productuoms",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productuomapp.NewProductUOM{
				ProductID:        sd.Products[0].ProductID,
				ConversionFactor: 1.0,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"name","error":"name is a required field"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-product-id",
			URL:        "/v1/products/productuoms",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productuomapp.NewProductUOM{
				ProductID:        "not-a-uuid",
				Name:             "Each",
				ConversionFactor: 1.0,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"product_id","error":"product_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty token",
			URL:        "/v1/products/productuoms",
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
			URL:        "/v1/products/productuoms",
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
			URL:        "/v1/products/productuoms",
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
			URL:        "/v1/products/productuoms",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: products.product_uoms"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "product-id-not-valid-fk",
			URL:        "/v1/products/productuoms",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &productuomapp.NewProductUOM{
				ProductID:        uuid.NewString(),
				Name:             "Each",
				ConversionFactor: 1.0,
			},
			ExpResp: errs.Newf(errs.Aborted, "foreign key violation"),
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
