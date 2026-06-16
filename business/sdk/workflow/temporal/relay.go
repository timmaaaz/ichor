package temporal

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/outbox"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// =============================================================================
// Relay — the polling publisher that drains the transactional outbox
// =============================================================================
//
// The relay is the SOLE cascade dispatcher after the F5 cutover. Each poll it
// claims a batch of pending workflow.cascade_outbox rows (FOR UPDATE SKIP LOCKED,
// ORDER BY seq), rebuilds the workflow.TriggerEvent from each row — using the exact
// enrichment the DelegateHandler used to do inline (extractEntityData /
// computeFieldChanges, now sourced here from the persisted payload) — re-hydrates
// the cascade loop-guard lineage onto the dispatch context, and calls
// WorkflowTrigger.OnEntityEvent. On success the row is deleted (delete-on-publish);
// on failure attempts is bumped and the row goes dead after MaxAttempts so it never
// head-of-line blocks the queue. A reaper sweeps aged dead rows.
//
// At-least-once + dedup: dispatch happens before the delete commits. A crash between
// a successful ExecuteWorkflow and the commit leaves the row pending; it is
// re-dispatched, but the deterministic workflow id (workflow-{ruleID}-{eventID},
// eventID = row id) + REJECT_DUPLICATE collapses the retry to exactly one execution
// (trigger.go). So at-least-once emission yields effectively-once execution.
//
// A polling relay is deliberately the simplest correct design (DESIGN §2); a
// LISTEN/NOTIFY fast-path or a broker is a latency/HA swap behind the same table and
// the same OnEntityEvent boundary.

// RelayConfig tunes the relay. Zero values fall back to the defaults below, so the
// composition root can override only what it needs (ICHOR_* at the call site).
type RelayConfig struct {
	PollInterval  time.Duration // how often to poll for pending rows (default 500ms)
	BatchSize     int           // max rows claimed per poll (default 100)
	MaxAttempts   int           // dispatch attempts before a row is marked dead (default 5)
	DeadRowWindow time.Duration // how long dead rows are retained before reaping (default 7d)
	ReapInterval  time.Duration // how often to reap aged dead rows (default 1h)
}

// EventDispatcher dispatches a rebuilt cascade event into the workflow engine.
// *WorkflowTrigger satisfies it; extracting the interface lets the relay's
// drain/retry/dead/reap logic be tested with a fake dispatcher, no Temporal stack.
type EventDispatcher interface {
	OnEntityEvent(ctx context.Context, event workflow.TriggerEvent) error
}

// The production dispatcher is the same WorkflowTrigger the delegate handler used.
var _ EventDispatcher = (*WorkflowTrigger)(nil)

func (c RelayConfig) withDefaults() RelayConfig {
	if c.PollInterval <= 0 {
		c.PollInterval = 500 * time.Millisecond
	}
	if c.BatchSize <= 0 {
		c.BatchSize = 100
	}
	if c.MaxAttempts <= 0 {
		c.MaxAttempts = 5
	}
	if c.DeadRowWindow <= 0 {
		c.DeadRowWindow = 7 * 24 * time.Hour
	}
	if c.ReapInterval <= 0 {
		c.ReapInterval = time.Hour
	}
	return c
}

// Relay drains the cascade outbox into the workflow engine via an EventDispatcher.
type Relay struct {
	log        *logger.Logger
	db         *sqlx.DB
	store      *outbox.Store
	dispatcher EventDispatcher
	cfg        RelayConfig
}

// NewRelay constructs a relay. db is the base pool it opens its poll/reap
// transactions on; dispatcher is the workflow dispatcher (in production the same
// *WorkflowTrigger the delegate handler drove pre-cutover).
func NewRelay(log *logger.Logger, db *sqlx.DB, dispatcher EventDispatcher, cfg RelayConfig) *Relay {
	return &Relay{
		log:        log,
		db:         db,
		store:      outbox.NewStore(log),
		dispatcher: dispatcher,
		cfg:        cfg.withDefaults(),
	}
}

// Run polls until ctx is cancelled, draining pending rows and periodically reaping
// dead ones. Intended to be launched in a goroutine by the composition root at
// cutover. It returns ctx.Err() when stopped.
func (r *Relay) Run(ctx context.Context) error {
	r.log.Info(ctx, "cascade relay starting",
		"poll_interval", r.cfg.PollInterval, "batch_size", r.cfg.BatchSize,
		"max_attempts", r.cfg.MaxAttempts, "reap_interval", r.cfg.ReapInterval)

	poll := time.NewTicker(r.cfg.PollInterval)
	defer poll.Stop()
	reap := time.NewTicker(r.cfg.ReapInterval)
	defer reap.Stop()

	for {
		select {
		case <-ctx.Done():
			r.log.Info(ctx, "cascade relay stopping", "reason", ctx.Err())
			return ctx.Err()
		case <-poll.C:
			if _, err := r.ProcessBatch(ctx); err != nil {
				r.log.Error(ctx, "cascade relay: poll batch failed", "error", err)
			}
		case <-reap.C:
			if _, err := r.Reap(ctx); err != nil {
				r.log.Error(ctx, "cascade relay: reap failed", "error", err)
			}
		}
	}
}

