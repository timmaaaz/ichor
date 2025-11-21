package pageaction_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/config/pageactionapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

// =============================================================================
// Button Create Tests
// =============================================================================

func createButton200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/page-actions/buttons",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &pageactionapp.NewButtonAction{
				PageConfigID:       sd.PageConfigs[0].ID,
				ActionOrder:        100,
				IsActive:           true,
				Label:              "Test Button",
				Icon:               "test-icon",
				TargetPath:         "/test/path",
				Variant:            "default",
				Alignment:          "right",
				ConfirmationPrompt: "Are you sure?",
			},
			GotResp: &pageactionapp.PageAction{},
			ExpResp: &pageactionapp.PageAction{
				PageConfigID: sd.PageConfigs[0].ID,
				ActionType:   "button",
				ActionOrder:  100,
				IsActive:     true,
				Button: &pageactionapp.ButtonAction{
					Label:              "Test Button",
					Icon:               "test-icon",
					TargetPath:         "/test/path",
					Variant:            "default",
					Alignment:          "right",
					ConfirmationPrompt: "Are you sure?",
				},
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*pageactionapp.PageAction)
				expResp := exp.(*pageactionapp.PageAction)
				expResp.ID = gotResp.ID
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func createButton400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing_required_fields",
			URL:        "/v1/config/page-actions/buttons",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input:      &pageactionapp.NewButtonAction{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"pageConfigId\",\"error\":\"pageConfigId is a required field\"},{\"field\":\"label\",\"error\":\"label is a required field\"},{\"field\":\"targetPath\",\"error\":\"targetPath is a required field\"},{\"field\":\"variant\",\"error\":\"variant is a required field\"},{\"field\":\"alignment\",\"error\":\"alignment is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func createButton401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "unauthorized",
			URL:        "/v1/config/page-actions/buttons",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &pageactionapp.NewButtonAction{
				PageConfigID: sd.PageConfigs[0].ID,
				Label:        "Test Button",
				TargetPath:   "/test/path",
				Variant:      "default",
				Alignment:    "right",
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
// Dropdown Create Tests
// =============================================================================

func createDropdown200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/page-actions/dropdowns",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &pageactionapp.NewDropdownAction{
				PageConfigID: sd.PageConfigs[0].ID,
				ActionOrder:  200,
				IsActive:     true,
				Label:        "Test Dropdown",
				Icon:         "dropdown-icon",
				Items: []pageactionapp.NewDropdownItem{
					{Label: "Item 1", TargetPath: "/path1", ItemOrder: 1},
					{Label: "Item 2", TargetPath: "/path2", ItemOrder: 2},
				},
			},
			GotResp: &pageactionapp.PageAction{},
			ExpResp: &pageactionapp.PageAction{
				PageConfigID: sd.PageConfigs[0].ID,
				ActionType:   "dropdown",
				ActionOrder:  200,
				IsActive:     true,
				Dropdown: &pageactionapp.DropdownAction{
					Label: "Test Dropdown",
					Icon:  "dropdown-icon",
					Items: []pageactionapp.DropdownItem{
						{Label: "Item 1", TargetPath: "/path1", ItemOrder: 1},
						{Label: "Item 2", TargetPath: "/path2", ItemOrder: 2},
					},
				},
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*pageactionapp.PageAction)
				expResp := exp.(*pageactionapp.PageAction)
				expResp.ID = gotResp.ID
				// Set dropdown item IDs from response
				for i := range expResp.Dropdown.Items {
					expResp.Dropdown.Items[i].ID = gotResp.Dropdown.Items[i].ID
				}
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func createDropdown400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing_items",
			URL:        "/v1/config/page-actions/dropdowns",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &pageactionapp.NewDropdownAction{
				PageConfigID: sd.PageConfigs[0].ID,
				Label:        "Test Dropdown",
				Items:        []pageactionapp.NewDropdownItem{}, // Empty items
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"items\",\"error\":\"items must contain at least 1 item\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func createDropdown401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "unauthorized",
			URL:        "/v1/config/page-actions/dropdowns",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &pageactionapp.NewDropdownAction{
				PageConfigID: sd.PageConfigs[0].ID,
				Label:        "Test Dropdown",
				Items: []pageactionapp.NewDropdownItem{
					{Label: "Item 1", TargetPath: "/path1", ItemOrder: 1},
				},
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
// Separator Create Tests
// =============================================================================

func createSeparator200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/page-actions/separators",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &pageactionapp.NewSeparatorAction{
				PageConfigID: sd.PageConfigs[0].ID,
				ActionOrder:  300,
				IsActive:     true,
			},
			GotResp: &pageactionapp.PageAction{},
			ExpResp: &pageactionapp.PageAction{
				PageConfigID: sd.PageConfigs[0].ID,
				ActionType:   "separator",
				ActionOrder:  300,
				IsActive:     true,
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*pageactionapp.PageAction)
				expResp := exp.(*pageactionapp.PageAction)
				expResp.ID = gotResp.ID
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func createSeparator400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing_page_config",
			URL:        "/v1/config/page-actions/separators",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input:      &pageactionapp.NewSeparatorAction{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"pageConfigId\",\"error\":\"pageConfigId is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func createSeparator401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "unauthorized",
			URL:        "/v1/config/page-actions/separators",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &pageactionapp.NewSeparatorAction{
				PageConfigID: sd.PageConfigs[0].ID,
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
// Batch Create Tests
// =============================================================================

func batchCreate200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "mixed_actions",
			URL:        "/v1/config/page-configs/actions/batch/" + sd.PageConfigs[1].ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &pageactionapp.BatchCreateRequest{
				Actions: []pageactionapp.BatchActionRequest{
					{
						ActionType: "button",
						Button: &pageactionapp.NewButtonAction{
							PageConfigID: sd.PageConfigs[1].ID,
							ActionOrder:  1,
							IsActive:     true,
							Label:        "Batch Button",
							TargetPath:   "/batch/path",
							Variant:      "default",
							Alignment:    "left",
						},
					},
					{
						ActionType: "dropdown",
						Dropdown: &pageactionapp.NewDropdownAction{
							PageConfigID: sd.PageConfigs[1].ID,
							ActionOrder:  2,
							IsActive:     true,
							Label:        "Batch Dropdown",
							Items: []pageactionapp.NewDropdownItem{
								{Label: "Batch Item", TargetPath: "/batch/item", ItemOrder: 1},
							},
						},
					},
					{
						ActionType: "separator",
						Separator: &pageactionapp.NewSeparatorAction{
							PageConfigID: sd.PageConfigs[1].ID,
							ActionOrder:  3,
							IsActive:     true,
						},
					},
				},
			},
			GotResp: &pageactionapp.PageActions{},
			ExpResp: &pageactionapp.PageActions{},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*pageactionapp.PageActions)
				if len(*gotResp) != 3 {
					return "expected 3 actions in response"
				}
				return ""
			},
		},
	}
}

func batchCreate400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "invalid_action",
			URL:        "/v1/config/page-configs/actions/batch/" + sd.PageConfigs[1].ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &pageactionapp.BatchCreateRequest{
				Actions: []pageactionapp.BatchActionRequest{
					{
						ActionType: "button",
						Button:     &pageactionapp.NewButtonAction{}, // Missing required fields
					},
				},
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate"),
			CmpFunc: func(got, exp any) string {
				// Just check that we got an error
				gotErr := got.(*errs.Error)
				if gotErr.Code != errs.InvalidArgument {
					return "expected InvalidArgument error"
				}
				return ""
			},
		},
	}
}

func batchCreate401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "unauthorized",
			URL:        "/v1/config/page-configs/actions/batch/" + sd.PageConfigs[1].ID,
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &pageactionapp.BatchCreateRequest{
				Actions: []pageactionapp.BatchActionRequest{
					{
						ActionType: "button",
						Button: &pageactionapp.NewButtonAction{
							PageConfigID: sd.PageConfigs[1].ID,
							Label:        "Test",
							TargetPath:   "/test",
							Variant:      "default",
							Alignment:    "left",
						},
					},
				},
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[USER]] rule[rule_admin_only]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
