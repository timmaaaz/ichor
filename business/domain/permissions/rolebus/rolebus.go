package rolebus

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
	ErrNotFound              = errors.New("role not found")
	ErrUnique                = errors.New("not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, role Role) error
	Update(ctx context.Context, role Role) error
	Delete(ctx context.Context, role Role) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Role, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, userID uuid.UUID) (Role, error)
	QueryAll(ctx context.Context) ([]Role, error)
}

// Business manages the set of APIs for user access.
type Business struct {
	log    *logger.Logger
	storer Storer
	del    *delegate.Delegate
}

// NewBusiness constructs a user business API for use.
func NewBusiness(log *logger.Logger, del *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:    log,
		del:    del,
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

// Create adds a new role to the system.
func (b *Business) Create(ctx context.Context, nr NewRole) (Role, error) {
	ctx, span := otel.AddSpan(ctx, "business.rolebus.create")
	defer span.End()

	role := Role{
		ID:          uuid.New(),
		Name:        nr.Name,
		Description: nr.Description,
	}

	if err := b.storer.Create(ctx, role); err != nil {
		return Role{}, fmt.Errorf("creating role: %w", err)
	}

	return role, nil
}

// Update modifies a role in the system.
func (b *Business) Update(ctx context.Context, role Role, ur UpdateRole) (Role, error) {
	ctx, span := otel.AddSpan(ctx, "business.userbus.update")
	defer span.End()

	err := convert.PopulateSameTypes(ur, &role)
	if err != nil {
		return Role{}, fmt.Errorf("populate same types: %w", err)
	}

	if err := b.storer.Update(ctx, role); err != nil {
		return Role{}, fmt.Errorf("updating role: %w", err)
	}

	// Inform permissions, need to clear cache
	if err := b.del.Call(ctx, ActionUpdatedData(role)); err != nil {
		return Role{}, fmt.Errorf("calling delegate: %w", err)
	}

	return role, nil
}

// Delete removes a role from the system.
func (b *Business) Delete(ctx context.Context, role Role) error {
	ctx, span := otel.AddSpan(ctx, "business.rolebus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, role); err != nil {
		return fmt.Errorf("deleting role: %w", err)
	}

	// Inform permissions, need to clear cache
	if err := b.del.Call(ctx, ActionDeletedData(role.ID)); err != nil {
		return fmt.Errorf("calling delegate: %w", err)
	}

	return nil
}

// Query retrieves a list of roles from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Role, error) {
	ctx, span := otel.AddSpan(ctx, "business.rolebus.query")
	defer span.End()

	roles, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("querying roles: %w", err)
	}

	return roles, nil
}

// Count returns the total number of roles.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.rolebus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the role by the specified ID.
func (b *Business) QueryByID(ctx context.Context, roleID uuid.UUID) (Role, error) {
	ctx, span := otel.AddSpan(ctx, "business.rolebus.querybyid")
	defer span.End()

	role, err := b.storer.QueryByID(ctx, roleID)
	if err != nil {
		return Role{}, fmt.Errorf("querying role: roleID[%s]: %w", roleID, err)
	}

	return role, nil
}

// QueryAll retrieves all roles from the system.
func (b *Business) QueryAll(ctx context.Context) ([]Role, error) {
	ctx, span := otel.AddSpan(ctx, "business.rolebus.queryall")
	defer span.End()

	roles, err := b.storer.QueryAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("querying all roles: %w", err)
	}

	return roles, nil
}
