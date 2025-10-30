package pageactiondb

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
)

// createButtonAction inserts a button action into the database.
func (s *Store) createButtonAction(ctx context.Context, baseAction dbPageAction, buttonData pageactionbus.ButtonAction) error {
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

	// Insert button-specific data
	const qButton = `
	INSERT INTO config.page_action_buttons (
		action_id, label, icon, target_path, variant, alignment, confirmation_prompt
	) VALUES (
		:action_id, :label, :icon, :target_path, :variant, :alignment, :confirmation_prompt
	)`

	dbButton := toDBButtonAction(baseAction.ID, buttonData)
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, qButton, dbButton); err != nil {
		return fmt.Errorf("namedexeccontext button: %w", err)
	}

	return nil
}

// updateButtonAction updates a button action in the database.
func (s *Store) updateButtonAction(ctx context.Context, baseAction dbPageAction, buttonData pageactionbus.ButtonAction) error {
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

	// Update button-specific data
	const qButton = `
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

	dbButton := toDBButtonAction(baseAction.ID, buttonData)
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, qButton, dbButton); err != nil {
		return fmt.Errorf("namedexeccontext button: %w", err)
	}

	return nil
}

// queryButtonByID retrieves a button action by its ID.
func (s *Store) queryButtonByID(ctx context.Context, actionID uuid.UUID) (pageactionbus.PageAction, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: actionID.String(),
	}

	const q = `
	SELECT
		a.id, a.page_config_id, a.action_type, a.action_order, a.is_active,
		b.label, b.icon, b.target_path, b.variant, b.alignment, b.confirmation_prompt
	FROM
		config.page_actions a
	INNER JOIN
		config.page_action_buttons b ON a.id = b.action_id
	WHERE
		a.id = :id`

	var result struct {
		dbPageAction
		Label              string `db:"label"`
		Icon               string `db:"icon"`
		TargetPath         string `db:"target_path"`
		Variant            string `db:"variant"`
		Alignment          string `db:"alignment"`
		ConfirmationPrompt string `db:"confirmation_prompt"`
	}

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &result); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return pageactionbus.PageAction{}, pageactionbus.ErrNotFound
		}
		return pageactionbus.PageAction{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	action := toBusPageAction(result.dbPageAction)
	action.Button = &pageactionbus.ButtonAction{
		Label:              result.Label,
		Icon:               result.Icon,
		TargetPath:         result.TargetPath,
		Variant:            result.Variant,
		Alignment:          result.Alignment,
		ConfirmationPrompt: result.ConfirmationPrompt,
	}

	return action, nil
}

// queryButtonsByPageConfigID retrieves all button actions for a page config.
func (s *Store) queryButtonsByPageConfigID(ctx context.Context, pageConfigID uuid.UUID) ([]pageactionbus.PageAction, error) {
	data := struct {
		PageConfigID string `db:"page_config_id"`
	}{
		PageConfigID: pageConfigID.String(),
	}

	const q = `
	SELECT
		a.id, a.page_config_id, a.action_type, a.action_order, a.is_active,
		b.label, b.icon, b.target_path, b.variant, b.alignment, b.confirmation_prompt
	FROM
		config.page_actions a
	INNER JOIN
		config.page_action_buttons b ON a.id = b.action_id
	WHERE
		a.page_config_id = :page_config_id
	ORDER BY
		a.action_order ASC, a.id ASC`

	var results []struct {
		dbPageAction
		Label              string `db:"label"`
		Icon               string `db:"icon"`
		TargetPath         string `db:"target_path"`
		Variant            string `db:"variant"`
		Alignment          string `db:"alignment"`
		ConfirmationPrompt string `db:"confirmation_prompt"`
	}

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &results); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	actions := make([]pageactionbus.PageAction, len(results))
	for i, result := range results {
		action := toBusPageAction(result.dbPageAction)
		action.Button = &pageactionbus.ButtonAction{
			Label:              result.Label,
			Icon:               result.Icon,
			TargetPath:         result.TargetPath,
			Variant:            result.Variant,
			Alignment:          result.Alignment,
			ConfirmationPrompt: result.ConfirmationPrompt,
		}
		actions[i] = action
	}

	return actions, nil
}
