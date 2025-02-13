package userapprovalcomment_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/users/status/commentapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/users/status/commentbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
	ctx := context.Background()

	busDomain := db.BusDomain

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 2, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu1 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
	}

	tu3 := apitest.User{
		User:  usrs[1],
		Token: apitest.Token(db.BusDomain.User, ath, usrs[1].Email.Address),
	}

	usrs, err = userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}
	tu2 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
	}

	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 10, userbus.Roles.User, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	userIDs := make([]uuid.UUID, 0, len(users))
	for _, u := range users {
		userIDs = append(userIDs, u.ID)
	}

	comments, err := commentbus.TestSeedUserApprovalComment(ctx, 10, userIDs[:5], userIDs[5:], busDomain.UserApprovalComment)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding approval comments : %w", err)
	}

	sd := apitest.SeedData{
		Users:                []apitest.User{tu1, tu3},
		Admins:               []apitest.User{tu2},
		UserApprovalComments: commentapp.ToAppUserApprovalComments(comments),
	}

	return sd, nil

}
