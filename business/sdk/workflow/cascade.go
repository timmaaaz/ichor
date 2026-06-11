package workflow

// CASCADE FILE
//
// Shared cascade-edge primitives: "which active rules would a given entity mutation
// trigger?" This is the substrate consumed by both the cascade-map endpoint (display)
// and the static loop detector (authoring-time enforcement). Keeping one
// implementation means both consumers see exactly the same edges.

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// FindDownstreamRules returns the ACTIVE automation rules whose trigger listens to the
// given (entityName, eventType), excluding excludeRuleID. An edge "this mutation could
// fire that rule" exists for every rule returned.
//
// entityName may be either a bare table name or "schema.table"; the schema prefix is
// stripped because the workflow entities table stores name and schema separately.
//
// Fail-soft: if the entity or trigger type is not registered in the workflow system,
// there are no downstream rules and (nil, nil) is returned (not an error). Only a
// genuine failure of the rule lookup returns an error.
func (b *Business) FindDownstreamRules(ctx context.Context, entityName, eventType string, excludeRuleID uuid.UUID) ([]AutomationRule, error) {
	ctx, span := otel.AddSpan(ctx, "business.workflowbus.finddownstreamrules")
	defer span.End()

	// Strip the schema prefix if entityName is "schema.table".
	tableName := entityName
	if idx := strings.LastIndex(entityName, "."); idx != -1 {
		tableName = entityName[idx+1:]
	}

	// Resolve the entity. Not registered in the workflow system → no downstream rules.
	entity, err := b.QueryEntityByName(ctx, tableName)
	if err != nil {
		b.log.Debug(ctx, "finddownstreamrules: entity not found", "entity_name", entityName, "error", err)
		return nil, nil
	}

	// Resolve the trigger type (on_create / on_update / on_delete). Unknown → no downstream rules.
	triggerType, err := b.QueryTriggerTypeByName(ctx, eventType)
	if err != nil {
		b.log.Debug(ctx, "finddownstreamrules: trigger type not found", "event_type", eventType, "error", err)
		return nil, nil
	}

	// All rules monitoring this entity.
	rules, err := b.QueryRulesByEntity(ctx, entity.ID)
	if err != nil {
		return nil, fmt.Errorf("finddownstreamrules: query rules by entity[%s]: %w", entity.ID, err)
	}

	// Keep active rules whose trigger type matches, excluding the source rule.
	downstream := make([]AutomationRule, 0, len(rules))
	for _, rule := range rules {
		if rule.ID == excludeRuleID {
			continue
		}
		if !rule.IsActive {
			continue
		}
		if rule.TriggerTypeID != triggerType.ID {
			continue
		}
		downstream = append(downstream, rule)
	}

	return downstream, nil
}
