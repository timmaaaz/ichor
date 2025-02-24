package restrictedcolumnbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByTableName, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID         = "restricted_column_id"
	OrderByTableName  = "table_name"
	OrderByColumnName = "column_name"
)
