package paperworkapp

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/paperwork/pdf"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Sentinel errors surfaced for HTTP-shape mapping. Relocated from
// paperworkbus in Phase 0g.F4-enrichment when cross-domain orchestration
// moved to this layer (reverses D-CONV-3; see docs/arch/paperwork.md).
var (
	ErrOrderNotFound         = errors.New("paperwork: order not found")
	ErrPONotFound            = errors.New("paperwork: purchase order not found")
	ErrTransferNotFound      = errors.New("paperwork: transfer order not found")
	ErrTransferNumberMissing = errors.New("paperwork: transfer order has no transfer_number")
)

// App orchestrates cross-domain reads for paperwork rendering. Per the
// app-layer orchestration pattern (directedworkapp/supervisorkpiapp/scanapp),
// it holds the sibling buses, assembles pdf.*Data, and delegates rendering to
// the pure pdf leaf. PR1 carries the 5 buses the bus held; PR2 adds the
// enrichment buses.
type App struct {
	log            *logger.Logger
	ordersBus      *ordersbus.Business
	orderLinesBus  *orderlineitemsbus.Business
	purchaseOrders *purchaseorderbus.Business
	purchaseLines  *purchaseorderlineitembus.Business
	transferOrders *transferorderbus.Business
}

// NewApp constructs the paperwork app.
func NewApp(
	log *logger.Logger,
	ordersBus *ordersbus.Business,
	orderLinesBus *orderlineitemsbus.Business,
	purchaseOrders *purchaseorderbus.Business,
	purchaseLines *purchaseorderlineitembus.Business,
	transferOrders *transferorderbus.Business,
) *App {
	return &App{
		log:            log,
		ordersBus:      ordersBus,
		orderLinesBus:  orderLinesBus,
		purchaseOrders: purchaseOrders,
		purchaseLines:  purchaseLines,
		transferOrders: transferOrders,
	}
}

// BuildPickSheet renders a pick sheet PDF for the given sales order.
func (a *App) BuildPickSheet(ctx context.Context, req PickSheetRequest) ([]byte, error) {
	order, err := a.ordersBus.QueryByID(ctx, req.OrderID)
	if err != nil {
		return nil, mapErr("buildpicksheet", fmt.Errorf("%w: %s", ErrOrderNotFound, err))
	}

	data := pdf.PickSheetData{
		TaskCode:     taskCodeFor("SO", order.Number),
		OrderNumber:  order.Number,
		CustomerName: "", // TODO(F4): resolve via customersBus (PR2)
		Zone:         req.Zone,
		Lines:        nil, // TODO(F4): populate from pickTaskBus (PR2)
	}
	out, err := pdf.PickSheet(data)
	if err != nil {
		return nil, mapErr("buildpicksheet", err)
	}
	return out, nil
}

// BuildReceiveCover renders a receive-cover PDF for the given purchase order.
func (a *App) BuildReceiveCover(ctx context.Context, req ReceiveCoverRequest) ([]byte, error) {
	po, err := a.purchaseOrders.QueryByID(ctx, req.PurchaseOrderID)
	if err != nil {
		return nil, mapErr("buildreceivecover", fmt.Errorf("%w: %s", ErrPONotFound, err))
	}

	data := pdf.ReceiveCoverData{
		TaskCode:   taskCodeFor("PO", po.OrderNumber),
		PONumber:   po.OrderNumber,
		VendorName: "", // TODO(F4): resolve via supplierBus (PR2)
		ExpectedAt: "", // TODO(F4): format po.ExpectedDeliveryDate (PR2)
		Lines:      nil, // TODO(F4): populate from purchaseLines (PR2)
	}
	out, err := pdf.ReceiveCover(data)
	if err != nil {
		return nil, mapErr("buildreceivecover", err)
	}
	return out, nil
}

// BuildTransferSheet renders a transfer sheet PDF for the given transfer order.
func (a *App) BuildTransferSheet(ctx context.Context, req TransferSheetRequest) ([]byte, error) {
	to, err := a.transferOrders.QueryByID(ctx, req.TransferID)
	if err != nil {
		return nil, mapErr("buildtransfersheet", fmt.Errorf("%w: %s", ErrTransferNotFound, err))
	}

	if to.TransferNumber == nil || *to.TransferNumber == "" {
		return nil, mapErr("buildtransfersheet", ErrTransferNumberMissing)
	}
	transferNum := *to.TransferNumber

	data := pdf.TransferSheetData{
		TaskCode:      taskCodeFor("XFER", transferNum),
		TransferNum:   transferNum,
		SourceWH:      "", // TODO(F4): resolve via warehouseBus (PR2)
		DestinationWH: "", // TODO(F4): resolve via warehouseBus (PR2)
		Lines:         nil, // TODO(F4): populate transfer line (PR2)
	}
	out, err := pdf.TransferSheet(data)
	if err != nil {
		return nil, mapErr("buildtransfersheet", err)
	}
	return out, nil
}

// taskCodeFor produces a deterministic prefix-qualified task code. Idempotent:
// it strips an existing "<prefix>-" head before re-prepending, so callers may
// pass bare or already-prefixed values. Relocated verbatim from paperworkbus.
//
//	taskCodeFor("SO", "12345")    → "SO-12345"
//	taskCodeFor("SO", "SO-12345") → "SO-12345"
func taskCodeFor(prefix, value string) string {
	return prefix + "-" + strings.TrimPrefix(value, prefix+"-")
}

// mapErr translates orchestration errors to errs.Error values. The not-found
// sentinels map to 404; transfer-number-missing to 400; everything else
// (including secondary-join failures, which indicate dangling FKs) to 500.
func mapErr(op string, err error) error {
	switch {
	case errors.Is(err, ErrOrderNotFound),
		errors.Is(err, ErrPONotFound),
		errors.Is(err, ErrTransferNotFound):
		return errs.New(errs.NotFound, err)
	case errors.Is(err, ErrTransferNumberMissing):
		return errs.New(errs.InvalidArgument, err)
	}
	return errs.Newf(errs.Internal, "%s: %s", op, err)
}
