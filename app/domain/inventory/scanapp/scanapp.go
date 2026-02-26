// Package scanapp provides a scan resolver that maps a barcode to its inventory context.
package scanapp

import (
	"context"
	"fmt"
	"sync"

	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/lottrackingsbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/serialnumberbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App orchestrates a barcode scan across four inventory domains.
type App struct {
	productBus       *productbus.Business
	inventoryItemBus *inventoryitembus.Business
	locationBus      *inventorylocationbus.Business
	lotTrackingsBus  *lottrackingsbus.Business
	serialNumberBus  *serialnumberbus.Business
}

// NewApp constructs a scan resolver App.
func NewApp(
	productBus *productbus.Business,
	inventoryItemBus *inventoryitembus.Business,
	locationBus *inventorylocationbus.Business,
	lotTrackingsBus *lottrackingsbus.Business,
	serialNumberBus *serialnumberbus.Business,
) *App {
	return &App{
		productBus:       productBus,
		inventoryItemBus: inventoryItemBus,
		locationBus:      locationBus,
		lotTrackingsBus:  lotTrackingsBus,
		serialNumberBus:  serialNumberBus,
	}
}

// Scan resolves a barcode to its most specific inventory context.
// Priority order: serial > lot > product > location.
// Returns type "unknown" with nil data when the barcode is not found.
// Individual domain query errors are treated as no-match (fail open) so a
// transient failure in one domain does not block results from others.
func (a *App) Scan(ctx context.Context, barcode string) (ScanResult, error) {
	pg, err := page.Parse("1", "1")
	if err != nil {
		panic(fmt.Sprintf("scanapp: invalid page constants: %v", err))
	}
	defaultOrder := order.NewBy("id", order.ASC)

	var (
		mu        sync.Mutex
		products  []productbus.Product
		locations []inventorylocationbus.InventoryLocation
		lots      []lottrackingsbus.LotTrackings
		serials   []serialnumberbus.SerialNumber
	)

	var wg sync.WaitGroup
	wg.Add(4)

	// goroutine 1: product by UPC
	go func() {
		defer wg.Done()
		results, err := a.productBus.Query(ctx, productbus.QueryFilter{UpcCode: &barcode}, defaultOrder, pg)
		if err != nil {
			return // fail open — barcode simply won't match as a product
		}
		mu.Lock()
		defer mu.Unlock()
		products = results
	}()

	// goroutine 2: location by exact code
	go func() {
		defer wg.Done()
		results, err := a.locationBus.Query(ctx, inventorylocationbus.QueryFilter{LocationCodeExact: &barcode}, defaultOrder, pg)
		if err != nil {
			return // fail open
		}
		mu.Lock()
		defer mu.Unlock()
		locations = results
	}()

	// goroutine 3: lot by lot number
	go func() {
		defer wg.Done()
		results, err := a.lotTrackingsBus.Query(ctx, lottrackingsbus.QueryFilter{LotNumber: &barcode}, defaultOrder, pg)
		if err != nil {
			return // fail open
		}
		mu.Lock()
		defer mu.Unlock()
		lots = results
	}()

	// goroutine 4: serial by serial number
	go func() {
		defer wg.Done()
		results, err := a.serialNumberBus.Query(ctx, serialnumberbus.QueryFilter{SerialNumber: &barcode}, defaultOrder, pg)
		if err != nil {
			return // fail open
		}
		mu.Lock()
		defer mu.Unlock()
		serials = results
	}()

	wg.Wait()

	// Priority: serial > lot > product > location
	if len(serials) > 0 {
		data, err := a.enrichSerial(ctx, serials[0])
		if err != nil {
			return ScanResult{}, fmt.Errorf("enrich serial: %w", err)
		}
		return ScanResult{Type: "serial", Data: data}, nil
	}

	if len(lots) > 0 {
		data, err := a.enrichLot(ctx, lots[0])
		if err != nil {
			return ScanResult{}, fmt.Errorf("enrich lot: %w", err)
		}
		return ScanResult{Type: "lot", Data: data}, nil
	}

	if len(products) > 0 {
		data, err := a.enrichProduct(ctx, products[0])
		if err != nil {
			return ScanResult{}, fmt.Errorf("enrich product: %w", err)
		}
		return ScanResult{Type: "product", Data: data}, nil
	}

	if len(locations) > 0 {
		data, err := a.enrichLocation(ctx, locations[0])
		if err != nil {
			return ScanResult{}, fmt.Errorf("enrich location: %w", err)
		}
		return ScanResult{Type: "location", Data: data}, nil
	}

	return ScanResult{Type: "unknown", Data: nil}, nil
}

func (a *App) enrichSerial(ctx context.Context, sn serialnumberbus.SerialNumber) (SerialScanResult, error) {
	loc, err := a.serialNumberBus.QueryLocationBySerialID(ctx, sn.SerialID)
	if err != nil {
		// Location lookup failure shouldn't block the result — return partial data.
		return SerialScanResult{
			SerialID:     sn.SerialID.String(),
			SerialNumber: sn.SerialNumber,
			ProductID:    sn.ProductID.String(),
			LotID:        sn.LotID.String(),
			Status:       sn.Status,
			LocationID:   sn.LocationID.String(),
		}, nil
	}

	return SerialScanResult{
		SerialID:      sn.SerialID.String(),
		SerialNumber:  sn.SerialNumber,
		ProductID:     sn.ProductID.String(),
		LotID:         sn.LotID.String(),
		Status:        sn.Status,
		LocationID:    loc.LocationID.String(),
		LocationCode:  loc.LocationCode,
		Aisle:         loc.Aisle,
		Rack:          loc.Rack,
		Shelf:         loc.Shelf,
		Bin:           loc.Bin,
		WarehouseName: loc.WarehouseName,
		ZoneName:      loc.ZoneName,
	}, nil
}

func (a *App) enrichLot(ctx context.Context, lot lottrackingsbus.LotTrackings) (LotScanResult, error) {
	locations, err := a.lotTrackingsBus.QueryLocationsByLotID(ctx, lot.LotID)
	if err != nil {
		locations = nil // fail open — return result without locations
	}

	locs := make([]LotLocationEntry, len(locations))
	for i, l := range locations {
		locs[i] = LotLocationEntry{
			LocationID:   l.LocationID.String(),
			LocationCode: l.LocationCode,
			Aisle:        l.Aisle,
			Rack:         l.Rack,
			Shelf:        l.Shelf,
			Bin:          l.Bin,
			Quantity:     l.Quantity,
		}
	}

	return LotScanResult{
		LotID:         lot.LotID.String(),
		LotNumber:     lot.LotNumber,
		ProductID:     lot.ProductID.String(),
		ProductName:   lot.ProductName,
		ProductSKU:    lot.ProductSKU,
		QualityStatus: lot.QualityStatus,
		Quantity:      lot.Quantity,
		Locations:     locs,
	}, nil
}

func (a *App) enrichProduct(ctx context.Context, product productbus.Product) (ProductScanResult, error) {
	// Use QueryWithLocationDetails to get location info per inventory row for this product.
	allPages, err := page.Parse("1", "100")
	if err != nil {
		panic(fmt.Sprintf("scanapp: invalid page constants: %v", err))
	}
	defaultOrder := order.NewBy("id", order.ASC)

	items, err := a.inventoryItemBus.QueryWithLocationDetails(
		ctx,
		inventoryitembus.QueryFilter{ProductID: &product.ProductID},
		defaultOrder,
		allPages,
	)
	if err != nil {
		items = nil
	}

	stock := make([]StockAtLocation, len(items))
	for i, item := range items {
		stock[i] = StockAtLocation{
			LocationID:       item.LocationID.String(),
			LocationCode:     item.LocationCode,
			Quantity:         item.Quantity,
			ReservedQuantity: item.ReservedQuantity,
		}
	}

	return ProductScanResult{
		ProductID:    product.ProductID.String(),
		Name:         product.Name,
		SKU:          product.SKU,
		TrackingType: product.TrackingType,
		StockSummary: stock,
	}, nil
}

func (a *App) enrichLocation(ctx context.Context, loc inventorylocationbus.InventoryLocation) (LocationScanResult, error) {
	locationCode := ""
	if loc.LocationCode != nil {
		locationCode = *loc.LocationCode
	}

	items, err := a.inventoryItemBus.QueryItemsWithProductAtLocation(ctx, loc.LocationID)
	if err != nil {
		items = nil
	}

	itemList := make([]ItemAtLocation, len(items))
	for i, item := range items {
		itemList[i] = ItemAtLocation{
			ProductID:    item.ProductID.String(),
			ProductName:  item.ProductName,
			ProductSKU:   item.ProductSKU,
			TrackingType: item.TrackingType,
			Quantity:     item.Quantity,
		}
	}

	return LocationScanResult{
		LocationID:   loc.LocationID.String(),
		LocationCode: locationCode,
		Aisle:        loc.Aisle,
		Rack:         loc.Rack,
		Shelf:        loc.Shelf,
		Bin:          loc.Bin,
		Items:        itemList,
	}, nil
}
