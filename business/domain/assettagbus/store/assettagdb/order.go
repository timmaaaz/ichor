package assettagdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/assettagbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	assettagbus.OrderByID:      "asset_tag_id",
	assettagbus.OrderByAssetID: "asset_id",
	assettagbus.OrderByTagID:   "tag_id",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}
	return "ORDER BY " + by + " " + orderBy.Direction, nil
}
