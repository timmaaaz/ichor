package settingsdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/config/settingsbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (settingsbus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	return &Store{
		log: s.log,
		db:  ec,
	}, nil
}

func (s *Store) Create(ctx context.Context, setting settingsbus.Setting) error {
	const q = `
    INSERT INTO config.settings (
        key, value, description, created_date, updated_date
    ) VALUES (
        :key, :value, :description, :created_date, :updated_date
    )`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBSetting(setting)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", settingsbus.ErrUniqueEntry)
		}
		return err
	}

	return nil
}

func (s *Store) Update(ctx context.Context, setting settingsbus.Setting) error {
	const q = `
    UPDATE config.settings
    SET
        value        = :value,
        description  = :description,
        updated_date = :updated_date
    WHERE
        key = :key`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBSetting(setting)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, setting settingsbus.Setting) error {
	const q = `DELETE FROM config.settings WHERE key = :key`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBSetting(setting)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Query(ctx context.Context, filter settingsbus.QueryFilter, orderBy order.By, page page.Page) ([]settingsbus.Setting, error) {
	data := map[string]interface{}{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
	    key, value, description, created_date, updated_date
	FROM
		config.settings`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbSettings []setting
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbSettings); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusSettings(dbSettings), nil
}

func (s *Store) Count(ctx context.Context, filter settingsbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        config.settings`

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

func (s *Store) QueryByKey(ctx context.Context, key string) (settingsbus.Setting, error) {
	data := struct {
		Key string `db:"key"`
	}{Key: key}

	const q = `
	SELECT
	    key, value, description, created_date, updated_date
	FROM
		config.settings
	WHERE
		key = :key`

	var dbSetting setting
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbSetting); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return settingsbus.Setting{}, settingsbus.ErrNotFound
		}
		return settingsbus.Setting{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusSetting(dbSetting), nil
}
