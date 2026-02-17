package agenttools

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// =============================================================================
// Parsed graph types for workflow analysis
// =============================================================================

// graphAction represents a parsed action from the API response.
type graphAction struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	ActionType  string          `json:"action_type"`
	Description string          `json:"description"`
	IsActive    bool            `json:"is_active"`
	Config      json.RawMessage `json:"action_config"`
}

// graphEdge represents a parsed edge from the API response.
type graphEdge struct {
	ID             string `json:"id"`
	SourceActionID string `json:"source_action_id"`
	TargetActionID string `json:"target_action_id"`
	EdgeType       string `json:"edge_type"`
	SourceOutput   string `json:"source_output"`
	EdgeOrder      int    `json:"edge_order"`
}

// workflowGraph holds parsed and indexed workflow graph data.
type workflowGraph struct {
	actions    []graphAction
	edges      []graphEdge
	byID       map[string]*graphAction
	byName     map[string]*graphAction
	outgoing   map[string][]graphEdge // source_action_id -> edges
	incoming   map[string][]graphEdge // target_action_id -> edges
	startNodes []string               // action IDs targeted by start edges
}

// parseWorkflowGraph parses actions and edges JSON into an indexed graph.
func parseWorkflowGraph(actionsJSON, edgesJSON json.RawMessage) (*workflowGraph, error) {
	var actions []graphAction
	if err := json.Unmarshal(actionsJSON, &actions); err != nil {
		return nil, fmt.Errorf("parsing actions: %w", err)
	}

	var edges []graphEdge
	if err := json.Unmarshal(edgesJSON, &edges); err != nil {
		return nil, fmt.Errorf("parsing edges: %w", err)
	}

	g := &workflowGraph{
		actions:  actions,
		edges:    edges,
		byID:     make(map[string]*graphAction, len(actions)),
		byName:   make(map[string]*graphAction, len(actions)),
		outgoing: make(map[string][]graphEdge),
		incoming: make(map[string][]graphEdge),
	}

	for i := range actions {
		g.byID[actions[i].ID] = &actions[i]
		g.byName[actions[i].Name] = &actions[i]
	}

	for _, edge := range edges {
		if edge.EdgeType == "start" {
			g.startNodes = append(g.startNodes, edge.TargetActionID)
		}
		if edge.SourceActionID != "" {
			g.outgoing[edge.SourceActionID] = append(g.outgoing[edge.SourceActionID], edge)
		}
		g.incoming[edge.TargetActionID] = append(g.incoming[edge.TargetActionID], edge)
	}

	// Sort outgoing edges by EdgeOrder for deterministic traversal.
	for id := range g.outgoing {
		edges := g.outgoing[id]
		sort.Slice(edges, func(i, j int) bool {
			if edges[i].EdgeOrder != edges[j].EdgeOrder {
				return edges[i].EdgeOrder < edges[j].EdgeOrder
			}
			return edges[i].TargetActionID < edges[j].TargetActionID
		})
		g.outgoing[id] = edges
	}

	return g, nil
}

// findAction returns an action by name or ID.
func (g *workflowGraph) findAction(identifier string) *graphAction {
	if a, ok := g.byName[identifier]; ok {
		return a
	}
	if a, ok := g.byID[identifier]; ok {
		return a
	}
	// Fallback: match by action_type (e.g. "create_alert" instead of the action name).
	// If multiple match, return the first one.
	for i := range g.actions {
		if g.actions[i].ActionType == identifier {
			return &g.actions[i]
		}
	}
	return nil
}

// findActionsByType returns all actions matching the given action_type.
func (g *workflowGraph) findActionsByType(actionType string) []*graphAction {
	var matches []*graphAction
	for i := range g.actions {
		if g.actions[i].ActionType == actionType {
			matches = append(matches, &g.actions[i])
		}
	}
	return matches
}

// calculateDepth returns the minimum edge count from any start node to the target.
// Returns -1 if the target is not reachable.
func (g *workflowGraph) calculateDepth(targetID string) int {
	if len(g.startNodes) == 0 {
		return -1
	}

	type entry struct {
		id    string
		depth int
	}

	visited := make(map[string]bool)
	queue := make([]entry, 0, len(g.startNodes))

	for _, startID := range g.startNodes {
		if startID == targetID {
			return 0
		}
		queue = append(queue, entry{startID, 0})
		visited[startID] = true
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, edge := range g.outgoing[current.id] {
			if edge.TargetActionID == targetID {
				return current.depth + 1
			}
			if !visited[edge.TargetActionID] {
				visited[edge.TargetActionID] = true
				queue = append(queue, entry{edge.TargetActionID, current.depth + 1})
			}
		}
	}

	return -1
}

// =============================================================================
// Summary generation
// =============================================================================

// workflowSummary holds computed summary statistics for a workflow.
type workflowSummary struct {
	NodeCount       int      `json:"node_count"`
	EdgeCount       int      `json:"edge_count"`
	BranchCount     int      `json:"branch_count"`
	ActionTypesUsed []string `json:"action_types_used"`
	FlowOutline     string   `json:"flow_outline"`
}

