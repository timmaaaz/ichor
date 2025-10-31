package dbtest

import (
	"github.com/timmaaaz/ichor/business/domain/core/pagebus"
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)

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

// Dedicated Orders Page Config
var ordersPageConfig = &tablebuilder.Config{
	Title:           "Orders Management",
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
			Source: "orders",
			Schema: "sales",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "orders.id"},
					{Name: "number", Alias: "order_number", TableColumn: "orders.number"},
					{Name: "due_date", TableColumn: "orders.due_date"},
					{Name: "customer_id", TableColumn: "orders.customer_id"},
					{Name: "order_fulfillment_status_id", TableColumn: "orders.order_fulfillment_status_id"},
					{Name: "created_date", TableColumn: "orders.created_date"},
					{Name: "updated_date", TableColumn: "orders.updated_date"},
				},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "customers",
						Schema:           "sales",
						RelationshipFrom: "orders.customer_id",
						RelationshipTo:   "customers.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "id", Alias: "customer_id_fk", TableColumn: "customers.id"},
							{Name: "name", Alias: "customer_name", TableColumn: "customers.name"},
							{Name: "contact_id", Alias: "customer_contact_id", TableColumn: "customers.contact_id"},
						},
					},
					{
						Table:            "order_fulfillment_statuses",
						Schema:           "sales",
						RelationshipFrom: "orders.order_fulfillment_status_id",
						RelationshipTo:   "order_fulfillment_statuses.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "status_name", TableColumn: "order_fulfillment_statuses.name"},
							{Name: "description", Alias: "status_description", TableColumn: "order_fulfillment_statuses.description"},
						},
					},
				},
				ClientComputedColumns: []tablebuilder.ComputedColumn{
					{
						Name:       "days_until_due",
						Expression: "Math.ceil((new Date(due_date) - new Date()) / (1000 * 60 * 60 * 24))",
					},
					{
						Name:       "is_overdue",
						Expression: "new Date(due_date) < new Date() && status_name !== 'Delivered'",
					},
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "due_date",
					Direction: "desc",
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
			"customer_name": {
				Name:       "customer_name",
				Header:     "Customer",
				Width:      200,
				Sortable:   true,
				Filterable: true,
			},
			"due_date": {
				Name:     "due_date",
				Header:   "Due Date",
				Width:    120,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "2006-01-02",
				},
			},
			"status_name": {
				Name:         "status_name",
				Header:       "Status",
				Width:        150,
				Sortable:     true,
				Filterable:   true,
				CellTemplate: "status",
			},
			"days_until_due": {
				Name:   "days_until_due",
				Header: "Days Until Due",
				Width:  120,
				Align:  "center",
				Format: &tablebuilder.FormatConfig{
					Type:      "number",
					Precision: 0,
				},
			},
			"id": {
				Name:   "id",
				Header: "Actions",
				Width:  100,
				Link: &tablebuilder.LinkConfig{
					URL:   "/sales/orders/{id}",
					Label: "View",
				},
			},
		},
		ConditionalFormatting: []tablebuilder.ConditionalFormat{
			{
				Column:     "is_overdue",
				Condition:  "eq",
				Value:      true,
				Color:      "#d32f2f",
				Background: "#ffebee",
				Icon:       "alert-triangle",
			},
			{
				Column:     "days_until_due",
				Condition:  "lte",
				Value:      3,
				Condition2: "gt",
				Value2:     0,
				Color:      "#f57c00",
				Background: "#fff3e0",
				Icon:       "clock",
			},
		},
		Pagination: &tablebuilder.PaginationConfig{
			Enabled:         true,
			PageSizes:       []int{10, 25, 50, 100},
			DefaultPageSize: 25,
		},
	},
	Permissions: tablebuilder.Permissions{
		Roles:   []string{"admin", "sales"},
		Actions: []string{"view", "edit", "export"},
	},
}

// Suppliers Page Config
var suppliersPageConfig = &tablebuilder.Config{
	Title:           "Suppliers Management",
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
			Source: "suppliers",
			Schema: "procurement",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "suppliers.id"},
					{Name: "name", TableColumn: "suppliers.name"},
					{Name: "payment_terms", TableColumn: "suppliers.payment_terms"},
					{Name: "lead_time_days", TableColumn: "suppliers.lead_time_days"},
					{Name: "rating", TableColumn: "suppliers.rating"},
					{Name: "is_active", TableColumn: "suppliers.is_active"},
					{Name: "contact_infos_id", TableColumn: "suppliers.contact_infos_id"},
					{Name: "created_date", TableColumn: "suppliers.created_date"},
					{Name: "updated_date", TableColumn: "suppliers.updated_date"},
				},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "contact_infos",
						Schema:           "core",
						RelationshipFrom: "suppliers.contact_infos_id",
						RelationshipTo:   "contact_infos.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "primary_phone_number", Alias: "primary_phone_number", TableColumn: "contact_infos.primary_phone_number"},
							{Name: "secondary_phone_number", Alias: "secondary_phone_number", TableColumn: "contact_infos.secondary_phone_number"},
							{Name: "email_address", Alias: "email_address", TableColumn: "contact_infos.email_address"},
						},
					},
				},
				ClientComputedColumns: []tablebuilder.ComputedColumn{
					{
						Name:       "rating_stars",
						Expression: "Math.round(rating * 2) / 2",
					},
					{
						Name:       "performance_level",
						Expression: "rating >= 4.5 ? 'excellent' : rating >= 3.5 ? 'good' : rating >= 2.5 ? 'fair' : 'poor'",
					},
				},
			},
			Filters: []tablebuilder.Filter{
				{
					Column:   "is_active",
					Operator: "eq",
					Value:    true,
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "rating",
					Direction: "desc",
				},
				{
					Column:    "name",
					Direction: "asc",
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"name": {
				Name:       "name",
				Header:     "Supplier Name",
				Width:      200,
				Sortable:   true,
				Filterable: true,
			},
			"contact_email": {
				Name:       "contact_email",
				Header:     "Email",
				Width:      200,
				Filterable: true,
			},
			"contact_phone": {
				Name:       "contact_phone",
				Header:     "Phone",
				Width:      150,
				Filterable: true,
			},
			"payment_terms": {
				Name:       "payment_terms",
				Header:     "Payment Terms",
				Width:      150,
				Sortable:   true,
				Filterable: true,
			},
			"lead_time_days": {
				Name:     "lead_time_days",
				Header:   "Lead Time (days)",
				Width:    130,
				Align:    "center",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:      "number",
					Precision: 0,
				},
			},
			"rating": {
				Name:     "rating",
				Header:   "Rating",
				Width:    100,
				Align:    "center",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:      "number",
					Precision: 2,
				},
			},
			"performance_level": {
				Name:         "performance_level",
				Header:       "Performance",
				Width:        120,
				Align:        "center",
				CellTemplate: "badge",
			},
			"is_active": {
				Name:   "is_active",
				Header: "Active",
				Width:  80,
				Align:  "center",
				Format: &tablebuilder.FormatConfig{
					Type: "boolean",
				},
				Editable: &tablebuilder.EditableConfig{
					Type: "boolean",
				},
			},
			"supplier_id": {
				Name:   "supplier_id",
				Header: "Actions",
				Width:  100,
				Link: &tablebuilder.LinkConfig{
					URL:   "/procurement/suppliers/{supplier_id}",
					Label: "View",
				},
			},
		},
		ConditionalFormatting: []tablebuilder.ConditionalFormat{
			{
				Column:     "performance_level",
				Condition:  "eq",
				Value:      "excellent",
				Color:      "#1b5e20",
				Background: "#c8e6c9",
				Icon:       "star",
			},
			{
				Column:     "performance_level",
				Condition:  "eq",
				Value:      "good",
				Color:      "#2e7d32",
				Background: "#e8f5e9",
				Icon:       "check-circle",
			},
			{
				Column:     "performance_level",
				Condition:  "eq",
				Value:      "fair",
				Color:      "#f57c00",
				Background: "#fff3e0",
				Icon:       "alert-circle",
			},
			{
				Column:     "performance_level",
				Condition:  "eq",
				Value:      "poor",
				Color:      "#c62828",
				Background: "#ffebee",
				Icon:       "x-circle",
			},
			{
				Column:     "is_active",
				Condition:  "eq",
				Value:      false,
				Color:      "#757575",
				Background: "#f5f5f5",
			},
		},
		Pagination: &tablebuilder.PaginationConfig{
			Enabled:         true,
			PageSizes:       []int{10, 25, 50, 100},
			DefaultPageSize: 25,
		},
	},
	Permissions: tablebuilder.Permissions{
		Roles:   []string{"admin", "procurement", "inventory_manager"},
		Actions: []string{"view", "edit", "export"},
	},
}

