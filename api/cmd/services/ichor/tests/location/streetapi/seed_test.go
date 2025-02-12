package street_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/location/cityapp"
	"github.com/timmaaaz/ichor/app/domain/location/streetapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	"github.com/timmaaaz/ichor/business/domain/location/countrybus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}
	tu1 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
	}
	usrs, err = userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}
	tu2 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
	}

	allCountries, err := busDomain.Country.Query(ctx, countrybus.QueryFilter{}, order.NewBy(countrybus.OrderByNumber, order.ASC), page.MustParse("1", "99"))
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying countries : %w", err)
	}

	allRegions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "99"))
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying regions : %w", err)
	}

	var regionIDs []uuid.UUID

	for _, region := range allRegions {
		regionIDs = append(regionIDs, region.ID)
	}

	cities, err := citybus.TestSeedCities(ctx, 50, regionIDs, busDomain.City)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding cities : %w", err)
	}
	appCities := cityapp.ToAppCities(cities)

	cityIDs := make([]uuid.UUID, 0, len(cities))
	for _, city := range cities {
		cityIDs = append(cityIDs, city.ID)
	}

	streets, err := streetbus.TestSeedStreets(ctx, 50, cityIDs, busDomain.Street)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding streets : %w", err)
	}
	appStreets := streetapp.ToAppStreets(streets)

	sd := apitest.SeedData{
		Users:     []apitest.User{tu1},
		Admins:    []apitest.User{tu2},
		Countries: allCountries,
		Regions:   allRegions,
		Cities:    appCities,
		Streets:   appStreets,
	}

	return sd, nil
}
