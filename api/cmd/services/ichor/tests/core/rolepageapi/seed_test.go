package rolepage_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/pageapp"
	"github.com/timmaaaz/ichor/app/domain/core/rolepageapp"
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
		return apitest.SeedData{}, fmt.Errorf("seeding users: %w", err)
	}
	tu1 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
	}

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users: %w", err)
	}
	tu2 := apitest.User{
		User:  admins[0],
		Token: apitest.Token(db.BusDomain.User, ath, admins[0].Email.Address),
	}

	// Seed roles for table access and role-page relationships
	roles, err := rolebus.TestSeedRoles(ctx, 3, busDomain.Role)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding roles: %w", err)
	}

	// Assign roles to both users
	roleIDs := make(uuid.UUIDs, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}

	userIDs := uuid.UUIDs{tu1.ID, tu2.ID}
	_, err = userrolebus.TestSeedUserRoles(ctx, userIDs, roleIDs, busDomain.UserRole)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding user roles: %w", err)
	}

	// Seed table access for all roles
	_, err = tableaccessbus.TestSeedTableAccess(ctx, roleIDs, busDomain.TableAccess)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding table access: %w", err)
	}

	// Seed pages for role-page relationships
	pages, err := pagebus.TestSeedPages(ctx, 12, busDomain.Page)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding roles: %w", err)
	}

	// Create role-page mappings using the first role and first few pages
	rolePages, err := rolepagebus.TestGenerateSeedRolePages(ctx, busDomain.RolePage,
		[]uuid.UUID{roles[0].ID},
		[]uuid.UUID{
			pages[0].ID,
			pages[1].ID,
			pages[2].ID,
			pages[3].ID,
			pages[4].ID,
		},
	)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding role pages: %w", err)
	}

	sd := apitest.SeedData{
		Users:     []apitest.User{tu1},
		Admins:    []apitest.User{tu2},
		Pages:     toAppPages(pages),
		RolePages: toAppRolePages(rolePages),
	}

	return sd, nil
}

// =============================================================================

func toAppPage(bus pagebus.Page) pageapp.Page {
	return pageapp.Page{
		ID:        bus.ID.String(),
		Path:      bus.Path,
		Name:      bus.Name,
		Module:    bus.Module,
		Icon:      bus.Icon,
		SortOrder: bus.SortOrder,
		IsActive:  bus.IsActive,
	}
}

func toAppPages(bus []pagebus.Page) []pageapp.Page {
	app := make([]pageapp.Page, len(bus))
	for i, b := range bus {
		app[i] = toAppPage(b)
	}
	return app
}

func toAppRolePage(bus rolepagebus.RolePage) rolepageapp.RolePage {
	return rolepageapp.RolePage{
		ID:        bus.ID.String(),
		RoleID:    bus.RoleID.String(),
		PageID:    bus.PageID.String(),
		CanAccess: bus.CanAccess,
	}
}

func toAppRolePages(bus []rolepagebus.RolePage) []rolepageapp.RolePage {
	app := make([]rolepageapp.RolePage, len(bus))
	for i, b := range bus {
		app[i] = toAppRolePage(b)
	}
	return app
}
