package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Engine is the main workflow orchestration engine
// Singleton pattern - use NewEngine() to get instance
type Engine struct {
	log         *logger.Logger
	db          *sqlx.DB
	workflowBus *Business
	delegate    *delegate.Delegate

	// Sub-components
	triggerProcessor *TriggerProcessor
	dependencies     *DependencyResolver
	executor         *ActionExecutor
	// queue            *QueueManager

	// State management
	mu               sync.RWMutex
	isInitialized    bool
	activeExecutions map[uuid.UUID]*WorkflowExecution
	executionHistory []*WorkflowExecution
	stats            WorkflowStats
	config           WorkflowConfig
}

var (
	engineInstance *Engine
	engineOnce     sync.Once
)

// ResetEngineForTesting resets the singleton engine instance.
// This should ONLY be used in tests to ensure each test gets a fresh engine
// with its own database connection.
func ResetEngineForTesting() {
	engineInstance = nil
	engineOnce = sync.Once{}
}

// GetRegistry returns the action registry from the executor
func (e *Engine) GetRegistry() *ActionRegistry {
	if e.executor != nil {
		return e.executor.GetRegistry()
	}
	return nil
}

// GetActionExecutor returns the action executor (if you need direct access)
func (e *Engine) GetActionExecutor() *ActionExecutor {
	return e.executor
}

// NewEngine creates or returns the singleton workflow engine instance.
func NewEngine(log *logger.Logger, db *sqlx.DB, del *delegate.Delegate, workflowBus *Business) *Engine {
	engineOnce.Do(func() {
		engineInstance = &Engine{
			log:              log,
			db:               db,
			workflowBus:      workflowBus,
			delegate:         del,
			activeExecutions: make(map[uuid.UUID]*WorkflowExecution),
			executionHistory: make([]*WorkflowExecution, 0),
			config: WorkflowConfig{
				MaxParallelRules:      5,
				MaxParallelActions:    10,
				DefaultTimeout:        5 * time.Minute,
				RetryFailedActions:    true,
				MaxRetries:            3,
				StopOnCriticalFailure: true,
			},
		}
		log.Info(context.Background(), "🔄 Workflow engine instance created")
	})
	return engineInstance
}

// Initialize initializes the workflow engine and its components
func (e *Engine) Initialize(ctx context.Context, workflowBus *Business) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.isInitialized {
		return nil
	}

	e.log.Info(ctx, "Initializing workflow engine...")

	// Initialize sub-components
	e.triggerProcessor = NewTriggerProcessor(e.log, e.db, workflowBus)
	e.dependencies = NewDependencyResolver(e.log, e.db, workflowBus)
	e.executor = NewActionExecutor(e.log, e.db, workflowBus)
	// e.queue = NewQueueManager(e.log)

	// Initialize each component
	if err := e.triggerProcessor.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize trigger processor: %w", err)
	}

	if err := e.dependencies.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize dependency resolver: %w", err)
	}

	// if err := e.executor.Initialize(ctx); err != nil {
	// 	return fmt.Errorf("failed to initialize action executor: %w", err)
	// }

	// if err := e.queue.Initialize(ctx); err != nil {
	// 	return fmt.Errorf("failed to initialize queue manager: %w", err)
	// }

	// Register rule change handlers for immediate cache invalidation
	if e.delegate != nil {
		e.registerRuleChangeHandlers(ctx)
	}

	e.isInitialized = true
	e.log.Info(ctx, "✅ Workflow engine initialized successfully")

	return nil
}

// registerRuleChangeHandlers registers delegate handlers for rule lifecycle events.
// When rules are created, updated, activated, or deactivated, the trigger processor
// cache is immediately refreshed so changes take effect without waiting for cache timeout.
func (e *Engine) registerRuleChangeHandlers(ctx context.Context) {
	invalidateCache := func(ctx context.Context, data delegate.Data) error {
		if e.triggerProcessor != nil {
			if err := e.triggerProcessor.RefreshRules(ctx); err != nil {
				e.log.Error(ctx, "failed to refresh rules cache", "error", err)
			} else {
				ruleCount := e.triggerProcessor.GetActiveRuleCount()
				e.log.Info(ctx, "rules cache refreshed due to rule change", "action", data.Action, "ruleCount", ruleCount)
			}
		}
		return nil
	}

	e.delegate.Register(DomainName, ActionRuleCreated, invalidateCache)
	e.delegate.Register(DomainName, ActionRuleUpdated, invalidateCache)
	e.delegate.Register(DomainName, ActionRuleDeleted, invalidateCache)
	e.delegate.Register(DomainName, ActionRuleActivated, invalidateCache)
	e.delegate.Register(DomainName, ActionRuleDeactivated, invalidateCache)

	e.log.Info(ctx, "registered rule change handlers for cache invalidation")
}