// Order Line Items Page Config
var orderLineItemsPageConfig = &tablebuilder.Config{
	Title:           "Order Line Items",
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
			Source: "order_line_items",
			Schema: "sales",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "order_line_items.id"},
					{Name: "order_id", TableColumn: "order_line_items.order_id"},
					{Name: "product_id", TableColumn: "order_line_items.product_id"},
					{Name: "quantity", TableColumn: "order_line_items.quantity"},
					{Name: "discount", TableColumn: "order_line_items.discount"},
					{Name: "line_item_fulfillment_statuses_id", TableColumn: "order_line_items.line_item_fulfillment_statuses_id"},
					{Name: "created_by", TableColumn: "order_line_items.created_by"},
					{Name: "created_date", TableColumn: "order_line_items.created_date"},
					{Name: "updated_by", TableColumn: "order_line_items.updated_by"},
					{Name: "updated_date", TableColumn: "order_line_items.updated_date"},
				},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "orders",
						Schema:           "sales",
						RelationshipFrom: "order_line_items.order_id",
						RelationshipTo:   "orders.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "number", Alias: "order_number", TableColumn: "orders.number"},
							{Name: "order_date", Alias: "order_date", TableColumn: "orders.order_date"},
							{Name: "customer_id", Alias: "order_customer_id", TableColumn: "orders.customer_id"},
						},
						ForeignTables: []tablebuilder.ForeignTable{
							{
								Table:            "customers",
								Schema:           "sales",
								RelationshipFrom: "orders.customer_id",
								RelationshipTo:   "customers.id",
								JoinType:         "left",
								Columns: []tablebuilder.ColumnDefinition{
									{Name: "name", Alias: "customer_name", TableColumn: "customers.name"},
								},
							},
						},
					},
					{
						Table:            "products",
						Schema:           "products",
						RelationshipFrom: "order_line_items.product_id",
						RelationshipTo:   "products.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "product_name", TableColumn: "products.name"},
							{Name: "sku", Alias: "product_sku", TableColumn: "products.sku"},
						},
					},
					{
						Table:            "line_item_fulfillment_statuses",
						Schema:           "sales",
						RelationshipFrom: "order_line_items.line_item_fulfillment_statuses_id",
						RelationshipTo:   "line_item_fulfillment_statuses.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "fulfillment_status_name", TableColumn: "line_item_fulfillment_statuses.name"},
							{Name: "description", Alias: "fulfillment_status_description", TableColumn: "line_item_fulfillment_statuses.description"},
						},
					},
					{
						Table:            "users",
						Schema:           "core",
						RelationshipFrom: "order_line_items.created_by",
						RelationshipTo:   "users.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "username", Alias: "created_by_username", TableColumn: "users.username"},
							{Name: "first_name", Alias: "created_by_first_name", TableColumn: "users.first_name"},
							{Name: "last_name", Alias: "created_by_last_name", TableColumn: "users.last_name"},
						},
					},
				},
				ClientComputedColumns: []tablebuilder.ComputedColumn{
					{
						Name:       "line_total",
						Expression: "quantity * (product_unit_price || 0) * (1 - (discount || 0) / 100)",
					},
					{
						Name:       "discount_amount",
						Expression: "(quantity * (product_unit_price || 0)) * ((discount || 0) / 100)",
					},
					{
						Name:       "created_by_full_name",
						Expression: "(created_by_first_name || '') + ' ' + (created_by_last_name || '')",
					},
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "order_date",
					Direction: "desc",
				},
				{
					Column:    "id",
					Direction: "asc",
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
				Link: &tablebuilder.LinkConfig{
					URL:   "/sales/orders/{order_id}",
					Label: "{order_number}",
				},
			},
			"customer_name": {
				Name:       "customer_name",
				Header:     "Customer",
				Width:      180,
				Sortable:   true,
				Filterable: true,
			},
			"product_name": {
				Name:       "product_name",
				Header:     "Product",
				Width:      200,
				Sortable:   true,
				Filterable: true,
			},
			"product_sku": {
				Name:       "product_sku",
				Header:     "SKU",
				Width:      120,
				Filterable: true,
			},
			"quantity": {
				Name:     "quantity",
				Header:   "Qty",
				Width:    80,
				Align:    "center",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:      "number",
					Precision: 0,
				},
				Editable: &tablebuilder.EditableConfig{
					Type:        "number",
					Placeholder: "0",
				},
			},
			"discount": {
				Name:     "discount",
				Header:   "Discount %",
				Width:    100,
				Align:    "right",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:      "number",
					Precision: 2,
				},
				Editable: &tablebuilder.EditableConfig{
					Type:        "number",
					Placeholder: "0.00",
				},
			},
			"discount_amount": {
				Name:   "discount_amount",
				Header: "Discount $",
				Width:  100,
				Align:  "right",
				Format: &tablebuilder.FormatConfig{
					Type:      "currency",
					Currency:  "USD",
					Precision: 2,
				},
			},
			"line_total": {
				Name:   "line_total",
				Header: "Line Total",
				Width:  120,
				Align:  "right",
				Format: &tablebuilder.FormatConfig{
					Type:      "currency",
					Currency:  "USD",
					Precision: 2,
				},
			},
			"fulfillment_status_name": {
				Name:         "fulfillment_status_name",
				Header:       "Status",
				Width:        130,
				Sortable:     true,
				Filterable:   true,
				CellTemplate: "status",
			},
			"order_date": {
				Name:     "order_date",
				Header:   "Order Date",
				Width:    120,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "2006-01-02",
				},
			},
			"created_by_full_name": {
				Name:   "created_by_full_name",
				Header: "Created By",
				Width:  150,
			},
			"created_date": {
				Name:     "created_date",
				Header:   "Created",
				Width:    150,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02 15:04",
				},
			},
			"id": {
				Name:   "id",
				Header: "Actions",
				Width:  100,
				Link: &tablebuilder.LinkConfig{
					URL:   "/sales/order-line-items/{id}",
					Label: "View",
				},
			},
		},
		ConditionalFormatting: []tablebuilder.ConditionalFormat{
			{
				Column:     "fulfillment_status_name",
				Condition:  "eq",
				Value:      "Pending",
				Color:      "#f57c00",
				Background: "#fff3e0",
				Icon:       "clock",
			},
			{
				Column:     "fulfillment_status_name",
				Condition:  "eq",
				Value:      "Shipped",
				Color:      "#1976d2",
				Background: "#e3f2fd",
				Icon:       "truck",
			},
			{
				Column:     "fulfillment_status_name",
				Condition:  "eq",
				Value:      "Delivered",
				Color:      "#388e3c",
				Background: "#e8f5e9",
				Icon:       "check-circle",
			},
			{
				Column:     "fulfillment_status_name",
				Condition:  "eq",
				Value:      "Cancelled",
				Color:      "#d32f2f",
				Background: "#ffebee",
				Icon:       "x-circle",
			},
			{
				Column:     "discount",
				Condition:  "gt",
				Value:      0,
				Color:      "#c62828",
				Background: "#ffebee",
				Icon:       "percent",
			},
		},
		Pagination: &tablebuilder.PaginationConfig{
			Enabled:         true,
			PageSizes:       []int{10, 25, 50, 100},
			DefaultPageSize: 25,
		},
	},
	Permissions: tablebuilder.Permissions{
		Roles:   []string{"admin", "sales", "order_manager"},
		Actions: []string{"view", "edit", "export"},
	},
}

// Product Categories Page Config
var categoriesPageConfig = &tablebuilder.Config{
	Title:           "Product Categories",
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
			Source: "product_categories",
			Schema: "products",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "product_categories.id"},
					{Name: "name", TableColumn: "product_categories.name"},
					{Name: "description", TableColumn: "product_categories.description"},
					{Name: "created_date", TableColumn: "product_categories.created_date"},
					{Name: "updated_date", TableColumn: "product_categories.updated_date"},
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "name",
					Direction: "asc",
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"name": {
				Name:       "name",
				Header:     "Category Name",
				Width:      250,
				Sortable:   true,
				Filterable: true,
				Editable: &tablebuilder.EditableConfig{
					Type:        "text",
					Placeholder: "Category name",
				},
			},
			"description": {
				Name:       "description",
				Header:     "Description",
				Width:      400,
				Filterable: true,
				Editable: &tablebuilder.EditableConfig{
					Type:        "textarea",
					Placeholder: "Category description",
				},
			},
			"created_date": {
				Name:     "created_date",
				Header:   "Created",
				Width:    150,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02 15:04",
				},
			},
			"updated_date": {
				Name:     "updated_date",
				Header:   "Last Updated",
				Width:    150,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02 15:04",
				},
			},
			"product_category_id": {
				Name:   "product_category_id",
				Header: "Actions",
				Width:  100,
				Link: &tablebuilder.LinkConfig{
					URL:   "/products/categories/{product_category_id}",
					Label: "View",
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
		Roles:   []string{"admin", "product_manager"},
		Actions: []string{"view", "edit", "delete", "export"},
	},
}

// =============================================================================
// PAGE CONFIGS
// =============================================================================
const sortOrder = iota

var allPages = []pagebus.NewPage{
	// DASHBOARD MODULE
	{
		Path:       "/dashboard",
		Name:       "Main Dashboard",
		Module:     "dashboard",
		Icon:       "material-symbols:space-dashboard",
		SortOrder:  1,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/dashboard/analytics",
		Name:       "Analytics Dashboard",
		Module:     "dashboard",
		Icon:       "material-symbols:analytics",
		SortOrder:  2,
		IsActive:   true,
		ShowInMenu: true,
	},

	// SALES MODULE
	{
		Path:       "/sales",
		Name:       "Sales Dashboard",
		Module:     "sales",
		Icon:       "material-symbols:point-of-sale",
		SortOrder:  3,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/sales/orders",
		Name:       "Order Management",
		Module:     "sales",
		Icon:       "material-symbols:receipt-long",
		SortOrder:  4,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/sales/orders/new",
		Name:       "New Order Form",
		Module:     "sales",
		Icon:       "material-symbols:add-shopping-cart",
		SortOrder:  5,
		IsActive:   true,
		ShowInMenu: false,
	},
	{
		Path:       "/sales/customers",
		Name:       "Customer Management",
		Module:     "sales",
		Icon:       "material-symbols:group",
		SortOrder:  6,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/sales/customers/new",
		Name:       "New Customer Form",
		Module:     "sales",
		Icon:       "material-symbols:person-add",
		SortOrder:  7,
		IsActive:   true,
		ShowInMenu: false,
	},
	{
		Path:       "/sales/reports",
		Name:       "Sales Reports",
		Module:     "sales",
		Icon:       "material-symbols:assessment",
		SortOrder:  8,
		IsActive:   true,
		ShowInMenu: true,
	},

	// INVENTORY MODULE
	{
		Path:       "/inventory",
		Name:       "Inventory Dashboard",
		Module:     "inventory",
		Icon:       "material-symbols:inventory-2",
		SortOrder:  9,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/inventory/items",
		Name:       "Item Management",
		Module:     "inventory",
		Icon:       "material-symbols:category",
		SortOrder:  10,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/inventory/warehouses",
		Name:       "Warehouse Management",
		Module:     "inventory",
		Icon:       "material-symbols:warehouse",
		SortOrder:  11,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/inventory/transfers",
		Name:       "Transfer Orders",
		Module:     "inventory",
		Icon:       "material-symbols:sync-alt",
		SortOrder:  12,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/inventory/adjustments",
		Name:       "Stock Adjustments",
		Module:     "inventory",
		Icon:       "material-symbols:tune",
		SortOrder:  13,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/inventory/reports",
		Name:       "Inventory Reports",
		Module:     "inventory",
		Icon:       "material-symbols:summarize",
		SortOrder:  14,
		IsActive:   true,
		ShowInMenu: true,
	},

	// PROCUREMENT MODULE
	{
		Path:       "/procurement",
		Name:       "Procurement Dashboard",
		Module:     "procurement",
		Icon:       "material-symbols:shopping-cart",
		SortOrder:  15,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/procurement/suppliers",
		Name:       "Supplier Management",
		Module:     "procurement",
		Icon:       "material-symbols:local-shipping",
		SortOrder:  16,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/procurement/orders",
		Name:       "Purchase Orders",
		Module:     "procurement",
		Icon:       "material-symbols:shopping-bag",
		SortOrder:  17,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/procurement/approvals",
		Name:       "Approval Queue",
		Module:     "procurement",
		Icon:       "material-symbols:check-circle",
		SortOrder:  18,
		IsActive:   true,
		ShowInMenu: true,
	},

	// ASSETS MODULE
	{
		Path:       "/assets",
		Name:       "Asset Dashboard",
		Module:     "assets",
		Icon:       "material-symbols:apartment",
		SortOrder:  19,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/assets/list",
		Name:       "Asset List",
		Module:     "assets",
		Icon:       "material-symbols:list",
		SortOrder:  20,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/assets/requests",
		Name:       "Asset Requests",
		Module:     "assets",
		Icon:       "material-symbols:request-quote",
		SortOrder:  21,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/assets/maintenance",
		Name:       "Maintenance Schedule",
		Module:     "assets",
		Icon:       "material-symbols:build",
		SortOrder:  22,
		IsActive:   true,
		ShowInMenu: true,
	},

	// HR MODULE
	{
		Path:       "/hr",
		Name:       "HR Dashboard",
		Module:     "hr",
		Icon:       "material-symbols:badge",
		SortOrder:  23,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/hr/employees",
		Name:       "Employee Directory",
		Module:     "hr",
		Icon:       "material-symbols:groups",
		SortOrder:  24,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/hr/onboarding",
		Name:       "Onboarding",
		Module:     "hr",
		Icon:       "material-symbols:how-to-reg",
		SortOrder:  25,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/hr/offices",
		Name:       "Office Management",
		Module:     "hr",
		Icon:       "material-symbols:business",
		SortOrder:  26,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/hr/reports",
		Name:       "HR Reports",
		Module:     "hr",
		Icon:       "material-symbols:bar-chart",
		SortOrder:  27,
		IsActive:   true,
		ShowInMenu: true,
	},

	// ADMIN MODULE
	{
		Path:       "/admin",
		Name:       "Admin Dashboard",
		Module:     "admin",
		Icon:       "material-symbols:admin-panel-settings",
		SortOrder:  28,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/admin/users",
		Name:       "User Management",
		Module:     "admin",
		Icon:       "material-symbols:manage-accounts",
		SortOrder:  29,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/admin/roles",
		Name:       "Role Management",
		Module:     "admin",
		Icon:       "material-symbols:security",
		SortOrder:  30,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/admin/config",
		Name:       "System Configuration",
		Module:     "admin",
		Icon:       "material-symbols:settings-applications",
		SortOrder:  31,
		IsActive:   true,
		ShowInMenu: true,
	},
	{
		Path:       "/admin/audit",
		Name:       "Audit Logs",
		Module:     "admin",
		Icon:       "material-symbols:history",
		SortOrder:  32,
		IsActive:   true,
		ShowInMenu: true,
	},
}

// =============================================================================
// ADMIN MODULE CONFIGS
// =============================================================================

