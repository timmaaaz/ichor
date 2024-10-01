package assettypeapp

import (
	"encoding/json"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/assettypebus"
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

type AssetType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Encode implements the encoder interface.
func (app AssetType) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppAssetType(bus assettypebus.AssetType) AssetType {
	return AssetType{
		ID:          bus.ID.String(),
		Name:        bus.Name,
		Description: bus.Description,
	}
}

func ToAppAssetTypes(bus []assettypebus.AssetType) []AssetType {
	app := make([]AssetType, len(bus))
	for i, v := range bus {
		app[i] = ToAppAssetType(v)
	}
	return app
}

// =============================================================================

type NewAssetType struct {
	Name        string `json:"name" validate:"required,min=3,max=31"`
	Description string `json:"description" validate:"required,min=3,max=127"`
}

// Decode implements the decoder interface.
func (app *NewAssetType) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app NewAssetType) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func ToBusNewAssetType(app NewAssetType) assettypebus.NewAssetType {
	return assettypebus.NewAssetType{
		Name:        app.Name,
		Description: app.Description,
	}
}

// =============================================================================

type UpdateAssetType struct {
	Name        *string `json:"name" validate:"omitempty,min=3,max=31"`
	Description *string `json:"description" validate:"omitempty,min=3,max=127"`
}

// Decode implements the decoder interface.
func (app *UpdateAssetType) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateAssetType) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func ToBusUpdateAssetType(app UpdateAssetType) assettypebus.UpdateAssetType {
	return assettypebus.UpdateAssetType{
		Name:        app.Name,
		Description: app.Description,
	}
}
