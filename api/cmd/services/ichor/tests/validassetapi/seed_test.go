package validasset_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/assets/validassetapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"

	"github.com/timmaaaz/ichor/app/domain/assets/assettypeapp"
	"github.com/timmaaaz/ichor/app/domain/assets/validassetapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"

	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus"
	"github.com/timmaaaz/ichor/business/domain/permissions/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/userrolebus"

	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	usr, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}
	tu1 := apitest.User{
		User:  usr[0],
		Token: apitest.Token(db.BusDomain.User, ath, usr[0].Email.Address),
	}
	admin, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}
	tu2 := apitest.User{
		User:  admin[0],
		Token: apitest.Token(db.BusDomain.User, ath, admin[0].Email.Address),
	}

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

	tas, err := busDomain.TableAccess.QueryByRoleIDs(ctx, usrRoleIDs)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying table access : %w", err)
	}

	// Update only tu1's role permissions
	for _, ta := range tas {
		// Only update for the asset table
		if ta.TableName == validassetapi.RouteTable {
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

	sd := apitest.SeedData{
		Users:       []apitest.User{tu1},
		Admins:      []apitest.User{tu2},
		ValidAssets: validassetapp.ToAppValidAssets(as),
		AssetTypes:  assettypeapp.ToAppAssetTypes(ats),
	}

	return sd, nil
}
