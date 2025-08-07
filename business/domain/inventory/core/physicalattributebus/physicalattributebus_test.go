package physicalattributebus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/physicalattributebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_PhysicalAttributes(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_PhysicalAttributes")

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

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 5, strIDs, busDomain.ContactInfos)
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

	physicalAttributes, err := physicalattributebus.TestSeedPhysicalAttributes(ctx, 20, productIDs, busDomain.PhysicalAttribute)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding physical attribute : %w", err)
	}

	return unitest.SeedData{
		Admins:             []unitest.User{{User: admins[0]}},
		PhysicalAttributes: physicalAttributes,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []physicalattributebus.PhysicalAttribute{
				sd.PhysicalAttributes[0],
				sd.PhysicalAttributes[1],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.PhysicalAttribute.Query(ctx, physicalattributebus.QueryFilter{}, physicalattributebus.DefaultOrderBy, page.MustParse("1", "2"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]physicalattributebus.PhysicalAttribute)
				if !exists {
					return "error occurred"
				}
				expResp := exp.([]physicalattributebus.PhysicalAttribute)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: physicalattributebus.PhysicalAttribute{
				ProductID:           sd.PhysicalAttributes[0].ProductID,
				Length:              physicalattributebus.NewDimension(34.2),
				Width:               physicalattributebus.NewDimension(20.1),
				Height:              physicalattributebus.NewDimension(15.7),
				Weight:              physicalattributebus.NewDimension(20),
				WeightUnit:          "lbs",
				Color:               "green",
				Size:                "xl",
				Material:            "gamma",
				StorageRequirements: "cold",
				HazmatClass:         "normal",
				ShelfLifeDays:       10,
			},
			ExcFunc: func(ctx context.Context) any {
				newPC := physicalattributebus.NewPhysicalAttribute{
					ProductID:           sd.PhysicalAttributes[0].ProductID,
					Length:              physicalattributebus.NewDimension(34.2),
					Width:               physicalattributebus.NewDimension(20.1),
					Height:              physicalattributebus.NewDimension(15.7),
					Weight:              physicalattributebus.NewDimension(20),
					WeightUnit:          "lbs",
					Color:               "green",
					Size:                "xl",
					Material:            "gamma",
					StorageRequirements: "cold",
					HazmatClass:         "normal",
					ShelfLifeDays:       10,
				}

				pc, err := busDomain.PhysicalAttribute.Create(ctx, newPC)
				if err != nil {
					return err
				}
				return pc
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(physicalattributebus.PhysicalAttribute)
				if !exists {
					return fmt.Sprintf("got is not a contact info: %v", got)
				}

				expResp := exp.(physicalattributebus.PhysicalAttribute)

				expResp.AttributeID = gotResp.AttributeID
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
			ExpResp: physicalattributebus.PhysicalAttribute{
				AttributeID:         sd.PhysicalAttributes[0].AttributeID,
				ProductID:           sd.PhysicalAttributes[1].ProductID,
				Length:              physicalattributebus.NewDimension(25),
				Width:               physicalattributebus.NewDimension(24),
				Height:              physicalattributebus.NewDimension(23),
				Weight:              physicalattributebus.NewDimension(22),
				WeightUnit:          "lbs",
				Color:               "blue",
				Size:                "xxl",
				Material:            "delta",
				StorageRequirements: "warm",
				HazmatClass:         "high",
				ShelfLifeDays:       7,
				CreatedDate:         sd.PhysicalAttributes[0].CreatedDate,
			},
			ExcFunc: func(ctx context.Context) any {
				upc := physicalattributebus.UpdatePhysicalAttribute{
					ProductID:           &sd.PhysicalAttributes[1].ProductID,
					Length:              physicalattributebus.NewDimension(25).ToPtr(),
					Width:               physicalattributebus.NewDimension(24).ToPtr(),
					Height:              physicalattributebus.NewDimension(23).ToPtr(),
					Weight:              physicalattributebus.NewDimension(22).ToPtr(),
					WeightUnit:          dbtest.StringPointer("lbs"),
					Color:               dbtest.StringPointer("blue"),
					Size:                dbtest.StringPointer("xxl"),
					Material:            dbtest.StringPointer("delta"),
					StorageRequirements: dbtest.StringPointer("warm"),
					HazmatClass:         dbtest.StringPointer("high"),
					ShelfLifeDays:       dbtest.IntPointer(7),
				}

				got, err := busDomain.PhysicalAttribute.Update(ctx, sd.PhysicalAttributes[0], upc)
				if err != nil {
					return err
				}

				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(physicalattributebus.PhysicalAttribute)
				if !exists {
					return "got is not a product category"
				}

				expResp := exp.(physicalattributebus.PhysicalAttribute)
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
				err := busDomain.PhysicalAttribute.Delete(ctx, sd.PhysicalAttributes[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
