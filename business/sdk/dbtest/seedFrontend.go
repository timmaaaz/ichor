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
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitemstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
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
	"github.com/timmaaaz/ichor/business/sdk/dbtest/seedmodels"
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

	// Purchase Order Statuses
	poStatuses, err := purchaseorderstatusbus.TestSeedPurchaseOrderStatuses(ctx, 5, busDomain.PurchaseOrderStatus)
	if err != nil {
		return fmt.Errorf("seeding purchase order statuses : %w", err)
	}
	poStatusIDs := make(uuid.UUIDs, len(poStatuses))
	for i, ps := range poStatuses {
		poStatusIDs[i] = ps.ID
	}

	// Purchase Order Line Item Statuses
	poLineItemStatuses, err := purchaseorderlineitemstatusbus.TestSeedPurchaseOrderLineItemStatuses(ctx, 5, busDomain.PurchaseOrderLineItemStatus)
	if err != nil {
		return fmt.Errorf("seeding purchase order line item statuses : %w", err)
	}
	poLineItemStatusIDs := make(uuid.UUIDs, len(poLineItemStatuses))
	for i, pols := range poLineItemStatuses {
		poLineItemStatusIDs[i] = pols.ID
	}

	// Purchase Orders
	purchaseOrders, err := purchaseorderbus.TestSeedPurchaseOrders(ctx, 10, supplierIDs, poStatusIDs, warehouseIDs, strIDs, userIDs, busDomain.PurchaseOrder)
	if err != nil {
		return fmt.Errorf("seeding purchase orders : %w", err)
	}
	purchaseOrderIDs := make(uuid.UUIDs, len(purchaseOrders))
	for i, po := range purchaseOrders {
		purchaseOrderIDs[i] = po.ID
	}

	// Purchase Order Line Items
	_, err = purchaseorderlineitembus.TestSeedPurchaseOrderLineItems(ctx, 25, purchaseOrderIDs, supplierProductIDs, poLineItemStatusIDs, userIDs, busDomain.PurchaseOrderLineItem)
	if err != nil {
		return fmt.Errorf("seeding purchase order line items : %w", err)
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
	_, err = configStore.Create(ctx, "orders_dashboard", "orders_base", seedmodels.OrdersConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating stored config: %w", err)
	}

	_, err = configStore.Create(ctx, "products_dashboard", "products", seedmodels.PageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating stored config: %w", err)
	}

	_, err = configStore.Create(ctx, "inventory_dashboard", "inventory_items", seedmodels.ComplexConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating stored config: %w", err)
	}

	// Create dedicated page configs for orders, suppliers, categories, and order line items
	_, err = configStore.Create(ctx, "orders_page", "orders", seedmodels.OrdersPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating orders page config: %w", err)
	}

	_, err = configStore.Create(ctx, "suppliers_page", "suppliers", seedmodels.SuppliersPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating suppliers page config: %w", err)
	}

	_, err = configStore.Create(ctx, "categories_page", "product_categories", seedmodels.CategoriesPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating categories page config: %w", err)
	}

	_, err = configStore.Create(ctx, "order_line_items_page", "order_line_items", seedmodels.OrderLineItemsPageConfig, admins[0].ID)
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
	_, err = configStore.Create(ctx, "assets_list_page", "assets", seedmodels.AssetsListPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating assets list page config: %w", err)
	}

	_, err = configStore.Create(ctx, "assets_requests_page", "user_assets", seedmodels.AssetsRequestsPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating assets requests page config: %w", err)
	}

	// HR Module Configs
	_, err = configStore.Create(ctx, "hr_employees_page", "users", seedmodels.HrEmployeesPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating hr employees page config: %w", err)
	}

	_, err = configStore.Create(ctx, "hr_offices_page", "offices", seedmodels.HrOfficesPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating hr offices page config: %w", err)
	}

	// Inventory Module Configs
	_, err = configStore.Create(ctx, "inventory_warehouses_page", "warehouses", seedmodels.InventoryWarehousesPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating inventory warehouses page config: %w", err)
	}

	_, err = configStore.Create(ctx, "inventory_items_page", "inventory_items", seedmodels.InventoryItemsPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating inventory items page config: %w", err)
	}

	_, err = configStore.Create(ctx, "inventory_adjustments_page", "inventory_adjustments", seedmodels.InventoryAdjustmentsPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating inventory adjustments page config: %w", err)
	}

	_, err = configStore.Create(ctx, "inventory_transfers_page", "transfer_orders", seedmodels.InventoryTransfersPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating inventory transfers page config: %w", err)
	}

	// Sales Module Configs
	_, err = configStore.Create(ctx, "sales_customers_page", "customers", seedmodels.SalesCustomersPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating sales customers page config: %w", err)
	}

	// Procurement Module Configs
	_, err = configStore.Create(ctx, "procurement_purchase_orders_config", "purchase_orders", seedmodels.PurchaseOrderPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating procurement purchase orders config: %w", err)
	}

	_, err = configStore.Create(ctx, "procurement_line_items_config", "purchase_order_line_items", seedmodels.PurchaseOrderLineItemPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating procurement line items config: %w", err)
	}

	_, err = configStore.Create(ctx, "procurement_approvals_open_config", "purchase_orders", seedmodels.ProcurementOpenApprovalsPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating procurement approvals open config: %w", err)
	}

	_, err = configStore.Create(ctx, "procurement_approvals_closed_config", "purchase_orders", seedmodels.ProcurementClosedApprovalsPageConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating procurement approvals closed config: %w", err)
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

	// Query Procurement Module Configs
	procurementPurchaseOrdersConfigStored, err := configStore.QueryByName(ctx, "procurement_purchase_orders_config")
	if err != nil {
		return fmt.Errorf("querying procurement purchase orders config: %w", err)
	}

	procurementLineItemsConfigStored, err := configStore.QueryByName(ctx, "procurement_line_items_config")
	if err != nil {
		return fmt.Errorf("querying procurement line items config: %w", err)
	}

	procurementApprovalsOpenConfigStored, err := configStore.QueryByName(ctx, "procurement_approvals_open_config")
	if err != nil {
		return fmt.Errorf("querying procurement approvals open config: %w", err)
	}

	procurementApprovalsClosedConfigStored, err := configStore.QueryByName(ctx, "procurement_approvals_closed_config")
	if err != nil {
		return fmt.Errorf("querying procurement approvals closed config: %w", err)
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
	// Create Procurement Module Pages
	// =========================================================================

	// Purchase Orders Page
	procurementPurchaseOrdersPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "procurement_purchase_orders",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating procurement purchase orders page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Purchase Orders",
		PageConfigID: procurementPurchaseOrdersPage.ID,
		ConfigID:     procurementPurchaseOrdersConfigStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating procurement purchase orders page tab: %w", err)
	}

	// Purchase Order Line Items Page
	procurementLineItemsPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "procurement_line_items",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating procurement line items page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Line Items",
		PageConfigID: procurementLineItemsPage.ID,
		ConfigID:     procurementLineItemsConfigStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating procurement line items page tab: %w", err)
	}

	// Procurement Approvals Page (multi-tab: open, closed)
	procurementApprovalsPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "procurement_approvals",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating procurement approvals page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Open",
		PageConfigID: procurementApprovalsPage.ID,
		ConfigID:     procurementApprovalsOpenConfigStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating procurement approvals open tab: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Closed",
		PageConfigID: procurementApprovalsPage.ID,
		ConfigID:     procurementApprovalsClosedConfigStored.ID,
		IsDefault:    false,
		TabOrder:     2,
	})
	if err != nil {
		return fmt.Errorf("creating procurement approvals closed tab: %w", err)
	}

	// Procurement Dashboard (multi-tab: purchase orders, line items, suppliers, approvals)
	procurementDashboardPage, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
		Name:      "procurement_dashboard",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating procurement dashboard page: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Purchase Orders",
		PageConfigID: procurementDashboardPage.ID,
		ConfigID:     procurementPurchaseOrdersConfigStored.ID,
		IsDefault:    true,
		TabOrder:     1,
	})
	if err != nil {
		return fmt.Errorf("creating procurement dashboard purchase orders tab: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Line Items",
		PageConfigID: procurementDashboardPage.ID,
		ConfigID:     procurementLineItemsConfigStored.ID,
		IsDefault:    false,
		TabOrder:     2,
	})
	if err != nil {
		return fmt.Errorf("creating procurement dashboard line items tab: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Suppliers",
		PageConfigID: procurementDashboardPage.ID,
		ConfigID:     suppliersPageStored.ID,
		IsDefault:    false,
		TabOrder:     3,
	})
	if err != nil {
		return fmt.Errorf("creating procurement dashboard suppliers tab: %w", err)
	}

	_, err = configStore.CreatePageTabConfig(ctx, tablebuilder.PageTabConfig{
		Label:        "Approvals",
		PageConfigID: procurementDashboardPage.ID,
		ConfigID:     procurementApprovalsOpenConfigStored.ID,
		IsDefault:    false,
		TabOrder:     4,
	})
	if err != nil {
		return fmt.Errorf("creating procurement dashboard approvals tab: %w", err)
	}

	// =========================================================================

	// Form 1: Single entity - Users only (using generator)
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

	userFormFields := seedmodels.GetUserFormFields(userForm.ID, userEntity.ID)
	for _, ff := range userFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating user form field : %w", err)
		}
	}

	// Form 2: Single entity - Assets only (using generator)
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

	assetFormFields := seedmodels.GetAssetFormFields(assetForm.ID, assetEntity.ID)
	for _, ff := range assetFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
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
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			EntitySchema: "core",
			EntityTable:  "users",
			Name:         "username",
			FieldOrder:   1,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			EntitySchema: "core",
			EntityTable:  "users",
			Name:         "first_name",
			FieldOrder:   2,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			EntitySchema: "core",
			EntityTable:  "users",
			Name:         "last_name",
			FieldOrder:   3,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			EntitySchema: "core",
			EntityTable:  "users",
			Name:         "email",
			FieldOrder:   4,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			EntitySchema: "core",
			EntityTable:  "users",
			Name:         "password",
			FieldOrder:   5,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			EntitySchema: "core",
			EntityTable:  "users",
			Name:         "password_confirm",
			FieldOrder:   6,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			EntitySchema: "core",
			EntityTable:  "users",
			Name:         "birthday",
			FieldOrder:   7,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			EntitySchema: "core",
			EntityTable:  "users",
			Name:         "roles",
			FieldOrder:   8,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			EntitySchema: "core",
			EntityTable:  "users",
			Name:         "system_roles",
			FieldOrder:   9,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			EntitySchema: "core",
			EntityTable:  "users",
			Name:         "enabled",
			FieldOrder:   10,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			EntitySchema: "core",
			EntityTable:  "users",
			Name:         "requested_by",
			FieldOrder:   11,
			Config:       json.RawMessage(`{}`),
		},
		// Asset fields (order 12-14)
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     assetEntity.ID,
			EntitySchema: "assets",
			EntityTable:  "assets",
			Name:         "asset_condition_id",
			FieldOrder:   12,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     assetEntity.ID,
			EntitySchema: "assets",
			EntityTable:  "assets",
			Name:         "valid_asset_id",
			FieldOrder:   13,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     assetEntity.ID,
			EntitySchema: "assets",
			EntityTable:  "assets",
			Name:         "serial_number",
			FieldOrder:   14,
			Config:       json.RawMessage(`{}`),
		},
	}

	for _, ff := range multiFormFields {
		_, err = busDomain.FormField.Create(ctx, formfieldbus.NewFormField{
			FormID:       ff.FormID,
			EntityID:     ff.EntityID,
			EntitySchema: ff.EntitySchema,
			EntityTable:  ff.EntityTable,
			Name:         ff.Name,
			FieldOrder:   ff.FieldOrder,
		})
		if err != nil {
			return fmt.Errorf("creating multi-entity form field : %w", err)
		}
	}

	// =============================================================================
	// COMPOSITE FORMS
	// =============================================================================

	// Composite Form 1: Full Customer (Customer + Contact Info + Delivery Address)
	fullCustomerForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Full Customer Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating full customer form : %w", err)
	}

	customerEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "customers")
	if err != nil {
		return fmt.Errorf("querying customer entity : %w", err)
	}

	contactInfoEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "contact_infos")
	if err != nil {
		return fmt.Errorf("querying contact_infos entity : %w", err)
	}

	streetEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "streets")
	if err != nil {
		return fmt.Errorf("querying streets entity : %w", err)
	}

	fullCustomerFields := seedmodels.GetFullCustomerFormFields(
		fullCustomerForm.ID,
		customerEntity.ID,
		contactInfoEntity.ID,
		streetEntity.ID,
	)

	for _, ff := range fullCustomerFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating full customer form field : %w", err)
		}
	}

	// Composite Form 2: Full Supplier (Supplier + Contact Info)
	fullSupplierForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Full Supplier Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating full supplier form : %w", err)
	}

	supplierEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "suppliers")
	if err != nil {
		return fmt.Errorf("querying supplier entity : %w", err)
	}

	fullSupplierFields := seedmodels.GetFullSupplierFormFields(
		fullSupplierForm.ID,
		supplierEntity.ID,
		contactInfoEntity.ID,
	)

	for _, ff := range fullSupplierFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating full supplier form field : %w", err)
		}
	}

	// Composite Form 3: Full Sales Order (Order + Line Items)
	fullSalesOrderForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Full Sales Order Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating full sales order form : %w", err)
	}

	orderEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "orders")
	if err != nil {
		return fmt.Errorf("querying orders entity : %w", err)
	}

	orderLineItemEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "order_line_items")
	if err != nil {
		return fmt.Errorf("querying order_line_items entity : %w", err)
	}

	fullSalesOrderFields := seedmodels.GetFullSalesOrderFormFields(
		fullSalesOrderForm.ID,
		orderEntity.ID,
		orderLineItemEntity.ID,
	)

	for _, ff := range fullSalesOrderFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating full sales order form field : %w", err)
		}
	}

	// Composite Form 4: Full Purchase Order (PO + Line Items)
	fullPurchaseOrderForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Full Purchase Order Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating full purchase order form : %w", err)
	}

	purchaseOrderEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "purchase_orders")
	if err != nil {
		return fmt.Errorf("querying purchase_orders entity : %w", err)
	}

	poLineItemEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "purchase_order_line_items")
	if err != nil {
		return fmt.Errorf("querying purchase_order_line_items entity : %w", err)
	}

	fullPurchaseOrderFields := seedmodels.GetFullPurchaseOrderFormFields(
		fullPurchaseOrderForm.ID,
		purchaseOrderEntity.ID,
		poLineItemEntity.ID,
	)

	for _, ff := range fullPurchaseOrderFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating full purchase order form field : %w", err)
		}
	}

	// =============================================================================
	// SIMPLE FORMS (Dropdown-based for foreign keys)
	// =============================================================================

	// Simple Form 1: Role
	roleForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Role Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating role form : %w", err)
	}

	roleEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "roles")
	if err != nil {
		return fmt.Errorf("querying roles entity : %w", err)
	}

	roleFields := seedmodels.GetRoleFormFields(roleForm.ID, roleEntity.ID)
	for _, ff := range roleFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating role form field : %w", err)
		}
	}

	// Simple Form 2: Customer (dropdown version)
	simpleCustomerForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Customer Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating simple customer form : %w", err)
	}

	simpleCustomerFields := seedmodels.GetCustomerFormFields(simpleCustomerForm.ID, customerEntity.ID)
	for _, ff := range simpleCustomerFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating simple customer form field : %w", err)
		}
	}

	// Simple Form 3: Sales Order (dropdown version)
	simpleSalesOrderForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Sales Order Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating simple sales order form : %w", err)
	}

	simpleSalesOrderFields := seedmodels.GetSalesOrderFormFields(simpleSalesOrderForm.ID, orderEntity.ID)
	for _, ff := range simpleSalesOrderFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating simple sales order form field : %w", err)
		}
	}

	// Simple Form 4: Supplier (dropdown version)
	simpleSupplierForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Supplier Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating simple supplier form : %w", err)
	}

	simpleSupplierFields := seedmodels.GetSupplierFormFields(simpleSupplierForm.ID, supplierEntity.ID)
	for _, ff := range simpleSupplierFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating simple supplier form field : %w", err)
		}
	}

	// Simple Form 5: Purchase Order (dropdown version)
	simplePurchaseOrderForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Purchase Order Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating simple purchase order form : %w", err)
	}

	simplePurchaseOrderFields := seedmodels.GetPurchaseOrderFormFields(simplePurchaseOrderForm.ID, purchaseOrderEntity.ID)
	for _, ff := range simplePurchaseOrderFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating simple purchase order form field : %w", err)
		}
	}

	// Simple Form 6: Warehouse
	warehouseForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Warehouse Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating warehouse form : %w", err)
	}

	warehouseEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "warehouses")
	if err != nil {
		return fmt.Errorf("querying warehouses entity : %w", err)
	}

	warehouseFields := seedmodels.GetWarehouseFormFields(warehouseForm.ID, warehouseEntity.ID)
	for _, ff := range warehouseFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating warehouse form field : %w", err)
		}
	}

	// Simple Form 7: Inventory Adjustment
	inventoryAdjustmentForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Inventory Adjustment Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating inventory adjustment form : %w", err)
	}

	inventoryAdjustmentEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "inventory_adjustments")
	if err != nil {
		return fmt.Errorf("querying inventory_adjustments entity : %w", err)
	}

	inventoryAdjustmentFields := seedmodels.GetInventoryAdjustmentFormFields(inventoryAdjustmentForm.ID, inventoryAdjustmentEntity.ID)
	for _, ff := range inventoryAdjustmentFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating inventory adjustment form field : %w", err)
		}
	}

	// Simple Form 8: Transfer Order
	transferOrderForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Transfer Order Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating transfer order form : %w", err)
	}

	transferOrderEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "transfer_orders")
	if err != nil {
		return fmt.Errorf("querying transfer_orders entity : %w", err)
	}

	transferOrderFields := seedmodels.GetTransferOrderFormFields(transferOrderForm.ID, transferOrderEntity.ID)
	for _, ff := range transferOrderFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating transfer order form field : %w", err)
		}
	}

	// Simple Form 9: Inventory Item
	inventoryItemForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Inventory Item Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating inventory item form : %w", err)
	}

	inventoryItemEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "inventory_items")
	if err != nil {
		return fmt.Errorf("querying inventory_items entity : %w", err)
	}

	inventoryItemFields := seedmodels.GetInventoryItemFormFields(inventoryItemForm.ID, inventoryItemEntity.ID)
	for _, ff := range inventoryItemFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating inventory item form field : %w", err)
		}
	}

	// Simple Form 10: Office
	officeForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Office Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating office form : %w", err)
	}

	officeEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "offices")
	if err != nil {
		return fmt.Errorf("querying offices entity : %w", err)
	}

	officeFields := seedmodels.GetOfficeFormFields(officeForm.ID, officeEntity.ID)
	for _, ff := range officeFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating office form field : %w", err)
		}
	}

	// =============================================================================
	// REFERENCE DATA FORMS (Admin-managed, no inline creation)
	// =============================================================================

	// Reference Form 1: Country
	countryForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Country Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating country form : %w", err)
	}

	countryEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "countries")
	if err != nil {
		return fmt.Errorf("querying countries entity : %w", err)
	}

	countryFields := seedmodels.GetCountryFormFields(countryForm.ID, countryEntity.ID)
	for _, ff := range countryFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating country form field : %w", err)
		}
	}

	// Reference Form 2: Region
	regionForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Region Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating region form : %w", err)
	}

	regionEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "regions")
	if err != nil {
		return fmt.Errorf("querying regions entity : %w", err)
	}

	regionFields := seedmodels.GetRegionFormFields(regionForm.ID, regionEntity.ID)
	for _, ff := range regionFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating region form field : %w", err)
		}
	}

	// Reference Form 3: User Approval Status
	userApprovalStatusForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "User Approval Status Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating user approval status form : %w", err)
	}

	userApprovalStatusEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "user_approval_status")
	if err != nil {
		return fmt.Errorf("querying user_approval_status entity : %w", err)
	}

	userApprovalStatusFields := seedmodels.GetUserApprovalStatusFormFields(userApprovalStatusForm.ID, userApprovalStatusEntity.ID)
	for _, ff := range userApprovalStatusFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating user approval status form field : %w", err)
		}
	}

	// Reference Form 4: Asset Approval Status
	assetApprovalStatusForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Asset Approval Status Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating asset approval status form : %w", err)
	}

	assetApprovalStatusEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "approval_status")
	if err != nil {
		return fmt.Errorf("querying approval_status entity : %w", err)
	}

	assetApprovalStatusFields := seedmodels.GetAssetApprovalStatusFormFields(assetApprovalStatusForm.ID, assetApprovalStatusEntity.ID)
	for _, ff := range assetApprovalStatusFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating asset approval status form field : %w", err)
		}
	}

	// Reference Form 5: Asset Fulfillment Status
	assetFulfillmentStatusForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Asset Fulfillment Status Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating asset fulfillment status form : %w", err)
	}

	assetFulfillmentStatusEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "fulfillment_status")
	if err != nil {
		return fmt.Errorf("querying fulfillment_status entity : %w", err)
	}

	assetFulfillmentStatusFields := seedmodels.GetAssetFulfillmentStatusFormFields(assetFulfillmentStatusForm.ID, assetFulfillmentStatusEntity.ID)
	for _, ff := range assetFulfillmentStatusFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating asset fulfillment status form field : %w", err)
		}
	}

	// Reference Form 6: Order Fulfillment Status
	orderFulfillmentStatusForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Order Fulfillment Status Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating order fulfillment status form : %w", err)
	}

	orderFulfillmentStatusEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "order_fulfillment_statuses")
	if err != nil {
		return fmt.Errorf("querying order_fulfillment_statuses entity : %w", err)
	}

	orderFulfillmentStatusFields := seedmodels.GetOrderFulfillmentStatusFormFields(orderFulfillmentStatusForm.ID, orderFulfillmentStatusEntity.ID)
	for _, ff := range orderFulfillmentStatusFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating order fulfillment status form field : %w", err)
		}
	}

	// Reference Form 7: Line Item Fulfillment Status
	lineItemFulfillmentStatusForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Line Item Fulfillment Status Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating line item fulfillment status form : %w", err)
	}

	lineItemFulfillmentStatusEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "line_item_fulfillment_statuses")
	if err != nil {
		return fmt.Errorf("querying line_item_fulfillment_statuses entity : %w", err)
	}

	lineItemFulfillmentStatusFields := seedmodels.GetLineItemFulfillmentStatusFormFields(lineItemFulfillmentStatusForm.ID, lineItemFulfillmentStatusEntity.ID)
	for _, ff := range lineItemFulfillmentStatusFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating line item fulfillment status form field : %w", err)
		}
	}

	// Reference Form 8: Purchase Order Status
	purchaseOrderStatusForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Purchase Order Status Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating purchase order status form : %w", err)
	}

	purchaseOrderStatusEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "purchase_order_statuses")
	if err != nil {
		return fmt.Errorf("querying purchase_order_statuses entity : %w", err)
	}

	purchaseOrderStatusFields := seedmodels.GetPurchaseOrderStatusFormFields(purchaseOrderStatusForm.ID, purchaseOrderStatusEntity.ID)
	for _, ff := range purchaseOrderStatusFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating purchase order status form field : %w", err)
		}
	}

	// Reference Form 9: PO Line Item Status
	poLineItemStatusForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Purchase Order Line Item Status Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating po line item status form : %w", err)
	}

	poLineItemStatusEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "purchase_order_line_item_statuses")
	if err != nil {
		return fmt.Errorf("querying purchase_order_line_item_statuses entity : %w", err)
	}

	poLineItemStatusFields := seedmodels.GetPOLineItemStatusFormFields(poLineItemStatusForm.ID, poLineItemStatusEntity.ID)
	for _, ff := range poLineItemStatusFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating po line item status form field : %w", err)
		}
	}

	// Reference Form 10: Asset Type
	assetTypeForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Asset Type Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating asset type form : %w", err)
	}

	assetTypeEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "asset_types")
	if err != nil {
		return fmt.Errorf("querying asset_types entity : %w", err)
	}

	assetTypeFields := seedmodels.GetAssetTypeFormFields(assetTypeForm.ID, assetTypeEntity.ID)
	for _, ff := range assetTypeFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating asset type form field : %w", err)
		}
	}

	// Reference Form 11: Asset Condition
	assetConditionForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Asset Condition Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating asset condition form : %w", err)
	}

	assetConditionEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "asset_conditions")
	if err != nil {
		return fmt.Errorf("querying asset_conditions entity : %w", err)
	}

	assetConditionFields := seedmodels.GetAssetConditionFormFields(assetConditionForm.ID, assetConditionEntity.ID)
	for _, ff := range assetConditionFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating asset condition form field : %w", err)
		}
	}

	// Reference Form 12: Product Category
	productCategoryForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Product Category Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating product category form : %w", err)
	}

	productCategoryEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "product_categories")
	if err != nil {
		return fmt.Errorf("querying product_categories entity : %w", err)
	}

	productCategoryFields := seedmodels.GetProductCategoryFormFields(productCategoryForm.ID, productCategoryEntity.ID)
	for _, ff := range productCategoryFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating product category form field : %w", err)
		}
	}

	// =============================================================================
	// HIGH-PRIORITY TRANSACTIONAL FORMS (Referenced in inline_create)
	// =============================================================================

	// High Priority Form 1: City
	cityForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "City Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating city form : %w", err)
	}

	cityEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "cities")
	if err != nil {
		return fmt.Errorf("querying cities entity : %w", err)
	}

	cityFields := seedmodels.GetCityFormFields(cityForm.ID, cityEntity.ID)
	for _, ff := range cityFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating city form field : %w", err)
		}
	}

	// High Priority Form 2: Street (entity already declared)
	streetForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Street Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating street form : %w", err)
	}

	streetFormFields := seedmodels.GetStreetFormFields(streetForm.ID, streetEntity.ID)
	for _, ff := range streetFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating street form field : %w", err)
		}
	}

	// High Priority Form 3: Contact Info (entity already declared)
	contactInfoForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Contact Info Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating contact info form : %w", err)
	}

	contactInfoFormFields := seedmodels.GetContactInfoFormFields(contactInfoForm.ID, contactInfoEntity.ID)
	for _, ff := range contactInfoFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating contact info form field : %w", err)
		}
	}

	// High Priority Form 4: Title
	titleForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Title Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating title form : %w", err)
	}

	titleEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "titles")
	if err != nil {
		return fmt.Errorf("querying titles entity : %w", err)
	}

	titleFormFields := seedmodels.GetTitleFormFields(titleForm.ID, titleEntity.ID)
	for _, ff := range titleFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating title form field : %w", err)
		}
	}

	// High Priority Form 5: Product
	productForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Product Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating product form : %w", err)
	}

	productEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "products")
	if err != nil {
		return fmt.Errorf("querying products entity : %w", err)
	}

	productFormFields := seedmodels.GetProductFormFields(productForm.ID, productEntity.ID)
	for _, ff := range productFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating product form field : %w", err)
		}
	}

	// High Priority Form 6: Inventory Location
	inventoryLocationForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Inventory Location Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating inventory location form : %w", err)
	}

	inventoryLocationEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "inventory_locations")
	if err != nil {
		return fmt.Errorf("querying inventory_locations entity : %w", err)
	}

	inventoryLocationFormFields := seedmodels.GetInventoryLocationFormFields(inventoryLocationForm.ID, inventoryLocationEntity.ID)
	for _, ff := range inventoryLocationFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating inventory location form field : %w", err)
		}
	}

	// High Priority Form 7: Valid Asset
	validAssetForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Valid Asset Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating valid asset form : %w", err)
	}

	validAssetEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "valid_assets")
	if err != nil {
		return fmt.Errorf("querying valid_assets entity : %w", err)
	}

	validAssetFormFields := seedmodels.GetValidAssetFormFields(validAssetForm.ID, validAssetEntity.ID)
	for _, ff := range validAssetFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating valid asset form field : %w", err)
		}
	}

	// High Priority Form 8: Supplier Product
	supplierProductForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Supplier Product Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating supplier product form : %w", err)
	}

	supplierProductEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "supplier_products")
	if err != nil {
		return fmt.Errorf("querying supplier_products entity : %w", err)
	}

	supplierProductFormFields := seedmodels.GetSupplierProductFormFields(supplierProductForm.ID, supplierProductEntity.ID)
	for _, ff := range supplierProductFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating supplier product form field : %w", err)
		}
	}

	// High Priority Form 9: Sales Order Line Item (entity already declared)
	salesOrderLineItemForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Sales Order Line Item Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating sales order line item form : %w", err)
	}

	salesOrderLineItemFormFields := seedmodels.GetSalesOrderLineItemFormFields(salesOrderLineItemForm.ID, orderLineItemEntity.ID)
	for _, ff := range salesOrderLineItemFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating sales order line item form field : %w", err)
		}
	}

	// High Priority Form 10: Purchase Order Line Item (entity already declared)
	purchaseOrderLineItemForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Purchase Order Line Item Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating purchase order line item form : %w", err)
	}

	purchaseOrderLineItemFormFields := seedmodels.GetPurchaseOrderLineItemFormFields(purchaseOrderLineItemForm.ID, poLineItemEntity.ID)
	for _, ff := range purchaseOrderLineItemFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating purchase order line item form field : %w", err)
		}
	}

	// =============================================================================
	// MEDIUM-PRIORITY TRANSACTIONAL FORMS (Domain completeness)
	// =============================================================================

	// Medium Priority Form 1: Home
	homeForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Home Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating home form : %w", err)
	}

	homeEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "homes")
	if err != nil {
		return fmt.Errorf("querying homes entity : %w", err)
	}

	homeFormFields := seedmodels.GetHomeFormFields(homeForm.ID, homeEntity.ID)
	for _, ff := range homeFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating home form field : %w", err)
		}
	}

	// Medium Priority Form 2: Tag
	tagForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Tag Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating tag form : %w", err)
	}

	tagEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "tags")
	if err != nil {
		return fmt.Errorf("querying tags entity : %w", err)
	}

	tagFormFields := seedmodels.GetTagFormFields(tagForm.ID, tagEntity.ID)
	for _, ff := range tagFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating tag form field : %w", err)
		}
	}

	// Medium Priority Form 3: User Approval Comment
	userApprovalCommentForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "User Approval Comment Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating user approval comment form : %w", err)
	}

	userApprovalCommentEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "user_approval_comments")
	if err != nil {
		return fmt.Errorf("querying user_approval_comments entity : %w", err)
	}

	userApprovalCommentFormFields := seedmodels.GetUserApprovalCommentFormFields(userApprovalCommentForm.ID, userApprovalCommentEntity.ID)
	for _, ff := range userApprovalCommentFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating user approval comment form field : %w", err)
		}
	}

	// Medium Priority Form 4: Brand
	brandForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Brand Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating brand form : %w", err)
	}

	brandEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "brands")
	if err != nil {
		return fmt.Errorf("querying brands entity : %w", err)
	}

	brandFormFields := seedmodels.GetBrandFormFields(brandForm.ID, brandEntity.ID)
	for _, ff := range brandFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating brand form field : %w", err)
		}
	}

	// Medium Priority Form 5: Zone
	zoneForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Zone Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating zone form : %w", err)
	}

	zoneEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "zones")
	if err != nil {
		return fmt.Errorf("querying zones entity : %w", err)
	}

	zoneFormFields := seedmodels.GetZoneFormFields(zoneForm.ID, zoneEntity.ID)
	for _, ff := range zoneFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating zone form field : %w", err)
		}
	}

	// Medium Priority Form 6: User Asset
	userAssetForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "User Asset Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating user asset form : %w", err)
	}

	userAssetEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "user_assets")
	if err != nil {
		return fmt.Errorf("querying user_assets entity : %w", err)
	}

	userAssetFormFields := seedmodels.GetUserAssetFormFields(userAssetForm.ID, userAssetEntity.ID)
	for _, ff := range userAssetFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating user asset form field : %w", err)
		}
	}

	// Medium Priority Form 7: Automation Rule
	automationRuleForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Automation Rule Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating automation rule form : %w", err)
	}

	automationRuleEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "automation_rules")
	if err != nil {
		return fmt.Errorf("querying automation_rules entity : %w", err)
	}

	automationRuleFormFields := seedmodels.GetAutomationRuleFormFields(automationRuleForm.ID, automationRuleEntity.ID)
	for _, ff := range automationRuleFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating automation rule form field : %w", err)
		}
	}

	// Medium Priority Form 8: Rule Action
	ruleActionForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Rule Action Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating rule action form : %w", err)
	}

	ruleActionEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "rule_actions")
	if err != nil {
		return fmt.Errorf("querying rule_actions entity : %w", err)
	}

	ruleActionFormFields := seedmodels.GetRuleActionFormFields(ruleActionForm.ID, ruleActionEntity.ID)
	for _, ff := range ruleActionFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating rule action form field : %w", err)
		}
	}

	// Medium Priority Form 9: Entity
	entityForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Entity Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating entity form : %w", err)
	}

	entityEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "entities")
	if err != nil {
		return fmt.Errorf("querying entities entity : %w", err)
	}

	entityFormFields := seedmodels.GetEntityFormFields(entityForm.ID, entityEntity.ID)
	for _, ff := range entityFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating entity form field : %w", err)
		}
	}

	// Medium Priority Form 10: User (using proper generator instead of inline)
	userFormProp, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "User Creation Form (Proper)",
	})
	if err != nil {
		return fmt.Errorf("creating user form (proper) : %w", err)
	}

	userFormProperFields := seedmodels.GetUserFormFields(userFormProp.ID, userEntity.ID)
	for _, ff := range userFormProperFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating user form field : %w", err)
		}
	}

	// =============================================================================
	// LOWER-PRIORITY TRANSACTIONAL FORMS (Utility/tracking)
	// =============================================================================

	// Lower Priority Form 1: Physical Attribute
	physicalAttributeForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Physical Attribute Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating physical attribute form : %w", err)
	}

	physicalAttributeEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "physical_attributes")
	if err != nil {
		return fmt.Errorf("querying physical_attributes entity : %w", err)
	}

	physicalAttributeFormFields := seedmodels.GetPhysicalAttributeFormFields(physicalAttributeForm.ID, physicalAttributeEntity.ID)
	for _, ff := range physicalAttributeFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating physical attribute form field : %w", err)
		}
	}

	// Lower Priority Form 2: Product Cost
	productCostForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Product Cost Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating product cost form : %w", err)
	}

	productCostEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "product_costs")
	if err != nil {
		return fmt.Errorf("querying product_costs entity : %w", err)
	}

	productCostFormFields := seedmodels.GetProductCostFormFields(productCostForm.ID, productCostEntity.ID)
	for _, ff := range productCostFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating product cost form field : %w", err)
		}
	}

	// Lower Priority Form 3: Cost History
	costHistoryForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Cost History Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating cost history form : %w", err)
	}

	costHistoryEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "cost_history")
	if err != nil {
		return fmt.Errorf("querying cost_history entity : %w", err)
	}

	costHistoryFormFields := seedmodels.GetCostHistoryFormFields(costHistoryForm.ID, costHistoryEntity.ID)
	for _, ff := range costHistoryFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating cost history form field : %w", err)
		}
	}

	// Lower Priority Form 4: Quality Metric
	qualityMetricForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Quality Metric Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating quality metric form : %w", err)
	}

	qualityMetricEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "quality_metrics")
	if err != nil {
		return fmt.Errorf("querying quality_metrics entity : %w", err)
	}

	qualityMetricFormFields := seedmodels.GetQualityMetricFormFields(qualityMetricForm.ID, qualityMetricEntity.ID)
	for _, ff := range qualityMetricFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating quality metric form field : %w", err)
		}
	}

	// Lower Priority Form 5: Serial Number
	serialNumberForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Serial Number Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating serial number form : %w", err)
	}

	serialNumberEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "serial_numbers")
	if err != nil {
		return fmt.Errorf("querying serial_numbers entity : %w", err)
	}

	serialNumberFormFields := seedmodels.GetSerialNumberFormFields(serialNumberForm.ID, serialNumberEntity.ID)
	for _, ff := range serialNumberFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating serial number form field : %w", err)
		}
	}

	// Lower Priority Form 6: Lot Tracking
	lotTrackingForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Lot Tracking Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating lot tracking form : %w", err)
	}

	lotTrackingEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "lot_trackings")
	if err != nil {
		return fmt.Errorf("querying lot_trackings entity : %w", err)
	}

	lotTrackingFormFields := seedmodels.GetLotTrackingFormFields(lotTrackingForm.ID, lotTrackingEntity.ID)
	for _, ff := range lotTrackingFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating lot tracking form field : %w", err)
		}
	}

	// Lower Priority Form 7: Quality Inspection
	qualityInspectionForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Quality Inspection Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating quality inspection form : %w", err)
	}

	qualityInspectionEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "quality_inspections")
	if err != nil {
		return fmt.Errorf("querying quality_inspections entity : %w", err)
	}

	qualityInspectionFormFields := seedmodels.GetQualityInspectionFormFields(qualityInspectionForm.ID, qualityInspectionEntity.ID)
	for _, ff := range qualityInspectionFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating quality inspection form field : %w", err)
		}
	}

	// Lower Priority Form 8: Inventory Transaction
	inventoryTransactionForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Inventory Transaction Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating inventory transaction form : %w", err)
	}

	inventoryTransactionEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "inventory_transactions")
	if err != nil {
		return fmt.Errorf("querying inventory_transactions entity : %w", err)
	}

	inventoryTransactionFormFields := seedmodels.GetInventoryTransactionFormFields(inventoryTransactionForm.ID, inventoryTransactionEntity.ID)
	for _, ff := range inventoryTransactionFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating inventory transaction form field : %w", err)
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
