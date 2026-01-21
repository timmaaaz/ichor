package seedmodels

import (
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)

// =============================================================================
// TABLE CONFIGS
// =============================================================================
var TableConfig = &tablebuilder.Config{
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
			"products.id": {
				Name:   "products.id",
				Header: "ID",
				Width:  100,
				Type:   "uuid",
			},
			"products.name": {
				Name:       "products.name",
				Header:     "Product Name",
				Width:      200,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"products.sku": {
				Name:       "products.sku",
				Header:     "SKU",
				Width:      150,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
				Editable: &tablebuilder.EditableConfig{
					Type:        "text",
					Placeholder: "SKU-12345",
				},
			},
			"products.is_active": {
				Name:       "products.is_active",
				Header:     "Is Active",
				Width:      100,
				Type:       "boolean",
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
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"current_quantity": {
				Name:   "current_quantity",
				Header: "Current Stock",
				Width:  120,
				Align:  "right",
				Type:   "number",
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
				Type:         "computed",
				CellTemplate: "status",
			},
			"stock_percentage": {
				Name:   "stock_percentage",
				Header: "Capacity",
				Width:  100,
				Align:  "right",
				Type:   "computed",
				Format: &tablebuilder.FormatConfig{
					Type:      "percent",
					Precision: 1,
				},
			},
			"product_id": {
				Name:       "product_id",
				Header:     "Product",
				Width:      200,
				Type:       "lookup",
				Filterable: true,
				Link: &tablebuilder.LinkConfig{
					URL:   "/products/products/{product_id}",
					Label: "View Product",
				},
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "products.products",
					LabelColumn: "products.name",
					ValueColumn: "products.id",
				},
			},
			"inventory_items.id": {
				Name:   "inventory_items.id",
				Header: "ID",
				Width:  100,
				Type:   "uuid",
			},
			"inventory_items.reorder_point": {
				Name:   "inventory_items.reorder_point",
				Header: "Reorder Point",
				Width:  100,
				Type:   "number",
			},
			"inventory_items.maximum_stock": {
				Name:   "inventory_items.maximum_stock",
				Header: "Maximum Stock",
				Width:  100,
				Type:   "number",
			},
			"products.sku": {
				Name:   "products.sku",
				Header: "SKU",
				Width:  120,
				Type:   "string",
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

var OrdersConfig = &tablebuilder.Config{
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
					{Name: "orders_id", Alias: "orders_id", TableColumn: "orders.id"},
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
			"orders_id": {
				Name:   "orders_id",
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
				Name:       "order_fulfillment_status_id",
				Header:     "Fulfillment Status",
				Width:      150,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "sales.order_fulfillment_statuses",
					LabelColumn: "order_fulfillment_statuses.name",
					ValueColumn: "order_fulfillment_statuses.id",
				},
			},
			"order_customer_id": {
				Name:       "order_customer_id",
				Header:     "Customer",
				Width:      200,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "sales.customers",
					LabelColumn: "customers.notes",
					ValueColumn: "customers.id",
				},
			},
			"customer_id": {
				Name:       "customer_id",
				Header:     "Customer",
				Width:      200,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "sales.customers",
					LabelColumn: "customers.name",
					ValueColumn: "customers.id",
				},
			},
			"customer_contact_info_id": {
				Name:       "customer_contact_info_id",
				Header:     "Contact Info",
				Width:      200,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "core.contact_infos",
					LabelColumn: "contact_infos.email_address",
					ValueColumn: "contact_infos.id",
				},
			},
			"customer_delivery_address_id": {
				Name:       "customer_delivery_address_id",
				Header:     "Delivery Address",
				Width:      200,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "geography.addresses",
					LabelColumn: "addresses.street",
					ValueColumn: "addresses.id",
				},
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

