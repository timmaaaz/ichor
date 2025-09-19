package physicalattribute_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/products/physicalattributeapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/inventory/core/physical-attributes/%s", sd.PhysicalAttributes[1].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &physicalattributeapp.UpdatePhysicalAttribute{
				ProductID:           &sd.Products[3].ProductID,
				Length:              dbtest.StringPointer("32.2"),
				Width:               dbtest.StringPointer("20.145"),
				Height:              dbtest.StringPointer("13.98"),
				Weight:              dbtest.StringPointer("10.5"),
				WeightUnit:          dbtest.StringPointer("lbs"),
				Color:               dbtest.StringPointer("red"),
				Size:                dbtest.StringPointer("m"),
				Material:            dbtest.StringPointer("bronze"),
				StorageRequirements: dbtest.StringPointer("cold"),
				HazmatClass:         dbtest.StringPointer("high"),
				ShelfLifeDays:       dbtest.StringPointer("6"),
			},
			GotResp: &physicalattributeapp.PhysicalAttribute{},
			ExpResp: &physicalattributeapp.PhysicalAttribute{
				ID:                  sd.PhysicalAttributes[1].ID,
				ProductID:           sd.Products[3].ProductID,
				Length:              "32.2",
				Width:               "20.145",
				Height:              "13.98",
				Weight:              "10.5",
				WeightUnit:          "lbs",
				Color:               "red",
				Size:                "m",
				Material:            "bronze",
				StorageRequirements: "cold",
				HazmatClass:         "high",
				ShelfLifeDays:       "6",
				CreatedDate:         sd.PhysicalAttributes[1].CreatedDate,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*physicalattributeapp.PhysicalAttribute)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*physicalattributeapp.PhysicalAttribute)
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "malformed-product-uuid",
			URL:        fmt.Sprintf("/v1/inventory/core/physical-attributes/%s", sd.PhysicalAttributes[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &physicalattributeapp.UpdatePhysicalAttribute{
				ProductID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"product_id","error":"product_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-brand-uuid",
			URL:        fmt.Sprintf("/v1/inventory/core/physical-attributes/%s", "not-a-uuid"),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &physicalattributeapp.UpdatePhysicalAttribute{
				ProductID: &sd.Products[0].ProductID,
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
			URL:        fmt.Sprintf("/v1/inventory/core/physical-attributes/%s", sd.PhysicalAttributes[0].ID),
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
			URL:        fmt.Sprintf("/v1/inventory/core/physical-attributes/%s", sd.PhysicalAttributes[0].ID),
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
			URL:        fmt.Sprintf("/v1/inventory/core/physical-attributes/%s", sd.PhysicalAttributes[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: physical_attributes"),
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
			Name:       "attribute-dne",
			URL:        fmt.Sprintf("/v1/inventory/core/physical-attributes/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &physicalattributeapp.UpdatePhysicalAttribute{
				ProductID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "physical attribute not found"),
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
			URL:        fmt.Sprintf("/v1/inventory/core/physical-attributes/%s", sd.PhysicalAttributes[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &physicalattributeapp.UpdatePhysicalAttribute{
				ProductID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
