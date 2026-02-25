package supplierproductapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

type App struct {
	supplierproductbus *supplierproductbus.Business
	auth               *auth.Auth
}

func NewApp(supplierproductbus *supplierproductbus.Business) *App {
	return &App{
		supplierproductbus: supplierproductbus,
	}
}

func NewAppWithAuth(supplierproductbus *supplierproductbus.Business, auth *auth.Auth) *App {
	return &App{
		supplierproductbus: supplierproductbus,
		auth:               auth,
	}
}

func (a *App) Create(ctx context.Context, app NewSupplierProduct) (SupplierProduct, error) {
	np, err := toBusNewSupplierProduct(app)
	if err != nil {
		return SupplierProduct{}, err
	}

	sp, err := a.supplierproductbus.Create(ctx, np)
	if err != nil {
		if errors.Is(err, supplierproductbus.ErrUniqueEntry) {
			return SupplierProduct{}, errs.New(errs.AlreadyExists, supplierproductbus.ErrUniqueEntry)
		}
		if errors.Is(err, supplierproductbus.ErrForeignKeyViolation) {
			return SupplierProduct{}, errs.New(errs.Aborted, supplierproductbus.ErrForeignKeyViolation)
		}
		return SupplierProduct{}, errs.Newf(errs.Internal, "create: product cost[%+v]: %s", sp, err)
	}

	return ToAppSupplierProduct(sp), err
}

func (a *App) Update(ctx context.Context, app UpdateSupplierProduct, id uuid.UUID) (SupplierProduct, error) {
	usp, err := toBusUpdateSupplierProduct(app)
	if err != nil {
		return SupplierProduct{}, errs.New(errs.InvalidArgument, err)
	}

	sp, err := a.supplierproductbus.QueryByID(ctx, id)
	if err != nil {
		return SupplierProduct{}, errs.New(errs.NotFound, supplierproductbus.ErrNotFound)
	}

	supplierProduct, err := a.supplierproductbus.Update(ctx, sp, usp)
	if err != nil {
		if errors.Is(err, supplierproductbus.ErrForeignKeyViolation) {
			return SupplierProduct{}, errs.New(errs.Aborted, supplierproductbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, supplierproductbus.ErrNotFound) {
			return SupplierProduct{}, errs.New(errs.NotFound, err)
		}
		return SupplierProduct{}, errs.Newf(errs.Internal, "update: supplier product[%+v]: %s", supplierProduct, err)
	}

	return ToAppSupplierProduct(supplierProduct), nil
}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	sp, err := a.supplierproductbus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, supplierproductbus.ErrNotFound)
	}

	err = a.supplierproductbus.Delete(ctx, sp)
	if err != nil {
		if errors.Is(err, supplierproductbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "delete: supplier product[%+v]: %s", sp, err)
	}

	return nil
}

func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[SupplierProduct], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[SupplierProduct]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[SupplierProduct]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[SupplierProduct]{}, errs.NewFieldsError("orderby", err)
	}

	supplierProducts, err := a.supplierproductbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[SupplierProduct]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.supplierproductbus.Count(ctx, filter)
	if err != nil {
		return query.Result[SupplierProduct]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppSupplierProducts(supplierProducts), total, page), nil
}

func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (SupplierProduct, error) {
	sp, err := a.supplierproductbus.QueryByID(ctx, id)
	if err != nil {
		return SupplierProduct{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppSupplierProduct(sp), nil
}

// QueryByIDs retrieves multiple supplier products by their IDs.
func (a *App) QueryByIDs(ctx context.Context, ids []string) (SupplierProducts, error) {
	uuids, err := toBusIDs(ids)
	if err != nil {
		return nil, errs.New(errs.InvalidArgument, err)
	}

	sps, err := a.supplierproductbus.QueryByIDs(ctx, uuids)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "querybyids: %s", err)
	}

	return ToAppSupplierProducts(sps), nil
}
