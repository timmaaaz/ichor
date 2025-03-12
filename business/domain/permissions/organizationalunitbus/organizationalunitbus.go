package organizationalunitbus

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/convert"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("role not found")
	ErrUnique                = errors.New("organizational unit is not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, ou OrganizationalUnit) error
	Update(ctx context.Context, ou OrganizationalUnit) error
	Delete(ctx context.Context, ou OrganizationalUnit) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]OrganizationalUnit, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, userID uuid.UUID) (OrganizationalUnit, error)
	QueryByParentID(ctx context.Context, parentID uuid.UUID) ([]OrganizationalUnit, error)
}

// Business manages the set of APIs for user access.
type Business struct {
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs a user business API for use.
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

// Create adds a new org unit to the system
func (b *Business) Create(ctx context.Context, nou NewOrganizationalUnit) (OrganizationalUnit, error) {
	ctx, span := otel.AddSpan(ctx, "business.organizationalunitbus.create")
	defer span.End()

	var path string
	var level int

	if nou.ParentID != uuid.Nil {
		parent, err := b.storer.QueryByID(ctx, nou.ParentID)
		if err != nil {
			return OrganizationalUnit{}, fmt.Errorf("querying parent organizational unit: %w", err)
		}

		path = fmt.Sprintf("%s.%s", parent.Path, strings.ReplaceAll(nou.Name, " ", "_"))
		level = parent.Level + 1 // Calculate level based on parent's level
	} else {
		path = strings.ReplaceAll(nou.Name, " ", "_") // Root level units have their name as the path
		level = 0                                     // Root level is 0
	}

	ou := OrganizationalUnit{
		ID:                    uuid.New(),
		ParentID:              nou.ParentID,
		Name:                  nou.Name,
		Level:                 level, // Use calculated level instead of nou.Level
		Path:                  path,
		CanInheritPermissions: nou.CanInheritPermissions,
		CanRollupData:         nou.CanRollupData,
		UnitType:              nou.UnitType,
		IsActive:              nou.IsActive,
	}

	if err := b.storer.Create(ctx, ou); err != nil {
		return OrganizationalUnit{}, fmt.Errorf("creating organizational unit: %w", err)
	}

	return ou, nil
}

// // Update modifies a org unit in the system
// func (b *Business) Update(ctx context.Context, ou OrganizationalUnit, uou UpdateOrganizationalUnit) (OrganizationalUnit, error) {
// 	ctx, span := otel.AddSpan(ctx, "business.organizationalroleunit.update")
// 	defer span.End()

// 	err := convert.PopulateSameTypes(uou, &ou)
// 	if err != nil {
// 		return OrganizationalUnit{}, fmt.Errorf("populate same types: %w", err)
// 	}

// 	if err := b.storer.Update(ctx, ou); err != nil {
// 		return OrganizationalUnit{}, fmt.Errorf("updating organizational unit: %w", err)
// 	}

//		return ou, nil
//	}
//

// Update modifies an org unit in the system and updates paths for all children
func (b *Business) Update(ctx context.Context, ou OrganizationalUnit, uou UpdateOrganizationalUnit) (OrganizationalUnit, error) {
	ctx, span := otel.AddSpan(ctx, "business.organizationalroleunit.update")
	defer span.End()

	// Store the original name and path before updating
	originalName := ou.Name
	originalPath := ou.Path

	// Populate the fields to update
	err := convert.PopulateSameTypes(uou, &ou)
	if err != nil {
		return OrganizationalUnit{}, fmt.Errorf("populate same types: %w", err)
	}

	// Check if name has changed - this affects paths
	nameChanged := uou.Name != nil && *uou.Name != originalName

	// If name changed, we need to update the path for this unit
	if nameChanged {
		// Get parent path if this isn't a root node
		if ou.ParentID != uuid.Nil {
			parent, err := b.storer.QueryByID(ctx, ou.ParentID)
			if err != nil {
				return OrganizationalUnit{}, fmt.Errorf("querying parent organizational unit: %w", err)
			}

			// Update path based on parent's path and new name
			ou.Path = fmt.Sprintf("%s.%s", parent.Path, strings.ReplaceAll(ou.Name, " ", "_"))
		} else {
			// Root level units have their name as the path
			ou.Path = strings.ReplaceAll(ou.Name, " ", "_")
		}

		// Since we're calculating path, the path field should not be directly set by the caller
		// If it was provided, override it with our calculated value
		if uou.Path != nil {
			*uou.Path = ou.Path
		}
	}

	// Update the organizational unit
	if err := b.storer.Update(ctx, ou); err != nil {
		return OrganizationalUnit{}, fmt.Errorf("updating organizational unit: %w", err)
	}

	// If name changed or parent ID changed, update all children's paths
	// We store the original parentID for proper comparison
	originalParentID := ou.ParentID
	parentChanged := uou.ParentID != nil && *uou.ParentID != originalParentID

	if nameChanged || parentChanged {
		if err := b.updateChildrenPaths(ctx, ou.ID, originalPath, ou.Path); err != nil {
			return OrganizationalUnit{}, fmt.Errorf("updating children paths: %w", err)
		}
	}

	return ou, nil
}

// Delete removes an org unit and all its descendants from the system.
func (b *Business) Delete(ctx context.Context, ou OrganizationalUnit) error {
	ctx, span := otel.AddSpan(ctx, "business.organizationalunitbus.delete")
	defer span.End()

	// Recursively delete all descendants first
	if err := b.deleteDescendants(ctx, ou.ID); err != nil {
		return fmt.Errorf("deleting descendants: %w", err)
	}

	// Then delete the unit itself
	if err := b.storer.Delete(ctx, ou); err != nil {
		return fmt.Errorf("deleting organizational unit: %w", err)
	}

	return nil
}

// Query retrieves a list of org units from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]OrganizationalUnit, error) {
	ctx, span := otel.AddSpan(ctx, "business.organizationalunitbus.query")
	defer span.End()

	ous, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("querying organizational units: %w", err)
	}

	return ous, nil
}

