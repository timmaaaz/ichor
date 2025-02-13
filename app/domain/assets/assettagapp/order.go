package assettagapp

import (
	"github.com/timmaaaz/ichor/business/domain/assets/assettagbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("asset_tag_id", order.ASC)

var orderByFields = map[string]string{
	"asset_tag_id": assettagbus.OrderByID,
	"asset_id":     assettagbus.OrderByAssetID,
	"tag_id":       assettagbus.OrderByTagID,
}
