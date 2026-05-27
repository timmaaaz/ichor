// Package paperworkapp orchestrates cross-domain reads for paperwork
// rendering. It holds the sibling buses, assembles pdf.*Data, and delegates
// rendering to the pure pdf leaf at business/domain/paperwork/pdf/.
// Orchestration was relocated here from paperworkbus in Phase 0g.F4-enrichment
// (reverses D-CONV-3; see docs/arch/paperwork.md).
package paperworkapp

import "github.com/google/uuid"

// PickSheetRequest carries inputs for rendering a pick sheet PDF.
//
// Zone is optional. When non-empty the resulting sheet is filtered to lines
// whose pick locations fall within the named zone. Filtering is inert until
// line items are populated (deferred to phase 0g.F4 PR2).
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
