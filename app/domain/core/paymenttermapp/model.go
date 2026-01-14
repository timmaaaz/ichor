package paymenttermapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/paymenttermbus"
)

// QueryParams represents the query parameters that can be used.
type QueryParams struct {
	Page        string
	Rows        string
	OrderBy     string
	ID          string
	Name        string
	Description string
}

type PaymentTerm struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Encode implements the encoder interface.
func (app PaymentTerm) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppPaymentTerm(bus paymenttermbus.PaymentTerm) PaymentTerm {
	return PaymentTerm{
		ID:          bus.ID.String(),
		Name:        bus.Name,
		Description: bus.Description,
	}
}

func ToAppPaymentTerms(bus []paymenttermbus.PaymentTerm) []PaymentTerm {
	app := make([]PaymentTerm, len(bus))
	for i, v := range bus {
		app[i] = ToAppPaymentTerm(v)
	}
	return app
}

// =============================================================================

type NewPaymentTerm struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"omitempty,min=3,max=500"`
}

// Decode implements the decoder interface.
func (app *NewPaymentTerm) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app NewPaymentTerm) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func ToBusNewPaymentTerm(app NewPaymentTerm) paymenttermbus.NewPaymentTerm {
	return paymenttermbus.NewPaymentTerm{
		Name:        app.Name,
		Description: app.Description,
	}
}

// =============================================================================

type UpdatePaymentTerm struct {
	Name        *string `json:"name" validate:"omitempty,min=3,max=100"`
	Description *string `json:"description" validate:"omitempty,min=3,max=500"`
}

// Decode implements the decoder interface.
func (app *UpdatePaymentTerm) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app UpdatePaymentTerm) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func ToBusUpdatePaymentTerm(app UpdatePaymentTerm) paymenttermbus.UpdatePaymentTerm {
	return paymenttermbus.UpdatePaymentTerm{
		Name:        app.Name,
		Description: app.Description,
	}
}
