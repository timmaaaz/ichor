package outbox

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Writer persists cascade delegate events to the outbox. It is the explicit
// persistence step a cascade-relevant bus performs in its unit of work, right
// where it would otherwise (or additionally) call delegate.Call.
//
// Two facts an Outbox row needs are not present in delegate.Data and are supplied
// as injected dependencies so this package depends on neither the api/cmd/build
// registry nor the temporal package:
//   - entityForDomain resolves the workflow entity name from the delegate domain
//     (built from workflowdomains.Registrations() by the composition root).
//   - lineage extracts the serialized cascade loop-guard visited-set from the
//     context (it rides an unexported temporal context key; the composition root
//     injects an adapter). Returns nil for a human/non-workflow write.
//
// A nil *Writer is a no-op: until the composition root injects a real Writer at the
// F5 cutover, every bus's b.outbox is nil and Emit does nothing — so the Emit calls
// can be added across the buses ahead of time without changing behavior (the inert
// window that makes the cutover a single atomic flip — DESIGN §6).
type Writer struct {
	log             *logger.Logger
	db              *sqlx.DB
	store           *Store
	entityForDomain map[string]string
	lineage         func(context.Context) []byte
}

// NewWriter constructs a live outbox Writer. db is the base pool used only as a
// degraded fallback when no transaction rides the context. entityForDomain maps a
// delegate domain to its workflow entity name. lineage extracts the serialized
// cascade lineage from the context (nil when none / when no extractor is wired).
func NewWriter(
	log *logger.Logger,
	db *sqlx.DB,
	entityForDomain map[string]string,
	lineage func(context.Context) []byte,
) *Writer {
	return &Writer{
		log:             log,
		db:              db,
		store:           NewStore(log),
		entityForDomain: entityForDomain,
		lineage:         lineage,
	}
}

// Emit persists one cascade event in the originating write's transaction and
// returns any error so the bus can propagate it (return err) and let
// mid.BeginCommitRollback roll back both the entity row and the outbox row
// together (DESIGN §4). A nil *Writer no-ops (inert until cutover).
func (w *Writer) Emit(ctx context.Context, data delegate.Data) error {
	if w == nil {
		return nil
	}

	// bouncer: (DESIGN §2 follow-up) a volume-guard pre-filter belongs here — skip
	// emitting when no active rule listens for (domain, event_type). v1 writes
	// always and relies on delete-on-publish; correctness is identical either way.

	eventType, ok := eventTypeForAction(data.Action)
	if !ok {
		// Cascade buses only emit created/updated/deleted; anything else (e.g. a
		// rule-lifecycle action) is not a cascade event and must not become a row
		// no rule could match.
		w.log.Warn(ctx, "outbox: skip emit for non-CRUD delegate action",
			"domain", data.Domain, "action", data.Action)
		return nil
	}

	entityName := w.entityForDomain[data.Domain]
	if entityName == "" {
		// Not fatal: the row still persists, but no rule will match a blank entity.
		// Surfacing it catches a domain that emits without a Registrations() entry.
		w.log.Warn(ctx, "outbox: no entity mapping for delegate domain",
			"domain", data.Domain, "action", data.Action)
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("outbox: marshal delegate data: %w", err)
	}

	var lineage []byte
	if w.lineage != nil {
		lineage = w.lineage(ctx)
	}

	row := Outbox{
		ID:         uuid.New(),
		Domain:     data.Domain,
		Action:     data.Action,
		EventType:  eventType,
		EntityName: entityName,
		Payload:    payload,
		Lineage:    lineage,
	}

	// Run the INSERT on the originating transaction so it commits or rolls back
	// atomically with the entity write. Fall back to the base pool only when no tx
	// rides the context (a non-tx write path); that fallback is NOT atomic, so warn
	// loudly — the on-a-tx trip-wire (DESIGN §8) flags any covered path that lands
	// here unexpectedly.
	ec, ok := sqldb.GetTxExecutor(ctx)
	if !ok {
		w.log.Warn(ctx, "outbox: no transaction on context — emitting on base pool (NOT atomic with entity write)",
			"domain", data.Domain, "entity", entityName, "event_type", eventType)
		ec = w.db
	}

	if err := w.store.Insert(ctx, ec, row); err != nil {
		return fmt.Errorf("outbox emit: %w", err)
	}

	w.log.Info(ctx, "outbox: emitted cascade event",
		"id", row.ID, "domain", data.Domain, "entity", entityName,
		"event_type", eventType, "entity_id", entityIDFromData(data))

	return nil
}
