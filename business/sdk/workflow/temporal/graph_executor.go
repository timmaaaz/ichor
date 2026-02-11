package temporal

import (
	"sort"

	"github.com/google/uuid"
)

// GraphExecutor traverses the workflow graph respecting edge types and output ports.
// All operations are deterministic for Temporal replay safety.
//
// Edge traversal rules:
//   - EdgeTypeStart:    Skip (only used for initial dispatch via GetStartActions)
//   - SourceOutput=nil: Always follow (always edges, start edges)
//   - SourceOutput=X:   Follow if result["output"] == X
type GraphExecutor struct {
	graph         GraphDefinition
	actionsByID   map[uuid.UUID]ActionNode
	edgesBySource map[uuid.UUID][]ActionEdge // source_action_id -> outgoing edges (sorted by SortOrder)
	incomingCount map[uuid.UUID]int          // target_action_id -> number of incoming edges
}

// NewGraphExecutor creates an executor from a graph definition.
// Builds indexes for O(1) lookups and pre-sorts edges for deterministic traversal.
func NewGraphExecutor(graph GraphDefinition) *GraphExecutor {
	e := &GraphExecutor{
		graph:         graph,
		actionsByID:   make(map[uuid.UUID]ActionNode),
		edgesBySource: make(map[uuid.UUID][]ActionEdge),
		incomingCount: make(map[uuid.UUID]int),
	}

	// Index actions by ID.
	for _, action := range graph.Actions {
		e.actionsByID[action.ID] = action
	}

	// Index edges by source and count incoming edges.
	for _, edge := range graph.Edges {
		if edge.SourceActionID != nil {
			e.edgesBySource[*edge.SourceActionID] = append(
				e.edgesBySource[*edge.SourceActionID], edge)
		} else {
			// Start edges (source_action_id = nil) stored under uuid.Nil.
			e.edgesBySource[uuid.Nil] = append(e.edgesBySource[uuid.Nil], edge)
		}
		e.incomingCount[edge.TargetActionID]++
	}

	// Sort edges by SortOrder for deterministic execution, with UUID tie-breaking.
	// When SortOrder values are equal, sort.Slice (unstable) could produce
	// non-deterministic ordering. TargetActionID string comparison guarantees
	// consistent ordering for Temporal replay safety.
	for sourceID := range e.edgesBySource {
		edges := e.edgesBySource[sourceID]
		sort.Slice(edges, func(i, j int) bool {
			if edges[i].SortOrder != edges[j].SortOrder {
				return edges[i].SortOrder < edges[j].SortOrder
			}
			return edges[i].TargetActionID.String() < edges[j].TargetActionID.String()
		})
		e.edgesBySource[sourceID] = edges
	}

	return e
}

// GetStartActions returns all actions targeted by start edges (source_action_id = nil).
// Returns actions in SortOrder of their start edges.
func (e *GraphExecutor) GetStartActions() []ActionNode {
	startEdges := e.edgesBySource[uuid.Nil]
	actions := make([]ActionNode, 0, len(startEdges))

	for _, edge := range startEdges {
		if edge.EdgeType == EdgeTypeStart {
			if action, ok := e.actionsByID[edge.TargetActionID]; ok {
				actions = append(actions, action)
			}
		}
	}

	return actions
}

// GetNextActions returns the next actions to execute based on output ports.
//
// Routing rules:
//   - EdgeTypeStart:    Skip (start edges are only for initial dispatch)
//   - SourceOutput=nil: Always follow (always edges, sequence edges with no output constraint)
//   - SourceOutput=X:   Follow if result["output"] == X
//
// If the action result has no "output" key, it defaults to "success".
// Returns nil if no outgoing edges exist (end of path) or if edges exist
// but none match (e.g., condition with no matching branch). Callers should
// treat nil as "nothing more to execute on this path."
// Multiple returned actions indicate parallel execution.
func (e *GraphExecutor) GetNextActions(sourceActionID uuid.UUID, result map[string]any) []ActionNode {
	edges := e.edgesBySource[sourceActionID]
	if len(edges) == 0 {
		return nil
	}

	// Extract which output port the action chose.
	actionOutput, _ := result["output"].(string)
	if actionOutput == "" {
		actionOutput = "success" // Default for actions that don't set output
	}

	var nextActions []ActionNode

	for _, edge := range edges {
		shouldFollow := false

		switch {
		case edge.EdgeType == EdgeTypeStart:
			// Start edges are only used for initial dispatch via GetStartActions.
			continue

		case edge.SourceOutput == nil:
			// NULL source_output = always follow (always edges, unconstrained sequence)
			shouldFollow = true

		case *edge.SourceOutput == actionOutput:
			// Output port matches the action's chosen output
			shouldFollow = true
		}

		if shouldFollow {
			if action, ok := e.actionsByID[edge.TargetActionID]; ok {
				nextActions = append(nextActions, action)
			}
		}
	}

	if len(nextActions) == 0 {
		return nil
	}
	return nextActions
}

