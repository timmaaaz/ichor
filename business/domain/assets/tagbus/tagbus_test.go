package tagbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/domain/assets/tagbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_Tags(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_Tags")

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

	tags, err := tagbus.TestSeedTag(ctx, 10, busDomain.Tag)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding approval statues : %w", err)
	}

	return unitest.SeedData{
		Tags: tags,
	}, nil
}

func query(busdomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "query",
			ExpResp: []tagbus.Tag{
				{ID: sd.Tags[0].ID, Name: sd.Tags[0].Name, Description: sd.Tags[0].Description},
				{ID: sd.Tags[1].ID, Name: sd.Tags[1].Name, Description: sd.Tags[1].Description},
				{ID: sd.Tags[2].ID, Name: sd.Tags[2].Name, Description: sd.Tags[2].Description},
				{ID: sd.Tags[3].ID, Name: sd.Tags[3].Name, Description: sd.Tags[3].Description},
				{ID: sd.Tags[4].ID, Name: sd.Tags[4].Name, Description: sd.Tags[4].Description},
			},
			ExcFunc: func(ctx context.Context) any {
				aprvlStatuses, err := busdomain.Tag.Query(ctx, tagbus.QueryFilter{}, order.NewBy(tagbus.OrderByName, order.ASC), page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return aprvlStatuses
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
			ExpResp: tagbus.Tag{
				Description: sd.Tags[0].Description,
				Name:        "Test Tag",
			},
			ExcFunc: func(ctx context.Context) any {
				aprvlStatus, err := busDomain.Tag.Create(ctx, tagbus.NewTag{
					Name:        "Test Tag",
					Description: sd.Tags[0].Description,
				})
				if err != nil {
					return err
				}
				return aprvlStatus
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(tagbus.Tag)
				if !exists {
					return fmt.Sprintf("got is not a tag %v", got)
				}

				expResp := exp.(tagbus.Tag)
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
			ExpResp: tagbus.Tag{
				ID:          sd.Tags[0].ID,
				Description: sd.Tags[1].Description,
				Name:        "Updated Tag",
			},
			ExcFunc: func(ctx context.Context) any {
				aprvlStatus, err := busDomain.Tag.Update(ctx, sd.Tags[0], tagbus.UpdateTag{
					Name:        dbtest.StringPointer("Updated Tag"),
					Description: &sd.Tags[1].Description,
				})
				if err != nil {
					return err
				}
				return aprvlStatus
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(tagbus.Tag)
				if !exists {
					return fmt.Sprintf("got is not a tag %v", got)
				}
				expResp := exp.(tagbus.Tag)
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
				err := busDomain.Tag.Delete(ctx, sd.Tags[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
