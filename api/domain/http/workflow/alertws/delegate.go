package alertws

import (
	"context"
	"encoding/json"

	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// AlertHubDelegateHandler handles delegate events for AlertHub.
type AlertHubDelegateHandler struct {
	alertHub *AlertHub
	log      *logger.Logger
}

// NewAlertHubDelegateHandler creates a handler for role change events.
func NewAlertHubDelegateHandler(alertHub *AlertHub, log *logger.Logger) *AlertHubDelegateHandler {
	return &AlertHubDelegateHandler{
		alertHub: alertHub,
		log:      log,
	}
}

// RegisterRoleChanges registers delegate handlers for user role changes.
func (h *AlertHubDelegateHandler) RegisterRoleChanges(del *delegate.Delegate) {
	del.Register(userrolebus.DomainName, userrolebus.ActionCreated, h.handleRoleChange)
	del.Register(userrolebus.DomainName, userrolebus.ActionDeleted, h.handleRoleChange)
}

// handleRoleChange refreshes roles for all connections of the affected user.
func (h *AlertHubDelegateHandler) handleRoleChange(ctx context.Context, data delegate.Data) error {
	// Parse the event to get userID
	var params userrolebus.ActionCreatedParms
	if err := json.Unmarshal(data.RawParams, &params); err != nil {
		h.log.Error(ctx, "failed to parse role change event", "error", err)
		return nil // Don't fail the business operation
	}

	userID := params.Entity.UserID

	// Use AlertHub to refresh roles (it handles fetching and ID conversion)
	if err := h.alertHub.RefreshUserRoles(ctx, userID); err != nil {
		h.log.Error(ctx, "failed to refresh websocket roles",
			"user_id", userID, "error", err)
		return nil // Don't fail the business operation
	}

	h.log.Debug(ctx, "refreshed roles for websocket connections",
		"user_id", userID)

	return nil
}
