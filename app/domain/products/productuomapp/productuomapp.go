package productuomapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/products/productuombus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer API functions for the product UOM domain.
type App struct {
	productuombus *productuombus.Business
}

// NewApp constructs a product UOM App.
func NewApp(productuombus *productuombus.Business) *App {
	return &App{productuombus: productuombus}
}

// Create adds a new product UOM.
func (a *App) Create(ctx context.Context, app NewProductUOM) (ProductUOM, error) {
	np, err := toBusNewProductUOM(app)
	if err != nil {
		return ProductUOM{}, errs.New(errs.InvalidArgument, err)
	}

	uom, err := a.productuombus.Create(ctx, np)
	if err != nil {
		if errors.Is(err, productuombus.ErrUniqueEntry) {
			return ProductUOM{}, errs.New(errs.AlreadyExists, productuombus.ErrUniqueEntry)
		}
		return ProductUOM{}, errs.Newf(errs.Internal, "create: uom[%+v]: %s", uom, err)
	}

	return ToAppProductUOM(uom), nil
}

// Update modifies an existing product UOM.
func (a *App) Update(ctx context.Context, app UpdateProductUOM, uomID uuid.UUID) (ProductUOM, error) {
	uom, err := a.productuombus.QueryByID(ctx, uomID)
	if err != nil {
		return ProductUOM{}, errs.New(errs.NotFound, productuombus.ErrNotFound)
	}

	updated, err := a.productuombus.Update(ctx, uom, toBusUpdateProductUOM(app))
	if err != nil {
		if errors.Is(err, productuombus.ErrUniqueEntry) {
			return ProductUOM{}, errs.New(errs.AlreadyExists, productuombus.ErrUniqueEntry)
		}
		if errors.Is(err, productuombus.ErrNotFound) {
			return ProductUOM{}, errs.New(errs.NotFound, err)
		}
		return ProductUOM{}, errs.Newf(errs.Internal, "update: uom[%+v]: %s", updated, err)
	}

	return ToAppProductUOM(updated), nil
}

// Delete removes a product UOM.
func (a *App) Delete(ctx context.Context, uomID uuid.UUID) error {
	uom, err := a.productuombus.QueryByID(ctx, uomID)
	if err != nil {
		return errs.New(errs.NotFound, productuombus.ErrNotFound)
	}

	if err := a.productuombus.Delete(ctx, uom); err != nil {
		return errs.Newf(errs.Internal, "delete: uom[%+v]: %s", uom, err)
	}

	return nil
}

// Query returns a list of product UOMs.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[ProductUOM], error) {
	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[ProductUOM]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[ProductUOM]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[ProductUOM]{}, errs.NewFieldsError("orderby", err)
	}

	uoms, err := a.productuombus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return query.Result[ProductUOM]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.productuombus.Count(ctx, filter)
	if err != nil {
		return query.Result[ProductUOM]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppProductUOMs(uoms), total, pg), nil
}

// QueryByID returns a single product UOM by ID.
func (a *App) QueryByID(ctx context.Context, uomID uuid.UUID) (ProductUOM, error) {
	uom, err := a.productuombus.QueryByID(ctx, uomID)
	if err != nil {
		return ProductUOM{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppProductUOM(uom), nil
}
