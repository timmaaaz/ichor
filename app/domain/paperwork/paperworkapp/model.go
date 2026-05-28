// Package paperworkapp orchestrates cross-domain reads for paperwork
// rendering. It holds the sibling buses, assembles pdf.*Data, and delegates
// rendering to the pure pdf leaf at business/domain/paperwork/pdf/.
// Orchestration was relocated here from the deleted paperworkbus in Phase
// 0g.F4-enrichment so cross-domain reads live in the app layer, matching
// directedworkapp/supervisorkpiapp/scanapp.
package paperworkapp

import "github.com/google/uuid"

// PickSheetRequest carries inputs for rendering a pick sheet PDF.
//
// Zone is optional. Today it is rendered in the sheet header only; per-line
// filtering to pick locations within the named zone is deferred to phase
// 0g.F4 PR2 (it needs the location→zone mapping not yet wired here).
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
