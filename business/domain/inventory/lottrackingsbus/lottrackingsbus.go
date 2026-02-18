package lottrackingsbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("lot not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("lot entry is not unique")
	ErrForeignKeyViolation   = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, lotTrackings LotTrackings) error
	Update(ctx context.Context, lotTrackings LotTrackings) error
	Delete(ctx context.Context, lotTrackings LotTrackings) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]LotTrackings, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, lotTrackingsID uuid.UUID) (LotTrackings, error)
}

// Business manages the set of APIs for brand access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a brand business API for use.
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:      log,
		delegate: delegate,
		storer:   storer,
	}
}

// NewWithTx constructs a new Business value replacing the Storer
// value with a Storer value that is currently inside a transaction.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	return &Business{
		log:      b.log,
		delegate: b.delegate,
		storer:   storer,
	}, nil
}

func (b *Business) Create(ctx context.Context, nlt NewLotTrackings) (LotTrackings, error) {
	ctx, span := otel.AddSpan(ctx, "business.lottrackingsbus.create")
	defer span.End()

	now := time.Now()

	lt := LotTrackings{
		LotID:             uuid.New(),
		SupplierProductID: nlt.SupplierProductID,
		LotNumber:         nlt.LotNumber,
		ManufactureDate:   nlt.ManufactureDate,
		ExpirationDate:    nlt.ExpirationDate,
		RecievedDate:      nlt.RecievedDate,
		Quantity:          nlt.Quantity,
		QualityStatus:     nlt.QualityStatus,
		CreatedDate:       now,
		UpdatedDate:       now,
	}

	err := b.storer.Create(ctx, lt)
	if err != nil {
		return LotTrackings{}, fmt.Errorf("create: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(lt)); err != nil {
		b.log.Error(ctx, "lottrackingsbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return lt, nil
}

// Update modifies a lot tracking in the system.
func (b *Business) Update(ctx context.Context, lt LotTrackings, ul UpdateLotTrackings) (LotTrackings, error) {
	ctx, span := otel.AddSpan(ctx, "business.lottrackingsbus.update")
	defer span.End()

	before := lt

	if ul.SupplierProductID != nil {
		lt.SupplierProductID = *ul.SupplierProductID
	}
	if ul.LotNumber != nil {
		lt.LotNumber = *ul.LotNumber
	}
	if ul.ManufactureDate != nil {
		lt.ManufactureDate = *ul.ManufactureDate
	}
	if ul.ExpirationDate != nil {
		lt.ExpirationDate = *ul.ExpirationDate
	}
	if ul.RecievedDate != nil {
		lt.RecievedDate = *ul.RecievedDate
	}
	if ul.Quantity != nil {
		lt.Quantity = *ul.Quantity
	}
	if ul.QualityStatus != nil {
		lt.QualityStatus = *ul.QualityStatus
	}

	lt.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, lt); err != nil {
		return LotTrackings{}, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(before, lt)); err != nil {
		b.log.Error(ctx, "lottrackingsbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return lt, nil
}

// Delete removes a lot tracking from the system.
func (b *Business) Delete(ctx context.Context, lt LotTrackings) error {
	ctx, span := otel.AddSpan(ctx, "business.lottrackingsbus.delete")
	defer span.End()

	err := b.storer.Delete(ctx, lt)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(lt)); err != nil {
		b.log.Error(ctx, "lottrackingsbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves a list of lot trackings based on the provided query filter,
// order, and pagination options.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]LotTrackings, error) {
	ctx, span := otel.AddSpan(ctx, "business.lottrackingsbus.query")
	defer span.End()

	items, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return items, nil
}

// Count returns the total number of lot trackings.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.lottrackingsbus.count")
	defer span.End()

	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}

// QueryByID retrieves a lot tracking by its unique ID.
func (b *Business) QueryByID(ctx context.Context, lotTrackingsID uuid.UUID) (LotTrackings, error) {
	ctx, span := otel.AddSpan(ctx, "business.lottrackingsbus.querybyid")
	defer span.End()

	lt, err := b.storer.QueryByID(ctx, lotTrackingsID)
	if err != nil {
		return LotTrackings{}, fmt.Errorf("query by id: %w", err)
	}

	return lt, nil
}