// Dedicated Orders Page Config
var OrdersTableConfig = &tablebuilder.Config{
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
						Expression: "daysUntil(due_date)",
					},
					{
						Name:       "is_overdue",
						Expression: "isOverdue(due_date) && status_name != 'Delivered'",
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
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"customer_name": {
				Name:       "customer_name",
				Header:     "Customer",
				Width:      200,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"orders.due_date": {
				Name:     "orders.due_date",
				Header:   "Due Date",
				Width:    120,
				Sortable: true,
				Type:     "datetime",
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
				Type:         "status",
				CellTemplate: "status",
			},
			"days_until_due": {
				Name:   "days_until_due",
				Header: "Days Until Due",
				Width:  120,
				Align:  "center",
				Type:   "number",
				Format: &tablebuilder.FormatConfig{
					Type:      "number",
					Precision: 0,
				},
			},
			"orders.id": {
				Name:   "orders.id",
				Header: "Actions",
				Width:  100,
				Type:   "uuid",
				Link: &tablebuilder.LinkConfig{
					URL:   "/sales/orders/{orders.id}",
					Label: "View",
				},
			},
			"is_overdue": {
				Name:   "is_overdue",
				Header: "Overdue",
				Width:  80,
				Type:   "boolean",
			},
			"orders.customer_id": {
				Name:       "orders.customer_id",
				Header:     "Customer",
				Width:      200,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "sales.customers",
					LabelColumn: "customers.name",
					ValueColumn: "customers.id",
				},
			},
			"orders.order_fulfillment_status_id": {
				Name:       "orders.order_fulfillment_status_id",
				Header:     "Fulfillment Status",
				Width:      150,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "sales.order_fulfillment_statuses",
					LabelColumn: "order_fulfillment_statuses.name",
					ValueColumn: "order_fulfillment_statuses.id",
				},
			},
			"orders.created_date": {
				Name:   "orders.created_date",
				Header: "Created Date",
				Width:  120,
				Type:   "datetime",
			},
			"orders.updated_date": {
				Name:   "orders.updated_date",
				Header: "Updated Date",
				Width:  120,
				Type:   "datetime",
			},
			"customer_id_fk": {
				Name:   "customer_id_fk",
				Header: "Customer FK",
				Width:  100,
				Type:   "uuid",
			},
			"customer_contact_id": {
				Name:   "customer_contact_id",
				Header: "Customer Contact ID",
				Width:  100,
				Type:   "uuid",
			},
			"status_description": {
				Name:   "status_description",
				Header: "Status Description",
				Width:  200,
				Type:   "string",
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
var SuppliersTableConfig = &tablebuilder.Config{
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
					{Name: "id", Alias: "suppliers_id", TableColumn: "suppliers.id"},
					{Name: "name", Alias: "suppliers_name", TableColumn: "suppliers.name"},
					{Name: "lead_time_days", Alias: "suppliers_lead_time_days", TableColumn: "suppliers.lead_time_days"},
					{Name: "rating", Alias: "suppliers_rating", TableColumn: "suppliers.rating"},
					{Name: "is_active", Alias: "suppliers_is_active", TableColumn: "suppliers.is_active"},
					{Name: "contact_infos_id", Alias: "suppliers_contact_infos_id", TableColumn: "suppliers.contact_infos_id"},
					{Name: "created_date", Alias: "suppliers_created_date", TableColumn: "suppliers.created_date"},
					{Name: "updated_date", Alias: "suppliers_updated_date", TableColumn: "suppliers.updated_date"},
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
					{
						Table:            "payment_terms",
						Schema:           "core",
						RelationshipFrom: "suppliers.payment_term_id",
						RelationshipTo:   "payment_terms.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "payment_terms", TableColumn: "payment_terms.name"},
						},
					},
				},
				ClientComputedColumns: []tablebuilder.ComputedColumn{
					{
						Name:       "rating_stars",
						Expression: "round(suppliers_rating * 2) / 2",
					},
					{
						Name:       "performance_level",
						Expression: "suppliers_rating >= 4.5 ? 'excellent' : suppliers_rating >= 3.5 ? 'good' : suppliers_rating >= 2.5 ? 'fair' : 'poor'",
					},
				},
			},
			Filters: []tablebuilder.Filter{
				{
					Column:   "suppliers.is_active",
					Operator: "eq",
					Value:    true,
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "suppliers.rating",
					Direction: "desc",
				},
				{
					Column:    "suppliers.name",
					Direction: "asc",
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"suppliers_name": {
				Name:       "suppliers_name",
				Header:     "Supplier Name",
				Width:      200,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"payment_terms": {
				Name:       "payment_terms",
				Header:     "Payment Terms",
				Width:      150,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"suppliers_lead_time_days": {
				Name:     "suppliers_lead_time_days",
				Header:   "Lead Time (days)",
				Width:    130,
				Align:    "center",
				Type:     "number",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:      "number",
					Precision: 0,
				},
			},
			"suppliers_rating": {
				Name:     "suppliers_rating",
				Header:   "Rating",
				Width:    100,
				Align:    "center",
				Type:     "number",
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
				Type:         "computed",
				CellTemplate: "badge",
			},
			"suppliers_is_active": {
				Name:   "suppliers_is_active",
				Header: "Active",
				Width:  80,
				Align:  "center",
				Type:   "boolean",
				Format: &tablebuilder.FormatConfig{
					Type: "boolean",
				},
				Editable: &tablebuilder.EditableConfig{
					Type: "boolean",
				},
			},
			"suppliers_id": {
				Name:   "suppliers_id",
				Header: "Actions",
				Width:  100,
				Type:   "uuid",
				Link: &tablebuilder.LinkConfig{
					URL:   "/procurement/suppliers/{suppliers_id}",
					Label: "View",
				},
			},
			"suppliers_contact_infos_id": {
				Name:       "suppliers_contact_infos_id",
				Header:     "Contact Info",
				Width:      200,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "core.contact_infos",
					LabelColumn: "contact_infos.email_address",
					ValueColumn: "contact_infos.id",
				},
			},
			"suppliers_created_date": {
				Name:   "suppliers_created_date",
				Header: "Created Date",
				Width:  120,
				Type:   "datetime",
			},
			"suppliers_updated_date": {
				Name:   "suppliers_updated_date",
				Header: "Updated Date",
				Width:  120,
				Type:   "datetime",
			},
			"primary_phone_number": {
				Name:   "primary_phone_number",
				Header: "Primary Phone",
				Width:  150,
				Type:   "string",
			},
			"secondary_phone_number": {
				Name:   "secondary_phone_number",
				Header: "Secondary Phone",
				Width:  150,
				Type:   "string",
			},
			"email_address": {
				Name:   "email_address",
				Header: "Email",
				Width:  200,
				Type:   "string",
			},
			"rating_stars": {
				Name:   "rating_stars",
				Header: "Rating Stars",
				Width:  100,
				Type:   "computed",
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
var OrderLineItemsTableConfig = &tablebuilder.Config{
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
				Type:       "string",
				Sortable:   true,
				Filterable: true,
				Link: &tablebuilder.LinkConfig{
					URL:   "/sales/orders/{order_line_items.order_id}",
					Label: "{order_number}",
				},
			},
			"customer_name": {
				Name:       "customer_name",
				Header:     "Customer",
				Width:      180,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"product_name": {
				Name:       "product_name",
				Header:     "Product",
				Width:      200,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"product_sku": {
				Name:       "product_sku",
				Header:     "SKU",
				Width:      120,
				Type:       "string",
				Filterable: true,
			},
			"order_line_items.quantity": {
				Name:     "order_line_items.quantity",
				Header:   "Qty",
				Width:    80,
				Align:    "center",
				Type:     "number",
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
			"order_line_items.discount": {
				Name:     "order_line_items.discount",
				Header:   "Discount %",
				Width:    100,
				Align:    "right",
				Type:     "number",
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
				Type:   "computed",
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
				Type:   "computed",
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
				Type:         "status",
				CellTemplate: "status",
			},
			"order_date": {
				Name:     "order_date",
				Header:   "Order Date",
				Width:    120,
				Type:     "datetime",
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
				Type:   "computed",
			},
			"order_line_items.created_date": {
				Name:     "order_line_items.created_date",
				Header:   "Created",
				Width:    150,
				Type:     "datetime",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02 15:04",
				},
			},
			"order_line_items.id": {
				Name:   "order_line_items.id",
				Header: "Actions",
				Width:  100,
				Type:   "uuid",
				Link: &tablebuilder.LinkConfig{
					URL:   "/sales/order-line-items/{order_line_items.id}",
					Label: "View",
				},
			},
			"order_line_items.order_id": {
				Name:       "order_line_items.order_id",
				Header:     "Order",
				Width:      150,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "sales.orders",
					LabelColumn: "orders.number",
					ValueColumn: "orders.id",
				},
			},
			"order_line_items.product_id": {
				Name:       "order_line_items.product_id",
				Header:     "Product",
				Width:      200,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "products.products",
					LabelColumn: "products.name",
					ValueColumn: "products.id",
				},
			},
			"order_line_items.line_item_fulfillment_statuses_id": {
				Name:       "order_line_items.line_item_fulfillment_statuses_id",
				Header:     "Fulfillment Status",
				Width:      150,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "sales.line_item_fulfillment_statuses",
					LabelColumn: "line_item_fulfillment_statuses.name",
					ValueColumn: "line_item_fulfillment_statuses.id",
				},
			},
			"order_line_items.created_by": {
				Name:   "order_line_items.created_by",
				Header: "Created By ID",
				Width:  100,
				Type:   "uuid",
			},
			"order_line_items.updated_by": {
				Name:   "order_line_items.updated_by",
				Header: "Updated By ID",
				Width:  100,
				Type:   "uuid",
			},
			"order_line_items.updated_date": {
				Name:   "order_line_items.updated_date",
				Header: "Updated Date",
				Width:  120,
				Type:   "datetime",
			},
			"order_customer_id": {
				Name:       "order_customer_id",
				Header:     "Customer",
				Width:      200,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "sales.customers",
					LabelColumn: "customers.name",
					ValueColumn: "customers.id",
				},
			},
			"fulfillment_status_description": {
				Name:   "fulfillment_status_description",
				Header: "Status Description",
				Width:  200,
				Type:   "string",
			},
			"created_by_username": {
				Name:   "created_by_username",
				Header: "Creator Username",
				Width:  150,
				Type:   "string",
			},
			"created_by_first_name": {
				Name:   "created_by_first_name",
				Header: "Creator First Name",
				Width:  150,
				Type:   "string",
			},
			"created_by_last_name": {
				Name:   "created_by_last_name",
				Header: "Creator Last Name",
				Width:  150,
				Type:   "string",
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
var CategoriesTableConfig = &tablebuilder.Config{
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
			"product_categories.name": {
				Name:       "product_categories.name",
				Header:     "Category Name",
				Width:      250,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
				Editable: &tablebuilder.EditableConfig{
					Type:        "text",
					Placeholder: "Category name",
				},
			},
			"product_categories.description": {
				Name:       "product_categories.description",
				Header:     "Description",
				Width:      400,
				Type:       "string",
				Filterable: true,
				Editable: &tablebuilder.EditableConfig{
					Type:        "textarea",
					Placeholder: "Category description",
				},
			},
			"product_categories.created_date": {
				Name:     "product_categories.created_date",
				Header:   "Created",
				Width:    150,
				Type:     "datetime",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02 15:04",
				},
			},
			"product_categories.updated_date": {
				Name:     "product_categories.updated_date",
				Header:   "Last Updated",
				Width:    150,
				Type:     "datetime",
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
				Type:   "uuid",
				Link: &tablebuilder.LinkConfig{
					URL:   "/products/categories/{product_category_id}",
					Label: "View",
				},
			},
			"product_categories.id": {
				Name:   "product_categories.id",
				Header: "ID",
				Width:  100,
				Type:   "uuid",
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
// ASSETS MODULE CONFIGS
// =============================================================================

// Assets List Page Config
var AssetsListTableConfig = &tablebuilder.Config{
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
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"asset_type_name": {
				Name:       "asset_type_name",
				Header:     "Type",
				Width:      150,
				Type:       "status",
				Sortable:   true,
				Filterable: true,
			},
			"assets.serial_number": {
				Name:       "assets.serial_number",
				Header:     "Serial Number",
				Width:      180,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"valid_assets.model_number": {
				Name:       "valid_assets.model_number",
				Header:     "Model",
				Width:      150,
				Type:       "string",
				Filterable: true,
			},
			"condition_name": {
				Name:       "condition_name",
				Header:     "Condition",
				Width:      120,
				Type:       "status",
				Sortable:   true,
				Filterable: true,
			},
			"valid_assets.price": {
				Name:     "valid_assets.price",
				Header:   "Value",
				Width:    120,
				Align:    "right",
				Type:     "number",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:      "currency",
					Currency:  "USD",
					Precision: 2,
				},
			},
			"assets.last_maintenance_time": {
				Name:     "assets.last_maintenance_time",
				Header:   "Last Maintenance",
				Width:    150,
				Type:     "datetime",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "2006-01-02",
				},
			},
			"assets.id": {
				Name:   "assets.id",
				Header: "Actions",
				Width:  100,
				Type:   "uuid",
				Link: &tablebuilder.LinkConfig{
					URL:   "/assets/list/{assets.id}",
					Label: "View",
				},
			},
			"valid_assets.maintenance_interval": {
				Name:   "valid_assets.maintenance_interval",
				Header: "Maintenance Interval",
				Width:  150,
				Type:   "number",
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
var AssetsRequestsTableConfig = &tablebuilder.Config{
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
				Type:       "computed",
				Sortable:   true,
				Filterable: true,
			},
			"asset_name": {
				Name:       "asset_name",
				Header:     "Asset",
				Width:      200,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"asset_serial_number": {
				Name:       "asset_serial_number",
				Header:     "Serial Number",
				Width:      150,
				Type:       "string",
				Filterable: true,
			},
			"approval_status_name": {
				Name:         "approval_status_name",
				Header:       "Approval Status",
				Width:        150,
				Sortable:     true,
				Filterable:   true,
				Type:         "status",
				CellTemplate: "status",
			},
			"fulfillment_status_name": {
				Name:         "fulfillment_status_name",
				Header:       "Fulfillment Status",
				Width:        150,
				Sortable:     true,
				Filterable:   true,
				Type:         "status",
				CellTemplate: "status",
			},
			"approved_by_username": {
				Name:       "approved_by_username",
				Header:     "Approved By",
				Width:      150,
				Type:       "string",
				Filterable: true,
			},
			"user_assets.date_received": {
				Name:     "user_assets.date_received",
				Header:   "Date Received",
				Width:    120,
				Type:     "datetime",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "2006-01-02",
				},
			},
			"user_assets.id": {
				Name:   "user_assets.id",
				Header: "Actions",
				Width:  100,
				Type:   "uuid",
				Link: &tablebuilder.LinkConfig{
					URL:   "/assets/requests/{user_assets.id}",
					Label: "View",
				},
			},
			"user_assets.last_maintenance": {
				Name:   "user_assets.last_maintenance",
				Header: "Last Maintenance",
				Width:  150,
				Type:   "datetime",
			},
			"user_first_name": {
				Name:   "user_first_name",
				Header: "First Name",
				Width:  150,
				Type:   "string",
			},
			"user_last_name": {
				Name:   "user_last_name",
				Header: "Last Name",
				Width:  150,
				Type:   "string",
			},
			"username": {
				Name:   "username",
				Header: "Username",
				Width:  150,
				Type:   "string",
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
var HrEmployeesTableConfig = &tablebuilder.Config{
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
				Type:       "computed",
				Sortable:   true,
				Filterable: true,
			},
			"users.email": {
				Name:       "users.email",
				Header:     "Email",
				Width:      250,
				Type:       "string",
				Filterable: true,
			},
			"title_name": {
				Name:       "title_name",
				Header:     "Title",
				Width:      180,
				Type:       "status",
				Sortable:   true,
				Filterable: true,
			},
			"office_name": {
				Name:       "office_name",
				Header:     "Office",
				Width:      150,
				Type:       "status",
				Sortable:   true,
				Filterable: true,
			},
			"office_location": {
				Name:       "office_location",
				Header:     "Location",
				Width:      200,
				Type:       "computed",
				Filterable: true,
			},
			"users.date_hired": {
				Name:     "users.date_hired",
				Header:   "Date Hired",
				Width:    120,
				Type:     "datetime",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "2006-01-02",
				},
			},
			"users.enabled": {
				Name:   "users.enabled",
				Header: "Active",
				Width:  80,
				Align:  "center",
				Type:   "boolean",
				Format: &tablebuilder.FormatConfig{
					Type: "boolean",
				},
			},
			"users.id": {
				Name:   "users.id",
				Header: "Actions",
				Width:  100,
				Type:   "uuid",
				Link: &tablebuilder.LinkConfig{
					URL:   "/hr/employees/{users.id}",
					Label: "View",
				},
			},
			"users.first_name": {
				Name:   "users.first_name",
				Header: "First Name",
				Width:  150,
				Type:   "string",
			},
			"users.last_name": {
				Name:   "users.last_name",
				Header: "Last Name",
				Width:  150,
				Type:   "string",
			},
			"office_street": {
				Name:   "office_street",
				Header: "Office Street",
				Width:  200,
				Type:   "string",
			},
			"office_city": {
				Name:   "office_city",
				Header: "Office City",
				Width:  150,
				Type:   "string",
			},
			"office_region": {
				Name:   "office_region",
				Header: "Office Region",
				Width:  150,
				Type:   "string",
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
var HrOfficesTableConfig = &tablebuilder.Config{
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
					{Name: "id", Alias: "offices_id", TableColumn: "offices.id"},
					{Name: "name", Alias: "offices_name", TableColumn: "offices.name"},
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
					Column:    "offices.name",
					Direction: "asc",
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"offices_name": {
				Name:       "offices_name",
				Header:     "Office Name",
				Width:      200,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"full_address": {
				Name:       "full_address",
				Header:     "Address",
				Width:      400,
				Type:       "computed",
				Filterable: true,
			},
			"city_name": {
				Name:       "city_name",
				Header:     "City",
				Width:      150,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"region_name": {
				Name:       "region_name",
				Header:     "State/Region",
				Width:      150,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"country_name": {
				Name:       "country_name",
				Header:     "Country",
				Width:      120,
				Type:       "string",
				Filterable: true,
			},
			"offices_id": {
				Name:   "offices_id",
				Header: "Actions",
				Width:  100,
				Type:   "uuid",
				Link: &tablebuilder.LinkConfig{
					URL:   "/hr/offices/{offices_id}",
					Label: "View",
				},
			},
			"street_line_1": {
				Name:   "street_line_1",
				Header: "Street Line 1",
				Width:  200,
				Type:   "string",
			},
			"street_line_2": {
				Name:   "street_line_2",
				Header: "Street Line 2",
				Width:  200,
				Type:   "string",
			},
			"streets.postal_code": {
				Name:   "streets.postal_code",
				Header: "Postal Code",
				Width:  100,
				Type:   "string",
			},
			"region_code": {
				Name:   "region_code",
				Header: "Region Code",
				Width:  100,
				Type:   "string",
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
var InventoryWarehousesTableConfig = &tablebuilder.Config{
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
					Column:    "warehouses.name",
					Direction: "asc",
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"warehouses.name": {
				Name:       "warehouses.name",
				Header:     "Warehouse Name",
				Width:      200,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"location": {
				Name:       "location",
				Header:     "Location",
				Width:      200,
				Type:       "computed",
				Filterable: true,
			},
			"street_line_1": {
				Name:       "street_line_1",
				Header:     "Address",
				Width:      250,
				Type:       "string",
				Filterable: true,
			},
			"warehouses.is_active": {
				Name:   "warehouses.is_active",
				Header: "Active",
				Width:  80,
				Type:   "boolean",
				Align:  "center",
				Format: &tablebuilder.FormatConfig{
					Type: "boolean",
				},
			},
			"created_by_username": {
				Name:       "created_by_username",
				Header:     "Created By",
				Width:      150,
				Type:       "string",
				Filterable: true,
			},
			"warehouses.created_date": {
				Name:     "warehouses.created_date",
				Header:   "Created",
				Width:    150,
				Type:     "datetime",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02 15:04",
				},
			},
			"warehouses.id": {
				Name:   "warehouses.id",
				Header: "Actions",
				Width:  100,
				Type:   "uuid",
				Link: &tablebuilder.LinkConfig{
					URL:   "/inventory/warehouses/{warehouses.id}",
					Label: "View",
				},
			},
			"warehouses.updated_date": {
				Name:   "warehouses.updated_date",
				Header: "Updated Date",
				Width:  150,
				Type:   "datetime",
			},
			"streets.postal_code": {
				Name:   "streets.postal_code",
				Header: "Postal Code",
				Width:  100,
				Type:   "string",
			},
			"city_name": {
				Name:   "city_name",
				Header: "City",
				Width:  150,
				Type:   "string",
			},
			"region_code": {
				Name:   "region_code",
				Header: "Region Code",
				Width:  100,
				Type:   "string",
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
var InventoryItemsTableConfig = ComplexConfig

// Inventory Adjustments Page Config
var InventoryAdjustmentsTableConfig = &tablebuilder.Config{
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
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"product_sku": {
				Name:       "product_sku",
				Header:     "SKU",
				Width:      120,
				Type:       "string",
				Filterable: true,
			},
			"warehouse_name": {
				Name:       "warehouse_name",
				Header:     "Warehouse",
				Width:      150,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"location_code": {
				Name:       "location_code",
				Header:     "Location",
				Width:      150,
				Type:       "computed",
				Filterable: true,
			},
			"inventory_adjustments.quantity_change": {
				Name:     "inventory_adjustments.quantity_change",
				Header:   "Qty Change",
				Width:    100,
				Type:     "number",
				Align:    "right",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:      "number",
					Precision: 0,
				},
			},
			"inventory_adjustments.reason_code": {
				Name:       "inventory_adjustments.reason_code",
				Header:     "Reason",
				Width:      120,
				Type:       "status",
				Filterable: true,
			},
			"adjusted_by_username": {
				Name:       "adjusted_by_username",
				Header:     "Adjusted By",
				Width:      130,
				Type:       "string",
				Filterable: true,
			},
			"approved_by_username": {
				Name:       "approved_by_username",
				Header:     "Approved By",
				Width:      130,
				Type:       "string",
				Filterable: true,
			},
			"inventory_adjustments.adjustment_date": {
				Name:     "inventory_adjustments.adjustment_date",
				Header:   "Date",
				Width:    150,
				Type:     "datetime",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02 15:04",
				},
			},
			"inventory_adjustments.id": {
				Name:   "inventory_adjustments.id",
				Header: "Actions",
				Width:  100,
				Type:   "uuid",
				Link: &tablebuilder.LinkConfig{
					URL:   "/inventory/adjustments/{inventory_adjustments.id}",
					Label: "View",
				},
			},
			"inventory_adjustments.notes": {
				Name:   "inventory_adjustments.notes",
				Header: "Notes",
				Width:  200,
				Type:   "string",
			},
			"inventory_adjustments.created_date": {
				Name:   "inventory_adjustments.created_date",
				Header: "Created Date",
				Width:  150,
				Type:   "datetime",
			},
			"inventory_locations.aisle": {
				Name:   "inventory_locations.aisle",
				Header: "Aisle",
				Width:  80,
				Type:   "string",
			},
			"inventory_locations.rack": {
				Name:   "inventory_locations.rack",
				Header: "Rack",
				Width:  80,
				Type:   "string",
			},
			"inventory_locations.shelf": {
				Name:   "inventory_locations.shelf",
				Header: "Shelf",
				Width:  80,
				Type:   "string",
			},
			"inventory_locations.bin": {
				Name:   "inventory_locations.bin",
				Header: "Bin",
				Width:  80,
				Type:   "string",
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
var InventoryTransfersTableConfig = &tablebuilder.Config{
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
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"product_sku": {
				Name:       "product_sku",
				Header:     "SKU",
				Width:      120,
				Type:       "string",
				Filterable: true,
			},
			"from_location": {
				Name:       "from_location",
				Header:     "From",
				Width:      180,
				Type:       "computed",
				Filterable: true,
			},
			"to_location": {
				Name:       "to_location",
				Header:     "To",
				Width:      180,
				Type:       "computed",
				Filterable: true,
			},
			"transfer_orders.quantity": {
				Name:     "transfer_orders.quantity",
				Header:   "Quantity",
				Width:    90,
				Type:     "number",
				Align:    "right",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:      "number",
					Precision: 0,
				},
			},
			"transfer_orders.status": {
				Name:         "transfer_orders.status",
				Header:       "Status",
				Width:        120,
				Type:         "status",
				Sortable:     true,
				Filterable:   true,
				CellTemplate: "status",
			},
			"requested_by_username": {
				Name:       "requested_by_username",
				Header:     "Requested By",
				Width:      130,
				Type:       "string",
				Filterable: true,
			},
			"approved_by_username": {
				Name:       "approved_by_username",
				Header:     "Approved By",
				Width:      130,
				Type:       "string",
				Filterable: true,
			},
			"transfer_orders.transfer_date": {
				Name:     "transfer_orders.transfer_date",
				Header:   "Transfer Date",
				Width:    150,
				Type:     "datetime",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02 15:04",
				},
			},
			"transfer_orders.id": {
				Name:   "transfer_orders.id",
				Header: "Actions",
				Width:  100,
				Type:   "uuid",
				Link: &tablebuilder.LinkConfig{
					URL:   "/inventory/transfers/{transfer_orders.id}",
					Label: "View",
				},
			},
			"transfer_orders.created_date": {
				Name:   "transfer_orders.created_date",
				Header: "Created Date",
				Width:  150,
				Type:   "datetime",
			},
			"from_aisle": {
				Name:   "from_aisle",
				Header: "From Aisle",
				Width:  100,
				Type:   "string",
			},
			"from_rack": {
				Name:   "from_rack",
				Header: "From Rack",
				Width:  100,
				Type:   "string",
			},
			"from_warehouse_name": {
				Name:   "from_warehouse_name",
				Header: "From Warehouse",
				Width:  150,
				Type:   "string",
			},
			"to_aisle": {
				Name:   "to_aisle",
				Header: "To Aisle",
				Width:  100,
				Type:   "string",
			},
			"to_rack": {
				Name:   "to_rack",
				Header: "To Rack",
				Width:  100,
				Type:   "string",
			},
			"to_warehouse_name": {
				Name:   "to_warehouse_name",
				Header: "To Warehouse",
				Width:  150,
				Type:   "string",
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
var SalesCustomersTableConfig = &tablebuilder.Config{
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
					Column:    "customers.name",
					Direction: "asc",
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"customers.name": {
				Name:       "customers.name",
				Header:     "Customer Name",
				Width:      200,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"contact_full_name": {
				Name:       "contact_full_name",
				Header:     "Contact Person",
				Width:      180,
				Type:       "computed",
				Filterable: true,
			},
			"contact_email": {
				Name:       "contact_email",
				Header:     "Email",
				Width:      200,
				Type:       "string",
				Filterable: true,
			},
			"contact_phone": {
				Name:       "contact_phone",
				Header:     "Phone",
				Width:      150,
				Type:       "string",
				Filterable: true,
			},
			"delivery_location": {
				Name:       "delivery_location",
				Header:     "Location",
				Width:      180,
				Type:       "computed",
				Filterable: true,
			},
			"created_by_username": {
				Name:       "created_by_username",
				Header:     "Created By",
				Width:      130,
				Type:       "string",
				Filterable: true,
			},
			"customers.created_date": {
				Name:     "customers.created_date",
				Header:   "Created",
				Width:    150,
				Type:     "datetime",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02 15:04",
				},
			},
			"customers.id": {
				Name:   "customers.id",
				Header: "Actions",
				Width:  100,
				Type:   "uuid",
				Link: &tablebuilder.LinkConfig{
					URL:   "/sales/customers/{customers.id}",
					Label: "View",
				},
			},
			"customers.notes": {
				Name:   "customers.notes",
				Header: "Notes",
				Width:  200,
				Type:   "string",
			},
			"customers.updated_date": {
				Name:   "customers.updated_date",
				Header: "Updated Date",
				Width:  150,
				Type:   "datetime",
			},
			"contact_first_name": {
				Name:   "contact_first_name",
				Header: "Contact First Name",
				Width:  150,
				Type:   "string",
			},
			"contact_last_name": {
				Name:   "contact_last_name",
				Header: "Contact Last Name",
				Width:  150,
				Type:   "string",
			},
			"delivery_street": {
				Name:   "delivery_street",
				Header: "Delivery Street",
				Width:  200,
				Type:   "string",
			},
			"delivery_city": {
				Name:   "delivery_city",
				Header: "Delivery City",
				Width:  150,
				Type:   "string",
			},
			"delivery_region_code": {
				Name:   "delivery_region_code",
				Header: "Delivery Region",
				Width:  100,
				Type:   "string",
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

var PurchaseOrderTableConfig = &tablebuilder.Config{
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
							{Name: "lead_time_days", Alias: "supplier_lead_time_days", TableColumn: "suppliers.lead_time_days"},
						},
						ForeignTables: []tablebuilder.ForeignTable{
							{
								Table:            "payment_terms",
								Schema:           "core",
								RelationshipFrom: "suppliers.payment_term_id",
								RelationshipTo:   "payment_terms.id",
								JoinType:         "left",
								Columns: []tablebuilder.ColumnDefinition{
									{Name: "name", Alias: "supplier_payment_terms", TableColumn: "payment_terms.name"},
								},
							},
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
					{
						Table:            "currencies",
						Schema:           "core",
						RelationshipFrom: "purchase_orders.currency_id",
						RelationshipTo:   "currencies.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "code", Alias: "currency_code", TableColumn: "currencies.code"},
							{Name: "symbol", Alias: "currency_symbol", TableColumn: "currencies.symbol"},
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
						Expression: "currency_code + ' ' + total_amount.toFixed(2)",
					},
					{
						Name:       "delivery_status",
						Expression: "hasValue(actual_delivery_date) ? 'delivered' : (isOverdue(expected_delivery_date) ? 'overdue' : 'pending')",
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
			"purchase_orders.order_number": {
				Name:       "purchase_orders.order_number",
				Header:     "PO Number",
				Width:      150,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"supplier_name": {
				Name:       "supplier_name",
				Header:     "Supplier",
				Width:      200,
				Type:       "string",
				Filterable: true,
			},
			"status_name": {
				Name:       "status_name",
				Header:     "Status",
				Width:      120,
				Type:       "status",
				Filterable: true,
			},
			"warehouse_name": {
				Name:       "warehouse_name",
				Header:     "Delivery Warehouse",
				Width:      180,
				Type:       "string",
				Filterable: true,
			},
			"purchase_orders.order_date": {
				Name:     "purchase_orders.order_date",
				Header:   "Order Date",
				Width:    130,
				Type:     "datetime",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02",
				},
			},
			"purchase_orders.expected_delivery_date": {
				Name:     "purchase_orders.expected_delivery_date",
				Header:   "Expected Delivery",
				Width:    150,
				Type:     "datetime",
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
				Type:   "computed",
			},
			"delivery_status": {
				Name:       "delivery_status",
				Header:     "Delivery Status",
				Width:      130,
				Type:       "computed",
				Filterable: true,
			},
			"requested_by_full_name": {
				Name:       "requested_by_full_name",
				Header:     "Requested By",
				Width:      150,
				Type:       "computed",
				Filterable: true,
			},
			"purchase_orders.id": {
				Name:   "purchase_orders.id",
				Header: "Actions",
				Width:  100,
				Type:   "uuid",
				Link: &tablebuilder.LinkConfig{
					URL:   "/procurement/purchase-orders/{purchase_orders.id}",
					Label: "View",
				},
			},
			"purchase_orders.supplier_id": {
				Name:       "purchase_orders.supplier_id",
				Header:     "Supplier",
				Width:      200,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "procurement.suppliers",
					LabelColumn: "suppliers.name",
					ValueColumn: "suppliers.id",
				},
			},
			"purchase_orders.purchase_order_status_id": {
				Name:       "purchase_orders.purchase_order_status_id",
				Header:     "Status",
				Width:      150,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "procurement.purchase_order_statuses",
					LabelColumn: "purchase_order_statuses.name",
					ValueColumn: "purchase_order_statuses.id",
				},
			},
			"purchase_orders.delivery_warehouse_id": {
				Name:       "purchase_orders.delivery_warehouse_id",
				Header:     "Warehouse",
				Width:      200,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "inventory.warehouses",
					LabelColumn: "warehouses.name",
					ValueColumn: "warehouses.id",
				},
			},
			"purchase_orders.actual_delivery_date": {
				Name:   "purchase_orders.actual_delivery_date",
				Header: "Actual Delivery",
				Width:  150,
				Type:   "datetime",
			},
			"purchase_orders.subtotal": {
				Name:   "purchase_orders.subtotal",
				Header: "Subtotal",
				Width:  100,
				Type:   "number",
			},
			"purchase_orders.tax_amount": {
				Name:   "purchase_orders.tax_amount",
				Header: "Tax Amount",
				Width:  100,
				Type:   "number",
			},
			"purchase_orders.shipping_cost": {
				Name:   "purchase_orders.shipping_cost",
				Header: "Shipping Cost",
				Width:  100,
				Type:   "number",
			},
			"purchase_orders.total_amount": {
				Name:   "purchase_orders.total_amount",
				Header: "Total Amount",
				Width:  100,
				Type:   "number",
			},
			"currency_code": {
				Name:   "currency_code",
				Header: "Currency",
				Width:  80,
				Type:   "string",
			},
			"currency_symbol": {
				Name:   "currency_symbol",
				Header: "Symbol",
				Width:  60,
				Type:   "string",
			},
			"purchase_orders.requested_by": {
				Name:   "purchase_orders.requested_by",
				Header: "Requested By ID",
				Width:  100,
				Type:   "uuid",
			},
			"purchase_orders.approved_by": {
				Name:   "purchase_orders.approved_by",
				Header: "Approved By ID",
				Width:  100,
				Type:   "uuid",
			},
			"purchase_orders.approved_date": {
				Name:   "purchase_orders.approved_date",
				Header: "Approved Date",
				Width:  150,
				Type:   "datetime",
			},
			"purchase_orders.notes": {
				Name:   "purchase_orders.notes",
				Header: "Notes",
				Width:  200,
				Type:   "string",
			},
			"purchase_orders.created_date": {
				Name:   "purchase_orders.created_date",
				Header: "Created Date",
				Width:  150,
				Type:   "datetime",
			},
			"purchase_orders.updated_date": {
				Name:   "purchase_orders.updated_date",
				Header: "Updated Date",
				Width:  150,
				Type:   "datetime",
			},
			"supplier_payment_terms": {
				Name:   "supplier_payment_terms",
				Header: "Payment Terms",
				Width:  120,
				Type:   "string",
			},
			"supplier_lead_time_days": {
				Name:   "supplier_lead_time_days",
				Header: "Lead Time Days",
				Width:  100,
				Type:   "number",
			},
			"status_description": {
				Name:   "status_description",
				Header: "Status Description",
				Width:  200,
				Type:   "string",
			},
			"warehouse_code": {
				Name:   "warehouse_code",
				Header: "Warehouse Code",
				Width:  100,
				Type:   "string",
			},
			"requested_by_username": {
				Name:   "requested_by_username",
				Header: "Requestor Username",
				Width:  150,
				Type:   "string",
			},
			"requested_by_first_name": {
				Name:   "requested_by_first_name",
				Header: "Requestor First Name",
				Width:  150,
				Type:   "string",
			},
			"requested_by_last_name": {
				Name:   "requested_by_last_name",
				Header: "Requestor Last Name",
				Width:  150,
				Type:   "string",
			},
			"approved_by_username": {
				Name:   "approved_by_username",
				Header: "Approver Username",
				Width:  150,
				Type:   "string",
			},
			"approved_by_first_name": {
				Name:   "approved_by_first_name",
				Header: "Approver First Name",
				Width:  150,
				Type:   "string",
			},
			"approved_by_last_name": {
				Name:   "approved_by_last_name",
				Header: "Approver Last Name",
				Width:  150,
				Type:   "string",
			},
			"approved_by_full_name": {
				Name:   "approved_by_full_name",
				Header: "Approved By",
				Width:  150,
				Type:   "computed",
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

var PurchaseOrderLineItemTableConfig = &tablebuilder.Config{
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
					{Name: "id", Alias: "line_item_id", TableColumn: "purchase_order_line_items.id"},
					{Name: "purchase_order_id", Alias: "line_item_purchase_order_id", TableColumn: "purchase_order_line_items.purchase_order_id"},
					{Name: "supplier_product_id", Alias: "line_item_supplier_product_id", TableColumn: "purchase_order_line_items.supplier_product_id"},
					{Name: "quantity_ordered", Alias: "line_item_quantity_ordered", TableColumn: "purchase_order_line_items.quantity_ordered"},
					{Name: "quantity_received", Alias: "line_item_quantity_received", TableColumn: "purchase_order_line_items.quantity_received"},
					{Name: "quantity_cancelled", Alias: "line_item_quantity_cancelled", TableColumn: "purchase_order_line_items.quantity_cancelled"},
					{Name: "unit_cost", Alias: "line_item_unit_cost", TableColumn: "purchase_order_line_items.unit_cost"},
					{Name: "discount", Alias: "line_item_discount", TableColumn: "purchase_order_line_items.discount"},
					{Name: "line_total", Alias: "line_item_line_total", TableColumn: "purchase_order_line_items.line_total"},
					{Name: "line_item_status_id", Alias: "line_item_line_item_status_id", TableColumn: "purchase_order_line_items.line_item_status_id"},
					{Name: "expected_delivery_date", Alias: "line_item_expected_delivery_date", TableColumn: "purchase_order_line_items.expected_delivery_date"},
					{Name: "actual_delivery_date", Alias: "line_item_actual_delivery_date", TableColumn: "purchase_order_line_items.actual_delivery_date"},
					{Name: "notes", Alias: "line_item_notes", TableColumn: "purchase_order_line_items.notes"},
					{Name: "created_by", Alias: "line_item_created_by", TableColumn: "purchase_order_line_items.created_by"},
					{Name: "created_date", Alias: "line_item_created_date", TableColumn: "purchase_order_line_items.created_date"},
					{Name: "updated_by", Alias: "line_item_updated_by", TableColumn: "purchase_order_line_items.updated_by"},
					{Name: "updated_date", Alias: "line_item_updated_date", TableColumn: "purchase_order_line_items.updated_date"},
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
						Expression: "line_item_quantity_ordered - line_item_quantity_received - line_item_quantity_cancelled",
					},
					{
						Name:       "fulfillment_percentage",
						Expression: "line_item_quantity_ordered > 0 ? ((line_item_quantity_received / line_item_quantity_ordered) * 100).toFixed(1) : '0.0'",
					},
					{
						Name:       "line_status",
						Expression: "line_item_quantity_received >= line_item_quantity_ordered ? 'complete' : line_item_quantity_cancelled > 0 ? 'partial' : 'pending'",
					},
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "line_item_created_date",
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
				Type:       "string",
				Filterable: true,
			},
			"supplier_name": {
				Name:       "supplier_name",
				Header:     "Supplier",
				Width:      180,
				Type:       "string",
				Filterable: true,
			},
			"product_name": {
				Name:       "product_name",
				Header:     "Product",
				Width:      200,
				Type:       "string",
				Filterable: true,
			},
			"product_sku": {
				Name:       "product_sku",
				Header:     "SKU",
				Width:      120,
				Type:       "string",
				Filterable: true,
			},
			"supplier_part_number": {
				Name:       "supplier_part_number",
				Header:     "Supplier Part #",
				Width:      150,
				Type:       "string",
				Filterable: true,
			},
			"line_item_quantity_ordered": {
				Name:       "line_item_quantity_ordered",
				Header:     "Qty Ordered",
				Width:      110,
				Type:       "number",
				Sortable:   true,
				Filterable: true,
			},
			"line_item_quantity_received": {
				Name:       "line_item_quantity_received",
				Header:     "Qty Received",
				Width:      120,
				Type:       "number",
				Sortable:   true,
				Filterable: true,
			},
			"quantity_pending": {
				Name:   "quantity_pending",
				Header: "Qty Pending",
				Width:  110,
				Type:   "computed",
			},
			"fulfillment_percentage": {
				Name:   "fulfillment_percentage",
				Header: "Fulfillment %",
				Width:  120,
				Type:   "computed",
			},
			"line_item_unit_cost": {
				Name:     "line_item_unit_cost",
				Header:   "Unit Cost",
				Width:    100,
				Type:     "number",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "currency",
					Format: "USD",
				},
			},
			"line_item_line_total": {
				Name:     "line_item_line_total",
				Header:   "Line Total",
				Width:    110,
				Type:     "number",
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
				Type:       "status",
				Filterable: true,
			},
			"line_item_expected_delivery_date": {
				Name:     "line_item_expected_delivery_date",
				Header:   "Expected Delivery",
				Width:    150,
				Type:     "datetime",
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
				Type:       "string",
				Filterable: true,
			},
			"line_item_id": {
				Name:   "line_item_id",
				Header: "Actions",
				Width:  100,
				Type:   "uuid",
				Link: &tablebuilder.LinkConfig{
					URL:   "/procurement/purchase-order-line-items/{line_item_id}",
					Label: "View",
				},
			},
			"line_item_purchase_order_id": {
				Name:       "line_item_purchase_order_id",
				Header:     "Purchase Order",
				Width:      150,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "procurement.purchase_orders",
					LabelColumn: "purchase_orders.order_number",
					ValueColumn: "purchase_orders.id",
				},
			},
			"line_item_supplier_product_id": {
				Name:       "line_item_supplier_product_id",
				Header:     "Supplier Product",
				Width:      200,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "procurement.supplier_products",
					LabelColumn: "supplier_products.supplier_part_number",
					ValueColumn: "supplier_products.id",
				},
			},
			"line_item_quantity_cancelled": {
				Name:   "line_item_quantity_cancelled",
				Header: "Qty Cancelled",
				Width:  110,
				Type:   "number",
			},
			"line_item_discount": {
				Name:   "line_item_discount",
				Header: "Discount",
				Width:  100,
				Type:   "number",
			},
			"line_item_line_item_status_id": {
				Name:       "line_item_line_item_status_id",
				Header:     "Status",
				Width:      150,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "procurement.purchase_order_line_item_statuses",
					LabelColumn: "purchase_order_line_item_statuses.name",
					ValueColumn: "purchase_order_line_item_statuses.id",
				},
			},
			"line_item_actual_delivery_date": {
				Name:   "line_item_actual_delivery_date",
				Header: "Actual Delivery",
				Width:  150,
				Type:   "datetime",
			},
			"line_item_notes": {
				Name:   "line_item_notes",
				Header: "Notes",
				Width:  200,
				Type:   "string",
			},
			"line_item_created_by": {
				Name:   "line_item_created_by",
				Header: "Created By ID",
				Width:  100,
				Type:   "uuid",
			},
			"line_item_created_date": {
				Name:   "line_item_created_date",
				Header: "Created Date",
				Width:  150,
				Type:   "datetime",
			},
			"line_item_updated_by": {
				Name:   "line_item_updated_by",
				Header: "Updated By ID",
				Width:  100,
				Type:   "uuid",
			},
			"line_item_updated_date": {
				Name:   "line_item_updated_date",
				Header: "Updated Date",
				Width:  150,
				Type:   "datetime",
			},
			"po_supplier_id": {
				Name:       "po_supplier_id",
				Header:     "Supplier",
				Width:      200,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "procurement.suppliers",
					LabelColumn: "suppliers.name",
					ValueColumn: "suppliers.id",
				},
			},
			"po_order_date": {
				Name:   "po_order_date",
				Header: "PO Order Date",
				Width:  150,
				Type:   "datetime",
			},
			"product_id": {
				Name:       "product_id",
				Header:     "Product",
				Width:      200,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "products.products",
					LabelColumn: "products.name",
					ValueColumn: "products.id",
				},
			},
			"line_status": {
				Name:   "line_status",
				Header: "Line Status",
				Width:  100,
				Type:   "computed",
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
var ProcurementOpenApprovalsTableConfig = &tablebuilder.Config{
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
							{Name: "lead_time_days", Alias: "supplier_lead_time_days", TableColumn: "suppliers.lead_time_days"},
						},
						ForeignTables: []tablebuilder.ForeignTable{
							{
								Table:            "payment_terms",
								Schema:           "core",
								RelationshipFrom: "suppliers.payment_term_id",
								RelationshipTo:   "payment_terms.id",
								JoinType:         "left",
								Columns: []tablebuilder.ColumnDefinition{
									{Name: "name", Alias: "supplier_payment_terms", TableColumn: "payment_terms.name"},
								},
							},
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
						Table:            "currencies",
						Schema:           "core",
						RelationshipFrom: "purchase_orders.currency_id",
						RelationshipTo:   "currencies.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "code", Alias: "currency_code", TableColumn: "currencies.code"},
							{Name: "symbol", Alias: "currency_symbol", TableColumn: "currencies.symbol"},
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
						Expression: "currency_code + ' ' + total_amount.toFixed(2)",
					},
					{
						Name:       "days_pending",
						Expression: "daysSince(order_date)",
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
			"purchase_orders.order_number": {
				Name:       "purchase_orders.order_number",
				Header:     "PO Number",
				Width:      150,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"supplier_name": {
				Name:       "supplier_name",
				Header:     "Supplier",
				Width:      200,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"status_name": {
				Name:       "status_name",
				Header:     "Status",
				Width:      120,
				Type:       "status",
				Filterable: true,
			},
			"warehouse_name": {
				Name:       "warehouse_name",
				Header:     "Delivery Warehouse",
				Width:      180,
				Type:       "string",
				Filterable: true,
			},
			"purchase_orders.order_date": {
				Name:     "purchase_orders.order_date",
				Header:   "Order Date",
				Width:    130,
				Type:     "datetime",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02",
				},
			},
			"purchase_orders.expected_delivery_date": {
				Name:     "purchase_orders.expected_delivery_date",
				Header:   "Expected Delivery",
				Width:    150,
				Type:     "datetime",
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
				Type:   "computed",
			},
			"days_pending": {
				Name:   "days_pending",
				Header: "Days Pending",
				Width:  120,
				Type:   "computed",
			},
			"requested_by_full_name": {
				Name:       "requested_by_full_name",
				Header:     "Requested By",
				Width:      150,
				Type:       "computed",
				Filterable: true,
			},
			"purchase_orders.id": {
				Name:   "purchase_orders.id",
				Header: "Actions",
				Width:  100,
				Type:   "uuid",
				Link: &tablebuilder.LinkConfig{
					URL:   "/procurement/purchase-orders/{purchase_orders.id}/approve",
					Label: "Review",
				},
			},
			"purchase_orders.supplier_id": {
				Name:       "purchase_orders.supplier_id",
				Header:     "Supplier",
				Width:      200,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "procurement.suppliers",
					LabelColumn: "suppliers.name",
					ValueColumn: "suppliers.id",
				},
			},
			"purchase_orders.purchase_order_status_id": {
				Name:       "purchase_orders.purchase_order_status_id",
				Header:     "Status",
				Width:      150,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "procurement.purchase_order_statuses",
					LabelColumn: "purchase_order_statuses.name",
					ValueColumn: "purchase_order_statuses.id",
				},
			},
			"purchase_orders.delivery_warehouse_id": {
				Name:       "purchase_orders.delivery_warehouse_id",
				Header:     "Warehouse",
				Width:      200,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "inventory.warehouses",
					LabelColumn: "warehouses.name",
					ValueColumn: "warehouses.id",
				},
			},
			"purchase_orders.subtotal": {
				Name:   "purchase_orders.subtotal",
				Header: "Subtotal",
				Width:  100,
				Type:   "number",
			},
			"purchase_orders.tax_amount": {
				Name:   "purchase_orders.tax_amount",
				Header: "Tax Amount",
				Width:  100,
				Type:   "number",
			},
			"purchase_orders.shipping_cost": {
				Name:   "purchase_orders.shipping_cost",
				Header: "Shipping Cost",
				Width:  100,
				Type:   "number",
			},
			"purchase_orders.total_amount": {
				Name:   "purchase_orders.total_amount",
				Header: "Total Amount",
				Width:  100,
				Type:   "number",
			},
			"currency_code": {
				Name:   "currency_code",
				Header: "Currency",
				Width:  80,
				Type:   "string",
			},
			"currency_symbol": {
				Name:   "currency_symbol",
				Header: "Symbol",
				Width:  60,
				Type:   "string",
			},
			"purchase_orders.requested_by": {
				Name:   "purchase_orders.requested_by",
				Header: "Requested By ID",
				Width:  100,
				Type:   "uuid",
			},
			"purchase_orders.notes": {
				Name:   "purchase_orders.notes",
				Header: "Notes",
				Width:  200,
				Type:   "string",
			},
			"purchase_orders.created_date": {
				Name:   "purchase_orders.created_date",
				Header: "Created Date",
				Width:  150,
				Type:   "datetime",
			},
			"supplier_payment_terms": {
				Name:   "supplier_payment_terms",
				Header: "Payment Terms",
				Width:  120,
				Type:   "string",
			},
			"supplier_lead_time_days": {
				Name:   "supplier_lead_time_days",
				Header: "Lead Time Days",
				Width:  100,
				Type:   "number",
			},
			"status_description": {
				Name:   "status_description",
				Header: "Status Description",
				Width:  200,
				Type:   "string",
			},
			"warehouse_code": {
				Name:   "warehouse_code",
				Header: "Warehouse Code",
				Width:  100,
				Type:   "string",
			},
			"requested_by_username": {
				Name:   "requested_by_username",
				Header: "Requestor Username",
				Width:  150,
				Type:   "string",
			},
			"requested_by_first_name": {
				Name:   "requested_by_first_name",
				Header: "Requestor First Name",
				Width:  150,
				Type:   "string",
			},
			"requested_by_last_name": {
				Name:   "requested_by_last_name",
				Header: "Requestor Last Name",
				Width:  150,
				Type:   "string",
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
var ProcurementClosedApprovalsTableConfig = &tablebuilder.Config{
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
							{Name: "lead_time_days", Alias: "supplier_lead_time_days", TableColumn: "suppliers.lead_time_days"},
						},
						ForeignTables: []tablebuilder.ForeignTable{
							{
								Table:            "payment_terms",
								Schema:           "core",
								RelationshipFrom: "suppliers.payment_term_id",
								RelationshipTo:   "payment_terms.id",
								JoinType:         "left",
								Columns: []tablebuilder.ColumnDefinition{
									{Name: "name", Alias: "supplier_payment_terms", TableColumn: "payment_terms.name"},
								},
							},
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
					{
						Table:            "currencies",
						Schema:           "core",
						RelationshipFrom: "purchase_orders.currency_id",
						RelationshipTo:   "currencies.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "code", Alias: "currency_code", TableColumn: "currencies.code"},
							{Name: "symbol", Alias: "currency_symbol", TableColumn: "currencies.symbol"},
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
						Expression: "currency_code + ' ' + total_amount.toFixed(2)",
					},
					{
						Name:       "delivery_status",
						Expression: "hasValue(actual_delivery_date) ? 'delivered' : (isOverdue(expected_delivery_date) ? 'overdue' : 'pending')",
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
			"purchase_orders.order_number": {
				Name:       "purchase_orders.order_number",
				Header:     "PO Number",
				Width:      150,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"supplier_name": {
				Name:       "supplier_name",
				Header:     "Supplier",
				Width:      200,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"status_name": {
				Name:       "status_name",
				Header:     "Status",
				Width:      120,
				Type:       "status",
				Filterable: true,
			},
			"warehouse_name": {
				Name:       "warehouse_name",
				Header:     "Delivery Warehouse",
				Width:      180,
				Type:       "string",
				Filterable: true,
			},
			"purchase_orders.order_date": {
				Name:     "purchase_orders.order_date",
				Header:   "Order Date",
				Width:    130,
				Type:     "datetime",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02",
				},
			},
			"purchase_orders.approved_date": {
				Name:     "purchase_orders.approved_date",
				Header:   "Approved Date",
				Width:    130,
				Type:     "datetime",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "2006-01-02",
				},
			},
			"purchase_orders.expected_delivery_date": {
				Name:     "purchase_orders.expected_delivery_date",
				Header:   "Expected Delivery",
				Width:    150,
				Type:     "datetime",
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
				Type:   "computed",
			},
			"delivery_status": {
				Name:       "delivery_status",
				Header:     "Delivery Status",
				Width:      130,
				Type:       "computed",
				Filterable: true,
			},
			"requested_by_full_name": {
				Name:       "requested_by_full_name",
				Header:     "Requested By",
				Width:      150,
				Type:       "computed",
				Filterable: true,
			},
			"approved_by_full_name": {
				Name:       "approved_by_full_name",
				Header:     "Approved By",
				Width:      150,
				Type:       "computed",
				Filterable: true,
			},
			"purchase_orders.id": {
				Name:   "purchase_orders.id",
				Header: "Actions",
				Width:  100,
				Type:   "uuid",
				Link: &tablebuilder.LinkConfig{
					URL:   "/procurement/purchase-orders/{purchase_orders.id}",
					Label: "View",
				},
			},
			"purchase_orders.supplier_id": {
				Name:       "purchase_orders.supplier_id",
				Header:     "Supplier",
				Width:      200,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "procurement.suppliers",
					LabelColumn: "suppliers.name",
					ValueColumn: "suppliers.id",
				},
			},
			"purchase_orders.purchase_order_status_id": {
				Name:       "purchase_orders.purchase_order_status_id",
				Header:     "Status",
				Width:      150,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "procurement.purchase_order_statuses",
					LabelColumn: "purchase_order_statuses.name",
					ValueColumn: "purchase_order_statuses.id",
				},
			},
			"purchase_orders.delivery_warehouse_id": {
				Name:       "purchase_orders.delivery_warehouse_id",
				Header:     "Warehouse",
				Width:      200,
				Type:       "lookup",
				Filterable: true,
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "inventory.warehouses",
					LabelColumn: "warehouses.name",
					ValueColumn: "warehouses.id",
				},
			},
			"purchase_orders.actual_delivery_date": {
				Name:   "purchase_orders.actual_delivery_date",
				Header: "Actual Delivery",
				Width:  150,
				Type:   "datetime",
			},
			"purchase_orders.subtotal": {
				Name:   "purchase_orders.subtotal",
				Header: "Subtotal",
				Width:  100,
				Type:   "number",
			},
			"purchase_orders.tax_amount": {
				Name:   "purchase_orders.tax_amount",
				Header: "Tax Amount",
				Width:  100,
				Type:   "number",
			},
			"purchase_orders.shipping_cost": {
				Name:   "purchase_orders.shipping_cost",
				Header: "Shipping Cost",
				Width:  100,
				Type:   "number",
			},
			"purchase_orders.total_amount": {
				Name:   "purchase_orders.total_amount",
				Header: "Total Amount",
				Width:  100,
				Type:   "number",
			},
			"currency_code": {
				Name:   "currency_code",
				Header: "Currency",
				Width:  80,
				Type:   "string",
			},
			"currency_symbol": {
				Name:   "currency_symbol",
				Header: "Symbol",
				Width:  60,
				Type:   "string",
			},
			"purchase_orders.requested_by": {
				Name:   "purchase_orders.requested_by",
				Header: "Requested By ID",
				Width:  100,
				Type:   "uuid",
			},
			"purchase_orders.approved_by": {
				Name:   "purchase_orders.approved_by",
				Header: "Approved By ID",
				Width:  100,
				Type:   "uuid",
			},
			"purchase_orders.notes": {
				Name:   "purchase_orders.notes",
				Header: "Notes",
				Width:  200,
				Type:   "string",
			},
			"purchase_orders.created_date": {
				Name:   "purchase_orders.created_date",
				Header: "Created Date",
				Width:  150,
				Type:   "datetime",
			},
			"supplier_payment_terms": {
				Name:   "supplier_payment_terms",
				Header: "Payment Terms",
				Width:  120,
				Type:   "string",
			},
			"supplier_lead_time_days": {
				Name:   "supplier_lead_time_days",
				Header: "Lead Time Days",
				Width:  100,
				Type:   "number",
			},
			"status_description": {
				Name:   "status_description",
				Header: "Status Description",
				Width:  200,
				Type:   "string",
			},
			"warehouse_code": {
				Name:   "warehouse_code",
				Header: "Warehouse Code",
				Width:  100,
				Type:   "string",
			},
			"requested_by_username": {
				Name:   "requested_by_username",
				Header: "Requestor Username",
				Width:  150,
				Type:   "string",
			},
			"requested_by_first_name": {
				Name:   "requested_by_first_name",
				Header: "Requestor First Name",
				Width:  150,
				Type:   "string",
			},
			"requested_by_last_name": {
				Name:   "requested_by_last_name",
				Header: "Requestor Last Name",
				Width:  150,
				Type:   "string",
			},
			"approved_by_username": {
				Name:   "approved_by_username",
				Header: "Approver Username",
				Width:  150,
				Type:   "string",
			},
			"approved_by_first_name": {
				Name:   "approved_by_first_name",
				Header: "Approver First Name",
				Width:  150,
				Type:   "string",
			},
			"approved_by_last_name": {
				Name:   "approved_by_last_name",
				Header: "Approver Last Name",
				Width:  150,
				Type:   "string",
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

// ProductsWithPricesLookup provides a dropdown lookup for products with their selling prices.
// Used by order entry forms to populate unit_price when product is selected.
//
// JOIN Strategy: Uses INNER JOIN to only return products with valid prices.
// Currency: Filters to USD only via currencies.code join (multi-currency support deferred).
// Price History: Sorts by name ascending for dropdown UX.
//
// KNOWN LIMITATION: May return duplicate products if multiple price records
// match the filters (same currency). Frontend should deduplicate by product_id
// if needed. Consider creating a database view with DISTINCT ON for guaranteed uniqueness.
var ProductsWithPricesLookup = &tablebuilder.Config{
	Title:           "Products with Selling Prices",
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
					{Name: "description", TableColumn: "products.description"},
					{Name: "sku", TableColumn: "products.sku"},
					{Name: "is_active", TableColumn: "products.is_active"},
				},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "product_costs",
						Schema:           "products",
						RelationshipFrom: "products.id",
						RelationshipTo:   "product_costs.product_id",
						JoinType:         "inner",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "selling_price", TableColumn: "product_costs.selling_price"},
							{Name: "purchase_cost", TableColumn: "product_costs.purchase_cost"},
						},
					},
					{
						Table:            "currencies",
						Schema:           "core",
						RelationshipFrom: "product_costs.currency_id",
						RelationshipTo:   "currencies.id",
						JoinType:         "inner",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "code", Alias: "currency_code", TableColumn: "currencies.code"},
						},
					},
				},
			},
			Filters: []tablebuilder.Filter{
				{
					Column:   "products.is_active",
					Operator: "eq",
					Value:    true,
				},
				{
					Column:   "currencies.code",
					Operator: "eq",
					Value:    "USD",
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "products.name",
					Direction: "asc",
				},
			},
			Rows: 1000,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"products.id": {
				Name:   "products.id",
				Header: "ID",
				Width:  100,
				Type:   "uuid",
			},
			"products.name": {
				Name:       "products.name",
				Header:     "Product Name",
				Width:      250,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"products.description": {
				Name:   "products.description",
				Header: "Description",
				Width:  300,
				Type:   "string",
			},
			"products.sku": {
				Name:       "products.sku",
				Header:     "SKU",
				Width:      150,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"products.is_active": {
				Name:       "products.is_active",
				Header:     "Active",
				Width:      100,
				Type:       "boolean",
				Sortable:   true,
				Filterable: true,
			},
			"product_costs.selling_price": {
				Name:       "product_costs.selling_price",
				Header:     "Selling Price",
				Width:      150,
				Type:       "number",
				Align:      "right",
				Sortable:   true,
				Filterable: true,
				Format: &tablebuilder.FormatConfig{
					Type:      "currency",
					Precision: 2,
				},
			},
			"product_costs.purchase_cost": {
				Name:       "product_costs.purchase_cost",
				Header:     "Cost Price",
				Width:      150,
				Type:       "number",
				Align:      "right",
				Sortable:   true,
				Format: &tablebuilder.FormatConfig{
					Type:      "currency",
					Precision: 2,
				},
			},
			"currency_code": {
				Name:   "currency_code",
				Header: "Currency",
				Width:  80,
				Type:   "string",
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
		Actions: []string{"view"},
	},
}
