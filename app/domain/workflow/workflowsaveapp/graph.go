package workflowsaveapp

import (
	"fmt"
	"strconv"
	"strings"
)

// ValidateGraph validates the workflow graph structure for cycles and reachability.
// It ensures:
// 1. Exactly one start edge exists
// 2. No cycles exist in the graph
// 3. All actions are reachable from the start
func ValidateGraph(actions []SaveActionRequest, edges []SaveEdgeRequest) error {
	if len(actions) == 0 {
		return fmt.Errorf("at least one action is required")
	}

	// Count start edges
	startEdgeCount := 0
	for _, edge := range edges {
		if edge.EdgeType == "start" {
			startEdgeCount++
		}
	}

	if startEdgeCount == 0 {
		return fmt.Errorf("exactly one start edge is required")
	}
	if startEdgeCount > 1 {
		return fmt.Errorf("only one start edge is allowed, found %d", startEdgeCount)
	}

	// Build a mapping from various reference formats to a canonical node ID
	// Actions can be referenced by:
	// - "temp:N" for new actions (index into actions array)
	// - UUID string for existing actions (matching action.ID)
	// We normalize everything to temp:N format internally
	actionRefToNodeID := buildActionRefMap(actions)

	// Build adjacency list for cycle detection
	// We add a virtual "start_node" with in-degree 0 that connects to start edge targets
	adjacency := make(map[string][]string)
	inDegree := make(map[string]int)

	const virtualStartNode = "__start__"
	adjacency[virtualStartNode] = []string{}
	inDegree[virtualStartNode] = 0

	// Initialize all action nodes using canonical temp:N format
	for i := range actions {
		nodeID := fmt.Sprintf("temp:%d", i)
		adjacency[nodeID] = []string{}
		inDegree[nodeID] = 0
	}

	// Build edges using the reference map
	for _, edge := range edges {
		targetID, err := resolveActionRef(edge.TargetActionID, actionRefToNodeID, len(actions))
		if err != nil {
			return fmt.Errorf("invalid target_action_id %q: %w", edge.TargetActionID, err)
		}

		if edge.EdgeType == "start" {
			// Start edges connect from virtual start node to target
			adjacency[virtualStartNode] = append(adjacency[virtualStartNode], targetID)
			inDegree[targetID]++
			continue
		}

		sourceID, err := resolveActionRef(edge.SourceActionID, actionRefToNodeID, len(actions))
		if err != nil {
			return fmt.Errorf("invalid source_action_id %q: %w", edge.SourceActionID, err)
		}

		adjacency[sourceID] = append(adjacency[sourceID], targetID)
		inDegree[targetID]++
	}

	// Detect cycles using Kahn's algorithm (topological sort)
	if err := detectCycles(adjacency, inDegree); err != nil {
		return err
	}

	// Check reachability from start
	if err := checkReachability(actions, edges, actionRefToNodeID); err != nil {
		return err
	}

	return nil
}

// buildActionRefMap creates a mapping from various action reference formats to canonical node IDs.
// This handles both temp:N references and UUID references for existing actions.
func buildActionRefMap(actions []SaveActionRequest) map[string]string {
	refMap := make(map[string]string)

	for i, action := range actions {
		canonicalID := fmt.Sprintf("temp:%d", i)

		// Map temp:N to itself
		refMap[canonicalID] = canonicalID

		// If action has an ID (existing action), map UUID to temp:N
		if action.ID != nil && *action.ID != "" {
			refMap[*action.ID] = canonicalID
		}
	}

	return refMap
}

// resolveActionRef converts an action reference to a canonical node ID.
func resolveActionRef(ref string, refMap map[string]string, actionCount int) (string, error) {
	if ref == "" {
		return "", fmt.Errorf("action reference cannot be empty")
	}

	// Check if we have a direct mapping
	if nodeID, ok := refMap[ref]; ok {
		return nodeID, nil
	}

	// Handle temp:N format that might not be in the map yet
	if strings.HasPrefix(ref, "temp:") {
		indexStr := strings.TrimPrefix(ref, "temp:")
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			return "", fmt.Errorf("invalid temp index: %w", err)
		}
		if index < 0 || index >= actionCount {
			return "", fmt.Errorf("temp index %d out of range (0-%d)", index, actionCount-1)
		}
		return ref, nil
	}

	// Unknown reference - it's a UUID that doesn't match any action
	return "", fmt.Errorf("action reference %q not found in actions list", ref)
}

// detectCycles uses Kahn's algorithm to detect cycles in the graph.
// Returns an error if a cycle is detected.
func detectCycles(adjacency map[string][]string, inDegree map[string]int) error {
	// Make a copy of inDegree since we'll modify it
	inDegreeCopy := make(map[string]int, len(inDegree))
	for k, v := range inDegree {
		inDegreeCopy[k] = v
	}

	// Queue of nodes with no incoming edges
	var queue []string
	for node, degree := range inDegreeCopy {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	processedCount := 0
	for len(queue) > 0 {
		// Pop from queue
		node := queue[0]
		queue = queue[1:]
		processedCount++

		// Decrease in-degree of neighbors
		for _, neighbor := range adjacency[node] {
			inDegreeCopy[neighbor]--
			if inDegreeCopy[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// If we couldn't process all nodes, there's a cycle
	if processedCount != len(inDegree) {
		return fmt.Errorf("cycle detected in workflow graph")
	}

	return nil
}

// checkReachability verifies all actions are reachable from the start edge.
func checkReachability(actions []SaveActionRequest, edges []SaveEdgeRequest, actionRefToNodeID map[string]string) error {
	if len(actions) == 0 {
		return nil
	}

	// Build adjacency list using canonical temp:N format
	adjacency := make(map[string][]string)
	for i := range actions {
		adjacency[fmt.Sprintf("temp:%d", i)] = []string{}
	}

	// Find start node and build edges
	var startNode string
	for _, edge := range edges {
		if edge.EdgeType == "start" {
			targetID, _ := resolveActionRef(edge.TargetActionID, actionRefToNodeID, len(actions))
			startNode = targetID
			continue
		}

		sourceID, err := resolveActionRef(edge.SourceActionID, actionRefToNodeID, len(actions))
		if err != nil {
			continue // Already validated in ValidateGraph
		}
		targetID, err := resolveActionRef(edge.TargetActionID, actionRefToNodeID, len(actions))
		if err != nil {
			continue
		}
		adjacency[sourceID] = append(adjacency[sourceID], targetID)
	}

	if startNode == "" {
		return fmt.Errorf("no start edge found")
	}

	// BFS from start node
	visited := make(map[string]bool)
	queue := []string{startNode}
	visited[startNode] = true

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		for _, neighbor := range adjacency[node] {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}

	// Check if all actions are reachable
	for i := range actions {
		nodeID := fmt.Sprintf("temp:%d", i)
		if !visited[nodeID] {
			return fmt.Errorf("action[%d] (%s) is not reachable from start", i, actions[i].Name)
		}
	}

	return nil
}
