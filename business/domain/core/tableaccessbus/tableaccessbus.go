package tableaccessbus

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
	ErrNotFound              = errors.New("role not found")
	ErrUnique                = errors.New("table access is not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrNonexistentTableName  = errors.New("table does not exist")
)

// VirtualTables defines tables that exist for permission purposes but don't
// have actual database tables. These tables bypass database existence validation
// while still maintaining full RBAC permission checks.
var VirtualTables = map[string]bool{
	"introspection": true, // Database metadata introspection endpoints
}

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, ta TableAccess) error
	Update(ctx context.Context, ta TableAccess) error
	Delete(ctx context.Context, ta TableAccess) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]TableAccess, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, userID uuid.UUID) (TableAccess, error)
	QueryByRoleIDs(ctx context.Context, roleIDs []uuid.UUID) ([]TableAccess, error)
	QueryAll(ctx context.Context) ([]TableAccess, error)
}

// Business manages the set of APIs for user access.
type Business struct {
	log    *logger.Logger
	del    *delegate.Delegate
	storer Storer
}

// NewBusiness constructs a user business API for use.
func NewBusiness(log *logger.Logger, del *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:    log,
		storer: storer,
		del:    del,
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

// Create adds a new table access to the system
func (b *Business) Create(ctx context.Context, nta NewTableAccess) (TableAccess, error) {
	ctx, span := otel.AddSpan(ctx, "business.tableaccess.create")
	defer span.End()

	ta := TableAccess{
		ID:        uuid.New(),
		RoleID:    nta.RoleID,
		TableName: nta.TableName,
		CanCreate: nta.CanCreate,
		CanRead:   nta.CanRead,
		CanUpdate: nta.CanUpdate,
		CanDelete: nta.CanDelete,
	}

	if err := b.storer.Create(ctx, ta); err != nil {
		return TableAccess{}, fmt.Errorf("creating table access: %w", err)
	}

	return ta, nil
}

// Update modifies a table access in the system
func (b *Business) Update(ctx context.Context, ta TableAccess, uta UpdateTableAccess) (TableAccess, error) {
	ctx, span := otel.AddSpan(ctx, "business.tableaccess.update")
	defer span.End()

	if uta.RoleID != nil {
		ta.RoleID = *uta.RoleID
	}
	if uta.TableName != nil {
		ta.TableName = *uta.TableName
	}
	if uta.CanCreate != nil {
		ta.CanCreate = *uta.CanCreate
	}
	if uta.CanRead != nil {
		ta.CanRead = *uta.CanRead
	}
	if uta.CanUpdate != nil {
		ta.CanUpdate = *uta.CanUpdate
	}
	if uta.CanDelete != nil {
		ta.CanDelete = *uta.CanDelete
	}

	if err := b.storer.Update(ctx, ta); err != nil {
		return TableAccess{}, fmt.Errorf("updating table access: %w", err)
	}

	// Inform permissions, need to clear cache
	if err := b.del.Call(ctx, ActionUpdatedData(ta)); err != nil {
		return TableAccess{}, fmt.Errorf("calling delegate: %w", err)
	}

	return ta, nil
}

// Delete removes a table access from the system
func (b *Business) Delete(ctx context.Context, ta TableAccess) error {
	ctx, span := otel.AddSpan(ctx, "business.tableaccess.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, ta); err != nil {
		return fmt.Errorf("deleting table access: %w", err)
	}

	// Inform permissions, need to clear cache
	if err := b.del.Call(ctx, ActionDeletedData(ta.ID)); err != nil {
		return fmt.Errorf("calling delegate: %w", err)
	}

	return nil
}

// Query retrieves a list of table accesses from the system
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]TableAccess, error) {
	ctx, span := otel.AddSpan(ctx, "business.tableaccess.query")
	defer span.End()

	ta, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("querying table access: %w", err)
	}

	return ta, nil
}

// Count returns the number of table accesses in the system
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.tableaccess.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID retrieves a table access by its ID
func (b *Business) QueryByID(ctx context.Context, tableAccessID uuid.UUID) (TableAccess, error) {
	ctx, span := otel.AddSpan(ctx, "business.tableaccess.querybyid")
	defer span.End()

	ta, err := b.storer.QueryByID(ctx, tableAccessID)
	if err != nil {
		return TableAccess{}, fmt.Errorf("querying table access by ID: %w", err)
	}

	return ta, nil
}

// QueryByRoleIDs retrieves a list of table accesses by their role IDs
func (b *Business) QueryByRoleIDs(ctx context.Context, roleIDs []uuid.UUID) ([]TableAccess, error) {
	ctx, span := otel.AddSpan(ctx, "business.tableaccess.QueryByRoleIDs")
	defer span.End()

	ta, err := b.storer.QueryByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, fmt.Errorf("querying table access by role IDs: %w", err)
	}

	return ta, nil
}

// QueryAll retrieves all table accesses from the system
func (b *Business) QueryAll(ctx context.Context) ([]TableAccess, error) {
	ctx, span := otel.AddSpan(ctx, "business.tableaccess.QueryAll")
	defer span.End()

	ta, err := b.storer.QueryAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("querying all table access: %w", err)
	}

	return ta, nil
}
