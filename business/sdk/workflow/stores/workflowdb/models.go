package workflowdb

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/sqldb/nulltypes"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// Table name constants
const (
	TableAutomationRules      = "automation_rules"
	TableRuleActions          = "rule_actions"
	TableActionTemplates      = "action_templates"
	TableRuleDependencies     = "rule_dependencies"
	TableTriggerTypes         = "trigger_types"
	TableEntityTypes          = "entity_types"
	TableEntities             = "entities"
	TableAutomationExecutions = "automation_executions"
)

// triggerType represents types of triggers (on_create, on_update, etc.)
type triggerType struct {
	ID            string         `db:"id"`
	Name          string         `db:"name"`
	Description   string         `db:"description"`
	IsActive      bool           `db:"is_active"` // Indicates if the trigger type is active
	DeactivatedBy sql.NullString `db:"deactivated_by"`
}

// toCoreTriggerType converts a store triggerType to core TriggerType
func toCoreTriggerType(dbTriggerType triggerType) workflow.TriggerType {

	deactivatedBy := uuid.Nil
	if dbTriggerType.DeactivatedBy.Valid {
		deactivatedBy = uuid.MustParse(dbTriggerType.DeactivatedBy.String)
	}

	tt := workflow.TriggerType{
		ID:            uuid.MustParse(dbTriggerType.ID),
		Name:          dbTriggerType.Name,
		Description:   dbTriggerType.Description,
		IsActive:      dbTriggerType.IsActive,
		DeactivatedBy: deactivatedBy,
	}
	return tt
}

// toDBTriggerType converts a core TriggerType to store values
func toDBTriggerType(tt workflow.TriggerType) triggerType {
	valid := false
	if tt.DeactivatedBy != uuid.Nil {
		valid = true
	}
	return triggerType{
		ID:          tt.ID.String(),
		Name:        tt.Name,
		Description: tt.Description,
		IsActive:    tt.IsActive,
		DeactivatedBy: sql.NullString{
			String: tt.DeactivatedBy.String(),
			Valid:  valid,
		},
	}
}

// entityType represents types of entities (table, view, etc.)
type entityType struct {
	ID            string         `db:"id"`
	Name          string         `db:"name"`
	Description   string         `db:"description"`
	IsActive      bool           `db:"is_active"` // Indicates if the entity type is active
	DeactivatedBy sql.NullString `db:"deactivated_by"`
}

// toCoreEntityType converts a store entityType to core EntityType
func toCoreEntityType(dbEntityType entityType) workflow.EntityType {
	deactivatedBy := uuid.Nil
	if dbEntityType.DeactivatedBy.Valid {
		deactivatedBy = uuid.MustParse(dbEntityType.DeactivatedBy.String)
	}

	et := workflow.EntityType{
		ID:            uuid.MustParse(dbEntityType.ID),
		Name:          dbEntityType.Name,
		Description:   dbEntityType.Description,
		IsActive:      dbEntityType.IsActive,
		DeactivatedBy: deactivatedBy,
	}
	return et
}

func toCoreEntityTypeSlice(dbEntityTypes []entityType) []workflow.EntityType {
	etSlice := make([]workflow.EntityType, len(dbEntityTypes))
	for i, dbET := range dbEntityTypes {
		etSlice[i] = toCoreEntityType(dbET)
	}
	return etSlice
}

// toDBEntityType converts a core EntityType to store values
func toDBEntityType(et workflow.EntityType) entityType {
	valid := false
	if et.DeactivatedBy != uuid.Nil {
		valid = true
	}

	return entityType{
		ID:          et.ID.String(),
		Name:        et.Name,
		Description: et.Description,
		IsActive:    et.IsActive,
		DeactivatedBy: sql.NullString{
			String: et.DeactivatedBy.String(),
			Valid:  valid,
		},
	}
}

