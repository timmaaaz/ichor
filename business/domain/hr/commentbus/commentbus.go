package commentbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
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
	ErrNotFound              = errors.New("user status comment not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("user status comment entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, userApprovalComment UserApprovalComment) error
	Update(ctx context.Context, userApprovalComment UserApprovalComment) error
	Delete(ctx context.Context, userApprovalComment UserApprovalComment) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]UserApprovalComment, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, userApprovalCommentID uuid.UUID) (UserApprovalComment, error)
}

type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
	outbox   *outbox.Writer
	userbus  *userbus.Business
}

// NewBusiness constructs a new user status comment business API for use.
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, userbus *userbus.Business, storer Storer) *Business {
	return &Business{
		log:      log,
		delegate: delegate,
		storer:   storer,
		userbus:  userbus,
	}
}

// NewWithTx constructs a new business value that will use the
// specified transaction in any store related calls.
//
// This function seems like it could be implemented only once, but same with a lot of other pieces of this
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

	// Copy every field and override only the storer. The prior explicit-field
	// list silently dropped userbus (*userbus.Business), which Create
	// dereferences (SetUnderReview) — so a tx-bound commentbus nil-panicked.
	// Copy-then-override can't drop a dependency.
	nb := *b
	nb.storer = storer

	// Create calls userbus.SetUnderReview, which WRITES to core.users. Rebind the nested userbus
	// onto the same tx so that write joins the caller's transaction instead of autocommitting on
	// the base pool — otherwise a rolled-back form submit (e.g. via formdataapp) would leave the
	// comment rolled back but the user's status change orphaned.
	if b.userbus != nil {
		ub, err := b.userbus.NewWithTx(tx)
		if err != nil {
			return nil, err
		}
		nb.userbus = ub
	}

	return &nb, nil
}

// Create creates a new user status comment to the system.
func (b *Business) Create(ctx context.Context, nuac NewUserApprovalComment) (UserApprovalComment, error) {
	ctx, span := otel.AddSpan(ctx, "business.user.status.comment.Create")
	defer span.End()

	now := time.Now().UTC().Truncate(time.Second)
	if nuac.CreatedDate != nil {
		now = nuac.CreatedDate.Truncate(time.Second)
	}

	uac := UserApprovalComment{
		ID:          uuid.New(),
		Comment:     nuac.Comment,
		CommenterID: nuac.CommenterID,
		UserID:      nuac.UserID,
		CreatedDate: now,
	}

	if err := b.storer.Create(ctx, uac); err != nil {
		return UserApprovalComment{}, fmt.Errorf("store create: %w", err)
	}

	if err := b.userbus.SetUnderReview(ctx, nuac.UserID); err != nil {
		return UserApprovalComment{}, fmt.Errorf("userbus set under review: %w", err)
	}

	// Fire delegate event for workflow automation
	evtData := ActionCreatedData(uac)
	if err := b.outbox.Emit(ctx, evtData); err != nil {
		return UserApprovalComment{}, fmt.Errorf("emit cascade event: %w", err)
	}
	if err := b.delegate.Call(ctx, ActionCreatedData(uac)); err != nil {
		b.log.Error(ctx, "commentbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return uac, nil
}

// Update modifies information about an user status comment.
func (b *Business) Update(ctx context.Context, uac UserApprovalComment, uuac UpdateUserApprovalComment) (UserApprovalComment, error) {
	ctx, span := otel.AddSpan(ctx, "business.user.status.comment.Update")
	defer span.End()

	before := uac

	if uuac.Comment != nil {
		uac.Comment = *uuac.Comment
	}

	if err := b.storer.Update(ctx, uac); err != nil {
		return UserApprovalComment{}, fmt.Errorf("store update: %w", err)
	}

	// Fire delegate event for workflow automation
	evtData := ActionUpdatedData(before, uac)
	if err := b.outbox.Emit(ctx, evtData); err != nil {
		return UserApprovalComment{}, fmt.Errorf("emit cascade event: %w", err)
	}
	if err := b.delegate.Call(ctx, ActionUpdatedData(before, uac)); err != nil {
		b.log.Error(ctx, "commentbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return uac, nil
}

// Delete removes an user status comment from the system.
func (b *Business) Delete(ctx context.Context, uac UserApprovalComment) error {
	ctx, span := otel.AddSpan(ctx, "business.user.status.comment.Delete")
	defer span.End()

	if err := b.storer.Delete(ctx, uac); err != nil {
		return fmt.Errorf("store delete: %w", err)
	}

	// Fire delegate event for workflow automation
	evtData := ActionDeletedData(uac)
	if err := b.outbox.Emit(ctx, evtData); err != nil {
		return fmt.Errorf("emit cascade event: %w", err)
	}
	if err := b.delegate.Call(ctx, ActionDeletedData(uac)); err != nil {
		b.log.Error(ctx, "commentbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query returns a list of user status commentes
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]UserApprovalComment, error) {
	ctx, span := otel.AddSpan(ctx, "business.user.status.comment.Query")
	defer span.End()

	comment, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return comment, nil
}

// Count returns the total number of user status commentes
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.user.status.comment.Count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the user status comment by the specified ID.
func (b *Business) QueryByID(ctx context.Context, commentID uuid.UUID) (UserApprovalComment, error) {
	ctx, span := otel.AddSpan(ctx, "business.user.status.comment.QueryByID")
	defer span.End()

	comment, err := b.storer.QueryByID(ctx, commentID)
	if err != nil {
		return UserApprovalComment{}, fmt.Errorf("query: user status comment[%s]: %w", commentID, err)
	}

	return comment, nil
}
