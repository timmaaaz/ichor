package productcategorybus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_ProductCategories(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_ProductCategories")

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

	categories, err := productcategorybus.TestSeedProductCategories(ctx, 15, busDomain.ProductCategory)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding product categories : %w", err)
	}

	return unitest.SeedData{
		Admins:            []unitest.User{{User: admins[0]}},
		ProductCategories: categories,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {

	return []unitest.Table{
		{
			Name:    "query",
			ExpResp: sd.ProductCategories[:5],
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.ProductCategory.Query(ctx, productcategorybus.QueryFilter{}, productcategorybus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}

				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]productcategorybus.ProductCategory)
				if !exists {
					return "error occurred"
				}

				expResp := exp.([]productcategorybus.ProductCategory)

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: productcategorybus.ProductCategory{
				Name:        "NewCategory",
				Description: "a new category",
			},
			ExcFunc: func(ctx context.Context) any {
				newPC := productcategorybus.NewProductCategory{
					Name:        "NewCategory",
					Description: "a new category",
				}

				pc, err := busDomain.ProductCategory.Create(ctx, newPC)
				if err != nil {
					return err
				}
				return pc
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(productcategorybus.ProductCategory)
				if !exists {
					return fmt.Sprintf("got is not a contact info: %v", got)
				}

				expResp := exp.(productcategorybus.ProductCategory)

				expResp.ProductCategoryID = gotResp.ProductCategoryID
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
			ExpResp: productcategorybus.ProductCategory{
				ProductCategoryID: sd.ProductCategories[0].ProductCategoryID,
				Name:              "Updated Category",
				Description:       "updated category description",
				CreatedDate:       sd.ProductCategories[0].CreatedDate,
			},
			ExcFunc: func(ctx context.Context) any {
				upc := productcategorybus.UpdateProductCategory{
					Name:        dbtest.StringPointer("Updated Category"),
					Description: dbtest.StringPointer("updated category description"),
				}

				got, err := busDomain.ProductCategory.Update(ctx, sd.ProductCategories[0], upc)
				if err != nil {
					return err
				}

				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(productcategorybus.ProductCategory)
				if !exists {
					return "got is not a product category"
				}

				expResp := exp.(productcategorybus.ProductCategory)
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
				err := busDomain.ProductCategory.Delete(ctx, sd.ProductCategories[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
