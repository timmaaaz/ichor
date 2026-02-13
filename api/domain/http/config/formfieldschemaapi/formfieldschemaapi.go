// Package formfieldschemaapi provides HTTP endpoints for form field type discovery.
// These endpoints return field type metadata and JSON schemas that describe
// the configuration options available for each form field type.
package formfieldschemaapi

import (
	"context"
	"net/http"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct{}

func newAPI() *api {
	return &api{}
}

// queryFieldTypes handles GET /v1/config/form-field-types
func (a *api) queryFieldTypes(_ context.Context, _ *http.Request) web.Encoder {
	return FieldTypes(GetFieldTypes())
}

// queryFieldTypeSchema handles GET /v1/config/form-field-types/{type}/schema
func (a *api) queryFieldTypeSchema(_ context.Context, r *http.Request) web.Encoder {
	fieldType := web.Param(r, "type")

	info, found := getFieldTypeSchema(fieldType)
	if !found {
		return errs.Newf(errs.NotFound, "field type %q not found", fieldType)
	}

	return info
}
