package assetconditionapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/assetconditionbus"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string
	ID      string
	Name    string
}

type AssetCondition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (app AssetCondition) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppAssetCondition(bus assetconditionbus.AssetCondition) AssetCondition {
	return AssetCondition{
		ID:   bus.ID.String(),
		Name: bus.Name,
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
	Name string `json:"name" validate:"required,min=3,max=100"`
}

func (app *NewAssetCondition) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewAssetCondition) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "Validate: %s", err)
	}

	return nil
}

func toBusNewAssetCondition(app NewAssetCondition) (assetconditionbus.NewAssetCondition, error) {

	return assetconditionbus.NewAssetCondition{
		Name: app.Name, // TODO: Look at defining custom type
	}, nil
}

type UpdateAssetCondition struct {
	Name *string `json:"name" validate:"required,min=3,max=100"`
}

func (app *UpdateAssetCondition) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateAssetCondition) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateAssetCondition(app UpdateAssetCondition) (assetconditionbus.UpdateAssetCondition, error) {
	var name *string

	if app.Name != nil {
		name = app.Name
	}

	return assetconditionbus.UpdateAssetCondition{
		Name: name,
	}, nil
}
