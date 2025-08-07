package assettagapp

import (
	"github.com/timmaaaz/ichor/business/domain/assets/assettagbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("id", order.ASC)

var orderByFields = map[string]string{
	"id":             assettagbus.OrderByID,
	"valid_asset_id": assettagbus.OrderByValidAssetID,
	"tag_id":         assettagbus.OrderByTagID,
}
