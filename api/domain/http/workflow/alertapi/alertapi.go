// Package alertapi provides HTTP handlers for workflow alerts.
package alertapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
	"github.com/timmaaaz/ichor/foundation/web"
)

// orderByFields maps API field names to database field names for ordering.
var orderByFields = map[string]string{
	"id":          alertbus.OrderByID,
	"alertType":   alertbus.OrderByAlertType,
	"severity":    alertbus.OrderBySeverity,
	"status":      alertbus.OrderByStatus,
	"createdDate": alertbus.OrderByCreatedDate,
	"updatedDate": alertbus.OrderByUpdatedDate,
}

type api struct {
	log           *logger.Logger
	alertBus      *alertbus.Business
	userRoleBus   *userrolebus.Business
	workflowQueue *rabbitmq.WorkflowQueue
}

func newAPI(log *logger.Logger, alertBus *alertbus.Business, userRoleBus *userrolebus.Business, workflowQueue *rabbitmq.WorkflowQueue) *api {
	return &api{
		log:           log,
		alertBus:      alertBus,
		userRoleBus:   userRoleBus,
		workflowQueue: workflowQueue,
	}
}

// query returns all alerts (admin only).
func (a *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp := parseQueryParams(r)

	filter, err := parseFilter(qp)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, alertbus.DefaultOrderBy)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	alerts, err := a.alertBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.alertBus.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toAppAlerts(alerts), total, pg)
}

// queryMine returns alerts for the authenticated user.
func (a *api) queryMine(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	roleIDs, err := a.getUserRoleIDs(ctx, userID)
	if err != nil {
		return errs.Newf(errs.Internal, "get user roles: %s", err)
	}

	qp := parseQueryParams(r)

	filter, err := parseFilter(qp)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, alertbus.DefaultOrderBy)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	alerts, err := a.alertBus.QueryMine(ctx, userID, roleIDs, filter, orderBy, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query mine: %s", err)
	}

	total, err := a.alertBus.CountMine(ctx, userID, roleIDs, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count mine: %s", err)
	}

	return query.NewResult(toAppAlerts(alerts), total, pg)
}

// queryByID returns a single alert by ID.
func (a *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	id, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// Verify user is authenticated
	_, err = mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	alert, err := a.alertBus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, alertbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "query by id: %s", err)
	}

	return toAppAlert(alert)
}

// acknowledge marks an alert as acknowledged by the user.
func (a *api) acknowledge(ctx context.Context, r *http.Request) web.Encoder {
	id, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	roleIDs, err := a.getUserRoleIDs(ctx, userID)
	if err != nil {
		return errs.Newf(errs.Internal, "get user roles: %s", err)
	}

	var req AcknowledgeRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	alert, err := a.alertBus.Acknowledge(ctx, id, userID, roleIDs, req.Notes, time.Now())
	if err != nil {
		if errors.Is(err, alertbus.ErrNotRecipient) {
			return errs.New(errs.PermissionDenied, err)
		}
		if errors.Is(err, alertbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return errs.New(errs.AlreadyExists, fmt.Errorf("already acknowledged"))
		}
		return errs.Newf(errs.Internal, "acknowledge: %s", err)
	}

	return toAppAlert(alert)
}

// dismiss marks an alert as dismissed by the user.
func (a *api) dismiss(ctx context.Context, r *http.Request) web.Encoder {
	id, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	roleIDs, err := a.getUserRoleIDs(ctx, userID)
	if err != nil {
		return errs.Newf(errs.Internal, "get user roles: %s", err)
	}

	alert, err := a.alertBus.Dismiss(ctx, id, userID, roleIDs, time.Now())
	if err != nil {
		if errors.Is(err, alertbus.ErrNotRecipient) {
			return errs.New(errs.PermissionDenied, err)
		}
		if errors.Is(err, alertbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "dismiss: %s", err)
	}

	return toAppAlert(alert)
}

// getUserRoleIDs fetches role IDs for the current user.
func (a *api) getUserRoleIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	userRoles, err := a.userRoleBus.QueryByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	roleIDs := make([]uuid.UUID, len(userRoles))
	for i, ur := range userRoles {
		roleIDs[i] = ur.RoleID
	}
	return roleIDs, nil
}

