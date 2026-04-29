// Package paperworkapp is a thin pass-through over paperworkbus. The bus
// owns cross-domain orchestration; the app layer translates bus error
// sentinels via mapBusErr and delegates rendering. App-layer request
// types mirror bus-layer types so future enrichment can land in the app
// layer without breaking the bus contract.
package paperworkapp

import "github.com/google/uuid"

// PickSheetRequest mirrors paperworkbus.PickSheetRequest at the app layer.
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
