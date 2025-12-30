// Package pageactionbus provides business logic for page action configuration management.
package pageactionbus

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound            = errors.New("page action not found")
	ErrUniqueEntry         = errors.New("page action entry is not unique")
	ErrForeignKeyViolation = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data. All create/update operations are atomic single-table operations.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)

	// Atomic create operations (single table each)
	CreateBaseAction(ctx context.Context, action PageAction) error
	CreateButtonData(ctx context.Context, actionID uuid.UUID, button ButtonAction) error
	CreateDropdownData(ctx context.Context, actionID uuid.UUID, dropdown DropdownAction) error
	CreateDropdownItem(ctx context.Context, dropdownActionID uuid.UUID, item NewDropdownItem) error

	// Atomic update operations
	UpdateBaseAction(ctx context.Context, action PageAction) error
	UpdateButtonData(ctx context.Context, actionID uuid.UUID, button ButtonAction) error
	UpdateDropdownData(ctx context.Context, actionID uuid.UUID, dropdown DropdownAction) error
	DeleteDropdownItems(ctx context.Context, dropdownActionID uuid.UUID) error

	// Delete
	Delete(ctx context.Context, action PageAction) error

	// Query operations
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]PageAction, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, actionID uuid.UUID) (PageAction, error)
	QueryByPageConfigID(ctx context.Context, pageConfigID uuid.UUID) (ActionsGroupedByType, error)
}

// Business manages the set of APIs for page action access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a page action business API for use.
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:      log,
		delegate: delegate,
		storer:   storer,
	}
}

// NewWithTx constructs a new Business value replacing the Storer
// value with a Storer value that is currently inside a transaction.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	return &Business{
		log:      b.log,
		delegate: b.delegate,
		storer:   storer,
	}, nil
}

