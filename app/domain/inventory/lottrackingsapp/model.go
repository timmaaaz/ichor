package lottrackingsapp

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/lottrackingsbus"
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
	ExpiryBefore      string
	ExpiryAfter       string
	RecievedDate      string
	Quantity          string
	QualityStatus     string
	CreatedDate       string
	UpdatedDate       string
}

type LotTrackings struct {
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

func (app LotTrackings) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppLotTracking(bus lottrackingsbus.LotTrackings) LotTrackings {
	return LotTrackings{
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

func ToAppLotTrackings(bus []lottrackingsbus.LotTrackings) []LotTrackings {
	app := make([]LotTrackings, len(bus))
	for i, v := range bus {
		app[i] = ToAppLotTracking(v)
	}
	return app
}

type NewLotTrackings struct {
	SupplierProductID string `json:"supplier_product_id" validate:"required,min=36,max=36"`
	LotNumber         string `json:"lot_number" validate:"required"`
	ManufactureDate   string `json:"manufacture_date" validate:"required"`
	ExpirationDate    string `json:"expiration_date" validate:"required"`
	RecievedDate      string `json:"received_date" validate:"required"`
	Quantity          string `json:"quantity" validate:"required"`
	QualityStatus     string `json:"quality_status" validate:"required"`
}

func (app *NewLotTrackings) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewLotTrackings) Validate() error {
	if err := errs.Check(app); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

func toBusNewLotTrackings(app NewLotTrackings) (lottrackingsbus.NewLotTrackings, error) {
	supplierProductID, err := uuid.Parse(app.SupplierProductID)
	if err != nil {
		return lottrackingsbus.NewLotTrackings{}, errs.Newf(errs.InvalidArgument, "parse supplierProductID: %s", err)
	}

	manufactureDate, err := time.Parse(timeutil.FORMAT, app.ManufactureDate)
	if err != nil {
		return lottrackingsbus.NewLotTrackings{}, errs.Newf(errs.InvalidArgument, "parse manufactureDate: %s", err)
	}

	expirationDate, err := time.Parse(timeutil.FORMAT, app.ExpirationDate)
	if err != nil {
		return lottrackingsbus.NewLotTrackings{}, errs.Newf(errs.InvalidArgument, "parse expirationDate: %s", err)
	}

	receivedDate, err := time.Parse(timeutil.FORMAT, app.RecievedDate)
	if err != nil {
		return lottrackingsbus.NewLotTrackings{}, errs.Newf(errs.InvalidArgument, "parse receivedDate: %s", err)
	}

	quantity, err := strconv.Atoi(app.Quantity)
	if err != nil {
		return lottrackingsbus.NewLotTrackings{}, errs.Newf(errs.InvalidArgument, "parse quantity: %s", err)
	}

	bus := lottrackingsbus.NewLotTrackings{
		SupplierProductID: supplierProductID,
		LotNumber:         app.LotNumber,
		ManufactureDate:   manufactureDate,
		ExpirationDate:    expirationDate,
		RecievedDate:      receivedDate,
		Quantity:          quantity,
		QualityStatus:     app.QualityStatus,
	}
	return bus, nil
}

type UpdateLotTrackings struct {
	SupplierProductID *string `json:"supplier_product_id" validate:"omitempty,min=36,max=36"`
	LotNumber         *string `json:"lot_number"`
	ManufactureDate   *string `json:"manufacture_date"`
	ExpirationDate    *string `json:"expiration_date"`
	RecievedDate      *string `json:"received_date"`
	Quantity          *string `json:"quantity"`
	QualityStatus     *string `json:"quality_status"`
}

func (app *UpdateLotTrackings) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateLotTrackings) Validate() error {
	if err := errs.Check(app); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

func toBusUpdateLotTrackings(app UpdateLotTrackings) (lottrackingsbus.UpdateLotTrackings, error) {
	bus := lottrackingsbus.UpdateLotTrackings{
		LotNumber:     app.LotNumber,
		QualityStatus: app.QualityStatus,
	}

	if app.SupplierProductID != nil {
		supplierProductID, err := uuid.Parse(*app.SupplierProductID)
		if err != nil {
			return lottrackingsbus.UpdateLotTrackings{}, errs.Newf(errs.InvalidArgument, "parse supplierProductID: %s", err)
		}
		bus.SupplierProductID = &supplierProductID
	}

	if app.ManufactureDate != nil {
		manufactureDate, err := time.Parse(timeutil.FORMAT, *app.ManufactureDate)
		if err != nil {
			return lottrackingsbus.UpdateLotTrackings{}, errs.Newf(errs.InvalidArgument, "parse manufactureDate: %s", err)
		}
		bus.ManufactureDate = &manufactureDate
	}

	if app.ExpirationDate != nil {
		expirationDate, err := time.Parse(timeutil.FORMAT, *app.ExpirationDate)
		if err != nil {
			return lottrackingsbus.UpdateLotTrackings{}, errs.Newf(errs.InvalidArgument, "parse expirationDate: %s", err)
		}
		bus.ExpirationDate = &expirationDate
	}

	if app.RecievedDate != nil {
		receivedDate, err := time.Parse(timeutil.FORMAT, *app.RecievedDate)
		if err != nil {
			return lottrackingsbus.UpdateLotTrackings{}, errs.Newf(errs.InvalidArgument, "parse receivedDate: %s", err)
		}
		bus.RecievedDate = &receivedDate
	}

	if app.Quantity != nil {
		quantity, err := strconv.Atoi(*app.Quantity)
		if err != nil {
			return lottrackingsbus.UpdateLotTrackings{}, errs.Newf(errs.InvalidArgument, "parse quantity: %s", err)
		}
		bus.Quantity = &quantity
	}

	return bus, nil
}
