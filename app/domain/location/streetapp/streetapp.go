package streetapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the street domain.
type App struct {
	streetBus *streetbus.Business
	auth      *auth.Auth
}

// NewApp constructs a street app API for use.
func NewApp(streetBus *streetbus.Business) *App {
	return &App{
		streetBus: streetBus,
	}
}

// NewAppWithAuth constructs a street app API for use with auth support.
func NewAppWithAuth(streetBus *streetbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:      ath,
		streetBus: streetBus,
	}
}

// Create adds a new street to the system.
func (a *App) Create(ctx context.Context, app NewStreet) (Street, error) {
	ns, err := toBusNewStreet(app)
	if err != nil {
		return Street{}, errs.New(errs.InvalidArgument, err)
	}

	street, err := a.streetBus.Create(ctx, ns)
	if err != nil {
		// No unique constraint in the db yet.
		// if errors.Is(err, streetbus.ErrUniqueEntry) {
		// 	return Street{}, errs.New(errs.Aborted, streetbus.ErrUniqueEntry)
		// }
		return Street{}, errs.Newf(errs.Internal, "create: street[%+v]: %s", street, err)
	}

	return ToAppStreet(street), nil
}

// Update updates an existing street.
func (a *App) Update(ctx context.Context, app UpdateStreet, id uuid.UUID) (Street, error) {
	us, err := toBusUpdateStreet(app)
	if err != nil {
		return Street{}, errs.New(errs.InvalidArgument, err)
	}

	st, err := a.streetBus.QueryByID(ctx, id)
	if err != nil {
		return Street{}, errs.New(errs.NotFound, streetbus.ErrNotFound)
	}

	street, err := a.streetBus.Update(ctx, st, us)
	if err != nil {
		if errors.Is(err, streetbus.ErrNotFound) {
			return Street{}, errs.New(errs.NotFound, err)
		}
		return Street{}, errs.Newf(errs.Internal, "update: street[%+v]: %s", street, err)
	}

	return ToAppStreet(street), nil
}

// Delete removes an existing street.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	st, err := a.streetBus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, streetbus.ErrNotFound)
	}

	err = a.streetBus.Delete(ctx, st)
	if err != nil {
		return errs.Newf(errs.Internal, "delete: street[%+v]: %s", st, err)
	}

	return nil
}

// Query returns a list of streets based on the filter, order and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Street], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Street]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Street]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Street]{}, errs.NewFieldsError("orderby", err)
	}

	streets, err := a.streetBus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[Street]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.streetBus.Count(ctx, filter)
	if err != nil {
		return query.Result[Street]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppStreets(streets), total, page), nil
}

// QueryByID retrieves a single street by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (Street, error) {
	street, err := a.streetBus.QueryByID(ctx, id)
	if err != nil {
		return Street{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppStreet(street), nil
}
