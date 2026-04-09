package directedworkapi_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
	"github.com/timmaaaz/ichor/business/domain/sales/lineitemfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// DirectedWorkSeedData holds directed-work specific test state.
type DirectedWorkSeedData struct {
	apitest.SeedData

	// Worker is a user with an assigned pending pick task.
	Worker apitest.User

	// Unassigned is a user with no tasks (expects null response).
	Unassigned apitest.User

	// WorkerPickTask is the pending pick task seeded for Worker.
	WorkerPickTask picktaskbus.PickTask
}

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (DirectedWorkSeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	// =========================================================================
	// Users — two workers: one gets a task assigned, one stays empty.
	// =========================================================================

	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 2, userbus.Roles.User, busDomain.User)
	if err != nil {
		return DirectedWorkSeedData{}, fmt.Errorf("seeding users: %w", err)
	}
	worker := apitest.User{
		User:  users[0],
		Token: apitest.Token(busDomain.User, ath, users[0].Email.Address),
	}
	unassigned := apitest.User{
		User:  users[1],
		Token: apitest.Token(busDomain.User, ath, users[1].Email.Address),
	}

	// =========================================================================
	// Geography — required for warehouses and contact infos.
	// =========================================================================

	const warehouseCount = 1

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return DirectedWorkSeedData{}, fmt.Errorf("querying regions: %w", err)
	}
	regionIDs := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		regionIDs = append(regionIDs, r.ID)
	}

	ctys, err := citybus.TestSeedCities(ctx, warehouseCount, regionIDs, busDomain.City)
	if err != nil {
		return DirectedWorkSeedData{}, fmt.Errorf("seeding cities: %w", err)
	}
	ctyIDs := make([]uuid.UUID, 0, len(ctys))
	for _, c := range ctys {
		ctyIDs = append(ctyIDs, c.ID)
	}

	strs, err := streetbus.TestSeedStreets(ctx, warehouseCount, ctyIDs, busDomain.Street)
	if err != nil {
		return DirectedWorkSeedData{}, fmt.Errorf("seeding streets: %w", err)
	}
	strIDs := make([]uuid.UUID, 0, len(strs))
	for _, s := range strs {
		strIDs = append(strIDs, s.ID)
	}

	// =========================================================================
	// Warehouse infrastructure
	// =========================================================================

	warehouses, err := warehousebus.TestSeedWarehouses(ctx, warehouseCount, worker.ID, strIDs, busDomain.Warehouse)
	if err != nil {
		return DirectedWorkSeedData{}, fmt.Errorf("seeding warehouses: %w", err)
	}
	warehouseIDs := make(uuid.UUIDs, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	zones, err := zonebus.TestSeedZone(ctx, 1, warehouseIDs, busDomain.Zones)
	if err != nil {
		return DirectedWorkSeedData{}, fmt.Errorf("seeding zones: %w", err)
	}
	zoneIDs := make([]uuid.UUID, len(zones))
	for i, z := range zones {
		zoneIDs[i] = z.ZoneID
	}

	inventoryLocations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 1, warehouseIDs, zoneIDs, busDomain.InventoryLocation)
	if err != nil {
		return DirectedWorkSeedData{}, fmt.Errorf("seeding inventory locations: %w", err)
	}
	locationIDs := make([]uuid.UUID, len(inventoryLocations))
	for i, il := range inventoryLocations {
		locationIDs[i] = il.LocationID
	}

	// =========================================================================
	// Products
	// =========================================================================

	tzs, err := busDomain.Timezone.Query(ctx, timezonebus.QueryFilter{}, timezonebus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		return DirectedWorkSeedData{}, fmt.Errorf("querying timezones: %w", err)
	}
	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	contacts, err := contactinfosbus.TestSeedContactInfos(ctx, 1, strIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return DirectedWorkSeedData{}, fmt.Errorf("seeding contact infos: %w", err)
	}
	contactIDs := make(uuid.UUIDs, len(contacts))
	for i, c := range contacts {
		contactIDs[i] = c.ID
	}

	brands, err := brandbus.TestSeedBrands(ctx, 1, contactIDs, busDomain.Brand)
	if err != nil {
		return DirectedWorkSeedData{}, fmt.Errorf("seeding brands: %w", err)
	}
	brandIDs := make(uuid.UUIDs, len(brands))
	for i, b := range brands {
		brandIDs[i] = b.BrandID
	}

	pc, err := productcategorybus.TestSeedProductCategories(ctx, 1, busDomain.ProductCategory)
	if err != nil {
		return DirectedWorkSeedData{}, fmt.Errorf("seeding product categories: %w", err)
	}
	pcIDs := make(uuid.UUIDs, len(pc))
	for i, p := range pc {
		pcIDs[i] = p.ProductCategoryID
	}

	products, err := productbus.TestSeedProducts(ctx, 1, brandIDs, pcIDs, busDomain.Product)
	if err != nil {
		return DirectedWorkSeedData{}, fmt.Errorf("seeding products: %w", err)
	}
	productIDs := make(uuid.UUIDs, len(products))
	for i, p := range products {
		productIDs[i] = p.ProductID
	}

	// =========================================================================
	// Sales: Customers → Orders → Line Items
	// =========================================================================

	customers, err := customersbus.TestSeedCustomers(ctx, 1, strIDs, contactIDs, uuid.UUIDs{worker.ID}, busDomain.Customers)
	if err != nil {
		return DirectedWorkSeedData{}, fmt.Errorf("seeding customers: %w", err)
	}
	customerIDs := make(uuid.UUIDs, len(customers))
	for i, c := range customers {
		customerIDs[i] = c.ID
	}

	ofStatuses, err := orderfulfillmentstatusbus.TestSeedOrderFulfillmentStatuses(ctx, busDomain.OrderFulfillmentStatus)
	if err != nil {
		return DirectedWorkSeedData{}, fmt.Errorf("seeding order fulfillment statuses: %w", err)
	}
	ofIDs := make(uuid.UUIDs, len(ofStatuses))
	for i, s := range ofStatuses {
		ofIDs[i] = s.ID
	}

	currencies, err := currencybus.TestSeedCurrencies(ctx, 1, busDomain.Currency)
	if err != nil {
		return DirectedWorkSeedData{}, fmt.Errorf("seeding currencies: %w", err)
	}
	currencyIDs := make(uuid.UUIDs, len(currencies))
	for i, c := range currencies {
		currencyIDs[i] = c.ID
	}

	orders, err := ordersbus.TestSeedOrders(ctx, 1, uuid.UUIDs{worker.ID}, customerIDs, ofIDs, currencyIDs, busDomain.Order)
	if err != nil {
		return DirectedWorkSeedData{}, fmt.Errorf("seeding orders: %w", err)
	}
	orderIDs := make(uuid.UUIDs, len(orders))
	for i, o := range orders {
		orderIDs[i] = o.ID
	}

	liStatuses, err := lineitemfulfillmentstatusbus.TestSeedLineItemFulfillmentStatuses(ctx, busDomain.LineItemFulfillmentStatus)
	if err != nil {
		return DirectedWorkSeedData{}, fmt.Errorf("seeding line item fulfillment statuses: %w", err)
	}
	liStatusIDs := make(uuid.UUIDs, len(liStatuses))
	for i, s := range liStatuses {
		liStatusIDs[i] = s.ID
	}

	lineItems, err := orderlineitemsbus.TestSeedOrderLineItems(ctx, 1, orderIDs, productIDs, liStatusIDs, uuid.UUIDs{worker.ID}, busDomain.OrderLineItem)
	if err != nil {
		return DirectedWorkSeedData{}, fmt.Errorf("seeding order line items: %w", err)
	}
	lineItemIDs := make(uuid.UUIDs, len(lineItems))
	for i, li := range lineItems {
		lineItemIDs[i] = li.ID
	}

	// Map line items back to their order IDs for FK consistency.
	salesOrderIDs := make(uuid.UUIDs, len(lineItems))
	for i, li := range lineItems {
		salesOrderIDs[i] = li.OrderID
	}

	// =========================================================================
	// Inventory items (needed so the pick task store can reference inventory).
	// =========================================================================

	_, err = inventoryitembus.TestSeedInventoryItems(ctx, 1, locationIDs, productIDs, busDomain.InventoryItem)
	if err != nil {
		return DirectedWorkSeedData{}, fmt.Errorf("seeding inventory items: %w", err)
	}

	// =========================================================================
	// Pick Task — one pending task assigned to worker.
	// =========================================================================

	tasks, err := picktaskbus.TestSeedPickTasks(
		ctx, 1,
		salesOrderIDs, lineItemIDs, productIDs, locationIDs,
		[]uuid.UUID{worker.ID},    // createdByIDs
		[]uuid.UUID{worker.ID},    // assigneeIDs — assigns task to worker
		busDomain.PickTask,
	)
	if err != nil {
		return DirectedWorkSeedData{}, fmt.Errorf("seeding pick tasks: %w", err)
	}

	return DirectedWorkSeedData{
		Worker:         worker,
		Unassigned:     unassigned,
		WorkerPickTask: tasks[0],
	}, nil
}
