package data_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/dataapp"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func pageTabConfigCreate200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/data/page/tab",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: dataapp.NewPageTabConfig{
				Label:        "New Tab",
				PageConfigID: sd.PageConfigs[0].ID.String(),
				ConfigID:     sd.SimpleTableConfig.ID.String(),
				IsDefault:    "false",
				TabOrder:     "3",
			},
			GotResp: &dataapp.PageTabConfig{},
			ExpResp: &dataapp.PageTabConfig{
				Label:        "New Tab",
				PageConfigID: sd.PageConfigs[0].ID.String(),
				ConfigID:     sd.SimpleTableConfig.ID.String(),
				IsDefault:    "false",
				TabOrder:     "3",
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*dataapp.PageTabConfig)
				if !exists {
					return "could not convert got to *dataapp.PageTabConfig"
				}
				expResp, exists := exp.(*dataapp.PageTabConfig)
				if !exists {
					return "could not convert exp to *dataapp.PageTabConfig"
				}

				// Copy over the generated ID field from got to exp
				expResp.ID = gotResp.ID

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func pageTabConfigUpdate200(sd apitest.SeedData) []apitest.Table {
	label := "Updated Tab Label"
	isDefault := true
	tabOrder := 5

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/data/page/tab/%s", sd.PageTabConfigs[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: dataapp.UpdatePageTabConfig{
				Label:     &label,
				IsDefault: dbtest.StringPointer(fmt.Sprintf("%t", isDefault)),
				TabOrder:  dbtest.StringPointer(fmt.Sprintf("%d", tabOrder)),
			},
			GotResp: &dataapp.PageTabConfig{},
			ExpResp: &dataapp.PageTabConfig{
				ID:           sd.PageTabConfigs[0].ID.String(),
				Label:        label,
				PageConfigID: sd.PageTabConfigs[0].PageConfigID.String(),
				ConfigID:     sd.PageTabConfigs[0].ConfigID.String(),
				IsDefault:    "true",
				TabOrder:     "5",
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*dataapp.PageTabConfig)
				if !exists {
					return "could not convert got to *dataapp.PageTabConfig"
				}
				expResp, exists := exp.(*dataapp.PageTabConfig)
				if !exists {
					return "could not convert exp to *dataapp.PageTabConfig"
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func pageTabConfigDelete200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/data/page/tab/%s", sd.PageTabConfigs[5].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusNoContent,
		},
	}
}
