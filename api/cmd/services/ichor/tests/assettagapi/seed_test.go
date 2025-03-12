package assettag_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/assets/assettagapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/assets/assettagapp"
	"github.com/timmaaaz/ichor/app/domain/assets/tagapp"
	"github.com/timmaaaz/ichor/app/domain/assets/validassetapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/assets/assettagbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/assets/tagbus"
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

	// =================== Asset =================
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

	asIDs := make([]uuid.UUID, len(as))
	for i, asset := range as {
		asIDs[i] = asset.ID
	}

	// =================== Tags ====================

	tags, err := tagbus.TestSeedTag(ctx, 15, busDomain.Tag)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding tags : %w", err)
	}

	tIDs := make([]uuid.UUID, 0, len(tags))
	for _, t := range tags {
		tIDs = append(tIDs, t.ID)
	}

	// =================== Asset-Tag ====================

	assetTags, err := assettagbus.TestSeedAssetTag(ctx, 11, asIDs, tIDs, busDomain.AssetTag)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding asset tags : %w", err)
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
	tas, err := busDomain.TableAccess.QueryByRoleIDs(ctx, []uuid.UUID{ur1.RoleID})
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying table access : %w", err)
	}

	// Update only tu1's role permissions
	for _, ta := range tas {
		// Only update for the asset table
		if ta.TableName == assettagapi.RouteTable {
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
		Tags:        tagapp.ToAppTags(tags),
		AssetTags:   assettagapp.ToAppAssetTags(assetTags),
	}

	return sd, nil
}
