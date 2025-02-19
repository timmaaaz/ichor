package userrolebus

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
	ErrUniqueEmail           = errors.New("email is not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, ur UserRole) error
	Update(ctx context.Context, ur UserRole) error
	Delete(ctx context.Context, ur UserRole) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]UserRole, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, userID uuid.UUID) (UserRole, error)
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

// Create adds a new user role to the system
func (b *Business) Create(ctx context.Context, nur NewUserRole) (UserRole, error) {
	ctx, span := otel.AddSpan(ctx, "business.userrolebus.create")
	defer span.End()

	ur := UserRole{
		ID:     uuid.New(),
		UserID: nur.UserID,
		RoleID: nur.RoleID,
	}

	if err := b.storer.Create(ctx, ur); err != nil {
		return UserRole{}, err
	}

	return ur, nil
}

// Update modifies a user role in the system
func (b *Business) Update(ctx context.Context, ur UserRole, uur UpdateUserRole) error {
	ctx, span := otel.AddSpan(ctx, "business.userrolebus.update")
	defer span.End()

	if uur.RoleID != nil {
		ur.RoleID = *uur.RoleID
	}

	if err := b.storer.Update(ctx, ur); err != nil {
		return fmt.Errorf("updating role: %w", err)
	}

	return nil
}

// Delete removes a user role from the system.
func (b *Business) Delete(ctx context.Context, ur UserRole) error {
	ctx, span := otel.AddSpan(ctx, "business.userrolebus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, ur); err != nil {
		return fmt.Errorf("deleting user role: %w", err)
	}

	return nil
}

// Query retrieves a list of user roles from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]UserRole, error) {
	ctx, span := otel.AddSpan(ctx, "business.userrolebus.query")
	defer span.End()

	urs, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("querying user roles: %w", err)
	}

	return urs, nil
}

// Count returns the total number of user roles.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.userrolebus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the user role by the specified ID.
func (b *Business) QueryByID(ctx context.Context, urID uuid.UUID) (UserRole, error) {
	ctx, span := otel.AddSpan(ctx, "business.userrolebus.querybyid")
	defer span.End()

	ur, err := b.storer.QueryByID(ctx, urID)
	if err != nil {
		return UserRole{}, fmt.Errorf("querying role: userRoleID[%s]: %w", urID, err)
	}

	return ur, nil
}
