// Package labelbus provides business access to label catalog and printing.
package labelbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound   = errors.New("label not found")
	ErrUniqueCode = errors.New("label code already exists")
)

// Storer declares the behavior needed to persist and retrieve label catalog entries.
type Storer interface {
	Create(ctx context.Context, lc LabelCatalog) error
	Update(ctx context.Context, lc LabelCatalog) error
	Delete(ctx context.Context, lc LabelCatalog) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]LabelCatalog, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, id uuid.UUID) (LabelCatalog, error)
	QueryByCode(ctx context.Context, code string) (LabelCatalog, error)
}

// Printer declares the behavior for sending ZPL bytes to a physical printer.
type Printer interface {
	SendZPL(ctx context.Context, zpl []byte) error
}

// Business manages the set of APIs for label access.
type Business struct {
	log      *logger.Logger
	delegate *delegate.Delegate
	storer   Storer
	printer  Printer
}

// NewBusiness constructs a label business API for use.
func NewBusiness(log *logger.Logger, d *delegate.Delegate, storer Storer, printer Printer) *Business {
	return &Business{log: log, delegate: d, storer: storer, printer: printer}
}

// Create inserts a new label into the catalog.
func (b *Business) Create(ctx context.Context, nlc NewLabelCatalog) (LabelCatalog, error) {
	lc := LabelCatalog{
		ID:          uuid.New(),
		Code:        nlc.Code,
		Type:        nlc.Type,
		EntityRef:   nlc.EntityRef,
		PayloadJSON: nlc.PayloadJSON,
		CreatedDate: time.Now(),
	}
	if err := b.storer.Create(ctx, lc); err != nil {
		return LabelCatalog{}, fmt.Errorf("create: %w", err)
	}

	// Fire delegate event for workflow automation. Errors are logged
	// but not returned — the DB write already succeeded and the caller
	// has a valid LabelCatalog back.
	if err := b.delegate.Call(ctx, ActionCreatedData(lc)); err != nil {
		b.log.Error(ctx, "labelbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return lc, nil
}

// QueryByCode retrieves a label by its stable code.
func (b *Business) QueryByCode(ctx context.Context, code string) (LabelCatalog, error) {
	lc, err := b.storer.QueryByCode(ctx, code)
	if err != nil {
		return LabelCatalog{}, fmt.Errorf("querybycode: %w", err)
	}
	return lc, nil
}

// Query retrieves a page of labels matching the filter.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pg page.Page) ([]LabelCatalog, error) {
	labels, err := b.storer.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	return labels, nil
}

// Print renders ZPL for the given catalog label id and sends it to the printer.
func (b *Business) Print(ctx context.Context, id uuid.UUID) error {
	lc, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return fmt.Errorf("querybyid: %w", err)
	}
	zpl, err := Render(lc)
	if err != nil {
		return fmt.Errorf("render: %w", err)
	}
	if err := b.printer.SendZPL(ctx, zpl); err != nil {
		return fmt.Errorf("sendzpl: %w", err)
	}
	b.log.Info(ctx, "label printed", "code", lc.Code, "bytes", len(zpl))
	return nil
}

// PrintZPL sends pre-rendered ZPL bytes directly to the printer. Used by
// transaction-label flows (render-print endpoint) that skip label_catalog.
func (b *Business) PrintZPL(ctx context.Context, zpl []byte) error {
	if err := b.printer.SendZPL(ctx, zpl); err != nil {
		return fmt.Errorf("sendzpl: %w", err)
	}
	b.log.Info(ctx, "transaction label printed", "bytes", len(zpl))
	return nil
}

// SeedCreate inserts a fully-formed LabelCatalog (caller supplies the ID).
// Seed-only — preserves deterministic UUIDs across reseeds. bus.Create
// assigns uuid.New internally, which would break determinism.
func (b *Business) SeedCreate(ctx context.Context, lc LabelCatalog) error {
	if lc.CreatedDate.IsZero() {
		lc.CreatedDate = time.Now()
	}
	if err := b.storer.Create(ctx, lc); err != nil {
		return fmt.Errorf("seedcreate: %w", err)
	}
	return nil
}
