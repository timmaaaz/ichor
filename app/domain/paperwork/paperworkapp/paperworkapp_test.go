// Package paperworkapp_test verifies the paperwork app renders non-empty PDFs
// containing the expected SO-/PO-/XFER- task codes when the requested order
// exists, and surfaces the right sentinel errors when it does not.
//
// The acceptance bar is intentionally minimal: each Build* method must produce
// bytes that begin with the "%PDF-" magic and contain the expected task code
// somewhere in the document. Visual layout is covered by the pdf subpackage
// tests (picksheet_test.go etc.); cross-domain wiring is covered here.
package paperworkapp_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/paperwork/paperworkapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitemstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
	"github.com/timmaaaz/ichor/business/domain/sales/lineitemfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// pdfMagic is the 5-byte header every well-formed PDF begins with.
var pdfMagic = []byte("%PDF-")

// paperworkSeed bundles the cross-domain fixtures the paperwork integration
// test needs. All entities are seeded once per test database (the currency
// seed helper uses fixed codes "TS0..TSn" and cannot be called twice without
// a unique-violation), so sub-tests share the same fixture set.
type paperworkSeed struct {
	order         ordersbus.Order
	purchaseOrder purchaseorderbus.PurchaseOrder
	transferOrder transferorderbus.TransferOrder

	// transferOrderNoNum is a second, dedicated transfer order reserved for the
	// TransferNumberMissing sub-test, which clears its TransferNumber. Keeping
	// it separate from transferOrder means no sub-test mutates a fixture another
	// reads, so the transfer sub-tests are order-independent.
	transferOrderNoNum transferorderbus.TransferOrder

	// enrichment fields (F4)
	customerName  string
	pickTaskCount int
	vendorName    string
	srcWHName     string
	dstWHName     string
}

// newTestApp constructs a paperworkapp.App wired to the test database's full
// BusDomain — the 11-bus signature introduced in F4-enrichment.
func newTestApp(db *dbtest.Database) *paperworkapp.App {
	return paperworkapp.NewApp(
		db.Log,
		db.BusDomain.Order,
		db.BusDomain.Customers,
		db.BusDomain.PickTask,
		db.BusDomain.PurchaseOrder,
		db.BusDomain.PurchaseOrderLineItem,
		db.BusDomain.Supplier,
		db.BusDomain.SupplierProduct,
		db.BusDomain.TransferOrder,
		db.BusDomain.Warehouse,
		db.BusDomain.InventoryLocation,
		db.BusDomain.Product,
	)
}

