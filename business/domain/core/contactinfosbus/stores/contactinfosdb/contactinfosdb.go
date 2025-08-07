package contactinfosdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"

	"github.com/jmoiron/sqlx"
)

// Store manages the set of APIs for assets database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (contactinfosbus.Storer, error) {
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

// Create inserts a new user asset into the database.
func (s *Store) Create(ctx context.Context, ass contactinfosbus.ContactInfos) error {
	const q = `
    INSERT INTO contact_infos (
        id, first_name, last_name, email_address, primary_phone_number, secondary_phone_number, street_id,
		available_hours_start, available_hours_end, timezone, preferred_contact_type, notes
    ) VALUES (
		:id, :first_name, :last_name, :email_address, :primary_phone_number, :secondary_phone_number, :street_id,
		:available_hours_start, :available_hours_end, :timezone, :preferred_contact_type, :notes
	)
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBContactInfos(ass)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", contactinfosbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Update replaces a user asset document in the database.
func (s *Store) Update(ctx context.Context, ass contactinfosbus.ContactInfos) error {
	const q = `
	UPDATE
		contact_infos
	SET
		id = :id,
		first_name = :first_name,
        last_name = :last_name,
        primary_phone_number = :primary_phone_number,
        email_address = :email_address,
        street_id = :street_id,
		secondary_phone_number = :secondary_phone_number,
        available_hours_start = :available_hours_start,
        available_hours_end = :available_hours_end,
        timezone = :timezone,
        preferred_contact_type = :preferred_contact_type,
		notes = :notes
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBContactInfos(ass)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", contactinfosbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Delete removes an user asset from the database.
func (s *Store) Delete(ctx context.Context, ass contactinfosbus.ContactInfos) error {
	const q = `
	DELETE FROM
		contact_infos
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBContactInfos(ass)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of user assets from the database.
func (s *Store) Query(ctx context.Context, filter contactinfosbus.QueryFilter, orderBy order.By, page page.Page) ([]contactinfosbus.ContactInfos, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
    SELECT
		id, first_name, last_name, email_address, primary_phone_number, street_id, 
		secondary_phone_number, available_hours_start, available_hours_end, timezone, preferred_contact_type, notes
    FROM
        contact_infos`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var ci []contactInfos
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &ci); err != nil {
		return nil, fmt.Errorf("namedselectcontext: %w", err)
	}

	return toBusContactInfoss(ci), nil
}

// Count returns the number of assets in the database.
func (s *Store) Count(ctx context.Context, filter contactinfosbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        contact_infos`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedqueryint: %w", err)
	}

	return count.Count, nil
}

// QueryByID retrieves a single asset from the database by its ID.
func (s *Store) QueryByID(ctx context.Context, userContactInfosID uuid.UUID) (contactinfosbus.ContactInfos, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: userContactInfosID.String(),
	}

	const q = `
    SELECT
        id, first_name, last_name, email_address, primary_phone_number, street_id,
		secondary_phone_number, available_hours_start, available_hours_end, timezone, preferred_contact_type, notes
    FROM
        contact_infos
    WHERE
        id = :id
    `
	var ci contactInfos

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &ci); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return contactinfosbus.ContactInfos{}, contactinfosbus.ErrNotFound
		}
		return contactinfosbus.ContactInfos{}, fmt.Errorf("querystruct: %w", err)
	}

	return toBusContactInfos(ci), nil
}
