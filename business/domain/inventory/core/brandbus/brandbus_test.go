package brandbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfobus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_Brand(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_Brand")

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

	contactInfo, err := contactinfobus.TestSeedContactInfo(ctx, 15, busDomain.ContactInfo)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding contact info : %w", err)
	}

	contactIDs := make(uuid.UUIDs, len(contactInfo))
	for i, c := range contactInfo {
		contactIDs[i] = c.ID
	}

	brands, err := brandbus.TestSeedBrands(ctx, 25, contactIDs, busDomain.Brand)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding brands : %w", err)
	}

	return unitest.SeedData{
		Admins:      []unitest.User{{User: admins[0]}},
		ContactInfo: contactInfo,
		Brands:      brands,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []brandbus.Brand{
				sd.Brands[0],
				sd.Brands[1],
				sd.Brands[2],
				sd.Brands[3],
				sd.Brands[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.Brand.Query(ctx, brandbus.QueryFilter{}, order.NewBy(brandbus.OrderByName, order.ASC), page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]brandbus.Brand)
				if !exists {
					return fmt.Sprintf("got is not a slice of brands: %v", got)
				}

				expResp := exp.([]brandbus.Brand)

				return cmp.Diff(gotResp, expResp)

			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: brandbus.Brand{
				Name:          "NewBrand",
				ContactInfoID: sd.ContactInfo[0].ID,
			},
			ExcFunc: func(ctx context.Context) any {
				newBrand := brandbus.NewBrand{
					Name:          "NewBrand",
					ContactInfoID: sd.ContactInfo[0].ID,
				}

				ci, err := busDomain.Brand.Create(ctx, newBrand)
				if err != nil {
					return err
				}

				return ci
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(brandbus.Brand)
				if !exists {
					return fmt.Sprintf("got is not a brand: %v", got)
				}

				expResp := exp.(brandbus.Brand)

				expResp.BrandID = gotResp.BrandID
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
			ExpResp: brandbus.Brand{
				BrandID:       sd.Brands[0].BrandID,
				Name:          "UpdatedBrand",
				ContactInfoID: sd.ContactInfo[0].ID,
				CreatedDate:   sd.Brands[0].CreatedDate,
			},
			ExcFunc: func(ctx context.Context) any {
				uc := brandbus.UpdateBrand{
					ContactInfoID: &sd.ContactInfo[0].ID,
					Name:          dbtest.StringPointer("UpdatedBrand"),
				}

				got, err := busDomain.Brand.Update(ctx, sd.Brands[0], uc)
				if err != nil {
					return err
				}

				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(brandbus.Brand)
				if !exists {
					return fmt.Sprintf("got is not a brand: %v", got)
				}

				expResp := exp.(brandbus.Brand)

				expResp.BrandID = gotResp.BrandID
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
				err := busDomain.Brand.Delete(ctx, sd.Brands[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
