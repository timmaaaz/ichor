package seedmodels

import (
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)



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
				Type:   "uuid",
				Hidden: true,
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
					Format: "yyyy-MM-dd",
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
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "yyyy-MM-dd",
				},
			},
			"order_created_date": {
				Name:   "order_created_date",
				Header: "Created Date",
				Width:  120,
				Type:   "datetime",
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "yyyy-MM-dd",
				},
			},
			"order_updated_date": {
				Name:   "order_updated_date",
				Header: "Updated Date",
				Width:  120,
				Type:   "datetime",
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "yyyy-MM-dd",
				},
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
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "yyyy-MM-dd",
				},
			},
			"customer_updated_date": {
				Name:   "customer_updated_date",
				Header: "Customer Updated",
				Width:  120,
				Type:   "datetime",
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "yyyy-MM-dd",
				},
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
					{Name: "due_date", Alias: "due_date", TableColumn: "orders.due_date"},
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
							{Name: "name", Alias: "customer_name", TableColumn: "customers.name"},
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
			"due_date": {
				Name:     "due_date",
				Header:   "Due Date",
				Width:    120,
				Sortable: true,
				Type:     "datetime",
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "yyyy-MM-dd",
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
			"order_number": {
				Name:     "order_number",
				Header:   "Order #",
				Width:    150,
				Type:     "string",
				Sortable: true,
				Link: &tablebuilder.LinkConfig{
					URL:         "/sales/orders/{orders.id}",
					LabelColumn: "order_number",
				},
			},
			"orders.id": {
				Name:   "orders.id",
				Header: "Order ID",
				Type:   "uuid",
				Hidden: true,
			},
			"is_overdue": {
				Name:   "is_overdue",
				Header: "Overdue",
				Width:  80,
				Type:   "boolean",
			},
			"customer_name": {
				Name:     "customer_name",
				Header:   "Customer",
				Width:    200,
				Type:     "string",
				Sortable: true,
				Link: &tablebuilder.LinkConfig{
					URL:         "/sales/customers/{orders.customer_id}",
					LabelColumn: "customer_name",
				},
			},
			"orders.customer_id": {
				Name:   "orders.customer_id",
				Header: "Customer ID",
				Type:   "uuid",
				Hidden: true,
			},
			"orders.order_fulfillment_status_id": {
				Hidden: true,
			},
			"orders.created_date": {
				Name:   "orders.created_date",
				Header: "Created Date",
				Width:  120,
				Type:   "datetime",
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "MM-dd-yyyy",
				},
			},
			"orders.updated_date": {
				Name:   "orders.updated_date",
				Header: "Updated Date",
				Width:  120,
				Type:   "datetime",
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "MM-dd-yyyy",
				},
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
				Column:    "days_until_due",
				Condition: "lt",
				Value:     0,
				Color:     "red",
			},
			{
				Column:    "days_until_due",
				Condition: "lte",
				Value:     7,
				Color:     "yellow",
			},
			{
				Column:    "is_overdue",
				Condition: "eq",
				Value:     true,
				Color:     "red",
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
					{Name: "quantity", Alias: "quantity", TableColumn: "order_line_items.quantity"},
					{Name: "discount", Alias: "discount", TableColumn: "order_line_items.discount"},
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
							{Name: "unit_price", Alias: "product_unit_price", TableColumn: "products.unit_price"},
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
					{
						Table:            "users",
						Schema:           "core",
						Alias:            "updated_by_user",
						RelationshipFrom: "order_line_items.updated_by",
						RelationshipTo:   "users.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "username", Alias: "updated_by_username", TableColumn: "updated_by_user.username"},
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
				Link: &tablebuilder.LinkConfig{
					URL:         "/sales/order-line-items/{order_line_items.id}",
					LabelColumn: "product_name",
				},
			},
			"product_sku": {
				Name:       "product_sku",
				Header:     "SKU",
				Width:      120,
				Type:       "string",
				Filterable: true,
			},
			"quantity": {
				Name:     "quantity",
				Header:   "Qty",
				Width:    80,
				Align:    "center",
				Type:     "number",
				Sortable: true,
				Hidden:   true,
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
				Type:     "number",
				Sortable: true,
				Hidden:   true,
				Format: &tablebuilder.FormatConfig{
					Type:      "number",
					Precision: 2,
				},
				Editable: &tablebuilder.EditableConfig{
					Type:        "number",
					Placeholder: "0.00",
				},
			},
			"product_unit_price": {
				Name:   "product_unit_price",
				Header: "Unit Price",
				Width:  100,
				Align:  "right",
				Type:   "number",
				Hidden: true,
				Format: &tablebuilder.FormatConfig{
					Type:      "currency",
					Precision: 2,
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
					Format: "yyyy-MM-dd",
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
					Format: "yyyy-MM-dd HH:mm",
				},
			},
			"order_line_items.id": {
				Name:   "order_line_items.id",
				Header: "Line Item ID",
				Type:   "uuid",
				Hidden: true,
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
				Header: "Created By",
				Type:   "uuid",
				Hidden: true,
			},
			"order_line_items.updated_by": {
				Name:   "order_line_items.updated_by",
				Header: "Updated By",
				Type:   "uuid",
				Hidden: true,
			},
			"order_line_items.updated_date": {
				Name:   "order_line_items.updated_date",
				Header: "Updated Date",
				Width:  120,
				Type:   "datetime",
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "MM-dd-yyyy",
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
			"updated_by_username": {
				Name:       "updated_by_username",
				Header:     "Updated By",
				Width:      150,
				Type:       "string",
				Filterable: true,
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
				Link: &tablebuilder.LinkConfig{
					URL:         "/sales/customers/{customers.id}",
					LabelColumn: "customers.name",
				},
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
					Format: "yyyy-MM-dd HH:mm",
				},
			},
			"customers.id": {
				Name:   "customers.id",
				Header: "Customer ID",
				Type:   "uuid",
				Hidden: true,
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
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "MM-dd-yyyy",
				},
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