// Users Management Page Config
var adminUsersPageConfig = &tablebuilder.Config{
	Title:           "User Management",
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
			Source: "users",
			Schema: "core",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "users.id"},
					{Name: "username", TableColumn: "users.username"},
					{Name: "first_name", TableColumn: "users.first_name"},
					{Name: "last_name", TableColumn: "users.last_name"},
					{Name: "email", TableColumn: "users.email"},
					{Name: "enabled", TableColumn: "users.enabled"},
					{Name: "date_hired", TableColumn: "users.date_hired"},
					{Name: "created_date", TableColumn: "users.created_date"},
				},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "titles",
						Schema:           "hr",
						RelationshipFrom: "users.title_id",
						RelationshipTo:   "titles.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "title_name", TableColumn: "titles.name"},
						},
					},
					{
						Table:            "offices",
						Schema:           "hr",
						RelationshipFrom: "users.office_id",
						RelationshipTo:   "offices.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "office_name", TableColumn: "offices.name"},
						},
					},
					{
						Table:            "user_approval_status",
						Schema:           "hr",
						RelationshipFrom: "users.user_approval_status_id",
						RelationshipTo:   "user_approval_status.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "approval_status_name", TableColumn: "user_approval_status.name"},
						},
					},
				},
				ClientComputedColumns: []tablebuilder.ComputedColumn{
					{
						Name:       "full_name",
						Expression: "first_name + ' ' + last_name",
					},
					{
						Name:       "days_employed",
						Expression: "date_hired ? Math.floor((new Date() - new Date(date_hired)) / (1000 * 60 * 60 * 24)) : null",
					},
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "last_name",
					Direction: "asc",
				},
				{
					Column:    "first_name",
					Direction: "asc",
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"full_name": {
				Name:       "full_name",
				Header:     "Name",
				Width:      200,
				Sortable:   true,
				Filterable: true,
			},
			"username": {
				Name:       "username",
				Header:     "Username",
				Width:      150,
				Sortable:   true,
				Filterable: true,
			},
			"email": {
				Name:       "email",
				Header:     "Email",
				Width:      250,
				Filterable: true,
			},
			"title_name": {
				Name:       "title_name",
				Header:     "Title",
				Width:      150,
				Sortable:   true,
				Filterable: true,
			},
			"office_name": {
				Name:       "office_name",
				Header:     "Office",
				Width:      150,
				Sortable:   true,
				Filterable: true,
			},
			"approval_status_name": {
				Name:       "approval_status_name",
				Header:     "Status",
				Width:      120,
				Sortable:   true,
				Filterable: true,
			},
			"enabled": {
				Name:   "enabled",
				Header: "Active",
				Width:  80,
				Align:  "center",
				Format: &tablebuilder.FormatConfig{
					Type: "boolean",
				},
			},
			"date_hired": {
				Name:     "date_hired",
				Header:   "Date Hired",
				Width:    120,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "2006-01-02",
				},
			},
			"days_employed": {
				Name:   "days_employed",
				Header: "Days Employed",
				Width:  120,
				Align:  "right",
				Format: &tablebuilder.FormatConfig{
					Type:      "number",
					Precision: 0,
				},
			},
			"id": {
				Name:   "id",
				Header: "Actions",
				Width:  100,
				Link: &tablebuilder.LinkConfig{
					URL:   "/admin/users/{id}",
					Label: "View",
				},
			},
		},
		ConditionalFormatting: []tablebuilder.ConditionalFormat{
			{
				Column:     "enabled",
				Condition:  "eq",
				Value:      false,
				Color:      "#757575",
				Background: "#f5f5f5",
			},
			{
				Column:     "approval_status_name",
				Condition:  "eq",
				Value:      "Pending",
				Color:      "#f57c00",
				Background: "#fff3e0",
				Icon:       "clock",
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
		Actions: []string{"view", "edit", "export"},
	},
}

// Roles Management Page Config
var adminRolesPageConfig = &tablebuilder.Config{
	Title:           "Role Management",
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
			Source: "roles",
			Schema: "core",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "roles.id"},
					{Name: "name", TableColumn: "roles.name"},
					{Name: "description", TableColumn: "roles.description"},
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "name",
					Direction: "asc",
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"name": {
				Name:       "name",
				Header:     "Role Name",
				Width:      200,
				Sortable:   true,
				Filterable: true,
			},
			"description": {
				Name:       "description",
				Header:     "Description",
				Width:      400,
				Filterable: true,
			},
			"id": {
				Name:   "id",
				Header: "Actions",
				Width:  100,
				Link: &tablebuilder.LinkConfig{
					URL:   "/admin/roles/{id}",
					Label: "View",
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
		Actions: []string{"view", "edit", "delete", "export"},
	},
}

// Table Access (Permissions) Config
var adminTableAccessPageConfig = &tablebuilder.Config{
	Title:           "Table Permissions",
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
			Source: "table_access",
			Schema: "core",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "table_access.id"},
					{Name: "table_name", TableColumn: "table_access.table_name"},
					{Name: "can_create", TableColumn: "table_access.can_create"},
					{Name: "can_read", TableColumn: "table_access.can_read"},
					{Name: "can_update", TableColumn: "table_access.can_update"},
					{Name: "can_delete", TableColumn: "table_access.can_delete"},
				},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "roles",
						Schema:           "core",
						RelationshipFrom: "table_access.role_id",
						RelationshipTo:   "roles.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "role_name", TableColumn: "roles.name"},
						},
					},
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "table_name",
					Direction: "asc",
				},
			},
			Rows: 100,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"role_name": {
				Name:       "role_name",
				Header:     "Role",
				Width:      150,
				Sortable:   true,
				Filterable: true,
			},
			"table_name": {
				Name:       "table_name",
				Header:     "Table",
				Width:      200,
				Sortable:   true,
				Filterable: true,
			},
			"can_create": {
				Name:   "can_create",
				Header: "Create",
				Width:  80,
				Align:  "center",
				Format: &tablebuilder.FormatConfig{
					Type: "boolean",
				},
			},
			"can_read": {
				Name:   "can_read",
				Header: "Read",
				Width:  80,
				Align:  "center",
				Format: &tablebuilder.FormatConfig{
					Type: "boolean",
				},
			},
			"can_update": {
				Name:   "can_update",
				Header: "Update",
				Width:  80,
				Align:  "center",
				Format: &tablebuilder.FormatConfig{
					Type: "boolean",
				},
			},
			"can_delete": {
				Name:   "can_delete",
				Header: "Delete",
				Width:  80,
				Align:  "center",
				Format: &tablebuilder.FormatConfig{
					Type: "boolean",
				},
			},
		},
		ConditionalFormatting: []tablebuilder.ConditionalFormat{
			{
				Column:     "can_create",
				Condition:  "eq",
				Value:      true,
				Color:      "#2e7d32",
				Background: "#e8f5e9",
				Icon:       "check",
			},
			{
				Column:     "can_read",
				Condition:  "eq",
				Value:      true,
				Color:      "#2e7d32",
				Background: "#e8f5e9",
				Icon:       "check",
			},
			{
				Column:     "can_update",
				Condition:  "eq",
				Value:      true,
				Color:      "#2e7d32",
				Background: "#e8f5e9",
				Icon:       "check",
			},
			{
				Column:     "can_delete",
				Condition:  "eq",
				Value:      true,
				Color:      "#2e7d32",
				Background: "#e8f5e9",
				Icon:       "check",
			},
		},
		Pagination: &tablebuilder.PaginationConfig{
			Enabled:         true,
			PageSizes:       []int{25, 50, 100, 200},
			DefaultPageSize: 50,
		},
	},
	Permissions: tablebuilder.Permissions{
		Roles:   []string{"admin"},
		Actions: []string{"view", "edit"},
	},
}

// Audit Logs Page Config
var adminAuditPageConfig = &tablebuilder.Config{
	Title:           "Audit Logs",
	WidgetType:      "table",
	Visualization:   "table",
	PositionX:       0,
	PositionY:       0,
	Width:           12,
	Height:          8,
	RefreshInterval: 60,
	RefreshMode:     "polling",
	DataSource: []tablebuilder.DataSource{
		{
			Type:   "query",
			Source: "automation_executions",
			Schema: "workflow",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "automation_executions.id"},
					{Name: "entity_type", TableColumn: "automation_executions.entity_type"},
					{Name: "status", TableColumn: "automation_executions.status"},
					{Name: "error_message", TableColumn: "automation_executions.error_message"},
					{Name: "execution_time_ms", TableColumn: "automation_executions.execution_time_ms"},
					{Name: "executed_at", TableColumn: "automation_executions.executed_at"},
				},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "automation_rules",
						Schema:           "workflow",
						RelationshipFrom: "automation_executions.automation_rules_id",
						RelationshipTo:   "automation_rules.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "rule_name", TableColumn: "automation_rules.name"},
						},
					},
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "executed_at",
					Direction: "desc",
				},
			},
			Rows: 100,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"executed_at": {
				Name:     "executed_at",
				Header:   "Execution Time",
				Width:    180,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02 15:04:05",
				},
			},
			"rule_name": {
				Name:       "rule_name",
				Header:     "Rule",
				Width:      200,
				Sortable:   true,
				Filterable: true,
			},
			"entity_type": {
				Name:       "entity_type",
				Header:     "Entity Type",
				Width:      150,
				Sortable:   true,
				Filterable: true,
			},
			"status": {
				Name:         "status",
				Header:       "Status",
				Width:        120,
				Sortable:     true,
				Filterable:   true,
				CellTemplate: "status",
			},
			"execution_time_ms": {
				Name:     "execution_time_ms",
				Header:   "Duration (ms)",
				Width:    120,
				Align:    "right",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:      "number",
					Precision: 0,
				},
			},
			"error_message": {
				Name:   "error_message",
				Header: "Error",
				Width:  300,
			},
			"id": {
				Name:   "id",
				Header: "Details",
				Width:  100,
				Link: &tablebuilder.LinkConfig{
					URL:   "/admin/audit/{id}",
					Label: "View",
				},
			},
		},
		ConditionalFormatting: []tablebuilder.ConditionalFormat{
			{
				Column:     "status",
				Condition:  "eq",
				Value:      "success",
				Color:      "#2e7d32",
				Background: "#e8f5e9",
				Icon:       "check-circle",
			},
			{
				Column:     "status",
				Condition:  "eq",
				Value:      "failed",
				Color:      "#c62828",
				Background: "#ffebee",
				Icon:       "x-circle",
			},
			{
				Column:     "status",
				Condition:  "eq",
				Value:      "partial",
				Color:      "#f57c00",
				Background: "#fff3e0",
				Icon:       "alert-circle",
			},
		},
		Pagination: &tablebuilder.PaginationConfig{
			Enabled:         true,
			PageSizes:       []int{25, 50, 100, 200},
			DefaultPageSize: 50,
		},
	},
	Permissions: tablebuilder.Permissions{
		Roles:   []string{"admin"},
		Actions: []string{"view", "export"},
	},
}

