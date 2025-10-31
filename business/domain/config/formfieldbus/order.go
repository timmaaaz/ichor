package formfieldbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByFieldOrder, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID         = "id"
	OrderByFormID     = "form_id"
	OrderByName       = "name"
	OrderByFieldOrder = "field_order"
	OrderByFieldType  = "field_type"
)
