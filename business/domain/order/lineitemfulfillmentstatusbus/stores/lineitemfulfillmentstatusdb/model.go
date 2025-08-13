package lineitemfulfillmentstatusdb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/order/lineitemfulfillmentstatusbus"
)

type orderFulfillmentStatus struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
}

func toBusLineItemFulfillmentStatus(db orderFulfillmentStatus) lineitemfulfillmentstatusbus.LineItemFulfillmentStatus {
	return lineitemfulfillmentstatusbus.LineItemFulfillmentStatus{
		ID:          db.ID,
		Name:        db.Name,
		Description: db.Description,
	}
}

func toBusLineItemFulfillmentStatuses(dbs []orderFulfillmentStatus) []lineitemfulfillmentstatusbus.LineItemFulfillmentStatus {
	app := make([]lineitemfulfillmentstatusbus.LineItemFulfillmentStatus, len(dbs))
	for i, db := range dbs {
		app[i] = toBusLineItemFulfillmentStatus(db)
	}
	return app
}

func toDBLineItemFulfillmentStatus(app lineitemfulfillmentstatusbus.LineItemFulfillmentStatus) orderFulfillmentStatus {
	return orderFulfillmentStatus{
		ID:          app.ID,
		Name:        app.Name,
		Description: app.Description,
	}
}
