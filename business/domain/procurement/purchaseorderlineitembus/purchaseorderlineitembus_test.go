package purchaseorderlineitembus_test

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitemstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_PurchaseOrderLineItem(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_PurchaseOrderLineItem")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding user: %w", err)
	}
	userIDs := make([]uuid.UUID, 0, len(admins))
	for _, a := range admins {
		userIDs = append(userIDs, a.ID)
	}

	count := 5

	// Regions for addresses
	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("querying regions: %w", err)
	}
	regionIDs := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		regionIDs = append(regionIDs, r.ID)
	}

	cities, err := citybus.TestSeedCities(ctx, count, regionIDs, busDomain.City)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding cities: %w", err)
	}
	cityIDs := make([]uuid.UUID, 0, len(cities))
	for _, c := range cities {
		cityIDs = append(cityIDs, c.ID)
	}

	streets, err := streetbus.TestSeedStreets(ctx, count, cityIDs, busDomain.Street)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding streets: %w", err)
	}
	streetIDs := make([]uuid.UUID, 0, len(streets))
	for _, s := range streets {
		streetIDs = append(streetIDs, s.ID)
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

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, count, streetIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding contact info: %w", err)
	}
	contactInfoIDs := make([]uuid.UUID, 0, len(contactInfos))
	for _, ci := range contactInfos {
		contactInfoIDs = append(contactInfoIDs, ci.ID)
	}

	// Suppliers
	suppliers, err := supplierbus.TestSeedSuppliers(ctx, count, contactInfoIDs, busDomain.Supplier)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding suppliers: %w", err)
	}
	supplierIDs := make([]uuid.UUID, 0, len(suppliers))
	for _, s := range suppliers {
		supplierIDs = append(supplierIDs, s.SupplierID)
	}

	// Products for supplier products
	brands, err := brandbus.TestSeedBrands(ctx, 5, contactInfoIDs, busDomain.Brand)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding brands: %w", err)
	}
	brandIDs := make(uuid.UUIDs, len(brands))
	for i, b := range brands {
		brandIDs[i] = b.BrandID
	}

	productCategories, err := productcategorybus.TestSeedProductCategories(ctx, 10, busDomain.ProductCategory)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding product categories: %w", err)
	}
	productCategoryIDs := make(uuid.UUIDs, len(productCategories))
	for i, pc := range productCategories {
		productCategoryIDs[i] = pc.ProductCategoryID
	}

	products, err := productbus.TestSeedProducts(ctx, 20, brandIDs, productCategoryIDs, busDomain.Product)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding products: %w", err)
	}
	productIDs := make([]uuid.UUID, 0, len(products))
	for _, p := range products {
		productIDs = append(productIDs, p.ProductID)
	}

	// Supplier products
	supplierProducts, err := supplierproductbus.TestSeedSupplierProducts(ctx, count, productIDs, supplierIDs, busDomain.SupplierProduct)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding supplier products: %w", err)
	}
	supplierProductIDs := make([]uuid.UUID, 0, len(supplierProducts))
	for _, sp := range supplierProducts {
		supplierProductIDs = append(supplierProductIDs, sp.SupplierProductID)
	}

	// Warehouses
	warehouses, err := warehousebus.TestSeedWarehouses(ctx, count, userIDs[0], streetIDs, busDomain.Warehouse)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding warehouses: %w", err)
	}
	warehouseIDs := make([]uuid.UUID, 0, len(warehouses))
	for _, w := range warehouses {
		warehouseIDs = append(warehouseIDs, w.ID)
	}

	// Purchase order statuses
	poStatuses, err := purchaseorderstatusbus.TestSeedPurchaseOrderStatuses(ctx, 3, busDomain.PurchaseOrderStatus)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding purchase order statuses: %w", err)
	}
	poStatusIDs := make([]uuid.UUID, 0, len(poStatuses))
	for _, pos := range poStatuses {
		poStatusIDs = append(poStatusIDs, pos.ID)
	}

	// Purchase order line item statuses
	poliStatuses, err := purchaseorderlineitemstatusbus.TestSeedPurchaseOrderLineItemStatuses(ctx, 3, busDomain.PurchaseOrderLineItemStatus)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding purchase order line item statuses: %w", err)
	}
	poliStatusIDs := make([]uuid.UUID, 0, len(poliStatuses))
	for _, polis := range poliStatuses {
		poliStatusIDs = append(poliStatusIDs, polis.ID)
	}

	// Currencies
	currencies, err := currencybus.TestSeedCurrencies(ctx, 5, busDomain.Currency)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding currencies: %w", err)
	}
	currencyIDs := make(uuid.UUIDs, len(currencies))
	for i, c := range currencies {
		currencyIDs[i] = c.ID
	}

	// Purchase orders
	purchaseOrders, err := purchaseorderbus.TestSeedPurchaseOrders(ctx, count, supplierIDs, poStatusIDs, warehouseIDs, streetIDs, userIDs, currencyIDs, busDomain.PurchaseOrder)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding purchase orders: %w", err)
	}
	purchaseOrderIDs := make([]uuid.UUID, 0, len(purchaseOrders))
	for _, po := range purchaseOrders {
		purchaseOrderIDs = append(purchaseOrderIDs, po.ID)
	}

	// Purchase order line items
	polis, err := purchaseorderlineitembus.TestSeedPurchaseOrderLineItems(ctx, count, purchaseOrderIDs, supplierProductIDs, poliStatusIDs, userIDs, busDomain.PurchaseOrderLineItem)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding purchase order line items: %w", err)
	}

	return unitest.SeedData{
		Admins:                        []unitest.User{{User: admins[0]}},
		PurchaseOrders:                purchaseOrders,
		PurchaseOrderStatuses:         poStatuses,
		PurchaseOrderLineItemStatuses: poliStatuses,
		PurchaseOrderLineItems:        polis,
		Suppliers:                     suppliers,
		SupplierProducts:              supplierProducts,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []purchaseorderlineitembus.PurchaseOrderLineItem{
				sd.PurchaseOrderLineItems[0],
				sd.PurchaseOrderLineItems[1],
				sd.PurchaseOrderLineItems[2],
				sd.PurchaseOrderLineItems[3],
				sd.PurchaseOrderLineItems[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.PurchaseOrderLineItem.Query(ctx, purchaseorderlineitembus.QueryFilter{}, purchaseorderlineitembus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]purchaseorderlineitembus.PurchaseOrderLineItem)
				sort.Slice(gotResp, func(i, j int) bool {
					return gotResp[i].ID.String() < gotResp[j].ID.String()
				})

				if !exists {
					return fmt.Sprintf("expected []purchaseorderlineitembus.PurchaseOrderLineItem, got %T", got)
				}

				expResp, exists := exp.([]purchaseorderlineitembus.PurchaseOrderLineItem)
				if !exists {
					return fmt.Sprintf("expected []purchaseorderlineitembus.PurchaseOrderLineItem, got %T", exp)
				}
				sort.Slice(expResp, func(i, j int) bool {
					return expResp[i].ID.String() < expResp[j].ID.String()
				})

				for i, got := range gotResp {
					expResp[i].CreatedDate = got.CreatedDate
					expResp[i].UpdatedDate = got.UpdatedDate
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: purchaseorderlineitembus.PurchaseOrderLineItem{
				PurchaseOrderID:      sd.PurchaseOrders[0].ID,
				SupplierProductID:    sd.SupplierProducts[0].SupplierProductID,
				QuantityOrdered:      100,
				QuantityReceived:     0,
				QuantityCancelled:    0,
				UnitCost:             50.00,
				Discount:             5.00,
				LineTotal:            4995.00,
				LineItemStatusID:     sd.PurchaseOrderLineItemStatuses[0].ID,
				ExpectedDeliveryDate: time.Now().UTC().Add(time.Hour * 24 * 7),
				Notes:                "Test line item",
				CreatedBy:            sd.Admins[0].ID,
				UpdatedBy:            sd.Admins[0].ID,
			},
			ExcFunc: func(ctx context.Context) any {
				poli, err := busDomain.PurchaseOrderLineItem.Create(ctx, purchaseorderlineitembus.NewPurchaseOrderLineItem{
					PurchaseOrderID:      sd.PurchaseOrders[0].ID,
					SupplierProductID:    sd.SupplierProducts[0].SupplierProductID,
					QuantityOrdered:      100,
					UnitCost:             50.00,
					Discount:             5.00,
					LineTotal:            4995.00,
					LineItemStatusID:     sd.PurchaseOrderLineItemStatuses[0].ID,
					ExpectedDeliveryDate: time.Now().UTC().Add(time.Hour * 24 * 7),
					Notes:                "Test line item",
					CreatedBy:            sd.Admins[0].ID,
				})
				if err != nil {
					return err
				}
				return poli
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(purchaseorderlineitembus.PurchaseOrderLineItem)
				if !exists {
					return fmt.Sprintf("expected purchaseorderlineitembus.PurchaseOrderLineItem, got %T", got)
				}

				expResp, exists := exp.(purchaseorderlineitembus.PurchaseOrderLineItem)
				if !exists {
					return fmt.Sprintf("expected purchaseorderlineitembus.PurchaseOrderLineItem, got %T", exp)
				}

				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate
				expResp.ExpectedDeliveryDate = gotResp.ExpectedDeliveryDate
				expResp.ActualDeliveryDate = gotResp.ActualDeliveryDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	newQuantity := 200
	return []unitest.Table{
		{
			Name: "Update",
			ExpResp: purchaseorderlineitembus.PurchaseOrderLineItem{
				ID:                   sd.PurchaseOrderLineItems[0].ID,
				PurchaseOrderID:      sd.PurchaseOrderLineItems[0].PurchaseOrderID,
				SupplierProductID:    sd.PurchaseOrderLineItems[0].SupplierProductID,
				QuantityOrdered:      newQuantity,
				QuantityReceived:     sd.PurchaseOrderLineItems[0].QuantityReceived,
				QuantityCancelled:    sd.PurchaseOrderLineItems[0].QuantityCancelled,
				UnitCost:             sd.PurchaseOrderLineItems[0].UnitCost,
				Discount:             sd.PurchaseOrderLineItems[0].Discount,
				LineTotal:            sd.PurchaseOrderLineItems[0].LineTotal,
				LineItemStatusID:     sd.PurchaseOrderLineItems[0].LineItemStatusID,
				ExpectedDeliveryDate: sd.PurchaseOrderLineItems[0].ExpectedDeliveryDate,
				ActualDeliveryDate:   sd.PurchaseOrderLineItems[0].ActualDeliveryDate,
				Notes:                sd.PurchaseOrderLineItems[0].Notes,
				CreatedBy:            sd.PurchaseOrderLineItems[0].CreatedBy,
				UpdatedBy:            sd.PurchaseOrderLineItems[0].UpdatedBy,
				CreatedDate:          sd.PurchaseOrderLineItems[0].CreatedDate,
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.PurchaseOrderLineItem.Update(ctx, sd.PurchaseOrderLineItems[0], purchaseorderlineitembus.UpdatePurchaseOrderLineItem{
					QuantityOrdered: &newQuantity,
				})
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(purchaseorderlineitembus.PurchaseOrderLineItem)
				if !exists {
					return fmt.Sprintf("expected purchaseorderlineitembus.PurchaseOrderLineItem, got %T", got)
				}

				expResp, exists := exp.(purchaseorderlineitembus.PurchaseOrderLineItem)
				if !exists {
					return fmt.Sprintf("expected purchaseorderlineitembus.PurchaseOrderLineItem, got %T", exp)
				}

				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.PurchaseOrderLineItem.Delete(ctx, sd.PurchaseOrderLineItems[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
