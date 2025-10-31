package formbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/timmaaaz/ichor/business/domain/config/formbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_Form(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_Form")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, queryByID(db.BusDomain, sd), "queryByID")
	unitest.Run(t, queryByName(db.BusDomain, sd), "queryByName")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")

}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	forms, err := formbus.TestSeedForms(ctx, 10, busDomain.Form)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding forms : %w", err)
	}

	return unitest.SeedData{
		Forms: forms,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "query",
			ExpResp: []formbus.Form{
				{ID: sd.Forms[0].ID, Name: sd.Forms[0].Name},
				{ID: sd.Forms[1].ID, Name: sd.Forms[1].Name},
				{ID: sd.Forms[2].ID, Name: sd.Forms[2].Name},
				{ID: sd.Forms[3].ID, Name: sd.Forms[3].Name},
				{ID: sd.Forms[4].ID, Name: sd.Forms[4].Name},
			},
			ExcFunc: func(ctx context.Context) any {
				forms, err := busDomain.Form.Query(ctx, formbus.QueryFilter{}, order.NewBy(formbus.OrderByName, order.ASC), page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return forms
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func queryByID(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "queryByID",
			ExpResp: sd.Forms[0],
			ExcFunc: func(ctx context.Context) any {
				form, err := busDomain.Form.QueryByID(ctx, sd.Forms[0].ID)
				if err != nil {
					return err
				}
				return form
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func queryByName(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "queryByName",
			ExpResp: sd.Forms[0],
			ExcFunc: func(ctx context.Context) any {
				form, err := busDomain.Form.QueryByName(ctx, sd.Forms[0].Name)
				if err != nil {
					return err
				}
				return form
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "create",
			ExpResp: formbus.Form{
				Name: "Test Form",
			},
			ExcFunc: func(ctx context.Context) any {
				form, err := busDomain.Form.Create(ctx, formbus.NewForm{
					Name: "Test Form",
				})
				if err != nil {
					return err
				}
				return form
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(formbus.Form)
				if !exists {
					return fmt.Sprintf("got is not a form %v", got)
				}

				expResp := exp.(formbus.Form)
				expResp.ID = gotResp.ID

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "update",
			ExpResp: formbus.Form{
				ID:   sd.Forms[0].ID,
				Name: "Updated Form",
			},
			ExcFunc: func(ctx context.Context) any {
				form, err := busDomain.Form.Update(ctx, sd.Forms[0], formbus.UpdateForm{
					Name: dbtest.StringPointer("Updated Form"),
				})
				if err != nil {
					return err
				}
				return form
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(formbus.Form)
				if !exists {
					return fmt.Sprintf("got is not a form %v", got)
				}
				expResp := exp.(formbus.Form)
				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.Form.Delete(ctx, sd.Forms[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
