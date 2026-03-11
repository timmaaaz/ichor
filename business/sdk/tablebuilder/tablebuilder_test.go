package tablebuilder_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
	"github.com/timmaaaz/ichor/business/domain/sales/lineitemfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
	"github.com/timmaaaz/ichor/foundation/logger"
)

/*
TODOs
- add a check that any entry in "ids" is a legitimate table name, easy to miss especially with views
- figure out aliases on views with doubled up names i.e. multiple "street_id" or something
	- We'll need this for links
*/

func Test_TableBuilder(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_TableBuilder")
	log := logger.New(io.Discard, logger.LevelInfo, "ADMIN", func(context.Context) string { return "00000000-0000-0000-0000-000000000000" })

	configStore := tablebuilder.NewConfigStore(log, db.DB)
	store := tablebuilder.NewStore(log, db.DB)

	sd, err := insertSeedData(db.BusDomain, configStore)
	if err != nil {
		t.Fatalf("seed data: %v", err)
	}

	t.Run("pagination_correctness", func(t *testing.T) {
		t.Parallel()
		// productsList config has 20 seeded products, no filter, page size 10.
		pg1 := page.MustParse("1", "10")
		result1, err := store.QueryByPage(testCtx(t), productsList, pg1)
		if err != nil {
			t.Fatalf("QueryByPage page1: %v", err)
		}
		if len(result1.Data) != 10 {
			t.Errorf("page1 rows = %d, want 10", len(result1.Data))
		}
		if result1.Meta.Total != 20 {
			t.Errorf("Total = %d, want 20", result1.Meta.Total)
		}
		if result1.Meta.TotalPages != 2 {
			t.Errorf("TotalPages = %d, want 2", result1.Meta.TotalPages)
		}

		pg2 := page.MustParse("2", "10")
		result2, err := store.QueryByPage(testCtx(t), productsList, pg2)
		if err != nil {
			t.Fatalf("QueryByPage page2: %v", err)
		}
		if len(result2.Data) != 10 {
			t.Errorf("page2 rows = %d, want 10", len(result2.Data))
		}

		// No ID should appear on both pages
		page1IDs := make(map[string]bool)
		for _, row := range result1.Data {
			if id, ok := row["products.id"]; ok {
				page1IDs[fmt.Sprintf("%v", id)] = true
			}
		}
		for _, row := range result2.Data {
			if id, ok := row["products.id"]; ok {
				idStr := fmt.Sprintf("%v", id)
				if page1IDs[idStr] {
					t.Errorf("ID %s appears on both page 1 and page 2", idStr)
				}
			}
		}
	})

	t.Run("simple_inventory_items", func(t *testing.T) {
		t.Parallel()
		params := tablebuilder.QueryParams{Page: 1, Rows: 50}
		result, err := store.FetchTableData(testCtx(t), inventoryItems, params)
		if err != nil {
			t.Fatalf("FetchTableData inventory items: %v", err)
		}

		// Filter is quantity > 0; we may get fewer than 30 rows
		if len(result.Data) == 0 {
			t.Error("expected at least 1 inventory item row, got 0")
		}

		// Every row must have the expected fields
		requiredFields := []string{"inventory_items.id", "current_stock"}
		for i, row := range result.Data {
			for _, field := range requiredFields {
				if _, ok := row[field]; !ok {
					t.Errorf("row %d missing field %q", i, field)
				}
			}
		}

		// Rows sorted DESC by quantity: each row's current_stock <= previous
		for i := 1; i < len(result.Data); i++ {
			prev := toFloat(result.Data[i-1]["current_stock"])
			curr := toFloat(result.Data[i]["current_stock"])
			if curr > prev {
				t.Errorf("rows not sorted desc by quantity: row[%d]=%v > row[%d]=%v", i, curr, i-1, prev)
			}
		}
	})

	t.Run("orders_view", func(t *testing.T) {
		t.Parallel()
		// currentOrders uses sales.orders_base view (verified present in migrate.sql)
		params := tablebuilder.QueryParams{Page: 1, Rows: 50}
		result, err := store.FetchTableData(testCtx(t), currentOrders, params)
		if err != nil {
			t.Fatalf("FetchTableData orders: %v", err)
		}
		if len(result.Data) != 5 {
			t.Errorf("rows = %d, want 5 (seeded orders)", len(result.Data))
		}

		requiredFields := []string{"orders.id", "order_number"}
		for i, row := range result.Data {
			for _, field := range requiredFields {
				if _, ok := row[field]; !ok {
					t.Errorf("row %d missing field %q", i, field)
				}
			}
		}
	})

	t.Run("inventory_with_joins_computed_columns", func(t *testing.T) {
		t.Parallel()
		params := tablebuilder.QueryParams{Page: 1, Rows: 50}
		result, err := store.FetchTableData(testCtx(t), currentInventoryProducts, params)
		if err != nil {
			t.Fatalf("FetchTableData complex: %v", err)
		}
		if len(result.Data) == 0 {
			t.Fatal("expected rows from complex join query, got 0")
		}

		for i, row := range result.Data {
			// Foreign table join produced product_name
			if _, ok := row["product_name"]; !ok {
				t.Errorf("row %d missing product_name (join failed)", i)
			}
			// ClientComputedColumns (JS expressions) are evaluated client-side;
			// the server includes the column in metadata but its value is nil.
			// Just verify the key exists in the row.
			if _, ok := row["stock_status"]; !ok {
				t.Errorf("row %d missing stock_status key (column absent from row)", i)
			}
		}
	})

	t.Run("inventory_adjustments_deep_join", func(t *testing.T) {
		t.Parallel()
		// inventoryAdjustmentsPageConfig has 3-level nested join:
		// inventory_adjustments → inventory_locations → warehouses
		// and a location_code computed column.
		params := tablebuilder.QueryParams{Page: 1, Rows: 50}
		result, err := store.FetchTableData(testCtx(t), inventoryAdjustmentsPageConfig, params)
		if err != nil {
			t.Fatalf("FetchTableData adjustments: %v", err)
		}

		// Log actual columns to aid debugging when column names don't match.
		t.Logf("actual result columns: %v", result.Meta.Columns)

		// Verify expected columns present in metadata
		expectedCols := []string{
			"product_name",
			"product_sku",
			"warehouse_name",
			"location_code",
			"inventory_adjustments.quantity_change",
			"inventory_adjustments.reason_code",
			"adjusted_by_username",
			"approved_by_username",
			"inventory_adjustments.adjustment_date",
		}
		colSet := make(map[string]bool)
		for _, col := range result.Meta.Columns {
			colSet[col.Field] = true
		}
		for _, expected := range expectedCols {
			if !colSet[expected] {
				t.Errorf("column %q not found in result metadata", expected)
			}
		}
	})

	t.Run("stored_config_roundtrip", func(t *testing.T) {
		t.Parallel()
		// One of the seeded configs is "products_list"
		loaded, err := configStore.LoadConfigByName(testCtx(t), "products_list")
		if err != nil {
			t.Fatalf("LoadConfigByName: %v", err)
		}
		if loaded.Title != productsList.Title {
			t.Errorf("loaded Title = %q, want %q", loaded.Title, productsList.Title)
		}

		// Fetch data using the loaded config
		params := tablebuilder.QueryParams{Page: 1, Rows: 10}
		result, err := store.FetchTableData(testCtx(t), loaded, params)
		if err != nil {
			t.Fatalf("FetchTableData with stored config: %v", err)
		}
		if len(result.Data) == 0 {
			t.Error("expected rows from stored config query, got 0")
		}
	})

	t.Run("configstore_crud", func(t *testing.T) {
		t.Parallel()
		testCfg := &tablebuilder.Config{
			Title:         "CRUD Test Config",
			WidgetType:    "table",
			Visualization: "table",
			DataSource: []tablebuilder.DataSource{{Source: "products", Schema: "products", Select: tablebuilder.SelectConfig{Columns: []tablebuilder.ColumnDefinition{{Name: "id", TableColumn: "products.id"}}}}},
			VisualSettings: tablebuilder.VisualSettings{Columns: map[string]tablebuilder.ColumnConfig{"products.id": {Type: "uuid"}}},
		}

		created, err := configStore.Create(testCtx(t), "crud_test", "CRUD test desc", testCfg, sd.Admins[0].ID)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		// Use cmpopts.IgnoreFields to skip time.Time fields (CreatedDate, UpdatedDate)
		// which include monotonic clock readings that cause false cmp.Diff mismatches.
		ignoreTime := cmpopts.IgnoreFields(tablebuilder.StoredConfig{}, "CreatedDate", "UpdatedDate")

		// QueryByID
		byID, err := configStore.QueryByID(testCtx(t), created.ID)
		if err != nil {
			t.Fatalf("QueryByID: %v", err)
		}
		// NormalizeJSONFields aligns json.RawMessage bytes between got/exp so cmp.Diff
		// won't flag key-order differences introduced by Postgres round-trip.
		dbtest.NormalizeJSONFields(byID, created)
		if diff := cmp.Diff(created, byID, ignoreTime); diff != "" {
			t.Errorf("QueryByID mismatch (-want +got):\n%s", diff)
		}

		// QueryByName
		byName, err := configStore.QueryByName(testCtx(t), "crud_test")
		if err != nil {
			t.Fatalf("QueryByName: %v", err)
		}
		dbtest.NormalizeJSONFields(byName, created)
		if diff := cmp.Diff(created, byName, ignoreTime); diff != "" {
			t.Errorf("QueryByName mismatch (-want +got):\n%s", diff)
		}

		// Update
		testCfg.Title = "Updated CRUD Config"
		updated, err := configStore.Update(testCtx(t), created.ID, "crud_test", "Updated desc", testCfg, sd.Admins[0].ID)
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
		if updated.Description != "Updated desc" {
			t.Errorf("Description = %q, want %q", updated.Description, "Updated desc")
		}

		// Delete
		if err := configStore.Delete(testCtx(t), created.ID); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		_, err = configStore.QueryByID(testCtx(t), created.ID)
		if err == nil {
			t.Error("expected error after Delete, got nil")
		}
	})

	t.Run("page_config_crud", func(t *testing.T) {
		t.Parallel()
		pageConfig := tablebuilder.PageConfig{
			Name:      "Test Page Config",
			UserID:    sd.Admins[0].ID,
			IsDefault: true,
		}

		saved, err := configStore.CreatePageConfig(testCtx(t), pageConfig)
		if err != nil {
			t.Fatalf("CreatePageConfig: %v", err)
		}

		byID, err := configStore.QueryPageByID(testCtx(t), saved.ID)
		if err != nil {
			t.Fatalf("QueryPageByID: %v", err)
		}
		if diff := cmp.Diff(saved, byID); diff != "" {
			t.Errorf("QueryPageByID mismatch (-want +got):\n%s", diff)
		}

		byName, err := configStore.QueryPageByName(testCtx(t), "Test Page Config")
		if err != nil {
			t.Fatalf("QueryPageByName: %v", err)
		}
		if diff := cmp.Diff(saved, byName); diff != "" {
			t.Errorf("QueryPageByName mismatch (-want +got):\n%s", diff)
		}

		saved.Name = "Updated Page Config"
		updated, err := configStore.UpdatePageConfig(testCtx(t), *saved)
		if err != nil {
			t.Fatalf("UpdatePageConfig: %v", err)
		}
		if updated.Name != "Updated Page Config" {
			t.Errorf("Name = %q, want %q", updated.Name, "Updated Page Config")
		}

		if err := configStore.DeletePageConfig(testCtx(t), saved.ID); err != nil {
			t.Fatalf("DeletePageConfig: %v", err)
		}
	})

	t.Run("dynamic_filters", func(t *testing.T) {
		t.Parallel()
		// Unfiltered: get all products
		paramsAll := tablebuilder.QueryParams{Page: 1, Rows: 50}
		resultAll, err := store.FetchTableData(testCtx(t), productsList, paramsAll)
		if err != nil {
			t.Fatalf("FetchTableData unfiltered: %v", err)
		}

		// Filtered: pass a dynamic filter via QueryParams.Filters
		paramsFiltered := tablebuilder.QueryParams{
			Page: 1,
			Rows: 50,
			Filters: []tablebuilder.Filter{
				{Column: "products.is_active", Operator: "eq", Value: true},
			},
		}
		resultFiltered, err := store.FetchTableData(testCtx(t), productsList, paramsFiltered)
		if err != nil {
			t.Fatalf("FetchTableData filtered: %v", err)
		}

		// Filtered ≤ unfiltered (filter reduces or keeps same count)
		if len(resultFiltered.Data) > len(resultAll.Data) {
			t.Errorf("filtered count %d > unfiltered count %d", len(resultFiltered.Data), len(resultAll.Data))
		}
	})
}