// Entity represents a monitored database entity
type entity struct {
	ID            string         `db:"id"`
	Name          string         `db:"name"`
	EntityTypeID  string         `db:"entity_type_id"`
	SchemaName    string         `db:"schema_name"`
	IsActive      bool           `db:"is_active"`
	CreatedDate   time.Time      `db:"created_date"`
	DeactivatedBy sql.NullString `db:"deactivated_by"`
}

// toCoreEntity converts a store entity to core Entity
func toCoreEntity(dbEntity entity) workflow.Entity {
	deactivatedBy := uuid.Nil
	if dbEntity.DeactivatedBy.Valid {
		deactivatedBy = uuid.MustParse(dbEntity.DeactivatedBy.String)
	}

	return workflow.Entity{
		ID:            uuid.MustParse(dbEntity.ID),
		Name:          dbEntity.Name,
		EntityTypeID:  uuid.MustParse(dbEntity.EntityTypeID),
		SchemaName:    dbEntity.SchemaName,
		IsActive:      dbEntity.IsActive,
		CreatedDate:   dbEntity.CreatedDate,
		DeactivatedBy: deactivatedBy,
	}
}

func toCoreEntitySlice(dbEntities []entity) []workflow.Entity {
	entities := make([]workflow.Entity, len(dbEntities))
	for i, dbEntity := range dbEntities {
		entities[i] = toCoreEntity(dbEntity)
	}
	return entities
}

// toDBEntity converts a core Entity to store values
func toDBEntity(e workflow.Entity) entity {
	deactivatedBy := sql.NullString{
		String: e.DeactivatedBy.String(),
		Valid:  e.DeactivatedBy != uuid.Nil,
	}
	return entity{
		ID:            e.ID.String(),
		Name:          e.Name,
		EntityTypeID:  e.EntityTypeID.String(),
		SchemaName:    e.SchemaName,
		IsActive:      e.IsActive,
		CreatedDate:   time.Now(),
		DeactivatedBy: deactivatedBy,
	}
}

// automationRule represents a workflow automation rule
type automationRule struct {
	ID                string          `db:"id"`
	Name              string          `db:"name"`
	Description       string          `db:"description"`
	EntityID          string          `db:"entity_id"`
	EntityTypeID      string          `db:"entity_type_id"`
	TriggerTypeID     string          `db:"trigger_type_id"`
	TriggerConditions json.RawMessage `db:"trigger_conditions"`
	IsActive          bool            `db:"is_active"`
	CreatedDate       time.Time       `db:"created_date"`
	UpdatedDate       time.Time       `db:"updated_date"`
	CreatedBy         string          `db:"created_by"`
	UpdatedBy         string          `db:"updated_by"`
	DeactivatedBy     sql.NullString  `db:"deactivated_by"`
}

// toCoreAutomationRule converts a store automationRule to core AutomationRule
func toCoreAutomationRule(dbRule automationRule) workflow.AutomationRule {
	deactivatedBy := uuid.Nil
	if dbRule.DeactivatedBy.Valid {
		deactivatedBy = uuid.MustParse(dbRule.DeactivatedBy.String)
	}

	ar := workflow.AutomationRule{
		ID:                uuid.MustParse(dbRule.ID),
		Name:              dbRule.Name,
		Description:       dbRule.Description,
		EntityID:          uuid.MustParse(dbRule.EntityID),
		EntityTypeID:      uuid.MustParse(dbRule.EntityTypeID),
		TriggerTypeID:     uuid.MustParse(dbRule.TriggerTypeID),
		TriggerConditions: dbRule.TriggerConditions,
		IsActive:          dbRule.IsActive,
		CreatedDate:       dbRule.CreatedDate,
		UpdatedDate:       dbRule.UpdatedDate,
		CreatedBy:         uuid.MustParse(dbRule.CreatedBy),
		UpdatedBy:         uuid.MustParse(dbRule.UpdatedBy),
		DeactivatedBy:     deactivatedBy,
	}

	return ar
}

func toCoreAutomationRuleSlice(dbRules []automationRule) []workflow.AutomationRule {
	rules := make([]workflow.AutomationRule, len(dbRules))
	for i, dbRule := range dbRules {
		rules[i] = toCoreAutomationRule(dbRule)
	}
	return rules
}

