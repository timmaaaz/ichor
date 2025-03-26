package lottrackingapp

import (
	"encoding/json"
	"fmt"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/lot/lottrackingbus"
	"github.com/timmaaaz/ichor/business/sdk/convert"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string

	LotID             string
	SupplierProductID string
	LotNumber         string
	ManufactureDate   string
	ExpirationDate    string
	RecievedDate      string
	Quantity          string
	QualityStatus     string
	CreatedDate       string
	UpdatedDate       string
}

type LotTracking struct {
	LotID             string `json:"lot_id"`
	SupplierProductID string `json:"supplier_product_id"`
	LotNumber         string `json:"lot_number"`
	ManufactureDate   string `json:"manufacture_date"`
	ExpirationDate    string `json:"expiration_date"`
	RecievedDate      string `json:"received_date"`
	Quantity          string `json:"quantity"`
	QualityStatus     string `json:"quality_status"`
	CreatedDate       string `json:"created_date"`
	UpdatedDate       string `json:"updated_date"`
}

func (app LotTracking) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppLotTracking(bus lottrackingbus.LotTracking) LotTracking {
	return LotTracking{
		LotID:             bus.LotID.String(),
		SupplierProductID: bus.SupplierProductID.String(),
		LotNumber:         bus.LotNumber,
		ManufactureDate:   bus.ManufactureDate.Format(timeutil.FORMAT),
		ExpirationDate:    bus.ExpirationDate.Format(timeutil.FORMAT),
		RecievedDate:      bus.RecievedDate.Format(timeutil.FORMAT),
		Quantity:          fmt.Sprintf("%d", bus.Quantity),
		QualityStatus:     bus.QualityStatus,
		CreatedDate:       bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:       bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

func ToAppLotTrackings(bus []lottrackingbus.LotTracking) []LotTracking {
	app := make([]LotTracking, len(bus))
	for i, v := range bus {
		app[i] = ToAppLotTracking(v)
	}
	return app
}

type NewLotTracking struct {
	SupplierProductID string `json:"supplier_product_id" validate:"required,min=36,max=36"`
	LotNumber         string `json:"lot_number" validate:"required"`
	ManufactureDate   string `json:"manufacture_date" validate:"required"`
	ExpirationDate    string `json:"expiration_date" validate:"required"`
	RecievedDate      string `json:"received_date" validate:"required"`
	Quantity          string `json:"quantity" validate:"required"`
	QualityStatus     string `json:"quality_status" validate:"required"`
}

func (app *NewLotTracking) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewLotTracking) Validate() error {
	if err := errs.Check(app); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

func toBusNewLotTracking(app NewLotTracking) (lottrackingbus.NewLotTracking, error) {
	dest := lottrackingbus.NewLotTracking{}
	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return lottrackingbus.NewLotTracking{}, fmt.Errorf("toBusNewLotTracking: %w", err)
	}

	return dest, nil
}

type UpdateLotTracking struct {
	SupplierProductID *string `json:"supplier_product_id" validate:"omitempty,min=36,max=36"`
	LotNumber         *string `json:"lot_number"`
	ManufactureDate   *string `json:"manufacture_date"`
	ExpirationDate    *string `json:"expiration_date"`
	RecievedDate      *string `json:"received_date"`
	Quantity          *string `json:"quantity"`
	QualityStatus     *string `json:"quality_status"`
}

func (app *UpdateLotTracking) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateLotTracking) Validate() error {
	if err := errs.Check(app); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

func toBusUpdateLotTracking(app UpdateLotTracking) (lottrackingbus.UpdateLotTracking, error) {
	dest := lottrackingbus.UpdateLotTracking{}
	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return lottrackingbus.UpdateLotTracking{}, fmt.Errorf("toBusUpdateLotTracking: %w", err)
	}

	return dest, nil
}
