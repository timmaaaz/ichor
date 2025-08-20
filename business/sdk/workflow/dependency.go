package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// DependencyNode represents a node in the dependency graph
type DependencyNode struct {
	RuleID   string   `json:"rule_id"`
	RuleName string   `json:"rule_name"`
	Parents  []string `json:"parents"`
	Children []string `json:"children"`
	Level    int      `json:"level"` // Execution batch level
}

// DependencyGraph represents the complete dependency graph
type DependencyGraph struct {
	Nodes    map[string]*DependencyNode `json:"nodes"`
	Levels   map[int][]string           `json:"levels"` // level -> rule_ids
	MaxLevel int                        `json:"max_level"`
}

// BatchOrder represents the execution order of rules
type BatchOrder struct {
	Batches           [][]string    `json:"batches"` // Array of batches, each batch contains rule_ids
	TotalBatches      int           `json:"total_batches"`
	EstimatedDuration time.Duration `json:"estimated_duration"`
}

// CycleDetectionResult represents the result of cycle detection
type CycleDetectionResult struct {
	HasCycles     bool       `json:"has_cycles"`
	Cycles        [][]string `json:"cycles"` // Array of cycles
	AffectedRules []string   `json:"affected_rules"`
}

// ValidationError represents a dependency validation error
type ValidationError struct {
	Type          string   `json:"type"` // cycle, missing_rule, self_dependency
	Message       string   `json:"message"`
	AffectedRules []string `json:"affected_rules"`
}

// DependencyResolver manages rule dependencies and execution order
type DependencyResolver struct {
	log *logger.Logger
	db  *sqlx.DB

	// Cached data
	dependencies []RuleDependency
	rules        []AutomationRule
	graph        *DependencyGraph
	lastLoadTime time.Time
	cacheTimeout time.Duration
}

// NewDependencyResolver creates a new dependency resolver
func NewDependencyResolver(log *logger.Logger, db *sqlx.DB) *DependencyResolver {
	return &DependencyResolver{
		log:          log,
		db:           db,
		cacheTimeout: 5 * time.Minute,
		graph: &DependencyGraph{
			Nodes:  make(map[string]*DependencyNode),
			Levels: make(map[int][]string),
		},
	}
}

// Initialize loads dependencies and builds the graph
func (dr *DependencyResolver) Initialize(ctx context.Context) error {
	dr.log.Info(ctx, "Initializing dependency resolver...")

	if err := dr.loadDependencies(ctx); err != nil {
		return fmt.Errorf("failed to load dependencies: %w", err)
	}

	if err := dr.buildDependencyGraph(ctx); err != nil {
		return fmt.Errorf("failed to build dependency graph: %w", err)
	}

	dr.log.Info(ctx, "Dependency resolver initialized successfully")
	return nil
}

// loadDependencies loads rule dependencies from the database
func (dr *DependencyResolver) loadDependencies(ctx context.Context) error {
	// Check cache validity
	if time.Since(dr.lastLoadTime) < dr.cacheTimeout && len(dr.dependencies) > 0 {
		return nil
	}

	// Load dependencies
	depQuery := `
        SELECT parent_rule_id, child_rule_id, created_date, updated_date
        FROM rule_dependencies
    `

	var deps []RuleDependency
	if err := dr.db.SelectContext(ctx, &deps, depQuery); err != nil {
		return fmt.Errorf("failed to load dependencies: %w", err)
	}

	// Load active rules
	ruleQuery := `
        SELECT id, name, description, entity_name, entity_type_id, 
               trigger_type_id, trigger_conditions, is_active,
               created_date, updated_date, created_by, updated_by
        FROM automation_rules
        WHERE is_active = true
    `

	var rules []AutomationRule
	if err := dr.db.SelectContext(ctx, &rules, ruleQuery); err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	dr.dependencies = deps
	dr.rules = rules
	dr.lastLoadTime = time.Now()

	dr.log.Info(ctx, "Loaded dependencies and rules",
		"dependencies", len(deps),
		"rules", len(rules))

	return nil
}

