package userassetapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/userassetbus"
	"github.com/timmaaaz/ichor/foundation/convert"
)

type QueryParams struct {
	Page                string
	Rows                string
	OrderBy             string
	ID                  string
	UserID              string
	AssetID             string
	ApprovedBy          string
	ConditionID         string
	ApprovalStatusID    string
	FulfillmentStatusID string
	DateReceived        string
	LastMaintenance     string
}

type UserAsset struct {
	ID                  string `json:"id"`
	UserID              string `json:"user_id"`
	AssetID             string `json:"asset_id"`
	ApprovedBy          string `json:"approved_by"`
	ApprovalStatusID    string `json:"approval_status_id"`
	FulfillmentStatusID string `json:"fulfillment_status_id"`

	DateReceived    string `json:"date_received"`
	LastMaintenance string `json:"last_maintenance"`
}

func (app UserAsset) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppUserAsset(bus userassetbus.UserAsset) UserAsset {
	return UserAsset{
		ID:                  bus.ID.String(),
		UserID:              bus.UserID.String(),
		AssetID:             bus.AssetID.String(),
		ApprovedBy:          bus.ApprovedBy.String(),
		ApprovalStatusID:    bus.ApprovalStatusID.String(),
		FulfillmentStatusID: bus.FulfillmentStatusID.String(),
		DateReceived:        bus.DateReceived.String(),
		LastMaintenance:     bus.LastMaintenance.String(),
	}
}

func ToAppUserAssets(bus []userassetbus.UserAsset) []UserAsset {
	app := make([]UserAsset, len(bus))
	for i, v := range bus {
		app[i] = ToAppUserAsset(v)
	}
	return app
}

// =========================================================================

type NewUserAsset struct {
	UserID              string `json:"user_id" validate:"required"`
	AssetID             string `json:"asset_id" validate:"required"`
	ApprovedBy          string `json:"approved_by" validate:"required"`
	ApprovalStatusID    string `json:"approval_status_id" validate:"required"`
	FulfillmentStatusID string `json:"fulfillment_status_id" validate:"required"`

	DateReceived    string `json:"date_received" validate:"required"`
	LastMaintenance string `json:"last_maintenance" validate:"required"`
}

// Decode implements the decoder interface.
func (app *NewUserAsset) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app NewUserAsset) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusNewUserAsset(app NewUserAsset) (userassetbus.NewUserAsset, error) {

	dst := &userassetbus.NewUserAsset{}

	err := convert.PopulateTypesFromStrings(app, dst)
	return *dst, err
}

// =========================================================================

type UpdateUserAsset struct {
	UserID              *string `json:"user_id"`
	AssetID             *string `json:"asset_id"`
	ApprovedBy          *string `json:"approved_by"`
	ApprovalStatusID    *string `json:"approval_status_id"`
	FulfillmentStatusID *string `json:"fulfillment_status_id"`

	DateReceived    *string `json:"date_received"`
	LastMaintenance *string `json:"last_maintenance"`
}

// Decode implements the decoder interface.
func (app *UpdateUserAsset) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateUserAsset) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateUserAsset(app UpdateUserAsset) (userassetbus.UpdateUserAsset, error) {
	uua := userassetbus.UpdateUserAsset{}

	err := convert.PopulateTypesFromStrings(app, &uua)

	return uua, err
}
