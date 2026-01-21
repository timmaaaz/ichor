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
					// orders table
					{Name: "orders_id", TableColumn: "orders.id"},
					{Name: "orders_number", Alias: "order_number", TableColumn: "orders.number"},
					{Name: "orders_order_date", Alias: "order_date", TableColumn: "orders.order_date"},
					{Name: "orders_due_date", Alias: "order_due_date", TableColumn: "orders.due_date"},
					{Name: "orders_created_date", Alias: "order_created_date", TableColumn: "orders.created_date"},
					{Name: "orders_updated_date", Alias: "order_updated_date", TableColumn: "orders.updated_date"},
					{Name: "orders_fulfillment_status_id", Alias: "order_fulfillment_status_id", TableColumn: "orders.fulfillment_status_id"},
					{Name: "orders_customer_id", Alias: "order_customer_id", TableColumn: "orders.customer_id"},

					// customers table
					{Name: "customers_id", Alias: "customer_id", TableColumn: "customers.id"},
					{Name: "customers_contact_infos_id", Alias: "customer_contact_info_id", TableColumn: "customers.contact_id"},
					{Name: "customers_delivery_address_id", Alias: "customer_delivery_address_id", TableColumn: "customers.delivery_address_id"},
					{Name: "customers_notes", Alias: "customer_notes", TableColumn: "customers.notes"},
					{Name: "customers_created_date", Alias: "customer_created_date", TableColumn: "customers.created_date"},
					{Name: "customers_updated_date", Alias: "customer_updated_date", TableColumn: "customers.updated_date"},

					// order_fulfillment_statuses table
					{Name: "order_fulfillment_statuses_name", Alias: "fulfillment_status_name", TableColumn: "order_fulfillment_statuses.name"},
					{Name: "order_fulfillment_statuses_description", Alias: "fulfillment_status_description", TableColumn: "order_fulfillment_statuses.description"},
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"orders.id": {
				Name:   "orders.id",
				Header: "Order ID",
				Width:  100,
				Type:   "uuid",
			},
			"order_number": {
				Name:       "order_number",
				Header:     "Order #",
				Width:      150,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"order_date": {
				Name:   "order_date",
				Header: "Order Date",
				Width:  120,
				Type:   "datetime",
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "2006-01-02",
				},
			},
			"fulfillment_status_name": {
				Name:   "fulfillment_status_name",
				Header: "Status",
				Width:  120,
				Type:   "status",
			},
			"order_due_date": {
				Name:   "order_due_date",
				Header: "Due Date",
				Width:  120,
				Type:   "datetime",
			},
			"order_created_date": {
				Name:   "order_created_date",
				Header: "Created Date",
				Width:  120,
				Type:   "datetime",
			},
			"order_updated_date": {
				Name:   "order_updated_date",
				Header: "Updated Date",
				Width:  120,
				Type:   "datetime",
			},
			"order_fulfillment_status_id": {
				Name:   "order_fulfillment_status_id",
				Header: "Fulfillment Status ID",
				Width:  100,
				Type:   "uuid",
			},
			"order_customer_id": {
				Name:   "order_customer_id",
				Header: "Customer ID",
				Width:  100,
				Type:   "uuid",
			},
			"customer_id": {
				Name:   "customer_id",
				Header: "Customer ID",
				Width:  100,
				Type:   "uuid",
			},
			"customer_contact_info_id": {
				Name:   "customer_contact_info_id",
				Header: "Contact Info ID",
				Width:  100,
				Type:   "uuid",
			},
			"customer_delivery_address_id": {
				Name:   "customer_delivery_address_id",
				Header: "Delivery Address ID",
				Width:  100,
				Type:   "uuid",
			},
			"customer_notes": {
				Name:   "customer_notes",
				Header: "Customer Notes",
				Width:  200,
				Type:   "string",
			},
			"customer_created_date": {
				Name:   "customer_created_date",
				Header: "Customer Created",
				Width:  120,
				Type:   "datetime",
			},
			"customer_updated_date": {
				Name:   "customer_updated_date",
				Header: "Customer Updated",
				Width:  120,
				Type:   "datetime",
			},
			"fulfillment_status_description": {
				Name:   "fulfillment_status_description",
				Header: "Status Description",
				Width:  200,
				Type:   "string",
			},
		},
		ConditionalFormatting: []tablebuilder.ConditionalFormat{},
	},
	Permissions: tablebuilder.Permissions{
		Roles:   []string{"admin", "sales"},
		Actions: []string{"view", "export"},
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
