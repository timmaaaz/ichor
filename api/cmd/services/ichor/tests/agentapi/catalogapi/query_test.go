package catalog_test

import (
	"fmt"
	"net/http"

	"github.com/timmaaaz/ichor/api/domain/http/agentapi/catalogapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func queryCatalog200(sd CatalogSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/agent/catalog",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &[]catalogapi.ConfigSurface{},
			ExpResp:    &[]catalogapi.ConfigSurface{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*[]catalogapi.ConfigSurface)
				if !ok {
					return "error casting catalog response"
				}

				if len(*gotResp) < 12 {
					return fmt.Sprintf("expected at least 12 config surfaces, got %d", len(*gotResp))
				}

				validCategories := map[string]bool{"ui": true, "workflow": true, "system": true}

				for _, surface := range *gotResp {
					if surface.Name == "" {
						return "config surface has empty name"
					}
					if !validCategories[surface.Category] {
						return fmt.Sprintf("config surface %q has invalid category %q", surface.Name, surface.Category)
					}
					if surface.Endpoints.List == "" && surface.Endpoints.Get == "" &&
						surface.Endpoints.Create == "" && surface.Endpoints.Update == "" &&
						surface.Endpoints.Delete == "" {
						return fmt.Sprintf("config surface %q has no endpoints", surface.Name)
					}
				}

				return ""
			},
		},
	}
}

func queryCatalog401(sd CatalogSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/v1/agent/catalog",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}
}
