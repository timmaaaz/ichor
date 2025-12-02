package assetapp

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
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
	LastMaintenance  string `json:"last_maintenance"`
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
	validAssetID, err := uuid.Parse(app.ValidAssetID)
	if err != nil {
		return assetbus.NewAsset{}, errs.Newf(errs.InvalidArgument, "parse validAssetID: %s", err)
	}

	assetConditionID, err := uuid.Parse(app.AssetConditionID)
	if err != nil {
		return assetbus.NewAsset{}, errs.Newf(errs.InvalidArgument, "parse assetConditionID: %s", err)
	}

	var lastMaintenance time.Time
	if app.LastMaintenance != "" {
		lastMaintenance, err = time.Parse(timeutil.FORMAT, app.LastMaintenance)
		if err != nil {
			return assetbus.NewAsset{}, errs.Newf(errs.InvalidArgument, "parse lastMaintenance: %s", err)
		}
	}

	bus := assetbus.NewAsset{
		ValidAssetID:     validAssetID,
		AssetConditionID: assetConditionID,
		LastMaintenance:  lastMaintenance,
		SerialNumber:     app.SerialNumber,
	}
	return bus, nil
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
	var validAssetID *uuid.UUID
	if app.ValidAssetID != nil {
		id, err := uuid.Parse(*app.ValidAssetID)
		if err != nil {
			return assetbus.UpdateAsset{}, errs.Newf(errs.InvalidArgument, "parse validAssetID: %s", err)
		}
		validAssetID = &id
	}

	var assetConditionID *uuid.UUID
	if app.AssetConditionID != nil {
		id, err := uuid.Parse(*app.AssetConditionID)
		if err != nil {
			return assetbus.UpdateAsset{}, errs.Newf(errs.InvalidArgument, "parse assetConditionID: %s", err)
		}
		assetConditionID = &id
	}

	var lastMaintenance *time.Time
	if app.LastMaintenance != nil {
		t, err := time.Parse(timeutil.FORMAT, *app.LastMaintenance)
		if err != nil {
			return assetbus.UpdateAsset{}, errs.Newf(errs.InvalidArgument, "parse lastMaintenance: %s", err)
		}
		lastMaintenance = &t
	}

	bus := assetbus.UpdateAsset{
		ValidAssetID:     validAssetID,
		AssetConditionID: assetConditionID,
		LastMaintenance:  lastMaintenance,
		SerialNumber:     app.SerialNumber,
	}
	return bus, nil
}