// testCtx returns a background context. Named testCtx to avoid shadowing any
// existing ctx variable in the package.
func testCtx(t *testing.T) context.Context {
	t.Helper()
	return context.Background()
}

// toFloat converts a TableRow value to float64 for numeric comparisons.
func toFloat(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	case int:
		return float64(val)
	}
	return 0
}

var productsList = &tablebuilder.Config{
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
			"products.id":        {Type: "uuid"},
			"products.name":      {Type: "string"},
			"products.sku":       {Type: "string"},
			"products.is_active": {Type: "boolean"},
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

var currentOrders = &tablebuilder.Config{
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
			// orders table columns (key = TableColumn when no Alias)
			"orders.id":                   {Type: "uuid"},
			"order_number":                {Type: "string", Name: "order_number", Header: "Order #", Width: 150, Sortable: true, Filterable: true},
			"order_date":                  {Type: "datetime", Name: "order_date", Header: "Order Date", Width: 120, Format: &tablebuilder.FormatConfig{Type: "date", Format: "yyyy-MM-dd"}},
			"order_due_date":              {Type: "datetime", Format: &tablebuilder.FormatConfig{Type: "date", Format: "yyyy-MM-dd"}},
			"order_created_date":          {Type: "datetime", Format: &tablebuilder.FormatConfig{Type: "date", Format: "yyyy-MM-dd"}},
			"order_updated_date":          {Type: "datetime", Format: &tablebuilder.FormatConfig{Type: "date", Format: "yyyy-MM-dd"}},
			"order_fulfillment_status_id": {Type: "uuid"},
			"order_customer_id":           {Type: "uuid"},
			// customers table columns
			"customer_id":                  {Type: "uuid"},
			"customer_contact_info_id":     {Type: "uuid"},
			"customer_delivery_address_id": {Type: "uuid"},
			"customer_notes":               {Type: "string"},
			"customer_created_date":        {Type: "datetime", Format: &tablebuilder.FormatConfig{Type: "date", Format: "yyyy-MM-dd"}},
			"customer_updated_date":        {Type: "datetime", Format: &tablebuilder.FormatConfig{Type: "date", Format: "yyyy-MM-dd"}},
			// order_fulfillment_statuses columns
			"fulfillment_status_name":        {Type: "string", Name: "fulfillment_status_name", Header: "Status", Width: 120},
			"fulfillment_status_description": {Type: "string"},
		},
		ConditionalFormatting: []tablebuilder.ConditionalFormat{},
	},
	Permissions: tablebuilder.Permissions{
		Roles:   []string{"admin", "sales"},
		Actions: []string{"view", "export"},
	},
}

