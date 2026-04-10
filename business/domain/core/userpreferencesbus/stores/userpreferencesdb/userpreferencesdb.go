package userpreferencesdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/core/userpreferencesbus"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for user preference database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (userpreferencesbus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	return &Store{
		log: s.log,
		db:  ec,
	}, nil
}

// Upsert inserts or updates a user preference.
func (s *Store) Upsert(ctx context.Context, pref userpreferencesbus.UserPreference) error {
	const q = `
    INSERT INTO core.user_preferences (
        user_id, key, value, updated_date
    ) VALUES (
        :user_id, :key, :value, :updated_date
    )
    ON CONFLICT (user_id, key) DO UPDATE SET
        value        = EXCLUDED.value,
        updated_date = EXCLUDED.updated_date`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUserPreference(pref)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a user preference by user ID and key.
func (s *Store) Delete(ctx context.Context, userID uuid.UUID, key string) error {
	data := struct {
		UserID uuid.UUID `db:"user_id"`
		Key    string    `db:"key"`
	}{
		UserID: userID,
		Key:    key,
	}

	const q = `DELETE FROM core.user_preferences WHERE user_id = :user_id AND key = :key`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// QueryByUser retrieves all preferences for a given user.
func (s *Store) QueryByUser(ctx context.Context, userID uuid.UUID) ([]userpreferencesbus.UserPreference, error) {
	data := struct {
		UserID uuid.UUID `db:"user_id"`
	}{
		UserID: userID,
	}

	const q = `
    SELECT
        user_id, key, value, updated_date
    FROM
        core.user_preferences
    WHERE
        user_id = :user_id
    ORDER BY
        key`

	var dbPrefs []userPreference
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbPrefs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusUserPreferences(dbPrefs), nil
}

// QueryByUserAndKey retrieves a single preference by user ID and key.
func (s *Store) QueryByUserAndKey(ctx context.Context, userID uuid.UUID, key string) (userpreferencesbus.UserPreference, error) {
	data := struct {
		UserID uuid.UUID `db:"user_id"`
		Key    string    `db:"key"`
	}{
		UserID: userID,
		Key:    key,
	}

	const q = `
    SELECT
        user_id, key, value, updated_date
    FROM
        core.user_preferences
    WHERE
        user_id = :user_id AND key = :key`

	var dbPref userPreference
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbPref); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return userpreferencesbus.UserPreference{}, userpreferencesbus.ErrNotFound
		}
		return userpreferencesbus.UserPreference{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusUserPreference(dbPref), nil
}
