package page_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/core/pageapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/pageapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/pagebus"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/rolepagebus"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
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
	usrs, err = userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}
	tu2 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
	}

	// =========================================================================
	// Permissions stuff
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

	// Update only tu1's role permissions
	for _, ta := range tas {
		// Only update for the pages table
		if ta.TableName == pageapi.RouteTable {
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
	// Seed Pages
	// =========================================================================
	pages, err := pagebus.TestSeedPages(ctx, 12, busDomain.Page)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding pages : %w", err)
	}

	// =========================================================================
	// Seed Role-Page associations
	// =========================================================================
	// Create page IDs slice
	pageIDs := make(uuid.UUIDs, len(pages))
	for i, p := range pages {
		pageIDs[i] = p.ID
	}

	// Associate tu1's roles with pages
	// First 3 pages: can_access=false (should be filtered out by query)
	// Next 3 pages: can_access=true (should be returned by query)
	// Remaining 6 pages: no association (should not appear in query)
	for i, pageID := range pageIDs[:6] {
		canAccess := i >= 3 // False for indices 0-2, true for indices 3-5
		for _, ur := range ur1 {
			_, err := busDomain.RolePage.Create(ctx, rolepagebus.NewRolePage{
				RoleID:    ur.RoleID,
				PageID:    pageID,
				CanAccess: canAccess,
			})
			if err != nil {
				return apitest.SeedData{}, fmt.Errorf("seeding role page for tu1: %w", err)
			}
		}
	}

	appPages := pageapp.ToAppPages(pages)

	return apitest.SeedData{
		Users:  []apitest.User{tu1},
		Admins: []apitest.User{tu2},
		Pages:  appPages,
	}, nil
}
