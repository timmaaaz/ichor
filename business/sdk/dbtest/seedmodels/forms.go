package seedmodels

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
)

// =============================================================================
// FORM FIELD GENERATORS
// =============================================================================
// These functions generate form field configurations for seeding the database.
// Each function corresponds to a table/entity and returns the fields needed
// to create records in that table through forms.

// GetCustomerFormFields returns form fields for creating customers (sales.customers)
func GetCustomerFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "name",
			Label:      "Customer Name",
			FieldType:  "text",
			FieldOrder: 1,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "contact_id",
			Label:      "Contact Information",
			FieldType:  "select",
			FieldOrder: 2,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "contact_infos", "display_field": "email_address"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "delivery_address_id",
			Label:      "Delivery Address",
			FieldType:  "select",
			FieldOrder: 3,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "streets", "display_field": "line_1"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "notes",
			Label:      "Notes",
			FieldType:  "textarea",
			FieldOrder: 4,
			Required:   false,
			Config:     json.RawMessage(`{}`),
		},
	}
}

// GetSalesOrderFormFields returns form fields for creating sales orders (sales.orders)
func GetSalesOrderFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "number",
			Label:      "Order Number",
			FieldType:  "text",
			FieldOrder: 1,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "customer_id",
			Label:      "Customer",
			FieldType:  "select",
			FieldOrder: 2,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "customers", "display_field": "name"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "due_date",
			Label:      "Due Date",
			FieldType:  "date",
			FieldOrder: 3,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "order_fulfillment_status_id",
			Label:      "Fulfillment Status",
			FieldType:  "select",
			FieldOrder: 4,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "order_fulfillment_statuses", "display_field": "name"}`),
		},
	}
}

// GetRoleFormFields returns form fields for creating roles (core.roles)
func GetRoleFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "name",
			Label:      "Role Name",
			FieldType:  "text",
			FieldOrder: 1,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "description",
			Label:      "Description",
			FieldType:  "textarea",
			FieldOrder: 2,
			Required:   false,
			Config:     json.RawMessage(`{}`),
		},
	}
}

// GetUserFormFields returns form fields for creating users (core.users)
func GetUserFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "username",
			Label:      "Username",
			FieldType:  "text",
			FieldOrder: 1,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "first_name",
			Label:      "First Name",
			FieldType:  "text",
			FieldOrder: 2,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "last_name",
			Label:      "Last Name",
			FieldType:  "text",
			FieldOrder: 3,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "email",
			Label:      "Email",
			FieldType:  "email",
			FieldOrder: 4,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "password",
			Label:      "Password",
			FieldType:  "password",
			FieldOrder: 5,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "password_confirm",
			Label:      "Confirm Password",
			FieldType:  "password",
			FieldOrder: 6,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "birthday",
			Label:      "Birthday",
			FieldType:  "date",
			FieldOrder: 7,
			Required:   false,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "title_id",
			Label:      "Job Title",
			FieldType:  "select",
			FieldOrder: 8,
			Required:   false,
			Config:     json.RawMessage(`{"entity": "titles", "display_field": "name"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "office_id",
			Label:      "Office",
			FieldType:  "select",
			FieldOrder: 9,
			Required:   false,
			Config:     json.RawMessage(`{"entity": "offices", "display_field": "name"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "roles",
			Label:      "Roles",
			FieldType:  "multiselect",
			FieldOrder: 10,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "system_roles",
			Label:      "System Roles",
			FieldType:  "multiselect",
			FieldOrder: 11,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "enabled",
			Label:      "Enabled",
			FieldType:  "boolean",
			FieldOrder: 12,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "requested_by",
			Label:      "Requested By",
			FieldType:  "select",
			FieldOrder: 13,
			Required:   false,
			Config:     json.RawMessage(`{"entity": "users", "display_field": "username"}`),
		},
	}
}

// GetSupplierFormFields returns form fields for creating suppliers (procurement.suppliers)
func GetSupplierFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "name",
			Label:      "Supplier Name",
			FieldType:  "text",
			FieldOrder: 1,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "contact_infos_id",
			Label:      "Contact Information",
			FieldType:  "select",
			FieldOrder: 2,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "contact_infos", "display_field": "email_address"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "payment_terms",
			Label:      "Payment Terms",
			FieldType:  "textarea",
			FieldOrder: 3,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "lead_time_days",
			Label:      "Lead Time (Days)",
			FieldType:  "number",
			FieldOrder: 4,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "rating",
			Label:      "Rating",
			FieldType:  "number",
			FieldOrder: 5,
			Required:   true,
			Config:     json.RawMessage(`{"min": 0, "max": 5, "step": 0.1}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "is_active",
			Label:      "Active",
			FieldType:  "boolean",
			FieldOrder: 6,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
	}
}