// Table Configs Management Page Config
var adminConfigPageConfig = &tablebuilder.Config{
	Title:           "Table Configurations",
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
			Source: "table_configs",
			Schema: "config",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "table_configs.id"},
					{Name: "name", TableColumn: "table_configs.name"},
					{Name: "description", TableColumn: "table_configs.description"},
					{Name: "created_date", TableColumn: "table_configs.created_date"},
					{Name: "updated_date", TableColumn: "table_configs.updated_date"},
				},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "users",
						Alias:            "created_by_user",
						Schema:           "core",
						RelationshipFrom: "table_configs.created_by",
						RelationshipTo:   "created_by_user.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "username", Alias: "created_by_username", TableColumn: "created_by_user.username"},
						},
					},
					{
						Table:            "users",
						Alias:            "updated_by_user",
						Schema:           "core",
						RelationshipFrom: "table_configs.updated_by",
						RelationshipTo:   "updated_by_user.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "username", Alias: "updated_by_username", TableColumn: "updated_by_user.username"},
						},
					},
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "updated_date",
					Direction: "desc",
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"name": {
				Name:       "name",
				Header:     "Config Name",
				Width:      250,
				Sortable:   true,
				Filterable: true,
			},
			"description": {
				Name:       "description",
				Header:     "Description",
				Width:      400,
				Filterable: true,
			},
			"created_by_username": {
				Name:       "created_by_username",
				Header:     "Created By",
				Width:      150,
				Filterable: true,
			},
			"updated_by_username": {
				Name:       "updated_by_username",
				Header:     "Updated By",
				Width:      150,
				Filterable: true,
			},
			"updated_date": {
				Name:     "updated_date",
				Header:   "Last Updated",
				Width:    180,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02 15:04",
				},
			},
			"id": {
				Name:   "id",
				Header: "Actions",
				Width:  100,
				Link: &tablebuilder.LinkConfig{
					URL:   "/admin/config/{id}",
					Label: "View",
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
		Actions: []string{"view", "edit", "delete", "export"},
	},
}

// =============================================================================
// ASSETS MODULE CONFIGS
// =============================================================================

// Assets List Page Config
var assetsListPageConfig = &tablebuilder.Config{
	Title:           "Asset List",
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
			Source: "assets",
			Schema: "assets",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "assets.id"},
					{Name: "serial_number", TableColumn: "assets.serial_number"},
					{Name: "last_maintenance_time", TableColumn: "assets.last_maintenance_time"},
				},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "valid_assets",
						Schema:           "assets",
						RelationshipFrom: "assets.valid_asset_id",
						RelationshipTo:   "valid_assets.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "asset_name", TableColumn: "valid_assets.name"},
							{Name: "price", TableColumn: "valid_assets.price"},
							{Name: "model_number", TableColumn: "valid_assets.model_number"},
							{Name: "maintenance_interval", TableColumn: "valid_assets.maintenance_interval"},
						},
						ForeignTables: []tablebuilder.ForeignTable{
							{
								Table:            "asset_types",
								Schema:           "assets",
								RelationshipFrom: "valid_assets.type_id",
								RelationshipTo:   "asset_types.id",
								JoinType:         "left",
								Columns: []tablebuilder.ColumnDefinition{
									{Name: "name", Alias: "asset_type_name", TableColumn: "asset_types.name"},
								},
							},
						},
					},
					{
						Table:            "asset_conditions",
						Schema:           "assets",
						RelationshipFrom: "assets.asset_condition_id",
						RelationshipTo:   "asset_conditions.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "condition_name", TableColumn: "asset_conditions.name"},
						},
					},
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "asset_name",
					Direction: "asc",
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"asset_name": {
				Name:       "asset_name",
				Header:     "Asset",
				Width:      250,
				Sortable:   true,
				Filterable: true,
			},
			"asset_type_name": {
				Name:       "asset_type_name",
				Header:     "Type",
				Width:      150,
				Sortable:   true,
				Filterable: true,
			},
			"serial_number": {
				Name:       "serial_number",
				Header:     "Serial Number",
				Width:      180,
				Sortable:   true,
				Filterable: true,
			},
			"model_number": {
				Name:       "model_number",
				Header:     "Model",
				Width:      150,
				Filterable: true,
			},
			"condition_name": {
				Name:       "condition_name",
				Header:     "Condition",
				Width:      120,
				Sortable:   true,
				Filterable: true,
			},
			"price": {
				Name:     "price",
				Header:   "Value",
				Width:    120,
				Align:    "right",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:      "currency",
					Currency:  "USD",
					Precision: 2,
				},
			},
			"last_maintenance_time": {
				Name:     "last_maintenance_time",
				Header:   "Last Maintenance",
				Width:    150,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "2006-01-02",
				},
			},
			"id": {
				Name:   "id",
				Header: "Actions",
				Width:  100,
				Link: &tablebuilder.LinkConfig{
					URL:   "/assets/list/{id}",
					Label: "View",
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
		Roles:   []string{"admin", "asset_manager"},
		Actions: []string{"view", "export"},
	},
}

// Asset Requests Page Config
var assetsRequestsPageConfig = &tablebuilder.Config{
	Title:           "Asset Requests",
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
			Source: "user_assets",
			Schema: "assets",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "user_assets.id"},
					{Name: "date_received", TableColumn: "user_assets.date_received"},
					{Name: "last_maintenance", TableColumn: "user_assets.last_maintenance"},
				},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "users",
						Alias:            "asset_owner",
						Schema:           "core",
						RelationshipFrom: "user_assets.user_id",
						RelationshipTo:   "asset_owner.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "first_name", Alias: "user_first_name", TableColumn: "asset_owner.first_name"},
							{Name: "last_name", Alias: "user_last_name", TableColumn: "asset_owner.last_name"},
							{Name: "username", Alias: "username", TableColumn: "asset_owner.username"},
						},
					},
					{
						Table:            "users",
						Alias:            "approved_by_user",
						Schema:           "core",
						RelationshipFrom: "user_assets.approved_by",
						RelationshipTo:   "approved_by_user.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "username", Alias: "approved_by_username", TableColumn: "approved_by_user.username"},
						},
					},
					{
						Table:            "assets",
						Schema:           "assets",
						RelationshipFrom: "user_assets.asset_id",
						RelationshipTo:   "assets.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "serial_number", Alias: "asset_serial_number", TableColumn: "assets.serial_number"},
						},
						ForeignTables: []tablebuilder.ForeignTable{
							{
								Table:            "valid_assets",
								Schema:           "assets",
								RelationshipFrom: "assets.valid_asset_id",
								RelationshipTo:   "valid_assets.id",
								JoinType:         "left",
								Columns: []tablebuilder.ColumnDefinition{
									{Name: "name", Alias: "asset_name", TableColumn: "valid_assets.name"},
								},
							},
						},
					},
					{
						Table:            "approval_status",
						Schema:           "assets",
						RelationshipFrom: "user_assets.approval_status_id",
						RelationshipTo:   "approval_status.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "approval_status_name", TableColumn: "approval_status.name"},
						},
					},
					{
						Table:            "fulfillment_status",
						Schema:           "assets",
						RelationshipFrom: "user_assets.fulfillment_status_id",
						RelationshipTo:   "fulfillment_status.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "fulfillment_status_name", TableColumn: "fulfillment_status.name"},
						},
					},
				},
				ClientComputedColumns: []tablebuilder.ComputedColumn{
					{
						Name:       "user_full_name",
						Expression: "user_first_name + ' ' + user_last_name",
					},
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "date_received",
					Direction: "desc",
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"user_full_name": {
				Name:       "user_full_name",
				Header:     "Requester",
				Width:      180,
				Sortable:   true,
				Filterable: true,
			},
			"asset_name": {
				Name:       "asset_name",
				Header:     "Asset",
				Width:      200,
				Sortable:   true,
				Filterable: true,
			},
			"asset_serial_number": {
				Name:       "asset_serial_number",
				Header:     "Serial Number",
				Width:      150,
				Filterable: true,
			},
			"approval_status_name": {
				Name:         "approval_status_name",
				Header:       "Approval Status",
				Width:        150,
				Sortable:     true,
				Filterable:   true,
				CellTemplate: "status",
			},
			"fulfillment_status_name": {
				Name:         "fulfillment_status_name",
				Header:       "Fulfillment Status",
				Width:        150,
				Sortable:     true,
				Filterable:   true,
				CellTemplate: "status",
			},
			"approved_by_username": {
				Name:       "approved_by_username",
				Header:     "Approved By",
				Width:      150,
				Filterable: true,
			},
			"date_received": {
				Name:     "date_received",
				Header:   "Date Received",
				Width:    120,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "2006-01-02",
				},
			},
			"id": {
				Name:   "id",
				Header: "Actions",
				Width:  100,
				Link: &tablebuilder.LinkConfig{
					URL:   "/assets/requests/{id}",
					Label: "View",
				},
			},
		},
		ConditionalFormatting: []tablebuilder.ConditionalFormat{
			{
				Column:     "approval_status_name",
				Condition:  "eq",
				Value:      "Pending",
				Color:      "#f57c00",
				Background: "#fff3e0",
				Icon:       "clock",
			},
			{
				Column:     "approval_status_name",
				Condition:  "eq",
				Value:      "Approved",
				Color:      "#2e7d32",
				Background: "#e8f5e9",
				Icon:       "check-circle",
			},
			{
				Column:     "fulfillment_status_name",
				Condition:  "eq",
				Value:      "Pending",
				Color:      "#f57c00",
				Background: "#fff3e0",
				Icon:       "clock",
			},
			{
				Column:     "fulfillment_status_name",
				Condition:  "eq",
				Value:      "Delivered",
				Color:      "#2e7d32",
				Background: "#e8f5e9",
				Icon:       "check-circle",
			},
		},
		Pagination: &tablebuilder.PaginationConfig{
			Enabled:         true,
			PageSizes:       []int{10, 25, 50, 100},
			DefaultPageSize: 25,
		},
	},
	Permissions: tablebuilder.Permissions{
		Roles:   []string{"admin", "asset_manager", "user"},
		Actions: []string{"view", "export"},
	},
}

// =============================================================================
// HR MODULE CONFIGS
// =============================================================================

