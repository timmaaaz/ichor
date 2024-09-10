package countrybus_test

import (
	"context"
	"fmt"
	"testing"

	"bitbucket.org/superiortechnologies/ichor/business/domain/location/countrybus"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/dbtest"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/page"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/unitest"
	"github.com/google/go-cmp/cmp"
)

func Test_Country(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_Country")

	unitest.Run(t, countryQuery(db.BusDomain), "country-query")
}

func countryQuery(busdomain dbtest.BusDomain) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "Alpha2",
			ExpResp: []countrybus.Country{
				{Number: 1, Name: "Andorra", Alpha2: "AD", Alpha3: "AND"},
				{Number: 2, Name: "United Arab Emirates", Alpha2: "AE", Alpha3: "ARE"},
				{Number: 3, Name: "Afghanistan", Alpha2: "AF", Alpha3: "AFG"},
				{Number: 4, Name: "Antigua and Barbuda", Alpha2: "AG", Alpha3: "ATG"},
				{Number: 5, Name: "Anguilla", Alpha2: "AI", Alpha3: "AIA"},
			},
			ExcFunc: func(ctx context.Context) any {
				ctrys, err := busdomain.Country.Query(ctx, countrybus.QueryFilter{}, countrybus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return ctrys
			},
			CmpFunc: func(got any, exp any) string {
				fmt.Println(got, exp)

				gotResp := got.([]countrybus.Country)
				expResp := exp.([]countrybus.Country)

				// Ignore id's, they are generated in seeding
				for i := range expResp {
					expResp[i].ID = gotResp[i].ID
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
	return table
}
