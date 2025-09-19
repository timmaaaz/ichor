package productcategoryapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"

	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the product category domain.
type App struct {
	productcategorybus *productcategorybus.Business
	auth               *auth.Auth
}

// NewApp constructs a product category app API for use.
func NewApp(productcategorybus *productcategorybus.Business) *App {
	return &App{
		productcategorybus: productcategorybus,
	}
}

// NewAppWithAuth constructs a product category app API for use with auth support.
func NewAppWithAuth(productcategorybus *productcategorybus.Business, ath *auth.Auth) *App {
	return &App{
		auth:               ath,
		productcategorybus: productcategorybus,
	}
}

// Create adds a new product category to the system.
func (a *App) Create(ctx context.Context, app NewProductCategory) (ProductCategory, error) {
	npc := toBusNewProductCategory(app)

	pc, err := a.productcategorybus.Create(ctx, npc)
	if err != nil {
		if errors.Is(err, productcategorybus.ErrUniqueEntry) {
			return ProductCategory{}, errs.New(errs.AlreadyExists, productcategorybus.ErrUniqueEntry)
		}
		return ProductCategory{}, errs.Newf(errs.Internal, "create: product category[%+v]: %s", pc, err)
	}

	return ToAppProductCategory(pc), err
}

// Update updates an existing product category.
func (a *App) Update(ctx context.Context, app UpdateProductCategory, id uuid.UUID) (ProductCategory, error) {
	upc := toBusUpdateProductCategory(app)

	st, err := a.productcategorybus.QueryByID(ctx, id)
	if err != nil {
		return ProductCategory{}, errs.New(errs.NotFound, productcategorybus.ErrNotFound)
	}

	pc, err := a.productcategorybus.Update(ctx, st, upc)
	if err != nil {
		if errors.Is(err, productcategorybus.ErrNotFound) {
			return ProductCategory{}, errs.New(errs.NotFound, err)
		}
		return ProductCategory{}, errs.Newf(errs.Internal, "update: product category[%+v]: %s", pc, err)
	}

	return ToAppProductCategory(pc), nil
}

// Delete removes an existing product category.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	pc, err := a.productcategorybus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, productcategorybus.ErrNotFound)
	}

	err = a.productcategorybus.Delete(ctx, pc)
	if err != nil {
		return errs.Newf(errs.Internal, "delete: product category[%+v]: %s", pc, err)
	}

	return nil
}

// Query returns a list of product categories based on the filter, order and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[ProductCategory], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[ProductCategory]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[ProductCategory]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[ProductCategory]{}, errs.NewFieldsError("orderby", err)
	}

	pcs, err := a.productcategorybus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[ProductCategory]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.productcategorybus.Count(ctx, filter)
	if err != nil {
		return query.Result[ProductCategory]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppProductCategories(pcs), total, page), nil
}

// QueryByID retrieves a single product category by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (ProductCategory, error) {
	pc, err := a.productcategorybus.QueryByID(ctx, id)
	if err != nil {
		return ProductCategory{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppProductCategory(pc), nil
}
