package paperworkbus

import (
	"context"
	"errors"

	"github.com/timmaaaz/ichor/foundation/logger"
)

// ErrNotImplemented is returned by every Build* method during the B2 scaffold
// phase. Phase 0g.B3 replaces these stubs with real PDF rendering.
var ErrNotImplemented = errors.New("paperwork: not implemented")

// Business manages paperwork rendering. Phase 0g.B2 holds only a logger;
// Phase 0g.B3 will add cross-domain dependencies (sales orders, purchase
// orders, transfers) when the rendering bodies land.
type Business struct {
	log *logger.Logger
}

// NewBusiness constructs a paperwork business API for use.
func NewBusiness(log *logger.Logger) *Business {
	return &Business{log: log}
}

// BuildPickSheet renders a pick sheet PDF for the given order. Returns
// ErrNotImplemented during B2; B3 fills in the rendering body.
func (b *Business) BuildPickSheet(ctx context.Context, req PickSheetRequest) ([]byte, error) {
	return nil, ErrNotImplemented
}

// BuildReceiveCover renders a receive-cover PDF for the given purchase order.
// Returns ErrNotImplemented during B2; B3 fills in the rendering body.
func (b *Business) BuildReceiveCover(ctx context.Context, req ReceiveCoverRequest) ([]byte, error) {
	return nil, ErrNotImplemented
}

// BuildTransferSheet renders a transfer sheet PDF for the given transfer.
// Returns ErrNotImplemented during B2; B3 fills in the rendering body.
func (b *Business) BuildTransferSheet(ctx context.Context, req TransferSheetRequest) ([]byte, error) {
	return nil, ErrNotImplemented
}