var inventoryItems = &tablebuilder.Config{
	Title:           "Inventory Items",
	WidgetType:      "table",
	Visualization:   "table",
	PositionX:       0,
	PositionY:       0,
	Width:           12,
	Height:          6,
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
					{Name: "quantity", Alias: "current_stock", TableColumn: "inventory_items.quantity"},
					{Name: "product_id", TableColumn: "inventory_items.product_id"},
					{Name: "location_id", TableColumn: "inventory_items.location_id"},
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
					Direction: "desc",
				},
			},
			Rows: 10,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"inventory_items.id": {Type: "uuid"},
			"current_stock": {
				Type:       "number",
				Name:       "current_stock",
				Header:     "Current Stock",
				Width:      120,
				Align:      "right",
				Sortable:   true,
				Filterable: true,
				Format: &tablebuilder.FormatConfig{
					Type:      "number",
					Precision: 0,
				},
			},
			"inventory_items.product_id": {
				Type:       "lookup",
				Name:       "inventory_items.product_id",
				Header:     "Product",
				Width:      200,
				Filterable: true,
				Link: &tablebuilder.LinkConfig{
					URL:   "/products/{inventory_items.product_id}",
					Label: "View Product",
				},
				Lookup: &tablebuilder.LookupConfig{
					Entity:      "products.products",
					LabelColumn: "products.name",
					ValueColumn: "products.id",
				},
			},
			"inventory_items.location_id": {
				Type:       "lookup",
				Name:       "inventory_items.location_id",
				Header:     "Location",
				Width:      200,
				Filterable: true,
				Link: &tablebuilder.LinkConfig{
					URL:   "/inventory/locations/{inventory_items.location_id}",
					Label: "View Location",
				},
				Lookup: &tablebuilder.LookupConfig{
					Entity:         "inventory.inventory_locations",
					LabelColumn:    "inventory_locations.aisle",
					ValueColumn:    "inventory_locations.id",
					DisplayColumns: []string{"inventory_locations.rack", "inventory_locations.shelf", "inventory_locations.bin"},
				},
			},
		},
		ConditionalFormatting: []tablebuilder.ConditionalFormat{},
	},
	Permissions: tablebuilder.Permissions{
		Roles:   []string{"admin", "inventory_manager"},
		Actions: []string{"view", "export"},
	},
}

