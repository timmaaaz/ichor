package customersbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID                = "id"
	OrderByName              = "name"
	OrderByContactID         = "contact_id"
	OrderByDeliveryAddressID = "delivery_address_id"
	OrderByNotes             = "notes"
	OrderByCreatedBy         = "created_by"
	OrderByUpdatedBy         = "updated_by"
	OrderByCreatedDate       = "created_date"
	OrderByUpdatedDate       = "updated_date"
)
