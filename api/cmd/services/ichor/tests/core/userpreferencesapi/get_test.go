package userpreferencesapi_test

import (
	"encoding/json"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/userpreferencesapp"
)

func get200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/users/" + sd.Admins[0].ID.String() + "/preferences/floor.font_scale",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &userpreferencesapp.UserPreference{},
			ExpResp: &userpreferencesapp.UserPreference{
				UserID: sd.Admins[0].ID.String(),
				Key:    "floor.font_scale",
				Value:  json.RawMessage(`"medium"`),
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

func getAll200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/users/" + sd.Admins[0].ID.String() + "/preferences",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &userpreferencesapp.UserPreferences{},
			ExpResp: &userpreferencesapp.UserPreferences{
				Items: []userpreferencesapp.UserPreference{
					{
						UserID: sd.Admins[0].ID.String(),
						Key:    "floor.font_scale",
						Value:  json.RawMessage(`"medium"`),
					},
				},
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*userpreferencesapp.UserPreferences)
				expResp := exp.(*userpreferencesapp.UserPreferences)

				for i := range expResp.Items {
					if i < len(gotResp.Items) {
						expResp.Items[i].UpdatedDate = gotResp.Items[i].UpdatedDate
					}
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}