// GetPurchaseOrderFormFields returns form fields for creating purchase orders (procurement.purchase_orders)
func GetPurchaseOrderFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "order_number",
			Label:      "Order Number",
			FieldType:  "text",
			FieldOrder: 1,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "supplier_id",
			Label:      "Supplier",
			FieldType:  "select",
			FieldOrder: 2,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "suppliers", "display_field": "name"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "purchase_order_status_id",
			Label:      "Status",
			FieldType:  "select",
			FieldOrder: 3,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "purchase_order_statuses", "display_field": "name"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "delivery_warehouse_id",
			Label:      "Delivery Warehouse",
			FieldType:  "select",
			FieldOrder: 4,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "warehouses", "display_field": "name"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "delivery_location_id",
			Label:      "Delivery Location",
			FieldType:  "select",
			FieldOrder: 5,
			Required:   false,
			Config:     json.RawMessage(`{"entity": "inventory_locations", "display_field": "aisle"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "delivery_street_id",
			Label:      "Delivery Street Address",
			FieldType:  "select",
			FieldOrder: 6,
			Required:   false,
			Config:     json.RawMessage(`{"entity": "streets", "display_field": "line_1"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "order_date",
			Label:      "Order Date",
			FieldType:  "date",
			FieldOrder: 7,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "expected_delivery_date",
			Label:      "Expected Delivery Date",
			FieldType:  "date",
			FieldOrder: 8,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "subtotal",
			Label:      "Subtotal",
			FieldType:  "number",
			FieldOrder: 9,
			Required:   true,
			Config:     json.RawMessage(`{"min": 0, "step": 0.01}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "tax_amount",
			Label:      "Tax Amount",
			FieldType:  "number",
			FieldOrder: 10,
			Required:   true,
			Config:     json.RawMessage(`{"min": 0, "step": 0.01}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "shipping_cost",
			Label:      "Shipping Cost",
			FieldType:  "number",
			FieldOrder: 11,
			Required:   true,
			Config:     json.RawMessage(`{"min": 0, "step": 0.01}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "currency",
			Label:      "Currency",
			FieldType:  "text",
			FieldOrder: 12,
			Required:   true,
			Config:     json.RawMessage(`{"default": "USD"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "notes",
			Label:      "Notes",
			FieldType:  "textarea",
			FieldOrder: 13,
			Required:   false,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "supplier_reference_number",
			Label:      "Supplier Reference Number",
			FieldType:  "text",
			FieldOrder: 14,
			Required:   false,
			Config:     json.RawMessage(`{}`),
		},
	}
}

// GetWarehouseFormFields returns form fields for creating warehouses (inventory.warehouses)
func GetWarehouseFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "code",
			Label:      "Warehouse Code",
			FieldType:  "text",
			FieldOrder: 1,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "name",
			Label:      "Warehouse Name",
			FieldType:  "text",
			FieldOrder: 2,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "street_id",
			Label:      "Address",
			FieldType:  "select",
			FieldOrder: 3,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "streets", "display_field": "line_1"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "is_active",
			Label:      "Active",
			FieldType:  "boolean",
			FieldOrder: 4,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
	}
}

