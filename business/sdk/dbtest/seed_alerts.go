package dbtest

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// seedAlerts creates a predictable set of workflow alerts for E2E testing.
// Distribution: 5 active (2 high, 2 medium, 1 low), 2 acknowledged, 1 dismissed.
// No active criticals seeded — overlay tests create their own via factory to avoid blocking other tests.
// All alerts are addressed to the admin user so floor worker E2E tests can see them.
func seedAlerts(ctx context.Context, log *logger.Logger, busDomain BusDomain, adminID uuid.UUID) error {
	log.Info(ctx, "Seeding workflow alerts for E2E testing...")

	now := time.Now()
	emptyCtx := json.RawMessage(`{}`)

	type alertSpec struct {
		severity         string
		alertType        string
		title            string
		message          string
		sourceEntityName string
		status           string
	}

	specs := []alertSpec{
		// Active — high (2)
		{
			severity:         alertbus.SeverityHigh,
			alertType:        "transfer_pending",
			title:            "High: Transfer Awaiting Approval",
			message:          "Transfer request from Zone A to Zone B has been pending approval for over 2 hours.",
			sourceEntityName: "transfers",
			status:           alertbus.StatusActive,
		},
		{
			severity:         alertbus.SeverityHigh,
			alertType:        "reorder_point_reached",
			title:            "High: Reorder Point Reached",
			message:          "Product SKU-2210 has reached its reorder threshold. Review procurement queue.",
			sourceEntityName: "inventory_locations",
			status:           alertbus.StatusActive,
		},
		// Active — medium (2)
		{
			severity:         alertbus.SeverityMedium,
			alertType:        "cycle_count_discrepancy",
			title:            "Cycle Count Discrepancy",
			message:          "Cycle count in Zone C shows a variance of -3 units for SKU-0987. Manual recount recommended.",
			sourceEntityName: "inventory_locations",
			status:           alertbus.StatusActive,
		},
		{
			severity:         alertbus.SeverityMedium,
			alertType:        "receiving_delay",
			title:            "Receiving Task Overdue",
			message:          "PO #5678 receiving task is 45 minutes past its expected completion time.",
			sourceEntityName: "receiving_tasks",
			status:           alertbus.StatusActive,
		},
		// Active — low (1)
		{
			severity:         alertbus.SeverityLow,
			alertType:        "low_stock_advisory",
			title:            "Low Stock Advisory",
			message:          "Product SKU-3301 is approaching minimum threshold. No immediate action required.",
			sourceEntityName: "inventory_locations",
			status:           alertbus.StatusActive,
		},
		// Acknowledged (2)
		{
			severity:  alertbus.SeverityCritical,
			alertType: "inventory_critical",
			title:     "Critical Alert — Resolved",
			message:   "Previously critical inventory issue has been addressed by the warehouse team.",
			status:    alertbus.StatusAcknowledged,
		},
		{
			severity:  alertbus.SeverityHigh,
			alertType: "transfer_pending",
			title:     "High Priority Alert — Acknowledged",
			message:   "Transfer approval delay was acknowledged and escalated to supervisor.",
			status:    alertbus.StatusAcknowledged,
		},
		// Dismissed (1)
		{
			severity:  alertbus.SeverityMedium,
			alertType: "routine_check",
			title:     "Routine Check Reminder",
			message:   "End-of-shift equipment check reminder. Dismissed — completed.",
			status:    alertbus.StatusDismissed,
		},
	}

	for _, s := range specs {
		alertID := uuid.New()

		alert := alertbus.Alert{
			ID:               alertID,
			AlertType:        s.alertType,
			Severity:         s.severity,
			Title:            s.title,
			Message:          s.message,
			Context:          emptyCtx,
			SourceEntityName: s.sourceEntityName,
			Status:           s.status,
			CreatedDate:      now,
			UpdatedDate:      now,
		}

		if err := busDomain.Alert.Create(ctx, alert); err != nil {
			log.Error(ctx, "Failed to seed alert", "title", s.title, "error", err)
			continue
		}

		recipient := alertbus.AlertRecipient{
			ID:            uuid.New(),
			AlertID:       alertID,
			RecipientType: "user",
			RecipientID:   adminID,
			CreatedDate:   now,
		}

		if err := busDomain.Alert.CreateRecipients(ctx, []alertbus.AlertRecipient{recipient}); err != nil {
			log.Error(ctx, "Failed to seed alert recipient", "alert_title", s.title, "error", err)
		}
	}

	log.Info(ctx, "Alert seeding complete", "count", len(specs))
	return nil
}
