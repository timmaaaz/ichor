// Package paperworkapp orchestrates paperwork rendering across domain
// dependencies. Phase 0g.B2 is a thin pass-through; B3 adds cross-domain
// data fetches before delegating to paperworkbus.
package paperworkapp

import "github.com/google/uuid"

// PickSheetRequest mirrors paperworkbus.PickSheetRequest at the app layer.
// Reserved for app-only enrichment (e.g. caller identity) in B3.
type PickSheetRequest struct {
	OrderID uuid.UUID
	Zone    string
}

// ReceiveCoverRequest mirrors paperworkbus.ReceiveCoverRequest at the app layer.
type ReceiveCoverRequest struct {
	PurchaseOrderID uuid.UUID
}

// TransferSheetRequest mirrors paperworkbus.TransferSheetRequest at the app layer.
type TransferSheetRequest struct {
	TransferID uuid.UUID
}
