package dbtest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/formbus"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
	"github.com/timmaaaz/ichor/business/domain/config/pageconfigbus"
	"github.com/timmaaaz/ichor/business/domain/config/pagecontentbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest/seedmodels"
	"github.com/timmaaaz/ichor/foundation/logger"
)

func seedForms(ctx context.Context, log *logger.Logger, busDomain BusDomain) error {
	// =========================================================================
	// Create Forms
	// =========================================================================

	// Form 1: Single entity - Users only (using generator)
	userForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "User Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating user form : %w", err)
	}

	userEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "users")
	if err != nil {
		return fmt.Errorf("querying user entity : %w", err)
	}

	userFormFields := seedmodels.GetUserFormFields(userForm.ID, userEntity.ID)
	for _, ff := range userFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating user form field : %w", err)
		}
	}

	// Form 2: Single entity - Assets only (using generator)
	assetForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Asset Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating asset form : %w", err)
	}

	assetEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "assets")
	if err != nil {
		return fmt.Errorf("querying asset entity : %w", err)
	}

	assetFormFields := seedmodels.GetAssetFormFields(assetForm.ID, assetEntity.ID)
	for _, ff := range assetFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating asset form field : %w", err)
		}
	}

	// Form 3: Multi-entity - User then Asset (with foreign key)
	multiForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "User and Asset Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating multi-entity form : %w", err)
	}

	multiFormFields := []formfieldbus.FormField{
		// User fields (order 1-11)
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			EntitySchema: "core",
			EntityTable:  "users",
			Name:         "username",
			FieldOrder:   1,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			EntitySchema: "core",
			EntityTable:  "users",
			Name:         "first_name",
			FieldOrder:   2,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			EntitySchema: "core",
			EntityTable:  "users",
			Name:         "last_name",
			FieldOrder:   3,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			EntitySchema: "core",
			EntityTable:  "users",
			Name:         "email",
			FieldOrder:   4,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			EntitySchema: "core",
			EntityTable:  "users",
			Name:         "password",
			FieldOrder:   5,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			EntitySchema: "core",
			EntityTable:  "users",
			Name:         "password_confirm",
			FieldOrder:   6,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			EntitySchema: "core",
			EntityTable:  "users",
			Name:         "birthday",
			FieldOrder:   7,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			EntitySchema: "core",
			EntityTable:  "users",
			Name:         "roles",
			FieldOrder:   8,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			EntitySchema: "core",
			EntityTable:  "users",
			Name:         "system_roles",
			FieldOrder:   9,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			EntitySchema: "core",
			EntityTable:  "users",
			Name:         "enabled",
			FieldOrder:   10,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     userEntity.ID,
			EntitySchema: "core",
			EntityTable:  "users",
			Name:         "requested_by",
			FieldOrder:   11,
			Config:       json.RawMessage(`{}`),
		},
		// Asset fields (order 12-14)
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     assetEntity.ID,
			EntitySchema: "assets",
			EntityTable:  "assets",
			Name:         "asset_condition_id",
			FieldOrder:   12,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     assetEntity.ID,
			EntitySchema: "assets",
			EntityTable:  "assets",
			Name:         "valid_asset_id",
			FieldOrder:   13,
			Config:       json.RawMessage(`{}`),
		},
		{
			ID:           uuid.New(),
			FormID:       multiForm.ID,
			EntityID:     assetEntity.ID,
			EntitySchema: "assets",
			EntityTable:  "assets",
			Name:         "serial_number",
			FieldOrder:   14,
			Config:       json.RawMessage(`{}`),
		},
	}

	for _, ff := range multiFormFields {
		_, err = busDomain.FormField.Create(ctx, formfieldbus.NewFormField{
			FormID:       ff.FormID,
			EntityID:     ff.EntityID,
			EntitySchema: ff.EntitySchema,
			EntityTable:  ff.EntityTable,
			Name:         ff.Name,
			FieldOrder:   ff.FieldOrder,
		})
		if err != nil {
			return fmt.Errorf("creating multi-entity form field : %w", err)
		}
	}

	// =============================================================================
	// COMPOSITE FORMS
	// =============================================================================

	// Composite Form 1: Full Customer (Customer + Contact Info + Delivery Address)
	fullCustomerForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Full Customer Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating full customer form : %w", err)
	}

	customerEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "customers")
	if err != nil {
		return fmt.Errorf("querying customer entity : %w", err)
	}

	contactInfoEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "contact_infos")
	if err != nil {
		return fmt.Errorf("querying contact_infos entity : %w", err)
	}

	streetEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "streets")
	if err != nil {
		return fmt.Errorf("querying streets entity : %w", err)
	}

	fullCustomerFields := seedmodels.GetFullCustomerFormFields(
		fullCustomerForm.ID,
		customerEntity.ID,
		contactInfoEntity.ID,
		streetEntity.ID,
	)

	for _, ff := range fullCustomerFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating full customer form field : %w", err)
		}
	}

	// Composite Form 2: Full Supplier (Supplier + Contact Info)
	fullSupplierForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Full Supplier Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating full supplier form : %w", err)
	}

	supplierEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "suppliers")
	if err != nil {
		return fmt.Errorf("querying supplier entity : %w", err)
	}

	fullSupplierFields := seedmodels.GetFullSupplierFormFields(
		fullSupplierForm.ID,
		supplierEntity.ID,
		contactInfoEntity.ID,
	)

	for _, ff := range fullSupplierFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating full supplier form field : %w", err)
		}
	}

	// Composite Form 3: Full Sales Order (Order + Line Items)
	fullSalesOrderForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Full Sales Order Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating full sales order form : %w", err)
	}

	orderEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "orders")
	if err != nil {
		return fmt.Errorf("querying orders entity : %w", err)
	}

	orderLineItemEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "order_line_items")
	if err != nil {
		return fmt.Errorf("querying order_line_items entity : %w", err)
	}

	fullSalesOrderFields := seedmodels.GetFullSalesOrderFormFields(
		fullSalesOrderForm.ID,
		orderEntity.ID,
		orderLineItemEntity.ID,
	)

	for _, ff := range fullSalesOrderFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating full sales order form field : %w", err)
		}
	}

	// Composite Form 4: Full Purchase Order (PO + Line Items)
	fullPurchaseOrderForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Full Purchase Order Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating full purchase order form : %w", err)
	}

	purchaseOrderEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "purchase_orders")
	if err != nil {
		return fmt.Errorf("querying purchase_orders entity : %w", err)
	}

	poLineItemEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "purchase_order_line_items")
	if err != nil {
		return fmt.Errorf("querying purchase_order_line_items entity : %w", err)
	}

	fullPurchaseOrderFields := seedmodels.GetFullPurchaseOrderFormFields(
		fullPurchaseOrderForm.ID,
		purchaseOrderEntity.ID,
		poLineItemEntity.ID,
	)

	for _, ff := range fullPurchaseOrderFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating full purchase order form field : %w", err)
		}
	}

	// =============================================================================
	// SIMPLE FORMS (Dropdown-based for foreign keys)
	// =============================================================================

	// Simple Form 1: Role
	roleForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Role Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating role form : %w", err)
	}

	roleEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "roles")
	if err != nil {
		return fmt.Errorf("querying roles entity : %w", err)
	}

	roleFields := seedmodels.GetRoleFormFields(roleForm.ID, roleEntity.ID)
	for _, ff := range roleFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating role form field : %w", err)
		}
	}

	// Simple Form 2: Customer (dropdown version)
	simpleCustomerForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Customer Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating simple customer form : %w", err)
	}

	simpleCustomerFields := seedmodels.GetCustomerFormFields(simpleCustomerForm.ID, customerEntity.ID)
	for _, ff := range simpleCustomerFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating simple customer form field : %w", err)
		}
	}

	// Simple Form 3: Sales Order (dropdown version)
	simpleSalesOrderForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Sales Order Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating simple sales order form : %w", err)
	}

	simpleSalesOrderFields := seedmodels.GetSalesOrderFormFields(simpleSalesOrderForm.ID, orderEntity.ID)
	for _, ff := range simpleSalesOrderFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating simple sales order form field : %w", err)
		}
	}

	// Simple Form 4: Supplier (dropdown version)
	simpleSupplierForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Supplier Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating simple supplier form : %w", err)
	}

	simpleSupplierFields := seedmodels.GetSupplierFormFields(simpleSupplierForm.ID, supplierEntity.ID)
	for _, ff := range simpleSupplierFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating simple supplier form field : %w", err)
		}
	}

	// Simple Form 5: Purchase Order (dropdown version)
	simplePurchaseOrderForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Purchase Order Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating simple purchase order form : %w", err)
	}

	simplePurchaseOrderFields := seedmodels.GetPurchaseOrderFormFields(simplePurchaseOrderForm.ID, purchaseOrderEntity.ID)
	for _, ff := range simplePurchaseOrderFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating simple purchase order form field : %w", err)
		}
	}

	// Simple Form 6: Warehouse
	warehouseForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Warehouse Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating warehouse form : %w", err)
	}

	warehouseEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "warehouses")
	if err != nil {
		return fmt.Errorf("querying warehouses entity : %w", err)
	}

	warehouseFields := seedmodels.GetWarehouseFormFields(warehouseForm.ID, warehouseEntity.ID)
	for _, ff := range warehouseFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating warehouse form field : %w", err)
		}
	}

	// Simple Form 7: Inventory Adjustment
	inventoryAdjustmentForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Inventory Adjustment Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating inventory adjustment form : %w", err)
	}

	inventoryAdjustmentEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "inventory_adjustments")
	if err != nil {
		return fmt.Errorf("querying inventory_adjustments entity : %w", err)
	}

	inventoryAdjustmentFields := seedmodels.GetInventoryAdjustmentFormFields(inventoryAdjustmentForm.ID, inventoryAdjustmentEntity.ID)
	for _, ff := range inventoryAdjustmentFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating inventory adjustment form field : %w", err)
		}
	}

	// Simple Form 8: Transfer Order
	transferOrderForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Transfer Order Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating transfer order form : %w", err)
	}

	transferOrderEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "transfer_orders")
	if err != nil {
		return fmt.Errorf("querying transfer_orders entity : %w", err)
	}

	transferOrderFields := seedmodels.GetTransferOrderFormFields(transferOrderForm.ID, transferOrderEntity.ID)
	for _, ff := range transferOrderFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating transfer order form field : %w", err)
		}
	}

	// Simple Form 9: Inventory Item
	inventoryItemForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Inventory Item Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating inventory item form : %w", err)
	}

	inventoryItemEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "inventory_items")
	if err != nil {
		return fmt.Errorf("querying inventory_items entity : %w", err)
	}

	inventoryItemFields := seedmodels.GetInventoryItemFormFields(inventoryItemForm.ID, inventoryItemEntity.ID)
	for _, ff := range inventoryItemFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating inventory item form field : %w", err)
		}
	}

	// Simple Form 10: Office
	officeForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Office Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating office form : %w", err)
	}

	officeEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "offices")
	if err != nil {
		return fmt.Errorf("querying offices entity : %w", err)
	}

	officeFields := seedmodels.GetOfficeFormFields(officeForm.ID, officeEntity.ID)
	for _, ff := range officeFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating office form field : %w", err)
		}
	}

	// =============================================================================
	// REFERENCE DATA FORMS (Admin-managed, no inline creation)
	// =============================================================================

	// Reference Form 1: Country
	countryForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Country Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating country form : %w", err)
	}

	countryEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "countries")
	if err != nil {
		return fmt.Errorf("querying countries entity : %w", err)
	}

	countryFields := seedmodels.GetCountryFormFields(countryForm.ID, countryEntity.ID)
	for _, ff := range countryFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating country form field : %w", err)
		}
	}

	// Reference Form 2: Region
	regionForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Region Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating region form : %w", err)
	}

	regionEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "regions")
	if err != nil {
		return fmt.Errorf("querying regions entity : %w", err)
	}

	regionFields := seedmodels.GetRegionFormFields(regionForm.ID, regionEntity.ID)
	for _, ff := range regionFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating region form field : %w", err)
		}
	}

	// Reference Form 3: User Approval Status
	userApprovalStatusForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "User Approval Status Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating user approval status form : %w", err)
	}

	userApprovalStatusEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "user_approval_status")
	if err != nil {
		return fmt.Errorf("querying user_approval_status entity : %w", err)
	}

	userApprovalStatusFields := seedmodels.GetUserApprovalStatusFormFields(userApprovalStatusForm.ID, userApprovalStatusEntity.ID)
	for _, ff := range userApprovalStatusFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating user approval status form field : %w", err)
		}
	}

	// Reference Form 4: Asset Approval Status
	assetApprovalStatusForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Asset Approval Status Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating asset approval status form : %w", err)
	}

	assetApprovalStatusEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "approval_status")
	if err != nil {
		return fmt.Errorf("querying approval_status entity : %w", err)
	}

	assetApprovalStatusFields := seedmodels.GetAssetApprovalStatusFormFields(assetApprovalStatusForm.ID, assetApprovalStatusEntity.ID)
	for _, ff := range assetApprovalStatusFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating asset approval status form field : %w", err)
		}
	}

	// Reference Form 5: Asset Fulfillment Status
	assetFulfillmentStatusForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Asset Fulfillment Status Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating asset fulfillment status form : %w", err)
	}

	assetFulfillmentStatusEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "fulfillment_status")
	if err != nil {
		return fmt.Errorf("querying fulfillment_status entity : %w", err)
	}

	assetFulfillmentStatusFields := seedmodels.GetAssetFulfillmentStatusFormFields(assetFulfillmentStatusForm.ID, assetFulfillmentStatusEntity.ID)
	for _, ff := range assetFulfillmentStatusFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating asset fulfillment status form field : %w", err)
		}
	}

	// Reference Form 6: Order Fulfillment Status
	orderFulfillmentStatusForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Order Fulfillment Status Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating order fulfillment status form : %w", err)
	}

	orderFulfillmentStatusEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "order_fulfillment_statuses")
	if err != nil {
		return fmt.Errorf("querying order_fulfillment_statuses entity : %w", err)
	}

	orderFulfillmentStatusFields := seedmodels.GetOrderFulfillmentStatusFormFields(orderFulfillmentStatusForm.ID, orderFulfillmentStatusEntity.ID)
	for _, ff := range orderFulfillmentStatusFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating order fulfillment status form field : %w", err)
		}
	}

	// Reference Form 7: Line Item Fulfillment Status
	lineItemFulfillmentStatusForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Line Item Fulfillment Status Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating line item fulfillment status form : %w", err)
	}

	lineItemFulfillmentStatusEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "line_item_fulfillment_statuses")
	if err != nil {
		return fmt.Errorf("querying line_item_fulfillment_statuses entity : %w", err)
	}

	lineItemFulfillmentStatusFields := seedmodels.GetLineItemFulfillmentStatusFormFields(lineItemFulfillmentStatusForm.ID, lineItemFulfillmentStatusEntity.ID)
	for _, ff := range lineItemFulfillmentStatusFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating line item fulfillment status form field : %w", err)
		}
	}

	// Reference Form 8: Purchase Order Status
	purchaseOrderStatusForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Purchase Order Status Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating purchase order status form : %w", err)
	}

	purchaseOrderStatusEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "purchase_order_statuses")
	if err != nil {
		return fmt.Errorf("querying purchase_order_statuses entity : %w", err)
	}

	purchaseOrderStatusFields := seedmodels.GetPurchaseOrderStatusFormFields(purchaseOrderStatusForm.ID, purchaseOrderStatusEntity.ID)
	for _, ff := range purchaseOrderStatusFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating purchase order status form field : %w", err)
		}
	}

	// Reference Form 9: PO Line Item Status
	poLineItemStatusForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Purchase Order Line Item Status Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating po line item status form : %w", err)
	}

	poLineItemStatusEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "purchase_order_line_item_statuses")
	if err != nil {
		return fmt.Errorf("querying purchase_order_line_item_statuses entity : %w", err)
	}

	poLineItemStatusFields := seedmodels.GetPOLineItemStatusFormFields(poLineItemStatusForm.ID, poLineItemStatusEntity.ID)
	for _, ff := range poLineItemStatusFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating po line item status form field : %w", err)
		}
	}

	// Reference Form 10: Asset Type
	assetTypeForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Asset Type Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating asset type form : %w", err)
	}

	assetTypeEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "asset_types")
	if err != nil {
		return fmt.Errorf("querying asset_types entity : %w", err)
	}

	assetTypeFields := seedmodels.GetAssetTypeFormFields(assetTypeForm.ID, assetTypeEntity.ID)
	for _, ff := range assetTypeFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating asset type form field : %w", err)
		}
	}

	// Reference Form 11: Asset Condition
	assetConditionForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Asset Condition Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating asset condition form : %w", err)
	}

	assetConditionEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "asset_conditions")
	if err != nil {
		return fmt.Errorf("querying asset_conditions entity : %w", err)
	}

	assetConditionFields := seedmodels.GetAssetConditionFormFields(assetConditionForm.ID, assetConditionEntity.ID)
	for _, ff := range assetConditionFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating asset condition form field : %w", err)
		}
	}

	// Reference Form 12: Product Category
	productCategoryForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Product Category Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating product category form : %w", err)
	}

	productCategoryEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "product_categories")
	if err != nil {
		return fmt.Errorf("querying product_categories entity : %w", err)
	}

	productCategoryFields := seedmodels.GetProductCategoryFormFields(productCategoryForm.ID, productCategoryEntity.ID)
	for _, ff := range productCategoryFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating product category form field : %w", err)
		}
	}

	// =============================================================================
	// HIGH-PRIORITY TRANSACTIONAL FORMS (Referenced in inline_create)
	// =============================================================================

	// High Priority Form 1: City
	cityForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "City Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating city form : %w", err)
	}

	cityEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "cities")
	if err != nil {
		return fmt.Errorf("querying cities entity : %w", err)
	}

	cityFields := seedmodels.GetCityFormFields(cityForm.ID, cityEntity.ID)
	for _, ff := range cityFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating city form field : %w", err)
		}
	}

	// High Priority Form 2: Street (entity already declared)
	streetForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Street Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating street form : %w", err)
	}

	streetFormFields := seedmodels.GetStreetFormFields(streetForm.ID, streetEntity.ID)
	for _, ff := range streetFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating street form field : %w", err)
		}
	}

	// High Priority Form 3: Contact Info (entity already declared)
	contactInfoForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Contact Info Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating contact info form : %w", err)
	}

	contactInfoFormFields := seedmodels.GetContactInfoFormFields(contactInfoForm.ID, contactInfoEntity.ID)
	for _, ff := range contactInfoFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating contact info form field : %w", err)
		}
	}

	// High Priority Form 4: Title
	titleForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Title Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating title form : %w", err)
	}

	titleEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "titles")
	if err != nil {
		return fmt.Errorf("querying titles entity : %w", err)
	}

	titleFormFields := seedmodels.GetTitleFormFields(titleForm.ID, titleEntity.ID)
	for _, ff := range titleFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating title form field : %w", err)
		}
	}

	// High Priority Form 5: Product
	productForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Product Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating product form : %w", err)
	}

	productEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "products")
	if err != nil {
		return fmt.Errorf("querying products entity : %w", err)
	}

	productFormFields := seedmodels.GetProductFormFields(productForm.ID, productEntity.ID)
	for _, ff := range productFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating product form field : %w", err)
		}
	}

	// High Priority Form 6: Inventory Location
	inventoryLocationForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Inventory Location Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating inventory location form : %w", err)
	}

	inventoryLocationEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "inventory_locations")
	if err != nil {
		return fmt.Errorf("querying inventory_locations entity : %w", err)
	}

	inventoryLocationFormFields := seedmodels.GetInventoryLocationFormFields(inventoryLocationForm.ID, inventoryLocationEntity.ID)
	for _, ff := range inventoryLocationFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating inventory location form field : %w", err)
		}
	}

	// High Priority Form 7: Valid Asset
	validAssetForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Valid Asset Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating valid asset form : %w", err)
	}

	validAssetEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "valid_assets")
	if err != nil {
		return fmt.Errorf("querying valid_assets entity : %w", err)
	}

	validAssetFormFields := seedmodels.GetValidAssetFormFields(validAssetForm.ID, validAssetEntity.ID)
	for _, ff := range validAssetFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating valid asset form field : %w", err)
		}
	}

	// High Priority Form 8: Supplier Product
	supplierProductForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Supplier Product Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating supplier product form : %w", err)
	}

	supplierProductEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "supplier_products")
	if err != nil {
		return fmt.Errorf("querying supplier_products entity : %w", err)
	}

	supplierProductFormFields := seedmodels.GetSupplierProductFormFields(supplierProductForm.ID, supplierProductEntity.ID)
	for _, ff := range supplierProductFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating supplier product form field : %w", err)
		}
	}

	// High Priority Form 9: Sales Order Line Item (entity already declared)
	salesOrderLineItemForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Sales Order Line Item Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating sales order line item form : %w", err)
	}

	salesOrderLineItemFormFields := seedmodels.GetSalesOrderLineItemFormFields(salesOrderLineItemForm.ID, orderLineItemEntity.ID)
	for _, ff := range salesOrderLineItemFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating sales order line item form field : %w", err)
		}
	}

	// High Priority Form 10: Purchase Order Line Item (entity already declared)
	purchaseOrderLineItemForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Purchase Order Line Item Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating purchase order line item form : %w", err)
	}

	purchaseOrderLineItemFormFields := seedmodels.GetPurchaseOrderLineItemFormFields(purchaseOrderLineItemForm.ID, poLineItemEntity.ID)
	for _, ff := range purchaseOrderLineItemFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating purchase order line item form field : %w", err)
		}
	}

	// =============================================================================
	// MEDIUM-PRIORITY TRANSACTIONAL FORMS (Domain completeness)
	// =============================================================================

	// Medium Priority Form 1: Home
	homeForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Home Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating home form : %w", err)
	}

	homeEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "homes")
	if err != nil {
		return fmt.Errorf("querying homes entity : %w", err)
	}

	homeFormFields := seedmodels.GetHomeFormFields(homeForm.ID, homeEntity.ID)
	for _, ff := range homeFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating home form field : %w", err)
		}
	}

	// Medium Priority Form 2: Tag
	tagForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Tag Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating tag form : %w", err)
	}

	tagEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "tags")
	if err != nil {
		return fmt.Errorf("querying tags entity : %w", err)
	}

	tagFormFields := seedmodels.GetTagFormFields(tagForm.ID, tagEntity.ID)
	for _, ff := range tagFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating tag form field : %w", err)
		}
	}

	// Medium Priority Form 3: User Approval Comment
	userApprovalCommentForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "User Approval Comment Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating user approval comment form : %w", err)
	}

	userApprovalCommentEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "user_approval_comments")
	if err != nil {
		return fmt.Errorf("querying user_approval_comments entity : %w", err)
	}

	userApprovalCommentFormFields := seedmodels.GetUserApprovalCommentFormFields(userApprovalCommentForm.ID, userApprovalCommentEntity.ID)
	for _, ff := range userApprovalCommentFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating user approval comment form field : %w", err)
		}
	}

	// Medium Priority Form 4: Brand
	brandForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Brand Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating brand form : %w", err)
	}

	brandEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "brands")
	if err != nil {
		return fmt.Errorf("querying brands entity : %w", err)
	}

	brandFormFields := seedmodels.GetBrandFormFields(brandForm.ID, brandEntity.ID)
	for _, ff := range brandFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating brand form field : %w", err)
		}
	}

	// Medium Priority Form 5: Zone
	zoneForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Zone Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating zone form : %w", err)
	}

	zoneEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "zones")
	if err != nil {
		return fmt.Errorf("querying zones entity : %w", err)
	}

	zoneFormFields := seedmodels.GetZoneFormFields(zoneForm.ID, zoneEntity.ID)
	for _, ff := range zoneFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating zone form field : %w", err)
		}
	}

	// Medium Priority Form 6: User Asset
	userAssetForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "User Asset Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating user asset form : %w", err)
	}

	userAssetEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "user_assets")
	if err != nil {
		return fmt.Errorf("querying user_assets entity : %w", err)
	}

	userAssetFormFields := seedmodels.GetUserAssetFormFields(userAssetForm.ID, userAssetEntity.ID)
	for _, ff := range userAssetFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating user asset form field : %w", err)
		}
	}

	// Medium Priority Form 7: Automation Rule
	automationRuleForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Automation Rule Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating automation rule form : %w", err)
	}

	automationRuleEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "automation_rules")
	if err != nil {
		return fmt.Errorf("querying automation_rules entity : %w", err)
	}

	automationRuleFormFields := seedmodels.GetAutomationRuleFormFields(automationRuleForm.ID, automationRuleEntity.ID)
	for _, ff := range automationRuleFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating automation rule form field : %w", err)
		}
	}

	// Medium Priority Form 8: Rule Action
	ruleActionForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Rule Action Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating rule action form : %w", err)
	}

	ruleActionEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "rule_actions")
	if err != nil {
		return fmt.Errorf("querying rule_actions entity : %w", err)
	}

	ruleActionFormFields := seedmodels.GetRuleActionFormFields(ruleActionForm.ID, ruleActionEntity.ID)
	for _, ff := range ruleActionFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating rule action form field : %w", err)
		}
	}

	// Medium Priority Form 9: Entity
	entityForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Entity Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating entity form : %w", err)
	}

	entityEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "entities")
	if err != nil {
		return fmt.Errorf("querying entities entity : %w", err)
	}

	entityFormFields := seedmodels.GetEntityFormFields(entityForm.ID, entityEntity.ID)
	for _, ff := range entityFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating entity form field : %w", err)
		}
	}

	// Medium Priority Form 10: User (using proper generator instead of inline)
	userFormProp, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "User Creation Form (Proper)",
	})
	if err != nil {
		return fmt.Errorf("creating user form (proper) : %w", err)
	}

	userFormProperFields := seedmodels.GetUserFormFields(userFormProp.ID, userEntity.ID)
	for _, ff := range userFormProperFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating user form field : %w", err)
		}
	}

	// =============================================================================
	// LOWER-PRIORITY TRANSACTIONAL FORMS (Utility/tracking)
	// =============================================================================

	// Lower Priority Form 1: Physical Attribute
	physicalAttributeForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Physical Attribute Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating physical attribute form : %w", err)
	}

	physicalAttributeEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "physical_attributes")
	if err != nil {
		return fmt.Errorf("querying physical_attributes entity : %w", err)
	}

	physicalAttributeFormFields := seedmodels.GetPhysicalAttributeFormFields(physicalAttributeForm.ID, physicalAttributeEntity.ID)
	for _, ff := range physicalAttributeFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating physical attribute form field : %w", err)
		}
	}

	// Lower Priority Form 2: Product Cost
	productCostForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Product Cost Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating product cost form : %w", err)
	}

	productCostEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "product_costs")
	if err != nil {
		return fmt.Errorf("querying product_costs entity : %w", err)
	}

	productCostFormFields := seedmodels.GetProductCostFormFields(productCostForm.ID, productCostEntity.ID)
	for _, ff := range productCostFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating product cost form field : %w", err)
		}
	}

	// Lower Priority Form 3: Cost History
	costHistoryForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Cost History Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating cost history form : %w", err)
	}

	costHistoryEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "cost_history")
	if err != nil {
		return fmt.Errorf("querying cost_history entity : %w", err)
	}

	costHistoryFormFields := seedmodels.GetCostHistoryFormFields(costHistoryForm.ID, costHistoryEntity.ID)
	for _, ff := range costHistoryFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating cost history form field : %w", err)
		}
	}

	// Lower Priority Form 4: Quality Metric
	qualityMetricForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Quality Metric Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating quality metric form : %w", err)
	}

	qualityMetricEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "quality_metrics")
	if err != nil {
		return fmt.Errorf("querying quality_metrics entity : %w", err)
	}

	qualityMetricFormFields := seedmodels.GetQualityMetricFormFields(qualityMetricForm.ID, qualityMetricEntity.ID)
	for _, ff := range qualityMetricFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating quality metric form field : %w", err)
		}
	}

	// Lower Priority Form 5: Serial Number
	serialNumberForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Serial Number Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating serial number form : %w", err)
	}

	serialNumberEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "serial_numbers")
	if err != nil {
		return fmt.Errorf("querying serial_numbers entity : %w", err)
	}

	serialNumberFormFields := seedmodels.GetSerialNumberFormFields(serialNumberForm.ID, serialNumberEntity.ID)
	for _, ff := range serialNumberFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating serial number form field : %w", err)
		}
	}

	// Lower Priority Form 6: Lot Tracking
	lotTrackingForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Lot Tracking Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating lot tracking form : %w", err)
	}

	lotTrackingEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "lot_trackings")
	if err != nil {
		return fmt.Errorf("querying lot_trackings entity : %w", err)
	}

	lotTrackingFormFields := seedmodels.GetLotTrackingFormFields(lotTrackingForm.ID, lotTrackingEntity.ID)
	for _, ff := range lotTrackingFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating lot tracking form field : %w", err)
		}
	}

	// Lower Priority Form 7: Quality Inspection
	qualityInspectionForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Quality Inspection Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating quality inspection form : %w", err)
	}

	qualityInspectionEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "quality_inspections")
	if err != nil {
		return fmt.Errorf("querying quality_inspections entity : %w", err)
	}

	qualityInspectionFormFields := seedmodels.GetQualityInspectionFormFields(qualityInspectionForm.ID, qualityInspectionEntity.ID)
	for _, ff := range qualityInspectionFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating quality inspection form field : %w", err)
		}
	}

	// Lower Priority Form 8: Inventory Transaction
	inventoryTransactionForm, err := busDomain.Form.Create(ctx, formbus.NewForm{
		Name: "Inventory Transaction Creation Form",
	})
	if err != nil {
		return fmt.Errorf("creating inventory transaction form : %w", err)
	}

	inventoryTransactionEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "inventory_transactions")
	if err != nil {
		return fmt.Errorf("querying inventory_transactions entity : %w", err)
	}

	inventoryTransactionFormFields := seedmodels.GetInventoryTransactionFormFields(inventoryTransactionForm.ID, inventoryTransactionEntity.ID)
	for _, ff := range inventoryTransactionFormFields {
		_, err = busDomain.FormField.Create(ctx, ff)
		if err != nil {
			return fmt.Errorf("creating inventory transaction form field : %w", err)
		}
	}


	// =========================================================================
	// NEW PAGE CONTENT SYSTEM EXAMPLE - Flexible content blocks
	// =========================================================================
	// Query table configs needed for demo pages
	adminUsersTableStored, err := busDomain.ConfigStore.QueryByName(ctx, "admin_users_table")
	if err != nil {
		return fmt.Errorf("querying admin users table config for demo: %w", err)
	}

	adminRolesTableStored, err := busDomain.ConfigStore.QueryByName(ctx, "admin_roles_table")
	if err != nil {
		return fmt.Errorf("querying admin roles table config for demo: %w", err)
	}

	// Create a new page config for "User Management Example"
	userManagementPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "user_management_example",
		UserID:    uuid.Nil, // System default
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating user management example page : %w", err)
	}

	// Content Block 1: Form at top (New User Form)
	// Full width on all screen sizes
	formBlock, err := busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID: userManagementPage.ID,
		ContentType:  pagecontentbus.ContentTypeForm,
		Label:        "Create New User",
		FormID:       userForm.ID, // Reference the user form we created earlier
		OrderIndex:   1,
		Layout:       json.RawMessage(`{"colSpan":{"xs":12}}`),
		IsVisible:    true,
		IsDefault:    false,
	})
	if err != nil {
		return fmt.Errorf("creating form content block : %w", err)
	}

	// Content Block 2: Tabs Container
	// This is a container that will hold the tab items
	tabsContainer, err := busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID: userManagementPage.ID,
		ContentType:  pagecontentbus.ContentTypeTabs,
		Label:        "User Lists",
		OrderIndex:   2,
		Layout:       json.RawMessage(`{"colSpan":{"xs":12},"containerType":"tabs"}`),
		IsVisible:    true,
		IsDefault:    false,
	})
	if err != nil {
		return fmt.Errorf("creating tabs container : %w", err)
	}

	// Tab 1: Active Users (using admin users table config)
	// This is a CHILD of the tabs container
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  userManagementPage.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Active Users",
		TableConfigID: adminUsersTableStored.ID, // Reference existing table config
		OrderIndex:    1,
		ParentID:      tabsContainer.ID, // This makes it a child of the tabs container
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true, // This tab is active by default
	})
	if err != nil {
		return fmt.Errorf("creating active users tab : %w", err)
	}

	// Tab 2: Roles (using roles table config)
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  userManagementPage.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Roles",
		TableConfigID: adminRolesTableStored.ID, // Reference existing table config
		OrderIndex:    2,
		ParentID:      tabsContainer.ID, // Child of tabs container
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating roles tab : %w", err)
	}

	// Tab 3: Permissions (using table access config if available)
	adminTableAccessTableStored, err := busDomain.ConfigStore.QueryByName(ctx, "admin_table_access_page")
	if err == nil {
		// Only create this tab if the table config exists
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  userManagementPage.ID,
			ContentType:   pagecontentbus.ContentTypeTable,
			Label:         "Permissions",
			TableConfigID: adminTableAccessTableStored.ID,
			OrderIndex:    3,
			ParentID:      tabsContainer.ID, // Child of tabs container
			Layout:        json.RawMessage(`{}`),
			IsVisible:     true,
			IsDefault:     false,
		})
		if err != nil {
			return fmt.Errorf("creating permissions tab : %w", err)
		}
	}

	// Log success
	log.Info(ctx, "✅ Created User Management Example page with flexible content blocks",
		"page_config_id", userManagementPage.ID,
		"form_block_id", formBlock.ID,
		"tabs_container_id", tabsContainer.ID)

	// =========================================================================
	// Create Sample Charts Dashboard
	// Demonstrates remaining chart types not distributed to other pages
	// =========================================================================

	// Query remaining chart configs for sample dashboard (those not queried earlier)
	stackedBarRegionStored, _ := busDomain.ConfigStore.QueryByName(ctx, "seed_stacked_bar_region")
	stackedAreaCumulativeStored, _ := busDomain.ConfigStore.QueryByName(ctx, "seed_stacked_area_cumulative")
	comboRevenueOrdersStored, _ := busDomain.ConfigStore.QueryByName(ctx, "seed_combo_revenue_orders")
	waterfallProfitStored, _ := busDomain.ConfigStore.QueryByName(ctx, "seed_waterfall_profit")

	// Only create dashboard if at least some chart configs exist
	if stackedBarRegionStored != nil {
		sampleChartsDashboardPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
			Name:      "sample_charts_dashboard",
			UserID:    uuid.Nil,
			IsDefault: true,
		})
		if err != nil {
			return fmt.Errorf("creating sample charts dashboard page: %w", err)
		}

		orderIndex := 1

		// Row 1: Stacked charts (2 across)
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  sampleChartsDashboardPage.ID,
			ContentType:   pagecontentbus.ContentTypeChart,
			Label:         "Sales by Region",
			ChartConfigID: stackedBarRegionStored.ID,
			OrderIndex:    orderIndex,
			Layout:        json.RawMessage(`{"colSpan":{"xs":12,"md":6}}`),
			IsVisible:     true,
			IsDefault:     true,
		})
		if err != nil {
			return fmt.Errorf("creating stacked bar chart content: %w", err)
		}
		orderIndex++

		if stackedAreaCumulativeStored != nil {
			_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
				PageConfigID:  sampleChartsDashboardPage.ID,
				ContentType:   pagecontentbus.ContentTypeChart,
				Label:         "Cumulative Revenue",
				ChartConfigID: stackedAreaCumulativeStored.ID,
				OrderIndex:    orderIndex,
				Layout:        json.RawMessage(`{"colSpan":{"xs":12,"md":6}}`),
				IsVisible:     true,
				IsDefault:     false,
			})
			if err != nil {
				return fmt.Errorf("creating stacked area chart content: %w", err)
			}
			orderIndex++
		}

		// Row 2: Combo + Waterfall (2 across)
		if comboRevenueOrdersStored != nil {
			_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
				PageConfigID:  sampleChartsDashboardPage.ID,
				ContentType:   pagecontentbus.ContentTypeChart,
				Label:         "Revenue vs Orders",
				ChartConfigID: comboRevenueOrdersStored.ID,
				OrderIndex:    orderIndex,
				Layout:        json.RawMessage(`{"colSpan":{"xs":12,"md":6}}`),
				IsVisible:     true,
				IsDefault:     false,
			})
			if err != nil {
				return fmt.Errorf("creating combo chart content: %w", err)
			}
			orderIndex++
		}

		if waterfallProfitStored != nil {
			_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
				PageConfigID:  sampleChartsDashboardPage.ID,
				ContentType:   pagecontentbus.ContentTypeChart,
				Label:         "Profit Breakdown",
				ChartConfigID: waterfallProfitStored.ID,
				OrderIndex:    orderIndex,
				Layout:        json.RawMessage(`{"colSpan":{"xs":12,"md":6}}`),
				IsVisible:     true,
				IsDefault:     false,
			})
			if err != nil {
				return fmt.Errorf("creating waterfall chart content: %w", err)
			}
		}

		log.Info(ctx, "✅ Created Sample Charts Dashboard page",
			"page_config_id", sampleChartsDashboardPage.ID)
	}

	return nil
}
