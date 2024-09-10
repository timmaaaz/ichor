package regionbus

import "bitbucket.org/superiortechnologies/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort. Usually we order by id but
// in the case of countries, they almost always should be by their country.
var DefaultOrderBy = order.NewBy(OrderByName, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID        = "region_id"
	OrderByName      = "name"
	OrderByCode      = "code"
	OrderByCountryID = "country_id"
)
