package pagecontentdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/config/pagecontentbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for page content database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the API for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (pagecontentbus.Storer, error) {
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

// Create inserts a new page content block into the database.
func (s *Store) Create(ctx context.Context, content pagecontentbus.PageContent) error {
	const q = `
		INSERT INTO config.page_content (
			id, page_config_id, content_type, label,
			table_config_id, form_id, order_index,
			parent_id, layout, is_visible, is_default
		) VALUES (
			:id, :page_config_id, :content_type, :label,
			:table_config_id, :form_id, :order_index,
			:parent_id, :layout, :is_visible, :is_default
		)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPageContent(content)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies an existing page content block in the database.
func (s *Store) Update(ctx context.Context, content pagecontentbus.PageContent) error {
	const q = `
		UPDATE config.page_content SET
			page_config_id = :page_config_id,
			content_type = :content_type,
			label = :label,
			table_config_id = :table_config_id,
			form_id = :form_id,
			order_index = :order_index,
			parent_id = :parent_id,
			layout = :layout,
			is_visible = :is_visible,
			is_default = :is_default
		WHERE id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPageContent(content)); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return fmt.Errorf("namedexeccontext: %w", pagecontentbus.ErrNotFound)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a page content block from the database.
func (s *Store) Delete(ctx context.Context, contentID uuid.UUID) error {
	data := struct {
		ID uuid.UUID `db:"id"`
	}{
		ID: contentID,
	}

	const q = `DELETE FROM config.page_content WHERE id = :id`

	rowsAffected, err := sqldb.NamedExecContextWithCount(ctx, s.log, s.db, q, data)
	if err != nil {
		return fmt.Errorf("namedexeccontextwithcount: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("namedexeccontextwithcount: %w", pagecontentbus.ErrNotFound)
	}

	return nil
}

// Query retrieves a list of page content blocks based on filters.
func (s *Store) Query(ctx context.Context, filter pagecontentbus.QueryFilter, orderBy order.By, pageReq page.Page) ([]pagecontentbus.PageContent, error) {
	data := map[string]any{
		"offset":        (pageReq.Number() - 1) * pageReq.RowsPerPage(),
		"rows_per_page": pageReq.RowsPerPage(),
	}

	const q = `
	SELECT
		id, page_config_id, content_type, label,
		table_config_id, form_id, order_index,
		parent_id, layout, is_visible, is_default
	FROM
		config.page_content`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbContents []dbPageContent
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbContents); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusPageContents(dbContents), nil
}

// Count returns the total number of page content blocks matching the filter.
func (s *Store) Count(ctx context.Context, filter pagecontentbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		config.page_content`

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

// QueryByID retrieves a single page content block by ID.
func (s *Store) QueryByID(ctx context.Context, contentID uuid.UUID) (pagecontentbus.PageContent, error) {
	data := struct {
		ID uuid.UUID `db:"id"`
	}{
		ID: contentID,
	}

	const q = `
		SELECT
			id, page_config_id, content_type, label,
			table_config_id, form_id, order_index,
			parent_id, layout, is_visible, is_default
		FROM
			config.page_content
		WHERE
			id = :id`

	var dbContent dbPageContent
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbContent); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return pagecontentbus.PageContent{}, fmt.Errorf("namedquerystruct: %w", pagecontentbus.ErrNotFound)
		}
		return pagecontentbus.PageContent{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusPageContent(dbContent), nil
}

// QueryByPageConfigID retrieves all content blocks for a specific page config.
func (s *Store) QueryByPageConfigID(ctx context.Context, pageConfigID uuid.UUID) ([]pagecontentbus.PageContent, error) {
	data := struct {
		PageConfigID uuid.UUID `db:"page_config_id"`
	}{
		PageConfigID: pageConfigID,
	}

	const q = `
		SELECT
			id, page_config_id, content_type, label,
			table_config_id, form_id, order_index,
			parent_id, layout, is_visible, is_default
		FROM
			config.page_content
		WHERE
			page_config_id = :page_config_id
		ORDER BY
			order_index ASC`

	var dbContents []dbPageContent
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbContents); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusPageContents(dbContents), nil
}

// QueryWithChildren retrieves content blocks and nests children under parents.
// This is especially useful for tabs, where tab items are children of a tabs container.
func (s *Store) QueryWithChildren(ctx context.Context, pageConfigID uuid.UUID) ([]pagecontentbus.PageContent, error) {
	// Query all content for this page config
	allContent, err := s.QueryByPageConfigID(ctx, pageConfigID)
	if err != nil {
		return nil, err
	}

	// Build map for quick lookup
	contentMap := make(map[uuid.UUID]*pagecontentbus.PageContent)
	for i := range allContent {
		contentMap[allContent[i].ID] = &allContent[i]
		allContent[i].Children = []pagecontentbus.PageContent{} // Initialize children slice
	}

	// Nest children under parents and collect top-level IDs
	topLevelIDs := []uuid.UUID{}
	for i := range allContent {
		if allContent[i].ParentID == uuid.Nil {
			// Top-level content (no parent) - save ID for later
			topLevelIDs = append(topLevelIDs, allContent[i].ID)
		} else {
			// Child content - add to parent's children slice
			if parent, ok := contentMap[allContent[i].ParentID]; ok {
				parent.Children = append(parent.Children, allContent[i])
			}
		}
	}

	// Build top-level result from map (after children have been populated)
	topLevel := make([]pagecontentbus.PageContent, 0, len(topLevelIDs))
	for _, id := range topLevelIDs {
		if content, ok := contentMap[id]; ok {
			topLevel = append(topLevel, *content)
		}
	}

	return topLevel, nil
}
