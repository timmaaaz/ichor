// Package approvalapi provides HTTP handlers for workflow approval requests.
package approvalapi

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
	"github.com/timmaaaz/ichor/foundation/web"
)

// orderByFields maps API field names to database field names for ordering.
var orderByFields = map[string]string{
	"createdDate": approvalrequestbus.OrderByCreatedDate,
	"status":      approvalrequestbus.OrderByStatus,
}

type api struct {
	log            *logger.Logger
	approvalBus    *approvalrequestbus.Business
	userBus        *userbus.Business
	asyncCompleter *temporal.AsyncCompleter
	workflowQueue  *rabbitmq.WorkflowQueue
}

func newAPI(cfg Config) *api {
	return &api{
		log:            cfg.Log,
		approvalBus:    cfg.ApprovalBus,
		userBus:        cfg.UserBus,
		asyncCompleter: cfg.AsyncCompleter,
		workflowQueue:  cfg.WorkflowQueue,
	}
}

// query returns all approval requests (admin only).
func (a *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp := parseQueryParams(r)

	filter := parseFilter(qp)

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, approvalrequestbus.DefaultOrderBy)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	items, err := a.approvalBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.approvalBus.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	appItems := toAppApprovals(items)
	a.enrichApproversSlice(ctx, items, appItems)

	return query.NewResult(appItems, total, pg)
}

// queryMine returns approval requests where the authenticated user is an approver.
func (a *api) queryMine(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	qp := parseQueryParams(r)

	filter := parseFilter(qp)
	filter.ApproverID = &userID

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, approvalrequestbus.DefaultOrderBy)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	items, err := a.approvalBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.approvalBus.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	appItems := toAppApprovals(items)
	a.enrichApproversSlice(ctx, items, appItems)

	return query.NewResult(appItems, total, pg)
}

// queryByID returns a single approval request by ID.
func (a *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	id, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	_, err = mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	req, err := a.approvalBus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, approvalrequestbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "query by id: %s", err)
	}

	app := toAppApproval(req)
	a.enrichSingleApproval(ctx, req, &app)

	return app
}

// resolve handles the approval/rejection of a pending approval request.
// Authorization: user must be an approver or have ADMIN role in claims.
func (a *api) resolve(ctx context.Context, r *http.Request) web.Encoder {
	id, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	var req ResolveRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if req.Resolution != "approved" && req.Resolution != "rejected" {
		return errs.Newf(errs.InvalidArgument, "resolution must be 'approved' or 'rejected'")
	}

	// Authorization: check if user is approver or admin.
	isApprover, err := a.approvalBus.IsApprover(ctx, id, userID)
	if err != nil {
		return errs.Newf(errs.Internal, "check approver: %s", err)
	}

	if !isApprover {
		claims := mid.GetClaims(ctx)
		if !hasAdminRole(claims.Roles) {
			return errs.New(errs.PermissionDenied, approvalrequestbus.ErrNotApprover)
		}
	}

	// Resolve in DB (atomic, checks pending status).
	approval, err := a.approvalBus.Resolve(ctx, id, userID, req.Resolution, req.Reason)
	if err != nil {
		if errors.Is(err, approvalrequestbus.ErrAlreadyResolved) {
			return errs.Newf(errs.FailedPrecondition, "approval already resolved")
		}
		return errs.Newf(errs.Internal, "resolve: %s", err)
	}

	// Complete the Temporal activity if we have an async completer and a task token.
	if a.asyncCompleter != nil && approval.TaskToken != "" {
		taskToken, err := base64.StdEncoding.DecodeString(approval.TaskToken)
		if err != nil {
			a.log.Error(ctx, "failed to decode task token",
				"approval_id", id,
				"error", err)
		} else {
			output := temporal.ActionActivityOutput{
				ActionName: approval.ActionName,
				Result: map[string]any{
					"output":      req.Resolution,
					"approval_id": approval.ID.String(),
					"resolved_by": userID.String(),
					"reason":      req.Reason,
				},
				Success: true,
			}

			if err := a.asyncCompleter.Complete(ctx, taskToken, output); err != nil {
				a.log.Error(ctx, "failed to complete Temporal activity",
					"approval_id", id,
					"error", err)
			}
		}
	}

	a.publishApprovalResolved(ctx, approval, userID)

	return toAppApproval(approval)
}

// hasAdminRole checks if the ADMIN role is present in the claims roles.
func hasAdminRole(roles []string) bool {
	for _, r := range roles {
		if r == "ADMIN" {
			return true
		}
	}
	return false
}

