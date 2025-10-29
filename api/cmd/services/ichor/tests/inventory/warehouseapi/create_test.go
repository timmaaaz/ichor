package warehouse_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/warehouseapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/warehouses",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &warehouseapp.NewWarehouse{
				Name:      "New Warehouse test",
				StreetID:  sd.Streets[0].ID,
				CreatedBy: sd.Admins[0].ID.String(),
			},
			GotResp: &warehouseapp.Warehouse{},
			ExpResp: &warehouseapp.Warehouse{
				Name:      "New Warehouse test",
				StreetID:  sd.Streets[0].ID,
				CreatedBy: sd.Admins[0].ID.String(),
				UpdatedBy: sd.Admins[0].ID.String(),
				IsActive:  true,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*warehouseapp.Warehouse)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*warehouseapp.Warehouse)
				expResp.ID = gotResp.ID
				expResp.Code = gotResp.Code
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "auto-generated code",
			URL:        "/v1/inventory/warehouses",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &warehouseapp.NewWarehouse{
				Name:      "Auto Code Warehouse",
				StreetID:  sd.Streets[0].ID,
				CreatedBy: sd.Admins[0].ID.String(),
			},
			GotResp: &warehouseapp.Warehouse{},
			ExpResp: &warehouseapp.Warehouse{
				Name:      "Auto Code Warehouse",
				StreetID:  sd.Streets[0].ID,
				CreatedBy: sd.Admins[0].ID.String(),
				UpdatedBy: sd.Admins[0].ID.String(),
				IsActive:  true,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*warehouseapp.Warehouse)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*warehouseapp.Warehouse)
				expResp.ID = gotResp.ID
				expResp.Code = gotResp.Code
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				// Verify code was auto-generated in WH-XXXXX format
				if len(gotResp.Code) != 8 || gotResp.Code[:3] != "WH-" {
					return "code was not auto-generated in expected format WH-XXXXX"
				}

				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "manual code preserved",
			URL:        "/v1/inventory/warehouses",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &warehouseapp.NewWarehouse{
				Code:      "CUSTOM-WH-123",
				Name:      "Manual Code Warehouse",
				StreetID:  sd.Streets[0].ID,
				CreatedBy: sd.Admins[0].ID.String(),
			},
			GotResp: &warehouseapp.Warehouse{},
			ExpResp: &warehouseapp.Warehouse{
				Code:      "CUSTOM-WH-123",
				Name:      "Manual Code Warehouse",
				StreetID:  sd.Streets[0].ID,
				CreatedBy: sd.Admins[0].ID.String(),
				UpdatedBy: sd.Admins[0].ID.String(),
				IsActive:  true,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*warehouseapp.Warehouse)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*warehouseapp.Warehouse)
				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(got, exp)
			},
		},
	}
}

func create400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing name",
			URL:        "/v1/inventory/warehouses",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &warehouseapp.NewWarehouse{
				StreetID:  sd.Streets[0].ID,
				CreatedBy: sd.Admins[0].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"name","error":"name is a required field"}]`),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing street_id",
			URL:        "/v1/inventory/warehouses",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &warehouseapp.NewWarehouse{
				Name:      "New Warehouse",
				CreatedBy: sd.Admins[0].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"street_id","error":"street_id is a required field"}]`),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty token",
			URL:        "/v1/inventory/warehouses",
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
			URL:        "/v1/inventory/warehouses",
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
			URL:        "/v1/inventory/warehouses",
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
			URL:        "/v1/inventory/warehouses",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: warehouses"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
