package pageactiondb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for page actions database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (pageactionbus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// CreateBaseAction inserts only the base action record (atomic operation).
func (s *Store) CreateBaseAction(ctx context.Context, action pageactionbus.PageAction) error {
	const q = `
	INSERT INTO config.page_actions (
		id, page_config_id, action_type, action_order, is_active
	) VALUES (
		:id, :page_config_id, :action_type, :action_order, :is_active
	)`

	dbAction := toDBPageAction(action)
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbAction); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", pageactionbus.ErrUniqueEntry)
		}
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", pageactionbus.ErrForeignKeyViolation)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// CreateButtonData inserts only button-specific data (atomic operation).
func (s *Store) CreateButtonData(ctx context.Context, actionID uuid.UUID, button pageactionbus.ButtonAction) error {
	const q = `
	INSERT INTO config.page_action_buttons (
		action_id, label, icon, target_path, variant, alignment, confirmation_prompt
	) VALUES (
		:action_id, :label, :icon, :target_path, :variant, :alignment, :confirmation_prompt
	)`

	dbButton := toDBButtonAction(actionID, button)
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbButton); err != nil {
		return fmt.Errorf("namedexeccontext button: %w", err)
	}

	return nil
}

// CreateDropdownData inserts only dropdown-specific data (atomic operation).
func (s *Store) CreateDropdownData(ctx context.Context, actionID uuid.UUID, dropdown pageactionbus.DropdownAction) error {
	const q = `
	INSERT INTO config.page_action_dropdowns (
		action_id, label, icon
	) VALUES (
		:action_id, :label, :icon
	)`

	dbDropdown := toDBDropdownAction(actionID, dropdown)
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbDropdown); err != nil {
		return fmt.Errorf("namedexeccontext dropdown: %w", err)
	}

	return nil
}

// CreateDropdownItem inserts a single dropdown item (atomic operation).
func (s *Store) CreateDropdownItem(ctx context.Context, dropdownActionID uuid.UUID, item pageactionbus.NewDropdownItem) error {
	const q = `
	INSERT INTO config.page_action_dropdown_items (
		id, dropdown_action_id, label, target_path, item_order
	) VALUES (
		:id, :dropdown_action_id, :label, :target_path, :item_order
	)`

	dbItem := toDBDropdownItem(dropdownActionID, item, uuid.New())
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbItem); err != nil {
		return fmt.Errorf("namedexeccontext item: %w", err)
	}

	return nil
}

