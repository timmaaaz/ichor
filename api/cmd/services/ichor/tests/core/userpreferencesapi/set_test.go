package userpreferencesapi_test

import (
	"encoding/json"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/userpreferencesapp"
)

func set200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "create-new",
			URL:        "/v1/users/" + sd.Admins[0].ID.String() + "/preferences/floor.theme",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPut,
			Input: &userpreferencesapp.NewUserPreference{
				Value: json.RawMessage(`"dark"`),
			},
			GotResp: &userpreferencesapp.UserPreference{},
			ExpResp: &userpreferencesapp.UserPreference{
				UserID: sd.Admins[0].ID.String(),
				Key:    "floor.theme",
				Value:  json.RawMessage(`"dark"`),
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*userpreferencesapp.UserPreference)
				expResp := exp.(*userpreferencesapp.UserPreference)

				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
		{
			Name:       "upsert-existing",
			URL:        "/v1/users/" + sd.Admins[0].ID.String() + "/preferences/floor.font_scale",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPut,
			Input: &userpreferencesapp.NewUserPreference{
				Value: json.RawMessage(`"large"`),
			},
			GotResp: &userpreferencesapp.UserPreference{},
			ExpResp: &userpreferencesapp.UserPreference{
				UserID: sd.Admins[0].ID.String(),
				Key:    "floor.font_scale",
				Value:  json.RawMessage(`"large"`),
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*userpreferencesapp.UserPreference)
				expResp := exp.(*userpreferencesapp.UserPreference)

				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}
