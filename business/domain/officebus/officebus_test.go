package officebus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	"github.com/timmaaaz/ichor/business/domain/officebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_Office(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_Office")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("querying regions : %w", err)
	}

	ids := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		ids = append(ids, r.ID)
	}

	ctys, err := citybus.TestSeedCities(ctx, 10, ids, busDomain.City)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding cities : %w", err)
	}

	ctyIDs := make([]uuid.UUID, 0, len(ctys))
	for _, c := range ctys {
		ctyIDs = append(ctyIDs, c.ID)
	}

	strs, err := streetbus.TestSeedStreets(ctx, 10, ctyIDs, busDomain.Street)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding streets : %w", err)
	}

	streetIDs := make(uuid.UUIDs, len(strs))
	for i, street := range strs {
		streetIDs[i] = street.ID
	}

	offices, err := officebus.TestSeedOffices(ctx, 10, streetIDs, busDomain.Office)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding offices : %w", err)
	}

	return unitest.SeedData{
		Offices: offices,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "Query",
			ExpResp: []officebus.Office{
				sd.Offices[0],
				sd.Offices[1],
				sd.Offices[2],
				sd.Offices[3],
				sd.Offices[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.Office.Query(ctx, officebus.QueryFilter{}, officebus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(exp, got)
			},
		},
		{
			Name:    "Query by id",
			ExpResp: sd.Offices[0],
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.Office.QueryByID(ctx, sd.Offices[0].ID)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(exp, got)
			},
		},
	}
	return table
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "Create",
			ExpResp: officebus.Office{
				Name:     "Test Office",
				StreetID: sd.Offices[0].StreetID,
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.Office.Create(ctx, officebus.NewOffice{
					Name:     "Test Office",
					StreetID: sd.Offices[0].StreetID,
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(officebus.Office)
				if !exists {
					return fmt.Sprintf("got is not an asset type: %v", got)
				}

				expResp := exp.(officebus.Office)
				expResp.ID = gotResp.ID

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
	return table
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "Update",
			ExpResp: officebus.Office{
				ID:       sd.Offices[0].ID,
				Name:     "Updated Office",
				StreetID: sd.Offices[0].StreetID,
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.Office.Update(ctx, sd.Offices[0], officebus.UpdateOffice{
					Name:     dbtest.StringPointer("Updated Office"),
					StreetID: &sd.Offices[0].StreetID,
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(officebus.Office)
				if !exists {
					return fmt.Sprintf("got is not an asset type: %v", got)
				}

				expResp := exp.(officebus.Office)
				return cmp.Diff(expResp, gotResp)
			},
		},
	}
	return table
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "Delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.Office.Delete(ctx, sd.Offices[0])
				if err != nil {
					return err
				}
				return nil
			},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}
	return table
}
