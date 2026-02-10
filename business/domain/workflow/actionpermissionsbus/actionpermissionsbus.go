// Package actionpermissionsbus provides business logic for managing action permissions
// that control which roles can execute workflow actions manually.
package actionpermissionsbus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound = errors.New("action permission not found")
	ErrUnique   = errors.New("action permission already exists for role")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, ap ActionPermission) error
	Update(ctx context.Context, ap ActionPermission) error
	Delete(ctx context.Context, ap ActionPermission) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, pg page.Page) ([]ActionPermission, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, id uuid.UUID) (ActionPermission, error)
	QueryByRoleAndAction(ctx context.Context, roleID uuid.UUID, actionType string) (ActionPermission, error)
	QueryByRoleIDs(ctx context.Context, roleIDs []uuid.UUID, actionType string) ([]ActionPermission, error)
}

// Business manages action permission operations.
type Business struct {
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs an action permissions business API for use.
func NewBusiness(log *logger.Logger, storer Storer) *Business {
	return &Business{
		log:    log,
		storer: storer,
	}
}

// NewWithTx constructs a new business value that will use the
// specified transaction in any store related calls.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	bus := Business{
		log:    b.log,
		storer: storer,
	}

	return &bus, nil
}

// Create adds a new action permission to the system.
func (b *Business) Create(ctx context.Context, nap NewActionPermission) (ActionPermission, error) {
	ctx, span := otel.AddSpan(ctx, "business.actionpermissionsbus.create")
	defer span.End()

	now := time.Now()

	// Set default empty constraints if not provided
	constraints := nap.Constraints
	if constraints == nil {
		constraints = json.RawMessage("{}")
	}

	ap := ActionPermission{
		ID:          uuid.New(),
		RoleID:      nap.RoleID,
		ActionType:  nap.ActionType,
		IsAllowed:   nap.IsAllowed,
		Constraints: constraints,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := b.storer.Create(ctx, ap); err != nil {
		if errors.Is(err, ErrUnique) {
			return ActionPermission{}, ErrUnique
		}
		return ActionPermission{}, fmt.Errorf("create action permission: %w", err)
	}

	return ap, nil
}

// Update modifies an existing action permission.
func (b *Business) Update(ctx context.Context, ap ActionPermission, uap UpdateActionPermission) (ActionPermission, error) {
	ctx, span := otel.AddSpan(ctx, "business.actionpermissionsbus.update")
	defer span.End()

	if uap.IsAllowed != nil {
		ap.IsAllowed = *uap.IsAllowed
	}

	if uap.Constraints != nil {
		ap.Constraints = *uap.Constraints
	}

	ap.UpdatedAt = time.Now()

	if err := b.storer.Update(ctx, ap); err != nil {
		return ActionPermission{}, fmt.Errorf("update action permission: %w", err)
	}

	return ap, nil
}

// Delete removes an action permission from the system.
func (b *Business) Delete(ctx context.Context, ap ActionPermission) error {
	ctx, span := otel.AddSpan(ctx, "business.actionpermissionsbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, ap); err != nil {
		return fmt.Errorf("delete action permission: %w", err)
	}

	return nil
}

// Query returns a list of action permissions based on filter criteria.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pg page.Page) ([]ActionPermission, error) {
	ctx, span := otel.AddSpan(ctx, "business.actionpermissionsbus.query")
	defer span.End()

	perms, err := b.storer.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return nil, fmt.Errorf("query action permissions: %w", err)
	}

	return perms, nil
}

// Count returns the total number of action permissions based on filter criteria.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.actionpermissionsbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID returns a single action permission by its ID.
func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (ActionPermission, error) {
	ctx, span := otel.AddSpan(ctx, "business.actionpermissionsbus.querybyid")
	defer span.End()

	ap, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return ActionPermission{}, fmt.Errorf("query action permission: id[%s]: %w", id, err)
	}

	return ap, nil
}

// QueryByRoleAndAction returns a permission by role ID and action type.
func (b *Business) QueryByRoleAndAction(ctx context.Context, roleID uuid.UUID, actionType string) (ActionPermission, error) {
	ctx, span := otel.AddSpan(ctx, "business.actionpermissionsbus.querybyroleandaction")
	defer span.End()

	ap, err := b.storer.QueryByRoleAndAction(ctx, roleID, actionType)
	if err != nil {
		return ActionPermission{}, fmt.Errorf("query action permission: roleID[%s] actionType[%s]: %w", roleID, actionType, err)
	}

	return ap, nil
}

// CanUserExecuteAction checks if a user (via their roles) has permission to execute
// a specific action type. Returns true if ANY of the user's roles grants permission.
func (b *Business) CanUserExecuteAction(ctx context.Context, userID uuid.UUID, actionType string, userRoles []uuid.UUID) (bool, error) {
	ctx, span := otel.AddSpan(ctx, "business.actionpermissionsbus.canuserexecuteaction")
	defer span.End()

	if len(userRoles) == 0 {
		return false, nil
	}

	perms, err := b.storer.QueryByRoleIDs(ctx, userRoles, actionType)
	if err != nil {
		return false, fmt.Errorf("query permissions for roles: %w", err)
	}

	// If any role has permission with is_allowed=true, user can execute
	for _, perm := range perms {
		if perm.IsAllowed {
			return true, nil
		}
	}

	return false, nil
}

// GetAllowedActionsForRoles returns all action types that are allowed for the given roles.
func (b *Business) GetAllowedActionsForRoles(ctx context.Context, roleIDs []uuid.UUID) ([]string, error) {
	ctx, span := otel.AddSpan(ctx, "business.actionpermissionsbus.getallowedactionsforroles")
	defer span.End()

	if len(roleIDs) == 0 {
		return []string{}, nil
	}

	// Query all permissions for these roles
	filter := QueryFilter{}
	perms, err := b.storer.Query(ctx, filter, DefaultOrderBy, page.MustParse("1", "1000"))
	if err != nil {
		return nil, fmt.Errorf("query permissions: %w", err)
	}

	// Build map of role IDs for quick lookup
	roleSet := make(map[uuid.UUID]bool)
	for _, id := range roleIDs {
		roleSet[id] = true
	}

	// Collect unique action types that are allowed
	allowedActions := make(map[string]bool)
	for _, perm := range perms {
		if roleSet[perm.RoleID] && perm.IsAllowed {
			allowedActions[perm.ActionType] = true
		}
	}

	// Convert to slice
	result := make([]string, 0, len(allowedActions))
	for actionType := range allowedActions {
		result = append(result, actionType)
	}

	return result, nil
}
