package paperworkapp

import (
	"context"
	"errors"
	"strings"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/paperwork/pdf"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/sdk/page"
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
// the pure pdf leaf. F4-enrichment wires 11 buses for full cross-domain joins.
type App struct {
	log                *logger.Logger
	orders             *ordersbus.Business
	customers          *customersbus.Business
	pickTasks          *picktaskbus.Business
	purchaseOrders     *purchaseorderbus.Business
	purchaseLines      *purchaseorderlineitembus.Business
	suppliers          *supplierbus.Business
	supplierProducts   *supplierproductbus.Business
	transferOrders     *transferorderbus.Business
	warehouses         *warehousebus.Business
	inventoryLocations *inventorylocationbus.Business
	products           *productbus.Business
}

// NewApp constructs the paperwork app.
func NewApp(
	log *logger.Logger,
	orders *ordersbus.Business,
	customers *customersbus.Business,
	pickTasks *picktaskbus.Business,
	purchaseOrders *purchaseorderbus.Business,
	purchaseLines *purchaseorderlineitembus.Business,
	suppliers *supplierbus.Business,
	supplierProducts *supplierproductbus.Business,
	transferOrders *transferorderbus.Business,
	warehouses *warehousebus.Business,
	inventoryLocations *inventorylocationbus.Business,
	products *productbus.Business,
) *App {
	return &App{
		log:                log,
		orders:             orders,
		customers:          customers,
		pickTasks:          pickTasks,
		purchaseOrders:     purchaseOrders,
		purchaseLines:      purchaseLines,
		suppliers:          suppliers,
		supplierProducts:   supplierProducts,
		transferOrders:     transferOrders,
		warehouses:         warehouses,
		inventoryLocations: inventoryLocations,
		products:           products,
	}
}

// BuildPickSheet renders a pick sheet PDF for the given sales order.
func (a *App) BuildPickSheet(ctx context.Context, req PickSheetRequest) ([]byte, error) {
	order, err := a.orders.QueryByID(ctx, req.OrderID)
	if err != nil {
		if errors.Is(err, ordersbus.ErrNotFound) {
			return nil, errs.New(errs.NotFound, ErrOrderNotFound)
		}
		return nil, errs.Newf(errs.Internal, "buildpicksheet: order: %s", err)
	}

	customer, err := a.customers.QueryByID(ctx, order.CustomerID)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "buildpicksheet: customer: %s", err)
	}

	tasks, err := a.pickTasks.Query(ctx, picktaskbus.QueryFilter{SalesOrderID: &order.ID}, picktaskbus.DefaultOrderBy, page.MustParse("1", "500"))
	if err != nil {
		return nil, errs.Newf(errs.Internal, "buildpicksheet: picktasks: %s", err)
	}

	lines := make([]pdf.PickSheetLine, 0, len(tasks))
	for _, tk := range tasks {
		prod, err := a.products.QueryByID(ctx, tk.ProductID)
		if err != nil {
			return nil, errs.Newf(errs.Internal, "buildpicksheet: product: %s", err)
		}
		loc, err := a.inventoryLocations.QueryByID(ctx, tk.LocationID)
		if err != nil {
			return nil, errs.Newf(errs.Internal, "buildpicksheet: location: %s", err)
		}
		lines = append(lines, pdf.PickSheetLine{
			LocationCode: derefStr(loc.LocationCode),
			SKU:          prod.SKU,
			ProductName:  prod.Name,
			Quantity:     tk.QuantityToPick,
		})
	}

	data := pdf.PickSheetData{
		TaskCode:     taskCodeFor("SO", order.Number),
		OrderNumber:  order.Number,
		CustomerName: customer.Name,
		Zone:         req.Zone,
		Lines:        lines,
	}
	out, err := pdf.PickSheet(data)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "buildpicksheet: render: %s", err)
	}
	return out, nil
}

