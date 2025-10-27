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

// =============================================================================
// FORM CONFIGS
// =============================================================================