// Employees Page Config
var hrEmployeesPageConfig = &tablebuilder.Config{
	Title:           "Employee Directory",
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
			Source: "users",
			Schema: "core",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "users.id"},
					{Name: "first_name", TableColumn: "users.first_name"},
					{Name: "last_name", TableColumn: "users.last_name"},
					{Name: "email", TableColumn: "users.email"},
					{Name: "enabled", TableColumn: "users.enabled"},
					{Name: "date_hired", TableColumn: "users.date_hired"},
				},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "titles",
						Schema:           "hr",
						RelationshipFrom: "users.title_id",
						RelationshipTo:   "titles.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "title_name", TableColumn: "titles.name"},
						},
					},
					{
						Table:            "offices",
						Schema:           "hr",
						RelationshipFrom: "users.office_id",
						RelationshipTo:   "offices.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "office_name", TableColumn: "offices.name"},
						},
						ForeignTables: []tablebuilder.ForeignTable{
							{
								Table:            "streets",
								Schema:           "geography",
								RelationshipFrom: "offices.street_id",
								RelationshipTo:   "streets.id",
								JoinType:         "left",
								Columns: []tablebuilder.ColumnDefinition{
									{Name: "line_1", Alias: "office_street", TableColumn: "streets.line_1"},
								},
								ForeignTables: []tablebuilder.ForeignTable{
									{
										Table:            "cities",
										Schema:           "geography",
										RelationshipFrom: "streets.city_id",
										RelationshipTo:   "cities.id",
										JoinType:         "left",
										Columns: []tablebuilder.ColumnDefinition{
											{Name: "name", Alias: "office_city", TableColumn: "cities.name"},
										},
										ForeignTables: []tablebuilder.ForeignTable{
											{
												Table:            "regions",
												Schema:           "geography",
												RelationshipFrom: "cities.region_id",
												RelationshipTo:   "regions.id",
												JoinType:         "left",
												Columns: []tablebuilder.ColumnDefinition{
													{Name: "name", Alias: "office_region", TableColumn: "regions.name"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				ClientComputedColumns: []tablebuilder.ComputedColumn{
					{
						Name:       "full_name",
						Expression: "first_name + ' ' + last_name",
					},
					{
						Name:       "office_location",
						Expression: "(office_city || '') + (office_region ? ', ' + office_region : '')",
					},
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "last_name",
					Direction: "asc",
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"full_name": {
				Name:       "full_name",
				Header:     "Name",
				Width:      200,
				Sortable:   true,
				Filterable: true,
			},
			"email": {
				Name:       "email",
				Header:     "Email",
				Width:      250,
				Filterable: true,
			},
			"title_name": {
				Name:       "title_name",
				Header:     "Title",
				Width:      180,
				Sortable:   true,
				Filterable: true,
			},
			"office_name": {
				Name:       "office_name",
				Header:     "Office",
				Width:      150,
				Sortable:   true,
				Filterable: true,
			},
			"office_location": {
				Name:       "office_location",
				Header:     "Location",
				Width:      200,
				Filterable: true,
			},
			"date_hired": {
				Name:     "date_hired",
				Header:   "Date Hired",
				Width:    120,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "2006-01-02",
				},
			},
			"enabled": {
				Name:   "enabled",
				Header: "Active",
				Width:  80,
				Align:  "center",
				Format: &tablebuilder.FormatConfig{
					Type: "boolean",
				},
			},
			"id": {
				Name:   "id",
				Header: "Actions",
				Width:  100,
				Link: &tablebuilder.LinkConfig{
					URL:   "/hr/employees/{id}",
					Label: "View",
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
		Roles:   []string{"admin", "hr"},
		Actions: []string{"view", "export"},
	},
}

// Offices Page Config
var hrOfficesPageConfig = &tablebuilder.Config{
	Title:           "Office Management",
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
			Source: "offices",
			Schema: "hr",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "offices.id"},
					{Name: "name", TableColumn: "offices.name"},
				},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "streets",
						Schema:           "geography",
						RelationshipFrom: "offices.street_id",
						RelationshipTo:   "streets.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "line_1", Alias: "street_line_1", TableColumn: "streets.line_1"},
							{Name: "line_2", Alias: "street_line_2", TableColumn: "streets.line_2"},
							{Name: "postal_code", TableColumn: "streets.postal_code"},
						},
						ForeignTables: []tablebuilder.ForeignTable{
							{
								Table:            "cities",
								Schema:           "geography",
								RelationshipFrom: "streets.city_id",
								RelationshipTo:   "cities.id",
								JoinType:         "left",
								Columns: []tablebuilder.ColumnDefinition{
									{Name: "name", Alias: "city_name", TableColumn: "cities.name"},
								},
								ForeignTables: []tablebuilder.ForeignTable{
									{
										Table:            "regions",
										Schema:           "geography",
										RelationshipFrom: "cities.region_id",
										RelationshipTo:   "regions.id",
										JoinType:         "left",
										Columns: []tablebuilder.ColumnDefinition{
											{Name: "name", Alias: "region_name", TableColumn: "regions.name"},
											{Name: "code", Alias: "region_code", TableColumn: "regions.code"},
										},
										ForeignTables: []tablebuilder.ForeignTable{
											{
												Table:            "countries",
												Schema:           "geography",
												RelationshipFrom: "regions.country_id",
												RelationshipTo:   "countries.id",
												JoinType:         "left",
												Columns: []tablebuilder.ColumnDefinition{
													{Name: "name", Alias: "country_name", TableColumn: "countries.name"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				ClientComputedColumns: []tablebuilder.ComputedColumn{
					{
						Name:       "full_address",
						Expression: "street_line_1 + (street_line_2 ? ', ' + street_line_2 : '') + ', ' + city_name + ', ' + region_code + ' ' + postal_code",
					},
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "name",
					Direction: "asc",
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"name": {
				Name:       "name",
				Header:     "Office Name",
				Width:      200,
				Sortable:   true,
				Filterable: true,
			},
			"full_address": {
				Name:       "full_address",
				Header:     "Address",
				Width:      400,
				Filterable: true,
			},
			"city_name": {
				Name:       "city_name",
				Header:     "City",
				Width:      150,
				Sortable:   true,
				Filterable: true,
			},
			"region_name": {
				Name:       "region_name",
				Header:     "State/Region",
				Width:      150,
				Sortable:   true,
				Filterable: true,
			},
			"country_name": {
				Name:       "country_name",
				Header:     "Country",
				Width:      120,
				Filterable: true,
			},
			"id": {
				Name:   "id",
				Header: "Actions",
				Width:  100,
				Link: &tablebuilder.LinkConfig{
					URL:   "/hr/offices/{id}",
					Label: "View",
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
		Roles:   []string{"admin", "hr"},
		Actions: []string{"view", "edit", "export"},
	},
}

// =============================================================================
// INVENTORY MODULE CONFIGS
// =============================================================================

// Warehouses Page Config
var inventoryWarehousesPageConfig = &tablebuilder.Config{
	Title:           "Warehouse Management",
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
			Source: "warehouses",
			Schema: "inventory",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "warehouses.id"},
					{Name: "name", TableColumn: "warehouses.name"},
					{Name: "is_active", TableColumn: "warehouses.is_active"},
					{Name: "created_date", TableColumn: "warehouses.created_date"},
					{Name: "updated_date", TableColumn: "warehouses.updated_date"},
				},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "streets",
						Schema:           "geography",
						RelationshipFrom: "warehouses.street_id",
						RelationshipTo:   "streets.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "line_1", Alias: "street_line_1", TableColumn: "streets.line_1"},
							{Name: "postal_code", TableColumn: "streets.postal_code"},
						},
						ForeignTables: []tablebuilder.ForeignTable{
							{
								Table:            "cities",
								Schema:           "geography",
								RelationshipFrom: "streets.city_id",
								RelationshipTo:   "cities.id",
								JoinType:         "left",
								Columns: []tablebuilder.ColumnDefinition{
									{Name: "name", Alias: "city_name", TableColumn: "cities.name"},
								},
								ForeignTables: []tablebuilder.ForeignTable{
									{
										Table:            "regions",
										Schema:           "geography",
										RelationshipFrom: "cities.region_id",
										RelationshipTo:   "regions.id",
										JoinType:         "left",
										Columns: []tablebuilder.ColumnDefinition{
											{Name: "code", Alias: "region_code", TableColumn: "regions.code"},
										},
									},
								},
							},
						},
					},
					{
						Table:            "users",
						Schema:           "core",
						RelationshipFrom: "warehouses.created_by",
						RelationshipTo:   "users.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "username", Alias: "created_by_username", TableColumn: "users.username"},
						},
					},
				},
				ClientComputedColumns: []tablebuilder.ComputedColumn{
					{
						Name:       "location",
						Expression: "city_name + ', ' + region_code",
					},
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "name",
					Direction: "asc",
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"name": {
				Name:       "name",
				Header:     "Warehouse Name",
				Width:      200,
				Sortable:   true,
				Filterable: true,
			},
			"location": {
				Name:       "location",
				Header:     "Location",
				Width:      200,
				Filterable: true,
			},
			"street_line_1": {
				Name:       "street_line_1",
				Header:     "Address",
				Width:      250,
				Filterable: true,
			},
			"is_active": {
				Name:   "is_active",
				Header: "Active",
				Width:  80,
				Align:  "center",
				Format: &tablebuilder.FormatConfig{
					Type: "boolean",
				},
			},
			"created_by_username": {
				Name:       "created_by_username",
				Header:     "Created By",
				Width:      150,
				Filterable: true,
			},
			"created_date": {
				Name:     "created_date",
				Header:   "Created",
				Width:    150,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02 15:04",
				},
			},
			"id": {
				Name:   "id",
				Header: "Actions",
				Width:  100,
				Link: &tablebuilder.LinkConfig{
					URL:   "/inventory/warehouses/{id}",
					Label: "View",
				},
			},
		},
		ConditionalFormatting: []tablebuilder.ConditionalFormat{
			{
				Column:     "is_active",
				Condition:  "eq",
				Value:      false,
				Color:      "#757575",
				Background: "#f5f5f5",
			},
		},
		Pagination: &tablebuilder.PaginationConfig{
			Enabled:         true,
			PageSizes:       []int{10, 25, 50, 100},
			DefaultPageSize: 25,
		},
	},
	Permissions: tablebuilder.Permissions{
		Roles:   []string{"admin", "inventory_manager"},
		Actions: []string{"view", "edit", "export"},
	},
}

// Inventory Items Page Config (reusing ComplexConfig variable)
var inventoryItemsPageConfig = ComplexConfig

// Inventory Adjustments Page Config
var inventoryAdjustmentsPageConfig = &tablebuilder.Config{
	Title:           "Stock Adjustments",
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
			Source: "inventory_adjustments",
			Schema: "inventory",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "inventory_adjustments.id"},
					{Name: "quantity_change", TableColumn: "inventory_adjustments.quantity_change"},
					{Name: "reason_code", TableColumn: "inventory_adjustments.reason_code"},
					{Name: "notes", TableColumn: "inventory_adjustments.notes"},
					{Name: "adjustment_date", TableColumn: "inventory_adjustments.adjustment_date"},
					{Name: "created_date", TableColumn: "inventory_adjustments.created_date"},
				},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "products",
						Schema:           "products",
						RelationshipFrom: "inventory_adjustments.product_id",
						RelationshipTo:   "products.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "product_name", TableColumn: "products.name"},
							{Name: "sku", Alias: "product_sku", TableColumn: "products.sku"},
						},
					},
					{
						Table:            "inventory_locations",
						Schema:           "inventory",
						RelationshipFrom: "inventory_adjustments.location_id",
						RelationshipTo:   "inventory_locations.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "aisle", TableColumn: "inventory_locations.aisle"},
							{Name: "rack", TableColumn: "inventory_locations.rack"},
							{Name: "shelf", TableColumn: "inventory_locations.shelf"},
							{Name: "bin", TableColumn: "inventory_locations.bin"},
						},
						ForeignTables: []tablebuilder.ForeignTable{
							{
								Table:            "warehouses",
								Schema:           "inventory",
								RelationshipFrom: "inventory_locations.warehouse_id",
								RelationshipTo:   "warehouses.id",
								JoinType:         "left",
								Columns: []tablebuilder.ColumnDefinition{
									{Name: "name", Alias: "warehouse_name", TableColumn: "warehouses.name"},
								},
							},
						},
					},
					{
						Table:            "users",
						Alias:            "adjusted_by_user",
						Schema:           "core",
						RelationshipFrom: "inventory_adjustments.adjusted_by",
						RelationshipTo:   "adjusted_by_user.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "username", Alias: "adjusted_by_username", TableColumn: "adjusted_by_user.username"},
						},
					},
					{
						Table:            "users",
						Alias:            "approved_by_user",
						Schema:           "core",
						RelationshipFrom: "inventory_adjustments.approved_by",
						RelationshipTo:   "approved_by_user.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "username", Alias: "approved_by_username", TableColumn: "approved_by_user.username"},
						},
					},
				},
				ClientComputedColumns: []tablebuilder.ComputedColumn{
					{
						Name:       "location_code",
						Expression: "aisle + '-' + rack + '-' + shelf + '-' + bin",
					},
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "adjustment_date",
					Direction: "desc",
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
				Width:      200,
				Sortable:   true,
				Filterable: true,
			},
			"product_sku": {
				Name:       "product_sku",
				Header:     "SKU",
				Width:      120,
				Filterable: true,
			},
			"warehouse_name": {
				Name:       "warehouse_name",
				Header:     "Warehouse",
				Width:      150,
				Sortable:   true,
				Filterable: true,
			},
			"location_code": {
				Name:       "location_code",
				Header:     "Location",
				Width:      150,
				Filterable: true,
			},
			"quantity_change": {
				Name:     "quantity_change",
				Header:   "Qty Change",
				Width:    100,
				Align:    "right",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:      "number",
					Precision: 0,
				},
			},
			"reason_code": {
				Name:       "reason_code",
				Header:     "Reason",
				Width:      120,
				Filterable: true,
			},
			"adjusted_by_username": {
				Name:       "adjusted_by_username",
				Header:     "Adjusted By",
				Width:      130,
				Filterable: true,
			},
			"approved_by_username": {
				Name:       "approved_by_username",
				Header:     "Approved By",
				Width:      130,
				Filterable: true,
			},
			"adjustment_date": {
				Name:     "adjustment_date",
				Header:   "Date",
				Width:    150,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02 15:04",
				},
			},
			"id": {
				Name:   "id",
				Header: "Actions",
				Width:  100,
				Link: &tablebuilder.LinkConfig{
					URL:   "/inventory/adjustments/{id}",
					Label: "View",
				},
			},
		},
		ConditionalFormatting: []tablebuilder.ConditionalFormat{
			{
				Column:     "quantity_change",
				Condition:  "lt",
				Value:      0,
				Color:      "#c62828",
				Background: "#ffebee",
				Icon:       "trending-down",
			},
			{
				Column:     "quantity_change",
				Condition:  "gt",
				Value:      0,
				Color:      "#2e7d32",
				Background: "#e8f5e9",
				Icon:       "trending-up",
			},
		},
		Pagination: &tablebuilder.PaginationConfig{
			Enabled:         true,
			PageSizes:       []int{10, 25, 50, 100},
			DefaultPageSize: 25,
		},
	},
	Permissions: tablebuilder.Permissions{
		Roles:   []string{"admin", "inventory_manager"},
		Actions: []string{"view", "export"},
	},
}

