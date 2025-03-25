package supplierapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"

	"github.com/timmaaaz/ichor/business/domain/supplier/supplierbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the supplier domain.
type App struct {
	supplierbus *supplierbus.Business
	auth        *auth.Auth
}

// NewApp constructs a supplier app API for use.
func NewApp(supplierbus *supplierbus.Business) *App {
	return &App{
		supplierbus: supplierbus,
	}
}

// NewAppWithAuth constructs a supplier app API for use with auth support.
func NewAppWithAuth(supplierbus *supplierbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:        ath,
		supplierbus: supplierbus,
	}
}

// Create adds a new supplier to the system.
func (a *App) Create(ctx context.Context, app NewSupplier) (Supplier, error) {
	nb, err := toBusNewSupplier(app)
	if err != nil {
		return Supplier{}, errs.New(errs.InvalidArgument, err)
	}

	supplier, err := a.supplierbus.Create(ctx, nb)
	if err != nil {
		if errors.Is(err, supplierbus.ErrUniqueEntry) {
			return Supplier{}, errs.New(errs.AlreadyExists, err)
		}
		if errors.Is(err, supplierbus.ErrForeignKeyViolation) {
			return Supplier{}, errs.New(errs.Aborted, err)
		}
		return Supplier{}, errs.Newf(errs.Internal, "create: supplier[%+v]: %s", supplier, err)
	}

	return ToAppSupplier(supplier), err
}

// Update updates an existing supplier.
func (a *App) Update(ctx context.Context, app UpdateSupplier, id uuid.UUID) (Supplier, error) {
	upc, err := toBusUpdateSupplier(app)
	if err != nil {
		return Supplier{}, errs.New(errs.InvalidArgument, err)
	}

	st, err := a.supplierbus.QueryByID(ctx, id)
	if err != nil {
		return Supplier{}, errs.New(errs.NotFound, supplierbus.ErrNotFound)
	}

	supplier, err := a.supplierbus.Update(ctx, st, upc)
	if err != nil {
		if errors.Is(err, supplierbus.ErrForeignKeyViolation) {
			return Supplier{}, errs.New(errs.Aborted, supplierbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, supplierbus.ErrNotFound) {
			return Supplier{}, errs.New(errs.NotFound, err)
		}
		return Supplier{}, errs.Newf(errs.Internal, "update: supplier[%+v]: %s", supplier, err)
	}

	return ToAppSupplier(supplier), nil
}

// Delete removes an existing supplier.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	supplier, err := a.supplierbus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, supplierbus.ErrNotFound)
	}

	err = a.supplierbus.Delete(ctx, supplier)
	if err != nil {
		return errs.Newf(errs.Internal, "delete: supplier[%+v]: %s", supplier, err)
	}

	return nil
}

// Query returns a list of suppliers based on the filter, order and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Supplier], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Supplier]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Supplier]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Supplier]{}, errs.NewFieldsError("orderby", err)
	}

	suppliers, err := a.supplierbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[Supplier]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.supplierbus.Count(ctx, filter)
	if err != nil {
		return query.Result[Supplier]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppSuppliers(suppliers), total, page), nil
}

// QueryByID retrieves a single supplier by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (Supplier, error) {
	supplier, err := a.supplierbus.QueryByID(ctx, id)
	if err != nil {
		return Supplier{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppSupplier(supplier), nil
}
