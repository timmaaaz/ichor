package all

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/assets/approvalstatusapp"
	"github.com/timmaaaz/ichor/app/domain/assets/assetapp"
	"github.com/timmaaaz/ichor/app/domain/assets/assetconditionapp"
	"github.com/timmaaaz/ichor/app/domain/assets/assettagapp"
	"github.com/timmaaaz/ichor/app/domain/assets/assettypeapp"
	"github.com/timmaaaz/ichor/app/domain/assets/fulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/domain/assets/tagapp"
	"github.com/timmaaaz/ichor/app/domain/assets/userassetapp"
	"github.com/timmaaaz/ichor/app/domain/assets/validassetapp"
	"github.com/timmaaaz/ichor/app/domain/config/formapp"
	"github.com/timmaaaz/ichor/app/domain/config/formfieldapp"
	"github.com/timmaaaz/ichor/app/domain/core/contactinfosapp"
	"github.com/timmaaaz/ichor/app/domain/core/roleapp"
	"github.com/timmaaaz/ichor/app/domain/core/tableaccessapp"
	"github.com/timmaaaz/ichor/app/domain/core/userapp"
	"github.com/timmaaaz/ichor/app/domain/core/userroleapp.go"
	"github.com/timmaaaz/ichor/app/domain/geography/cityapp"
	"github.com/timmaaaz/ichor/app/domain/geography/streetapp"
	"github.com/timmaaaz/ichor/app/domain/hr/approvalapp"
	"github.com/timmaaaz/ichor/app/domain/hr/commentapp"
	"github.com/timmaaaz/ichor/app/domain/hr/homeapp"
	"github.com/timmaaaz/ichor/app/domain/hr/officeapp"
	"github.com/timmaaaz/ichor/app/domain/hr/reportstoapp"
	"github.com/timmaaaz/ichor/app/domain/hr/titleapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/inspectionapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventoryadjustmentapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventoryitemapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventorylocationapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventorytransactionapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/lottrackingsapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/serialnumberapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/transferorderapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/warehouseapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/zoneapp"
	"github.com/timmaaaz/ichor/app/domain/procurement/supplierapp"
	"github.com/timmaaaz/ichor/app/domain/procurement/supplierproductapp"
	"github.com/timmaaaz/ichor/app/domain/products/brandapp"
	"github.com/timmaaaz/ichor/app/domain/products/costhistoryapp"
	"github.com/timmaaaz/ichor/app/domain/products/metricsapp"
	"github.com/timmaaaz/ichor/app/domain/products/physicalattributeapp"
	"github.com/timmaaaz/ichor/app/domain/products/productapp"
	"github.com/timmaaaz/ichor/app/domain/products/productcategoryapp"
	"github.com/timmaaaz/ichor/app/domain/products/productcostapp"
	"github.com/timmaaaz/ichor/app/domain/sales/customersapp"
	"github.com/timmaaaz/ichor/app/domain/sales/lineitemfulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/domain/sales/orderfulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/domain/sales/orderlineitemsapp"
	"github.com/timmaaaz/ichor/app/domain/sales/ordersapp"
	"github.com/timmaaaz/ichor/app/sdk/formdataregistry"
)

