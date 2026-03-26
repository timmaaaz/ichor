package notificationbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound = errors.New("notification not found")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, notification Notification) error
	QueryByID(ctx context.Context, id uuid.UUID) (Notification, error)
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, pg page.Page) ([]Notification, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	MarkAsRead(ctx context.Context, id uuid.UUID, readDate time.Time) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID, readDate time.Time) (int, error)
}

// Business manages notification operations.
type Business struct {
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs a notification business API for use.
func NewBusiness(log *logger.Logger, storer Storer) *Business {
	return &Business{
		log:    log,
		storer: storer,
	}
}

// NewWithTx constructs a new Business value replacing the Storer with a
// Storer that uses the specified transaction.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	return &Business{
		log:    b.log,
		storer: storer,
	}, nil
}

// Create adds a new notification to the system.
func (b *Business) Create(ctx context.Context, nn NewNotification) (Notification, error) {
	ctx, span := otel.AddSpan(ctx, "business.notificationbus.create")
	defer span.End()

	now := time.Now()

	notification := Notification{
		ID:               uuid.New(),
		UserID:           nn.UserID,
		Title:            nn.Title,
		Message:          nn.Message,
		Priority:         nn.Priority,
		IsRead:           false,
		SourceEntityName: nn.SourceEntityName,
		SourceEntityID:   nn.SourceEntityID,
		ActionURL:        nn.ActionURL,
		CreatedDate:      now,
	}

	if err := b.storer.Create(ctx, notification); err != nil {
		return Notification{}, fmt.Errorf("create: %w", err)
	}

	return notification, nil
}

// QueryByID finds the notification by the specified ID.
func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (Notification, error) {
	ctx, span := otel.AddSpan(ctx, "business.notificationbus.querybyid")
	defer span.End()

	notification, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return Notification{}, fmt.Errorf("query: notificationID[%s]: %w", id, err)
	}

	return notification, nil
}

// Query retrieves a list of notifications from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pg page.Page) ([]Notification, error) {
	ctx, span := otel.AddSpan(ctx, "business.notificationbus.query")
	defer span.End()

	notifications, err := b.storer.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return notifications, nil
}

// Count returns the total number of notifications matching the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.notificationbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// MarkAsRead marks a single notification as read.
func (b *Business) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	ctx, span := otel.AddSpan(ctx, "business.notificationbus.markasread")
	defer span.End()

	if err := b.storer.MarkAsRead(ctx, id, time.Now()); err != nil {
		return fmt.Errorf("markasread: notificationID[%s]: %w", id, err)
	}

	return nil
}

// MarkAllAsRead marks all unread notifications for a user as read.
func (b *Business) MarkAllAsRead(ctx context.Context, userID uuid.UUID) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.notificationbus.markallasread")
	defer span.End()

	count, err := b.storer.MarkAllAsRead(ctx, userID, time.Now())
	if err != nil {
		return 0, fmt.Errorf("markallasread: userID[%s]: %w", userID, err)
	}

	return count, nil
}
