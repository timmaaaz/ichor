// Package paperworkbus renders paperwork PDFs (pick sheets, receive cover
// sheets, transfer sheets) for floor-worker workflows. Build* methods
// query sibling domains (sales orders, purchase orders, transfers),
// transform the result into pdf-package data types, and delegate to the
// pdf subpackage for rendering. The bus is renderer-shaped: 5 sibling-bus
// pointers + log, no Storer, no DB writes.
package paperworkbus

import "github.com/google/uuid"

// PickSheetRequest carries inputs for rendering a pick sheet PDF.
//
// Zone is optional. When non-empty the resulting sheet is filtered to lines
// whose pick locations fall within the named zone. Filtering is inert until
// line items are populated (deferred to phase 0g.F4).
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
