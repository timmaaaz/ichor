package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store reads and writes workflow.cascade_outbox rows. Every method takes the
// sqlx.ExtContext to run on explicitly, because the executor genuinely varies by
// caller: Insert runs on the originating entity write's transaction (so the row
// commits or rolls back atomically with the entity), whereas the relay's
// FetchPending/DeletePublished/MarkAttempt run on the relay's own poll transaction
// (so FOR UPDATE SKIP LOCKED holds the row lock for the dispatch). This mirrors the
// sqldb.* helpers, which all accept an sqlx.ExtContext.
type Store struct {
	log *logger.Logger
}

// NewStore constructs an outbox store.
func NewStore(log *logger.Logger) *Store {
	return &Store{log: log}
}

// dbOutbox is the database row model. last_error / published_at are nullable;
// lineage is read via COALESCE(lineage, CAST('null' AS jsonb)) so a SQL NULL never
// reaches the json.RawMessage scan (this codebase cannot Scan a NULL jsonb into
// json.RawMessage — see cascade_lineage_test.go; CAST(...) not '::' which the
// sqldb named-query parser mangles).
type dbOutbox struct {
	ID          string          `db:"id"`
	Seq         int64           `db:"seq"`
	Domain      string          `db:"domain"`
	Action      string          `db:"action"`
	EventType   string          `db:"event_type"`
	EntityName  string          `db:"entity_name"`
	Payload     json.RawMessage `db:"payload"`
	Lineage     json.RawMessage `db:"lineage"`
	CreatedAt   time.Time       `db:"created_at"`
	Attempts    int             `db:"attempts"`
	LastError   sql.NullString  `db:"last_error"`
	PublishedAt sql.NullTime    `db:"published_at"`
	Dead        bool            `db:"dead"`
}

func toOutbox(r dbOutbox) Outbox {
	o := Outbox{
		ID:         uuid.MustParse(r.ID),
		Seq:        r.Seq,
		Domain:     r.Domain,
		Action:     r.Action,
		EventType:  r.EventType,
		EntityName: r.EntityName,
		Payload:    r.Payload,
		CreatedAt:  r.CreatedAt,
		Attempts:   r.Attempts,
		Dead:       r.Dead,
	}
	// 'null'::jsonb (the COALESCE sentinel) decodes to an empty lineage; only carry
	// non-null lineage bytes forward so the relay re-hydrates a real visited-set.
	if len(r.Lineage) > 0 && string(r.Lineage) != "null" {
		o.Lineage = r.Lineage
	}
	if r.LastError.Valid {
		o.LastError = r.LastError.String
	}
	if r.PublishedAt.Valid {
		t := r.PublishedAt.Time
		o.PublishedAt = &t
	}
	return o
}

// Insert persists a pending outbox row on the given executor. seq/created_at/
// attempts/published_at/dead take their column defaults. A nil Lineage is stored
// as SQL NULL. This is the write that must share the entity write's transaction.
func (s *Store) Insert(ctx context.Context, ec sqlx.ExtContext, o Outbox) error {
	data := struct {
		ID         string          `db:"id"`
		Domain     string          `db:"domain"`
		Action     string          `db:"action"`
		EventType  string          `db:"event_type"`
		EntityName string          `db:"entity_name"`
		Payload    json.RawMessage `db:"payload"`
		Lineage    json.RawMessage `db:"lineage"`
	}{
		ID:         o.ID.String(),
		Domain:     o.Domain,
		Action:     o.Action,
		EventType:  o.EventType,
		EntityName: o.EntityName,
		Payload:    o.Payload,
		Lineage:    o.Lineage,
	}

	const q = `
	INSERT INTO workflow.cascade_outbox
		(id, domain, action, event_type, entity_name, payload, lineage)
	VALUES
		(:id, :domain, :action, :event_type, :entity_name, :payload, :lineage)`

	if err := sqldb.NamedExecContext(ctx, s.log, ec, q, data); err != nil {
		return fmt.Errorf("insert cascade_outbox: %w", err)
	}
	return nil
}

// FetchPending claims up to limit pending rows in seq order, locking them with
// FOR UPDATE SKIP LOCKED so a second relay process safely skips them. The caller
// MUST run this inside a transaction; the lock is released only on commit/rollback.
func (s *Store) FetchPending(ctx context.Context, ec sqlx.ExtContext, limit int) ([]Outbox, error) {
	data := struct {
		Limit int `db:"limit"`
	}{
		Limit: limit,
	}

	const q = `
	SELECT
		id, seq, domain, action, event_type, entity_name, payload,
		COALESCE(lineage, CAST('null' AS jsonb)) AS lineage,
		created_at, attempts, last_error, published_at, dead
	FROM
		workflow.cascade_outbox
	WHERE
		published_at IS NULL AND dead = false
	ORDER BY
		seq
	LIMIT :limit
	FOR UPDATE SKIP LOCKED`

	var rows []dbOutbox
	if err := sqldb.NamedQuerySlice(ctx, s.log, ec, q, data, &rows); err != nil {
		return nil, fmt.Errorf("fetch pending cascade_outbox: %w", err)
	}

	out := make([]Outbox, len(rows))
	for i, r := range rows {
		out[i] = toOutbox(r)
	}
	return out, nil
}

// DeletePublished removes a row after its event dispatched successfully
// (delete-on-publish — the table self-cleans for the happy path).
func (s *Store) DeletePublished(ctx context.Context, ec sqlx.ExtContext, id uuid.UUID) error {
	data := struct {
		ID string `db:"id"`
	}{
		ID: id.String(),
	}

	const q = `DELETE FROM workflow.cascade_outbox WHERE id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, ec, q, data); err != nil {
		return fmt.Errorf("delete published cascade_outbox: %w", err)
	}
	return nil
}

// MarkAttempt records a failed dispatch: it increments attempts, stores the last
// error, and sets dead when the relay has exhausted its retry budget (dead rows are
// skipped by FetchPending so they never head-of-line block the queue).
func (s *Store) MarkAttempt(ctx context.Context, ec sqlx.ExtContext, id uuid.UUID, lastErr string, dead bool) error {
	data := struct {
		ID        string `db:"id"`
		LastError string `db:"last_error"`
		Dead      bool   `db:"dead"`
	}{
		ID:        id.String(),
		LastError: lastErr,
		Dead:      dead,
	}

	const q = `
	UPDATE workflow.cascade_outbox
	SET attempts = attempts + 1, last_error = :last_error, dead = :dead
	WHERE id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, ec, q, data); err != nil {
		return fmt.Errorf("mark attempt cascade_outbox: %w", err)
	}
	return nil
}

// Reap deletes dead rows older than olderThan (the dead-row retention window) and
// returns how many were removed. Pending rows are never reaped — only rows the
// relay gave up on after exhausting retries.
func (s *Store) Reap(ctx context.Context, ec sqlx.ExtContext, olderThan time.Time) (int64, error) {
	data := struct {
		OlderThan time.Time `db:"older_than"`
	}{
		OlderThan: olderThan,
	}

	const q = `DELETE FROM workflow.cascade_outbox WHERE dead = true AND created_at < :older_than`

	n, err := sqldb.NamedExecContextWithCount(ctx, s.log, ec, q, data)
	if err != nil {
		return 0, fmt.Errorf("reap cascade_outbox: %w", err)
	}
	return n, nil
}
