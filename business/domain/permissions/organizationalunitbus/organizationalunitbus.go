package organizationalunitbus

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("role not found")
	ErrUnique                = errors.New("organizational unit is not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, ou OrganizationalUnit) error
	Update(ctx context.Context, ou OrganizationalUnit) error
	Delete(ctx context.Context, ou OrganizationalUnit) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]OrganizationalUnit, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, userID uuid.UUID) (OrganizationalUnit, error)
}

// Business manages the set of APIs for user access.
type Business struct {
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs a user business API for use.
func NewBusiness(log *logger.Logger, storer Storer) *Business {
	return &Business{
		log:    log,
		storer: storer,
	}
}

// NewWithTx constructs a new business value that will use the
// specified transaction in any store related calls.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	bus := Business{
		log:    b.log,
		storer: storer,
	}

	return &bus, nil
}

// Create adds a new org unit to the system
func (b *Business) Create(ctx context.Context, nou NewOrganizationalUnit) (OrganizationalUnit, error) {
	ctx, span := otel.AddSpan(ctx, "business.organizationalunitbus.create")
	defer span.End()

	ou := OrganizationalUnit{
		ID:                    uuid.New(),
		ParentID:              nou.ParentID,
		Name:                  nou.Name,
		Level:                 nou.Level,
		Path:                  nou.Path,
		CanInheritPermissions: nou.CanInheritPermissions,
		CanRollupData:         nou.CanRollupData,
		UnitType:              nou.UnitType,
		IsActive:              nou.IsActive,
	}

	if err := b.storer.Create(ctx, ou); err != nil {
		return OrganizationalUnit{}, fmt.Errorf("creating organizational unit: %w", err)
	}

	return ou, nil
}

// Update modifies a org unit in the system
func (b *Business) Update(ctx context.Context, ou OrganizationalUnit, uou UpdateOrganizationalUnit) (OrganizationalUnit, error) {
	ctx, span := otel.AddSpan(ctx, "business.organizationalroleunits.update")
	defer span.End()