package assettagapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/assettagbus"
)

func parseFilter(qp QueryParams) (assettagbus.QueryFilter, error) {

	var filter assettagbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return assettagbus.QueryFilter{}, errs.NewFieldsError("asset_tag_id", err)
		}
		filter.ID = &id
	}

	if qp.AssetID != "" {
		id, err := uuid.Parse(qp.AssetID)
		if err != nil {
			return assettagbus.QueryFilter{}, errs.NewFieldsError("asset_id", err)
		}
		filter.AssetID = &id
	}

	if qp.TagID != "" {
		id, err := uuid.Parse(qp.TagID)
		if err != nil {
			return assettagbus.QueryFilter{}, errs.NewFieldsError("tag_id", err)
		}
		filter.TagID = &id
	}

	return filter, nil

}
