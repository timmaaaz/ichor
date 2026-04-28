package paperworkbus

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/paperwork/paperworkbus/pdf"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Sentinel errors surfaced to the app layer for HTTP-shape mapping.
//
// ErrOrderNotFound, ErrPONotFound, ErrTransferNotFound map to errs.NotFound
// (HTTP 404). ErrTransferNumberMissing maps to errs.InvalidArgument (HTTP 400) —
// a transfer order without a transfer_number is malformed for paperwork
// rendering even though the row is otherwise valid.
var (
	ErrOrderNotFound         = errors.New("paperwork: order not found")
	ErrPONotFound            = errors.New("paperwork: purchase order not found")
	ErrTransferNotFound      = errors.New("paperwork: transfer order not found")
	ErrTransferNumberMissing = errors.New("paperwork: transfer order has no transfer_number")
)

// Business manages paperwork rendering. Per D-CONV-3 (renderer-shaped), there
// is no Storer interface and no DB writes — Business holds 5 sibling-bus
// pointers + log and delegates rendering to the pdf subpackage.
type Business struct {
	log            *logger.Logger
	ordersBus      *ordersbus.Business
	orderLinesBus  *orderlineitemsbus.Business
	purchaseOrders *purchaseorderbus.Business
	purchaseLines  *purchaseorderlineitembus.Business
	transferOrders *transferorderbus.Business
}

// NewBusiness constructs a paperwork business API for use. The 5 sibling-bus
// dependencies are required: pick sheets, receive covers, and transfer sheets
// each fan out into one or more sibling reads.
func NewBusiness(
	log *logger.Logger,
	ordersBus *ordersbus.Business,
	orderLinesBus *orderlineitemsbus.Business,
	purchaseOrders *purchaseorderbus.Business,
	purchaseLines *purchaseorderlineitembus.Business,
	transferOrders *transferorderbus.Business,
) *Business {
	return &Business{
		log:            log,
		ordersBus:      ordersBus,
		orderLinesBus:  orderLinesBus,
		purchaseOrders: purchaseOrders,
		purchaseLines:  purchaseLines,
		transferOrders: transferOrders,
	}
}

// BuildPickSheet renders a pick sheet PDF for the given sales order.
//
// B3 scope: emit a non-empty PDF whose taskCode embeds the SO-prefixed order
// number. Lines are intentionally left empty here — F4 (frontend enrichment)
// will populate line item rows once the upstream allocation pipeline lands.
// Customer name is left blank for the same reason (FK-only on Order; resolution
// requires a join not yet plumbed through the bus).
func (b *Business) BuildPickSheet(ctx context.Context, req PickSheetRequest) ([]byte, error) {
	order, err := b.ordersBus.QueryByID(ctx, req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrOrderNotFound, err)
	}

	data := pdf.PickSheetData{
		TaskCode:     taskCodeFor("SO", order.Number),
		OrderNumber:  order.Number,
		CustomerName: "", // TODO(F4): resolve via customers join
		Zone:         req.Zone,
		Lines:        nil, // TODO(F4): populate from b.orderLinesBus
	}
	return pdf.PickSheet(data)
}

// BuildReceiveCover renders a receive-cover PDF for the given purchase order.
// See BuildPickSheet for the F4 enrichment story applied to vendor name + lines.
func (b *Business) BuildReceiveCover(ctx context.Context, req ReceiveCoverRequest) ([]byte, error) {
	po, err := b.purchaseOrders.QueryByID(ctx, req.PurchaseOrderID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrPONotFound, err)
	}

	data := pdf.ReceiveCoverData{
		TaskCode:   taskCodeFor("PO", po.OrderNumber),
		PONumber:   po.OrderNumber,
		VendorName: "", // TODO(F4): resolve via suppliers join
		ExpectedAt: "", // TODO(F4): format po.ExpectedDeliveryDate
		Lines:      nil, // TODO(F4): populate from b.purchaseLines
	}
	return pdf.ReceiveCover(data)
}

// BuildTransferSheet renders a transfer sheet PDF for the given transfer order.
//
// TransferOrder.TransferNumber is *string (nullable in the schema). Paperwork
// requires it — the task code IS the transfer number with an "XFER-" prefix
// guard — so a nil/empty TransferNumber is rejected via ErrTransferNumberMissing.
func (b *Business) BuildTransferSheet(ctx context.Context, req TransferSheetRequest) ([]byte, error) {
	to, err := b.transferOrders.QueryByID(ctx, req.TransferID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTransferNotFound, err)
	}

	if to.TransferNumber == nil || *to.TransferNumber == "" {
		return nil, ErrTransferNumberMissing
	}
	transferNum := *to.TransferNumber

	data := pdf.TransferSheetData{
		TaskCode:      taskCodeFor("XFER", transferNum),
		TransferNum:   transferNum,
		SourceWH:      "", // TODO(F4): resolve via warehouses join
		DestinationWH: "", // TODO(F4): resolve via warehouses join
		Lines:         nil, // TODO(F4): populate transfer lines
	}
	return pdf.TransferSheet(data)
}

// taskCodeFor produces a deterministic prefix-qualified task code from a
// human-readable identifier. The function is idempotent: it strips an existing
// "<prefix>-" head before re-prepending, so callers may pass either bare
// numbers or already-prefixed values without producing double-prefixed output.
//
//	taskCodeFor("SO", "12345")    → "SO-12345"
//	taskCodeFor("SO", "SO-12345") → "SO-12345"   (no double-prefix)
//	taskCodeFor("PO", "PO-1")     → "PO-1"
//
// Note: only the EXACT "<prefix>-" head is stripped. taskCodeFor("SO", "ORD-9")
// produces "SO-ORD-9" because "ORD-" is not the configured prefix; this is
// intentional and is tracked as a B4 follow-up where seed-side number formats
// are normalized.
func taskCodeFor(prefix, value string) string {
	return prefix + "-" + strings.TrimPrefix(value, prefix+"-")
}
