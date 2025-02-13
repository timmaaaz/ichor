package titlebus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
)

func TestNewTitle(n int) []NewTitle {
	newTitles := make([]NewTitle, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		nt := NewTitle{
			Name:        "Title" + fmt.Sprintf("%d", idx),
			Description: "Title" + fmt.Sprintf("%d Description", idx),
		}

		newTitles[i] = nt
	}

	return newTitles
}

func TestSeedTitles(ctx context.Context, n int, api *Business) ([]Title, error) {
	newTitles := TestNewTitle(n)

	titles := make([]Title, len(newTitles))

	for i, nt := range newTitles {
		title, err := api.Create(ctx, nt)
		if err != nil {
			return nil, fmt.Errorf("seeding title: idx: %d : %w", i, err)
		}

		titles[i] = title
	}

	sort.Slice(titles, func(i, j int) bool {
		return titles[i].Name <= titles[j].Name
	})

	return titles, nil
}
