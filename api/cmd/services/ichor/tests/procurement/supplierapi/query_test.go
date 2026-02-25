package supplierapi_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"

	"github.com/timmaaaz/ichor/app/domain/procurement/supplierapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/procurement/suppliers?page=1&rows=10",
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &query.Result[supplierapp.Supplier]{},
			ExpResp: &query.Result[supplierapp.Supplier]{
				Page:        1,
				RowsPerPage: 10,
				Total:       10,
				Items:       sd.Suppliers[:10],
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/procurement/suppliers/" + sd.Suppliers[0].SupplierID,
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &supplierapp.Supplier{},
			ExpResp:    &sd.Suppliers[0],
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func queryByIDs200(sd apitest.SeedData) []apitest.Table {
	ids := []string{sd.Suppliers[0].SupplierID, sd.Suppliers[1].SupplierID}
	expected := supplierapp.Suppliers{sd.Suppliers[0], sd.Suppliers[1]}

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/procurement/suppliers/batch",
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "POST",
			Input: &supplierapp.QueryByIDsRequest{
				IDs: ids,
			},
			GotResp: &supplierapp.Suppliers{},
			ExpResp: &expected,
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
