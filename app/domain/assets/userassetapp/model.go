package userassetapp

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/assets/userassetbus"
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
		DateReceived:        bus.DateReceived.Format(time.RFC3339),
		LastMaintenance:     bus.LastMaintenance.Format(time.RFC3339),
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
	userID, err := uuid.Parse(app.UserID)
	if err != nil {
		return userassetbus.NewUserAsset{}, errs.Newf(errs.InvalidArgument, "parse userID: %s", err)
	}

	assetID, err := uuid.Parse(app.AssetID)
	if err != nil {
		return userassetbus.NewUserAsset{}, errs.Newf(errs.InvalidArgument, "parse assetID: %s", err)
	}

	approvedBy, err := uuid.Parse(app.ApprovedBy)
	if err != nil {
		return userassetbus.NewUserAsset{}, errs.Newf(errs.InvalidArgument, "parse approvedBy: %s", err)
	}

	approvalStatusID, err := uuid.Parse(app.ApprovalStatusID)
	if err != nil {
		return userassetbus.NewUserAsset{}, errs.Newf(errs.InvalidArgument, "parse approvalStatusID: %s", err)
	}

	fulfillmentStatusID, err := uuid.Parse(app.FulfillmentStatusID)
	if err != nil {
		return userassetbus.NewUserAsset{}, errs.Newf(errs.InvalidArgument, "parse fulfillmentStatusID: %s", err)
	}

	dateReceived, err := time.Parse(time.RFC3339, app.DateReceived)
	if err != nil {
		return userassetbus.NewUserAsset{}, errs.Newf(errs.InvalidArgument, "parse dateReceived: %s", err)
	}

	lastMaintenance, err := time.Parse(time.RFC3339, app.LastMaintenance)
	if err != nil {
		return userassetbus.NewUserAsset{}, errs.Newf(errs.InvalidArgument, "parse lastMaintenance: %s", err)
	}

	bus := userassetbus.NewUserAsset{
		UserID:              userID,
		AssetID:             assetID,
		ApprovedBy:          approvedBy,
		ApprovalStatusID:    approvalStatusID,
		FulfillmentStatusID: fulfillmentStatusID,
		DateReceived:        dateReceived,
		LastMaintenance:     lastMaintenance,
	}
	return bus, nil
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
	var userID *uuid.UUID
	if app.UserID != nil {
		id, err := uuid.Parse(*app.UserID)
		if err != nil {
			return userassetbus.UpdateUserAsset{}, errs.Newf(errs.InvalidArgument, "parse userID: %s", err)
		}
		userID = &id
	}

	var assetID *uuid.UUID
	if app.AssetID != nil {
		id, err := uuid.Parse(*app.AssetID)
		if err != nil {
			return userassetbus.UpdateUserAsset{}, errs.Newf(errs.InvalidArgument, "parse assetID: %s", err)
		}
		assetID = &id
	}

	var approvedBy *uuid.UUID
	if app.ApprovedBy != nil {
		id, err := uuid.Parse(*app.ApprovedBy)
		if err != nil {
			return userassetbus.UpdateUserAsset{}, errs.Newf(errs.InvalidArgument, "parse approvedBy: %s", err)
		}
		approvedBy = &id
	}

	var approvalStatusID *uuid.UUID
	if app.ApprovalStatusID != nil {
		id, err := uuid.Parse(*app.ApprovalStatusID)
		if err != nil {
			return userassetbus.UpdateUserAsset{}, errs.Newf(errs.InvalidArgument, "parse approvalStatusID: %s", err)
		}
		approvalStatusID = &id
	}

	var fulfillmentStatusID *uuid.UUID
	if app.FulfillmentStatusID != nil {
		id, err := uuid.Parse(*app.FulfillmentStatusID)
		if err != nil {
			return userassetbus.UpdateUserAsset{}, errs.Newf(errs.InvalidArgument, "parse fulfillmentStatusID: %s", err)
		}
		fulfillmentStatusID = &id
	}

	var dateReceived *time.Time
	if app.DateReceived != nil {
		t, err := time.Parse(time.RFC3339, *app.DateReceived)
		if err != nil {
			return userassetbus.UpdateUserAsset{}, errs.Newf(errs.InvalidArgument, "parse dateReceived: %s", err)
		}
		dateReceived = &t
	}

	var lastMaintenance *time.Time
	if app.LastMaintenance != nil {
		t, err := time.Parse(time.RFC3339, *app.LastMaintenance)
		if err != nil {
			return userassetbus.UpdateUserAsset{}, errs.Newf(errs.InvalidArgument, "parse lastMaintenance: %s", err)
		}
		lastMaintenance = &t
	}

	bus := userassetbus.UpdateUserAsset{
		UserID:              userID,
		AssetID:             assetID,
		ApprovedBy:          approvedBy,
		ApprovalStatusID:    approvalStatusID,
		FulfillmentStatusID: fulfillmentStatusID,
		DateReceived:        dateReceived,
		LastMaintenance:     lastMaintenance,
	}
	return bus, nil
}
