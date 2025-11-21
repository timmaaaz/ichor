package pagecontentbus

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

// TestNewPageContents generates n NewPageContent structs for testing.
func TestNewPageContents(n int, pageConfigID uuid.UUID) []NewPageContent {
	newContents := make([]NewPageContent, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		// Create a simple layout (compact format to match PostgreSQL JSONB normalization)
		layout := json.RawMessage(`{"colSpan":{"default":12}}`)

		npc := NewPageContent{
			PageConfigID: pageConfigID,
			ContentType:  ContentTypeText,
			Label:        fmt.Sprintf("Content%d", idx),
			OrderIndex:   i + 1,
			Layout:       layout,
			IsVisible:    true,
			IsDefault:    i == 0,
		}

		newContents[i] = npc
	}

	return newContents
}

// TestSeedPageContents seeds the database with n page content blocks and returns them sorted by order_index.
func TestSeedPageContents(ctx context.Context, n int, pageConfigID uuid.UUID, api *Business) ([]PageContent, error) {
	newContents := TestNewPageContents(n, pageConfigID)

	contents := make([]PageContent, len(newContents))

	for i, npc := range newContents {
		content, err := api.Create(ctx, npc)
		if err != nil {
			return nil, fmt.Errorf("seeding page content: idx: %d : %w", i, err)
		}

		contents[i] = content
	}

	sort.Slice(contents, func(i, j int) bool {
		return contents[i].OrderIndex <= contents[j].OrderIndex
	})

	return contents, nil
}

// TestNewPageContentWithChildren generates page content with parent-child relationships for testing tabs.
func TestNewPageContentWithChildren(pageConfigID uuid.UUID) []NewPageContent {
	layout := json.RawMessage(`{"colSpan":{"default":12}}`)
	tabLayout := json.RawMessage(`{"colSpan":{"default":12},"containerType":"tabs"}`)

	return []NewPageContent{
		{
			PageConfigID: pageConfigID,
			ContentType:  ContentTypeText,
			Label:        "Header Text",
			OrderIndex:   1,
			Layout:       layout,
			IsVisible:    true,
			IsDefault:    false,
		},
		{
			PageConfigID: pageConfigID,
			ContentType:  ContentTypeTabs,
			Label:        "Tabs Container",
			OrderIndex:   2,
			Layout:       tabLayout,
			IsVisible:    true,
			IsDefault:    false,
		},
	}
}