func TestPaperworkApp_Integration(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "TestPaperworkApp_Integration")
	app := newTestApp(db)

	ctx := context.Background()
	seed := seedAll(t, ctx, db.BusDomain)

	// --- Sheet A: Pick -------------------------------------------------------

	t.Run("BuildPickSheet", func(t *testing.T) {
		got, err := app.BuildPickSheet(ctx, paperworkapp.PickSheetRequest{OrderID: seed.order.ID})
		if err != nil {
			t.Fatalf("BuildPickSheet: unexpected error: %v", err)
		}
		if !bytes.HasPrefix(got, pdfMagic) {
			t.Fatalf("BuildPickSheet: output does not start with %%PDF- (first 8 bytes: %q)", firstN(got, 8))
		}
		expected := expectedTaskCode("SO", seed.order.Number)
		if !bytes.Contains(got, []byte(expected)) {
			t.Errorf("BuildPickSheet: PDF does not contain task code %q", expected)
		}
	})

	t.Run("BuildPickSheet_Enriched", func(t *testing.T) {
		got, err := app.BuildPickSheet(ctx, paperworkapp.PickSheetRequest{OrderID: seed.order.ID})
		if err != nil {
			t.Fatalf("BuildPickSheet_Enriched: unexpected error: %v", err)
		}
		if !bytes.Contains(got, []byte(seed.customerName)) {
			t.Errorf("BuildPickSheet_Enriched: PDF does not contain customer name %q", seed.customerName)
		}
		if seed.pickTaskCount == 0 {
			t.Errorf("BuildPickSheet_Enriched: no pick tasks were seeded — enrichment test is vacuous")
		}
	})

	t.Run("BuildPickSheet_NotFound", func(t *testing.T) {
		_, err := app.BuildPickSheet(ctx, paperworkapp.PickSheetRequest{OrderID: uuid.New()})
		var appErr *errs.Error
		if !errors.As(err, &appErr) || !appErr.Code.Equal(errs.NotFound) {
			t.Fatalf("BuildPickSheet_NotFound: want NotFound errs.Error for missing order, got %v", err)
		}
	})

	// --- Sheet B: Receive ----------------------------------------------------

	t.Run("BuildReceiveCover", func(t *testing.T) {
		got, err := app.BuildReceiveCover(ctx, paperworkapp.ReceiveCoverRequest{PurchaseOrderID: seed.purchaseOrder.ID})
		if err != nil {
			t.Fatalf("BuildReceiveCover: unexpected error: %v", err)
		}
		if !bytes.HasPrefix(got, pdfMagic) {
			t.Fatalf("BuildReceiveCover: output does not start with %%PDF- (first 8 bytes: %q)", firstN(got, 8))
		}
		expected := expectedTaskCode("PO", seed.purchaseOrder.OrderNumber)
		if !bytes.Contains(got, []byte(expected)) {
			t.Errorf("BuildReceiveCover: PDF does not contain task code %q", expected)
		}
	})

	t.Run("BuildReceiveCover_Enriched", func(t *testing.T) {
		got, err := app.BuildReceiveCover(ctx, paperworkapp.ReceiveCoverRequest{PurchaseOrderID: seed.purchaseOrder.ID})
		if err != nil {
			t.Fatalf("BuildReceiveCover_Enriched: unexpected error: %v", err)
		}
		if !bytes.Contains(got, []byte(seed.vendorName)) {
			t.Errorf("BuildReceiveCover_Enriched: PDF does not contain vendor name %q", seed.vendorName)
		}
		expectedDate := seed.purchaseOrder.ExpectedDeliveryDate.Format("2006-01-02")
		if !bytes.Contains(got, []byte(expectedDate)) {
			t.Errorf("BuildReceiveCover_Enriched: PDF does not contain expected date %q", expectedDate)
		}
	})

	t.Run("BuildReceiveCover_NotFound", func(t *testing.T) {
		_, err := app.BuildReceiveCover(ctx, paperworkapp.ReceiveCoverRequest{PurchaseOrderID: uuid.New()})
		var appErr *errs.Error
		if !errors.As(err, &appErr) || !appErr.Code.Equal(errs.NotFound) {
			t.Fatalf("BuildReceiveCover_NotFound: want NotFound errs.Error for missing PO, got %v", err)
		}
	})

	// --- Sheet C: Transfer ---------------------------------------------------

	t.Run("BuildTransferSheet", func(t *testing.T) {
		if seed.transferOrder.TransferNumber == nil {
			t.Fatalf("seedAll returned transfer order with nil TransferNumber")
		}

		got, err := app.BuildTransferSheet(ctx, paperworkapp.TransferSheetRequest{TransferID: seed.transferOrder.TransferID})
		if err != nil {
			t.Fatalf("BuildTransferSheet: unexpected error: %v", err)
		}
		if !bytes.HasPrefix(got, pdfMagic) {
			t.Fatalf("BuildTransferSheet: output does not start with %%PDF- (first 8 bytes: %q)", firstN(got, 8))
		}
		// TransferNumber already begins with "XFER-" (per testutil.go), and
		// taskCodeFor is idempotent, so the embedded code matches the raw
		// transfer number.
		expected := []byte(*seed.transferOrder.TransferNumber)
		if !bytes.Contains(got, expected) {
			t.Errorf("BuildTransferSheet: PDF does not contain transfer number %q", expected)
		}
	})

	t.Run("BuildTransferSheet_Enriched", func(t *testing.T) {
		if seed.transferOrder.TransferNumber == nil {
			t.Skip("transfer order has no transfer number — skipping enrichment test")
		}
		if seed.srcWHName == seed.dstWHName {
			t.Fatalf("test is vacuous: srcWHName == dstWHName == %q; from/to locations must be in different warehouses", seed.srcWHName)
		}
		got, err := app.BuildTransferSheet(ctx, paperworkapp.TransferSheetRequest{TransferID: seed.transferOrder.TransferID})
		if err != nil {
			t.Fatalf("BuildTransferSheet_Enriched: unexpected error: %v", err)
		}
		if !bytes.Contains(got, []byte(seed.srcWHName)) {
			t.Errorf("BuildTransferSheet_Enriched: PDF does not contain source WH name %q", seed.srcWHName)
		}
		if !bytes.Contains(got, []byte(seed.dstWHName)) {
			t.Errorf("BuildTransferSheet_Enriched: PDF does not contain destination WH name %q", seed.dstWHName)
		}
	})

	t.Run("BuildTransferSheet_NotFound", func(t *testing.T) {
		_, err := app.BuildTransferSheet(ctx, paperworkapp.TransferSheetRequest{TransferID: uuid.New()})
		var appErr *errs.Error
		if !errors.As(err, &appErr) || !appErr.Code.Equal(errs.NotFound) {
			t.Fatalf("BuildTransferSheet_NotFound: want NotFound errs.Error for missing transfer, got %v", err)
		}
	})

	// BuildTransferSheet_TransferNumberMissing clears the TransferNumber on a
	// dedicated transfer order (seed.transferOrderNoNum), so it no longer
	// depends on running last or on any other transfer sub-test's state.
	t.Run("BuildTransferSheet_TransferNumberMissing", func(t *testing.T) {
		empty := ""
		updated, err := db.BusDomain.TransferOrder.Update(ctx, seed.transferOrderNoNum, transferorderbus.UpdateTransferOrder{TransferNumber: &empty})
		if err != nil {
			t.Fatalf("clear transfer number: %v", err)
		}
		_, err = app.BuildTransferSheet(ctx, paperworkapp.TransferSheetRequest{TransferID: updated.TransferID})
		var appErr *errs.Error
		if !errors.As(err, &appErr) || !appErr.Code.Equal(errs.InvalidArgument) {
			t.Fatalf("want InvalidArgument for empty transfer number, got %v", err)
		}
	})
}

