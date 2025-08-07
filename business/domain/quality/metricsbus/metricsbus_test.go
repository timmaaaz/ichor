package metricsbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/quality/metricsbus"
	"github.com/timmaaaz/ichor/business/domain/quality/metricsbus/types"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_Metrics(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_Metrics")

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

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 5, busDomain.ContactInfos)
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

	metrics, err := metricsbus.TestSeedMetrics(ctx, 40, productIDs, busDomain.Metrics)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding metrics : %w", err)
	}

	return unitest.SeedData{
		Admins:            []unitest.User{{User: admins[0]}},
		ContactInfos:      contactInfos,
		Brands:            brand,
		Products:          products,
		ProductCategories: productCategories,
		Metrics:           metrics,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []metricsbus.Metric{
				sd.Metrics[0],
				sd.Metrics[1],
				sd.Metrics[2],
				sd.Metrics[3],
				sd.Metrics[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.Metrics.Query(ctx, metricsbus.QueryFilter{}, metricsbus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]metricsbus.Metric)
				if !exists {
					return fmt.Sprintf("got is not a slice of metrics: %v", got)
				}

				expResp := exp.([]metricsbus.Metric)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: metricsbus.Metric{
				ProductID:         sd.Products[3].ProductID,
				ReturnRate:        types.RoundedFloat{Value: 3.23},
				DefectRate:        types.RoundedFloat{Value: 7.32},
				MeasurementPeriod: types.MustParseInterval("3 Days"),
			},
			ExcFunc: func(ctx context.Context) any {
				newMetric := metricsbus.NewMetric{
					ProductID:         sd.Products[3].ProductID,
					ReturnRate:        types.RoundedFloat{Value: 3.23},
					DefectRate:        types.RoundedFloat{Value: 7.32},
					MeasurementPeriod: types.MustParseInterval("3 Days"),
				}

				s, err := busDomain.Metrics.Create(ctx, newMetric)
				if err != nil {
					return err
				}

				return s
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(metricsbus.Metric)
				if !exists {
					return fmt.Sprintf("got is not a product cost: %v", got)
				}

				expResp := exp.(metricsbus.Metric)

				expResp.MetricID = gotResp.MetricID
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
			ExpResp: metricsbus.Metric{
				MetricID:          sd.Metrics[1].MetricID,
				ProductID:         sd.Products[3].ProductID,
				ReturnRate:        types.RoundedFloat{Value: 3.23},
				DefectRate:        types.RoundedFloat{Value: 7.32},
				MeasurementPeriod: types.MustParseInterval("3 Days"),
				CreatedDate:       sd.Metrics[1].CreatedDate,
			},
			ExcFunc: func(ctx context.Context) any {
				newMetric := metricsbus.UpdateMetric{
					ProductID:         &sd.Products[3].ProductID,
					ReturnRate:        &types.RoundedFloat{Value: 3.23},
					DefectRate:        &types.RoundedFloat{Value: 7.32},
					MeasurementPeriod: types.MustParseInterval("3 Days").Ptr(),
				}

				s, err := busDomain.Metrics.Update(ctx, sd.Metrics[1], newMetric)
				if err != nil {
					return err
				}

				return s
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(metricsbus.Metric)
				if !exists {
					return fmt.Sprintf("got is not a product cost: %v", got)
				}

				expResp := exp.(metricsbus.Metric)

				expResp.MetricID = gotResp.MetricID
				expResp.CreatedDate = gotResp.CreatedDate
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
				err := busDomain.Metrics.Delete(ctx, sd.Metrics[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
