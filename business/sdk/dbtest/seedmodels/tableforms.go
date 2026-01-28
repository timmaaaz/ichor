package seedmodels

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
)

// =============================================================================
// REFERENCE DATA FORM FIELD GENERATORS
// =============================================================================
// These functions generate form field configurations for reference/lookup tables
// that are admin-managed and stable (is_reference_data=true, allow_inline_create=false)

// GetCountryFormFields returns form fields for creating countries (geography.countries)
// Reference data - ISO country list, admin-managed only
func GetCountryFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "geography",
			EntityTable:  "countries",
			Name:         "number",
			Label:        "Country Number",
			FieldType:    "number",
			FieldOrder:   1,
			Required:     true,
			Config:       json.RawMessage(`{}`),
		},
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "geography",
			EntityTable:  "countries",
			Name:         "name",
			Label:        "Country Name",
			FieldType:    "text",
			FieldOrder:   2,
			Required:     true,
			Config:       json.RawMessage(`{}`),
		},
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "geography",
			EntityTable:  "countries",
			Name:         "alpha_2",
			Label:        "Alpha-2 Code",
			FieldType:    "text",
			FieldOrder:   3,
			Required:     true,
			Config:       json.RawMessage(`{"maxLength": 2}`),
		},
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "geography",
			EntityTable:  "countries",
			Name:         "alpha_3",
			Label:        "Alpha-3 Code",
			FieldType:    "text",
			FieldOrder:   4,
			Required:     true,
			Config:       json.RawMessage(`{"maxLength": 3}`),
		},
	}
}

// GetRegionFormFields returns form fields for creating regions (geography.regions)
// Reference data - State/province data, admin-managed only
func GetRegionFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "geography",
			EntityTable:  "regions",
			Name:         "country_id",
			Label:        "Country",
			FieldType:    "smart-combobox",
			FieldOrder:   1,
			Required:     true,
			Config:       json.RawMessage(`{"entity": "geography.countries", "display_field": "name"}`),
		},
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "geography",
			EntityTable:  "regions",
			Name:         "name",
			Label:        "Region Name",
			FieldType:    "text",
			FieldOrder:   2,
			Required:     true,
			Config:       json.RawMessage(`{}`),
		},
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "geography",
			EntityTable:  "regions",
			Name:         "code",
			Label:        "Region Code",
			FieldType:    "text",
			FieldOrder:   3,
			Required:     true,
			Config:       json.RawMessage(`{}`),
		},
	}
}

// GetUserApprovalStatusFormFields returns form fields for user approval statuses (hr.user_approval_status)
// Reference data - Workflow status values, admin-managed only
func GetUserApprovalStatusFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "hr",
			EntityTable:  "user_approval_status",
			Name:         "name",
			Label:        "Status Name",
			FieldType:    "text",
			FieldOrder:   1,
			Required:     true,
			Config:       json.RawMessage(`{}`),
		},
	}
}

// GetAssetApprovalStatusFormFields returns form fields for asset approval statuses (assets.approval_status)
// Reference data - Asset approval workflow statuses, admin-managed only
func GetAssetApprovalStatusFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "assets",
			EntityTable:  "approval_status",
			Name:         "name",
			Label:        "Status Name",
			FieldType:    "text",
			FieldOrder:   1,
			Required:     true,
			Config:       json.RawMessage(`{}`),
		},
	}
}

// GetAssetFulfillmentStatusFormFields returns form fields for asset fulfillment statuses (assets.fulfillment_status)
// Reference data - Asset fulfillment workflow statuses, admin-managed only
func GetAssetFulfillmentStatusFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "assets",
			EntityTable:  "fulfillment_status",
			Name:         "name",
			Label:        "Status Name",
			FieldType:    "text",
			FieldOrder:   1,
			Required:     true,
			Config:       json.RawMessage(`{}`),
		},
	}
}

// GetOrderFulfillmentStatusFormFields returns form fields for order fulfillment statuses (sales.order_fulfillment_statuses)
// Reference data - Sales order workflow statuses, admin-managed only
func GetOrderFulfillmentStatusFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "sales",
			EntityTable:  "order_fulfillment_statuses",
			Name:         "name",
			Label:        "Status Name",
			FieldType:    "text",
			FieldOrder:   1,
			Required:     true,
			Config:       json.RawMessage(`{}`),
		},
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "sales",
			EntityTable:  "order_fulfillment_statuses",
			Name:         "description",
			Label:        "Description",
			FieldType:    "textarea",
			FieldOrder:   2,
			Required:     false,
			Config:       json.RawMessage(`{}`),
		},
	}
}

// GetLineItemFulfillmentStatusFormFields returns form fields for line item fulfillment statuses (sales.line_item_fulfillment_statuses)
// Reference data - Sales line item workflow statuses, admin-managed only
func GetLineItemFulfillmentStatusFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "sales",
			EntityTable:  "line_item_fulfillment_statuses",
			Name:         "name",
			Label:        "Status Name",
			FieldType:    "text",
			FieldOrder:   1,
			Required:     true,
			Config:       json.RawMessage(`{}`),
		},
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "sales",
			EntityTable:  "line_item_fulfillment_statuses",
			Name:         "description",
			Label:        "Description",
			FieldType:    "textarea",
			FieldOrder:   2,
			Required:     false,
			Config:       json.RawMessage(`{}`),
		},
	}
}

// GetPurchaseOrderStatusFormFields returns form fields for purchase order statuses (procurement.purchase_order_statuses)
// Reference data - Purchase order workflow statuses, admin-managed only
func GetPurchaseOrderStatusFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "procurement",
			EntityTable:  "purchase_order_statuses",
			Name:         "name",
			Label:        "Status Name",
			FieldType:    "text",
			FieldOrder:   1,
			Required:     true,
			Config:       json.RawMessage(`{}`),
		},
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "procurement",
			EntityTable:  "purchase_order_statuses",
			Name:         "description",
			Label:        "Description",
			FieldType:    "textarea",
			FieldOrder:   2,
			Required:     false,
			Config:       json.RawMessage(`{}`),
		},
	}
}

// GetPOLineItemStatusFormFields returns form fields for PO line item statuses (procurement.purchase_order_line_item_statuses)
// Reference data - Purchase order line item workflow statuses, admin-managed only
func GetPOLineItemStatusFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "procurement",
			EntityTable:  "purchase_order_line_item_statuses",
			Name:         "name",
			Label:        "Status Name",
			FieldType:    "text",
			FieldOrder:   1,
			Required:     true,
			Config:       json.RawMessage(`{}`),
		},
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "procurement",
			EntityTable:  "purchase_order_line_item_statuses",
			Name:         "description",
			Label:        "Description",
			FieldType:    "textarea",
			FieldOrder:   2,
			Required:     false,
			Config:       json.RawMessage(`{}`),
		},
	}
}

