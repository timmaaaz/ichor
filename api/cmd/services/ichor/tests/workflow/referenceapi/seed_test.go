package reference_test

import (
	"context"
	"fmt"

	"github.com/timmaaaz/ichor/api/domain/http/workflow/referenceapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// ReferenceSeedData holds reference data test data.
type ReferenceSeedData struct {
	apitest.SeedData
	TriggerTypes []workflow.TriggerType
	EntityTypes  []workflow.EntityType
	Entities     []workflow.Entity
}

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (ReferenceSeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	// Create regular user
	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return ReferenceSeedData{}, fmt.Errorf("seeding users: %w", err)
	}
	tu1 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
	}

	// Seed trigger types (or get existing)
	triggerTypes, err := workflow.TestSeedTriggerTypes(ctx, 3, busDomain.Workflow)
	if err != nil {
		return ReferenceSeedData{}, fmt.Errorf("seeding trigger types: %w", err)
	}

	// Get entity types (already seeded in database migrations)
	entityTypes, err := workflow.GetEntityTypes(ctx, busDomain.Workflow)
	if err != nil {
		return ReferenceSeedData{}, fmt.Errorf("getting entity types: %w", err)
	}

	// Get entities (already seeded in database migrations)
	entities, err := workflow.GetEntities(ctx, busDomain.Workflow)
	if err != nil {
		return ReferenceSeedData{}, fmt.Errorf("getting entities: %w", err)
	}

	return ReferenceSeedData{
		SeedData: apitest.SeedData{
			Users: []apitest.User{tu1},
		},
		TriggerTypes: triggerTypes,
		EntityTypes:  entityTypes,
		Entities:     entities,
	}, nil
}

// toAppTriggerType converts a business trigger type to an API response.
func toAppTriggerType(tt workflow.TriggerType) referenceapi.TriggerType {
	return referenceapi.TriggerType{
		ID:          tt.ID,
		Name:        tt.Name,
		Description: tt.Description,
		IsActive:    tt.IsActive,
	}
}

// toAppEntityType converts a business entity type to an API response.
func toAppEntityType(et workflow.EntityType) referenceapi.EntityType {
	return referenceapi.EntityType{
		ID:          et.ID,
		Name:        et.Name,
		Description: et.Description,
		IsActive:    et.IsActive,
	}
}

// toAppEntity converts a business entity to an API response.
func toAppEntity(e workflow.Entity) referenceapi.Entity {
	return referenceapi.Entity{
		ID:           e.ID,
		Name:         e.Name,
		EntityTypeID: e.EntityTypeID,
		SchemaName:   e.SchemaName,
		IsActive:     e.IsActive,
	}
}
