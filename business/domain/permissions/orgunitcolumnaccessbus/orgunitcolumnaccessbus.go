package orgunitcolumnaccessbus

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/convert"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("role not found")
	ErrUnique                = errors.New("org unit column access organization is not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrColumnNotExists       = errors.New("column does not exist")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Exists(ctx context.Context, ouca OrgUnitColumnAccess) error
	Create(ctx context.Context, ouca OrgUnitColumnAccess) error
	Update(ctx context.Context, ouca OrgUnitColumnAccess) error
	Delete(ctx context.Context, ouca OrgUnitColumnAccess) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]OrgUnitColumnAccess, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, id uuid.UUID) (OrgUnitColumnAccess, error)
}

// Business manages the set of APIs for org unit column access.
type Business struct {
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs a org unit column access business API for use.
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

// Create adds a new org unit column access to the system
func (b *Business) Create(ctx context.Context, new NewOrgUnitColumnAccess) (OrgUnitColumnAccess, error) {
	ctx, span := otel.AddSpan(ctx, "business.orgunitcolumnaccess.create")
	defer span.End()

	ouca := OrgUnitColumnAccess{
		ID:                    uuid.New(),
		OrganizationalUnitID:  new.OrganizationalUnitID,
		TableName:             new.TableName,
		ColumnName:            new.ColumnName,
		CanRead:               new.CanRead,
		CanUpdate:             new.CanUpdate,
		CanInheritPermissions: new.CanInheritPermissions,
		CanRollupData:         new.CanRollupData,
	}

	if err := b.storer.Exists(ctx, ouca); err != nil {
		return OrgUnitColumnAccess{}, fmt.Errorf("checking if org unit column access column exists: %w", err)
	}

	if err := b.storer.Create(ctx, ouca); err != nil {
		if errors.Is(err, ErrUnique) {
			return OrgUnitColumnAccess{}, fmt.Errorf("create: %w", ErrUnique)
		}
		return OrgUnitColumnAccess{}, err
	}

	return ouca, nil
}

// Update makes changes to a org unit column access in the system
func (b *Business) Update(ctx context.Context, ouca OrgUnitColumnAccess, update UpdateOrgUnitColumnAccess) (OrgUnitColumnAccess, error) {
	ctx, span := otel.AddSpan(ctx, "business.orgunitcolumnaccess.update")
	defer span.End()

	err := convert.PopulateSameTypes(update, &ouca)
	if err != nil {
		return OrgUnitColumnAccess{}, fmt.Errorf("update: %w", err)
	}

	if err := b.storer.Update(ctx, ouca); err != nil {
		return OrgUnitColumnAccess{}, fmt.Errorf("update: %w", err)
	}

	return ouca, nil
}

// Delete removes a org unit column access from the system
func (b *Business) Delete(ctx context.Context, ouca OrgUnitColumnAccess) error {
	ctx, span := otel.AddSpan(ctx, "business.orgunitcolumnaccess.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, ouca); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of org unit column access from the system
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]OrgUnitColumnAccess, error) {
	ctx, span := otel.AddSpan(ctx, "business.orgunitcolumnaccess.query")
	defer span.End()

	oucas, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return oucas, nil
}

// Count retrieves the total number of org unit column access from the system
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.orgunitcolumnaccess.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID retrieves a org unit column access by its ID
func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (OrgUnitColumnAccess, error) {
	ctx, span := otel.AddSpan(ctx, "business.orgunitcolumnaccess.querybyid")
	defer span.End()

	ouca, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return OrgUnitColumnAccess{}, fmt.Errorf("query by id: %w", err)
	}

	return ouca, nil
}