// GetAssetTypeFormFields returns form fields for asset types (assets.asset_types)
// Reference data - Asset classification, admin-managed only
func GetAssetTypeFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "assets",
			EntityTable:  "asset_types",
			Name:         "name",
			Label:        "Asset Type Name",
			FieldType:    "text",
			FieldOrder:   1,
			Required:     true,
			Config:       json.RawMessage(`{}`),
		},
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "assets",
			EntityTable:  "asset_types",
			Name:         "description",
			Label:        "Description",
			FieldType:    "textarea",
			FieldOrder:   2,
			Required:     false,
			Config:       json.RawMessage(`{}`),
		},
	}
}

// GetAssetConditionFormFields returns form fields for asset conditions (assets.asset_conditions)
// Reference data - Asset condition classification, admin-managed only
func GetAssetConditionFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "assets",
			EntityTable:  "asset_conditions",
			Name:         "name",
			Label:        "Condition Name",
			FieldType:    "text",
			FieldOrder:   1,
			Required:     true,
			Config:       json.RawMessage(`{}`),
		},
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "assets",
			EntityTable:  "asset_conditions",
			Name:         "description",
			Label:        "Description",
			FieldType:    "textarea",
			FieldOrder:   2,
			Required:     false,
			Config:       json.RawMessage(`{}`),
		},
	}
}

// GetProductCategoryFormFields returns form fields for product categories (products.product_categories)
// Reference data - Product taxonomy, admin-managed only
func GetProductCategoryFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "products",
			EntityTable:  "product_categories",
			Name:         "name",
			Label:        "Category Name",
			FieldType:    "text",
			FieldOrder:   1,
			Required:     true,
			Config:       json.RawMessage(`{}`),
		},
		{
			FormID:       formID,
			EntityID:     entityID,
			EntitySchema: "products",
			EntityTable:  "product_categories",
			Name:         "description",
			Label:        "Description",
			FieldType:    "textarea",
			FieldOrder:   2,
			Required:     false,
			Config:       json.RawMessage(`{}`),
		},
		// Audit field - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "product_categories", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 3, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
	}
}

// =============================================================================
// TRANSACTIONAL DATA FORM FIELD GENERATORS
// =============================================================================
// These functions generate form field configurations for user-created transactional tables
// (is_reference_data=false, allow_inline_create=true)

// GetCityFormFields returns form fields for cities (geography.cities)
func GetCityFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID: formID, EntityID: entityID, EntitySchema: "geography", EntityTable: "cities",
			Name: "region_id", Label: "Region", FieldType: "smart-combobox", FieldOrder: 1, Required: true, Config: json.RawMessage(`{"entity": "geography.regions", "display_field": "name"}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "geography", EntityTable: "cities",
			Name: "name", Label: "City Name", FieldType: "text", FieldOrder: 2, Required: true, Config: json.RawMessage(`{}`),
		},
	}
}

// GetStreetFormFields returns form fields for streets (geography.streets)
func GetStreetFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		// ID field - required for update operations to work correctly
		{FormID: formID, EntityID: entityID, EntitySchema: "geography", EntityTable: "streets", Name: "id", Label: "ID", FieldType: "hidden", FieldOrder: 0, Required: false, Config: json.RawMessage(`{"hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "geography", EntityTable: "streets", Name: "line_1", Label: "Street Address", FieldType: "text", FieldOrder: 1, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "geography", EntityTable: "streets", Name: "line_2", Label: "Address Line 2", FieldType: "text", FieldOrder: 2, Required: false, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "geography", EntityTable: "streets", Name: "postal_code", Label: "Postal Code", FieldType: "text", FieldOrder: 3, Required: false, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "geography", EntityTable: "streets", Name: "city_id", Label: "City", FieldType: "smart-combobox", FieldOrder: 4, Required: true, Config: json.RawMessage(`{"entity": "geography.cities", "display_field": "name", "inline_create": {"enabled": true, "form_name": "City Creation Form", "button_text": "Create City"}}`)},
	}
}

