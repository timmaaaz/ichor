package inspectionbus_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfobus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/lot/lottrackingbus"
	"github.com/timmaaaz/ichor/business/domain/quality/inspectionbus"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierproductbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_Inspections(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_Inspections")

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

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 6, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	usrIDs := make([]uuid.UUID, 0, len(usrs))
	for _, u := range usrs {
		usrIDs = append(usrIDs, u.ID)
	}

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding user : %w", err)
	}

	contactInfo, err := contactinfobus.TestSeedContactInfo(ctx, 5, busDomain.ContactInfo)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding contact info : %w", err)
	}

	contactIDs := make(uuid.UUIDs, len(contactInfo))
	for i, c := range contactInfo {
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

	suppliers, err := supplierbus.TestSeedSuppliers(ctx, 25, contactIDs, busDomain.Supplier)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding suppliers : %w", err)
	}

	supplierIDs := make(uuid.UUIDs, len(suppliers))
	for i, s := range suppliers {
		supplierIDs[i] = s.SupplierID
	}

	supplierProducts, err := supplierproductbus.TestSeedSupplierProducts(ctx, 10, productIDs, supplierIDs, busDomain.SupplierProduct)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding supplier product : %w", err)
	}

	supplierProductIDs := make(uuid.UUIDs, len(supplierProducts))
	for i, sp := range supplierProducts {
		supplierProductIDs[i] = sp.SupplierProductID
	}

	lotTracking, err := lottrackingbus.TestSeedLotTracking(ctx, 15, supplierProductIDs, busDomain.LotTracking)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding lot tracking : %w", err)
	}

	lotTrackingIDs := make(uuid.UUIDs, len(lotTracking))
	for i, lt := range lotTracking {
		lotTrackingIDs[i] = lt.LotID
	}

	inspections, err := inspectionbus.TestSeedInspections(ctx, 10, productIDs, usrIDs, lotTrackingIDs, busDomain.Inspection)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding inspections : %w", err)
	}

	return unitest.SeedData{
		Products:    products,
		LotTracking: lotTracking,
		Users:       []unitest.User{{User: usrs[0]}, {User: usrs[1]}, {User: usrs[2]}, {User: usrs[3]}, {User: usrs[4]}, {User: usrs[5]}},
		Admins:      []unitest.User{{User: admins[0]}},
		Inspections: inspections,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []inspectionbus.Inspection{
				sd.Inspections[0],
				sd.Inspections[1],
				sd.Inspections[2],
				sd.Inspections[3],
				sd.Inspections[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.Inspection.Query(ctx, inspectionbus.QueryFilter{}, inspectionbus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]inspectionbus.Inspection)
				if !exists {
					return fmt.Sprintf("got is not a slice of metrics: %v", got)
				}

				expResp := exp.([]inspectionbus.Inspection)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {

	now := time.Now()
	later := now.AddDate(0, 0, 1)

	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: inspectionbus.Inspection{
				ProductID:          sd.Inspections[5].ProductID,
				InspectorID:        sd.Inspections[5].InspectorID,
				LotID:              sd.Inspections[5].LotID,
				Status:             "Pending",
				Notes:              "Initial inspection",
				InspectionDate:     now,
				NextInspectionDate: later,
			},
			ExcFunc: func(ctx context.Context) any {

				newInspection := inspectionbus.NewInspection{
					ProductID:          sd.Inspections[5].ProductID,
					InspectorID:        sd.Inspections[5].InspectorID,
					LotID:              sd.Inspections[5].LotID,
					Status:             "Pending",
					Notes:              "Initial inspection",
					InspectionDate:     now,
					NextInspectionDate: later,
				}

				got, err := busDomain.Inspection.Create(ctx, newInspection)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(inspectionbus.Inspection)
				if !exists {
					return fmt.Sprintf("got is not a inspection: %v", got)
				}

				expResp := exp.(inspectionbus.Inspection)
				expResp.InspectionID = gotResp.InspectionID
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
			ExpResp: inspectionbus.Inspection{
				InspectionID:       sd.Inspections[5].InspectionID,
				ProductID:          sd.Inspections[4].ProductID,
				InspectorID:        sd.Inspections[4].InspectorID,
				LotID:              sd.Inspections[4].LotID,
				Status:             "In Progress",
				Notes:              "Updated inspection",
				InspectionDate:     sd.Inspections[4].InspectionDate,
				NextInspectionDate: sd.Inspections[4].NextInspectionDate,
				CreatedDate:        sd.Inspections[5].CreatedDate,
			},
			ExcFunc: func(ctx context.Context) any {
				updatedInspection := inspectionbus.UpdateInspection{
					ProductID:          &sd.Inspections[4].ProductID,
					InspectorID:        &sd.Inspections[4].InspectorID,
					LotID:              &sd.Inspections[4].LotID,
					Status:             dbtest.StringPointer("In Progress"),
					Notes:              dbtest.StringPointer("Updated inspection"),
					InspectionDate:     &sd.Inspections[4].InspectionDate,
					NextInspectionDate: &sd.Inspections[4].NextInspectionDate,
				}
				got, err := busDomain.Inspection.Update(ctx, sd.Inspections[5], updatedInspection)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(inspectionbus.Inspection)
				if !exists {
					return fmt.Sprintf("got is not a inspection: %v", got)
				}

				expResp := exp.(inspectionbus.Inspection)
				expResp.InspectionDate = gotResp.InspectionDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "Delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.Inspection.Delete(ctx, sd.Inspections[5])
				if err != nil {
					return err
				}
				return nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