// Inventory Transfers Page Config
var inventoryTransfersPageConfig = &tablebuilder.Config{
	Title:           "Transfer Orders",
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
			Source: "transfer_orders",
			Schema: "inventory",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "transfer_orders.id"},
					{Name: "quantity", TableColumn: "transfer_orders.quantity"},
					{Name: "status", TableColumn: "transfer_orders.status"},
					{Name: "transfer_date", TableColumn: "transfer_orders.transfer_date"},
					{Name: "created_date", TableColumn: "transfer_orders.created_date"},
				},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "products",
						Schema:           "products",
						RelationshipFrom: "transfer_orders.product_id",
						RelationshipTo:   "products.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "product_name", TableColumn: "products.name"},
							{Name: "sku", Alias: "product_sku", TableColumn: "products.sku"},
						},
					},
					{
						Table:            "inventory_locations",
						Alias:            "from_location",
						Schema:           "inventory",
						RelationshipFrom: "transfer_orders.from_location_id",
						RelationshipTo:   "from_location.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "aisle", Alias: "from_aisle", TableColumn: "from_location.aisle"},
							{Name: "rack", Alias: "from_rack", TableColumn: "from_location.rack"},
						},
						ForeignTables: []tablebuilder.ForeignTable{
							{
								Table:            "warehouses",
								Alias:            "from_warehouse",
								Schema:           "inventory",
								RelationshipFrom: "from_location.warehouse_id",
								RelationshipTo:   "from_warehouse.id",
								JoinType:         "left",
								Columns: []tablebuilder.ColumnDefinition{
									{Name: "name", Alias: "from_warehouse_name", TableColumn: "from_warehouse.name"},
								},
							},
						},
					},
					{
						Table:            "inventory_locations",
						Alias:            "to_location",
						Schema:           "inventory",
						RelationshipFrom: "transfer_orders.to_location_id",
						RelationshipTo:   "to_location.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "aisle", Alias: "to_aisle", TableColumn: "to_location.aisle"},
							{Name: "rack", Alias: "to_rack", TableColumn: "to_location.rack"},
						},
						ForeignTables: []tablebuilder.ForeignTable{
							{
								Table:            "warehouses",
								Alias:            "to_warehouse",
								Schema:           "inventory",
								RelationshipFrom: "to_location.warehouse_id",
								RelationshipTo:   "to_warehouse.id",
								JoinType:         "left",
								Columns: []tablebuilder.ColumnDefinition{
									{Name: "name", Alias: "to_warehouse_name", TableColumn: "to_warehouse.name"},
								},
							},
						},
					},
					{
						Table:            "users",
						Alias:            "requested_by_user",
						Schema:           "core",
						RelationshipFrom: "transfer_orders.requested_by",
						RelationshipTo:   "requested_by_user.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "username", Alias: "requested_by_username", TableColumn: "requested_by_user.username"},
						},
					},
					{
						Table:            "users",
						Alias:            "approved_by_user",
						Schema:           "core",
						RelationshipFrom: "transfer_orders.approved_by",
						RelationshipTo:   "approved_by_user.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "username", Alias: "approved_by_username", TableColumn: "approved_by_user.username"},
						},
					},
				},
				ClientComputedColumns: []tablebuilder.ComputedColumn{
					{
						Name:       "from_location",
						Expression: "from_warehouse_name + ' (' + from_aisle + '-' + from_rack + ')'",
					},
					{
						Name:       "to_location",
						Expression: "to_warehouse_name + ' (' + to_aisle + '-' + to_rack + ')'",
					},
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "transfer_date",
					Direction: "desc",
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
				Width:      200,
				Sortable:   true,
				Filterable: true,
			},
			"product_sku": {
				Name:       "product_sku",
				Header:     "SKU",
				Width:      120,
				Filterable: true,
			},
			"from_location": {
				Name:       "from_location",
				Header:     "From",
				Width:      180,
				Filterable: true,
			},
			"to_location": {
				Name:       "to_location",
				Header:     "To",
				Width:      180,
				Filterable: true,
			},
			"quantity": {
				Name:     "quantity",
				Header:   "Quantity",
				Width:    90,
				Align:    "right",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:      "number",
					Precision: 0,
				},
			},
			"status": {
				Name:         "status",
				Header:       "Status",
				Width:        120,
				Sortable:     true,
				Filterable:   true,
				CellTemplate: "status",
			},
			"requested_by_username": {
				Name:       "requested_by_username",
				Header:     "Requested By",
				Width:      130,
				Filterable: true,
			},
			"approved_by_username": {
				Name:       "approved_by_username",
				Header:     "Approved By",
				Width:      130,
				Filterable: true,
			},
			"transfer_date": {
				Name:     "transfer_date",
				Header:   "Transfer Date",
				Width:    150,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02 15:04",
				},
			},
			"id": {
				Name:   "id",
				Header: "Actions",
				Width:  100,
				Link: &tablebuilder.LinkConfig{
					URL:   "/inventory/transfers/{id}",
					Label: "View",
				},
			},
		},
		ConditionalFormatting: []tablebuilder.ConditionalFormat{
			{
				Column:     "status",
				Condition:  "eq",
				Value:      "pending",
				Color:      "#f57c00",
				Background: "#fff3e0",
				Icon:       "clock",
			},
			{
				Column:     "status",
				Condition:  "eq",
				Value:      "completed",
				Color:      "#2e7d32",
				Background: "#e8f5e9",
				Icon:       "check-circle",
			},
			{
				Column:     "status",
				Condition:  "eq",
				Value:      "cancelled",
				Color:      "#c62828",
				Background: "#ffebee",
				Icon:       "x-circle",
			},
		},
		Pagination: &tablebuilder.PaginationConfig{
			Enabled:         true,
			PageSizes:       []int{10, 25, 50, 100},
			DefaultPageSize: 25,
		},
	},
	Permissions: tablebuilder.Permissions{
		Roles:   []string{"admin", "inventory_manager"},
		Actions: []string{"view", "export"},
	},
}

// =============================================================================
// SALES MODULE CONFIGS
// =============================================================================

// Customers Page Config
var salesCustomersPageConfig = &tablebuilder.Config{
	Title:           "Customer Management",
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
			Source: "customers",
			Schema: "sales",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "customers.id"},
					{Name: "name", TableColumn: "customers.name"},
					{Name: "notes", TableColumn: "customers.notes"},
					{Name: "created_date", TableColumn: "customers.created_date"},
					{Name: "updated_date", TableColumn: "customers.updated_date"},
				},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "contact_infos",
						Schema:           "core",
						RelationshipFrom: "customers.contact_id",
						RelationshipTo:   "contact_infos.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "first_name", Alias: "contact_first_name", TableColumn: "contact_infos.first_name"},
							{Name: "last_name", Alias: "contact_last_name", TableColumn: "contact_infos.last_name"},
							{Name: "email_address", Alias: "contact_email", TableColumn: "contact_infos.email_address"},
							{Name: "primary_phone_number", Alias: "contact_phone", TableColumn: "contact_infos.primary_phone_number"},
						},
					},
					{
						Table:            "streets",
						Schema:           "geography",
						RelationshipFrom: "customers.delivery_address_id",
						RelationshipTo:   "streets.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "line_1", Alias: "delivery_street", TableColumn: "streets.line_1"},
						},
						ForeignTables: []tablebuilder.ForeignTable{
							{
								Table:            "cities",
								Schema:           "geography",
								RelationshipFrom: "streets.city_id",
								RelationshipTo:   "cities.id",
								JoinType:         "left",
								Columns: []tablebuilder.ColumnDefinition{
									{Name: "name", Alias: "delivery_city", TableColumn: "cities.name"},
								},
								ForeignTables: []tablebuilder.ForeignTable{
									{
										Table:            "regions",
										Schema:           "geography",
										RelationshipFrom: "cities.region_id",
										RelationshipTo:   "regions.id",
										JoinType:         "left",
										Columns: []tablebuilder.ColumnDefinition{
											{Name: "code", Alias: "delivery_region_code", TableColumn: "regions.code"},
										},
									},
								},
							},
						},
					},
					{
						Table:            "users",
						Schema:           "core",
						RelationshipFrom: "customers.created_by",
						RelationshipTo:   "users.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "username", Alias: "created_by_username", TableColumn: "users.username"},
						},
					},
				},
				ClientComputedColumns: []tablebuilder.ComputedColumn{
					{
						Name:       "contact_full_name",
						Expression: "contact_first_name + ' ' + contact_last_name",
					},
					{
						Name:       "delivery_location",
						Expression: "delivery_city + ', ' + delivery_region_code",
					},
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "name",
					Direction: "asc",
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"name": {
				Name:       "name",
				Header:     "Customer Name",
				Width:      200,
				Sortable:   true,
				Filterable: true,
			},
			"contact_full_name": {
				Name:       "contact_full_name",
				Header:     "Contact Person",
				Width:      180,
				Filterable: true,
			},
			"contact_email": {
				Name:       "contact_email",
				Header:     "Email",
				Width:      200,
				Filterable: true,
			},
			"contact_phone": {
				Name:       "contact_phone",
				Header:     "Phone",
				Width:      150,
				Filterable: true,
			},
			"delivery_location": {
				Name:       "delivery_location",
				Header:     "Location",
				Width:      180,
				Filterable: true,
			},
			"created_by_username": {
				Name:       "created_by_username",
				Header:     "Created By",
				Width:      130,
				Filterable: true,
			},
			"created_date": {
				Name:     "created_date",
				Header:   "Created",
				Width:    150,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02 15:04",
				},
			},
			"id": {
				Name:   "id",
				Header: "Actions",
				Width:  100,
				Link: &tablebuilder.LinkConfig{
					URL:   "/sales/customers/{id}",
					Label: "View",
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
		Roles:   []string{"admin", "sales"},
		Actions: []string{"view", "edit", "export"},
	},
}

// =============================================================================
// PROCUREMENT CONFIGS
// =============================================================================

