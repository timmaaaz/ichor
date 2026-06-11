// Package ruleapp provides the application layer for automation-rule lifecycle
// operations whose business invariants must hold regardless of transport — today,
// the cascade-loop guard enforced when a rule is activated.
//
// It mirrors the save path (workflowsaveapp): the cascade enforcement lives in the
// app layer rather than the HTTP handler (so every caller of the app gets it, not
// just one HTTP route) and rather than the business layer (so the *ActionRegistry,
// which DetectCascadeLoops needs, stays out of the bus). The bus remains the
// primitive the app orchestrates.
package ruleapp

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// App manages the set of app-layer functions for automation-rule lifecycle operations.
type App struct {
	log         *logger.Logger
	workflowBus *workflow.Business
	registry    *workflow.ActionRegistry
}

// NewApp constructs a rule app API for use.
func NewApp(log *logger.Logger, workflowBus *workflow.Business, registry *workflow.ActionRegistry) *App {
	return &App{
		log:         log,
		workflowBus: workflowBus,
		registry:    registry,
	}
}

// Activate reactivates a rule after enforcing the cascade-loop guard: a rule that would
// close a provable re-arming cascade loop against the currently-active set is rejected
// (InvalidArgument) before the state change. Enforcement is active-only (DESIGN §10).
func (a *App) Activate(ctx context.Context, rule workflow.AutomationRule) error {
	if err := a.enforceActivateCascades(ctx, rule); err != nil {
		return err
	}
	if err := a.workflowBus.ActivateRule(ctx, rule); err != nil {
		return errs.Newf(errs.Internal, "activate rule: %s", err)
	}
	return nil
}

// Deactivate deactivates a rule. Deactivation can never create a cascade loop, so no guard
// is needed; it lives here for lifecycle symmetry with Activate, keeping the rule
// active-state transition owned by a single app-layer home.
func (a *App) Deactivate(ctx context.Context, rule workflow.AutomationRule) error {
	if err := a.workflowBus.DeactivateRule(ctx, rule); err != nil {
		return errs.Newf(errs.Internal, "deactivate rule: %s", err)
	}
	return nil
}

// enforceActivateCascades runs the static cascade-loop detector for a rule about to be
// activated, blocking activation (InvalidArgument) on a provable re-arming loop and logging
// the non-blocking warning/info tiers. Mirrors workflowsaveapp.enforceCascades.
func (a *App) enforceActivateCascades(ctx context.Context, rule workflow.AutomationRule) error {
	actions, err := a.workflowBus.QueryRoleActionsViewByRuleID(ctx, rule.ID)
	if err != nil {
		return errs.Newf(errs.Internal, "query actions: %s", err)
	}

	candActions := make([]workflow.CandidateAction, 0, len(actions))
	for _, av := range actions {
		if !av.IsActive {
			continue
		}
		candActions = append(candActions, workflow.CandidateAction{
			ActionType: av.TemplateActionType,
			Config:     av.ActionConfig,
		})
	}

	cand := workflow.CandidateRule{
		RuleID:            rule.ID,
		Name:              rule.Name,
		IsActive:          true,
		EntityID:          rule.EntityID,
		TriggerTypeID:     rule.TriggerTypeID,
		TriggerConditions: rule.TriggerConditions,
		Actions:           candActions,
	}

	analysis, err := a.workflowBus.DetectCascadeLoops(ctx, a.registry, cand)
	if err != nil {
		return errs.Newf(errs.Internal, "cascade analysis: %s", err)
	}
	if analysis.HasErrors() {
		return errs.Newf(errs.InvalidArgument, "cascade loop: %s", analysis.ErrorSummary())
	}
	for _, f := range analysis.Warnings {
		a.log.Warn(ctx, "ruleapp: possible cascade loop on activate", "rule_id", rule.ID, "detail", f.Reason)
	}
	for _, f := range analysis.Info {
		a.log.Info(ctx, "ruleapp: cascade datapoint on activate", "rule_id", rule.ID, "detail", f.Reason)
	}
	return nil
}

// CascadeMapEntry is the app-layer projection of one of a rule's actions and the active
// downstream rules its declared entity mutation would trigger. The transport layer shapes
// this into the cascade-map response DTO.
type CascadeMapEntry struct {
	ActionID       uuid.UUID
	ActionName     string
	ActionType     string
	ModifiesEntity string
	TriggersEvent  string
	ModifiedFields []string
	Downstream     []DownstreamRule
}

// DownstreamRule is one rule whose trigger an action's mutation could satisfy.
type DownstreamRule struct {
	RuleID            uuid.UUID
	RuleName          string
	EventType         string
	EntityName        string
	TriggerConditions *json.RawMessage
}

// CascadeMap returns a rule and, per action, the entity mutation it declares plus the active
// downstream rules that mutation would trigger — read-only cascade awareness for the authoring
// UI. The registry-aware derivation ("what does this action modify") and the downstream-rule
// lookup are domain concerns, so they live here rather than in the HTTP handler (which only
// formats the result). Mirrors the layering of the other rule-lifecycle operations.
func (a *App) CascadeMap(ctx context.Context, ruleID uuid.UUID) (workflow.AutomationRule, []CascadeMapEntry, error) {
	if a.registry == nil {
		return workflow.AutomationRule{}, nil, errs.Newf(errs.Internal, "action registry not configured")
	}

	rule, err := a.workflowBus.QueryRuleByID(ctx, ruleID)
	if err != nil {
		if errors.Is(err, workflow.ErrNotFound) {
			return workflow.AutomationRule{}, nil, errs.New(errs.NotFound, err)
		}
		return workflow.AutomationRule{}, nil, errs.Newf(errs.Internal, "query rule: %s", err)
	}

	actions, err := a.workflowBus.QueryRoleActionsViewByRuleID(ctx, ruleID)
	if err != nil {
		return workflow.AutomationRule{}, nil, errs.Newf(errs.Internal, "query actions: %s", err)
	}

	entries := make([]CascadeMapEntry, 0, len(actions))
	for _, action := range actions {
		entry := CascadeMapEntry{
			ActionID:   action.ID,
			ActionName: action.Name,
			ActionType: action.TemplateActionType,
		}

		handler, exists := a.registry.Get(action.TemplateActionType)
		if exists {
			if modifier, ok := handler.(workflow.EntityModifier); ok {
				for _, mod := range modifier.GetEntityModifications(action.ActionConfig) {
					entry.ModifiesEntity = mod.EntityName
					entry.TriggersEvent = mod.EventType
					entry.ModifiedFields = mod.Fields

					downstream, err := a.workflowBus.FindDownstreamRules(ctx, mod.EntityName, mod.EventType, ruleID)
					if err != nil {
						// Fail-soft: show no downstreams for this modification rather than failing
						// the whole endpoint.
						a.log.Error(ctx, "ruleapp: cascade map find downstream rules", "entity", mod.EntityName, "event", mod.EventType, "error", err)
					}
					for _, dr := range downstream {
						entry.Downstream = append(entry.Downstream, DownstreamRule{
							RuleID:            dr.ID,
							RuleName:          dr.Name,
							EventType:         mod.EventType,
							EntityName:        mod.EntityName,
							TriggerConditions: dr.TriggerConditions,
						})
					}
				}
			}
		}

		entries = append(entries, entry)
	}

	return rule, entries, nil
}
