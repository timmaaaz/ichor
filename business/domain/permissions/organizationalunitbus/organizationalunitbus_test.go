package organizationalunitbus_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/domain/permissions/organizationalunitbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_OrganizationalUnit(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_OrganizationalUnit")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}
	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	// unitest.Run(t, update(db.BusDomain, sd), "update")
	// unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	orgUnits, err := organizationalunitbus.TestSeedOrganizationalUnits(ctx, 5, busDomain.OrganizationalUnit)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding organizational units : %w", err)
	}

	return unitest.SeedData{
		OrgUnits: orgUnits,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []organizationalunitbus.OrganizationalUnit{
				sd.OrgUnits[0],
				sd.OrgUnits[1],
				sd.OrgUnits[2],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.OrganizationalUnit.Query(ctx, organizationalunitbus.QueryFilter{}, organizationalunitbus.DefaultOrderBy, page.MustParse("1", "3"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]organizationalunitbus.OrganizationalUnit)
				if !exists {
					return "error occurred"
				}

				return cmp.Diff(gotResp, exp.([]organizationalunitbus.OrganizationalUnit))
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: organizationalunitbus.OrganizationalUnit{
				ParentID:              sd.OrgUnits[0].ID,
				Name:                  "Name5",
				Level:                 1,
				Path:                  strings.Join([]string{sd.OrgUnits[0].Path, "Name5"}, "."),
				CanInheritPermissions: true,
				CanRollupData:         true,
				UnitType:              "UnitType5",
				IsActive:              true,
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.OrganizationalUnit.Create(ctx, organizationalunitbus.NewOrganizationalUnit{
					ParentID:              sd.OrgUnits[0].ID,
					Name:                  "Name5",
					CanInheritPermissions: true,
					CanRollupData:         true,
					UnitType:              "UnitType5",
					IsActive:              true,
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(organizationalunitbus.OrganizationalUnit)
				if !exists {
					return "error occurred"
				}
				expResp, exists := exp.(organizationalunitbus.OrganizationalUnit)
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
			ExpResp: organizationalunitbus.OrganizationalUnit{
				ID:                    sd.OrgUnits[0].ID,
				ParentID:              sd.OrgUnits[0].ParentID,
				Name:                  "NewName0",
				Level:                 sd.OrgUnits[0].Level,
				Path:                  sd.OrgUnits[0].Path,
				CanInheritPermissions: sd.OrgUnits[0].CanInheritPermissions,
				CanRollupData:         sd.OrgUnits[0].CanRollupData,
				UnitType:              sd.OrgUnits[0].UnitType,
				IsActive:              sd.OrgUnits[0].IsActive,
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.OrganizationalUnit.Update(ctx, sd.OrgUnits[0], organizationalunitbus.UpdateOrganizationalUnit{
					Name: dbtest.StringPointer("NewName0"),
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(organizationalunitbus.OrganizationalUnit)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(organizationalunitbus.OrganizationalUnit)
				if !exists {
					return "error occurred"
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
				err := busDomain.OrganizationalUnit.Delete(ctx, sd.OrgUnits[0])
				if err != nil {
					return err
				}
				return nil
			},
			CmpFunc: func(got, exp any) string {
				if got != nil {
					return "error occurred"
				}
				return ""
			},
		},
	}
}
