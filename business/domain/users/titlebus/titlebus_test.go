package titlebus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/timmaaaz/ichor/business/domain/users/titlebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_Title(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_Title")

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

	t, err := titlebus.TestSeedTitles(ctx, 10, busDomain.Title)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding fulfillment statues : %w", err)
	}

	return unitest.SeedData{
		Title: t,
	}, nil
}

func query(busdomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "query",
			ExpResp: []titlebus.Title{
				{ID: sd.Title[0].ID, Name: sd.Title[0].Name, Description: sd.Title[0].Description},
				{ID: sd.Title[1].ID, Name: sd.Title[1].Name, Description: sd.Title[1].Description},
				{ID: sd.Title[2].ID, Name: sd.Title[2].Name, Description: sd.Title[2].Description},
				{ID: sd.Title[3].ID, Name: sd.Title[3].Name, Description: sd.Title[3].Description},
				{ID: sd.Title[4].ID, Name: sd.Title[4].Name, Description: sd.Title[4].Description},
			},
			ExcFunc: func(ctx context.Context) any {
				fulfillmentstatuses, err := busdomain.Title.Query(ctx, titlebus.QueryFilter{}, order.NewBy(titlebus.OrderByName, order.ASC), page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return fulfillmentstatuses
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
			ExpResp: titlebus.Title{
				Description: sd.Title[0].Description,
				Name:        "Test Title",
			},
			ExcFunc: func(ctx context.Context) any {
				titles, err := busDomain.Title.Create(ctx, titlebus.NewTitle{
					Name:        "Test Title",
					Description: sd.Title[0].Description,
				})
				if err != nil {
					return err
				}
				return titles
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(titlebus.Title)
				if !exists {
					return fmt.Sprintf("got is not an approval status %v", got)
				}

				expResp := exp.(titlebus.Title)
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
			ExpResp: titlebus.Title{
				ID:          sd.Title[0].ID,
				Description: sd.Title[1].Description,
				Name:        "Updated Title",
			},
			ExcFunc: func(ctx context.Context) any {
				aprvlStatus, err := busDomain.Title.Update(ctx, sd.Title[0], titlebus.UpdateTitle{
					Name:        dbtest.StringPointer("Updated Title"),
					Description: &sd.Title[1].Description,
				})
				if err != nil {
					return err
				}
				return aprvlStatus
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(titlebus.Title)
				if !exists {
					return fmt.Sprintf("got is not an approval status %v", got)
				}
				expResp := exp.(titlebus.Title)
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
				err := busDomain.Title.Delete(ctx, sd.Title[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