// buildDependencyGraph builds the dependency graph from loaded data
func (dr *DependencyResolver) buildDependencyGraph(ctx context.Context) error {
	// Reset graph
	dr.graph = &DependencyGraph{
		Nodes:  make(map[string]*DependencyNode),
		Levels: make(map[int][]string),
	}

	// Initialize nodes for all rules
	for _, rule := range dr.rules {
		if rule.IsActive {
			dr.graph.Nodes[rule.ID.String()] = &DependencyNode{
				RuleID:   rule.ID.String(),
				RuleName: rule.Name,
				Parents:  make([]string, 0),
				Children: make([]string, 0),
				Level:    0,
			}
		}
	}

	// Build parent/child relationships
	for _, dep := range dr.dependencies {
		parentNode, parentExists := dr.graph.Nodes[dep.ParentRuleID.String()]
		childNode, childExists := dr.graph.Nodes[dep.ChildRuleID.String()]

		if parentExists && childExists {
			parentNode.Children = append(parentNode.Children, dep.ChildRuleID.String())
			childNode.Parents = append(childNode.Parents, dep.ParentRuleID.String())
		}
	}

	// Calculate levels using topological sort
	visited := make(map[string]bool)
	calculating := make(map[string]bool)

	var calculateLevel func(ruleID string) int
	calculateLevel = func(ruleID string) int {
		if calculating[ruleID] {
			// Cycle detected during level calculation
			return 0
		}
		if visited[ruleID] {
			return dr.graph.Nodes[ruleID].Level
		}

		calculating[ruleID] = true
		node := dr.graph.Nodes[ruleID]

		maxParentLevel := -1
		for _, parentID := range node.Parents {
			parentLevel := calculateLevel(parentID)
			if parentLevel > maxParentLevel {
				maxParentLevel = parentLevel
			}
		}

		node.Level = maxParentLevel + 1
		calculating[ruleID] = false
		visited[ruleID] = true

		return node.Level
	}

	// Calculate levels for all nodes
	maxLevel := 0
	for ruleID := range dr.graph.Nodes {
		level := calculateLevel(ruleID)
		if level > maxLevel {
			maxLevel = level
		}
	}

	// Group by levels
	for _, node := range dr.graph.Nodes {
		if _, exists := dr.graph.Levels[node.Level]; !exists {
			dr.graph.Levels[node.Level] = make([]string, 0)
		}
		dr.graph.Levels[node.Level] = append(dr.graph.Levels[node.Level], node.RuleID)
	}

	dr.graph.MaxLevel = maxLevel

	return nil
}

// CalculateBatchOrder calculates the execution order for given rules
func (dr *DependencyResolver) CalculateBatchOrder(ctx context.Context, matchedRules []RuleMatchResult) (*BatchOrder, error) {
	// Ensure graph is up to date
	if err := dr.loadDependencies(ctx); err != nil {
		return nil, err
	}
	if err := dr.buildDependencyGraph(ctx); err != nil {
		return nil, err
	}

	// Create a set of matched rule IDs for quick lookup
	matchedRuleIDs := make(map[string]bool)
	for _, match := range matchedRules {
		matchedRuleIDs[match.Rule.ID.String()] = true
	}

	// Build batches based on dependency levels
	batches := make([][]string, 0)
	totalDuration := time.Duration(0)

	for level := 0; level <= dr.graph.MaxLevel; level++ {
		levelRules := dr.graph.Levels[level]
		batchRules := make([]string, 0)

		// Filter to only include matched rules
		for _, ruleID := range levelRules {
			if matchedRuleIDs[ruleID] {
				batchRules = append(batchRules, ruleID)
			}
		}

		if len(batchRules) > 0 {
			batches = append(batches, batchRules)
			// Estimate duration (can be enhanced with historical data)
			batchDuration := time.Duration(len(batchRules)) * 2 * time.Second
			totalDuration += batchDuration
		}
	}

	return &BatchOrder{
		Batches:           batches,
		TotalBatches:      len(batches),
		EstimatedDuration: totalDuration,
	}, nil
}

