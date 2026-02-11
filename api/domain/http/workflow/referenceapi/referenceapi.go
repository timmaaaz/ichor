// Package referenceapi provides HTTP layer endpoints for workflow reference data.
//
// These endpoints return read-only lookups for:
//   - Trigger types (on_create, on_update, on_delete, scheduled)
//   - Entity types (tables, views that can be monitored)
//   - Entities (specific monitored tables/views)
//   - Action types (all registered types with config schemas and output ports)
//
// All endpoints require authentication but no special permissions beyond basic access.
// These endpoints are critical for the visual workflow editor UI, providing dropdown
// options and configuration schemas for building automation rules.
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
	log            *logger.Logger
	workflowBus    *workflow.Business
	actionRegistry *workflow.ActionRegistry
}

func newAPI(cfg Config) *api {
	return &api{
		log:            cfg.Log,
		workflowBus:    cfg.WorkflowBus,
		actionRegistry: cfg.ActionRegistry,
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
// Returns all 17 registered action types with config schemas and output ports.
func (a *api) queryActionTypes(ctx context.Context, r *http.Request) web.Encoder {
	return ActionTypes(GetActionTypes(a.actionRegistry))
}

// queryTemplates handles GET /v1/workflow/templates
func (a *api) queryTemplates(ctx context.Context, r *http.Request) web.Encoder {
	templates, err := a.workflowBus.QueryAllTemplates(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "query templates: %s", err)
	}
	return toActionTemplates(templates)
}

// queryActiveTemplates handles GET /v1/workflow/templates/active
func (a *api) queryActiveTemplates(ctx context.Context, r *http.Request) web.Encoder {
	templates, err := a.workflowBus.QueryActiveTemplates(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "query active templates: %s", err)
	}
	return toActionTemplates(templates)
}

// queryActionTypeSchema handles GET /v1/workflow/action-types/{type}/schema
func (a *api) queryActionTypeSchema(ctx context.Context, r *http.Request) web.Encoder {
	actionType := web.Param(r, "type")

	schema, found := getActionTypeSchema(actionType, a.actionRegistry)
	if !found {
		return errs.Newf(errs.NotFound, "action type %q not found", actionType)
	}

	return schema
}
