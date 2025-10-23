package all

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/assets/assetapp"
	"github.com/timmaaaz/ichor/app/domain/core/userapp"
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

	// =========================================================================
	// TODO: ADD MORE ENTITY REGISTRATIONS HERE
	// =========================================================================
	//
	// Follow the pattern above for each new entity:
	// 1. Import the app package
	// 2. Add registration block
	// 3. Test with forms
	//
	// Common entities to add:
	// - products
	// - customers
	// - orders
	// - addresses
	// - inventory items
	// - etc.

	return registry, nil
}
