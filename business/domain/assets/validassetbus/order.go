package validassetbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByName, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID             = "id"
	OrderByTypeID         = "type_id"
	OrderByName           = "name"
	OrderByEstPrice       = "est_price"
	OrderByPrice          = "price"
	OrderByMaintenance    = "maintenance_interval"
	OrderByLifeExpectancy = "life_expectancy"
	OrderBySerialNumber   = "serial_number"
	OrderByModelNumber    = "model_number"
	OrderByIsEnabled      = "is_enabled"
	OrderByDateCreated    = "date_created"
	OrderByDateUpdated    = "date_updated"
	OrderByCreatedBy      = "created_by"
	OrderByUpdatedBy      = "updated_by"
)
