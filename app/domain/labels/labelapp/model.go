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

// ToAppLabels exposes the internal slice converter so seed harnesses can
// pre-shape integration test fixtures.
func ToAppLabels(bus []labelbus.LabelCatalog) Labels {
	return toAppLabels(bus)
}

// =============================================================================

// NewLabel is the input shape for POST /v1/labels (catalog row creation).
type NewLabel struct {
	Code        string `json:"code" validate:"required,min=1,max=32"`
	Type        string `json:"type" validate:"required,oneof=location container lot serial product receiving pick"`
	EntityRef   string `json:"entity_ref,omitempty" validate:"omitempty,max=128"`
	PayloadJSON string `json:"payload_json,omitempty"`
}

// Decode implements the decoder interface.
func (app *NewLabel) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks that the new label is well-formed.
func (app NewLabel) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewLabel(app NewLabel) labelbus.NewLabelCatalog {
	return labelbus.NewLabelCatalog{
		Code:        app.Code,
		Type:        app.Type,
		EntityRef:   app.EntityRef,
		PayloadJSON: app.PayloadJSON,
	}
}

// UpdateLabel carries optional patch fields for PUT /v1/labels/{label_id}.
type UpdateLabel struct {
	Code        *string `json:"code,omitempty" validate:"omitempty,min=1,max=32"`
	Type        *string `json:"type,omitempty" validate:"omitempty,oneof=location container lot serial product receiving pick"`
	EntityRef   *string `json:"entity_ref,omitempty" validate:"omitempty,max=128"`
	PayloadJSON *string `json:"payload_json,omitempty"`
}

// Decode implements the decoder interface.
func (app *UpdateLabel) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateLabel) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateLabel(app UpdateLabel) labelbus.UpdateLabelCatalog {
	return labelbus.UpdateLabelCatalog{
		Code:        app.Code,
		Type:        app.Type,
		EntityRef:   app.EntityRef,
		PayloadJSON: app.PayloadJSON,
	}
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
// printer — no catalog row, no DB write. Type must be a renderable
// label type (location, container, product); lot/serial are accepted
// by NewLabel/UpdateLabel for catalog row creation but cannot render
// until D-001/D-002 ship.
type RenderPrintRequest struct {
	Type    string          `json:"type" validate:"required,oneof=location container product"`
	Payload json.RawMessage `json:"payload" validate:"required"`
	Copies  int             `json:"copies,omitempty" validate:"omitempty,min=1,max=100"`
}

// Decode implements the decoder interface.
func (app *RenderPrintRequest) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// maxRenderPayloadBytes caps the raw JSON payload accepted by the
// render-print endpoint. Realistic label payloads (product with long
// names + lot/UPC) are well under 1KB; 64KB gives generous headroom
// while preventing an authenticated client from POSTing a multi-MB
// body through the validate:"required" tag alone.
const maxRenderPayloadBytes = 64 * 1024

// Validate checks the data in the model is considered clean.
func (app RenderPrintRequest) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	if len(app.Payload) == 0 {
		return errs.Newf(errs.InvalidArgument, "validate: payload is empty")
	}
	if len(app.Payload) > maxRenderPayloadBytes {
		return errs.Newf(errs.InvalidArgument, "validate: payload exceeds %d bytes", maxRenderPayloadBytes)
	}
	return nil
}
