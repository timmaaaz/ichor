package inventoryadjustmentapp

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/movement/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/sdk/convert"
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
	QuantityChange        string
	ReasonCode            string
	Notes                 string
	AdjustmentDate        string
	CreatedDate           string
	UpdatedDate           string
}

type InventoryAdjustment struct {
	InventoryAdjustmentID string `json:"adjustment_id"`
	ProductID             string `json:"product_id"`
	LocationID            string `json:"location_id"`
	AdjustedBy            string `json:"adjusted_by"`
	ApprovedBy            string `json:"approved_by"`
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
	return InventoryAdjustment{
		InventoryAdjustmentID: bus.InventoryAdjustmentID.String(),
		ProductID:             bus.ProductID.String(),
		LocationID:            bus.LocationID.String(),
		AdjustedBy:            bus.AdjustedBy.String(),
		ApprovedBy:            bus.ApprovedBy.String(),
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
	ApprovedBy     string `json:"approved_by" validate:"required,min=36,max=36"`
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
	dest := inventoryadjustmentbus.NewInventoryAdjustment{}

	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return inventoryadjustmentbus.NewInventoryAdjustment{}, fmt.Errorf("toBusNewInventoryTransaction: %w", err)
	}

	return dest, nil
}

type UpdateInventoryAdjustment struct {
	ProductID      *string `json:"product_id" validate:"omitempty,min=36,max=36"`
	LocationID     *string `json:"location_id" validate:"omitempty,min=36,max=36"`
	AdjustedBy     *string `json:"adjusted_by" validate:"omitempty,min=36,max=36"`
	ApprovedBy     *string `json:"approved_by" validate:"omitempty,min=36,max=36"`
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
	dest := inventoryadjustmentbus.UpdateInventoryAdjustment{}

	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return inventoryadjustmentbus.UpdateInventoryAdjustment{}, fmt.Errorf("toBusUpdateInventoryAdjustment: %w", err)
	}

	return dest, nil
}
