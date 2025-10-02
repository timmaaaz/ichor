package data_test

import (
	"encoding/json"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/dataapp"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)

var newConfig = &tablebuilder.Config{
	Title:           "Current Orders and Associated data",
	WidgetType:      "table",
	Visualization:   "table",
	PositionX:       0,
	PositionY:       0,
	Width:           6,
	Height:          4,
	RefreshInterval: 300,
	RefreshMode:     "polling",
	DataSource: []tablebuilder.DataSource{
		{
			Type:   "view",
			Source: "orders_base",
			Schema: "sales",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					// order
					{Name: "orders_id", TableColumn: "orders.id"},
					{Name: "orders_number", Alias: "order_number", TableColumn: "orders.number"},
					{Name: "orders_order_date", Alias: "order_date", TableColumn: "orders.created_date"},
					{Name: "orders_due_date", Alias: "order_due_date", TableColumn: "orders.due_date"},
					{Name: "orders_created_date", Alias: "order_created_date", TableColumn: "orders.created_date"},
					{Name: "orders_updated_date", Alias: "order_updated_date", TableColumn: "orders.updated_date"},
					{Name: "orders_fulfillment_status_id", Alias: "order_fulfillment_status_id", TableColumn: "orders.fulfillment_status_id"},
					{Name: "orders_customer_id", Alias: "order_customer_id", TableColumn: "orders.customer_id"},
					// customers
					{Name: "customers_id", Alias: "customer_id", TableColumn: "customers.id"},
					{Name: "customers_contact_infos_id", Alias: "customer_contact_info_id", TableColumn: "customers.contact_id"},
					{Name: "customers_delivery_address_id", Alias: "customer_delivery_address_id", TableColumn: "customers.delivery_address_id"},
					{Name: "customers_notes", Alias: "customer_notes", TableColumn: "customers.notes"},
					{Name: "customers_created_date", Alias: "customer_created_date", TableColumn: "customers.created_date"},
					{Name: "customers_updated_date", Alias: "customer_updated_date", TableColumn: "customers.updated_date"},
					// order_fulfillment_statuses
					{Name: "order_fulfillment_statuses_name", Alias: "order_fulfillment_statuses_name", TableColumn: "order_fulfillment_statuses.name"},
					{Name: "order_fulfillment_statuses_description", Alias: "order_fulfillment_statuses_description", TableColumn: "order_fulfillment_statuses.description"},
				},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "products",
						Schema:           "products",                   // Optional, defaults to public
						RelationshipFrom: "inventory_items.product_id", // CHANGED
						RelationshipTo:   "products.id",                // CHANGED
						JoinType:         "inner",                      // Optional, defaults to inner
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "id", Alias: "product_id", TableColumn: "products.id"},
							{Name: "name", Alias: "product_name", TableColumn: "products.name"},
							{Name: "sku", TableColumn: "products.sku"},
						},
					},
				},
				ClientComputedColumns: []tablebuilder.ComputedColumn{
					{
						Name:       "stock_status",
						Expression: "current_quantity <= reorder_point ? 'low' : 'normal'",
					},
				},
			},
			Filters: []tablebuilder.Filter{
				{
					Column:   "quantity",
					Operator: "gt",
					Value:    0,
				},
			},
			Limit: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns:               map[string]tablebuilder.ColumnConfig{},
		ConditionalFormatting: []tablebuilder.ConditionalFormat{},
	},
	Permissions: tablebuilder.Permissions{
		Roles:   []string{"admin", "inventory_manager"},
		Actions: []string{"view", "export", "adjust"},
	},
}

func create200(sd apitest.SeedData) []apitest.Table {

	jsonConfig, err := json.Marshal(newConfig)
	if err != nil {
		panic(err)
	}

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/data",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: dataapp.NewTableConfig{
				Name:        "Orders",
				Description: "Current orders with customer and fulfillment status",
				Config:      jsonConfig,
			},
			GotResp: &dataapp.TableConfig{},
			ExpResp: &dataapp.TableConfig{
				Name:        "Orders",
				Description: "Current orders with customer and fulfillment status",
				Config:      jsonConfig,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*dataapp.TableConfig)
				if !exists {
					return "could not convert got to *dataapp.TableConfig"
				}
				expResp, exists := exp.(*dataapp.TableConfig)
				if !exists {
					return "could not convert exp to *dataapp.TableConfig"
				}

				// Copy over the generated fields from got to exp
				expResp.ID = gotResp.ID
				expResp.CreatedBy = gotResp.CreatedBy
				expResp.UpdatedBy = gotResp.UpdatedBy
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				dbtest.NormalizeJSONFields(gotResp, &expResp)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}