var purchaseOrderPageConfig = &tablebuilder.Config{
	Title:           "Purchase Orders",
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
			Source: "purchase_orders",
			Schema: "procurement",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "purchase_orders.id"},
					{Name: "order_number", TableColumn: "purchase_orders.order_number"},
					{Name: "supplier_id", TableColumn: "purchase_orders.supplier_id"},
					{Name: "purchase_order_status_id", TableColumn: "purchase_orders.purchase_order_status_id"},
					{Name: "delivery_warehouse_id", TableColumn: "purchase_orders.delivery_warehouse_id"},
					{Name: "order_date", TableColumn: "purchase_orders.order_date"},
					{Name: "expected_delivery_date", TableColumn: "purchase_orders.expected_delivery_date"},
					{Name: "actual_delivery_date", TableColumn: "purchase_orders.actual_delivery_date"},
					{Name: "subtotal", TableColumn: "purchase_orders.subtotal"},
					{Name: "tax_amount", TableColumn: "purchase_orders.tax_amount"},
					{Name: "shipping_cost", TableColumn: "purchase_orders.shipping_cost"},
					{Name: "total_amount", TableColumn: "purchase_orders.total_amount"},
					{Name: "currency", TableColumn: "purchase_orders.currency"},
					{Name: "requested_by", TableColumn: "purchase_orders.requested_by"},
					{Name: "approved_by", TableColumn: "purchase_orders.approved_by"},
					{Name: "approved_date", TableColumn: "purchase_orders.approved_date"},
					{Name: "notes", TableColumn: "purchase_orders.notes"},
					{Name: "created_date", TableColumn: "purchase_orders.created_date"},
					{Name: "updated_date", TableColumn: "purchase_orders.updated_date"},
				},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "suppliers",
						Schema:           "procurement",
						RelationshipFrom: "purchase_orders.supplier_id",
						RelationshipTo:   "suppliers.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "supplier_name", TableColumn: "suppliers.name"},
							{Name: "payment_terms", Alias: "supplier_payment_terms", TableColumn: "suppliers.payment_terms"},
							{Name: "lead_time_days", Alias: "supplier_lead_time_days", TableColumn: "suppliers.lead_time_days"},
						},
					},
					{
						Table:            "purchase_order_statuses",
						Schema:           "procurement",
						RelationshipFrom: "purchase_orders.purchase_order_status_id",
						RelationshipTo:   "purchase_order_statuses.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "status_name", TableColumn: "purchase_order_statuses.name"},
							{Name: "description", Alias: "status_description", TableColumn: "purchase_order_statuses.description"},
						},
					},
					{
						Table:            "warehouses",
						Schema:           "inventory",
						RelationshipFrom: "purchase_orders.delivery_warehouse_id",
						RelationshipTo:   "warehouses.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "warehouse_name", TableColumn: "warehouses.name"},
							{Name: "code", Alias: "warehouse_code", TableColumn: "warehouses.code"},
						},
					},
					{
						Table:            "users",
						Schema:           "core",
						RelationshipFrom: "purchase_orders.requested_by",
						RelationshipTo:   "users.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "username", Alias: "requested_by_username", TableColumn: "users.username"},
							{Name: "first_name", Alias: "requested_by_first_name", TableColumn: "users.first_name"},
							{Name: "last_name", Alias: "requested_by_last_name", TableColumn: "users.last_name"},
						},
					},
					{
						Table:            "users",
						Schema:           "core",
						RelationshipFrom: "purchase_orders.approved_by",
						RelationshipTo:   "users.id",
						JoinType:         "left",
						Alias:            "approver",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "username", Alias: "approved_by_username", TableColumn: "approver.username"},
							{Name: "first_name", Alias: "approved_by_first_name", TableColumn: "approver.first_name"},
							{Name: "last_name", Alias: "approved_by_last_name", TableColumn: "approver.last_name"},
						},
					},
				},
				ClientComputedColumns: []tablebuilder.ComputedColumn{
					{
						Name:       "requested_by_full_name",
						Expression: "requested_by_first_name + ' ' + requested_by_last_name",
					},
					{
						Name:       "approved_by_full_name",
						Expression: "approved_by_first_name ? (approved_by_first_name + ' ' + approved_by_last_name) : 'N/A'",
					},
					{
						Name:       "formatted_total",
						Expression: "currency + ' ' + total_amount.toFixed(2)",
					},
					{
						Name:       "delivery_status",
						Expression: "actual_delivery_date ? 'delivered' : (new Date(expected_delivery_date) < new Date() ? 'overdue' : 'pending')",
					},
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "order_date",
					Direction: "desc",
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"order_number": {
				Name:       "order_number",
				Header:     "PO Number",
				Width:      150,
				Sortable:   true,
				Filterable: true,
			},
			"supplier_name": {
				Name:       "supplier_name",
				Header:     "Supplier",
				Width:      200,
				Filterable: true,
			},
			"status_name": {
				Name:       "status_name",
				Header:     "Status",
				Width:      120,
				Filterable: true,
			},
			"warehouse_name": {
				Name:       "warehouse_name",
				Header:     "Delivery Warehouse",
				Width:      180,
				Filterable: true,
			},
			"order_date": {
				Name:     "order_date",
				Header:   "Order Date",
				Width:    130,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02",
				},
			},
			"expected_delivery_date": {
				Name:     "expected_delivery_date",
				Header:   "Expected Delivery",
				Width:    150,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02",
				},
			},
			"formatted_total": {
				Name:   "formatted_total",
				Header: "Total Amount",
				Width:  130,
			},
			"delivery_status": {
				Name:       "delivery_status",
				Header:     "Delivery Status",
				Width:      130,
				Filterable: true,
			},
			"requested_by_full_name": {
				Name:       "requested_by_full_name",
				Header:     "Requested By",
				Width:      150,
				Filterable: true,
			},
			"id": {
				Name:   "id",
				Header: "Actions",
				Width:  100,
				Link: &tablebuilder.LinkConfig{
					URL:   "/procurement/purchase-orders/{id}",
					Label: "View",
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
		Roles:   []string{"admin", "procurement"},
		Actions: []string{"view", "create", "edit", "export"},
	},
}

var purchaseOrderLineItemPageConfig = &tablebuilder.Config{
	Title:           "Purchase Order Line Items",
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
			Source: "purchase_order_line_items",
			Schema: "procurement",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "purchase_order_line_items.id"},
					{Name: "purchase_order_id", TableColumn: "purchase_order_line_items.purchase_order_id"},
					{Name: "supplier_product_id", TableColumn: "purchase_order_line_items.supplier_product_id"},
					{Name: "quantity_ordered", TableColumn: "purchase_order_line_items.quantity_ordered"},
					{Name: "quantity_received", TableColumn: "purchase_order_line_items.quantity_received"},
					{Name: "quantity_cancelled", TableColumn: "purchase_order_line_items.quantity_cancelled"},
					{Name: "unit_cost", TableColumn: "purchase_order_line_items.unit_cost"},
					{Name: "discount", TableColumn: "purchase_order_line_items.discount"},
					{Name: "line_total", TableColumn: "purchase_order_line_items.line_total"},
					{Name: "line_item_status_id", TableColumn: "purchase_order_line_items.line_item_status_id"},
					{Name: "expected_delivery_date", TableColumn: "purchase_order_line_items.expected_delivery_date"},
					{Name: "actual_delivery_date", TableColumn: "purchase_order_line_items.actual_delivery_date"},
					{Name: "notes", TableColumn: "purchase_order_line_items.notes"},
					{Name: "created_by", TableColumn: "purchase_order_line_items.created_by"},
					{Name: "created_date", TableColumn: "purchase_order_line_items.created_date"},
					{Name: "updated_by", TableColumn: "purchase_order_line_items.updated_by"},
					{Name: "updated_date", TableColumn: "purchase_order_line_items.updated_date"},
				},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "purchase_orders",
						Schema:           "procurement",
						RelationshipFrom: "purchase_order_line_items.purchase_order_id",
						RelationshipTo:   "purchase_orders.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "order_number", Alias: "po_number", TableColumn: "purchase_orders.order_number"},
							{Name: "supplier_id", Alias: "po_supplier_id", TableColumn: "purchase_orders.supplier_id"},
							{Name: "order_date", Alias: "po_order_date", TableColumn: "purchase_orders.order_date"},
						},
						ForeignTables: []tablebuilder.ForeignTable{
							{
								Table:            "suppliers",
								Schema:           "procurement",
								RelationshipFrom: "purchase_orders.supplier_id",
								RelationshipTo:   "suppliers.id",
								JoinType:         "left",
								Columns: []tablebuilder.ColumnDefinition{
									{Name: "name", Alias: "supplier_name", TableColumn: "suppliers.name"},
								},
							},
						},
					},
					{
						Table:            "supplier_products",
						Schema:           "procurement",
						RelationshipFrom: "purchase_order_line_items.supplier_product_id",
						RelationshipTo:   "supplier_products.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "supplier_part_number", Alias: "supplier_part_number", TableColumn: "supplier_products.supplier_part_number"},
							{Name: "product_id", Alias: "product_id", TableColumn: "supplier_products.product_id"},
						},
						ForeignTables: []tablebuilder.ForeignTable{
							{
								Table:            "products",
								Schema:           "products",
								RelationshipFrom: "supplier_products.product_id",
								RelationshipTo:   "products.id",
								JoinType:         "left",
								Columns: []tablebuilder.ColumnDefinition{
									{Name: "name", Alias: "product_name", TableColumn: "products.name"},
									{Name: "sku", Alias: "product_sku", TableColumn: "products.sku"},
								},
							},
						},
					},
					{
						Table:            "purchase_order_line_item_statuses",
						Schema:           "procurement",
						RelationshipFrom: "purchase_order_line_items.line_item_status_id",
						RelationshipTo:   "purchase_order_line_item_statuses.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "line_item_status_name", TableColumn: "purchase_order_line_item_statuses.name"},
						},
					},
					{
						Table:            "users",
						Schema:           "core",
						RelationshipFrom: "purchase_order_line_items.created_by",
						RelationshipTo:   "users.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "username", Alias: "created_by_username", TableColumn: "users.username"},
						},
					},
				},
				ClientComputedColumns: []tablebuilder.ComputedColumn{
					{
						Name:       "quantity_pending",
						Expression: "quantity_ordered - quantity_received - quantity_cancelled",
					},
					{
						Name:       "fulfillment_percentage",
						Expression: "quantity_ordered > 0 ? ((quantity_received / quantity_ordered) * 100).toFixed(1) : '0.0'",
					},
					{
						Name:       "line_status",
						Expression: "quantity_received >= quantity_ordered ? 'complete' : quantity_cancelled > 0 ? 'partial' : 'pending'",
					},
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "created_date",
					Direction: "desc",
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"po_number": {
				Name:       "po_number",
				Header:     "PO Number",
				Width:      150,
				Filterable: true,
			},
			"supplier_name": {
				Name:       "supplier_name",
				Header:     "Supplier",
				Width:      180,
				Filterable: true,
			},
			"product_name": {
				Name:       "product_name",
				Header:     "Product",
				Width:      200,
				Filterable: true,
			},
			"product_sku": {
				Name:       "product_sku",
				Header:     "SKU",
				Width:      120,
				Filterable: true,
			},
			"supplier_part_number": {
				Name:       "supplier_part_number",
				Header:     "Supplier Part #",
				Width:      150,
				Filterable: true,
			},
			"quantity_ordered": {
				Name:       "quantity_ordered",
				Header:     "Qty Ordered",
				Width:      110,
				Sortable:   true,
				Filterable: true,
			},
			"quantity_received": {
				Name:       "quantity_received",
				Header:     "Qty Received",
				Width:      120,
				Sortable:   true,
				Filterable: true,
			},
			"quantity_pending": {
				Name:   "quantity_pending",
				Header: "Qty Pending",
				Width:  110,
			},
			"fulfillment_percentage": {
				Name:   "fulfillment_percentage",
				Header: "Fulfillment %",
				Width:  120,
			},
			"unit_cost": {
				Name:     "unit_cost",
				Header:   "Unit Cost",
				Width:    100,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "currency",
					Format: "USD",
				},
			},
			"line_total": {
				Name:     "line_total",
				Header:   "Line Total",
				Width:    110,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "currency",
					Format: "USD",
				},
			},
			"line_item_status_name": {
				Name:       "line_item_status_name",
				Header:     "Status",
				Width:      120,
				Filterable: true,
			},
			"expected_delivery_date": {
				Name:     "expected_delivery_date",
				Header:   "Expected Delivery",
				Width:    150,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02",
				},
			},
			"created_by_username": {
				Name:       "created_by_username",
				Header:     "Created By",
				Width:      130,
				Filterable: true,
			},
			"id": {
				Name:   "id",
				Header: "Actions",
				Width:  100,
				Link: &tablebuilder.LinkConfig{
					URL:   "/procurement/purchase-order-line-items/{id}",
					Label: "View",
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
		Roles:   []string{"admin", "procurement"},
		Actions: []string{"view", "edit", "export"},
	},
}

