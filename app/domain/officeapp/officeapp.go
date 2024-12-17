package officeapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/officebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the office domain.
type App struct {
	officeBus *officebus.Business
	auth      *auth.Auth
}

// NewApp constructs a office app API for use.
func NewApp(officeBus *officebus.Business) *App {
	return &App{
		officeBus: officeBus,
	}
}

// NewAppWithAuth constructs a office app API for use with auth support.
func NewAppWithAuth(officeBus *officebus.Business, ath *auth.Auth) *App {
	return &App{
		auth:      ath,
		officeBus: officeBus,
	}
}

// Create adds a new office to the system.
func (a *App) Create(ctx context.Context, app NewOffice) (Office, error) {
	na, err := toBusNewOffice(app)
	if err != nil {
		return Office{}, errs.New(errs.InvalidArgument, err)
	}

	ass, err := a.officeBus.Create(ctx, na)
	if err != nil {
		if errors.Is(err, officebus.ErrUniqueEntry) {
			return Office{}, errs.New(errs.Aborted, officebus.ErrUniqueEntry)
		}
		return Office{}, errs.Newf(errs.Internal, "create: ass[%+v]: %s", ass, err)
	}

	return ToAppOffice(ass), err
}

// Update updates an existing office.
func (a *App) Update(ctx context.Context, app UpdateOffice, id uuid.UUID) (Office, error) {
	uo, err := toBusUpdateOffice(app)
	if err != nil {
		return Office{}, errs.New(errs.InvalidArgument, err)
	}

	oID, err := a.officeBus.QueryByID(ctx, id)
	if err != nil {
		return Office{}, errs.New(errs.NotFound, officebus.ErrNotFound)
	}

	office, err := a.officeBus.Update(ctx, oID, uo)
	if err != nil {
		if errors.Is(err, officebus.ErrNotFound) {
			return Office{}, errs.New(errs.NotFound, err)
		}
		return Office{}, errs.Newf(errs.Internal, "update: office[%+v]: %s", office, err)
	}

	return ToAppOffice(office), nil
}

// Delete removes an existing office.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	oID, err := a.officeBus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, officebus.ErrNotFound)
	}

	err = a.officeBus.Delete(ctx, oID)
	if err != nil {
		return errs.Newf(errs.Internal, "delete: office[%+v]: %s", oID, err)
	}

	return nil
}

// Query returns a list of offices based on the filter, order and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Office], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Office]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Office]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Office]{}, errs.NewFieldsError("orderby", err)
	}

	offices, err := a.officeBus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[Office]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.officeBus.Count(ctx, filter)
	if err != nil {
		return query.Result[Office]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppOffices(offices), total, page), nil
}

// QueryByID retrieves a single office by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (Office, error) {
	office, err := a.officeBus.QueryByID(ctx, id)
	if err != nil {
		return Office{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppOffice(office), nil
}
