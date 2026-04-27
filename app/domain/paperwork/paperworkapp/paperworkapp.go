package paperworkapp

import (
	"context"
	"errors"

	"github.com/timmaaaz/ichor/app/sdk/errs"
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

// mapBusErr translates paperworkbus errors to errs.Error values. Phase 0g.B2
// only sees ErrNotImplemented; B3 expands this with NotFound / InvalidArgument
// mappings as real handler bodies land.
func mapBusErr(op string, err error) error {
	if errors.Is(err, paperworkbus.ErrNotImplemented) {
		return errs.New(errs.Unimplemented, paperworkbus.ErrNotImplemented)
	}
	return errs.Newf(errs.Internal, "%s: %s", op, err)
}
