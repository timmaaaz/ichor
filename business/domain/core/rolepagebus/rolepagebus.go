package rolepagebus

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/outbox"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("role page not found")
	ErrUnique                = errors.New("not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, rolePage RolePage) error
	Update(ctx context.Context, rolePage RolePage) error
	Delete(ctx context.Context, rolePage RolePage) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]RolePage, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, rolePageID uuid.UUID) (RolePage, error)
}

// Business manages the set of APIs for role page access.
type Business struct {
	log    *logger.Logger
	storer Storer
	del    *delegate.Delegate
	outbox *outbox.Writer
}

// NewBusiness constructs a role page business API for use.
func NewBusiness(log *logger.Logger, del *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:    log,
		del:    del,
		storer: storer,
	}
}

// NewWithTx constructs a new business value that will use the
// specified transaction in any store related calls.
// WithOutbox returns a copy of the Business wired to the cascade outbox Writer.
// Inert until the Writer is injected at the F2 cutover (nil Writer -> Emit no-ops).
func (b *Business) WithOutbox(w *outbox.Writer) *Business {
	nb := *b
	nb.outbox = w
	return &nb
}

func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	nb := *b
	nb.storer = storer
	return &nb, nil
}

// Create adds a new role page mapping to the system.
func (b *Business) Create(ctx context.Context, nrp NewRolePage) (RolePage, error) {
	ctx, span := otel.AddSpan(ctx, "business.rolepagebus.create")
	defer span.End()

	return outbox.WriteAtomic(ctx, b.outbox, b, (*Business).NewWithTx,
		func(ctx context.Context, b *Business) (RolePage, error) {
			rolePage := RolePage{
				ID:        uuid.New(),
				RoleID:    nrp.RoleID,
				PageID:    nrp.PageID,
				CanAccess: nrp.CanAccess,
			}

			if err := b.storer.Create(ctx, rolePage); err != nil {
				return RolePage{}, fmt.Errorf("creating role page: %w", err)
			}

			// Fire delegate event for workflow automation
			evtData := ActionCreatedData(rolePage)
			if err := b.outbox.Emit(ctx, evtData); err != nil {
				return RolePage{}, fmt.Errorf("emit cascade event: %w", err)
			}
			if err := b.del.Call(ctx, ActionCreatedData(rolePage)); err != nil {
				b.log.Error(ctx, "rolepagebus: delegate call failed", "action", ActionCreated, "err", err)
			}

			return rolePage, nil
		})
}

// Update modifies a role page mapping in the system.
func (b *Business) Update(ctx context.Context, rolePage RolePage, urp UpdateRolePage) (RolePage, error) {
	ctx, span := otel.AddSpan(ctx, "business.rolepagebus.update")
	defer span.End()

	return outbox.WriteAtomic(ctx, b.outbox, b, (*Business).NewWithTx,
		func(ctx context.Context, b *Business) (RolePage, error) {
			before := rolePage

			if urp.CanAccess != nil {
				rolePage.CanAccess = *urp.CanAccess
			}

			if err := b.storer.Update(ctx, rolePage); err != nil {
				return RolePage{}, fmt.Errorf("updating role page: %w", err)
			}

			// Fire delegate event for workflow automation
			evtData := ActionUpdatedData(before, rolePage)
			if err := b.outbox.Emit(ctx, evtData); err != nil {
				return RolePage{}, fmt.Errorf("emit cascade event: %w", err)
			}
			if err := b.del.Call(ctx, ActionUpdatedData(before, rolePage)); err != nil {
				b.log.Error(ctx, "rolepagebus: delegate call failed", "action", ActionUpdated, "err", err)
			}

			return rolePage, nil
		})
}

// Delete removes a role page mapping from the system.
func (b *Business) Delete(ctx context.Context, rolePage RolePage) error {
	ctx, span := otel.AddSpan(ctx, "business.rolepagebus.delete")
	defer span.End()

	return outbox.WriteAtomicVoid(ctx, b.outbox, b, (*Business).NewWithTx,
		func(ctx context.Context, b *Business) error {
			if err := b.storer.Delete(ctx, rolePage); err != nil {
				return fmt.Errorf("deleting role page: %w", err)
			}

			// Fire delegate event for workflow automation
			evtData := ActionDeletedData(rolePage)
			if err := b.outbox.Emit(ctx, evtData); err != nil {
				return fmt.Errorf("emit cascade event: %w", err)
			}
			if err := b.del.Call(ctx, ActionDeletedData(rolePage)); err != nil {
				b.log.Error(ctx, "rolepagebus: delegate call failed", "action", ActionDeleted, "err", err)
			}

			return nil
		})
}

// Query retrieves a list of role page mappings from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]RolePage, error) {
	ctx, span := otel.AddSpan(ctx, "business.rolepagebus.query")
	defer span.End()

	rolePages, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("querying role pages: %w", err)
	}

	return rolePages, nil
}

// Count returns the total number of role page mappings.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.rolepagebus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the role page mapping by the specified ID.
func (b *Business) QueryByID(ctx context.Context, rolePageID uuid.UUID) (RolePage, error) {
	ctx, span := otel.AddSpan(ctx, "business.rolepagebus.querybyid")
	defer span.End()

	rolePage, err := b.storer.QueryByID(ctx, rolePageID)
	if err != nil {
		return RolePage{}, fmt.Errorf("querying role page: rolePageID[%s]: %w", rolePageID, err)
	}

	return rolePage, nil
}
