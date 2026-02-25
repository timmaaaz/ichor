package lottrackingsapi_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"

	"github.com/timmaaaz/ichor/app/domain/inventory/lottrackingsapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func queryByProductID200(sd apitest.SeedData) []apitest.Table {
	// Each supplier product references a unique product (15 SPs from 30 products, consecutive indices).
	// Each supplier product has exactly 1 lot tracking (15 LTs for 15 SPs).
	targetProductID := sd.SupplierProducts[0].ProductID
	targetSupplierProductID := sd.SupplierProducts[0].SupplierProductID

	// Find the lot tracking that references the first supplier product.
	var expectedItem lottrackingsapp.LotTrackings
	for _, lt := range sd.LotTrackings {
		if lt.SupplierProductID == targetSupplierProductID {
			expectedItem = lt
			break
		}
	}

	table := []apitest.Table{
		{
			Name:       "by-product-id",
			URL:        "/v1/inventory/lot-trackings?page=1&rows=10&product_id=" + targetProductID,
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &query.Result[lottrackingsapp.LotTrackings]{},
			ExpResp: &query.Result[lottrackingsapp.LotTrackings]{
				Page:        1,
				RowsPerPage: 10,
				Total:       1,
				Items:       []lottrackingsapp.LotTrackings{expectedItem},
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/lot-trackings?page=1&rows=10",
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &query.Result[lottrackingsapp.LotTrackings]{},
			ExpResp: &query.Result[lottrackingsapp.LotTrackings]{
				Page:        1,
				RowsPerPage: 10,
				Total:       15,
				Items:       sd.LotTrackings[:10],
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
			URL:        "/v1/inventory/lot-trackings/" + sd.LotTrackings[0].LotID,
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &lottrackingsapp.LotTrackings{},
			ExpResp:    &sd.LotTrackings[0],
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
