package contactinfobus

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/convert"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("contactInfo not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("contactInfo entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, contactInfo ContactInfo) error
	Update(ctx context.Context, contactInfo ContactInfo) error
	Delete(ctx context.Context, contactInfo ContactInfo) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ContactInfo, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, contactInfoID uuid.UUID) (ContactInfo, error)
}

// Business manages the set of APIs for contactInfo access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a contactInfo business API for use.
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

// Create inserts a new contactInfo into the database.
func (b *Business) Create(ctx context.Context, nci NewContactInfo) (ContactInfo, error) {
	ctx, span := otel.AddSpan(ctx, "business.contactInfobus.create")
	defer span.End()

	contactInfo := ContactInfo{
		ID:                   uuid.New(),
		FirstName:            nci.FirstName,
		LastName:             nci.LastName,
		EmailAddress:         nci.EmailAddress,
		PrimaryPhone:         nci.PrimaryPhone,
		SecondaryPhone:       nci.SecondaryPhone,
		Address:              nci.Address,
		AvailableHoursStart:  nci.AvailableHoursStart,
		AvailableHoursEnd:    nci.AvailableHoursEnd,
		Timezone:             nci.Timezone,
		PreferredContactType: nci.PreferredContactType,
		Notes:                nci.Notes,
	}

	if err := b.storer.Create(ctx, contactInfo); err != nil {
		return ContactInfo{}, fmt.Errorf("create: %w", err)
	}

	return contactInfo, nil
}

// Update replaces an contactInfo document in the database.
func (b *Business) Update(ctx context.Context, ci ContactInfo, uci UpdateContactInfo) (ContactInfo, error) {
	ctx, span := otel.AddSpan(ctx, "business.contactInfobus.update")
	defer span.End()

	if err := convert.PopulateSameTypes(uci, &ci); err != nil {
		return ContactInfo{}, fmt.Errorf("populate contactInfo from update contactInfo: %w", err)
	}

	if err := b.storer.Update(ctx, ci); err != nil {
		return ContactInfo{}, fmt.Errorf("update: %w", err)
	}

	return ci, nil
}

// Delete removes the specified contactInfo.
func (b *Business) Delete(ctx context.Context, ci ContactInfo) error {
	ctx, span := otel.AddSpan(ctx, "business.contactInfobus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, ci); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of contactInfos from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ContactInfo, error) {
	ctx, span := otel.AddSpan(ctx, "business.contactInfobus.Query")
	defer span.End()

	contacts, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return contacts, nil
}

// Count returns the total number of contactInfos.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.contactInfobus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the contactInfo by the specified ID.
func (b *Business) QueryByID(ctx context.Context, contactInfoID uuid.UUID) (ContactInfo, error) {
	ctx, span := otel.AddSpan(ctx, "business.contactInfobus.querybyid")
	defer span.End()

	contactInfo, err := b.storer.QueryByID(ctx, contactInfoID)
	if err != nil {
		return ContactInfo{}, fmt.Errorf("query: contactInfoID[%s]: %w", contactInfoID, err)
	}

	return contactInfo, nil
}
