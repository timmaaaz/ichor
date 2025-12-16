package seedmodels

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
)

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// mergeConfigJSON merges execution metadata into existing field Config JSON.
// This preserves important FK field configuration (entity, display_field, inline_create)
// while adding execution_order and parent relationship metadata.
func mergeConfigJSON(existingConfig json.RawMessage, executionMetadata map[string]interface{}) json.RawMessage {
	if len(existingConfig) == 0 {
		// No existing config, just serialize the execution metadata
		if merged, err := json.Marshal(executionMetadata); err == nil {
			return merged
		}
		return json.RawMessage("{}")
	}

	// Parse existing config
	var existing map[string]interface{}
	if err := json.Unmarshal(existingConfig, &existing); err != nil {
		// If unmarshal fails, just return execution metadata
		if merged, err := json.Marshal(executionMetadata); err == nil {
			return merged
		}
		return existingConfig
	}

	// Merge execution metadata into existing config
	for key, value := range executionMetadata {
		existing[key] = value
	}

	// Serialize merged config
	if merged, err := json.Marshal(existing); err == nil {
		return merged
	}

	// Fallback to original config if marshal fails
	return existingConfig
}

// =============================================================================
// COMPOSITE FORM FIELD GENERATORS
// =============================================================================
// These functions generate complex multi-entity forms that combine data from
// multiple tables in a single transaction. They reference the simple table
// form generators from tableforms.go.

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
		FormID:       formID,
		EntityID:     customerEntityID,
		EntitySchema: "sales",
		EntityTable:  "customers",
		Name:         "name",
		Label:        "Customer Name",
		FieldType:    "text",
		FieldOrder:   order,
		Required:     true,
		Config:       json.RawMessage(`{"execution_order": 3}`),
	})
	order++

	// Section 2: Contact Information (created first, execution_order: 1)
	contactFields := GetContactInfoFormFields(formID, contactEntityID)
	for i := range contactFields {
		contactFields[i].FieldOrder = order
		// Mark that this creates the contact_id for customer
		contactFields[i].Config = mergeConfigJSON(contactFields[i].Config, map[string]interface{}{
			"execution_order": 1,
			"parent_entity":   "customers",
			"parent_field":    "contact_id",
		})
		order++
	}
	fields = append(fields, contactFields...)

	// Section 3: Delivery Address (created second, execution_order: 2)
	streetFields := GetStreetFormFields(formID, streetEntityID)
	for i := range streetFields {
		streetFields[i].FieldOrder = order
		streetFields[i].Label = "Delivery " + streetFields[i].Label // Prefix with "Delivery"
		streetFields[i].Config = mergeConfigJSON(streetFields[i].Config, map[string]interface{}{
			"execution_order": 2,
			"parent_entity":   "customers",
			"parent_field":    "delivery_address_id",
		})
		order++
	}
	fields = append(fields, streetFields...)

	// Section 4: Customer notes (part of customer entity, execution_order: 3)
	fields = append(fields, formfieldbus.NewFormField{
		FormID:       formID,
		EntityID:     customerEntityID,
		EntitySchema: "sales",
		EntityTable:  "customers",
		Name:         "notes",
		Label:        "Customer Notes",
		FieldType:    "textarea",
		FieldOrder:   order,
		Required:     false,
		Config:       json.RawMessage(`{"execution_order": 3}`),
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
		FormID:       formID,
		EntityID:     supplierEntityID,
		EntitySchema: "procurement",
		EntityTable:  "suppliers",
		Name:         "name",
		Label:        "Supplier Name",
		FieldType:    "text",
		FieldOrder:   order,
		Required:     true,
		Config:       json.RawMessage(`{"execution_order": 2}`),
	})
	order++

	// Section 2: Contact Information (created first)
	contactFields := GetContactInfoFormFields(formID, contactEntityID)
	for i := range contactFields {
		contactFields[i].FieldOrder = order
		contactFields[i].Config = mergeConfigJSON(contactFields[i].Config, map[string]interface{}{
			"execution_order": 1,
			"parent_entity":   "suppliers",
			"parent_field":    "contact_infos_id",
		})
		order++
	}
	fields = append(fields, contactFields...)

	// Section 3: Supplier details (execution_order: 2)
	supplierDetailFields := []formfieldbus.NewFormField{
		{
			FormID:       formID,
			EntityID:     supplierEntityID,
			EntitySchema: "procurement",
			EntityTable:  "suppliers",
			Name:         "payment_terms",
			Label:        "Payment Terms",
			FieldType:    "textarea",
			FieldOrder:   order,
			Required:     true,
			Config:       json.RawMessage(`{"execution_order": 2}`),
		},
		{
			FormID:       formID,
			EntityID:     supplierEntityID,
			EntitySchema: "procurement",
			EntityTable:  "suppliers",
			Name:         "lead_time_days",
			Label:        "Lead Time (Days)",
			FieldType:    "number",
			FieldOrder:   order + 1,
			Required:     true,
			Config:       json.RawMessage(`{"execution_order": 2}`),
		},
		{
			FormID:       formID,
			EntityID:     supplierEntityID,
			EntitySchema: "procurement",
			EntityTable:  "suppliers",
			Name:         "rating",
			Label:        "Rating",
			FieldType:    "number",
			FieldOrder:   order + 2,
			Required:     true,
			Config:       json.RawMessage(`{"execution_order": 2, "min": 0, "max": 5, "step": 0.1}`),
		},
		{
			FormID:       formID,
			EntityID:     supplierEntityID,
			EntitySchema: "procurement",
			EntityTable:  "suppliers",
			Name:         "is_active",
			Label:        "Active",
			FieldType:    "boolean",
			FieldOrder:   order + 3,
			Required:     true,
			Config:       json.RawMessage(`{"execution_order": 2}`),
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
		orderFields[i].Config = mergeConfigJSON(orderFields[i].Config, map[string]interface{}{
			"execution_order": 1,
		})
		order++
	}
	fields = append(fields, orderFields...)

	// Section 2: Line Items (created second, references order_id from step 1)
	// Using lineitems field type for card-based repeatable line items UI
	minQuantity := 1
	maxQuantity := 10000

	lineItemsConfig := formfieldbus.LineItemsFieldConfig{
		ExecutionOrder: 2,
		Entity:         "sales.order_line_items",
		ParentField:    "order_id",
		Fields: []formfieldbus.LineItemField{
			{
				Name:     "product_id",
				Label:    "Product",
				Type:     "dropdown",
				Required: true,
				DropdownConfig: &formfieldbus.DropdownConfig{
					TableConfigName: "products_lookup",
					LabelColumn:     "name",
					ValueColumn:     "id",
				},
			},
			{
				Name:     "quantity",
				Label:    "Quantity",
				Type:     "number",
				Required: true,
				Validation: &formfieldbus.ValidationConfig{
					Min: &minQuantity,
					Max: &maxQuantity,
				},
			},
			{
				Name:     "discount",
				Label:    "Discount",
				Type:     "text",
				Required: false,
			},
			{
				Name:     "line_item_fulfillment_statuses_id",
				Label:    "Fulfillment Status",
				Type:     "dropdown",
				Required: true,
			},
			{
				Name:     "created_by",
				Label:    "Created By",
				Type:     "text",
				Required: true,
			},
		},
		ItemLabel:         "Order Items",
		SingularItemLabel: "Item",
		MinItems:          1,
		MaxItems:          100,
	}

	configJSON, err := lineItemsConfig.ToJSON()
	if err != nil {
		// Fallback to empty config if marshal fails
		configJSON = json.RawMessage("{}")
	}

	fields = append(fields, formfieldbus.NewFormField{
		FormID:       formID,
		EntityID:     lineItemEntityID,
		EntitySchema: "sales",
		EntityTable:  "order_line_items",
		Name:         "line_items",
		Label:        "Order Items",
		FieldType:    "lineitems",
		FieldOrder:   order,
		Required:     true,
		Config:       configJSON,
	})

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
		poFields[i].Config = mergeConfigJSON(poFields[i].Config, map[string]interface{}{
			"execution_order": 1,
		})
		order++
	}
	fields = append(fields, poFields...)

	// Section 2: Line Items (created second, references purchase_order_id from step 1)
	lineItemFields := GetPurchaseOrderLineItemFormFields(formID, lineItemEntityID)
	for i := range lineItemFields {
		lineItemFields[i].FieldOrder = order
		lineItemFields[i].Config = mergeConfigJSON(lineItemFields[i].Config, map[string]interface{}{
			"execution_order": 2,
			"parent_entity":   "purchase_orders",
			"parent_field":    "purchase_order_id",
			"repeatable":      true,
		})
		order++
	}
	fields = append(fields, lineItemFields...)

	return fields
}