// ExecuteWorkflow executes a complete workflow for the given trigger event
func (e *Engine) ExecuteWorkflow(ctx context.Context, event TriggerEvent) (*WorkflowExecution, error) {
	if !e.isInitialized {
		if err := e.Initialize(ctx, e.workflowBus); err != nil {
			return nil, fmt.Errorf("failed to initialize engine: %w", err)
		}
	}

	executionID := uuid.New()
	startTime := time.Now()

	// Create execution plan
	executionPlan, err := e.createExecutionPlan(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("failed to create execution plan: %w", err)
	}

	e.log.Info(ctx, "Execution plan created",
		"executionID", executionID,
		"matchedRules", executionPlan.MatchedRuleCount,
		"batches", executionPlan.TotalBatches,
	)

	// Initialize workflow execution
	workflowExecution := &WorkflowExecution{
		ExecutionID:   executionID,
		TriggerEvent:  event,
		ExecutionPlan: *executionPlan,
		CurrentBatch:  0,
		Status:        StatusPending,
		StartedAt:     startTime,
		BatchResults:  make([]BatchResult, 0),
		Errors:        make([]string, 0),
	}

	// Add to active executions
	e.mu.Lock()
	e.activeExecutions[executionID] = workflowExecution
	e.mu.Unlock()

	// Execute the workflow
	err = e.executeWorkflowInternal(ctx, workflowExecution)

	// Complete execution
	completedAt := time.Now()
	duration := completedAt.Sub(startTime)
	workflowExecution.CompletedAt = &completedAt
	workflowExecution.TotalDuration = &duration

	if err != nil {
		workflowExecution.Status = StatusFailed
		workflowExecution.Errors = append(workflowExecution.Errors, err.Error())
	} else if len(workflowExecution.Errors) > 0 {
		workflowExecution.Status = StatusFailed
	} else {
		workflowExecution.Status = StatusCompleted
	}

	// Update statistics
	e.updateStats(workflowExecution)

	// Move to history and remove from active
	e.mu.Lock()
	e.executionHistory = append(e.executionHistory, workflowExecution)
	delete(e.activeExecutions, executionID)
	e.mu.Unlock()

	return workflowExecution, nil
}

// createExecutionPlan creates an execution plan for the given trigger event
func (e *Engine) createExecutionPlan(ctx context.Context, event TriggerEvent) (*ExecutionPlan, error) {
	// Step 1: Find matching rules using trigger processor
	triggerResult, err := e.triggerProcessor.ProcessEvent(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("failed to process trigger event: %w", err)
	}

	if len(triggerResult.MatchedRules) == 0 {
		return &ExecutionPlan{
			PlanID:           uuid.New(),
			TriggerEvent:     event,
			MatchedRuleCount: 0,
			ExecutionBatches: []ExecutionBatch{},
			TotalBatches:     0,
			CreatedAt:        time.Now(),
		}, nil
	}

	// Step 2: Calculate batch order using dependencies
	batchOrder, err := e.dependencies.CalculateBatchOrder(ctx, triggerResult.MatchedRules)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate batch order: %w", err)
	}

	// Step 3: Create execution batches
	executionBatches := make([]ExecutionBatch, 0, len(batchOrder.Batches))
	totalEstimatedDuration := time.Duration(0)

	for i, batchRuleIDs := range batchOrder.Batches {
		if len(batchRuleIDs) > 0 {
			estimatedDuration := e.estimateBatchDuration(len(batchRuleIDs))
			totalEstimatedDuration += estimatedDuration

			executionBatches = append(executionBatches, ExecutionBatch{
				BatchNumber:       i,
				RuleIDs:           batchRuleIDs,
				CanRunParallel:    len(batchRuleIDs) > 1,
				EstimatedDuration: estimatedDuration,
				DependencyLevel:   i,
			})
		}
	}

	return &ExecutionPlan{
		PlanID:                 uuid.New(),
		TriggerEvent:           event,
		MatchedRuleCount:       len(triggerResult.MatchedRules),
		ExecutionBatches:       executionBatches,
		TotalBatches:           len(executionBatches),
		EstimatedTotalDuration: totalEstimatedDuration,
		CreatedAt:              time.Now(),
	}, nil
}

