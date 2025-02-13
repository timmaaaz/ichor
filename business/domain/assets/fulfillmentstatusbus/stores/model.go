package fulfillmentstatusdb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus"
)

type fulfillmentStatus struct {
	ID     uuid.UUID `db:"fulfillment_status_id"`
	Name   string    `db:"name"`
	IconID uuid.UUID `db:"icon_id"`
}

func toDBFulfillmentStatus(fs fulfillmentstatusbus.FulfillmentStatus) fulfillmentStatus {
	return fulfillmentStatus{
		ID:     fs.ID,
		Name:   fs.Name,
		IconID: fs.IconID,
	}
}

func toBusFulfillmentStatus(dbFS fulfillmentStatus) fulfillmentstatusbus.FulfillmentStatus {
	return fulfillmentstatusbus.FulfillmentStatus{
		ID:     dbFS.ID,
		Name:   dbFS.Name,
		IconID: dbFS.IconID,
	}
}

func toBusFulfillmentStatuses(dbFS []fulfillmentStatus) []fulfillmentstatusbus.FulfillmentStatus {
	fulfillmentStatuses := make([]fulfillmentstatusbus.FulfillmentStatus, len(dbFS))
	for i, fs := range dbFS {
		fulfillmentStatuses[i] = toBusFulfillmentStatus(fs)
	}

	return fulfillmentStatuses
}