// FindConvergencePoint detects if multiple branches converge to a common node.
// Returns the closest common node reachable by ALL branches, or nil if no
// such node exists (fire-and-forget branches with no common downstream node).
// Returns nil for single branches or empty/nil input.
//
// IMPORTANT: This function must be deterministic for Temporal replay.
// All map iterations sort keys by UUID string before processing.
func (e *GraphExecutor) FindConvergencePoint(branches []ActionNode) *ActionNode {
	if len(branches) <= 1 {
		return nil
	}

	// Track which nodes each branch can reach.
	reachable := make(map[uuid.UUID]int) // node_id -> count of branches that reach it

	for _, branch := range branches {
		visited := e.findReachableNodes(branch.ID)
		for nodeID := range visited {
			reachable[nodeID]++
		}
	}

	// DETERMINISM: Sort node IDs before iteration.
	nodeIDs := make([]uuid.UUID, 0, len(reachable))
	for nodeID := range reachable {
		nodeIDs = append(nodeIDs, nodeID)
	}
	sort.Slice(nodeIDs, func(i, j int) bool {
		return nodeIDs[i].String() < nodeIDs[j].String()
	})

	// Find the closest node reachable by ALL branches.
	// "Closest" means the smallest maximum depth across all branches -
	// we need every branch to reach the convergence point, so the
	// limiting factor is the longest path from any branch.
	// If multiple candidates have the same maxDepth, the first in UUID
	// string sort order wins (deterministic tie-breaking for Temporal replay).
	var convergencePoint *ActionNode
	bestMaxDepth := -1

	for _, nodeID := range nodeIDs {
		count := reachable[nodeID]
		if count == len(branches) {
			// Calculate max depth across ALL branches to this candidate.
			maxDepth := 0
			for _, branch := range branches {
				depth := e.calculateMinDepth(branch.ID, nodeID)
				if depth > maxDepth {
					maxDepth = depth
				}
			}
			if bestMaxDepth == -1 || maxDepth < bestMaxDepth {
				bestMaxDepth = maxDepth
				action := e.actionsByID[nodeID]
				convergencePoint = &action
			}
		}
	}

	return convergencePoint
}

// findReachableNodes performs iterative BFS to find all nodes reachable from startID.
// Uses iterative approach (not recursive) to prevent stack overflow on deep graphs.
// Follows ALL outgoing edges regardless of type (condition results don't
// matter for reachability analysis - both branches are structurally reachable).
func (e *GraphExecutor) findReachableNodes(startID uuid.UUID) map[uuid.UUID]bool {
	visited := make(map[uuid.UUID]bool)
	// Slice-based queue is O(n) per dequeue but acceptable for typical
	// workflow graphs (< 100 nodes).
	queue := []uuid.UUID{startID}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true

		for _, edge := range e.edgesBySource[current] {
			if !visited[edge.TargetActionID] {
				queue = append(queue, edge.TargetActionID)
			}
		}
	}

	return visited
}

// calculateMinDepth calculates minimum edge count from source to target using BFS.
// Returns -1 if target is not reachable from source.
func (e *GraphExecutor) calculateMinDepth(sourceID, targetID uuid.UUID) int {
	if sourceID == targetID {
		return 0
	}

	visited := make(map[uuid.UUID]bool)
	queue := []struct {
		id    uuid.UUID
		depth int
	}{{sourceID, 0}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current.id] {
			continue
		}
		visited[current.id] = true

		for _, edge := range e.edgesBySource[current.id] {
			if edge.TargetActionID == targetID {
				return current.depth + 1
			}
			if !visited[edge.TargetActionID] {
				queue = append(queue, struct {
					id    uuid.UUID
					depth int
				}{edge.TargetActionID, current.depth + 1})
			}
		}
	}

	return -1
}

// PathLeadsTo checks if starting from actionID eventually reaches targetID.
func (e *GraphExecutor) PathLeadsTo(actionID, targetID uuid.UUID) bool {
	reachable := e.findReachableNodes(actionID)
	return reachable[targetID]
}

// Graph returns the underlying graph definition.
func (e *GraphExecutor) Graph() GraphDefinition {
	return e.graph
}

// GetAction retrieves an action by ID.
func (e *GraphExecutor) GetAction(id uuid.UUID) (ActionNode, bool) {
	action, ok := e.actionsByID[id]
	return action, ok
}

// HasMultipleIncoming returns true if the node has more than one incoming edge.
// This indicates a potential convergence point where parallel branches rejoin.
func (e *GraphExecutor) HasMultipleIncoming(actionID uuid.UUID) bool {
	return e.incomingCount[actionID] > 1
}
