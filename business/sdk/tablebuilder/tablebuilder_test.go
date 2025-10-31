package tablebuilder_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
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
		t.Fatalf("failed to insert seed data: %v", err)
	}

	simpleExample(context.Background(), store)
	simpleExample2(context.Background(), store)
	complexExample(context.Background(), store)
	storedConfigExample(context.Background(), store, configStore, sd)
	paginationExample(context.Background(), store)
	inventoryAdjustmentsExample(context.Background(), store)
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
			"current_stock": {
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
			"product_id": {
				Name:   "product_id",
				Header: "Product",
				Width:  200,
				Link: &tablebuilder.LinkConfig{
					URL:   "/products/{product_id}",
					Label: "View Product",
				},
			},
			"location_id": {
				Name:   "location_id",
				Header: "Location",
				Width:  200,
				Link: &tablebuilder.LinkConfig{
					URL:   "/inventory/locations/{location_id}",
					Label: "View Location",
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

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, count, strIDs, busDomain.ContactInfos)
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

	orders, err := ordersbus.TestSeedOrders(ctx, count, uuid.UUIDs{admins[0].ID}, customerIDs, oflIDs, busDomain.Order)
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
		log.Printf("Error saving config: %v", err)
		return unitest.SeedData{}, err
	}
	cfg2, err := configStore.Create(ctx, "current_orders", "Current Orders", currentOrders, admins[0].ID)
	if err != nil {
		log.Printf("Error saving config: %v", err)
		return unitest.SeedData{}, err
	}
	cfg3, err := configStore.Create(ctx, "inventory_items", "Inventory Items", inventoryItems, admins[0].ID)
	if err != nil {
		log.Printf("Error saving config: %v", err)
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

func simpleExample2(ctx context.Context, store *tablebuilder.Store) {
	params := tablebuilder.QueryParams{
		Page: 1,
		Rows: 10,
	}

	result, err := store.FetchTableData(ctx, currentOrders, params)
	if err != nil {
		log.Printf("Error fetching data: %v", err)
		return
	}

	fmt.Printf("\n=== Simple Example 2: Orders View ===\n")

	fullJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Full JSON result:\n%s\n\n", fullJSON)

	// fmt.Printf("Returned %d rows\n\n", len(result.Data))

	// printResults(result)
	// printMetadata(result)
}

func simpleExample(ctx context.Context, store *tablebuilder.Store) {
	params := tablebuilder.QueryParams{
		Page: 1,
		Rows: 10,
	}

	result, err := store.FetchTableData(ctx, inventoryItems, params)
	if err != nil {
		log.Printf("Error fetching data: %v", err)
		return
	}

	fmt.Printf("\n=== Simple Example: Inventory Items ===\n")

	fullJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Full JSON result:\n%s\n\n", fullJSON)

	// fmt.Printf("Returned %d rows\n\n", len(result.Data))

	// printResults(result)
	// printMetadata(result)
}

func complexExample(ctx context.Context, store *tablebuilder.Store) {

	params := tablebuilder.QueryParams{
		Page: 1,
		Rows: 10,
	}

	result, err := store.FetchTableData(ctx, currentInventoryProducts, params)
	if err != nil {
		log.Printf("Error fetching complex data: %v", err)
		return
	}

	fmt.Printf("\n=== Complex Example: Inventory with Joins ===\n")

	fullJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Full JSON result:\n%s\n\n", fullJSON)

	// fmt.Printf("Returned %d rows\n\n", len(result.Data))

	// printResults(result)
	// printMetadata(result)
	// printRelationships(result)
}

func storedConfigExample(ctx context.Context, store *tablebuilder.Store, configStore *tablebuilder.ConfigStore, sd unitest.SeedData) {
	config := &tablebuilder.Config{
		Title:           "Orders Dashboard",
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
				Type:   "view",
				Source: "orders_base",
				Schema: "sales",
				Select: tablebuilder.SelectConfig{
					Columns: []tablebuilder.ColumnDefinition{
						{Name: "orders_id", Alias: "order_id", TableColumn: "orders.id"},
						{Name: "orders_number", Alias: "order_number", TableColumn: "orders.number"},
						{Name: "customers_name", Alias: "customer_name", TableColumn: "customers.name"},
						{Name: "order_fulfillment_statuses_name", Alias: "status", TableColumn: "order_fulfillment_statuses.name"},
					},
				},
				Filters: []tablebuilder.Filter{
					{
						Column:   "order_fulfillment_statuses_name",
						Operator: "in",
						Value:    []string{"PENDING", "PROCESSING", "SHIPPED"},
					},
				},
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
						URL:   "/orders/{order_id}",
						Label: "View Order",
					},
				},
				"customer_name": {
					Name:       "customer_name",
					Header:     "Customer",
					Width:      200,
					Sortable:   true,
					Filterable: true,
				},
				"status": {
					Name:   "status",
					Header: "Status",
					Width:  120,
				},
			},
		},
		Permissions: tablebuilder.Permissions{
			Roles:   []string{"admin", "sales"},
			Actions: []string{"view"},
		},
	}

	stored, err := configStore.Create(ctx, "orders_dashboard", "Main orders dashboard configuration", config, sd.Admins[0].ID)
	if err != nil {
		log.Printf("Error saving config: %v", err)
		return
	}

	fmt.Printf("\n=== Stored Config Example ===\n")
	fmt.Printf("Saved configuration with ID: %s\n\n", stored.ID)

	// Load and use the configuration
	loadedConfig, err := configStore.LoadConfig(ctx, stored.ID)
	if err != nil {
		log.Printf("Error loading config: %v", err)
		return
	}

	params := tablebuilder.QueryParams{
		Page: 1,
		Rows: 25,
	}

	result, err := store.FetchTableData(ctx, loadedConfig, params)
	if err != nil {
		log.Printf("Error fetching data with stored config: %v", err)
		return
	}

	fullJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Full JSON result:\n%s\n\n", fullJSON)

	// printResults(result)
	// printMetadata(result)
}

