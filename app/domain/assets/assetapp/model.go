package assetapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
	"github.com/timmaaaz/ichor/foundation/convert"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

type QueryParams struct {
	Page            string
	Rows            string
	OrderBy         string
	ID              string
	ValidAssetID    string
	ConditionID     string
	SerialNumber    string
	LastMaintenance string
}

type Asset struct {
	ID               string `json:"id"`
	ValidAssetID     string `json:"valid_asset_id"`
	AssetConditionID string `json:"asset_condition_id"`
	LastMaintenance  string `json:"last_maintenance"`
	SerialNumber     string `json:"serial_number"`
}

func (app Asset) Encode() ([]byte, string, error) {

	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppAsset(bus assetbus.Asset) Asset {

	return Asset{
		ID:               bus.ID.String(),
		ValidAssetID:     bus.ValidAssetID.String(),
		AssetConditionID: bus.AssetConditionID.String(),
		LastMaintenance:  bus.LastMaintenance.Format(timeutil.FORMAT),
		SerialNumber:     bus.SerialNumber,
	}
}

func ToAppAssets(bus []assetbus.Asset) []Asset {
	app := make([]Asset, len(bus))
	for i, v := range bus {
		app[i] = ToAppAsset(v)
	}
	return app
}

// =============================================================================

type NewAsset struct {
	ValidAssetID     string `json:"valid_asset_id" validate:"required"`
	AssetConditionID string `json:"asset_condition_id" validate:"required"`
	LastMaintenance  string `json:"last_maintenance" validate:"required"`
	SerialNumber     string `json:"serial_number" validate:"required"`
}

func (app *NewAsset) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewAsset) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusNewAsset(app NewAsset) (assetbus.NewAsset, error) {
	dest := assetbus.NewAsset{}

	err := convert.PopulateTypesFromStrings(app, &dest)
	return dest, err
}

type UpdateAsset struct {
	ValidAssetID     *string `json:"valid_asset_id" validate:"omitempty,min=36,max=36"`
	AssetConditionID *string `json:"asset_condition_id" validate:"omitempty,min=36,max=36"`
	LastMaintenance  *string `json:"last_maintenance"`
	SerialNumber     *string `json:"serial_number"`
}

// Decode implements the decoder interface.
func (app *UpdateAsset) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateAsset) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateAsset(app UpdateAsset) (assetbus.UpdateAsset, error) {
	dest := assetbus.UpdateAsset{}

	err := convert.PopulateTypesFromStrings(app, &dest)

	return dest, err
}
