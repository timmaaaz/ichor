// Package configschemaapi provides HTTP endpoints for config JSONB schema discovery.
// These endpoints return JSON schemas describing the structure of opaque JSONB
// columns in table configs and page content layout.
package configschemaapi

import (
	"context"
	"net/http"

	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct{}

func newAPI() *api {
	return &api{}
}

// queryTableConfigSchema handles GET /v1/config/schemas/table-config
func (a *api) queryTableConfigSchema(_ context.Context, _ *http.Request) web.Encoder {
	return SchemaInfo{
		Name:        "Table Config",
		Description: "JSON schema for the config.table_configs.config JSONB column",
		Schema:      configSchemas["table_config"],
	}
}

// queryLayoutSchema handles GET /v1/config/schemas/layout
func (a *api) queryLayoutSchema(_ context.Context, _ *http.Request) web.Encoder {
	return SchemaInfo{
		Name:        "Layout Config",
		Description: "JSON schema for the config.page_content.layout JSONB column",
		Schema:      configSchemas["layout"],
	}
}

// queryContentTypes handles GET /v1/config/schemas/content-types
func (a *api) queryContentTypes(_ context.Context, _ *http.Request) web.Encoder {
	return ContentTypes(contentTypes)
}
