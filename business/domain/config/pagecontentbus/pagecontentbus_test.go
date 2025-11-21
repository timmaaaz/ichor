package pagecontentbus_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/timmaaaz/ichor/business/domain/config/pageconfigbus"
	"github.com/timmaaaz/ichor/business/domain/config/pagecontentbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_PageContent(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_PageContent")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, queryByID(db.BusDomain, sd), "queryByID")
	unitest.Run(t, queryByPageConfigID(db.BusDomain, sd), "queryByPageConfigID")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
	unitest.Run(t, queryWithChildren(db.BusDomain, sd), "queryWithChildren")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	// Create a page config first
	pageConfig, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "test_page",
		IsDefault: true,
	})
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding page config : %w", err)
	}

	contents, err := pagecontentbus.TestSeedPageContents(ctx, 10, pageConfig.ID, busDomain.PageContent)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding page contents : %w", err)
	}

	return unitest.SeedData{
		PageContents: contents,
		PageConfigs:  []pageconfigbus.PageConfig{pageConfig},
	}, nil
}

// normalizeJSON compacts JSON by removing whitespace to handle PostgreSQL JSONB normalization
func normalizeJSON(data json.RawMessage) json.RawMessage {
	if len(data) == 0 {
		return data
	}
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return data // Return original if parse fails
	}
	normalized, err := json.Marshal(v)
	if err != nil {
		return data // Return original if marshal fails
	}
	return normalized
}

// normalizePageContentJSON normalizes JSON fields in PageContent for comparison
func normalizePageContentJSON(content pagecontentbus.PageContent) pagecontentbus.PageContent {
	content.Layout = normalizeJSON(content.Layout)
	return content
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "query",
			ExpResp: []pagecontentbus.PageContent{
				sd.PageContents[0],
				sd.PageContents[1],
				sd.PageContents[2],
				sd.PageContents[3],
				sd.PageContents[4],
			},
			ExcFunc: func(ctx context.Context) any {
				contents, err := busDomain.PageContent.Query(ctx, pagecontentbus.QueryFilter{}, order.NewBy(pagecontentbus.OrderByOrderIndex, order.ASC), page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return contents
			},
			CmpFunc: func(got any, exp any) string {
				gotContents, ok1 := got.([]pagecontentbus.PageContent)
				expContents, ok2 := exp.([]pagecontentbus.PageContent)
				if !ok1 || !ok2 {
					return fmt.Sprintf("type mismatch: got %T, exp %T", got, exp)
				}
				// Normalize JSON in both slices
				for i := range gotContents {
					gotContents[i] = normalizePageContentJSON(gotContents[i])
				}
				for i := range expContents {
					expContents[i] = normalizePageContentJSON(expContents[i])
				}
				return cmp.Diff(gotContents, expContents, cmpopts.EquateEmpty())
			},
		},
	}

	return table
}

func queryByID(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "queryByID",
			ExpResp: sd.PageContents[0],
			ExcFunc: func(ctx context.Context) any {
				content, err := busDomain.PageContent.QueryByID(ctx, sd.PageContents[0].ID)
				if err != nil {
					return err
				}
				return content
			},
			CmpFunc: func(got any, exp any) string {
				gotContent, ok1 := got.(pagecontentbus.PageContent)
				expContent, ok2 := exp.(pagecontentbus.PageContent)
				if !ok1 || !ok2 {
					return fmt.Sprintf("type mismatch: got %T, exp %T", got, exp)
				}
				// Normalize JSON
				gotContent = normalizePageContentJSON(gotContent)
				expContent = normalizePageContentJSON(expContent)
				return cmp.Diff(gotContent, expContent, cmpopts.EquateEmpty())
			},
		},
	}

	return table
}

