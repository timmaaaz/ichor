package dbtest

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/inspectionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/lotlocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/lottrackingsbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/serialnumberbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitemstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest/seedmodels"
)

func seedProcurement(ctx context.Context, busDomain BusDomain, foundation FoundationSeed, geoHR GeographyHRSeed, products ProductsSeed, inventory InventorySeed) error {
	contactIDs := make(uuid.UUIDs, len(geoHR.ContactInfos))
	for i, c := range geoHR.ContactInfos {
		contactIDs[i] = c.ID
	}

	strIDs := make([]uuid.UUID, 0, len(geoHR.Streets))
	for _, s := range geoHR.Streets {
		strIDs = append(strIDs, s.ID)
	}

	productIDs := make([]uuid.UUID, 0, len(products.Products))
	for _, p := range products.Products {
		productIDs = append(productIDs, p.ProductID)
	}

	warehouseIDs := make(uuid.UUIDs, len(inventory.Warehouses))
	for i, w := range inventory.Warehouses {
		warehouseIDs[i] = w.ID
	}

	inventoryLocationsIDs := make([]uuid.UUID, len(inventory.InventoryLocations))
	for i, il := range inventory.InventoryLocations {
		inventoryLocationsIDs[i] = il.LocationID
	}

	userIDs := make([]uuid.UUID, 0, len(foundation.Admins))
	for _, a := range foundation.Admins {
		userIDs = append(userIDs, a.ID)
	}

	currencyIDs := make(uuid.UUIDs, len(foundation.Currencies))
	for i, c := range foundation.Currencies {
		currencyIDs[i] = c.ID
	}

	suppliers, err := supplierbus.TestSeedSuppliers(ctx, 25, contactIDs, busDomain.Supplier)
	if err != nil {
		return fmt.Errorf("seeding suppliers : %w", err)
	}

	supplierIDs := make(uuid.UUIDs, len(suppliers))
	for i, s := range suppliers {
		supplierIDs[i] = s.SupplierID
	}

	supplierProducts, err := supplierproductbus.TestSeedSupplierProducts(ctx, 10, productIDs, supplierIDs, busDomain.SupplierProduct)
	if err != nil {
		return fmt.Errorf("seeding supplier product : %w", err)
	}

	supplierProductIDs := make(uuid.UUIDs, len(supplierProducts))
	for i, sp := range supplierProducts {
		supplierProductIDs[i] = sp.SupplierProductID
	}

	// Purchase Order Statuses - Create meaningful statuses
	poStatuses := make([]purchaseorderstatusbus.PurchaseOrderStatus, 0, len(seedmodels.PurchaseOrderStatusData))
	for _, data := range seedmodels.PurchaseOrderStatusData {
		ps, err := busDomain.PurchaseOrderStatus.Create(ctx, purchaseorderstatusbus.NewPurchaseOrderStatus{
			Name:        data.Name,
			Description: data.Description,
			SortOrder:   data.SortOrder,
		})
		if err != nil {
			return fmt.Errorf("seeding purchase order status %s: %w", data.Name, err)
		}
		poStatuses = append(poStatuses, ps)
	}
	poStatusIDs := make(uuid.UUIDs, len(poStatuses))
	for i, ps := range poStatuses {
		poStatusIDs[i] = ps.ID
	}

	// Purchase Order Line Item Statuses - Create meaningful statuses
	poLineItemStatuses := make([]purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus, 0, len(seedmodels.PurchaseOrderLineItemStatusData))
	for _, data := range seedmodels.PurchaseOrderLineItemStatusData {
		pols, err := busDomain.PurchaseOrderLineItemStatus.Create(ctx, purchaseorderlineitemstatusbus.NewPurchaseOrderLineItemStatus{
			Name:        data.Name,
			Description: data.Description,
			SortOrder:   data.SortOrder,
		})
		if err != nil {
			return fmt.Errorf("seeding purchase order line item status %s: %w", data.Name, err)
		}
		poLineItemStatuses = append(poLineItemStatuses, pols)
	}
	poLineItemStatusIDs := make(uuid.UUIDs, len(poLineItemStatuses))
	for i, pols := range poLineItemStatuses {
		poLineItemStatusIDs[i] = pols.ID
	}

	// Purchase Orders
	purchaseOrders, err := purchaseorderbus.TestSeedPurchaseOrdersHistorical(ctx, 10, 120, supplierIDs, poStatusIDs, warehouseIDs, strIDs, userIDs, currencyIDs, busDomain.PurchaseOrder)
	if err != nil {
		return fmt.Errorf("seeding purchase orders : %w", err)
	}
	purchaseOrderIDs := make(uuid.UUIDs, len(purchaseOrders))
	for i, po := range purchaseOrders {
		purchaseOrderIDs[i] = po.ID
	}

	// Purchase Order Line Items
	_, err = purchaseorderlineitembus.TestSeedPurchaseOrderLineItems(ctx, 25, purchaseOrderIDs, supplierProductIDs, poLineItemStatusIDs, userIDs, busDomain.PurchaseOrderLineItem)
	if err != nil {
		return fmt.Errorf("seeding purchase order line items : %w", err)
	}

	lotTrackings, err := lottrackingsbus.TestSeedLotTrackings(ctx, 15, supplierProductIDs, busDomain.LotTrackings)
	if err != nil {
		return fmt.Errorf("seeding lot tracking : %w", err)
	}
	lotTrackingsIDs := make(uuid.UUIDs, len(lotTrackings))
	for i, lt := range lotTrackings {
		lotTrackingsIDs[i] = lt.LotID
	}

	_, err = lotlocationbus.TestSeedLotLocations(ctx, 15, lotTrackingsIDs, inventoryLocationsIDs, busDomain.LotLocation)
	if err != nil {
		return fmt.Errorf("seeding lot locations : %w", err)
	}

	_, err = inspectionbus.TestSeedInspections(ctx, 10, productIDs, userIDs, lotTrackingsIDs, busDomain.Inspection)
	if err != nil {
		return fmt.Errorf("seeding inspections : %w", err)
	}

	_, err = serialnumberbus.TestSeedSerialNumbers(ctx, 50, lotTrackingsIDs, productIDs, inventoryLocationsIDs, busDomain.SerialNumber)
	if err != nil {
		return fmt.Errorf("seeding serial numbers : %w", err)
	}

	return nil
}
