package paperworkapi_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/api/domain/http/paperwork/paperworkapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// PaperworkSeed extends apitest.SeedData with paperwork-specific fixture IDs.
// The apitest harness's SeedData is a fixed shape; paperwork tests carry
// the seeded entity IDs in this side-channel struct.
type PaperworkSeed struct {
	apitest.SeedData
	OrderID    uuid.UUID
	OrderNum   string
	PurchaseID uuid.UUID
	POOrderNum string
	TransferID uuid.UUID
	TransferNo string
}

// insertSeedData stages everything the paperwork integration tests need:
//   - one admin (full permissions, used for happy-path 200)
//   - one regular user with the three paperwork-relevant table_access rows
//     downgraded to 0 (used for 403 cases)
//   - one sales order, one purchase order, one transfer order (used as
//     query targets for happy-path 200)
//
// Mirrors labelapi/seed_test.go role-downgrade pattern.
func insertSeedData(db *dbtest.Database, ath *auth.Auth) (PaperworkSeed, error) {
	ctx := context.Background()
	bd := db.BusDomain

	// Users
	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, bd.User)
	if err != nil {
		return PaperworkSeed{}, fmt.Errorf("seeding users: %w", err)
	}
	tu1 := apitest.User{User: usrs[0], Token: apitest.Token(bd.User, ath, usrs[0].Email.Address)}

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, bd.User)
	if err != nil {
		return PaperworkSeed{}, fmt.Errorf("seeding admins: %w", err)
	}
	tu2 := apitest.User{User: admins[0], Token: apitest.Token(bd.User, ath, admins[0].Email.Address)}

	// Domain entities — sales order, purchase order, transfer order are
	// seeded together through a single helper because their FK chains share
	// fixed-code seeders (currencies, streets, contact info, etc.) that can
	// only be invoked once per test database. Mirrors paperworkbus_test.go's
	// seedAll function — see business/domain/paperwork/paperworkbus.
	pwk, err := seedPaperworkEntities(ctx, bd, admins[0].ID)
	if err != nil {
		return PaperworkSeed{}, fmt.Errorf("seeding paperwork entities: %w", err)
	}

	// Permissions — downgrade tu1's table_access for all three paperwork
	// RouteTables so 403 cases fire reliably.
	roles, err := rolebus.TestSeedRoles(ctx, 2, bd.Role)
	if err != nil {
		return PaperworkSeed{}, fmt.Errorf("seeding roles: %w", err)
	}
	roleIDs := make([]uuid.UUID, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}

	if _, err := userrolebus.TestSeedUserRoles(ctx, []uuid.UUID{tu1.ID, tu2.ID}, roleIDs, bd.UserRole); err != nil {
		return PaperworkSeed{}, fmt.Errorf("seeding user roles: %w", err)
	}
	if _, err := tableaccessbus.TestSeedTableAccess(ctx, roleIDs, bd.TableAccess); err != nil {
		return PaperworkSeed{}, fmt.Errorf("seeding table access: %w", err)
	}

	ur1, err := bd.UserRole.QueryByUserID(ctx, tu1.ID)
	if err != nil {
		return PaperworkSeed{}, fmt.Errorf("querying user1 roles: %w", err)
	}
	usrRoleIDs := make(uuid.UUIDs, len(ur1))
	for i, r := range ur1 {
		usrRoleIDs[i] = r.RoleID
	}
	tas, err := bd.TableAccess.QueryByRoleIDs(ctx, usrRoleIDs)
	if err != nil {
		return PaperworkSeed{}, fmt.Errorf("querying table access: %w", err)
	}

	downgrade := func(table string) error {
		for _, ta := range tas {
			if ta.TableName != table {
				continue
			}
			update := tableaccessbus.UpdateTableAccess{
				CanCreate: dbtest.BoolPointer(false),
				CanUpdate: dbtest.BoolPointer(false),
				CanDelete: dbtest.BoolPointer(false),
				CanRead:   dbtest.BoolPointer(false),
			}
			if _, err := bd.TableAccess.Update(ctx, ta, update); err != nil {
				return fmt.Errorf("downgrading %s: %w", table, err)
			}
		}
		return nil
	}
	for _, table := range []string{
		paperworkapi.RouteTablePickSheet,
		paperworkapi.RouteTableReceiveCover,
		paperworkapi.RouteTableTransferSheet,
	} {
		if err := downgrade(table); err != nil {
			return PaperworkSeed{}, err
		}
	}

	return PaperworkSeed{
		SeedData: apitest.SeedData{
			Admins: []apitest.User{tu2},
			Users:  []apitest.User{tu1},
		},
		OrderID:    pwk.orderID,
		OrderNum:   pwk.orderNum,
		PurchaseID: pwk.poID,
		POOrderNum: pwk.poNum,
		TransferID: pwk.transferID,
		TransferNo: pwk.transferNum,
	}, nil
}