// computeSummary generates aggregate statistics and a flow outline.
func (g *workflowGraph) computeSummary() workflowSummary {
	// Unique action types.
	typeSet := make(map[string]bool)
	for _, a := range g.actions {
		if a.ActionType != "" {
			typeSet[a.ActionType] = true
		}
	}
	types := make([]string, 0, len(typeSet))
	for t := range typeSet {
		types = append(types, t)
	}
	sort.Strings(types)

	// Branch count: nodes with >1 outgoing edge.
	branchCount := 0
	for _, edges := range g.outgoing {
		if len(edges) > 1 {
			branchCount++
		}
	}

	return workflowSummary{
		NodeCount:       len(g.actions),
		EdgeCount:       len(g.edges),
		BranchCount:     branchCount,
		ActionTypesUsed: types,
		FlowOutline:     g.generateFlowOutline(),
	}
}

// generateFlowOutline produces a human-readable text description of the workflow flow.
func (g *workflowGraph) generateFlowOutline() string {
	if len(g.actions) == 0 {
		return "(empty workflow)"
	}
	if len(g.startNodes) == 0 {
		return "(no start edge defined)"
	}

	var lines []string
	visited := make(map[string]bool)

	for _, startID := range g.startNodes {
		g.walkNode(startID, 0, "", visited, &lines)
	}

	return strings.Join(lines, "\n")
}

// walkNode recursively traverses the graph and builds outline lines.
func (g *workflowGraph) walkNode(nodeID string, depth int, edgeLabel string, visited map[string]bool, lines *[]string) {
	action := g.byID[nodeID]
	if action == nil {
		return
	}

	prefix := strings.Repeat("  ", depth)

	label := action.Name
	if action.ActionType != "" {
		label += " (" + action.ActionType + ")"
	}
	if action.ActionType == "create_alert" {
		if rs := formatAlertRecipients(action.Config); rs != "" {
			label += " — alerts: " + rs
		}
	}
	if edgeLabel != "" {
		label = "[" + edgeLabel + "] " + label
	}
	*lines = append(*lines, prefix+label)

	if visited[nodeID] {
		*lines = append(*lines, prefix+"  ↻ (continues above)")
		return
	}
	visited[nodeID] = true

	outgoing := g.outgoing[nodeID]
	if len(outgoing) == 0 {
		return
	}

	if len(outgoing) == 1 {
		g.walkNode(outgoing[0].TargetActionID, depth, formatEdgeLabel(outgoing[0]), visited, lines)
		return
	}

	for _, edge := range outgoing {
		g.walkNode(edge.TargetActionID, depth+1, formatEdgeLabel(edge), visited, lines)
	}
}

// formatEdgeLabel creates a display label for an edge.
func formatEdgeLabel(edge graphEdge) string {
	if edge.SourceOutput != "" {
		return edge.SourceOutput
	}
	if edge.EdgeType == "always" {
		return "always"
	}
	return ""
}

// formatAlertRecipients extracts a human-readable recipient summary from an
// enriched create_alert config. Returns empty string if parsing fails or there
// are no recipients.
func formatAlertRecipients(config json.RawMessage) string {
	var cfg struct {
		Recipients struct {
			Users []struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"users"`
			Roles []struct {
				Name string `json:"name"`
			} `json:"roles"`
		} `json:"recipients"`
	}
	if err := json.Unmarshal(config, &cfg); err != nil {
		return ""
	}

	var parts []string
	for _, u := range cfg.Recipients.Users {
		if u.Name == "" {
			continue
		}
		if u.Email != "" {
			parts = append(parts, u.Name+" ("+u.Email+")")
		} else {
			parts = append(parts, u.Name)
		}
	}
	for _, r := range cfg.Recipients.Roles {
		if r.Name != "" {
			parts = append(parts, r.Name+" (role)")
		}
	}

	return strings.Join(parts, ", ")
}

// =============================================================================
// Node explanation
// =============================================================================

// edgeDescription describes an edge in human-readable terms.
type edgeDescription struct {
	ActionID   string `json:"action_id"`
	ActionName string `json:"action_name"`
	EdgeType   string `json:"edge_type"`
	OutputPort string `json:"output_port,omitempty"`
}

// nodeExplanation holds detailed information about a single workflow node.
type nodeExplanation struct {
	Action         graphAction       `json:"action"`
	DepthFromStart int               `json:"depth_from_start"`
	IncomingFrom   []edgeDescription `json:"incoming_from"`
	OutgoingTo     []edgeDescription `json:"outgoing_to"`
	ActionTypeInfo json.RawMessage   `json:"action_type_info,omitempty"`
}

// explainNode builds a detailed explanation for a single action node.
func (g *workflowGraph) explainNode(action *graphAction) nodeExplanation {
	var incoming []edgeDescription
	for _, edge := range g.incoming[action.ID] {
		desc := edgeDescription{
			EdgeType:   edge.EdgeType,
			OutputPort: edge.SourceOutput,
		}
		if src := g.byID[edge.SourceActionID]; src != nil {
			desc.ActionID = src.ID
			desc.ActionName = src.Name
		} else if edge.EdgeType == "start" {
			desc.ActionName = "(trigger)"
		}
		incoming = append(incoming, desc)
	}

	var outgoing []edgeDescription
	for _, edge := range g.outgoing[action.ID] {
		desc := edgeDescription{
			EdgeType:   edge.EdgeType,
			OutputPort: edge.SourceOutput,
		}
		if tgt := g.byID[edge.TargetActionID]; tgt != nil {
			desc.ActionID = tgt.ID
			desc.ActionName = tgt.Name
		}
		outgoing = append(outgoing, desc)
	}

	return nodeExplanation{
		Action:         *action,
		DepthFromStart: g.calculateDepth(action.ID),
		IncomingFrom:   incoming,
		OutgoingTo:     outgoing,
	}
}
