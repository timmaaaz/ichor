package countrybus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort. Usually we order by id but
// in the case of countries, they almost always should be alphabetical.
var DefaultOrderBy = order.NewBy(OrderByNumber, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID     = "id"
	OrderByNumber = "number"
	OrderByName   = "name"
	OrderByAlpha2 = "alpha_2"
	OrderByAlpha3 = "alpha_3"
)
