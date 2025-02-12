package userbus_test

import (
	"context"
	"fmt"
	"net/mail"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	"github.com/timmaaaz/ichor/business/domain/officebus"
	"github.com/timmaaaz/ichor/business/domain/titlebus"
	"github.com/timmaaaz/ichor/business/domain/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
	"golang.org/x/crypto/bcrypt"
)

func Test_User(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_User")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	unitest.Run(t, create(db.BusDomain), "create")
	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

// =============================================================================

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 2, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu1 := unitest.User{
		User: usrs[0],
	}

	tu2 := unitest.User{
		User: usrs[1],
	}

	// -------------------------------------------------------------------------

	/*
		usrs, err = userbus.TestSeedUsersWithNoFKs(ctx, 2, userbus.Roles.User, busDomain.User)
		if err != nil {
			return unitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
		}

			tu3 := unitest.User{
				User: usrs[0],
			}

			tu4 := unitest.User{
				User: usrs[1],
			}
	*/

	// -------------------------------------------------------------------------

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("querying regions : %w", err)
	}

	ids := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		ids = append(ids, r.ID)
	}

	cities, err := citybus.TestSeedCities(ctx, 10, ids, busDomain.City)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding cities : %w", err)
	}

	cityIDs := make([]uuid.UUID, 0, len(ids))
	for _, c := range cities {
		cityIDs = append(cityIDs, c.ID)
	}

	streets, err := streetbus.TestSeedStreets(ctx, 7, cityIDs, busDomain.Street)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding streets : %w", err)
	}

	streetIDs := make([]uuid.UUID, 0, len(streets))
	for _, s := range streets {
		streetIDs = append(streetIDs, s.ID)
	}

	offices, err := officebus.TestSeedOffices(ctx, 5, streetIDs, busDomain.Office)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding offices : %w", err)
	}

	officeIDs := make([]uuid.UUID, 0, len(offices))
	for _, o := range offices {
		officeIDs = append(officeIDs, o.ID)
	}

	titles, err := titlebus.TestSeedTitles(ctx, 5, busDomain.Title)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding titles : %w", err)
	}

	titleIDs := make([]uuid.UUID, 0, len(titles))
	for _, t := range titles {
		titleIDs = append(titleIDs, t.ID)
	}

	requestor, err := userbus.TestSeedUsersWithNoFKs(ctx, 3, userbus.Roles.User, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("test user : %w", err)
	}

	requestorIDs := make([]uuid.UUID, 0, len(requestor))
	for _, req := range requestor {
		requestorIDs = append(requestorIDs, req.ID)
	}

	users, err := userbus.TestSeedUsers(ctx, 5, userbus.Roles.User, requestorIDs, titleIDs, officeIDs, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	// -------------------------------------------------------------------------

	sd := unitest.SeedData{
		Users: []unitest.User{
			{User: users[0]}, {User: users[1]}, {User: users[2]},
			{User: users[3]}, {User: users[4]},
			{User: requestor[0]}, {User: requestor[1]}, {User: requestor[2]}},
		Admins: []unitest.User{tu1, tu2},
	}

	return sd, nil
}

