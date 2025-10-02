package data_test

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/dataapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventoryitemapp"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

// NOTE: This is just json.Marshal() on the result of the simple configuration
// in tablebuilder_test.go. These get the same execution result just with
// different query criteria.

func execute200(sd apitest.SeedData) []apitest.Table {

	q := dataapp.TableQuery{
		Page:    1,
		Rows:    10,
		Sort:    []dataapp.SortParam{},
		Filters: []dataapp.FilterParam{},
		Dynamic: map[string]any{},
	}

	// Sort inventory items by quantity descending, then take first 10
	sortedItems := make([]inventoryitemapp.InventoryItem, len(sd.InventoryItems))
	copy(sortedItems, sd.InventoryItems)

	sort.Slice(sortedItems, func(i, j int) bool {
		return sortedItems[i].Quantity > sortedItems[j].Quantity
	})

	// Build expected data from sorted items
	var expData []map[string]any
	for i := 0; i < 10 && i < len(sortedItems); i++ {
		item := sortedItems[i]
		parsedValue, _ := strconv.ParseFloat(item.Quantity, 64)

		expData = append(expData, map[string]any{
			"ids": map[string]any{
				"id":       item.ID,
				"location": item.LocationID,
				"product":  item.ProductID,
			},
			"quantity": map[string]any{
				"tableColumn": "inventory_items.quantity",
				"value":       parsedValue,
			},
		})
	}

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/data/execute/%s", sd.SimpleTableConfig.ID),
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Input:      q,
			Method:     http.MethodPost,
			GotResp:    &dataapp.TableData{},
			ExpResp: &dataapp.TableData{
				Data: expData,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*dataapp.TableData)
				if !exists {
					return "could not convert got to *dataapp.TableConfig"
				}
				expResp, exists := exp.(*dataapp.TableData)
				if !exists {
					return "could not convert exp to *dataapp.TableConfig"
				}

				// Copy over the generated fields from got to exp
				expResp.Meta = gotResp.Meta

				dbtest.NormalizeJSONFields(gotResp, &expResp)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func executeByName200(sd apitest.SeedData) []apitest.Table {

	q := dataapp.TableQuery{
		Page:    1,
		Rows:    10,
		Sort:    []dataapp.SortParam{},
		Filters: []dataapp.FilterParam{},
		Dynamic: map[string]any{},
	}

	// Sort inventory items by quantity descending, then take first 10
	sortedItems := make([]inventoryitemapp.InventoryItem, len(sd.InventoryItems))
	copy(sortedItems, sd.InventoryItems)

	sort.Slice(sortedItems, func(i, j int) bool {
		return sortedItems[i].Quantity > sortedItems[j].Quantity
	})

	// Build expected data from sorted items
	var expData []map[string]any
	for i := 0; i < 10 && i < len(sortedItems); i++ {
		item := sortedItems[i]
		parsedValue, _ := strconv.ParseFloat(item.Quantity, 64)

		expData = append(expData, map[string]any{
			"ids": map[string]any{
				"id":       item.ID,
				"location": item.LocationID,
				"product":  item.ProductID,
			},
			"quantity": map[string]any{
				"tableColumn": "inventory_items.quantity",
				"value":       parsedValue,
			},
		})
	}

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/data/execute/name/orders_dashboard",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Input:      q,
			Method:     http.MethodPost,
			GotResp:    &dataapp.TableData{},
			ExpResp: &dataapp.TableData{
				Data: expData,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*dataapp.TableData)
				if !exists {
					return "could not convert got to *dataapp.TableConfig"
				}
				expResp, exists := exp.(*dataapp.TableData)
				if !exists {
					return "could not convert exp to *dataapp.TableConfig"
				}

				// Copy over the generated fields from got to exp
				expResp.Meta = gotResp.Meta

				dbtest.NormalizeJSONFields(gotResp, &expResp)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}
