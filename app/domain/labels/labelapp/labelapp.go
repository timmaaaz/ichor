// Package labelapp provides the app-layer orchestration for label catalog
// lookups, catalog printing (POST /v1/labels/print), and on-the-fly
// transaction-label printing (POST /v1/labels/render-print). All printing
// flows through labelbus.Business and the injected Printer.
package labelapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/labels/labelbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app-layer APIs for the label domain.
type App struct {
	bus *labelbus.Business
}

// NewApp constructs a label app API for use.
func NewApp(bus *labelbus.Business) *App {
	return &App{bus: bus}
}

// Print renders the referenced catalog label and dispatches it to the
// printer. Copies defaults to 1; each copy is a fresh SendZPL call so a
// dropped connection only loses one label.
func (a *App) Print(ctx context.Context, req PrintRequest) error {
	id, err := uuid.Parse(req.LabelID)
	if err != nil {
		return errs.Newf(errs.InvalidArgument, "parse label_id: %s", err)
	}

	copies := req.Copies
	if copies < 1 {
		copies = 1
	}

	lc, err := a.bus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, labelbus.ErrNotFound) {
			return errs.New(errs.NotFound, labelbus.ErrNotFound)
		}
		return errs.Newf(errs.Internal, "querybyid label[%s]: %s", id, err)
	}

	zpl, err := labelbus.Render(lc)
	if err != nil {
		return errs.Newf(errs.Internal, "render label[%s]: %s", id, err)
	}

	for i := 0; i < copies; i++ {
		if err := a.bus.PrintZPL(ctx, zpl); err != nil {
			return errs.Newf(errs.Internal, "printzpl label[%s] copy[%d]: %s", id, i+1, err)
		}
	}
	return nil
}

// RenderPrint renders an ad-hoc payload in-memory (no catalog row, no DB
// write) and sends it to the printer. Used by transaction-label flows
// like receiving and pick where each label is unique to a single event.
func (a *App) RenderPrint(ctx context.Context, req RenderPrintRequest) error {
	lc := labelbus.LabelCatalog{
		Type:        req.Type,
		Code:        req.Code,
		PayloadJSON: string(req.Payload),
	}
	zpl, err := labelbus.Render(lc)
	if err != nil {
		return errs.Newf(errs.InvalidArgument, "render: %s", err)
	}

	copies := req.Copies
	if copies < 1 {
		copies = 1
	}

	for i := 0; i < copies; i++ {
		if err := a.bus.PrintZPL(ctx, zpl); err != nil {
			return errs.Newf(errs.Internal, "printzpl copy[%d]: %s", i+1, err)
		}
	}
	return nil
}

// Create inserts a new label catalog row.
func (a *App) Create(ctx context.Context, app NewLabel) (Label, error) {
	lc, err := a.bus.Create(ctx, toBusNewLabel(app))
	if err != nil {
		if errors.Is(err, labelbus.ErrUniqueCode) {
			return Label{}, errs.New(errs.Aborted, labelbus.ErrUniqueCode)
		}
		return Label{}, errs.Newf(errs.Internal, "create: %s", err)
	}
	return toAppLabel(lc), nil
}

// Update applies a partial patch to an existing label.
func (a *App) Update(ctx context.Context, labelID uuid.UUID, app UpdateLabel) (Label, error) {
	lc, err := a.bus.QueryByID(ctx, labelID)
	if err != nil {
		if errors.Is(err, labelbus.ErrNotFound) {
			return Label{}, errs.New(errs.NotFound, labelbus.ErrNotFound)
		}
		return Label{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	updated, err := a.bus.Update(ctx, lc, toBusUpdateLabel(app))
	if err != nil {
		if errors.Is(err, labelbus.ErrUniqueCode) {
			return Label{}, errs.New(errs.Aborted, labelbus.ErrUniqueCode)
		}
		return Label{}, errs.Newf(errs.Internal, "update: %s", err)
	}
	return toAppLabel(updated), nil
}

// Delete removes a label catalog row.
func (a *App) Delete(ctx context.Context, labelID uuid.UUID) error {
	lc, err := a.bus.QueryByID(ctx, labelID)
	if err != nil {
		if errors.Is(err, labelbus.ErrNotFound) {
			return errs.New(errs.NotFound, labelbus.ErrNotFound)
		}
		return errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	if err := a.bus.Delete(ctx, lc); err != nil {
		return errs.Newf(errs.Internal, "delete: %s", err)
	}
	return nil
}

// QueryByID returns the catalog label with the given ID.
func (a *App) QueryByID(ctx context.Context, labelID uuid.UUID) (Label, error) {
	lc, err := a.bus.QueryByID(ctx, labelID)
	if err != nil {
		if errors.Is(err, labelbus.ErrNotFound) {
			return Label{}, errs.New(errs.NotFound, labelbus.ErrNotFound)
		}
		return Label{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}
	return toAppLabel(lc), nil
}

// Query returns catalog labels matching the optional Type filter. No Count
// call is exposed on the Business yet, so this returns a plain slice with
// standard pagination.
func (a *App) Query(ctx context.Context, qp QueryParams) (Labels, error) {
	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return nil, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return nil, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return nil, errs.NewFieldsError("orderby", err)
	}

	labels, err := a.bus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "query: %s", err)
	}
	return toAppLabels(labels), nil
}