// Count returns the total number of org units.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.organizationalunitbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID retrieves a single org unit from the system.
func (b *Business) QueryByID(ctx context.Context, userID uuid.UUID) (OrganizationalUnit, error) {
	ctx, span := otel.AddSpan(ctx, "business.organizationalunitbus.querybyid")
	defer span.End()

	ou, err := b.storer.QueryByID(ctx, userID)
	if err != nil {
		return OrganizationalUnit{}, fmt.Errorf("querying organizational unit: %w", err)
	}

	return ou, nil
}

// QueryByParentID retrieves a list of org units that are children of a given parent.
func (b *Business) QueryByParentID(ctx context.Context, parentID uuid.UUID) ([]OrganizationalUnit, error) {
	ctx, span := otel.AddSpan(ctx, "business.organizationalunitbus.querybyparentid")
	defer span.End()

	ous, err := b.storer.QueryByParentID(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("querying organizational units by parent ID: %w", err)
	}

	return ous, nil
}

// updateChildrenPaths recursively updates paths for all children of a given org unit
func (b *Business) updateChildrenPaths(ctx context.Context, parentID uuid.UUID, oldParentPath, newParentPath string) error {
	ctx, span := otel.AddSpan(ctx, "business.organizationalroleunit.updateChildrenPaths")
	defer span.End()

	// Get all immediate children
	children, err := b.storer.QueryByParentID(ctx, parentID)
	if err != nil {
		return fmt.Errorf("querying children: %w", err)
	}

	// No children, nothing to update
	if len(children) == 0 {
		return nil
	}

	// Update each child's path and its children recursively
	for _, child := range children {
		// Store original path before updating
		originalChildPath := child.Path

		// Extract the node-specific part of the path (the last segment)
		pathParts := strings.Split(child.Path, ".")
		nodePath := pathParts[len(pathParts)-1]

		// Construct new path by combining new parent path with node's own name
		newChildPath := fmt.Sprintf("%s.%s", newParentPath, nodePath)

		// Update this child's path directly in storage to avoid recursive Update calls
		child.Path = newChildPath
		if err := b.storer.Update(ctx, child); err != nil {
			return fmt.Errorf("updating child path: %w", err)
		}

		// Recursively update this child's children
		if err := b.updateChildrenPaths(ctx, child.ID, originalChildPath, newChildPath); err != nil {
			return fmt.Errorf("updating grandchildren paths: %w", err)
		}
	}
	return nil
}

// deleteDescendants recursively deletes all descendants of a given org unit
func (b *Business) deleteDescendants(ctx context.Context, parentID uuid.UUID) error {
	ctx, span := otel.AddSpan(ctx, "business.organizationalunitbus.deleteDescendants")
	defer span.End()

	// Get all immediate children
	children, err := b.storer.QueryByParentID(ctx, parentID)
	if err != nil {
		return fmt.Errorf("querying children: %w", err)
	}

	// No children, nothing to delete
	if len(children) == 0 {
		return nil
	}

	// Recursively delete each child's descendants, then the child itself
	for _, child := range children {
		if err := b.deleteDescendants(ctx, child.ID); err != nil {
			return fmt.Errorf("deleting child descendants: %w", err)
		}

		if err := b.storer.Delete(ctx, child); err != nil {
			return fmt.Errorf("deleting child: %w", err)
		}
	}

	return nil
}