// CreateButton inserts a new button action into the database.
func (b *Business) CreateButton(ctx context.Context, nba NewButtonAction) (PageAction, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageactionbus.createbutton")
	defer span.End()

	action := PageAction{
		ID:           uuid.New(),
		PageConfigID: nba.PageConfigID,
		ActionType:   ActionTypeButton,
		ActionOrder:  nba.ActionOrder,
		IsActive:     nba.IsActive,
		Button: &ButtonAction{
			Label:              nba.Label,
			Icon:               nba.Icon,
			TargetPath:         nba.TargetPath,
			Variant:            nba.Variant,
			Alignment:          nba.Alignment,
			ConfirmationPrompt: nba.ConfirmationPrompt,
		},
	}

	if err := b.storer.CreateBaseAction(ctx, action); err != nil {
		return PageAction{}, fmt.Errorf("create base: %w", err)
	}

	if err := b.storer.CreateButtonData(ctx, action.ID, *action.Button); err != nil {
		return PageAction{}, fmt.Errorf("create button data: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(action)); err != nil {
		b.log.Error(ctx, "pageactionbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return action, nil
}

// CreateDropdown inserts a new dropdown action with items into the database.
func (b *Business) CreateDropdown(ctx context.Context, nda NewDropdownAction) (PageAction, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageactionbus.createdropdown")
	defer span.End()

	action := PageAction{
		ID:           uuid.New(),
		PageConfigID: nda.PageConfigID,
		ActionType:   ActionTypeDropdown,
		ActionOrder:  nda.ActionOrder,
		IsActive:     nda.IsActive,
		Dropdown: &DropdownAction{
			Label: nda.Label,
			Icon:  nda.Icon,
		},
	}

	if err := b.storer.CreateBaseAction(ctx, action); err != nil {
		return PageAction{}, fmt.Errorf("create base: %w", err)
	}

	if err := b.storer.CreateDropdownData(ctx, action.ID, *action.Dropdown); err != nil {
		return PageAction{}, fmt.Errorf("create dropdown data: %w", err)
	}

	for _, item := range nda.Items {
		if err := b.storer.CreateDropdownItem(ctx, action.ID, item); err != nil {
			return PageAction{}, fmt.Errorf("create dropdown item: %w", err)
		}
	}

	createdAction, err := b.storer.QueryByID(ctx, action.ID)
	if err != nil {
		return PageAction{}, fmt.Errorf("querybyid: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(createdAction)); err != nil {
		b.log.Error(ctx, "pageactionbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return createdAction, nil
}

// CreateSeparator inserts a new separator action into the database.
func (b *Business) CreateSeparator(ctx context.Context, nsa NewSeparatorAction) (PageAction, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageactionbus.createseparator")
	defer span.End()

	action := PageAction{
		ID:           uuid.New(),
		PageConfigID: nsa.PageConfigID,
		ActionType:   ActionTypeSeparator,
		ActionOrder:  nsa.ActionOrder,
		IsActive:     nsa.IsActive,
	}

	if err := b.storer.CreateBaseAction(ctx, action); err != nil {
		return PageAction{}, fmt.Errorf("create base: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(action)); err != nil {
		b.log.Error(ctx, "pageactionbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return action, nil
}

// UpdateButton replaces a button action in the database.
func (b *Business) UpdateButton(ctx context.Context, action PageAction, uba UpdateButtonAction) (PageAction, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageactionbus.updatebutton")
	defer span.End()

	if action.Button == nil {
		return PageAction{}, fmt.Errorf("action is not a button")
	}

	if uba.PageConfigID != nil {
		action.PageConfigID = *uba.PageConfigID
	}

	if uba.ActionOrder != nil {
		action.ActionOrder = *uba.ActionOrder
	}

	if uba.IsActive != nil {
		action.IsActive = *uba.IsActive
	}

	if uba.Label != nil {
		action.Button.Label = *uba.Label
	}

	if uba.Icon != nil {
		action.Button.Icon = *uba.Icon
	}

	if uba.TargetPath != nil {
		action.Button.TargetPath = *uba.TargetPath
	}

	if uba.Variant != nil {
		action.Button.Variant = *uba.Variant
	}

	if uba.Alignment != nil {
		action.Button.Alignment = *uba.Alignment
	}

	if uba.ConfirmationPrompt != nil {
		action.Button.ConfirmationPrompt = *uba.ConfirmationPrompt
	}

	if err := b.storer.UpdateBaseAction(ctx, action); err != nil {
		return PageAction{}, fmt.Errorf("update base: %w", err)
	}

	if err := b.storer.UpdateButtonData(ctx, action.ID, *action.Button); err != nil {
		return PageAction{}, fmt.Errorf("update button data: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(action)); err != nil {
		b.log.Error(ctx, "pageactionbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return action, nil
}

// UpdateDropdown replaces a dropdown action and its items in the database.
func (b *Business) UpdateDropdown(ctx context.Context, action PageAction, uda UpdateDropdownAction) (PageAction, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageactionbus.updatedropdown")
	defer span.End()

	if action.Dropdown == nil {
		return PageAction{}, fmt.Errorf("action is not a dropdown")
	}

	if uda.PageConfigID != nil {
		action.PageConfigID = *uda.PageConfigID
	}

	if uda.ActionOrder != nil {
		action.ActionOrder = *uda.ActionOrder
	}

	if uda.IsActive != nil {
		action.IsActive = *uda.IsActive
	}

	if uda.Label != nil {
		action.Dropdown.Label = *uda.Label
	}

	if uda.Icon != nil {
		action.Dropdown.Icon = *uda.Icon
	}

	var items []NewDropdownItem
	if uda.Items != nil {
		items = *uda.Items
	} else {
		// Keep existing items
		items = make([]NewDropdownItem, len(action.Dropdown.Items))
		for i, item := range action.Dropdown.Items {
			items[i] = NewDropdownItem{
				Label:      item.Label,
				TargetPath: item.TargetPath,
				ItemOrder:  item.ItemOrder,
			}
		}
	}

	if err := b.storer.UpdateBaseAction(ctx, action); err != nil {
		return PageAction{}, fmt.Errorf("update base: %w", err)
	}

	if err := b.storer.UpdateDropdownData(ctx, action.ID, *action.Dropdown); err != nil {
		return PageAction{}, fmt.Errorf("update dropdown data: %w", err)
	}

	if err := b.storer.DeleteDropdownItems(ctx, action.ID); err != nil {
		return PageAction{}, fmt.Errorf("delete dropdown items: %w", err)
	}

	for _, item := range items {
		if err := b.storer.CreateDropdownItem(ctx, action.ID, item); err != nil {
			return PageAction{}, fmt.Errorf("create dropdown item: %w", err)
		}
	}

	updatedAction, err := b.storer.QueryByID(ctx, action.ID)
	if err != nil {
		return PageAction{}, fmt.Errorf("querybyid: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(updatedAction)); err != nil {
		b.log.Error(ctx, "pageactionbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return updatedAction, nil
}

// UpdateSeparator replaces a separator action in the database.
func (b *Business) UpdateSeparator(ctx context.Context, action PageAction, usa UpdateSeparatorAction) (PageAction, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageactionbus.updateseparator")
	defer span.End()

	if action.ActionType != ActionTypeSeparator {
		return PageAction{}, fmt.Errorf("action is not a separator")
	}

	if usa.PageConfigID != nil {
		action.PageConfigID = *usa.PageConfigID
	}

	if usa.ActionOrder != nil {
		action.ActionOrder = *usa.ActionOrder
	}

	if usa.IsActive != nil {
		action.IsActive = *usa.IsActive
	}

	if err := b.storer.UpdateBaseAction(ctx, action); err != nil {
		return PageAction{}, fmt.Errorf("update base: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(action)); err != nil {
		b.log.Error(ctx, "pageactionbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return action, nil
}

// Delete removes the specified page action.
func (b *Business) Delete(ctx context.Context, action PageAction) error {
	ctx, span := otel.AddSpan(ctx, "business.pageactionbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, action); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(action)); err != nil {
		b.log.Error(ctx, "pageactionbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves a list of page actions from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pg page.Page) ([]PageAction, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageactionbus.query")
	defer span.End()

	actions, err := b.storer.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	// Fetch full details for each action (including button/dropdown data)
	fullActions := make([]PageAction, len(actions))
	for i, action := range actions {
		fullAction, err := b.storer.QueryByID(ctx, action.ID)
		if err != nil {
			return nil, fmt.Errorf("querybyid: %w", err)
		}
		fullActions[i] = fullAction
	}

	return fullActions, nil
}

// Count returns the total number of page actions.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageactionbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the page action by the specified ID.
func (b *Business) QueryByID(ctx context.Context, actionID uuid.UUID) (PageAction, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageactionbus.querybyid")
	defer span.End()

	action, err := b.storer.QueryByID(ctx, actionID)
	if err != nil {
		return PageAction{}, fmt.Errorf("query: actionID[%s]: %w", actionID, err)
	}

	return action, nil
}

// QueryByPageConfigID retrieves all actions for a specific page config, grouped by type.
func (b *Business) QueryByPageConfigID(ctx context.Context, pageConfigID uuid.UUID) (ActionsGroupedByType, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageactionbus.querybypageconfigid")
	defer span.End()

	actions, err := b.storer.QueryByPageConfigID(ctx, pageConfigID)
	if err != nil {
		return ActionsGroupedByType{}, fmt.Errorf("query: pageConfigID[%s]: %w", pageConfigID, err)
	}

	return actions, nil
}
