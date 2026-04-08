package supervisorkpiapi_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/inventory/supervisorkpiapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventoryadjustmentapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/putawaytaskapp"
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
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
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
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu1 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(busDomain.User, ath, usrs[0].Email.Address),
	}

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admin : %w", err)
	}

	tu2 := apitest.User{
		User:  admins[0],
		Token: apitest.Token(busDomain.User, ath, admins[0].Email.Address),
	}

	// =========================================================================
	// Geography + Warehouses + Locations
	// =========================================================================

	const count = 2

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying regions : %w", err)
	}

	regionIDs := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		regionIDs = append(regionIDs, r.ID)
	}

	ctys, err := citybus.TestSeedCities(ctx, count, regionIDs, busDomain.City)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding cities : %w", err)
	}

	ctyIDs := make([]uuid.UUID, 0, len(ctys))
	for _, c := range ctys {
		ctyIDs = append(ctyIDs, c.ID)
	}

	strs, err := streetbus.TestSeedStreets(ctx, count, ctyIDs, busDomain.Street)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding streets : %w", err)
	}

	strIDs := make([]uuid.UUID, 0, len(strs))
	for _, s := range strs {
		strIDs = append(strIDs, s.ID)
	}

	tzs, err := busDomain.Timezone.Query(ctx, timezonebus.QueryFilter{}, timezonebus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying timezones : %w", err)
	}

	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	contacts, err := contactinfosbus.TestSeedContactInfos(ctx, count, strIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding contact info : %w", err)
	}

	contactIDs := make(uuid.UUIDs, len(contacts))
	for i, c := range contacts {
		contactIDs[i] = c.ID
	}

	warehouses, err := warehousebus.TestSeedWarehouses(ctx, count, tu1.ID, strIDs, busDomain.Warehouse)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding warehouses : %w", err)
	}

	warehouseIDs := make(uuid.UUIDs, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	zones, err := zonebus.TestSeedZone(ctx, 4, warehouseIDs, busDomain.Zones)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding zones : %w", err)
	}

	zoneIDs := make([]uuid.UUID, len(zones))
	for i, z := range zones {
		zoneIDs[i] = z.ZoneID
	}

	inventoryLocations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 5, warehouseIDs, zoneIDs, busDomain.InventoryLocation)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding inventory locations : %w", err)
	}

	locationIDs := make([]uuid.UUID, len(inventoryLocations))
	for i, il := range inventoryLocations {
		locationIDs[i] = il.LocationID
	}

	// =========================================================================
	// Products
	// =========================================================================

	brands, err := brandbus.TestSeedBrands(ctx, 2, contactIDs, busDomain.Brand)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding brands : %w", err)
	}

	brandIDs := make(uuid.UUIDs, len(brands))
	for i, b := range brands {
		brandIDs[i] = b.BrandID
	}

	pc, err := productcategorybus.TestSeedProductCategories(ctx, 2, busDomain.ProductCategory)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding product categories : %w", err)
	}

	pcIDs := make(uuid.UUIDs, len(pc))
	for i, p := range pc {
		pcIDs[i] = p.ProductCategoryID
	}

	products, err := productbus.TestSeedProducts(ctx, 5, brandIDs, pcIDs, busDomain.Product)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding products : %w", err)
	}

	productIDs := make(uuid.UUIDs, len(products))
	for i, p := range products {
		productIDs[i] = p.ProductID
	}

	// =========================================================================
	// KPI-relevant entities
	// =========================================================================

	createdByIDs := []uuid.UUID{tu1.ID, tu2.ID}

	tasks, err := putawaytaskbus.TestSeedPutAwayTasks(ctx, 3, productIDs, locationIDs, createdByIDs, nil, busDomain.PutAwayTask)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding put-away tasks : %w", err)
	}

	adjustments, err := inventoryadjustmentbus.TestSeedInventoryAdjustments(ctx, 4, productIDs, locationIDs, createdByIDs, busDomain.InventoryAdjustment)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding inventory adjustments : %w", err)
	}

	// =========================================================================
	// Permissions
	// =========================================================================

	roles, err := rolebus.TestSeedRoles(ctx, 2, busDomain.Role)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding roles : %w", err)
	}

	roleIDs := make(uuid.UUIDs, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}

	userIDs := uuid.UUIDs{tu1.ID, tu2.ID}

	_, err = userrolebus.TestSeedUserRoles(ctx, userIDs, roleIDs, busDomain.UserRole)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding user roles : %w", err)
	}

	_, err = tableaccessbus.TestSeedTableAccess(ctx, roleIDs, busDomain.TableAccess)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding table access : %w", err)
	}

	ur1, err := busDomain.UserRole.QueryByUserID(ctx, tu1.ID)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying user1 roles : %w", err)
	}

	usrRoleIDs := make(uuid.UUIDs, len(ur1))
	for i, r := range ur1 {
		usrRoleIDs[i] = r.RoleID
	}

	tas, err := busDomain.TableAccess.QueryByRoleIDs(ctx, usrRoleIDs)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying table access : %w", err)
	}

	for _, ta := range tas {
		if ta.TableName == supervisorkpiapi.RouteTable {
			update := tableaccessbus.UpdateTableAccess{
				CanCreate: dbtest.BoolPointer(false),
				CanUpdate: dbtest.BoolPointer(false),
				CanDelete: dbtest.BoolPointer(false),
				CanRead:   dbtest.BoolPointer(true),
			}
			_, err := busDomain.TableAccess.Update(ctx, ta, update)
			if err != nil {
				return apitest.SeedData{}, fmt.Errorf("updating table access : %w", err)
			}
		}
	}

	return apitest.SeedData{
		Admins:               []apitest.User{tu2},
		Users:                []apitest.User{tu1},
		PutAwayTasks:         putawaytaskapp.ToAppPutAwayTasks(tasks),
		InventoryAdjustments: inventoryadjustmentapp.ToAppInventoryAdjustments(adjustments),
	}, nil
}
