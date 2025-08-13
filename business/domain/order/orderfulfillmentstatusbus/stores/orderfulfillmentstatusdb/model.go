package orderfulfillmentstatusdb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/order/orderfulfillmentstatusbus"
)

type orderFulfillmentStatus struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
}

func toBusOrderFulfillmentStatus(db orderFulfillmentStatus) orderfulfillmentstatusbus.OrderFulfillmentStatus {
	return orderfulfillmentstatusbus.OrderFulfillmentStatus{
		ID:          db.ID,
		Name:        db.Name,
		Description: db.Description,
	}
}

func toBusOrderFulfillmentStatuses(dbs []orderFulfillmentStatus) []orderfulfillmentstatusbus.OrderFulfillmentStatus {
	app := make([]orderfulfillmentstatusbus.OrderFulfillmentStatus, len(dbs))
	for i, db := range dbs {
		app[i] = toBusOrderFulfillmentStatus(db)
	}
	return app
}

func toDBOrderFulfillmentStatus(app orderfulfillmentstatusbus.OrderFulfillmentStatus) orderFulfillmentStatus {
	return orderFulfillmentStatus{
		ID:          app.ID,
		Name:        app.Name,
		Description: app.Description,
	}
}
