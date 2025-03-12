package userorganizationbus

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	ErrUnique                = errors.New("user organization is not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, uo UserOrganization) error
	Update(ctx context.Context, uo UserOrganization) error
	Delete(ctx context.Context, uo UserOrganization) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]UserOrganization, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, userID uuid.UUID) (UserOrganization, error)
	QueryByUserID(ctx context.Context, userID uuid.UUID) (UserOrganization, error)
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

// Create adds a new user organization to the system
func (b *Business) Create(ctx context.Context, nuo NewUserOrganization) (UserOrganization, error) {
	ctx, span := otel.AddSpan(ctx, "business.userorganizationbus.create")
	defer span.End()

	uo := UserOrganization{
		ID:                   uuid.New(),
		UserID:               nuo.UserID,
		OrganizationalUnitID: nuo.OrganizationalUnitID,
		RoleID:               nuo.RoleID,
		IsUnitManager:        nuo.IsUnitManager,
		StartDate:            nuo.StartDate,
		EndDate:              nuo.EndDate,
		CreatedBy:            nuo.CreatedBy,
		CreatedAt:            time.Now(),
	}

	if err := b.storer.Create(ctx, uo); err != nil {

		return UserOrganization{}, fmt.Errorf("create: %w", err)
	}

	return uo, nil
}

// Update modifies a user organization in the system
func (b *Business) Update(ctx context.Context, uo UserOrganization, uuo UpdateUserOrganization) (UserOrganization, error) {
	ctx, span := otel.AddSpan(ctx, "business.userorganizationbus.update")
	defer span.End()

	err := convert.PopulateSameTypes(uuo, &uo)
	if err != nil {
		return UserOrganization{}, fmt.Errorf("update: %w", err)
	}

	if err := b.storer.Update(ctx, uo); err != nil {
		return UserOrganization{}, fmt.Errorf("update: %w", err)
	}

	return uo, nil
}

// Delete removes a user organization from the system
func (b *Business) Delete(ctx context.Context, uo UserOrganization) error {
	ctx, span := otel.AddSpan(ctx, "business.userorganizationbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, uo); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of user organizations from the system
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]UserOrganization, error) {
	ctx, span := otel.AddSpan(ctx, "business.userorganizationbus.query")
	defer span.End()

	uos, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return uos, nil
}

// Count returns the total number of user organizations
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.userorganizationbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID retrieves a user organization by its ID
func (b *Business) QueryByID(ctx context.Context, uoID uuid.UUID) (UserOrganization, error) {
	ctx, span := otel.AddSpan(ctx, "business.userorganizationbus.querybyid")
	defer span.End()

	uo, err := b.storer.QueryByID(ctx, uoID)
	if err != nil {
		return UserOrganization{}, fmt.Errorf("query by id: id[%s]: %w", uoID, err)
	}

	return uo, nil
}

// QueryByUserID retrieves a list of user organizations by user ID
func (b *Business) QueryByUserID(ctx context.Context, userID uuid.UUID) (UserOrganization, error) {
	ctx, span := otel.AddSpan(ctx, "business.userorganizationbus.querybyuserid")
	defer span.End()

	uos, err := b.storer.QueryByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return UserOrganization{}, ErrNotFound
		}
		return UserOrganization{}, fmt.Errorf("query by user id: id[%s]: %w", userID, err)
	}

	return uos, nil
}
