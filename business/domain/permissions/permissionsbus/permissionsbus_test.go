package permissionsbus_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus"
	"github.com/timmaaaz/ichor/business/domain/permissions/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_Permissions(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_Permissions")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, query(db.BusDomain, sd), "query")
	// unitest.Run(t, create(db.BusDomain, sd), "create")
	// unitest.Run(t, update(db.BusDomain, sd), "update")
	// unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 3, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	roles, err := rolebus.TestSeedRoles(ctx, 4, busDomain.Role)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding roles : %w", err)
	}
	roleIDs := make(uuid.UUIDs, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}

	userRoles, err := userrolebus.TestSeedUserRoles(ctx, 3, usrs[0].ID, roleIDs, busDomain.UserRole)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding user roles : %w", err)
	}

	tables := []string{"countries", "regions", "cities"}
	tableAccesses, err := tableaccessbus.TestSeedTableAccesses(ctx, 3, roleIDs[0], tables, busDomain.TableAccess)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding table accesses : %w", err)
	}

	seedUsers := make([]unitest.User, len(usrs))
	for i, u := range usrs {
		seedUsers[i] = unitest.User{
			User: u,
		}
	}

	sd := unitest.SeedData{
		Users:         seedUsers,
		Roles:         roles,
		UserRoles:     userRoles,
		TableAccesses: tableAccesses,
	}
	return sd, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: permissionsbus.UserPermissions{
				UserID:   sd.Users[0].ID,
				Username: sd.Users[0].Username.String(),
				Roles: []permissionsbus.UserRole{
					{
						RoleID: sd.Roles[0].ID,
						Name:   sd.Roles[0].Name,
						Tables: []permissionsbus.TableAccess{
							{
								TableName: "countries",
								CanCreate: true,
								CanRead:   true,
								CanUpdate: true,
								CanDelete: true,
							},
							{
								TableName: "regions",
								CanCreate: true,
								CanRead:   true,
								CanUpdate: true,
								CanDelete: true,
							},
							{
								TableName: "cities",
								CanCreate: true,
								CanRead:   true,
								CanUpdate: true,
								CanDelete: true,
							},
						},
					},
				},
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.Permissions.QueryUserPermissions(ctx, sd.Users[0].ID)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(exp, got any) string {
				gotResp, exists := got.(permissionsbus.UserPermissions)
				if !exists {
					return fmt.Sprintf("expected permissionsbus.UserPermissions, got %T", got)
				}

				expResp := exp.(permissionsbus.UserPermissions)

				// Match RoleID from the actual response since it might be dynamically generated
				expResp.Roles[0].RoleID = gotResp.Roles[0].RoleID

				// Create a comparison option that ignores table order
				sortTables := cmp.Transformer("SortTables", func(in []permissionsbus.TableAccess) []permissionsbus.TableAccess {
					if len(in) == 0 {
						return in
					}
					out := append([]permissionsbus.TableAccess{}, in...) // Copy to avoid mutating input
					sort.Slice(out, func(i, j int) bool {
						return out[i].TableName < out[j].TableName
					})
					return out
				})

				// Compare with the custom sorting option
				return cmp.Diff(expResp, gotResp, sortTables)
			},
		},
	}
}
