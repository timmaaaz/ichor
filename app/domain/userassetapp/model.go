package userassetapp

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/userassetbus"
)

const TimeLayout = "2006-01-02 15:04:05 -0700 MST"

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
	ConditionID         string `json:"condition_id"`
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
		ConditionID:         bus.ConditionID.String(),
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
	ConditionID         string `json:"condition_id" validate:"required"`
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
	var userID, assetID, approvedBy, conditionID, approvalStatusID, fulfillmentStatusID uuid.UUID
	var dateReceived, lastMaintenance time.Time
	var err error

	if app.UserID != "" {
		userID, err = uuid.Parse(app.UserID)
		if err != nil {
			return userassetbus.NewUserAsset{}, err
		}
	}

	if app.ConditionID != "" {
		conditionID, err = uuid.Parse(app.ConditionID)
		if err != nil {
			return userassetbus.NewUserAsset{}, err
		}
	}

	if app.ApprovalStatusID != "" {
		approvalStatusID, err = uuid.Parse(app.ApprovalStatusID)
		if err != nil {
			return userassetbus.NewUserAsset{}, err
		}
	}

	if app.FulfillmentStatusID != "" {
		fulfillmentStatusID, err = uuid.Parse(app.FulfillmentStatusID)
		if err != nil {
			return userassetbus.NewUserAsset{}, err
		}
	}

	if app.AssetID != "" {
		assetID, err = uuid.Parse(app.AssetID)
		if err != nil {
			return userassetbus.NewUserAsset{}, err
		}
	}

	if app.ApprovedBy != "" {
		approvedBy, err = uuid.Parse(app.ApprovedBy)
		if err != nil {
			return userassetbus.NewUserAsset{}, err
		}
	}

	if app.DateReceived != "" {
		dateReceived, err = time.Parse(TimeLayout, app.DateReceived)
		if err != nil {
			return userassetbus.NewUserAsset{}, err
		}
	}

	if app.LastMaintenance != "" {
		lastMaintenance, err = time.Parse(TimeLayout, app.LastMaintenance)
		if err != nil {
			return userassetbus.NewUserAsset{}, err
		}
	}

	return userassetbus.NewUserAsset{
		UserID:              userID,
		AssetID:             assetID,
		ApprovedBy:          approvedBy,
		ConditionID:         conditionID,
		ApprovalStatusID:    approvalStatusID,
		FulfillmentStatusID: fulfillmentStatusID,
		DateReceived:        dateReceived,
		LastMaintenance:     lastMaintenance,
	}, nil
}

// =========================================================================

type UpdateUserAsset struct {
	UserID              *string `json:"user_id"`
	AssetID             *string `json:"asset_id"`
	ApprovedBy          *string `json:"approved_by"`
	ConditionID         *string `json:"condition_id"`
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
	var userID, assetID, approvedBy, conditionID, approvalStatusID, fulfillmentStatusID *uuid.UUID
	var dateReceived, lastMaintenance *time.Time

	if app.UserID != nil {
		id, err := uuid.Parse(*app.UserID)
		if err != nil {
			return userassetbus.UpdateUserAsset{}, err
		}
		userID = &id
	}

	if app.ConditionID != nil {
		id, err := uuid.Parse(*app.ConditionID)
		if err != nil {
			return userassetbus.UpdateUserAsset{}, err
		}
		conditionID = &id
	}

	if app.ApprovalStatusID != nil {
		id, err := uuid.Parse(*app.ApprovalStatusID)
		if err != nil {
			return userassetbus.UpdateUserAsset{}, err
		}
		approvalStatusID = &id
	}

	if app.FulfillmentStatusID != nil {
		id, err := uuid.Parse(*app.FulfillmentStatusID)
		if err != nil {
			return userassetbus.UpdateUserAsset{}, err
		}
		fulfillmentStatusID = &id
	}

	if app.AssetID != nil {
		id, err := uuid.Parse(*app.AssetID)
		if err != nil {
			return userassetbus.UpdateUserAsset{}, err
		}
		assetID = &id
	}

	if app.ApprovedBy != nil {
		id, err := uuid.Parse(*app.ApprovedBy)
		if err != nil {
			return userassetbus.UpdateUserAsset{}, err
		}
		approvedBy = &id
	}

	if app.DateReceived != nil {
		dr, err := time.Parse(TimeLayout, *app.DateReceived)
		if err != nil {
			return userassetbus.UpdateUserAsset{}, err
		}
		dateReceived = &dr
	}

	if app.LastMaintenance != nil {
		lm, err := time.Parse(TimeLayout, *app.LastMaintenance)
		if err != nil {
			return userassetbus.UpdateUserAsset{}, err
		}
		lastMaintenance = &lm
	}

	return userassetbus.UpdateUserAsset{
		UserID:              userID,
		AssetID:             assetID,
		ApprovedBy:          approvedBy,
		ConditionID:         conditionID,
		ApprovalStatusID:    approvalStatusID,
		FulfillmentStatusID: fulfillmentStatusID,
		DateReceived:        dateReceived,
		LastMaintenance:     lastMaintenance,
	}, nil
}
