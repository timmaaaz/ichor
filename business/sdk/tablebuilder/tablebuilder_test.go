package tablebuilder_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/customersbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	"github.com/timmaaaz/ichor/business/domain/order/lineitemfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/order/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/order/orderlineitemsbus"
	"github.com/timmaaaz/ichor/business/domain/order/ordersbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/zonebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
	"github.com/timmaaaz/ichor/foundation/logger"
)

func Test_TableBuilder(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_TableBuilder")
	log := logger.New(io.Discard, logger.LevelInfo, "ADMIN", func(context.Context) string { return "00000000-0000-0000-0000-000000000000" })

	store := tablebuilder.NewStore(log, db.DB)
	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("failed to insert seed data: %v", err)
	}

	simpleExample(context.Background(), store)
	complexExample(context.Background(), store)
	storedConfigExample(context.Background(), store, tablebuilder.NewConfigStore(log, db.DB), sd)
	paginationExample(context.Background(), store)
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
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

	return unitest.SeedData{
		Admins:                   []unitest.User{{User: admins[0]}},
		Orders:                   orders,
		Products:                 products,
		OrderFulfillmentStatuses: ofls,
		OrderLineItems:           ols,
		Customers:                customers,
	}, nil
}

func simpleExample(ctx context.Context, store *tablebuilder.Store) {
	// Simple configuration for querying inventory items
	config := &tablebuilder.Config{
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
				Select: tablebuilder.SelectConfig{
					Columns: []tablebuilder.ColumnDefinition{
						{Name: "id", TableColumn: "inventory_items.id"},
						{Name: "quantity", TableColumn: "inventory_items.quantity"},
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
				Limit: 10,
			},
		},
		VisualSettings: tablebuilder.VisualSettings{
			Columns: map[string]tablebuilder.ColumnConfig{
				"quantity": {
					Name:       "quantity",
					Header:     "Quantity",
					Width:      100,
					Align:      "right",
					Sortable:   true,
					Filterable: true,
					Format: &tablebuilder.FormatConfig{
						Type:      "number",
						Precision: 0,
					},
				},
			},
		},
		Permissions: tablebuilder.Permissions{
			Roles:   []string{"admin", "inventory_manager"},
			Actions: []string{"view", "export"},
		},
	}

	// Execute query
	params := tablebuilder.QueryParams{
		Page:  1,
		Limit: 10,
	}

	result, err := store.FetchTableData(ctx, config, params)
	if err != nil {
		log.Printf("Error fetching data: %v", err)
		return
	}

	fmt.Printf("Simple query returned %d rows\n", len(result.Data))
	printResults(result)

	_ = 1
}

func complexExample(ctx context.Context, store *tablebuilder.Store) {
	// Complex configuration matching the original TypeScript example
	config := &tablebuilder.Config{
		Title:           "Current Inventory at Warehouse A",
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
				Type:   "query",
				Source: "inventory_items",
				Select: tablebuilder.SelectConfig{
					Columns: []tablebuilder.ColumnDefinition{
						{Name: "id"},
						{Name: "quantity", Alias: "current_quantity", TableColumn: "inventory_items.quantity"},
						{Name: "reorder_point", TableColumn: "inventory_items.reorder_point"},
						{Name: "maximum_stock", TableColumn: "inventory_items.maximum_stock"},
					},
					ForeignTables: []tablebuilder.ForeignTable{
						{
							Table:            "products",
							RelationshipFrom: "inventory_items.product_id", // CHANGED
							RelationshipTo:   "products.id",                // CHANGED
							JoinType:         "inner",                      // Optional, defaults to inner
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
					},
				},
				Filters: []tablebuilder.Filter{
					{
						Column:   "quantity",
						Operator: "gt",
						Value:    0,
					},
				},
				Limit: 50,
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
					Header: "Quantity",
					Width:  100,
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

	params := tablebuilder.QueryParams{}

	result, err := store.FetchTableData(ctx, config, params)
	if err != nil {
		log.Printf("Error fetching complex data: %v", err)
		return
	}

	fmt.Printf("Complex query returned %d rows\n", len(result.Data))
	printResults(result)
}

func storedConfigExample(ctx context.Context, store *tablebuilder.Store, configStore *tablebuilder.ConfigStore, sd unitest.SeedData) {
	// Create a configuration to save
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
				Select: tablebuilder.SelectConfig{
					Columns: []tablebuilder.ColumnDefinition{
						{Name: "order_id", TableColumn: "orders.id"},
						{Name: "order_number", TableColumn: "orders.order_number"},
						{Name: "customer_name", TableColumn: "customers.name"},
						{Name: "order_fulfillment_statuses_name", TableColumn: "order_fulfillment_statuses.name"},
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
				},
				"customer_name": {
					Name:       "customer_name",
					Header:     "Customer",
					Width:      200,
					Sortable:   true,
					Filterable: true,
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

	fmt.Printf("Saved configuration with ID: %s\n", stored.ID)

	// Load and use the configuration
	loadedConfig, err := configStore.LoadConfig(ctx, stored.ID)
	if err != nil {
		log.Printf("Error loading config: %v", err)
		return
	}

	// Use the loaded configuration
	params := tablebuilder.QueryParams{
		Page:  1,
		Limit: 25,
	}

	result, err := store.FetchTableData(ctx, loadedConfig, params)
	if err != nil {
		log.Printf("Error fetching data with stored config: %v", err)
		return
	}

	printResults(result)
}

func paginationExample(ctx context.Context, store *tablebuilder.Store) {
	config := &tablebuilder.Config{
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
				Select: tablebuilder.SelectConfig{
					Columns: []tablebuilder.ColumnDefinition{
						{Name: "id"},
						{Name: "name"},
						{Name: "sku"},
						{Name: "is_active"},
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

	// Use with page.Page
	pg := page.MustParse("1", "10")

	result, err := store.QueryByPage(ctx, config, pg)
	if err != nil {
		log.Printf("Error fetching paginated data: %v", err)
		return
	}

	printResults(result)

	fmt.Printf("Page %d of %d (Total records: %d)\n",
		result.Meta.Page,
		result.Meta.TotalPages,
		result.Meta.Total)

	// Get next page
	pg = page.MustParse("2", "10")
	result, err = store.QueryByPage(ctx, config, pg)
	if err != nil {
		log.Printf("Error fetching page 2: %v", err)
		return
	}

	printResults(result)
}

func printResults(result *tablebuilder.TableData) {
	// Pretty print first few results
	for i, row := range result.Data {
		if i >= 3 {
			fmt.Println("...")
			break
		}

		jsonData, _ := json.MarshalIndent(row, "", "  ")
		fmt.Printf("Row %d: %s\n", i+1, string(jsonData))
	}

	fmt.Printf("Execution time: %dms\n", result.Meta.ExecutionTime)
}