// toDBAutomationRule converts a core AutomationRule to store values
func toDBAutomationRule(ar workflow.AutomationRule) automationRule {
	deactivatedBy := sql.NullString{
		String: ar.DeactivatedBy.String(),
		Valid:  ar.DeactivatedBy != uuid.Nil,
	}

	return automationRule{
		ID:                ar.ID.String(),
		Name:              ar.Name,
		Description:       ar.Description,
		EntityID:          ar.EntityID.String(),
		EntityTypeID:      ar.EntityTypeID.String(),
		TriggerTypeID:     ar.TriggerTypeID.String(),
		TriggerConditions: ar.TriggerConditions,
		IsActive:          ar.IsActive,
		CreatedDate:       ar.CreatedDate,
		UpdatedDate:       ar.UpdatedDate,
		CreatedBy:         ar.CreatedBy.String(),
		UpdatedBy:         ar.UpdatedBy.String(),
		DeactivatedBy:     deactivatedBy,
	}
}

// actionTemplate represents a reusable action configuration
type actionTemplate struct {
	ID            string          `db:"id"`
	Name          string          `db:"name"`
	Description   string          `db:"description"`
	ActionType    string          `db:"action_type"`
	DefaultConfig json.RawMessage `db:"default_config"`
	CreatedDate   time.Time       `db:"created_date"`
	CreatedBy     string          `db:"created_by"`
	IsActive      bool            `db:"is_active"`
	DeactivatedBy sql.NullString  `db:"deactivated_by"`
}

// toCoreActionTemplate converts a store ActionTemplate to core ActionTemplate
func toCoreActionTemplate(dbTemplate actionTemplate) workflow.ActionTemplate {
	deactivatedBy := uuid.Nil
	if dbTemplate.DeactivatedBy.Valid {
		deactivatedBy = uuid.MustParse(dbTemplate.DeactivatedBy.String)
	}

	at := workflow.ActionTemplate{
		ID:            uuid.MustParse(dbTemplate.ID),
		Name:          dbTemplate.Name,
		Description:   dbTemplate.Description,
		ActionType:    dbTemplate.ActionType,
		DefaultConfig: dbTemplate.DefaultConfig,
		CreatedDate:   dbTemplate.CreatedDate,
		CreatedBy:     uuid.MustParse(dbTemplate.CreatedBy),
		IsActive:      dbTemplate.IsActive,
		DeactivatedBy: deactivatedBy,
	}
	return at
}

// toDBActionTemplate converts a core ActionTemplate to store values
func toDBActionTemplate(at workflow.ActionTemplate) actionTemplate {
	deactivatedBy := sql.NullString{
		String: at.DeactivatedBy.String(),
		Valid:  at.DeactivatedBy != uuid.Nil,
	}

	return actionTemplate{
		ID:            at.ID.String(),
		Name:          at.Name,
		Description:   at.Description,
		ActionType:    at.ActionType,
		DefaultConfig: at.DefaultConfig,
		CreatedDate:   time.Now(),
		CreatedBy:     at.CreatedBy.String(),
		IsActive:      at.IsActive,
		DeactivatedBy: deactivatedBy,
	}
}

// ruleAction represents an action within a rule
type ruleAction struct {
	ID                string          `db:"id"`
	AutomationRulesID string          `db:"automation_rules_id"`
	Name              string          `db:"name"`
	Description       string          `db:"description"`
	ActionConfig      json.RawMessage `db:"action_config"`
	ExecutionOrder    int             `db:"execution_order"`
	IsActive          bool            `db:"is_active"`
	TemplateID        sql.NullString  `db:"template_id"`
}

