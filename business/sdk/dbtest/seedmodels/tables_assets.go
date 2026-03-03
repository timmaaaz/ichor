package seedmodels

import (
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)



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
				Link: &tablebuilder.LinkConfig{
					URL:         "/assets/list/{assets.id}",
					LabelColumn: "asset_name",
				},
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
					Format: "yyyy-MM-dd",
				},
			},
			"assets.id": {
				Name:   "assets.id",
				Header: "Asset ID",
				Type:   "uuid",
				Hidden: true,
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
				Link: &tablebuilder.LinkConfig{
					URL:         "/assets/requests/{user_assets.id}",
					LabelColumn: "asset_name",
				},
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
					Format: "yyyy-MM-dd",
				},
			},
			"user_assets.id": {
				Name:   "user_assets.id",
				Header: "User Asset ID",
				Type:   "uuid",
				Hidden: true,
			},
			"user_assets.last_maintenance": {
				Name:   "user_assets.last_maintenance",
				Header: "Last Maintenance",
				Width:  150,
				Type:   "datetime",
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "MM-dd-yyyy",
				},
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
