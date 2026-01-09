package introspectionapi

import (
	"context"
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/introspectionapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	introspectionApp *introspectionapp.App
}

func newAPI(introspectionApp *introspectionapp.App) *api {
	return &api{
		introspectionApp: introspectionApp,
	}
}

func (api *api) querySchemas(ctx context.Context, r *http.Request) web.Encoder {
	schemas, err := api.introspectionApp.QuerySchemas(ctx)
	if err != nil {
		return errs.NewError(err)
	}

	return schemas
}

func (api *api) queryTables(ctx context.Context, r *http.Request) web.Encoder {
	schema := web.Param(r, "schema")

	tables, err := api.introspectionApp.QueryTables(ctx, schema)
	if err != nil {
		return errs.NewError(err)
	}

	return tables
}

func (api *api) queryColumns(ctx context.Context, r *http.Request) web.Encoder {
	schema := web.Param(r, "schema")
	table := web.Param(r, "table")

	columns, err := api.introspectionApp.QueryColumns(ctx, schema, table)
	if err != nil {
		return errs.NewError(err)
	}

	return columns
}

func (api *api) queryRelationships(ctx context.Context, r *http.Request) web.Encoder {
	schema := web.Param(r, "schema")
	table := web.Param(r, "table")

	relationships, err := api.introspectionApp.QueryRelationships(ctx, schema, table)
	if err != nil {
		return errs.NewError(err)
	}

	return relationships
}

func (api *api) queryReferencingTables(ctx context.Context, r *http.Request) web.Encoder {
	schema := web.Param(r, "schema")
	table := web.Param(r, "table")

	tables, err := api.introspectionApp.QueryReferencingTables(ctx, schema, table)
	if err != nil {
		return errs.NewError(err)
	}

	return tables
}
