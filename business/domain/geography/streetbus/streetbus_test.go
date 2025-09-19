package streetbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_Street(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_Street")

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

	return unitest.SeedData{
		Cities:  ctys,
		Regions: regions,
		Streets: strs,
	}, nil
}

// =============================================================================

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "query",
			ExpResp: []streetbus.Street{
				{ID: sd.Streets[0].ID, CityID: sd.Streets[0].CityID, Line1: sd.Streets[0].Line1, Line2: sd.Streets[0].Line2, PostalCode: sd.Streets[0].PostalCode},
				{ID: sd.Streets[1].ID, CityID: sd.Streets[1].CityID, Line1: sd.Streets[1].Line1, Line2: sd.Streets[1].Line2, PostalCode: sd.Streets[1].PostalCode},
				{ID: sd.Streets[2].ID, CityID: sd.Streets[2].CityID, Line1: sd.Streets[2].Line1, Line2: sd.Streets[2].Line2, PostalCode: sd.Streets[2].PostalCode},
				{ID: sd.Streets[3].ID, CityID: sd.Streets[3].CityID, Line1: sd.Streets[3].Line1, Line2: sd.Streets[3].Line2, PostalCode: sd.Streets[3].PostalCode},
				{ID: sd.Streets[4].ID, CityID: sd.Streets[4].CityID, Line1: sd.Streets[4].Line1, Line2: sd.Streets[4].Line2, PostalCode: sd.Streets[4].PostalCode},
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.Street.Query(ctx, streetbus.QueryFilter{}, streetbus.DefaultOrderBy, page.MustParse("1", "5"))
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
			Name: "create",
			ExpResp: streetbus.Street{
				CityID:     sd.Cities[0].ID,
				Line1:      "Test Line 1",
				Line2:      "Test Line 2",
				PostalCode: "54321",
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.Street.Create(ctx, streetbus.NewStreet{
					CityID:     sd.Cities[0].ID,
					Line1:      "Test Line 1",
					Line2:      "Test Line 2",
					PostalCode: "54321",
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(streetbus.Street)
				if !exists {
					return fmt.Sprintf("got is not a street: %v", got)
				}
				expResp := exp.(streetbus.Street)

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
			Name: "update",
			ExpResp: streetbus.Street{
				ID:         sd.Streets[0].ID,
				CityID:     sd.Streets[0].CityID,
				Line1:      "Updated Line 1",
				Line2:      "Updated Line 2",
				PostalCode: "12345",
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.Street.Update(ctx, sd.Streets[0], streetbus.UpdateStreet{
					Line1:      dbtest.StringPointer("Updated Line 1"),
					Line2:      dbtest.StringPointer("Updated Line 2"),
					PostalCode: dbtest.StringPointer("12345"),
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(streetbus.Street)
				if !exists {
					return fmt.Sprintf("got is not a street: %v", got)
				}
				expResp := exp.(streetbus.Street)

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
	return table
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.Street.Delete(ctx, sd.Streets[0])
				if err != nil {
					return err
				}
				return nil
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