// firstN returns the first n bytes of b for diagnostic printing without
// panicking on short inputs.
func firstN(b []byte, n int) []byte {
	if len(b) < n {
		return b
	}
	return b[:n]
}

// expectedTaskCode mirrors paperworkapp.taskCodeFor — kept as a separate test
// helper because the production helper is unexported. The two MUST stay in
// lockstep; if production grows a knob, update this helper.
func expectedTaskCode(prefix, value string) string {
	return prefix + "-" + strings.TrimPrefix(value, prefix+"-")
}

// seedAll provisions the union of FK chains the three Build* methods need:
// addresses, customers, suppliers, warehouses, zones, locations, products,
// statuses, currency, plus exactly one Order, PurchaseOrder, and
// TransferOrder. Also seeds pick tasks (for sheet A) and PO line items (for
// sheet B). Returns a paperworkSeed with enrichment metadata.
//
// All seeding happens in a single function so currency, status, and other
// fixed-code seeders are invoked exactly once — calling them twice in the
// same test database produces unique-key violations.
func seedAll(t *testing.T, ctx context.Context, busDomain dbtest.BusDomain) paperworkSeed {
	t.Helper()

	// USERS -----------------------------------------------------------------

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	adminID := admins[0].ID

	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 6, userbus.Roles.User, busDomain.User)
	if err != nil {
		t.Fatalf("seed users: %v", err)
	}
	userIDs := make([]uuid.UUID, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	// ADDRESS / CONTACT CHAIN ----------------------------------------------

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		t.Fatalf("query regions: %v", err)
	}
	regionIDs := make(uuid.UUIDs, len(regions))
	for i, r := range regions {
		regionIDs[i] = r.ID
	}

	cities, err := citybus.TestSeedCities(ctx, 5, regionIDs, busDomain.City)
	if err != nil {
		t.Fatalf("seed cities: %v", err)
	}
	cityIDs := make(uuid.UUIDs, len(cities))
	for i, c := range cities {
		cityIDs[i] = c.ID
	}

	streets, err := streetbus.TestSeedStreets(ctx, 5, cityIDs, busDomain.Street)
	if err != nil {
		t.Fatalf("seed streets: %v", err)
	}
	streetIDs := make(uuid.UUIDs, len(streets))
	for i, s := range streets {
		streetIDs[i] = s.ID
	}

	tzs, err := busDomain.Timezone.QueryAll(ctx)
	if err != nil {
		t.Fatalf("query timezones: %v", err)
	}
	tzIDs := make(uuid.UUIDs, len(tzs))
	for i, tz := range tzs {
		tzIDs[i] = tz.ID
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 5, streetIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		t.Fatalf("seed contact info: %v", err)
	}
	contactInfoIDs := make(uuid.UUIDs, len(contactInfos))
	for i, ci := range contactInfos {
		contactInfoIDs[i] = ci.ID
	}

	// SHARED CURRENCY (single call — fixed codes TS0..TS4) -----------------

	currencies, err := currencybus.TestSeedCurrencies(ctx, 5, busDomain.Currency)
	if err != nil {
		t.Fatalf("seed currencies: %v", err)
	}
	currencyIDs := make(uuid.UUIDs, len(currencies))
	for i, c := range currencies {
		currencyIDs[i] = c.ID
	}

	// ORDER (sales) --------------------------------------------------------

	customers, err := customersbus.TestSeedCustomers(ctx, 1, streetIDs, contactInfoIDs, uuid.UUIDs{adminID}, busDomain.Customers)
	if err != nil {
		t.Fatalf("seed customers: %v", err)
	}
	customerIDs := uuid.UUIDs{customers[0].ID}

	ofls, err := orderfulfillmentstatusbus.TestSeedOrderFulfillmentStatuses(ctx, busDomain.OrderFulfillmentStatus)
	if err != nil {
		t.Fatalf("seed order fulfillment statuses: %v", err)
	}
	oflIDs := make(uuid.UUIDs, len(ofls))
	for i, o := range ofls {
		oflIDs[i] = o.ID
	}

	orders, err := ordersbus.TestSeedOrders(ctx, 1, uuid.UUIDs{adminID}, customerIDs, oflIDs, currencyIDs, busDomain.Order)
	if err != nil {
		t.Fatalf("seed orders: %v", err)
	}

	// PURCHASE ORDER -------------------------------------------------------

	suppliers, err := supplierbus.TestSeedSuppliers(ctx, 1, contactInfoIDs, busDomain.Supplier)
	if err != nil {
		t.Fatalf("seed suppliers: %v", err)
	}
	supplierIDs := make(uuid.UUIDs, len(suppliers))
	for i, s := range suppliers {
		supplierIDs[i] = s.SupplierID
	}

	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 2, adminID, streetIDs, busDomain.Warehouse)
	if err != nil {
		t.Fatalf("seed warehouses: %v", err)
	}
	warehouseIDs := make(uuid.UUIDs, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	poStatuses, err := purchaseorderstatusbus.TestSeedPurchaseOrderStatuses(ctx, 3, busDomain.PurchaseOrderStatus)
	if err != nil {
		t.Fatalf("seed PO statuses: %v", err)
	}
	poStatusIDs := make(uuid.UUIDs, len(poStatuses))
	for i, s := range poStatuses {
		poStatusIDs[i] = s.ID
	}

	pos, err := purchaseorderbus.TestSeedPurchaseOrders(ctx, 1, supplierIDs, poStatusIDs, warehouseIDs, streetIDs, uuid.UUIDs{adminID}, currencyIDs, busDomain.PurchaseOrder)
	if err != nil {
		t.Fatalf("seed purchase orders: %v", err)
	}

	// TRANSFER ORDER -------------------------------------------------------

	brands, err := brandbus.TestSeedBrands(ctx, 1, contactInfoIDs, busDomain.Brand)
	if err != nil {
		t.Fatalf("seed brands: %v", err)
	}
	brandIDs := make(uuid.UUIDs, len(brands))
	for i, b := range brands {
		brandIDs[i] = b.BrandID
	}

	categories, err := productcategorybus.TestSeedProductCategories(ctx, 1, busDomain.ProductCategory)
	if err != nil {
		t.Fatalf("seed product categories: %v", err)
	}
	categoryIDs := make(uuid.UUIDs, len(categories))
	for i, c := range categories {
		categoryIDs[i] = c.ProductCategoryID
	}

	products, err := productbus.TestSeedProducts(ctx, 2, brandIDs, categoryIDs, busDomain.Product)
	if err != nil {
		t.Fatalf("seed products: %v", err)
	}
	productIDs := make([]uuid.UUID, len(products))
	for i, p := range products {
		productIDs[i] = p.ProductID
	}

	zones, err := zonebus.TestSeedZone(ctx, 4, warehouseIDs, busDomain.Zones)
	if err != nil {
		t.Fatalf("seed zones: %v", err)
	}

	locations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 4, warehouseIDs, zones, busDomain.InventoryLocation)
	if err != nil {
		t.Fatalf("seed inventory locations: %v", err)
	}
	if len(locations) < 2 {
		t.Fatalf("seed inventory locations: want >=2, got %d", len(locations))
	}

	// Move locations[1] to warehouseIDs[1] so from/to belong to DIFFERENT
	// warehouses — without this, TestNewInventoryLocation assigns warehouseIDs[0]
	// to every location, making BuildTransferSheet_Enriched a vacuous test
	// (srcWHName == dstWHName).
	updatedToLoc, err := busDomain.InventoryLocation.Update(ctx, locations[1], inventorylocationbus.UpdateInventoryLocation{
		WarehouseID: &warehouseIDs[1],
	})
	if err != nil {
		t.Fatalf("update to-location warehouse: %v", err)
	}
	locations[1] = updatedToLoc

	fromIDs := []uuid.UUID{locations[0].LocationID}
	toIDs := []uuid.UUID{locations[1].LocationID}

	// Seed two transfer orders: tos[0] is the shared fixture the happy-path
	// transfer sub-tests render; tos[1] is dedicated to the destructive
	// TransferNumberMissing sub-test (see paperworkSeed.transferOrderNoNum).
	tos, err := transferorderbus.TestSeedTransferOrders(ctx, 2, productIDs, fromIDs, toIDs, userIDs[:3], userIDs[3:], nil, busDomain.TransferOrder)
	if err != nil {
		t.Fatalf("seed transfer orders: %v", err)
	}
	if len(tos) < 2 {
		t.Fatalf("seed transfer orders: want >=2, got %d", len(tos))
	}

	// LINE ITEM FULFILLMENT STATUSES (FK for sales order line items + pick tasks)

	lineItemStatuses, err := lineitemfulfillmentstatusbus.TestSeedLineItemFulfillmentStatuses(ctx, busDomain.LineItemFulfillmentStatus)
	if err != nil {
		t.Fatalf("seed line item fulfillment statuses: %v", err)
	}
	lineItemStatusIDs := make(uuid.UUIDs, len(lineItemStatuses))
	for i, s := range lineItemStatuses {
		lineItemStatusIDs[i] = s.ID
	}

	// ORDER LINE ITEMS (FK parent for pick tasks) --------------------------

	orderLineItems, err := orderlineitemsbus.TestSeedOrderLineItems(
		ctx, 2,
		uuid.UUIDs{orders[0].ID},
		uuid.UUIDs(productIDs),
		lineItemStatusIDs,
		uuid.UUIDs(userIDs),
		busDomain.OrderLineItem,
	)
	if err != nil {
		t.Fatalf("seed order line items: %v", err)
	}
	orderLineItemIDs := make([]uuid.UUID, len(orderLineItems))
	for i, li := range orderLineItems {
		orderLineItemIDs[i] = li.ID
	}

	// PICK TASKS -----------------------------------------------------------

	pickTasks, err := picktaskbus.TestSeedPickTasks(
		ctx, 2,
		[]uuid.UUID{orders[0].ID},
		orderLineItemIDs,
		productIDs,
		[]uuid.UUID{locations[0].LocationID, locations[1].LocationID},
		userIDs[:3],
		nil,
		busDomain.PickTask,
	)
	if err != nil {
		t.Fatalf("seed pick tasks: %v", err)
	}

	// SUPPLIER PRODUCTS (FK for PO line items) -----------------------------

	supplierProducts, err := supplierproductbus.TestSeedSupplierProducts(
		ctx, 2,
		uuid.UUIDs(productIDs),
		supplierIDs,
		busDomain.SupplierProduct,
	)
	if err != nil {
		t.Fatalf("seed supplier products: %v", err)
	}
	supplierProductIDs := make(uuid.UUIDs, len(supplierProducts))
	for i, sp := range supplierProducts {
		supplierProductIDs[i] = sp.SupplierProductID
	}

	// PO LINE ITEM STATUSES (FK for PO line items) -------------------------

	poLineItemStatuses, err := purchaseorderlineitemstatusbus.TestSeedPurchaseOrderLineItemStatuses(
		ctx, 3,
		busDomain.PurchaseOrderLineItemStatus,
	)
	if err != nil {
		t.Fatalf("seed PO line item statuses: %v", err)
	}
	poLineItemStatusIDs := make(uuid.UUIDs, len(poLineItemStatuses))
	for i, s := range poLineItemStatuses {
		poLineItemStatusIDs[i] = s.ID
	}

	// PO LINE ITEMS --------------------------------------------------------

	_, err = purchaseorderlineitembus.TestSeedPurchaseOrderLineItems(
		ctx, 2,
		uuid.UUIDs{pos[0].ID},
		supplierProductIDs,
		poLineItemStatusIDs,
		uuid.UUIDs(userIDs),
		busDomain.PurchaseOrderLineItem,
	)
	if err != nil {
		t.Fatalf("seed purchase order line items: %v", err)
	}

	// RESOLVE WAREHOUSE NAMES FOR TRANSFER ENRICHMENT ----------------------

	fromLocFull, err := busDomain.InventoryLocation.QueryByID(ctx, locations[0].LocationID)
	if err != nil {
		t.Fatalf("query from location: %v", err)
	}
	toLocFull, err := busDomain.InventoryLocation.QueryByID(ctx, locations[1].LocationID)
	if err != nil {
		t.Fatalf("query to location: %v", err)
	}
	srcWH, err := busDomain.Warehouse.QueryByID(ctx, fromLocFull.WarehouseID)
	if err != nil {
		t.Fatalf("query src warehouse: %v", err)
	}
	dstWH, err := busDomain.Warehouse.QueryByID(ctx, toLocFull.WarehouseID)
	if err != nil {
		t.Fatalf("query dst warehouse: %v", err)
	}

	return paperworkSeed{
		order:              orders[0],
		purchaseOrder:      pos[0],
		transferOrder:      tos[0],
		transferOrderNoNum: tos[1],
		customerName:       customers[0].Name,
		pickTaskCount:      len(pickTasks),
		vendorName:         suppliers[0].Name,
		srcWHName:          srcWH.Name,
		dstWHName:          dstWH.Name,
	}
}

