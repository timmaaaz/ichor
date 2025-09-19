package approvalstatus_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/assets/approvalstatusapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/assets/approvalstatusapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
	ctx := context.Background()

	busDomain := db.BusDomain

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
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

	approvalstatus1, err := busDomain.ApprovalStatus.Query(ctx, approvalstatusbus.QueryFilter{}, order.NewBy(approvalstatusbus.OrderByName, order.ASC), page.MustParse("1", "2"))
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying approval statuses : %w", err)
	}

	approvalstatus2, err := busDomain.ApprovalStatus.Query(ctx, approvalstatusbus.QueryFilter{}, order.NewBy(approvalstatusbus.OrderByName, order.ASC), page.MustParse("2", "2"))
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying approval statuses : %w", err)
	}

	apprvls := append(approvalstatus1, approvalstatus2...)

	appApprvls := approvalstatusapp.ToAppApprovalStatuses(apprvls)

	// =========================================================================
	// Permissions stuff
	// =========================================================================
	roles, err := rolebus.TestSeedRoles(ctx, len(usrs), busDomain.Role)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding roles : %w", err)
	}

	roleIDs := make(uuid.UUIDs, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}
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

	tas, err := busDomain.TableAccess.QueryByRoleIDs(ctx, []uuid.UUID{roleIDs[0]})
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying table access : %w", err)
	}
	for _, ta := range tas {
		if ta.TableName == approvalstatusapi.RouteTable {
			continue
		}
		update := tableaccessbus.UpdateTableAccess{
			CanCreate: dbtest.BoolPointer(false),
			CanUpdate: dbtest.BoolPointer(false),
			CanDelete: dbtest.BoolPointer(false),
			CanRead:   dbtest.BoolPointer(true),
		}
		_, err = busDomain.TableAccess.Update(ctx, ta, update)
		if err != nil {
			return apitest.SeedData{}, fmt.Errorf("updating table access : %w", err)
		}
	}

	sd := apitest.SeedData{
		Users:            []apitest.User{tu1},
		Admins:           []apitest.User{tu2},
		ApprovalStatuses: appApprvls,
	}

	return sd, nil

}