// executeWorkflowInternal executes the workflow batches
func (e *Engine) executeWorkflowInternal(ctx context.Context, execution *WorkflowExecution) error {
	execution.Status = StatusRunning

	for batchIndex, batch := range execution.ExecutionPlan.ExecutionBatches {
		execution.CurrentBatch = batchIndex

		batchResult, err := e.executeBatch(ctx, batch, execution.TriggerEvent, execution.ExecutionID)
		if err != nil {
			return fmt.Errorf("batch %d execution failed: %w", batchIndex, err)
		}

		execution.BatchResults = append(execution.BatchResults, *batchResult)

		// Check if we should stop on critical failure
		if batchResult.BatchStatus == "failed" && e.config.StopOnCriticalFailure {
			execution.Errors = append(execution.Errors,
				fmt.Sprintf("Critical failure in batch %d, stopping execution", batchIndex))
			break
		}
	}

	return nil
}

// executeBatch executes a batch of rules
func (e *Engine) executeBatch(ctx context.Context, batch ExecutionBatch, event TriggerEvent, executionID uuid.UUID) (*BatchResult, error) {
	batchStartTime := time.Now()
	ruleResults := make([]RuleResult, 0, len(batch.RuleIDs))

	if batch.CanRunParallel && len(batch.RuleIDs) > 1 {
		// Execute rules in parallel
		results := e.executeRulesParallel(ctx, batch.RuleIDs, event, executionID)
		ruleResults = append(ruleResults, results...)
	} else {
		// Execute rules sequentially
		for _, ruleID := range batch.RuleIDs {
			result, err := e.executeRule(ctx, ruleID, event, executionID)
			if err != nil {
				e.log.Error(ctx, "Rule execution failed",
					"ruleID", ruleID,
					"error", err)
				result = &RuleResult{
					RuleID:       ruleID,
					RuleName:     "Unknown Rule",
					Status:       "failed",
					ErrorMessage: err.Error(),
					StartedAt:    time.Now(),
				}
			}
			ruleResults = append(ruleResults, *result)
		}
	}

	// Determine batch status
	hasFailures := false
	hasSuccesses := false
	for _, result := range ruleResults {
		if result.Status == "failed" {
			hasFailures = true
		} else if result.Status == "success" {
			hasSuccesses = true
		}
	}

	batchStatus := "completed"
	if hasFailures && hasSuccesses {
		batchStatus = "partial"
	} else if hasFailures {
		batchStatus = "failed"
	}

	return &BatchResult{
		BatchNumber: batch.BatchNumber,
		RuleResults: ruleResults,
		BatchStatus: batchStatus,
		StartedAt:   batchStartTime,
		CompletedAt: time.Now(),
		Duration:    time.Since(batchStartTime),
	}, nil
}

// executeRulesParallel executes multiple rules in parallel
func (e *Engine) executeRulesParallel(ctx context.Context, ruleIDs uuid.UUIDs, event TriggerEvent, executionID uuid.UUID) []RuleResult {
	parallelLimit := e.config.MaxParallelRules
	if len(ruleIDs) < parallelLimit {
		parallelLimit = len(ruleIDs)
	}

	results := make([]RuleResult, len(ruleIDs))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, parallelLimit)

	for i, ruleID := range ruleIDs {
		wg.Add(1)
		go func(index int, id uuid.UUID) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result, err := e.executeRule(ctx, id, event, executionID)
			if err != nil {
				result = &RuleResult{
					RuleID:       id,
					RuleName:     "Unknown Rule",
					Status:       "failed",
					ErrorMessage: err.Error(),
					StartedAt:    time.Now(),
				}
			}
			results[index] = *result
		}(i, ruleID)
	}

	wg.Wait()
	return results
}