var currentInventoryProducts = &tablebuilder.Config{
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
			// inventory_items columns
			"inventory_items.id":            {Type: "uuid"},
			"inventory_items.reorder_point": {Type: "number"},
			"inventory_items.maximum_stock": {Type: "number"},
			"product_name": {
				Type:       "string",
				Name:       "product_name",
				Header:     "Product",
				Width:      250,
				Sortable:   true,
				Filterable: true,
			},
			"current_quantity": {
				Type:   "number",
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
				Type:         "computed",
				Name:         "stock_status",
				Header:       "Status",
				Width:        100,
				Align:        "center",
				CellTemplate: "status",
			},
			"stock_percentage": {
				Type:   "computed",
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
				Type:   "uuid",
				Name:   "product_id",
				Header: "Product",
				Width:  200,
				Link: &tablebuilder.LinkConfig{
					URL:   "/products/products/{product_id}",
					Label: "View Product",
				},
			},
			"products.sku": {Type: "string"},
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
			// inventory_adjustments columns
			"inventory_adjustments.notes":        {Type: "string"},
			"inventory_adjustments.created_date": {Type: "datetime", Format: &tablebuilder.FormatConfig{Type: "date", Format: "yyyy-MM-dd"}},
			// foreign table columns
			"inventory_locations.aisle": {Type: "string"},
			"inventory_locations.rack":  {Type: "string"},
			"inventory_locations.shelf": {Type: "string"},
			"inventory_locations.bin":   {Type: "string"},
			"product_name": {
				Type:       "string",
				Name:       "product_name",
				Header:     "Product",
				Width:      200,
				Sortable:   true,
				Filterable: true,
			},
			"product_sku": {
				Type:       "string",
				Name:       "product_sku",
				Header:     "SKU",
				Width:      120,
				Filterable: true,
			},
			"warehouse_name": {
				Type:       "string",
				Name:       "warehouse_name",
				Header:     "Warehouse",
				Width:      150,
				Sortable:   true,
				Filterable: true,
			},
			"location_code": {
				Type:       "computed",
				Name:       "location_code",
				Header:     "Location",
				Width:      150,
				Filterable: true,
			},
			"inventory_adjustments.quantity_change": {
				Type:     "number",
				Name:     "inventory_adjustments.quantity_change",
				Header:   "Qty Change",
				Width:    100,
				Align:    "right",
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:      "number",
					Precision: 0,
				},
			},
			"inventory_adjustments.reason_code": {
				Type:       "string",
				Name:       "inventory_adjustments.reason_code",
				Header:     "Reason",
				Width:      120,
				Filterable: true,
			},
			"adjusted_by_username": {
				Type:       "string",
				Name:       "adjusted_by_username",
				Header:     "Adjusted By",
				Width:      130,
				Filterable: true,
			},
			"approved_by_username": {
				Type:       "string",
				Name:       "approved_by_username",
				Header:     "Approved By",
				Width:      130,
				Filterable: true,
			},
			"inventory_adjustments.adjustment_date": {
				Type:     "datetime",
				Name:     "inventory_adjustments.adjustment_date",
				Header:   "Date",
				Width:    150,
				Sortable: true,
				Format: &tablebuilder.FormatConfig{
					Type:   "datetime",
					Format: "yyyy-MM-dd HH:mm",
				},
			},
			"inventory_adjustments.id": {
				Type:   "uuid",
				Name:   "inventory_adjustments.id",
				Header: "Actions",
				Width:  100,
				Link: &tablebuilder.LinkConfig{
					URL:   "/inventory/adjustments/{inventory_adjustments.id}",
					Label: "View",
				},
			},
		},
		ConditionalFormatting: []tablebuilder.ConditionalFormat{
			{
				Column:     "inventory_adjustments.quantity_change",
				Condition:  "lt",
				Value:      0,
				Color:      "#c62828",
				Background: "#ffebee",
				Icon:       "trending-down",
			},
			{
				Column:     "inventory_adjustments.quantity_change",
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

func insertSeedData(busDomain dbtest.BusDomain, configStore *tablebuilder.ConfigStore) (unitest.SeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding user : %w", err)
	}
	userIDs := make([]uuid.UUID, 0, len(admins))
	for _, a := range admins {
		userIDs = append(userIDs, a.ID)
	}

	count := 5

	// ADDRESSES
	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("querying regions : %w", err)
	}
	ids := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		ids = append(ids, r.ID)
	}

	ctys, err := citybus.TestSeedCities(ctx, count, ids, busDomain.City)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding cities : %w", err)
	}

	ctyIDs := make([]uuid.UUID, 0, len(ctys))
	for _, c := range ctys {
		ctyIDs = append(ctyIDs, c.ID)
	}

	strs, err := streetbus.TestSeedStreets(ctx, count, ctyIDs, busDomain.Street)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding streets : %w", err)
	}
	strIDs := make([]uuid.UUID, 0, len(strs))
	for _, s := range strs {
		strIDs = append(strIDs, s.ID)
	}

	// Query timezones from seed data
	tzs, err := busDomain.Timezone.QueryAll(ctx)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("querying timezones : %w", err)
	}
	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, count, strIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding contact info : %w", err)
	}
	contactInfoIDs := make([]uuid.UUID, 0, len(contactInfos))
	for _, ci := range contactInfos {
		contactInfoIDs = append(contactInfoIDs, ci.ID)
	}

	customers, err := customersbus.TestSeedCustomers(ctx, count, strIDs, contactInfoIDs, uuid.UUIDs{admins[0].ID}, busDomain.Customers)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding customers : %w", err)
	}
	customerIDs := make([]uuid.UUID, 0, len(customers))
	for _, c := range customers {
		customerIDs = append(customerIDs, c.ID)
	}

	ofls, err := orderfulfillmentstatusbus.TestSeedOrderFulfillmentStatuses(ctx, busDomain.OrderFulfillmentStatus)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding order fulfillment statuses: %w", err)
	}
	oflIDs := make([]uuid.UUID, 0, len(ofls))
	for _, ofl := range ofls {
		oflIDs = append(oflIDs, ofl.ID)
	}

	currencies, err := currencybus.TestSeedCurrencies(ctx, 5, busDomain.Currency)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding currencies: %w", err)
	}
	currencyIDs := make(uuid.UUIDs, len(currencies))
	for i, c := range currencies {
		currencyIDs[i] = c.ID
	}

	orders, err := ordersbus.TestSeedOrders(ctx, count, uuid.UUIDs{admins[0].ID}, customerIDs, oflIDs, currencyIDs, busDomain.Order)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding Orders: %w", err)
	}
	orderIDs := make([]uuid.UUID, 0, len(orders))
	for _, o := range orders {
		orderIDs = append(orderIDs, o.ID)
	}

	contactIDs := make(uuid.UUIDs, len(contactInfos))
	for i, c := range contactInfos {
		contactIDs[i] = c.ID
	}

	brand, err := brandbus.TestSeedBrands(ctx, 5, contactIDs, busDomain.Brand)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding brand : %w", err)
	}

	brandIDs := make(uuid.UUIDs, len(brand))
	for i, b := range brand {
		brandIDs[i] = b.BrandID
	}

	productCategories, err := productcategorybus.TestSeedProductCategories(ctx, 10, busDomain.ProductCategory)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding product category : %w", err)
	}

	productCategoryIDs := make(uuid.UUIDs, len(productCategories))

	for i, pc := range productCategories {
		productCategoryIDs[i] = pc.ProductCategoryID
	}

	products, err := productbus.TestSeedProducts(ctx, 20, brandIDs, productCategoryIDs, busDomain.Product)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding product : %w", err)
	}
	productIDs := make([]uuid.UUID, 0, len(products))
	for _, p := range products {
		productIDs = append(productIDs, p.ProductID)
	}

	olStatuses, err := lineitemfulfillmentstatusbus.TestSeedLineItemFulfillmentStatuses(ctx, busDomain.LineItemFulfillmentStatus)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding line item fulfillment statuses: %w", err)
	}
	olStatusIDs := make([]uuid.UUID, 0, len(olStatuses))
	for _, ols := range olStatuses {
		olStatusIDs = append(olStatusIDs, ols.ID)
	}

	ols, err := orderlineitemsbus.TestSeedOrderLineItems(ctx, count, orderIDs, productIDs, olStatusIDs, userIDs, busDomain.OrderLineItem)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding Order Line Items: %w", err)
	}

	// WAREHOUSES
	warehouses, err := warehousebus.TestSeedWarehouses(ctx, count, admins[0].ID, strIDs, busDomain.Warehouse)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding warehouses : %w", err)
	}

	warehouseIDs := make(uuid.UUIDs, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	zones, err := zonebus.TestSeedZone(ctx, 12, warehouseIDs, busDomain.Zones)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding zones : %w", err)
	}

	zoneIDs := make([]uuid.UUID, len(zones))
	for i, z := range zones {
		zoneIDs[i] = z.ZoneID
	}

	inventoryLocations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 25, warehouseIDs, zoneIDs, busDomain.InventoryLocation)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding inventory locations : %w", err)
	}

	inventoryLocationsIDs := make([]uuid.UUID, len(inventoryLocations))
	for i, il := range inventoryLocations {
		inventoryLocationsIDs[i] = il.LocationID
	}

	_, err = inventoryitembus.TestSeedInventoryItems(ctx, 30, inventoryLocationsIDs, productIDs, busDomain.InventoryItem)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding inventory products : %w", err)
	}

	// SEED CONFIGS
	cfg1, err := configStore.Create(ctx, "products_list", "Products List", productsList, admins[0].ID)
	if err != nil {
		return unitest.SeedData{}, err
	}
	cfg2, err := configStore.Create(ctx, "current_orders", "Current Orders", currentOrders, admins[0].ID)
	if err != nil {
		return unitest.SeedData{}, err
	}
	cfg3, err := configStore.Create(ctx, "inventory_items", "Inventory Items", inventoryItems, admins[0].ID)
	if err != nil {
		return unitest.SeedData{}, err
	}

	storedConfigs := []tablebuilder.StoredConfig{*cfg1, *cfg2, *cfg3}

	return unitest.SeedData{
		Admins:                   []unitest.User{{User: admins[0]}},
		Orders:                   orders,
		Products:                 products,
		OrderFulfillmentStatuses: ofls,
		OrderLineItems:           ols,
		Customers:                customers,
		TableBuilderConfigs:      storedConfigs,
	}, nil
}


// =============================================================================
// Chart Type Tests
// =============================================================================