func paginationExample(ctx context.Context, store *tablebuilder.Store) {

	fmt.Printf("\n=== Pagination Example ===\n")

	// Page 1
	pg := page.MustParse("1", "10")
	result, err := store.QueryByPage(ctx, productsList, pg)
	if err != nil {
		log.Printf("Error fetching paginated data: %v", err)
		return
	}

	fullJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Full JSON result:\n%s\n\n", fullJSON)

	// fmt.Printf("Page 1 of %d (Total records: %d)\n", result.Meta.TotalPages, result.Meta.Total)
	// printResults(result)

	// Page 2
	pg = page.MustParse("2", "10")
	result, err = store.QueryByPage(ctx, productsList, pg)
	if err != nil {
		log.Printf("Error fetching page 2: %v", err)
		return
	}

	// fmt.Printf("\nPage 2 of %d (Total records: %d)\n", result.Meta.TotalPages, result.Meta.Total)
	// printResults(result)

	fullJSON, _ = json.MarshalIndent(result, "", "  ")
	fmt.Printf("Full JSON result:\n%s\n\n", fullJSON)
}

func inventoryAdjustmentsExample(ctx context.Context, store *tablebuilder.Store) {
	fmt.Printf("\n=== Inventory Adjustments Example ===\n")

	params := tablebuilder.QueryParams{
		Page: 1,
		Rows: 25,
	}

	result, err := store.FetchTableData(ctx, inventoryAdjustmentsPageConfig, params)
	if err != nil {
		log.Printf("Error fetching inventory adjustments data: %v", err)
		return
	}

	fullJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Full JSON result:\n%s\n\n", fullJSON)

	// Verify that we have the expected columns
	expectedColumns := []string{
		"product_name",
		"product_sku",
		"warehouse_name",
		"location_code",
		"quantity_change",
		"reason_code",
		"adjusted_by_username",
		"approved_by_username",
		"adjustment_date",
	}

	fmt.Printf("Verifying expected columns are present...\n")
	for _, expected := range expectedColumns {
		found := false
		for _, col := range result.Meta.Columns {
			if col.Field == expected {
				found = true
				break
			}
		}
		if found {
			fmt.Printf("✓ Column '%s' found\n", expected)
		} else {
			fmt.Printf("✗ Column '%s' NOT found\n", expected)
		}
	}
}