// toCoreRuleAction converts a store ruleAction to core RuleAction
func toCoreRuleAction(dbAction ruleAction) workflow.RuleAction {
	ra := workflow.RuleAction{
		ID:               uuid.MustParse(dbAction.ID),
		AutomationRuleID: uuid.MustParse(dbAction.AutomationRulesID),
		Description:      dbAction.Description,
		Name:             dbAction.Name,
		ActionConfig:     dbAction.ActionConfig,
		ExecutionOrder:   dbAction.ExecutionOrder,
		IsActive:         dbAction.IsActive,
	}
	if dbAction.TemplateID.Valid {
		templateID := uuid.MustParse(dbAction.TemplateID.String)
		ra.TemplateID = &templateID
	}
	return ra
}

func toCoreRuleActionSlice(dbActions []ruleAction) []workflow.RuleAction {
	actions := make([]workflow.RuleAction, len(dbActions))
	for i, dbAction := range dbActions {
		actions[i] = toCoreRuleAction(dbAction)
	}
	return actions
}

// toDBRuleAction converts a core RuleAction to store values
func toDBRuleAction(ra workflow.RuleAction) ruleAction {
	dbAction := ruleAction{
		ID:                ra.ID.String(),
		AutomationRulesID: ra.AutomationRuleID.String(),
		Name:              ra.Name,
		Description:       ra.Description,
		ActionConfig:      ra.ActionConfig,
		ExecutionOrder:    ra.ExecutionOrder,
		IsActive:          ra.IsActive,
	}
	if ra.TemplateID != nil {
		dbAction.TemplateID = sql.NullString{String: ra.TemplateID.String(), Valid: true}
	}
	return dbAction
}

// ruleDependency represents a dependency between rules
type ruleDependency struct {
	ParentRuleID string `db:"parent_rule_id"`
	ChildRuleID  string `db:"child_rule_id"`
}

// toCoreRuleDependency converts a store ruleDependency to core RuleDependency
func toCoreRuleDependency(dbDep ruleDependency) workflow.RuleDependency {
	return workflow.RuleDependency{
		ParentRuleID: uuid.MustParse(dbDep.ParentRuleID),
		ChildRuleID:  uuid.MustParse(dbDep.ChildRuleID),
	}
}

func toCoreRuleDependencySlice(dbDeps []ruleDependency) []workflow.RuleDependency {
	dependencies := make([]workflow.RuleDependency, len(dbDeps))
	for i, dbDep := range dbDeps {
		dependencies[i] = toCoreRuleDependency(dbDep)
	}
	return dependencies
}

// toDBRuleDependency converts a core RuleDependency to store values
func toDBRuleDependency(rd workflow.RuleDependency) ruleDependency {
	return ruleDependency{
		ParentRuleID: rd.ParentRuleID.String(),
		ChildRuleID:  rd.ChildRuleID.String(),
	}
}

// automationExecution represents an execution record of an automation rule
type automationExecution struct {
	ID                string          `db:"id"`
	AutomationRulesID string          `db:"automation_rules_id"`
	EntityType        string          `db:"entity_type"`
	TriggerData       json.RawMessage `db:"trigger_data"`
	ActionsExecuted   json.RawMessage `db:"actions_executed"`
	Status            string          `db:"status"` // 'success', 'failed', 'partial'
	ErrorMessage      sql.NullString  `db:"error_message"`
	ExecutionTimeMs   sql.NullInt32   `db:"execution_time_ms"`
	ExecutedAt        time.Time       `db:"executed_at"`
}

// toCoreAutomationExecution converts a store automationExecution to core AutomationExecution
func toCoreAutomationExecution(dbExec automationExecution) workflow.AutomationExecution {
	ae := workflow.AutomationExecution{
		ID:               uuid.MustParse(dbExec.ID),
		AutomationRuleID: uuid.MustParse(dbExec.AutomationRulesID),
		EntityType:       dbExec.EntityType,
		TriggerData:      dbExec.TriggerData,
		ActionsExecuted:  dbExec.ActionsExecuted,
		Status:           workflow.ExecutionStatus(dbExec.Status),
		ExecutedAt:       dbExec.ExecutedAt,
	}
	if dbExec.ErrorMessage.Valid {
		ae.ErrorMessage = dbExec.ErrorMessage.String
	}
	if dbExec.ExecutionTimeMs.Valid {
		ae.ExecutionTimeMs = int(dbExec.ExecutionTimeMs.Int32)
	}
	return ae
}

