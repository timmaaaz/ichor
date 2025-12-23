package formdataapi_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/assets/assetconditionapp"
	"github.com/timmaaaz/ichor/app/domain/assets/assettypeapp"
	"github.com/timmaaaz/ichor/app/domain/assets/validassetapp"
	"github.com/timmaaaz/ichor/app/domain/config/formapp"
	"github.com/timmaaaz/ichor/app/domain/core/contactinfosapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventoryitemapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventorylocationapp"
	"github.com/timmaaaz/ichor/app/domain/products/brandapp"
	"github.com/timmaaaz/ichor/app/domain/products/productapp"
	"github.com/timmaaaz/ichor/app/domain/products/productcategoryapp"
	"github.com/timmaaaz/ichor/app/domain/products/productcostapp"
	"github.com/timmaaaz/ichor/app/domain/sales/customersapp"
	"github.com/timmaaaz/ichor/app/domain/sales/lineitemfulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/domain/sales/orderfulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
	"github.com/timmaaaz/ichor/business/domain/config/formbus"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/products/productcostbus"
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
	"github.com/timmaaaz/ichor/business/domain/sales/lineitemfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/dbtest/seedmodels"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu1 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
	}

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admin : %w", err)
	}

	tu2 := apitest.User{
		User:  admins[0],
		Token: apitest.Token(db.BusDomain.User, ath, admins[0].Email.Address),
	}

	warehouseCount := 5

	// ADDRESSES
	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying regions : %w", err)
	}

	ids := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		ids = append(ids, r.ID)
	}

	ctys, err := citybus.TestSeedCities(ctx, warehouseCount, ids, busDomain.City)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding cities : %w", err)
	}

	ctyIDs := make([]uuid.UUID, 0, len(ctys))
	for _, c := range ctys {
		ctyIDs = append(ctyIDs, c.ID)
	}

	strs, err := streetbus.TestSeedStreets(ctx, warehouseCount, ctyIDs, busDomain.Street)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding streets : %w", err)
	}
	strIDs := make([]uuid.UUID, 0, len(strs))
	for _, s := range strs {
		strIDs = append(strIDs, s.ID)
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

	contacts, err := contactinfosbus.TestSeedContactInfos(ctx, 5, strIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding contact info : %w", err)
	}

	contactIDs := make(uuid.UUIDs, len(contacts))
	for i, c := range contacts {
		contactIDs[i] = c.ID
	}

	brands, err := brandbus.TestSeedBrands(ctx, 10, contactIDs, busDomain.Brand)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding brands : %w", err)
	}

	brandIDs := make(uuid.UUIDs, len(brands))
	for i, b := range brands {
		brandIDs[i] = b.BrandID
	}

	pc, err := productcategorybus.TestSeedProductCategories(ctx, 10, busDomain.ProductCategory)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding product category : %w", err)
	}

	pcIDs := make(uuid.UUIDs, len(pc))
	for i, p := range pc {
		pcIDs[i] = p.ProductCategoryID
	}

	products, err := productbus.TestSeedProducts(ctx, 30, brandIDs, pcIDs, busDomain.Product)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding product : %w", err)
	}

	productIDs := make(uuid.UUIDs, len(products))
	for i, p := range products {
		productIDs[i] = p.ProductID
	}

	// WAREHOUSES
	warehouses, err := warehousebus.TestSeedWarehouses(ctx, warehouseCount, tu1.ID, strIDs, busDomain.Warehouse)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding warehouses : %w", err)
	}

	warehouseIDs := make(uuid.UUIDs, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	zones, err := zonebus.TestSeedZone(ctx, 15, warehouseIDs, busDomain.Zones)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding zones : %w", err)
	}

	zoneIDs := make([]uuid.UUID, len(zones))
	for i, z := range zones {
		zoneIDs[i] = z.ZoneID
	}

	inventoryLocations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 25, warehouseIDs, zoneIDs, busDomain.InventoryLocation)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding inventory locations : %w", err)
	}

	inventoryLocationIDs := make([]uuid.UUID, len(inventoryLocations))
	for i, il := range inventoryLocations {
		inventoryLocationIDs[i] = il.LocationID
	}

	inventoryItems, err := inventoryitembus.TestSeedInventoryItems(ctx, 50, inventoryLocationIDs, productIDs, busDomain.InventoryItem)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding inventory items : %w", err)
	}

	// =========================================================================
	// Permissions stuff
	// =========================================================================
	roles, err := rolebus.TestSeedRoles(ctx, 2, busDomain.Role)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding roles : %w", err)
	}

	roleIDs := make(uuid.UUIDs, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}

	// Include both users for permissions
	userIDs := make(uuid.UUIDs, 2)
	userIDs[0] = tu1.ID
	userIDs[1] = tu2.ID

	_, err = userrolebus.TestSeedUserRoles(ctx, userIDs, roleIDs, busDomain.UserRole)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding user roles : %w", err)
	}

	_, err = tableaccessbus.TestSeedTableAccess(ctx, roleIDs, busDomain.TableAccess)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding table access : %w", err)
	}

	// We need to ensure ONLY tu1's permissions are updated
	ur1, err := busDomain.UserRole.QueryByUserID(ctx, tu1.ID)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying user1 roles : %w", err)
	}

	// Only get table access for tu1's role specifically
	usrRoleIDs := make(uuid.UUIDs, len(ur1))
	for i, r := range ur1 {
		usrRoleIDs[i] = r.RoleID
	}

	// ASSETS
	ats, err := assettypebus.TestSeedAssetTypes(ctx, 10, busDomain.AssetType)
	if err != nil {
		return apitest.SeedData{}, err
	}
	atIDs := make([]uuid.UUID, 0, len(ats))
	for _, at := range ats {
		atIDs = append(atIDs, at.ID)
	}

	as, err := validassetbus.TestSeedValidAssets(ctx, 20, atIDs, tu1.ID, busDomain.ValidAsset)
	if err != nil {
		return apitest.SeedData{}, err
	}

	validAssetIDs := make([]uuid.UUID, len(as))
	for i, asset := range as {
		validAssetIDs[i] = asset.ID
	}

	assetConditions, err := assetconditionbus.TestSeedAssetConditions(ctx, 5, busDomain.AssetCondition)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding asset conditions : %w", err)
	}

	conditionIDs := make([]uuid.UUID, len(assetConditions))
	for i, condition := range assetConditions {
		conditionIDs[i] = condition.ID
	}

	_, err = assetbus.TestSeedAssets(ctx, 15, validAssetIDs, conditionIDs, busDomain.Asset)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding assets : %w", err)
	}

	// =========================================================================

	// Form 1: Single entity - Users only
	userForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "User Creation Form",
	})
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("creating user form : %w", err)
	}

	userEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "users")
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying user entity : %w", err)
	}

	userFormFields := []formfieldbus.FormField{
		{
			ID:           uuid.New(),
			FormID:       userForm.ID,
			EntityID:     userEntity.ID,
			Name:         "username",
			FieldOrder:   1,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "core",
			EntityTable:  "users",
		},
		{
			ID:           uuid.New(),
			FormID:       userForm.ID,
			EntityID:     userEntity.ID,
			Name:         "first_name",
			FieldOrder:   2,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "core",
			EntityTable:  "users",
		},
		{
			ID:           uuid.New(),
			FormID:       userForm.ID,
			EntityID:     userEntity.ID,
			Name:         "last_name",
			FieldOrder:   3,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "core",
			EntityTable:  "users",
		},
		{
			ID:           uuid.New(),
			FormID:       userForm.ID,
			EntityID:     userEntity.ID,
			Name:         "email",
			FieldOrder:   4,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "core",
			EntityTable:  "users",
		},
		{
			ID:           uuid.New(),
			FormID:       userForm.ID,
			EntityID:     userEntity.ID,
			Name:         "password",
			FieldOrder:   5,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "core",
			EntityTable:  "users",
		},
		{
			ID:           uuid.New(),
			FormID:       userForm.ID,
			EntityID:     userEntity.ID,
			Name:         "password_confirm",
			FieldOrder:   6,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "core",
			EntityTable:  "users",
		},
		{
			ID:           uuid.New(),
			FormID:       userForm.ID,
			EntityID:     userEntity.ID,
			Name:         "birthday",
			FieldOrder:   7,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "core",
			EntityTable:  "users",
		},
		{
			ID:           uuid.New(),
			FormID:       userForm.ID,
			EntityID:     userEntity.ID,
			Name:         "roles",
			FieldOrder:   8,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "core",
			EntityTable:  "users",
		},
		{
			ID:           uuid.New(),
			FormID:       userForm.ID,
			EntityID:     userEntity.ID,
			Name:         "system_roles",
			FieldOrder:   9,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "core",
			EntityTable:  "users",
		},
		{
			ID:           uuid.New(),
			FormID:       userForm.ID,
			EntityID:     userEntity.ID,
			Name:         "enabled",
			FieldOrder:   10,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "core",
			EntityTable:  "users",
		},
		{
			ID:           uuid.New(),
			FormID:       userForm.ID,
			EntityID:     userEntity.ID,
			Name:         "requested_by",
			FieldOrder:   11,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "core",
			EntityTable:  "users",
		},
	}

	for _, ff := range userFormFields {
		_, err = busDomain.FormField.Create(ctx, formfieldbus.NewFormField{
			EntitySchema: ff.EntitySchema,
			EntityTable:  ff.EntityTable,
			FormID:       ff.FormID,
			EntityID:     ff.EntityID,
			Name:         ff.Name,
			FieldOrder:   ff.FieldOrder,
		})
		if err != nil {
			return apitest.SeedData{}, fmt.Errorf("creating user form field : %w", err)
		}
	}

	// Form 2: Single entity - Assets only
	assetForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Asset Creation Form",
	})
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("creating asset form : %w", err)
	}

	assetEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "assets")
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying asset entity : %w", err)
	}

	assetFormFields := []formfieldbus.FormField{
		{
			ID:           uuid.New(),
			FormID:       assetForm.ID,
			EntityID:     assetEntity.ID,
			Name:         "valid_asset_id",
			FieldOrder:   1,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "assets",
			EntityTable:  "assets",
		},
		{
			ID:           uuid.New(),
			FormID:       assetForm.ID,
			EntityID:     assetEntity.ID,
			Name:         "serial_number",
			FieldOrder:   2,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "assets",
			EntityTable:  "assets",
		},
		{
			ID:           uuid.New(),
			FormID:       assetForm.ID,
			EntityID:     assetEntity.ID,
			Name:         "asset_condition_id",
			FieldOrder:   3,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "assets",
			EntityTable:  "assets",
		},
	}

	for _, ff := range assetFormFields {
		_, err = busDomain.FormField.Create(ctx, formfieldbus.NewFormField{
			EntitySchema: ff.EntitySchema,
			EntityTable:  ff.EntityTable,
			FormID:       ff.FormID,
			EntityID:     ff.EntityID,
			Name:         ff.Name,
			FieldOrder:   ff.FieldOrder,
		})
		if err != nil {
			return apitest.SeedData{}, fmt.Errorf("creating asset form field : %w", err)
		}
	}

	// Form 3: Multi-entity - User then Asset (with foreign key)
	multiForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "User and Asset Creation Form",
	})
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("creating multi-entity form : %w", err)
	}

	multiFormFields := []formfieldbus.FormField{
		// User fields (order 1-11)
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			Name:         "username",
			FieldOrder:   1,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "core",
			EntityTable:  "users",
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			Name:         "first_name",
			FieldOrder:   2,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "core",
			EntityTable:  "users",
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			Name:         "last_name",
			FieldOrder:   3,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "core",
			EntityTable:  "users",
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			Name:         "email",
			FieldOrder:   4,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "core",
			EntityTable:  "users",
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			Name:         "password",
			FieldOrder:   5,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "core",
			EntityTable:  "users",
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			Name:         "password_confirm",
			FieldOrder:   6,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "core",
			EntityTable:  "users",
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			Name:         "birthday",
			FieldOrder:   7,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "core",
			EntityTable:  "users",
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			Name:         "roles",
			FieldOrder:   8,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "core",
			EntityTable:  "users",
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			Name:         "system_roles",
			FieldOrder:   9,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "core",
			EntityTable:  "users",
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			Name:         "enabled",
			FieldOrder:   10,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "core",
			EntityTable:  "users",
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			Name:         "requested_by",
			FieldOrder:   11,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "core",
			EntityTable:  "users",
		},
		// Asset fields (order 12-14)
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     assetEntity.ID,
			Name:         "asset_condition_id",
			FieldOrder:   12,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "assets",
			EntityTable:  "assets",
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     assetEntity.ID,
			Name:         "valid_asset_id",
			FieldOrder:   13,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "assets",
			EntityTable:  "assets",
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     assetEntity.ID,
			Name:         "serial_number",
			FieldOrder:   14,
			Config:       json.RawMessage(`{}`),
			EntitySchema: "assets",
			EntityTable:  "assets",
		},
	}

	for _, ff := range multiFormFields {
		_, err = busDomain.FormField.Create(ctx, formfieldbus.NewFormField{
			EntitySchema: ff.EntitySchema,
			EntityTable:  ff.EntityTable,
			FormID:       ff.FormID,
			EntityID:     ff.EntityID,
			Name:         ff.Name,
			FieldOrder:   ff.FieldOrder,
		})
		if err != nil {
			return apitest.SeedData{}, fmt.Errorf("creating multi-entity form field : %w", err)
		}
	}

	// =========================================================================
	// Sales Orders Form (for Phase 4 testing)
	// =========================================================================

	// Create product costs for price lookup
	productCosts, err := productcostbus.TestSeedProductCosts(ctx, len(products), productIDs, busDomain.ProductCost)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding product costs : %w", err)
	}

	// Create customers
	customerIDs := []uuid.UUID{tu1.ID}
	customers, err := customersbus.TestSeedCustomers(ctx, 2, strIDs, contactIDs, customerIDs, busDomain.Customers)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding customers : %w", err)
	}

	// Create order fulfillment statuses
	orderStatuses, err := orderfulfillmentstatusbus.TestSeedOrderFulfillmentStatuses(ctx, busDomain.OrderFulfillmentStatus)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding order fulfillment statuses : %w", err)
	}

	// Create line item fulfillment statuses
	lineItemStatuses, err := lineitemfulfillmentstatusbus.TestSeedLineItemFulfillmentStatuses(ctx, busDomain.LineItemFulfillmentStatus)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding line item fulfillment statuses : %w", err)
	}

	// Create sales order form with line items
	salesOrderForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Sales Order Creation Form",
	})
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("creating sales order form : %w", err)
	}

	orderEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "orders")
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying order entity : %w", err)
	}

	lineItemEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "order_line_items")
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying line item entity : %w", err)
	}

	salesOrderFormFields := seedmodels.GetFullSalesOrderFormFields(
		salesOrderForm.ID,
		orderEntity.ID,
		lineItemEntity.ID,
	)

	for _, ff := range salesOrderFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return apitest.SeedData{}, fmt.Errorf("creating sales order form field : %w", err)
		}
	}

	forms := []formapp.Form{
		formapp.ToAppForm(userForm),
		formapp.ToAppForm(assetForm),
		formapp.ToAppForm(multiForm),
		formapp.ToAppForm(salesOrderForm),
	}

	return apitest.SeedData{
		Admins:                     []apitest.User{tu2},
		Users:                      []apitest.User{tu1},
		ProductCategories:          productcategoryapp.ToAppProductCategories(pc),
		ContactInfos:               contactinfosapp.ToAppContactInfos(contacts),
		Brands:                     brandapp.ToAppBrands(brands),
		Products:                   productapp.ToAppProducts(products),
		ProductCosts:               productcostapp.ToAppProductCosts(productCosts),
		InventoryLocations:         inventorylocationapp.ToAppInventoryLocations(inventoryLocations),
		InventoryItems:             inventoryitemapp.ToAppInventoryItems(inventoryItems),
		Forms:                      forms,
		AssetTypes:                 assettypeapp.ToAppAssetTypes(ats),
		AssetConditions:            assetconditionapp.ToAppAssetConditions(assetConditions),
		ValidAssets:                validassetapp.ToAppValidAssets(as),
		Customers:                  customersapp.ToAppCustomers(customers),
		OrderFulfillmentStatuses:   orderfulfillmentstatusapp.ToAppOrderFulfillmentStatuses(orderStatuses),
		LineItemFulfillmentStatuses: lineitemfulfillmentstatusapp.ToAppLineItemFulfillmentStatuses(lineItemStatuses),
	}, nil
}