// UpdateBaseAction updates only the base action record (atomic operation).
func (s *Store) UpdateBaseAction(ctx context.Context, action pageactionbus.PageAction) error {
	const q = `
	UPDATE config.page_actions
	SET
		page_config_id = :page_config_id,
		action_order = :action_order,
		is_active = :is_active
	WHERE
		id = :id`

	dbAction := toDBPageAction(action)
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbAction); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// UpdateButtonData updates only button-specific data (atomic operation).
func (s *Store) UpdateButtonData(ctx context.Context, actionID uuid.UUID, button pageactionbus.ButtonAction) error {
	const q = `
	UPDATE config.page_action_buttons
	SET
		label = :label,
		icon = :icon,
		target_path = :target_path,
		variant = :variant,
		alignment = :alignment,
		confirmation_prompt = :confirmation_prompt
	WHERE
		action_id = :action_id`

	dbButton := toDBButtonAction(actionID, button)
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbButton); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// UpdateDropdownData updates only dropdown-specific data (atomic operation).
func (s *Store) UpdateDropdownData(ctx context.Context, actionID uuid.UUID, dropdown pageactionbus.DropdownAction) error {
	const q = `
	UPDATE config.page_action_dropdowns
	SET
		label = :label,
		icon = :icon
	WHERE
		action_id = :action_id`

	dbDropdown := toDBDropdownAction(actionID, dropdown)
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbDropdown); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// DeleteDropdownItems deletes all items for a dropdown (atomic operation).
func (s *Store) DeleteDropdownItems(ctx context.Context, dropdownActionID uuid.UUID) error {
	const q = `
	DELETE FROM config.page_action_dropdown_items
	WHERE dropdown_action_id = :dropdown_action_id`

	data := map[string]any{"dropdown_action_id": dropdownActionID}
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a page action from the database (cascades to type-specific tables).
func (s *Store) Delete(ctx context.Context, action pageactionbus.PageAction) error {
	const q = `
	DELETE FROM config.page_actions
	WHERE id = :id`

	data := struct {
		ID uuid.UUID `db:"id"`
	}{
		ID: action.ID,
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of page actions from the database.
// Note: This returns base actions only; use QueryByID for full action data.
func (s *Store) Query(ctx context.Context, filter pageactionbus.QueryFilter, orderBy order.By, pg page.Page) ([]pageactionbus.PageAction, error) {
	data := map[string]any{
		"offset":        (pg.Number() - 1) * pg.RowsPerPage(),
		"rows_per_page": pg.RowsPerPage(),
	}

	const q = `
	SELECT
		id, page_config_id, action_type, action_order, is_active
	FROM
		config.page_actions`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbActions []dbPageAction
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbActions); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusPageActions(dbActions), nil
}

// Count returns the total number of page actions matching the filter.
func (s *Store) Count(ctx context.Context, filter pageactionbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		config.page_actions`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}

// QueryByID retrieves a single page action with full type-specific data by its ID.
func (s *Store) QueryByID(ctx context.Context, actionID uuid.UUID) (pageactionbus.PageAction, error) {
	// First, get the action type
	data := struct {
		ID string `db:"id"`
	}{
		ID: actionID.String(),
	}

	const qType = `
	SELECT action_type
	FROM config.page_actions
	WHERE id = :id`

	var actionType struct {
		ActionType string `db:"action_type"`
	}

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, qType, data, &actionType); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return pageactionbus.PageAction{}, pageactionbus.ErrNotFound
		}
		return pageactionbus.PageAction{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	// Route to appropriate query based on type
	switch pageactionbus.ActionType(actionType.ActionType) {
	case pageactionbus.ActionTypeButton:
		return s.queryButtonByID(ctx, actionID)
	case pageactionbus.ActionTypeDropdown:
		return s.queryDropdownByID(ctx, actionID)
	case pageactionbus.ActionTypeSeparator:
		return s.querySeparatorByID(ctx, actionID)
	default:
		return pageactionbus.PageAction{}, fmt.Errorf("unknown action type: %s", actionType.ActionType)
	}
}

// querySeparatorByID retrieves a separator action by its ID.
func (s *Store) querySeparatorByID(ctx context.Context, actionID uuid.UUID) (pageactionbus.PageAction, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: actionID.String(),
	}

	const q = `
	SELECT
		id, page_config_id, action_type, action_order, is_active
	FROM
		config.page_actions
	WHERE
		id = :id`

	var dbAction dbPageAction
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbAction); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return pageactionbus.PageAction{}, pageactionbus.ErrNotFound
		}
		return pageactionbus.PageAction{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusPageAction(dbAction), nil
}

// QueryByPageConfigID retrieves all actions for a page config, grouped by type.
func (s *Store) QueryByPageConfigID(ctx context.Context, pageConfigID uuid.UUID) (pageactionbus.ActionsGroupedByType, error) {
	var result pageactionbus.ActionsGroupedByType

	// Query buttons
	buttons, err := s.queryButtonsByPageConfigID(ctx, pageConfigID)
	if err != nil {
		return result, fmt.Errorf("query buttons: %w", err)
	}
	result.Buttons = buttons

	// Query dropdowns
	dropdowns, err := s.queryDropdownsByPageConfigID(ctx, pageConfigID)
	if err != nil {
		return result, fmt.Errorf("query dropdowns: %w", err)
	}
	result.Dropdowns = dropdowns

	// Query separators
	separators, err := s.querySeparatorsByPageConfigID(ctx, pageConfigID)
	if err != nil {
		return result, fmt.Errorf("query separators: %w", err)
	}
	result.Separators = separators

	return result, nil
}

// querySeparatorsByPageConfigID retrieves all separator actions for a page config.
func (s *Store) querySeparatorsByPageConfigID(ctx context.Context, pageConfigID uuid.UUID) ([]pageactionbus.PageAction, error) {
	data := struct {
		PageConfigID string        `db:"page_config_id"`
		ActionType   string        `db:"action_type"`
	}{
		PageConfigID: pageConfigID.String(),
		ActionType:   string(pageactionbus.ActionTypeSeparator),
	}

	const q = `
	SELECT
		id, page_config_id, action_type, action_order, is_active
	FROM
		config.page_actions
	WHERE
		page_config_id = :page_config_id
		AND action_type = :action_type
	ORDER BY
		action_order ASC, id ASC`

	var dbActions []dbPageAction
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbActions); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusPageActions(dbActions), nil
}
