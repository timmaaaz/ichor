package orgunitcolumnaccessbus_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/organizationalunitbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/orgunitcolumnaccessbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus"
	"github.com/timmaaaz/ichor/business/domain/permissions/userorganizationbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_OrgUnitColumnAccess(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_OrganizationalUnit")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}
	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, createColNotExists(db.BusDomain, sd), "createColNotExists")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
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

	orgUnitColAccess, err := orgunitcolumnaccessbus.TestSeedOrgUnitColumnAccesses(ctx, orgUnitIDs, busDomain.OrgUnitColAccess)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding org unit column accesses : %w", err)
	}

	return unitest.SeedData{
		Users:              seedUsers,
		Roles:              roles,
		OrgUnits:           orgUnits,
		UserOrgs:           userOrgs,
		OrgUnitColAccesses: orgUnitColAccess,
	}, nil

}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {

	// Make sorted copy for testing
	sortedOrgUnitColAccesses := make([]orgunitcolumnaccessbus.OrgUnitColumnAccess, len(sd.OrgUnitColAccesses))
	copy(sortedOrgUnitColAccesses, sd.OrgUnitColAccesses)

	sort.Slice(sortedOrgUnitColAccesses, func(i, j int) bool {
		return sortedOrgUnitColAccesses[i].OrganizationalUnitID.String() < sortedOrgUnitColAccesses[j].OrganizationalUnitID.String()
	})

	return []unitest.Table{
		{
			Name: "OrgUnitColumnAccess",
			ExpResp: []orgunitcolumnaccessbus.OrgUnitColumnAccess{
				sortedOrgUnitColAccesses[0],
				sortedOrgUnitColAccesses[1],
				sortedOrgUnitColAccesses[2],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.OrgUnitColAccess.Query(ctx, orgunitcolumnaccessbus.QueryFilter{}, orgunitcolumnaccessbus.DefaultOrderBy, page.MustParse("1", "3"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(exp, got any) string {
				gotResp, exists := got.([]orgunitcolumnaccessbus.OrgUnitColumnAccess)
				if !exists {
					return "got is not a []orgunitcolumnaccessbus.OrgUnitColumnAccess"

				}
				expResp, exists := exp.([]orgunitcolumnaccessbus.OrgUnitColumnAccess)
				if !exists {
					return "exp is not a []orgunitcolumnaccessbus.OrgUnitColumnAccess"
				}

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "OrgUnitColumnAccess",
			ExpResp: orgunitcolumnaccessbus.OrgUnitColumnAccess{
				OrganizationalUnitID:  sd.OrgUnits[0].ID,
				TableName:             "contact_info",
				ColumnName:            "primary_phone_number",
				CanRead:               true,
				CanUpdate:             true,
				CanInheritPermissions: true,
				CanRollupData:         true,
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.OrgUnitColAccess.Create(ctx, orgunitcolumnaccessbus.NewOrgUnitColumnAccess{
					OrganizationalUnitID:  sd.OrgUnits[0].ID,
					TableName:             "contact_info",
					ColumnName:            "primary_phone_number",
					CanRead:               true,
					CanUpdate:             true,
					CanInheritPermissions: true,
					CanRollupData:         true,
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(exp, got any) string {
				gotResp, exists := got.(orgunitcolumnaccessbus.OrgUnitColumnAccess)
				if !exists {
					return "got is not a orgunitcolumnaccessbus.OrgUnitColumnAccess"
				}
				expResp, exists := exp.(orgunitcolumnaccessbus.OrgUnitColumnAccess)
				if !exists {
					return "exp is not a orgunitcolumnaccessbus.OrgUnitColumnAccess"
				}

				expResp.ID = gotResp.ID

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func createColNotExists(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {

	tmp := fmt.Errorf("exists: %w", orgunitcolumnaccessbus.ErrColumnNotExists)
	exp := fmt.Errorf("checking if org unit column access column exists: %w", tmp)

	return []unitest.Table{
		{
			Name:    "OrgUnitColumnAccess",
			ExpResp: exp,
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.OrgUnitColAccess.Create(ctx, orgunitcolumnaccessbus.NewOrgUnitColumnAccess{
					OrganizationalUnitID:  sd.OrgUnits[0].ID,
					TableName:             "asdf",
					ColumnName:            "asdf",
					CanRead:               true,
					CanUpdate:             true,
					CanInheritPermissions: true,
					CanRollupData:         true,
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(exp, got any) string {
				gotErr, ok := got.(error)
				if !ok {
					return fmt.Sprintf("expected error, got %T", got)
				}

				expErr, ok := exp.(error)
				if !ok {
					return fmt.Sprintf("expected error, got %T", exp)
				}

				if gotErr.Error() != expErr.Error() {
					return fmt.Sprintf("expected error %q, got %q", expErr, gotErr)
				}

				return "" // Return empty string on success
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	exp := sd.OrgUnitColAccesses[0]
	exp.CanRollupData = true

	return []unitest.Table{
		{
			Name:    "OrgUnitColumnAccess",
			ExpResp: exp,
			ExcFunc: func(ctx context.Context) any {
				update := orgunitcolumnaccessbus.UpdateOrgUnitColumnAccess{
					CanRollupData: dbtest.BoolPointer(true),
				}
				resp, err := busDomain.OrgUnitColAccess.Update(ctx, sd.OrgUnitColAccesses[0], update)
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(exp, got any) string {
				gotResp, exists := got.(orgunitcolumnaccessbus.OrgUnitColumnAccess)
				if !exists {
					return "got is not a orgunitcolumnaccessbus.OrgUnitColumnAccess"
				}
				expResp, exists := exp.(orgunitcolumnaccessbus.OrgUnitColumnAccess)
				if !exists {
					return "exp is not a orgunitcolumnaccessbus.OrgUnitColumnAccess"
				}

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "Delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.OrgUnitColAccess.Delete(ctx, sd.OrgUnitColAccesses[0])
				if err != nil {
					return err
				}
				return nil
			},
			CmpFunc: func(exp, got any) string {
				if got != nil {
					return "expected nil"
				}
				return ""
			},
		},
	}
}
