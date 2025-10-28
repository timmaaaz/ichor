package pagebus

import (
	"context"
	"fmt"
)

// TestSeedPages creates N test pages for testing purposes.
func TestSeedPages(ctx context.Context, n int, api *Business) ([]Page, error) {
	pages := make([]Page, n)
	for i := 0; i < n; i++ {
		page, err := api.Create(ctx, NewPage{
			Path:      fmt.Sprintf("/test/path%d", i),
			Name:      fmt.Sprintf("Test Page %d", i),
			Module:    fmt.Sprintf("module%d", i),
			Icon:      "test-icon",
			SortOrder: i * 100,
			IsActive:  true,
		})
		if err != nil {
			return nil, fmt.Errorf("creating page : %w", err)
		}

		pages[i] = page
	}

	return pages, nil
}