// testAlert creates a test alert for the authenticated user (for E2E WebSocket testing).
func (a *api) testAlert(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	now := time.Now()
	alertID := uuid.New()

	// Create alert in database
	alert := alertbus.Alert{
		ID:          alertID,
		AlertType:   "test_alert",
		Severity:    alertbus.SeverityMedium,
		Title:       "Test Alert",
		Message:     "This is a test alert for E2E testing",
		Status:      alertbus.StatusActive,
		CreatedDate: now,
		UpdatedDate: now,
	}

	if err := a.alertBus.Create(ctx, alert); err != nil {
		return errs.Newf(errs.Internal, "create alert: %s", err)
	}

	// Create recipient record for the user
	recipient := alertbus.AlertRecipient{
		ID:            uuid.New(),
		AlertID:       alertID,
		RecipientType: "user",
		RecipientID:   userID,
		CreatedDate:   now,
	}

	if err := a.alertBus.CreateRecipients(ctx, []alertbus.AlertRecipient{recipient}); err != nil {
		return errs.Newf(errs.Internal, "create recipient: %s", err)
	}

	// Publish to RabbitMQ (if available) for WebSocket delivery
	if a.workflowQueue != nil {
		alertPayload := map[string]interface{}{
			"alert": map[string]interface{}{
				"id":          alertID.String(),
				"alertType":   "test_alert",
				"severity":    alertbus.SeverityMedium,
				"title":       "Test Alert",
				"message":     "This is a test alert for E2E testing",
				"status":      alertbus.StatusActive,
				"createdDate": now.Format(time.RFC3339),
				"updatedDate": now.Format(time.RFC3339),
			},
		}

		msg := &rabbitmq.Message{
			Type:       "alert",
			EntityName: "workflow.alerts",
			EntityID:   alertID,
			UserID:     userID, // Target this user
			Payload:    alertPayload,
		}

		if err := a.workflowQueue.Publish(ctx, rabbitmq.QueueTypeAlert, msg); err != nil {
			a.log.Error(ctx, "failed to publish test alert to rabbitmq", "error", err)
			// Don't fail the request - the alert is still created in the database
		}
	}

	return toAppAlert(alert)
}

// acknowledgeSelected acknowledges multiple alerts by ID.
func (a *api) acknowledgeSelected(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	roleIDs, err := a.getUserRoleIDs(ctx, userID)
	if err != nil {
		return errs.Newf(errs.Internal, "get user roles: %s", err)
	}

	var req BulkSelectedRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}
	if err := req.Validate(); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	alertIDs, err := parseUUIDs(req.IDs)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	count, skipped, err := a.alertBus.AcknowledgeSelected(ctx, alertIDs, userID, roleIDs, req.Notes, time.Now())
	if err != nil {
		return errs.Newf(errs.Internal, "acknowledge selected: %s", err)
	}

	return BulkActionResult{Count: count, Skipped: skipped}
}

// acknowledgeAll acknowledges all active alerts for the user.
func (a *api) acknowledgeAll(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	roleIDs, err := a.getUserRoleIDs(ctx, userID)
	if err != nil {
		return errs.Newf(errs.Internal, "get user roles: %s", err)
	}

	var req BulkAllRequest
	if err := web.Decode(r, &req); err != nil {
		req = BulkAllRequest{} // Optional body
	}

	count, err := a.alertBus.AcknowledgeAll(ctx, userID, roleIDs, req.Notes, time.Now())
	if err != nil {
		return errs.Newf(errs.Internal, "acknowledge all: %s", err)
	}

	return BulkActionResult{Count: count, Skipped: 0}
}

// dismissSelected dismisses multiple alerts by ID.
func (a *api) dismissSelected(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	roleIDs, err := a.getUserRoleIDs(ctx, userID)
	if err != nil {
		return errs.Newf(errs.Internal, "get user roles: %s", err)
	}

	var req BulkSelectedRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}
	if err := req.Validate(); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	alertIDs, err := parseUUIDs(req.IDs)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	count, skipped, err := a.alertBus.DismissSelected(ctx, alertIDs, userID, roleIDs, time.Now())
	if err != nil {
		return errs.Newf(errs.Internal, "dismiss selected: %s", err)
	}

	return BulkActionResult{Count: count, Skipped: skipped}
}

// dismissAll dismisses all active alerts for the user.
func (a *api) dismissAll(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	roleIDs, err := a.getUserRoleIDs(ctx, userID)
	if err != nil {
		return errs.Newf(errs.Internal, "get user roles: %s", err)
	}

	count, err := a.alertBus.DismissAll(ctx, userID, roleIDs, time.Now())
	if err != nil {
		return errs.Newf(errs.Internal, "dismiss all: %s", err)
	}

	return BulkActionResult{Count: count, Skipped: 0}
}

// parseUUIDs converts a slice of string IDs to UUIDs.
func parseUUIDs(ids []string) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("invalid id at index %d: %w", i, err)
		}
		result[i] = parsed
	}
	return result, nil
}
