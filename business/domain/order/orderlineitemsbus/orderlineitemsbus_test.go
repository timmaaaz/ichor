package orderlineitemsbus_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/customersbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	"github.com/timmaaaz/ichor/business/domain/order/lineitemfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/order/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/order/orderlineitemsbus"
	"github.com/timmaaaz/ichor/business/domain/order/ordersbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_OrderLineItem(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_OrderLineItem")

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
		return unitest.SeedData{}, fmt.Errorf("seeding user : %w", err)
	}
	userIDs := make([]uuid.UUID, 0, len(admins))
	for _, a := range admins {
		userIDs = append(userIDs, a.ID)
	}

	count := 5

	// ADDRESSES
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

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, count, strIDs, busDomain.ContactInfos)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding contact info : %w", err)
	}
	contactInfoIDs := make([]uuid.UUID, 0, len(contactInfos))
	for _, ci := range contactInfos {
		contactInfoIDs = append(contactInfoIDs, ci.ID)
	}

	customers, err := customersbus.TestSeedCustomers(ctx, count, strIDs, contactInfoIDs, uuid.UUIDs{admins[0].ID}, busDomain.Customers)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding customers : %w", err)
	}
	customerIDs := make([]uuid.UUID, 0, len(customers))
	for _, c := range customers {
		customerIDs = append(customerIDs, c.ID)
	}

	ofls, err := orderfulfillmentstatusbus.TestSeedOrderFulfillmentStatuses(ctx, busDomain.OrderFulfillmentStatus)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding order fulfillment statuses: %w", err)
	}
	oflIDs := make([]uuid.UUID, 0, len(ofls))
	for _, ofl := range ofls {
		oflIDs = append(oflIDs, ofl.ID)
	}

	orders, err := ordersbus.TestSeedOrders(ctx, count, uuid.UUIDs{admins[0].ID}, customerIDs, oflIDs, busDomain.Order)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding Orders: %w", err)
	}
	orderIDs := make([]uuid.UUID, 0, len(orders))
	for _, o := range orders {
		orderIDs = append(orderIDs, o.ID)
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
	productIDs := make([]uuid.UUID, 0, len(products))
	for _, p := range products {
		productIDs = append(productIDs, p.ProductID)
	}

	olStatuses, err := lineitemfulfillmentstatusbus.TestSeedLineItemFulfillmentStatuses(ctx, busDomain.LineItemFulfillmentStatus)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding line item fulfillment statuses: %w", err)
	}
	olStatusIDs := make([]uuid.UUID, 0, len(olStatuses))
	for _, ols := range olStatuses {
		olStatusIDs = append(olStatusIDs, ols.ID)
	}

	ols, err := orderlineitemsbus.TestSeedOrderLineItems(ctx, count, orderIDs, productIDs, olStatusIDs, userIDs, busDomain.OrderLineItem)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding Order Line Items: %w", err)
	}

	return unitest.SeedData{
		Admins:                   []unitest.User{{User: admins[0]}},
		Orders:                   orders,
		Products:                 products,
		OrderFulfillmentStatuses: ofls,
		OrderLineItems:           ols,
		Customers:                customers,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {

	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []orderlineitemsbus.OrderLineItem{
				sd.OrderLineItems[0],
				sd.OrderLineItems[1],
				sd.OrderLineItems[2],
				sd.OrderLineItems[3],
				sd.OrderLineItems[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.OrderLineItem.Query(ctx, orderlineitemsbus.QueryFilter{}, orderlineitemsbus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]orderlineitemsbus.OrderLineItem)
				sort.Slice(gotResp, func(i, j int) bool {
					return gotResp[i].ID.String() < gotResp[j].ID.String()
				})

				if !exists {
					return fmt.Sprintf("expected []orderlineitemsbus.OrderLineItem, got %T", got)
				}

				expResp, exists := exp.([]orderlineitemsbus.OrderLineItem)
				if !exists {
					return fmt.Sprintf("expected []orderlineitemsbus.OrderLineItem, got %T", exp)
				}
				sort.Slice(expResp, func(i, j int) bool {
					return expResp[i].ID.String() < expResp[j].ID.String()
				})

				for i, got := range gotResp {
					expResp[i].CreatedDate = got.CreatedDate // Ignore ID for comparison
					expResp[i].UpdatedDate = got.UpdatedDate // Ignore ID for comparison
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
			ExpResp: orderlineitemsbus.OrderLineItem{
				OrderID:                       sd.Orders[0].ID,
				ProductID:                     sd.Products[0].ProductID,
				Quantity:                      1,
				LineItemFulfillmentStatusesID: sd.OrderLineItems[0].LineItemFulfillmentStatusesID,
				CreatedBy:                     sd.Admins[0].ID,
				UpdatedBy:                     sd.Admins[0].ID,
			},
			ExcFunc: func(ctx context.Context) any {
				ofs, err := busDomain.OrderLineItem.Create(ctx, orderlineitemsbus.NewOrderLineItem{
					OrderID:                       sd.Orders[0].ID,
					ProductID:                     sd.Products[0].ProductID,
					Quantity:                      1,
					LineItemFulfillmentStatusesID: sd.OrderLineItems[0].LineItemFulfillmentStatusesID,
					CreatedBy:                     sd.Admins[0].ID,
				})
				if err != nil {
					return err
				}
				return ofs
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(orderlineitemsbus.OrderLineItem)
				if !exists {
					return fmt.Sprintf("expected orderlineitemsbus.OrderLineItem, got %T", got)
				}

				expResp, exists := exp.(orderlineitemsbus.OrderLineItem)
				if !exists {
					return fmt.Sprintf("expected orderlineitemsbus.OrderLineItem, got %T", exp)
				}

				expResp.ID = gotResp.ID // Ignore ID for comparison

				expResp.CreatedDate = gotResp.CreatedDate // Ignore ID for comparison
				expResp.UpdatedDate = gotResp.UpdatedDate // Ignore ID for comparison

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Update",
			ExpResp: orderlineitemsbus.OrderLineItem{
				ID:                            sd.OrderLineItems[0].ID,
				OrderID:                       sd.OrderLineItems[0].OrderID,
				ProductID:                     sd.Products[1].ProductID,
				Quantity:                      sd.OrderLineItems[0].Quantity,
				LineItemFulfillmentStatusesID: sd.OrderLineItems[0].LineItemFulfillmentStatusesID,
				CreatedBy:                     sd.OrderLineItems[0].CreatedBy,
				UpdatedBy:                     sd.OrderLineItems[0].UpdatedBy,
				CreatedDate:                   sd.OrderLineItems[0].CreatedDate,
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.OrderLineItem.Update(ctx, sd.OrderLineItems[0], orderlineitemsbus.UpdateOrderLineItem{
					ProductID: &sd.Products[1].ProductID,
				})
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(orderlineitemsbus.OrderLineItem)
				if !exists {
					return fmt.Sprintf("expected orderlineitemsbus.OrderLineItem, got %T", got)
				}

				expResp, exists := exp.(orderlineitemsbus.OrderLineItem)
				if !exists {
					return fmt.Sprintf("expected orderlineitemsbus.OrderLineItem, got %T", exp)
				}

				expResp.UpdatedDate = gotResp.UpdatedDate // Ignore ID for comparison

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
				err := busDomain.OrderLineItem.Delete(ctx, sd.OrderLineItems[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
