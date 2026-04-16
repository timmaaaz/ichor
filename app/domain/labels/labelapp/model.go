package labelapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/labels/labelbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

// QueryParams filters catalog labels via the list endpoint.
type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string
	Type    string
}

// Label is the API-facing shape of a catalog label.
type Label struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Type        string `json:"type"`
	EntityRef   string `json:"entity_ref,omitempty"`
	PayloadJSON string `json:"payload_json,omitempty"`
	CreatedDate string `json:"created_date"`
}

// Encode implements the web.Encoder interface.
func (app Label) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// Labels is a slice wrapper so it implements web.Encoder directly.
type Labels []Label

// Encode implements the web.Encoder interface.
func (app Labels) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func toAppLabel(bus labelbus.LabelCatalog) Label {
	return Label{
		ID:          bus.ID.String(),
		Code:        bus.Code,
		Type:        bus.Type,
		EntityRef:   bus.EntityRef,
		PayloadJSON: bus.PayloadJSON,
		CreatedDate: bus.CreatedDate.Format(timeutil.FORMAT),
	}
}

func toAppLabels(bus []labelbus.LabelCatalog) Labels {
	out := make(Labels, len(bus))
	for i, v := range bus {
		out[i] = toAppLabel(v)
	}
	return out
}

// =============================================================================

// PrintRequest is the body of POST /v1/labels/print. A catalog label row is
// looked up by LabelID, rendered to ZPL, and dispatched to the printer.
type PrintRequest struct {
	LabelID string `json:"label_id" validate:"required,min=36,max=36"`
	Copies  int    `json:"copies,omitempty" validate:"omitempty,min=1,max=100"`
}

// Decode implements the decoder interface.
func (app *PrintRequest) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app PrintRequest) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

// RenderPrintRequest is the body of POST /v1/labels/render-print. A
// transaction-label payload is rendered in-memory and dispatched to the
// printer — no catalog row, no DB write.
type RenderPrintRequest struct {
	Type    string          `json:"type" validate:"required,oneof=receiving pick"`
	Payload json.RawMessage `json:"payload" validate:"required"`
	Copies  int             `json:"copies,omitempty" validate:"omitempty,min=1,max=100"`
}

// Decode implements the decoder interface.
func (app *RenderPrintRequest) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app RenderPrintRequest) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	if len(app.Payload) == 0 {
		return errs.Newf(errs.InvalidArgument, "validate: payload is empty")
	}
	return nil
}
