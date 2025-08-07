package assettagbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID           = "id"
	OrderByValidAssetID = "valid_asset_id"
	OrderByTagID        = "tag_id"
)
