package transferorderbus_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_TransferOrders(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_TransferOrders")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// Find a pending transfer order for the lifecycle chain (approve → claim → execute).
	// After UUID sort, status-to-index mapping is non-deterministic with mixed statuses.
	lifecycleID := findTransferOrderIDByStatus(t, sd.TransferOrders, transferorderbus.StatusPending)

	// -------------------------------------------------------------------------
	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, approve(db.BusDomain, sd, lifecycleID), "approve")
	unitest.Run(t, claim(db.BusDomain, sd, lifecycleID), "claim")
	unitest.Run(t, execute(db.BusDomain, sd, lifecycleID), "execute")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

// findTransferOrderIDByStatus returns the TransferID of the first order with the given status.
func findTransferOrderIDByStatus(t *testing.T, orders []transferorderbus.TransferOrder, status string) uuid.UUID {
	t.Helper()
	for _, to := range orders {
		if to.Status == status {
			return to.TransferID
		}
	}
	t.Fatalf("no transfer order with status %q found in seed data", status)
	return uuid.Nil
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()
	warehouseCount := 5

	// USERS
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

	productIDs := make(uuid.UUIDs, len(products))
	for i, p := range products {
		productIDs[i] = p.ProductID
	}

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

	inventoryLocationIDs := make([]uuid.UUID, len(inventoryLocations))
	for i, il := range inventoryLocations {
		inventoryLocationIDs[i] = il.LocationID
	}

	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 10, userbus.Roles.User, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}
	userIDs := make([]uuid.UUID, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	transferOrders, err := transferorderbus.TestSeedTransferOrders(ctx, 20, productIDs, inventoryLocationIDs[:15], inventoryLocationIDs[15:], userIDs[:4], userIDs[4:], nil, busDomain.TransferOrder)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding transfer orders : %w", err)
	}

	return unitest.SeedData{
		Products:           products,
		Admins:             []unitest.User{{User: admins[0]}},
		InventoryLocations: inventoryLocations,
		TransferOrders:     transferOrders,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "query",
			ExpResp: sd.TransferOrders[:5],
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.TransferOrder.Query(ctx, transferorderbus.QueryFilter{}, transferorderbus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return fmt.Errorf("querying transfer orders: %w", err)
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]transferorderbus.TransferOrder)
				if !exists {
					return "unexpected response type"
				}

				expResp := exp.([]transferorderbus.TransferOrder)
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	now := time.Now()
	return []unitest.Table{{
		Name: "create",
		ExpResp: transferorderbus.TransferOrder{
			ProductID:      sd.Products[0].ProductID,
			FromLocationID: sd.InventoryLocations[0].LocationID,
			ToLocationID:   sd.InventoryLocations[3].LocationID,
			RequestedByID:  sd.TransferOrders[2].RequestedByID,
			ApprovedByID:   sd.TransferOrders[4].ApprovedByID,
			Quantity:       10,
			Status:         transferorderbus.StatusPending,
			TransferDate:   now,
		},
		ExcFunc: func(ctx context.Context) any {
			got, err := busDomain.TransferOrder.Create(ctx, transferorderbus.NewTransferOrder{
				ProductID:      sd.Products[0].ProductID,
				FromLocationID: sd.InventoryLocations[0].LocationID,
				ToLocationID:   sd.InventoryLocations[3].LocationID,
				RequestedByID:  sd.TransferOrders[2].RequestedByID,
				ApprovedByID:   sd.TransferOrders[4].ApprovedByID,
				Quantity:       10,
				Status:         transferorderbus.StatusPending,
				TransferDate:   now,
			})
			if err != nil {
				return fmt.Errorf("creating transfer order: %w", err)
			}
			return got
		},
		CmpFunc: func(got, exp any) string {
			gotResp, exists := got.(transferorderbus.TransferOrder)
			if !exists {
				return fmt.Sprintf("got is not an transferorderbus.TransferOrder: %v", got)
			}
			expResp := exp.(transferorderbus.TransferOrder)

			expResp.TransferID = gotResp.TransferID
			expResp.CreatedDate = gotResp.CreatedDate
			expResp.UpdatedDate = gotResp.UpdatedDate

			return cmp.Diff(gotResp, expResp)
		},
	}}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	now := time.Now()

	return []unitest.Table{{
		Name: "update",
		ExpResp: transferorderbus.TransferOrder{
			TransferID:     sd.TransferOrders[0].TransferID,
			ProductID:      sd.Products[0].ProductID,
			FromLocationID: sd.InventoryLocations[0].LocationID,
			ToLocationID:   sd.InventoryLocations[3].LocationID,
			RequestedByID:  sd.TransferOrders[2].RequestedByID,
			ApprovedByID:   sd.TransferOrders[4].ApprovedByID,
			Quantity:       15,
			Status:         transferorderbus.StatusPending,
			TransferDate:   now,
			CreatedDate:    sd.TransferOrders[0].CreatedDate,
		},
		ExcFunc: func(ctx context.Context) any {
			updateTO := transferorderbus.UpdateTransferOrder{

				ProductID:      &sd.Products[0].ProductID,
				FromLocationID: &sd.InventoryLocations[0].LocationID,
				ToLocationID:   &sd.InventoryLocations[3].LocationID,
				RequestedByID:  &sd.TransferOrders[2].RequestedByID,
				ApprovedByID:   sd.TransferOrders[4].ApprovedByID,
				Quantity:       dbtest.IntPointer(15),
				Status:         dbtest.StringPointer(transferorderbus.StatusPending),
				TransferDate:   &now,
			}

			got, err := busDomain.TransferOrder.Update(ctx, sd.TransferOrders[0], updateTO)
			if err != nil {
				return fmt.Errorf("updating transfer order: %w", err)
			}

			return got
		},
		CmpFunc: func(got, exp any) string {
			gotResp, exists := got.(transferorderbus.TransferOrder)
			if !exists {
				return "got is not a transfer order"
			}

			expResp := exp.(transferorderbus.TransferOrder)

			expResp.TransferID = gotResp.TransferID
			expResp.UpdatedDate = gotResp.UpdatedDate

			return cmp.Diff(gotResp, expResp)
		},
	}}
}

