package fulfillmentstatusdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for fulfillment status database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (fulfillmentstatusbus.Storer, error) {
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

// Create inserts a new fulfillment status into the database.
func (s *Store) Create(ctx context.Context, fs fulfillmentstatusbus.FulfillmentStatus) error {
	const q = `
    INSERT INTO assets.fulfillment_status (
        id, icon_id, name, primary_color, secondary_color, icon
    ) VALUES (
        :id, :icon_id, :name, :primary_color, :secondary_color, :icon
    )
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBFulfillmentStatus(fs)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", fulfillmentstatusbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update replaces an approval status document in the database.
func (s *Store) Update(ctx context.Context, fs fulfillmentstatusbus.FulfillmentStatus) error {
	const q = `
	UPDATE assets.fulfillment_status
	SET
	    icon_id = :icon_id,
        name = :name,
        primary_color = :primary_color,
        secondary_color = :secondary_color,
        icon = :icon
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBFulfillmentStatus(fs)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", fulfillmentstatusbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes an approval status from the database.
func (s *Store) Delete(ctx context.Context, as fulfillmentstatusbus.FulfillmentStatus) error {
	const q = `
	DELETE FROM
		assets.fulfillment_status
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBFulfillmentStatus(as)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of approval statuses from the database.
func (s *Store) Query(ctx context.Context, filter fulfillmentstatusbus.QueryFilter, orderBy order.By, page page.Page) ([]fulfillmentstatusbus.FulfillmentStatus, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, icon_id, name, primary_color, secondary_color, icon
	FROM
		assets.fulfillment_status
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var fulfillmentStatuses []fulfillmentStatus

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &fulfillmentStatuses); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusFulfillmentStatuses(fulfillmentStatuses), nil
}

// Count returns the total number of approval statuses
func (s *Store) Count(ctx context.Context, filter fulfillmentstatusbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        assets.fulfillment_status`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerysingle: %w", err)
	}

	return count.Count, nil
}

// QueryByID finds the approval status by the specified ID.
func (s *Store) QueryByID(ctx context.Context, aprvlStatusID uuid.UUID) (fulfillmentstatusbus.FulfillmentStatus, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: aprvlStatusID.String(),
	}

	const q = `
    SELECT
        id, icon_id, name, primary_color, secondary_color, icon
    FROM
        assets.fulfillment_status
    WHERE
        id = :id
    `

	var fulfillmentStatus fulfillmentStatus
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &fulfillmentStatus); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return fulfillmentstatusbus.FulfillmentStatus{}, fmt.Errorf("db: %w", fulfillmentstatusbus.ErrNotFound)
		}
		return fulfillmentstatusbus.FulfillmentStatus{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusFulfillmentStatus(fulfillmentStatus), nil
}