// ProcessBatch claims and dispatches one batch of pending rows in a single poll
// transaction, returning how many rows were processed (dispatched + deleted, or
// marked failed). Exported so tests can drain the outbox deterministically rather
// than waiting on the poll ticker.
//
// The poll transaction holds the FOR UPDATE locks on the claimed rows for the
// duration of dispatch (an OnEntityEvent call is a Temporal RPC + an execution-record
// write on a separate connection). With a single server-side relay and small batches
// this is acceptable; SKIP LOCKED keeps a second relay correct regardless.
func (r *Relay) ProcessBatch(ctx context.Context) (int, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return 0, fmt.Errorf("begin relay tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	rows, err := r.store.FetchPending(ctx, tx, r.cfg.BatchSize)
	if err != nil {
		return 0, fmt.Errorf("fetch pending: %w", err)
	}
	if len(rows) == 0 {
		return 0, nil // rollback (no-op) via defer releases the empty tx
	}

	for _, row := range rows {
		r.dispatchRow(ctx, tx, row)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit relay tx: %w", err)
	}
	committed = true
	return len(rows), nil
}

// dispatchRow rebuilds and dispatches a single row's event, then records the
// outcome on the poll transaction (delete on success, attempt/dead on failure).
func (r *Relay) dispatchRow(ctx context.Context, tx sqlx.ExtContext, row outbox.Outbox) {
	event, ok := r.buildEvent(ctx, row)
	if !ok {
		// An undecodable payload can never succeed; retiring it immediately keeps it
		// from blocking the queue and leaves a queryable dead row.
		if err := r.store.MarkAttempt(ctx, tx, row.ID, "relay: undecodable payload", true); err != nil {
			r.log.Error(ctx, "cascade relay: mark dead failed", "id", row.ID, "error", err)
		}
		return
	}

	dispatchCtx := contextWithLineage(ctx, decodeLineage(row.Lineage))

	if err := r.dispatcher.OnEntityEvent(dispatchCtx, event); err != nil {
		attempts := row.Attempts + 1
		dead := attempts >= r.cfg.MaxAttempts
		r.log.Error(ctx, "cascade relay: dispatch failed",
			"id", row.ID, "entity", row.EntityName, "event_type", row.EventType,
			"attempts", attempts, "dead", dead, "error", err)
		if mErr := r.store.MarkAttempt(ctx, tx, row.ID, err.Error(), dead); mErr != nil {
			r.log.Error(ctx, "cascade relay: mark attempt failed", "id", row.ID, "error", mErr)
		}
		return
	}

	if err := r.store.DeletePublished(ctx, tx, row.ID); err != nil {
		// Delete failed after a successful dispatch: the row stays pending and will be
		// re-dispatched, but the deterministic workflow id + REJECT_DUPLICATE makes the
		// retry a no-op execution. Log and move on.
		r.log.Error(ctx, "cascade relay: delete-on-publish failed", "id", row.ID, "error", err)
	}
}

// buildEvent reconstructs the TriggerEvent from a persisted outbox row, mirroring
// DelegateHandler.handleEvent exactly but sourcing the delegate.Data from the row's
// payload instead of a live ctx. ok is false only when the payload itself cannot be
// decoded (a corrupt row), which the caller retires as dead.
func (r *Relay) buildEvent(ctx context.Context, row outbox.Outbox) (workflow.TriggerEvent, bool) {
	var data delegate.Data
	if err := json.Unmarshal(row.Payload, &data); err != nil {
		r.log.Error(ctx, "cascade relay: unmarshal payload failed", "id", row.ID, "error", err)
		return workflow.TriggerEvent{}, false
	}

	var params workflow.DelegateEventParams
	if err := json.Unmarshal(data.RawParams, &params); err != nil {
		r.log.Error(ctx, "cascade relay: unmarshal params failed", "id", row.ID, "error", err)
		return workflow.TriggerEvent{}, false
	}

	event := workflow.TriggerEvent{
		EventType:  row.EventType,
		EntityName: row.EntityName,
		EntityID:   params.EntityID,
		Timestamp:  row.CreatedAt,
		UserID:     params.UserID,
		EventID:    row.ID, // dedup key — trigger derives workflow-{ruleID}-{eventID}
	}

	// Extract entity raw data if present (JSON round-trip, app-layer string IDs).
	if params.Entity != nil {
		entityID, rawData, err := extractEntityData(params.Entity)
		if err == nil {
			event.RawData = rawData
			if event.EntityID == uuid.Nil {
				event.EntityID = entityID
			}
		}
	}

	// Compute FieldChanges for on_update events by diffing before/after snapshots.
	if row.EventType == workflow.EventTypeOnUpdate && params.BeforeEntity != nil && event.RawData != nil {
		_, beforeData, err := extractEntityData(params.BeforeEntity)
		if err == nil {
			event.FieldChanges = computeFieldChanges(beforeData, event.RawData)
		}
	}

	return event, true
}

// Reap deletes dead rows older than the retention window and returns the count.
func (r *Relay) Reap(ctx context.Context) (int64, error) {
	n, err := r.store.Reap(ctx, r.db, time.Now().Add(-r.cfg.DeadRowWindow))
	if err != nil {
		return 0, err
	}
	if n > 0 {
		r.log.Info(ctx, "cascade relay: reaped dead outbox rows", "count", n)
	}
	return n, nil
}

// decodeLineage turns a stored lineage JSON blob into a WorkflowLineage. A nil/empty
// blob (a human or non-workflow write) decodes to the zero value, which correctly
// starts a fresh cascade chain.
func decodeLineage(b []byte) WorkflowLineage {
	if len(b) == 0 {
		return WorkflowLineage{}
	}
	var l WorkflowLineage
	if err := json.Unmarshal(b, &l); err != nil {
		return WorkflowLineage{}
	}
	return l
}