// GetContactInfoFormFields returns form fields for contact infos (core.contact_infos)
func GetContactInfoFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		// ID field - required for update operations to work correctly
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "contact_infos", Name: "id", Label: "ID", FieldType: "hidden", FieldOrder: 0, Required: false, Config: json.RawMessage(`{"hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "contact_infos", Name: "first_name", Label: "Contact First Name", FieldType: "text", FieldOrder: 1, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "contact_infos", Name: "last_name", Label: "Contact Last Name", FieldType: "text", FieldOrder: 2, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "contact_infos", Name: "primary_phone_number", Label: "Primary Phone", FieldType: "tel", FieldOrder: 3, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "contact_infos", Name: "secondary_phone_number", Label: "Secondary Phone", FieldType: "tel", FieldOrder: 4, Required: false, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "contact_infos", Name: "email_address", Label: "Email", FieldType: "email", FieldOrder: 5, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "contact_infos", Name: "street_id", Label: "Street Address", FieldType: "smart-combobox", FieldOrder: 6, Required: true, Config: json.RawMessage(`{"entity": "geography.streets", "display_field": "line_1", "inline_create": {"enabled": true, "form_name": "Street Creation Form", "button_text": "Create Street"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "contact_infos", Name: "available_hours_start", Label: "Available From", FieldType: "time", FieldOrder: 7, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "contact_infos", Name: "available_hours_end", Label: "Available Until", FieldType: "time", FieldOrder: 8, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "contact_infos", Name: "timezone_id", Label: "Timezone", FieldType: "smart-combobox", FieldOrder: 9, Required: true, Config: json.RawMessage(`{"entity": "geography.timezones", "display_field": "display_name"}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "contact_infos", Name: "preferred_contact_type", Label: "Preferred Contact Method", FieldType: "enum", FieldOrder: 10, Required: true, Config: json.RawMessage(`{"enum_name": "public.contact_type"}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "contact_infos", Name: "notes", Label: "Notes", FieldType: "textarea", FieldOrder: 11, Required: false, Config: json.RawMessage(`{}`)},
	}
}

// GetTitleFormFields returns form fields for titles (hr.titles)
func GetTitleFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID: formID, EntityID: entityID, EntitySchema: "hr", EntityTable: "titles",
			Name: "name", Label: "Title Name", FieldType: "text", FieldOrder: 1, Required: true, Config: json.RawMessage(`{}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "hr", EntityTable: "titles",
			Name: "description", Label: "Description", FieldType: "textarea", FieldOrder: 2, Required: false, Config: json.RawMessage(`{}`),
		},
	}
}

// GetOfficeFormFields returns form fields for offices (hr.offices)
func GetOfficeFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "hr", EntityTable: "offices", Name: "name", Label: "Office Name", FieldType: "text", FieldOrder: 1, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "hr", EntityTable: "offices", Name: "street_id", Label: "Address", FieldType: "smart-combobox", FieldOrder: 2, Required: true, Config: json.RawMessage(`{"entity": "geography.streets", "display_field": "line_1", "inline_create": {"enabled": true, "form_name": "Street Creation Form", "button_text": "Create Street"}}`)},
	}
}

// GetHomeFormFields returns form fields for homes (hr.homes)
func GetHomeFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID: formID, EntityID: entityID, EntitySchema: "hr", EntityTable: "homes",
			Name: "type", Label: "Home Type", FieldType: "text", FieldOrder: 1, Required: true, Config: json.RawMessage(`{}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "hr", EntityTable: "homes",
			Name: "user_id", Label: "User", FieldType: "smart-combobox", FieldOrder: 2, Required: true, Config: json.RawMessage(`{"entity": "core.users", "display_field": "username"}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "hr", EntityTable: "homes",
			Name: "address_1", Label: "Address Line 1", FieldType: "text", FieldOrder: 3, Required: true, Config: json.RawMessage(`{}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "hr", EntityTable: "homes",
			Name: "address_2", Label: "Address Line 2", FieldType: "text", FieldOrder: 4, Required: false, Config: json.RawMessage(`{}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "hr", EntityTable: "homes",
			Name: "city", Label: "City", FieldType: "text", FieldOrder: 5, Required: true, Config: json.RawMessage(`{}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "hr", EntityTable: "homes",
			Name: "state", Label: "State", FieldType: "text", FieldOrder: 6, Required: true, Config: json.RawMessage(`{}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "hr", EntityTable: "homes",
			Name: "zip_code", Label: "Zip Code", FieldType: "text", FieldOrder: 7, Required: true, Config: json.RawMessage(`{}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "hr", EntityTable: "homes",
			Name: "country", Label: "Country", FieldType: "text", FieldOrder: 8, Required: true, Config: json.RawMessage(`{}`),
		},
		// Audit field - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "hr", EntityTable: "homes", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 9, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
	}
}

// GetTagFormFields returns form fields for tags (assets.tags)
func GetTagFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "tags",
			Name: "name", Label: "Tag Name", FieldType: "text", FieldOrder: 1, Required: true, Config: json.RawMessage(`{}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "tags",
			Name: "description", Label: "Description", FieldType: "textarea", FieldOrder: 2, Required: false, Config: json.RawMessage(`{}`),
		},
	}
}

// GetUserApprovalCommentFormFields returns form fields for user approval comments (hr.user_approval_comments)
func GetUserApprovalCommentFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID: formID, EntityID: entityID, EntitySchema: "hr", EntityTable: "user_approval_comments",
			Name: "comment", Label: "Comment", FieldType: "textarea", FieldOrder: 1, Required: true, Config: json.RawMessage(`{}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "hr", EntityTable: "user_approval_comments",
			Name: "commenter_id", Label: "Commenter", FieldType: "smart-combobox", FieldOrder: 2, Required: true, Config: json.RawMessage(`{"entity": "core.users", "display_field": "username"}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "hr", EntityTable: "user_approval_comments",
			Name: "user_id", Label: "User", FieldType: "smart-combobox", FieldOrder: 3, Required: true, Config: json.RawMessage(`{"entity": "core.users", "display_field": "username"}`),
		},
		// Audit field - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "hr", EntityTable: "user_approval_comments", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 4, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
	}
}

// GetBrandFormFields returns form fields for brands (products.brands)
func GetBrandFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "brands",
			Name: "name", Label: "Brand Name", FieldType: "text", FieldOrder: 1, Required: true, Config: json.RawMessage(`{}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "brands",
			Name: "contact_infos_id", Label: "Contact Info", FieldType: "smart-combobox", FieldOrder: 2, Required: true, Config: json.RawMessage(`{"entity": "core.contact_infos", "display_field": "email_address"}`),
		},
		// Audit field - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "brands", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 3, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
	}
}

// GetProductFormFields returns form fields for products (products.products)
func GetProductFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		// ID field - required for update operations to work correctly
		{
			FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "products",
			Name: "id", Label: "ID", FieldType: "hidden", FieldOrder: 0, Required: false, Config: json.RawMessage(`{"hidden": true}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "products",
			Name: "sku", Label: "SKU", FieldType: "text", FieldOrder: 1, Required: true, Config: json.RawMessage(`{}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "products",
			Name: "name", Label: "Product Name", FieldType: "text", FieldOrder: 2, Required: true, Config: json.RawMessage(`{}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "products",
			Name: "brand_id", Label: "Brand", FieldType: "smart-combobox", FieldOrder: 3, Required: true, Config: json.RawMessage(`{"entity": "products.brands", "display_field": "name"}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "products",
			Name: "category_id", Label: "Category", FieldType: "smart-combobox", FieldOrder: 4, Required: true, Config: json.RawMessage(`{"entity": "products.product_categories", "display_field": "name"}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "products",
			Name: "description", Label: "Description", FieldType: "textarea", FieldOrder: 5, Required: true, Config: json.RawMessage(`{}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "products",
			Name: "model_number", Label: "Model Number", FieldType: "text", FieldOrder: 6, Required: false, Config: json.RawMessage(`{}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "products",
			Name: "upc_code", Label: "UPC Code", FieldType: "text", FieldOrder: 7, Required: true, Config: json.RawMessage(`{}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "products",
			Name: "status", Label: "Status", FieldType: "text", FieldOrder: 8, Required: true, Config: json.RawMessage(`{}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "products",
			Name: "is_active", Label: "Is Active", FieldType: "boolean", FieldOrder: 9, Required: true, Config: json.RawMessage(`{}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "products",
			Name: "is_perishable", Label: "Is Perishable", FieldType: "boolean", FieldOrder: 10, Required: true, Config: json.RawMessage(`{}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "products",
			Name: "handling_instructions", Label: "Handling Instructions", FieldType: "textarea", FieldOrder: 11, Required: false, Config: json.RawMessage(`{}`),
		},
		{
			FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "products",
			Name: "units_per_case", Label: "Units Per Case", FieldType: "number", FieldOrder: 12, Required: true, Config: json.RawMessage(`{}`),
		},
		// Audit field - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "products", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 13, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
	}
}

// GetPhysicalAttributeFormFields returns form fields for physical attributes (products.physical_attributes)
func GetPhysicalAttributeFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "physical_attributes", Name: "product_id", Label: "Product", FieldType: "smart-combobox", FieldOrder: 1, Required: true, Config: json.RawMessage(`{"entity": "products.products", "display_field": "name"}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "physical_attributes", Name: "length", Label: "Length", FieldType: "number", FieldOrder: 2, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "physical_attributes", Name: "width", Label: "Width", FieldType: "number", FieldOrder: 3, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "physical_attributes", Name: "height", Label: "Height", FieldType: "number", FieldOrder: 4, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "physical_attributes", Name: "weight", Label: "Weight", FieldType: "number", FieldOrder: 5, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "physical_attributes", Name: "weight_unit", Label: "Weight Unit", FieldType: "text", FieldOrder: 6, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "physical_attributes", Name: "color", Label: "Color", FieldType: "text", FieldOrder: 7, Required: false, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "physical_attributes", Name: "size", Label: "Size", FieldType: "text", FieldOrder: 8, Required: false, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "physical_attributes", Name: "material", Label: "Material", FieldType: "text", FieldOrder: 9, Required: false, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "physical_attributes", Name: "storage_requirements", Label: "Storage Requirements", FieldType: "textarea", FieldOrder: 10, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "physical_attributes", Name: "hazmat_class", Label: "Hazmat Class", FieldType: "text", FieldOrder: 11, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "physical_attributes", Name: "shelf_life_days", Label: "Shelf Life (Days)", FieldType: "number", FieldOrder: 12, Required: true, Config: json.RawMessage(`{}`)},
		// Audit field - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "physical_attributes", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 13, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
	}
}

// GetProductCostFormFields returns form fields for product costs (products.product_costs)
func GetProductCostFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "product_costs", Name: "product_id", Label: "Product", FieldType: "smart-combobox", FieldOrder: 1, Required: true, Config: json.RawMessage(`{"entity": "products.products", "display_field": "name"}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "product_costs", Name: "purchase_cost", Label: "Purchase Cost", FieldType: "number", FieldOrder: 2, Required: true, Config: json.RawMessage(`{}`)},
		// Audit field - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "product_costs", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 3, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
	}
}

// GetCostHistoryFormFields returns form fields for cost history (products.cost_history)
func GetCostHistoryFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "cost_history", Name: "product_id", Label: "Product", FieldType: "smart-combobox", FieldOrder: 1, Required: true, Config: json.RawMessage(`{"entity": "products.products", "display_field": "name"}`)},
		// Audit field - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "cost_history", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 2, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
	}
}

// GetQualityMetricFormFields returns form fields for quality metrics (products.quality_metrics)
func GetQualityMetricFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "quality_metrics", Name: "product_id", Label: "Product", FieldType: "smart-combobox", FieldOrder: 1, Required: true, Config: json.RawMessage(`{"entity": "products.products", "display_field": "name"}`)},
		// Audit field - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "products", EntityTable: "quality_metrics", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 2, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
	}
}

// GetWarehouseFormFields returns form fields for warehouses (inventory.warehouses)
func GetWarehouseFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "warehouses", Name: "code", Label: "Warehouse Code", FieldType: "text", FieldOrder: 1, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "warehouses", Name: "name", Label: "Warehouse Name", FieldType: "text", FieldOrder: 2, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "warehouses", Name: "street_id", Label: "Address", FieldType: "smart-combobox", FieldOrder: 3, Required: true, Config: json.RawMessage(`{"entity": "geography.streets", "display_field": "line_1", "inline_create": {"enabled": true, "form_name": "Street Creation Form", "button_text": "Create Street"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "warehouses", Name: "is_active", Label: "Active", FieldType: "boolean", FieldOrder: 4, Required: true, Config: json.RawMessage(`{}`)},
		// Audit fields - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "warehouses", Name: "created_by", Label: "Created By", FieldType: "hidden", FieldOrder: 5, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$me}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "warehouses", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 6, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "warehouses", Name: "updated_by", Label: "Updated By", FieldType: "hidden", FieldOrder: 7, Required: false, Config: json.RawMessage(`{"default_value": "{{$me}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "warehouses", Name: "updated_date", Label: "Updated Date", FieldType: "hidden", FieldOrder: 8, Required: false, Config: json.RawMessage(`{"default_value": "{{$now}}", "hidden": true}`)},
	}
}

// GetZoneFormFields returns form fields for zones (inventory.zones)
func GetZoneFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "zones", Name: "warehouse_id", Label: "Warehouse", FieldType: "smart-combobox", FieldOrder: 1, Required: true, Config: json.RawMessage(`{"entity": "inventory.warehouses", "display_field": "name"}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "zones", Name: "name", Label: "Zone Name", FieldType: "text", FieldOrder: 2, Required: true, Config: json.RawMessage(`{}`)},
		// Audit field - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "zones", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 3, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
	}
}

// GetInventoryLocationFormFields returns form fields for inventory locations (inventory.inventory_locations)
func GetInventoryLocationFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "inventory_locations", Name: "zone_id", Label: "Zone", FieldType: "smart-combobox", FieldOrder: 1, Required: true, Config: json.RawMessage(`{"entity": "inventory.zones", "display_field": "name"}`)},
		// Audit field - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "inventory_locations", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 2, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
	}
}

// GetInventoryItemFormFields returns form fields for inventory items (inventory.inventory_items)
func GetInventoryItemFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "inventory_items", Name: "product_id", Label: "Product", FieldType: "smart-combobox", FieldOrder: 1, Required: true, Config: json.RawMessage(`{"entity": "products.products", "display_field": "name", "inline_create": {"enabled": true, "form_name": "Product Creation Form", "button_text": "Create Product"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "inventory_items", Name: "location_id", Label: "Location", FieldType: "smart-combobox", FieldOrder: 2, Required: true, Config: json.RawMessage(`{"entity": "inventory.inventory_locations", "display_field": "aisle", "inline_create": {"enabled": true, "form_name": "Inventory Location Creation Form", "button_text": "Create Inventory Location"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "inventory_items", Name: "quantity", Label: "Quantity", FieldType: "number", FieldOrder: 3, Required: true, Config: json.RawMessage(`{"min": 0, "max": 1000000}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "inventory_items", Name: "reserved_quantity", Label: "Reserved Quantity", FieldType: "number", FieldOrder: 4, Required: true, Config: json.RawMessage(`{"min": 0, "max": 1000000}`)},
		// Audit field - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "inventory_items", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 5, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
	}
}

// GetSerialNumberFormFields returns form fields for serial numbers (inventory.serial_numbers)
func GetSerialNumberFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "serial_numbers", Name: "product_id", Label: "Product", FieldType: "smart-combobox", FieldOrder: 1, Required: true, Config: json.RawMessage(`{"entity": "products.products", "display_field": "name"}`)},
		// Audit field - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "serial_numbers", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 2, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
	}
}

// GetLotTrackingFormFields returns form fields for lot trackings (inventory.lot_trackings)
func GetLotTrackingFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "lot_trackings", Name: "supplier_product_id", Label: "Supplier Product", FieldType: "smart-combobox", FieldOrder: 1, Required: true, Config: json.RawMessage(`{"entity": "procurement.supplier_products", "display_field": "supplier_part_number"}`)},
		// Audit field - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "lot_trackings", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 2, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
	}
}

