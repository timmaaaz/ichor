// Package executionapi provides HTTP handlers for workflow execution history endpoints.
// These endpoints provide read-only access to execution records for debugging and auditing.
package executionapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// api holds dependencies for execution HTTP handlers.
type api struct {
	log         *logger.Logger
	workflowBus *workflow.Business
}

// newAPI creates a new execution API handler.
func newAPI(cfg Config) *api {
	return &api{
		log:         cfg.Log,
		workflowBus: cfg.WorkflowBus,
	}
}

// query handles GET /v1/workflow/executions
func (a *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp := parseQueryParams(r)

	filter, err := parseFilter(qp)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, workflow.DefaultExecutionOrderBy)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	executions, err := a.workflowBus.QueryExecutionsPaginated(ctx, filter, orderBy, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.workflowBus.CountExecutions(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toExecutionResponses(executions), total, pg)
}

// queryByID handles GET /v1/workflow/executions/{id}
func (a *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	id, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	execution, err := a.workflowBus.QueryExecutionByID(ctx, id)
	if err != nil {
		if errors.Is(err, workflow.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	return toExecutionDetail(execution)
}
