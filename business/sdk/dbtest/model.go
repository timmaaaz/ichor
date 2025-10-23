package dbtest

import "github.com/timmaaaz/ichor/business/sdk/tablebuilder"

// =============================================================================
// TABLE CONFIGS
// =============================================================================
var PageConfig = &tablebuilder.Config{
	Title:           "Products List",
	WidgetType:      "table",
	Visualization:   "table",
	PositionX:       0,
	PositionY:       0,
	Width:           12,
	Height:          8,
	RefreshInterval: 300,
	RefreshMode:     "polling",
	DataSource: []tablebuilder.DataSource{
		{
			Type:   "query",
			Source: "products",
			Schema: "products",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "products.id"},
					{Name: "name", TableColumn: "products.name"},
					{Name: "sku", TableColumn: "products.sku"},
					{Name: "is_active", TableColumn: "products.is_active"},
				},
			},
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"name": {
				Name:       "name",
				Header:     "Product Name",
				Width:      200,
				Sortable:   true,
				Filterable: true,
			},
			"sku": {
				Name:       "sku",
				Header:     "SKU",
				Width:      150,
				Sortable:   true,
				Filterable: true,
				Editable: &tablebuilder.EditableConfig{
					Type:        "text",
					Placeholder: "SKU-12345",
				},
			},
			"is_active": {
				Name:       "is_active",
				Header:     "Is Active",
				Width:      100,
				Sortable:   true,
				Filterable: true,
				Editable: &tablebuilder.EditableConfig{
					Type: "boolean",
				},
			},
		},
		Pagination: &tablebuilder.PaginationConfig{
			Enabled:         true,
			PageSizes:       []int{10, 25, 50, 100},
			DefaultPageSize: 25,
		},
	},
	Permissions: tablebuilder.Permissions{
		Roles:   []string{"admin"},
		Actions: []string{"view"},
	},
}

var ComplexConfig = &tablebuilder.Config{
	Title:           "Current Inventory with Products",
	WidgetType:      "table",
	Visualization:   "table",
	PositionX:       0,
	PositionY:       0,
	Width:           12,
	Height:          8,
	RefreshInterval: 300,
	RefreshMode:     "polling",
	DataSource: []tablebuilder.DataSource{
		{
			Type:   "query",
			Source: "inventory_items",
			Schema: "inventory",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "inventory_items.id"},
					{Name: "quantity", Alias: "current_quantity", TableColumn: "inventory_items.quantity"},
					{Name: "reorder_point", TableColumn: "inventory_items.reorder_point"},
					{Name: "maximum_stock", TableColumn: "inventory_items.maximum_stock"},
				},
				ForeignTables: []tablebuilder.ForeignTable{

					{
						Table:            "products",
						Schema:           "products",
						RelationshipFrom: "inventory_items.product_id",
						RelationshipTo:   "products.id",
						JoinType:         "inner",
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
					{
						Name:       "stock_percentage",
						Expression: "(current_quantity / maximum_stock) * 100",
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
			Sort: []tablebuilder.Sort{
				{
					Column:    "quantity",
					Direction: "asc",
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"product_name": {
				Name:       "product_name",
				Header:     "Product",
				Width:      250,
				Sortable:   true,
				Filterable: true,
			},
			"current_quantity": {
				Name:   "current_quantity",
				Header: "Current Stock",
				Width:  120,
				Align:  "right",
				Format: &tablebuilder.FormatConfig{
					Type:      "number",
					Precision: 0,
				},
			},
			"stock_status": {
				Name:         "stock_status",
				Header:       "Status",
				Width:        100,
				Align:        "center",
				CellTemplate: "status",
			},
			"stock_percentage": {
				Name:   "stock_percentage",
				Header: "Capacity",
				Width:  100,
				Align:  "right",
				Format: &tablebuilder.FormatConfig{
					Type:      "percent",
					Precision: 1,
				},
			},
			"product_id": {
				Name:   "product_id",
				Header: "Product",
				Width:  200,
				Link: &tablebuilder.LinkConfig{
					URL:   "/products/products/{product_id}",
					Label: "View Product",
				},
			},
		},
		ConditionalFormatting: []tablebuilder.ConditionalFormat{
			{
				Column:     "stock_status",
				Condition:  "eq",
				Value:      "low",
				Color:      "#ff4444",
				Background: "#ffebee",
				Icon:       "alert-circle",
			},
			{
				Column:     "stock_status",
				Condition:  "eq",
				Value:      "normal",
				Color:      "#00C851",
				Background: "#e8f5e9",
				Icon:       "check-circle",
			},
		},
	},
	Permissions: tablebuilder.Permissions{
		Roles:   []string{"admin", "inventory_manager"},
		Actions: []string{"view", "export", "adjust"},
	},
}

var ordersConfig = &tablebuilder.Config{
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
			"order_number": {
				Name:       "order_number",
				Header:     "Order #",
				Width:      150,
				Sortable:   true,
				Filterable: true,
			},
			"order_date": {
				Name:   "order_date",
				Header: "Order Date",
				Width:  120,
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "2006-01-02",
				},
			},
			"fulfillment_status_name": {
				Name:   "fulfillment_status_name",
				Header: "Status",
				Width:  120,
			},
		},
		ConditionalFormatting: []tablebuilder.ConditionalFormat{},
	},
	Permissions: tablebuilder.Permissions{
		Roles:   []string{"admin", "sales"},
		Actions: []string{"view", "export"},
	},
}

// =============================================================================
// PAGE CONFIGS
// =============================================================================

// =============================================================================
// FORM CONFIGS
// =============================================================================
