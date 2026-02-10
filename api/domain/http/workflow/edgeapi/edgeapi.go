package edgeapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// api holds dependencies for edge HTTP handlers.
type api struct {
	log         *logger.Logger
	workflowBus *workflow.Business
}

// newAPI creates a new edge API handler.
func newAPI(log *logger.Logger, workflowBus *workflow.Business) *api {
	return &api{
		log:         log,
		workflowBus: workflowBus,
	}
}

// create handles POST /v1/workflow/rules/{ruleID}/edges
func (a *api) create(ctx context.Context, r *http.Request) web.Encoder {
	ruleID, err := uuid.Parse(web.Param(r, "ruleID"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	var req CreateEdgeRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// Validate request
	if validationErr := req.Validate(); validationErr != nil {
		return errs.New(errs.FailedPrecondition, validationErr)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Verify rule exists
	_, err = a.workflowBus.QueryRuleByID(ctx, ruleID)
	if err != nil {
		if errors.Is(err, workflow.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "query rule: %s", err)
	}

	// Verify target action exists and belongs to this rule
	targetAction, err := a.workflowBus.QueryActionByID(ctx, req.TargetActionID)
	if err != nil {
		if errors.Is(err, workflow.ErrNotFound) {
			return errs.Newf(errs.NotFound, "target action not found: %s", req.TargetActionID)
		}
		return errs.Newf(errs.Internal, "query target action: %s", err)
	}
	if targetAction.AutomationRuleID != ruleID {
		return errs.New(errs.FailedPrecondition, errors.New("target action does not belong to this rule"))
	}

	// If source action is specified, verify it exists and belongs to this rule
	if req.SourceActionID != nil {
		sourceAction, err := a.workflowBus.QueryActionByID(ctx, *req.SourceActionID)
		if err != nil {
			if errors.Is(err, workflow.ErrNotFound) {
				return errs.Newf(errs.NotFound, "source action not found: %s", *req.SourceActionID)
			}
			return errs.Newf(errs.Internal, "query source action: %s", err)
		}
		if sourceAction.AutomationRuleID != ruleID {
			return errs.New(errs.FailedPrecondition, errors.New("source action does not belong to this rule"))
		}
	}

	// Create the edge
	newEdge := toNewActionEdge(ruleID, req)
	edge, err := a.workflowBus.CreateActionEdge(ctx, newEdge)
	if err != nil {
		return errs.Newf(errs.Internal, "create edge: %s", err)
	}

	a.log.Info(ctx, "edge created", "edge_id", edge.ID, "rule_id", ruleID, "created_by", userID)

	return toEdgeResponse(edge)
}

// query handles GET /v1/workflow/rules/{ruleID}/edges
func (a *api) query(ctx context.Context, r *http.Request) web.Encoder {
	ruleID, err := uuid.Parse(web.Param(r, "ruleID"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// Verify rule exists
	_, err = a.workflowBus.QueryRuleByID(ctx, ruleID)
	if err != nil {
		if errors.Is(err, workflow.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "query rule: %s", err)
	}

	edges, err := a.workflowBus.QueryEdgesByRuleID(ctx, ruleID)
	if err != nil {
		return errs.Newf(errs.Internal, "query edges: %s", err)
	}

	return toEdgeResponses(edges)
}

// queryByID handles GET /v1/workflow/rules/{ruleID}/edges/{edgeID}
func (a *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	ruleID, err := uuid.Parse(web.Param(r, "ruleID"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	edgeID, err := uuid.Parse(web.Param(r, "edgeID"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// Verify rule exists
	_, err = a.workflowBus.QueryRuleByID(ctx, ruleID)
	if err != nil {
		if errors.Is(err, workflow.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "query rule: %s", err)
	}

	edge, err := a.workflowBus.QueryEdgeByID(ctx, edgeID)
	if err != nil {
		if errors.Is(err, workflow.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "query edge: %s", err)
	}

	// Verify edge belongs to the specified rule
	if edge.RuleID != ruleID {
		return errs.New(errs.NotFound, errors.New("edge does not belong to this rule"))
	}

	return toEdgeResponse(edge)
}

// delete handles DELETE /v1/workflow/rules/{ruleID}/edges/{edgeID}
func (a *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	ruleID, err := uuid.Parse(web.Param(r, "ruleID"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	edgeID, err := uuid.Parse(web.Param(r, "edgeID"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Verify rule exists
	_, err = a.workflowBus.QueryRuleByID(ctx, ruleID)
	if err != nil {
		if errors.Is(err, workflow.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "query rule: %s", err)
	}

	// Verify edge exists and belongs to this rule
	edge, err := a.workflowBus.QueryEdgeByID(ctx, edgeID)
	if err != nil {
		if errors.Is(err, workflow.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "query edge: %s", err)
	}

	if edge.RuleID != ruleID {
		return errs.New(errs.NotFound, errors.New("edge does not belong to this rule"))
	}

	// Delete the edge
	if err := a.workflowBus.DeleteActionEdge(ctx, edgeID); err != nil {
		return errs.Newf(errs.Internal, "delete edge: %s", err)
	}

	a.log.Info(ctx, "edge deleted", "edge_id", edgeID, "rule_id", ruleID, "deleted_by", userID)

	return nil
}

// deleteAll handles DELETE /v1/workflow/rules/{ruleID}/edges
// Deletes all edges for a rule (useful for resetting action graph)
func (a *api) deleteAll(ctx context.Context, r *http.Request) web.Encoder {
	ruleID, err := uuid.Parse(web.Param(r, "ruleID"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Verify rule exists
	_, err = a.workflowBus.QueryRuleByID(ctx, ruleID)
	if err != nil {
		if errors.Is(err, workflow.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "query rule: %s", err)
	}

	// Delete all edges for this rule
	if err := a.workflowBus.DeleteEdgesByRuleID(ctx, ruleID); err != nil {
		return errs.Newf(errs.Internal, "delete edges: %s", err)
	}

	a.log.Info(ctx, "all edges deleted for rule", "rule_id", ruleID, "deleted_by", userID)

	return nil
}
