package action_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/workflow/actionpermissionsbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

// ActionSeedData holds action-specific test data.
type ActionSeedData struct {
	apitest.SeedData

	// Users with different permission levels
	AdminUser             apitest.User
	UserWithAlertPerm     apitest.User // Has create_alert permission
	UserWithInventoryPerm apitest.User // Has allocate_inventory permission
	UserNoPermissions     apitest.User // Has no action permissions

	// Roles
	AlertRole     rolebus.Role
	InventoryRole rolebus.Role
	BasicRole     rolebus.Role

	// Action Permissions (for reference in tests)
	AlertPermissions     []actionpermissionsbus.ActionPermission
	InventoryPermissions []actionpermissionsbus.ActionPermission

	// Pre-created execution for status tests
	CompletedExecutionID uuid.UUID
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

	// 6. Create a completed execution for status tests
	// For now, just use a random UUID - actual execution would need to be created via ActionService
	completedExecID := uuid.New()

	return ActionSeedData{
		SeedData: apitest.SeedData{
			Admins: []apitest.User{adminUser},
			Users:  []apitest.User{userWithAlertPerm, userWithInventoryPerm, userNoPermissions},
		},
		AdminUser:             adminUser,
		UserWithAlertPerm:     userWithAlertPerm,
		UserWithInventoryPerm: userWithInventoryPerm,
		UserNoPermissions:     userNoPermissions,
		AlertRole:             alertRole,
		InventoryRole:         inventoryRole,
		BasicRole:             basicRole,
		AlertPermissions:      alertPerms,
		InventoryPermissions:  invPerms,
		CompletedExecutionID:  completedExecID,
	}, nil
}
