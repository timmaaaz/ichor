package currencyapp

import (
	"encoding/json"
	"time"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
)

// QueryParams represents the set of possible query parameters for currency queries.
type QueryParams struct {
	Page     string
	Rows     string
	OrderBy  string
	ID       string
	Code     string
	Name     string
	IsActive string
}

// =============================================================================

// Currency represents information about an individual currency.
type Currency struct {
	ID            string  `json:"id"`
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Symbol        string  `json:"symbol"`
	Locale        string  `json:"locale"`
	DecimalPlaces int     `json:"decimal_places"`
	IsActive      bool    `json:"is_active"`
	SortOrder     int     `json:"sort_order"`
	CreatedBy     *string `json:"created_by"`
	CreatedDate   string  `json:"created_date"`
	UpdatedBy     *string `json:"updated_by"`
	UpdatedDate   string  `json:"updated_date"`
}

// Encode implements the encoder interface.
func (app Currency) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToAppCurrency converts a business currency to an app currency.
func ToAppCurrency(bus currencybus.Currency) Currency {
	var createdBy, updatedBy *string
	if bus.CreatedBy != nil {
		s := bus.CreatedBy.String()
		createdBy = &s
	}
	if bus.UpdatedBy != nil {
		s := bus.UpdatedBy.String()
		updatedBy = &s
	}

	return Currency{
		ID:            bus.ID.String(),
		Code:          bus.Code,
		Name:          bus.Name,
		Symbol:        bus.Symbol,
		Locale:        bus.Locale,
		DecimalPlaces: bus.DecimalPlaces,
		IsActive:      bus.IsActive,
		SortOrder:     bus.SortOrder,
		CreatedBy:     createdBy,
		CreatedDate:   bus.CreatedDate.Format(time.RFC3339),
		UpdatedBy:     updatedBy,
		UpdatedDate:   bus.UpdatedDate.Format(time.RFC3339),
	}
}

// ToAppCurrencies converts a slice of business currencies to app currencies.
func ToAppCurrencies(bus []currencybus.Currency) []Currency {
	app := make([]Currency, len(bus))
	for i, v := range bus {
		app[i] = ToAppCurrency(v)
	}
	return app
}

// =============================================================================

// Currencies is a collection wrapper that implements the Encoder interface.
type Currencies []Currency

// Encode implements the encoder interface.
func (app Currencies) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// =============================================================================

// NewCurrency contains information needed to create a new currency.
type NewCurrency struct {
	Code          string `json:"code" validate:"required,len=3"`
	Name          string `json:"name" validate:"required,min=1,max=100"`
	Symbol        string `json:"symbol" validate:"required,min=1,max=10"`
	Locale        string `json:"locale" validate:"required,min=2,max=10"`
	DecimalPlaces int    `json:"decimal_places" validate:"gte=0,lte=4"`
	IsActive      bool   `json:"is_active"`
	SortOrder     int    `json:"sort_order" validate:"gte=0"`
}

// Decode implements the decoder interface.
func (app *NewCurrency) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app NewCurrency) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewCurrency(app NewCurrency) currencybus.NewCurrency {
	return currencybus.NewCurrency{
		Code:          app.Code,
		Name:          app.Name,
		Symbol:        app.Symbol,
		Locale:        app.Locale,
		DecimalPlaces: app.DecimalPlaces,
		IsActive:      app.IsActive,
		SortOrder:     app.SortOrder,
	}
}

// =============================================================================

// UpdateCurrency contains information needed to update a currency.
type UpdateCurrency struct {
	Code          *string `json:"code" validate:"omitempty,len=3"`
	Name          *string `json:"name" validate:"omitempty,min=1,max=100"`
	Symbol        *string `json:"symbol" validate:"omitempty,min=1,max=10"`
	Locale        *string `json:"locale" validate:"omitempty,min=2,max=10"`
	DecimalPlaces *int    `json:"decimal_places" validate:"omitempty,gte=0,lte=4"`
	IsActive      *bool   `json:"is_active"`
	SortOrder     *int    `json:"sort_order" validate:"omitempty,gte=0"`
}

// Decode implements the decoder interface.
func (app *UpdateCurrency) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateCurrency) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateCurrency(app UpdateCurrency) currencybus.UpdateCurrency {
	return currencybus.UpdateCurrency{
		Code:          app.Code,
		Name:          app.Name,
		Symbol:        app.Symbol,
		Locale:        app.Locale,
		DecimalPlaces: app.DecimalPlaces,
		IsActive:      app.IsActive,
		SortOrder:     app.SortOrder,
	}
}
