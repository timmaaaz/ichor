package customersapp

import (
	"github.com/timmaaaz/ichor/business/domain/core/customersbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"id":                  customersbus.OrderByID,
	"name":                customersbus.OrderByName,
	"contact_id":          customersbus.OrderByContactID,
	"delivery_address_id": customersbus.OrderByDeliveryAddressID,
	"notes":               customersbus.OrderByNotes,
	"created_by":          customersbus.OrderByCreatedBy,
	"updated_by":          customersbus.OrderByUpdatedBy,
	"created_date":        customersbus.OrderByCreatedDate,
	"updated_date":        customersbus.OrderByUpdatedDate,
}
