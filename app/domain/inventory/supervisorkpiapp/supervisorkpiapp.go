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

	approvalStatus := approvalrequestbus.StatusPending
	adjustmentStatus := inventoryadjustmentbus.ApprovalStatusPending
	transferStatus := transferorderbus.StatusPending
	// inspectionbus has no exported status constants; "pending" is the only
	// status used in seeds and the Create method default.
	inspectionStatus := "pending"
	putAwayStatus := putawaytaskbus.Statuses.Pending
	alertStatus := alertbus.StatusActive

	approvalCount, err := a.approvalRequestBus.Count(ctx, approvalrequestbus.QueryFilter{
		Status: &approvalStatus,
	})
	if err != nil {
		return KPIs{}, err
	}
	kpis.PendingApprovals = approvalCount

	adjustmentCount, err := a.inventoryAdjustmentBus.Count(ctx, inventoryadjustmentbus.QueryFilter{
		ApprovalStatus: &adjustmentStatus,
	})
	if err != nil {
		return KPIs{}, err
	}
	kpis.PendingAdjustments = adjustmentCount

	transferCount, err := a.transferOrderBus.Count(ctx, transferorderbus.QueryFilter{
		Status: &transferStatus,
	})
	if err != nil {
		return KPIs{}, err
	}
	kpis.PendingTransfers = transferCount

	inspectionCount, err := a.inspectionBus.Count(ctx, inspectionbus.QueryFilter{
		Status: &inspectionStatus,
	})
	if err != nil {
		return KPIs{}, err
	}
	kpis.PendingInspections = inspectionCount

	putAwayCount, err := a.putAwayTaskBus.Count(ctx, putawaytaskbus.QueryFilter{
		Status: &putAwayStatus,
	})
	if err != nil {
		return KPIs{}, err
	}
	kpis.PendingPutAwayTasks = putAwayCount

	alertCount, err := a.alertBus.Count(ctx, alertbus.QueryFilter{
		Status: &alertStatus,
	})
	if err != nil {
		return KPIs{}, err
	}
	kpis.ActiveAlerts = alertCount

	return kpis, nil
}
