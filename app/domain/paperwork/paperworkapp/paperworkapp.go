package paperworkapp

import (
	"context"
	"errors"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/paperwork/paperworkbus"
)

// App is a thin wrapper over paperworkbus.Business. Cross-domain
// orchestration (querying sibling buses for orders/POs/transfers)
// lives in the bus per the renderer-shaped design (D-CONV-3); this
// layer translates errors and delegates rendering.
type App struct {
	bus *paperworkbus.Business
}

// NewApp constructs an App.
func NewApp(bus *paperworkbus.Business) *App {
	return &App{bus: bus}
}

// BuildPickSheet delegates to the bus and translates bus-layer errors into
// HTTP-shaped errs.Error values so the API handler can stay free of any
// business import.
func (a *App) BuildPickSheet(ctx context.Context, req PickSheetRequest) ([]byte, error) {
	pdf, err := a.bus.BuildPickSheet(ctx, paperworkbus.PickSheetRequest{
		OrderID: req.OrderID,
		Zone:    req.Zone,
	})
	if err != nil {
		return nil, mapBusErr("buildpicksheet", err)
	}
	return pdf, nil
}

// BuildReceiveCover delegates to the bus.
func (a *App) BuildReceiveCover(ctx context.Context, req ReceiveCoverRequest) ([]byte, error) {
	pdf, err := a.bus.BuildReceiveCover(ctx, paperworkbus.ReceiveCoverRequest{
		PurchaseOrderID: req.PurchaseOrderID,
	})
	if err != nil {
		return nil, mapBusErr("buildreceivecover", err)
	}
	return pdf, nil
}

// BuildTransferSheet delegates to the bus.
func (a *App) BuildTransferSheet(ctx context.Context, req TransferSheetRequest) ([]byte, error) {
	pdf, err := a.bus.BuildTransferSheet(ctx, paperworkbus.TransferSheetRequest{
		TransferID: req.TransferID,
	})
	if err != nil {
		return nil, mapBusErr("buildtransfersheet", err)
	}
	return pdf, nil
}

// mapBusErr translates paperworkbus errors to errs.Error values. The four
// sentinels surfaced by paperworkbus map onto two HTTP shapes: NotFound (the
// referenced order/PO/transfer does not exist) and InvalidArgument (the
// transfer order exists but lacks the transfer_number paperwork requires).
// Anything else becomes Internal.
func mapBusErr(op string, err error) error {
	switch {
	case errors.Is(err, paperworkbus.ErrOrderNotFound),
		errors.Is(err, paperworkbus.ErrPONotFound),
		errors.Is(err, paperworkbus.ErrTransferNotFound):
		return errs.New(errs.NotFound, err)
	case errors.Is(err, paperworkbus.ErrTransferNumberMissing):
		return errs.New(errs.InvalidArgument, err)
	}
	return errs.Newf(errs.Internal, "%s: %s", op, err)
}