// BuildReceiveCover renders a receive-cover PDF for the given purchase order.
func (a *App) BuildReceiveCover(ctx context.Context, req ReceiveCoverRequest) ([]byte, error) {
	po, err := a.purchaseOrders.QueryByID(ctx, req.PurchaseOrderID)
	if err != nil {
		if errors.Is(err, purchaseorderbus.ErrNotFound) {
			return nil, errs.New(errs.NotFound, ErrPONotFound)
		}
		return nil, errs.Newf(errs.Internal, "buildreceivecover: po: %s", err)
	}

	supplier, err := a.suppliers.QueryByID(ctx, po.SupplierID)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "buildreceivecover: supplier: %s", err)
	}

	lineItems, err := a.purchaseLines.QueryByPurchaseOrderID(ctx, po.ID)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "buildreceivecover: purchase lines: %s", err)
	}

	lines := make([]pdf.ReceiveCoverLine, 0, len(lineItems))
	for _, li := range lineItems {
		sp, err := a.supplierProducts.QueryByID(ctx, li.SupplierProductID)
		if err != nil {
			return nil, errs.Newf(errs.Internal, "buildreceivecover: supplier product: %s", err)
		}
		prod, err := a.products.QueryByID(ctx, sp.ProductID)
		if err != nil {
			return nil, errs.Newf(errs.Internal, "buildreceivecover: product: %s", err)
		}
		lines = append(lines, pdf.ReceiveCoverLine{
			SKU:         prod.SKU,
			ProductName: prod.Name,
			UPC:         prod.UpcCode,
			Expected:    li.QuantityOrdered,
		})
	}

	data := pdf.ReceiveCoverData{
		TaskCode:   taskCodeFor("PO", po.OrderNumber),
		PONumber:   po.OrderNumber,
		VendorName: supplier.Name,
		ExpectedAt: po.ExpectedDeliveryDate.Format("2006-01-02"),
		Lines:      lines,
	}
	out, err := pdf.ReceiveCover(data)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "buildreceivecover: render: %s", err)
	}
	return out, nil
}

// BuildTransferSheet renders a transfer sheet PDF for the given transfer order.
func (a *App) BuildTransferSheet(ctx context.Context, req TransferSheetRequest) ([]byte, error) {
	to, err := a.transferOrders.QueryByID(ctx, req.TransferID)
	if err != nil {
		if errors.Is(err, transferorderbus.ErrNotFound) {
			return nil, errs.New(errs.NotFound, ErrTransferNotFound)
		}
		return nil, errs.Newf(errs.Internal, "buildtransfersheet: transfer: %s", err)
	}

	if to.TransferNumber == nil || *to.TransferNumber == "" {
		return nil, errs.New(errs.InvalidArgument, ErrTransferNumberMissing)
	}
	transferNum := *to.TransferNumber

	fromLoc, err := a.inventoryLocations.QueryByID(ctx, to.FromLocationID)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "buildtransfersheet: from location: %s", err)
	}

	toLoc, err := a.inventoryLocations.QueryByID(ctx, to.ToLocationID)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "buildtransfersheet: to location: %s", err)
	}

	fromWH, err := a.warehouses.QueryByID(ctx, fromLoc.WarehouseID)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "buildtransfersheet: from warehouse: %s", err)
	}

	toWH, err := a.warehouses.QueryByID(ctx, toLoc.WarehouseID)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "buildtransfersheet: to warehouse: %s", err)
	}

	prod, err := a.products.QueryByID(ctx, to.ProductID)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "buildtransfersheet: product: %s", err)
	}

	lines := []pdf.TransferSheetLine{
		{
			SKU:          prod.SKU,
			ProductName:  prod.Name,
			FromLocation: derefStr(fromLoc.LocationCode),
			ToLocation:   derefStr(toLoc.LocationCode),
			Quantity:     to.Quantity,
		},
	}

	data := pdf.TransferSheetData{
		TaskCode:      taskCodeFor("XFER", transferNum),
		TransferNum:   transferNum,
		SourceWH:      fromWH.Name,
		DestinationWH: toWH.Name,
		Lines:         lines,
	}
	out, err := pdf.TransferSheet(data)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "buildtransfersheet: render: %s", err)
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

// derefStr returns the pointed-to string or "" if nil.
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
