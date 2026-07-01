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

// Labels is a slice wrapper used by converters and test fixtures.
type Labels []Label

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
	// Code length max=12 reflects the practical render budget on the
	// established Zebra GK420t 4"×6" media: ≈12 alphanumeric chars
	// fit a Code128 BY4 barcode at 812-dot width after start/check/stop
	// (~140 dots) + quiet zones (~80 dots) at 44 dots/char. Schema
	// column is VARCHAR(32) for future media flexibility; the
	// validator is tightened here so we fail loud at ingress instead
	// of silently producing labels with clipped barcodes.
	Code        string `json:"code" validate:"required,min=1,max=12"`
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
	// Code: see NewLabel.Code for the max=12 rationale.
	Code        *string `json:"code,omitempty" validate:"omitempty,min=1,max=12"`
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
// until D-001/D-002 ship. NewLabel/UpdateLabel oneof tags also still
// accept "receiving" and "pick" for catalog parity with pre-0g.B1
// rows; Phase 0g.B2 deletes those rows and tightens both validators.
//
// Code populates the human-readable + barcode field for location and
// container labels (which don't carry a structured payload). Payload
// is the JSON body for product labels. Validate() enforces that the
// right field is present for the requested Type.
type RenderPrintRequest struct {
	Type    string          `json:"type" validate:"required,oneof=location container product"`
	// Code: see NewLabel.Code for the max=12 rationale.
	Code    string          `json:"code,omitempty" validate:"omitempty,max=12"`
	Payload json.RawMessage `json:"payload,omitempty"`
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
	switch app.Type {
	case "location", "container":
		if app.Code == "" {
			return errs.Newf(errs.InvalidArgument, "validate: code is required for type=%s", app.Type)
		}
	case "product":
		if len(app.Payload) == 0 {
			return errs.Newf(errs.InvalidArgument, "validate: payload is required for type=product")
		}
	}
	if len(app.Payload) > maxRenderPayloadBytes {
		return errs.Newf(errs.InvalidArgument, "validate: payload exceeds %d bytes", maxRenderPayloadBytes)
	}
	return nil
}
