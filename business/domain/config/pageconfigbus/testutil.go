package pageconfigbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
)

// TestNewPageConfigs generates n NewPageConfig structs for testing.
// All configs are created as default configs (IsDefault: true, UserID: zero UUID)
// to satisfy the database constraint: default configs must have NULL user_id.
func TestNewPageConfigs(n int) []NewPageConfig {
	newConfigs := make([]NewPageConfig, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		npc := NewPageConfig{
			Name:      fmt.Sprintf("PageConfig%d", idx),
			IsDefault: true, // All test configs are default (user_id will be NULL/zero)
		}

		newConfigs[i] = npc
	}

	return newConfigs
}

// TestSeedPageConfigs seeds the database with n page configs and returns them sorted by name.
func TestSeedPageConfigs(ctx context.Context, n int, api *Business) ([]PageConfig, error) {
	newConfigs := TestNewPageConfigs(n)

	configs := make([]PageConfig, len(newConfigs))

	for i, npc := range newConfigs {
		config, err := api.Create(ctx, npc)
		if err != nil {
			return nil, fmt.Errorf("seeding page config: idx: %d : %w", i, err)
		}

		configs[i] = config
	}

	sort.Slice(configs, func(i, j int) bool {
		return configs[i].Name <= configs[j].Name
	})

	return configs, nil
}
