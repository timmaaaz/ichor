package physicalattribute_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/core/physicalattributeapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/core/physicalattribute",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &physicalattributeapp.NewPhysicalAttribute{
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
			},
			GotResp: &physicalattributeapp.PhysicalAttribute{},
			ExpResp: &physicalattributeapp.PhysicalAttribute{
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
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*physicalattributeapp.PhysicalAttribute)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*physicalattributeapp.PhysicalAttribute)
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
			URL:        "/v1/inventory/core/physicalattribute",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &physicalattributeapp.NewPhysicalAttribute{
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
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"product_id\",\"error\":\"product_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-length",
			URL:        "/v1/inventory/core/physicalattribute",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &physicalattributeapp.NewPhysicalAttribute{
				ProductID:           sd.Products[3].ProductID,
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
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"length\",\"error\":\"length is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-width",
			URL:        "/v1/inventory/core/physicalattribute",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &physicalattributeapp.NewPhysicalAttribute{
				ProductID:           sd.Products[3].ProductID,
				Length:              "32.2",
				Height:              "13.98",
				Weight:              "10.5",
				WeightUnit:          "lbs",
				Color:               "red",
				Size:                "m",
				Material:            "bronze",
				StorageRequirements: "cold",
				HazmatClass:         "high",
				ShelfLifeDays:       "6",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"width\",\"error\":\"width is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-height",
			URL:        "/v1/inventory/core/physicalattribute",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &physicalattributeapp.NewPhysicalAttribute{
				ProductID:           sd.Products[3].ProductID,
				Length:              "32.2",
				Width:               "20.145",
				Weight:              "10.5",
				WeightUnit:          "lbs",
				Color:               "red",
				Size:                "m",
				Material:            "bronze",
				StorageRequirements: "cold",
				HazmatClass:         "high",
				ShelfLifeDays:       "6",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"height\",\"error\":\"height is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-weight",
			URL:        "/v1/inventory/core/physicalattribute",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &physicalattributeapp.NewPhysicalAttribute{
				ProductID:           sd.Products[3].ProductID,
				Length:              "32.2",
				Width:               "20.145",
				Height:              "13.98",
				WeightUnit:          "lbs",
				Color:               "red",
				Size:                "m",
				Material:            "bronze",
				StorageRequirements: "cold",
				HazmatClass:         "high",
				ShelfLifeDays:       "6",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"weight\",\"error\":\"weight is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-weight-unit",
			URL:        "/v1/inventory/core/physicalattribute",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &physicalattributeapp.NewPhysicalAttribute{
				ProductID:           sd.Products[3].ProductID,
				Length:              "32.2",
				Width:               "20.145",
				Height:              "13.98",
				Weight:              "10.5",
				Color:               "red",
				Size:                "m",
				Material:            "bronze",
				StorageRequirements: "cold",
				HazmatClass:         "high",
				ShelfLifeDays:       "6",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"weight_unit\",\"error\":\"weight_unit is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-storage-requirement",
			URL:        "/v1/inventory/core/physicalattribute",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &physicalattributeapp.NewPhysicalAttribute{
				ProductID:     sd.Products[3].ProductID,
				Length:        "32.2",
				Width:         "20.145",
				Height:        "13.98",
				Weight:        "10.5",
				WeightUnit:    "lbs",
				Color:         "red",
				Size:          "m",
				Material:      "bronze",
				HazmatClass:   "high",
				ShelfLifeDays: "6",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"storage_requirements\",\"error\":\"storage_requirements is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-hazmat-class",
			URL:        "/v1/inventory/core/physicalattribute",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &physicalattributeapp.NewPhysicalAttribute{
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
				ShelfLifeDays:       "6",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"hazmat_class\",\"error\":\"hazmat_class is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-shelf-life-days",
			URL:        "/v1/inventory/core/physicalattribute",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &physicalattributeapp.NewPhysicalAttribute{
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
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"shelf_life_days\",\"error\":\"shelf_life_days is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-product-id",
			URL:        "/v1/inventory/core/physicalattribute",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &physicalattributeapp.NewPhysicalAttribute{
				ProductID:           "not-a-uuid",
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
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"product_id","error":"product_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "product-id-not-valid-fk",
			URL:        "/v1/inventory/core/physicalattribute",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &physicalattributeapp.NewPhysicalAttribute{
				ProductID:           uuid.NewString(),
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
			},
			ExpResp: errs.Newf(errs.Aborted, "foreign key violation"),
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
			URL:        "/v1/inventory/core/physicalattribute",
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
			URL:        "/v1/inventory/core/physicalattribute",
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
			URL:        "/v1/inventory/core/physicalattribute",
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
			URL:        "/v1/inventory/core/physicalattribute",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: physical_attributes"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
