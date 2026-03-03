package seedmodels

import (
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)



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
					{Name: "first_name", Alias: "first_name", TableColumn: "users.first_name"},
					{Name: "last_name", Alias: "last_name", TableColumn: "users.last_name"},
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
			"first_name": {
				Name:   "first_name",
				Header: "First Name",
				Width:  100,
				Type:   "string",
				Hidden: true,
			},
			"last_name": {
				Name:   "last_name",
				Header: "Last Name",
				Width:  100,
				Type:   "string",
				Hidden: true,
			},
			"full_name": {
				Name:       "full_name",
				Header:     "Name",
				Width:      200,
				Type:       "computed",
				Sortable:   true,
				Filterable: true,
				Link: &tablebuilder.LinkConfig{
					URL:         "/hr/employees/{users.id}",
					LabelColumn: "full_name",
				},
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
					Format: "yyyy-MM-dd",
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
				Header: "User ID",
				Type:   "uuid",
				Hidden: true,
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
							{Name: "postal_code", Alias: "postal_code", TableColumn: "streets.postal_code"},
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
			"postal_code": {
				Name:   "postal_code",
				Header: "Postal Code",
				Width:  100,
				Type:   "string",
				Hidden: true,
			},
			"offices_name": {
				Name:       "offices_name",
				Header:     "Office Name",
				Width:      200,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
				Link: &tablebuilder.LinkConfig{
					URL:         "/hr/offices/{offices_id}",
					LabelColumn: "offices_name",
				},
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
				Header: "Office ID",
				Type:   "uuid",
				Hidden: true,
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
