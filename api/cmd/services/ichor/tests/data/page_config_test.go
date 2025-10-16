package data_test

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/dataapp"
)

func pageConfigCreate200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/data/page",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: dataapp.NewPageConfig{
				Name:      "Test Dashboard",
				UserID:    "", // Empty string for default configs (will be NULL in DB)
				IsDefault: "true",
			},
			GotResp: &dataapp.PageConfig{},
			ExpResp: &dataapp.PageConfig{
				Name:      "Test Dashboard",
				UserID:    "", // Default configs have no user_id
				IsDefault: "true",
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*dataapp.PageConfig)
				if !exists {
					return "could not convert got to *dataapp.PageConfig"
				}
				expResp, exists := exp.(*dataapp.PageConfig)
				if !exists {
					return "could not convert exp to *dataapp.PageConfig"
				}

				// Copy over the generated ID field from got to exp
				expResp.ID = gotResp.ID

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func pageConfigUpdate200(sd apitest.SeedData) []apitest.Table {
	name := "Updated Dashboard Name"
	isDefault := "true"

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/data/page/%s", sd.PageConfigs[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: dataapp.UpdatePageConfig{
				Name:      &name,
				IsDefault: &isDefault,
			},
			GotResp: &dataapp.PageConfig{},
			ExpResp: &dataapp.PageConfig{
				ID:        sd.PageConfigs[0].ID.String(),
				Name:      name,
				UserID:    "", // When updated to default, user_id is cleared
				IsDefault: "true",
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*dataapp.PageConfig)
				if !exists {
					return "could not convert got to *dataapp.PageConfig"
				}
				expResp, exists := exp.(*dataapp.PageConfig)
				if !exists {
					return "could not convert exp to *dataapp.PageConfig"
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func pageConfigDelete200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/data/page/%s", sd.PageConfigs[2].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusNoContent,
		},
	}
}

func pageConfigQueryByName200(sd apitest.SeedData) []apitest.Table {
	urlEncoded := url.QueryEscape("Dashboard Home")

	var expTabs = []dataapp.PageTabConfig{}

	for _, p := range sd.PageTabConfigs {
		if p.PageConfigID == sd.PageConfigs[0].ID {
			expTabs = append(expTabs, dataapp.ToAppPageTabConfig(p))
		}
	}

	// This test verifies that QueryPageByName returns the default page configuration.
	// Only one default page config is allowed per page name (enforced by database constraint).
	// This serves as the fallback configuration for all users.
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/data/page/name/%s", urlEncoded),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &dataapp.FullPageConfig{},
			ExpResp: &dataapp.FullPageConfig{
				PageConfig: dataapp.PageConfig{
					ID:        sd.PageConfigs[0].ID.String(),
					Name:      "Dashboard Home",
					UserID:    "", // Default configs have no user_id
					IsDefault: "true",
				},
				PageTabs: expTabs,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*dataapp.FullPageConfig)
				if !exists {
					return "could not convert got to *dataapp.FullPageConfig"
				}
				expResp, exists := exp.(*dataapp.FullPageConfig)
				if !exists {
					return "could not convert exp to *dataapp.FullPageConfig"
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func pageConfigQueryByID200(sd apitest.SeedData) []apitest.Table {

	var expTabs = []dataapp.PageTabConfig{}

	for _, p := range sd.PageTabConfigs {
		if p.PageConfigID == sd.PageConfigs[1].ID {
			expTabs = append(expTabs, dataapp.ToAppPageTabConfig(p))
		}
	}

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/data/page/id/%s", sd.PageConfigs[1].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &dataapp.FullPageConfig{},
			ExpResp: &dataapp.FullPageConfig{
				PageConfig: dataapp.PageConfig{
					ID:        sd.PageConfigs[1].ID.String(),
					Name:      "Inventory Overview",
					UserID:    sd.PageConfigs[1].UserID.String(),
					IsDefault: "false",
				},
				PageTabs: expTabs,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*dataapp.FullPageConfig)
				if !exists {
					return "could not convert got to *dataapp.FullPageConfig"
				}
				expResp, exists := exp.(*dataapp.FullPageConfig)
				if !exists {
					return "could not convert exp to *dataapp.FullPageConfig"
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func pageConfigQueryByNameAndUserID200(sd apitest.SeedData) []apitest.Table {
	urlEncodedName := url.QueryEscape("Inventory Overview")

	var expTabs = []dataapp.PageTabConfig{}

	for _, p := range sd.PageTabConfigs {
		if p.PageConfigID == sd.PageConfigs[1].ID {
			expTabs = append(expTabs, dataapp.ToAppPageTabConfig(p))
		}
	}

	// This test verifies that QueryPageByNameAndUserID returns a specific user's page configuration.
	// This allows each user to have their own customized version of a page (e.g., Jake's version of the orders page).
	// Multiple users can have configs with the same page name, but only one per user+name combination.
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/data/page/name/%s/user/%s", urlEncodedName, sd.Admins[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &dataapp.FullPageConfig{},
			ExpResp: &dataapp.FullPageConfig{
				PageConfig: dataapp.PageConfig{
					ID:        sd.PageConfigs[1].ID.String(),
					Name:      "Inventory Overview",
					UserID:    sd.PageConfigs[1].UserID.String(),
					IsDefault: "false",
				},
				PageTabs: expTabs,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*dataapp.FullPageConfig)
				if !exists {
					return "could not convert got to *dataapp.FullPageConfig"
				}
				expResp, exists := exp.(*dataapp.FullPageConfig)
				if !exists {
					return "could not convert exp to *dataapp.FullPageConfig"
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}