func Test_ChartTypes(t *testing.T) {
	t.Parallel()

	t.Run("chart type constants", func(t *testing.T) {
		t.Parallel()
		// Verify all chart type constants are defined
		chartTypes := []string{
			tablebuilder.ChartTypeLine,
			tablebuilder.ChartTypeBar,
			tablebuilder.ChartTypeStackedBar,
			tablebuilder.ChartTypeStackedArea,
			tablebuilder.ChartTypeCombo,
			tablebuilder.ChartTypeKPI,
			tablebuilder.ChartTypeGauge,
			tablebuilder.ChartTypePie,
			tablebuilder.ChartTypeWaterfall,
			tablebuilder.ChartTypeFunnel,
			tablebuilder.ChartTypeHeatmap,
			tablebuilder.ChartTypeTreemap,
			tablebuilder.ChartTypeGantt,
		}

		expected := []string{
			"line", "bar", "stacked-bar", "stacked-area", "combo",
			"kpi", "gauge", "pie", "waterfall", "funnel",
			"heatmap", "treemap", "gantt",
		}

		for i, ct := range chartTypes {
			if ct != expected[i] {
				t.Errorf("chart type mismatch: got %s, want %s", ct, expected[i])
			}
		}
	})

	t.Run("chart response serialization", func(t *testing.T) {
		t.Parallel()
		// Create a sample KPI chart response
		kpiResponse := tablebuilder.ChartResponse{
			Type:  tablebuilder.ChartTypeKPI,
			Title: "Total Revenue",
			KPI: &tablebuilder.KPIData{
				Value:         125000.50,
				PreviousValue: 100000.00,
				Change:        25.0,
				Trend:         "up",
				Label:         "Revenue",
				Format:        "currency",
			},
			Meta: tablebuilder.ChartMeta{
				ExecutionTime: 45,
				RowsProcessed: 1,
			},
		}

		jsonBytes, err := json.Marshal(kpiResponse)
		if err != nil {
			t.Fatalf("failed to marshal KPI response: %v", err)
		}

		var unmarshaled tablebuilder.ChartResponse
		if err := json.Unmarshal(jsonBytes, &unmarshaled); err != nil {
			t.Fatalf("failed to unmarshal KPI response: %v", err)
		}

		if unmarshaled.Type != tablebuilder.ChartTypeKPI {
			t.Errorf("type mismatch: got %s, want %s", unmarshaled.Type, tablebuilder.ChartTypeKPI)
		}
		if unmarshaled.KPI == nil {
			t.Fatal("KPI data is nil after unmarshaling")
		}
		if unmarshaled.KPI.Value != 125000.50 {
			t.Errorf("KPI value mismatch: got %f, want %f", unmarshaled.KPI.Value, 125000.50)
		}
	})

	t.Run("categorical chart response", func(t *testing.T) {
		t.Parallel()
		// Create a sample line chart response
		lineResponse := tablebuilder.ChartResponse{
			Type:       tablebuilder.ChartTypeLine,
			Title:      "Monthly Revenue",
			Categories: []string{"Jan", "Feb", "Mar", "Apr"},
			Series: []tablebuilder.SeriesData{
				{
					Name: "Revenue",
					Data: []float64{10000, 12000, 15000, 18000},
				},
				{
					Name: "Costs",
					Data: []float64{8000, 9000, 10000, 11000},
				},
			},
			Meta: tablebuilder.ChartMeta{
				ExecutionTime: 120,
				RowsProcessed: 4,
			},
		}

		jsonBytes, err := json.Marshal(lineResponse)
		if err != nil {
			t.Fatalf("failed to marshal line response: %v", err)
		}

		var unmarshaled tablebuilder.ChartResponse
		if err := json.Unmarshal(jsonBytes, &unmarshaled); err != nil {
			t.Fatalf("failed to unmarshal line response: %v", err)
		}

		if len(unmarshaled.Categories) != 4 {
			t.Errorf("categories length mismatch: got %d, want 4", len(unmarshaled.Categories))
		}
		if len(unmarshaled.Series) != 2 {
			t.Errorf("series length mismatch: got %d, want 2", len(unmarshaled.Series))
		}
	})

	t.Run("combo chart with dual axis", func(t *testing.T) {
		t.Parallel()
		comboResponse := tablebuilder.ChartResponse{
			Type:       tablebuilder.ChartTypeCombo,
			Title:      "Revenue vs Growth Rate",
			Categories: []string{"Q1", "Q2", "Q3", "Q4"},
			Series: []tablebuilder.SeriesData{
				{
					Name:       "Revenue",
					Type:       "bar",
					YAxisIndex: 0,
					Data:       []float64{100000, 120000, 150000, 180000},
				},
				{
					Name:       "Growth Rate",
					Type:       "line",
					YAxisIndex: 1,
					Data:       []float64{0, 20, 25, 20},
				},
			},
			Meta: tablebuilder.ChartMeta{
				ExecutionTime: 80,
				RowsProcessed: 4,
			},
		}

		jsonBytes, err := json.Marshal(comboResponse)
		if err != nil {
			t.Fatalf("failed to marshal combo response: %v", err)
		}

		var unmarshaled tablebuilder.ChartResponse
		if err := json.Unmarshal(jsonBytes, &unmarshaled); err != nil {
			t.Fatalf("failed to unmarshal combo response: %v", err)
		}

		if unmarshaled.Series[0].Type != "bar" {
			t.Errorf("first series type mismatch: got %s, want bar", unmarshaled.Series[0].Type)
		}
		if unmarshaled.Series[1].YAxisIndex != 1 {
			t.Errorf("second series yAxisIndex mismatch: got %d, want 1", unmarshaled.Series[1].YAxisIndex)
		}
	})

	t.Run("chart visual settings", func(t *testing.T) {
		t.Parallel()
		settings := tablebuilder.ChartVisualSettings{
			ChartType:      tablebuilder.ChartTypeLine,
			CategoryColumn: "month",
			ValueColumns:   []string{"revenue", "costs"},
			XAxis: &tablebuilder.AxisConfig{
				Title: "Month",
				Type:  "category",
			},
			YAxis: &tablebuilder.AxisConfig{
				Title:  "Amount",
				Type:   "value",
				Format: "currency",
			},
			Legend: &tablebuilder.LegendConfig{
				Show:     true,
				Position: "top",
			},
			Colors: []string{"#3498db", "#e74c3c"},
		}

		jsonBytes, err := json.Marshal(settings)
		if err != nil {
			t.Fatalf("failed to marshal chart settings: %v", err)
		}

		var unmarshaled tablebuilder.ChartVisualSettings
		if err := json.Unmarshal(jsonBytes, &unmarshaled); err != nil {
			t.Fatalf("failed to unmarshal chart settings: %v", err)
		}

		if unmarshaled.ChartType != tablebuilder.ChartTypeLine {
			t.Errorf("chart type mismatch: got %s, want %s", unmarshaled.ChartType, tablebuilder.ChartTypeLine)
		}
		if len(unmarshaled.ValueColumns) != 2 {
			t.Errorf("value columns length mismatch: got %d, want 2", len(unmarshaled.ValueColumns))
		}
	})

	t.Run("KPI config thresholds", func(t *testing.T) {
		t.Parallel()
		kpiConfig := tablebuilder.KPIConfig{
			Label:             "Active Users",
			Format:            "number",
			CompareColumn:     "previous_users",
			TargetValue:       1000,
			ThresholdWarning:  800,
			ThresholdCritical: 500,
		}

		jsonBytes, err := json.Marshal(kpiConfig)
		if err != nil {
			t.Fatalf("failed to marshal KPI config: %v", err)
		}

		var unmarshaled tablebuilder.KPIConfig
		if err := json.Unmarshal(jsonBytes, &unmarshaled); err != nil {
			t.Fatalf("failed to unmarshal KPI config: %v", err)
		}

		if unmarshaled.ThresholdWarning != 800 {
			t.Errorf("threshold warning mismatch: got %f, want 800", unmarshaled.ThresholdWarning)
		}
	})
}

