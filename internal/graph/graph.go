package graph

import (
    "fmt"
    "github.com/Kartik-r/design-to-code/pkg/types"
)

type Graph struct {
    nodesByID    map[string]*types.Node
    edges        []*types.Edge
    adjacencyOut map[string][]*types.Edge  // nodeID → edges going OUT
    adjacencyIn  map[string][]*types.Edge  // nodeID → edges coming IN
}

func New() *Graph {
    return &Graph{
        nodesByID:    make(map[string]*types.Node),
        edges:        make([]*types.Edge, 0),
        adjacencyOut: make(map[string][]*types.Edge),
        adjacencyIn:  make(map[string][]*types.Edge),
    }
}

func (g *Graph) AddNode(n *types.Node) {
    if n == nil { return }
    if _, exists := g.nodesByID[n.ID]; !exists {
        g.nodesByID[n.ID] = n
    }
}

func (g *Graph) AddEdge(e *types.Edge) {
    if e == nil { return }
    g.edges = append(g.edges, e)
    g.adjacencyOut[e.From] = append(g.adjacencyOut[e.From], e)
    g.adjacencyIn[e.To] = append(g.adjacencyIn[e.To], e)
}

func (g *Graph) GetNode(id string) *types.Node { return g.nodesByID[id] }
func (g *Graph) HasNode(id string) bool { _, ok := g.nodesByID[id]; return ok }
func (g *Graph) NodeCount() int { return len(g.nodesByID) }
func (g *Graph) EdgeCount() int { return len(g.edges) }

func (g *Graph) GetAllNodes() []*types.Node {
    nodes := make([]*types.Node, 0, len(g.nodesByID))
    for _, n := range g.nodesByID { nodes = append(nodes, n) }
    return nodes
}

func (g *Graph) GetAllEdges() []*types.Edge { return g.edges }

func (g *Graph) NodeCountByType() map[types.NodeType]int {
    counts := make(map[types.NodeType]int)
    for _, n := range g.nodesByID { counts[n.Type]++ }
    return counts
}

func (g *Graph) Summary() string {
    return fmt.Sprintf("Graph: %d nodes, %d edges", g.NodeCount(), g.EdgeCount())
}

// GetDependencies returns all nodes that nodeID directly calls or imports.
func (g *Graph) GetDependencies(nodeID string) []*types.Node {
	outEdges := g.adjacencyOut[nodeID]
	result := make([]*types.Node, 0, len(outEdges))
	seen := make(map[string]bool)
	for _, edge := range outEdges {
		if !seen[edge.To] {
			seen[edge.To] = true
			if n := g.nodesByID[edge.To]; n != nil {
				result = append(result, n)
			}
		}
	}
	return result
}

// GetCallers returns all nodes that directly call or import nodeID.
func (g *Graph) GetCallers(nodeID string) []*types.Node {
	inEdges := g.adjacencyIn[nodeID]
	result := make([]*types.Node, 0, len(inEdges))
	seen := make(map[string]bool)
	for _, edge := range inEdges {
		if !seen[edge.From] {
			seen[edge.From] = true
			if n := g.nodesByID[edge.From]; n != nil {
				result = append(result, n)
			}
		}
	}
	return result
}

// GetImpacted returns all nodes transitively affected if nodeID changes.
// Uses BFS following incoming edges — "who calls the callers?"
func (g *Graph) GetImpacted(nodeID string) []*types.Node {
	visited := make(map[string]bool)
	queue := []string{nodeID}
	result := make([]*types.Node, 0)
	visited[nodeID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, edge := range g.adjacencyIn[current] {
			if !visited[edge.From] {
				visited[edge.From] = true
				if n := g.nodesByID[edge.From]; n != nil {
					result = append(result, n)
				}
				queue = append(queue, edge.From)
			}
		}
	}
	return result
}

// GetTransitiveDependencies returns all nodes nodeID transitively depends on.
// Uses BFS following outgoing edges.
func (g *Graph) GetTransitiveDependencies(nodeID string) []*types.Node {
	visited := make(map[string]bool)
	queue := []string{nodeID}
	result := make([]*types.Node, 0)
	visited[nodeID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, edge := range g.adjacencyOut[current] {
			if !visited[edge.To] {
				visited[edge.To] = true
				if n := g.nodesByID[edge.To]; n != nil {
					result = append(result, n)
				}
				queue = append(queue, edge.To)
			}
		}
	}
	return result
}