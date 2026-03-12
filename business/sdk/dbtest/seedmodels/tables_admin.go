package seedmodels

import (
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)



// =============================================================================
// ADMIN MODULE CONFIGS
// =============================================================================

// AdminUsersTableConfig is the Users Management Page Config
var AdminUsersTableConfig = &tablebuilder.Config{
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
						Expression: "hasValue(date_hired) ? daysSince(date_hired) : nil",
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
				Type:       "computed",
				Sortable:   true,
				Filterable: true,
			},
			"users.username": {
				Name:       "users.username",
				Header:     "Username",
				Width:      150,
				Type:       "string",
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
				Width:      150,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"office_name": {
				Name:       "office_name",
				Header:     "Office",
				Width:      150,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"approval_status_name": {
				Name:       "approval_status_name",
				Header:     "Status",
				Width:      120,
				Sortable:   true,
				Filterable: true,
				Type:       "status",
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
			"users.date_hired": {
				Name:     "users.date_hired",
				Header:   "Date Hired",
				Width:    120,
				Type:     "datetime",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "yyyy-MM-dd",
				},
			},
			"days_employed": {
				Name:   "days_employed",
				Header: "Days Employed",
				Width:  120,
				Align:  "right",
				Type:   "computed",
				Format: &tablebuilder.FormatConfig{
					Type:      "number",
					Precision: 0,
				},
			},
			"users.id": {
				Name:   "users.id",
				Header: "Actions",
				Width:  100,
				Type:   "uuid",
				Link: &tablebuilder.LinkConfig{
					URL:   "/admin/users/{users.id}",
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
			"users.created_date": {
				Name:   "users.created_date",
				Header: "Created Date",
				Width:  150,
				Type:   "datetime",
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "yyyy-MM-dd",
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


// AdminRolesTableConfig is the Roles Management Page Config
var AdminRolesTableConfig = &tablebuilder.Config{
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
			"roles.name": {
				Name:       "roles.name",
				Header:     "Role Name",
				Width:      200,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"roles.description": {
				Name:       "roles.description",
				Header:     "Description",
				Width:      400,
				Type:       "string",
				Filterable: true,
			},
			"roles.id": {
				Name:   "roles.id",
				Header: "Actions",
				Width:  100,
				Type:   "uuid",
				Link: &tablebuilder.LinkConfig{
					URL:   "/admin/roles/{roles.id}",
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


// AdminTableAccessTableConfig is the Table Access (Permissions) Config
var AdminTableAccessTableConfig = &tablebuilder.Config{
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
			"table_access.id": {
				Name:   "table_access.id",
				Header: "ID",
				Type:   "uuid",
				Hidden: true,
			},
			"role_name": {
				Name:       "role_name",
				Header:     "Role",
				Width:      150,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"table_access.table_name": {
				Name:       "table_access.table_name",
				Header:     "Table",
				Width:      200,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"table_access.can_create": {
				Name:   "table_access.can_create",
				Header: "Create",
				Width:  80,
				Type:   "boolean",
				Align:  "center",
				Format: &tablebuilder.FormatConfig{
					Type: "boolean",
				},
			},
			"table_access.can_read": {
				Name:   "table_access.can_read",
				Header: "Read",
				Width:  80,
				Type:   "boolean",
				Align:  "center",
				Format: &tablebuilder.FormatConfig{
					Type: "boolean",
				},
			},
			"table_access.can_update": {
				Name:   "table_access.can_update",
				Header: "Update",
				Width:  80,
				Type:   "boolean",
				Align:  "center",
				Format: &tablebuilder.FormatConfig{
					Type: "boolean",
				},
			},
			"table_access.can_delete": {
				Name:   "table_access.can_delete",
				Header: "Delete",
				Width:  80,
				Type:   "boolean",
				Align:  "center",
				Format: &tablebuilder.FormatConfig{
					Type: "boolean",
				},
			},
		},
		ConditionalFormatting: []tablebuilder.ConditionalFormat{
			{
				Column:     "table_access.can_create",
				Condition:  "eq",
				Value:      true,
				Color:      "#2e7d32",
				Background: "#e8f5e9",
				Icon:       "check",
			},
			{
				Column:     "table_access.can_read",
				Condition:  "eq",
				Value:      true,
				Color:      "#2e7d32",
				Background: "#e8f5e9",
				Icon:       "check",
			},
			{
				Column:     "table_access.can_update",
				Condition:  "eq",
				Value:      true,
				Color:      "#2e7d32",
				Background: "#e8f5e9",
				Icon:       "check",
			},
			{
				Column:     "table_access.can_delete",
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


// AdminAuditTableConfig is the Audit Logs Page Config
var AdminAuditTableConfig = &tablebuilder.Config{
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
			"automation_executions.executed_at": {
				Name:     "automation_executions.executed_at",
				Header:   "Execution Time",
				Width:    180,
				Type:     "datetime",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "yyyy-MM-dd HH:mm:ss",
				},
			},
			"rule_name": {
				Name:       "rule_name",
				Header:     "Rule",
				Width:      200,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"automation_executions.entity_type": {
				Name:       "automation_executions.entity_type",
				Header:     "Entity Type",
				Width:      150,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"automation_executions.status": {
				Name:         "automation_executions.status",
				Header:       "Status",
				Width:        120,
				Type:         "status",
				Sortable:     true,
				Filterable:   true,
				CellTemplate: "status",
			},
			"automation_executions.execution_time_ms": {
				Name:     "automation_executions.execution_time_ms",
				Header:   "Duration (ms)",
				Width:    120,
				Type:     "number",
				Align:    "right",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:      "number",
					Precision: 0,
				},
			},
			"automation_executions.error_message": {
				Name:   "automation_executions.error_message",
				Header: "Error",
				Width:  300,
				Type:   "string",
			},
			"automation_executions.id": {
				Name:   "automation_executions.id",
				Header: "Details",
				Width:  100,
				Type:   "uuid",
				Link: &tablebuilder.LinkConfig{
					URL:   "/admin/audit/{id}",
					Label: "View",
				},
			},
		},
		ConditionalFormatting: []tablebuilder.ConditionalFormat{
			{
				Column:     "automation_executions.status",
				Condition:  "eq",
				Value:      "success",
				Color:      "#2e7d32",
				Background: "#e8f5e9",
				Icon:       "check-circle",
			},
			{
				Column:     "automation_executions.status",
				Condition:  "eq",
				Value:      "failed",
				Color:      "#c62828",
				Background: "#ffebee",
				Icon:       "x-circle",
			},
			{
				Column:     "automation_executions.status",
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


// AdminConfigTableConfig is the Table Configs Management Page Config
var AdminConfigTableConfig = &tablebuilder.Config{
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
					{Name: "id", Alias: "table_configs_id", TableColumn: "table_configs.id"},
					{Name: "name", Alias: "table_configs_name", TableColumn: "table_configs.name"},
					{Name: "description", Alias: "table_configs_description", TableColumn: "table_configs.description"},
					{Name: "created_date", Alias: "table_configs_created_date", TableColumn: "table_configs.created_date"},
					{Name: "updated_date", Alias: "table_configs_updated_date", TableColumn: "table_configs.updated_date"},
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
					Column:    "table_configs.updated_date",
					Direction: "desc",
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"table_configs_id": {
				Name:   "table_configs_id",
				Header: "Actions",
				Width:  100,
				Type:   "uuid",
				Link: &tablebuilder.LinkConfig{
					URL:   "/admin/config/{table_configs_id}",
					Label: "View",
				},
			},
			"table_configs_name": {
				Name:       "table_configs_name",
				Header:     "Config Name",
				Width:      250,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"table_configs_description": {
				Name:       "table_configs_description",
				Header:     "Description",
				Width:      400,
				Type:       "string",
				Filterable: true,
			},
			"table_configs_created_date": {
				Name:     "table_configs_created_date",
				Header:   "Created Date",
				Width:    180,
				Type:     "datetime",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "yyyy-MM-dd HH:mm",
				},
			},
			"created_by_username": {
				Name:       "created_by_username",
				Header:     "Created By",
				Width:      150,
				Type:       "string",
				Filterable: true,
			},
			"updated_by_username": {
				Name:       "updated_by_username",
				Header:     "Updated By",
				Width:      150,
				Type:       "string",
				Filterable: true,
			},
			"table_configs_updated_date": {
				Name:     "table_configs_updated_date",
				Header:   "Last Updated",
				Width:    180,
				Type:     "datetime",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "yyyy-MM-dd HH:mm",
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
