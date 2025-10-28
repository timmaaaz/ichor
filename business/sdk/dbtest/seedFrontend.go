package dbtest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettagbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/assets/tagbus"
	"github.com/timmaaaz/ichor/business/domain/assets/userassetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
	"github.com/timmaaaz/ichor/business/domain/config/formbus"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/rolepagebus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/hr/commentbus"
	"github.com/timmaaaz/ichor/business/domain/hr/officebus"
	"github.com/timmaaaz/ichor/business/domain/hr/reportstobus"
	"github.com/timmaaaz/ichor/business/domain/hr/titlebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inspectionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/lottrackingsbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/serialnumberbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/costhistorybus"
	"github.com/timmaaaz/ichor/business/domain/products/metricsbus"
	"github.com/timmaaaz/ichor/business/domain/products/physicalattributebus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/products/productcostbus"
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
	"github.com/timmaaaz/ichor/business/domain/sales/lineitemfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
	"github.com/timmaaaz/ichor/foundation/logger"
)

func InsertSeedData(log *logger.Logger, cfg sqldb.Config) error {
	db, err := sqldb.Open(cfg)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()
	busDomain := newBusDomains(log, db)

	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return fmt.Errorf("seeding user : %w", err)
	}
	userIDs := make([]uuid.UUID, 0, len(admins))
	for _, a := range admins {
		userIDs = append(userIDs, a.ID)
	}

	// Extra users for hierarchy
	reporters, err := userbus.TestSeedUsersWithNoFKs(ctx, 20, userbus.Roles.User, busDomain.User)
	if err != nil {
		return fmt.Errorf("seeding reporter : %w", err)
	}

	reporterIDs := make([]uuid.UUID, len(reporters))
	for i, r := range reporters {
		reporterIDs[i] = r.ID
	}

	bosses, err := userbus.TestSeedUsersWithNoFKs(ctx, 10, userbus.Roles.User, busDomain.User)
	if err != nil {
		return fmt.Errorf("seeding reporter : %w", err)
	}

	bossIDs := make([]uuid.UUID, len(bosses))
	for i, b := range bosses {
		bossIDs[i] = b.ID
	}

	_, err = reportstobus.TestSeedReportsTo(ctx, 30, reporterIDs, bossIDs, busDomain.ReportsTo)
	if err != nil {
		return fmt.Errorf("seeding reportsto : %w", err)
	}

	_, err = commentbus.TestSeedUserApprovalComment(ctx, 10, reporterIDs[:5], reporterIDs[5:], busDomain.UserApprovalComment)
	if err != nil {
		return fmt.Errorf("seeding approval comments : %w", err)
	}

	_, err = titlebus.TestSeedTitles(ctx, 10, busDomain.Title)
	if err != nil {
		return fmt.Errorf("seeding fulfillment statues : %w", err)
	}

	count := 5

	assetTypes, err := assettypebus.TestSeedAssetTypes(ctx, 10, busDomain.AssetType)
	if err != nil {
		return fmt.Errorf("seeding asset types : %w", err)
	}
	assetTypeIDs := make([]uuid.UUID, 0, len(assetTypes))
	for _, at := range assetTypes {
		assetTypeIDs = append(assetTypeIDs, at.ID)
	}

	validAssets, err := validassetbus.TestSeedValidAssets(ctx, 10, assetTypeIDs, admins[0].ID, busDomain.ValidAsset)
	if err != nil {
		return fmt.Errorf("seeding assets : %w", err)
	}
	validAssetIDs := make([]uuid.UUID, 0, len(validAssets))
	for _, va := range validAssets {
		validAssetIDs = append(validAssetIDs, va.ID)
	}

	conditions, err := assetconditionbus.TestSeedAssetConditions(ctx, 8, busDomain.AssetCondition)
	if err != nil {
		return fmt.Errorf("seeding asset conditions : %w", err)
	}

	conditionIDs := make([]uuid.UUID, len(conditions))
	for i, c := range conditions {
		conditionIDs[i] = c.ID
	}

	assets, err := assetbus.TestSeedAssets(ctx, 15, validAssetIDs, conditionIDs, busDomain.Asset)
	if err != nil {
		return fmt.Errorf("seeding assets : %w", err)
	}
	assetIDs := make([]uuid.UUID, 0, len(assets))
	for _, a := range assets {
		assetIDs = append(assetIDs, a.ID)
	}

	approvalStatuses, err := approvalstatusbus.TestSeedApprovalStatus(ctx, 12, busDomain.ApprovalStatus)
	if err != nil {
		return fmt.Errorf("seeding approval statuses : %w", err)
	}
	approvalStatusIDs := make([]uuid.UUID, len(approvalStatuses))
	for i, as := range approvalStatuses {
		approvalStatusIDs[i] = as.ID
	}

	fulfillmentStatuses, err := fulfillmentstatusbus.TestSeedFulfillmentStatus(ctx, 8, busDomain.FulfillmentStatus)
	if err != nil {
		return fmt.Errorf("seeding fulfillment statuses : %w", err)
	}
	fulfillmentStatusIDs := make([]uuid.UUID, len(fulfillmentStatuses))
	for i, fs := range fulfillmentStatuses {
		fulfillmentStatusIDs[i] = fs.ID
	}

	_, err = userassetbus.TestSeedUserAssets(ctx, 25, reporterIDs[:15], assetIDs, reporterIDs[15:], conditionIDs, approvalStatusIDs, fulfillmentStatusIDs, busDomain.UserAsset)
	if err != nil {
		return fmt.Errorf("seeding user assets : %w", err)
	}

	tags, err := tagbus.TestSeedTag(ctx, 10, busDomain.Tag)
	if err != nil {
		return fmt.Errorf("seeding approval statues : %w", err)
	}
	tagIDs := make([]uuid.UUID, 0, len(tags))
	for _, t := range tags {
		tagIDs = append(tagIDs, t.ID)
	}

	_, err = assettagbus.TestSeedAssetTag(ctx, 20, validAssetIDs, tagIDs, busDomain.AssetTag)
	if err != nil {
		return fmt.Errorf("seeding asset tags : %w", err)
	}

	// ADDRESSES
	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return fmt.Errorf("querying regions : %w", err)
	}
	ids := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		ids = append(ids, r.ID)
	}

	ctys, err := citybus.TestSeedCities(ctx, count, ids, busDomain.City)
	if err != nil {
		return fmt.Errorf("seeding cities : %w", err)
	}

	ctyIDs := make([]uuid.UUID, 0, len(ctys))
	for _, c := range ctys {
		ctyIDs = append(ctyIDs, c.ID)
	}

	strs, err := streetbus.TestSeedStreets(ctx, count, ctyIDs, busDomain.Street)
	if err != nil {
		return fmt.Errorf("seeding streets : %w", err)
	}
	strIDs := make([]uuid.UUID, 0, len(strs))
	for _, s := range strs {
		strIDs = append(strIDs, s.ID)
	}

	_, err = officebus.TestSeedOffices(ctx, 10, strIDs, busDomain.Office)
	if err != nil {
		return fmt.Errorf("seeding offices : %w", err)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, count, strIDs, busDomain.ContactInfos)
	if err != nil {
		return fmt.Errorf("seeding contact info : %w", err)
	}
	contactInfoIDs := make([]uuid.UUID, 0, len(contactInfos))
	for _, ci := range contactInfos {
		contactInfoIDs = append(contactInfoIDs, ci.ID)
	}

	customers, err := customersbus.TestSeedCustomers(ctx, count, strIDs, contactInfoIDs, uuid.UUIDs{admins[0].ID}, busDomain.Customers)
	if err != nil {
		return fmt.Errorf("seeding customers : %w", err)
	}
	customerIDs := make([]uuid.UUID, 0, len(customers))
	for _, c := range customers {
		customerIDs = append(customerIDs, c.ID)
	}

	ofls, err := orderfulfillmentstatusbus.TestSeedOrderFulfillmentStatuses(ctx, busDomain.OrderFulfillmentStatus)
	if err != nil {
		return fmt.Errorf("seeding order fulfillment statuses: %w", err)
	}
	oflIDs := make([]uuid.UUID, 0, len(ofls))
	for _, ofl := range ofls {
		oflIDs = append(oflIDs, ofl.ID)
	}

	orders, err := ordersbus.TestSeedOrders(ctx, count, uuid.UUIDs{admins[0].ID}, customerIDs, oflIDs, busDomain.Order)
	if err != nil {
		return fmt.Errorf("seeding Orders: %w", err)
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
		return fmt.Errorf("seeding brand : %w", err)
	}

	brandIDs := make(uuid.UUIDs, len(brand))
	for i, b := range brand {
		brandIDs[i] = b.BrandID
	}

	productCategories, err := productcategorybus.TestSeedProductCategories(ctx, 10, busDomain.ProductCategory)
	if err != nil {
		return fmt.Errorf("seeding product category : %w", err)
	}

	productCategoryIDs := make(uuid.UUIDs, len(productCategories))

	for i, pc := range productCategories {
		productCategoryIDs[i] = pc.ProductCategoryID
	}

	products, err := productbus.TestSeedProducts(ctx, 20, brandIDs, productCategoryIDs, busDomain.Product)
	if err != nil {
		return fmt.Errorf("seeding product : %w", err)
	}
	productIDs := make([]uuid.UUID, 0, len(products))
	for _, p := range products {
		productIDs = append(productIDs, p.ProductID)
	}

	_, err = productcostbus.TestSeedProductCosts(ctx, 20, productIDs, busDomain.ProductCost)
	if err != nil {
		return fmt.Errorf("seeding product cost : %w", err)
	}

	_, err = physicalattributebus.TestSeedPhysicalAttributes(ctx, 20, productIDs, busDomain.PhysicalAttribute)
	if err != nil {
		return fmt.Errorf("seeding physical attribute : %w", err)
	}

	_, err = metricsbus.TestSeedMetrics(ctx, 40, productIDs, busDomain.Metrics)
	if err != nil {
		return fmt.Errorf("seeding metrics : %w", err)
	}

	_, err = costhistorybus.TestSeedCostHistories(ctx, 40, productIDs, busDomain.CostHistory)
	if err != nil {
		return fmt.Errorf("seeding cost history : %w", err)
	}

	olStatuses, err := lineitemfulfillmentstatusbus.TestSeedLineItemFulfillmentStatuses(ctx, busDomain.LineItemFulfillmentStatus)
	if err != nil {
		return fmt.Errorf("seeding line item fulfillment statuses: %w", err)
	}
	olStatusIDs := make([]uuid.UUID, 0, len(olStatuses))
	for _, ols := range olStatuses {
		olStatusIDs = append(olStatusIDs, ols.ID)
	}

	_, err = orderlineitemsbus.TestSeedOrderLineItems(ctx, count, orderIDs, productIDs, olStatusIDs, userIDs, busDomain.OrderLineItem)
	if err != nil {
		return fmt.Errorf("seeding Order Line Items: %w", err)
	}

	warehouseCount := 5

	// WAREHOUSES
	warehouses, err := warehousebus.TestSeedWarehouses(ctx, warehouseCount, admins[0].ID, strIDs, busDomain.Warehouse)
	if err != nil {
		return fmt.Errorf("seeding warehouses : %w", err)
	}

	warehouseIDs := make(uuid.UUIDs, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	zones, err := zonebus.TestSeedZone(ctx, 12, warehouseIDs, busDomain.Zones)
	if err != nil {
		return fmt.Errorf("seeding zones : %w", err)
	}

	zoneIDs := make([]uuid.UUID, len(zones))
	for i, z := range zones {
		zoneIDs[i] = z.ZoneID
	}

	inventoryLocations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 25, warehouseIDs, zoneIDs, busDomain.InventoryLocation)
	if err != nil {
		return fmt.Errorf("seeding inventory locations : %w", err)
	}

	inventoryLocationsIDs := make([]uuid.UUID, len(inventoryLocations))
	for i, il := range inventoryLocations {
		inventoryLocationsIDs[i] = il.LocationID
	}

	_, err = inventoryitembus.TestSeedInventoryItems(ctx, 30, inventoryLocationsIDs, productIDs, busDomain.InventoryItem)
	if err != nil {
		return fmt.Errorf("seeding inventory products : %w", err)
	}

	_, err = inventoryitembus.TestSeedInventoryItems(ctx, 30, inventoryLocationsIDs, productIDs, busDomain.InventoryItem)
	if err != nil {
		return fmt.Errorf("seeding inventory products : %w", err)
	}

	suppliers, err := supplierbus.TestSeedSuppliers(ctx, 25, contactIDs, busDomain.Supplier)
	if err != nil {
		return fmt.Errorf("seeding suppliers : %w", err)
	}

	supplierIDs := make(uuid.UUIDs, len(suppliers))
	for i, s := range suppliers {
		supplierIDs[i] = s.SupplierID
	}

	supplierProducts, err := supplierproductbus.TestSeedSupplierProducts(ctx, 10, productIDs, supplierIDs, busDomain.SupplierProduct)
	if err != nil {
		return fmt.Errorf("seeding supplier product : %w", err)
	}

	supplierProductIDs := make(uuid.UUIDs, len(supplierProducts))
	for i, sp := range supplierProducts {
		supplierProductIDs[i] = sp.SupplierProductID
	}

	lotTrackings, err := lottrackingsbus.TestSeedLotTrackings(ctx, 15, supplierProductIDs, busDomain.LotTrackings)
	if err != nil {
		return fmt.Errorf("seeding lot tracking : %w", err)
	}
	lotTrackingsIDs := make(uuid.UUIDs, len(lotTrackings))
	for i, lt := range lotTrackings {
		lotTrackingsIDs[i] = lt.LotID
	}

	_, err = inspectionbus.TestSeedInspections(ctx, 10, productIDs, userIDs, lotTrackingsIDs, busDomain.Inspection)
	if err != nil {
		return fmt.Errorf("seeding inspections : %w", err)
	}

	_, err = serialnumberbus.TestSeedSerialNumbers(ctx, 50, lotTrackingsIDs, productIDs, inventoryLocationsIDs, busDomain.SerialNumber)
	if err != nil {
		return fmt.Errorf("seeding serial numbers : %w", err)
	}

	_, err = transferorderbus.TestSeedTransferOrders(ctx, 20, productIDs, inventoryLocationsIDs[:15], inventoryLocationsIDs[15:], reporterIDs[:4], bossIDs[4:], busDomain.TransferOrder)
	if err != nil {
		return fmt.Errorf("seeding transfer orders : %w", err)
	}

	_, err = inventorytransactionbus.TestSeedInventoryTransaction(ctx, 40, inventoryLocationsIDs, productIDs, userIDs, busDomain.InventoryTransaction)

	if err != nil {
		return fmt.Errorf("seeding inventory transactions : %w", err)
	}

	_, err = inventoryadjustmentbus.TestSeedInventoryAdjustments(ctx, 20, productIDs, inventoryLocationsIDs, reporterIDs[:2], reporterIDs[2:], busDomain.InventoryAdjustment)
	if err != nil {
		return fmt.Errorf("seeding inventory adjustments : %w", err)
	}

	configStore := tablebuilder.NewConfigStore(log, db)
	_, err = configStore.Create(ctx, "orders_dashboard", "orders_base", ordersConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating stored config: %w", err)
	}

	_, err = configStore.Create(ctx, "products_dashboard", "products", PageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating stored config: %w", err)
	}

	_, err = configStore.Create(ctx, "inventory_dashboard", "inventory_items", ComplexConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating stored config: %w", err)
	}

	// Create dedicated page configs for orders, suppliers, categories, and order line items
	_, err = configStore.Create(ctx, "orders_page", "orders", ordersPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating orders page config: %w", err)
	}

	_, err = configStore.Create(ctx, "suppliers_page", "suppliers", suppliersPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating suppliers page config: %w", err)
	}

	_, err = configStore.Create(ctx, "categories_page", "product_categories", categoriesPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating categories page config: %w", err)
	}

	_, err = configStore.Create(ctx, "order_line_items_page", "order_line_items", orderLineItemsPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating order line items page config: %w", err)
	}

	// Admin Module Configs
	_, err = configStore.Create(ctx, "admin_users_page", "users", adminUsersPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating admin users page config: %w", err)
	}

	_, err = configStore.Create(ctx, "admin_roles_page", "roles", adminRolesPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating admin roles page config: %w", err)
	}

	_, err = configStore.Create(ctx, "admin_table_access_page", "table_access", adminTableAccessPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating admin table access page config: %w", err)
	}

	_, err = configStore.Create(ctx, "admin_audit_page", "automation_executions", adminAuditPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating admin audit page config: %w", err)
	}

	_, err = configStore.Create(ctx, "admin_config_page", "table_configs", adminConfigPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating admin config page config: %w", err)
	}

	// Assets Module Configs
	_, err = configStore.Create(ctx, "assets_list_page", "assets", assetsListPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating assets list page config: %w", err)
	}

	_, err = configStore.Create(ctx, "assets_requests_page", "user_assets", assetsRequestsPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating assets requests page config: %w", err)
	}

	// HR Module Configs
	_, err = configStore.Create(ctx, "hr_employees_page", "users", hrEmployeesPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating hr employees page config: %w", err)
	}

	_, err = configStore.Create(ctx, "hr_offices_page", "offices", hrOfficesPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating hr offices page config: %w", err)
	}

	// Inventory Module Configs
	_, err = configStore.Create(ctx, "inventory_warehouses_page", "warehouses", inventoryWarehousesPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating inventory warehouses page config: %w", err)
	}

	_, err = configStore.Create(ctx, "inventory_items_page", "inventory_items", inventoryItemsPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating inventory items page config: %w", err)
	}

	_, err = configStore.Create(ctx, "inventory_adjustments_page", "inventory_adjustments", inventoryAdjustmentsPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating inventory adjustments page config: %w", err)
	}

	_, err = configStore.Create(ctx, "inventory_transfers_page", "transfer_orders", inventoryTransfersPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating inventory transfers page config: %w", err)
	}

	// Sales Module Configs
	_, err = configStore.Create(ctx, "sales_customers_page", "customers", salesCustomersPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating sales customers page config: %w", err)
	}

	// Create SYSTEM-WIDE default page (user_id = NULL or uuid.Nil)
	// This is the template that all users fall back to if they don't have their own version
	defaultPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "default_dashboard",
		UserID:    uuid.Nil, // ✅ CORRECT - no user association, this is system-wide
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating default page config: %w", err)
	}

	// Get the stored config IDs we just created
	ordersConfigStored, err := configStore.QueryByName(ctx, "orders_dashboard")
	if err != nil {
		return fmt.Errorf("querying orders config: %w", err)
	}

	productsConfigStored, err := configStore.QueryByName(ctx, "products_dashboard")
	if err != nil {
		return fmt.Errorf("querying products config: %w", err)
	}

	inventoryConfigStored, err := configStore.QueryByName(ctx, "inventory_dashboard")
	if err != nil {
		return fmt.Errorf("querying inventory config: %w", err)
	}

	// Get the stored config IDs for suppliers and categories to add to default dashboard
	suppliersConfigStored, err := configStore.QueryByName(ctx, "suppliers_page")
	if err != nil {
		return fmt.Errorf("querying suppliers config: %w", err)
	}

	categoriesConfigStored, err := configStore.QueryByName(ctx, "categories_page")
	if err != nil {
		return fmt.Errorf("querying categories config: %w", err)
	}

	orderLineItemsConfigStored, err := configStore.QueryByName(ctx, "order_line_items_page")
	if err != nil {
		return fmt.Errorf("querying order line items config: %w", err)
	}

	// Create tabs for the SYSTEM default page
	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Orders",
		PageConfigID: defaultPage.ID,
		ConfigID:     ordersConfigStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating orders tab: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Products",
		PageConfigID: defaultPage.ID,
		ConfigID:     productsConfigStored.ID,
		IsDefault:    false,
		TabOrder:     2,
	})
	if err != nil {
		return fmt.Errorf("creating products tab: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Inventory",
		PageConfigID: defaultPage.ID,
		ConfigID:     inventoryConfigStored.ID,
		IsDefault:    false,
		TabOrder:     3,
	})
	if err != nil {
		return fmt.Errorf("creating inventory tab: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Suppliers",
		PageConfigID: defaultPage.ID,
		ConfigID:     suppliersConfigStored.ID,
		IsDefault:    false,
		TabOrder:     4,
	})
	if err != nil {
		return fmt.Errorf("creating suppliers tab: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Categories",
		PageConfigID: defaultPage.ID,
		ConfigID:     categoriesConfigStored.ID,
		IsDefault:    false,
		TabOrder:     5,
	})
	if err != nil {
		return fmt.Errorf("creating categories tab: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Order Line Items",
		PageConfigID: defaultPage.ID,
		ConfigID:     orderLineItemsConfigStored.ID,
		IsDefault:    false,
		TabOrder:     6,
	})
	if err != nil {
		return fmt.Errorf("creating order line items tab: %w", err)
	}

	// =========================================================================
	// Create dedicated page configs for Orders, Suppliers, and Categories
	// =========================================================================

	// Get the stored config IDs for the new pages
	ordersPageStored, err := configStore.QueryByName(ctx, "orders_page")
	if err != nil {
		return fmt.Errorf("querying orders page config: %w", err)
	}

	suppliersPageStored, err := configStore.QueryByName(ctx, "suppliers_page")
	if err != nil {
		return fmt.Errorf("querying suppliers page config: %w", err)
	}

	categoriesPageStored, err := configStore.QueryByName(ctx, "categories_page")
	if err != nil {
		return fmt.Errorf("querying categories page config: %w", err)
	}

	orderLineItemsPageStored, err := configStore.QueryByName(ctx, "order_line_items_page")
	if err != nil {
		return fmt.Errorf("querying order line items page config: %w", err)
	}

	// Query Admin Module Configs
	adminUsersPageStored, err := configStore.QueryByName(ctx, "admin_users_page")
	if err != nil {
		return fmt.Errorf("querying admin users page config: %w", err)
	}

	adminRolesPageStored, err := configStore.QueryByName(ctx, "admin_roles_page")
	if err != nil {
		return fmt.Errorf("querying admin roles page config: %w", err)
	}

	adminTableAccessPageStored, err := configStore.QueryByName(ctx, "admin_table_access_page")
	if err != nil {
		return fmt.Errorf("querying admin table access page config: %w", err)
	}

	adminAuditPageStored, err := configStore.QueryByName(ctx, "admin_audit_page")
	if err != nil {
		return fmt.Errorf("querying admin audit page config: %w", err)
	}

	adminConfigPageStored, err := configStore.QueryByName(ctx, "admin_config_page")
	if err != nil {
		return fmt.Errorf("querying admin config page config: %w", err)
	}

	// Query Assets Module Configs
	assetsListPageStored, err := configStore.QueryByName(ctx, "assets_list_page")
	if err != nil {
		return fmt.Errorf("querying assets list page config: %w", err)
	}

	assetsRequestsPageStored, err := configStore.QueryByName(ctx, "assets_requests_page")
	if err != nil {
		return fmt.Errorf("querying assets requests page config: %w", err)
	}

	// Query HR Module Configs
	hrEmployeesPageStored, err := configStore.QueryByName(ctx, "hr_employees_page")
	if err != nil {
		return fmt.Errorf("querying hr employees page config: %w", err)
	}

	hrOfficesPageStored, err := configStore.QueryByName(ctx, "hr_offices_page")
	if err != nil {
		return fmt.Errorf("querying hr offices page config: %w", err)
	}

	// Query Inventory Module Configs
	inventoryWarehousesPageStored, err := configStore.QueryByName(ctx, "inventory_warehouses_page")
	if err != nil {
		return fmt.Errorf("querying inventory warehouses page config: %w", err)
	}

	inventoryItemsPageStored, err := configStore.QueryByName(ctx, "inventory_items_page")
	if err != nil {
		return fmt.Errorf("querying inventory items page config: %w", err)
	}

	inventoryAdjustmentsPageStored, err := configStore.QueryByName(ctx, "inventory_adjustments_page")
	if err != nil {
		return fmt.Errorf("querying inventory adjustments page config: %w", err)
	}

	inventoryTransfersPageStored, err := configStore.QueryByName(ctx, "inventory_transfers_page")
	if err != nil {
		return fmt.Errorf("querying inventory transfers page config: %w", err)
	}

	// Query Sales Module Configs
	salesCustomersPageStored, err := configStore.QueryByName(ctx, "sales_customers_page")
	if err != nil {
		return fmt.Errorf("querying sales customers page config: %w", err)
	}

	// Create Orders Page
	ordersPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "orders_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating orders page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Orders",
		PageConfigID: ordersPage.ID,
		ConfigID:     ordersPageStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating orders page tab: %w", err)
	}

	// Create Suppliers Page
	suppliersPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "suppliers_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating suppliers page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Suppliers",
		PageConfigID: suppliersPage.ID,
		ConfigID:     suppliersPageStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating suppliers page tab: %w", err)
	}

	// Create Categories Page
	categoriesPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "categories_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating categories page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Categories",
		PageConfigID: categoriesPage.ID,
		ConfigID:     categoriesPageStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating categories page tab: %w", err)
	}

	// Create Order Line Items Page
	orderLineItemsPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "order_line_items_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating order line items page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Order Line Items",
		PageConfigID: orderLineItemsPage.ID,
		ConfigID:     orderLineItemsPageStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating order line items page tab: %w", err)
	}

	// =========================================================================
	// Create Admin Module Pages
	// =========================================================================

	// Admin Users Page
	adminUsersPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "admin_users_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating admin users page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Users",
		PageConfigID: adminUsersPage.ID,
		ConfigID:     adminUsersPageStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating admin users page tab: %w", err)
	}

	// Admin Roles Page
	adminRolesPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "admin_roles_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating admin roles page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Roles",
		PageConfigID: adminRolesPage.ID,
		ConfigID:     adminRolesPageStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating admin roles page tab: %w", err)
	}

	// Admin Dashboard Page (multi-tab: users, roles, table access)
	adminDashboardPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "admin_dashboard_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating admin dashboard page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Users",
		PageConfigID: adminDashboardPage.ID,
		ConfigID:     adminUsersPageStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating admin dashboard users tab: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Roles",
		PageConfigID: adminDashboardPage.ID,
		ConfigID:     adminRolesPageStored.ID,
		IsDefault:    false,
		TabOrder:     2,
	})
	if err != nil {
		return fmt.Errorf("creating admin dashboard roles tab: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Permissions",
		PageConfigID: adminDashboardPage.ID,
		ConfigID:     adminTableAccessPageStored.ID,
		IsDefault:    false,
		TabOrder:     3,
	})
	if err != nil {
		return fmt.Errorf("creating admin dashboard permissions tab: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Audit Logs",
		PageConfigID: adminDashboardPage.ID,
		ConfigID:     adminAuditPageStored.ID,
		IsDefault:    false,
		TabOrder:     4,
	})
	if err != nil {
		return fmt.Errorf("creating admin dashboard audit tab: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Configurations",
		PageConfigID: adminDashboardPage.ID,
		ConfigID:     adminConfigPageStored.ID,
		IsDefault:    false,
		TabOrder:     5,
	})
	if err != nil {
		return fmt.Errorf("creating admin dashboard config tab: %w", err)
	}

	// =========================================================================
	// Create Assets Module Pages
	// =========================================================================

	// Assets List Page
	assetsPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "assets_list_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating assets page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Assets",
		PageConfigID: assetsPage.ID,
		ConfigID:     assetsListPageStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating assets page tab: %w", err)
	}

	// Asset Requests Page
	assetsRequestsPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "assets_requests_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating assets requests page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Requests",
		PageConfigID: assetsRequestsPage.ID,
		ConfigID:     assetsRequestsPageStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating assets requests page tab: %w", err)
	}

	// Assets Dashboard (multi-tab: assets, requests)
	assetsDashboardPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "assets_dashboard_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating assets dashboard page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Assets",
		PageConfigID: assetsDashboardPage.ID,
		ConfigID:     assetsListPageStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating assets dashboard assets tab: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Requests",
		PageConfigID: assetsDashboardPage.ID,
		ConfigID:     assetsRequestsPageStored.ID,
		IsDefault:    false,
		TabOrder:     2,
	})
	if err != nil {
		return fmt.Errorf("creating assets dashboard requests tab: %w", err)
	}

	// =========================================================================
	// Create HR Module Pages
	// =========================================================================

	// HR Employees Page
	hrEmployeesPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "hr_employees_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating hr employees page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Employees",
		PageConfigID: hrEmployeesPage.ID,
		ConfigID:     hrEmployeesPageStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating hr employees page tab: %w", err)
	}

	// HR Offices Page
	hrOfficesPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "hr_offices_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating hr offices page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Offices",
		PageConfigID: hrOfficesPage.ID,
		ConfigID:     hrOfficesPageStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating hr offices page tab: %w", err)
	}

	// HR Dashboard (multi-tab: employees, offices)
	hrDashboardPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "hr_dashboard_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating hr dashboard page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Employees",
		PageConfigID: hrDashboardPage.ID,
		ConfigID:     hrEmployeesPageStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating hr dashboard employees tab: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Offices",
		PageConfigID: hrDashboardPage.ID,
		ConfigID:     hrOfficesPageStored.ID,
		IsDefault:    false,
		TabOrder:     2,
	})
	if err != nil {
		return fmt.Errorf("creating hr dashboard offices tab: %w", err)
	}

	// =========================================================================
	// Create Inventory Module Pages
	// =========================================================================

	// Inventory Warehouses Page
	inventoryWarehousesPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "inventory_warehouses_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory warehouses page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Warehouses",
		PageConfigID: inventoryWarehousesPage.ID,
		ConfigID:     inventoryWarehousesPageStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating inventory warehouses page tab: %w", err)
	}

	// Inventory Items Page
	inventoryItemsPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "inventory_items_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory items page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Items",
		PageConfigID: inventoryItemsPage.ID,
		ConfigID:     inventoryItemsPageStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating inventory items page tab: %w", err)
	}

	// Inventory Adjustments Page
	inventoryAdjustmentsPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "inventory_adjustments_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory adjustments page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Adjustments",
		PageConfigID: inventoryAdjustmentsPage.ID,
		ConfigID:     inventoryAdjustmentsPageStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating inventory adjustments page tab: %w", err)
	}

	// Inventory Transfers Page
	inventoryTransfersPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "inventory_transfers_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory transfers page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Transfers",
		PageConfigID: inventoryTransfersPage.ID,
		ConfigID:     inventoryTransfersPageStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating inventory transfers page tab: %w", err)
	}

	// Inventory Dashboard (multi-tab: warehouses, items, adjustments, transfers)
	inventoryDashboardPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "inventory_dashboard_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory dashboard page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Items",
		PageConfigID: inventoryDashboardPage.ID,
		ConfigID:     inventoryItemsPageStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating inventory dashboard items tab: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Warehouses",
		PageConfigID: inventoryDashboardPage.ID,
		ConfigID:     inventoryWarehousesPageStored.ID,
		IsDefault:    false,
		TabOrder:     2,
	})
	if err != nil {
		return fmt.Errorf("creating inventory dashboard warehouses tab: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Adjustments",
		PageConfigID: inventoryDashboardPage.ID,
		ConfigID:     inventoryAdjustmentsPageStored.ID,
		IsDefault:    false,
		TabOrder:     3,
	})
	if err != nil {
		return fmt.Errorf("creating inventory dashboard adjustments tab: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Transfers",
		PageConfigID: inventoryDashboardPage.ID,
		ConfigID:     inventoryTransfersPageStored.ID,
		IsDefault:    false,
		TabOrder:     4,
	})
	if err != nil {
		return fmt.Errorf("creating inventory dashboard transfers tab: %w", err)
	}

	// =========================================================================
	// Create Sales Module Pages
	// =========================================================================

	// Sales Customers Page
	salesCustomersPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "sales_customers_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating sales customers page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Customers",
		PageConfigID: salesCustomersPage.ID,
		ConfigID:     salesCustomersPageStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating sales customers page tab: %w", err)
	}

	// Sales Dashboard (multi-tab: orders, customers)
	salesDashboardPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "sales_dashboard_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating sales dashboard page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Orders",
		PageConfigID: salesDashboardPage.ID,
		ConfigID:     ordersPageStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating sales dashboard orders tab: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Customers",
		PageConfigID: salesDashboardPage.ID,
		ConfigID:     salesCustomersPageStored.ID,
		IsDefault:    false,
		TabOrder:     2,
	})
	if err != nil {
		return fmt.Errorf("creating sales dashboard customers tab: %w", err)
	}

	// =========================================================================

	// Form 1: Single entity - Users only
	userForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "User Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating user form : %w", err)
	}

	userEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "users")
	if err != nil {
		return fmt.Errorf("querying user entity : %w", err)
	}

	userFormFields := []formfieldbus.FormField{
		{
			ID:         uuid.New(),
			FormID:     userForm.ID,
			EntityID:   userEntity.ID,
			Name:       "username",
			FieldOrder: 1,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     userForm.ID,
			EntityID:   userEntity.ID,
			Name:       "first_name",
			FieldOrder: 2,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     userForm.ID,
			EntityID:   userEntity.ID,
			Name:       "last_name",
			FieldOrder: 3,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     userForm.ID,
			EntityID:   userEntity.ID,
			Name:       "email",
			FieldOrder: 4,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     userForm.ID,
			EntityID:   userEntity.ID,
			Name:       "password",
			FieldOrder: 5,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     userForm.ID,
			EntityID:   userEntity.ID,
			Name:       "password_confirm",
			FieldOrder: 6,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     userForm.ID,
			EntityID:   userEntity.ID,
			Name:       "birthday",
			FieldOrder: 7,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     userForm.ID,
			EntityID:   userEntity.ID,
			Name:       "roles",
			FieldOrder: 8,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     userForm.ID,
			EntityID:   userEntity.ID,
			Name:       "system_roles",
			FieldOrder: 9,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     userForm.ID,
			EntityID:   userEntity.ID,
			Name:       "enabled",
			FieldOrder: 10,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     userForm.ID,
			EntityID:   userEntity.ID,
			Name:       "requested_by",
			FieldOrder: 11,
			Config:     json.RawMessage(`{}`),
		},
	}

	for _, ff := range userFormFields {
		_, err = busDomain.FormField.Create(ctx, formfieldbus.NewFormField{
			FormID:     ff.FormID,
			EntityID:   ff.EntityID,
			Name:       ff.Name,
			FieldOrder: ff.FieldOrder,
		})
		if err != nil {
			return fmt.Errorf("creating user form field : %w", err)
		}
	}

	// Form 2: Single entity - Assets only
	assetForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Asset Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating asset form : %w", err)
	}

	assetEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "assets")
	if err != nil {
		return fmt.Errorf("querying asset entity : %w", err)
	}

	assetFormFields := []formfieldbus.FormField{
		{
			ID:         uuid.New(),
			FormID:     assetForm.ID,
			EntityID:   assetEntity.ID,
			Name:       "valid_asset_id",
			FieldOrder: 1,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     assetForm.ID,
			EntityID:   assetEntity.ID,
			Name:       "serial_number",
			FieldOrder: 2,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     assetForm.ID,
			EntityID:   assetEntity.ID,
			Name:       "asset_condition_id",
			FieldOrder: 3,
			Config:     json.RawMessage(`{}`),
		},
	}

	for _, ff := range assetFormFields {
		_, err = busDomain.FormField.Create(ctx, formfieldbus.NewFormField{
			FormID:     ff.FormID,
			EntityID:   ff.EntityID,
			Name:       ff.Name,
			FieldOrder: ff.FieldOrder,
		})
		if err != nil {
			return fmt.Errorf("creating asset form field : %w", err)
		}
	}

	// Form 3: Multi-entity - User then Asset (with foreign key)
	multiForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "User and Asset Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating multi-entity form : %w", err)
	}

	multiFormFields := []formfieldbus.FormField{
		// User fields (order 1-11)
		{
			ID:         uuid.New(),
			FormID:     multiForm.ID,
			EntityID:   userEntity.ID,
			Name:       "username",
			FieldOrder: 1,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     multiForm.ID,
			EntityID:   userEntity.ID,
			Name:       "first_name",
			FieldOrder: 2,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     multiForm.ID,
			EntityID:   userEntity.ID,
			Name:       "last_name",
			FieldOrder: 3,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     multiForm.ID,
			EntityID:   userEntity.ID,
			Name:       "email",
			FieldOrder: 4,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     multiForm.ID,
			EntityID:   userEntity.ID,
			Name:       "password",
			FieldOrder: 5,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     multiForm.ID,
			EntityID:   userEntity.ID,
			Name:       "password_confirm",
			FieldOrder: 6,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     multiForm.ID,
			EntityID:   userEntity.ID,
			Name:       "birthday",
			FieldOrder: 7,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     multiForm.ID,
			EntityID:   userEntity.ID,
			Name:       "roles",
			FieldOrder: 8,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     multiForm.ID,
			EntityID:   userEntity.ID,
			Name:       "system_roles",
			FieldOrder: 9,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     multiForm.ID,
			EntityID:   userEntity.ID,
			Name:       "enabled",
			FieldOrder: 10,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     multiForm.ID,
			EntityID:   userEntity.ID,
			Name:       "requested_by",
			FieldOrder: 11,
			Config:     json.RawMessage(`{}`),
		},
		// Asset fields (order 12-14)
		{
			ID:         uuid.New(),
			FormID:     multiForm.ID,
			EntityID:   assetEntity.ID,
			Name:       "asset_condition_id",
			FieldOrder: 12,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     multiForm.ID,
			EntityID:   assetEntity.ID,
			Name:       "valid_asset_id",
			FieldOrder: 13,
			Config:     json.RawMessage(`{}`),
		},
		{
			ID:         uuid.New(),
			FormID:     multiForm.ID,
			EntityID:   assetEntity.ID,
			Name:       "serial_number",
			FieldOrder: 14,
			Config:     json.RawMessage(`{}`),
		},
	}

	for _, ff := range multiFormFields {
		_, err = busDomain.FormField.Create(ctx, formfieldbus.NewFormField{
			FormID:     ff.FormID,
			EntityID:   ff.EntityID,
			Name:       ff.Name,
			FieldOrder: ff.FieldOrder,
		})
		if err != nil {
			return fmt.Errorf("creating multi-entity form field : %w", err)
		}
	}

	// PAGES
	var pageIDs uuid.UUIDs

	for _, page := range allPages {
		p, err := busDomain.Page.Create(ctx, page)
		if err != nil {
			return fmt.Errorf("creating page %s : %w", page.Name, err)
		}
		pageIDs = append(pageIDs, p.ID)
	}

	// all user roles
	urs, err := busDomain.UserRole.Query(ctx, userrolebus.QueryFilter{}, userrolebus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		return fmt.Errorf("querying user roles : %w", err)
	}

	r, err := busDomain.Role.QueryByID(ctx, urs[0].RoleID)
	if err != nil {
		return fmt.Errorf("querying role : %w", err)
	}

	// Add all pages to role
	for i := range allPages {
		_, err = busDomain.RolePage.Create(ctx, rolepagebus.NewRolePage{
			RoleID:    r.ID,
			PageID:    pageIDs[i],
			CanAccess: true,
		})
		if err != nil {
			return fmt.Errorf("creating role-page association : %w", err)
		}
	}

	return nil
}
