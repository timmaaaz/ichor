package tagbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
)

func TestNewTag(n int) []NewTag {
	newTag := make([]NewTag, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		nt := NewTag{
			Name:        fmt.Sprintf("Tag%d", idx),
			Description: fmt.Sprintf("Tag%d Description", idx),
		}

		newTag[i] = nt
	}

	return newTag
}

func TestSeedTag(ctx context.Context, n int, api *Business) ([]Tag, error) {
	newTags := TestNewTag(n)

	tags := make([]Tag, len(newTags))
	for i, nt := range newTags {
		tag, err := api.Create(ctx, nt)
		if err != nil {
			return nil, fmt.Errorf("seeding tag: idx: %d : %w", i, err)
		}

		tags[i] = tag
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Name <= tags[j].Name
	})

	return tags, nil

}