// buildFormDataRegistry creates and populates the form data registry with all entity registrations.
//
// This function centralizes all entity registrations for the dynamic form data service.
// Each entity requires four functions:
//  1. DecodeNew - Decode and validate JSON for CREATE operations
//  2. CreateFunc - Execute CREATE via app layer
//  3. DecodeUpdate - Decode and validate JSON for UPDATE operations
//  4. UpdateFunc - Execute UPDATE via app layer
//
// # Adding New Entities
//
// To register a new entity:
//
//  1. Add the app layer import at the top of this file
//  2. Add registration block following the pattern below
//  3. Test with a simple single-entity form first
//  4. Then test multi-entity forms with foreign keys
//
// Example for products entity:
//
//	if err := registry.Register(formdataregistry.EntityRegistration{
//	    Name: "products",
//	    DecodeNew: func(data json.RawMessage) (interface{}, error) {
//	        var app productapp.NewProduct
//	        if err := json.Unmarshal(data, &app); err != nil {
//	            return nil, err
//	        }
//	        if err := app.Validate(); err != nil {
//	            return nil, err
//	        }
//	        return app, nil
//	    },
//	    CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
//	        return productApp.Create(ctx, model.(productapp.NewProduct))
//	    },
//	    DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
//	        var app productapp.UpdateProduct
//	        if err := json.Unmarshal(data, &app); err != nil {
//	            return nil, err
//	        }
//	        if err := app.Validate(); err != nil {
//	            return nil, err
//	        }
//	        return app, nil
//	    },
//	    UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
//	        return productApp.Update(ctx, id, model.(productapp.UpdateProduct))
//	    },
//	}); err != nil {
//	    return nil, fmt.Errorf("register products: %w", err)
//	}
func buildFormDataRegistry(
	userApp *userapp.App,
	assetApp *assetapp.App,
	roleApp *roleapp.App,
	tableAccessApp *tableaccessapp.App,
	userRoleApp *userroleapp.App,
	contactInfosApp *contactinfosapp.App,
	assetConditionApp *assetconditionapp.App,
	assetTypeApp *assettypeapp.App,
	fulfillmentStatusApp *fulfillmentstatusapp.App,
	tagApp *tagapp.App,
	assetTagApp *assettagapp.App,
	validAssetApp *validassetapp.App,
	userAssetApp *userassetapp.App,
	approvalStatusApp *approvalstatusapp.App,
	cityApp *cityapp.App,
	streetApp *streetapp.App,
	commentApp *commentapp.App,
	approvalApp *approvalapp.App,
	reportsToApp *reportstoapp.App,
	officeApp *officeapp.App,
	homeApp *homeapp.App,
	titleApp *titleapp.App,
	inspectionApp *inspectionapp.App,
	inventoryAdjustmentApp *inventoryadjustmentapp.App,
	inventoryLocationApp *inventorylocationapp.App,
	inventoryTransactionApp *inventorytransactionapp.App,
	serialNumberApp *serialnumberapp.App,
	transferOrderApp *transferorderapp.App,
	warehouseApp *warehouseapp.App,
	zoneApp *zoneapp.App,
	inventoryItemApp *inventoryitemapp.App,
	lotTrackingsApp *lottrackingsapp.App,
	supplierApp *supplierapp.App,
	supplierProductApp *supplierproductapp.App,
	brandApp *brandapp.App,
	costHistoryApp *costhistoryapp.App,
	metricsApp *metricsapp.App,
	physicalAttributeApp *physicalattributeapp.App,
	productCategoryApp *productcategoryapp.App,
	productCostApp *productcostapp.App,
	productApp *productapp.App,
	customersApp *customersapp.App,
	orderLineItemsApp *orderlineitemsapp.App,
	ordersApp *ordersapp.App,
	lineItemFulfillmentStatusApp *lineitemfulfillmentstatusapp.App,
	orderFulfillmentStatusApp *orderfulfillmentstatusapp.App,
	formApp *formapp.App,
	formFieldApp *formfieldapp.App,
) (*formdataregistry.Registry, error) {
	registry := formdataregistry.New()

	// =========================================================================
	// CORE DOMAIN ENTITIES
	// =========================================================================

	// Register users entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "users",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app userapp.NewUser
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return userApp.Create(ctx, model.(userapp.NewUser))
		},
		CreateModel: userapp.NewUser{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app userapp.UpdateUser
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			// Note: userApp.Update gets user ID from context, not parameter
			// This may not work as expected for form data updates
			// Consider using a different approach for user updates
			return userApp.UpdateNoMid(ctx, model.(userapp.UpdateUser), id)
		},
		UpdateModel: userapp.UpdateUser{},
	}); err != nil {
		return nil, fmt.Errorf("register users: %w", err)
	}

	// =========================================================================
	// ASSETS DOMAIN ENTITIES
	// =========================================================================

	// Register assets entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "assets",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app assetapp.NewAsset
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return assetApp.Create(ctx, model.(assetapp.NewAsset))
		},
		CreateModel: assetapp.NewAsset{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app assetapp.UpdateAsset
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return assetApp.Update(ctx, model.(assetapp.UpdateAsset), id)
		},
		UpdateModel: assetapp.UpdateAsset{},
	}); err != nil {
		return nil, fmt.Errorf("register assets: %w", err)
	}

	// Register roles entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "roles",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app roleapp.NewRole
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return roleApp.Create(ctx, model.(roleapp.NewRole))
		},
		CreateModel: roleapp.NewRole{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app roleapp.UpdateRole
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return roleApp.Update(ctx, model.(roleapp.UpdateRole), id)
		},
		UpdateModel: roleapp.UpdateRole{},
	}); err != nil {
		return nil, fmt.Errorf("register roles: %w", err)
	}

	// Register table_access entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "table_access",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app tableaccessapp.NewTableAccess
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return tableAccessApp.Create(ctx, model.(tableaccessapp.NewTableAccess))
		},
		CreateModel: tableaccessapp.NewTableAccess{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app tableaccessapp.UpdateTableAccess
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return tableAccessApp.Update(ctx, model.(tableaccessapp.UpdateTableAccess), id)
		},
		UpdateModel: tableaccessapp.UpdateTableAccess{},
	}); err != nil {
		return nil, fmt.Errorf("register table_access: %w", err)
	}

	// Register user_roles entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "user_roles",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app userroleapp.NewUserRole
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return userRoleApp.Create(ctx, model.(userroleapp.NewUserRole))
		},
		CreateModel: userroleapp.NewUserRole{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app userroleapp.NewUserRole // Note: No UpdateUserRole, using New for both
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return userRoleApp.Create(ctx, model.(userroleapp.NewUserRole)) // Using Create for updates too
		},
		UpdateModel: userroleapp.NewUserRole{},
	}); err != nil {
		return nil, fmt.Errorf("register user_roles: %w", err)
	}

	// Register contact_infos entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "contact_infos",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app contactinfosapp.NewContactInfos
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return contactInfosApp.Create(ctx, model.(contactinfosapp.NewContactInfos))
		},
		CreateModel: contactinfosapp.NewContactInfos{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app contactinfosapp.UpdateContactInfos
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return contactInfosApp.Update(ctx, model.(contactinfosapp.UpdateContactInfos), id)
		},
		UpdateModel: contactinfosapp.UpdateContactInfos{},
	}); err != nil {
		return nil, fmt.Errorf("register contact_infos: %w", err)
	}

	// =========================================================================
	// ASSETS DOMAIN ENTITIES (continued)
	// =========================================================================

	// Register asset_conditions entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "asset_conditions",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app assetconditionapp.NewAssetCondition
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return assetConditionApp.Create(ctx, model.(assetconditionapp.NewAssetCondition))
		},
		CreateModel: assetconditionapp.NewAssetCondition{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app assetconditionapp.UpdateAssetCondition
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return assetConditionApp.Update(ctx, model.(assetconditionapp.UpdateAssetCondition), id)
		},
		UpdateModel: assetconditionapp.UpdateAssetCondition{},
	}); err != nil {
		return nil, fmt.Errorf("register asset_conditions: %w", err)
	}

	// Register asset_types entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "asset_types",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app assettypeapp.NewAssetType
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return assetTypeApp.Create(ctx, model.(assettypeapp.NewAssetType))
		},
		CreateModel: assettypeapp.NewAssetType{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app assettypeapp.UpdateAssetType
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return assetTypeApp.Update(ctx, model.(assettypeapp.UpdateAssetType), id)
		},
		UpdateModel: assettypeapp.UpdateAssetType{},
	}); err != nil {
		return nil, fmt.Errorf("register asset_types: %w", err)
	}

	// Register fulfillment_status entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "fulfillment_status",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app fulfillmentstatusapp.NewFulfillmentStatus
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return fulfillmentStatusApp.Create(ctx, model.(fulfillmentstatusapp.NewFulfillmentStatus))
		},
		CreateModel: fulfillmentstatusapp.NewFulfillmentStatus{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app fulfillmentstatusapp.UpdateFulfillmentStatus
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return fulfillmentStatusApp.Update(ctx, model.(fulfillmentstatusapp.UpdateFulfillmentStatus), id)
		},
		UpdateModel: fulfillmentstatusapp.UpdateFulfillmentStatus{},
	}); err != nil {
		return nil, fmt.Errorf("register fulfillment_status: %w", err)
	}

	// Register tags entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "tags",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app tagapp.NewTag
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return tagApp.Create(ctx, model.(tagapp.NewTag))
		},
		CreateModel: tagapp.NewTag{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app tagapp.UpdateTag
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return tagApp.Update(ctx, model.(tagapp.UpdateTag), id)
		},
		UpdateModel: tagapp.UpdateTag{},
	}); err != nil {
		return nil, fmt.Errorf("register tags: %w", err)
	}

	// Register asset_tags entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "asset_tags",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app assettagapp.NewAssetTag
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return assetTagApp.Create(ctx, model.(assettagapp.NewAssetTag))
		},
		CreateModel: assettagapp.NewAssetTag{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app assettagapp.UpdateAssetTag
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return assetTagApp.Update(ctx, model.(assettagapp.UpdateAssetTag), id)
		},
		UpdateModel: assettagapp.UpdateAssetTag{},
	}); err != nil {
		return nil, fmt.Errorf("register asset_tags: %w", err)
	}

	// Register valid_assets entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "valid_assets",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app validassetapp.NewValidAsset
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return validAssetApp.Create(ctx, model.(validassetapp.NewValidAsset))
		},
		CreateModel: validassetapp.NewValidAsset{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app validassetapp.UpdateValidAsset
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return validAssetApp.Update(ctx, model.(validassetapp.UpdateValidAsset), id)
		},
		UpdateModel: validassetapp.UpdateValidAsset{},
	}); err != nil {
		return nil, fmt.Errorf("register valid_assets: %w", err)
	}

	// Register user_assets entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "user_assets",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app userassetapp.NewUserAsset
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return userAssetApp.Create(ctx, model.(userassetapp.NewUserAsset))
		},
		CreateModel: userassetapp.NewUserAsset{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app userassetapp.UpdateUserAsset
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return userAssetApp.Update(ctx, model.(userassetapp.UpdateUserAsset), id)
		},
		UpdateModel: userassetapp.UpdateUserAsset{},
	}); err != nil {
		return nil, fmt.Errorf("register user_assets: %w", err)
	}

	// Register approval_status entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "approval_status",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app approvalstatusapp.NewApprovalStatus
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return approvalStatusApp.Create(ctx, model.(approvalstatusapp.NewApprovalStatus))
		},
		CreateModel: approvalstatusapp.NewApprovalStatus{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app approvalstatusapp.UpdateApprovalStatus
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return approvalStatusApp.Update(ctx, model.(approvalstatusapp.UpdateApprovalStatus), id)
		},
		UpdateModel: approvalstatusapp.UpdateApprovalStatus{},
	}); err != nil {
		return nil, fmt.Errorf("register approval_status: %w", err)
	}

	// =========================================================================
	// GEOGRAPHY DOMAIN ENTITIES
	// =========================================================================

	// Register cities entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "cities",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app cityapp.NewCity
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return cityApp.Create(ctx, model.(cityapp.NewCity))
		},
		CreateModel: cityapp.NewCity{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app cityapp.UpdateCity
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return cityApp.Update(ctx, model.(cityapp.UpdateCity), id)
		},
		UpdateModel: cityapp.UpdateCity{},
	}); err != nil {
		return nil, fmt.Errorf("register cities: %w", err)
	}

	// Register streets entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "streets",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app streetapp.NewStreet
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return streetApp.Create(ctx, model.(streetapp.NewStreet))
		},
		CreateModel: streetapp.NewStreet{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app streetapp.UpdateStreet
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return streetApp.Update(ctx, model.(streetapp.UpdateStreet), id)
		},
		UpdateModel: streetapp.UpdateStreet{},
	}); err != nil {
		return nil, fmt.Errorf("register streets: %w", err)
	}

	// =========================================================================
	// HR DOMAIN ENTITIES
	// =========================================================================

	// Register user_approval_comments entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "user_approval_comments",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app commentapp.NewUserApprovalComment
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return commentApp.Create(ctx, model.(commentapp.NewUserApprovalComment))
		},
		CreateModel: commentapp.NewUserApprovalComment{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app commentapp.UpdateUserApprovalComment
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return commentApp.Update(ctx, model.(commentapp.UpdateUserApprovalComment), id)
		},
		UpdateModel: commentapp.UpdateUserApprovalComment{},
	}); err != nil {
		return nil, fmt.Errorf("register user_approval_comments: %w", err)
	}

	// Register user_approval_status entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "user_approval_status",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app approvalapp.NewUserApprovalStatus
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return approvalApp.Create(ctx, model.(approvalapp.NewUserApprovalStatus))
		},
		CreateModel: approvalapp.NewUserApprovalStatus{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app approvalapp.UpdateUserApprovalStatus
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return approvalApp.Update(ctx, model.(approvalapp.UpdateUserApprovalStatus), id)
		},
		UpdateModel: approvalapp.UpdateUserApprovalStatus{},
	}); err != nil {
		return nil, fmt.Errorf("register user_approval_status: %w", err)
	}

	// Register reports_to entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "reports_to",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app reportstoapp.NewReportsTo
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return reportsToApp.Create(ctx, model.(reportstoapp.NewReportsTo))
		},
		CreateModel: reportstoapp.NewReportsTo{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app reportstoapp.UpdateReportsTo
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return reportsToApp.Update(ctx, model.(reportstoapp.UpdateReportsTo), id)
		},
		UpdateModel: reportstoapp.UpdateReportsTo{},
	}); err != nil {
		return nil, fmt.Errorf("register reports_to: %w", err)
	}

	// Register offices entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "offices",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app officeapp.NewOffice
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return officeApp.Create(ctx, model.(officeapp.NewOffice))
		},
		CreateModel: officeapp.NewOffice{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app officeapp.UpdateOffice
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return officeApp.Update(ctx, model.(officeapp.UpdateOffice), id)
		},
		UpdateModel: officeapp.UpdateOffice{},
	}); err != nil {
		return nil, fmt.Errorf("register offices: %w", err)
	}

	// Register homes entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "homes",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app homeapp.NewHome
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return homeApp.Create(ctx, model.(homeapp.NewHome))
		},
		CreateModel: homeapp.NewHome{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app homeapp.UpdateHome
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return homeApp.Update(ctx, model.(homeapp.UpdateHome))
		},
		UpdateModel: homeapp.UpdateHome{},
	}); err != nil {
		return nil, fmt.Errorf("register homes: %w", err)
	}

	// Register titles entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "titles",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app titleapp.NewTitle
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return titleApp.Create(ctx, model.(titleapp.NewTitle))
		},
		CreateModel: titleapp.NewTitle{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app titleapp.UpdateTitle
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return titleApp.Update(ctx, model.(titleapp.UpdateTitle), id)
		},
		UpdateModel: titleapp.UpdateTitle{},
	}); err != nil {
		return nil, fmt.Errorf("register titles: %w", err)
	}

	// =========================================================================
	// INVENTORY DOMAIN ENTITIES
	// =========================================================================

	// Register inspections entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "inspections",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app inspectionapp.NewInspection
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return inspectionApp.Create(ctx, model.(inspectionapp.NewInspection))
		},
		CreateModel: inspectionapp.NewInspection{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app inspectionapp.UpdateInspection
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return inspectionApp.Update(ctx, model.(inspectionapp.UpdateInspection), id)
		},
		UpdateModel: inspectionapp.UpdateInspection{},
	}); err != nil {
		return nil, fmt.Errorf("register inspections: %w", err)
	}

	// Register inventory_adjustments entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "inventory_adjustments",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app inventoryadjustmentapp.NewInventoryAdjustment
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return inventoryAdjustmentApp.Create(ctx, model.(inventoryadjustmentapp.NewInventoryAdjustment))
		},
		CreateModel: inventoryadjustmentapp.NewInventoryAdjustment{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app inventoryadjustmentapp.UpdateInventoryAdjustment
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return inventoryAdjustmentApp.Update(ctx, id, model.(inventoryadjustmentapp.UpdateInventoryAdjustment))
		},
		UpdateModel: inventoryadjustmentapp.UpdateInventoryAdjustment{},
	}); err != nil {
		return nil, fmt.Errorf("register inventory_adjustments: %w", err)
	}

	// Register inventory_locations entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "inventory_locations",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app inventorylocationapp.NewInventoryLocation
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return inventoryLocationApp.Create(ctx, model.(inventorylocationapp.NewInventoryLocation))
		},
		CreateModel: inventorylocationapp.NewInventoryLocation{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app inventorylocationapp.UpdateInventoryLocation
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return inventoryLocationApp.Update(ctx, model.(inventorylocationapp.UpdateInventoryLocation), id)
		},
		UpdateModel: inventorylocationapp.UpdateInventoryLocation{},
	}); err != nil {
		return nil, fmt.Errorf("register inventory_locations: %w", err)
	}

	// Register inventory_transactions entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "inventory_transactions",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app inventorytransactionapp.NewInventoryTransaction
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return inventoryTransactionApp.Create(ctx, model.(inventorytransactionapp.NewInventoryTransaction))
		},
		CreateModel: inventorytransactionapp.NewInventoryTransaction{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app inventorytransactionapp.UpdateInventoryTransaction
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return inventoryTransactionApp.Update(ctx, id, model.(inventorytransactionapp.UpdateInventoryTransaction))
		},
		UpdateModel: inventorytransactionapp.UpdateInventoryTransaction{},
	}); err != nil {
		return nil, fmt.Errorf("register inventory_transactions: %w", err)
	}

	// Register serial_numbers entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "serial_numbers",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app serialnumberapp.NewSerialNumber
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return serialNumberApp.Create(ctx, model.(serialnumberapp.NewSerialNumber))
		},
		CreateModel: serialnumberapp.NewSerialNumber{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app serialnumberapp.UpdateSerialNumber
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return serialNumberApp.Update(ctx, model.(serialnumberapp.UpdateSerialNumber), id)
		},
		UpdateModel: serialnumberapp.UpdateSerialNumber{},
	}); err != nil {
		return nil, fmt.Errorf("register serial_numbers: %w", err)
	}

	// Register transfer_orders entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "transfer_orders",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app transferorderapp.NewTransferOrder
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return transferOrderApp.Create(ctx, model.(transferorderapp.NewTransferOrder))
		},
		CreateModel: transferorderapp.NewTransferOrder{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app transferorderapp.UpdateTransferOrder
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return transferOrderApp.Update(ctx, id, model.(transferorderapp.UpdateTransferOrder))
		},
		UpdateModel: transferorderapp.UpdateTransferOrder{},
	}); err != nil {
		return nil, fmt.Errorf("register transfer_orders: %w", err)
	}

	// Register warehouses entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "warehouses",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app warehouseapp.NewWarehouse
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return warehouseApp.Create(ctx, model.(warehouseapp.NewWarehouse))
		},
		CreateModel: warehouseapp.NewWarehouse{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app warehouseapp.UpdateWarehouse
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return warehouseApp.Update(ctx, model.(warehouseapp.UpdateWarehouse), id)
		},
		UpdateModel: warehouseapp.UpdateWarehouse{},
	}); err != nil {
		return nil, fmt.Errorf("register warehouses: %w", err)
	}

	// Register zones entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "zones",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app zoneapp.NewZone
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return zoneApp.Create(ctx, model.(zoneapp.NewZone))
		},
		CreateModel: zoneapp.NewZone{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app zoneapp.UpdateZone
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return zoneApp.Update(ctx, model.(zoneapp.UpdateZone), id)
		},
		UpdateModel: zoneapp.UpdateZone{},
	}); err != nil {
		return nil, fmt.Errorf("register zones: %w", err)
	}

	// Register inventory_items entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "inventory_items",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app inventoryitemapp.NewInventoryItem
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return inventoryItemApp.Create(ctx, model.(inventoryitemapp.NewInventoryItem))
		},
		CreateModel: inventoryitemapp.NewInventoryItem{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app inventoryitemapp.UpdateInventoryItem
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return inventoryItemApp.Update(ctx, model.(inventoryitemapp.UpdateInventoryItem), id)
		},
		UpdateModel: inventoryitemapp.UpdateInventoryItem{},
	}); err != nil {
		return nil, fmt.Errorf("register inventory_items: %w", err)
	}

	// Register lot_trackings entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "lot_trackings",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app lottrackingsapp.NewLotTrackings
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return lotTrackingsApp.Create(ctx, model.(lottrackingsapp.NewLotTrackings))
		},
		CreateModel: lottrackingsapp.NewLotTrackings{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app lottrackingsapp.UpdateLotTrackings
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return lotTrackingsApp.Update(ctx, model.(lottrackingsapp.UpdateLotTrackings), id)
		},
		UpdateModel: lottrackingsapp.UpdateLotTrackings{},
	}); err != nil {
		return nil, fmt.Errorf("register lot_trackings: %w", err)
	}

	// =========================================================================
	// PROCUREMENT DOMAIN ENTITIES
	// =========================================================================

	// Register suppliers entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "suppliers",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app supplierapp.NewSupplier
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return supplierApp.Create(ctx, model.(supplierapp.NewSupplier))
		},
		CreateModel: supplierapp.NewSupplier{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app supplierapp.UpdateSupplier
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return supplierApp.Update(ctx, model.(supplierapp.UpdateSupplier), id)
		},
		UpdateModel: supplierapp.UpdateSupplier{},
	}); err != nil {
		return nil, fmt.Errorf("register suppliers: %w", err)
	}

	// Register supplier_products entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "supplier_products",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app supplierproductapp.NewSupplierProduct
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return supplierProductApp.Create(ctx, model.(supplierproductapp.NewSupplierProduct))
		},
		CreateModel: supplierproductapp.NewSupplierProduct{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app supplierproductapp.UpdateSupplierProduct
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return supplierProductApp.Update(ctx, model.(supplierproductapp.UpdateSupplierProduct), id)
		},
		UpdateModel: supplierproductapp.UpdateSupplierProduct{},
	}); err != nil {
		return nil, fmt.Errorf("register supplier_products: %w", err)
	}

	// =========================================================================
	// PRODUCTS DOMAIN ENTITIES
	// =========================================================================

	// Register brands entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "brands",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app brandapp.NewBrand
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return brandApp.Create(ctx, model.(brandapp.NewBrand))
		},
		CreateModel: brandapp.NewBrand{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app brandapp.UpdateBrand
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return brandApp.Update(ctx, model.(brandapp.UpdateBrand), id)
		},
		UpdateModel: brandapp.UpdateBrand{},
	}); err != nil {
		return nil, fmt.Errorf("register brands: %w", err)
	}

	// Register cost_history entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "cost_history",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app costhistoryapp.NewCostHistory
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return costHistoryApp.Create(ctx, model.(costhistoryapp.NewCostHistory))
		},
		CreateModel: costhistoryapp.NewCostHistory{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app costhistoryapp.UpdateCostHistory
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return costHistoryApp.Update(ctx, model.(costhistoryapp.UpdateCostHistory), id)
		},
		UpdateModel: costhistoryapp.UpdateCostHistory{},
	}); err != nil {
		return nil, fmt.Errorf("register cost_history: %w", err)
	}

	// Register metrics entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "metrics",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app metricsapp.NewMetric
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return metricsApp.Create(ctx, model.(metricsapp.NewMetric))
		},
		CreateModel: metricsapp.NewMetric{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app metricsapp.UpdateMetric
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return metricsApp.Update(ctx, model.(metricsapp.UpdateMetric), id)
		},
		UpdateModel: metricsapp.UpdateMetric{},
	}); err != nil {
		return nil, fmt.Errorf("register metrics: %w", err)
	}

	// Register physical_attributes entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "physical_attributes",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app physicalattributeapp.NewPhysicalAttribute
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return physicalAttributeApp.Create(ctx, model.(physicalattributeapp.NewPhysicalAttribute))
		},
		CreateModel: physicalattributeapp.NewPhysicalAttribute{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app physicalattributeapp.UpdatePhysicalAttribute
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return physicalAttributeApp.Update(ctx, model.(physicalattributeapp.UpdatePhysicalAttribute), id)
		},
		UpdateModel: physicalattributeapp.UpdatePhysicalAttribute{},
	}); err != nil {
		return nil, fmt.Errorf("register physical_attributes: %w", err)
	}

	// Register product_categories entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "product_categories",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app productcategoryapp.NewProductCategory
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return productCategoryApp.Create(ctx, model.(productcategoryapp.NewProductCategory))
		},
		CreateModel: productcategoryapp.NewProductCategory{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app productcategoryapp.UpdateProductCategory
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return productCategoryApp.Update(ctx, model.(productcategoryapp.UpdateProductCategory), id)
		},
		UpdateModel: productcategoryapp.UpdateProductCategory{},
	}); err != nil {
		return nil, fmt.Errorf("register product_categories: %w", err)
	}

	// Register product_costs entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "product_costs",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app productcostapp.NewProductCost
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return productCostApp.Create(ctx, model.(productcostapp.NewProductCost))
		},
		CreateModel: productcostapp.NewProductCost{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app productcostapp.UpdateProductCost
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return productCostApp.Update(ctx, model.(productcostapp.UpdateProductCost), id)
		},
		UpdateModel: productcostapp.UpdateProductCost{},
	}); err != nil {
		return nil, fmt.Errorf("register product_costs: %w", err)
	}

	// Register products entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "products",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app productapp.NewProduct
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return productApp.Create(ctx, model.(productapp.NewProduct))
		},
		CreateModel: productapp.NewProduct{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app productapp.UpdateProduct
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return productApp.Update(ctx, model.(productapp.UpdateProduct), id)
		},
		UpdateModel: productapp.UpdateProduct{},
	}); err != nil {
		return nil, fmt.Errorf("register products: %w", err)
	}

	// =========================================================================
	// SALES DOMAIN ENTITIES
	// =========================================================================

	// Register customers entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "customers",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app customersapp.NewCustomers
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return customersApp.Create(ctx, model.(customersapp.NewCustomers))
		},
		CreateModel: customersapp.NewCustomers{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app customersapp.UpdateCustomers
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return customersApp.Update(ctx, model.(customersapp.UpdateCustomers), id)
		},
		UpdateModel: customersapp.UpdateCustomers{},
	}); err != nil {
		return nil, fmt.Errorf("register customers: %w", err)
	}

	// Register order_line_items entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "order_line_items",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app orderlineitemsapp.NewOrderLineItem
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return orderLineItemsApp.Create(ctx, model.(orderlineitemsapp.NewOrderLineItem))
		},
		CreateModel: orderlineitemsapp.NewOrderLineItem{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app orderlineitemsapp.UpdateOrderLineItem
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return orderLineItemsApp.Update(ctx, model.(orderlineitemsapp.UpdateOrderLineItem), id)
		},
		UpdateModel: orderlineitemsapp.UpdateOrderLineItem{},
	}); err != nil {
		return nil, fmt.Errorf("register order_line_items: %w", err)
	}

	// Register orders entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "orders",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app ordersapp.NewOrder
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return ordersApp.Create(ctx, model.(ordersapp.NewOrder))
		},
		CreateModel: ordersapp.NewOrder{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app ordersapp.UpdateOrder
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return ordersApp.Update(ctx, model.(ordersapp.UpdateOrder), id)
		},
		UpdateModel: ordersapp.UpdateOrder{},
	}); err != nil {
		return nil, fmt.Errorf("register orders: %w", err)
	}

	// Register line_item_fulfillment_status entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "line_item_fulfillment_status",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app lineitemfulfillmentstatusapp.NewLineItemFulfillmentStatus
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return lineItemFulfillmentStatusApp.Create(ctx, model.(lineitemfulfillmentstatusapp.NewLineItemFulfillmentStatus))
		},
		CreateModel: lineitemfulfillmentstatusapp.NewLineItemFulfillmentStatus{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app lineitemfulfillmentstatusapp.UpdateLineItemFulfillmentStatus
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return lineItemFulfillmentStatusApp.Update(ctx, model.(lineitemfulfillmentstatusapp.UpdateLineItemFulfillmentStatus), id)
		},
		UpdateModel: lineitemfulfillmentstatusapp.UpdateLineItemFulfillmentStatus{},
	}); err != nil {
		return nil, fmt.Errorf("register line_item_fulfillment_status: %w", err)
	}

	// Register order_fulfillment_status entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "order_fulfillment_status",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app orderfulfillmentstatusapp.NewOrderFulfillmentStatus
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return orderFulfillmentStatusApp.Create(ctx, model.(orderfulfillmentstatusapp.NewOrderFulfillmentStatus))
		},
		CreateModel: orderfulfillmentstatusapp.NewOrderFulfillmentStatus{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app orderfulfillmentstatusapp.UpdateOrderFulfillmentStatus
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return orderFulfillmentStatusApp.Update(ctx, model.(orderfulfillmentstatusapp.UpdateOrderFulfillmentStatus), id)
		},
		UpdateModel: orderfulfillmentstatusapp.UpdateOrderFulfillmentStatus{},
	}); err != nil {
		return nil, fmt.Errorf("register order_fulfillment_status: %w", err)
	}

	// =========================================================================
	// CONFIG DOMAIN ENTITIES
	// =========================================================================

	// Register forms entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "forms",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app formapp.NewForm
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return formApp.Create(ctx, model.(formapp.NewForm))
		},
		CreateModel: formapp.NewForm{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app formapp.UpdateForm
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return formApp.Update(ctx, model.(formapp.UpdateForm), id)
		},
		UpdateModel: formapp.UpdateForm{},
	}); err != nil {
		return nil, fmt.Errorf("register forms: %w", err)
	}

	// Register form_fields entity
	if err := registry.Register(formdataregistry.EntityRegistration{
		Name: "form_fields",
		DecodeNew: func(data json.RawMessage) (interface{}, error) {
			var app formfieldapp.NewFormField
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
			return formFieldApp.Create(ctx, model.(formfieldapp.NewFormField))
		},
		CreateModel: formfieldapp.NewFormField{},
		DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
			var app formfieldapp.UpdateFormField
			if err := json.Unmarshal(data, &app); err != nil {
				return nil, err
			}
			if err := app.Validate(); err != nil {
				return nil, err
			}
			return app, nil
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
			return formFieldApp.Update(ctx, model.(formfieldapp.UpdateFormField), id)
		},
		UpdateModel: formfieldapp.UpdateFormField{},
	}); err != nil {
		return nil, fmt.Errorf("register form_fields: %w", err)
	}

	return registry, nil
}