func Test_ChartTransformer(t *testing.T) {
	t.Parallel()

	t.Run("transform KPI from table data", func(t *testing.T) {
		t.Parallel()
		transformer := tablebuilder.NewChartTransformer()

		data := &tablebuilder.TableData{
			Data: []tablebuilder.TableRow{
				{"total_revenue": float64(125000.50)},
			},
		}

		config := &tablebuilder.Config{
			Title:         "Total Revenue",
			WidgetType:    "kpi",
			Visualization: "kpi",
		}

		result, err := transformer.Transform(data, config)
		if err != nil {
			t.Fatalf("transform failed: %v", err)
		}

		if result.Type != tablebuilder.ChartTypeKPI {
			t.Errorf("type mismatch: got %s, want %s", result.Type, tablebuilder.ChartTypeKPI)
		}
		if result.KPI == nil {
			t.Fatal("KPI data is nil")
		}
		if result.KPI.Value != 125000.50 {
			t.Errorf("KPI value mismatch: got %f, want %f", result.KPI.Value, 125000.50)
		}
	})

	t.Run("transform line chart from table data", func(t *testing.T) {
		t.Parallel()
		transformer := tablebuilder.NewChartTransformer()

		data := &tablebuilder.TableData{
			Data: []tablebuilder.TableRow{
				{"month": "Jan", "revenue": float64(10000), "costs": float64(8000)},
				{"month": "Feb", "revenue": float64(12000), "costs": float64(9000)},
				{"month": "Mar", "revenue": float64(15000), "costs": float64(10000)},
			},
		}

		config := &tablebuilder.Config{
			Title:         "Monthly Revenue",
			WidgetType:    "chart",
			Visualization: "line",
		}

		result, err := transformer.Transform(data, config)
		if err != nil {
			t.Fatalf("transform failed: %v", err)
		}

		if result.Type != tablebuilder.ChartTypeLine {
			t.Errorf("type mismatch: got %s, want %s", result.Type, tablebuilder.ChartTypeLine)
		}
		if len(result.Categories) != 3 {
			t.Errorf("categories length mismatch: got %d, want 3", len(result.Categories))
		}
		// Should have 2 series (revenue and costs)
		if len(result.Series) < 1 {
			t.Errorf("series length mismatch: got %d, want at least 1", len(result.Series))
		}
	})

	t.Run("transform bar chart", func(t *testing.T) {
		t.Parallel()
		transformer := tablebuilder.NewChartTransformer()

		data := &tablebuilder.TableData{
			Data: []tablebuilder.TableRow{
				{"product": "Widget A", "sales": float64(150)},
				{"product": "Widget B", "sales": float64(200)},
				{"product": "Widget C", "sales": float64(100)},
			},
		}

		config := &tablebuilder.Config{
			Title:         "Product Sales",
			WidgetType:    "chart",
			Visualization: "bar",
		}

		result, err := transformer.Transform(data, config)
		if err != nil {
			t.Fatalf("transform failed: %v", err)
		}

		if result.Type != tablebuilder.ChartTypeBar {
			t.Errorf("type mismatch: got %s, want %s", result.Type, tablebuilder.ChartTypeBar)
		}
	})

	t.Run("transform pie chart", func(t *testing.T) {
		t.Parallel()
		transformer := tablebuilder.NewChartTransformer()

		data := &tablebuilder.TableData{
			Data: []tablebuilder.TableRow{
				{"category": "Electronics", "value": float64(45)},
				{"category": "Clothing", "value": float64(30)},
				{"category": "Food", "value": float64(25)},
			},
		}

		config := &tablebuilder.Config{
			Title:         "Sales by Category",
			WidgetType:    "chart",
			Visualization: "pie",
		}

		result, err := transformer.Transform(data, config)
		if err != nil {
			t.Fatalf("transform failed: %v", err)
		}

		if result.Type != tablebuilder.ChartTypePie {
			t.Errorf("type mismatch: got %s, want %s", result.Type, tablebuilder.ChartTypePie)
		}
		if len(result.Series) != 3 {
			t.Errorf("series length mismatch: got %d, want 3", len(result.Series))
		}
	})

	t.Run("transformer handles empty data", func(t *testing.T) {
		t.Parallel()
		transformer := tablebuilder.NewChartTransformer()

		data := &tablebuilder.TableData{
			Data: []tablebuilder.TableRow{},
		}

		config := &tablebuilder.Config{
			Title:         "Empty KPI",
			WidgetType:    "kpi",
			Visualization: "kpi",
		}

		result, err := transformer.Transform(data, config)
		if err != nil {
			t.Fatalf("transform failed: %v", err)
		}

		if result.KPI == nil {
			t.Fatal("KPI data should not be nil for empty data")
		}
		if result.KPI.Value != 0 {
			t.Errorf("empty KPI should have value 0, got %f", result.KPI.Value)
		}
	})

	t.Run("transformer rejects nil data", func(t *testing.T) {
		t.Parallel()
		transformer := tablebuilder.NewChartTransformer()

		config := &tablebuilder.Config{
			Title:         "Test",
			Visualization: "kpi",
		}

		_, err := transformer.Transform(nil, config)
		if err == nil {
			t.Error("expected error for nil data")
		}
	})

	t.Run("transformer rejects table type", func(t *testing.T) {
		t.Parallel()
		transformer := tablebuilder.NewChartTransformer()

		data := &tablebuilder.TableData{
			Data: []tablebuilder.TableRow{
				{"col": "value"},
			},
		}

		config := &tablebuilder.Config{
			Title:         "Table",
			WidgetType:    "table",
			Visualization: "table",
		}

		_, err := transformer.Transform(data, config)
		if err == nil {
			t.Error("expected error for table type")
		}
	})

	t.Run("transform gauge chart", func(t *testing.T) {
		t.Parallel()
		transformer := tablebuilder.NewChartTransformer()

		data := &tablebuilder.TableData{
			Data: []tablebuilder.TableRow{
				{"revenue": float64(750000)},
			},
		}

		config := &tablebuilder.Config{
			Title:         "Revenue Progress",
			WidgetType:    "chart",
			Visualization: "gauge",
		}

		result, err := transformer.Transform(data, config)
		if err != nil {
			t.Fatalf("transform failed: %v", err)
		}

		if result.Type != tablebuilder.ChartTypeGauge {
			t.Errorf("type mismatch: got %s, want %s", result.Type, tablebuilder.ChartTypeGauge)
		}
		if result.KPI == nil {
			t.Fatal("KPI data is nil")
		}
		if result.KPI.Value != 750000 {
			t.Errorf("KPI value mismatch: got %f, want %f", result.KPI.Value, 750000.0)
		}
		// Gauge should have target/min/max
		if result.KPI.Max == 0 {
			t.Error("gauge max should be set")
		}
	})

	t.Run("transform stacked bar chart", func(t *testing.T) {
		t.Parallel()
		transformer := tablebuilder.NewChartTransformer()

		data := &tablebuilder.TableData{
			Data: []tablebuilder.TableRow{
				{"region": "North", "q1": float64(100), "q2": float64(120)},
				{"region": "South", "q1": float64(80), "q2": float64(90)},
				{"region": "East", "q1": float64(150), "q2": float64(160)},
			},
		}

		config := &tablebuilder.Config{
			Title:         "Sales by Region",
			WidgetType:    "chart",
			Visualization: "stacked-bar",
		}

		result, err := transformer.Transform(data, config)
		if err != nil {
			t.Fatalf("transform failed: %v", err)
		}

		if result.Type != tablebuilder.ChartTypeStackedBar {
			t.Errorf("type mismatch: got %s, want %s", result.Type, tablebuilder.ChartTypeStackedBar)
		}
		if len(result.Categories) != 3 {
			t.Errorf("categories length mismatch: got %d, want 3", len(result.Categories))
		}
		// Verify stack property is set
		for _, series := range result.Series {
			if series.Stack == "" {
				t.Error("stacked chart series should have stack property set")
			}
		}
	})

	t.Run("transform stacked area chart", func(t *testing.T) {
		t.Parallel()
		transformer := tablebuilder.NewChartTransformer()

		data := &tablebuilder.TableData{
			Data: []tablebuilder.TableRow{
				{"month": "Jan", "revenue": float64(1000)},
				{"month": "Feb", "revenue": float64(1200)},
			},
		}

		config := &tablebuilder.Config{
			Title:         "Cumulative Revenue",
			WidgetType:    "chart",
			Visualization: "stacked-area",
		}

		result, err := transformer.Transform(data, config)
		if err != nil {
			t.Fatalf("transform failed: %v", err)
		}

		if result.Type != tablebuilder.ChartTypeStackedArea {
			t.Errorf("type mismatch: got %s, want %s", result.Type, tablebuilder.ChartTypeStackedArea)
		}
	})

	t.Run("transform combo chart with series config", func(t *testing.T) {
		t.Parallel()
		transformer := tablebuilder.NewChartTransformer()

		data := &tablebuilder.TableData{
			Data: []tablebuilder.TableRow{
				{"month": "Jan", "revenue": float64(50000), "orders": float64(100)},
				{"month": "Feb", "revenue": float64(60000), "orders": float64(120)},
			},
		}

		config := &tablebuilder.Config{
			Title:         "Revenue vs Orders",
			WidgetType:    "chart",
			Visualization: "combo",
		}

		result, err := transformer.Transform(data, config)
		if err != nil {
			t.Fatalf("transform failed: %v", err)
		}

		if result.Type != tablebuilder.ChartTypeCombo {
			t.Errorf("type mismatch: got %s, want %s", result.Type, tablebuilder.ChartTypeCombo)
		}
		// Combo should have series with types
		foundBar := false
		foundLine := false
		for _, series := range result.Series {
			if series.Type == "bar" {
				foundBar = true
			}
			if series.Type == "line" {
				foundLine = true
			}
		}
		if !foundBar || !foundLine {
			t.Error("combo chart should have both bar and line series types")
		}
	})

	t.Run("transform waterfall chart", func(t *testing.T) {
		t.Parallel()
		transformer := tablebuilder.NewChartTransformer()

		data := &tablebuilder.TableData{
			Data: []tablebuilder.TableRow{
				{"name": "Revenue", "value": float64(500000)},
				{"name": "COGS", "value": float64(-200000)},
				{"name": "OpEx", "value": float64(-100000)},
				{"name": "Net Profit", "value": float64(200000)},
			},
		}

		config := &tablebuilder.Config{
			Title:         "Profit Breakdown",
			WidgetType:    "chart",
			Visualization: "waterfall",
		}

		result, err := transformer.Transform(data, config)
		if err != nil {
			t.Fatalf("transform failed: %v", err)
		}

		if result.Type != tablebuilder.ChartTypeWaterfall {
			t.Errorf("type mismatch: got %s, want %s", result.Type, tablebuilder.ChartTypeWaterfall)
		}
		if len(result.Categories) != 4 {
			t.Errorf("categories length mismatch: got %d, want 4", len(result.Categories))
		}
	})

	t.Run("transform funnel chart", func(t *testing.T) {
		t.Parallel()
		transformer := tablebuilder.NewChartTransformer()

		data := &tablebuilder.TableData{
			Data: []tablebuilder.TableRow{
				{"stage": "Leads", "count": float64(1000)},
				{"stage": "Qualified", "count": float64(600)},
				{"stage": "Proposals", "count": float64(300)},
				{"stage": "Won", "count": float64(100)},
			},
		}

		config := &tablebuilder.Config{
			Title:         "Sales Pipeline",
			WidgetType:    "chart",
			Visualization: "funnel",
			VisualSettings: tablebuilder.VisualSettings{
				Columns: map[string]tablebuilder.ColumnConfig{
					"_chart": {
						CellTemplate: `{"chartType":"funnel","categoryColumn":"stage","valueColumns":["count"]}`,
					},
				},
			},
		}

		result, err := transformer.Transform(data, config)
		if err != nil {
			t.Fatalf("transform failed: %v", err)
		}

		if result.Type != tablebuilder.ChartTypeFunnel {
			t.Errorf("type mismatch: got %s, want %s", result.Type, tablebuilder.ChartTypeFunnel)
		}
		if len(result.Series) != 1 {
			t.Errorf("series length mismatch: got %d, want 1", len(result.Series))
		}
		if len(result.Categories) != 4 {
			t.Errorf("categories length mismatch: got %d, want 4", len(result.Categories))
		}
		if len(result.Series[0].Data) != 4 {
			t.Errorf("series data length mismatch: got %d, want 4", len(result.Series[0].Data))
		}
	})

	t.Run("transform heatmap chart", func(t *testing.T) {
		t.Parallel()
		transformer := tablebuilder.NewChartTransformer()

		data := &tablebuilder.TableData{
			Data: []tablebuilder.TableRow{
				{"day": "Mon", "hour": "9", "count": float64(10)},
				{"day": "Mon", "hour": "10", "count": float64(15)},
				{"day": "Tue", "hour": "9", "count": float64(12)},
				{"day": "Tue", "hour": "10", "count": float64(8)},
			},
		}

		config := &tablebuilder.Config{
			Title:         "Activity Heatmap",
			WidgetType:    "chart",
			Visualization: "heatmap",
			VisualSettings: tablebuilder.VisualSettings{
				Columns: map[string]tablebuilder.ColumnConfig{
					"_chart": {
						CellTemplate: `{"chartType":"heatmap","xCategoryColumn":"hour","yCategoryColumn":"day","valueColumns":["count"]}`,
					},
				},
			},
		}

		result, err := transformer.Transform(data, config)
		if err != nil {
			t.Fatalf("transform failed: %v", err)
		}

		if result.Type != tablebuilder.ChartTypeHeatmap {
			t.Errorf("type mismatch: got %s, want %s", result.Type, tablebuilder.ChartTypeHeatmap)
		}
		if result.Heatmap == nil {
			t.Fatal("heatmap data is nil")
		}
		if len(result.Heatmap.XCategories) != 2 {
			t.Errorf("xCategories length mismatch: got %d, want 2", len(result.Heatmap.XCategories))
		}
		if len(result.Heatmap.YCategories) != 2 {
			t.Errorf("yCategories length mismatch: got %d, want 2", len(result.Heatmap.YCategories))
		}
	})

	t.Run("transform treemap chart", func(t *testing.T) {
		t.Parallel()
		transformer := tablebuilder.NewChartTransformer()

		data := &tablebuilder.TableData{
			Data: []tablebuilder.TableRow{
				{"category": "Electronics", "revenue": float64(50000)},
				{"category": "Clothing", "revenue": float64(30000)},
				{"category": "Food", "revenue": float64(20000)},
			},
		}

		config := &tablebuilder.Config{
			Title:         "Revenue Breakdown",
			WidgetType:    "chart",
			Visualization: "treemap",
			VisualSettings: tablebuilder.VisualSettings{
				Columns: map[string]tablebuilder.ColumnConfig{
					"_chart": {
						CellTemplate: `{"chartType":"treemap","categoryColumn":"category","valueColumns":["revenue"]}`,
					},
				},
			},
		}

		result, err := transformer.Transform(data, config)
		if err != nil {
			t.Fatalf("transform failed: %v", err)
		}

		if result.Type != tablebuilder.ChartTypeTreemap {
			t.Errorf("type mismatch: got %s, want %s", result.Type, tablebuilder.ChartTypeTreemap)
		}
		if result.Treemap == nil {
			t.Fatal("treemap data is nil")
		}
		if len(result.Treemap.Children) != 3 {
			t.Errorf("treemap children length mismatch: got %d, want 3", len(result.Treemap.Children))
		}
	})

	t.Run("transform gantt chart", func(t *testing.T) {
		t.Parallel()
		transformer := tablebuilder.NewChartTransformer()

		data := &tablebuilder.TableData{
			Data: []tablebuilder.TableRow{
				{"task": "Planning", "start": "2025-01-01", "end": "2025-01-15", "progress": float64(100)},
				{"task": "Development", "start": "2025-01-10", "end": "2025-02-15", "progress": float64(50)},
				{"task": "Testing", "start": "2025-02-01", "end": "2025-02-28", "progress": float64(0)},
			},
		}

		config := &tablebuilder.Config{
			Title:         "Project Timeline",
			WidgetType:    "chart",
			Visualization: "gantt",
			VisualSettings: tablebuilder.VisualSettings{
				Columns: map[string]tablebuilder.ColumnConfig{
					"_chart": {
						CellTemplate: `{"chartType":"gantt","nameColumn":"task","startColumn":"start","endColumn":"end","progressColumn":"progress"}`,
					},
				},
			},
		}

		result, err := transformer.Transform(data, config)
		if err != nil {
			t.Fatalf("transform failed: %v", err)
		}

		if result.Type != tablebuilder.ChartTypeGantt {
			t.Errorf("type mismatch: got %s, want %s", result.Type, tablebuilder.ChartTypeGantt)
		}
		if len(result.Gantt) != 3 {
			t.Errorf("gantt data length mismatch: got %d, want 3", len(result.Gantt))
		}
		// Verify first task
		if result.Gantt[0].Name != "Planning" {
			t.Errorf("gantt task name mismatch: got %s, want Planning", result.Gantt[0].Name)
		}
		if result.Gantt[0].Progress != 100 {
			t.Errorf("gantt progress mismatch: got %d, want 100", result.Gantt[0].Progress)
		}
	})

	t.Run("transform with metadata", func(t *testing.T) {
		t.Parallel()
		transformer := tablebuilder.NewChartTransformer()

		data := &tablebuilder.TableData{
			Data: []tablebuilder.TableRow{
				{"value": float64(100)},
			},
		}

		config := &tablebuilder.Config{
			Title:         "Test KPI",
			Visualization: "kpi",
		}

		result, err := transformer.Transform(data, config)
		if err != nil {
			t.Fatalf("transform failed: %v", err)
		}

		// Should have metadata
		if result.Meta.RowsProcessed != 1 {
			t.Errorf("rowsProcessed mismatch: got %d, want 1", result.Meta.RowsProcessed)
		}
		if result.Meta.ExecutionTime < 0 {
			t.Error("executionTime should be non-negative")
		}
	})

	t.Run("KPI with trend calculation", func(t *testing.T) {
		t.Parallel()
		transformer := tablebuilder.NewChartTransformer()

		// Two rows - current and previous for trend calculation
		data := &tablebuilder.TableData{
			Data: []tablebuilder.TableRow{
				{"revenue": float64(120000)}, // Current
				{"revenue": float64(100000)}, // Previous
			},
		}

		config := &tablebuilder.Config{
			Title:         "Revenue with Trend",
			Visualization: "kpi",
		}

		result, err := transformer.Transform(data, config)
		if err != nil {
			t.Fatalf("transform failed: %v", err)
		}

		if result.KPI.Value != 120000 {
			t.Errorf("KPI value mismatch: got %f, want %f", result.KPI.Value, 120000.0)
		}
		if result.KPI.PreviousValue != 100000 {
			t.Errorf("KPI previous value mismatch: got %f, want %f", result.KPI.PreviousValue, 100000.0)
		}
		if result.KPI.Trend != "up" {
			t.Errorf("KPI trend mismatch: got %s, want up", result.KPI.Trend)
		}
		// 20% increase
		expectedChange := 20.0
		if result.KPI.Change < expectedChange-0.1 || result.KPI.Change > expectedChange+0.1 {
			t.Errorf("KPI change mismatch: got %f, want ~%f", result.KPI.Change, expectedChange)
		}
	})

	t.Run("rejects unsupported chart type", func(t *testing.T) {
		t.Parallel()
		transformer := tablebuilder.NewChartTransformer()

		data := &tablebuilder.TableData{
			Data: []tablebuilder.TableRow{
				{"value": float64(100)},
			},
		}

		config := &tablebuilder.Config{
			Title:         "Unknown Chart",
			Visualization: "unknown-type",
		}

		_, err := transformer.Transform(data, config)
		if err == nil {
			t.Error("expected error for unsupported chart type")
		}
	})
}

