package purchaseorderbus_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

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
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_PurchaseOrder(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_PurchaseOrder")

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

	return unitest.SeedData{
		Admins:                []unitest.User{{User: admins[0]}},
		PurchaseOrders:        purchaseOrders,
		PurchaseOrderStatuses: poStatuses,
		Suppliers:             suppliers,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []purchaseorderbus.PurchaseOrder{
				sd.PurchaseOrders[0],
				sd.PurchaseOrders[1],
				sd.PurchaseOrders[2],
				sd.PurchaseOrders[3],
				sd.PurchaseOrders[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.PurchaseOrder.Query(ctx, purchaseorderbus.QueryFilter{}, purchaseorderbus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]purchaseorderbus.PurchaseOrder)
				sort.Slice(gotResp, func(i, j int) bool {
					return gotResp[i].ID.String() < gotResp[j].ID.String()
				})

				if !exists {
					return fmt.Sprintf("expected []purchaseorderbus.PurchaseOrder, got %T", got)
				}

				expResp, exists := exp.([]purchaseorderbus.PurchaseOrder)
				if !exists {
					return fmt.Sprintf("expected []purchaseorderbus.PurchaseOrder, got %T", exp)
				}
				sort.Slice(expResp, func(i, j int) bool {
					return expResp[i].ID.String() < expResp[j].ID.String()
				})

				for i, got := range gotResp {
					expResp[i].CreatedDate = got.CreatedDate
					expResp[i].UpdatedDate = got.UpdatedDate
					expResp[i].OrderDate = got.OrderDate
					expResp[i].ExpectedDeliveryDate = got.ExpectedDeliveryDate
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	subtotal := 5000.00
	tax := subtotal * 0.08
	shipping := 100.00
	total := subtotal + tax + shipping

	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: purchaseorderbus.PurchaseOrder{
				OrderNumber:              "PO-TEST-001",
				SupplierID:               sd.Suppliers[0].SupplierID,
				PurchaseOrderStatusID:    sd.PurchaseOrderStatuses[0].ID,
				DeliveryWarehouseID:      sd.PurchaseOrders[0].DeliveryWarehouseID,
				DeliveryLocationID:       uuid.Nil,
				DeliveryStreetID:         sd.PurchaseOrders[0].DeliveryStreetID,
				Subtotal:                 subtotal,
				TaxAmount:                tax,
				ShippingCost:             shipping,
				TotalAmount:              total,
				CurrencyID:                 sd.PurchaseOrders[0].CurrencyID,
				RequestedBy:              sd.Admins[0].ID,
				ApprovedBy:               uuid.Nil,
				Notes:                    "Test purchase order",
				SupplierReferenceNumber:  "SUP-REF-TEST",
				CreatedBy:                sd.Admins[0].ID,
				UpdatedBy:                sd.Admins[0].ID,
			},
			ExcFunc: func(ctx context.Context) any {
				po, err := busDomain.PurchaseOrder.Create(ctx, purchaseorderbus.NewPurchaseOrder{
					OrderNumber:              "PO-TEST-001",
					SupplierID:               sd.Suppliers[0].SupplierID,
					PurchaseOrderStatusID:    sd.PurchaseOrderStatuses[0].ID,
					DeliveryWarehouseID:      sd.PurchaseOrders[0].DeliveryWarehouseID,
					DeliveryLocationID:       uuid.Nil,
					DeliveryStreetID:         sd.PurchaseOrders[0].DeliveryStreetID,
					Subtotal:                 subtotal,
					TaxAmount:                tax,
					ShippingCost:             shipping,
					TotalAmount:              total,
					CurrencyID:                 sd.PurchaseOrders[0].CurrencyID,
					RequestedBy:              sd.Admins[0].ID,
					Notes:                    "Test purchase order",
					SupplierReferenceNumber:  "SUP-REF-TEST",
					CreatedBy:                sd.Admins[0].ID,
				})
				if err != nil {
					return err
				}
				return po
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(purchaseorderbus.PurchaseOrder)
				if !exists {
					return fmt.Sprintf("expected purchaseorderbus.PurchaseOrder, got %T", got)
				}

				expResp, exists := exp.(purchaseorderbus.PurchaseOrder)
				if !exists {
					return fmt.Sprintf("expected purchaseorderbus.PurchaseOrder, got %T", exp)
				}

				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate
				expResp.OrderDate = gotResp.OrderDate
				expResp.ExpectedDeliveryDate = gotResp.ExpectedDeliveryDate
				expResp.ActualDeliveryDate = gotResp.ActualDeliveryDate
				expResp.ApprovedDate = gotResp.ApprovedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	newNotes := "Updated purchase order notes"
	return []unitest.Table{
		{
			Name: "Update",
			ExpResp: purchaseorderbus.PurchaseOrder{
				ID:                       sd.PurchaseOrders[0].ID,
				OrderNumber:              sd.PurchaseOrders[0].OrderNumber,
				SupplierID:               sd.PurchaseOrders[0].SupplierID,
				PurchaseOrderStatusID:    sd.PurchaseOrders[0].PurchaseOrderStatusID,
				DeliveryWarehouseID:      sd.PurchaseOrders[0].DeliveryWarehouseID,
				DeliveryLocationID:       sd.PurchaseOrders[0].DeliveryLocationID,
				DeliveryStreetID:         sd.PurchaseOrders[0].DeliveryStreetID,
				OrderDate:                sd.PurchaseOrders[0].OrderDate,
				ExpectedDeliveryDate:     sd.PurchaseOrders[0].ExpectedDeliveryDate,
				ActualDeliveryDate:       sd.PurchaseOrders[0].ActualDeliveryDate,
				Subtotal:                 sd.PurchaseOrders[0].Subtotal,
				TaxAmount:                sd.PurchaseOrders[0].TaxAmount,
				ShippingCost:             sd.PurchaseOrders[0].ShippingCost,
				TotalAmount:              sd.PurchaseOrders[0].TotalAmount,
				CurrencyID:                 sd.PurchaseOrders[0].CurrencyID,
				RequestedBy:              sd.PurchaseOrders[0].RequestedBy,
				ApprovedBy:               sd.PurchaseOrders[0].ApprovedBy,
				ApprovedDate:             sd.PurchaseOrders[0].ApprovedDate,
				Notes:                    newNotes,
				SupplierReferenceNumber:  sd.PurchaseOrders[0].SupplierReferenceNumber,
				CreatedBy:                sd.PurchaseOrders[0].CreatedBy,
				UpdatedBy:                sd.PurchaseOrders[0].UpdatedBy,
				CreatedDate:              sd.PurchaseOrders[0].CreatedDate,
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.PurchaseOrder.Update(ctx, sd.PurchaseOrders[0], purchaseorderbus.UpdatePurchaseOrder{
					Notes: &newNotes,
				})
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(purchaseorderbus.PurchaseOrder)
				if !exists {
					return fmt.Sprintf("expected purchaseorderbus.PurchaseOrder, got %T", got)
				}

				expResp, exists := exp.(purchaseorderbus.PurchaseOrder)
				if !exists {
					return fmt.Sprintf("expected purchaseorderbus.PurchaseOrder, got %T", exp)
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
				err := busDomain.PurchaseOrder.Delete(ctx, sd.PurchaseOrders[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
