// Package paperworkbus_test verifies the paperwork bus renders non-empty PDFs
// containing the expected SO-/PO-/XFER- task codes when the requested order
// exists, and surfaces the right sentinel errors when it does not.
//
// The acceptance bar is intentionally minimal: each Build* method must produce
// bytes that begin with the "%PDF-" magic and contain the expected task code
// somewhere in the document. Visual layout is covered by the pdf subpackage
// tests (picksheet_test.go etc.); cross-domain wiring is covered here.
package paperworkbus_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/domain/paperwork/paperworkbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
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
}

func TestPaperworkBus_Integration(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "TestPaperworkBus_Integration")

	bus := paperworkbus.NewBusiness(
		db.Log,
		db.BusDomain.Order,
		db.BusDomain.OrderLineItem,
		db.BusDomain.PurchaseOrder,
		db.BusDomain.PurchaseOrderLineItem,
		db.BusDomain.TransferOrder,
	)

	ctx := context.Background()
	seed := seedAll(t, ctx, db.BusDomain)

	t.Run("BuildPickSheet", func(t *testing.T) {
		got, err := bus.BuildPickSheet(ctx, paperworkbus.PickSheetRequest{OrderID: seed.order.ID})
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

	t.Run("BuildPickSheet_NotFound", func(t *testing.T) {
		_, err := bus.BuildPickSheet(ctx, paperworkbus.PickSheetRequest{OrderID: uuid.New()})
		if !errors.Is(err, paperworkbus.ErrOrderNotFound) {
			t.Fatalf("BuildPickSheet: want ErrOrderNotFound for missing order, got %v", err)
		}
	})

	t.Run("BuildReceiveCover", func(t *testing.T) {
		got, err := bus.BuildReceiveCover(ctx, paperworkbus.ReceiveCoverRequest{PurchaseOrderID: seed.purchaseOrder.ID})
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

	t.Run("BuildTransferSheet", func(t *testing.T) {
		if seed.transferOrder.TransferNumber == nil {
			t.Fatalf("seedAll returned transfer order with nil TransferNumber")
		}

		got, err := bus.BuildTransferSheet(ctx, paperworkbus.TransferSheetRequest{TransferID: seed.transferOrder.TransferID})
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
}

// firstN returns the first n bytes of b for diagnostic printing without
// panicking on short inputs.
func firstN(b []byte, n int) []byte {
	if len(b) < n {
		return b
	}
	return b[:n]
}

// expectedTaskCode mirrors paperworkbus.taskCodeFor — kept as a separate test
// helper because the production helper is unexported. The two MUST stay in
// lockstep; if production grows a knob, update this helper.
func expectedTaskCode(prefix, value string) string {
	return prefix + "-" + strings.TrimPrefix(value, prefix+"-")
}

// seedAll provisions the union of FK chains the three Build* methods need:
// addresses, customers, suppliers, warehouses, zones, locations, products,
// statuses, currency, plus exactly one Order, PurchaseOrder, and
// TransferOrder. Returns the three seeded entities.
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
	fromIDs := []uuid.UUID{locations[0].LocationID}
	toIDs := []uuid.UUID{locations[1].LocationID}

	tos, err := transferorderbus.TestSeedTransferOrders(ctx, 1, productIDs, fromIDs, toIDs, userIDs[:3], userIDs[3:], nil, busDomain.TransferOrder)
	if err != nil {
		t.Fatalf("seed transfer orders: %v", err)
	}

	return paperworkSeed{
		order:         orders[0],
		purchaseOrder: pos[0],
		transferOrder: tos[0],
	}
}

// TestTaskCodeFor_Idempotent_ViaBuildPickSheet verifies the idempotent
// strip-then-prepend contract called out in paperworkbus.go: feeding a value
// that already begins with the "<prefix>-" head must NOT produce a
// double-prefixed task code in the rendered PDF.
//
// Drives the assertion through the public BuildPickSheet path by overriding
// a seeded order's Number to "SO-IDEMP-1". taskCodeFor("SO", "SO-IDEMP-1")
// must resolve to "SO-IDEMP-1", not "SO-SO-IDEMP-1".
func TestTaskCodeFor_Idempotent_ViaBuildPickSheet(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "TestTaskCodeFor_Idempotent")

	bus := paperworkbus.NewBusiness(
		db.Log,
		db.BusDomain.Order,
		db.BusDomain.OrderLineItem,
		db.BusDomain.PurchaseOrder,
		db.BusDomain.PurchaseOrderLineItem,
		db.BusDomain.TransferOrder,
	)

	ctx := context.Background()
	seed := seedAll(t, ctx, db.BusDomain)

	const overlay = "SO-IDEMP-1"
	number := overlay
	updated, err := db.BusDomain.Order.Update(ctx, seed.order, ordersbus.UpdateOrder{Number: &number})
	if err != nil {
		t.Fatalf("update order number: %v", err)
	}

	got, err := bus.BuildPickSheet(ctx, paperworkbus.PickSheetRequest{OrderID: updated.ID})
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
