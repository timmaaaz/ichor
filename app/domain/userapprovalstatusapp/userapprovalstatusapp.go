package userapprovalstatusapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/userapprovalstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the approval status domain.
type App struct {
	userapprovalstatusbus *userapprovalstatusbus.Business
	auth                  *auth.Auth
}

// NewApp constructs an approval status app API for use.
func NewApp(approvalstatusBus *userapprovalstatusbus.Business) *App {
	return &App{
		userapprovalstatusbus: approvalstatusBus,
	}
}

// NewAppWithAuth constructs an approval status app API for use with auth support.
func NewAppWithAuth(approvalstatusBus *userapprovalstatusbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:                  ath,
		userapprovalstatusbus: approvalstatusBus,
	}
}

// Create adds a new approval status to the system
func (a *App) Create(ctx context.Context, app NewUserApprovalStatus) (UserApprovalStatus, error) {
	nas, err := toBusNewUserApprovalStatus(app)
	if err != nil {
		return UserApprovalStatus{}, errs.New(errs.InvalidArgument, err)
	}

	as, err := a.userapprovalstatusbus.Create(ctx, nas)
	if err != nil {
		return UserApprovalStatus{}, err
	}

	return ToAppUserApprovalStatus(as), nil
}

// Update updates an existing approval status
func (a *App) Update(ctx context.Context, app UpdateUserApprovalStatus, id uuid.UUID) (UserApprovalStatus, error) {
	uas, err := toBusUpdateUserApprovalStatus(app)
	if err != nil {
		return UserApprovalStatus{}, errs.New(errs.InvalidArgument, err)
	}

	as, err := a.userapprovalstatusbus.QueryByID(ctx, id)
	if err != nil {
		return UserApprovalStatus{}, errs.New(errs.NotFound, userapprovalstatusbus.ErrNotFound)
	}

	updated, err := a.userapprovalstatusbus.Update(ctx, as, uas)
	if err != nil {
		if errors.Is(err, userapprovalstatusbus.ErrNotFound) {
			return UserApprovalStatus{}, errs.New(errs.NotFound, err)
		}
		return UserApprovalStatus{}, errs.Newf(errs.Internal, "update: user approvalStatus[%+v]: %s", updated, err)
	}

	return ToAppUserApprovalStatus(updated), nil
}

// Delete removes an existing approval status
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	as, err := a.userapprovalstatusbus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, userapprovalstatusbus.ErrNotFound)
	}

	err = a.userapprovalstatusbus.Delete(ctx, as)
	if err != nil {
		return errs.Newf(errs.Internal, "delete user approval status[%+v]: %s", as, err)
	}

	return nil
}

// Query returns a list of approval statuses based on the filter, order and page
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[UserApprovalStatus], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[UserApprovalStatus]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[UserApprovalStatus]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[UserApprovalStatus]{}, errs.NewFieldsError("orderby", err)
	}

	as, err := a.userapprovalstatusbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[UserApprovalStatus]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.userapprovalstatusbus.Count(ctx, filter)
	if err != nil {
		return query.Result[UserApprovalStatus]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppUserApprovalStatuses(as), total, page), nil
}

// QueryByID retrieves the approval status by ID
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (UserApprovalStatus, error) {
	as, err := a.userapprovalstatusbus.QueryByID(ctx, id)
	if err != nil {
		return UserApprovalStatus{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppUserApprovalStatus(as), nil
}
