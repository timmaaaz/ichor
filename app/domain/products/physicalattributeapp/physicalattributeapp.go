package physicalattributeapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"

	"github.com/timmaaaz/ichor/business/domain/products/physicalattributebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the physical attribute domain.
type App struct {
	physicalattributebus *physicalattributebus.Business
	auth                 *auth.Auth
}

// NewApp constructs a physical attribute app API for use.
func NewApp(physicalattributebus *physicalattributebus.Business) *App {
	return &App{
		physicalattributebus: physicalattributebus,
	}
}

// NewAppWithAuth constructs a physical attribute app API for use with auth support.
func NewAppWithAuth(physicalattributebus *physicalattributebus.Business, ath *auth.Auth) *App {
	return &App{
		auth:                 ath,
		physicalattributebus: physicalattributebus,
	}
}

// Create adds a new physical attribute to the system.
func (a *App) Create(ctx context.Context, app NewPhysicalAttribute) (PhysicalAttribute, error) {
	npa, err := toBusNewPhysicalAttribute(app)
	if err != nil {
		return PhysicalAttribute{}, errs.New(errs.InvalidArgument, err)
	}

	physicalAttribute, err := a.physicalattributebus.Create(ctx, npa)
	if err != nil {
		if errors.Is(err, physicalattributebus.ErrUniqueEntry) {
			return PhysicalAttribute{}, errs.New(errs.AlreadyExists, physicalattributebus.ErrUniqueEntry)
		}
		if errors.Is(err, physicalattributebus.ErrForeignKeyViolation) {
			return PhysicalAttribute{}, errs.New(errs.Aborted, physicalattributebus.ErrForeignKeyViolation)
		}
		return PhysicalAttribute{}, errs.Newf(errs.Internal, "create: physical attribute[%+v]: %s", physicalAttribute, err)
	}

	return ToAppPhysicalAttribute(physicalAttribute), err
}

// Update updates an existing physical attribute.
func (a *App) Update(ctx context.Context, app UpdatePhysicalAttribute, id uuid.UUID) (PhysicalAttribute, error) {
	upa, err := toBusUpdatePhysicalAttribute(app)
	if err != nil {
		return PhysicalAttribute{}, errs.New(errs.InvalidArgument, err)
	}

	st, err := a.physicalattributebus.QueryByID(ctx, id)
	if err != nil {
		return PhysicalAttribute{}, errs.New(errs.NotFound, physicalattributebus.ErrNotFound)
	}

	physicalAttribute, err := a.physicalattributebus.Update(ctx, st, upa)
	if err != nil {
		if errors.Is(err, physicalattributebus.ErrForeignKeyViolation) {
			return PhysicalAttribute{}, errs.New(errs.Aborted, physicalattributebus.ErrForeignKeyViolation)
		}
		if errors.Is(err, physicalattributebus.ErrNotFound) {
			return PhysicalAttribute{}, errs.New(errs.NotFound, err)
		}
		return PhysicalAttribute{}, errs.Newf(errs.Internal, "update: physical attribute[%+v]: %s", physicalAttribute, err)
	}

	return ToAppPhysicalAttribute(physicalAttribute), nil
}

// Delete removes an existing physical attribute.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	physicalAttribute, err := a.physicalattributebus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, physicalattributebus.ErrNotFound)
	}

	err = a.physicalattributebus.Delete(ctx, physicalAttribute)
	if err != nil {
		return errs.Newf(errs.Internal, "delete: physical attribute[%+v]: %s", physicalAttribute, err)
	}

	return nil
}

// Query returns a list of physical attributes based on the filter, order and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[PhysicalAttribute], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[PhysicalAttribute]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[PhysicalAttribute]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[PhysicalAttribute]{}, errs.NewFieldsError("orderby", err)
	}

	physicalAttributes, err := a.physicalattributebus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[PhysicalAttribute]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.physicalattributebus.Count(ctx, filter)
	if err != nil {
		return query.Result[PhysicalAttribute]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppPhysicalAttributes(physicalAttributes), total, page), nil
}

// QueryByID retrieves a single physical attribute by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (PhysicalAttribute, error) {
	physicalAttribute, err := a.physicalattributebus.QueryByID(ctx, id)
	if err != nil {
		return PhysicalAttribute{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppPhysicalAttribute(physicalAttribute), nil
}
