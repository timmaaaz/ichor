package productcostapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/finance/productcostbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the productCost domain.
type App struct {
	productcostbus *productcostbus.Business
	auth           *auth.Auth
}

// NewApp constructs a product cost app API for use.
func NewApp(productcostbus *productcostbus.Business) *App {
	return &App{
		productcostbus: productcostbus,
	}
}

// NewAppWithAuth constructs a product cost app API for use with auth support.
func NewAppWithAuth(productcostbus *productcostbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:           ath,
		productcostbus: productcostbus,
	}
}

// Create adds a new product cost to the system.
func (a *App) Create(ctx context.Context, app NewProductCost) (ProductCost, error) {
	nb, err := toBusNewProductCost(app)
	if err != nil {
		return ProductCost{}, errs.New(errs.InvalidArgument, err)
	}

	productCost, err := a.productcostbus.Create(ctx, nb)
	if err != nil {
		if errors.Is(err, productcostbus.ErrUniqueEntry) {
			return ProductCost{}, errs.New(errs.AlreadyExists, productcostbus.ErrUniqueEntry)
		}
		if errors.Is(err, productcostbus.ErrForeignKeyViolation) {
			return ProductCost{}, errs.New(errs.Aborted, productcostbus.ErrForeignKeyViolation)
		}
		return ProductCost{}, errs.Newf(errs.Internal, "create: product cost[%+v]: %s", productCost, err)
	}

	return ToAppProductCost(productCost), err
}

// Update updates an existing product cost.
func (a *App) Update(ctx context.Context, app UpdateProductCost, id uuid.UUID) (ProductCost, error) {
	upc, err := toBusUpdateProductCost(app)
	if err != nil {
		return ProductCost{}, errs.New(errs.InvalidArgument, err)
	}

	st, err := a.productcostbus.QueryByID(ctx, id)
	if err != nil {
		return ProductCost{}, errs.New(errs.NotFound, productcostbus.ErrNotFound)
	}

	productCost, err := a.productcostbus.Update(ctx, st, upc)
	if err != nil {
		if errors.Is(err, productcostbus.ErrForeignKeyViolation) {
			return ProductCost{}, errs.New(errs.Aborted, productcostbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, productcostbus.ErrNotFound) {
			return ProductCost{}, errs.New(errs.NotFound, err)
		}
		return ProductCost{}, errs.Newf(errs.Internal, "update: productCost[%+v]: %s", productCost, err)
	}

	return ToAppProductCost(productCost), nil
}

// Delete removes an existing productCost.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	productCost, err := a.productcostbus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, productcostbus.ErrNotFound)
	}

	err = a.productcostbus.Delete(ctx, productCost)
	if err != nil {
		return errs.Newf(errs.Internal, "delete: productCost[%+v]: %s", productCost, err)
	}

	return nil
}

// Query returns a list of productCosts based on the filter, order and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[ProductCost], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[ProductCost]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[ProductCost]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[ProductCost]{}, errs.NewFieldsError("orderby", err)
	}

	productCosts, err := a.productcostbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[ProductCost]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.productcostbus.Count(ctx, filter)
	if err != nil {
		return query.Result[ProductCost]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppProductCosts(productCosts), total, page), nil
}

// QueryByID retrieves a single productCost by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (ProductCost, error) {
	productCost, err := a.productcostbus.QueryByID(ctx, id)
	if err != nil {
		return ProductCost{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppProductCost(productCost), nil
}
