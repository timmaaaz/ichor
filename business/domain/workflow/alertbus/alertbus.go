// Package alertbus provides the core business logic for workflow alerts.
package alertbus

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
	ErrNotFound     = errors.New("alert not found")
	ErrNotRecipient = errors.New("user is not a recipient of this alert")
	ErrAlreadyAcked = errors.New("alert already acknowledged by this user")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, alert Alert) error
	CreateRecipients(ctx context.Context, recipients []AlertRecipient) error
	CreateAcknowledgment(ctx context.Context, ack AlertAcknowledgment) error
	QueryByID(ctx context.Context, alertID uuid.UUID) (Alert, error)
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, pg page.Page) ([]Alert, error)
	QueryByUserID(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, filter QueryFilter, orderBy order.By, pg page.Page) ([]Alert, error)
	QueryRecipientsByAlertID(ctx context.Context, alertID uuid.UUID) ([]AlertRecipient, error)
	QueryRecipientsByAlertIDs(ctx context.Context, alertIDs []uuid.UUID) (map[uuid.UUID][]AlertRecipient, error)
	QueryAcknowledgmentsByAlertID(ctx context.Context, alertID uuid.UUID) ([]AlertAcknowledgment, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	CountByUserID(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, filter QueryFilter) (int, error)
	CountByUserIDGroupedBySeverity(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, status string) (map[string]int, error)
	UpdateStatus(ctx context.Context, alertID uuid.UUID, status string, now time.Time) error
	IsRecipient(ctx context.Context, alertID, userID uuid.UUID, roleIDs []uuid.UUID) (bool, error)
	FilterRecipientAlerts(ctx context.Context, alertIDs []uuid.UUID, userID uuid.UUID, roleIDs []uuid.UUID) ([]uuid.UUID, error)
	QueryActiveByUserID(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) ([]uuid.UUID, error)
	AcknowledgeMultiple(ctx context.Context, alertIDs []uuid.UUID, userID uuid.UUID, notes string, now time.Time) (int, error)
	DismissMultiple(ctx context.Context, alertIDs []uuid.UUID, now time.Time) (int, error)
	ResolveRelatedAlerts(ctx context.Context, sourceEntityID uuid.UUID, alertType string, excludeAlertID uuid.UUID, now time.Time) (int, error)
}

// Business manages alert operations.
type Business struct {
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs an alert business API for use.
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

// Create creates a new alert (called by CreateAlertHandler).
func (b *Business) Create(ctx context.Context, alert Alert) error {
	ctx, span := otel.AddSpan(ctx, "business.alertbus.create")
	defer span.End()

	if err := b.storer.Create(ctx, alert); err != nil {
		return fmt.Errorf("create alert: %w", err)
	}

	return nil
}

// CreateRecipients adds multiple recipients to an alert (batch insert).
func (b *Business) CreateRecipients(ctx context.Context, recipients []AlertRecipient) error {
	ctx, span := otel.AddSpan(ctx, "business.alertbus.createrecipients")
	defer span.End()

	if len(recipients) == 0 {
		return nil
	}

	if err := b.storer.CreateRecipients(ctx, recipients); err != nil {
		return fmt.Errorf("create recipients: %w", err)
	}

	return nil
}

// QueryByID returns a single alert by ID.
func (b *Business) QueryByID(ctx context.Context, alertID uuid.UUID) (Alert, error) {
	ctx, span := otel.AddSpan(ctx, "business.alertbus.querybyid")
	defer span.End()

	alert, err := b.storer.QueryByID(ctx, alertID)
	if err != nil {
		return Alert{}, fmt.Errorf("query alert: alertID[%s]: %w", alertID, err)
	}

	return alert, nil
}

// QueryRecipientsByAlertID returns all recipients for a given alert.
func (b *Business) QueryRecipientsByAlertID(ctx context.Context, alertID uuid.UUID) ([]AlertRecipient, error) {
	ctx, span := otel.AddSpan(ctx, "business.alertbus.queryrecipientsbyalertid")
	defer span.End()

	recipients, err := b.storer.QueryRecipientsByAlertID(ctx, alertID)
	if err != nil {
		return nil, fmt.Errorf("query recipients: alertID[%s]: %w", alertID, err)
	}

	return recipients, nil
}

// QueryRecipientsByAlertIDs returns recipients for multiple alerts, keyed by alert ID.
func (b *Business) QueryRecipientsByAlertIDs(ctx context.Context, alertIDs []uuid.UUID) (map[uuid.UUID][]AlertRecipient, error) {
	ctx, span := otel.AddSpan(ctx, "business.alertbus.queryrecipientsbyalertids")
	defer span.End()

	result, err := b.storer.QueryRecipientsByAlertIDs(ctx, alertIDs)
	if err != nil {
		return nil, fmt.Errorf("query recipients: %w", err)
	}

	return result, nil
}

// Query returns all alerts (admin only - no recipient filtering).
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pg page.Page) ([]Alert, error) {
	ctx, span := otel.AddSpan(ctx, "business.alertbus.query")
	defer span.End()

	alerts, err := b.storer.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return nil, fmt.Errorf("query alerts: %w", err)
	}

	return alerts, nil
}

