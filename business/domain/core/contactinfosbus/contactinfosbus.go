package contactinfosbus

import (
	"context"
	"errors"
	"fmt"

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
	ErrNotFound              = errors.New("contactInfos not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("contactInfos entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, contactInfos ContactInfos) error
	Update(ctx context.Context, contactInfos ContactInfos) error
	Delete(ctx context.Context, contactInfos ContactInfos) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ContactInfos, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, contactInfosID uuid.UUID) (ContactInfos, error)
}

// Business manages the set of APIs for contactInfos access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a contactInfos business API for use.
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

// Create inserts a new contactInfos into the database.
func (b *Business) Create(ctx context.Context, nci NewContactInfos) (ContactInfos, error) {
	ctx, span := otel.AddSpan(ctx, "business.contactInfosbus.create")
	defer span.End()

	contactInfos := ContactInfos{
		ID:                   uuid.New(),
		FirstName:            nci.FirstName,
		LastName:             nci.LastName,
		EmailAddress:         nci.EmailAddress,
		PrimaryPhone:         nci.PrimaryPhone,
		SecondaryPhone:       nci.SecondaryPhone,
		StreetID:             nci.StreetID,
		DeliveryAddressID:    nci.DeliveryAddressID,
		AvailableHoursStart:  nci.AvailableHoursStart,
		AvailableHoursEnd:    nci.AvailableHoursEnd,
		Timezone:             nci.Timezone,
		PreferredContactType: nci.PreferredContactType,
		Notes:                nci.Notes,
	}

	if err := b.storer.Create(ctx, contactInfos); err != nil {
		return ContactInfos{}, fmt.Errorf("create: %w", err)
	}

	return contactInfos, nil
}

// Update replaces an contactInfos document in the database.
func (b *Business) Update(ctx context.Context, ci ContactInfos, uci UpdateContactInfos) (ContactInfos, error) {
	ctx, span := otel.AddSpan(ctx, "business.contactInfosbus.update")
	defer span.End()

	if uci.FirstName != nil {
		ci.FirstName = *uci.FirstName
	}
	if uci.LastName != nil {
		ci.LastName = *uci.LastName
	}
	if uci.EmailAddress != nil {
		ci.EmailAddress = *uci.EmailAddress
	}
	if uci.PrimaryPhone != nil {
		ci.PrimaryPhone = *uci.PrimaryPhone
	}
	if uci.SecondaryPhone != nil {
		ci.SecondaryPhone = *uci.SecondaryPhone
	}
	if uci.StreetID != nil {
		ci.StreetID = *uci.StreetID
	}
	if uci.DeliveryAddressID != nil {
		ci.DeliveryAddressID = *uci.DeliveryAddressID
	}
	if uci.AvailableHoursStart != nil {
		ci.AvailableHoursStart = *uci.AvailableHoursStart
	}
	if uci.AvailableHoursEnd != nil {
		ci.AvailableHoursEnd = *uci.AvailableHoursEnd
	}
	if uci.Timezone != nil {
		ci.Timezone = *uci.Timezone
	}
	if uci.PreferredContactType != nil {
		ci.PreferredContactType = *uci.PreferredContactType
	}
	if uci.Notes != nil {
		ci.Notes = *uci.Notes
	}

	if err := b.storer.Update(ctx, ci); err != nil {
		return ContactInfos{}, fmt.Errorf("update: %w", err)
	}

	return ci, nil
}

// Delete removes the specified contactInfos.
func (b *Business) Delete(ctx context.Context, ci ContactInfos) error {
	ctx, span := otel.AddSpan(ctx, "business.contactInfosbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, ci); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of contactInfoss from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ContactInfos, error) {
	ctx, span := otel.AddSpan(ctx, "business.contactInfosbus.Query")
	defer span.End()

	contacts, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return contacts, nil
}

// Count returns the total number of contactInfoss.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.contactInfosbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the contactInfos by the specified ID.
func (b *Business) QueryByID(ctx context.Context, contactInfosID uuid.UUID) (ContactInfos, error) {
	ctx, span := otel.AddSpan(ctx, "business.contactInfosbus.querybyid")
	defer span.End()

	contactInfos, err := b.storer.QueryByID(ctx, contactInfosID)
	if err != nil {
		return ContactInfos{}, fmt.Errorf("query: contactInfosID[%s]: %w", contactInfosID, err)
	}

	return contactInfos, nil
}
