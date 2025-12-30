package customersbus

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
	ErrNotFound              = errors.New("customers not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("customers entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, customers Customers) error
	Update(ctx context.Context, customers Customers) error
	Delete(ctx context.Context, customers Customers) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Customers, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, customersID uuid.UUID) (Customers, error)
}

// Business manages the set of APIs for customers access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a customers business API for use.
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

// Create inserts a new customers into the database.
func (b *Business) Create(ctx context.Context, nci NewCustomers) (Customers, error) {
	ctx, span := otel.AddSpan(ctx, "business.customersbus.create")
	defer span.End()

	now := time.Now().UTC()
	if nci.CreatedDate != nil {
		now = *nci.CreatedDate // Use provided date for seeding
	}

	customers := Customers{
		ID:                uuid.New(),
		Name:              nci.Name,
		ContactID:         nci.ContactID,
		DeliveryAddressID: nci.DeliveryAddressID,
		Notes:             nci.Notes,
		CreatedBy:         nci.CreatedBy,
		UpdatedBy:         nci.CreatedBy,
		CreatedDate:       now,
		UpdatedDate:       now,
	}

	if err := b.storer.Create(ctx, customers); err != nil {
		return Customers{}, fmt.Errorf("create: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(customers)); err != nil {
		b.log.Error(ctx, "customersbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return customers, nil
}

// Update replaces an customers document in the database.
func (b *Business) Update(ctx context.Context, ci Customers, uci UpdateCustomers) (Customers, error) {
	ctx, span := otel.AddSpan(ctx, "business.customersbus.update")
	defer span.End()

	if uci.Name != nil {
		ci.Name = *uci.Name
	}
	if uci.ContactID != nil {
		ci.ContactID = *uci.ContactID
	}
	if uci.DeliveryAddressID != nil {
		ci.DeliveryAddressID = *uci.DeliveryAddressID
	}
	if uci.Notes != nil {
		ci.Notes = *uci.Notes
	}
	if uci.UpdatedBy != nil {
		ci.UpdatedBy = *uci.UpdatedBy
	}

	ci.UpdatedDate = time.Now().UTC()

	if err := b.storer.Update(ctx, ci); err != nil {
		return Customers{}, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(ci)); err != nil {
		b.log.Error(ctx, "customersbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return ci, nil
}

// Delete removes the specified customers.
func (b *Business) Delete(ctx context.Context, ci Customers) error {
	ctx, span := otel.AddSpan(ctx, "business.customersbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, ci); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(ci)); err != nil {
		b.log.Error(ctx, "customersbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves a list of customerss from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Customers, error) {
	ctx, span := otel.AddSpan(ctx, "business.customersbus.Query")
	defer span.End()

	contacts, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return contacts, nil
}

// Count returns the total number of customerss.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.customersbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the customers by the specified ID.
func (b *Business) QueryByID(ctx context.Context, customersID uuid.UUID) (Customers, error) {
	ctx, span := otel.AddSpan(ctx, "business.customersbus.querybyid")
	defer span.End()

	customers, err := b.storer.QueryByID(ctx, customersID)
	if err != nil {
		return Customers{}, fmt.Errorf("query: customersID[%s]: %w", customersID, err)
	}

	return customers, nil
}
