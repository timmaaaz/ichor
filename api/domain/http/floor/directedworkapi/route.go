package directedworkapi

import (
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inspectionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Config holds the dependencies for the directed-work API routes.
type Config struct {
	Log               *logger.Logger
	PickTaskBus       *picktaskbus.Business
	PutAwayTaskBus    *putawaytaskbus.Business
	CycleCountItemBus *cyclecountitembus.Business
	InspectionBus     *inspectionbus.Business
	TransferOrderBus  *transferorderbus.Business
	OrdersBus         *ordersbus.Business
	AuthClient        *authclient.Client
}