// GetInventoryAdjustmentFormFields returns form fields for creating inventory adjustments (inventory.inventory_adjustments)
func GetInventoryAdjustmentFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "product_id",
			Label:      "Product",
			FieldType:  "select",
			FieldOrder: 1,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "products", "display_field": "name"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "location_id",
			Label:      "Location",
			FieldType:  "select",
			FieldOrder: 2,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "inventory_locations", "display_field": "aisle"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "quantity_change",
			Label:      "Quantity Change",
			FieldType:  "number",
			FieldOrder: 3,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "reason_code",
			Label:      "Reason Code",
			FieldType:  "text",
			FieldOrder: 4,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "notes",
			Label:      "Notes",
			FieldType:  "textarea",
			FieldOrder: 5,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "adjustment_date",
			Label:      "Adjustment Date",
			FieldType:  "date",
			FieldOrder: 6,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "approved_by",
			Label:      "Approved By",
			FieldType:  "select",
			FieldOrder: 7,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "users", "display_field": "username"}`),
		},
	}
}

// GetTransferOrderFormFields returns form fields for creating transfer orders (inventory.transfer_orders)
func GetTransferOrderFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "product_id",
			Label:      "Product",
			FieldType:  "select",
			FieldOrder: 1,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "products", "display_field": "name"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "from_location_id",
			Label:      "From Location",
			FieldType:  "select",
			FieldOrder: 2,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "inventory_locations", "display_field": "aisle"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "to_location_id",
			Label:      "To Location",
			FieldType:  "select",
			FieldOrder: 3,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "inventory_locations", "display_field": "aisle"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "quantity",
			Label:      "Quantity",
			FieldType:  "number",
			FieldOrder: 4,
			Required:   true,
			Config:     json.RawMessage(`{"min": 1}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "status",
			Label:      "Status",
			FieldType:  "text",
			FieldOrder: 5,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "transfer_date",
			Label:      "Transfer Date",
			FieldType:  "date",
			FieldOrder: 6,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "approved_by",
			Label:      "Approved By",
			FieldType:  "select",
			FieldOrder: 7,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "users", "display_field": "username"}`),
		},
	}
}

// GetInventoryItemFormFields returns form fields for creating inventory items (inventory.inventory_items)
func GetInventoryItemFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "product_id",
			Label:      "Product",
			FieldType:  "select",
			FieldOrder: 1,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "products", "display_field": "name"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "location_id",
			Label:      "Location",
			FieldType:  "select",
			FieldOrder: 2,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "inventory_locations", "display_field": "aisle"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "quantity",
			Label:      "Quantity",
			FieldType:  "number",
			FieldOrder: 3,
			Required:   true,
			Config:     json.RawMessage(`{"min": 0}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "reserved_quantity",
			Label:      "Reserved Quantity",
			FieldType:  "number",
			FieldOrder: 4,
			Required:   true,
			Config:     json.RawMessage(`{"min": 0}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "allocated_quantity",
			Label:      "Allocated Quantity",
			FieldType:  "number",
			FieldOrder: 5,
			Required:   true,
			Config:     json.RawMessage(`{"min": 0}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "minimum_stock",
			Label:      "Minimum Stock",
			FieldType:  "number",
			FieldOrder: 6,
			Required:   true,
			Config:     json.RawMessage(`{"min": 0}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "maximum_stock",
			Label:      "Maximum Stock",
			FieldType:  "number",
			FieldOrder: 7,
			Required:   true,
			Config:     json.RawMessage(`{"min": 0}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "reorder_point",
			Label:      "Reorder Point",
			FieldType:  "number",
			FieldOrder: 8,
			Required:   true,
			Config:     json.RawMessage(`{"min": 0}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "economic_order_quantity",
			Label:      "Economic Order Quantity",
			FieldType:  "number",
			FieldOrder: 9,
			Required:   true,
			Config:     json.RawMessage(`{"min": 0}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "safety_stock",
			Label:      "Safety Stock",
			FieldType:  "number",
			FieldOrder: 10,
			Required:   true,
			Config:     json.RawMessage(`{"min": 0}`),
		},
	}
}

