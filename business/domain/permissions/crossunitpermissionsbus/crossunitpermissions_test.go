package crossunitpermissionsbus_test

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/crossunitpermissionsbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/organizationalunitbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_CrossUnitPermissions(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_OrganizationalUnit")

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

	cups, err := crossunitpermissionsbus.TestSeedCrossUnitPermissions(ctx, orgUnitIDs, orgUnitIDs, userIDs[0], busDomain.CrossUnitPermissions)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding cross unit permissions : %w", err)
	}

	return unitest.SeedData{
		Roles:                roles,
		OrgUnits:             orgUnits,
		Users:                seedUsers,
		CrossUnitPermissions: cups,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	// Make sorted copy for testing
	crossUnitPermissions := make([]crossunitpermissionsbus.CrossUnitPermission, len(sd.CrossUnitPermissions))
	copy(crossUnitPermissions, sd.CrossUnitPermissions)

	sort.Slice(crossUnitPermissions, func(i, j int) bool {
		return crossUnitPermissions[i].ID.String() < crossUnitPermissions[j].ID.String()
	})

	return []unitest.Table{
		{
			Name: "CrossUnitPermissions",
			ExpResp: []crossunitpermissionsbus.CrossUnitPermission{
				crossUnitPermissions[0],
				crossUnitPermissions[1],
				crossUnitPermissions[2],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.CrossUnitPermissions.Query(ctx, crossunitpermissionsbus.QueryFilter{}, crossunitpermissionsbus.DefaultOrderBy, page.MustParse("1", "3"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(exp, got any) string {
				gotResp, exists := got.([]crossunitpermissionsbus.CrossUnitPermission)
				if !exists {
					return "expected []crossunitpermissionsbus.CrossUnitPermission"
				}

				expResp, exists := exp.([]crossunitpermissionsbus.CrossUnitPermission)
				if !exists {
					return "expected []crossunitpermissionsbus.CrossUnitPermission"
				}

				for i := range expResp {
					if gotResp[i].ValidFrom.Format(time.RFC3339) == expResp[i].ValidFrom.Format(time.RFC3339) {
						expResp[i].ValidFrom = gotResp[i].ValidFrom
					}
					if gotResp[i].ValidUntil.Format(time.RFC3339) == expResp[i].ValidUntil.Format(time.RFC3339) {
						expResp[i].ValidUntil = gotResp[i].ValidUntil
					}

				}

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	now := time.Now()

	return []unitest.Table{
		{
			Name: "CrossUnitPermissions",
			ExpResp: crossunitpermissionsbus.CrossUnitPermission{
				SourceUnitID: sd.OrgUnits[0].ID,
				TargetUnitID: sd.OrgUnits[1].ID,
				CanRead:      true,
				CanUpdate:    false,
				GrantedBy:    sd.Users[0].ID,
				ValidFrom:    now,
				ValidUntil:   now.AddDate(0, 0, 1),
				Reason:       "testing",
			},
			ExcFunc: func(ctx context.Context) any {
				ncup := crossunitpermissionsbus.NewCrossUnitPermission{
					SourceUnitID: sd.OrgUnits[0].ID,
					TargetUnitID: sd.OrgUnits[1].ID,
					CanRead:      true,
					CanUpdate:    false,
					GrantedBy:    sd.Users[0].ID,
					ValidFrom:    now,
					ValidUntil:   now.AddDate(0, 0, 1),
					Reason:       "testing",
				}

				cup, err := busDomain.CrossUnitPermissions.Create(ctx, ncup)
				if err != nil {
					return err
				}
				return cup
			},
			CmpFunc: func(exp, got any) string {
				gotResp, exists := got.(crossunitpermissionsbus.CrossUnitPermission)
				if !exists {
					return "got is not a crossunitpermissionsbus.CrossUnitPermission"
				}
				expResp, exists := exp.(crossunitpermissionsbus.CrossUnitPermission)
				if !exists {
					return "exp is not a crossunitpermissionsbus.CrossUnitPermission"
				}

				expResp.ID = gotResp.ID

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	exp := sd.CrossUnitPermissions[0]
	exp.CanRead = false
	exp.CanUpdate = true

	return []unitest.Table{
		{
			Name:    "CrossUnitPermissions",
			ExpResp: exp,
			ExcFunc: func(ctx context.Context) any {
				ucup := crossunitpermissionsbus.UpdateCrossUnitPermission{
					CanRead:   &exp.CanRead,
					CanUpdate: &exp.CanUpdate,
				}

				cup, err := busDomain.CrossUnitPermissions.Update(ctx, exp, ucup)
				if err != nil {
					return err
				}
				return cup
			},
			CmpFunc: func(exp, got any) string {
				gotResp, exists := got.(crossunitpermissionsbus.CrossUnitPermission)
				if !exists {
					return "got is not a crossunitpermissionsbus.CrossUnitPermission"
				}
				expResp, exists := exp.(crossunitpermissionsbus.CrossUnitPermission)
				if !exists {
					return "exp is not a crossunitpermissionsbus.CrossUnitPermission"
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
				err := busDomain.CrossUnitPermissions.Delete(ctx, sd.CrossUnitPermissions[0])
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
