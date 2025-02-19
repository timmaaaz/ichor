package brandapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"

	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the brand domain.
type App struct {
	brandbus *brandbus.Business
	auth     *auth.Auth
}

// NewApp constructs a brand app API for use.
func NewApp(brandbus *brandbus.Business) *App {
	return &App{
		brandbus: brandbus,
	}
}

// NewAppWithAuth constructs a brand app API for use with auth support.
func NewAppWithAuth(brandbus *brandbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:     ath,
		brandbus: brandbus,
	}
}

// Create adds a new brand to the system.
func (a *App) Create(ctx context.Context, app NewBrand) (Brand, error) {
	nb, err := toBusNewBrand(app)
	if err != nil {
		return Brand{}, errs.New(errs.InvalidArgument, err)
	}

	brand, err := a.brandbus.Create(ctx, nb)
	if err != nil {
		if errors.Is(err, brandbus.ErrUniqueEntry) {
			return Brand{}, errs.New(errs.AlreadyExists, brandbus.ErrUniqueEntry)
		}
		if errors.Is(err, brandbus.ErrForeignKeyViolation) {
			return Brand{}, errs.New(errs.Aborted, brandbus.ErrForeignKeyViolation)
		}
		return Brand{}, errs.Newf(errs.Internal, "create: brand[%+v]: %s", brand, err)
	}

	return ToAppBrand(brand), err
}

// Update updates an existing brand.
func (a *App) Update(ctx context.Context, app UpdateBrand, id uuid.UUID) (Brand, error) {
	ub, err := toBusUpdateBrand(app)
	if err != nil {
		return Brand{}, errs.New(errs.InvalidArgument, err)
	}

	st, err := a.brandbus.QueryByID(ctx, id)
	if err != nil {
		return Brand{}, errs.New(errs.NotFound, brandbus.ErrNotFound)
	}

	brand, err := a.brandbus.Update(ctx, st, ub)
	if err != nil {
		if errors.Is(err, brandbus.ErrForeignKeyViolation) {
			return Brand{}, errs.New(errs.Aborted, brandbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, brandbus.ErrNotFound) {
			return Brand{}, errs.New(errs.NotFound, err)
		}
		return Brand{}, errs.Newf(errs.Internal, "update: brand[%+v]: %s", brand, err)
	}

	return ToAppBrand(brand), nil
}

// Delete removes an existing brand.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	brand, err := a.brandbus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, brandbus.ErrNotFound)
	}

	err = a.brandbus.Delete(ctx, brand)
	if err != nil {
		return errs.Newf(errs.Internal, "delete: brand[%+v]: %s", brand, err)
	}

	return nil
}

// Query returns a list of brands based on the filter, order and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Brand], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Brand]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Brand]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Brand]{}, errs.NewFieldsError("orderby", err)
	}

	brands, err := a.brandbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[Brand]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.brandbus.Count(ctx, filter)
	if err != nil {
		return query.Result[Brand]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppBrands(brands), total, page), nil
}

// QueryByID retrieves a single brand by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (Brand, error) {
	brand, err := a.brandbus.QueryByID(ctx, id)
	if err != nil {
		return Brand{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppBrand(brand), nil
}
