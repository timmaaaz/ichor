package assetconditionapp

import (
	"encoding/json"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/assetconditionbus"
)

// QueryParams represents the query parameters that can be used.
type QueryParams struct {
	Page        string
	Rows        string
	OrderBy     string
	ID          string
	Name        string
	Description string
}

type AssetCondition struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Encode implements the encoder interface.
func (app AssetCondition) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppAssetCondition(bus assetconditionbus.AssetCondition) AssetCondition {
	return AssetCondition{
		ID:          bus.ID.String(),
		Name:        bus.Name,
		Description: bus.Description,
	}
}

func ToAppAssetConditions(bus []assetconditionbus.AssetCondition) []AssetCondition {
	app := make([]AssetCondition, len(bus))
	for i, v := range bus {
		app[i] = ToAppAssetCondition(v)
	}
	return app
}

// =============================================================================

type NewAssetCondition struct {
	Name        string `json:"name" validate:"required,min=3,max=31"`
	Description string `json:"description" validate:"required,min=3,max=127"`
}

// Decode implements the decoder interface.
func (app *NewAssetCondition) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app NewAssetCondition) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func ToBusNewAssetCondition(app NewAssetCondition) assetconditionbus.NewAssetCondition {
	return assetconditionbus.NewAssetCondition{
		Name:        app.Name,
		Description: app.Description,
	}
}

// =============================================================================

type UpdateAssetCondition struct {
	Name        *string `json:"name" validate:"omitempty,min=3,max=31"`
	Description *string `json:"description" validate:"omitempty,min=3,max=127"`
}

// Decode implements the decoder interface.
func (app *UpdateAssetCondition) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateAssetCondition) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func ToBusUpdateAssetCondition(app UpdateAssetCondition) assetconditionbus.UpdateAssetCondition {
	return assetconditionbus.UpdateAssetCondition{
		Name:        app.Name,
		Description: app.Description,
	}
}