func approve(busDomain dbtest.BusDomain, sd unitest.SeedData, lifecycleID uuid.UUID) []unitest.Table {
	approverID := sd.Admins[0].ID
	return []unitest.Table{
		{
			Name:    "approve-pending-succeeds",
			ExpResp: transferorderbus.StatusApproved,
			ExcFunc: func(ctx context.Context) any {
				// Query the lifecycle order (known to be pending).
				to, err := busDomain.TransferOrder.QueryByID(ctx, lifecycleID)
				if err != nil {
					return fmt.Errorf("query: %w", err)
				}
				approved, err := busDomain.TransferOrder.Approve(ctx, to, approverID, "approved for transfer")
				if err != nil {
					return fmt.Errorf("approving transfer order: %w", err)
				}
				return approved.Status
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "approve-approved-fails",
			ExpResp: transferorderbus.ErrInvalidTransferStatus,
			ExcFunc: func(ctx context.Context) any {
				// Re-query to get the approved order from previous test.
				to, err := busDomain.TransferOrder.QueryByID(ctx, lifecycleID)
				if err != nil {
					return fmt.Errorf("query: %w", err)
				}
				_, err = busDomain.TransferOrder.Approve(ctx, to, approverID, "double approve")
				return err
			},
			CmpFunc: func(got, exp any) string {
				gotErr, ok := got.(error)
				if !ok {
					return "expected error"
				}
				expErr := exp.(error)
				if !errors.Is(gotErr, expErr) {
					return fmt.Sprintf("got %v, exp %v", gotErr, expErr)
				}
				return ""
			},
		},
	}
}

func claim(busDomain dbtest.BusDomain, sd unitest.SeedData, lifecycleID uuid.UUID) []unitest.Table {
	claimerID := sd.Admins[0].ID

	// Find a non-approved order for the failure test (claim requires approved status).
	var failTO transferorderbus.TransferOrder
	for _, to := range sd.TransferOrders {
		if to.Status != transferorderbus.StatusApproved && to.TransferID != lifecycleID {
			failTO = to
			break
		}
	}

	return []unitest.Table{
		{
			Name:    "claim-pending-fails",
			ExpResp: transferorderbus.ErrInvalidTransferStatus,
			ExcFunc: func(ctx context.Context) any {
				_, err := busDomain.TransferOrder.Claim(ctx, failTO, claimerID)
				return err
			},
			CmpFunc: func(got, exp any) string {
				gotErr, ok := got.(error)
				if !ok {
					return "expected error"
				}
				expErr := exp.(error)
				if !errors.Is(gotErr, expErr) {
					return fmt.Sprintf("got %v, exp %v", gotErr, expErr)
				}
				return ""
			},
		},
		{
			Name:    "claim-approved-succeeds",
			ExpResp: transferorderbus.StatusInTransit,
			ExcFunc: func(ctx context.Context) any {
				// Re-query lifecycle order which is now approved from the approve test.
				to, err := busDomain.TransferOrder.QueryByID(ctx, lifecycleID)
				if err != nil {
					return fmt.Errorf("query: %w", err)
				}
				claimed, err := busDomain.TransferOrder.Claim(ctx, to, claimerID)
				if err != nil {
					return fmt.Errorf("claiming transfer order: %w", err)
				}
				return claimed.Status
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func execute(busDomain dbtest.BusDomain, sd unitest.SeedData, lifecycleID uuid.UUID) []unitest.Table {
	executorID := sd.Admins[0].ID

	// Find a non-in-transit order for the failure test (execute requires in_transit status).
	var failTO transferorderbus.TransferOrder
	for _, to := range sd.TransferOrders {
		if to.Status != transferorderbus.StatusInTransit && to.TransferID != lifecycleID {
			failTO = to
			break
		}
	}

	return []unitest.Table{
		{
			Name:    "execute-pending-fails",
			ExpResp: transferorderbus.ErrInvalidTransferStatus,
			ExcFunc: func(ctx context.Context) any {
				_, err := busDomain.TransferOrder.Execute(ctx, failTO, executorID)
				return err
			},
			CmpFunc: func(got, exp any) string {
				gotErr, ok := got.(error)
				if !ok {
					return "expected error"
				}
				expErr := exp.(error)
				if !errors.Is(gotErr, expErr) {
					return fmt.Sprintf("got %v, exp %v", gotErr, expErr)
				}
				return ""
			},
		},
		{
			Name:    "execute-in-transit-succeeds",
			ExpResp: transferorderbus.StatusCompleted,
			ExcFunc: func(ctx context.Context) any {
				// Re-query lifecycle order which is now in_transit from the claim test.
				to, err := busDomain.TransferOrder.QueryByID(ctx, lifecycleID)
				if err != nil {
					return fmt.Errorf("query: %w", err)
				}
				completed, err := busDomain.TransferOrder.Execute(ctx, to, executorID)
				if err != nil {
					return fmt.Errorf("executing transfer order: %w", err)
				}
				return completed.Status
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func TestTransferOrder_ClaimedByIDFilter(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_TransferOrder_ClaimedByIDFilter")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("seeding: %s", err)
	}

	ctx := context.Background()

	// Set claimed_by on a known transfer order via Update so we have something
	// to filter on. Seeded orders never have ClaimedByID populated because
	// NewTransferOrder/Create don't accept it.
	target := sd.TransferOrders[0]
	claimer := sd.Admins[0].ID
	if _, err := db.BusDomain.TransferOrder.Update(ctx, target, transferorderbus.UpdateTransferOrder{
		ClaimedByID: &claimer,
	}); err != nil {
		t.Fatalf("update claimed_by: %s", err)
	}

	// Positive case: filter returns only the rows whose ClaimedByID matches.
	filter := transferorderbus.QueryFilter{ClaimedByID: &claimer}
	got, err := db.BusDomain.TransferOrder.Query(ctx, filter, transferorderbus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("query with filter: %s", err)
	}
	if len(got) == 0 {
		t.Fatal("expected at least one transfer with ClaimedByID filter; got 0")
	}
	for _, to := range got {
		if to.ClaimedByID == nil || *to.ClaimedByID != claimer {
			t.Errorf("filter leaked: got transfer %s with ClaimedByID=%v", to.TransferID, to.ClaimedByID)
		}
	}

	// Negative case: a random UUID that is not a claimer should return zero rows,
	// proving the WHERE clause is actually being applied.
	other := uuid.New()
	gotNone, err := db.BusDomain.TransferOrder.Query(ctx, transferorderbus.QueryFilter{ClaimedByID: &other}, transferorderbus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("query with no-match filter: %s", err)
	}
	if len(gotNone) != 0 {
		t.Errorf("expected zero rows for unused ClaimedByID; got %d", len(gotNone))
	}
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{{
		Name: "delete",
		ExcFunc: func(ctx context.Context) any {
			err := busDomain.TransferOrder.Delete(ctx, sd.TransferOrders[0])
			if err != nil {
				return fmt.Errorf("deleting transfer order: %w", err)
			}
			return nil
		},
		CmpFunc: func(got, exp any) string {
			return cmp.Diff(got, exp)
		},
	}}
}
