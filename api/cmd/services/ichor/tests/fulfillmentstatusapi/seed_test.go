package fulfillmentstatus_test

import (
	"context"
	"fmt"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/fulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/fulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/userbus"
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

	fulfillmentstatus1, err := busDomain.FulfillmentStatus.Query(ctx, fulfillmentstatusbus.QueryFilter{}, order.NewBy(fulfillmentstatusbus.OrderByName, order.ASC), page.MustParse("1", "2"))
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying approval statuses : %w", err)
	}

	fulfillmentstatus2, err := busDomain.FulfillmentStatus.Query(ctx, fulfillmentstatusbus.QueryFilter{}, order.NewBy(fulfillmentstatusbus.OrderByName, order.ASC), page.MustParse("2", "2"))
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying approval statuses : %w", err)
	}

	fulfillments := append(fulfillmentstatus1, fulfillmentstatus2...)

	appFulfillments := fulfillmentstatusapp.ToAppFulfillmentStatuses(fulfillments)

	sd := apitest.SeedData{
		Users:               []apitest.User{tu1},
		Admins:              []apitest.User{tu2},
		FulfillmentStatuses: appFulfillments,
	}

	return sd, nil

}