func pageConfigsExample(ctx context.Context, store *tablebuilder.Store, configStore *tablebuilder.ConfigStore, sd unitest.SeedData) {
	// Create a new page configuration
	pageConfig := tablebuilder.PageConfig{
		Name:      "Inventory Overview",
		UserID:    sd.Admins[0].ID,
		IsDefault: true,
	}
	savedPage, err := configStore.CreatePageConfig(ctx, pageConfig)
	if err != nil {
		log.Printf("Error saving page config: %v", err)
		return
	}

	// Save some tabs
	tab1 := tablebuilder.PageTabConfig{
		PageConfigID: savedPage.ID,
		Label:        "Products List",
		ConfigID:     sd.TableBuilderConfigs[0].ID,
		IsDefault:    true,
		TabOrder:     1,
	}
	tab2 := tablebuilder.PageTabConfig{
		PageConfigID: savedPage.ID,
		Label:        "Inventory Items",
		ConfigID:     sd.TableBuilderConfigs[2].ID,
		IsDefault:    false,
		TabOrder:     2,
	}
	tab3 := tablebuilder.PageTabConfig{
		PageConfigID: savedPage.ID,
		Label:        "Current Inventory",
		ConfigID:     sd.TableBuilderConfigs[1].ID,
		IsDefault:    false,
		TabOrder:     3,
	}

	ret1, err := configStore.CreatePageTabConfig(ctx, tab2)
	if err != nil {
		log.Printf("Error saving tab2 config: %v", err)
		return
	}
	_, err = configStore.CreatePageTabConfig(ctx, tab3)
	if err != nil {
		log.Printf("Error saving tab3 config: %v", err)
		return
	}
	_, err = configStore.CreatePageTabConfig(ctx, tab1)
	if err != nil {
		log.Printf("Error saving tab1 config: %v", err)
		return
	}

	// QueryByID
	queriedPage, err := configStore.QueryPageByID(ctx, savedPage.ID)
	if err != nil {
		log.Printf("Error querying page by ID: %v", err)
		return
	}
	if cmp.Diff(queriedPage, savedPage) != "" {
		log.Printf("Queried page does not match saved page")
		return
	}

	// QueryByName
	queriedPageByName, err := configStore.QueryPageByName(ctx, "Inventory Overview")
	if err != nil {
		log.Printf("Error querying page by name: %v", err)
		return
	}
	if cmp.Diff(queriedPageByName, savedPage) != "" {
		log.Printf("Queried page by name does not match saved page")
		return
	}

	// QueryTabsByID
	queriedTab, err := configStore.QueryPageTabConfigByID(ctx, ret1.ID)
	if err != nil {
		log.Printf("Error querying tabs by page ID: %v", err)
		return
	}
	if !cmp.Equal(queriedTab, ret1) {
		log.Printf("Queried tabs do not match saved tabs")
		return
	}

	// QueryPageTabConfigsByPageID
	queriedTabs, err := configStore.QueryPageTabConfigsByPageID(ctx, savedPage.ID)
	if err != nil {
		log.Printf("Error querying tabs by page ID: %v", err)
		return
	}
	if len(queriedTabs) != 3 {
		log.Printf("Expected 3 tabs, got %d", len(queriedTabs))
		return
	}

	// Update Page
	savedPage.Name = "Updated Inventory Overview"
	updatedPage, err := configStore.UpdatePageConfig(ctx, *savedPage)
	if err != nil {
		log.Printf("Error updating page: %v", err)
		return
	}
	if updatedPage.Name != "Updated Inventory Overview" {
		log.Printf("Page name not updated")
		return
	}

	// Update tab
	tab1.Label = "Updated Products List"
	updatedTab, err := configStore.UpdatePageTabConfig(ctx, tab1)
	if err != nil {
		log.Printf("Error updating tab: %v", err)
		return
	}
	if updatedTab.Label != "Updated Products List" {
		log.Printf("Tab label not updated")
		return
	}

	// Delete tab
	err = configStore.DeletePageTabConfig(ctx, tab2.ID)
	if err != nil {
		log.Printf("Error deleting tab: %v", err)
		return
	}
	deletedTabs, err := configStore.QueryPageTabConfigsByPageID(ctx, savedPage.ID)
	if err != nil {
		log.Printf("Error querying tabs after deletion: %v", err)
		return
	}
	if len(deletedTabs) != 2 {
		log.Printf("Expected 2 tabs after deletion, got %d", len(deletedTabs))
		return
	}

	// Delete page
	err = configStore.DeletePageConfig(ctx, savedPage.ID)
	if err != nil {
		log.Printf("Error deleting page: %v", err)
		return
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func printResults(result *tablebuilder.TableData) {
	fmt.Println("--- Data Rows ---")
	for i, row := range result.Data {
		if i >= 3 {
			fmt.Printf("... and %d more rows\n", len(result.Data)-3)
			break
		}

		jsonData, _ := json.MarshalIndent(row, "", "  ")
		fmt.Printf("Row %d: %s\n", i+1, string(jsonData))
	}
	fmt.Printf("\nExecution time: %dms\n", result.Meta.ExecutionTime)
}

func printMetadata(result *tablebuilder.TableData) {
	if len(result.Meta.Columns) == 0 {
		return
	}

	fmt.Println("\n--- Column Metadata ---")
	for _, col := range result.Meta.Columns {
		fmt.Printf("Field: %-25s | Display: %-20s | DB: %-20s | Type: %-10s",
			col.Field,
			col.DisplayName,
			col.DatabaseName,
			col.Type,
		)

		if col.SourceTable != "" {
			fmt.Printf(" | Source: %s.%s", col.SourceTable, col.SourceColumn)
		}

		if col.IsPrimaryKey {
			fmt.Printf(" [PK]")
		}
		if col.IsForeignKey {
			fmt.Printf(" [FK->%s]", col.RelatedTable)
		}

		fmt.Println()
	}
}

func printRelationships(result *tablebuilder.TableData) {
	if len(result.Meta.Relationships) == 0 {
		return
	}

	fmt.Println("\n--- Relationships ---")
	for _, rel := range result.Meta.Relationships {
		fmt.Printf("%s.%s -> %s.%s (%s)\n",
			rel.FromTable,
			rel.FromColumn,
			rel.ToTable,
			rel.ToColumn,
			rel.Type,
		)
	}
}
