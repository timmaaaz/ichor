// Package paperworkbus provides business access to PDF paperwork generation
// for warehouse workflows (pick sheets, receive cover sheets, transfer sheets).
//
// Paperwork is a renderer, not a persister: there is no Storer interface and
// no DB writes. The bus orchestrates rendering of PDF bytes from line-item
// data fetched from sibling domains (sales orders, purchase orders, transfers).
//
// Phase 0g.B2 ships scaffolding only — every Build* method returns
// ErrNotImplemented. Phase 0g.B3 wires gofpdf + boombuler/barcode and fills
// in the rendering bodies.
package paperworkbus

import "github.com/google/uuid"

// PickSheetRequest carries inputs for rendering a pick sheet PDF.
//
// Zone is optional. When non-empty the resulting sheet is filtered to lines
// whose pick locations fall within the named zone (Phase 0g.B3 behavior).
type PickSheetRequest struct {
	OrderID uuid.UUID
	Zone    string
}

// ReceiveCoverRequest carries inputs for rendering a receive-cover PDF.
type ReceiveCoverRequest struct {
	PurchaseOrderID uuid.UUID
}

// TransferSheetRequest carries inputs for rendering a transfer sheet PDF.
type TransferSheetRequest struct {
	TransferID uuid.UUID
}
