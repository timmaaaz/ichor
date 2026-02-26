// Package notificationsapi provides the HTTP handler for the notifications summary endpoint.
package notificationsapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	log         *logger.Logger
	alertBus    *alertbus.Business
	approvalBus *approvalrequestbus.Business
	userRoleBus *userrolebus.Business
}

func newAPI(cfg Config) *api {
	return &api{
		log:         cfg.Log,
		alertBus:    cfg.AlertBus,
		approvalBus: cfg.ApprovalBus,
		userRoleBus: cfg.UserRoleBus,
	}
}

// summary returns a consolidated notifications summary for the authenticated user.
func (a *api) summary(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	roleIDs, err := a.getUserRoleIDs(ctx, userID)
	if err != nil {
		return errs.Newf(errs.Internal, "get user roles: %s", err)
	}

	result := NotificationSummary{}

	// Alert severity counts (active + non-expired only).
	severityCounts, err := a.alertBus.CountMineBySeverity(ctx, userID, roleIDs, alertbus.StatusActive)
	if err != nil {
		a.log.Error(ctx, "failed to count alerts by severity", "error", err)
	} else {
		result.Alerts = AlertSummary{
			Critical: severityCounts[alertbus.SeverityCritical],
			High:     severityCounts[alertbus.SeverityHigh],
			Medium:   severityCounts[alertbus.SeverityMedium],
			Low:      severityCounts[alertbus.SeverityLow],
			Info:     severityCounts["info"],
		}
		for _, v := range severityCounts {
			result.Alerts.TotalActive += v
		}
	}

	// Pending approval count for this user.
	status := approvalrequestbus.StatusPending
	filter := approvalrequestbus.QueryFilter{
		ApproverID: &userID,
		Status:     &status,
	}
	pendingCount, err := a.approvalBus.Count(ctx, filter)
	if err != nil {
		a.log.Error(ctx, "failed to count pending approvals", "error", err)
	} else {
		result.Approvals = ApprovalSummary{PendingCount: pendingCount}
	}

	return result
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