// GetQualityInspectionFormFields returns form fields for quality inspections (inventory.quality_inspections)
func GetQualityInspectionFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "quality_inspections", Name: "product_id", Label: "Product", FieldType: "smart-combobox", FieldOrder: 1, Required: true, Config: json.RawMessage(`{"entity": "products.products", "display_field": "name"}`)},
		// Audit field - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "quality_inspections", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 2, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
	}
}

// GetInventoryTransactionFormFields returns form fields for inventory transactions (inventory.inventory_transactions)
func GetInventoryTransactionFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "inventory_transactions", Name: "product_id", Label: "Product", FieldType: "smart-combobox", FieldOrder: 1, Required: true, Config: json.RawMessage(`{"entity": "products.products", "display_field": "name"}`)},
		// Audit field - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "inventory_transactions", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 2, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
	}
}

// GetInventoryAdjustmentFormFields returns form fields for inventory adjustments (inventory.inventory_adjustments)
func GetInventoryAdjustmentFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "inventory_adjustments", Name: "product_id", Label: "Product", FieldType: "smart-combobox", FieldOrder: 1, Required: true, Config: json.RawMessage(`{"entity": "products.products", "display_field": "name", "inline_create": {"enabled": true, "form_name": "Product Creation Form", "button_text": "Create Product"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "inventory_adjustments", Name: "location_id", Label: "Location", FieldType: "smart-combobox", FieldOrder: 2, Required: true, Config: json.RawMessage(`{"entity": "inventory.inventory_locations", "display_field": "aisle", "inline_create": {"enabled": true, "form_name": "Inventory Location Creation Form", "button_text": "Create Inventory Location"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "inventory_adjustments", Name: "quantity_change", Label: "Quantity Change", FieldType: "number", FieldOrder: 3, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "inventory_adjustments", Name: "reason_code", Label: "Reason Code", FieldType: "text", FieldOrder: 4, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "inventory_adjustments", Name: "notes", Label: "Notes", FieldType: "textarea", FieldOrder: 5, Required: false, Config: json.RawMessage(`{}`)},
		// Audit field - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "inventory_adjustments", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 6, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
	}
}

