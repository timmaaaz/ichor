package paperworkapi_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/api/domain/http/paperwork/paperworkapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

// PaperworkSeed extends apitest.SeedData with paperwork-specific fixture IDs.
// The apitest harness's SeedData is a fixed shape; paperwork tests carry
// the seeded entity IDs in this side-channel struct.
type PaperworkSeed struct {
	apitest.SeedData
	OrderID    uuid.UUID
	OrderNum   string
	PurchaseID uuid.UUID
	POOrderNum string
	TransferID uuid.UUID
	TransferNo string
}

// insertSeedData stages everything the paperwork integration tests need:
//   - one admin (full permissions, used for happy-path 200)
//   - one regular user with the three paperwork-relevant table_access rows
//     downgraded to 0 (used for 403 cases)
//   - one sales order, one purchase order, one transfer order (used as
//     query targets for happy-path 200)
//
// Mirrors labelapi/seed_test.go role-downgrade pattern.
func insertSeedData(db *dbtest.Database, ath *auth.Auth) (PaperworkSeed, error) {
	ctx := context.Background()
	bd := db.BusDomain

	// Users
	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, bd.User)
	if err != nil {
		return PaperworkSeed{}, fmt.Errorf("seeding users: %w", err)
	}
	tu1 := apitest.User{User: usrs[0], Token: apitest.Token(bd.User, ath, usrs[0].Email.Address)}

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, bd.User)
	if err != nil {
		return PaperworkSeed{}, fmt.Errorf("seeding admins: %w", err)
	}
	tu2 := apitest.User{User: admins[0], Token: apitest.Token(bd.User, ath, admins[0].Email.Address)}

	// Domain entities — Task 8 will replace these stubs with real seed code
	// using the same testutil patterns as paperworkbus_test.go's seedAll.
	orderID, orderNum, err := seedOneSalesOrder(ctx, bd)
	if err != nil {
		return PaperworkSeed{}, fmt.Errorf("seeding sales order: %w", err)
	}
	poID, poNum, err := seedOnePurchaseOrder(ctx, bd)
	if err != nil {
		return PaperworkSeed{}, fmt.Errorf("seeding purchase order: %w", err)
	}
	transferID, transferNum, err := seedOneTransferOrder(ctx, bd)
	if err != nil {
		return PaperworkSeed{}, fmt.Errorf("seeding transfer order: %w", err)
	}

	// Permissions — downgrade tu1's table_access for all three paperwork
	// RouteTables so 403 cases fire reliably.
	roles, err := rolebus.TestSeedRoles(ctx, 2, bd.Role)
	if err != nil {
		return PaperworkSeed{}, fmt.Errorf("seeding roles: %w", err)
	}
	roleIDs := make([]uuid.UUID, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}

	if _, err := userrolebus.TestSeedUserRoles(ctx, []uuid.UUID{tu1.ID, tu2.ID}, roleIDs, bd.UserRole); err != nil {
		return PaperworkSeed{}, fmt.Errorf("seeding user roles: %w", err)
	}
	if _, err := tableaccessbus.TestSeedTableAccess(ctx, roleIDs, bd.TableAccess); err != nil {
		return PaperworkSeed{}, fmt.Errorf("seeding table access: %w", err)
	}

	ur1, err := bd.UserRole.QueryByUserID(ctx, tu1.ID)
	if err != nil {
		return PaperworkSeed{}, fmt.Errorf("querying user1 roles: %w", err)
	}
	usrRoleIDs := make(uuid.UUIDs, len(ur1))
	for i, r := range ur1 {
		usrRoleIDs[i] = r.RoleID
	}
	tas, err := bd.TableAccess.QueryByRoleIDs(ctx, usrRoleIDs)
	if err != nil {
		return PaperworkSeed{}, fmt.Errorf("querying table access: %w", err)
	}

	downgrade := func(table string) error {
		for _, ta := range tas {
			if ta.TableName != table {
				continue
			}
			update := tableaccessbus.UpdateTableAccess{
				CanCreate: dbtest.BoolPointer(false),
				CanUpdate: dbtest.BoolPointer(false),
				CanDelete: dbtest.BoolPointer(false),
				CanRead:   dbtest.BoolPointer(false),
			}
			if _, err := bd.TableAccess.Update(ctx, ta, update); err != nil {
				return fmt.Errorf("downgrading %s: %w", table, err)
			}
		}
		return nil
	}
	for _, table := range []string{
		paperworkapi.RouteTablePickSheet,
		paperworkapi.RouteTableReceiveCover,
		paperworkapi.RouteTableTransferSheet,
	} {
		if err := downgrade(table); err != nil {
			return PaperworkSeed{}, err
		}
	}

	return PaperworkSeed{
		SeedData: apitest.SeedData{
			Admins: []apitest.User{tu2},
			Users:  []apitest.User{tu1},
		},
		OrderID:    orderID,
		OrderNum:   orderNum,
		PurchaseID: poID,
		POOrderNum: poNum,
		TransferID: transferID,
		TransferNo: transferNum,
	}, nil
}

// Seed-helper stubs — Task 8 fills these in (the bus_test.go seedAll
// function in paperworkbus_test.go is the reference implementation).
func seedOneSalesOrder(_ context.Context, _ dbtest.BusDomain) (uuid.UUID, string, error) {
	return uuid.Nil, "", fmt.Errorf("VFY: implement at task time using ordersbus seed pattern")
}
func seedOnePurchaseOrder(_ context.Context, _ dbtest.BusDomain) (uuid.UUID, string, error) {
	return uuid.Nil, "", fmt.Errorf("VFY: implement at task time using purchaseorderbus.testutil")
}
func seedOneTransferOrder(_ context.Context, _ dbtest.BusDomain) (uuid.UUID, string, error) {
	return uuid.Nil, "", fmt.Errorf("VFY: implement at task time using transferorderbus.testutil")
}