// DetectCycles detects cycles in the dependency graph
func (dr *DependencyResolver) DetectCycles() *CycleDetectionResult {
	cycles := make([][]string, 0)
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)
	currentPath := make([]string, 0)

	var dfs func(ruleID string)
	dfs = func(ruleID string) {
		if recursionStack[ruleID] {
			// Found a cycle - extract it from current path
			cycleStart := -1
			for i, id := range currentPath {
				if id == ruleID {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cycle := append([]string{}, currentPath[cycleStart:]...)
				cycle = append(cycle, ruleID)
				cycles = append(cycles, cycle)
			}
			return
		}

		if visited[ruleID] {
			return
		}

		visited[ruleID] = true
		recursionStack[ruleID] = true
		currentPath = append(currentPath, ruleID)

		if node, exists := dr.graph.Nodes[ruleID]; exists {
			for _, childID := range node.Children {
				dfs(childID)
			}
		}

		recursionStack[ruleID] = false
		currentPath = currentPath[:len(currentPath)-1]
	}

	// Check all nodes for cycles
	for ruleID := range dr.graph.Nodes {
		if !visited[ruleID] {
			dfs(ruleID)
		}
	}

	// Collect affected rules
	affectedRules := make(map[string]bool)
	for _, cycle := range cycles {
		for _, ruleID := range cycle {
			affectedRules[ruleID] = true
		}
	}

	affectedRulesList := make([]string, 0, len(affectedRules))
	for ruleID := range affectedRules {
		affectedRulesList = append(affectedRulesList, ruleID)
	}

	return &CycleDetectionResult{
		HasCycles:     len(cycles) > 0,
		Cycles:        cycles,
		AffectedRules: affectedRulesList,
	}
}

// ValidateDependencies validates a set of dependencies
func (dr *DependencyResolver) ValidateDependencies(newDependencies []RuleDependency) []ValidationError {
	errors := make([]ValidationError, 0)

	// Check for self-dependencies
	for _, dep := range newDependencies {
		if dep.ParentRuleID == dep.ChildRuleID {
			errors = append(errors, ValidationError{
				Type:          "self_dependency",
				Message:       fmt.Sprintf("Rule cannot depend on itself: %s", dep.ParentRuleID),
				AffectedRules: []string{dep.ParentRuleID.String()},
			})
		}
	}

	// Check for missing rules
	allRuleIDs := make(map[string]bool)
	for _, rule := range dr.rules {
		allRuleIDs[rule.ID.String()] = true
	}

	for _, dep := range newDependencies {
		if !allRuleIDs[dep.ParentRuleID.String()] {
			errors = append(errors, ValidationError{
				Type:          "missing_rule",
				Message:       fmt.Sprintf("Parent rule not found: %s", dep.ParentRuleID),
				AffectedRules: []string{dep.ParentRuleID.String()},
			})
		}
		if !allRuleIDs[dep.ChildRuleID.String()] {
			errors = append(errors, ValidationError{
				Type:          "missing_rule",
				Message:       fmt.Sprintf("Child rule not found: %s", dep.ChildRuleID),
				AffectedRules: []string{dep.ChildRuleID.String()},
			})
		}
	}

	// Check for cycles (simulate adding new dependencies)
	if len(newDependencies) > 0 {
		tempGraph := dr.simulateDependencyGraph(newDependencies)
		cycleResult := dr.detectCyclesInGraph(tempGraph)

		if cycleResult.HasCycles {
			cycleStrings := make([]string, 0)
			for _, cycle := range cycleResult.Cycles {
				cycleStr := ""
				for i, ruleID := range cycle {
					if i > 0 {
						cycleStr += " -> "
					}
					cycleStr += ruleID
				}
				cycleStrings = append(cycleStrings, cycleStr)
			}

			errors = append(errors, ValidationError{
				Type:          "cycle",
				Message:       fmt.Sprintf("Adding dependencies would create cycles: %s", cycleStrings),
				AffectedRules: cycleResult.AffectedRules,
			})
		}
	}

	return errors
}

// simulateDependencyGraph creates a temporary graph with new dependencies
func (dr *DependencyResolver) simulateDependencyGraph(newDeps []RuleDependency) *DependencyGraph {
	tempGraph := &DependencyGraph{
		Nodes:  make(map[string]*DependencyNode),
		Levels: make(map[int][]string),
	}

	// Copy existing nodes
	for ruleID, node := range dr.graph.Nodes {
		tempGraph.Nodes[ruleID] = &DependencyNode{
			RuleID:   node.RuleID,
			RuleName: node.RuleName,
			Parents:  append([]string{}, node.Parents...),
			Children: append([]string{}, node.Children...),
			Level:    node.Level,
		}
	}

	// Add new dependencies
	allDeps := append(dr.dependencies, newDeps...)

	// Rebuild relationships
	for _, dep := range allDeps {
		if parentNode, exists := tempGraph.Nodes[dep.ParentRuleID.String()]; exists {
			if childNode, exists := tempGraph.Nodes[dep.ChildRuleID.String()]; exists {
				// Check if dependency already exists
				hasChild := false
				for _, child := range parentNode.Children {
					if child == dep.ChildRuleID.String() {
						hasChild = true
						break
					}
				}
				if !hasChild {
					parentNode.Children = append(parentNode.Children, dep.ChildRuleID.String())
					childNode.Parents = append(childNode.Parents, dep.ParentRuleID.String())
				}
			}
		}
	}

	return tempGraph
}

