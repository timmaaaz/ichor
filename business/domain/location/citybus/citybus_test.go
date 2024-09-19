package citybus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_City(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_City")

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

// =============================================================================

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

	return unitest.SeedData{
		Cities:  ctys,
		Regions: regions,
	}, nil
}

// =============================================================================

func query(busdomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "query",
			ExpResp: []citybus.City{
				{ID: sd.Cities[0].ID, RegionID: sd.Cities[0].RegionID, Name: sd.Cities[0].Name},
				{ID: sd.Cities[1].ID, RegionID: sd.Cities[1].RegionID, Name: sd.Cities[1].Name},
				{ID: sd.Cities[2].ID, RegionID: sd.Cities[2].RegionID, Name: sd.Cities[2].Name},
				{ID: sd.Cities[3].ID, RegionID: sd.Cities[3].RegionID, Name: sd.Cities[3].Name},
				{ID: sd.Cities[4].ID, RegionID: sd.Cities[4].RegionID, Name: sd.Cities[4].Name},
			},
			ExcFunc: func(ctx context.Context) any {
				ctrys, err := busdomain.City.Query(ctx, citybus.QueryFilter{}, order.NewBy(citybus.OrderByName, order.ASC), page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return ctrys
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "create",
			ExpResp: citybus.City{
				RegionID: sd.Regions[0].ID,
				Name:     "Test City",
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.City.Create(ctx, citybus.NewCity{
					RegionID: sd.Regions[0].ID,
					Name:     "Test City",
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(citybus.City)
				if !exists {
					return fmt.Sprintf("got is not a city: %v", got)
				}
				expResp := exp.(citybus.City)
				expResp.ID = gotResp.ID

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
	return table
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "update",
			ExpResp: citybus.City{
				ID:       sd.Cities[0].ID,
				RegionID: sd.Cities[1].RegionID,
				Name:     "Updated City",
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.City.Update(ctx, sd.Cities[0], citybus.UpdateCity{
					RegionID: &sd.Cities[1].RegionID,
					Name:     dbtest.StringPointer("Updated City"),
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(citybus.City)
				if !exists {
					return fmt.Sprintf("got is not a city: %v", got)
				}
				expResp := exp.(citybus.City)

				return cmp.Diff(gotResp, expResp)
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
				err := busDomain.City.Delete(ctx, sd.Cities[0])
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
