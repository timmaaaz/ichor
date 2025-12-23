package purchaseorderapi_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/procurement/purchaseorderapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/contactinfosapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventorylocationapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/warehouseapp"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderapp"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderstatusapp"
	"github.com/timmaaaz/ichor/app/domain/procurement/supplierapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	// Seed users
	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users: %w", err)
	}

	tu1 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
	}

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admin: %w", err)
	}

	tu2 := apitest.User{
		User:  admins[0],
		Token: apitest.Token(db.BusDomain.User, ath, admins[0].Email.Address),
	}

	// Seed addresses
	count := 5
	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying regions: %w", err)
	}

	regionIDs := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		regionIDs = append(regionIDs, r.ID)
	}

	cities, err := citybus.TestSeedCities(ctx, count, regionIDs, busDomain.City)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding cities: %w", err)
	}

	cityIDs := make([]uuid.UUID, 0, len(cities))
	for _, c := range cities {
		cityIDs = append(cityIDs, c.ID)
	}

	streets, err := streetbus.TestSeedStreets(ctx, count, cityIDs, busDomain.Street)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding streets: %w", err)
	}

	streetIDs := make([]uuid.UUID, 0, len(streets))
	for _, s := range streets {
		streetIDs = append(streetIDs, s.ID)
	}

	// Query timezones from seed data
	tzs, err := busDomain.Timezone.Query(ctx, timezonebus.QueryFilter{}, timezonebus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying timezones : %w", err)
	}
	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	// Seed contact infos
	contacts, err := contactinfosbus.TestSeedContactInfos(ctx, count, streetIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding contact infos: %w", err)
	}

	contactIDs := make(uuid.UUIDs, len(contacts))
	for i, c := range contacts {
		contactIDs[i] = c.ID
	}

	// Seed suppliers
	suppliers, err := supplierbus.TestSeedSuppliers(ctx, 10, contactIDs, busDomain.Supplier)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding suppliers: %w", err)
	}

	supplierIDs := make([]uuid.UUID, len(suppliers))
	for i, s := range suppliers {
		supplierIDs[i] = s.SupplierID
	}

	// Seed purchase order statuses
	statuses, err := purchaseorderstatusbus.TestSeedPurchaseOrderStatuses(ctx, 5, busDomain.PurchaseOrderStatus)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding purchase order statuses: %w", err)
	}

	statusIDs := make([]uuid.UUID, len(statuses))
	for i, s := range statuses {
		statusIDs[i] = s.ID
	}

	// Seed warehouses
	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 3, tu1.ID, streetIDs, busDomain.Warehouse)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding warehouses: %w", err)
	}

	warehouseIDs := make([]uuid.UUID, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	// Seed zones
	zones, err := zonebus.TestSeedZone(ctx, 3, warehouseIDs, busDomain.Zones)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding zones: %w", err)
	}

	zoneIDs := make([]uuid.UUID, len(zones))
	for i, z := range zones {
		zoneIDs[i] = z.ZoneID
	}

	// Seed inventory locations
	locations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 5, warehouseIDs, zoneIDs, busDomain.InventoryLocation)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding inventory locations: %w", err)
	}

	// Seed purchase orders
	userIDs := []uuid.UUID{tu1.ID, tu2.ID}
	purchaseOrders, err := purchaseorderbus.TestSeedPurchaseOrders(ctx, 10, supplierIDs, statusIDs, warehouseIDs, streetIDs, userIDs, busDomain.PurchaseOrder)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding purchase orders: %w", err)
	}

	// Seed permissions
	roles, err := rolebus.TestSeedRoles(ctx, 2, busDomain.Role)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding roles: %w", err)
	}

	roleIDs := make(uuid.UUIDs, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}

	userIDsForRoles := make(uuid.UUIDs, 2)
	userIDsForRoles[0] = tu1.ID
	userIDsForRoles[1] = tu2.ID

	_, err = userrolebus.TestSeedUserRoles(ctx, userIDsForRoles, roleIDs, busDomain.UserRole)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding user roles: %w", err)
	}

	_, err = tableaccessbus.TestSeedTableAccess(ctx, roleIDs, busDomain.TableAccess)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding table access: %w", err)
	}

	// Update permissions for tu1 (read-only)
	ur1, err := busDomain.UserRole.QueryByUserID(ctx, tu1.ID)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying user1 roles: %w", err)
	}

	usrRoleIDs := make(uuid.UUIDs, len(ur1))
	for i, r := range ur1 {
		usrRoleIDs[i] = r.RoleID
	}

	tas, err := busDomain.TableAccess.QueryByRoleIDs(ctx, usrRoleIDs)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying table access: %w", err)
	}

	for _, ta := range tas {
		if ta.TableName == purchaseorderapi.RouteTable {
			update := tableaccessbus.UpdateTableAccess{
				CanCreate: dbtest.BoolPointer(false),
				CanUpdate: dbtest.BoolPointer(false),
				CanDelete: dbtest.BoolPointer(false),
				CanRead:   dbtest.BoolPointer(true),
			}
			_, err := busDomain.TableAccess.Update(ctx, ta, update)
			if err != nil {
				return apitest.SeedData{}, fmt.Errorf("updating table access: %w", err)
			}
		}
	}

	return apitest.SeedData{
		Admins:                []apitest.User{tu2},
		Users:                 []apitest.User{tu1},
		ContactInfos:          contactinfosapp.ToAppContactInfos(contacts),
		Suppliers:             supplierapp.ToAppSuppliers(suppliers),
		PurchaseOrderStatuses: purchaseorderstatusapp.ToAppPurchaseOrderStatuses(statuses),
		Warehouses:            warehouseapp.ToAppWarehouses(warehouses),
		InventoryLocations:    inventorylocationapp.ToAppInventoryLocations(locations),
		PurchaseOrders:        purchaseorderapp.ToAppPurchaseOrders(purchaseOrders),
	}, nil
}