// GetOfficeFormFields returns form fields for creating offices (hr.offices)
func GetOfficeFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "name",
			Label:      "Office Name",
			FieldType:  "text",
			FieldOrder: 1,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "street_id",
			Label:      "Address",
			FieldType:  "select",
			FieldOrder: 2,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "streets", "display_field": "line_1"}`),
		},
	}
}

// GetAssetFormFields returns form fields for creating assets (assets.assets)
func GetAssetFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "valid_asset_id",
			Label:      "Asset Type",
			FieldType:  "select",
			FieldOrder: 1,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "valid_assets", "display_field": "name"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "serial_number",
			Label:      "Serial Number",
			FieldType:  "text",
			FieldOrder: 2,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "asset_condition_id",
			Label:      "Condition",
			FieldType:  "select",
			FieldOrder: 3,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "asset_conditions", "display_field": "name"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "last_maintenance_time",
			Label:      "Last Maintenance",
			FieldType:  "date",
			FieldOrder: 4,
			Required:   false,
			Config:     json.RawMessage(`{}`),
		},
	}
}

// =============================================================================
// DEPENDENCY ENTITY FORM FIELDS (Building blocks for composite forms)
// =============================================================================

// GetContactInfoFormFields returns form fields for creating contact info (core.contact_infos)
func GetContactInfoFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "first_name",
			Label:      "Contact First Name",
			FieldType:  "text",
			FieldOrder: 1,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "last_name",
			Label:      "Contact Last Name",
			FieldType:  "text",
			FieldOrder: 2,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "primary_phone_number",
			Label:      "Primary Phone",
			FieldType:  "tel",
			FieldOrder: 3,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "secondary_phone_number",
			Label:      "Secondary Phone",
			FieldType:  "tel",
			FieldOrder: 4,
			Required:   false,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "email_address",
			Label:      "Email Address",
			FieldType:  "email",
			FieldOrder: 5,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "street_id",
			Label:      "Contact Address",
			FieldType:  "select",
			FieldOrder: 6,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "streets", "display_field": "line_1"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "available_hours_start",
			Label:      "Available From",
			FieldType:  "time",
			FieldOrder: 7,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "available_hours_end",
			Label:      "Available Until",
			FieldType:  "time",
			FieldOrder: 8,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "timezone",
			Label:      "Timezone",
			FieldType:  "text",
			FieldOrder: 9,
			Required:   true,
			Config:     json.RawMessage(`{"default": "America/New_York"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "preferred_contact_type",
			Label:      "Preferred Contact Method",
			FieldType:  "select",
			FieldOrder: 10,
			Required:   true,
			Config:     json.RawMessage(`{"options": ["phone", "email", "mail", "fax"]}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "notes",
			Label:      "Contact Notes",
			FieldType:  "textarea",
			FieldOrder: 11,
			Required:   false,
			Config:     json.RawMessage(`{}`),
		},
	}
}

// GetStreetFormFields returns form fields for creating street addresses (geography.streets)
func GetStreetFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "line_1",
			Label:      "Street Address",
			FieldType:  "text",
			FieldOrder: 1,
			Required:   true,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "line_2",
			Label:      "Address Line 2",
			FieldType:  "text",
			FieldOrder: 2,
			Required:   false,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "postal_code",
			Label:      "Postal Code",
			FieldType:  "text",
			FieldOrder: 3,
			Required:   false,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "city_id",
			Label:      "City",
			FieldType:  "select",
			FieldOrder: 4,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "cities", "display_field": "name"}`),
		},
	}
}

