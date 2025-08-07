package assettagapp

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/assets/assettagbus"
)

type QueryParams struct {
	Page         string
	Rows         string
	OrderBy      string
	ID           string
	ValidAssetID string
	TagID        string
}

// =============================================================================

type AssetTag struct {
	ID           string `json:"id"`
	ValidAssetID string `json:"valid_asset_id"`
	TagID        string `json:"tag_id"`
}

func (app AssetTag) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppAssetTag(bus assettagbus.AssetTag) AssetTag {
	return AssetTag{
		ID:           bus.ID.String(),
		ValidAssetID: bus.ValidAssetID.String(),
		TagID:        bus.TagID.String(),
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
	ValidAssetID string `json:"valid_asset_id" validate:"required"`
	TagID        string `json:"tag_id" validate:"required"`
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
	var ValidAssetID, tagID uuid.UUID
	var err error

	if ValidAssetID, err = uuid.Parse(app.ValidAssetID); err != nil {
		return assettagbus.NewAssetTag{}, err
	}

	if tagID, err = uuid.Parse(app.TagID); err != nil {
		return assettagbus.NewAssetTag{}, err
	}

	return assettagbus.NewAssetTag{
		ValidAssetID: ValidAssetID,
		TagID:        tagID,
	}, nil
}

// =============================================================================

type UpdateAssetTag struct {
	ValidAssetID *string `json:"valid_asset_id"`
	TagID        *string `json:"tag_id"`
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
	var ValidAssetID, tagID *uuid.UUID

	if app.ValidAssetID != nil {
		id, err := uuid.Parse(*app.ValidAssetID)
		if err != nil {
			return assettagbus.UpdateAssetTag{}, err
		}

		ValidAssetID = &id
	}

	if app.TagID != nil {
		id, err := uuid.Parse(*app.TagID)
		if err != nil {
			return assettagbus.UpdateAssetTag{}, err
		}
		tagID = &id
	}

	return assettagbus.UpdateAssetTag{
		ValidAssetID: ValidAssetID,
		TagID:        tagID,
	}, nil
}
