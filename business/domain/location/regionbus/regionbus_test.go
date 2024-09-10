package regionbus_test

import (
	"context"
	"fmt"
	"testing"

	"bitbucket.org/superiortechnologies/ichor/business/domain/location/regionbus"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/dbtest"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/page"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/unitest"
	"github.com/google/go-cmp/cmp"
)

func Test_region(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_region")

	sd, err := getSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, regionQuery(db.BusDomain, sd), "query")
}

func getSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("querying regions : %w", err)
	}

	sd := unitest.SeedData{
		Regions: regions,
	}
	return sd, nil
}

func regionQuery(busdomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "Alpha2",
			ExpResp: []regionbus.Region{
				{ID: sd.Regions[0].ID, CountryID: sd.Regions[0].CountryID, Name: sd.Regions[0].Name, Code: sd.Regions[0].Code},
				{ID: sd.Regions[1].ID, CountryID: sd.Regions[1].CountryID, Name: sd.Regions[1].Name, Code: sd.Regions[1].Code},
				{ID: sd.Regions[2].ID, CountryID: sd.Regions[2].CountryID, Name: sd.Regions[2].Name, Code: sd.Regions[2].Code},
				{ID: sd.Regions[3].ID, CountryID: sd.Regions[3].CountryID, Name: sd.Regions[3].Name, Code: sd.Regions[3].Code},
				{ID: sd.Regions[4].ID, CountryID: sd.Regions[4].CountryID, Name: sd.Regions[4].Name, Code: sd.Regions[4].Code},
			},
			ExcFunc: func(ctx context.Context) any {
				ctrys, err := busdomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
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