// Test_ColumnTypeFromVisualSettings verifies that column types are correctly
// read from VisualSettings.Columns[].Type and override the default "string" type.
func Test_ColumnTypeFromVisualSettings(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_ColumnTypeFromVisualSettings")
	log := logger.New(io.Discard, logger.LevelInfo, "ADMIN", func(context.Context) string { return "00000000-0000-0000-0000-000000000000" })

	store := tablebuilder.NewStore(log, db.DB)

	// Seed minimal data for the test - we need at least one product
	busDomain := db.BusDomain
	ctx := context.Background()

	// Seed required data for the test
	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		t.Fatalf("seeding users: %v", err)
	}

	_ = users

	// Test configuration with explicit types in VisualSettings
	configWithTypes := &tablebuilder.Config{
		Title:         "Column Types Test",
		WidgetType:    "table",
		Visualization: "table",
		DataSource: []tablebuilder.DataSource{
			{
				Type:   "query",
				Source: "users",
				Schema: "core",
				Select: tablebuilder.SelectConfig{
					Columns: []tablebuilder.ColumnDefinition{
						{Name: "id", TableColumn: "users.id"},
						{Name: "username", TableColumn: "users.username"},
						{Name: "enabled", TableColumn: "users.enabled"},
						{Name: "created_date", TableColumn: "users.created_date"},
					},
				},
			},
		},
		VisualSettings: tablebuilder.VisualSettings{
			Columns: map[string]tablebuilder.ColumnConfig{
				"users.id": {
					Name: "users.id",
					Type: "uuid",
				},
				"users.username": {
					Name: "users.username",
					Type: "string",
				},
				"users.enabled": {
					Name: "users.enabled",
					Type: "boolean",
				},
				"users.created_date": {
					Name:   "users.created_date",
					Type:   "datetime",
					Format: &tablebuilder.FormatConfig{Type: "date", Format: "yyyy-MM-dd"},
				},
			},
		},
	}

	params := tablebuilder.QueryParams{
		Page: 1,
		Rows: 10,
	}

	result, err := store.FetchTableData(ctx, configWithTypes, params)
	if err != nil {
		t.Fatalf("failed to fetch table data: %v", err)
	}

	// Create a map for easy lookup
	columnTypes := make(map[string]string)
	for _, col := range result.Meta.Columns {
		columnTypes[col.Field] = col.Type
	}

	// Test cases: field name -> expected type
	expectedTypes := map[string]string{
		"users.id":           "uuid",     // Explicit type from VisualSettings
		"users.username":     "string",   // Explicit type from VisualSettings
		"users.enabled":      "boolean",  // Explicit type from VisualSettings
		"users.created_date": "datetime", // Explicit type from VisualSettings
	}

	for field, expectedType := range expectedTypes {
		actualType, found := columnTypes[field]
		if !found {
			t.Errorf("column %q not found in metadata", field)
			continue
		}
		if actualType != expectedType {
			t.Errorf("column %q type mismatch: got %q, want %q", field, actualType, expectedType)
		}
	}
}