// TestTaskCodeFor_Idempotent_ViaBuildPickSheet verifies the idempotent
// strip-then-prepend contract called out in paperworkapp.go: feeding a value
// that already begins with the "<prefix>-" head must NOT produce a
// double-prefixed task code in the rendered PDF.
//
// Drives the assertion through the public BuildPickSheet path by overriding
// a seeded order's Number to "SO-IDEMP-1". taskCodeFor("SO", "SO-IDEMP-1")
// must resolve to "SO-IDEMP-1", not "SO-SO-IDEMP-1".
func TestTaskCodeFor_Idempotent_ViaBuildPickSheet(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "TestTaskCodeFor_Idempotent")
	app := newTestApp(db)

	ctx := context.Background()
	seed := seedAll(t, ctx, db.BusDomain)

	const overlay = "SO-IDEMP-1"
	number := overlay
	updated, err := db.BusDomain.Order.Update(ctx, seed.order, ordersbus.UpdateOrder{Number: &number})
	if err != nil {
		t.Fatalf("update order number: %v", err)
	}

	got, err := app.BuildPickSheet(ctx, paperworkapp.PickSheetRequest{OrderID: updated.ID})
	if err != nil {
		t.Fatalf("BuildPickSheet: %v", err)
	}

	if !bytes.Contains(got, []byte(overlay)) {
		t.Fatalf("PDF missing expected task code %q", overlay)
	}
	if bytes.Contains(got, []byte("SO-"+overlay)) {
		t.Fatalf("PDF contains double-prefixed task code %q (taskCodeFor not idempotent)", "SO-"+overlay)
	}
}