// GetTransferOrderFormFields returns form fields for transfer orders (inventory.transfer_orders)
func GetTransferOrderFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "transfer_orders", Name: "product_id", Label: "Product", FieldType: "smart-combobox", FieldOrder: 1, Required: true, Config: json.RawMessage(`{"entity": "products.products", "display_field": "name", "inline_create": {"enabled": true, "form_name": "Product Creation Form", "button_text": "Create Product"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "transfer_orders", Name: "from_location_id", Label: "From Location", FieldType: "smart-combobox", FieldOrder: 2, Required: true, Config: json.RawMessage(`{"entity": "inventory.inventory_locations", "display_field": "aisle", "inline_create": {"enabled": true, "form_name": "Inventory Location Creation Form", "button_text": "Create Inventory Location"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "transfer_orders", Name: "to_location_id", Label: "To Location", FieldType: "smart-combobox", FieldOrder: 3, Required: true, Config: json.RawMessage(`{"entity": "inventory.inventory_locations", "display_field": "aisle", "inline_create": {"enabled": true, "form_name": "Inventory Location Creation Form", "button_text": "Create Inventory Location"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "transfer_orders", Name: "quantity", Label: "Quantity", FieldType: "number", FieldOrder: 4, Required: true, Config: json.RawMessage(`{"min": 1, "max": 1000000}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "transfer_orders", Name: "status", Label: "Status", FieldType: "text", FieldOrder: 5, Required: true, Config: json.RawMessage(`{}`)},
		// Audit field - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "inventory", EntityTable: "transfer_orders", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 6, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
	}
}

// GetValidAssetFormFields returns form fields for valid assets (assets.valid_assets)
func GetValidAssetFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "valid_assets", Name: "type_id", Label: "Asset Type", FieldType: "smart-combobox", FieldOrder: 1, Required: true, Config: json.RawMessage(`{"entity": "assets.asset_types", "display_field": "name"}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "valid_assets", Name: "name", Label: "Asset Name", FieldType: "text", FieldOrder: 2, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "valid_assets", Name: "est_price", Label: "Estimated Price", FieldType: "number", FieldOrder: 3, Required: false, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "valid_assets", Name: "price", Label: "Price", FieldType: "number", FieldOrder: 4, Required: false, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "valid_assets", Name: "serial_number", Label: "Serial Number", FieldType: "text", FieldOrder: 5, Required: false, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "valid_assets", Name: "model_number", Label: "Model Number", FieldType: "text", FieldOrder: 6, Required: false, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "valid_assets", Name: "is_enabled", Label: "Is Enabled", FieldType: "boolean", FieldOrder: 7, Required: true, Config: json.RawMessage(`{}`)},
		// Audit fields - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "valid_assets", Name: "created_by", Label: "Created By", FieldType: "hidden", FieldOrder: 8, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$me}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "valid_assets", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 9, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "valid_assets", Name: "updated_by", Label: "Updated By", FieldType: "hidden", FieldOrder: 10, Required: false, Config: json.RawMessage(`{"default_value": "{{$me}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "valid_assets", Name: "updated_date", Label: "Updated Date", FieldType: "hidden", FieldOrder: 11, Required: false, Config: json.RawMessage(`{"default_value": "{{$now}}", "hidden": true}`)},
	}
}

// GetAssetFormFields returns form fields for assets (assets.assets)
func GetAssetFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "assets", Name: "valid_asset_id", Label: "Asset Type", FieldType: "smart-combobox", FieldOrder: 1, Required: true, Config: json.RawMessage(`{"entity": "assets.valid_assets", "display_field": "name", "inline_create": {"enabled": true, "form_name": "Valid Asset Creation Form", "button_text": "Create Valid Asset"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "assets", Name: "serial_number", Label: "Serial Number", FieldType: "text", FieldOrder: 2, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "assets", Name: "asset_condition_id", Label: "Condition", FieldType: "smart-combobox", FieldOrder: 3, Required: true, Config: json.RawMessage(`{"entity": "assets.asset_conditions", "display_field": "name", "inline_create": {"enabled": true, "form_name": "Asset Condition Creation Form", "button_text": "Create Asset Condition"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "assets", Name: "last_maintenance_time", Label: "Last Maintenance", FieldType: "date", FieldOrder: 4, Required: false, Config: json.RawMessage(`{}`)},
	}
}

// GetUserAssetFormFields returns form fields for user assets (assets.user_assets)
func GetUserAssetFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "user_assets", Name: "user_id", Label: "User", FieldType: "smart-combobox", FieldOrder: 1, Required: true, Config: json.RawMessage(`{"entity": "core.users", "display_field": "username"}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "user_assets", Name: "asset_id", Label: "Asset", FieldType: "smart-combobox", FieldOrder: 2, Required: true, Config: json.RawMessage(`{"entity": "assets.assets", "display_field": "serial_number"}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "user_assets", Name: "approval_status_id", Label: "Approval Status", FieldType: "smart-combobox", FieldOrder: 3, Required: true, Config: json.RawMessage(`{"entity": "assets.approval_status", "label_column": "name", "value_column": "id", "default_value_create": "WAITING"}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "user_assets", Name: "fulfillment_status_id", Label: "Fulfillment Status", FieldType: "smart-combobox", FieldOrder: 4, Required: true, Config: json.RawMessage(`{"entity": "assets.fulfillment_status", "label_column": "name", "value_column": "id", "default_value_create": "WAITING"}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "user_assets", Name: "last_maintenance", Label: "Last Maintenance", FieldType: "date", FieldOrder: 5, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "user_assets", Name: "date_received", Label: "Date Received", FieldType: "date", FieldOrder: 6, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "assets", EntityTable: "user_assets", Name: "approved_by", Label: "Approved By", FieldType: "smart-combobox", FieldOrder: 7, Required: true, Config: json.RawMessage(`{"entity": "core.users", "display_field": "username"}`)},
	}
}

