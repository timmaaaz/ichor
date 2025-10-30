package pageaction_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/config/pageactionapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/config/pageactionapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	// =========================================================================
	// Seed Users
	// =========================================================================

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}
	tu1 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
	}

	usrs, err = userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}
	tu2 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
	}

	// =========================================================================
	// Seed Roles and Permissions
	// =========================================================================

	roles, err := rolebus.TestSeedRoles(ctx, 12, busDomain.Role)
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

	tas, err := busDomain.TableAccess.QueryByRoleIDs(ctx, usrRoleIDs)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying table access : %w", err)
	}

	// Update only tu1's permissions to deny access to page actions
	for _, ta := range tas {
		if ta.TableName == pageactionapi.RouteTablePageActions ||
			ta.TableName == pageactionapi.RouteTablePageActionButtons ||
			ta.TableName == pageactionapi.RouteTablePageActionDropdowns {
			update := tableaccessbus.UpdateTableAccess{
				CanCreate: dbtest.BoolPointer(false),
				CanUpdate: dbtest.BoolPointer(false),
				CanDelete: dbtest.BoolPointer(false),
				CanRead:   dbtest.BoolPointer(false),
			}
			_, err := busDomain.TableAccess.Update(ctx, ta, update)
			if err != nil {
				return apitest.SeedData{}, fmt.Errorf("updating table access : %w", err)
			}
		}
	}

	// =========================================================================
	// Seed Page Configs using tablebuilder
	// =========================================================================

	configStore := tablebuilder.NewConfigStore(db.Log, db.DB)
	pageConfigIDs := make([]uuid.UUID, 3)

	for i := 0; i < 3; i++ {
		pageConfig, err := configStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
			Name:      fmt.Sprintf("Test Page Config %d", i),
			UserID:    uuid.Nil,
			IsDefault: true,
		})
		if err != nil {
			return apitest.SeedData{}, fmt.Errorf("creating page config : %w", err)
		}
		pageConfigIDs[i] = pageConfig.ID
	}

	// =========================================================================
	// Seed Page Actions
	// =========================================================================

	actions, err := pageactionbus.TestSeedPageActions(ctx, 12, pageConfigIDs, busDomain.PageAction)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding page actions : %w", err)
	}

	appActions := pageactionapp.ToAppPageActions(actions)

	// Store first page config ID for tests that need it
	sd := apitest.SeedData{
		Users:       []apitest.User{tu1},
		Admins:      []apitest.User{tu2},
		PageActions: appActions,
	}

	// Store page config IDs in PageConfigs field for test access
	for _, id := range pageConfigIDs {
		sd.PageConfigs = append(sd.PageConfigs, tablebuilder.PageConfig{ID: id})
	}

	return sd, nil
}