// paperworkEntities bundles the three seeded entity IDs + identifying
// numbers the apitest happy-path 200 cases drive against the handler.
type paperworkEntities struct {
	orderID     uuid.UUID
	orderNum    string
	poID        uuid.UUID
	poNum       string
	transferID  uuid.UUID
	transferNum string
}

// seedPaperworkEntities provisions the union of FK chains the three Build*
// methods need: addresses (region/city/street), contact info, currency,
// customers, suppliers, warehouses, zones, locations, brands, categories,
// products, statuses — plus exactly one Order, PurchaseOrder, and
// TransferOrder. Mirrors paperworkbus_test.go's seedAll one-for-one.
//
// All seeding happens in a single function so currency, status, and other
// fixed-code seeders are invoked exactly once — calling them twice in the
// same test database produces unique-key violations.
func seedPaperworkEntities(ctx context.Context, bd dbtest.BusDomain, adminID uuid.UUID) (paperworkEntities, error) {
	// USERS — non-admin users used as TransferOrder requesters/approvers.
	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 6, userbus.Roles.User, bd.User)
	if err != nil {
		return paperworkEntities{}, fmt.Errorf("seed transfer-order users: %w", err)
	}
	userIDs := make([]uuid.UUID, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	// ADDRESS / CONTACT CHAIN.
	regions, err := bd.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return paperworkEntities{}, fmt.Errorf("query regions: %w", err)
	}
	regionIDs := make(uuid.UUIDs, len(regions))
	for i, r := range regions {
		regionIDs[i] = r.ID
	}

	cities, err := citybus.TestSeedCities(ctx, 5, regionIDs, bd.City)
	if err != nil {
		return paperworkEntities{}, fmt.Errorf("seed cities: %w", err)
	}
	cityIDs := make(uuid.UUIDs, len(cities))
	for i, c := range cities {
		cityIDs[i] = c.ID
	}

	streets, err := streetbus.TestSeedStreets(ctx, 5, cityIDs, bd.Street)
	if err != nil {
		return paperworkEntities{}, fmt.Errorf("seed streets: %w", err)
	}
	streetIDs := make(uuid.UUIDs, len(streets))
	for i, s := range streets {
		streetIDs[i] = s.ID
	}

	tzs, err := bd.Timezone.QueryAll(ctx)
	if err != nil {
		return paperworkEntities{}, fmt.Errorf("query timezones: %w", err)
	}
	tzIDs := make(uuid.UUIDs, len(tzs))
	for i, tz := range tzs {
		tzIDs[i] = tz.ID
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 5, streetIDs, tzIDs, bd.ContactInfos)
	if err != nil {
		return paperworkEntities{}, fmt.Errorf("seed contact info: %w", err)
	}
	contactInfoIDs := make(uuid.UUIDs, len(contactInfos))
	for i, ci := range contactInfos {
		contactInfoIDs[i] = ci.ID
	}

	// SHARED CURRENCY (single call — fixed codes TS0..TS4).
	currencies, err := currencybus.TestSeedCurrencies(ctx, 5, bd.Currency)
	if err != nil {
		return paperworkEntities{}, fmt.Errorf("seed currencies: %w", err)
	}
	currencyIDs := make(uuid.UUIDs, len(currencies))
	for i, c := range currencies {
		currencyIDs[i] = c.ID
	}

	// ORDER (sales).
	customers, err := customersbus.TestSeedCustomers(ctx, 1, streetIDs, contactInfoIDs, uuid.UUIDs{adminID}, bd.Customers)
	if err != nil {
		return paperworkEntities{}, fmt.Errorf("seed customers: %w", err)
	}
	customerIDs := uuid.UUIDs{customers[0].ID}

	ofls, err := orderfulfillmentstatusbus.TestSeedOrderFulfillmentStatuses(ctx, bd.OrderFulfillmentStatus)
	if err != nil {
		return paperworkEntities{}, fmt.Errorf("seed order fulfillment statuses: %w", err)
	}
	oflIDs := make(uuid.UUIDs, len(ofls))
	for i, o := range ofls {
		oflIDs[i] = o.ID
	}

	orders, err := ordersbus.TestSeedOrders(ctx, 1, uuid.UUIDs{adminID}, customerIDs, oflIDs, currencyIDs, bd.Order)
	if err != nil {
		return paperworkEntities{}, fmt.Errorf("seed orders: %w", err)
	}

	// PURCHASE ORDER.
	suppliers, err := supplierbus.TestSeedSuppliers(ctx, 1, contactInfoIDs, bd.Supplier)
	if err != nil {
		return paperworkEntities{}, fmt.Errorf("seed suppliers: %w", err)
	}
	supplierIDs := make(uuid.UUIDs, len(suppliers))
	for i, s := range suppliers {
		supplierIDs[i] = s.SupplierID
	}

	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 2, adminID, streetIDs, bd.Warehouse)
	if err != nil {
		return paperworkEntities{}, fmt.Errorf("seed warehouses: %w", err)
	}
	warehouseIDs := make(uuid.UUIDs, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	poStatuses, err := purchaseorderstatusbus.TestSeedPurchaseOrderStatuses(ctx, 3, bd.PurchaseOrderStatus)
	if err != nil {
		return paperworkEntities{}, fmt.Errorf("seed PO statuses: %w", err)
	}
	poStatusIDs := make(uuid.UUIDs, len(poStatuses))
	for i, s := range poStatuses {
		poStatusIDs[i] = s.ID
	}

	pos, err := purchaseorderbus.TestSeedPurchaseOrders(ctx, 1, supplierIDs, poStatusIDs, warehouseIDs, streetIDs, uuid.UUIDs{adminID}, currencyIDs, bd.PurchaseOrder)
	if err != nil {
		return paperworkEntities{}, fmt.Errorf("seed purchase orders: %w", err)
	}

	// TRANSFER ORDER.
	brands, err := brandbus.TestSeedBrands(ctx, 1, contactInfoIDs, bd.Brand)
	if err != nil {
		return paperworkEntities{}, fmt.Errorf("seed brands: %w", err)
	}
	brandIDs := make(uuid.UUIDs, len(brands))
	for i, b := range brands {
		brandIDs[i] = b.BrandID
	}

	categories, err := productcategorybus.TestSeedProductCategories(ctx, 1, bd.ProductCategory)
	if err != nil {
		return paperworkEntities{}, fmt.Errorf("seed product categories: %w", err)
	}
	categoryIDs := make(uuid.UUIDs, len(categories))
	for i, c := range categories {
		categoryIDs[i] = c.ProductCategoryID
	}

	products, err := productbus.TestSeedProducts(ctx, 2, brandIDs, categoryIDs, bd.Product)
	if err != nil {
		return paperworkEntities{}, fmt.Errorf("seed products: %w", err)
	}
	productIDs := make([]uuid.UUID, len(products))
	for i, p := range products {
		productIDs[i] = p.ProductID
	}

	zones, err := zonebus.TestSeedZone(ctx, 4, warehouseIDs, bd.Zones)
	if err != nil {
		return paperworkEntities{}, fmt.Errorf("seed zones: %w", err)
	}

	locations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 4, warehouseIDs, zones, bd.InventoryLocation)
	if err != nil {
		return paperworkEntities{}, fmt.Errorf("seed inventory locations: %w", err)
	}
	if len(locations) < 2 {
		return paperworkEntities{}, fmt.Errorf("seed inventory locations: want >=2, got %d", len(locations))
	}
	fromIDs := []uuid.UUID{locations[0].LocationID}
	toIDs := []uuid.UUID{locations[1].LocationID}

	tos, err := transferorderbus.TestSeedTransferOrders(ctx, 1, productIDs, fromIDs, toIDs, userIDs[:3], userIDs[3:], nil, bd.TransferOrder)
	if err != nil {
		return paperworkEntities{}, fmt.Errorf("seed transfer orders: %w", err)
	}
	if tos[0].TransferNumber == nil {
		return paperworkEntities{}, fmt.Errorf("seed transfer orders: TransferNumber is nil")
	}

	return paperworkEntities{
		orderID:     orders[0].ID,
		orderNum:    orders[0].Number,
		poID:        pos[0].ID,
		poNum:       pos[0].OrderNumber,
		transferID:  tos[0].TransferID,
		transferNum: *tos[0].TransferNumber,
	}, nil
}