// GetAutomationRuleFormFields returns form fields for automation rules (workflow.automation_rules)
func GetAutomationRuleFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "workflow", EntityTable: "automation_rules", Name: "name", Label: "Rule Name", FieldType: "text", FieldOrder: 1, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "workflow", EntityTable: "automation_rules", Name: "description", Label: "Description", FieldType: "textarea", FieldOrder: 2, Required: false, Config: json.RawMessage(`{}`)},
		// Audit fields - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "workflow", EntityTable: "automation_rules", Name: "created_by", Label: "Created By", FieldType: "hidden", FieldOrder: 3, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$me}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "workflow", EntityTable: "automation_rules", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 4, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "workflow", EntityTable: "automation_rules", Name: "updated_by", Label: "Updated By", FieldType: "hidden", FieldOrder: 5, Required: false, Config: json.RawMessage(`{"default_value": "{{$me}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "workflow", EntityTable: "automation_rules", Name: "updated_date", Label: "Updated Date", FieldType: "hidden", FieldOrder: 6, Required: false, Config: json.RawMessage(`{"default_value": "{{$now}}", "hidden": true}`)},
	}
}

// GetRuleActionFormFields returns form fields for rule actions (workflow.rule_actions)
func GetRuleActionFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "workflow", EntityTable: "rule_actions", Name: "automation_rules_id", Label: "Automation Rule", FieldType: "smart-combobox", FieldOrder: 1, Required: true, Config: json.RawMessage(`{"entity": "workflow.automation_rules", "display_field": "name"}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "workflow", EntityTable: "rule_actions", Name: "name", Label: "Action Name", FieldType: "text", FieldOrder: 2, Required: true, Config: json.RawMessage(`{}`)},
	}
}

// GetEntityFormFields returns form fields for entities (workflow.entities)
func GetEntityFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "workflow", EntityTable: "entities", Name: "name", Label: "Entity Name", FieldType: "text", FieldOrder: 1, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "workflow", EntityTable: "entities", Name: "entity_type_id", Label: "Entity Type", FieldType: "smart-combobox", FieldOrder: 2, Required: true, Config: json.RawMessage(`{"entity": "workflow.entity_types", "display_field": "name"}`)},
	}
}

// GetCustomerFormFields returns form fields for customers (sales.customers)
func GetCustomerFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		// ID field - required for update operations to work correctly
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "customers", Name: "id", Label: "ID", FieldType: "hidden", FieldOrder: 0, Required: false, Config: json.RawMessage(`{"hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "customers", Name: "name", Label: "Customer Name", FieldType: "text", FieldOrder: 1, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "customers", Name: "contact_id", Label: "Contact Information", FieldType: "smart-combobox", FieldOrder: 2, Required: true, Config: json.RawMessage(`{"entity": "core.contact_infos", "display_field": "email_address", "inline_create": {"enabled": true, "form_name": "Contact Info Creation Form", "button_text": "Create Contact Info"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "customers", Name: "delivery_address_id", Label: "Delivery Address", FieldType: "smart-combobox", FieldOrder: 3, Required: true, Config: json.RawMessage(`{"entity": "geography.streets", "display_field": "line_1", "inline_create": {"enabled": true, "form_name": "Street Creation Form", "button_text": "Create Street"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "customers", Name: "notes", Label: "Notes", FieldType: "textarea", FieldOrder: 4, Required: false, Config: json.RawMessage(`{}`)},
		// Audit fields - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "customers", Name: "created_by", Label: "Created By", FieldType: "hidden", FieldOrder: 5, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$me}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "customers", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 6, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "customers", Name: "updated_by", Label: "Updated By", FieldType: "hidden", FieldOrder: 7, Required: false, Config: json.RawMessage(`{"default_value": "{{$me}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "customers", Name: "updated_date", Label: "Updated Date", FieldType: "hidden", FieldOrder: 8, Required: false, Config: json.RawMessage(`{"default_value": "{{$now}}", "hidden": true}`)},
	}
}

