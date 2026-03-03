package seedmodels

import (
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)



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
					{Name: "reorder_point", Alias: "reorder_point", TableColumn: "inventory_items.reorder_point"},
					{Name: "maximum_stock", Alias: "maximum_stock", TableColumn: "inventory_items.maximum_stock"},
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
				Order:      10, // Display first
				Sortable:   true,
				Filterable: true,
				Link: &tablebuilder.LinkConfig{
					URL:         "/products/products/{product_id}",
					LabelColumn: "product_name",
				},
			},
			"current_quantity": {
				Name:   "current_quantity",
				Header: "Current Stock",
				Width:  120,
				Align:  "right",
				Type:   "number",
				Order:  20, // Display second
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
				Order:        30, // Display third
				CellTemplate: "status",
			},
			"stock_percentage": {
				Name:   "stock_percentage",
				Header: "Capacity",
				Width:  100,
				Align:  "right",
				Type:   "computed",
				Order:  40, // Display fourth
				Format: &tablebuilder.FormatConfig{
					Type:      "percent",
					Precision: 1,
				},
			},
			"product_id": {
				Hidden: true,
			},
			"inventory_items.id": {
				Hidden: true,
			},
			"reorder_point": {
				Name:   "reorder_point",
				Header: "Reorder Point",
				Width:  100,
				Type:   "number",
				Hidden: true,
			},
			"maximum_stock": {
				Name:   "maximum_stock",
				Header: "Maximum Stock",
				Width:  100,
				Type:   "number",
				Hidden: true,
			},
			"products.sku": {
				Name:   "products.sku",
				Header: "SKU",
				Width:  120,
				Type:   "string",
				Order:  50, // Display fifth
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
				Link: &tablebuilder.LinkConfig{
					URL:         "/inventory/warehouses/{warehouses.id}",
					LabelColumn: "warehouses.name",
				},
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
					Format: "yyyy-MM-dd HH:mm",
				},
			},
			"warehouses.id": {
				Name:   "warehouses.id",
				Header: "Warehouse ID",
				Type:   "uuid",
				Hidden: true,
			},
			"warehouses.updated_date": {
				Name:   "warehouses.updated_date",
				Header: "Updated Date",
				Width:  150,
				Type:   "datetime",
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "MM-dd-yyyy",
				},
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
							{Name: "aisle", Alias: "aisle", TableColumn: "inventory_locations.aisle"},
							{Name: "rack", Alias: "rack", TableColumn: "inventory_locations.rack"},
							{Name: "shelf", Alias: "shelf", TableColumn: "inventory_locations.shelf"},
							{Name: "bin", Alias: "bin", TableColumn: "inventory_locations.bin"},
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
			"aisle": {
				Name:   "aisle",
				Header: "Aisle",
				Width:  80,
				Type:   "string",
				Hidden: true,
			},
			"rack": {
				Name:   "rack",
				Header: "Rack",
				Width:  80,
				Type:   "string",
				Hidden: true,
			},
			"shelf": {
				Name:   "shelf",
				Header: "Shelf",
				Width:  80,
				Type:   "string",
				Hidden: true,
			},
			"bin": {
				Name:   "bin",
				Header: "Bin",
				Width:  80,
				Type:   "string",
				Hidden: true,
			},
			"product_name": {
				Name:       "product_name",
				Header:     "Product",
				Width:      200,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
				Link: &tablebuilder.LinkConfig{
					URL:         "/inventory/adjustments/{inventory_adjustments.id}",
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
					Format: "yyyy-MM-dd HH:mm",
				},
			},
			"inventory_adjustments.id": {
				Name:   "inventory_adjustments.id",
				Header: "Adjustment ID",
				Type:   "uuid",
				Hidden: true,
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
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "MM-dd-yyyy",
				},
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
				Link: &tablebuilder.LinkConfig{
					URL:         "/inventory/transfers/{transfer_orders.id}",
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
					Format: "yyyy-MM-dd HH:mm",
				},
			},
			"transfer_orders.id": {
				Name:   "transfer_orders.id",
				Header: "Transfer ID",
				Type:   "uuid",
				Hidden: true,
			},
			"transfer_orders.created_date": {
				Name:   "transfer_orders.created_date",
				Header: "Created Date",
				Width:  150,
				Type:   "datetime",
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "MM-dd-yyyy",
				},
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
