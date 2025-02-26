package userorganizationbus_test

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/organizationalunitbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus"
	"github.com/timmaaaz/ichor/business/domain/permissions/userorganizationbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_UserOrganization(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_OrganizationalUnit")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}
	unitest.Run(t, query(db.BusDomain, sd), "query")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	roles, err := rolebus.TestSeedRoles(ctx, busDomain.Role)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding roles : %w", err)
	}

	roleIDs := make(uuid.UUIDs, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}

	orgUnits, err := organizationalunitbus.TestSeedOrganizationalUnits(ctx, busDomain.OrganizationalUnit)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding organizational units : %w", err)
	}
	orgUnitIDs := make(uuid.UUIDs, len(orgUnits))
	for i, ou := range orgUnits {
		orgUnitIDs[i] = ou.ID
	}

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, len(orgUnits)+1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}
	seedUsers := make([]unitest.User, len(usrs))
	for i, u := range usrs {
		seedUsers[i] = unitest.User{
			User: u,
		}
	}
	userIDs := make(uuid.UUIDs, len(usrs))
	for i, u := range usrs {
		userIDs[i] = u.ID
	}

	userOrgs, err := userorganizationbus.TestSeedUserOrganizations(ctx, orgUnitIDs, userIDs, roleIDs, busDomain.UserOrganization)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding user organizations : %w", err)
	}

	return unitest.SeedData{
		Users:    seedUsers,
		Roles:    roles,
		OrgUnits: orgUnits,
		UserOrgs: userOrgs,
	}, nil

}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {

	// Make copy for testing
	sortedUserOrgs := make([]userorganizationbus.UserOrganization, len(sd.UserOrgs))
	copy(sortedUserOrgs, sd.UserOrgs)

	// Sort by user id
	sort.Slice(sortedUserOrgs, func(i, j int) bool {
		return sortedUserOrgs[i].UserID.String() < sortedUserOrgs[j].UserID.String()
	})

	return []unitest.Table{
		{
			Name: "UserOrganization",
			ExpResp: []userorganizationbus.UserOrganization{
				sortedUserOrgs[0],
				sortedUserOrgs[1],
				sortedUserOrgs[2],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.UserOrganization.Query(ctx, userorganizationbus.QueryFilter{}, userorganizationbus.DefaultOrderBy, page.MustParse("1", "3"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(exp, got any) string {
				gotResp, exists := got.([]userorganizationbus.UserOrganization)
				if !exists {
					return "got is not a []userorganizationbus.UserOrganization"

				}
				expResp, exists := exp.([]userorganizationbus.UserOrganization)
				if !exists {
					return "exp is not a []userorganizationbus.UserOrganization"
				}

				for i := range gotResp {
					if gotResp[i].CreatedAt.Format(time.RFC3339) == expResp[i].CreatedAt.Format(time.RFC3339) {
						expResp[i].CreatedAt = gotResp[i].CreatedAt
					}
				}

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}
