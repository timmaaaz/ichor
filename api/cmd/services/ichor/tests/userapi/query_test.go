package user_test

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/users/userapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
)

func query200(sd apitest.SeedData) []apitest.Table {
	usrs := make([]userbus.User, 0, len(sd.Admins)+len(sd.Users))

	for _, adm := range sd.Admins {
		adm.User.PasswordHash = nil // Endpoint wont return this
		usrs = append(usrs, adm.User)
	}

	for _, usr := range sd.Users {
		usr.User.PasswordHash = nil // Endpoint wont return this
		usrs = append(usrs, usr.User)
	}

	sort.Slice(usrs, func(i, j int) bool {
		return usrs[i].ID.String() <= usrs[j].ID.String()
	})

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/users?page=1&rows=10&orderBy=user_id,ASC&username=Username",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[userapp.User]{},
			ExpResp: &query.Result[userapp.User]{
				Page:        1,
				RowsPerPage: 10,
				Total:       len(usrs),
				Items:       toAppUsers(usrs),
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func queryByID200(sd apitest.SeedData) []apitest.Table {

	sd.Users[0].PasswordHash = nil // Endpoint wont return this

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/users/%s", sd.Users[0].ID),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &userapp.User{},
			ExpResp:    toAppUserPtr(sd.Users[0].User),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
