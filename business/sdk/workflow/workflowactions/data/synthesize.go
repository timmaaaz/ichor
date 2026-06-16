package data

import (
	"context"
	"fmt"

	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/outbox"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// fireSynthesizedEvent announces a successful generic raw-SQL write to the workflow
// trigger system (P4 M1 — update_field / create_entity / transition_status). It
// resolves the schema-qualified target to a delegate (domain, bare-entity) via the
// injected reverse map and fires on the same channel a real bus write uses, so the
// write cascades to any rule whose trigger matches it. The lineage visited-set rides
// ctx into delegate.Call, so the P1 runtime loop guard applies for free.
//
// Degrades safely: a nil delegate or map (synthesis disabled, e.g. RegisterCoreActions
// / unit tests) or a target absent from the map (whitelisted-but-unmapped) completes
// the write and fires nothing — logged, never panics, never false-cascades.
//
// Callers MUST invoke this only on a confirmed write (success path), and suppress it
// when nothing was written (records_affected == 0, invalid transition) so no phantom
// cascade fires.
//
// F2: the synthesized event is persisted DURABLY to the transactional outbox (w.Emit)
// in addition to the best-effort in-process delegate. The caller runs the write and
// this call inside one transaction (ctx carries it), so the outbox row commits or
// rolls back atomically with the write. Returning the Emit error lets the caller roll
// back. A nil Writer (pre-cutover / synthesis-disabled callers) no-ops the emit; a nil
// delegate skips the best-effort call. entityMap == nil disables synthesis entirely.
func fireSynthesizedEvent(
	ctx context.Context,
	log *logger.Logger,
	del *delegate.Delegate,
	w *outbox.Writer,
	entityMap map[string]EntityRef,
	target, action string,
	params workflow.DelegateEventParams,
) error {
	if entityMap == nil {
		return nil
	}
	ref, ok := entityMap[target]
	if !ok {
		log.Warn(ctx, "workflow synthesize: cascade skipped, no domain mapping",
			"target", target, "action", action)
		return nil
	}

	evtData := workflow.SyntheticEventData(ref.Domain, action, params)

	if err := w.Emit(ctx, evtData); err != nil {
		return fmt.Errorf("synthesize: emit cascade event: %w", err)
	}

	if del != nil {
		if err := del.Call(ctx, evtData); err != nil {
			log.Error(ctx, "workflow synthesize: delegate call failed",
				"target", target, "domain", ref.Domain, "action", action, "err", err)
		}
	}

	log.Info(ctx, "workflow synthesize: cascade event fired",
		"target", target, "domain", ref.Domain, "entity", ref.Entity, "action", action)
	return nil
}
