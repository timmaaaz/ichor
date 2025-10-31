package pageactiondb

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
)

// createDropdownAction inserts a dropdown action with items into the database.
func (s *Store) createDropdownAction(ctx context.Context, baseAction dbPageAction, dropdownData pageactionbus.DropdownAction, items []pageactionbus.NewDropdownItem) error {
	// Insert base action first
	const qBase = `
	INSERT INTO config.page_actions (
		id, page_config_id, action_type, action_order, is_active
	) VALUES (
		:id, :page_config_id, :action_type, :action_order, :is_active
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, qBase, baseAction); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", pageactionbus.ErrUniqueEntry)
		}
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", pageactionbus.ErrForeignKeyViolation)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	// Insert dropdown-specific data
	const qDropdown = `
	INSERT INTO config.page_action_dropdowns (
		action_id, label, icon
	) VALUES (
		:action_id, :label, :icon
	)`

	dbDropdown := toDBDropdownAction(baseAction.ID, dropdownData)
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, qDropdown, dbDropdown); err != nil {
		return fmt.Errorf("namedexeccontext dropdown: %w", err)
	}

	// Insert dropdown items
	if len(items) > 0 {
		const qItems = `
		INSERT INTO config.page_action_dropdown_items (
			id, dropdown_action_id, label, target_path, item_order
		) VALUES (
			:id, :dropdown_action_id, :label, :target_path, :item_order
		)`

		for _, item := range items {
			dbItem := toDBDropdownItem(baseAction.ID, item, uuid.New())
			if err := sqldb.NamedExecContext(ctx, s.log, s.db, qItems, dbItem); err != nil {
				return fmt.Errorf("namedexeccontext item: %w", err)
			}
		}
	}

	return nil
}

// updateDropdownAction updates a dropdown action and replaces all items.
func (s *Store) updateDropdownAction(ctx context.Context, baseAction dbPageAction, dropdownData pageactionbus.DropdownAction, newItems []pageactionbus.NewDropdownItem) error {
	// Update base action
	const qBase = `
	UPDATE config.page_actions
	SET
		page_config_id = :page_config_id,
		action_order = :action_order,
		is_active = :is_active
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, qBase, baseAction); err != nil {
		return fmt.Errorf("namedexeccontext base: %w", err)
	}

	// Update dropdown-specific data
	const qDropdown = `
	UPDATE config.page_action_dropdowns
	SET
		label = :label,
		icon = :icon
	WHERE
		action_id = :action_id`

	dbDropdown := toDBDropdownAction(baseAction.ID, dropdownData)
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, qDropdown, dbDropdown); err != nil {
		return fmt.Errorf("namedexeccontext dropdown: %w", err)
	}

	// Delete existing items
	const qDeleteItems = `
	DELETE FROM config.page_action_dropdown_items
	WHERE dropdown_action_id = :action_id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, qDeleteItems, map[string]any{"action_id": baseAction.ID}); err != nil {
		return fmt.Errorf("namedexeccontext delete items: %w", err)
	}

	// Insert new items
	if len(newItems) > 0 {
		const qItems = `
		INSERT INTO config.page_action_dropdown_items (
			id, dropdown_action_id, label, target_path, item_order
		) VALUES (
			:id, :dropdown_action_id, :label, :target_path, :item_order
		)`

		for _, item := range newItems {
			dbItem := toDBDropdownItem(baseAction.ID, item, uuid.New())
			if err := sqldb.NamedExecContext(ctx, s.log, s.db, qItems, dbItem); err != nil {
				return fmt.Errorf("namedexeccontext item: %w", err)
			}
		}
	}

	return nil
}

// queryDropdownByID retrieves a dropdown action with its items by ID.
func (s *Store) queryDropdownByID(ctx context.Context, actionID uuid.UUID) (pageactionbus.PageAction, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: actionID.String(),
	}

	// Query dropdown action
	const qDropdown = `
	SELECT
		a.id, a.page_config_id, a.action_type, a.action_order, a.is_active,
		d.label, d.icon
	FROM
		config.page_actions a
	INNER JOIN
		config.page_action_dropdowns d ON a.id = d.action_id
	WHERE
		a.id = :id`

	var dropdownResult struct {
		dbPageAction
		Label string `db:"label"`
		Icon  string `db:"icon"`
	}

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, qDropdown, data, &dropdownResult); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return pageactionbus.PageAction{}, pageactionbus.ErrNotFound
		}
		return pageactionbus.PageAction{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	// Query dropdown items
	const qItems = `
	SELECT
		id, dropdown_action_id, label, target_path, item_order
	FROM
		config.page_action_dropdown_items
	WHERE
		dropdown_action_id = :id
	ORDER BY
		item_order ASC, id ASC`

	var items []dbDropdownItem
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, qItems, data, &items); err != nil {
		return pageactionbus.PageAction{}, fmt.Errorf("namedqueryslice items: %w", err)
	}

	action := toBusPageAction(dropdownResult.dbPageAction)
	dbDropdown := dbDropdownAction{
		ActionID: action.ID,
		Label:    dropdownResult.Label,
		Icon:     dropdownResult.Icon,
	}
	dropdown := toBusDropdownAction(dbDropdown, items)
	action.Dropdown = &dropdown

	return action, nil
}

// queryDropdownsByPageConfigID retrieves all dropdown actions for a page config.
func (s *Store) queryDropdownsByPageConfigID(ctx context.Context, pageConfigID uuid.UUID) ([]pageactionbus.PageAction, error) {
	data := struct {
		PageConfigID string `db:"page_config_id"`
	}{
		PageConfigID: pageConfigID.String(),
	}

	// Query all dropdown actions
	const qDropdowns = `
	SELECT
		a.id, a.page_config_id, a.action_type, a.action_order, a.is_active,
		d.label, d.icon
	FROM
		config.page_actions a
	INNER JOIN
		config.page_action_dropdowns d ON a.id = d.action_id
	WHERE
		a.page_config_id = :page_config_id
	ORDER BY
		a.action_order ASC, a.id ASC`

	var dropdownResults []struct {
		dbPageAction
		Label string `db:"label"`
		Icon  string `db:"icon"`
	}

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, qDropdowns, data, &dropdownResults); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	// For each dropdown, query its items
	actions := make([]pageactionbus.PageAction, len(dropdownResults))
	for i, result := range dropdownResults {
		action := toBusPageAction(result.dbPageAction)

		// Query items for this dropdown
		const qItems = `
		SELECT
			id, dropdown_action_id, label, target_path, item_order
		FROM
			config.page_action_dropdown_items
		WHERE
			dropdown_action_id = :action_id
		ORDER BY
			item_order ASC, id ASC`

		itemData := struct {
			ActionID string `db:"action_id"`
		}{
			ActionID: action.ID.String(),
		}

		var items []dbDropdownItem
		if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, qItems, itemData, &items); err != nil {
			return nil, fmt.Errorf("namedqueryslice items: %w", err)
		}

		dbDropdown := dbDropdownAction{
			ActionID: action.ID,
			Label:    result.Label,
			Icon:     result.Icon,
		}
		dropdown := toBusDropdownAction(dbDropdown, items)
		action.Dropdown = &dropdown

		actions[i] = action
	}

	return actions, nil
}
