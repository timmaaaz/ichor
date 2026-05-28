package paperworkapp

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
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

// Sentinel errors surfaced for HTTP-shape mapping. Relocated from the deleted
// paperworkbus in Phase 0g.F4-enrichment, when cross-domain orchestration moved
// to the app layer to match directedworkapp/supervisorkpiapp/scanapp.
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
		return nil, a.internal(ctx, "buildpicksheet: order", err)
	}

	customer, err := a.customers.QueryByID(ctx, order.CustomerID)
	if err != nil {
		return nil, a.internal(ctx, "buildpicksheet: customer", err)
	}

	// One sales order never approaches 1000 distinct pick tasks; the high
	// ceiling matches pickingapp's order-scoped "fetch all" queries and keeps
	// the sheet from silently dropping lines (a missed line = a missed pick).
	tasks, err := a.pickTasks.Query(ctx, picktaskbus.QueryFilter{SalesOrderID: &order.ID}, picktaskbus.DefaultOrderBy, page.MustParse("1", "1000"))
	if err != nil {
		return nil, a.internal(ctx, "buildpicksheet: picktasks", err)
	}

	// A pick sheet lists work still to be done. Skip terminal tasks so a
	// re-printed sheet never shows already-picked, short-picked, or voided
	// lines as if they still need picking, and collect the IDs to resolve.
	active := make([]picktaskbus.PickTask, 0, len(tasks))
	productIDs := make([]uuid.UUID, 0, len(tasks))
	locationIDs := make([]uuid.UUID, 0, len(tasks))
	for _, tk := range tasks {
		switch tk.Status {
		case picktaskbus.Statuses.Completed, picktaskbus.Statuses.ShortPicked, picktaskbus.Statuses.Cancelled:
			continue
		}
		active = append(active, tk)
		productIDs = append(productIDs, tk.ProductID)
		locationIDs = append(locationIDs, tk.LocationID)
	}

	// Resolve every product and location in one batched query each, rather than
	// two per task (avoids N+1).
	productByID, err := a.productsByID(ctx, "buildpicksheet", productIDs)
	if err != nil {
		return nil, err
	}
	locationByID, err := a.locationsByID(ctx, "buildpicksheet", locationIDs)
	if err != nil {
		return nil, err
	}

	lines := make([]pdf.PickSheetLine, 0, len(active))
	for _, tk := range active {
		prod, ok := productByID[tk.ProductID]
		if !ok {
			return nil, a.internal(ctx, "buildpicksheet: product", fmt.Errorf("product %s not found for pick task %s", tk.ProductID, tk.ID))
		}
		loc, ok := locationByID[tk.LocationID]
		if !ok {
			return nil, a.internal(ctx, "buildpicksheet: location", fmt.Errorf("location %s not found for pick task %s", tk.LocationID, tk.ID))
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
		return nil, a.internal(ctx, "buildpicksheet: render", err)
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
		return nil, a.internal(ctx, "buildreceivecover: po", err)
	}

	supplier, err := a.suppliers.QueryByID(ctx, po.SupplierID)
	if err != nil {
		return nil, a.internal(ctx, "buildreceivecover: supplier", err)
	}

	lineItems, err := a.purchaseLines.QueryByPurchaseOrderID(ctx, po.ID)
	if err != nil {
		return nil, a.internal(ctx, "buildreceivecover: purchase lines", err)
	}

	// Resolve every supplier product, then every underlying product, in one
	// batched query each rather than two per line (avoids N+1).
	spIDs := make([]uuid.UUID, len(lineItems))
	for i, li := range lineItems {
		spIDs[i] = li.SupplierProductID
	}
	supplierProductByID, err := a.supplierProductsByID(ctx, "buildreceivecover", spIDs)
	if err != nil {
		return nil, err
	}
	productIDs := make([]uuid.UUID, 0, len(supplierProductByID))
	for _, sp := range supplierProductByID {
		productIDs = append(productIDs, sp.ProductID)
	}
	productByID, err := a.productsByID(ctx, "buildreceivecover", productIDs)
	if err != nil {
		return nil, err
	}

	lines := make([]pdf.ReceiveCoverLine, 0, len(lineItems))
	for _, li := range lineItems {
		sp, ok := supplierProductByID[li.SupplierProductID]
		if !ok {
			return nil, a.internal(ctx, "buildreceivecover: supplier product", fmt.Errorf("supplier product %s referenced by a po line not found", li.SupplierProductID))
		}
		prod, ok := productByID[sp.ProductID]
		if !ok {
			return nil, a.internal(ctx, "buildreceivecover: product", fmt.Errorf("product %s not found for supplier product %s", sp.ProductID, sp.SupplierProductID))
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
		return nil, a.internal(ctx, "buildreceivecover: render", err)
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
		return nil, a.internal(ctx, "buildtransfersheet: transfer", err)
	}

	if to.TransferNumber == nil || *to.TransferNumber == "" {
		return nil, errs.New(errs.InvalidArgument, ErrTransferNumberMissing)
	}
	transferNum := *to.TransferNumber

	fromLoc, err := a.inventoryLocations.QueryByID(ctx, to.FromLocationID)
	if err != nil {
		return nil, a.internal(ctx, "buildtransfersheet: from location", err)
	}

	toLoc, err := a.inventoryLocations.QueryByID(ctx, to.ToLocationID)
	if err != nil {
		return nil, a.internal(ctx, "buildtransfersheet: to location", err)
	}

	fromWH, err := a.warehouses.QueryByID(ctx, fromLoc.WarehouseID)
	if err != nil {
		return nil, a.internal(ctx, "buildtransfersheet: from warehouse", err)
	}

	toWH, err := a.warehouses.QueryByID(ctx, toLoc.WarehouseID)
	if err != nil {
		return nil, a.internal(ctx, "buildtransfersheet: to warehouse", err)
	}

	prod, err := a.products.QueryByID(ctx, to.ProductID)
	if err != nil {
		return nil, a.internal(ctx, "buildtransfersheet: product", err)
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
		return nil, a.internal(ctx, "buildtransfersheet: render", err)
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

// internal logs a server-side paperwork failure (a primary-query DB error, a
// secondary cross-domain-join failure, or a render fault) and returns an opaque
// errs.Internal. The detail is recorded for debugging but never surfaced to the
// client — mirroring directedworkapp's observable-anomaly logging. Secondary
// joins route here on purpose: a broken FK (e.g. an order referencing a missing
// customer) is a server-side data-integrity problem, not a client 404.
func (a *App) internal(ctx context.Context, op string, err error) error {
	a.log.Error(ctx, "paperwork: "+op, "error", err)
	return errs.Newf(errs.Internal, "%s: %s", op, err)
}

// productsByID resolves products in a single batched query, keyed by product
// ID. ids may contain duplicates. A query failure is logged and surfaced as
// errs.Internal (op identifies the calling sheet).
func (a *App) productsByID(ctx context.Context, op string, ids []uuid.UUID) (map[uuid.UUID]productbus.Product, error) {
	prods, err := a.products.QueryByIDs(ctx, dedupe(ids))
	if err != nil {
		return nil, a.internal(ctx, op+": products", err)
	}
	m := make(map[uuid.UUID]productbus.Product, len(prods))
	for _, p := range prods {
		m[p.ProductID] = p
	}
	return m, nil
}

// locationsByID resolves inventory locations in a single batched query, keyed
// by location ID. ids may contain duplicates.
func (a *App) locationsByID(ctx context.Context, op string, ids []uuid.UUID) (map[uuid.UUID]inventorylocationbus.InventoryLocation, error) {
	locs, err := a.inventoryLocations.QueryByIDs(ctx, dedupe(ids))
	if err != nil {
		return nil, a.internal(ctx, op+": locations", err)
	}
	m := make(map[uuid.UUID]inventorylocationbus.InventoryLocation, len(locs))
	for _, l := range locs {
		m[l.LocationID] = l
	}
	return m, nil
}

// supplierProductsByID resolves supplier products in a single batched query,
// keyed by supplier-product ID. ids may contain duplicates.
func (a *App) supplierProductsByID(ctx context.Context, op string, ids []uuid.UUID) (map[uuid.UUID]supplierproductbus.SupplierProduct, error) {
	sps, err := a.supplierProducts.QueryByIDs(ctx, dedupe(ids))
	if err != nil {
		return nil, a.internal(ctx, op+": supplier products", err)
	}
	m := make(map[uuid.UUID]supplierproductbus.SupplierProduct, len(sps))
	for _, sp := range sps {
		m[sp.SupplierProductID] = sp
	}
	return m, nil
}

// dedupe returns the unique IDs in ids. Order is not preserved; the result
// feeds an IN-clause and a lookup map, neither of which is order-sensitive.
func dedupe(ids []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(ids))
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// derefStr returns the pointed-to string or "" if nil.
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
