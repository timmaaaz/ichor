package inventoryitembus_test

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/lottrackingsbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/serialnumberbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	suptypes "github.com/timmaaaz/ichor/business/domain/procurement/supplierbus/types"
	spttypes "github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus/types"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_InventoryItem(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_InventoryItem")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------
	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding user : %w", err)
	}

	adminIDs := make([]uuid.UUID, len(admins))
	for i, admin := range admins {
		adminIDs[i] = admin.ID
	}

	count := 5

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("querying regions : %w", err)
	}

	ids := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		ids = append(ids, r.ID)
	}

	ctys, err := citybus.TestSeedCities(ctx, count, ids, busDomain.City)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding cities : %w", err)
	}

	ctyIDs := make([]uuid.UUID, 0, len(ctys))
	for _, c := range ctys {
		ctyIDs = append(ctyIDs, c.ID)
	}

	strs, err := streetbus.TestSeedStreets(ctx, count, ctyIDs, busDomain.Street)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding streets : %w", err)
	}
	strIDs := make([]uuid.UUID, 0, len(strs))
	for _, s := range strs {
		strIDs = append(strIDs, s.ID)
	}

	// Query timezones from seed data
	tzs, err := busDomain.Timezone.QueryAll(ctx)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("querying timezones : %w", err)
	}
	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 5, strIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding contact info : %w", err)
	}

	contactIDs := make(uuid.UUIDs, len(contactInfos))
	for i, c := range contactInfos {
		contactIDs[i] = c.ID
	}

	brand, err := brandbus.TestSeedBrands(ctx, 5, contactIDs, busDomain.Brand)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding brand : %w", err)
	}

	brandIDs := make(uuid.UUIDs, len(brand))
	for i, b := range brand {
		brandIDs[i] = b.BrandID
	}

	productCategories, err := productcategorybus.TestSeedProductCategories(ctx, 10, busDomain.ProductCategory)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding product category : %w", err)
	}

	productCategoryIDs := make(uuid.UUIDs, len(productCategories))

	for i, pc := range productCategories {
		productCategoryIDs[i] = pc.ProductCategoryID
	}

	products, err := productbus.TestSeedProducts(ctx, 20, brandIDs, productCategoryIDs, busDomain.Product)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding product : %w", err)
	}

	productIDs := make([]uuid.UUID, len(products))
	for i, p := range products {
		productIDs[i] = p.ProductID
	}

	warehouseCount := 5

	// WAREHOUSES
	warehouses, err := warehousebus.TestSeedWarehouses(ctx, warehouseCount, adminIDs[0], strIDs, busDomain.Warehouse)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding warehouses : %w", err)
	}

	warehouseIDs := make(uuid.UUIDs, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	zones, err := zonebus.TestSeedZone(ctx, 12, warehouseIDs, busDomain.Zones)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding zones : %w", err)
	}

	zoneIDs := make([]uuid.UUID, len(zones))
	for i, z := range zones {
		zoneIDs[i] = z.ZoneID
	}

	inventoryLocations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 25, warehouseIDs, zoneIDs, busDomain.InventoryLocation)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding inventory locations : %w", err)
	}

	inventoryLocationsIDs := make([]uuid.UUID, len(inventoryLocations))
	for i, il := range inventoryLocations {
		inventoryLocationsIDs[i] = il.LocationID
	}

	inventoryItems, err := inventoryitembus.TestSeedInventoryItems(ctx, 30, inventoryLocationsIDs, productIDs, busDomain.InventoryItem)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding inventory products : %w", err)
	}

	return unitest.SeedData{
		Admins:             []unitest.User{{User: admins[0]}},
		Products:           products,
		InventoryLocations: inventoryLocations,
		InventoryItems:     inventoryItems,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	// Filter to items belonging to Products[1] (grid positions 25-29, exactly 5 items).
	// Using a ProductID filter scopes the query to test-specific rows and avoids
	// contamination from the global seed's inventory items.
	p1ID := sd.Products[1].ProductID
	expItems := make([]inventoryitembus.InventoryItem, 0, 5)
	for _, item := range sd.InventoryItems {
		if item.ProductID == p1ID {
			expItems = append(expItems, item)
		}
	}

	return []unitest.Table{
		{
			Name:    "Query",
			ExpResp: expItems,
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.InventoryItem.Query(ctx, inventoryitembus.QueryFilter{ProductID: &p1ID}, inventoryitembus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]inventoryitembus.InventoryItem)
				if !exists {
					return fmt.Sprintf("got is not a slice of inventory products: %v", got)
				}

				expResp := exp.([]inventoryitembus.InventoryItem)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: inventoryitembus.InventoryItem{
				LocationID:            sd.InventoryLocations[5].LocationID,
				ProductID:             sd.Products[1].ProductID,
				Quantity:              10,
				ReservedQuantity:      15,
				AllocatedQuantity:     20,
				MinimumStock:          1,
				MaximumStock:          100,
				ReorderPoint:          5,
				EconomicOrderQuantity: 25,
				SafetyStock:           40,
				AvgDailyUsage:         6,
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.InventoryItem.Create(ctx, inventoryitembus.NewInventoryItem{
					LocationID:            sd.InventoryLocations[5].LocationID,
					ProductID:             sd.Products[1].ProductID,
					Quantity:              10,
					ReservedQuantity:      15,
					AllocatedQuantity:     20,
					MinimumStock:          1,
					MaximumStock:          100,
					ReorderPoint:          5,
					EconomicOrderQuantity: 25,
					SafetyStock:           40,
					AvgDailyUsage:         6,
				})
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(inventoryitembus.InventoryItem)
				if !exists {
					return fmt.Sprintf("got is not an inventory product: %v", got)
				}

				expResp := exp.(inventoryitembus.InventoryItem)
				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Update",
			ExpResp: inventoryitembus.InventoryItem{
				ID:                    sd.InventoryItems[0].ID,
				LocationID:            sd.InventoryLocations[6].LocationID,
				ProductID:             sd.Products[1].ProductID,
				Quantity:              20,
				ReservedQuantity:      25,
				AllocatedQuantity:     30,
				MinimumStock:          2,
				MaximumStock:          102,
				ReorderPoint:          10,
				EconomicOrderQuantity: 50,
				SafetyStock:           60,
				AvgDailyUsage:         10,
				CreatedDate:           sd.InventoryItems[0].CreatedDate,
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.InventoryItem.Update(ctx, sd.InventoryItems[0], inventoryitembus.UpdateInventoryItem{
					LocationID:            &sd.InventoryLocations[6].LocationID,
					ProductID:             &sd.Products[1].ProductID,
					Quantity:              dbtest.IntPointer(20),
					ReservedQuantity:      dbtest.IntPointer(25),
					AllocatedQuantity:     dbtest.IntPointer(30),
					MinimumStock:          dbtest.IntPointer(2),
					MaximumStock:          dbtest.IntPointer(102),
					ReorderPoint:          dbtest.IntPointer(10),
					EconomicOrderQuantity: dbtest.IntPointer(50),
					SafetyStock:           dbtest.IntPointer(60),
					AvgDailyUsage:         dbtest.IntPointer(10),
				})
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(inventoryitembus.InventoryItem)
				if !exists {
					return fmt.Sprintf("got is not an inventory product: %v", got)
				}

				expResp := exp.(inventoryitembus.InventoryItem)
				expResp.ID = gotResp.ID
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Delete",
			ExcFunc: func(ctx context.Context) any {
				return busDomain.InventoryItem.Delete(ctx, sd.InventoryItems[0])
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Test_QueryAvailableForAllocation exercises the strategy-based allocation
// query including FEFO, FIFO, and LIFO ordering.
// ---------------------------------------------------------------------------

func Test_QueryAvailableForAllocation(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_QueryAvailableForAllocation")
	bd := db.BusDomain
	ctx := context.Background()

	// --- Seed base entities (geography → contacts → products → locations) ---

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, bd.User)
	if err != nil {
		t.Fatalf("seeding users: %s", err)
	}

	regions, err := bd.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		t.Fatalf("querying regions: %s", err)
	}
	regionIDs := make([]uuid.UUID, len(regions))
	for i, r := range regions {
		regionIDs[i] = r.ID
	}

	cities, err := citybus.TestSeedCities(ctx, 2, regionIDs, bd.City)
	if err != nil {
		t.Fatalf("seeding cities: %s", err)
	}
	cityIDs := make([]uuid.UUID, len(cities))
	for i, c := range cities {
		cityIDs[i] = c.ID
	}

	streets, err := streetbus.TestSeedStreets(ctx, 2, cityIDs, bd.Street)
	if err != nil {
		t.Fatalf("seeding streets: %s", err)
	}
	streetIDs := make([]uuid.UUID, len(streets))
	for i, s := range streets {
		streetIDs[i] = s.ID
	}

	tzs, err := bd.Timezone.QueryAll(ctx)
	if err != nil {
		t.Fatalf("querying timezones: %s", err)
	}
	tzIDs := make([]uuid.UUID, len(tzs))
	for i, tz := range tzs {
		tzIDs[i] = tz.ID
	}

	contacts, err := contactinfosbus.TestSeedContactInfos(ctx, 2, streetIDs, tzIDs, bd.ContactInfos)
	if err != nil {
		t.Fatalf("seeding contacts: %s", err)
	}
	contactIDs := make([]uuid.UUID, len(contacts))
	for i, c := range contacts {
		contactIDs[i] = c.ID
	}

	brands, err := brandbus.TestSeedBrands(ctx, 1, contactIDs[:1], bd.Brand)
	if err != nil {
		t.Fatalf("seeding brands: %s", err)
	}

	categories, err := productcategorybus.TestSeedProductCategories(ctx, 1, bd.ProductCategory)
	if err != nil {
		t.Fatalf("seeding categories: %s", err)
	}

	products, err := productbus.TestSeedProducts(ctx, 1, uuid.UUIDs{brands[0].BrandID}, uuid.UUIDs{categories[0].ProductCategoryID}, bd.Product)
	if err != nil {
		t.Fatalf("seeding products: %s", err)
	}
	productID := products[0].ProductID

	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 1, admins[0].ID, streetIDs[:1], bd.Warehouse)
	if err != nil {
		t.Fatalf("seeding warehouses: %s", err)
	}

	zones, err := zonebus.TestSeedZone(ctx, 1, uuid.UUIDs{warehouses[0].ID}, bd.Zones)
	if err != nil {
		t.Fatalf("seeding zones: %s", err)
	}

	locations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 3, uuid.UUIDs{warehouses[0].ID}, uuid.UUIDs{zones[0].ZoneID}, bd.InventoryLocation)
	if err != nil {
		t.Fatalf("seeding locations: %s", err)
	}
	locIDs := make([]uuid.UUID, len(locations))
	for i, l := range locations {
		locIDs[i] = l.LocationID
	}

	// --- Create 3 inventory items for the same product at different locations ---
	// Each has available quantity (quantity > reserved + allocated).

	var items []inventoryitembus.InventoryItem
	for i, locID := range locIDs {
		item, err := bd.InventoryItem.Create(ctx, inventoryitembus.NewInventoryItem{
			ProductID:  productID,
			LocationID: locID,
			Quantity:   100 + i,
		})
		if err != nil {
			t.Fatalf("creating inventory item %d: %s", i, err)
		}
		items = append(items, item)
	}

	// --- Seed procurement chain: supplier → supplier product → lot trackings ---

	supplier, err := bd.Supplier.Create(ctx, supplierbus.NewSupplier{
		ContactInfosID: contactIDs[0],
		Name:           "Test Supplier",
		LeadTimeDays:   7,
		Rating:         suptypes.NewRoundedFloat(4.5),
		IsActive:       true,
	})
	if err != nil {
		t.Fatalf("creating supplier: %s", err)
	}

	sp, err := bd.SupplierProduct.Create(ctx, supplierproductbus.NewSupplierProduct{
		SupplierID:         supplier.SupplierID,
		ProductID:          productID,
		SupplierPartNumber: "SP-001",
		MinOrderQuantity:   1,
		MaxOrderQuantity:   1000,
		LeadTimeDays:       5,
		UnitCost:           spttypes.MustParseMoney("10.00"),
		IsPrimarySupplier:  true,
	})
	if err != nil {
		t.Fatalf("creating supplier product: %s", err)
	}

	// Create 3 lot trackings with controlled expiration dates.
	// L0 → expires 2025-06-01 (latest)
	// L1 → expires 2025-01-01 (earliest)
	// L2 → expires 2025-03-01 (middle)
	expirations := []time.Time{
		time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
	}

	lots := make([]lottrackingsbus.LotTrackings, 3)
	for i, exp := range expirations {
		lot, err := bd.LotTrackings.Create(ctx, lottrackingsbus.NewLotTrackings{
			SupplierProductID: sp.SupplierProductID,
			LotNumber:         fmt.Sprintf("LOT-%d", i),
			ManufactureDate:   exp.AddDate(0, -6, 0),
			ExpirationDate:    exp,
			RecievedDate:      time.Now(),
			Quantity:          100,
			QualityStatus:     "good",
		})
		if err != nil {
			t.Fatalf("creating lot tracking %d: %s", i, err)
		}
		lots[i] = lot
	}

	// Create serial numbers linking each inventory item's (product, location)
	// to the corresponding lot tracking.
	for i, item := range items {
		_, err := bd.SerialNumber.Create(ctx, serialnumberbus.NewSerialNumber{
			LotID:        lots[i].LotID,
			ProductID:    item.ProductID,
			LocationID:   item.LocationID,
			SerialNumber: fmt.Sprintf("SN-%d", i),
			Status:       "in_stock",
		})
		if err != nil {
			t.Fatalf("creating serial number %d: %s", i, err)
		}
	}

	// -----------------------------------------------------------------------
	// Test FEFO: should return items ordered by earliest expiration date.
	// Expected order: items[1] (2025-01-01), items[2] (2025-03-01), items[0] (2025-06-01)
	// -----------------------------------------------------------------------

	t.Run("fefo", func(t *testing.T) {
		got, err := bd.InventoryItem.QueryAvailableForAllocation(ctx, productID, nil, nil, "fefo", 10)
		if err != nil {
			t.Fatalf("FEFO query failed: %s", err)
		}

		if len(got) != 3 {
			t.Fatalf("expected 3 items, got %d", len(got))
		}

		// Verify ordering: earliest expiry first.
		expectedOrder := []uuid.UUID{items[1].ID, items[2].ID, items[0].ID}
		gotOrder := make([]uuid.UUID, len(got))
		for i, item := range got {
			gotOrder[i] = item.ID
		}

		if diff := cmp.Diff(expectedOrder, gotOrder); diff != "" {
			t.Errorf("FEFO ordering mismatch (-want +got):\n%s", diff)
		}
	})

	// -----------------------------------------------------------------------
	// Test FIFO: should return items ordered by created_date ASC.
	// -----------------------------------------------------------------------

	t.Run("fifo", func(t *testing.T) {
		got, err := bd.InventoryItem.QueryAvailableForAllocation(ctx, productID, nil, nil, "fifo", 10)
		if err != nil {
			t.Fatalf("FIFO query failed: %s", err)
		}

		if len(got) != 3 {
			t.Fatalf("expected 3 items, got %d", len(got))
		}

		// Verify ordering: oldest first.
		sorted := make([]inventoryitembus.InventoryItem, len(got))
		copy(sorted, got)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].CreatedDate.Before(sorted[j].CreatedDate)
		})

		for i := range got {
			if got[i].ID != sorted[i].ID {
				t.Errorf("FIFO ordering wrong at position %d: got %s, want %s", i, got[i].ID, sorted[i].ID)
			}
		}
	})

	// -----------------------------------------------------------------------
	// Test LIFO: should return items ordered by created_date DESC.
	// -----------------------------------------------------------------------

	t.Run("lifo", func(t *testing.T) {
		got, err := bd.InventoryItem.QueryAvailableForAllocation(ctx, productID, nil, nil, "lifo", 10)
		if err != nil {
			t.Fatalf("LIFO query failed: %s", err)
		}

		if len(got) != 3 {
			t.Fatalf("expected 3 items, got %d", len(got))
		}

		// Verify ordering: newest first.
		sorted := make([]inventoryitembus.InventoryItem, len(got))
		copy(sorted, got)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].CreatedDate.After(sorted[j].CreatedDate)
		})

		for i := range got {
			if got[i].ID != sorted[i].ID {
				t.Errorf("LIFO ordering wrong at position %d: got %s, want %s", i, got[i].ID, sorted[i].ID)
			}
		}
	})
}
