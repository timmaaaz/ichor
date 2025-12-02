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
	"github.com/timmaaaz/ichor/business/domain/config/pageconfigbus"
	"github.com/timmaaaz/ichor/business/domain/config/pagecontentbus"
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

	_, err = configStore.Create(ctx, "products_dashboard", "products", seedmodels.TableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating stored config: %w", err)
	}

	_, err = configStore.Create(ctx, "inventory_dashboard", "inventory_items", seedmodels.ComplexConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating stored config: %w", err)
	}

	// Create dedicated page configs for orders, suppliers, categories, and order line items
	_, err = configStore.Create(ctx, "orders_table", "orders", seedmodels.OrdersTableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating orders table config: %w", err)
	}

	_, err = configStore.Create(ctx, "suppliers_table", "suppliers", seedmodels.SuppliersTableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating suppliers table config: %w", err)
	}

	_, err = configStore.Create(ctx, "categories_table", "product_categories", seedmodels.CategoriesTableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating categories table config: %w", err)
	}

	_, err = configStore.Create(ctx, "order_line_items_table", "order_line_items", seedmodels.OrderLineItemsTableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating order line items table config: %w", err)
	}

	// Admin Module Configs
	_, err = configStore.Create(ctx, "admin_users_table", "users", adminUsersTableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating admin users table config: %w", err)
	}

	_, err = configStore.Create(ctx, "admin_roles_table", "roles", adminRolesTableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating admin roles table config: %w", err)
	}

	_, err = configStore.Create(ctx, "admin_table_access_table", "table_access", adminTableAccessTableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating admin table access table config: %w", err)
	}

	_, err = configStore.Create(ctx, "admin_audit_table", "automation_executions", adminAuditTableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating admin audit table config: %w", err)
	}

	_, err = configStore.Create(ctx, "admin_config_table", "table_configs", adminConfigTableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating admin config table config: %w", err)
	}

	// Assets Module Configs
	_, err = configStore.Create(ctx, "assets_list_table", "assets", seedmodels.AssetsListTableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating assets list table config: %w", err)
	}

	_, err = configStore.Create(ctx, "assets_requests_table", "user_assets", seedmodels.AssetsRequestsTableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating assets requests table config: %w", err)
	}

	// HR Module Configs
	_, err = configStore.Create(ctx, "hr_employees_table", "users", seedmodels.HrEmployeesTableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating hr employees table config: %w", err)
	}

	_, err = configStore.Create(ctx, "hr_offices_table", "offices", seedmodels.HrOfficesTableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating hr offices table config: %w", err)
	}

	// Inventory Module Configs
	_, err = configStore.Create(ctx, "inventory_warehouses_table", "warehouses", seedmodels.InventoryWarehousesTableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating inventory warehouses table config: %w", err)
	}

	_, err = configStore.Create(ctx, "inventory_items_table", "inventory_items", seedmodels.InventoryItemsTableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating inventory items table config: %w", err)
	}

	_, err = configStore.Create(ctx, "inventory_adjustments_table", "inventory_adjustments", seedmodels.InventoryAdjustmentsTableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating inventory adjustments table config: %w", err)
	}

	_, err = configStore.Create(ctx, "inventory_transfers_table", "transfer_orders", seedmodels.InventoryTransfersTableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating inventory transfers table config: %w", err)
	}

	// Sales Module Configs
	_, err = configStore.Create(ctx, "sales_customers_table", "customers", seedmodels.SalesCustomersTableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating sales customers table config: %w", err)
	}

	// Procurement Module Configs
	_, err = configStore.Create(ctx, "procurement_purchase_orders_config", "purchase_orders", seedmodels.PurchaseOrderTableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating procurement purchase orders config: %w", err)
	}

	_, err = configStore.Create(ctx, "procurement_line_items_config", "purchase_order_line_items", seedmodels.PurchaseOrderLineItemTableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating procurement line items config: %w", err)
	}

	_, err = configStore.Create(ctx, "procurement_approvals_open_config", "purchase_orders", seedmodels.ProcurementOpenApprovalsTableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating procurement approvals open config: %w", err)
	}

	_, err = configStore.Create(ctx, "procurement_approvals_closed_config", "purchase_orders", seedmodels.ProcurementClosedApprovalsTableConfig, admins[0].ID)
	if err != nil {
		return fmt.Errorf("creating procurement approvals closed config: %w", err)
	}

	// =========================================================================
	// Chart Configurations - 14 seed charts covering all chart types
	// =========================================================================
	for _, chartConfig := range seedmodels.ChartConfigs {
		_, err = configStore.Create(ctx, chartConfig.Name, chartConfig.Description, chartConfig.Config, admins[0].ID)
		if err != nil {
			return fmt.Errorf("creating chart config %s: %w", chartConfig.Name, err)
		}
	}

	// =========================================================================
	// Create dedicated page configs for Orders, Suppliers, and Categories
	// =========================================================================

	// Get the stored config IDs for the new pages
	ordersTableStored, err := configStore.QueryByName(ctx, "orders_table")
	if err != nil {
		return fmt.Errorf("querying orders table config: %w", err)
	}

	suppliersTableStored, err := configStore.QueryByName(ctx, "suppliers_table")
	if err != nil {
		return fmt.Errorf("querying suppliers table config: %w", err)
	}

	categoriesTableStored, err := configStore.QueryByName(ctx, "categories_table")
	if err != nil {
		return fmt.Errorf("querying categories table config: %w", err)
	}

	orderLineItemsTableStored, err := configStore.QueryByName(ctx, "order_line_items_table")
	if err != nil {
		return fmt.Errorf("querying order line items table config: %w", err)
	}

	// Query Admin Module Configs
	adminUsersTableStored, err := configStore.QueryByName(ctx, "admin_users_table")
	if err != nil {
		return fmt.Errorf("querying admin users table config: %w", err)
	}

	adminRolesTableStored, err := configStore.QueryByName(ctx, "admin_roles_table")
	if err != nil {
		return fmt.Errorf("querying admin roles table config: %w", err)
	}

	adminTableAccessTableStored, err := configStore.QueryByName(ctx, "admin_table_access_table")
	if err != nil {
		return fmt.Errorf("querying admin table access table config: %w", err)
	}

	adminAuditTableStored, err := configStore.QueryByName(ctx, "admin_audit_table")
	if err != nil {
		return fmt.Errorf("querying admin audit table config: %w", err)
	}

	adminConfigTableStored, err := configStore.QueryByName(ctx, "admin_config_table")
	if err != nil {
		return fmt.Errorf("querying admin config table config: %w", err)
	}

	// Query Assets Module Configs
	assetsListTableStored, err := configStore.QueryByName(ctx, "assets_list_table")
	if err != nil {
		return fmt.Errorf("querying assets list table config: %w", err)
	}

	assetsRequestsTableStored, err := configStore.QueryByName(ctx, "assets_requests_table")
	if err != nil {
		return fmt.Errorf("querying assets requests table config: %w", err)
	}

	// Query HR Module Configs
	hrEmployeesTableStored, err := configStore.QueryByName(ctx, "hr_employees_table")
	if err != nil {
		return fmt.Errorf("querying hr employees table config: %w", err)
	}

	hrOfficesTableStored, err := configStore.QueryByName(ctx, "hr_offices_table")
	if err != nil {
		return fmt.Errorf("querying hr offices table config: %w", err)
	}

	// Query Inventory Module Configs
	inventoryWarehousesTableStored, err := configStore.QueryByName(ctx, "inventory_warehouses_table")
	if err != nil {
		return fmt.Errorf("querying inventory warehouses table config: %w", err)
	}

	inventoryItemsTableStored, err := configStore.QueryByName(ctx, "inventory_items_table")
	if err != nil {
		return fmt.Errorf("querying inventory items table config: %w", err)
	}

	inventoryAdjustmentsTableStored, err := configStore.QueryByName(ctx, "inventory_adjustments_table")
	if err != nil {
		return fmt.Errorf("querying inventory adjustments table config: %w", err)
	}

	inventoryTransfersTableStored, err := configStore.QueryByName(ctx, "inventory_transfers_table")
	if err != nil {
		return fmt.Errorf("querying inventory transfers table config: %w", err)
	}

	// Query Sales Module Configs
	salesCustomersTableStored, err := configStore.QueryByName(ctx, "sales_customers_table")
	if err != nil {
		return fmt.Errorf("querying sales customers table config: %w", err)
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

	// Query Chart Configs for distribution across pages
	// These use _ for error since charts are optional - pages work without them
	kpiRevenueStored, _ := configStore.QueryByName(ctx, "seed_kpi_total_revenue")
	kpiOrdersStored, _ := configStore.QueryByName(ctx, "seed_kpi_order_count")
	gaugeRevenueStored, _ := configStore.QueryByName(ctx, "seed_gauge_revenue_target")
	lineMonthlySalesStored, _ := configStore.QueryByName(ctx, "seed_line_monthly_sales")
	barTopProductsStored, _ := configStore.QueryByName(ctx, "seed_bar_top_products")
	pieRevenueCategoryStored, _ := configStore.QueryByName(ctx, "seed_pie_revenue_category")
	funnelPipelineStored, _ := configStore.QueryByName(ctx, "seed_funnel_pipeline")
	heatmapSalesTimeStored, _ := configStore.QueryByName(ctx, "seed_heatmap_sales_time")
	treemapRevenueStored, _ := configStore.QueryByName(ctx, "seed_treemap_revenue")
	ganttProjectStored, _ := configStore.QueryByName(ctx, "seed_gantt_project")

	// Create Orders Page
	ordersPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "orders_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating orders page: %w", err)
	}

	ordersPageOrderIndex := 1

	// Add charts to Orders Page
	if lineMonthlySalesStored != nil {
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  ordersPage.ID,
			ContentType:   pagecontentbus.ContentTypeChart,
			Label:         "Monthly Sales Trend",
			ChartConfigID: lineMonthlySalesStored.ID,
			OrderIndex:    ordersPageOrderIndex,
			Layout:        json.RawMessage(`{"colSpan":{"xs":8,"sm":8,"md":8,"lg":8,"xl":8}}`),
			IsVisible:     true,
			IsDefault:     false,
		})
		if err != nil {
			return fmt.Errorf("creating orders page line chart: %w", err)
		}
		ordersPageOrderIndex++
	}

	if funnelPipelineStored != nil {
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  ordersPage.ID,
			ContentType:   pagecontentbus.ContentTypeChart,
			Label:         "Sales Pipeline",
			ChartConfigID: funnelPipelineStored.ID,
			OrderIndex:    ordersPageOrderIndex,
			Layout:        json.RawMessage(`{"colSpan":{"xs":4,"sm":4,"md":4,"lg":4,"xl":4}}`),
			IsVisible:     true,
			IsDefault:     false,
		})
		if err != nil {
			return fmt.Errorf("creating orders page funnel chart: %w", err)
		}
		ordersPageOrderIndex++
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  ordersPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: ordersTableStored.ID,
		OrderIndex:    ordersPageOrderIndex,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating orders page content: %w", err)
	}

	// Create Suppliers Page
	suppliersPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "suppliers_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating suppliers page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  suppliersPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: suppliersTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating suppliers page content: %w", err)
	}

	// Create Categories Page
	categoriesPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "categories_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating categories page: %w", err)
	}

	categoriesPageOrderIndex := 1

	// Add charts to Categories Page
	if pieRevenueCategoryStored != nil {
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  categoriesPage.ID,
			ContentType:   pagecontentbus.ContentTypeChart,
			Label:         "Revenue by Category",
			ChartConfigID: pieRevenueCategoryStored.ID,
			OrderIndex:    categoriesPageOrderIndex,
			Layout:        json.RawMessage(`{"colSpan":{"xs":12,"md":4}}`),
			IsVisible:     true,
			IsDefault:     false,
		})
		if err != nil {
			return fmt.Errorf("creating categories page pie chart: %w", err)
		}
		categoriesPageOrderIndex++
	}

	if barTopProductsStored != nil {
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  categoriesPage.ID,
			ContentType:   pagecontentbus.ContentTypeChart,
			Label:         "Top Products",
			ChartConfigID: barTopProductsStored.ID,
			OrderIndex:    categoriesPageOrderIndex,
			Layout:        json.RawMessage(`{"colSpan":{"xs":12,"md":4}}`),
			IsVisible:     true,
			IsDefault:     false,
		})
		if err != nil {
			return fmt.Errorf("creating categories page bar chart: %w", err)
		}
		categoriesPageOrderIndex++
	}

	if treemapRevenueStored != nil {
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  categoriesPage.ID,
			ContentType:   pagecontentbus.ContentTypeChart,
			Label:         "Revenue Breakdown",
			ChartConfigID: treemapRevenueStored.ID,
			OrderIndex:    categoriesPageOrderIndex,
			Layout:        json.RawMessage(`{"colSpan":{"xs":12,"md":4}}`),
			IsVisible:     true,
			IsDefault:     false,
		})
		if err != nil {
			return fmt.Errorf("creating categories page treemap chart: %w", err)
		}
		categoriesPageOrderIndex++
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  categoriesPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: categoriesTableStored.ID,
		OrderIndex:    categoriesPageOrderIndex,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating categories page content: %w", err)
	}

	// Create Order Line Items Page
	orderLineItemsPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "order_line_items_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating order line items page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  orderLineItemsPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: orderLineItemsTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating order line items page content: %w", err)
	}

	// =========================================================================
	// Create Admin Module Pages
	// =========================================================================

	// Admin Users Page
	adminUsersPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "admin_users_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating admin users page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  adminUsersPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: adminUsersTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating admin users page content: %w", err)
	}

	// Admin Roles Page
	adminRolesPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "admin_roles_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating admin roles page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  adminRolesPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: adminRolesTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating admin roles page content: %w", err)
	}

	// Admin Dashboard Page (multi-tab: users, roles, table access)
	adminDashboardPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "admin_dashboard_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating admin dashboard page: %w", err)
	}

	// Create tabs container (parent)
	adminDashboardTabsContainer, err := busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID: adminDashboardPage.ID,
		ContentType:  pagecontentbus.ContentTypeTabs,
		OrderIndex:   1,
		Layout:       json.RawMessage(`{"containerType":"tabs"}`),
		IsVisible:    true,
		IsDefault:    false,
	})
	if err != nil {
		return fmt.Errorf("creating admin dashboard tabs container: %w", err)
	}

	// Tab 1: Users
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  adminDashboardPage.ID,
		ParentID:      adminDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Users",
		TableConfigID: adminUsersTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating admin dashboard users tab: %w", err)
	}

	// Tab 2: Roles
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  adminDashboardPage.ID,
		ParentID:      adminDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Roles",
		TableConfigID: adminRolesTableStored.ID,
		OrderIndex:    2,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating admin dashboard roles tab: %w", err)
	}

	// Tab 3: Permissions
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  adminDashboardPage.ID,
		ParentID:      adminDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Permissions",
		TableConfigID: adminTableAccessTableStored.ID,
		OrderIndex:    3,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating admin dashboard permissions tab: %w", err)
	}

	// Tab 4: Audit Logs
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  adminDashboardPage.ID,
		ParentID:      adminDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Audit Logs",
		TableConfigID: adminAuditTableStored.ID,
		OrderIndex:    4,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating admin dashboard audit tab: %w", err)
	}

	// Tab 5: Configurations
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  adminDashboardPage.ID,
		ParentID:      adminDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Configurations",
		TableConfigID: adminConfigTableStored.ID,
		OrderIndex:    5,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating admin dashboard config tab: %w", err)
	}

	// =========================================================================
	// Create Assets Module Pages
	// =========================================================================

	// Assets List Page
	assetsPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "assets_list_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating assets page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  assetsPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: assetsListTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating assets page content: %w", err)
	}

	// Asset Requests Page
	assetsRequestsPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "assets_requests_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating assets requests page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  assetsRequestsPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: assetsRequestsTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating assets requests page content: %w", err)
	}

	// Assets Dashboard (multi-tab: assets, requests)
	assetsDashboardPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "assets_dashboard_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating assets dashboard page: %w", err)
	}

	// Create tabs container (parent)
	assetsDashboardTabsContainer, err := busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID: assetsDashboardPage.ID,
		ContentType:  pagecontentbus.ContentTypeTabs,
		OrderIndex:   1,
		Layout:       json.RawMessage(`{"containerType":"tabs"}`),
		IsVisible:    true,
		IsDefault:    false,
	})
	if err != nil {
		return fmt.Errorf("creating assets dashboard tabs container: %w", err)
	}

	// Tab 1: Assets
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  assetsDashboardPage.ID,
		ParentID:      assetsDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Assets",
		TableConfigID: assetsListTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating assets dashboard assets tab: %w", err)
	}

	// Tab 2: Requests
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  assetsDashboardPage.ID,
		ParentID:      assetsDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Requests",
		TableConfigID: assetsRequestsTableStored.ID,
		OrderIndex:    2,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating assets dashboard requests tab: %w", err)
	}

	// =========================================================================
	// Create HR Module Pages
	// =========================================================================

	// HR Employees Page
	hrEmployeesPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "hr_employees_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating hr employees page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  hrEmployeesPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: hrEmployeesTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating hr employees page content: %w", err)
	}

	// HR Offices Page
	hrOfficesPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "hr_offices_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating hr offices page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  hrOfficesPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: hrOfficesTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating hr offices page content: %w", err)
	}

	// HR Dashboard (multi-tab: employees, offices)
	hrDashboardPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "hr_dashboard_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating hr dashboard page: %w", err)
	}

	// Create tabs container (parent)
	hrDashboardTabsContainer, err := busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID: hrDashboardPage.ID,
		ContentType:  pagecontentbus.ContentTypeTabs,
		OrderIndex:   1,
		Layout:       json.RawMessage(`{"containerType":"tabs"}`),
		IsVisible:    true,
		IsDefault:    false,
	})
	if err != nil {
		return fmt.Errorf("creating hr dashboard tabs container: %w", err)
	}

	// Tab 1: Employees
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  hrDashboardPage.ID,
		ParentID:      hrDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Employees",
		TableConfigID: hrEmployeesTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating hr dashboard employees tab: %w", err)
	}

	// Tab 2: Offices
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  hrDashboardPage.ID,
		ParentID:      hrDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Offices",
		TableConfigID: hrOfficesTableStored.ID,
		OrderIndex:    2,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating hr dashboard offices tab: %w", err)
	}

	// =========================================================================
	// Create Inventory Module Pages
	// =========================================================================

	// Inventory Warehouses Page
	inventoryWarehousesPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "inventory_warehouses_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory warehouses page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  inventoryWarehousesPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: inventoryWarehousesTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory warehouses page content: %w", err)
	}

	// Inventory Items Page
	inventoryItemsPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "inventory_items_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory items page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  inventoryItemsPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: inventoryItemsTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory items page content: %w", err)
	}

	// Inventory Adjustments Page
	inventoryAdjustmentsPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "inventory_adjustments_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory adjustments page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  inventoryAdjustmentsPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: inventoryAdjustmentsTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory adjustments page content: %w", err)
	}

	// Inventory Transfers Page
	inventoryTransfersPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "inventory_transfers_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory transfers page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  inventoryTransfersPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: inventoryTransfersTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory transfers page content: %w", err)
	}

	// Inventory Dashboard (multi-tab: warehouses, items, adjustments, transfers)
	inventoryDashboardPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "inventory_dashboard_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory dashboard page: %w", err)
	}

	inventoryDashboardOrderIndex := 1

	// Add Heatmap chart to Inventory Dashboard (above tabs)
	if heatmapSalesTimeStored != nil {
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  inventoryDashboardPage.ID,
			ContentType:   pagecontentbus.ContentTypeChart,
			Label:         "Orders by Day and Hour",
			ChartConfigID: heatmapSalesTimeStored.ID,
			OrderIndex:    inventoryDashboardOrderIndex,
			Layout:        json.RawMessage(`{"colSpan":{"xs":12}}`),
			IsVisible:     true,
			IsDefault:     false,
		})
		if err != nil {
			return fmt.Errorf("creating inventory dashboard heatmap chart: %w", err)
		}
		inventoryDashboardOrderIndex++
	}

	// Create tabs container (parent)
	inventoryDashboardTabsContainer, err := busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID: inventoryDashboardPage.ID,
		ContentType:  pagecontentbus.ContentTypeTabs,
		OrderIndex:   inventoryDashboardOrderIndex,
		Layout:       json.RawMessage(`{"containerType":"tabs"}`),
		IsVisible:    true,
		IsDefault:    false,
	})
	if err != nil {
		return fmt.Errorf("creating inventory dashboard tabs container: %w", err)
	}

	// Tab 1: Items
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  inventoryDashboardPage.ID,
		ParentID:      inventoryDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Items",
		TableConfigID: inventoryItemsTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory dashboard items tab: %w", err)
	}

	// Tab 2: Warehouses
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  inventoryDashboardPage.ID,
		ParentID:      inventoryDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Warehouses",
		TableConfigID: inventoryWarehousesTableStored.ID,
		OrderIndex:    2,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating inventory dashboard warehouses tab: %w", err)
	}

	// Tab 3: Adjustments
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  inventoryDashboardPage.ID,
		ParentID:      inventoryDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Adjustments",
		TableConfigID: inventoryAdjustmentsTableStored.ID,
		OrderIndex:    3,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating inventory dashboard adjustments tab: %w", err)
	}

	// Tab 4: Transfers
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  inventoryDashboardPage.ID,
		ParentID:      inventoryDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Transfers",
		TableConfigID: inventoryTransfersTableStored.ID,
		OrderIndex:    4,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating inventory dashboard transfers tab: %w", err)
	}

	// =========================================================================
	// Create Sales Module Pages
	// =========================================================================

	// Sales Customers Page
	salesCustomersPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "sales_customers_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating sales customers page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  salesCustomersPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: salesCustomersTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating sales customers page content: %w", err)
	}

	// Sales Dashboard (multi-tab: orders, customers)
	salesDashboardPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "sales_dashboard_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating sales dashboard page: %w", err)
	}

	salesDashboardOrderIndex := 1

	// Add KPI charts row to Sales Dashboard (above tabs)
	if kpiRevenueStored != nil {
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  salesDashboardPage.ID,
			ContentType:   pagecontentbus.ContentTypeChart,
			Label:         "Total Revenue",
			ChartConfigID: kpiRevenueStored.ID,
			OrderIndex:    salesDashboardOrderIndex,
			Layout:        json.RawMessage(`{"colSpan":{"xs":12,"sm":6,"md":4}}`),
			IsVisible:     true,
			IsDefault:     false,
		})
		if err != nil {
			return fmt.Errorf("creating sales dashboard KPI revenue chart: %w", err)
		}
		salesDashboardOrderIndex++
	}

	if kpiOrdersStored != nil {
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  salesDashboardPage.ID,
			ContentType:   pagecontentbus.ContentTypeChart,
			Label:         "Total Orders",
			ChartConfigID: kpiOrdersStored.ID,
			OrderIndex:    salesDashboardOrderIndex,
			Layout:        json.RawMessage(`{"colSpan":{"xs":12,"sm":6,"md":4}}`),
			IsVisible:     true,
			IsDefault:     false,
		})
		if err != nil {
			return fmt.Errorf("creating sales dashboard KPI orders chart: %w", err)
		}
		salesDashboardOrderIndex++
	}

	if gaugeRevenueStored != nil {
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  salesDashboardPage.ID,
			ContentType:   pagecontentbus.ContentTypeChart,
			Label:         "Revenue Progress",
			ChartConfigID: gaugeRevenueStored.ID,
			OrderIndex:    salesDashboardOrderIndex,
			Layout:        json.RawMessage(`{"colSpan":{"xs":12,"sm":6,"md":4}}`),
			IsVisible:     true,
			IsDefault:     false,
		})
		if err != nil {
			return fmt.Errorf("creating sales dashboard gauge chart: %w", err)
		}
		salesDashboardOrderIndex++
	}

	// Create tabs container (parent)
	salesDashboardTabsContainer, err := busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID: salesDashboardPage.ID,
		ContentType:  pagecontentbus.ContentTypeTabs,
		OrderIndex:   salesDashboardOrderIndex,
		Layout:       json.RawMessage(`{"containerType":"tabs"}`),
		IsVisible:    true,
		IsDefault:    false,
	})
	if err != nil {
		return fmt.Errorf("creating sales dashboard tabs container: %w", err)
	}

	// Tab 1: Orders
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  salesDashboardPage.ID,
		ParentID:      salesDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Orders",
		TableConfigID: ordersTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating sales dashboard orders tab: %w", err)
	}

	// Tab 2: Customers
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  salesDashboardPage.ID,
		ParentID:      salesDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Customers",
		TableConfigID: salesCustomersTableStored.ID,
		OrderIndex:    2,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating sales dashboard customers tab: %w", err)
	}

	// =========================================================================
	// Create Procurement Module Pages
	// =========================================================================

	// Purchase Orders Page
	procurementPurchaseOrdersPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "procurement_purchase_orders",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating procurement purchase orders page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  procurementPurchaseOrdersPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: procurementPurchaseOrdersConfigStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating procurement purchase orders page content: %w", err)
	}

	// Purchase Order Line Items Page
	procurementLineItemsPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "procurement_line_items",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating procurement line items page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  procurementLineItemsPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: procurementLineItemsConfigStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating procurement line items page content: %w", err)
	}

	// Procurement Approvals Page (multi-tab: open, closed)
	procurementApprovalsPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "procurement_approvals",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating procurement approvals page: %w", err)
	}

	// Create tabs container (parent)
	procurementApprovalsTabsContainer, err := busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID: procurementApprovalsPage.ID,
		ContentType:  pagecontentbus.ContentTypeTabs,
		OrderIndex:   1,
		Layout:       json.RawMessage(`{"containerType":"tabs"}`),
		IsVisible:    true,
		IsDefault:    false,
	})
	if err != nil {
		return fmt.Errorf("creating procurement approvals tabs container: %w", err)
	}

	// Tab 1: Open
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  procurementApprovalsPage.ID,
		ParentID:      procurementApprovalsTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Open",
		TableConfigID: procurementApprovalsOpenConfigStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating procurement approvals open tab: %w", err)
	}

	// Tab 2: Closed
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  procurementApprovalsPage.ID,
		ParentID:      procurementApprovalsTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Closed",
		TableConfigID: procurementApprovalsClosedConfigStored.ID,
		OrderIndex:    2,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating procurement approvals closed tab: %w", err)
	}

	// Procurement Dashboard (multi-tab: purchase orders, line items, suppliers, approvals)
	procurementDashboardPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "procurement_dashboard",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating procurement dashboard page: %w", err)
	}

	procurementDashboardOrderIndex := 1

	// Add Gantt chart to Procurement Dashboard (above tabs)
	if ganttProjectStored != nil {
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  procurementDashboardPage.ID,
			ContentType:   pagecontentbus.ContentTypeChart,
			Label:         "Purchase Order Timeline",
			ChartConfigID: ganttProjectStored.ID,
			OrderIndex:    procurementDashboardOrderIndex,
			Layout:        json.RawMessage(`{"colSpan":{"xs":12}}`),
			IsVisible:     true,
			IsDefault:     false,
		})
		if err != nil {
			return fmt.Errorf("creating procurement dashboard gantt chart: %w", err)
		}
		procurementDashboardOrderIndex++
	}

	// Create tabs container (parent)
	procurementDashboardTabsContainer, err := busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID: procurementDashboardPage.ID,
		ContentType:  pagecontentbus.ContentTypeTabs,
		OrderIndex:   procurementDashboardOrderIndex,
		Layout:       json.RawMessage(`{"containerType":"tabs"}`),
		IsVisible:    true,
		IsDefault:    false,
	})
	if err != nil {
		return fmt.Errorf("creating procurement dashboard tabs container: %w", err)
	}

	// Tab 1: Purchase Orders
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  procurementDashboardPage.ID,
		ParentID:      procurementDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Purchase Orders",
		TableConfigID: procurementPurchaseOrdersConfigStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating procurement dashboard purchase orders tab: %w", err)
	}

	// Tab 2: Line Items
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  procurementDashboardPage.ID,
		ParentID:      procurementDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Line Items",
		TableConfigID: procurementLineItemsConfigStored.ID,
		OrderIndex:    2,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating procurement dashboard line items tab: %w", err)
	}

	// Tab 3: Suppliers
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  procurementDashboardPage.ID,
		ParentID:      procurementDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Suppliers",
		TableConfigID: suppliersTableStored.ID,
		OrderIndex:    3,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating procurement dashboard suppliers tab: %w", err)
	}

	// Tab 4: Approvals
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  procurementDashboardPage.ID,
		ParentID:      procurementDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Approvals",
		TableConfigID: procurementApprovalsOpenConfigStored.ID,
		OrderIndex:    4,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating procurement dashboard approvals tab: %w", err)
	}

	// =========================================================================
	// Seed Page Action Buttons
	// =========================================================================

	pageConfigIDs := map[string]uuid.UUID{
		"admin_users_page":            adminUsersPage.ID,
		"admin_roles_page":            adminRolesPage.ID,
		"assets_list_page":            assetsPage.ID,
		"hr_employees_page":           hrEmployeesPage.ID,
		"hr_offices_page":             hrOfficesPage.ID,
		"inventory_items_page":        inventoryItemsPage.ID,
		"inventory_warehouses_page":   inventoryWarehousesPage.ID,
		"inventory_transfers_page":    inventoryTransfersPage.ID,
		"inventory_adjustments_page":  inventoryAdjustmentsPage.ID,
		"suppliers_page":              suppliersPage.ID,
		"procurement_purchase_orders": procurementPurchaseOrdersPage.ID,
		"sales_customers_page":        salesCustomersPage.ID,
		"orders_page":                 ordersPage.ID,
	}

	if err := seedPageActionButtons(ctx, busDomain, pageConfigIDs); err != nil {
		return fmt.Errorf("seeding page action buttons: %w", err)
	}

	// =========================================================================
	// Create Forms
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

	// =========================================================================
	// NEW PAGE CONTENT SYSTEM EXAMPLE - Flexible content blocks
	// =========================================================================
	// This example demonstrates the new flexible page content system with:
	// 1. A form at the top of the page
	// 2. A tabs container below the form
	// 3. Multiple tabs with different table configs
	//
	// This shows how content can be mixed (forms + tables) and nested (tabs)

	// Create a new page config for "User Management Example"
	userManagementPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "user_management_example",
		UserID:    uuid.Nil, // System default
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating user management example page : %w", err)
	}

	// Content Block 1: Form at top (New User Form)
	// Full width on all screen sizes
	formBlock, err := busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID: userManagementPage.ID,
		ContentType:  pagecontentbus.ContentTypeForm,
		Label:        "Create New User",
		FormID:       userForm.ID, // Reference the user form we created earlier
		OrderIndex:   1,
		Layout:       json.RawMessage(`{"colSpan":{"xs":12}}`),
		IsVisible:    true,
		IsDefault:    false,
	})
	if err != nil {
		return fmt.Errorf("creating form content block : %w", err)
	}

	// Content Block 2: Tabs Container
	// This is a container that will hold the tab items
	tabsContainer, err := busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID: userManagementPage.ID,
		ContentType:  pagecontentbus.ContentTypeTabs,
		Label:        "User Lists",
		OrderIndex:   2,
		Layout:       json.RawMessage(`{"colSpan":{"xs":12},"containerType":"tabs"}`),
		IsVisible:    true,
		IsDefault:    false,
	})
	if err != nil {
		return fmt.Errorf("creating tabs container : %w", err)
	}

	// Tab 1: Active Users (using admin users table config)
	// This is a CHILD of the tabs container
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  userManagementPage.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Active Users",
		TableConfigID: adminUsersTableStored.ID, // Reference existing table config
		OrderIndex:    1,
		ParentID:      tabsContainer.ID, // This makes it a child of the tabs container
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true, // This tab is active by default
	})
	if err != nil {
		return fmt.Errorf("creating active users tab : %w", err)
	}

	// Tab 2: Roles (using roles table config)
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  userManagementPage.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Roles",
		TableConfigID: adminRolesTableStored.ID, // Reference existing table config
		OrderIndex:    2,
		ParentID:      tabsContainer.ID, // Child of tabs container
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating roles tab : %w", err)
	}

	// Tab 3: Permissions (using table access config if available)
	adminTableAccessTableStored, err = configStore.QueryByName(ctx, "admin_table_access_page")
	if err == nil {
		// Only create this tab if the table config exists
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  userManagementPage.ID,
			ContentType:   pagecontentbus.ContentTypeTable,
			Label:         "Permissions",
			TableConfigID: adminTableAccessTableStored.ID,
			OrderIndex:    3,
			ParentID:      tabsContainer.ID, // Child of tabs container
			Layout:        json.RawMessage(`{}`),
			IsVisible:     true,
			IsDefault:     false,
		})
		if err != nil {
			return fmt.Errorf("creating permissions tab : %w", err)
		}
	}

	// Log success
	log.Info(ctx, "✅ Created User Management Example page with flexible content blocks",
		"page_config_id", userManagementPage.ID,
		"form_block_id", formBlock.ID,
		"tabs_container_id", tabsContainer.ID)

	// =========================================================================
	// Create Sample Charts Dashboard
	// Demonstrates remaining chart types not distributed to other pages
	// =========================================================================

	// Query remaining chart configs for sample dashboard (those not queried earlier)
	stackedBarRegionStored, _ := configStore.QueryByName(ctx, "seed_stacked_bar_region")
	stackedAreaCumulativeStored, _ := configStore.QueryByName(ctx, "seed_stacked_area_cumulative")
	comboRevenueOrdersStored, _ := configStore.QueryByName(ctx, "seed_combo_revenue_orders")
	waterfallProfitStored, _ := configStore.QueryByName(ctx, "seed_waterfall_profit")

	// Only create dashboard if at least some chart configs exist
	if stackedBarRegionStored != nil {
		sampleChartsDashboardPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
			Name:      "sample_charts_dashboard",
			UserID:    uuid.Nil,
			IsDefault: true,
		})
		if err != nil {
			return fmt.Errorf("creating sample charts dashboard page: %w", err)
		}

		orderIndex := 1

		// Row 1: Stacked charts (2 across)
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  sampleChartsDashboardPage.ID,
			ContentType:   pagecontentbus.ContentTypeChart,
			Label:         "Sales by Region",
			ChartConfigID: stackedBarRegionStored.ID,
			OrderIndex:    orderIndex,
			Layout:        json.RawMessage(`{"colSpan":{"xs":12,"md":6}}`),
			IsVisible:     true,
			IsDefault:     true,
		})
		if err != nil {
			return fmt.Errorf("creating stacked bar chart content: %w", err)
		}
		orderIndex++

		if stackedAreaCumulativeStored != nil {
			_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
				PageConfigID:  sampleChartsDashboardPage.ID,
				ContentType:   pagecontentbus.ContentTypeChart,
				Label:         "Cumulative Revenue",
				ChartConfigID: stackedAreaCumulativeStored.ID,
				OrderIndex:    orderIndex,
				Layout:        json.RawMessage(`{"colSpan":{"xs":12,"md":6}}`),
				IsVisible:     true,
				IsDefault:     false,
			})
			if err != nil {
				return fmt.Errorf("creating stacked area chart content: %w", err)
			}
			orderIndex++
		}

		// Row 2: Combo + Waterfall (2 across)
		if comboRevenueOrdersStored != nil {
			_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
				PageConfigID:  sampleChartsDashboardPage.ID,
				ContentType:   pagecontentbus.ContentTypeChart,
				Label:         "Revenue vs Orders",
				ChartConfigID: comboRevenueOrdersStored.ID,
				OrderIndex:    orderIndex,
				Layout:        json.RawMessage(`{"colSpan":{"xs":12,"md":6}}`),
				IsVisible:     true,
				IsDefault:     false,
			})
			if err != nil {
				return fmt.Errorf("creating combo chart content: %w", err)
			}
			orderIndex++
		}

		if waterfallProfitStored != nil {
			_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
				PageConfigID:  sampleChartsDashboardPage.ID,
				ContentType:   pagecontentbus.ContentTypeChart,
				Label:         "Profit Breakdown",
				ChartConfigID: waterfallProfitStored.ID,
				OrderIndex:    orderIndex,
				Layout:        json.RawMessage(`{"colSpan":{"xs":12,"md":6}}`),
				IsVisible:     true,
				IsDefault:     false,
			})
			if err != nil {
				return fmt.Errorf("creating waterfall chart content: %w", err)
			}
		}

		log.Info(ctx, "✅ Created Sample Charts Dashboard page",
			"page_config_id", sampleChartsDashboardPage.ID)
	}

	return nil
}

// seedPageActionButtons creates "New" button actions for list pages
func seedPageActionButtons(ctx context.Context, busDomain BusDomain, pageConfigIDs map[string]uuid.UUID) error {
	// Get button definitions
	buttonDefs := seedmodels.GetNewButtonActionDefinitions()

	// Create button actions for each page config
	for configName, pageConfigID := range pageConfigIDs {
		buttonDef, exists := buttonDefs[configName]
		if !exists {
			// Skip if no button definition exists for this page config
			continue
		}

		buttonAction := seedmodels.CreateNewButtonAction(pageConfigID, buttonDef)

		_, err := busDomain.PageAction.CreateButton(ctx, buttonAction)
		if err != nil {
			return fmt.Errorf("creating button action for %s: %w", configName, err)
		}
	}

	return nil
}
