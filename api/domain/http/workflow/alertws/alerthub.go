package alertws

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/websocket"
)

// ID prefix constants for string-based Hub registration.
const (
	userIDPrefix = "user:"
	roleIDPrefix = "role:"
)

// AlertHub wraps the foundation Hub with user/role semantics.
type AlertHub struct {
	hub         *websocket.Hub
	userRoleBus *userrolebus.Business
	log         *logger.Logger
}

// NewAlertHub creates a new AlertHub.
func NewAlertHub(hub *websocket.Hub, userRoleBus *userrolebus.Business, log *logger.Logger) *AlertHub {
	return &AlertHub{
		hub:         hub,
		userRoleBus: userRoleBus,
		log:         log,
	}
}

// Hub returns the underlying foundation Hub.
func (ah *AlertHub) Hub() *websocket.Hub {
	return ah.hub
}

// RegisterClient fetches user roles and registers the client with user and role IDs.
func (ah *AlertHub) RegisterClient(ctx context.Context, client *websocket.Client, userID uuid.UUID) error {
	// Fetch user roles for role-targeted broadcasts
	roles, err := ah.userRoleBus.QueryByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("fetching user roles: %w", err)
	}

	// Build ID list: user ID + all role IDs
	ids := make([]string, 0, 1+len(roles))
	ids = append(ids, userIDToString(userID))
	for _, role := range roles {
		ids = append(ids, roleIDToString(role.RoleID))
	}

	// Register with foundation Hub using string IDs
	ah.hub.Register(ctx, client, ids)

	ah.log.Info(ctx, "alert client registered",
		"user_id", userID,
		"role_count", len(roles))

	return nil
}

// BroadcastToUser sends a message to all connections for a specific user.
func (ah *AlertHub) BroadcastToUser(userID uuid.UUID, message []byte) int {
	return ah.hub.BroadcastToID(userIDToString(userID), message)
}

// BroadcastToRole sends a message to all connections with a specific role.
func (ah *AlertHub) BroadcastToRole(roleID uuid.UUID, message []byte) int {
	return ah.hub.BroadcastToID(roleIDToString(roleID), message)
}

// BroadcastAll sends a message to all connected clients.
func (ah *AlertHub) BroadcastAll(message []byte) int {
	return ah.hub.BroadcastAll(message)
}

// RefreshUserRoles updates the role mappings for all connections of a user.
// Called when delegate fires role change event.
func (ah *AlertHub) RefreshUserRoles(ctx context.Context, userID uuid.UUID) error {
	// Fetch fresh roles
	roles, err := ah.userRoleBus.QueryByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("fetching updated roles: %w", err)
	}

	// Build new ID list
	ids := make([]string, 0, 1+len(roles))
	ids = append(ids, userIDToString(userID))
	for _, role := range roles {
		ids = append(ids, roleIDToString(role.RoleID))
	}

	// Update all clients for this user (exact ID match, not prefix)
	ah.hub.UpdateClientIDsForID(ctx, userIDToString(userID), ids)

	ah.log.Debug(ctx, "refreshed roles for alert connections",
		"user_id", userID,
		"role_count", len(roles))

	return nil
}

// Helper functions for ID conversion
func userIDToString(id uuid.UUID) string {
	return userIDPrefix + id.String()
}

func roleIDToString(id uuid.UUID) string {
	return roleIDPrefix + id.String()
}