// executeRule executes a single rule with all its actions
func (e *Engine) executeRule(ctx context.Context, ruleID uuid.UUID, event TriggerEvent, executionID uuid.UUID) (*RuleResult, error) {
	ruleStartTime := time.Now()

	// Create execution context for the rule
	executionContext := ActionExecutionContext{
		EntityID:      event.EntityID,
		EntityName:    event.EntityName,
		EventType:     event.EventType,
		FieldChanges:  event.FieldChanges,
		RawData:       event.RawData,
		Timestamp:     event.Timestamp,
		UserID:        event.UserID,
		RuleID:        &ruleID,
		RuleName:      "Unknown Rule", // Will be updated by action executor
		ExecutionID:   executionID,
		TriggerSource: TriggerSourceAutomation,
	}

	// Use graph-based execution (edges define execution flow)
	batchResult, err := e.executor.ExecuteRuleActionsGraph(ctx, ruleID, executionContext)
	if err != nil {
		return nil, fmt.Errorf("failed to execute rule actions: %w", err)
	}

	completedAt := time.Now()

	// Convert to RuleResult
	return &RuleResult{
		RuleID:        ruleID,
		RuleName:      batchResult.RuleName,
		Status:        batchResult.Status,
		ActionResults: batchResult.ActionResults,
		StartedAt:     ruleStartTime,
		CompletedAt:   &completedAt,
		Duration:      time.Since(ruleStartTime),
		ErrorMessage:  batchResult.ErrorMessage,
	}, nil
}

func (e *Engine) estimateBatchDuration(ruleCount int) time.Duration {
	// Basic estimation - could be enhanced with historical data
	baseRuleDuration := 2 * time.Second
	parallelEfficiency := 0.7
	return time.Duration(float64(ruleCount) * float64(baseRuleDuration) * parallelEfficiency)
}

func (e *Engine) updateStats(execution *WorkflowExecution) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.stats.TotalWorkflowsProcessed++

	if execution.Status == StatusCompleted {
		e.stats.SuccessfulWorkflows++
	} else if execution.Status == StatusFailed {
		e.stats.FailedWorkflows++
	}

	if execution.TotalDuration != nil {
		// Update average duration
		totalTime := time.Duration(e.stats.AverageWorkflowDuration.Nanoseconds() *
			int64(e.stats.TotalWorkflowsProcessed-1))
		totalTime += *execution.TotalDuration
		e.stats.AverageWorkflowDuration = totalTime / time.Duration(e.stats.TotalWorkflowsProcessed)
	}

	// Count rules and actions
	for _, batch := range execution.BatchResults {
		e.stats.TotalRulesExecuted += len(batch.RuleResults)
		for _, rule := range batch.RuleResults {
			e.stats.TotalActionsExecuted += len(rule.ActionResults)
		}
	}

	now := time.Now()
	e.stats.LastExecutionAt = &now
}

// Public methods for monitoring and management

// GetStats returns the current workflow statistics
func (e *Engine) GetStats() WorkflowStats {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.stats
}

// GetActiveExecutions returns the currently active executions
func (e *Engine) GetActiveExecutions() map[uuid.UUID]*WorkflowExecution {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[uuid.UUID]*WorkflowExecution)
	for k, v := range e.activeExecutions {
		result[k] = v
	}
	return result
}

// GetExecutionHistory returns the execution history (limited to last N)
func (e *Engine) GetExecutionHistory(limit int) []*WorkflowExecution {
	e.mu.RLock()
	defer e.mu.RUnlock()

	historyLen := len(e.executionHistory)
	if limit > historyLen {
		limit = historyLen
	}

	// Return the last N executions
	start := historyLen - limit
	if start < 0 {
		start = 0
	}

	result := make([]*WorkflowExecution, limit)
	copy(result, e.executionHistory[start:])
	return result
}

// CancelWorkflow attempts to cancel an active workflow
func (e *Engine) CancelWorkflow(ctx context.Context, executionID uuid.UUID) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	execution, exists := e.activeExecutions[executionID]
	if !exists {
		return fmt.Errorf("execution %s not found", executionID)
	}

	execution.Status = StatusCancelled
	now := time.Now()
	execution.CompletedAt = &now
	execution.Errors = append(execution.Errors, "Workflow cancelled by user")

	// Move to history
	e.executionHistory = append(e.executionHistory, execution)
	delete(e.activeExecutions, executionID)

	return nil
}

// UpdateConfig updates the engine configuration
func (e *Engine) UpdateConfig(config WorkflowConfig) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config = config
}

// GetConfig returns the current engine configuration
func (e *Engine) GetConfig() WorkflowConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.config
}
