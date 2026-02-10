package actionpermissionsbus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// TestSeedActionPermissions seeds action permissions for testing.
// It creates permissions for the given roles and action types.
func TestSeedActionPermissions(ctx context.Context, api *Business, roleID uuid.UUID, actionTypes []string) ([]ActionPermission, error) {
	permissions := make([]ActionPermission, len(actionTypes))

	for i, actionType := range actionTypes {
		perm, err := api.Create(ctx, NewActionPermission{
			RoleID:      roleID,
			ActionType:  actionType,
			IsAllowed:   true,
			Constraints: json.RawMessage("{}"),
		})
		if err != nil {
			return nil, fmt.Errorf("creating action permission: %w", err)
		}

		permissions[i] = perm
	}

	return permissions, nil
}

// TestSeedActionPermission seeds a single action permission for testing.
func TestSeedActionPermission(ctx context.Context, api *Business, roleID uuid.UUID, actionType string, isAllowed bool) (ActionPermission, error) {
	perm, err := api.Create(ctx, NewActionPermission{
		RoleID:      roleID,
		ActionType:  actionType,
		IsAllowed:   isAllowed,
		Constraints: json.RawMessage("{}"),
	})
	if err != nil {
		return ActionPermission{}, fmt.Errorf("creating action permission: %w", err)
	}

	return perm, nil
}
