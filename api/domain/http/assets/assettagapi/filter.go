package assettagapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/assets/assettagapp"
)

func parseQueryParams(r *http.Request) (assettagapp.QueryParams, error) {
	values := r.URL.Query()

	filter := assettagapp.QueryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("orderBy"),
		AssetID: values.Get("asset_id"),
		TagID:   values.Get("tag_id"),
	}

	return filter, nil
}