func toCoreAutomationExecutionSlice(dbExecs []automationExecution) []workflow.AutomationExecution {
	aeSlice := make([]workflow.AutomationExecution, len(dbExecs))
	for i, dbExec := range dbExecs {
		aeSlice[i] = toCoreAutomationExecution(dbExec)
	}
	return aeSlice
}

// toDBAutomationExecution converts a core AutomationExecution to store values
func toDBAutomationExecution(ae workflow.AutomationExecution) automationExecution {
	dbExec := automationExecution{
		ID:                ae.ID.String(),
		AutomationRulesID: ae.AutomationRuleID.String(),
		EntityType:        ae.EntityType,
		TriggerData:       ae.TriggerData,
		ActionsExecuted:   ae.ActionsExecuted,
		Status:            string(ae.Status),
		ExecutedAt:        time.Now(),
	}
	if ae.ErrorMessage != "" {
		dbExec.ErrorMessage = sql.NullString{String: ae.ErrorMessage, Valid: true}
	}
	if ae.ExecutionTimeMs > 0 {
		dbExec.ExecutionTimeMs = sql.NullInt32{Int32: int32(ae.ExecutionTimeMs), Valid: true}
	}
	return dbExec
}

// automationRulesView is a flattened view with joined data
type automationRulesView struct {
	ID                string          `db:"id"`
	Name              string          `db:"name"`
	Description       sql.NullString  `db:"description"`
	EntityID          sql.NullString  `db:"entity_id"`
	TriggerConditions json.RawMessage `db:"trigger_conditions"`
	Actions           json.RawMessage `db:"actions"`
	IsActive          bool            `db:"is_active"`
	CreatedDate       time.Time       `db:"created_date"`
	UpdatedDate       time.Time       `db:"updated_date"`
	CreatedBy         string          `db:"created_by"`
	UpdatedBy         string          `db:"updated_by"`
	// Trigger type information
	TriggerTypeID          sql.NullString `db:"trigger_type_id"`
	TriggerTypeName        sql.NullString `db:"trigger_type_name"`
	TriggerTypeDescription sql.NullString `db:"trigger_type_description"`
	// Entity type information
	EntityTypeID          sql.NullString `db:"entity_type_id"`
	EntityTypeName        sql.NullString `db:"entity_type_name"`
	EntityTypeDescription sql.NullString `db:"entity_type_description"`
	// Entity information
	EntityName       sql.NullString `db:"entity_name"`
	EntitySchemaName sql.NullString `db:"entity_schema_name"`
}

// toCoreAutomationRuleView converts a store AutomationRulesView to core AutomationRuleView
func toCoreAutomationRuleView(dbView automationRulesView) workflow.AutomationRuleView {
	view := workflow.AutomationRuleView{
		ID:                uuid.MustParse(dbView.ID),
		Name:              dbView.Name,
		TriggerConditions: dbView.TriggerConditions,
		Actions:           dbView.Actions,
		IsActive:          dbView.IsActive,
		CreatedDate:       dbView.CreatedDate,
		UpdatedDate:       dbView.UpdatedDate,
		CreatedBy:         uuid.MustParse(dbView.CreatedBy),
		UpdatedBy:         uuid.MustParse(dbView.UpdatedBy),
	}

	// Handle nullable fields
	if dbView.Description.Valid {
		view.Description = dbView.Description.String
	}
	if dbView.EntityID.Valid {
		entityID := uuid.MustParse(dbView.EntityID.String)
		view.EntityID = &entityID
	}
	if dbView.TriggerTypeID.Valid {
		triggerTypeID := uuid.MustParse(dbView.TriggerTypeID.String)
		view.TriggerTypeID = &triggerTypeID
	}
	if dbView.TriggerTypeName.Valid {
		view.TriggerTypeName = dbView.TriggerTypeName.String
	}
	if dbView.TriggerTypeDescription.Valid {
		view.TriggerTypeDescription = dbView.TriggerTypeDescription.String
	}
	if dbView.EntityTypeID.Valid {
		entityTypeID := uuid.MustParse(dbView.EntityTypeID.String)
		view.EntityTypeID = &entityTypeID
	}
	if dbView.EntityTypeName.Valid {
		view.EntityTypeName = dbView.EntityTypeName.String
	}
	if dbView.EntityTypeDescription.Valid {
		view.EntityTypeDescription = dbView.EntityTypeDescription.String
	}
	if dbView.EntityName.Valid {
		view.EntityName = dbView.EntityName.String
	}
	if dbView.EntitySchemaName.Valid {
		view.EntitySchemaName = dbView.EntitySchemaName.String
	}

	return view
}

