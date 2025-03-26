package lottrackingbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/convert"
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
	Create(ctx context.Context, lotTracking LotTracking) error
	Update(ctx context.Context, lotTracking LotTracking) error
	Delete(ctx context.Context, lotTracking LotTracking) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]LotTracking, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, lotTrackingID uuid.UUID) (LotTracking, error)
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

func (b *Business) Create(ctx context.Context, nlt NewLotTracking) (LotTracking, error) {
	ctx, span := otel.AddSpan(ctx, "business.lottrackingbus.create")
	defer span.End()

	now := time.Now()

	lt := LotTracking{
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
		return LotTracking{}, fmt.Errorf("create: %w", err)
	}

	return lt, nil
}

// Update modifies a lot tracking in the system.
func (b *Business) Update(ctx context.Context, lt LotTracking, ul UpdateLotTracking) (LotTracking, error) {
	ctx, span := otel.AddSpan(ctx, "business.lottrackingbus.update")
	defer span.End()

	err := convert.PopulateSameTypes(ul, &lt)
	if err != nil {
		return LotTracking{}, fmt.Errorf("populate lot tracking from update lot tracking: %w", err)
	}

	lt.UpdatedDate = time.Now()

	err = b.storer.Update(ctx, lt)
	if err != nil {
		return LotTracking{}, fmt.Errorf("update: %w", err)
	}

	return lt, nil
}

// Delete removes a lot tracking from the system.
func (b *Business) Delete(ctx context.Context, lt LotTracking) error {
	ctx, span := otel.AddSpan(ctx, "business.lottrackingbus.delete")
	defer span.End()

	err := b.storer.Delete(ctx, lt)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of lot trackings based on the provided query filter,
// order, and pagination options.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]LotTracking, error) {
	ctx, span := otel.AddSpan(ctx, "business.lottrackingbus.query")
	defer span.End()

	items, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return items, nil
}

// Count returns the total number of lot trackings.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.lottrackingbus.count")
	defer span.End()

	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}

// QueryByID retrieves a lot tracking by its unique ID.
func (b *Business) QueryByID(ctx context.Context, lotTrackingID uuid.UUID) (LotTracking, error) {
	ctx, span := otel.AddSpan(ctx, "business.lottrackingbus.querybyid")
	defer span.End()

	lt, err := b.storer.QueryByID(ctx, lotTrackingID)
	if err != nil {
		return LotTracking{}, fmt.Errorf("query by id: %w", err)
	}

	return lt, nil
}
