package seedmodels

import (
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)



// =============================================================================
// TABLE CONFIGS
// =============================================================================
// ProductsListTableConfig is used for the /products list page with a link to the detail page.
var ProductsListTableConfig = &tablebuilder.Config{
	Title:           "Products",
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
					{Name: "status", TableColumn: "products.status"},
					{Name: "is_active", TableColumn: "products.is_active"},
					{Name: "created_date", TableColumn: "products.created_date"},
				},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "product_categories",
						Schema:           "products",
						RelationshipFrom: "products.category_id",
						RelationshipTo:   "product_categories.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "category_name", TableColumn: "product_categories.name"},
						},
					},
					{
						Table:            "brands",
						Schema:           "products",
						RelationshipFrom: "products.brand_id",
						RelationshipTo:   "brands.id",
						JoinType:         "left",
						Columns: []tablebuilder.ColumnDefinition{
							{Name: "name", Alias: "brand_name", TableColumn: "brands.name"},
						},
					},
				},
			},
			Sort: []tablebuilder.Sort{
				{
					Column:    "products.name",
					Direction: "asc",
				},
			},
			Rows: 50,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"products.name": {
				Name:       "products.name",
				Header:     "Product Name",
				Width:      250,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
				Link: &tablebuilder.LinkConfig{
					URL:         "/products/{products.id}",
					LabelColumn: "products.name",
				},
			},
			"products.sku": {
				Name:       "products.sku",
				Header:     "SKU",
				Width:      150,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"category_name": {
				Name:       "category_name",
				Header:     "Category",
				Width:      180,
				Type:       "string",
				Sortable:   true,
				Filterable: true,
			},
			"brand_name": {
				Name:       "brand_name",
				Header:     "Brand",
				Width:      150,
				Type:       "string",
				Filterable: true,
			},
			"products.status": {
				Name:         "products.status",
				Header:       "Status",
				Width:        120,
				Type:         "status",
				Sortable:     true,
				Filterable:   true,
				CellTemplate: "status",
			},
			"products.is_active": {
				Name:       "products.is_active",
				Header:     "Active",
				Width:      80,
				Type:       "boolean",
				Align:      "center",
				Sortable:   true,
				Filterable: true,
				Format: &tablebuilder.FormatConfig{
					Type: "boolean",
				},
			},
			"products.created_date": {
				Name:     "products.created_date",
				Header:   "Created",
				Width:    150,
				Type:     "datetime",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "date",
					Format: "MM-dd-yyyy",
				},
			},
			"products.id": {
				Name:   "products.id",
				Header: "Product ID",
				Type:   "uuid",
				Hidden: true,
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
		Actions: []string{"view", "export"},
	},
}

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
				Type:   "uuid",
				Hidden: true,
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
				Link: &tablebuilder.LinkConfig{
					URL:         "/products/categories/{product_categories.id}",
					LabelColumn: "product_categories.name",
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
					Format: "yyyy-MM-dd HH:mm",
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
					Format: "yyyy-MM-dd HH:mm",
				},
			},
			"product_categories.id": {
				Name:   "product_categories.id",
				Header: "Category ID",
				Type:   "uuid",
				Hidden: true,
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
				Type:   "uuid",
				Hidden: true,
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
				Name:     "product_costs.purchase_cost",
				Header:   "Cost Price",
				Width:    150,
				Type:     "number",
				Align:    "right",
				Sortable: true,
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
