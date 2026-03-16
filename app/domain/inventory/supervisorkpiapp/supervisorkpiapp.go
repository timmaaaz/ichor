package supervisorkpiapp

import (
	"context"

	"github.com/timmaaaz/ichor/business/domain/inventory/inspectionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// App manages the supervisor KPI aggregation use case.
type App struct {
	log                    *logger.Logger
	approvalRequestBus     *approvalrequestbus.Business
	inventoryAdjustmentBus *inventoryadjustmentbus.Business
	transferOrderBus       *transferorderbus.Business
	inspectionBus          *inspectionbus.Business
	putAwayTaskBus         *putawaytaskbus.Business
	alertBus               *alertbus.Business
}

// NewApp constructs a supervisor KPI app.
func NewApp(
	log *logger.Logger,
	approvalRequestBus *approvalrequestbus.Business,
	inventoryAdjustmentBus *inventoryadjustmentbus.Business,
	transferOrderBus *transferorderbus.Business,
	inspectionBus *inspectionbus.Business,
	putAwayTaskBus *putawaytaskbus.Business,
	alertBus *alertbus.Business,
) *App {
	return &App{
		log:                    log,
		approvalRequestBus:     approvalRequestBus,
		inventoryAdjustmentBus: inventoryAdjustmentBus,
		transferOrderBus:       transferOrderBus,
		inspectionBus:          inspectionBus,
		putAwayTaskBus:         putAwayTaskBus,
		alertBus:               alertBus,
	}
}

// Query returns aggregated KPI counts for the supervisor dashboard.
func (a *App) Query(ctx context.Context) (KPIs, error) {
	var kpis KPIs

	pendingStatus := "pending"
	openStatus := "open"
	activeStatus := "active"
	pendingPATStatus := putawaytaskbus.Statuses.Pending

	approvalCount, err := a.approvalRequestBus.Count(ctx, approvalrequestbus.QueryFilter{
		Status: &pendingStatus,
	})
	if err != nil {
		return KPIs{}, err
	}
	kpis.PendingApprovals = approvalCount

	adjustmentCount, err := a.inventoryAdjustmentBus.Count(ctx, inventoryadjustmentbus.QueryFilter{
		ApprovalStatus: &pendingStatus,
	})
	if err != nil {
		return KPIs{}, err
	}
	kpis.PendingAdjustments = adjustmentCount

	transferCount, err := a.transferOrderBus.Count(ctx, transferorderbus.QueryFilter{
		Status: &pendingStatus,
	})
	if err != nil {
		return KPIs{}, err
	}
	kpis.PendingTransfers = transferCount

	inspectionCount, err := a.inspectionBus.Count(ctx, inspectionbus.QueryFilter{
		Status: &openStatus,
	})
	if err != nil {
		return KPIs{}, err
	}
	kpis.OpenInspections = inspectionCount

	putAwayCount, err := a.putAwayTaskBus.Count(ctx, putawaytaskbus.QueryFilter{
		Status: &pendingPATStatus,
	})
	if err != nil {
		return KPIs{}, err
	}
	kpis.PendingPutAwayTasks = putAwayCount

	alertCount, err := a.alertBus.Count(ctx, alertbus.QueryFilter{
		Status: &activeStatus,
	})
	if err != nil {
		return KPIs{}, err
	}
	kpis.ActiveAlerts = alertCount

	return kpis, nil
}
