package pageaction_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/config/pageactionapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

// =============================================================================
// Button Update Tests
// =============================================================================

func updateButton200(sd apitest.SeedData) []apitest.Table {
	// Find a button action in seed data
	var buttonAction *pageactionapp.PageAction
	for _, action := range sd.PageActions {
		if action.ActionType == "button" {
			buttonAction = &action
			break
		}
	}

	if buttonAction == nil {
		return []apitest.Table{}
	}

	newLabel := "Updated Button Label"
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/page-actions/buttons/" + buttonAction.ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &pageactionapp.UpdateButtonAction{
				Label: &newLabel,
			},
			GotResp: &pageactionapp.PageAction{},
			ExpResp: &pageactionapp.PageAction{
				ID:           buttonAction.ID,
				PageConfigID: buttonAction.PageConfigID,
				ActionType:   "button",
				ActionOrder:  buttonAction.ActionOrder,
				IsActive:     buttonAction.IsActive,
				Button: &pageactionapp.ButtonAction{
					Label:              newLabel,
					Icon:               buttonAction.Button.Icon,
					TargetPath:         buttonAction.Button.TargetPath,
					Variant:            buttonAction.Button.Variant,
					Alignment:          buttonAction.Button.Alignment,
					ConfirmationPrompt: buttonAction.Button.ConfirmationPrompt,
				},
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func updateButton400(sd apitest.SeedData) []apitest.Table {
	var buttonAction *pageactionapp.PageAction
	for _, action := range sd.PageActions {
		if action.ActionType == "button" {
			buttonAction = &action
			break
		}
	}

	if buttonAction == nil {
		return []apitest.Table{}
	}

	invalidVariant := "invalid_variant"
	return []apitest.Table{
		{
			Name:       "invalid_variant",
			URL:        "/v1/config/page-actions/buttons/" + buttonAction.ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &pageactionapp.UpdateButtonAction{
				Variant: &invalidVariant,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate"),
			CmpFunc: func(got, exp any) string {
				gotErr := got.(*errs.Error)
				if gotErr.Code != errs.InvalidArgument {
					return "expected InvalidArgument error"
				}
				return ""
			},
		},
	}
}

func updateButton401(sd apitest.SeedData) []apitest.Table {
	var buttonAction *pageactionapp.PageAction
	for _, action := range sd.PageActions {
		if action.ActionType == "button" {
			buttonAction = &action
			break
		}
	}

	if buttonAction == nil {
		return []apitest.Table{}
	}

	newLabel := "Updated Label"
	return []apitest.Table{
		{
			Name:       "unauthorized",
			URL:        "/v1/config/page-actions/buttons/" + buttonAction.ID,
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input: &pageactionapp.UpdateButtonAction{
				Label: &newLabel,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[USER]] rule[rule_admin_only]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

// =============================================================================
// Dropdown Update Tests
// =============================================================================

func updateDropdown200(sd apitest.SeedData) []apitest.Table {
	var dropdownAction *pageactionapp.PageAction
	for _, action := range sd.PageActions {
		if action.ActionType == "dropdown" {
			dropdownAction = &action
			break
		}
	}

	if dropdownAction == nil {
		return []apitest.Table{}
	}

	newLabel := "Updated Dropdown Label"
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/page-actions/dropdowns/" + dropdownAction.ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &pageactionapp.UpdateDropdownAction{
				Label: &newLabel,
			},
			GotResp: &pageactionapp.PageAction{},
			ExpResp: &pageactionapp.PageAction{},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*pageactionapp.PageAction)
				if gotResp.Dropdown.Label != newLabel {
					return "label not updated"
				}
				return ""
			},
		},
	}
}

func updateDropdown400(sd apitest.SeedData) []apitest.Table {
	var dropdownAction *pageactionapp.PageAction
	for _, action := range sd.PageActions {
		if action.ActionType == "dropdown" {
			dropdownAction = &action
			break
		}
	}

	if dropdownAction == nil {
		return []apitest.Table{}
	}

	emptyItems := []pageactionapp.NewDropdownItem{}
	return []apitest.Table{
		{
			Name:       "empty_items",
			URL:        "/v1/config/page-actions/dropdowns/" + dropdownAction.ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &pageactionapp.UpdateDropdownAction{
				Items: &emptyItems,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate"),
			CmpFunc: func(got, exp any) string {
				gotErr := got.(*errs.Error)
				if gotErr.Code != errs.InvalidArgument {
					return "expected InvalidArgument error"
				}
				return ""
			},
		},
	}
}

func updateDropdown401(sd apitest.SeedData) []apitest.Table {
	var dropdownAction *pageactionapp.PageAction
	for _, action := range sd.PageActions {
		if action.ActionType == "dropdown" {
			dropdownAction = &action
			break
		}
	}

	if dropdownAction == nil {
		return []apitest.Table{}
	}

	newLabel := "Updated Label"
	return []apitest.Table{
		{
			Name:       "unauthorized",
			URL:        "/v1/config/page-actions/dropdowns/" + dropdownAction.ID,
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input: &pageactionapp.UpdateDropdownAction{
				Label: &newLabel,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[USER]] rule[rule_admin_only]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

// =============================================================================
// Separator Update Tests
// =============================================================================

func updateSeparator200(sd apitest.SeedData) []apitest.Table {
	var separatorAction *pageactionapp.PageAction
	for _, action := range sd.PageActions {
		if action.ActionType == "separator" {
			separatorAction = &action
			break
		}
	}

	if separatorAction == nil {
		return []apitest.Table{}
	}

	newOrder := 999
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/page-actions/separators/" + separatorAction.ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &pageactionapp.UpdateSeparatorAction{
				ActionOrder: &newOrder,
			},
			GotResp: &pageactionapp.PageAction{},
			ExpResp: &pageactionapp.PageAction{},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*pageactionapp.PageAction)
				if gotResp.ActionOrder != newOrder {
					return "action order not updated"
				}
				return ""
			},
		},
	}
}

func updateSeparator400(sd apitest.SeedData) []apitest.Table {
	var separatorAction *pageactionapp.PageAction
	for _, action := range sd.PageActions {
		if action.ActionType == "separator" {
			separatorAction = &action
			break
		}
	}

	if separatorAction == nil {
		return []apitest.Table{}
	}

	invalidUUID := "not-a-uuid"
	return []apitest.Table{
		{
			Name:       "invalid_page_config_id",
			URL:        "/v1/config/page-actions/separators/" + separatorAction.ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &pageactionapp.UpdateSeparatorAction{
				PageConfigID: &invalidUUID,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate"),
			CmpFunc: func(got, exp any) string {
				gotErr := got.(*errs.Error)
				if gotErr.Code != errs.InvalidArgument {
					return "expected InvalidArgument error"
				}
				return ""
			},
		},
	}
}

func updateSeparator401(sd apitest.SeedData) []apitest.Table {
	var separatorAction *pageactionapp.PageAction
	for _, action := range sd.PageActions {
		if action.ActionType == "separator" {
			separatorAction = &action
			break
		}
	}

	if separatorAction == nil {
		return []apitest.Table{}
	}

	newOrder := 999
	return []apitest.Table{
		{
			Name:       "unauthorized",
			URL:        "/v1/config/page-actions/separators/" + separatorAction.ID,
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input: &pageactionapp.UpdateSeparatorAction{
				ActionOrder: &newOrder,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[USER]] rule[rule_admin_only]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}