// toCoreAutomationRuleViews converts multiple store views to core views
func toCoreAutomationRuleViews(dbViews []automationRulesView) []workflow.AutomationRuleView {
	views := make([]workflow.AutomationRuleView, len(dbViews))
	for i, dbView := range dbViews {
		views[i] = toCoreAutomationRuleView(dbView)
	}
	return views
}

// ruleActionView is a view with template information
type ruleActionView struct {
	ID                string          `db:"id"`
	AutomationRulesID sql.NullString  `db:"automation_rules_id"`
	Name              sql.NullString  `db:"name"`
	Description       sql.NullString  `db:"description"`
	ActionConfig      json.RawMessage `db:"action_config"`
	ExecutionOrder    sql.NullInt32   `db:"execution_order"`
	IsActive          sql.NullBool    `db:"is_active"`
	TemplateID        sql.NullString  `db:"template_id"`
	// Template information
	TemplateName          sql.NullString  `db:"template_name"`
	TemplateActionType    sql.NullString  `db:"template_action_type"`
	TemplateDefaultConfig json.RawMessage `db:"template_default_config"`
}

// toCoreRuleActionView converts a store ruleActionView to core RuleActionView
func toCoreRuleActionView(dbView ruleActionView) workflow.RuleActionView {
	view := workflow.RuleActionView{
		ID:           uuid.MustParse(dbView.ID),
		ActionConfig: dbView.ActionConfig,
	}

	// Handle nullable fields
	if dbView.AutomationRulesID.Valid {
		ruleID := uuid.MustParse(dbView.AutomationRulesID.String)
		view.AutomationRuleID = &ruleID
	}
	if dbView.Name.Valid {
		view.Name = dbView.Name.String
	}
	if dbView.Description.Valid {
		view.Description = dbView.Description.String
	}
	if dbView.ExecutionOrder.Valid {
		view.ExecutionOrder = int(dbView.ExecutionOrder.Int32)
	}
	if dbView.IsActive.Valid {
		view.IsActive = dbView.IsActive.Bool
	}
	if dbView.TemplateID.Valid {
		templateID := uuid.MustParse(dbView.TemplateID.String)
		view.TemplateID = &templateID
	}
	if dbView.TemplateName.Valid {
		view.TemplateName = dbView.TemplateName.String
	}
	if dbView.TemplateActionType.Valid {
		view.TemplateActionType = dbView.TemplateActionType.String
	}
	view.TemplateDefaultConfig = dbView.TemplateDefaultConfig

	return view
}

// toCoreRuleActionViews converts multiple store views to core views
func toCoreRuleActionViews(dbViews []ruleActionView) []workflow.RuleActionView {
	views := make([]workflow.RuleActionView, len(dbViews))
	for i, dbView := range dbViews {
		views[i] = toCoreRuleActionView(dbView)
	}
	return views
}

