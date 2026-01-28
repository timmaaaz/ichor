package seedmodels

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
)

// =============================================================================
// Form Registry Types
// =============================================================================

// FormGeneratorFunc generates form fields given form and entity IDs
type FormGeneratorFunc func(formID, entityID uuid.UUID) []formfieldbus.NewFormField

// FormRegistryEntry holds metadata about a registered form
type FormRegistryEntry struct {
	Name           string
	SupportsUpdate bool // If true, validation requires 'id' field for update operations
	Generator      FormGeneratorFunc
}

// =============================================================================
// Form Registry
// =============================================================================

// FormRegistry holds all registered form generators for validation
var FormRegistry []FormRegistryEntry

// RegisterForm adds a form to the registry
func RegisterForm(name string, supportsUpdate bool, gen FormGeneratorFunc) {
	FormRegistry = append(FormRegistry, FormRegistryEntry{
		Name:           name,
		SupportsUpdate: supportsUpdate,
		Generator:      gen,
	})
}

// =============================================================================
// Form Registration
// =============================================================================
// Forms are registered in init() to ensure they're available for validation.
// The supportsUpdate flag indicates whether the form is used for editing
// existing records (e.g., /sales/orders/:id edit page).

func init() {
	// =========================================================================
	// Composite Forms (multi-entity) - These typically support updates
	// =========================================================================
	// These use wrapper functions to normalize the varying parameter signatures

	RegisterForm("GetFullSalesOrderFormFields", true, wrapFullSalesOrderForm)
	RegisterForm("GetFullPurchaseOrderFormFields", true, wrapFullPurchaseOrderForm)
	RegisterForm("GetFullSupplierFormFields", true, wrapFullSupplierForm)
	RegisterForm("GetFullCustomerFormFields", true, wrapFullCustomerForm)

	// =========================================================================
	// Entity Forms That Support Updates (edit pages exist)
	// =========================================================================

	RegisterForm("GetSalesOrderFormFields", true, GetSalesOrderFormFields)
	RegisterForm("GetSalesOrderLineItemFormFields", true, GetSalesOrderLineItemFormFields)
	RegisterForm("GetPurchaseOrderFormFields", true, GetPurchaseOrderFormFields)
	RegisterForm("GetPurchaseOrderLineItemFormFields", true, GetPurchaseOrderLineItemFormFields)
	RegisterForm("GetSupplierFormFields", true, GetSupplierFormFields)
	RegisterForm("GetCustomerFormFields", true, GetCustomerFormFields)
	RegisterForm("GetProductFormFields", true, GetProductFormFields)
	RegisterForm("GetUserFormFields", true, GetUserFormFields)

	// =========================================================================
	// Reference/Lookup Forms (create-only, no edit page)
	// =========================================================================

	// Geography
	RegisterForm("GetCountryFormFields", false, GetCountryFormFields)
	RegisterForm("GetRegionFormFields", false, GetRegionFormFields)
	RegisterForm("GetCityFormFields", false, GetCityFormFields)
	RegisterForm("GetStreetFormFields", false, GetStreetFormFields)

	// Core
	RegisterForm("GetContactInfoFormFields", false, GetContactInfoFormFields)
	RegisterForm("GetRoleFormFields", false, GetRoleFormFields)
	RegisterForm("GetEntityFormFields", false, GetEntityFormFields)
	RegisterForm("GetTagFormFields", false, GetTagFormFields)

	// Status/Approval types
	RegisterForm("GetUserApprovalStatusFormFields", false, GetUserApprovalStatusFormFields)
	RegisterForm("GetAssetApprovalStatusFormFields", false, GetAssetApprovalStatusFormFields)
	RegisterForm("GetAssetFulfillmentStatusFormFields", false, GetAssetFulfillmentStatusFormFields)
	RegisterForm("GetOrderFulfillmentStatusFormFields", false, GetOrderFulfillmentStatusFormFields)
	RegisterForm("GetLineItemFulfillmentStatusFormFields", false, GetLineItemFulfillmentStatusFormFields)
	RegisterForm("GetPurchaseOrderStatusFormFields", false, GetPurchaseOrderStatusFormFields)
	RegisterForm("GetPOLineItemStatusFormFields", false, GetPOLineItemStatusFormFields)
	RegisterForm("GetUserApprovalCommentFormFields", false, GetUserApprovalCommentFormFields)

	// Asset types
	RegisterForm("GetAssetTypeFormFields", false, GetAssetTypeFormFields)
	RegisterForm("GetAssetConditionFormFields", false, GetAssetConditionFormFields)
	RegisterForm("GetValidAssetFormFields", false, GetValidAssetFormFields)
	RegisterForm("GetAssetFormFields", false, GetAssetFormFields)
	RegisterForm("GetUserAssetFormFields", false, GetUserAssetFormFields)

	// Products
	RegisterForm("GetProductCategoryFormFields", false, GetProductCategoryFormFields)
	RegisterForm("GetBrandFormFields", false, GetBrandFormFields)
	RegisterForm("GetPhysicalAttributeFormFields", false, GetPhysicalAttributeFormFields)
	RegisterForm("GetProductCostFormFields", false, GetProductCostFormFields)
	RegisterForm("GetCostHistoryFormFields", false, GetCostHistoryFormFields)
	RegisterForm("GetQualityMetricFormFields", false, GetQualityMetricFormFields)
	RegisterForm("GetSupplierProductFormFields", false, GetSupplierProductFormFields)

	// Inventory
	RegisterForm("GetWarehouseFormFields", false, GetWarehouseFormFields)
	RegisterForm("GetZoneFormFields", false, GetZoneFormFields)
	RegisterForm("GetInventoryLocationFormFields", false, GetInventoryLocationFormFields)
	RegisterForm("GetInventoryItemFormFields", false, GetInventoryItemFormFields)
	RegisterForm("GetSerialNumberFormFields", false, GetSerialNumberFormFields)
	RegisterForm("GetLotTrackingFormFields", false, GetLotTrackingFormFields)
	RegisterForm("GetQualityInspectionFormFields", false, GetQualityInspectionFormFields)
	RegisterForm("GetInventoryTransactionFormFields", false, GetInventoryTransactionFormFields)
	RegisterForm("GetInventoryAdjustmentFormFields", false, GetInventoryAdjustmentFormFields)
	RegisterForm("GetTransferOrderFormFields", false, GetTransferOrderFormFields)

	// HR
	RegisterForm("GetTitleFormFields", false, GetTitleFormFields)
	RegisterForm("GetOfficeFormFields", false, GetOfficeFormFields)
	RegisterForm("GetHomeFormFields", false, GetHomeFormFields)

	// Workflow
	RegisterForm("GetAutomationRuleFormFields", false, GetAutomationRuleFormFields)
	RegisterForm("GetRuleActionFormFields", false, GetRuleActionFormFields)
}

// =============================================================================
// Wrapper Functions for Composite Forms
// =============================================================================
// These normalize the varying parameter signatures of composite form generators
// to the standard FormGeneratorFunc signature for the registry.

func wrapFullSalesOrderForm(formID, entityID uuid.UUID) []formfieldbus.NewFormField {
	lineItemEntityID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	return GetFullSalesOrderFormFields(formID, entityID, lineItemEntityID)
}

func wrapFullPurchaseOrderForm(formID, entityID uuid.UUID) []formfieldbus.NewFormField {
	lineItemEntityID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	return GetFullPurchaseOrderFormFields(formID, entityID, lineItemEntityID)
}

func wrapFullSupplierForm(formID, entityID uuid.UUID) []formfieldbus.NewFormField {
	contactEntityID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	return GetFullSupplierFormFields(formID, entityID, contactEntityID)
}

func wrapFullCustomerForm(formID, entityID uuid.UUID) []formfieldbus.NewFormField {
	contactEntityID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	streetEntityID := uuid.MustParse("00000000-0000-0000-0000-000000000005")
	return GetFullCustomerFormFields(formID, entityID, contactEntityID, streetEntityID)
}