// Open Approvals - Purchase orders awaiting approval
var procurementOpenApprovalsPageConfig = &tablebuilder.Config{
	Title:           "Open Approvals",
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
			Source: "purchase_orders",
			Schema: "procurement",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "purchase_orders.id"},
					{Name: "order_number", TableColumn: "purchase_orders.order_number"},
					{Name: "supplier_id", TableColumn: "purchase_orders.supplier_id"},
					{Name: "purchase_order_status_id", TableColumn: "purchase_orders.purchase_order_status_id"},
					{Name: "delivery_warehouse_id", TableColumn: "purchase_orders.delivery_warehouse_id"},
					{Name: "order_date", TableColumn: "purchase_orders.order_date"},
					{Name: "expected_delivery_date", TableColumn: "purchase_orders.expected_delivery_date"},
					{Name: "subtotal", TableColumn: "purchase_orders.subtotal"},
					{Name: "tax_amount", TableColumn: "purchase_orders.tax_amount"},
					{Name: "shipping_cost", TableColumn: "purchase_orders.shipping_cost"},
					{Name: "total_amount", TableColumn: "purchase_orders.total_amount"},
					{Name: "currency", TableColumn: "purchase_orders.currency"},
					{Name: "requested_by", TableColumn: "purchase_orders.requested_by"},
					{Name: "notes", TableColumn: "purchase_orders.notes"},
					{Name: "created_date", TableColumn: "purchase_orders.created_date"},
				},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "suppliers",
						Schema:           "procurement",
						RelationshipFrom: "purchase_orders.supplier_id",
						RelationshipTo:   "suppliers.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "supplier_name", TableColumn: "suppliers.name"},
							{Name: "payment_terms", Alias: "supplier_payment_terms", TableColumn: "suppliers.payment_terms"},
							{Name: "lead_time_days", Alias: "supplier_lead_time_days", TableColumn: "suppliers.lead_time_days"},
						},
					},
					{
						Table:            "purchase_order_statuses",
						Schema:           "procurement",
						RelationshipFrom: "purchase_orders.purchase_order_status_id",
						RelationshipTo:   "purchase_order_statuses.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "status_name", TableColumn: "purchase_order_statuses.name"},
							{Name: "description", Alias: "status_description", TableColumn: "purchase_order_statuses.description"},
						},
					},
					{
						Table:            "warehouses",
						Schema:           "inventory",
						RelationshipFrom: "purchase_orders.delivery_warehouse_id",
						RelationshipTo:   "warehouses.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "warehouse_name", TableColumn: "warehouses.name"},
							{Name: "code", Alias: "warehouse_code", TableColumn: "warehouses.code"},
						},
					},
					{
						Table:            "users",
						Schema:           "core",
						RelationshipFrom: "purchase_orders.requested_by",
						RelationshipTo:   "users.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "username", Alias: "requested_by_username", TableColumn: "users.username"},
							{Name: "first_name", Alias: "requested_by_first_name", TableColumn: "users.first_name"},
							{Name: "last_name", Alias: "requested_by_last_name", TableColumn: "users.last_name"},
						},
					},
				},
				ClientComputedColumns: []tablebuilder.ComputedColumn{
					{
						Name:       "requested_by_full_name",
						Expression: "requested_by_first_name + ' ' + requested_by_last_name",
					},
					{
						Name:       "formatted_total",
						Expression: "currency + ' ' + total_amount.toFixed(2)",
					},
					{
						Name:       "days_pending",
						Expression: "Math.floor((new Date() - new Date(order_date)) / (1000 * 60 * 60 * 24))",
					},
				},
			},
			Filters: []tablebuilder.Filter{
				{
					Column:   "approved_by",
					Operator: "is_null",
					Value:    nil,
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "order_date",
					Direction: "desc",
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"order_number": {
				Name:       "order_number",
				Header:     "PO Number",
				Width:      150,
				Sortable:   true,
				Filterable: true,
			},
			"supplier_name": {
				Name:       "supplier_name",
				Header:     "Supplier",
				Width:      200,
				Sortable:   true,
				Filterable: true,
			},
			"status_name": {
				Name:       "status_name",
				Header:     "Status",
				Width:      120,
				Filterable: true,
			},
			"warehouse_name": {
				Name:       "warehouse_name",
				Header:     "Delivery Warehouse",
				Width:      180,
				Filterable: true,
			},
			"order_date": {
				Name:     "order_date",
				Header:   "Order Date",
				Width:    130,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02",
				},
			},
			"expected_delivery_date": {
				Name:     "expected_delivery_date",
				Header:   "Expected Delivery",
				Width:    150,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02",
				},
			},
			"formatted_total": {
				Name:   "formatted_total",
				Header: "Total Amount",
				Width:  130,
			},
			"days_pending": {
				Name:   "days_pending",
				Header: "Days Pending",
				Width:  120,
			},
			"requested_by_full_name": {
				Name:       "requested_by_full_name",
				Header:     "Requested By",
				Width:      150,
				Filterable: true,
			},
			"id": {
				Name:   "id",
				Header: "Actions",
				Width:  100,
				Link: &tablebuilder.LinkConfig{
					URL:   "/procurement/purchase-orders/{id}/approve",
					Label: "Review",
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
		Roles:   []string{"admin", "procurement", "manager"},
		Actions: []string{"view", "approve", "reject"},
	},
}

// Closed Approvals - Purchase orders that have been approved
var procurementClosedApprovalsPageConfig = &tablebuilder.Config{
	Title:           "Closed Approvals",
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
			Source: "purchase_orders",
			Schema: "procurement",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", TableColumn: "purchase_orders.id"},
					{Name: "order_number", TableColumn: "purchase_orders.order_number"},
					{Name: "supplier_id", TableColumn: "purchase_orders.supplier_id"},
					{Name: "purchase_order_status_id", TableColumn: "purchase_orders.purchase_order_status_id"},
					{Name: "delivery_warehouse_id", TableColumn: "purchase_orders.delivery_warehouse_id"},
					{Name: "order_date", TableColumn: "purchase_orders.order_date"},
					{Name: "expected_delivery_date", TableColumn: "purchase_orders.expected_delivery_date"},
					{Name: "actual_delivery_date", TableColumn: "purchase_orders.actual_delivery_date"},
					{Name: "subtotal", TableColumn: "purchase_orders.subtotal"},
					{Name: "tax_amount", TableColumn: "purchase_orders.tax_amount"},
					{Name: "shipping_cost", TableColumn: "purchase_orders.shipping_cost"},
					{Name: "total_amount", TableColumn: "purchase_orders.total_amount"},
					{Name: "currency", TableColumn: "purchase_orders.currency"},
					{Name: "requested_by", TableColumn: "purchase_orders.requested_by"},
					{Name: "approved_by", TableColumn: "purchase_orders.approved_by"},
					{Name: "approved_date", TableColumn: "purchase_orders.approved_date"},
					{Name: "notes", TableColumn: "purchase_orders.notes"},
					{Name: "created_date", TableColumn: "purchase_orders.created_date"},
				},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "suppliers",
						Schema:           "procurement",
						RelationshipFrom: "purchase_orders.supplier_id",
						RelationshipTo:   "suppliers.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "supplier_name", TableColumn: "suppliers.name"},
							{Name: "payment_terms", Alias: "supplier_payment_terms", TableColumn: "suppliers.payment_terms"},
							{Name: "lead_time_days", Alias: "supplier_lead_time_days", TableColumn: "suppliers.lead_time_days"},
						},
					},
					{
						Table:            "purchase_order_statuses",
						Schema:           "procurement",
						RelationshipFrom: "purchase_orders.purchase_order_status_id",
						RelationshipTo:   "purchase_order_statuses.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "status_name", TableColumn: "purchase_order_statuses.name"},
							{Name: "description", Alias: "status_description", TableColumn: "purchase_order_statuses.description"},
						},
					},
					{
						Table:            "warehouses",
						Schema:           "inventory",
						RelationshipFrom: "purchase_orders.delivery_warehouse_id",
						RelationshipTo:   "warehouses.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "warehouse_name", TableColumn: "warehouses.name"},
							{Name: "code", Alias: "warehouse_code", TableColumn: "warehouses.code"},
						},
					},
					{
						Table:            "users",
						Schema:           "core",
						RelationshipFrom: "purchase_orders.requested_by",
						RelationshipTo:   "users.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "username", Alias: "requested_by_username", TableColumn: "users.username"},
							{Name: "first_name", Alias: "requested_by_first_name", TableColumn: "users.first_name"},
							{Name: "last_name", Alias: "requested_by_last_name", TableColumn: "users.last_name"},
						},
					},
					{
						Table:            "users",
						Schema:           "core",
						RelationshipFrom: "purchase_orders.approved_by",
						RelationshipTo:   "users.id",
						JoinType:         "left",
						Alias:            "approver",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "username", Alias: "approved_by_username", TableColumn: "approver.username"},
							{Name: "first_name", Alias: "approved_by_first_name", TableColumn: "approver.first_name"},
							{Name: "last_name", Alias: "approved_by_last_name", TableColumn: "approver.last_name"},
						},
					},
				},
				ClientComputedColumns: []tablebuilder.ComputedColumn{
					{
						Name:       "requested_by_full_name",
						Expression: "requested_by_first_name + ' ' + requested_by_last_name",
					},
					{
						Name:       "approved_by_full_name",
						Expression: "approved_by_first_name + ' ' + approved_by_last_name",
					},
					{
						Name:       "formatted_total",
						Expression: "currency + ' ' + total_amount.toFixed(2)",
					},
					{
						Name:       "delivery_status",
						Expression: "actual_delivery_date ? 'delivered' : (new Date(expected_delivery_date) < new Date() ? 'overdue' : 'pending')",
					},
				},
			},
			Filters: []tablebuilder.Filter{
				{
					Column:   "approved_by",
					Operator: "is_not_null",
					Value:    nil,
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "approved_date",
					Direction: "desc",
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"order_number": {
				Name:       "order_number",
				Header:     "PO Number",
				Width:      150,
				Sortable:   true,
				Filterable: true,
			},
			"supplier_name": {
				Name:       "supplier_name",
				Header:     "Supplier",
				Width:      200,
				Sortable:   true,
				Filterable: true,
			},
			"status_name": {
				Name:       "status_name",
				Header:     "Status",
				Width:      120,
				Filterable: true,
			},
			"warehouse_name": {
				Name:       "warehouse_name",
				Header:     "Delivery Warehouse",
				Width:      180,
				Filterable: true,
			},
			"order_date": {
				Name:     "order_date",
				Header:   "Order Date",
				Width:    130,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02",
				},
			},
			"approved_date": {
				Name:     "approved_date",
				Header:   "Approved Date",
				Width:    130,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02",
				},
			},
			"expected_delivery_date": {
				Name:     "expected_delivery_date",
				Header:   "Expected Delivery",
				Width:    150,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02",
				},
			},
			"formatted_total": {
				Name:   "formatted_total",
				Header: "Total Amount",
				Width:  130,
			},
			"delivery_status": {
				Name:       "delivery_status",
				Header:     "Delivery Status",
				Width:      130,
				Filterable: true,
			},
			"requested_by_full_name": {
				Name:       "requested_by_full_name",
				Header:     "Requested By",
				Width:      150,
				Filterable: true,
			},
			"approved_by_full_name": {
				Name:       "approved_by_full_name",
				Header:     "Approved By",
				Width:      150,
				Filterable: true,
			},
			"id": {
				Name:   "id",
				Header: "Actions",
				Width:  100,
				Link: &tablebuilder.LinkConfig{
					URL:   "/procurement/purchase-orders/{id}",
					Label: "View",
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
		Roles:   []string{"admin", "procurement", "manager"},
		Actions: []string{"view", "export"},
	},
}

// =============================================================================
// FORM CONFIGS
// =============================================================================