// Count returns count of all alerts (admin only).
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.alertbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryMine returns alerts for a user (directly or via roles).
func (b *Business) QueryMine(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, filter QueryFilter, orderBy order.By, pg page.Page) ([]Alert, error) {
	ctx, span := otel.AddSpan(ctx, "business.alertbus.querymine")
	defer span.End()

	alerts, err := b.storer.QueryByUserID(ctx, userID, roleIDs, filter, orderBy, pg)
	if err != nil {
		return nil, fmt.Errorf("query user alerts: userID[%s]: %w", userID, err)
	}

	return alerts, nil
}

// IsRecipient checks whether a user (by user ID or role membership) is a recipient of an alert.
func (b *Business) IsRecipient(ctx context.Context, alertID, userID uuid.UUID, roleIDs []uuid.UUID) (bool, error) {
	ctx, span := otel.AddSpan(ctx, "business.alertbus.isrecipient")
	defer span.End()

	return b.storer.IsRecipient(ctx, alertID, userID, roleIDs)
}

// CountMine returns count of alerts for a user.
func (b *Business) CountMine(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.alertbus.countmine")
	defer span.End()

	return b.storer.CountByUserID(ctx, userID, roleIDs, filter)
}

// Acknowledge marks an alert as acknowledged by a user.
// Validates that user is a recipient before allowing acknowledgment.
func (b *Business) Acknowledge(ctx context.Context, alertID, userID uuid.UUID, roleIDs []uuid.UUID, notes string, now time.Time) (Alert, error) {
	ctx, span := otel.AddSpan(ctx, "business.alertbus.acknowledge")
	defer span.End()

	// Security check: verify user is a recipient
	isRecipient, err := b.storer.IsRecipient(ctx, alertID, userID, roleIDs)
	if err != nil {
		return Alert{}, fmt.Errorf("check recipient: %w", err)
	}
	if !isRecipient {
		return Alert{}, ErrNotRecipient
	}

	ack := AlertAcknowledgment{
		ID:               uuid.New(),
		AlertID:          alertID,
		AcknowledgedBy:   userID,
		AcknowledgedDate: now,
		Notes:            notes,
	}

	if err := b.storer.CreateAcknowledgment(ctx, ack); err != nil {
		return Alert{}, fmt.Errorf("create acknowledgment: %w", err)
	}

	if err := b.storer.UpdateStatus(ctx, alertID, StatusAcknowledged, now); err != nil {
		return Alert{}, fmt.Errorf("update status: %w", err)
	}

	return b.storer.QueryByID(ctx, alertID)
}

// Dismiss marks an alert as dismissed by a user.
// Validates that user is a recipient before allowing dismissal.
func (b *Business) Dismiss(ctx context.Context, alertID, userID uuid.UUID, roleIDs []uuid.UUID, now time.Time) (Alert, error) {
	ctx, span := otel.AddSpan(ctx, "business.alertbus.dismiss")
	defer span.End()

	// Security check: verify user is a recipient
	isRecipient, err := b.storer.IsRecipient(ctx, alertID, userID, roleIDs)
	if err != nil {
		return Alert{}, fmt.Errorf("check recipient: %w", err)
	}
	if !isRecipient {
		return Alert{}, ErrNotRecipient
	}

	if err := b.storer.UpdateStatus(ctx, alertID, StatusDismissed, now); err != nil {
		return Alert{}, fmt.Errorf("update status: %w", err)
	}

	return b.storer.QueryByID(ctx, alertID)
}

// AcknowledgeSelected acknowledges specific alerts by ID.
// Returns count of acknowledged alerts and count of skipped (non-recipient) alerts.
func (b *Business) AcknowledgeSelected(ctx context.Context, alertIDs []uuid.UUID, userID uuid.UUID, roleIDs []uuid.UUID, notes string, now time.Time) (count, skipped int, err error) {
	ctx, span := otel.AddSpan(ctx, "business.alertbus.acknowledgeselected")
	defer span.End()

	if len(alertIDs) == 0 {
		return 0, 0, nil
	}

	// Filter to only alerts user can access
	validIDs, err := b.storer.FilterRecipientAlerts(ctx, alertIDs, userID, roleIDs)
	if err != nil {
		return 0, 0, fmt.Errorf("filter recipient alerts: %w", err)
	}

	skipped = len(alertIDs) - len(validIDs)
	if len(validIDs) == 0 {
		return 0, skipped, nil
	}

	// Bulk acknowledge
	count, err = b.storer.AcknowledgeMultiple(ctx, validIDs, userID, notes, now)
	if err != nil {
		return 0, 0, fmt.Errorf("acknowledge multiple: %w", err)
	}

	return count, skipped, nil
}

