package user_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/users/userapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/users",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &userapp.NewUser{
				Username:        "agriffis",
				FirstName:       "Adam",
				LastName:        "Griffis",
				Email:           "adam@superiortech.io",
				Birthday:        "1980-01-01",
				Roles:           []string{"ADMIN"},
				SystemRoles:     []string{"ADMIN"},
				Password:        "123",
				PasswordConfirm: "123",
				Enabled:         true,
			},
			GotResp: &userapp.User{},
			ExpResp: &userapp.User{
				Username:    "agriffis",
				FirstName:   "Adam",
				LastName:    "Griffis",
				Birthday:    "1980-01-01",
				Email:       "adam@superiortech.io",
				Roles:       []string{"ADMIN"},
				SystemRoles: []string{"ADMIN"},
				Enabled:     true,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*userapp.User)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*userapp.User)

				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate
				expResp.RequestedBy = "00000000-0000-0000-0000-000000000000"
				expResp.ApprovedBy = "00000000-0000-0000-0000-000000000000"
				expResp.TitleID = "00000000-0000-0000-0000-000000000000"
				expResp.OfficeID = "00000000-0000-0000-0000-000000000000"
				expResp.WorkPhoneID = "00000000-0000-0000-0000-000000000000"
				expResp.CellPhoneID = "00000000-0000-0000-0000-000000000000"
				expResp.PasswordHash = gotResp.PasswordHash
				expResp.Birthday = "1980-01-01T00:00:00Z"
				expResp.DateHired = "0001-01-01T00:00:00Z"
				expResp.DateRequested = gotResp.DateRequested
				expResp.DateApproved = "0001-01-01T00:00:00Z"

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func create400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "missing-input",
			URL:        "/v1/users",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input:      &userapp.NewUser{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"username\",\"error\":\"username is a required field\"},{\"field\":\"firstName\",\"error\":\"firstName is a required field\"},{\"field\":\"lastName\",\"error\":\"lastName is a required field\"},{\"field\":\"email\",\"error\":\"email is a required field\"},{\"field\":\"birthday\",\"error\":\"birthday is a required field\"},{\"field\":\"roles\",\"error\":\"roles is a required field\"},{\"field\":\"systemRoles\",\"error\":\"systemRoles is a required field\"},{\"field\":\"password\",\"error\":\"password is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-role",
			URL:        "/v1/users",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &userapp.NewUser{
				Username:        "agriffis",
				FirstName:       "Adam",
				LastName:        "Griffis",
				Email:           "adam@superiortech.io",
				Birthday:        "1980-01-01",
				Roles:           []string{"SUPER"},
				SystemRoles:     []string{"ADMIN"},
				Password:        "123",
				PasswordConfirm: "123",
				Enabled:         true,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "parse: invalid role \"SUPER\""),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-name",
			URL:        "/v1/users",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &userapp.NewUser{
				Username:        "ag",
				FirstName:       "Adam",
				LastName:        "Griffis",
				Email:           "adam@superiortech.io",
				Birthday:        "1980-01-01",
				Roles:           []string{"ADMIN"},
				SystemRoles:     []string{"ADMIN"},
				Password:        "123",
				PasswordConfirm: "123",
				Enabled:         true,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "parse: invalid name \"ag\""),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func create401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        "/v1/users",
			Token:      "&nbsp;",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "badtoken",
			URL:        "/v1/users",
			Token:      sd.Admins[0].Token[:10],
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "badsig",
			URL:        "/v1/users",
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "wronguser",
			URL:        "/v1/users",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: users"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
