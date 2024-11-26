package assettypeapp

import "github.com/timmaaaz/ichor/business/sdk/order"

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"asset_type_id": "asset_type_id",
	"name":          "name",
}