// AcknowledgeAll acknowledges all active alerts for a user.
// Returns count of acknowledged alerts.
func (b *Business) AcknowledgeAll(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, notes string, now time.Time) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.alertbus.acknowledgeall")
	defer span.End()

	alertIDs, err := b.storer.QueryActiveByUserID(ctx, userID, roleIDs)
	if err != nil {
		return 0, fmt.Errorf("query active alerts: %w", err)
	}

	if len(alertIDs) == 0 {
		return 0, nil
	}

	count, err := b.storer.AcknowledgeMultiple(ctx, alertIDs, userID, notes, now)
	if err != nil {
		return 0, fmt.Errorf("acknowledge multiple: %w", err)
	}

	return count, nil
}

// DismissSelected dismisses specific alerts by ID.
// Returns count of dismissed alerts and count of skipped (non-recipient) alerts.
func (b *Business) DismissSelected(ctx context.Context, alertIDs []uuid.UUID, userID uuid.UUID, roleIDs []uuid.UUID, now time.Time) (count, skipped int, err error) {
	ctx, span := otel.AddSpan(ctx, "business.alertbus.dismissselected")
	defer span.End()

	if len(alertIDs) == 0 {
		return 0, 0, nil
	}

	// Filter to only alerts user can access
	validIDs, err := b.storer.FilterRecipientAlerts(ctx, alertIDs, userID, roleIDs)
	if err != nil {
		return 0, 0, fmt.Errorf("filter recipient alerts: %w", err)
	}

	skipped = len(alertIDs) - len(validIDs)
	if len(validIDs) == 0 {
		return 0, skipped, nil
	}

	// Bulk dismiss
	count, err = b.storer.DismissMultiple(ctx, validIDs, now)
	if err != nil {
		return 0, 0, fmt.Errorf("dismiss multiple: %w", err)
	}

	return count, skipped, nil
}

// DismissAll dismisses all active alerts for a user.
// Returns count of dismissed alerts.
func (b *Business) DismissAll(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, now time.Time) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.alertbus.dismissall")
	defer span.End()

	alertIDs, err := b.storer.QueryActiveByUserID(ctx, userID, roleIDs)
	if err != nil {
		return 0, fmt.Errorf("query active alerts: %w", err)
	}

	if len(alertIDs) == 0 {
		return 0, nil
	}

	count, err := b.storer.DismissMultiple(ctx, alertIDs, now)
	if err != nil {
		return 0, fmt.Errorf("dismiss multiple: %w", err)
	}

	return count, nil
}

// QueryAcknowledgmentsByAlertID returns all acknowledgments for an alert, enriched with acknowledger names.
func (b *Business) QueryAcknowledgmentsByAlertID(ctx context.Context, alertID uuid.UUID) ([]AlertAcknowledgment, error) {
	ctx, span := otel.AddSpan(ctx, "business.alertbus.queryacknowledgmentsbyalertid")
	defer span.End()

	acks, err := b.storer.QueryAcknowledgmentsByAlertID(ctx, alertID)
	if err != nil {
		return nil, fmt.Errorf("query acknowledgments: alertID[%s]: %w", alertID, err)
	}

	return acks, nil
}

// CountMineBySeverity returns a map of severity → count for the user's active (non-expired) alerts.
func (b *Business) CountMineBySeverity(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, status string) (map[string]int, error) {
	ctx, span := otel.AddSpan(ctx, "business.alertbus.countminebyseverity")
	defer span.End()

	return b.storer.CountByUserIDGroupedBySeverity(ctx, userID, roleIDs, status)
}

// ResolveRelatedAlerts marks prior active/acknowledged alerts as resolved when a new
// success alert is created for the same source entity and alert type.
// Returns the count of resolved alerts.
func (b *Business) ResolveRelatedAlerts(ctx context.Context, sourceEntityID uuid.UUID, alertType string, excludeAlertID uuid.UUID, now time.Time) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.alertbus.resolverelatedalerts")
	defer span.End()

	// Require both sourceEntityID and alertType to resolve related alerts
	if sourceEntityID == uuid.Nil || alertType == "" {
		return 0, nil
	}

	count, err := b.storer.ResolveRelatedAlerts(ctx, sourceEntityID, alertType, excludeAlertID, now)
	if err != nil {
		return 0, fmt.Errorf("resolve related alerts: %w", err)
	}

	if count > 0 {
		b.log.Info(ctx, "auto-resolved prior alerts",
			"source_entity_id", sourceEntityID,
			"alert_type", alertType,
			"resolved_count", count)
	}

	return count, nil
}
