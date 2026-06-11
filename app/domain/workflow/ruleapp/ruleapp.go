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
