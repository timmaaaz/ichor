package assettagbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID      = "asset_tag_id"
	OrderByAssetID = "asset_id"
	OrderByTagID   = "tag_id"
)
