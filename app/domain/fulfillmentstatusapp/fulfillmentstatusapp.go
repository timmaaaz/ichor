package fulfillmentstatusapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/fulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the fulfillment status domain.
type App struct {
	fulfillmentstatusbus *fulfillmentstatusbus.Business
	auth                 *auth.Auth
}

// NewApp constructs a fulfillment status app API for use.
func NewApp(fulfillmentstatusbus *fulfillmentstatusbus.Business) *App {
	return &App{
		fulfillmentstatusbus: fulfillmentstatusbus,
	}
}

// NewAppWithAuth constructs a fulfillment status app API for use with auth support.
func NewAppWithAuth(fulfillmentstatusbus *fulfillmentstatusbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:                 ath,
		fulfillmentstatusbus: fulfillmentstatusbus,
	}
}

// Create adds a new fulfillment status to the system
func (a *App) Create(ctx context.Context, app NewFulfillmentStatus) (FulfillmentStatus, error) {
	nfs, err := toBusNewFulfillmentStatus(app)
	if err != nil {
		return FulfillmentStatus{}, err
	}

	fs, err := a.fulfillmentstatusbus.Create(ctx, nfs)
	if err != nil {
		return FulfillmentStatus{}, err
	}

	return ToAppFulfillmentStatus(fs), nil
}

// Update updates an existing approval status
func (a *App) Update(ctx context.Context, app UpdateFulfillmentStatus, id uuid.UUID) (FulfillmentStatus, error) {
	uas, err := toBusUpdateFulfillmentStatus(app)
	if err != nil {
		return FulfillmentStatus{}, errs.New(errs.InvalidArgument, err)
	}

	as, err := a.fulfillmentstatusbus.QueryByID(ctx, id)
	if err != nil {
		return FulfillmentStatus{}, errs.New(errs.NotFound, fulfillmentstatusbus.ErrNotFound)
	}

	updated, err := a.fulfillmentstatusbus.Update(ctx, as, uas)
	if err != nil {
		if errors.Is(err, fulfillmentstatusbus.ErrNotFound) {
			return FulfillmentStatus{}, errs.New(errs.NotFound, err)
		}
		return FulfillmentStatus{}, errs.Newf(errs.Internal, "update: approvalStatus[%+v]: %s", updated, err)
	}

	return ToAppFulfillmentStatus(updated), nil
}

// Delete removes an existing approval status
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	as, err := a.fulfillmentstatusbus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, fulfillmentstatusbus.ErrNotFound)
	}

	err = a.fulfillmentstatusbus.Delete(ctx, as)
	if err != nil {
		return errs.Newf(errs.Internal, "delete approval status[%+v]: %s", as, err)
	}

	return nil
}

// Query returns a list of approval statuses based on the filter, order and page
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[FulfillmentStatus], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[FulfillmentStatus]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[FulfillmentStatus]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[FulfillmentStatus]{}, errs.NewFieldsError("orderby", err)
	}

	as, err := a.fulfillmentstatusbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[FulfillmentStatus]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.fulfillmentstatusbus.Count(ctx, filter)
	if err != nil {
		return query.Result[FulfillmentStatus]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppFulfillmentStatuses(as), total, page), nil
}

// QueryByID retrieves the approval status by ID
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (FulfillmentStatus, error) {
	as, err := a.fulfillmentstatusbus.QueryByID(ctx, id)
	if err != nil {
		return FulfillmentStatus{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppFulfillmentStatus(as), nil
}
