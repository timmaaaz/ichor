package dbtest

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
)

// InventorySeed holds the results of seeding inventory data.
type InventorySeed struct {
	Warehouses         []warehousebus.Warehouse
	InventoryLocations []inventorylocationbus.InventoryLocation
}

func seedInventory(ctx context.Context, busDomain BusDomain, foundation FoundationSeed, geoHR GeographyHRSeed, products ProductsSeed) (InventorySeed, error) {
	warehouseCount := 5

	strIDs := make([]uuid.UUID, 0, len(geoHR.Streets))
	for _, s := range geoHR.Streets {
		strIDs = append(strIDs, s.ID)
	}

	productIDs := make([]uuid.UUID, 0, len(products.Products))
	for _, p := range products.Products {
		productIDs = append(productIDs, p.ProductID)
	}

	reporterIDs := make([]uuid.UUID, len(foundation.Reporters))
	for i, r := range foundation.Reporters {
		reporterIDs[i] = r.ID
	}

	bossIDs := make([]uuid.UUID, len(foundation.Bosses))
	for i, b := range foundation.Bosses {
		bossIDs[i] = b.ID
	}

	userIDs := make([]uuid.UUID, 0, len(foundation.Admins))
	for _, a := range foundation.Admins {
		userIDs = append(userIDs, a.ID)
	}

	// WAREHOUSES
	warehouses, err := warehousebus.TestSeedWarehousesHistorical(ctx, warehouseCount, 365, foundation.Admins[0].ID, strIDs, busDomain.Warehouse)
	if err != nil {
		return InventorySeed{}, fmt.Errorf("seeding warehouses : %w", err)
	}

	warehouseIDs := make(uuid.UUIDs, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	zones, err := zonebus.TestSeedZone(ctx, 12, warehouseIDs, busDomain.Zones)
	if err != nil {
		return InventorySeed{}, fmt.Errorf("seeding zones : %w", err)
	}

	zoneIDs := make([]uuid.UUID, len(zones))
	for i, z := range zones {
		zoneIDs[i] = z.ZoneID
	}

	inventoryLocations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 25, warehouseIDs, zoneIDs, busDomain.InventoryLocation)
	if err != nil {
		return InventorySeed{}, fmt.Errorf("seeding inventory locations : %w", err)
	}

	inventoryLocationsIDs := make([]uuid.UUID, len(inventoryLocations))
	for i, il := range inventoryLocations {
		inventoryLocationsIDs[i] = il.LocationID
	}

	_, err = inventoryitembus.TestSeedInventoryItems(ctx, 30, inventoryLocationsIDs, productIDs, busDomain.InventoryItem)
	if err != nil {
		return InventorySeed{}, fmt.Errorf("seeding inventory products : %w", err)
	}

	_, err = transferorderbus.TestSeedTransferOrders(ctx, 20, productIDs, inventoryLocationsIDs[:15], inventoryLocationsIDs[15:], reporterIDs[:4], bossIDs[4:], busDomain.TransferOrder)
	if err != nil {
		return InventorySeed{}, fmt.Errorf("seeding transfer orders : %w", err)
	}

	_, err = inventorytransactionbus.TestSeedInventoryTransaction(ctx, 40, inventoryLocationsIDs, productIDs, userIDs, busDomain.InventoryTransaction)
	if err != nil {
		return InventorySeed{}, fmt.Errorf("seeding inventory transactions : %w", err)
	}

	_, err = inventoryadjustmentbus.TestSeedInventoryAdjustments(ctx, 20, productIDs, inventoryLocationsIDs, reporterIDs[:2], busDomain.InventoryAdjustment)
	if err != nil {
		return InventorySeed{}, fmt.Errorf("seeding inventory adjustments : %w", err)
	}

	return InventorySeed{
		Warehouses:         warehouses,
		InventoryLocations: inventoryLocations,
	}, nil
}
