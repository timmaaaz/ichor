package supplierbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/paymenttermbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus/types"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_Suppliers(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_Suppliers")

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

	ContactInfosIDs := make(uuid.UUIDs, len(contactInfos))
	for i, c := range contactInfos {
		ContactInfosIDs[i] = c.ID
	}

	suppliers, err := supplierbus.TestSeedSuppliers(ctx, 25, ContactInfosIDs, busDomain.Supplier)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding suppliers : %w", err)
	}

	paymentTerms, err := paymenttermbus.TestSeedPaymentTerms(ctx, 2, busDomain.PaymentTerm)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding payment terms : %w", err)
	}

	return unitest.SeedData{
		Admins:       []unitest.User{{User: admins[0]}},
		ContactInfos: contactInfos,
		Suppliers:    suppliers,
		PaymentTerms: paymentTerms,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []supplierbus.Supplier{
				sd.Suppliers[0],
				sd.Suppliers[1],
				sd.Suppliers[2],
				sd.Suppliers[3],
				sd.Suppliers[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.Supplier.Query(ctx, supplierbus.QueryFilter{}, supplierbus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]supplierbus.Supplier)
				if !exists {
					return fmt.Sprintf("got is not a slice of product costs: %v", got)
				}

				expResp := exp.([]supplierbus.Supplier)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: supplierbus.Supplier{
				ContactInfosID: sd.ContactInfos[0].ID,
				Name:           "Name",
				PaymentTermID:  &sd.PaymentTerms[0].ID,
				LeadTimeDays:   8,
				Rating:         types.NewRoundedFloat(8.76),
				IsActive:       true,
			},
			ExcFunc: func(ctx context.Context) any {
				newSupplier := supplierbus.NewSupplier{
					ContactInfosID: sd.ContactInfos[0].ID,
					Name:           "Name",
					PaymentTermID:  &sd.PaymentTerms[0].ID,
					LeadTimeDays:   8,
					Rating:         types.NewRoundedFloat(8.76),
					IsActive:       true,
				}

				s, err := busDomain.Supplier.Create(ctx, newSupplier)
				if err != nil {
					return err
				}

				return s
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(supplierbus.Supplier)
				if !exists {
					return fmt.Sprintf("got is not a product cost: %v", got)
				}

				expResp := exp.(supplierbus.Supplier)

				expResp.SupplierID = gotResp.SupplierID
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
			ExpResp: supplierbus.Supplier{
				ContactInfosID: sd.ContactInfos[2].ID,
				SupplierID:     sd.Suppliers[0].SupplierID,
				Name:           "UpdatedName",
				PaymentTermID:  &sd.PaymentTerms[1].ID,
				LeadTimeDays:   10,
				Rating:         types.MustParseRoundedFloat("9.87"),
				IsActive:       false,
				CreatedDate:    sd.Suppliers[0].CreatedDate,
			},
			ExcFunc: func(ctx context.Context) any {
				updateSupplier := supplierbus.UpdateSupplier{
					ContactInfosID: &sd.ContactInfos[2].ID,
					Name:           dbtest.StringPointer("UpdatedName"),
					PaymentTermID:  &sd.PaymentTerms[1].ID,
					LeadTimeDays:   dbtest.IntPointer(10),
					Rating:         types.NewRoundedFloat(9.87).ToPtr(),
					IsActive:       dbtest.BoolPointer(false),
				}

				s, err := busDomain.Supplier.Update(ctx, sd.Suppliers[0], updateSupplier)
				if err != nil {
					return err
				}

				return s
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(supplierbus.Supplier)
				if !exists {
					return fmt.Sprintf("got is not a supplier: %v", got)
				}

				expResp := exp.(supplierbus.Supplier)

				expResp.SupplierID = gotResp.SupplierID
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
				err := busDomain.Supplier.Delete(ctx, sd.Suppliers[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
