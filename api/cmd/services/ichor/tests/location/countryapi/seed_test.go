package country_test

import (
	"context"
	"fmt"

	"bitbucket.org/superiortechnologies/ichor/api/sdk/http/apitest"
	"bitbucket.org/superiortechnologies/ichor/app/sdk/auth"
	"bitbucket.org/superiortechnologies/ichor/business/domain/location/countrybus"
	"bitbucket.org/superiortechnologies/ichor/business/domain/userbus"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/dbtest"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/order"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/page"
)

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	usrs, err := userbus.TestSeedUsers(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu1 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
	}

	countries1, err := busDomain.Country.Query(ctx, countrybus.QueryFilter{}, order.NewBy(countrybus.OrderByNumber, order.ASC), page.MustParse("1", "99"))
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying countries : %w", err)
	}

	countries2, err := busDomain.Country.Query(ctx, countrybus.QueryFilter{}, order.NewBy(countrybus.OrderByNumber, order.ASC), page.MustParse("2", "99"))
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying countries : %w", err)
	}

	countries3, err := busDomain.Country.Query(ctx, countrybus.QueryFilter{}, order.NewBy(countrybus.OrderByNumber, order.ASC), page.MustParse("3", "99"))
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying countries : %w", err)
	}

	// combine all countries
	countries := append(countries1, countries2...)
	countries = append(countries, countries3...)

	sd := apitest.SeedData{
		Users:     []apitest.User{tu1},
		Countries: countries,
	}

	return sd, nil
}