// =============================================================================
// Approver enrichment
// =============================================================================

// enrichApproversSlice batch-enriches approver names for a list of approval requests.
func (a *api) enrichApproversSlice(ctx context.Context, busItems []approvalrequestbus.ApprovalRequest, appItems []Approval) {
	if a.userBus == nil || len(busItems) == 0 {
		return
	}

	// Collect unique approver + resolvedBy UUIDs.
	idSet := make(map[uuid.UUID]bool)
	for _, item := range busItems {
		for _, id := range item.Approvers {
			idSet[id] = true
		}
		if item.ResolvedBy != nil {
			idSet[*item.ResolvedBy] = true
		}
	}

	nameMap := a.resolveUserNames(ctx, idSet)

	for i, item := range busItems {
		details := make([]ApproverVM, 0, len(item.Approvers))
		for _, id := range item.Approvers {
			vm := ApproverVM{ID: id.String()}
			if name, ok := nameMap[id]; ok {
				vm.Name = name
			}
			details = append(details, vm)
		}
		appItems[i].ApproverDetails = details

		if item.ResolvedBy != nil {
			if name, ok := nameMap[*item.ResolvedBy]; ok {
				appItems[i].ResolvedByName = name
			}
		}
	}
}

// enrichSingleApproval enriches a single approval response with resolved names.
func (a *api) enrichSingleApproval(ctx context.Context, bus approvalrequestbus.ApprovalRequest, app *Approval) {
	if a.userBus == nil {
		return
	}

	idSet := make(map[uuid.UUID]bool)
	for _, id := range bus.Approvers {
		idSet[id] = true
	}
	if bus.ResolvedBy != nil {
		idSet[*bus.ResolvedBy] = true
	}

	nameMap := a.resolveUserNames(ctx, idSet)

	details := make([]ApproverVM, 0, len(bus.Approvers))
	for _, id := range bus.Approvers {
		vm := ApproverVM{ID: id.String()}
		if name, ok := nameMap[id]; ok {
			vm.Name = name
		}
		details = append(details, vm)
	}
	app.ApproverDetails = details

	if bus.ResolvedBy != nil {
		if name, ok := nameMap[*bus.ResolvedBy]; ok {
			app.ResolvedByName = name
		}
	}
}

// resolveUserNames fetches display names for a set of user UUIDs.
func (a *api) resolveUserNames(ctx context.Context, ids map[uuid.UUID]bool) map[uuid.UUID]string {
	nameMap := make(map[uuid.UUID]string, len(ids))
	for id := range ids {
		user, err := a.userBus.QueryByID(ctx, id)
		if err != nil {
			a.log.Error(ctx, "failed to resolve approver name", "user_id", id, "error", err)
			continue
		}
		nameMap[id] = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	}
	return nameMap
}

// publishApprovalResolved publishes a resolved approval event to RabbitMQ for WebSocket delivery.
func (a *api) publishApprovalResolved(ctx context.Context, approval approvalrequestbus.ApprovalRequest, resolvedBy uuid.UUID) {
	if a.workflowQueue == nil {
		return
	}

	payload := map[string]any{
		"approvalId": approval.ID.String(),
		"status":     approval.Status,
		"resolvedBy": resolvedBy.String(),
		"ruleId":     approval.RuleID.String(),
		"actionName": approval.ActionName,
	}
	if approval.ResolvedDate != nil {
		payload["resolvedDate"] = approval.ResolvedDate.Format(time.RFC3339)
	}

	msg := &rabbitmq.Message{
		Type:       "approval_resolved",
		EntityName: "workflow.approval_requests",
		EntityID:   approval.ID,
		UserID:     resolvedBy,
		Payload:    payload,
	}

	if err := a.workflowQueue.Publish(ctx, rabbitmq.QueueTypeAlert, msg); err != nil {
		a.log.Error(ctx, "failed to publish approval resolved event", "approval_id", approval.ID, "error", err)
	}
}

// parseQueryParams extracts query parameters from the request.
func parseQueryParams(r *http.Request) QueryParams {
	values := r.URL.Query()
	return QueryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("orderBy"),
		Status:  values.Get("status"),
	}
}

// parseFilter constructs a QueryFilter from query parameters.
func parseFilter(qp QueryParams) approvalrequestbus.QueryFilter {
	var filter approvalrequestbus.QueryFilter
	if qp.Status != "" {
		filter.Status = &qp.Status
	}
	return filter
}
