package approvalstatusapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the approval status domain.
type App struct {
	approvalstatusbus *approvalstatusbus.Business
	auth              *auth.Auth
}

// NewApp constructs an approval status app API for use.
func NewApp(approvalstatusBus *approvalstatusbus.Business) *App {
	return &App{
		approvalstatusbus: approvalstatusBus,
	}
}

// NewAppWithAuth constructs an approval status app API for use with auth support.
func NewAppWithAuth(approvalstatusBus *approvalstatusbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:              ath,
		approvalstatusbus: approvalstatusBus,
	}
}

// Create adds a new approval status to the system
func (a *App) Create(ctx context.Context, app NewApprovalStatus) (ApprovalStatus, error) {
	nas, err := toBusNewApprovalStatus(app)
	if err != nil {
		return ApprovalStatus{}, errs.New(errs.InvalidArgument, err)
	}

	as, err := a.approvalstatusbus.Create(ctx, nas)
	if err != nil {
		return ApprovalStatus{}, err
	}

	return ToAppApprovalStatus(as), nil
}

// Update updates an existing approval status
func (a *App) Update(ctx context.Context, app UpdateApprovalStatus, id uuid.UUID) (ApprovalStatus, error) {
	uas, err := toBusUpdateApprovalStatus(app)
	if err != nil {
		return ApprovalStatus{}, errs.New(errs.InvalidArgument, err)
	}

	as, err := a.approvalstatusbus.QueryByID(ctx, id)
	if err != nil {
		return ApprovalStatus{}, errs.New(errs.NotFound, approvalstatusbus.ErrNotFound)
	}

	updated, err := a.approvalstatusbus.Update(ctx, as, uas)
	if err != nil {
		if errors.Is(err, approvalstatusbus.ErrNotFound) {
			return ApprovalStatus{}, errs.New(errs.NotFound, err)
		}
		return ApprovalStatus{}, errs.Newf(errs.Internal, "update: approvalStatus[%+v]: %s", updated, err)
	}

	return ToAppApprovalStatus(updated), nil
}

// Delete removes an existing approval status
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	as, err := a.approvalstatusbus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, approvalstatusbus.ErrNotFound)
	}

	err = a.approvalstatusbus.Delete(ctx, as)
	if err != nil {
		return errs.Newf(errs.Internal, "delete approval status[%+v]: %s", as, err)
	}

	return nil
}

// Query returns a list of approval statuses based on the filter, order and page
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[ApprovalStatus], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[ApprovalStatus]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[ApprovalStatus]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[ApprovalStatus]{}, errs.NewFieldsError("orderby", err)
	}

	as, err := a.approvalstatusbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[ApprovalStatus]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.approvalstatusbus.Count(ctx, filter)
	if err != nil {
		return query.Result[ApprovalStatus]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppApprovalStatuses(as), total, page), nil
}

// QueryByID retrieves the approval status by ID
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (ApprovalStatus, error) {
	as, err := a.approvalstatusbus.QueryByID(ctx, id)
	if err != nil {
		return ApprovalStatus{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppApprovalStatus(as), nil
}
