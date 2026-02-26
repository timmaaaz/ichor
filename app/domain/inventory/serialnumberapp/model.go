package serialnumberapp

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/serialnumberbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

// SerialLocation is the app-layer representation of a serial number's storage location.
type SerialLocation struct {
	LocationID    string `json:"location_id"`
	LocationCode  string `json:"location_code"`
	Aisle         string `json:"aisle"`
	Rack          string `json:"rack"`
	Shelf         string `json:"shelf"`
	Bin           string `json:"bin"`
	WarehouseName string `json:"warehouse_name"`
	ZoneName      string `json:"zone_name"`
}

func (sl SerialLocation) Encode() ([]byte, string, error) {
	data, err := json.Marshal(sl)
	return data, "application/json", err
}

func toAppSerialLocation(bus serialnumberbus.SerialLocation) SerialLocation {
	return SerialLocation{
		LocationID:    bus.LocationID.String(),
		LocationCode:  bus.LocationCode,
		Aisle:         bus.Aisle,
		Rack:          bus.Rack,
		Shelf:         bus.Shelf,
		Bin:           bus.Bin,
		WarehouseName: bus.WarehouseName,
		ZoneName:      bus.ZoneName,
	}
}

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string

	SerialID     string
	LotID        string
	ProductID    string
	LocationID   string
	SerialNumber string
	Status       string
	CreatedDate  string
	UpdatedDate  string
}

type SerialNumber struct {
	SerialID     string `json:"serial_id"`
	LotID        string `json:"lot_id"`
	ProductID    string `json:"product_id"`
	LocationID   string `json:"location_id"`
	SerialNumber string `json:"serial_number"`
	Status       string `json:"status"`
	CreatedDate  string `json:"created_date"`
	UpdatedDate  string `json:"updated_date"`
}

func (app SerialNumber) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppSerialNumber(bus serialnumberbus.SerialNumber) SerialNumber {
	return SerialNumber{
		SerialID:     bus.SerialID.String(),
		LotID:        bus.LotID.String(),
		ProductID:    bus.ProductID.String(),
		LocationID:   bus.LocationID.String(),
		SerialNumber: bus.SerialNumber,
		Status:       bus.Status,
		CreatedDate:  bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:  bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

func ToAppSerialNumbers(bus []serialnumberbus.SerialNumber) []SerialNumber {
	app := make([]SerialNumber, len(bus))
	for i, v := range bus {
		app[i] = ToAppSerialNumber(v)
	}
	return app
}

type NewSerialNumber struct {
	LotID        string `json:"lot_id" validate:"required,min=36,max=36"`
	ProductID    string `json:"product_id" validate:"required,min=36,max=36"`
	LocationID   string `json:"location_id" validate:"required,min=36,max=36"`
	SerialNumber string `json:"serial_number" validate:"required"`
	Status       string `json:"status" validate:"required"`
}

func (app *NewSerialNumber) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewSerialNumber) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewSerialNumber(app NewSerialNumber) (serialnumberbus.NewSerialNumber, error) {
	lotID, err := uuid.Parse(app.LotID)
	if err != nil {
		return serialnumberbus.NewSerialNumber{}, errs.Newf(errs.InvalidArgument, "parse lotID: %s", err)
	}

	productID, err := uuid.Parse(app.ProductID)
	if err != nil {
		return serialnumberbus.NewSerialNumber{}, errs.Newf(errs.InvalidArgument, "parse productID: %s", err)
	}

	locationID, err := uuid.Parse(app.LocationID)
	if err != nil {
		return serialnumberbus.NewSerialNumber{}, errs.Newf(errs.InvalidArgument, "parse locationID: %s", err)
	}

	bus := serialnumberbus.NewSerialNumber{
		LotID:        lotID,
		ProductID:    productID,
		LocationID:   locationID,
		SerialNumber: app.SerialNumber,
		Status:       app.Status,
	}
	return bus, nil
}

type UpdateSerialNumber struct {
	LotID        *string `json:"lot_id" validate:"omitempty,min=36,max=36"`
	ProductID    *string `json:"product_id" validate:"omitempty,min=36,max=36"`
	LocationID   *string `json:"location_id" validate:"omitempty,min=36,max=36"`
	SerialNumber *string `json:"serial_number" validate:"omitempty"`
	Status       *string `json:"status" validate:"omitempty"`
}

func (app *UpdateSerialNumber) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateSerialNumber) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateSerialNumber(app UpdateSerialNumber) (serialnumberbus.UpdateSerialNumber, error) {
	bus := serialnumberbus.UpdateSerialNumber{
		SerialNumber: app.SerialNumber,
		Status:       app.Status,
	}

	if app.LotID != nil {
		lotID, err := uuid.Parse(*app.LotID)
		if err != nil {
			return serialnumberbus.UpdateSerialNumber{}, errs.Newf(errs.InvalidArgument, "parse lotID: %s", err)
		}
		bus.LotID = &lotID
	}

	if app.ProductID != nil {
		productID, err := uuid.Parse(*app.ProductID)
		if err != nil {
			return serialnumberbus.UpdateSerialNumber{}, errs.Newf(errs.InvalidArgument, "parse productID: %s", err)
		}
		bus.ProductID = &productID
	}

	if app.LocationID != nil {
		locationID, err := uuid.Parse(*app.LocationID)
		if err != nil {
			return serialnumberbus.UpdateSerialNumber{}, errs.Newf(errs.InvalidArgument, "parse locationID: %s", err)
		}
		bus.LocationID = &locationID
	}

	return bus, nil
}
