package cyclecountitemapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer APIs for cycle count item access.
type App struct {
	cycleCountItemBus *cyclecountitembus.Business
}

// NewApp constructs a cycle count item app.
func NewApp(cycleCountItemBus *cyclecountitembus.Business) *App {
	return &App{
		cycleCountItemBus: cycleCountItemBus,
	}
}

// Create adds a new cycle count item to the system.
func (a *App) Create(ctx context.Context, app NewCycleCountItem) (CycleCountItem, error) {
	ncci, err := toBusNewCycleCountItem(app)
	if err != nil {
		return CycleCountItem{}, errs.New(errs.InvalidArgument, err)
	}

	item, err := a.cycleCountItemBus.Create(ctx, ncci)
	if err != nil {
		if errors.Is(err, cyclecountitembus.ErrUniqueEntry) {
			return CycleCountItem{}, errs.New(errs.Aborted, err)
		}
		if errors.Is(err, cyclecountitembus.ErrForeignKeyViolation) {
			return CycleCountItem{}, errs.New(errs.Aborted, err)
		}
		return CycleCountItem{}, fmt.Errorf("create: %w", err)
	}

	return ToAppCycleCountItem(item), nil
}

// Update modifies an existing cycle count item.
func (a *App) Update(ctx context.Context, itemID uuid.UUID, app UpdateCycleCountItem) (CycleCountItem, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return CycleCountItem{}, errs.Newf(errs.Unauthenticated, "get user id: %s", err)
	}

	item, err := a.cycleCountItemBus.QueryByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, cyclecountitembus.ErrNotFound) {
			return CycleCountItem{}, errs.New(errs.NotFound, err)
		}
		return CycleCountItem{}, fmt.Errorf("querybyid: %w", err)
	}

	ucci, err := toBusUpdateCycleCountItem(app, userID)
	if err != nil {
		return CycleCountItem{}, errs.New(errs.InvalidArgument, err)
	}

	updated, err := a.cycleCountItemBus.Update(ctx, item, ucci)
	if err != nil {
		if errors.Is(err, cyclecountitembus.ErrForeignKeyViolation) {
			return CycleCountItem{}, errs.New(errs.Aborted, err)
		}
		return CycleCountItem{}, fmt.Errorf("update: %w", err)
	}

	return ToAppCycleCountItem(updated), nil
}

// Delete removes a cycle count item from the system.
func (a *App) Delete(ctx context.Context, itemID uuid.UUID) error {
	item, err := a.cycleCountItemBus.QueryByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, cyclecountitembus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return fmt.Errorf("delete [querybyid]: %w", err)
	}

	if err := a.cycleCountItemBus.Delete(ctx, item); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of cycle count items based on query parameters.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[CycleCountItem], error) {
	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[CycleCountItem]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[CycleCountItem]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[CycleCountItem]{}, errs.NewFieldsError("orderBy", err)
	}

	items, err := a.cycleCountItemBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return query.Result[CycleCountItem]{}, errs.Newf(errs.Internal, "query: %v", err)
	}

	total, err := a.cycleCountItemBus.Count(ctx, filter)
	if err != nil {
		return query.Result[CycleCountItem]{}, errs.Newf(errs.Internal, "count: %v", err)
	}

	return query.NewResult(ToAppCycleCountItems(items), total, pg), nil
}

// QueryByID retrieves a single cycle count item by ID.
func (a *App) QueryByID(ctx context.Context, itemID uuid.UUID) (CycleCountItem, error) {
	item, err := a.cycleCountItemBus.QueryByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, cyclecountitembus.ErrNotFound) {
			return CycleCountItem{}, errs.New(errs.NotFound, err)
		}
		return CycleCountItem{}, fmt.Errorf("querybyid: %w", err)
	}

	return ToAppCycleCountItem(item), nil
}
