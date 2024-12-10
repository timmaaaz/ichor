package assettagapp

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/assettagbus"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string
	ID      string
	AssetID string
	TagID   string
}

// =============================================================================

type AssetTag struct {
	ID      string `json:"id"`
	AssetID string `json:"asset_id"`
	TagID   string `json:"tag_id"`
}

func (app AssetTag) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppAssetTag(bus assettagbus.AssetTag) AssetTag {
	return AssetTag{
		ID:      bus.ID.String(),
		AssetID: bus.AssetID.String(),
		TagID:   bus.TagID.String(),
	}
}

func ToAppAssetTags(bus []assettagbus.AssetTag) []AssetTag {
	app := make([]AssetTag, len(bus))
	for i, v := range bus {
		app[i] = ToAppAssetTag(v)
	}
	return app
}

// =============================================================================

type NewAssetTag struct {
	AssetID string `json:"asset_id" validate:"required"`
	TagID   string `json:"tag_id" validate:"required"`
}

// Decode implements the decoder interface.
func (app *NewAssetTag) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app NewAssetTag) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusNewAssetTag(app NewAssetTag) (assettagbus.NewAssetTag, error) {
	var assetID, tagID uuid.UUID
	var err error

	if assetID, err = uuid.Parse(app.AssetID); err != nil {
		return assettagbus.NewAssetTag{}, err
	}

	if tagID, err = uuid.Parse(app.TagID); err != nil {
		return assettagbus.NewAssetTag{}, err
	}

	return assettagbus.NewAssetTag{
		AssetID: assetID,
		TagID:   tagID,
	}, nil
}

// =============================================================================

type UpdateAssetTag struct {
	AssetID *string `json:"asset_id"`
	TagID   *string `json:"tag_id"`
}

func (app *UpdateAssetTag) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateAssetTag) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateAssetTag(app UpdateAssetTag) (assettagbus.UpdateAssetTag, error) {
	var assetID, tagID *uuid.UUID

	if app.AssetID != nil {
		id, err := uuid.Parse(*app.AssetID)
		if err != nil {
			return assettagbus.UpdateAssetTag{}, err
		}

		assetID = &id
	}

	if app.TagID != nil {
		id, err := uuid.Parse(*app.TagID)
		if err != nil {
			return assettagbus.UpdateAssetTag{}, err
		}
		tagID = &id
	}

	return assettagbus.UpdateAssetTag{
		AssetID: assetID,
		TagID:   tagID,
	}, nil
}
