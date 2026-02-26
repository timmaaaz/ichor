package inventoryadjustmentapp

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string

	InventoryAdjustmentID string
	ProductID             string
	LocationID            string
	AdjustedBy            string
	ApprovedBy            string
	ApprovalStatus        string
	QuantityChange        string
	ReasonCode            string
	Notes                 string
	AdjustmentDate        string
	StartAdjustmentDate   string
	EndAdjustmentDate     string
	CreatedDate           string
	StartCreatedDate      string
	EndCreatedDate        string
	UpdatedDate           string
}

type InventoryAdjustment struct {
	InventoryAdjustmentID string `json:"adjustment_id"`
	ProductID             string `json:"product_id"`
	LocationID            string `json:"location_id"`
	AdjustedBy            string `json:"adjusted_by"`
	ApprovedBy            string `json:"approved_by"`
	ApprovalStatus        string `json:"approval_status"`
	QuantityChange        string `json:"quantity_change"`
	ReasonCode            string `json:"reason_code"`
	Notes                 string `json:"notes"`
	AdjustmentDate        string `json:"adjustment_date"`
	CreatedDate           string `json:"created_date"`
	UpdatedDate           string `json:"updated_date"`
}

func (app InventoryAdjustment) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppInventoryAdjustment(bus inventoryadjustmentbus.InventoryAdjustment) InventoryAdjustment {
	approvedBy := ""
	if bus.ApprovedBy != nil {
		approvedBy = bus.ApprovedBy.String()
	}

	return InventoryAdjustment{
		InventoryAdjustmentID: bus.InventoryAdjustmentID.String(),
		ProductID:             bus.ProductID.String(),
		LocationID:            bus.LocationID.String(),
		AdjustedBy:            bus.AdjustedBy.String(),
		ApprovedBy:            approvedBy,
		ApprovalStatus:        bus.ApprovalStatus,
		QuantityChange:        strconv.Itoa(bus.QuantityChange),
		ReasonCode:            bus.ReasonCode,
		Notes:                 bus.Notes,
		AdjustmentDate:        bus.AdjustmentDate.Format(timeutil.FORMAT),
		CreatedDate:           bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:           bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

func ToAppInventoryAdjustments(bus []inventoryadjustmentbus.InventoryAdjustment) []InventoryAdjustment {
	app := make([]InventoryAdjustment, len(bus))
	for i, v := range bus {
		app[i] = ToAppInventoryAdjustment(v)
	}
	return app
}

type NewInventoryAdjustment struct {
	ProductID      string `json:"product_id" validate:"required,min=36,max=36"`
	LocationID     string `json:"location_id" validate:"required,min=36,max=36"`
	AdjustedBy     string `json:"adjusted_by" validate:"required,min=36,max=36"`
	ApprovedBy     string `json:"approved_by" validate:"omitempty,min=36,max=36"`
	QuantityChange string `json:"quantity_change" validate:"required"`
	ReasonCode     string `json:"reason_code" validate:"required"`
	Notes          string `json:"notes" validate:"required"`
	AdjustmentDate string `json:"adjustment_date" validate:"required"`
}

func (app *NewInventoryAdjustment) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewInventoryAdjustment) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewInventoryAdjustment(app NewInventoryAdjustment) (inventoryadjustmentbus.NewInventoryAdjustment, error) {
	productID, err := uuid.Parse(app.ProductID)
	if err != nil {
		return inventoryadjustmentbus.NewInventoryAdjustment{}, errs.Newf(errs.InvalidArgument, "parse productID: %s", err)
	}

	locationID, err := uuid.Parse(app.LocationID)
	if err != nil {
		return inventoryadjustmentbus.NewInventoryAdjustment{}, errs.Newf(errs.InvalidArgument, "parse locationID: %s", err)
	}

	adjustedBy, err := uuid.Parse(app.AdjustedBy)
	if err != nil {
		return inventoryadjustmentbus.NewInventoryAdjustment{}, errs.Newf(errs.InvalidArgument, "parse adjustedBy: %s", err)
	}

	var approvedBy *uuid.UUID
	if app.ApprovedBy != "" {
		id, err := uuid.Parse(app.ApprovedBy)
		if err != nil {
			return inventoryadjustmentbus.NewInventoryAdjustment{}, errs.Newf(errs.InvalidArgument, "parse approvedBy: %s", err)
		}
		approvedBy = &id
	}

	quantityChange, err := strconv.Atoi(app.QuantityChange)
	if err != nil {
		return inventoryadjustmentbus.NewInventoryAdjustment{}, errs.Newf(errs.InvalidArgument, "parse quantityChange: %s", err)
	}

	adjustmentDate, err := time.Parse(timeutil.FORMAT, app.AdjustmentDate)
	if err != nil {
		return inventoryadjustmentbus.NewInventoryAdjustment{}, errs.Newf(errs.InvalidArgument, "parse adjustmentDate: %s", err)
	}

	bus := inventoryadjustmentbus.NewInventoryAdjustment{
		ProductID:      productID,
		LocationID:     locationID,
		AdjustedBy:     adjustedBy,
		ApprovedBy:     approvedBy,
		QuantityChange: quantityChange,
		ReasonCode:     app.ReasonCode,
		Notes:          app.Notes,
		AdjustmentDate: adjustmentDate,
	}
	return bus, nil
}

// ApproveRequest contains information needed to approve an inventory adjustment.
type ApproveRequest struct {
	ApprovedBy string `json:"approved_by" validate:"required,min=36,max=36"`
}

func (app *ApproveRequest) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app ApproveRequest) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

