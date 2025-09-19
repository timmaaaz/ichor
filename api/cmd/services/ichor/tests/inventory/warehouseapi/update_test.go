package warehouse_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/warehouseapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/warehouses/" + sd.Warehouses[0].ID,
			Token:      sd.Admins[0].Token,
			Method:     "PUT",
			StatusCode: 200,
			Input: &warehouseapp.UpdateWarehouse{
				Name:      "Updated Warehouse",
				IsActive:  false,
				UpdatedBy: sd.Admins[0].ID.String(),
			},
			GotResp: &warehouseapp.Warehouse{},
			ExpResp: &warehouseapp.Warehouse{
				Name:      "Updated Warehouse",
				IsActive:  false,
				UpdatedBy: sd.Admins[0].ID.String(),
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*warehouseapp.Warehouse)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*warehouseapp.Warehouse)
				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate
				expResp.CreatedBy = gotResp.CreatedBy
				expResp.StreetID = gotResp.StreetID

				return cmp.Diff(got, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/warehouses/" + sd.Warehouses[0].ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &warehouseapp.UpdateWarehouse{
				Name: "Updated Warehouse",
			},
			GotResp: &warehouseapp.Warehouse{},
			ExpResp: &warehouseapp.Warehouse{},
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
			URL:        fmt.Sprintf("/v1/warehouses/%s", sd.Warehouses[0].ID),
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
			URL:        fmt.Sprintf("/v1/warehouses/%s", sd.Warehouses[0].ID),
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
			URL:        fmt.Sprintf("/v1/warehouses/%s", sd.Warehouses[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: warehouses"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
