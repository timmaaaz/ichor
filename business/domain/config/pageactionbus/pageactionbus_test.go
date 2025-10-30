package pageactionbus_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_PageAction(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_PageAction")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, queryByID(db.BusDomain, sd), "queryByID")
	unitest.Run(t, queryByPageConfigID(db.BusDomain, sd), "queryByPageConfigID")
	unitest.Run(t, createButton(db.BusDomain, sd), "createButton")
	unitest.Run(t, createDropdown(db.BusDomain, sd), "createDropdown")
	unitest.Run(t, createSeparator(db.BusDomain, sd), "createSeparator")
	unitest.Run(t, updateButton(db.BusDomain, sd), "updateButton")
	unitest.Run(t, updateDropdown(db.BusDomain, sd), "updateDropdown")
	unitest.Run(t, updateSeparator(db.BusDomain, sd), "updateSeparator")
	unitest.Run(t, deleteAction(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	// Create test page configs using the ConfigStore
	pageConfigIDs := make([]uuid.UUID, 3)
	for i := 0; i < 3; i++ {
		id := uuid.New()
		pc, err := busDomain.ConfigStore.CreatePageConfig(ctx, tablebuilder.PageConfig{
			ID:        id,
			Name:      fmt.Sprintf("test-page-config-%d", i),
			UserID:    uuid.Nil, // NULL for default configs
			IsDefault: true,
		})
		if err != nil {
			return unitest.SeedData{}, fmt.Errorf("creating page config: %w", err)
		}
		pageConfigIDs[i] = pc.ID
	}

	// Seed page actions
	actions, err := pageactionbus.TestSeedPageActions(ctx, 15, pageConfigIDs, busDomain.PageAction)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding page actions: %w", err)
	}

	return unitest.SeedData{
		PageActions:   actions,
		PageConfigIDs: pageConfigIDs,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	// Query returns base actions only (no Button/Dropdown data)
	// Create expected responses with only base fields
	expected := make([]pageactionbus.PageAction, 5)
	for i := 0; i < 5; i++ {
		expected[i] = pageactionbus.PageAction{
			ID:           sd.PageActions[i].ID,
			PageConfigID: sd.PageActions[i].PageConfigID,
			ActionType:   sd.PageActions[i].ActionType,
			ActionOrder:  sd.PageActions[i].ActionOrder,
			IsActive:     sd.PageActions[i].IsActive,
			Button:       nil, // Query does not populate detail fields
			Dropdown:     nil, // Query does not populate detail fields
		}
	}

	table := []unitest.Table{
		{
			Name:    "query-all",
			ExpResp: expected,
			ExcFunc: func(ctx context.Context) any {
				actions, err := busDomain.PageAction.Query(ctx, pageactionbus.QueryFilter{}, order.NewBy(pageactionbus.OrderByID, order.ASC), page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return actions
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func queryByID(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	// Find one of each type for testing
	var buttonAction, dropdownAction, separatorAction pageactionbus.PageAction
	for _, action := range sd.PageActions {
		switch action.ActionType {
		case pageactionbus.ActionTypeButton:
			if buttonAction.ID == uuid.Nil {
				buttonAction = action
			}
		case pageactionbus.ActionTypeDropdown:
			if dropdownAction.ID == uuid.Nil {
				dropdownAction = action
			}
		case pageactionbus.ActionTypeSeparator:
			if separatorAction.ID == uuid.Nil {
				separatorAction = action
			}
		}
	}

	table := []unitest.Table{
		{
			Name:    "queryByID-button",
			ExpResp: buttonAction,
			ExcFunc: func(ctx context.Context) any {
				action, err := busDomain.PageAction.QueryByID(ctx, buttonAction.ID)
				if err != nil {
					return err
				}
				return action
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "queryByID-dropdown",
			ExpResp: dropdownAction,
			ExcFunc: func(ctx context.Context) any {
				action, err := busDomain.PageAction.QueryByID(ctx, dropdownAction.ID)
				if err != nil {
					return err
				}
				return action
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "queryByID-separator",
			ExpResp: separatorAction,
			ExcFunc: func(ctx context.Context) any {
				action, err := busDomain.PageAction.QueryByID(ctx, separatorAction.ID)
				if err != nil {
					return err
				}
				return action
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func queryByPageConfigID(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	// Find actions for the first page config
	expectedButtons := []pageactionbus.PageAction(nil)
	expectedDropdowns := []pageactionbus.PageAction(nil)
	expectedSeparators := []pageactionbus.PageAction(nil)

	for _, action := range sd.PageActions {
		if action.PageConfigID == sd.PageConfigIDs[0] {
			switch action.ActionType {
			case pageactionbus.ActionTypeButton:
				expectedButtons = append(expectedButtons, action)
			case pageactionbus.ActionTypeDropdown:
				expectedDropdowns = append(expectedDropdowns, action)
			case pageactionbus.ActionTypeSeparator:
				expectedSeparators = append(expectedSeparators, action)
			}
		}
	}

	// Sort each slice by action_order ASC, then by ID ASC (matching database query order)
	sortActions := func(actions []pageactionbus.PageAction) {
		if len(actions) > 0 {
			sort.Slice(actions, func(i, j int) bool {
				if actions[i].ActionOrder != actions[j].ActionOrder {
					return actions[i].ActionOrder < actions[j].ActionOrder
				}
				return actions[i].ID.String() < actions[j].ID.String()
			})
		}
	}
	sortActions(expectedButtons)
	sortActions(expectedDropdowns)
	sortActions(expectedSeparators)

	expected := pageactionbus.ActionsGroupedByType{
		Buttons:    expectedButtons,
		Dropdowns:  expectedDropdowns,
		Separators: expectedSeparators,
	}

	table := []unitest.Table{
		{
			Name:    "queryByPageConfigID",
			ExpResp: expected,
			ExcFunc: func(ctx context.Context) any {
				actions, err := busDomain.PageAction.QueryByPageConfigID(ctx, sd.PageConfigIDs[0])
				if err != nil {
					return err
				}
				return actions
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func createButton(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "create-button",
			ExpResp: pageactionbus.PageAction{
				PageConfigID: sd.PageConfigIDs[0],
				ActionType:   pageactionbus.ActionTypeButton,
				ActionOrder:  999,
				IsActive:     true,
				Button: &pageactionbus.ButtonAction{
					Label:              "Test Button",
					Icon:               "test-icon",
					TargetPath:         "/test/path",
					Variant:            "default",
					Alignment:          "right",
					ConfirmationPrompt: "Are you sure?",
				},
			},
			ExcFunc: func(ctx context.Context) any {
				action, err := busDomain.PageAction.CreateButton(ctx, pageactionbus.NewButtonAction{
					PageConfigID:       sd.PageConfigIDs[0],
					ActionOrder:        999,
					IsActive:           true,
					Label:              "Test Button",
					Icon:               "test-icon",
					TargetPath:         "/test/path",
					Variant:            "default",
					Alignment:          "right",
					ConfirmationPrompt: "Are you sure?",
				})
				if err != nil {
					return err
				}
				return action
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(pageactionbus.PageAction)
				if !exists {
					return fmt.Sprintf("got is not a page action: %v", got)
				}

				expResp := exp.(pageactionbus.PageAction)
				expResp.ID = gotResp.ID

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func createDropdown(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "create-dropdown",
			ExpResp: pageactionbus.PageAction{
				PageConfigID: sd.PageConfigIDs[0],
				ActionType:   pageactionbus.ActionTypeDropdown,
				ActionOrder:  999,
				IsActive:     true,
				Dropdown: &pageactionbus.DropdownAction{
					Label: "Test Dropdown",
					Icon:  "dropdown-icon",
					Items: []pageactionbus.DropdownItem{
						{
							Label:      "Item 1",
							TargetPath: "/item/1",
							ItemOrder:  0,
						},
						{
							Label:      "Item 2",
							TargetPath: "/item/2",
							ItemOrder:  1,
						},
					},
				},
			},
			ExcFunc: func(ctx context.Context) any {
				action, err := busDomain.PageAction.CreateDropdown(ctx, pageactionbus.NewDropdownAction{
					PageConfigID: sd.PageConfigIDs[0],
					ActionOrder:  999,
					IsActive:     true,
					Label:        "Test Dropdown",
					Icon:         "dropdown-icon",
					Items: []pageactionbus.NewDropdownItem{
						{
							Label:      "Item 1",
							TargetPath: "/item/1",
							ItemOrder:  0,
						},
						{
							Label:      "Item 2",
							TargetPath: "/item/2",
							ItemOrder:  1,
						},
					},
				})
				if err != nil {
					return err
				}
				return action
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(pageactionbus.PageAction)
				if !exists {
					return fmt.Sprintf("got is not a page action: %v", got)
				}

				expResp := exp.(pageactionbus.PageAction)
				expResp.ID = gotResp.ID

				// Match dropdown item IDs
				if gotResp.Dropdown != nil && expResp.Dropdown != nil {
					for i := range expResp.Dropdown.Items {
						if i < len(gotResp.Dropdown.Items) {
							expResp.Dropdown.Items[i].ID = gotResp.Dropdown.Items[i].ID
						}
					}
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func createSeparator(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "create-separator",
			ExpResp: pageactionbus.PageAction{
				PageConfigID: sd.PageConfigIDs[0],
				ActionType:   pageactionbus.ActionTypeSeparator,
				ActionOrder:  999,
				IsActive:     true,
			},
			ExcFunc: func(ctx context.Context) any {
				action, err := busDomain.PageAction.CreateSeparator(ctx, pageactionbus.NewSeparatorAction{
					PageConfigID: sd.PageConfigIDs[0],
					ActionOrder:  999,
					IsActive:     true,
				})
				if err != nil {
					return err
				}
				return action
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(pageactionbus.PageAction)
				if !exists {
					return fmt.Sprintf("got is not a page action: %v", got)
				}

				expResp := exp.(pageactionbus.PageAction)
				expResp.ID = gotResp.ID

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func updateButton(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	// Find a button action
	var buttonAction pageactionbus.PageAction
	for _, action := range sd.PageActions {
		if action.ActionType == pageactionbus.ActionTypeButton {
			buttonAction = action
			break
		}
	}

	newLabel := "Updated Button"
	newOrder := 500

	table := []unitest.Table{
		{
			Name: "update-button",
			ExpResp: pageactionbus.PageAction{
				ID:           buttonAction.ID,
				PageConfigID: buttonAction.PageConfigID,
				ActionType:   pageactionbus.ActionTypeButton,
				ActionOrder:  newOrder,
				IsActive:     buttonAction.IsActive,
				Button: &pageactionbus.ButtonAction{
					Label:              newLabel,
					Icon:               buttonAction.Button.Icon,
					TargetPath:         buttonAction.Button.TargetPath,
					Variant:            buttonAction.Button.Variant,
					Alignment:          buttonAction.Button.Alignment,
					ConfirmationPrompt: buttonAction.Button.ConfirmationPrompt,
				},
			},
			ExcFunc: func(ctx context.Context) any {
				action, err := busDomain.PageAction.UpdateButton(ctx, buttonAction, pageactionbus.UpdateButtonAction{
					Label:       &newLabel,
					ActionOrder: &newOrder,
				})
				if err != nil {
					return err
				}
				return action
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func updateDropdown(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	// Find a dropdown action
	var dropdownAction pageactionbus.PageAction
	for _, action := range sd.PageActions {
		if action.ActionType == pageactionbus.ActionTypeDropdown {
			dropdownAction = action
			break
		}
	}

	newLabel := "Updated Dropdown"
	newItems := []pageactionbus.NewDropdownItem{
		{
			Label:      "New Item 1",
			TargetPath: "/new/1",
			ItemOrder:  0,
		},
	}

	table := []unitest.Table{
		{
			Name: "update-dropdown",
			ExpResp: pageactionbus.PageAction{
				ID:           dropdownAction.ID,
				PageConfigID: dropdownAction.PageConfigID,
				ActionType:   pageactionbus.ActionTypeDropdown,
				ActionOrder:  dropdownAction.ActionOrder,
				IsActive:     dropdownAction.IsActive,
				Dropdown: &pageactionbus.DropdownAction{
					Label: newLabel,
					Icon:  dropdownAction.Dropdown.Icon,
					Items: []pageactionbus.DropdownItem{
						{
							Label:      "New Item 1",
							TargetPath: "/new/1",
							ItemOrder:  0,
						},
					},
				},
			},
			ExcFunc: func(ctx context.Context) any {
				action, err := busDomain.PageAction.UpdateDropdown(ctx, dropdownAction, pageactionbus.UpdateDropdownAction{
					Label: &newLabel,
					Items: &newItems,
				})
				if err != nil {
					return err
				}
				return action
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(pageactionbus.PageAction)
				if !exists {
					return fmt.Sprintf("got is not a page action: %v", got)
				}

				expResp := exp.(pageactionbus.PageAction)

				// Match dropdown item IDs
				if gotResp.Dropdown != nil && expResp.Dropdown != nil {
					for i := range expResp.Dropdown.Items {
						if i < len(gotResp.Dropdown.Items) {
							expResp.Dropdown.Items[i].ID = gotResp.Dropdown.Items[i].ID
						}
					}
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func updateSeparator(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	// Find a separator action
	var separatorAction pageactionbus.PageAction
	for _, action := range sd.PageActions {
		if action.ActionType == pageactionbus.ActionTypeSeparator {
			separatorAction = action
			break
		}
	}

	newOrder := 100

	table := []unitest.Table{
		{
			Name: "update-separator",
			ExpResp: pageactionbus.PageAction{
				ID:           separatorAction.ID,
				PageConfigID: separatorAction.PageConfigID,
				ActionType:   pageactionbus.ActionTypeSeparator,
				ActionOrder:  newOrder,
				IsActive:     separatorAction.IsActive,
			},
			ExcFunc: func(ctx context.Context) any {
				action, err := busDomain.PageAction.UpdateSeparator(ctx, separatorAction, pageactionbus.UpdateSeparatorAction{
					ActionOrder: &newOrder,
				})
				if err != nil {
					return err
				}
				return action
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func deleteAction(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.PageAction.Delete(ctx, sd.PageActions[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
