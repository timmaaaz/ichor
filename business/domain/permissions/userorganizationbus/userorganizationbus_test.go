package userorganizationbus_test

import (
	"context"
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

	unitest.Run(t, query(db.BusDomain), "query")
	// unitest.Run(t, create(db.BusDomain, sd), "create")
	// unitest.Run(t, update(db.BusDomain, sd), "update")
	// unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func query(busDomain dbtest.BusDomain) []unitest.Table {
	users, err := busDomain.User.Query(context.Background(), userbus.QueryFilter{}, userbus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		panic(err)
	}

	orgUnits, err := busDomain.OrganizationalUnit.Query(context.Background(), organizationalunitbus.QueryFilter{}, userbus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		panic(err)
	}

	roles, err := busDomain.Role.Query(context.Background(), rolebus.QueryFilter{}, userbus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		panic(err)
	}

	return []unitest.Table{
		{
			Name: "UserOrganization",
			ExpResp: []userorganizationbus.UserOrganization{
				{
					ID:                   uuid.Nil,
					OrganizationalUnitID: uuid.Nil,
					UserID:               uuid.Nil,
					RoleID:               uuid.Nil,
					IsUnitManager:        true,
					CreatedBy:            uuid.Nil,
				},
				{
					ID:                   uuid.Nil,
					OrganizationalUnitID: uuid.Nil,
					UserID:               uuid.Nil,
					RoleID:               uuid.Nil,
					IsUnitManager:        false,
					CreatedBy:            uuid.Nil,
				},
				{
					ID:                   uuid.Nil,
					OrganizationalUnitID: uuid.Nil,
					UserID:               uuid.Nil,
					RoleID:               uuid.Nil,
					IsUnitManager:        true,
					CreatedBy:            uuid.Nil,
				},
				{
					ID:                   uuid.Nil,
					OrganizationalUnitID: uuid.Nil,
					UserID:               uuid.Nil,
					RoleID:               uuid.Nil,
					IsUnitManager:        true,
					CreatedBy:            uuid.Nil,
				},
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.UserOrganization.Query(ctx, userorganizationbus.QueryFilter{}, userorganizationbus.DefaultOrderBy, page.MustParse("1", "4"))
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

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: userorganizationbus.UserOrganization{
				OrganizationalUnitID: sd.OrgUnits[0].ID,
				UserID:               sd.Users[len(sd.Users)-1].ID,
				RoleID:               sd.Roles[0].ID,
				CreatedBy:            sd.Users[len(sd.Users)-1].ID,
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.UserOrganization.Create(ctx, userorganizationbus.NewUserOrganization{
					OrganizationalUnitID: sd.OrgUnits[0].ID,
					UserID:               sd.Users[len(sd.Users)-1].ID,
					RoleID:               sd.Roles[0].ID,
					CreatedBy:            sd.Users[len(sd.Users)-1].ID,
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(exp, got any) string {
				gotResp, exists := got.(userorganizationbus.UserOrganization)
				if !exists {
					return "got is not a userorganizationbus.UserOrganization"
				}
				expResp, exists := exp.(userorganizationbus.UserOrganization)
				if !exists {
					return "exp is not a userorganizationbus.UserOrganization"
				}

				expResp.CreatedAt = gotResp.CreatedAt
				expResp.ID = gotResp.ID

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	exp := sd.UserOrgs[0]
	exp.RoleID = sd.Roles[1].ID
	exp.IsUnitManager = true

	return []unitest.Table{
		{
			Name:    "Update",
			ExpResp: exp,
			ExcFunc: func(ctx context.Context) any {
				uuo := userorganizationbus.UpdateUserOrganization{
					RoleID:        &sd.Roles[1].ID,
					IsUnitManager: dbtest.BoolPointer(true),
				}

				resp, err := busDomain.UserOrganization.Update(ctx, sd.UserOrgs[0], uuo)
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(exp, got any) string {
				gotResp, exists := got.(userorganizationbus.UserOrganization)
				if !exists {
					return "got is not a userorganizationbus.UserOrganization"
				}
				expResp, exists := exp.(userorganizationbus.UserOrganization)
				if !exists {
					return "exp is not a userorganizationbus.UserOrganization"
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
				err := busDomain.UserOrganization.Delete(ctx, sd.UserOrgs[0])
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
