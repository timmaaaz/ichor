package userrolebus_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus"
	"github.com/timmaaaz/ichor/business/domain/permissions/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_UserRole(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_UserRole")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 21, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	roles, err := rolebus.TestSeedRoles(ctx, busDomain.Role)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding roles : %w", err)
	}

	roleIDs := make(uuid.UUIDs, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}
	userIDs := make(uuid.UUIDs, len(usrs))
	for i, u := range usrs {
		userIDs[i] = u.ID
	}

	userRoles, err := userrolebus.TestSeedUserRoles(ctx, userIDs, roleIDs, busDomain.UserRole)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding user roles : %w", err)
	}

	seedUsers := make([]unitest.User, len(usrs))
	for i, u := range usrs {
		seedUsers[i] = unitest.User{
			User: u,
		}
	}

	sd := unitest.SeedData{
		Users:     seedUsers,
		Roles:     roles,
		UserRoles: userRoles,
	}
	return sd, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {

	// make copy and sort by user id for comparison
	sdUserRoles := make([]userrolebus.UserRole, len(sd.UserRoles))
	copy(sdUserRoles, sd.UserRoles)
	sort.Slice(sdUserRoles, func(i, j int) bool {
		return sdUserRoles[i].UserID.String() < sdUserRoles[j].UserID.String()
	})

	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []userrolebus.UserRole{
				sdUserRoles[0],
				sdUserRoles[1],
				sdUserRoles[2],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.UserRole.Query(ctx, userrolebus.QueryFilter{}, userrolebus.DefaultOrderBy, page.MustParse("1", "3"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]userrolebus.UserRole)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.([]userrolebus.UserRole)
				if !exists {
					return "error occurred"
				}

				// Sort arrays to make sure they are in the same order.
				sort.Slice(gotResp, func(i, j int) bool {
					return gotResp[i].ID.String() < gotResp[j].ID.String()
				})

				sort.Slice(expResp, func(i, j int) bool {
					return expResp[i].ID.String() < expResp[j].ID.String()
				})

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: userrolebus.UserRole{
				UserID: sd.Users[len(sd.Users)-1].ID,
				RoleID: sd.Roles[len(sd.Roles)-1].ID,
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.UserRole.Create(ctx, userrolebus.NewUserRole{
					UserID: sd.Users[len(sd.Users)-1].ID,
					RoleID: sd.Roles[len(sd.Roles)-1].ID,
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(userrolebus.UserRole)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(userrolebus.UserRole)
				if !exists {
					return "error occurred"
				}

				expResp.ID = gotResp.ID
				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Update",
			ExpResp: userrolebus.UserRole{
				UserID: sd.UserRoles[0].UserID,
				RoleID: sd.Roles[3].ID,
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.UserRole.Update(ctx, sd.UserRoles[0], userrolebus.UpdateUserRole{
					RoleID: &sd.Roles[3].ID,
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(userrolebus.UserRole)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(userrolebus.UserRole)
				if !exists {
					return "error occurred"
				}

				expResp.ID = gotResp.ID
				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Delete",
			ExcFunc: func(ctx context.Context) any {
				return busDomain.UserRole.Delete(ctx, sd.UserRoles[0])
			},
			CmpFunc: func(got, exp any) string {
				return ""
			},
		},
	}
}