// RejectRequest is the body for rejecting an inventory adjustment (no fields required).
type RejectRequest struct{}

func (app *RejectRequest) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app RejectRequest) Validate() error {
	return nil
}

type UpdateInventoryAdjustment struct {
	ProductID      *string `json:"product_id" validate:"omitempty,min=36,max=36"`
	LocationID     *string `json:"location_id" validate:"omitempty,min=36,max=36"`
	AdjustedBy     *string `json:"adjusted_by" validate:"omitempty,min=36,max=36"`
	ApprovedBy     *string `json:"approved_by" validate:"omitempty,min=36,max=36"`
	ApprovalStatus *string `json:"approval_status" validate:"omitempty,oneof=pending approved rejected"`
	QuantityChange *string `json:"quantity_change" validate:"omitempty"`
	ReasonCode     *string `json:"reason_code" validate:"omitempty"`
	Notes          *string `json:"notes" validate:"omitempty"`
	AdjustmentDate *string `json:"adjustment_date" validate:"omitempty"`
}

func (app *UpdateInventoryAdjustment) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateInventoryAdjustment) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateInventoryAdjustment(app UpdateInventoryAdjustment) (inventoryadjustmentbus.UpdateInventoryAdjustment, error) {
	bus := inventoryadjustmentbus.UpdateInventoryAdjustment{
		ReasonCode:     app.ReasonCode,
		Notes:          app.Notes,
		ApprovalStatus: app.ApprovalStatus,
	}

	if app.ProductID != nil {
		productID, err := uuid.Parse(*app.ProductID)
		if err != nil {
			return inventoryadjustmentbus.UpdateInventoryAdjustment{}, errs.Newf(errs.InvalidArgument, "parse productID: %s", err)
		}
		bus.ProductID = &productID
	}

	if app.LocationID != nil {
		locationID, err := uuid.Parse(*app.LocationID)
		if err != nil {
			return inventoryadjustmentbus.UpdateInventoryAdjustment{}, errs.Newf(errs.InvalidArgument, "parse locationID: %s", err)
		}
		bus.LocationID = &locationID
	}

	if app.AdjustedBy != nil {
		adjustedBy, err := uuid.Parse(*app.AdjustedBy)
		if err != nil {
			return inventoryadjustmentbus.UpdateInventoryAdjustment{}, errs.Newf(errs.InvalidArgument, "parse adjustedBy: %s", err)
		}
		bus.AdjustedBy = &adjustedBy
	}

	if app.ApprovedBy != nil {
		approvedBy, err := uuid.Parse(*app.ApprovedBy)
		if err != nil {
			return inventoryadjustmentbus.UpdateInventoryAdjustment{}, errs.Newf(errs.InvalidArgument, "parse approvedBy: %s", err)
		}
		bus.ApprovedBy = &approvedBy
	}

	if app.QuantityChange != nil {
		quantityChange, err := strconv.Atoi(*app.QuantityChange)
		if err != nil {
			return inventoryadjustmentbus.UpdateInventoryAdjustment{}, errs.Newf(errs.InvalidArgument, "parse quantityChange: %s", err)
		}
		bus.QuantityChange = &quantityChange
	}

	if app.AdjustmentDate != nil {
		adjustmentDate, err := time.Parse(timeutil.FORMAT, *app.AdjustmentDate)
		if err != nil {
			return inventoryadjustmentbus.UpdateInventoryAdjustment{}, errs.Newf(errs.InvalidArgument, "parse adjustmentDate: %s", err)
		}
		bus.AdjustmentDate = &adjustmentDate
	}

	return bus, nil
}