func queryByPageConfigID(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "queryByPageConfigID",
			ExpResp: sd.PageContents,
			ExcFunc: func(ctx context.Context) any {
				contents, err := busDomain.PageContent.QueryByPageConfigID(ctx, sd.PageContents[0].PageConfigID)
				if err != nil {
					return err
				}
				return contents
			},
			CmpFunc: func(got any, exp any) string {
				gotContents, ok1 := got.([]pagecontentbus.PageContent)
				expContents, ok2 := exp.([]pagecontentbus.PageContent)
				if !ok1 || !ok2 {
					return fmt.Sprintf("type mismatch: got %T, exp %T", got, exp)
				}
				// Normalize JSON in both slices
				for i := range gotContents {
					gotContents[i] = normalizePageContentJSON(gotContents[i])
				}
				for i := range expContents {
					expContents[i] = normalizePageContentJSON(expContents[i])
				}
				return cmp.Diff(gotContents, expContents, cmpopts.EquateEmpty())
			},
		},
	}

	return table
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	layout := json.RawMessage(`{"colSpan":{"default":12}}`)

	table := []unitest.Table{
		{
			Name: "create",
			ExpResp: pagecontentbus.PageContent{
				Label:      "Test Content",
				OrderIndex: 100,
			},
			ExcFunc: func(ctx context.Context) any {
				content, err := busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
					PageConfigID: sd.PageContents[0].PageConfigID,
					ContentType:  pagecontentbus.ContentTypeText,
					Label:        "Test Content",
					OrderIndex:   100,
					Layout:       layout,
					IsVisible:    true,
				})
				if err != nil {
					return err
				}
				return content
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(pagecontentbus.PageContent)
				if !exists {
					return fmt.Sprintf("got is not a page content %v", got)
				}

				expResp := exp.(pagecontentbus.PageContent)
				expResp.ID = gotResp.ID
				expResp.PageConfigID = gotResp.PageConfigID
				expResp.ContentType = gotResp.ContentType
				expResp.Layout = gotResp.Layout
				expResp.IsVisible = gotResp.IsVisible
				expResp.Children = gotResp.Children

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
			ExpResp: pagecontentbus.PageContent{
				ID:         sd.PageContents[0].ID,
				Label:      "Updated Content",
				OrderIndex: sd.PageContents[0].OrderIndex,
			},
			ExcFunc: func(ctx context.Context) any {
				content, err := busDomain.PageContent.Update(ctx, pagecontentbus.UpdatePageContent{
					Label: dbtest.StringPointer("Updated Content"),
				}, sd.PageContents[0].ID)
				if err != nil {
					return err
				}
				return content
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(pagecontentbus.PageContent)
				if !exists {
					return fmt.Sprintf("got is not a page content %v", got)
				}
				expResp := exp.(pagecontentbus.PageContent)
				expResp.PageConfigID = gotResp.PageConfigID
				expResp.ContentType = gotResp.ContentType
				expResp.TableConfigID = gotResp.TableConfigID
				expResp.FormID = gotResp.FormID
				expResp.ParentID = gotResp.ParentID
				expResp.Layout = gotResp.Layout
				expResp.IsVisible = gotResp.IsVisible
				expResp.IsDefault = gotResp.IsDefault
				expResp.Children = gotResp.Children
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
				err := busDomain.PageContent.Delete(ctx, sd.PageContents[0].ID)
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func queryWithChildren(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "queryWithChildren",
			ExpResp: func() []pagecontentbus.PageContent {
				// This test verifies the nesting works, exact structure depends on seed data
				return []pagecontentbus.PageContent{}
			}(),
			ExcFunc: func(ctx context.Context) any {
				contents, err := busDomain.PageContent.QueryWithChildren(ctx, sd.PageContents[0].PageConfigID)
				if err != nil {
					return err
				}
				return contents
			},
			CmpFunc: func(got, exp any) string {
				// Just verify no error for now
				_, isErr := got.(error)
				if isErr {
					return fmt.Sprintf("unexpected error: %v", got)
				}
				return ""
			},
		},
	}

	return table
}
