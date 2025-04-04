package inventorytransactionapp

import (
	"encoding/json"
	"fmt"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/movement/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/sdk/convert"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string

	InventoryTransactionID string
	ProductID              string
	LocationID             string
	UserID                 string
	Quantity               string
	TransactionType        string
	ReferenceNumber        string
	TransactionDate        string
	CreatedDate            string
	UpdatedDate            string
}

type InventoryTransaction struct {
	InventoryTransactionID string `json:"transaction_id"`
	ProductID              string `json:"product_id"`
	LocationID             string `json:"location_id"`
	UserID                 string `json:"user_id"`
	Quantity               string `json:"quantity"`
	TransactionType        string `json:"transaction_type"`
	ReferenceNumber        string `json:"reference_number"`
	TransactionDate        string `json:"transaction_date"`
	CreatedDate            string `json:"created_date"`
	UpdatedDate            string `json:"updated_date"`
}

func (app InventoryTransaction) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppInventoryTransaction(bus inventorytransactionbus.InventoryTransaction) InventoryTransaction {
	return InventoryTransaction{
		InventoryTransactionID: bus.InventoryTransactionID.String(),
		ProductID:              bus.ProductID.String(),
		LocationID:             bus.LocationID.String(),
		UserID:                 bus.UserID.String(),
		Quantity:               fmt.Sprintf("%d", bus.Quantity),
		TransactionType:        bus.TransactionType,
		ReferenceNumber:        bus.ReferenceNumber,
		TransactionDate:        bus.TransactionDate.Format(timeutil.FORMAT),
		CreatedDate:            bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:            bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

func ToAppInventoryTransactions(bus []inventorytransactionbus.InventoryTransaction) []InventoryTransaction {
	app := make([]InventoryTransaction, len(bus))
	for i, v := range bus {
		app[i] = ToAppInventoryTransaction(v)
	}
	return app
}

type NewInventoryTransaction struct {
	ProductID       string `json:"product_id" validate:"required,min=36,max=36"`
	LocationID      string `json:"location_id" validate:"required,min=36,max=36"`
	UserID          string `json:"user_id" validate:"required,min=36,max=36"`
	Quantity        string `json:"quantity" validate:"required"`
	TransactionType string `json:"transaction_type" validate:"required"`
	ReferenceNumber string `json:"reference_number" validate:"required"`
	TransactionDate string `json:"transaction_date" validate:"required"`
}

func (app *NewInventoryTransaction) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewInventoryTransaction) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewInventoryTransaction(app NewInventoryTransaction) (inventorytransactionbus.NewInventoryTransaction, error) {
	dest := inventorytransactionbus.NewInventoryTransaction{}
	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return inventorytransactionbus.NewInventoryTransaction{}, fmt.Errorf("toBusNewInventoryTransaction: %w", err)
	}

	return dest, nil
}

type UpdateInventoryTransaction struct {
	ProductID       *string `json:"product_id" validate:"omitempty,min=36,max=36"`
	LocationID      *string `json:"location_id" validate:"omitempty,min=36,max=36"`
	UserID          *string `json:"user_id" validate:"omitempty,min=36,max=36"`
	Quantity        *string `json:"quantity" validate:"omitempty"`
	TransactionType *string `json:"transaction_type" validate:"omitempty"`
	ReferenceNumber *string `json:"reference_number" validate:"omitempty"`
	TransactionDate *string `json:"transaction_date" validate:"omitempty"`
}

func (app *UpdateInventoryTransaction) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateInventoryTransaction) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateInventoryTransaction(app UpdateInventoryTransaction) (inventorytransactionbus.UpdateInventoryTransaction, error) {
	dest := inventorytransactionbus.UpdateInventoryTransaction{}
	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return inventorytransactionbus.UpdateInventoryTransaction{}, fmt.Errorf("toBusUpdateInventoryTransaction: %w", err)
	}

	return dest, nil
}
