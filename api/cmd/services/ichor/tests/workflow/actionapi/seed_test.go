package action_test

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
	"github.com/timmaaaz/ichor/business/domain/sales/lineitemfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	orderstypes "github.com/timmaaaz/ichor/business/domain/sales/ordersbus/types"
	"github.com/timmaaaz/ichor/business/domain/workflow/actionpermissionsbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// ActionSeedData holds action-specific test data.
type ActionSeedData struct {
	apitest.SeedData

	// Users with different permission levels
	AdminUser              apitest.User
	UserWithAlertPerm      apitest.User // Has create_alert permission
	UserWithInventoryPerm  apitest.User // Has allocate_inventory permission
	UserNoPermissions      apitest.User // Has no action permissions
	UserWithTransitionPerm apitest.User // Has transition_status permission
	UserWithButtonPerm     apitest.User // Has release_to_picking + claim/execute_transfer_order permission

	// Roles
	AlertRole      rolebus.Role
	InventoryRole  rolebus.Role
	BasicRole      rolebus.Role
	TransitionRole rolebus.Role

	// Action Permissions (for reference in tests)
	AlertPermissions      []actionpermissionsbus.ActionPermission
	InventoryPermissions  []actionpermissionsbus.ActionPermission
	TransitionPermissions []actionpermissionsbus.ActionPermission

	// Pre-created execution for status tests
	CompletedExecutionID uuid.UUID

	// Fulfillment statuses (for transition tests)
	PendingStatusID uuid.UUID // PENDING order fulfillment status
	PickingStatusID uuid.UUID // PICKING order fulfillment status

	// Orders for transition tests
	PendingOrderID           uuid.UUID // Order at PENDING status (valid source for transition)
	NonTransitionableOrderID uuid.UUID // Order at PICKING status (NOT in valid_from)

	// Entities for the configurable-action-button verbs (release/claim/execute)
	ReleaseOrderID           uuid.UUID // PENDING order WITH a line item + stock — releasable to picking
	ApprovedTransferOrderID  uuid.UUID // transfer order at APPROVED status — claimable
	InTransitTransferOrderID uuid.UUID // transfer order at IN_TRANSIT status — executable
}

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (ActionSeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	// 1. Create admin user (has all permissions via seeded admin role)
	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("seeding admins: %w", err)
	}
	adminUser := apitest.User{
		User:  admins[0],
		Token: apitest.Token(db.BusDomain.User, ath, admins[0].Email.Address),
	}

	// 2. Create custom roles for specific permissions
	alertRole, err := busDomain.Role.Create(ctx, rolebus.NewRole{Name: "alert_manager", Description: "Can manage alerts"})
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("creating alert role: %w", err)
	}

	inventoryRole, err := busDomain.Role.Create(ctx, rolebus.NewRole{Name: "inventory_manager", Description: "Can manage inventory"})
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("creating inventory role: %w", err)
	}

	basicRole, err := busDomain.Role.Create(ctx, rolebus.NewRole{Name: "basic_user", Description: "Basic user with no action permissions"})
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("creating basic role: %w", err)
	}

	transitionRole, err := busDomain.Role.Create(ctx, rolebus.NewRole{Name: "transition_manager", Description: "Can execute the generic data actions (transition_status, create_entity, update_field)"})
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("creating transition role: %w", err)
	}

	// 3. Create users
	alertUsers, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("seeding alert users: %w", err)
	}
	userWithAlertPerm := apitest.User{
		User:  alertUsers[0],
		Token: apitest.Token(db.BusDomain.User, ath, alertUsers[0].Email.Address),
	}

	invUsers, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("seeding inventory users: %w", err)
	}
	userWithInventoryPerm := apitest.User{
		User:  invUsers[0],
		Token: apitest.Token(db.BusDomain.User, ath, invUsers[0].Email.Address),
	}

	basicUsers, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("seeding basic users: %w", err)
	}
	userNoPermissions := apitest.User{
		User:  basicUsers[0],
		Token: apitest.Token(db.BusDomain.User, ath, basicUsers[0].Email.Address),
	}

	transitionUsers, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("seeding transition users: %w", err)
	}
	userWithTransitionPerm := apitest.User{
		User:  transitionUsers[0],
		Token: apitest.Token(db.BusDomain.User, ath, transitionUsers[0].Email.Address),
	}

	// 4. Assign roles to users
	_, err = busDomain.UserRole.Create(ctx, userrolebus.NewUserRole{
		UserID: alertUsers[0].ID,
		RoleID: alertRole.ID,
	})
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("assigning alert role: %w", err)
	}

	_, err = busDomain.UserRole.Create(ctx, userrolebus.NewUserRole{
		UserID: invUsers[0].ID,
		RoleID: inventoryRole.ID,
	})
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("assigning inventory role: %w", err)
	}

	_, err = busDomain.UserRole.Create(ctx, userrolebus.NewUserRole{
		UserID: basicUsers[0].ID,
		RoleID: basicRole.ID,
	})
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("assigning basic role: %w", err)
	}

	_, err = busDomain.UserRole.Create(ctx, userrolebus.NewUserRole{
		UserID: transitionUsers[0].ID,
		RoleID: transitionRole.ID,
	})
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("assigning transition role: %w", err)
	}

	// 5. Create action permissions using testutil
	alertPerms, err := actionpermissionsbus.TestSeedActionPermissions(
		ctx, busDomain.ActionPermissions, alertRole.ID,
		[]string{"create_alert", "send_notification"},
	)
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("creating alert permissions: %w", err)
	}

	invPerms, err := actionpermissionsbus.TestSeedActionPermissions(
		ctx, busDomain.ActionPermissions, inventoryRole.ID,
		[]string{"allocate_inventory"},
	)
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("creating inventory permissions: %w", err)
	}

	transitionPerms, err := actionpermissionsbus.TestSeedActionPermissions(
		ctx, busDomain.ActionPermissions, transitionRole.ID,
		[]string{"transition_status", "create_entity", "update_field"},
	)
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("creating transition permissions: %w", err)
	}

	// 6. Create a completed execution for status tests
	// For now, just use a random UUID - actual execution would need to be created via ActionService
	completedExecID := uuid.New()

	// 7. Seed geography chain needed for customers → orders
	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("querying regions: %w", err)
	}
	regionIDs := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		regionIDs = append(regionIDs, r.ID)
	}

	cities, err := citybus.TestSeedCities(ctx, 1, regionIDs, busDomain.City)
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("seeding cities: %w", err)
	}
	cityIDs := make([]uuid.UUID, 0, len(cities))
	for _, c := range cities {
		cityIDs = append(cityIDs, c.ID)
	}

	streets, err := streetbus.TestSeedStreets(ctx, 1, cityIDs, busDomain.Street)
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("seeding streets: %w", err)
	}
	streetIDs := make([]uuid.UUID, 0, len(streets))
	for _, s := range streets {
		streetIDs = append(streetIDs, s.ID)
	}

	tzs, err := busDomain.Timezone.Query(ctx, timezonebus.QueryFilter{}, timezonebus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("querying timezones: %w", err)
	}
	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 1, streetIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("seeding contact infos: %w", err)
	}
	contactInfoIDs := make([]uuid.UUID, 0, len(contactInfos))
	for _, ci := range contactInfos {
		contactInfoIDs = append(contactInfoIDs, ci.ID)
	}

	customers, err := customersbus.TestSeedCustomers(ctx, 1, streetIDs, contactInfoIDs, []uuid.UUID{admins[0].ID}, busDomain.Customers)
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("seeding customers: %w", err)
	}
	customerIDs := make([]uuid.UUID, 0, len(customers))
	for _, c := range customers {
		customerIDs = append(customerIDs, c.ID)
	}

	// 8. Query a currency (currencies are seeded by migrate.Seed)
	currencies, err := busDomain.Currency.Query(ctx, currencybus.QueryFilter{}, currencybus.DefaultOrderBy, page.MustParse("1", "1"))
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("querying currencies: %w", err)
	}
	if len(currencies) == 0 {
		return ActionSeedData{}, fmt.Errorf("no currencies found in seed data")
	}
	currencyID := currencies[0].ID

	// 9. Create order fulfillment statuses: PENDING (valid-from) and PICKING (not-in-valid-from)
	pendingStatus, err := busDomain.OrderFulfillmentStatus.Create(ctx, orderfulfillmentstatusbus.NewOrderFulfillmentStatus{
		Name:        "PENDING",
		Description: "Order is pending",
	})
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("creating PENDING status: %w", err)
	}

	pickingStatus, err := busDomain.OrderFulfillmentStatus.Create(ctx, orderfulfillmentstatusbus.NewOrderFulfillmentStatus{
		Name:        "PICKING",
		Description: "Order is being picked from warehouse",
	})
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("creating PICKING status: %w", err)
	}

	// 10. Create two orders at different statuses for transition tests
	now := time.Now()
	pendingOrder, err := busDomain.Order.Create(ctx, ordersbus.NewOrder{
		Number:              "TST-TRANSITION-001",
		CustomerID:          customerIDs[0],
		DueDate:             now.AddDate(0, 0, 14),
		FulfillmentStatusID: pendingStatus.ID,
		OrderDate:           now,
		Subtotal:            orderstypes.MustParseMoney("100.00"),
		TaxRate:             orderstypes.MustParsePercentage("8.00"),
		TaxAmount:           orderstypes.MustParseMoney("8.00"),
		ShippingCost:        orderstypes.MustParseMoney("10.00"),
		TotalAmount:         orderstypes.MustParseMoney("118.00"),
		CurrencyID:          currencyID,
		Notes:               "Transition test — pending order",
		CreatedBy:           admins[0].ID,
	})
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("creating pending order: %w", err)
	}

	pickingOrder, err := busDomain.Order.Create(ctx, ordersbus.NewOrder{
		Number:              "TST-TRANSITION-002",
		CustomerID:          customerIDs[0],
		DueDate:             now.AddDate(0, 0, 14),
		FulfillmentStatusID: pickingStatus.ID,
		OrderDate:           now,
		Subtotal:            orderstypes.MustParseMoney("200.00"),
		TaxRate:             orderstypes.MustParsePercentage("8.00"),
		TaxAmount:           orderstypes.MustParseMoney("16.00"),
		ShippingCost:        orderstypes.MustParseMoney("10.00"),
		TotalAmount:         orderstypes.MustParseMoney("226.00"),
		CurrencyID:          currencyID,
		Notes:               "Transition test — picking order (not in valid_from)",
		CreatedBy:           admins[0].ID,
	})
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("creating picking order: %w", err)
	}

	// 11. A role + user holding ONLY the configurable-action-button verbs, so the 200
	//     cases prove the P4 grant gates the real endpoint (not an admin bypass).
	buttonRole, err := busDomain.Role.Create(ctx, rolebus.NewRole{Name: "button_action_manager", Description: "Can execute release_to_picking + transfer verbs"})
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("creating button role: %w", err)
	}
	buttonUsers, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("seeding button users: %w", err)
	}
	userWithButtonPerm := apitest.User{
		User:  buttonUsers[0],
		Token: apitest.Token(db.BusDomain.User, ath, buttonUsers[0].Email.Address),
	}
	if _, err = busDomain.UserRole.Create(ctx, userrolebus.NewUserRole{UserID: buttonUsers[0].ID, RoleID: buttonRole.ID}); err != nil {
		return ActionSeedData{}, fmt.Errorf("assigning button role: %w", err)
	}
	if _, err = actionpermissionsbus.TestSeedActionPermissions(ctx, busDomain.ActionPermissions, buttonRole.ID,
		[]string{"release_to_picking", "claim_transfer_order", "execute_transfer_order"}); err != nil {
		return ActionSeedData{}, fmt.Errorf("creating button permissions: %w", err)
	}

	// 12. PROCESSING status — release_to_picking resolves PENDING/PROCESSING/PICKING by name
	//     and errors if any is missing (PENDING/PICKING already created above).
	if _, err = busDomain.OrderFulfillmentStatus.Create(ctx, orderfulfillmentstatusbus.NewOrderFulfillmentStatus{
		Name: "PROCESSING", Description: "Order is processing",
	}); err != nil {
		return ActionSeedData{}, fmt.Errorf("creating PROCESSING status: %w", err)
	}

	// 13. Product + warehouse/zone/locations + stock so release_to_picking can fan a line
	//     item into a pick_task, and the transfer orders have a product + from/to locations.
	brands, err := brandbus.TestSeedBrands(ctx, 1, contactInfoIDs, busDomain.Brand)
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("seeding brands: %w", err)
	}
	categories, err := productcategorybus.TestSeedProductCategories(ctx, 1, busDomain.ProductCategory)
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("seeding product categories: %w", err)
	}
	products, err := productbus.TestSeedProducts(ctx, 1, uuid.UUIDs{brands[0].BrandID}, uuid.UUIDs{categories[0].ProductCategoryID}, busDomain.Product)
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("seeding products: %w", err)
	}
	productID := products[0].ProductID

	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 1, admins[0].ID, streetIDs, busDomain.Warehouse)
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("seeding warehouses: %w", err)
	}
	zones, err := zonebus.TestSeedZone(ctx, 1, []uuid.UUID{warehouses[0].ID}, busDomain.Zones)
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("seeding zones: %w", err)
	}
	locations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 2, []uuid.UUID{warehouses[0].ID}, zones, busDomain.InventoryLocation)
	if err != nil || len(locations) < 2 {
		return ActionSeedData{}, fmt.Errorf("seeding inventory locations: %w", err)
	}
	if _, err = inventoryitembus.TestSeedInventoryItems(ctx, 1, []uuid.UUID{locations[0].LocationID}, []uuid.UUID{productID}, busDomain.InventoryItem); err != nil {
		return ActionSeedData{}, fmt.Errorf("seeding inventory items: %w", err)
	}

	// 14. Dedicated PENDING order with one line item for productID (separate from the
	//     transition-test orders so the release does not mutate shared state).
	releaseOrder, err := busDomain.Order.Create(ctx, ordersbus.NewOrder{
		Number:              "TST-RELEASE-001",
		CustomerID:          customerIDs[0],
		DueDate:             now.AddDate(0, 0, 14),
		FulfillmentStatusID: pendingStatus.ID,
		OrderDate:           now,
		Subtotal:            orderstypes.MustParseMoney("100.00"),
		TaxRate:             orderstypes.MustParsePercentage("8.00"),
		TaxAmount:           orderstypes.MustParseMoney("8.00"),
		ShippingCost:        orderstypes.MustParseMoney("10.00"),
		TotalAmount:         orderstypes.MustParseMoney("118.00"),
		CurrencyID:          currencyID,
		Notes:               "Release test — pending order with a line item",
		CreatedBy:           admins[0].ID,
	})
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("creating release order: %w", err)
	}
	lifs, err := lineitemfulfillmentstatusbus.TestSeedLineItemFulfillmentStatuses(ctx, busDomain.LineItemFulfillmentStatus)
	if err != nil || len(lifs) == 0 {
		return ActionSeedData{}, fmt.Errorf("seeding line item fulfillment statuses: %w", err)
	}
	if _, err = orderlineitemsbus.TestSeedOrderLineItems(ctx, 1, []uuid.UUID{releaseOrder.ID}, []uuid.UUID{productID}, []uuid.UUID{lifs[0].ID}, []uuid.UUID{admins[0].ID}, busDomain.OrderLineItem); err != nil {
		return ActionSeedData{}, fmt.Errorf("seeding release order line item: %w", err)
	}

	// 15. Transfer orders: one APPROVED (claimable), one IN_TRANSIT (executable).
	approvedTO, err := busDomain.TransferOrder.Create(ctx, transferorderbus.NewTransferOrder{
		ProductID: productID, FromLocationID: locations[0].LocationID, ToLocationID: locations[1].LocationID,
		RequestedByID: admins[0].ID, Quantity: 5, Status: transferorderbus.StatusApproved, TransferDate: now,
	})
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("creating approved transfer order: %w", err)
	}
	inTransitTO, err := busDomain.TransferOrder.Create(ctx, transferorderbus.NewTransferOrder{
		ProductID: productID, FromLocationID: locations[0].LocationID, ToLocationID: locations[1].LocationID,
		RequestedByID: admins[0].ID, Quantity: 3, Status: transferorderbus.StatusInTransit, TransferDate: now,
	})
	if err != nil {
		return ActionSeedData{}, fmt.Errorf("creating in-transit transfer order: %w", err)
	}

	return ActionSeedData{
		SeedData: apitest.SeedData{
			Admins: []apitest.User{adminUser},
			Users:  []apitest.User{userWithAlertPerm, userWithInventoryPerm, userNoPermissions, userWithTransitionPerm, userWithButtonPerm},
		},
		AdminUser:                adminUser,
		UserWithAlertPerm:        userWithAlertPerm,
		UserWithInventoryPerm:    userWithInventoryPerm,
		UserNoPermissions:        userNoPermissions,
		UserWithTransitionPerm:   userWithTransitionPerm,
		UserWithButtonPerm:       userWithButtonPerm,
		AlertRole:                alertRole,
		InventoryRole:            inventoryRole,
		BasicRole:                basicRole,
		TransitionRole:           transitionRole,
		AlertPermissions:         alertPerms,
		InventoryPermissions:     invPerms,
		TransitionPermissions:    transitionPerms,
		CompletedExecutionID:     completedExecID,
		PendingStatusID:          pendingStatus.ID,
		PickingStatusID:          pickingStatus.ID,
		PendingOrderID:           pendingOrder.ID,
		NonTransitionableOrderID: pickingOrder.ID,
		ReleaseOrderID:           releaseOrder.ID,
		ApprovedTransferOrderID:  approvedTO.TransferID,
		InTransitTransferOrderID: inTransitTO.TransferID,
	}, nil
}
