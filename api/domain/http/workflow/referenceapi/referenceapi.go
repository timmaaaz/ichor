// Package referenceapi provides HTTP layer endpoints for workflow reference data.
//
// These endpoints return read-only lookups for:
//   - Trigger types (on_create, on_update, on_delete, scheduled)
//   - Entity types (tables, views that can be monitored)
//   - Entities (specific monitored tables/views)
//   - Action types (create_alert, send_email, etc.) with config schemas
//
// All endpoints require authentication but no special permissions beyond basic access.
// These endpoints are critical for the visual workflow editor UI, providing dropdown
// options and configuration schemas for building automation rules.
//
// Schema Alignment: Action type schemas MUST match TypeScript types in src/types/workflow.ts.
// See GetActionTypes() for the exported function used in schema alignment tests.
package referenceapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	log         *logger.Logger
	workflowBus *workflow.Business
}

func newAPI(cfg Config) *api {
	return &api{
		log:         cfg.Log,
		workflowBus: cfg.WorkflowBus,
	}
}

// queryTriggerTypes handles GET /v1/workflow/trigger-types
func (a *api) queryTriggerTypes(ctx context.Context, r *http.Request) web.Encoder {
	types, err := a.workflowBus.QueryTriggerTypes(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "query trigger types: %s", err)
	}
	return toTriggerTypes(types)
}

// queryEntityTypes handles GET /v1/workflow/entity-types
func (a *api) queryEntityTypes(ctx context.Context, r *http.Request) web.Encoder {
	types, err := a.workflowBus.QueryEntityTypes(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "query entity types: %s", err)
	}
	return toEntityTypes(types)
}

// queryEntities handles GET /v1/workflow/entities
// Optional query parameter: entity_type_id (UUID string) - filters entities by entity type
func (a *api) queryEntities(ctx context.Context, r *http.Request) web.Encoder {
	// Optional filter by entity type
	entityTypeIDStr := r.URL.Query().Get("entity_type_id")

	var entities []workflow.Entity
	var err error

	if entityTypeIDStr != "" {
		entityTypeID, parseErr := uuid.Parse(entityTypeIDStr)
		if parseErr != nil {
			return errs.New(errs.InvalidArgument, parseErr)
		}
		entities, err = a.workflowBus.QueryEntitiesByType(ctx, entityTypeID)
	} else {
		entities, err = a.workflowBus.QueryEntities(ctx)
	}

	if err != nil {
		return errs.Newf(errs.Internal, "query entities: %s", err)
	}
	return toEntities(entities)
}

// queryActionTypes handles GET /v1/workflow/action-types
func (a *api) queryActionTypes(ctx context.Context, r *http.Request) web.Encoder {
	return ActionTypes(GetActionTypes())
}

// queryActionTypeSchema handles GET /v1/workflow/action-types/{type}/schema
func (a *api) queryActionTypeSchema(ctx context.Context, r *http.Request) web.Encoder {
	actionType := web.Param(r, "type")

	schema, found := getActionTypeSchema(actionType)
	if !found {
		return errs.Newf(errs.NotFound, "action type %q not found", actionType)
	}

	return schema
}
