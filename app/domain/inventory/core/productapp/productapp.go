package productapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"

	"github.com/timmaaaz/ichor/business/domain/inventory/core/productbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the product domain.
type App struct {
	productbus *productbus.Business
	auth       *auth.Auth
}

// NewApp constructs a product app API for use.
func NewApp(productbus *productbus.Business) *App {
	return &App{
		productbus: productbus,
	}
}

// NewAppWithAuth constructs a product app API for use with auth support.
func NewAppWithAuth(productbus *productbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:       ath,
		productbus: productbus,
	}
}

// Create adds a new product to the system.
func (a *App) Create(ctx context.Context, app NewProduct) (Product, error) {
	np, err := toBusNewProduct(app)
	if err != nil {
		return Product{}, errs.New(errs.InvalidArgument, err)
	}

	product, err := a.productbus.Create(ctx, np)
	if err != nil {
		if errors.Is(err, productbus.ErrUniqueEntry) {
			return Product{}, errs.New(errs.AlreadyExists, productbus.ErrUniqueEntry)
		}
		if errors.Is(err, productbus.ErrForeignKeyViolation) {
			return Product{}, errs.New(errs.Aborted, productbus.ErrForeignKeyViolation)
		}
		return Product{}, errs.Newf(errs.Internal, "create: product[%+v]: %s", product, err)
	}

	return ToAppProduct(product), err
}

// Update updates an existing product.
func (a *App) Update(ctx context.Context, app UpdateProduct, id uuid.UUID) (Product, error) {
	up, err := toBusUpdateProduct(app)
	if err != nil {
		return Product{}, errs.New(errs.InvalidArgument, err)
	}

	st, err := a.productbus.QueryByID(ctx, id)
	if err != nil {
		return Product{}, errs.New(errs.NotFound, productbus.ErrNotFound)
	}

	product, err := a.productbus.Update(ctx, st, up)
	if err != nil {
		if errors.Is(err, productbus.ErrForeignKeyViolation) {
			return Product{}, errs.New(errs.Aborted, productbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, productbus.ErrNotFound) {
			return Product{}, errs.New(errs.NotFound, err)
		}
		return Product{}, errs.Newf(errs.Internal, "update: product[%+v]: %s", product, err)
	}

	return ToAppProduct(product), nil
}

// Delete removes an existing product.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	product, err := a.productbus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, productbus.ErrNotFound)
	}

	err = a.productbus.Delete(ctx, product)
	if err != nil {
		return errs.Newf(errs.Internal, "delete: product[%+v]: %s", product, err)
	}

	return nil
}

// Query returns a list of products based on the filter, order and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Product], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Product]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Product]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Product]{}, errs.NewFieldsError("orderby", err)
	}

	products, err := a.productbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[Product]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.productbus.Count(ctx, filter)
	if err != nil {
		return query.Result[Product]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppProducts(products), total, page), nil
}

// QueryByID retrieves a single product by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (Product, error) {
	product, err := a.productbus.QueryByID(ctx, id)
	if err != nil {
		return Product{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppProduct(product), nil
}