// GetSalesOrderFormFields returns form fields for sales orders (sales.orders)
func GetSalesOrderFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		// ID field - required for update operations to work correctly
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "orders", Name: "id", Label: "ID", FieldType: "hidden", FieldOrder: 0, Required: false, Config: json.RawMessage(`{"hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "orders", Name: "number", Label: "Order Number", FieldType: "text", FieldOrder: 1, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "orders", Name: "customer_id", Label: "Customer", FieldType: "smart-combobox", FieldOrder: 2, Required: true, Config: json.RawMessage(`{"entity": "sales.customers", "display_field": "name", "inline_create": {"enabled": true, "form_name": "Customer Creation Form", "button_text": "Create Customer"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "orders", Name: "order_date", Label: "Order Date", FieldType: "date", FieldOrder: 3, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "orders", Name: "due_date", Label: "Due Date", FieldType: "date", FieldOrder: 4, Required: true, Config: json.RawMessage(`{"must_be_future": true, "min_date": "{{order_date}}"}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "orders", Name: "billing_address_id", Label: "Billing Address", FieldType: "smart-combobox", FieldOrder: 5, Required: false, Config: json.RawMessage(`{"entity": "geography.streets", "display_field": "line_1", "inline_create": {"enabled": true, "form_name": "Street Creation Form", "button_text": "Create Street"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "orders", Name: "shipping_address_id", Label: "Shipping Address", FieldType: "smart-combobox", FieldOrder: 6, Required: false, Config: json.RawMessage(`{"entity": "geography.streets", "display_field": "line_1", "inline_create": {"enabled": true, "form_name": "Street Creation Form", "button_text": "Create Street"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "orders", Name: "tax_rate", Label: "Tax Rate", FieldType: "percent", FieldOrder: 7, Required: false, Config: json.RawMessage(`{"min": 0, "max": 100, "precision": 2}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "orders", Name: "shipping_cost", Label: "Shipping Cost", FieldType: "currency", FieldOrder: 8, Required: false, Config: json.RawMessage(`{"min": 0, "max": 10000000, "precision": 2}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "orders", Name: "payment_term_id", Label: "Payment Terms", FieldType: "smart-combobox", FieldOrder: 9, Required: false, Config: json.RawMessage(`{"entity": "core.payment_terms", "display_field": "name"}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "orders", Name: "notes", Label: "Notes", FieldType: "textarea", FieldOrder: 10, Required: false, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "orders", Name: "order_fulfillment_status_id", Label: "Fulfillment Status", FieldType: "hidden", FieldOrder: 11, Required: false, Config: json.RawMessage(`{"entity": "sales.order_fulfillment_statuses", "label_column": "name", "value_column": "id", "default_value_create": "PENDING", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "orders", Name: "currency_id", Label: "Currency", FieldType: "hidden", FieldOrder: 12, Required: false, Config: json.RawMessage(`{"entity": "core.currencies", "label_column": "code", "value_column": "id", "default_value_create": "USD", "hidden": true}`)},
		// Audit fields - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "orders", Name: "created_by", Label: "Created By", FieldType: "hidden", FieldOrder: 13, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$me}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "orders", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 14, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "orders", Name: "updated_by", Label: "Updated By", FieldType: "hidden", FieldOrder: 15, Required: false, Config: json.RawMessage(`{"default_value": "{{$me}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "orders", Name: "updated_date", Label: "Updated Date", FieldType: "hidden", FieldOrder: 16, Required: false, Config: json.RawMessage(`{"default_value": "{{$now}}", "hidden": true}`)},
	}
}

// GetSalesOrderLineItemFormFields returns form fields for order line items (sales.order_line_items)
func GetSalesOrderLineItemFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		// ID field - required for update operations to work correctly
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "order_line_items", Name: "id", Label: "ID", FieldType: "hidden", FieldOrder: 0, Required: false, Config: json.RawMessage(`{"hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "order_line_items", Name: "product_id", Label: "Product", FieldType: "smart-combobox", FieldOrder: 1, Required: true, Config: json.RawMessage(`{"entity": "products.products", "display_field": "name", "inline_create": {"enabled": true, "form_name": "Product Creation Form", "button_text": "Create Product"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "order_line_items", Name: "description", Label: "Description", FieldType: "textarea", FieldOrder: 2, Required: false, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "order_line_items", Name: "quantity", Label: "Quantity", FieldType: "number", FieldOrder: 3, Required: true, Config: json.RawMessage(`{"min": 1, "max": 1000000}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "order_line_items", Name: "unit_price", Label: "Unit Price", FieldType: "currency", FieldOrder: 4, Required: true, Config: json.RawMessage(`{"min": 0, "max": 10000000, "precision": 2}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "order_line_items", Name: "discount", Label: "Discount", FieldType: "currency", FieldOrder: 5, Required: false, Config: json.RawMessage(`{"min":0,"precision":2,"depends_on":{"field":"discount_type","value_mappings":{"flat":{"type":"currency","label":"Discount ($)"},"percent":{"type":"percent","label":"Discount (%)","validation":{"max":100}}},"default":{"type":"currency"}}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "order_line_items", Name: "discount_type", Label: "Discount Type", FieldType: "enum", FieldOrder: 6, Required: false, Config: json.RawMessage(`{"enum_name": "sales.discount_type", "default_value_create": "flat"}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "order_line_items", Name: "line_item_fulfillment_statuses_id", Label: "Fulfillment Status", FieldType: "hidden", FieldOrder: 7, Required: false, Config: json.RawMessage(`{"entity": "sales.line_item_fulfillment_statuses", "label_column": "name", "value_column": "id", "default_value_create": "PENDING", "hidden": true}`)},
		// Audit fields - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "order_line_items", Name: "created_by", Label: "Created By", FieldType: "hidden", FieldOrder: 8, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$me}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "order_line_items", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 9, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "order_line_items", Name: "updated_by", Label: "Updated By", FieldType: "hidden", FieldOrder: 10, Required: false, Config: json.RawMessage(`{"default_value": "{{$me}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "sales", EntityTable: "order_line_items", Name: "updated_date", Label: "Updated Date", FieldType: "hidden", FieldOrder: 11, Required: false, Config: json.RawMessage(`{"default_value": "{{$now}}", "hidden": true}`)},
	}
}

// GetSupplierFormFields returns form fields for suppliers (procurement.suppliers)
func GetSupplierFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		// ID field - required for update operations to work correctly
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "suppliers", Name: "id", Label: "ID", FieldType: "hidden", FieldOrder: 0, Required: false, Config: json.RawMessage(`{"hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "suppliers", Name: "name", Label: "Supplier Name", FieldType: "text", FieldOrder: 1, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "suppliers", Name: "contact_infos_id", Label: "Contact Information", FieldType: "smart-combobox", FieldOrder: 2, Required: true, Config: json.RawMessage(`{"entity": "core.contact_infos", "display_field": "email_address", "inline_create": {"enabled": true, "form_name": "Contact Info Creation Form", "button_text": "Create Contact Info"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "suppliers", Name: "payment_term_id", Label: "Payment Terms", FieldType: "smart-combobox", FieldOrder: 3, Required: false, Config: json.RawMessage(`{"entity": "core.payment_terms", "display_field": "name"}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "suppliers", Name: "lead_time_days", Label: "Lead Time (Days)", FieldType: "number", FieldOrder: 4, Required: true, Config: json.RawMessage(`{"min": 0, "max": 730}`)},
		// Audit field - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "suppliers", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 5, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
	}
}

// GetPurchaseOrderFormFields returns form fields for purchase orders (procurement.purchase_orders)
func GetPurchaseOrderFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		// ID field - required for update operations to work correctly
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "purchase_orders", Name: "id", Label: "ID", FieldType: "hidden", FieldOrder: 0, Required: false, Config: json.RawMessage(`{"hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "purchase_orders", Name: "order_number", Label: "Order Number", FieldType: "text", FieldOrder: 1, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "purchase_orders", Name: "supplier_id", Label: "Supplier", FieldType: "smart-combobox", FieldOrder: 2, Required: true, Config: json.RawMessage(`{"entity": "procurement.suppliers", "display_field": "name", "value_column": "id", "inline_create": {"enabled": true, "form_name": "Supplier Creation Form", "button_text": "Create Supplier"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "purchase_orders", Name: "purchase_order_status_id", Label: "Status", FieldType: "smart-combobox", FieldOrder: 3, Required: true, Config: json.RawMessage(`{"entity": "procurement.purchase_order_statuses", "label_column": "name", "value_column": "id", "default_value_create": "DRAFT", "inline_create": {"enabled": true, "form_name": "Purchase Order Status Creation Form", "button_text": "Create Purchase Order Status"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "purchase_orders", Name: "delivery_warehouse_id", Label: "Delivery Warehouse", FieldType: "smart-combobox", FieldOrder: 4, Required: true, Config: json.RawMessage(`{"entity": "inventory.warehouses", "display_field": "name", "inline_create": {"enabled": true, "form_name": "Warehouse Creation Form", "button_text": "Create Warehouse"}}`)},
		// Audit fields - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "purchase_orders", Name: "created_by", Label: "Created By", FieldType: "hidden", FieldOrder: 5, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$me}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "purchase_orders", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 6, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "purchase_orders", Name: "updated_by", Label: "Updated By", FieldType: "hidden", FieldOrder: 7, Required: false, Config: json.RawMessage(`{"default_value": "{{$me}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "purchase_orders", Name: "updated_date", Label: "Updated Date", FieldType: "hidden", FieldOrder: 8, Required: false, Config: json.RawMessage(`{"default_value": "{{$now}}", "hidden": true}`)},
	}
}

// GetPurchaseOrderLineItemFormFields returns form fields for PO line items (procurement.purchase_order_line_items)
func GetPurchaseOrderLineItemFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		// ID field - required for update operations to work correctly
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "purchase_order_line_items", Name: "id", Label: "ID", FieldType: "hidden", FieldOrder: 0, Required: false, Config: json.RawMessage(`{"hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "purchase_order_line_items", Name: "supplier_product_id", Label: "Supplier Product", FieldType: "smart-combobox", FieldOrder: 1, Required: true, Config: json.RawMessage(`{"entity": "procurement.supplier_products", "display_field": "supplier_part_number", "value_column": "id", "inline_create": {"enabled": true, "form_name": "Supplier Product Creation Form", "button_text": "Create Supplier Product"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "purchase_order_line_items", Name: "quantity_ordered", Label: "Quantity Ordered", FieldType: "number", FieldOrder: 2, Required: true, Config: json.RawMessage(`{"min": 1, "max": 1000000}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "purchase_order_line_items", Name: "unit_cost", Label: "Unit Cost", FieldType: "number", FieldOrder: 3, Required: true, Config: json.RawMessage(`{"min": 0, "max": 10000000}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "purchase_order_line_items", Name: "line_item_status_id", Label: "Line Item Status", FieldType: "smart-combobox", FieldOrder: 4, Required: true, Config: json.RawMessage(`{"entity": "procurement.purchase_order_line_item_statuses", "label_column": "name", "value_column": "id", "default_value_create": "PENDING", "inline_create": {"enabled": true, "form_name": "Purchase Order Line Item Status Creation Form", "button_text": "Create Purchase Order Line Item Status"}}`)},
		// Audit fields - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "purchase_order_line_items", Name: "created_by", Label: "Created By", FieldType: "hidden", FieldOrder: 5, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$me}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "purchase_order_line_items", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 6, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "purchase_order_line_items", Name: "updated_by", Label: "Updated By", FieldType: "hidden", FieldOrder: 7, Required: false, Config: json.RawMessage(`{"default_value": "{{$me}}", "hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "purchase_order_line_items", Name: "updated_date", Label: "Updated Date", FieldType: "hidden", FieldOrder: 8, Required: false, Config: json.RawMessage(`{"default_value": "{{$now}}", "hidden": true}`)},
	}
}

// GetRoleFormFields returns form fields for roles (core.roles)
func GetRoleFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "roles", Name: "name", Label: "Role Name", FieldType: "text", FieldOrder: 1, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "roles", Name: "description", Label: "Description", FieldType: "textarea", FieldOrder: 2, Required: false, Config: json.RawMessage(`{}`)},
	}
}

// GetUserFormFields returns form fields for users (core.users)
func GetUserFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		// ID field - required for update operations to work correctly
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "users", Name: "id", Label: "ID", FieldType: "hidden", FieldOrder: 0, Required: false, Config: json.RawMessage(`{"hidden": true}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "users", Name: "username", Label: "Username", FieldType: "text", FieldOrder: 1, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "users", Name: "first_name", Label: "First Name", FieldType: "text", FieldOrder: 2, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "users", Name: "last_name", Label: "Last Name", FieldType: "text", FieldOrder: 3, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "users", Name: "email", Label: "Email", FieldType: "email", FieldOrder: 4, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "users", Name: "title_id", Label: "Title", FieldType: "smart-combobox", FieldOrder: 5, Required: false, Config: json.RawMessage(`{"entity": "hr.titles", "display_field": "name", "inline_create": {"enabled": true, "form_name": "Title Creation Form", "button_text": "Create Title"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "users", Name: "office_id", Label: "Office", FieldType: "smart-combobox", FieldOrder: 6, Required: false, Config: json.RawMessage(`{"entity": "hr.offices", "display_field": "name", "inline_create": {"enabled": true, "form_name": "Office Creation Form", "button_text": "Create Office"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "users", Name: "user_approval_status_id", Label: "Approval Status", FieldType: "smart-combobox", FieldOrder: 7, Required: true, Config: json.RawMessage(`{"entity": "hr.user_approval_status", "display_field": "name", "inline_create": {"enabled": true, "form_name": "User Approval Status Creation Form", "button_text": "Create User Approval Status"}}`)},
		// Audit field - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "core", EntityTable: "users", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 8, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
	}
}

// GetSupplierProductFormFields returns form fields for supplier products (procurement.supplier_products)
func GetSupplierProductFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "supplier_products", Name: "supplier_id", Label: "Supplier", FieldType: "smart-combobox", FieldOrder: 1, Required: true, Config: json.RawMessage(`{"entity": "procurement.suppliers", "display_field": "name", "value_column": "id", "inline_create": {"enabled": true, "form_name": "Supplier Creation Form", "button_text": "Create Supplier"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "supplier_products", Name: "product_id", Label: "Product", FieldType: "smart-combobox", FieldOrder: 2, Required: true, Config: json.RawMessage(`{"entity": "products.products", "display_field": "name", "inline_create": {"enabled": true, "form_name": "Product Creation Form", "button_text": "Create Product"}}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "supplier_products", Name: "supplier_part_number", Label: "Supplier Part Number", FieldType: "text", FieldOrder: 3, Required: true, Config: json.RawMessage(`{}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "supplier_products", Name: "min_order_quantity", Label: "Minimum Order Quantity", FieldType: "number", FieldOrder: 4, Required: true, Config: json.RawMessage(`{"min": 1, "max": 1000000}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "supplier_products", Name: "max_order_quantity", Label: "Maximum Order Quantity", FieldType: "number", FieldOrder: 5, Required: true, Config: json.RawMessage(`{"min": 1, "max": 1000000}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "supplier_products", Name: "lead_time_days", Label: "Lead Time (Days)", FieldType: "number", FieldOrder: 6, Required: true, Config: json.RawMessage(`{"min": 0, "max": 730}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "supplier_products", Name: "unit_cost", Label: "Unit Cost", FieldType: "number", FieldOrder: 7, Required: true, Config: json.RawMessage(`{"min": 0, "max": 10000000, "step": 0.01}`)},
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "supplier_products", Name: "is_primary_supplier", Label: "Is Primary Supplier", FieldType: "boolean", FieldOrder: 8, Required: true, Config: json.RawMessage(`{}`)},
		// Audit field - auto-populated by backend
		{FormID: formID, EntityID: entityID, EntitySchema: "procurement", EntityTable: "supplier_products", Name: "created_date", Label: "Created Date", FieldType: "hidden", FieldOrder: 9, Required: false, Config: json.RawMessage(`{"default_value_create": "{{$now}}", "hidden": true}`)},
	}
}
