package assettypeapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/assettypebus"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string
	ID      string
	Name    string
}

type AssetType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (app AssetType) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppAssetType(bus assettypebus.AssetType) AssetType {
	return AssetType{
		ID:   bus.ID.String(),
		Name: bus.Name,
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
	Name string `json:"name" validate:"required,min=3,max=100"`
}

func (app *NewAssetType) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewAssetType) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "Validate: %s", err)
	}

	return nil
}

func toBusNewAssetType(app NewAssetType) (assettypebus.NewAssetType, error) {
	return assettypebus.NewAssetType{
		Name: app.Name, // TODO: Look at defining custom type
	}, nil
}

type UpdateAssetType struct {
	Name *string `json:"name" validate:"required,min=3,max=100"`
}

func (app *UpdateAssetType) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateAssetType) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateAssetType(app UpdateAssetType) (assettypebus.UpdateAssetType, error) {
	var name *string

	if app.Name != nil {
		name = app.Name
	}

	return assettypebus.UpdateAssetType{
		Name: name,
	}, nil
}
