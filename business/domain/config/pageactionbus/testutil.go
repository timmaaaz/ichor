package pageactionbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

// TestNewButtonActions generates n NewButtonAction structs for testing.
func TestNewButtonActions(n int, pageConfigIDs []uuid.UUID) []NewButtonAction {
	actions := make([]NewButtonAction, n)

	idx := rand.Intn(10000)
	variants := []string{"default", "secondary", "outline", "ghost", "destructive"}
	alignments := []string{"left", "right"}

	for i := 0; i < n; i++ {
		idx++

		nba := NewButtonAction{
			PageConfigID:       pageConfigIDs[rand.Intn(len(pageConfigIDs))],
			ActionOrder:        i,
			IsActive:           rand.Intn(2) == 0,
			Label:              fmt.Sprintf("Button %d", idx),
			Icon:               fmt.Sprintf("icon-%d", idx),
			TargetPath:         fmt.Sprintf("/path/%d", idx),
			Variant:            variants[rand.Intn(len(variants))],
			Alignment:          alignments[rand.Intn(len(alignments))],
			ConfirmationPrompt: fmt.Sprintf("Confirm action %d?", idx),
		}

		actions[i] = nba
	}

	return actions
}

// TestNewDropdownActions generates n NewDropdownAction structs for testing.
func TestNewDropdownActions(n int, pageConfigIDs []uuid.UUID) []NewDropdownAction {
	actions := make([]NewDropdownAction, n)

	idx := rand.Intn(10000)

	for i := 0; i < n; i++ {
		idx++

		// Generate 2-5 items per dropdown
		numItems := 2 + rand.Intn(4)
		items := make([]NewDropdownItem, numItems)
		for j := 0; j < numItems; j++ {
			items[j] = NewDropdownItem{
				Label:      fmt.Sprintf("Item %d-%d", idx, j),
				TargetPath: fmt.Sprintf("/path/%d/%d", idx, j),
				ItemOrder:  j,
			}
		}

		nda := NewDropdownAction{
			PageConfigID: pageConfigIDs[rand.Intn(len(pageConfigIDs))],
			ActionOrder:  i,
			IsActive:     rand.Intn(2) == 0,
			Label:        fmt.Sprintf("Dropdown %d", idx),
			Icon:         fmt.Sprintf("icon-%d", idx),
			Items:        items,
		}

		actions[i] = nda
	}

	return actions
}

// TestNewSeparatorActions generates n NewSeparatorAction structs for testing.
func TestNewSeparatorActions(n int, pageConfigIDs []uuid.UUID) []NewSeparatorAction {
	actions := make([]NewSeparatorAction, n)

	for i := 0; i < n; i++ {
		nsa := NewSeparatorAction{
			PageConfigID: pageConfigIDs[rand.Intn(len(pageConfigIDs))],
			ActionOrder:  i,
			IsActive:     rand.Intn(2) == 0,
		}

		actions[i] = nsa
	}

	return actions
}

// TestSeedPageActions seeds the database with a mix of page actions and returns them sorted by ID.
func TestSeedPageActions(ctx context.Context, n int, pageConfigIDs []uuid.UUID, api *Business) ([]PageAction, error) {
	if len(pageConfigIDs) == 0 {
		return nil, fmt.Errorf("pageConfigIDs cannot be empty")
	}

	// Split n into buttons, dropdowns, and separators
	numButtons := n / 3
	numDropdowns := n / 3
	numSeparators := n - numButtons - numDropdowns

	var allActions []PageAction

	// Create buttons
	newButtons := TestNewButtonActions(numButtons, pageConfigIDs)
	for i, nb := range newButtons {
		action, err := api.CreateButton(ctx, nb)
		if err != nil {
			return nil, fmt.Errorf("seeding button: idx: %d : %w", i, err)
		}
		allActions = append(allActions, action)
	}

	// Create dropdowns
	newDropdowns := TestNewDropdownActions(numDropdowns, pageConfigIDs)
	for i, nd := range newDropdowns {
		action, err := api.CreateDropdown(ctx, nd)
		if err != nil {
			return nil, fmt.Errorf("seeding dropdown: idx: %d : %w", i, err)
		}
		allActions = append(allActions, action)
	}

	// Create separators
	newSeparators := TestNewSeparatorActions(numSeparators, pageConfigIDs)
	for i, ns := range newSeparators {
		action, err := api.CreateSeparator(ctx, ns)
		if err != nil {
			return nil, fmt.Errorf("seeding separator: idx: %d : %w", i, err)
		}
		allActions = append(allActions, action)
	}

	// Sort by ID for consistent test ordering
	sort.Slice(allActions, func(i, j int) bool {
		return allActions[i].ID.String() < allActions[j].ID.String()
	})

	return allActions, nil
}