// detectCyclesInGraph detects cycles in a specific graph
func (dr *DependencyResolver) detectCyclesInGraph(graph *DependencyGraph) *CycleDetectionResult {
	cycles := make([][]string, 0)
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)
	currentPath := make([]string, 0)

	var dfs func(ruleID string)
	dfs = func(ruleID string) {
		if recursionStack[ruleID] {
			cycleStart := -1
			for i, id := range currentPath {
				if id == ruleID {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cycle := append([]string{}, currentPath[cycleStart:]...)
				cycle = append(cycle, ruleID)
				cycles = append(cycles, cycle)
			}
			return
		}

		if visited[ruleID] {
			return
		}

		visited[ruleID] = true
		recursionStack[ruleID] = true
		currentPath = append(currentPath, ruleID)

		if node, exists := graph.Nodes[ruleID]; exists {
			for _, childID := range node.Children {
				dfs(childID)
			}
		}

		recursionStack[ruleID] = false
		currentPath = currentPath[:len(currentPath)-1]
	}

	for ruleID := range graph.Nodes {
		if !visited[ruleID] {
			dfs(ruleID)
		}
	}

	affectedRules := make(map[string]bool)
	for _, cycle := range cycles {
		for _, ruleID := range cycle {
			affectedRules[ruleID] = true
		}
	}

	affectedRulesList := make([]string, 0, len(affectedRules))
	for ruleID := range affectedRules {
		affectedRulesList = append(affectedRulesList, ruleID)
	}

	return &CycleDetectionResult{
		HasCycles:     len(cycles) > 0,
		Cycles:        cycles,
		AffectedRules: affectedRulesList,
	}
}

// GetRuleDependents returns the rules that depend on the given rule
func (dr *DependencyResolver) GetRuleDependents(ruleID string) []string {
	if node, exists := dr.graph.Nodes[ruleID]; exists {
		return node.Children
	}
	return []string{}
}

// GetRuleDependencies returns the rules that the given rule depends on
func (dr *DependencyResolver) GetRuleDependencies(ruleID string) []string {
	if node, exists := dr.graph.Nodes[ruleID]; exists {
		return node.Parents
	}
	return []string{}
}

// AddDependency adds a new dependency to the database
func (dr *DependencyResolver) AddDependency(ctx context.Context, parentRuleID, childRuleID string) error {
	prID, err := uuid.Parse(parentRuleID)
	if err != nil {
		return fmt.Errorf("adddependency: %w", err)
	}

	crID, err := uuid.Parse(childRuleID)
	if err != nil {
		return fmt.Errorf("adddependency: %w", err)
	}

	// Validate before adding
	newDep := RuleDependency{
		ParentRuleID: prID,
		ChildRuleID:  crID,
	}

	validationErrors := dr.ValidateDependencies([]RuleDependency{newDep})
	if len(validationErrors) > 0 {
		return fmt.Errorf("validation failed: %s", validationErrors[0].Message)
	}

	// Insert into database
	query := `
        INSERT INTO rule_dependencies (parent_rule_id, child_rule_id, created_date, updated_date)
        VALUES ($1, $2, NOW(), NOW())
    `

	if _, err := dr.db.ExecContext(ctx, query, parentRuleID, childRuleID); err != nil {
		return fmt.Errorf("failed to add dependency: %w", err)
	}

	// Reload to get fresh state
	dr.lastLoadTime = time.Time{} // Force reload
	if err := dr.loadDependencies(ctx); err != nil {
		return err
	}

	return dr.buildDependencyGraph(ctx)
}

// RemoveDependency removes a dependency from the database
func (dr *DependencyResolver) RemoveDependency(ctx context.Context, parentRuleID, childRuleID string) error {
	query := `
        DELETE FROM rule_dependencies 
        WHERE parent_rule_id = $1 AND child_rule_id = $2
    `

	if _, err := dr.db.ExecContext(ctx, query, parentRuleID, childRuleID); err != nil {
		return fmt.Errorf("failed to remove dependency: %w", err)
	}

	// Reload
	dr.lastLoadTime = time.Time{}
	if err := dr.loadDependencies(ctx); err != nil {
		return err
	}

	return dr.buildDependencyGraph(ctx)
}
