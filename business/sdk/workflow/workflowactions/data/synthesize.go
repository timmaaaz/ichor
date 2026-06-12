package data

import (
	"context"

	"github.com/timmaaaz/ichor/business/sdk/delegate"
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
func fireSynthesizedEvent(
	ctx context.Context,
	log *logger.Logger,
	del *delegate.Delegate,
	entityMap map[string]EntityRef,
	target, action string,
	params workflow.DelegateEventParams,
) {
	if del == nil || entityMap == nil {
		return
	}
	ref, ok := entityMap[target]
	if !ok {
		log.Warn(ctx, "workflow synthesize: cascade skipped, no domain mapping",
			"target", target, "action", action)
		return
	}
	if err := del.Call(ctx, workflow.SyntheticEventData(ref.Domain, action, params)); err != nil {
		log.Error(ctx, "workflow synthesize: delegate call failed",
			"target", target, "domain", ref.Domain, "action", action, "err", err)
		return
	}
	log.Info(ctx, "workflow synthesize: cascade event fired",
		"target", target, "domain", ref.Domain, "entity", ref.Entity, "action", action)
}
