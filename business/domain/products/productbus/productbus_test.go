package productbus_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus/stores/productdb"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/seedid"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_Product(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_Product")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)

	}

	// -------------------------------------------------------------------------
	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, seedCreate(db.BusDomain, sd), "seedCreate")
	// delegateFires takes *dbtest.Database (not BusDomain) — it needs db.DB and db.Log
	// to construct an observable parallel productbus.Business. See delegateFires godoc.
	unitest.Run(t, delegateFires(db, sd), "delegateFires")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {

	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding user : %w", err)
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

	return unitest.SeedData{
		Admins:            []unitest.User{{User: admins[0]}},
		Brands:            brand,
		ProductCategories: productCategories,
		Products:          products,
	}, nil

}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "query",
			ExpResp: sd.Products[:5],
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.Product.Query(ctx, productbus.QueryFilter{}, productbus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]productbus.Product)
				if !exists {
					return "error occurred"
				}

				expResp := exp.([]productbus.Product)

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {

	// Use a fixed index well outside the seeded range (rand.Intn(10000)+1..20)
	// to avoid UpcCode unique constraint violations.
	const newIdx = 99999
	product := sd.Products[0]

	newProduct := productbus.NewProduct{
		BrandID:              product.BrandID,
		ProductCategoryID:    product.ProductCategoryID,
		Name:                 fmt.Sprintf("Product%d", newIdx),
		Description:          fmt.Sprintf("Description%d", newIdx),
		SKU:                  fmt.Sprintf("SKU%d", newIdx),
		ModelNumber:          fmt.Sprintf("ModelNumber%d", newIdx),
		UpcCode:              fmt.Sprintf("UpcCode%d", newIdx),
		Status:               fmt.Sprintf("Status%d", newIdx),
		IsActive:             newIdx%2 != 0,
		IsPerishable:         false,
		HandlingInstructions: fmt.Sprintf("Handling instructions %d", newIdx),
		UnitsPerCase:         newIdx * 5,
	}

	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: productbus.Product{
				BrandID:              product.BrandID,
				ProductCategoryID:    product.ProductCategoryID,
				Name:                 newProduct.Name,
				Description:         newProduct.Description,
				SKU:                  newProduct.SKU,
				ModelNumber:          newProduct.ModelNumber,
				UpcCode:              newProduct.UpcCode,
				Status:               newProduct.Status,
				IsActive:             newProduct.IsActive,
				IsPerishable:         newProduct.IsPerishable,
				HandlingInstructions: newProduct.HandlingInstructions,
				UnitsPerCase:         newProduct.UnitsPerCase,
				TrackingType:         "none",
			},
			ExcFunc: func(ctx context.Context) any {
				p, err := busDomain.Product.Create(ctx, newProduct)
				if err != nil {
					return err
				}
				return p
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(productbus.Product)
				if !exists {
					return "error occurred"
				}

				expResp, ok := exp.(productbus.Product)
				if !ok {
					return "expected product, got something else"
				}
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate
				expResp.ProductID = gotResp.ProductID

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func seedCreate(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	// Two fixtures verify SeedCreate's behavior:
	//   A: zero CreatedDate / UpdatedDate / TrackingType — defaults applied
	//   B: caller-supplied CreatedDate + TrackingType — preserved verbatim
	stableA := seedid.Stable("test:productbus:seedCreate:fixture-A")
	stableB := seedid.Stable("test:productbus:seedCreate:fixture-B")

	historicalDate := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	productA := productbus.Product{
		ProductID:            stableA,
		SKU:                  "SEED-TEST-A",
		BrandID:              sd.Brands[0].BrandID,
		ProductCategoryID:    sd.ProductCategories[0].ProductCategoryID,
		Name:                 "SeedCreate Fixture A",
		Description:          "deterministic seed test",
		ModelNumber:          "SC-A",
		UpcCode:              "SeedTestA-UPC",
		Status:               "active",
		IsActive:             true,
		IsPerishable:         false,
		HandlingInstructions: "",
		UnitsPerCase:         12,
		// TrackingType, CreatedDate, UpdatedDate intentionally zero — must default.
	}

	productB := productbus.Product{
		ProductID:            stableB,
		SKU:                  "SEED-TEST-B",
		BrandID:              sd.Brands[0].BrandID,
		ProductCategoryID:    sd.ProductCategories[0].ProductCategoryID,
		Name:                 "SeedCreate Fixture B",
		Description:          "deterministic seed test (preserved fields)",
		ModelNumber:          "SC-B",
		UpcCode:              "SeedTestB-UPC",
		Status:               "active",
		IsActive:             true,
		IsPerishable:         true,
		HandlingInstructions: "keep cool",
		UnitsPerCase:         24,
		TrackingType:         "lot",
		CreatedDate:          historicalDate,
		UpdatedDate:          historicalDate,
	}

	return []unitest.Table{
		{
			Name:    "SeedCreate_DefaultsApplied",
			ExpResp: stableA,
			ExcFunc: func(ctx context.Context) any {
				if err := busDomain.Product.SeedCreate(ctx, productA); err != nil {
					return err
				}
				got, err := busDomain.Product.QueryByID(ctx, stableA)
				if err != nil {
					return err
				}
				if got.ProductID != stableA {
					return fmt.Errorf("ProductID drift: got %s want %s", got.ProductID, stableA)
				}
				if got.TrackingType != "none" {
					return fmt.Errorf("TrackingType default missed: got %q want %q", got.TrackingType, "none")
				}
				if got.CreatedDate.IsZero() {
					return fmt.Errorf("CreatedDate not defaulted")
				}
				if got.UpdatedDate.IsZero() {
					return fmt.Errorf("UpdatedDate not defaulted")
				}
				if !got.UpdatedDate.Equal(got.CreatedDate) {
					return fmt.Errorf("UpdatedDate (%s) should equal CreatedDate (%s) when both default", got.UpdatedDate, got.CreatedDate)
				}
				return got.ProductID
			},
			CmpFunc: func(got, exp any) string {
				gotID, ok := got.(uuid.UUID)
				if !ok {
					if err, isErr := got.(error); isErr {
						return err.Error()
					}
					return "expected uuid.UUID"
				}
				return cmp.Diff(exp.(uuid.UUID), gotID)
			},
		},
		{
			Name:    "SeedCreate_PreservesCallerFields",
			ExpResp: stableB,
			ExcFunc: func(ctx context.Context) any {
				if err := busDomain.Product.SeedCreate(ctx, productB); err != nil {
					return err
				}
				got, err := busDomain.Product.QueryByID(ctx, stableB)
				if err != nil {
					return err
				}
				if got.ProductID != stableB {
					return fmt.Errorf("ProductID drift: got %s want %s", got.ProductID, stableB)
				}
				if !got.CreatedDate.Equal(historicalDate) {
					return fmt.Errorf("CreatedDate not preserved: got %s want %s", got.CreatedDate, historicalDate)
				}
				if !got.UpdatedDate.Equal(historicalDate) {
					return fmt.Errorf("UpdatedDate not preserved: got %s want %s", got.UpdatedDate, historicalDate)
				}
				if got.TrackingType != "lot" {
					return fmt.Errorf("TrackingType not preserved: got %q want %q", got.TrackingType, "lot")
				}
				return got.ProductID
			},
			CmpFunc: func(got, exp any) string {
				gotID, ok := got.(uuid.UUID)
				if !ok {
					if err, isErr := got.(error); isErr {
						return err.Error()
					}
					return "expected uuid.UUID"
				}
				return cmp.Diff(exp.(uuid.UUID), gotID)
			},
		},
	}
}

// capturedDelegate records what a delegate handler observed.
type capturedDelegate struct {
	Domain string
	Action string
	Count  int
}

// delegateFires verifies that productbus.Create fires the ActionCreated
// delegate event. SeedCreate (added in this same PR) intentionally
// SKIPS this side-effect; pinning Create's behavior here makes that
// skip verifiable as deliberate rather than an oversight.
//
// The sub-test builds a parallel productbus.Business that shares the
// real DB connection (via productdb.NewStore on db.DB) but uses a
// fresh delegate.New so the handler can be intercepted. The pre-wired
// db.BusDomain.Product holds an unobservable production delegate; the
// only way to assert event firing without mocks is to construct an
// observable Business beside it.
func delegateFires(db *dbtest.Database, sd unitest.SeedData) []unitest.Table {
	del := delegate.New(db.Log)

	var (
		mu             sync.Mutex
		capturedDomain string
		capturedAction string
		capturedCount  int
	)
	del.Register(productbus.DomainName, productbus.ActionCreated, func(_ context.Context, data delegate.Data) error {
		mu.Lock()
		defer mu.Unlock()
		capturedDomain = data.Domain
		capturedAction = data.Action
		capturedCount++
		return nil
	})

	// Real storer against the same DB the rest of Test_Product uses.
	bus := productbus.NewBusiness(db.Log, del, productdb.NewStore(db.Log, db.DB))

	np := productbus.NewProduct{
		SKU:                  "DELEGATE-REAL-001",
		BrandID:              sd.Brands[0].BrandID,
		ProductCategoryID:    sd.ProductCategories[0].ProductCategoryID,
		Name:                 "Delegate Real-DB Fixture",
		Description:         "verifies Create fires ActionCreated through delegate",
		ModelNumber:          "DLG-REAL",
		UpcCode:              "DELEGATE-REAL-UPC-001",
		Status:               "active",
		IsActive:             true,
		IsPerishable:         false,
		HandlingInstructions: "",
		UnitsPerCase:         1,
		TrackingType:         "none",
	}

	return []unitest.Table{
		{
			Name: "Create_FiresActionCreated",
			ExpResp: capturedDelegate{
				Domain: productbus.DomainName,
				Action: productbus.ActionCreated,
				Count:  1,
			},
			ExcFunc: func(ctx context.Context) any {
				if _, err := bus.Create(ctx, np); err != nil {
					return err
				}
				mu.Lock()
				defer mu.Unlock()
				return capturedDelegate{
					Domain: capturedDomain,
					Action: capturedAction,
					Count:  capturedCount,
				}
			},
			CmpFunc: func(got, exp any) string {
				if err, isErr := got.(error); isErr {
					return err.Error()
				}
				return cmp.Diff(exp, got)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {

	idx := rand.Intn(200)
	product := sd.Products[idx%len(sd.Products)]

	idx++
	updateProduct := sd.Products[idx%len(sd.Products)]

	// Generate fresh unique values for unique-constrained fields to avoid
	// colliding with updateProduct's row which is still in the DB.
	newSKU := fmt.Sprintf("UpdatedSKU%d", idx)
	newUpcCode := fmt.Sprintf("UpdatedUpc%d", idx)
	newModelNumber := fmt.Sprintf("UpdatedModel%d", idx)

	expected := product
	expected.BrandID = updateProduct.BrandID
	expected.ProductCategoryID = updateProduct.ProductCategoryID
	expected.Name = updateProduct.Name
	expected.Description = updateProduct.Description
	expected.SKU = newSKU
	expected.UpcCode = newUpcCode
	expected.ModelNumber = newModelNumber
	expected.Status = updateProduct.Status
	expected.IsActive = updateProduct.IsActive
	expected.IsPerishable = updateProduct.IsPerishable
	expected.HandlingInstructions = updateProduct.HandlingInstructions
	expected.UnitsPerCase = updateProduct.UnitsPerCase

	return []unitest.Table{
		{
			Name:    "Update",
			ExpResp: expected,
			ExcFunc: func(ctx context.Context) any {
				up := productbus.UpdateProduct{
					BrandID:              &updateProduct.BrandID,
					ProductCategoryID:    &updateProduct.ProductCategoryID,
					Name:                 &updateProduct.Name,
					Description:          &updateProduct.Description,
					SKU:                  &newSKU,
					ModelNumber:          &newModelNumber,
					UpcCode:              &newUpcCode,
					Status:               &updateProduct.Status,
					IsActive:             &updateProduct.IsActive,
					IsPerishable:         &updateProduct.IsPerishable,
					HandlingInstructions: &updateProduct.HandlingInstructions,
					UnitsPerCase:         &updateProduct.UnitsPerCase,
				}
				p, err := busDomain.Product.Update(ctx, product, up)
				if err != nil {
					return err
				}

				return p
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(productbus.Product)
				if !exists {
					return "error occurred"
				}

				expResp, ok := exp.(productbus.Product)
				if !ok {
					return "expected product, got something else"
				}
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate
				expResp.ProductID = gotResp.ProductID

				return cmp.Diff(expResp, gotResp)
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
				err := busDomain.Product.Delete(ctx, sd.Products[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
