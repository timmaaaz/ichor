package serialnumberbus

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
	ErrNotFound              = errors.New("serialNumber not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("serialNumber entry is not unique")
	ErrForeignKeyViolation   = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, sn SerialNumber) error
	Update(ctx context.Context, sn SerialNumber) error
	Delete(ctx context.Context, sn SerialNumber) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]SerialNumber, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, snID uuid.UUID) (SerialNumber, error)
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

func (b *Business) Create(ctx context.Context, nsn NewSerialNumber) (SerialNumber, error) {
	ctx, span := otel.AddSpan(ctx, "business.serialnumberbus.create")
	defer span.End()

	now := time.Now()

	sn := SerialNumber{
		SerialID:     uuid.New(),
		ProductID:    nsn.ProductID,
		LocationID:   nsn.LocationID,
		SerialNumber: nsn.SerialNumber,
		LotID:        nsn.LotID,
		Status:       nsn.Status,
		UpdatedDate:  now,
		CreatedDate:  now,
	}

	err := b.storer.Create(ctx, sn)
	if err != nil {
		return sn, fmt.Errorf("create: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(sn)); err != nil {
		b.log.Error(ctx, "serialnumberbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return sn, nil
}

func (b *Business) Update(ctx context.Context, sn SerialNumber, usn UpdateSerialNumber) (SerialNumber, error) {
	ctx, span := otel.AddSpan(ctx, "business.serialnumberbus.update")
	defer span.End()

	if usn.LotID != nil {
		sn.LotID = *usn.LotID
	}
	if usn.ProductID != nil {
		sn.ProductID = *usn.ProductID
	}
	if usn.LocationID != nil {
		sn.LocationID = *usn.LocationID
	}
	if usn.SerialNumber != nil {
		sn.SerialNumber = *usn.SerialNumber
	}
	if usn.Status != nil {
		sn.Status = *usn.Status
	}

	sn.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, sn); err != nil {
		return sn, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(sn)); err != nil {
		b.log.Error(ctx, "serialnumberbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return sn, nil
}

func (b *Business) Delete(ctx context.Context, sn SerialNumber) error {
	ctx, span := otel.AddSpan(ctx, "business.serialnumberbus.delete")
	defer span.End()

	err := b.storer.Delete(ctx, sn)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(sn)); err != nil {
		b.log.Error(ctx, "serialnumberbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]SerialNumber, error) {
	ctx, span := otel.AddSpan(ctx, "business.serialnumberbus.query")
	defer span.End()

	snList, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("querying: %w", err)
	}

	return snList, nil
}

func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.serialnumberbus.count")
	defer span.End()

	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}

func (b *Business) QueryByID(ctx context.Context, snID uuid.UUID) (SerialNumber, error) {
	ctx, span := otel.AddSpan(ctx, "business.serialnumberbus.querybyid")
	defer span.End()

	sn, err := b.storer.QueryByID(ctx, snID)
	if err != nil {
		return SerialNumber{}, fmt.Errorf("querying by ID: %w", err)
	}

	return sn, nil
}