// notificationDelivery represents a delivery record for notifications
type notificationDelivery struct {
	ID                    string          `db:"id"`
	NotificationID        string          `db:"notification_id"`
	AutomationExecutionID string          `db:"automation_execution_id"`
	RuleID                string          `db:"rule_id"`
	ActionID              string          `db:"action_id"`
	RecipientID           string          `db:"recipient_id"`
	Channel               string          `db:"channel"`
	Status                string          `db:"status"`
	Attempts              int             `db:"attempts"`
	SentAt                sql.NullTime    `db:"sent_at"`
	DeliveredAt           sql.NullTime    `db:"delivered_at"`
	FailedAt              sql.NullTime    `db:"failed_at"`
	ErrorMessage          sql.NullString  `db:"error_message"`
	ProviderResponse      json.RawMessage `db:"provider_response"`
	CreatedDate           time.Time       `db:"created_date"`
	UpdatedDate           time.Time       `db:"updated_date"`
}

func toCoreNotificationDelivery(dbDelivery notificationDelivery) workflow.NotificationDelivery {
	return workflow.NotificationDelivery{
		ID:                    uuid.MustParse(dbDelivery.ID),
		NotificationID:        uuid.MustParse(dbDelivery.NotificationID),
		AutomationExecutionID: uuid.MustParse(dbDelivery.AutomationExecutionID),
		RuleID:                uuid.MustParse(dbDelivery.RuleID),
		ActionID:              uuid.MustParse(dbDelivery.ActionID),
		RecipientID:           uuid.MustParse(dbDelivery.RecipientID),
		Channel:               dbDelivery.Channel,
		Status:                workflow.DeliveryStatus(dbDelivery.Status),
		Attempts:              dbDelivery.Attempts,
		SentAt:                nulltypes.TimePtr(dbDelivery.SentAt),
		DeliveredAt:           nulltypes.TimePtr(dbDelivery.DeliveredAt),
		FailedAt:              nulltypes.TimePtr(dbDelivery.FailedAt),
		ErrorMessage:          nulltypes.StringPtr(dbDelivery.ErrorMessage),
		ProviderResponse:      dbDelivery.ProviderResponse,
		CreatedDate:           dbDelivery.CreatedDate,
		UpdatedDate:           dbDelivery.UpdatedDate,
	}
}

func toCoreNotificationDeliverySlice(dbDeliveries []notificationDelivery) []workflow.NotificationDelivery {
	deliveries := make([]workflow.NotificationDelivery, len(dbDeliveries))
	for i, dbDelivery := range dbDeliveries {
		deliveries[i] = toCoreNotificationDelivery(dbDelivery)
	}
	return deliveries
}

// toDBNotificationDelivery converts a core NotificationDelivery to a store notificationDelivery
func toDBNotificationDelivery(delivery workflow.NotificationDelivery) notificationDelivery {
	sentAt := sql.NullTime{}
	if delivery.SentAt != nil {
		sentAt = sql.NullTime{Time: *delivery.SentAt, Valid: true}
	}

	deliveredAt := sql.NullTime{}
	if delivery.DeliveredAt != nil {
		deliveredAt = sql.NullTime{Time: *delivery.DeliveredAt, Valid: true}
	}

	failedAt := sql.NullTime{}
	if delivery.FailedAt != nil {
		failedAt = sql.NullTime{Time: *delivery.FailedAt, Valid: true}
	}

	errorMessage := sql.NullString{}
	if delivery.ErrorMessage != nil {
		errorMessage = sql.NullString{String: *delivery.ErrorMessage, Valid: true}
	}

	return notificationDelivery{
		ID:                    delivery.ID.String(),
		NotificationID:        delivery.NotificationID.String(),
		AutomationExecutionID: delivery.AutomationExecutionID.String(),
		RuleID:                delivery.RuleID.String(),
		ActionID:              delivery.ActionID.String(),
		RecipientID:           delivery.RecipientID.String(),
		Channel:               delivery.Channel,
		Status:                string(delivery.Status),
		Attempts:              delivery.Attempts,
		SentAt:                sentAt,
		DeliveredAt:           deliveredAt,
		FailedAt:              failedAt,
		ErrorMessage:          errorMessage,
		ProviderResponse:      delivery.ProviderResponse,
		CreatedDate:           delivery.CreatedDate,
		UpdatedDate:           delivery.UpdatedDate,
	}

}
