// Package workflowsaveapi provides HTTP handlers for transactional workflow save operations.
package workflowsaveapi

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/workflow/workflowsaveapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/web"
)

// api holds dependencies for workflow save HTTP handlers.
type api struct {
	app *workflowsaveapp.App
}

// newAPI creates a new workflow save API handler.
func newAPI(app *workflowsaveapp.App) *api {
	return &api{app: app}
}

// save handles PUT /v1/workflow/rules/{id}/full
// Updates an existing workflow atomically (rule + actions + edges).
func (a *api) save(ctx context.Context, r *http.Request) web.Encoder {
	ruleID, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	var req workflowsaveapp.SaveWorkflowRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	resp, err := a.app.SaveWorkflow(ctx, ruleID, req)
	if err != nil {
		return errs.NewError(err)
	}

	return resp
}

// create handles POST /v1/workflow/rules/full
// Creates a new workflow atomically (rule + actions + edges).
func (a *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var req workflowsaveapp.SaveWorkflowRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	resp, err := a.app.CreateWorkflow(ctx, userID, req)
	if err != nil {
		return errs.NewError(err)
	}

	return resp
}

// =============================================================================
// Action Type Discovery API
// =============================================================================

// discoveryAPI provides action type discovery endpoints.
type discoveryAPI struct {
	registry *workflow.ActionRegistry
}

func newDiscoveryAPI(registry *workflow.ActionRegistry) *discoveryAPI {
	return &discoveryAPI{registry: registry}
}

// actionTypeInfo is the response shape for a single action type.
type actionTypeInfo struct {
	Type             string                `json:"type"`
	Description      string                `json:"description"`
	SupportsManual   bool                  `json:"supports_manual"`
	IsAsync          bool                  `json:"is_async"`
	Outputs          []workflow.OutputPort  `json:"outputs"`
}

// actionTypeList implements web.Encoder for the discovery response.
type actionTypeList []actionTypeInfo

func (l actionTypeList) Encode() ([]byte, string, error) {
	data, err := json.Marshal(l)
	return data, "application/json", err
}

// queryActionTypes handles GET /v1/workflow/action-types
func (d *discoveryAPI) queryActionTypes(_ context.Context, _ *http.Request) web.Encoder {
	types := d.registry.GetAll()
	sort.Strings(types)

	result := make(actionTypeList, 0, len(types))
	for _, actionType := range types {
		handler, _ := d.registry.Get(actionType)
		result = append(result, actionTypeInfo{
			Type:           actionType,
			Description:    handler.GetDescription(),
			SupportsManual: handler.SupportsManualExecution(),
			IsAsync:        handler.IsAsync(),
			Outputs:        d.registry.GetOutputPorts(actionType),
		})
	}

	return result
}