// TestBuildPickSheet_ExcludesTerminalTasks verifies that BuildPickSheet lists
// only work still to be done: once every pick task for an order reaches a
// terminal status, the rendered sheet must contain none of their product SKUs.
// Guards the active-only status filter in BuildPickSheet against regression
// (a re-printed sheet must never show already-handled lines as actionable).
func TestBuildPickSheet_ExcludesTerminalTasks(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "TestBuildPickSheet_ExcludesTerminalTasks")
	app := newTestApp(db)

	ctx := context.Background()
	seed := seedAll(t, ctx, db.BusDomain)

	// Collect the seeded tasks and the SKUs they reference.
	tasks, err := db.BusDomain.PickTask.Query(ctx, picktaskbus.QueryFilter{SalesOrderID: &seed.order.ID}, picktaskbus.DefaultOrderBy, page.MustParse("1", "1000"))
	if err != nil {
		t.Fatalf("query pick tasks: %v", err)
	}
	if len(tasks) == 0 {
		t.Fatal("no pick tasks seeded — exclusion test is vacuous")
	}
	skus := make([]string, 0, len(tasks))
	for _, tk := range tasks {
		prod, err := db.BusDomain.Product.QueryByID(ctx, tk.ProductID)
		if err != nil {
			t.Fatalf("query product %s: %v", tk.ProductID, err)
		}
		skus = append(skus, prod.SKU)
	}

	// Baseline: while the tasks are pending, their SKUs appear on the sheet.
	before, err := app.BuildPickSheet(ctx, paperworkapp.PickSheetRequest{OrderID: seed.order.ID})
	if err != nil {
		t.Fatalf("BuildPickSheet (baseline): %v", err)
	}
	for _, sku := range skus {
		if !bytes.Contains(before, []byte(sku)) {
			t.Fatalf("baseline pick sheet missing SKU %q for a pending task", sku)
		}
	}

	// Drive every task to a terminal status.
	cancelled := picktaskbus.Statuses.Cancelled
	for _, tk := range tasks {
		if _, err := db.BusDomain.PickTask.Update(ctx, tk, picktaskbus.UpdatePickTask{Status: &cancelled}); err != nil {
			t.Fatalf("cancel pick task %s: %v", tk.ID, err)
		}
	}

	after, err := app.BuildPickSheet(ctx, paperworkapp.PickSheetRequest{OrderID: seed.order.ID})
	if err != nil {
		t.Fatalf("BuildPickSheet (after cancel): %v", err)
	}
	if !bytes.HasPrefix(after, pdfMagic) {
		t.Fatalf("pick sheet output does not start with %%PDF- (first 8 bytes: %q)", firstN(after, 8))
	}
	for _, sku := range skus {
		if bytes.Contains(after, []byte(sku)) {
			t.Errorf("pick sheet still contains SKU %q after its task was cancelled; terminal tasks must be excluded", sku)
		}
	}
}