// =============================================================================

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	usrs := make([]userbus.User, 0, len(sd.Admins)+len(sd.Users))

	for _, adm := range sd.Admins {
		usrs = append(usrs, adm.User)
	}

	for _, usr := range sd.Users {
		usrs = append(usrs, usr.User)
	}

	sort.Slice(usrs, func(i, j int) bool {
		return usrs[i].FirstName.String() > usrs[j].FirstName.String()
	})

	table := []unitest.Table{
		{
			Name:    "all",
			ExpResp: usrs,
			ExcFunc: func(ctx context.Context) any {
				filter := userbus.QueryFilter{
					Username: dbtest.UserNamePointer("Username"),
				}

				resp, err := busDomain.User.Query(ctx, filter, order.NewBy(userbus.OrderByFirstName, order.DESC), page.MustParse("1", "10"))
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.([]userbus.User)
				if !exists {
					return "error occurred"
				}

				expResp := exp.([]userbus.User)

				for i := range gotResp {
					if gotResp[i].DateCreated.Format(time.RFC3339) == expResp[i].DateCreated.Format(time.RFC3339) {
						expResp[i].DateCreated = gotResp[i].DateCreated
					}

					if gotResp[i].DateUpdated.Format(time.RFC3339) == expResp[i].DateUpdated.Format(time.RFC3339) {
						expResp[i].DateUpdated = gotResp[i].DateUpdated
					}
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
		{
			Name:    "byid",
			ExpResp: sd.Users[0].User,
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.User.QueryByID(ctx, sd.Users[0].ID)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(userbus.User)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(userbus.User)

				if gotResp.DateCreated.Format(time.RFC3339) == expResp.DateCreated.Format(time.RFC3339) {
					expResp.DateCreated = gotResp.DateCreated
				}

				if gotResp.DateUpdated.Format(time.RFC3339) == expResp.DateUpdated.Format(time.RFC3339) {
					expResp.DateUpdated = gotResp.DateUpdated
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func create(busDomain dbtest.BusDomain) []unitest.Table {
	email, _ := mail.ParseAddress("jake@superiortech.io")

	table := []unitest.Table{
		{
			Name: "basic",
			ExpResp: userbus.User{
				Username:           userbus.MustParseName("jtimmer"),
				FirstName:          userbus.MustParseName("Jake"),
				LastName:           userbus.MustParseName("Timmer"),
				Email:              *email,
				Roles:              []userbus.Role{userbus.Roles.Admin},
				SystemRoles:        []userbus.Role{userbus.Roles.Admin},
				Enabled:            true,
				UserApprovalStatus: uuid.MustParse("89173300-3f4e-4606-872c-f34914bbee19"),
			},
			ExcFunc: func(ctx context.Context) any {
				nu := userbus.NewUser{
					Username:    userbus.MustParseName("jtimmer"),
					FirstName:   userbus.MustParseName("Jake"),
					LastName:    userbus.MustParseName("Timmer"),
					Email:       *email,
					Roles:       []userbus.Role{userbus.Roles.Admin},
					SystemRoles: []userbus.Role{userbus.Roles.Admin},
					Enabled:     true,
					Password:    "123",
				}

				resp, err := busDomain.User.Create(ctx, nu)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(userbus.User)
				if !exists {
					return "error occurred"
				}

				if err := bcrypt.CompareHashAndPassword(gotResp.PasswordHash, []byte("123")); err != nil {
					return err.Error()
				}

				expResp := exp.(userbus.User)

				expResp.ID = gotResp.ID
				expResp.PasswordHash = gotResp.PasswordHash
				expResp.DateCreated = gotResp.DateCreated
				expResp.DateUpdated = gotResp.DateUpdated
				expResp.DateRequested = gotResp.DateRequested

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	email, _ := mail.ParseAddress("ben@superiortech.io")

	table := []unitest.Table{
		{
			Name: "basic",
			ExpResp: userbus.User{
				ID:          sd.Users[0].ID,
				FirstName:   userbus.MustParseName("Ben"),
				Email:       *email,
				Roles:       []userbus.Role{userbus.Roles.Admin},
				Enabled:     true,
				DateCreated: sd.Users[0].DateCreated,
			},
			ExcFunc: func(ctx context.Context) any {
				uu := userbus.UpdateUser{
					FirstName: dbtest.UserNamePointer("Ben"),
					Email:     email,
					Roles:     []userbus.Role{userbus.Roles.Admin},
					Password:  dbtest.StringPointer("1234"),
				}

				resp, err := busDomain.User.Update(ctx, sd.Users[0].User, uu)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(userbus.User)
				if !exists {
					return "error occurred"
				}

				if err := bcrypt.CompareHashAndPassword(gotResp.PasswordHash, []byte("1234")); err != nil {
					return err.Error()
				}

				expResp := sd.Users[0].User
				expResp.FirstName = userbus.MustParseName("Ben")
				expResp.Email = *email
				expResp.Roles = []userbus.Role{userbus.Roles.Admin}
				expResp.PasswordHash = gotResp.PasswordHash
				expResp.DateUpdated = gotResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "user",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				if err := busDomain.User.Delete(ctx, sd.Users[1].User); err != nil {
					return err
				}

				return nil
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "admin",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				if err := busDomain.User.Delete(ctx, sd.Admins[1].User); err != nil {
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
