package notificationdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for notification database access.
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

// NewWithTx constructs a new Store value replacing the sqlx DB with a sqlx DB
// value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (notificationbus.Storer, error) {
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

// Create inserts a new notification into the database.
func (s *Store) Create(ctx context.Context, notification notificationbus.Notification) error {
	const q = `
	INSERT INTO workflow.notifications
		(id, user_id, title, message, priority, is_read, read_date, source_entity_name, source_entity_id, action_url, created_date)
	VALUES
		(:id, :user_id, :title, :message, :priority, :is_read, :read_date, :source_entity_name, :source_entity_id, :action_url, :created_date)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBNotification(notification)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// QueryByID gets the specified notification from the database.
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (notificationbus.Notification, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: id.String(),
	}

	const q = `
	SELECT
		id, user_id, title, message, priority, is_read, read_date,
		source_entity_name, source_entity_id, action_url, created_date
	FROM
		workflow.notifications
	WHERE
		id = :id`

	var dbN dbNotification
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbN); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return notificationbus.Notification{}, fmt.Errorf("db: %w", notificationbus.ErrNotFound)
		}
		return notificationbus.Notification{}, fmt.Errorf("db: %w", err)
	}

	return toBusNotification(dbN), nil
}

// Query retrieves a list of notifications from the database.
func (s *Store) Query(ctx context.Context, filter notificationbus.QueryFilter, orderBy order.By, pg page.Page) ([]notificationbus.Notification, error) {
	data := map[string]any{
		"offset":        (pg.Number() - 1) * pg.RowsPerPage(),
		"rows_per_page": pg.RowsPerPage(),
	}

	var buf bytes.Buffer
	buf.WriteString(`
	SELECT
		id, user_id, title, message, priority, is_read, read_date,
		source_entity_name, source_entity_id, action_url, created_date
	FROM
		workflow.notifications`)

	applyFilter(filter, data, &buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}
	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbNs []dbNotification
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbNs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusNotifications(dbNs), nil
}

// Count returns the total number of notifications matching the filter.
func (s *Store) Count(ctx context.Context, filter notificationbus.QueryFilter) (int, error) {
	data := map[string]any{}

	var buf bytes.Buffer
	buf.WriteString(`
	SELECT
		count(1)
	FROM
		workflow.notifications`)

	applyFilter(filter, data, &buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	return count.Count, nil
}

// MarkAsRead updates a single notification's is_read and read_date.
func (s *Store) MarkAsRead(ctx context.Context, id uuid.UUID, readDate time.Time) error {
	data := struct {
		ID       string    `db:"id"`
		ReadDate time.Time `db:"read_date"`
	}{
		ID:       id.String(),
		ReadDate: readDate,
	}

	const q = `
	UPDATE
		workflow.notifications
	SET
		is_read = true,
		read_date = :read_date
	WHERE
		id = :id`

	rows, err := sqldb.NamedExecContextWithCount(ctx, s.log, s.db, q, data)
	if err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("db: %w", notificationbus.ErrNotFound)
	}

	return nil
}

// MarkAllAsRead marks all unread notifications for a user as read.
func (s *Store) MarkAllAsRead(ctx context.Context, userID uuid.UUID, readDate time.Time) (int, error) {
	data := struct {
		UserID   string    `db:"user_id"`
		ReadDate time.Time `db:"read_date"`
	}{
		UserID:   userID.String(),
		ReadDate: readDate,
	}

	const q = `
	UPDATE
		workflow.notifications
	SET
		is_read = true,
		read_date = :read_date
	WHERE
		user_id = :user_id AND is_read = false`

	rows, err := sqldb.NamedExecContextWithCount(ctx, s.log, s.db, q, data)
	if err != nil {
		return 0, fmt.Errorf("namedexeccontext: %w", err)
	}

	return int(rows), nil
}

// Ensure Store implements notificationbus.Storer.
var _ notificationbus.Storer = (*Store)(nil)