// GetSalesOrderLineItemFormFields returns form fields for creating order line items (sales.order_line_items)
func GetSalesOrderLineItemFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "product_id",
			Label:      "Product",
			FieldType:  "select",
			FieldOrder: 1,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "products", "display_field": "name"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "quantity",
			Label:      "Quantity",
			FieldType:  "number",
			FieldOrder: 2,
			Required:   true,
			Config:     json.RawMessage(`{"min": 1}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "discount",
			Label:      "Discount",
			FieldType:  "number",
			FieldOrder: 3,
			Required:   false,
			Config:     json.RawMessage(`{"min": 0, "step": 0.01}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "line_item_fulfillment_statuses_id",
			Label:      "Fulfillment Status",
			FieldType:  "select",
			FieldOrder: 4,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "line_item_fulfillment_statuses", "display_field": "name"}`),
		},
	}
}

// GetPurchaseOrderLineItemFormFields returns form fields for creating PO line items (procurement.purchase_order_line_items)
func GetPurchaseOrderLineItemFormFields(formID uuid.UUID, entityID uuid.UUID) []formfieldbus.NewFormField {
	return []formfieldbus.NewFormField{
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "supplier_product_id",
			Label:      "Supplier Product",
			FieldType:  "select",
			FieldOrder: 1,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "supplier_products", "display_field": "supplier_part_number"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "quantity_ordered",
			Label:      "Quantity Ordered",
			FieldType:  "number",
			FieldOrder: 2,
			Required:   true,
			Config:     json.RawMessage(`{"min": 1}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "unit_cost",
			Label:      "Unit Cost",
			FieldType:  "number",
			FieldOrder: 3,
			Required:   true,
			Config:     json.RawMessage(`{"min": 0, "step": 0.01}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "discount",
			Label:      "Discount",
			FieldType:  "number",
			FieldOrder: 4,
			Required:   false,
			Config:     json.RawMessage(`{"min": 0, "step": 0.01, "default": 0.00}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "line_item_status_id",
			Label:      "Line Item Status",
			FieldType:  "select",
			FieldOrder: 5,
			Required:   true,
			Config:     json.RawMessage(`{"entity": "purchase_order_line_item_statuses", "display_field": "name"}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "expected_delivery_date",
			Label:      "Expected Delivery",
			FieldType:  "date",
			FieldOrder: 6,
			Required:   false,
			Config:     json.RawMessage(`{}`),
		},
		{
			FormID:     formID,
			EntityID:   entityID,
			Name:       "notes",
			Label:      "Line Item Notes",
			FieldType:  "textarea",
			FieldOrder: 7,
			Required:   false,
			Config:     json.RawMessage(`{}`),
		},
	}
}

// =============================================================================
// COMPOSITE FORM FIELDS (Multi-entity forms)
// =============================================================================

// GetFullCustomerFormFields returns a composite form for creating a customer
// along with contact info and delivery address in a single transaction.
// Execution order: 1) Contact Info, 2) Delivery Address, 3) Customer
func GetFullCustomerFormFields(
	formID uuid.UUID,
	customerEntityID uuid.UUID,
	contactEntityID uuid.UUID,
	streetEntityID uuid.UUID,
) []formfieldbus.NewFormField {
	var fields []formfieldbus.NewFormField
	order := 1

	// Section 1: Customer basic info (will be created last but shown first)
	fields = append(fields, formfieldbus.NewFormField{
		FormID:     formID,
		EntityID:   customerEntityID,
		Name:       "name",
		Label:      "Customer Name",
		FieldType:  "text",
		FieldOrder: order,
		Required:   true,
		Config:     json.RawMessage(`{"execution_order": 3}`),
	})
	order++

	// Section 2: Contact Information (created first, execution_order: 1)
	contactFields := GetContactInfoFormFields(formID, contactEntityID)
	for i := range contactFields {
		contactFields[i].FieldOrder = order
		// Mark that this creates the contact_id for customer
		contactFields[i].Config = json.RawMessage(`{"execution_order": 1, "parent_entity": "customers", "parent_field": "contact_id"}`)
		order++
	}
	fields = append(fields, contactFields...)

	// Section 3: Delivery Address (created second, execution_order: 2)
	streetFields := GetStreetFormFields(formID, streetEntityID)
	for i := range streetFields {
		streetFields[i].FieldOrder = order
		streetFields[i].Label = "Delivery " + streetFields[i].Label // Prefix with "Delivery"
		streetFields[i].Config = json.RawMessage(`{"execution_order": 2, "parent_entity": "customers", "parent_field": "delivery_address_id"}`)
		order++
	}
	fields = append(fields, streetFields...)

	// Section 4: Customer notes (part of customer entity, execution_order: 3)
	fields = append(fields, formfieldbus.NewFormField{
		FormID:     formID,
		EntityID:   customerEntityID,
		Name:       "notes",
		Label:      "Customer Notes",
		FieldType:  "textarea",
		FieldOrder: order,
		Required:   false,
		Config:     json.RawMessage(`{"execution_order": 3}`),
	})

	return fields
}

// GetFullSupplierFormFields returns a composite form for creating a supplier
// along with contact info in a single transaction.
// Execution order: 1) Contact Info, 2) Supplier
func GetFullSupplierFormFields(
	formID uuid.UUID,
	supplierEntityID uuid.UUID,
	contactEntityID uuid.UUID,
) []formfieldbus.NewFormField {
	var fields []formfieldbus.NewFormField
	order := 1

	// Section 1: Supplier basic info
	fields = append(fields, formfieldbus.NewFormField{
		FormID:     formID,
		EntityID:   supplierEntityID,
		Name:       "name",
		Label:      "Supplier Name",
		FieldType:  "text",
		FieldOrder: order,
		Required:   true,
		Config:     json.RawMessage(`{"execution_order": 2}`),
	})
	order++

	// Section 2: Contact Information (created first)
	contactFields := GetContactInfoFormFields(formID, contactEntityID)
	for i := range contactFields {
		contactFields[i].FieldOrder = order
		contactFields[i].Config = json.RawMessage(`{"execution_order": 1, "parent_entity": "suppliers", "parent_field": "contact_infos_id"}`)
		order++
	}
	fields = append(fields, contactFields...)

	// Section 3: Supplier details (execution_order: 2)
	supplierDetailFields := []formfieldbus.NewFormField{
		{
			FormID:     formID,
			EntityID:   supplierEntityID,
			Name:       "payment_terms",
			Label:      "Payment Terms",
			FieldType:  "textarea",
			FieldOrder: order,
			Required:   true,
			Config:     json.RawMessage(`{"execution_order": 2}`),
		},
		{
			FormID:     formID,
			EntityID:   supplierEntityID,
			Name:       "lead_time_days",
			Label:      "Lead Time (Days)",
			FieldType:  "number",
			FieldOrder: order + 1,
			Required:   true,
			Config:     json.RawMessage(`{"execution_order": 2}`),
		},
		{
			FormID:     formID,
			EntityID:   supplierEntityID,
			Name:       "rating",
			Label:      "Rating",
			FieldType:  "number",
			FieldOrder: order + 2,
			Required:   true,
			Config:     json.RawMessage(`{"execution_order": 2, "min": 0, "max": 5, "step": 0.1}`),
		},
		{
			FormID:     formID,
			EntityID:   supplierEntityID,
			Name:       "is_active",
			Label:      "Active",
			FieldType:  "boolean",
			FieldOrder: order + 3,
			Required:   true,
			Config:     json.RawMessage(`{"execution_order": 2}`),
		},
	}
	fields = append(fields, supplierDetailFields...)

	return fields
}

// GetFullSalesOrderFormFields returns a composite form for creating a sales order
// along with line items in a single transaction.
// Execution order: 1) Order, 2) Line Items
func GetFullSalesOrderFormFields(
	formID uuid.UUID,
	orderEntityID uuid.UUID,
	lineItemEntityID uuid.UUID,
) []formfieldbus.NewFormField {
	var fields []formfieldbus.NewFormField
	order := 1

	// Section 1: Order header fields (created first)
	orderFields := GetSalesOrderFormFields(formID, orderEntityID)
	for i := range orderFields {
		orderFields[i].FieldOrder = order
		orderFields[i].Config = json.RawMessage(`{"execution_order": 1}`)
		order++
	}
	fields = append(fields, orderFields...)

	// Section 2: Line Items (created second, references order_id from step 1)
	// Note: In a real form, you'd have a repeatable section for multiple line items
	// This shows the structure for one line item
	lineItemFields := GetSalesOrderLineItemFormFields(formID, lineItemEntityID)
	for i := range lineItemFields {
		lineItemFields[i].FieldOrder = order
		lineItemFields[i].Config = json.RawMessage(`{"execution_order": 2, "parent_entity": "orders", "parent_field": "order_id", "repeatable": true}`)
		order++
	}
	fields = append(fields, lineItemFields...)

	return fields
}

// GetFullPurchaseOrderFormFields returns a composite form for creating a purchase order
// along with line items in a single transaction.
// Execution order: 1) Purchase Order, 2) Line Items
func GetFullPurchaseOrderFormFields(
	formID uuid.UUID,
	poEntityID uuid.UUID,
	lineItemEntityID uuid.UUID,
) []formfieldbus.NewFormField {
	var fields []formfieldbus.NewFormField
	order := 1

	// Section 1: PO header fields (created first)
	poFields := GetPurchaseOrderFormFields(formID, poEntityID)
	for i := range poFields {
		poFields[i].FieldOrder = order
		poFields[i].Config = json.RawMessage(`{"execution_order": 1}`)
		order++
	}
	fields = append(fields, poFields...)

	// Section 2: Line Items (created second, references purchase_order_id from step 1)
	lineItemFields := GetPurchaseOrderLineItemFormFields(formID, lineItemEntityID)
	for i := range lineItemFields {
		lineItemFields[i].FieldOrder = order
		lineItemFields[i].Config = json.RawMessage(`{"execution_order": 2, "parent_entity": "purchase_orders", "parent_field": "purchase_order_id", "repeatable": true}`)
		order++
	}
	fields = append(fields, lineItemFields...)

	return fields
}
