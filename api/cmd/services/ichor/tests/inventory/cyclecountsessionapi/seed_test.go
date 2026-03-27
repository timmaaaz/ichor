package cyclecountsessionapi_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/inventory/cyclecountsessionapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountsessionapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventorylocationapp"
	"github.com/timmaaaz/ichor/app/domain/products/productapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	// =========================================================================
	// Users
	// =========================================================================

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users: %w", err)
	}
	tu1 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(busDomain.User, ath, usrs[0].Email.Address),
	}

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admins: %w", err)
	}
	tu2 := apitest.User{
		User:  admins[0],
		Token: apitest.Token(busDomain.User, ath, admins[0].Email.Address),
	}

	// =========================================================================
	// Cycle Count Sessions (only FK is created_by → users)
	// =========================================================================

	createdByIDs := []uuid.UUID{tu2.ID}
	sessions, err := cyclecountsessionbus.TestSeedCycleCountSessions(ctx, 4, createdByIDs, busDomain.CycleCountSession)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding cycle count sessions: %w", err)
	}

	// =========================================================================
	// Permissions: tu1 (User) = read-only on sessions table
	// =========================================================================

	roles, err := rolebus.TestSeedRoles(ctx, 2, busDomain.Role)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding roles: %w", err)
	}

	roleIDs := make([]uuid.UUID, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}

	userIDs := []uuid.UUID{tu1.ID, tu2.ID}

	_, err = userrolebus.TestSeedUserRoles(ctx, userIDs, roleIDs, busDomain.UserRole)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding user roles: %w", err)
	}

	_, err = tableaccessbus.TestSeedTableAccess(ctx, roleIDs, busDomain.TableAccess)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding table access: %w", err)
	}

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
		if ta.TableName == cyclecountsessionapi.RouteTable {
			update := tableaccessbus.UpdateTableAccess{
				CanCreate: dbtest.BoolPointer(false),
				CanUpdate: dbtest.BoolPointer(false),
				CanDelete: dbtest.BoolPointer(false),
				CanRead:   dbtest.BoolPointer(false),
			}
			if _, err := busDomain.TableAccess.Update(ctx, ta, update); err != nil {
				return apitest.SeedData{}, fmt.Errorf("updating table access: %w", err)
			}
		}
	}

	// =========================================================================
	// Return
	// =========================================================================

	return apitest.SeedData{
		Admins:             []apitest.User{tu2},
		Users:              []apitest.User{tu1},
		CycleCountSessions: cyclecountsessionapp.ToAppCycleCountSessions(sessions),
	}, nil
}

// insertCompleteFlowSeedData creates seed data for the complete flow test.
// This includes products and inventory locations needed for cycle count items.
func insertCompleteFlowSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	// =========================================================================
	// Users
	// =========================================================================

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users: %w", err)
	}
	tu1 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(busDomain.User, ath, usrs[0].Email.Address),
	}

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admins: %w", err)
	}
	tu2 := apitest.User{
		User:  admins[0],
		Token: apitest.Token(busDomain.User, ath, admins[0].Email.Address),
	}

	// =========================================================================
	// Geography → Warehouse → Zones → Locations
	// =========================================================================

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying regions: %w", err)
	}
	regionIDs := make([]uuid.UUID, len(regions))
	for i, r := range regions {
		regionIDs[i] = r.ID
	}

	ctys, err := citybus.TestSeedCities(ctx, 1, regionIDs, busDomain.City)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding cities: %w", err)
	}
	ctyIDs := make([]uuid.UUID, len(ctys))
	for i, c := range ctys {
		ctyIDs[i] = c.ID
	}

	strs, err := streetbus.TestSeedStreets(ctx, 1, ctyIDs, busDomain.Street)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding streets: %w", err)
	}
	strIDs := make([]uuid.UUID, len(strs))
	for i, s := range strs {
		strIDs[i] = s.ID
	}

	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 1, tu2.ID, strIDs, busDomain.Warehouse)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding warehouses: %w", err)
	}
	warehouseIDs := make([]uuid.UUID, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	zones, err := zonebus.TestSeedZone(ctx, 1, warehouseIDs, busDomain.Zones)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding zones: %w", err)
	}
	zoneIDs := make([]uuid.UUID, len(zones))
	for i, z := range zones {
		zoneIDs[i] = z.ZoneID
	}

	inventoryLocations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 2, warehouseIDs, zoneIDs, busDomain.InventoryLocation)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding inventory locations: %w", err)
	}

	// =========================================================================
	// Products
	// =========================================================================

	tzs, err := busDomain.Timezone.Query(ctx, timezonebus.QueryFilter{}, timezonebus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying timezones: %w", err)
	}
	tzIDs := make([]uuid.UUID, len(tzs))
	for i, tz := range tzs {
		tzIDs[i] = tz.ID
	}

	contacts, err := contactinfosbus.TestSeedContactInfos(ctx, 1, strIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding contacts: %w", err)
	}
	contactIDs := make([]uuid.UUID, len(contacts))
	for i, c := range contacts {
		contactIDs[i] = c.ID
	}

	brands, err := brandbus.TestSeedBrands(ctx, 1, contactIDs, busDomain.Brand)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding brands: %w", err)
	}
	brandIDs := make([]uuid.UUID, len(brands))
	for i, b := range brands {
		brandIDs[i] = b.BrandID
	}

	pcs, err := productcategorybus.TestSeedProductCategories(ctx, 1, busDomain.ProductCategory)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding product categories: %w", err)
	}
	pcIDs := make([]uuid.UUID, len(pcs))
	for i, pc := range pcs {
		pcIDs[i] = pc.ProductCategoryID
	}

	products, err := productbus.TestSeedProducts(ctx, 2, brandIDs, pcIDs, busDomain.Product)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding products: %w", err)
	}

	// =========================================================================
	// Cycle Count Session (1 session for the complete flow)
	// =========================================================================

	sessions, err := cyclecountsessionbus.TestSeedCycleCountSessions(ctx, 1, []uuid.UUID{tu2.ID}, busDomain.CycleCountSession)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding sessions: %w", err)
	}

	// =========================================================================
	// Permissions
	// =========================================================================

	roles, err := rolebus.TestSeedRoles(ctx, 2, busDomain.Role)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding roles: %w", err)
	}

	roleIDs := make([]uuid.UUID, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}

	_, err = userrolebus.TestSeedUserRoles(ctx, []uuid.UUID{tu1.ID, tu2.ID}, roleIDs, busDomain.UserRole)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding user roles: %w", err)
	}

	_, err = tableaccessbus.TestSeedTableAccess(ctx, roleIDs, busDomain.TableAccess)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding table access: %w", err)
	}

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
		if ta.TableName == cyclecountsessionapi.RouteTable {
			update := tableaccessbus.UpdateTableAccess{
				CanCreate: dbtest.BoolPointer(false),
				CanUpdate: dbtest.BoolPointer(false),
				CanDelete: dbtest.BoolPointer(false),
				CanRead:   dbtest.BoolPointer(true),
			}
			if _, err := busDomain.TableAccess.Update(ctx, ta, update); err != nil {
				return apitest.SeedData{}, fmt.Errorf("updating table access: %w", err)
			}
		}
	}

	// =========================================================================
	// Return
	// =========================================================================

	return apitest.SeedData{
		Admins:             []apitest.User{tu2},
		Users:              []apitest.User{tu1},
		CycleCountSessions: cyclecountsessionapp.ToAppCycleCountSessions(sessions),
		Products:           productapp.ToAppProducts(products),
		InventoryLocations: inventorylocationapp.ToAppInventoryLocations(inventoryLocations),
	}, nil
}
