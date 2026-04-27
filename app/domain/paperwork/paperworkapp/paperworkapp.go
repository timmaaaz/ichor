package paperworkapp

import (
	"context"

	"github.com/timmaaaz/ichor/business/domain/paperwork/paperworkbus"
)

// App is the application layer for paperwork. Currently a thin wrapper over
// paperworkbus; B3 expands it with cross-domain orchestration.
type App struct {
	bus *paperworkbus.Business
}

// NewApp constructs an App.
func NewApp(bus *paperworkbus.Business) *App {
	return &App{bus: bus}
}

// BuildPickSheet delegates to the bus.
func (a *App) BuildPickSheet(ctx context.Context, req PickSheetRequest) ([]byte, error) {
	return a.bus.BuildPickSheet(ctx, paperworkbus.PickSheetRequest{
		OrderID: req.OrderID,
		Zone:    req.Zone,
	})
}

// BuildReceiveCover delegates to the bus.
func (a *App) BuildReceiveCover(ctx context.Context, req ReceiveCoverRequest) ([]byte, error) {
	return a.bus.BuildReceiveCover(ctx, paperworkbus.ReceiveCoverRequest{
		PurchaseOrderID: req.PurchaseOrderID,
	})
}

// BuildTransferSheet delegates to the bus.
func (a *App) BuildTransferSheet(ctx context.Context, req TransferSheetRequest) ([]byte, error) {
	return a.bus.BuildTransferSheet(ctx, paperworkbus.TransferSheetRequest{
		TransferID: req.TransferID,
	})
}
